package skills

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
)

const (
	// DefaultMaxSkills is the default maximum number of Skills to load per query
	DefaultMaxSkills = 3
)

// MatchResult represents a match between a query and a Skill
type MatchResult struct {
	Metadata *Metadata
	Score    float64
}

// Matcher matches user queries to Skills based on relevance
type Matcher struct {
	maxSkills int
	llmClient *llm.Client
	cache     map[string][]string // cache key: query hash, value: Skill names
	cacheMu   sync.RWMutex
}

// NewMatcher creates a new Skills matcher
func NewMatcher() *Matcher {
	return &Matcher{
		maxSkills: DefaultMaxSkills,
		cache:     make(map[string][]string),
	}
}

// SetLLMClient sets the LLM client for semantic matching
func (m *Matcher) SetLLMClient(client *llm.Client) {
	m.llmClient = client
}

// SetMaxSkills sets the maximum number of Skills to return
func (m *Matcher) SetMaxSkills(max int) {
	m.maxSkills = max
}

// Match matches a query against Skills metadata and returns the most relevant Skills
// Tries LLM semantic matching first, falls back to keyword matching if LLM is unavailable or fails
func (m *Matcher) Match(query string, metadataList []*Metadata) []*Metadata {
	if len(metadataList) == 0 {
		return []*Metadata{}
	}

	// Try LLM semantic matching first if client is available
	if m.llmClient != nil {
		matched, err := m.MatchWithLLM(context.Background(), query, metadataList)
		if err == nil && len(matched) > 0 {
			return matched
		}
		// If LLM matching fails, fall through to keyword matching
	}

	// Fallback to keyword matching
	return m.matchWithKeywords(query, metadataList)
}

// matchWithKeywords matches using keyword-based scoring (original implementation)
func (m *Matcher) matchWithKeywords(query string, metadataList []*Metadata) []*Metadata {
	// Extract keywords from query
	keywords := extractKeywords(query)

	// Score each Skill
	results := make([]MatchResult, 0, len(metadataList))
	for _, md := range metadataList {
		score := m.scoreSkill(keywords, md)
		if score > 0 {
			results = append(results, MatchResult{
				Metadata: md,
				Score:    score,
			})
		}
	}

	// Sort by score (descending)
	sortByScore(results)

	// Return top N
	max := m.maxSkills
	if len(results) < max {
		max = len(results)
	}

	matched := make([]*Metadata, 0, max)
	for i := 0; i < max; i++ {
		matched = append(matched, results[i].Metadata)
	}

	return matched
}

// MatchWithLLM uses LLM to perform semantic matching of Skills
func (m *Matcher) MatchWithLLM(ctx context.Context, query string, metadataList []*Metadata) ([]*Metadata, error) {
	if m.llmClient == nil {
		return nil, fmt.Errorf("LLM client not set")
	}

	// Check cache first
	cacheKey := m.hashQuery(query)
	m.cacheMu.RLock()
	if cached, exists := m.cache[cacheKey]; exists {
		m.cacheMu.RUnlock()
		return m.metadataFromNames(cached, metadataList), nil
	}
	m.cacheMu.RUnlock()

	// Build prompt for LLM semantic matching
	prompt := m.buildMatchingPrompt(query, metadataList)

	// Call LLM API directly
	response, err := m.callLLMAPI(ctx, prompt)
	if err != nil {
		return nil, fmt.Errorf("LLM call failed: %w", err)
	}

	// Parse LLM response to extract Skill names
	skillNames, err := m.parseLLMResponse(response)
	if err != nil {
		return nil, fmt.Errorf("failed to parse LLM response: %w", err)
	}

	// Cache the result
	m.cacheMu.Lock()
	m.cache[cacheKey] = skillNames
	m.cacheMu.Unlock()

	// Convert Skill names to Metadata
	return m.metadataFromNames(skillNames, metadataList), nil
}

// callLLMAPI makes a direct API call to LLM for semantic matching
func (m *Matcher) callLLMAPI(ctx context.Context, prompt string) (string, error) {
	// Build messages
	messages := []llm.ChatMessage{
		{
			Role:    "system",
			Content: "You are a helpful assistant that matches user queries to relevant Skills based on semantic understanding. Return only a JSON array of Skill names that are relevant to the query. Example: [\"skill-name-1\", \"skill-name-2\"]",
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
		Model:    m.llmClient.Model(),
		Messages: messagesInterface,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build API URL (reuse LLM client's buildAPIURL logic)
	apiURL := m.buildAPIURL()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.llmClient.APIKey())

	// Execute request with retry
	client := &http.Client{Timeout: 30 * time.Second}
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = client.Do(req)
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Duration(i+1) * time.Second)
		}
	}
	if err != nil {
		return "", fmt.Errorf("request failed after retries: %w", err)
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

// buildAPIURL builds the API URL (similar to LLM client's buildAPIURL)
func (m *Matcher) buildAPIURL() string {
	baseURL := strings.TrimSuffix(m.llmClient.BaseURL(), "/")
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	return baseURL + "/v1/chat/completions"
}

// buildMatchingPrompt builds the prompt for LLM semantic matching
func (m *Matcher) buildMatchingPrompt(query string, metadataList []*Metadata) string {
	var builder strings.Builder
	builder.WriteString("User query: ")
	builder.WriteString(query)
	builder.WriteString("\n\n")
	builder.WriteString("Available Skills:\n")

	for i, md := range metadataList {
		builder.WriteString(fmt.Sprintf("%d. Name: %s\n   Description: %s\n", i+1, md.Name, md.Description))
	}

	builder.WriteString("\n")
	builder.WriteString("Based on semantic understanding, return a JSON array of Skill names that are relevant to the user query.\n")
	builder.WriteString("Example: [\"skill-name-1\", \"skill-name-2\"]\n")
	builder.WriteString("Return only the JSON array, no other text.")

	return builder.String()
}

// parseLLMResponse parses LLM response to extract Skill names
// Handles both JSON array format and plain text with Skill names
func (m *Matcher) parseLLMResponse(response string) ([]string, error) {
	// Try to parse as JSON array first
	var skillNames []string
	if err := json.Unmarshal([]byte(response), &skillNames); err == nil {
		// Limit to maxSkills
		if len(skillNames) > m.maxSkills {
			skillNames = skillNames[:m.maxSkills]
		}
		return skillNames, nil
	}

	// If JSON parsing fails, try to extract Skill names from text
	// Look for patterns like "skill-name" or ["skill-name"]
	response = strings.TrimSpace(response)
	
	// Try to find JSON array in the response
	startIdx := strings.Index(response, "[")
	endIdx := strings.LastIndex(response, "]")
	if startIdx != -1 && endIdx != -1 && endIdx > startIdx {
		jsonStr := response[startIdx : endIdx+1]
		if err := json.Unmarshal([]byte(jsonStr), &skillNames); err == nil {
			if len(skillNames) > m.maxSkills {
				skillNames = skillNames[:m.maxSkills]
			}
			return skillNames, nil
		}
	}

	return nil, fmt.Errorf("could not parse Skill names from LLM response: %s", response)
}

// metadataFromNames converts Skill names to Metadata objects
func (m *Matcher) metadataFromNames(names []string, metadataList []*Metadata) []*Metadata {
	nameMap := make(map[string]*Metadata)
	for _, md := range metadataList {
		nameMap[md.Name] = md
	}

	result := make([]*Metadata, 0, len(names))
	for _, name := range names {
		if md, exists := nameMap[name]; exists {
			result = append(result, md)
		}
	}

	return result
}

// hashQuery creates a hash of the query for caching
func (m *Matcher) hashQuery(query string) string {
	hash := sha256.Sum256([]byte(query))
	return hex.EncodeToString(hash[:])
}

// scoreSkill scores a Skill based on how well it matches the query keywords
// Scoring priority: exact name match > partial name match > description keyword match
func (m *Matcher) scoreSkill(keywords []string, metadata *Metadata) float64 {
	var score float64

	queryLower := strings.ToLower(strings.Join(keywords, " "))
	nameLower := strings.ToLower(metadata.Name)
	descLower := strings.ToLower(metadata.Description)

	// Exact name match (highest priority)
	if queryLower == nameLower {
		score += 100.0
	}

	// Partial name match
	if strings.Contains(nameLower, queryLower) || strings.Contains(queryLower, nameLower) {
		score += 50.0
	}

	// Check if any keyword matches name
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(nameLower, keywordLower) {
			score += 30.0
		}
	}

	// Description keyword match
	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		if strings.Contains(descLower, keywordLower) {
			score += 10.0
		}
	}

	return score
}

// extractKeywords extracts keywords from a query
// Simple implementation: split by spaces and filter common words
func extractKeywords(query string) []string {
	// Common words to filter out
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "and": true, "or": true,
		"but": true, "in": true, "on": true, "at": true, "to": true,
		"for": true, "of": true, "with": true, "by": true, "from": true,
		"is": true, "are": true, "was": true, "were": true, "be": true,
		"been": true, "have": true, "has": true, "had": true, "do": true,
		"does": true, "did": true, "will": true, "would": true, "could": true,
		"should": true, "may": true, "might": true, "must": true, "can": true,
		"this": true, "that": true, "these": true, "those": true,
		"i": true, "you": true, "he": true, "she": true, "it": true,
		"we": true, "they": true, "what": true, "which": true, "who": true,
		"when": true, "where": true, "why": true, "how": true,
	}

	words := strings.Fields(strings.ToLower(query))
	keywords := make([]string, 0, len(words))

	for _, word := range words {
		// Remove punctuation
		word = strings.Trim(word, ".,!?;:()[]{}'\"")
		if word == "" {
			continue
		}

		// Filter stop words
		if !stopWords[word] && len(word) > 1 {
			keywords = append(keywords, word)
		}
	}

	return keywords
}

// sortByScore sorts match results by score in descending order
func sortByScore(results []MatchResult) {
	// Simple bubble sort (fine for small lists)
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Score < results[j].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}
