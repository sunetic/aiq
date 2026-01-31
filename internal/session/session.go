package session

import (
	"encoding/json"
	"time"
)

// Message represents a single message in the conversation (legacy format, for backward compatibility)
type Message struct {
	Role      string    `json:"role"` // "user" or "assistant"
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// SessionMetadata contains metadata about the session
type SessionMetadata struct {
	CreatedAt    time.Time `json:"created_at"`
	LastUpdated  time.Time `json:"last_updated"`
	DataSource   string    `json:"data_source"`
	DatabaseType string    `json:"database_type"`
}

// Session represents a conversation session
type Session struct {
	Metadata    SessionMetadata   `json:"metadata"`
	Messages    []Message         `json:"messages,omitempty"`     // Legacy format, for backward compatibility
	RawMessages []json.RawMessage `json:"raw_messages,omitempty"` // Complete messages array (includes tool calls and results)
}

// NewSession creates a new session with the given data source and database type
func NewSession(dataSource, databaseType string) *Session {
	now := time.Now().UTC()
	return &Session{
		Metadata: SessionMetadata{
			CreatedAt:    now,
			LastUpdated:  now,
			DataSource:   dataSource,
			DatabaseType: databaseType,
		},
		Messages: make([]Message, 0),
	}
}

// UpdateLastUpdated updates the last updated timestamp
func (s *Session) UpdateLastUpdated() {
	s.Metadata.LastUpdated = time.Now().UTC()
}

// GetTimestamp generates a timestamp string for session file naming
// Format: YYYYMMDDHHMMSS (UTC)
func GetTimestamp() string {
	return time.Now().UTC().Format("20060102150405")
}
