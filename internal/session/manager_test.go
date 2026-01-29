package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestSaveSession(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test_session.json")

	sess := NewSession("test_source", "mysql")
	sess.AddMessage("user", "What is the total revenue?")
	sess.AddMessage("assistant", "SELECT SUM(revenue) FROM sales")

	err := SaveSession(sess, sessionPath)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(sessionPath); os.IsNotExist(err) {
		t.Fatal("Session file was not created")
	}

	// Verify file content
	data, err := os.ReadFile(sessionPath)
	if err != nil {
		t.Fatalf("Failed to read session file: %v", err)
	}

	var loaded Session
	if err := json.Unmarshal(data, &loaded); err != nil {
		t.Fatalf("Failed to unmarshal session file: %v", err)
	}

	if loaded.Metadata.DataSource != "test_source" {
		t.Errorf("Expected DataSource 'test_source', got '%s'", loaded.Metadata.DataSource)
	}

	if len(loaded.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(loaded.Messages))
	}
}

func TestLoadSession(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "test_session.json")

	// Create a valid session file
	sess := NewSession("test_source", "mysql")
	sess.AddMessage("user", "Query 1")
	sess.AddMessage("assistant", "Response 1")

	err := SaveSession(sess, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// Load session
	loaded, err := LoadSession(sessionPath)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	if loaded.Metadata.DataSource != "test_source" {
		t.Errorf("Expected DataSource 'test_source', got '%s'", loaded.Metadata.DataSource)
	}

	if loaded.Metadata.DatabaseType != "mysql" {
		t.Errorf("Expected DatabaseType 'mysql', got '%s'", loaded.Metadata.DatabaseType)
	}

	if len(loaded.Messages) != 2 {
		t.Errorf("Expected 2 messages, got %d", len(loaded.Messages))
	}

	if loaded.Messages[0].Content != "Query 1" {
		t.Errorf("Expected first message 'Query 1', got '%s'", loaded.Messages[0].Content)
	}
}

func TestLoadSessionNotFound(t *testing.T) {
	_, err := LoadSession("/nonexistent/path/session.json")
	if err == nil {
		t.Fatal("Expected error for nonexistent file")
	}
}

func TestLoadSessionInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "invalid.json")

	// Write invalid JSON
	err := os.WriteFile(sessionPath, []byte("{ invalid json }"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	_, err = LoadSession(sessionPath)
	if err == nil {
		t.Fatal("Expected error for invalid JSON")
	}
}

func TestLoadSessionMissingFields(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "incomplete.json")

	// Create incomplete session (missing DataSource)
	incomplete := struct {
		Metadata struct {
			DatabaseType string `json:"database_type"`
		} `json:"metadata"`
		Messages []Message `json:"messages"`
	}{
		Metadata: struct {
			DatabaseType string `json:"database_type"`
		}{
			DatabaseType: "mysql",
		},
		Messages: []Message{},
	}

	data, err := json.Marshal(incomplete)
	if err != nil {
		t.Fatalf("Failed to marshal incomplete session: %v", err)
	}

	err = os.WriteFile(sessionPath, data, 0600)
	if err != nil {
		t.Fatalf("Failed to write incomplete session: %v", err)
	}

	_, err = LoadSession(sessionPath)
	if err == nil {
		t.Fatal("Expected error for missing DataSource field")
	}
}

func TestLoadSessionInvalidMessageRole(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "invalid_role.json")

	// Create session with invalid role
	sess := NewSession("test", "mysql")
	sess.Messages = []Message{
		{
			Role:      "invalid_role",
			Content:   "Test",
			Timestamp: time.Now(),
		},
	}

	err := SaveSession(sess, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// LoadSession should validate and fail
	_, err = LoadSession(sessionPath)
	if err == nil {
		t.Fatal("Expected error for invalid message role")
	}
}

func TestLoadSessionEmptyMessageContent(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "empty_content.json")

	// Create session with empty content
	sess := NewSession("test", "mysql")
	sess.Messages = []Message{
		{
			Role:      "user",
			Content:   "",
			Timestamp: time.Now(),
		},
	}

	err := SaveSession(sess, sessionPath)
	if err != nil {
		t.Fatalf("Failed to save session: %v", err)
	}

	// LoadSession should validate and fail
	_, err = LoadSession(sessionPath)
	if err == nil {
		t.Fatal("Expected error for empty message content")
	}
}

func TestGetSessionFilePath(t *testing.T) {
	timestamp := "20260126100000"
	path, err := GetSessionFilePath(timestamp)
	if err != nil {
		t.Fatalf("GetSessionFilePath failed: %v", err)
	}

	expectedFileName := "session_20260126100000.json"
	if filepath.Base(path) != expectedFileName {
		t.Errorf("Expected filename '%s', got '%s'", expectedFileName, filepath.Base(path))
	}

	// Verify it contains .aiq
	homeDir, err := os.UserHomeDir()
	if err == nil {
		expectedPrefix := filepath.Join(homeDir, ".aiq")
		if !filepath.HasPrefix(path, expectedPrefix) {
			t.Errorf("Path should start with %s, got %s", expectedPrefix, path)
		}
	}
}

func TestSessionSaveAndLoadRoundTrip(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "roundtrip.json")

	// Create original session
	original := NewSession("test_source", "seekdb")
	original.AddMessage("user", "First query")
	original.AddMessage("assistant", "First response")
	original.AddMessage("user", "Second query")
	original.AddMessage("assistant", "Second response")

	// Save
	err := SaveSession(original, sessionPath)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Load
	loaded, err := LoadSession(sessionPath)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	// Verify metadata
	if loaded.Metadata.DataSource != original.Metadata.DataSource {
		t.Errorf("DataSource mismatch: expected '%s', got '%s'",
			original.Metadata.DataSource, loaded.Metadata.DataSource)
	}

	if loaded.Metadata.DatabaseType != original.Metadata.DatabaseType {
		t.Errorf("DatabaseType mismatch: expected '%s', got '%s'",
			original.Metadata.DatabaseType, loaded.Metadata.DatabaseType)
	}

	// Verify messages
	if len(loaded.Messages) != len(original.Messages) {
		t.Fatalf("Message count mismatch: expected %d, got %d",
			len(original.Messages), len(loaded.Messages))
	}

	for i, msg := range original.Messages {
		if loaded.Messages[i].Role != msg.Role {
			t.Errorf("Message %d role mismatch: expected '%s', got '%s'",
				i, msg.Role, loaded.Messages[i].Role)
		}
		if loaded.Messages[i].Content != msg.Content {
			t.Errorf("Message %d content mismatch: expected '%s', got '%s'",
				i, msg.Content, loaded.Messages[i].Content)
		}
	}
}

func TestSessionWithLongHistory(t *testing.T) {
	tmpDir := t.TempDir()
	sessionPath := filepath.Join(tmpDir, "long_history.json")

	// Create session with many messages
	sess := NewSession("test", "mysql")
	for i := 0; i < 50; i++ {
		if i%2 == 0 {
			sess.AddMessage("user", "Query")
		} else {
			sess.AddMessage("assistant", "Response")
		}
	}

	// Save
	err := SaveSession(sess, sessionPath)
	if err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	// Load
	loaded, err := LoadSession(sessionPath)
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}

	// Should be trimmed to 40 messages (20 pairs)
	if len(loaded.Messages) != 40 {
		t.Errorf("Expected 40 messages (trimmed), got %d", len(loaded.Messages))
	}
}
