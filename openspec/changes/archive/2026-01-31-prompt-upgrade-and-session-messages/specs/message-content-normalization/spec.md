## ADDED Requirements

### Requirement: Content field type safety

The system SHALL ensure all message content fields are strings (not objects) for LLM API compatibility.

#### Scenario: Normalize when loading from Session
- **WHEN** loading messages from Session (`json.RawMessage` → `map[string]interface{}`)
- **THEN** system checks content field type
- **AND** if content is object/array, converts to JSON string
- **AND** if content is null, handles appropriately (remove for tool messages, set empty string for assistant messages)
- **AND** if content is other type, converts to string

#### Scenario: Normalize when building messages
- **WHEN** building messages array in tool handler
- **THEN** system ensures all messages use `map[string]interface{}` format
- **AND** system converts `ChatMessage` structs to maps
- **AND** system ensures content field is always string type

#### Scenario: Final normalization before API call
- **WHEN** sending messages to LLM API
- **THEN** system performs final normalization pass
- **AND** system converts all messages to maps if needed
- **AND** system ensures content fields are strings
- **AND** system handles first message specially (most critical)

### Requirement: Assistant message content handling

The system SHALL handle assistant messages with tool_calls properly (content may be empty or null).

#### Scenario: Assistant message with tool_calls
- **WHEN** LLM returns assistant message with tool_calls but empty content
- **THEN** system sets content to empty string (not null, not omitted)
- **AND** system preserves tool_calls field
- **AND** message format: `{"role": "assistant", "content": "", "tool_calls": [...]}`

#### Scenario: Assistant message without tool_calls
- **WHEN** LLM returns assistant message with content but no tool_calls
- **THEN** system preserves content as-is
- **AND** message format: `{"role": "assistant", "content": "..."}`

## MODIFIED Requirements

### Requirement: Message serialization

The system SHALL serialize messages consistently to avoid type errors.

#### Scenario: Consistent message format
- **WHEN** serializing messages for Session storage
- **THEN** all messages are serialized as JSON objects
- **AND** content fields are always strings (never objects)
- **AND** tool_calls are preserved as arrays
- **AND** tool_call_id is preserved for tool messages

#### Scenario: Deserialization safety
- **WHEN** deserializing messages from Session
- **THEN** system normalizes content fields immediately
- **AND** system handles type mismatches gracefully
- **AND** system skips invalid messages rather than failing
