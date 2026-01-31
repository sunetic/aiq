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

// BaseURL returns the base URL of the LLM client
func (c *Client) BaseURL() string {
	return c.baseURL
}

// APIKey returns the API key of the LLM client
func (c *Client) APIKey() string {
	return c.apiKey
}

// Model returns the model name of the LLM client
func (c *Client) Model() string {
	return c.model
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Function represents a tool/function definition for LLM
type Function struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToolCall represents a tool call request from LLM
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // "function"
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON string that needs to be parsed
	} `json:"function"`
}

// ParseArguments parses the arguments JSON string into a map
// Handles various JSON encoding formats including double-encoded strings, nested quotes, and escape sequences
func (tc *ToolCall) ParseArguments() (map[string]interface{}, error) {
	const maxUnquoteDepth = 10
	argsStr := tc.Function.Arguments
	originalArgs := argsStr

	// Recursively unquote JSON strings until we can't unquote anymore or reach max depth
	for depth := 0; depth < maxUnquoteDepth; depth++ {
		trimmed := strings.TrimSpace(argsStr)
		// Check if it's a quoted JSON string
		if len(trimmed) < 2 || trimmed[0] != '"' || trimmed[len(trimmed)-1] != '"' {
			break
		}
		// Try to unquote the string
		var unquoted string
		if err := json.Unmarshal([]byte(trimmed), &unquoted); err != nil {
			// Can't unquote further, stop
			break
		}
		argsStr = unquoted
	}

	// Try to parse as JSON object
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
		// Truncate original args for error message (max 100 chars)
		truncatedOriginal := originalArgs
		if len(truncatedOriginal) > 100 {
			truncatedOriginal = truncatedOriginal[:100] + "..."
		}
		return nil, fmt.Errorf("failed to parse arguments after unquoting: %w (original: %s)", err, truncatedOriginal)
	}

	// Validate that we got a valid JSON object (not null, array, or primitive)
	if args == nil {
		truncatedOriginal := originalArgs
		if len(truncatedOriginal) > 100 {
			truncatedOriginal = truncatedOriginal[:100] + "..."
		}
		return nil, fmt.Errorf("failed to parse arguments: expected JSON object, got null (original: %s)", truncatedOriginal)
	}

	return args, nil
}

// ChatRequest represents a chat API request
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []interface{} `json:"messages"` // Can be ChatMessage or map[string]interface{} for tool messages
	Tools    []struct {
		Type     string   `json:"type"`
		Function Function `json:"function"`
	} `json:"tools,omitempty"`
	ToolChoice interface{} `json:"tool_choice,omitempty"` // "auto", "none", or {"type": "function", "function": {"name": "..."}}
}

// ChatResponse represents a chat API response
type ChatResponse struct {
	Choices []struct {
		Message struct {
			Role      string     `json:"role"`
			Content   string     `json:"content"`
			ToolCalls []ToolCall `json:"tool_calls,omitempty"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"` // "stop", "tool_calls", etc.
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// ChatResponseType indicates the type of response from LLM
type ChatResponseType int

const (
	ChatResponseTypeText ChatResponseType = iota // Regular text response
	ChatResponseTypeSQL                          // SQL query response
)

// ChatResponse represents a chat response with type
type ChatResponseWithType struct {
	Type    ChatResponseType
	Content string
}

// Chat handles general conversation and can return either text or SQL
// It intelligently determines if the user wants SQL or just wants to chat
func (c *Client) Chat(ctx context.Context, userInput string, schemaContext string, databaseType string, conversationHistory []ChatMessage) (*ChatResponseWithType, error) {
	// Build system message that allows both conversation and SQL generation
	systemMessage := fmt.Sprintf(`You are a helpful AI assistant for database queries. You can have natural conversations with users, or help them generate SQL queries.

When the user asks a question that requires a database query, respond with ONLY the SQL query (no explanations, no markdown code blocks, just the SQL).
When the user is just chatting (greetings, questions about how to use the tool, etc.), respond naturally as a helpful assistant.

IMPORTANT CONTEXT:
- Database engine type: %s (this is the database system type like MySQL/PostgreSQL/seekdb, NOT a schema or database name)
- Current database connection information is provided below

Database connection and schema information:
%s

CRITICAL RULES FOR SQL GENERATION:
1. Database type "%s" is the ENGINE TYPE (like MySQL, PostgreSQL, seekdb), NOT a database name or schema name
2. When user asks "show tables" or "list tables", generate: SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE() OR use SHOW TABLES (for MySQL/seekdb)
3. NEVER use the database engine type (like "%s") as a schema name in WHERE table_schema = '%s'
4. Always use the actual database name from the connection context, or use DATABASE() function to get current database
5. For MySQL/seekdb: Use SHOW TABLES; or SELECT table_name FROM information_schema.tables WHERE table_schema = DATABASE();
6. For PostgreSQL: Use SELECT tablename FROM pg_tables WHERE schemaname = 'public'; or SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';

Remember: If the user wants SQL, return ONLY the SQL query. If it's just conversation, respond naturally.`, databaseType, schemaContext, databaseType, databaseType, databaseType)

	// Build messages list
	messages := make([]ChatMessage, 0)

	// System message is always first
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: systemMessage,
	})

	// Add conversation history if provided
	if conversationHistory != nil && len(conversationHistory) > 0 {
		messages = append(messages, conversationHistory...)
	}

	// Current query is always last
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: userInput,
	})

	// Convert messages to []interface{}
	messagesInterface := make([]interface{}, len(messages))
	for i, msg := range messages {
		messagesInterface[i] = msg
	}

	// Create request
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messagesInterface,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the full API URL
	apiURL := c.buildAPIURL()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("request failed after retries: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp struct {
		Choices []struct {
			Message struct {
				Role      string     `json:"role"`
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s (type: %s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	// Extract response
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := chatResp.Choices[0].Message.Content
	if content == "" {
		return nil, fmt.Errorf("empty content in response")
	}

	// Determine if response is SQL or text
	// Check if content looks like SQL (contains SELECT, INSERT, UPDATE, DELETE, SHOW, etc.)
	contentUpper := strings.ToUpper(strings.TrimSpace(content))
	isSQL := strings.HasPrefix(contentUpper, "SELECT") ||
		strings.HasPrefix(contentUpper, "INSERT") ||
		strings.HasPrefix(contentUpper, "UPDATE") ||
		strings.HasPrefix(contentUpper, "DELETE") ||
		strings.HasPrefix(contentUpper, "CREATE") ||
		strings.HasPrefix(contentUpper, "ALTER") ||
		strings.HasPrefix(contentUpper, "DROP") ||
		strings.HasPrefix(contentUpper, "SHOW") ||
		strings.HasPrefix(contentUpper, "DESCRIBE") ||
		strings.HasPrefix(contentUpper, "DESC") ||
		strings.HasPrefix(contentUpper, "EXPLAIN") ||
		strings.Contains(contentUpper, " FROM ") ||
		strings.Contains(contentUpper, " WHERE ")

	if isSQL {
		// Clean SQL: remove markdown code block markers if present
		sql := cleanSQL(content)
		return &ChatResponseWithType{
			Type:    ChatResponseTypeSQL,
			Content: sql,
		}, nil
	}

	// Regular text response
	return &ChatResponseWithType{
		Type:    ChatResponseTypeText,
		Content: content,
	}, nil
}

// ChatWithTools handles conversation with tool support
// messages can include ChatMessage or map[string]interface{} for tool messages
func (c *Client) ChatWithTools(ctx context.Context, messages []interface{}, tools []Function) (*ChatResponse, error) {

	// Build tools array for request
	toolsArray := make([]struct {
		Type     string   `json:"type"`
		Function Function `json:"function"`
	}, len(tools))
	for i, tool := range tools {
		toolsArray[i] = struct {
			Type     string   `json:"type"`
			Function Function `json:"function"`
		}{
			Type:     "function",
			Function: tool,
		}
	}

	// Normalize messages to ensure content field is always a string (not an object)
	// This is critical for LLM API compatibility
	// Messages from Session are already normalized in mode.go and tool_handler.go,
	// but we do a final check here for safety
	normalizedMessages := make([]interface{}, 0, len(messages))
	for _, msg := range messages {
		var msgMap map[string]interface{}

		// Convert to map if needed
		if m, ok := msg.(map[string]interface{}); ok {
			msgMap = m
		} else if chatMsg, ok := msg.(ChatMessage); ok {
			// Convert ChatMessage struct to map
			msgMap = map[string]interface{}{
				"role":    chatMsg.Role,
				"content": chatMsg.Content,
			}
		} else {
			// Unknown type - try to convert via JSON
			if jsonBytes, err := json.Marshal(msg); err == nil {
				if json.Unmarshal(jsonBytes, &msgMap) != nil {
					continue // Skip if conversion fails
				}
			} else {
				continue // Skip if marshal fails
			}
		}

		// Ensure content field is always a string
		if content, exists := msgMap["content"]; exists && content != nil {
			if _, isString := content.(string); !isString {
				// Convert to string
				if jsonBytes, err := json.Marshal(content); err == nil {
					msgMap["content"] = string(jsonBytes)
				} else {
					msgMap["content"] = fmt.Sprintf("%v", content)
				}
			}
		}

		normalizedMessages = append(normalizedMessages, msgMap)
	}

	// Create request
	reqBody := ChatRequest{
		Model:      c.model,
		Messages:   normalizedMessages,
		Tools:      toolsArray,
		ToolChoice: "auto", // Let LLM decide when to use tools
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build the full API URL
	apiURL := c.buildAPIURL()

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
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
		return nil, fmt.Errorf("request failed after retries: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var chatResp struct {
		Choices []struct {
			Message struct {
				Role      string     `json:"role"`
				Content   string     `json:"content"`
				ToolCalls []ToolCall `json:"tool_calls,omitempty"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		} `json:"error,omitempty"`
	}
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for API errors
	if chatResp.Error != nil {
		return nil, fmt.Errorf("API error: %s (type: %s)", chatResp.Error.Message, chatResp.Error.Type)
	}

	// Extract response
	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	// Return the response as-is, including tool_calls
	// Let the caller handle tool_calls
	return &ChatResponse{
		Choices: chatResp.Choices,
	}, nil
}

// TranslateToSQL translates natural language to SQL using LLM
// conversationHistory can be nil or empty for backward compatibility
// Deprecated: Use Chat() instead for more flexible conversation handling
func (c *Client) TranslateToSQL(ctx context.Context, naturalLanguage string, schemaContext string, databaseType string, conversationHistory []ChatMessage) (string, error) {
	// Build prompt
	prompt := buildSQLPrompt(naturalLanguage, schemaContext, databaseType)

	// Build messages list
	messages := make([]ChatMessage, 0)

	// System message is always first
	messages = append(messages, ChatMessage{
		Role:    "system",
		Content: "You are a SQL expert. Translate natural language questions into precise SQL queries. Only return the SQL query, no explanations.",
	})

	// Add conversation history if provided
	if conversationHistory != nil && len(conversationHistory) > 0 {
		messages = append(messages, conversationHistory...)
	}

	// Current query is always last
	messages = append(messages, ChatMessage{
		Role:    "user",
		Content: prompt,
	})

	// Convert messages to []interface{}
	messagesInterface := make([]interface{}, len(messages))
	for i, msg := range messages {
		messagesInterface[i] = msg
	}

	// Create request
	reqBody := ChatRequest{
		Model:    c.model,
		Messages: messagesInterface,
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
