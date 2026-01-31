## MODIFIED Requirements

### Requirement: Conversation history management
The system SHALL maintain conversation history during a chat session, including tool execution summaries and state change tracking.

#### Scenario: Message storage
- **WHEN** a user enters a query
- **THEN** the system stores the user query as a message with role "user"
- **AND** stores the LLM's response as a message with role "assistant"
- **AND** includes timestamps for each message
- **AND** tracks tool execution outcomes (success/failure, error information, state changes)

#### Scenario: History limit
- **WHEN** conversation history exceeds the configured limit (default: 20 message pairs)
- **THEN** the system removes the oldest messages
- **AND** keeps the most recent messages within the limit

#### Scenario: Tool execution tracking
- **WHEN** tools are executed during conversation
- **THEN** system tracks tool execution details (tool name, arguments, status, error information, state changes)
- **AND** maintains last N tool executions (default: 5) for summary generation

#### Scenario: Generate tool execution summary
- **WHEN** system prepares to send request to LLM
- **THEN** system generates concise summary of recent tool executions (last 3-5 executions)
- **AND** includes summary in conversation context sent to LLM
- **AND** highlights state changes (resources created/deleted, dependencies resolved)

#### Scenario: Include summary in conversation context
- **WHEN** sending request to LLM
- **THEN** tool execution summary is included as separate section before user query
- **AND** summary is clearly marked (e.g., `<TOOL_EXECUTION_SUMMARY>`)
- **AND** summary is additive to full conversation history (does not replace it)

#### Scenario: Reset summary for new query
- **WHEN** user submits new query
- **THEN** system resets tool execution tracking
- **AND** starts fresh summary for new conversation turn
