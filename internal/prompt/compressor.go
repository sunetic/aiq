package prompt

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/aiq/aiq/internal/llm"
	"github.com/aiq/aiq/internal/skills"
)

const (
	// Compression thresholds as percentages of context window
	ThresholdCompressHistory = 0.80 // 80% - start compressing conversation history
	ThresholdEvictSkills     = 0.90 // 90% - start evicting low-priority Skills
	ThresholdAggressive      = 0.95 // 95% - aggressive compression

	// Default context window size (conservative estimate for most models)
	DefaultContextWindow = 100000 // 100k tokens
)

// Compressor manages prompt compression to stay within token limits
type Compressor struct {
	contextWindow int
	llmClient     *llm.Client
	compressionCache map[string]string // cache key: content hash, value: compressed content
	cacheMu       sync.RWMutex
}

// NewCompressor creates a new prompt compressor
func NewCompressor(contextWindow int) *Compressor {
	if contextWindow <= 0 {
		contextWindow = DefaultContextWindow
	}
	return &Compressor{
		contextWindow:    contextWindow,
		compressionCache: make(map[string]string),
	}
}

// SetLLMClient sets the LLM client for semantic compression
func (c *Compressor) SetLLMClient(client *llm.Client) {
	c.llmClient = client
}

// CompressionResult represents the result of compression
type CompressionResult struct {
	CompressedHistory []string
	RemainingSkills   []*skills.Skill
	Compressed        bool
}

// Compress compresses a prompt if it exceeds token thresholds
func (c *Compressor) Compress(
	conversationHistory []string,
	loadedSkills []*skills.Skill,
	systemPrompt string,
	currentQuery string,
) (*CompressionResult, error) {
	// Estimate total tokens
	totalTokens := EstimatePromptTokens(
		systemPrompt,
		strings.Join(conversationHistory, "\n"),
		currentQuery,
	)

	result := &CompressionResult{
		CompressedHistory: conversationHistory,
		RemainingSkills:   loadedSkills,
		Compressed:        false,
	}

	threshold := float64(totalTokens) / float64(c.contextWindow)

	// 80% threshold: Start LLM compression (moderate compression)
	if threshold >= ThresholdCompressHistory {
		if c.llmClient != nil {
			// Try LLM compression first
			compressedHistory, err := c.CompressHistoryWithLLM(context.Background(), conversationHistory, 0.5) // Target 50% reduction
			if err == nil {
				result.CompressedHistory = compressedHistory
				result.Compressed = true
			} else {
				// Fallback to simple truncation
				result.CompressedHistory = c.compressHistory(conversationHistory, 10) // Keep last 10 messages
				result.Compressed = true
			}
		} else {
			// No LLM client, use simple compression
			result.CompressedHistory = c.compressHistory(conversationHistory, 10) // Keep last 10 messages
			result.Compressed = true
		}

		// Re-estimate after compression
		totalTokens = EstimatePromptTokens(
			systemPrompt,
			strings.Join(result.CompressedHistory, "\n"),
			currentQuery,
		)
		threshold = float64(totalTokens) / float64(c.contextWindow)
	}

	// 90% threshold: Aggressive LLM compression + evict low-priority Skills
	if threshold >= ThresholdEvictSkills {
		if c.llmClient != nil {
			// More aggressive LLM compression
			compressedHistory, err := c.CompressHistoryWithLLM(context.Background(), conversationHistory, 0.3) // Target 70% reduction
			if err == nil {
				result.CompressedHistory = compressedHistory
			} else {
				result.CompressedHistory = c.compressHistory(conversationHistory, 5) // Keep only last 5 messages
			}
		} else {
			result.CompressedHistory = c.compressHistory(conversationHistory, 5) // Keep only last 5 messages
		}
		result.RemainingSkills = c.evictLowPrioritySkills(loadedSkills, skills.PriorityRelevant)
		result.Compressed = true

		totalTokens = EstimatePromptTokens(
			systemPrompt,
			strings.Join(result.CompressedHistory, "\n"),
			currentQuery,
		)
		threshold = float64(totalTokens) / float64(c.contextWindow)
	}

	// 95% threshold: Maximum compression (keep only essential context)
	if threshold >= ThresholdAggressive {
		if c.llmClient != nil {
			// Maximum LLM compression
			compressedHistory, err := c.CompressHistoryWithLLM(context.Background(), conversationHistory, 0.2) // Target 80% reduction
			if err == nil {
				result.CompressedHistory = compressedHistory
			} else {
				result.CompressedHistory = c.compressHistory(conversationHistory, 3) // Keep only last 3 messages
			}

			// Compress Skills content if needed
			if len(result.RemainingSkills) > 0 {
				compressedSkills, err := c.CompressSkillsWithLLM(context.Background(), result.RemainingSkills)
				if err == nil {
					result.RemainingSkills = compressedSkills
				}
			}
		} else {
			result.CompressedHistory = c.compressHistory(conversationHistory, 3) // Keep only last 3 messages
		}
		result.RemainingSkills = c.evictLowPrioritySkills(loadedSkills, skills.PriorityActive) // Keep only active Skills
		result.Compressed = true
	}

	return result, nil
}

// compressHistory compresses conversation history, keeping only the last N messages
func (c *Compressor) compressHistory(history []string, keepLast int) []string {
	if len(history) <= keepLast {
		return history
	}

	// Keep last N messages
	kept := history[len(history)-keepLast:]

	// Summarize the rest
	toSummarize := history[:len(history)-keepLast]
	summary := c.summarizeHistory(toSummarize)

	// Return summary + kept messages
	result := []string{summary}
	result = append(result, kept...)

	return result
}

// summarizeHistory creates a summary of conversation history
// Simple implementation: just indicate how many messages were compressed
func (c *Compressor) summarizeHistory(messages []string) string {
	return fmt.Sprintf("[Previous conversation: %d messages compressed]", len(messages))
}

// CompressHistoryWithLLM uses LLM to semantically compress conversation history
func (c *Compressor) CompressHistoryWithLLM(ctx context.Context, history []string, targetRatio float64) ([]string, error) {
	if c.llmClient == nil {
		return nil, fmt.Errorf("LLM client not set")
	}

	if len(history) == 0 {
		return history, nil
	}

	// Check cache
	historyStr := strings.Join(history, "\n")
	cacheKey := c.hashContent(historyStr)
	c.cacheMu.RLock()
	if cached, exists := c.compressionCache[cacheKey]; exists {
		c.cacheMu.RUnlock()
		// Parse cached result back to []string
		return strings.Split(cached, "\n---MESSAGE---\n"), nil
	}
	c.cacheMu.RUnlock()

	// Build compression prompt
	prompt := c.buildHistoryCompressionPrompt(history, targetRatio)

	// Call LLM
	response, err := c.callLLMForCompression(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM compression failed: %w", err)
	}

	// Parse response (expecting compressed history as text, one message per line or separated)
	compressed := c.parseCompressedHistory(response)

	// Cache the result
	c.cacheMu.Lock()
	c.compressionCache[cacheKey] = strings.Join(compressed, "\n---MESSAGE---\n")
	c.cacheMu.Unlock()

	return compressed, nil
}

// CompressSkillsWithLLM uses LLM to compress Skills content
func (c *Compressor) CompressSkillsWithLLM(ctx context.Context, skillsList []*skills.Skill) ([]*skills.Skill, error) {
	if c.llmClient == nil {
		return nil, fmt.Errorf("LLM client not set")
	}

	if len(skillsList) == 0 {
		return skillsList, nil
	}

	// Compress each Skill's content
	compressedSkills := make([]*skills.Skill, 0, len(skillsList))
	for _, skill := range skillsList {
		// Check cache
		cacheKey := c.hashContent(skill.Content)
		c.cacheMu.RLock()
		if cached, exists := c.compressionCache[cacheKey]; exists {
			c.cacheMu.RUnlock()
			compressedSkill := *skill
			compressedSkill.Content = cached
			compressedSkills = append(compressedSkills, &compressedSkill)
			continue
		}
		c.cacheMu.RUnlock()

		// Build compression prompt for Skill
		prompt := c.buildSkillCompressionPrompt(skill)

		// Call LLM
		compressedContent, err := c.callLLMForCompression(ctx, prompt)
		if err != nil {
			// If compression fails, keep original
			compressedSkills = append(compressedSkills, skill)
			continue
		}

		// Cache and use compressed content
		c.cacheMu.Lock()
		c.compressionCache[cacheKey] = compressedContent
		c.cacheMu.Unlock()

		compressedSkill := *skill
		compressedSkill.Content = compressedContent
		compressedSkills = append(compressedSkills, &compressedSkill)
	}

	return compressedSkills, nil
}

// buildHistoryCompressionPrompt builds the prompt for compressing conversation history
func (c *Compressor) buildHistoryCompressionPrompt(history []string, targetRatio float64) string {
	var builder strings.Builder
	builder.WriteString("Compress the following conversation history while preserving key information.\n\n")
	builder.WriteString(fmt.Sprintf("Target compression: reduce to approximately %.0f%% of original length.\n\n", targetRatio*100))
	builder.WriteString("PRESERVE:\n")
	builder.WriteString("- User preferences and important decisions\n")
	builder.WriteString("- Key query results and data\n")
	builder.WriteString("- Active context and ongoing tasks\n")
	builder.WriteString("- Important facts and conclusions\n\n")
	builder.WriteString("REMOVE:\n")
	builder.WriteString("- Redundant information\n")
	builder.WriteString("- Outdated context\n")
	builder.WriteString("- Irrelevant details\n\n")
	builder.WriteString("Conversation history:\n")
	for i, msg := range history {
		builder.WriteString(fmt.Sprintf("Message %d: %s\n", i+1, msg))
	}
	builder.WriteString("\nReturn the compressed conversation history, maintaining essential context.")
	return builder.String()
}

// buildSkillCompressionPrompt builds the prompt for compressing Skill content
func (c *Compressor) buildSkillCompressionPrompt(skill *skills.Skill) string {
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("Compress the following Skill content while preserving essential information.\n\n"))
	builder.WriteString(fmt.Sprintf("Skill: %s\n", skill.Name))
	builder.WriteString(fmt.Sprintf("Description: %s\n\n", skill.Description))
	builder.WriteString("PRESERVE:\n")
	builder.WriteString("- Relevant instructions and steps\n")
	builder.WriteString("- Key examples and code snippets\n")
	builder.WriteString("- Important context and prerequisites\n\n")
	builder.WriteString("REMOVE:\n")
	builder.WriteString("- Redundant explanations\n")
	builder.WriteString("- Outdated information\n")
	builder.WriteString("- Irrelevant details\n\n")
	builder.WriteString("Skill content:\n")
	builder.WriteString(skill.Content)
	builder.WriteString("\n\nReturn the compressed Skill content, maintaining essential information.")
	return builder.String()
}

// callLLMForCompression calls LLM API for compression (similar to matcher's callLLMAPI)
func (c *Compressor) callLLMForCompression(ctx context.Context, prompt string) (string, error) {
	// Build messages
	messages := []llm.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant that compresses text while preserving key information. Return only the compressed content, no explanations.",
		},
		{
			Role:    "user",
			Content: prompt,
		},
	}

	// Convert messages to []interface{}
	messagesInterface := make([]interface{}, len(messages))
	for i, msg := range messages {
		messagesInterface[i] = msg
	}

	// Create request body
	reqBody := struct {
		Model    string        `json:"model"`
		Messages []interface{} `json:"messages"`
	}{
		Model:    c.llmClient.Model(),
		Messages: messagesInterface,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build API URL
	apiURL := c.buildAPIURL()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.llmClient.APIKey())

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s (type: %s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	// Extract response
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	content := chatResp.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("empty content in response")
	}

	return content, nil
}

// parseCompressedHistory parses LLM response into compressed history
func (c *Compressor) parseCompressedHistory(response string) []string {
	// Try to split by common separators
	if strings.Contains(response, "\n---MESSAGE---\n") {
		return strings.Split(response, "\n---MESSAGE---\n")
	}
	if strings.Contains(response, "\n\n") {
		// Split by double newline (common for message separation)
		return strings.Split(response, "\n\n")
	}
	// Fallback: treat as single message
	return []string{response}
}

// buildAPIURL builds the API URL (similar to matcher's buildAPIURL)
func (c *Compressor) buildAPIURL() string {
	baseURL := strings.TrimSuffix(c.llmClient.BaseURL(), "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	return baseURL + "/v1/chat/completions"
}

// hashContent creates a hash of content for caching
func (c *Compressor) hashContent(content string) string {
	hash := sha256.Sum256([]byte(content))
	return hex.EncodeToString(hash[:])
}

// evictLowPrioritySkills removes Skills with priority lower than the given threshold
func (c *Compressor) evictLowPrioritySkills(skillList []*skills.Skill, minPriority skills.Priority) []*skills.Skill {
	var remaining []*skills.Skill
	for _, skill := range skillList {
		if skill.Priority >= minPriority {
			remaining = append(remaining, skill)
		}
	}
	return remaining
}
