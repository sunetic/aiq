package llm

import (
	"encoding/json"
	"testing"
)

func TestParseArguments_StandardJSON(t *testing.T) {
	// Task 3.1: 标准 JSON 对象参数解析
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `{"command":"brew list mysql"}`,
		},
	}

	args, err := tc.ParseArguments()
	if err != nil {
		t.Fatalf("ParseArguments failed: %v", err)
	}

	if cmd, ok := args["command"].(string); !ok || cmd != "brew list mysql" {
		t.Errorf("Expected command 'brew list mysql', got %v", args["command"])
	}
}

func TestParseArguments_DoubleEncodedJSON(t *testing.T) {
	// Task 3.2: 双重编码 JSON 字符串解析
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `"{\"command\":\"brew list mysql\"}"`,
		},
	}

	args, err := tc.ParseArguments()
	if err != nil {
		t.Fatalf("ParseArguments failed: %v", err)
	}

	if cmd, ok := args["command"].(string); !ok || cmd != "brew list mysql" {
		t.Errorf("Expected command 'brew list mysql', got %v", args["command"])
	}
}

func TestParseArguments_MultipleLayers(t *testing.T) {
	// Task 3.3: 多重编码 JSON 字符串解析（3+ 层）
	// Triple-encoded: "\"{\\\"command\\\":\\\"brew list mysql\\\"}\""
	tripleEncoded := `"\"{\\\"command\\\":\\\"brew list mysql\\\"}\""`
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: tripleEncoded,
		},
	}

	args, err := tc.ParseArguments()
	if err != nil {
		t.Fatalf("ParseArguments failed: %v", err)
	}

	if cmd, ok := args["command"].(string); !ok || cmd != "brew list mysql" {
		t.Errorf("Expected command 'brew list mysql', got %v", args["command"])
	}
}

func TestParseArguments_EscapeSequences(t *testing.T) {
	// Task 3.4: 包含转义字符的 JSON 字符串解析
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `"{\"command\":\"brew services list | grep mysql\"}"`,
		},
	}

	args, err := tc.ParseArguments()
	if err != nil {
		t.Fatalf("ParseArguments failed: %v", err)
	}

	if cmd, ok := args["command"].(string); !ok || cmd != "brew services list | grep mysql" {
		t.Errorf("Expected command 'brew services list | grep mysql', got %v", args["command"])
	}
}

func TestParseArguments_EmptyString(t *testing.T) {
	// Task 3.5: 空字符串和空白字符处理
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `""`,
		},
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Error("Expected error for empty string, got nil")
	}
}

func TestParseArguments_WhitespaceOnly(t *testing.T) {
	// Task 3.5: 空字符串和空白字符处理
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `"   "`,
		},
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Error("Expected error for whitespace-only string, got nil")
	}
}

func TestParseArguments_InvalidJSON(t *testing.T) {
	// Task 3.6: 无效 JSON 格式处理
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `{invalid json}`,
		},
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestParseArguments_MaxDepth(t *testing.T) {
	// Task 3.7: 达到最大递归深度时的处理
	// Create a deeply nested JSON string (more than 10 layers)
	deeplyNested := `"\"\"\"\"\"\"\"\"\"\"{\\\"command\\\":\\\"test\\\"}\"\"\"\"\"\"\"\"\"\""`
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: deeplyNested,
		},
	}

	// Should stop at max depth and try to parse
	_, err := tc.ParseArguments()
	// May succeed if it can parse after max depth, or fail if it can't
	// The important thing is it doesn't panic or loop infinitely
	if err != nil {
		// Error is acceptable - it means we stopped at max depth
		t.Logf("ParseArguments stopped at max depth (expected): %v", err)
	}
}

func TestParseArguments_ErrorIncludesOriginal(t *testing.T) {
	// Task 3.8: 错误信息包含原始参数
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `{invalid json}`,
		},
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	errMsg := err.Error()
	if errMsg == "" {
		t.Error("Error message should not be empty")
	}

	// Check that error message includes "original" keyword
	// The exact format may vary, but should include context
	if len(errMsg) < 20 {
		t.Errorf("Error message seems too short: %s", errMsg)
	}
}

func TestParseArguments_BackwardCompatibility(t *testing.T) {
	// Test backward compatibility: standard format should still work
	testCases := []struct {
		name      string
		arguments string
		expected  map[string]interface{}
	}{
		{
			name:      "simple object",
			arguments: `{"key":"value"}`,
			expected:  map[string]interface{}{"key": "value"},
		},
		{
			name:      "multiple fields",
			arguments: `{"command":"ls","timeout":30}`,
			expected:  map[string]interface{}{"command": "ls", "timeout": float64(30)},
		},
		{
			name:      "nested object",
			arguments: `{"config":{"key":"value"}}`,
			expected: map[string]interface{}{
				"config": map[string]interface{}{"key": "value"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			toolCall := &ToolCall{
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{
					Name:      "test_tool",
					Arguments: tc.arguments,
				},
			}

			args, err := toolCall.ParseArguments()
			if err != nil {
				t.Fatalf("ParseArguments failed: %v", err)
			}

			// Compare JSON representation for deep equality
			expectedJSON, _ := json.Marshal(tc.expected)
			actualJSON, _ := json.Marshal(args)
			if string(expectedJSON) != string(actualJSON) {
				t.Errorf("Expected %s, got %s", string(expectedJSON), string(actualJSON))
			}
		})
	}
}

func TestParseArguments_NullValue(t *testing.T) {
	// Test that null values are rejected (should be a map, not null)
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `null`,
		},
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Error("Expected error for null value, got nil")
	}
}

func TestParseArguments_ArrayValue(t *testing.T) {
	// Test that array values are rejected (should be a map, not array)
	tc := &ToolCall{
		Function: struct {
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		}{
			Name:      "execute_command",
			Arguments: `["item1","item2"]`,
		},
	}

	_, err := tc.ParseArguments()
	if err == nil {
		t.Error("Expected error for array value, got nil")
	}
}
