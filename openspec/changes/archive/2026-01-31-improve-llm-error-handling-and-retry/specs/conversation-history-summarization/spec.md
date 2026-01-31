## ADDED Requirements

### Requirement: Tool execution summary generation
The system SHALL generate concise summaries of recent tool executions highlighting successes, failures, and state changes.

#### Scenario: Track recent tool executions
- **WHEN** tools are executed during conversation
- **THEN** system tracks last N tool executions (default: 5) with execution details (tool name, arguments, status, error information)

#### Scenario: Generate summary before LLM call
- **WHEN** system prepares to send request to LLM
- **THEN** system generates summary of recent tool executions (last 3-5 executions)
- **AND** includes summary in conversation context sent to LLM

#### Scenario: Include execution status in summary
- **WHEN** generating tool execution summary
- **THEN** summary includes status for each execution (SUCCESS or FAILED with error type)

#### Scenario: Highlight state changes in summary
- **WHEN** generating tool execution summary
- **THEN** summary explicitly calls out state changes (resources created/deleted, dependencies resolved)

#### Scenario: Include error information in summary
- **WHEN** tool execution failed
- **THEN** summary includes error type and affected resources from structured error information

#### Scenario: Keep summary concise
- **WHEN** generating tool execution summary
- **THEN** summary includes only last 3-5 tool executions
- **AND** uses concise format to minimize token usage

#### Scenario: Reset summary for new user query
- **WHEN** user submits new query
- **THEN** system resets tool execution tracking
- **AND** starts fresh summary for new conversation turn

### Requirement: State change detection
The system SHALL detect state changes from tool executions that may affect retry decisions.

#### Scenario: Detect DDL operations as state changes
- **WHEN** tool execution performs DDL operation (CREATE/DROP/ALTER table)
- **THEN** system marks execution as causing state change
- **AND** includes state change information in summary

#### Scenario: Detect dependency resolution
- **WHEN** tool execution successfully deletes resource that was blocking previous operations
- **THEN** system marks this as dependency resolution
- **AND** highlights in summary that dependencies have been resolved

#### Scenario: Track resource creation
- **WHEN** tool execution creates new resource (table, column, etc.)
- **THEN** system tracks resource creation in state changes
- **AND** includes in summary

#### Scenario: Track resource deletion
- **WHEN** tool execution deletes resource (table, column, etc.)
- **THEN** system tracks resource deletion in state changes
- **AND** includes in summary

### Requirement: Summary format
The system SHALL format tool execution summaries in a clear, LLM-readable format.

#### Scenario: Use structured summary format
- **WHEN** generating tool execution summary
- **THEN** summary uses structured format with clear sections:
  - List of recent tool executions with status
  - Explicit state changes section
  - Dependencies resolved section (if applicable)

#### Scenario: Include summary in conversation context
- **WHEN** sending request to LLM
- **THEN** summary is included as separate section before user query
- **AND** clearly marked as tool execution summary (e.g., `<TOOL_EXECUTION_SUMMARY>`)

#### Scenario: Summary is additive to full history
- **WHEN** sending request to LLM
- **THEN** summary is provided in addition to full conversation history
- **AND** does not replace full history
