package session

import (
	"encoding/json"
	"time"
)

const (
	// DefaultHistoryLimit is the default maximum number of message pairs to keep
	DefaultHistoryLimit = 20
)

// AddMessage adds a message to the conversation history
func (s *Session) AddMessage(role, content string) {
	message := Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now().UTC(),
	}

	s.Messages = append(s.Messages, message)
	s.UpdateLastUpdated()

	// Trim history if it exceeds the limit
	s.trimHistory(DefaultHistoryLimit)
}

// GetHistory returns all conversation messages
func (s *Session) GetHistory() []Message {
	return s.Messages
}

// ClearHistory clears all conversation messages
func (s *Session) ClearHistory() {
	s.Messages = make([]Message, 0)
	s.UpdateLastUpdated()
}

// trimHistory trims the conversation history to keep only the most recent messages
// Keeps the most recent `limit` message pairs (limit * 2 messages total)
func (s *Session) trimHistory(limit int) {
	maxMessages := limit * 2 // Each pair has user + assistant message

	if len(s.Messages) <= maxMessages {
		return
	}

	// Keep only the most recent messages
	trimCount := len(s.Messages) - maxMessages
	s.Messages = s.Messages[trimCount:]
}

// GetHistoryLimit returns the current history limit
func GetHistoryLimit() int {
	return DefaultHistoryLimit
}

// AddRawMessages adds complete messages array to session (includes tool calls and results)
// This preserves full conversation context including tool executions
// rawMessages should be the new messages to append (excluding system message and user input that were already in session)
func (s *Session) AddRawMessages(rawMessages []json.RawMessage) {
	s.RawMessages = append(s.RawMessages, rawMessages...)
	s.UpdateLastUpdated()

	// Trim history if it exceeds the limit (keep last N messages)
	s.trimRawHistory(DefaultHistoryLimit * 10) // More messages since we include tool calls/results
}

// GetRawMessages returns complete messages array (includes tool calls and results)
func (s *Session) GetRawMessages() []json.RawMessage {
	return s.RawMessages
}

// SetRawMessages sets the complete messages array and trims if needed
func (s *Session) SetRawMessages(rawMessages []json.RawMessage) {
	s.RawMessages = rawMessages
	s.UpdateLastUpdated()

	// Trim history if it exceeds the limit (keep last N messages)
	s.trimRawHistory(DefaultHistoryLimit * 10) // More messages since we include tool calls/results
}

// trimRawHistory trims the raw messages history to keep only the most recent messages
func (s *Session) trimRawHistory(maxMessages int) {
	if len(s.RawMessages) <= maxMessages {
		return
	}

	// Keep only the most recent messages
	trimCount := len(s.RawMessages) - maxMessages
	s.RawMessages = s.RawMessages[trimCount:]
}
