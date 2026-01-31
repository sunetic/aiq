# Multi-Turn Conversation Specification

## Overview

This specification defines the requirements for adding multi-turn conversation support to AIQ's chat mode, enabling context-aware queries and session persistence.

## Requirements

### REQ-1: Conversation History Management

The system SHALL maintain conversation history during a chat session.

#### REQ-1.1: Message Storage
- **WHEN** a user enters a query
- **THEN** the system stores the user query as a message with role "user"
- **AND** stores the LLM's response as a message with role "assistant"
- **AND** includes timestamps for each message
- **AND** tracks tool execution outcomes (success/failure, error information, state changes)

#### REQ-1.2: History Limit
- **WHEN** conversation history exceeds the configured limit (default: 20 message pairs)
- **THEN** the system removes the oldest messages
- **AND** keeps the most recent messages within the limit

#### REQ-1.3: Tool Execution Tracking
- **WHEN** tools are executed during conversation
- **THEN** system tracks tool execution details (tool name, arguments, status, error information, state changes)
- **AND** maintains last N tool executions (default: 5) for summary generation

#### REQ-1.4: Generate Tool Execution Summary
- **WHEN** system prepares to send request to LLM
- **THEN** system generates concise summary of recent tool executions (last 3-5 executions)
- **AND** includes summary in conversation context sent to LLM
- **AND** highlights state changes (resources created/deleted, dependencies resolved)

#### REQ-1.5: Include Summary in Conversation Context
- **WHEN** sending request to LLM
- **THEN** tool execution summary is included as separate section before user query
- **AND** summary is clearly marked (e.g., `<TOOL_EXECUTION_SUMMARY>`)
- **AND** summary is additive to full conversation history (does not replace it)

#### REQ-1.6: Reset Summary for New Query
- **WHEN** user submits new query
- **THEN** system resets tool execution tracking
- **AND** starts fresh summary for new conversation turn

### REQ-2: Context-Aware LLM Queries

The system SHALL send conversation history along with the current query to the LLM.

#### REQ-2.1: Message Format
- **WHEN** sending a query to the LLM
- **THEN** the system formats messages as:
  - System message: SQL expert instructions
  - Previous conversation messages (user + assistant pairs)
  - Current user query

#### REQ-2.2: Message Ordering
- **WHEN** building the message list for LLM
- **THEN** messages are ordered chronologically
- **AND** the current query is the last message

### REQ-3: Session Persistence

The system SHALL save conversation sessions to disk when exiting chat mode.

#### REQ-3.1: Session File Format
- **WHEN** saving a session
- **THEN** the system creates a JSON file at `~/.aiqconfig/session_<timestamp>.json`
- **AND** includes metadata (created_at, last_updated, data_source, database_type)
- **AND** includes all conversation messages

#### REQ-3.2: Session File Naming
- **WHEN** creating a session file
- **THEN** the filename format is `session_YYYYMMDDHHMMSS.json`
- **AND** uses UTC timestamp

#### REQ-3.3: Exit Message
- **WHEN** exiting chat mode
- **THEN** the system displays:
  ```
  Current session saved to ~/.aiqconfig/session_20260126100000.json
  Run 'aiq -s ~/.aiqconfig/session_20260126100000.json' to continue.
  ```

### REQ-4: Session Restoration

The system SHALL support restoring previous conversations via command-line flag.

#### REQ-4.1: Command-Line Flag
- **WHEN** starting AIQ with `-s <file>` or `--session <file>`
- **THEN** the system loads the session file
- **AND** restores conversation history
- **AND** uses the same data source from the session

#### REQ-4.2: Session Validation
- **WHEN** loading a session file
- **THEN** the system validates JSON format
- **AND** checks for required fields (metadata, messages)
- **AND** handles corrupted or invalid files gracefully

#### REQ-4.3: Data Source Restoration
- **WHEN** restoring a session
- **THEN** the system attempts to use the same data source from session metadata
- **AND** if data source no longer exists, prompts user to select a new one

### REQ-5: Conversation History Display

The system SHALL provide a way to view conversation history during a session.

#### REQ-5.1: History Command
- **WHEN** user enters `/history` command
- **THEN** the system displays all conversation messages
- **AND** shows role, content, and timestamp for each message

#### REQ-5.2: Clear History Command
- **WHEN** user enters `/clear` command
- **THEN** the system clears conversation history
- **AND** prompts for confirmation
- **AND** starts fresh conversation from that point

### REQ-6: Backward Compatibility

The system SHALL maintain backward compatibility with existing single-query behavior.

#### REQ-6.1: Default Behavior
- **WHEN** starting chat mode without session flag
- **THEN** the system creates a new empty session
- **AND** conversation history starts empty
- **AND** first query works as before (no context)

#### REQ-6.2: Single Query Mode
- **WHEN** conversation history is empty
- **THEN** the system sends only system message and current query
- **AND** behaves identically to current implementation

## Implementation Details

### Session Structure

```go
type Session struct {
    Metadata SessionMetadata `json:"metadata"`
    Messages []Message       `json:"messages"`
}

type SessionMetadata struct {
    CreatedAt    time.Time `json:"created_at"`
    LastUpdated  time.Time `json:"last_updated"`
    DataSource  string    `json:"data_source"`
    DatabaseType string   `json:"database_type"`
}

type Message struct {
    Role      string    `json:"role"`      // "user" or "assistant"
    Content   string    `json:"content"`
    Timestamp time.Time `json:"timestamp"`
}
```

### LLM Client Modification

```go
// TranslateToSQL now accepts conversation history
func (c *Client) TranslateToSQL(
    ctx context.Context,
    naturalLanguage string,
    schemaContext string,
    databaseType string,
    conversationHistory []ChatMessage, // New parameter
) (string, error)
```

### CLI Flag Addition

```go
var sessionFile string

rootCmd.PersistentFlags().StringVarP(
    &sessionFile,
    "session", "s",
    "",
    "Restore conversation from session file",
)
```

## Constraints

- Maximum conversation history: 20 message pairs (configurable)
- Session file size limit: 10MB (to prevent abuse)
- Message content length: No explicit limit (handled by LLM token limits)
- Session file encoding: UTF-8 JSON

## Error Handling

- **Invalid session file**: Display error, create new session
- **Missing data source**: Prompt user to select new data source
- **File write failure**: Display warning, continue without saving
- **LLM context too long**: Trim oldest messages automatically
