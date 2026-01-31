package builtin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

// HTTPTool handles HTTP requests
type HTTPTool struct{}

// NewHTTPTool creates a new HTTP tool
func NewHTTPTool() *HTTPTool {
	return &HTTPTool{}
}

// HTTPRequestParams represents parameters for HTTP request
type HTTPRequestParams struct {
	Method  string            `json:"method"` // GET, POST, PUT, DELETE
	URL     string            `json:"url"`
	Headers map[string]string `json:"headers,omitempty"`
	Body    string            `json:"body,omitempty"`
	Timeout int               `json:"timeout,omitempty"` // Timeout in seconds, default 30
}

// HTTPResponse represents HTTP response
type HTTPResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
}

// Execute executes an HTTP request
func (t *HTTPTool) Execute(ctx context.Context, params map[string]interface{}) (interface{}, error) {
	// Parse parameters
	var httpParams HTTPRequestParams
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal params: %w", err)
	}
	if err := json.Unmarshal(paramsJSON, &httpParams); err != nil {
		return nil, fmt.Errorf("failed to parse params: %w", err)
	}

	// Validate method
	method := httpParams.Method
	if method == "" {
		method = "GET"
	}
	switch method {
	case "GET", "POST", "PUT", "DELETE":
		// Valid method
	default:
		return nil, fmt.Errorf("unsupported HTTP method: %s", method)
	}

	// Validate URL
	if httpParams.URL == "" {
		return nil, fmt.Errorf("URL is required")
	}
	_, err = url.Parse(httpParams.URL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	// Set timeout
	timeout := 30 * time.Second
	if httpParams.Timeout > 0 {
		timeout = time.Duration(httpParams.Timeout) * time.Second
	}

	// Create request
	var bodyReader io.Reader
	if httpParams.Body != "" && (method == "POST" || method == "PUT") {
		bodyReader = bytes.NewReader([]byte(httpParams.Body))
	}

	req, err := http.NewRequestWithContext(ctx, method, httpParams.URL, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	for key, value := range httpParams.Headers {
		req.Header.Set(key, value)
	}

	// Set default Content-Type if body is provided and not set
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request
	client := &http.Client{
		Timeout: timeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Build response
	response := HTTPResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    make(map[string]string),
		Body:       string(bodyBytes),
	}

	// Copy headers
	for key, values := range resp.Header {
		if len(values) > 0 {
			response.Headers[key] = values[0]
		}
	}

	return response, nil
}

// GetDefinition returns the tool definition for LLM
func (t *HTTPTool) GetDefinition() map[string]interface{} {
	return map[string]interface{}{
		"type": "function",
		"function": map[string]interface{}{
			"name":        "http_request",
			"description": "Make HTTP requests (GET, POST, PUT, DELETE). Use this to fetch data from APIs or send data to endpoints.",
			"parameters": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"method": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"GET", "POST", "PUT", "DELETE"},
						"description": "HTTP method",
						"default":     "GET",
					},
					"url": map[string]interface{}{
						"type":        "string",
						"description": "URL to request",
					},
					"headers": map[string]interface{}{
						"type":        "object",
						"description": "HTTP headers as key-value pairs",
					},
					"body": map[string]interface{}{
						"type":        "string",
						"description": "Request body (for POST/PUT)",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds (default: 30)",
					},
					"risk_level": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"low", "medium", "high"},
						"description": "Optional: Risk level assessment for this operation. 'low' = safe to execute automatically (e.g., GET, HEAD), 'medium'/'high' = requires user confirmation (e.g., POST, DELETE). If not provided, system will assess risk conservatively.",
					},
					"task_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"definitive", "exploratory"},
						"description": "Optional: Task type classification. 'definitive' = task is clear and complete, 'exploratory' = task requires information gathering or multi-step process. If not provided, system will infer from context.",
					},
					"output_mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"full", "streaming"},
						"description": "Optional: Output display mode. 'full' = display all results to user (for definitive tasks), 'streaming' = real-time streaming with truncation (for exploratory tasks). If not provided, inferred from task_type.",
					},
				},
				"required": []string{"url"},
			},
		},
	}
}
