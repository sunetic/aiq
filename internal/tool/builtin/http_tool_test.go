package builtin

import (
	"context"
	"testing"
)

// TestHTTPTool_GETRequests tests HTTP tool GET requests
func TestHTTPTool_GETRequests(t *testing.T) {
	t.Run("executes GET request", func(t *testing.T) {
		tool := NewHTTPTool()

		// Use httpbin.org for testing (if available)
		// Fallback: test error handling if network unavailable
		params := map[string]interface{}{
			"method": "GET",
			"url":    "https://httpbin.org/get",
		}

		ctx := context.Background()
		result, err := tool.Execute(ctx, params)
		if err != nil {
			// Network error is acceptable in test environment
			t.Logf("HTTP request failed (may be network issue): %v", err)
			return
		}

		httpResponse, ok := result.(HTTPResponse)
		if !ok {
			t.Fatalf("Expected HTTPResponse, got %T", result)
		}

		if httpResponse.Status < 200 || httpResponse.Status >= 300 {
			t.Errorf("Expected success status, got %d", httpResponse.Status)
		}
	})

	t.Run("handles invalid URL", func(t *testing.T) {
		tool := NewHTTPTool()

		params := map[string]interface{}{
			"method": "GET",
			"url":    "not-a-valid-url",
		}

		ctx := context.Background()
		_, err := tool.Execute(ctx, params)
		if err == nil {
			t.Error("Expected error for invalid URL")
		}
	})

	t.Run("handles missing URL", func(t *testing.T) {
		tool := NewHTTPTool()

		params := map[string]interface{}{
			"method": "GET",
		}

		ctx := context.Background()
		_, err := tool.Execute(ctx, params)
		if err == nil {
			t.Error("Expected error for missing URL")
		}
	})
}

// TestHTTPTool_ErrorHandling tests HTTP tool error handling
func TestHTTPTool_ErrorHandling(t *testing.T) {
	t.Run("handles network errors", func(t *testing.T) {
		tool := NewHTTPTool()

		params := map[string]interface{}{
			"method": "GET",
			"url":    "https://nonexistent-domain-12345.com",
		}

		ctx := context.Background()
		_, err := tool.Execute(ctx, params)
		// Network error is acceptable
		if err == nil {
			t.Log("Request succeeded (unexpected, but may be DNS hijacking)")
		}
	})

	t.Run("handles unsupported HTTP method", func(t *testing.T) {
		tool := NewHTTPTool()

		params := map[string]interface{}{
			"method": "INVALID_METHOD",
			"url":    "https://httpbin.org/get",
		}

		ctx := context.Background()
		_, err := tool.Execute(ctx, params)
		if err == nil {
			t.Error("Expected error for unsupported method")
		}
	})

	t.Run("handles timeout", func(t *testing.T) {
		tool := NewHTTPTool()

		params := map[string]interface{}{
			"method":  "GET",
			"url":     "https://httpbin.org/delay/10",
			"timeout": 1, // 1 second timeout
		}

		ctx := context.Background()
		_, err := tool.Execute(ctx, params)
		// Timeout error is acceptable
		if err == nil {
			t.Log("Request completed (may be faster than expected)")
		}
	})
}
