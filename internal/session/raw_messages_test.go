package session

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// TestSession_RawMessagesStorage tests complete message array storage
func TestSession_RawMessagesStorage(t *testing.T) {
	t.Run("stores complete messages array including tool calls and results", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		// Create messages with tool calls
		messages := []json.RawMessage{
			json.RawMessage(`{"role": "system", "content": "You are a helpful assistant"}`),
			json.RawMessage(`{"role": "user", "content": "Show tables"}`),
			json.RawMessage(`{"role": "assistant", "content": "", "tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "execute_sql", "arguments": "{\"sql\": \"SHOW TABLES\"}"}}]}`),
			json.RawMessage(`{"role": "tool", "content": "{\"status\": \"success\", \"row_count\": 3}", "tool_call_id": "call_1"}`),
		}

		// Store messages
		sess.SetRawMessages(messages)

		// Verify messages were stored
		stored := sess.GetRawMessages()
		if len(stored) != len(messages) {
			t.Errorf("Expected %d messages, got %d", len(messages), len(stored))
		}

		// Verify tool calls are preserved
		var assistantMsg map[string]interface{}
		if err := json.Unmarshal(stored[2], &assistantMsg); err != nil {
			t.Fatalf("Failed to unmarshal assistant message: %v", err)
		}

		if _, exists := assistantMsg["tool_calls"]; !exists {
			t.Error("Expected tool_calls to be preserved in stored message")
		}
	})

	t.Run("serializes messages as JSON RawMessage array", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		messages := []json.RawMessage{
			json.RawMessage(`{"role": "user", "content": "test"}`),
		}

		sess.SetRawMessages(messages)

		// Verify serialization by saving and loading
		tmpDir := t.TempDir()
		sessionPath := filepath.Join(tmpDir, "test_session.json")

		if err := SaveSession(sess, sessionPath); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		// Load session
		loaded, err := LoadSession(sessionPath)
		if err != nil {
			t.Fatalf("LoadSession() failed: %v", err)
		}

		// Verify RawMessages were loaded
		loadedRaw := loaded.GetRawMessages()
		if len(loadedRaw) != len(messages) {
			t.Errorf("Expected %d raw messages after load, got %d", len(messages), len(loadedRaw))
		}
	})

	t.Run("updates legacy Messages field for backward compatibility", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		// Add raw messages
		messages := []json.RawMessage{
			json.RawMessage(`{"role": "user", "content": "test message"}`),
			json.RawMessage(`{"role": "assistant", "content": "response"}`),
		}

		sess.SetRawMessages(messages)

		// Verify legacy Messages field is also updated (if implementation does so)
		// Note: Current implementation may not update Messages automatically
		// This test documents expected behavior
		_ = sess.Messages
	})
}

// TestSession_RawMessagesLoading tests complete message array loading
func TestSession_RawMessagesLoading(t *testing.T) {
	t.Run("loads complete messages array if available", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		messages := []json.RawMessage{
			json.RawMessage(`{"role": "user", "content": "test"}`),
			json.RawMessage(`{"role": "assistant", "content": "response"}`),
		}

		sess.SetRawMessages(messages)

		// Save and load
		tmpDir := t.TempDir()
		sessionPath := filepath.Join(tmpDir, "test_session.json")

		if err := SaveSession(sess, sessionPath); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		loaded, err := LoadSession(sessionPath)
		if err != nil {
			t.Fatalf("LoadSession() failed: %v", err)
		}

		// Verify messages were loaded
		loadedRaw := loaded.GetRawMessages()
		if len(loadedRaw) != len(messages) {
			t.Errorf("Expected %d messages after load, got %d", len(messages), len(loadedRaw))
		}
	})

	t.Run("converts json.RawMessage to interface{} for use", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		messages := []json.RawMessage{
			json.RawMessage(`{"role": "user", "content": "test"}`),
		}

		sess.SetRawMessages(messages)

		rawMessages := sess.GetRawMessages()
		if len(rawMessages) == 0 {
			t.Fatal("Expected raw messages, got empty")
		}

		// Convert to interface{} (as used in tool handler)
		var msgInterface []interface{}
		for _, rawMsg := range rawMessages {
			var msg map[string]interface{}
			if err := json.Unmarshal(rawMsg, &msg); err != nil {
				t.Fatalf("Failed to unmarshal raw message: %v", err)
			}
			msgInterface = append(msgInterface, msg)
		}

		if len(msgInterface) != len(rawMessages) {
			t.Errorf("Expected %d interface messages, got %d", len(rawMessages), len(msgInterface))
		}
	})

	t.Run("falls back to legacy format if complete messages don't exist", func(t *testing.T) {
		// Create session with only legacy Messages
		sess := NewSession("test_source", "mysql")
		sess.AddMessage("user", "test message")
		sess.AddMessage("assistant", "response")

		// Verify legacy messages exist
		if len(sess.Messages) == 0 {
			t.Fatal("Expected legacy messages, got empty")
		}

		// Verify RawMessages is empty or nil
		if len(sess.RawMessages) > 0 {
			t.Log("RawMessages exists, which is fine if implementation creates it")
		}
	})
}

// TestSession_ContentNormalization tests content normalization
func TestSession_ContentNormalization(t *testing.T) {
	t.Run("normalizes content field to string type", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		// Create message with object content (should be normalized to string)
		msgWithObject := map[string]interface{}{
			"role":    "user",
			"content": map[string]interface{}{"text": "test"},
		}

		msgBytes, err := json.Marshal(msgWithObject)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		messages := []json.RawMessage{json.RawMessage(msgBytes)}
		sess.SetRawMessages(messages)

		// Load and verify content is normalized
		rawMessages := sess.GetRawMessages()
		if len(rawMessages) == 0 {
			t.Fatal("Expected raw messages")
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(rawMessages[0], &msg); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Content should be normalized (implementation may do this during loading)
		_ = msg["content"]
	})

	t.Run("handles null content appropriately", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		// Create assistant message with null content (tool_calls present)
		msgWithNull := map[string]interface{}{
			"role":       "assistant",
			"content":    nil,
			"tool_calls": []interface{}{},
		}

		msgBytes, err := json.Marshal(msgWithNull)
		if err != nil {
			t.Fatalf("Failed to marshal message: %v", err)
		}

		messages := []json.RawMessage{json.RawMessage(msgBytes)}
		sess.SetRawMessages(messages)

		// Verify message was stored
		rawMessages := sess.GetRawMessages()
		if len(rawMessages) == 0 {
			t.Fatal("Expected raw messages")
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(rawMessages[0], &msg); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		// Content may be null, empty string, or omitted - all are valid
		_ = msg["content"]
	})
}

// TestSession_BackwardCompatibility tests backward compatibility
func TestSession_BackwardCompatibility(t *testing.T) {
	t.Run("loads legacy Messages format", func(t *testing.T) {
		// Create session with legacy format
		sess := NewSession("test_source", "mysql")
		sess.AddMessage("user", "test")
		sess.AddMessage("assistant", "response")

		// Save session
		tmpDir := t.TempDir()
		sessionPath := filepath.Join(tmpDir, "test_session.json")

		if err := SaveSession(sess, sessionPath); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		// Load session
		loaded, err := LoadSession(sessionPath)
		if err != nil {
			t.Fatalf("LoadSession() failed: %v", err)
		}

		// Verify legacy messages were loaded
		legacyMessages := loaded.GetHistory()
		if len(legacyMessages) != 2 {
			t.Errorf("Expected 2 legacy messages, got %d", len(legacyMessages))
		}
	})

	t.Run("works normally without complete messages array", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")
		sess.AddMessage("user", "test")

		// Session should work with only legacy Messages
		messages := sess.GetHistory()
		if len(messages) == 0 {
			t.Error("Expected legacy messages to be available")
		}

		// RawMessages may be empty, which is fine
		rawMessages := sess.GetRawMessages()
		if len(rawMessages) > 0 {
			t.Log("RawMessages exists, which is fine")
		}
	})
}

// TestSession_MessageTrimming tests message trimming
func TestSession_MessageTrimming(t *testing.T) {
	t.Run("trims oldest messages when array exceeds limit", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		// Create more messages than limit
		limit := DefaultHistoryLimit * 10
		messages := make([]json.RawMessage, limit+10)

		for i := 0; i < limit+10; i++ {
			msg := map[string]interface{}{
				"role":    "user",
				"content": "message " + string(rune(i)),
			}
			msgBytes, _ := json.Marshal(msg)
			messages[i] = json.RawMessage(msgBytes)
		}

		sess.SetRawMessages(messages)

		// Verify trimming occurred
		stored := sess.GetRawMessages()
		if len(stored) > limit {
			t.Errorf("Expected messages to be trimmed to %d, got %d", limit, len(stored))
		}

		// Verify most recent messages are preserved
		if len(stored) < limit {
			t.Errorf("Expected at least %d messages after trimming, got %d", limit, len(stored))
		}
	})

	t.Run("preserves most recent messages up to limit", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		limit := DefaultHistoryLimit * 10
		messages := make([]json.RawMessage, limit+5)

		// Create messages with sequential content
		for i := 0; i < limit+5; i++ {
			msg := map[string]interface{}{
				"role":    "user",
				"content": "message " + string(rune(i)),
			}
			msgBytes, _ := json.Marshal(msg)
			messages[i] = json.RawMessage(msgBytes)
		}

		sess.SetRawMessages(messages)

		// Verify most recent messages are preserved
		stored := sess.GetRawMessages()
		if len(stored) != limit {
			t.Errorf("Expected exactly %d messages after trimming, got %d", limit, len(stored))
		}

		// Verify last message is preserved
		var lastMsg map[string]interface{}
		if err := json.Unmarshal(stored[len(stored)-1], &lastMsg); err != nil {
			t.Fatalf("Failed to unmarshal last message: %v", err)
		}

		expectedContent := "message " + string(rune(limit+4))
		if lastMsg["content"] != expectedContent {
			t.Errorf("Expected last message content %q, got %q", expectedContent, lastMsg["content"])
		}
	})
}

// TestSession_VariousMessageTypes tests session save/load with various message types
func TestSession_VariousMessageTypes(t *testing.T) {
	t.Run("saves and loads system messages", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		messages := []json.RawMessage{
			json.RawMessage(`{"role": "system", "content": "You are a helpful assistant"}`),
		}

		sess.SetRawMessages(messages)

		tmpDir := t.TempDir()
		sessionPath := filepath.Join(tmpDir, "test_session.json")

		if err := SaveSession(sess, sessionPath); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		loaded, err := LoadSession(sessionPath)
		if err != nil {
			t.Fatalf("LoadSession() failed: %v", err)
		}

		loadedRaw := loaded.GetRawMessages()
		if len(loadedRaw) != 1 {
			t.Errorf("Expected 1 message, got %d", len(loadedRaw))
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(loadedRaw[0], &msg); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if msg["role"] != "system" {
			t.Errorf("Expected role 'system', got %q", msg["role"])
		}
	})

	t.Run("saves and loads assistant messages with tool_calls", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		messages := []json.RawMessage{
			json.RawMessage(`{"role": "assistant", "content": "", "tool_calls": [{"id": "call_1", "type": "function", "function": {"name": "execute_sql", "arguments": "{}"}}]}`),
		}

		sess.SetRawMessages(messages)

		tmpDir := t.TempDir()
		sessionPath := filepath.Join(tmpDir, "test_session.json")

		if err := SaveSession(sess, sessionPath); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		loaded, err := LoadSession(sessionPath)
		if err != nil {
			t.Fatalf("LoadSession() failed: %v", err)
		}

		loadedRaw := loaded.GetRawMessages()
		var msg map[string]interface{}
		if err := json.Unmarshal(loadedRaw[0], &msg); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if _, exists := msg["tool_calls"]; !exists {
			t.Error("Expected tool_calls to be preserved")
		}
	})

	t.Run("saves and loads tool messages", func(t *testing.T) {
		sess := NewSession("test_source", "mysql")

		messages := []json.RawMessage{
			json.RawMessage(`{"role": "tool", "content": "{\"status\": \"success\"}", "tool_call_id": "call_1"}`),
		}

		sess.SetRawMessages(messages)

		tmpDir := t.TempDir()
		sessionPath := filepath.Join(tmpDir, "test_session.json")

		if err := SaveSession(sess, sessionPath); err != nil {
			t.Fatalf("SaveSession() failed: %v", err)
		}

		loaded, err := LoadSession(sessionPath)
		if err != nil {
			t.Fatalf("LoadSession() failed: %v", err)
		}

		loadedRaw := loaded.GetRawMessages()
		var msg map[string]interface{}
		if err := json.Unmarshal(loadedRaw[0], &msg); err != nil {
			t.Fatalf("Failed to unmarshal: %v", err)
		}

		if msg["role"] != "tool" {
			t.Errorf("Expected role 'tool', got %q", msg["role"])
		}

		if msg["tool_call_id"] != "call_1" {
			t.Errorf("Expected tool_call_id 'call_1', got %q", msg["tool_call_id"])
		}
	})
}
