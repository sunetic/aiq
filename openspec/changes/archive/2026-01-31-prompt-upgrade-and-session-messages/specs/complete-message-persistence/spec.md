## ADDED Requirements

### Requirement: Complete messages array storage

The system SHALL store complete messages array including tool calls and results in Session for full conversation context preservation.

#### Scenario: Save complete messages after tool execution
- **WHEN** tool execution completes and LLM returns final response
- **THEN** system saves entire messages array (system, user, assistant with tool_calls, tool messages with results) to Session
- **AND** messages array is serialized as JSON (`json.RawMessage` array)
- **AND** legacy `Messages []Message` field is also updated for backward compatibility

#### Scenario: Load complete messages on next turn
- **WHEN** user starts new conversation turn
- **THEN** system loads complete messages array from Session if available
- **AND** system converts `json.RawMessage` to `[]interface{}` for use in tool handler
- **AND** system skips old system message and uses new one (with updated Skills)
- **AND** if no complete messages exist, falls back to legacy `Messages []Message` format

#### Scenario: Message array structure
- **WHEN** messages array is saved
- **THEN** each message can be:
  - System message: `{"role": "system", "content": "..."}`
  - User message: `{"role": "user", "content": "..."}`
  - Assistant message: `{"role": "assistant", "content": "...", "tool_calls": [...]}`
  - Tool message: `{"role": "tool", "content": "...", "tool_call_id": "..."}`
- **AND** all messages preserve their original structure including tool calls and results

### Requirement: Backward compatibility

The system SHALL maintain backward compatibility with existing Session files.

#### Scenario: Legacy Session file
- **WHEN** Session file only contains legacy `Messages []Message` field
- **THEN** system loads legacy messages and converts to conversation history format
- **AND** system works normally without complete messages array
- **AND** new messages are saved in both formats (legacy and complete)

#### Scenario: New Session file
- **WHEN** Session file contains `RawMessages []json.RawMessage` field
- **THEN** system prioritizes complete messages array
- **AND** system uses complete messages for full context
- **AND** legacy messages field is still updated for compatibility

## MODIFIED Requirements

### Requirement: Conversation history loading

The system SHALL load complete conversation context including tool execution details.

#### Scenario: Load with tool context
- **WHEN** loading conversation history from Session
- **THEN** system loads complete messages array including:
  - Previous tool calls with parameters
  - Tool execution results (including truncated outputs)
  - Assistant responses with tool_calls
- **AND** LLM receives full context for better decision-making

#### Scenario: Context preservation
- **WHEN** user asks follow-up question
- **THEN** LLM can see:
  - Previous tool executions and their results
  - Table structures created/modified
  - Error messages and retry attempts
- **AND** LLM makes informed decisions based on full history
