package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client represents an LLM API client
type Client struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

// NewClient creates a new LLM client
func NewClient(baseURL, apiKey, model string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a chat API request
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
}

// ChatResponse represents a chat API response
type ChatResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type     string `json:"type"`
	} `json:"error,omitempty"`
}

// TranslateToSQL translates natural language to SQL using LLM
func (c *Client) TranslateToSQL(ctx context.Context, naturalLanguage string, schemaContext string, databaseType string) (string, error) {
	// Build prompt
	prompt := buildSQLPrompt(naturalLanguage, schemaContext, databaseType)

	// Create request
	reqBody := ChatRequest{
		Model: c.model,
		Messages: []ChatMessage{
			{
				Role:    "system",
				Content: "You are a SQL expert. Translate natural language questions into precise SQL queries. Only return the SQL query, no explanations.",
			},
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the full API URL
	apiURL := c.buildAPIURL()
	
	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	// Execute request with retry
	var resp *http.Response
	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		resp, err = c.client.Do(req)
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
	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s (type: %s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	// Extract SQL
	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}

	sql := chatResp.Choices[0].Message.Content
	if sql == "" {
		return "", fmt.Errorf("empty SQL in response")
	}

	// Clean SQL: remove markdown code block markers if present
	sql = cleanSQL(sql)

	return sql, nil
}

// buildAPIURL builds the full API URL from the base URL
// Handles different URL formats:
// - https://api.openai.com/v1 -> https://api.openai.com/v1/chat/completions
// - https://api.openai.com/v1/chat/completions -> https://api.openai.com/v1/chat/completions (no change)
// - https://api.example.com -> https://api.example.com/v1/chat/completions
func (c *Client) buildAPIURL() string {
	baseURL := strings.TrimSuffix(c.baseURL, "/")
	
	// If URL already ends with /chat/completions, use it as-is
	if strings.HasSuffix(baseURL, "/chat/completions") {
		return baseURL
	}
	
	// If URL ends with /v1, append /chat/completions
	if strings.HasSuffix(baseURL, "/v1") {
		return baseURL + "/chat/completions"
	}
	
	// Otherwise, append /v1/chat/completions
	return baseURL + "/v1/chat/completions"
}

// cleanSQL removes markdown code block markers and extra whitespace from SQL
func cleanSQL(sql string) string {
	sql = strings.TrimSpace(sql)
	
	// Remove markdown code block markers (```sql, ```, etc.)
	sql = strings.TrimPrefix(sql, "```sql")
	sql = strings.TrimPrefix(sql, "```SQL")
	sql = strings.TrimPrefix(sql, "```")
	sql = strings.TrimSuffix(sql, "```")
	
	// Remove any leading/trailing whitespace again
	sql = strings.TrimSpace(sql)
	
	return sql
}

func buildSQLPrompt(naturalLanguage string, schemaContext string, databaseType string) string {
	prompt := fmt.Sprintf(`You are a SQL expert for %s database.

Given the following database schema:

%s

Translate this natural language question into a %s SQL query:

%s

Return only the SQL query, no explanations. Use %s syntax.`, databaseType, schemaContext, databaseType, naturalLanguage, databaseType)
	return prompt
}
