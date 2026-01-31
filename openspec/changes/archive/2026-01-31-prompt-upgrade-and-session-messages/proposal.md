## Why

1. **Prompt upgrade detection**: When users modify prompt files in `~/.aiq/prompts`, upgrades become problematic - should we use existing prompts or overwrite with new versions? Need a detection mechanism to handle this gracefully.

2. **Session message persistence**: Tool execution details (commands, parameters, outputs) were not being persisted in conversation history, causing LLM to lose context across turns. Need to save complete messages array including tool calls and results.

3. **Message serialization issues**: Messages array serialization caused content field type errors when loading from Session, requiring normalization to ensure LLM API compatibility.

4. **Redundant output**: LLM was repeating query results that were already displayed in table format, creating redundant output.

## What Changes

### 1. Prompt Version Detection and Upgrade Mechanism
- **Content hash comparison**: Compare SHA256 hashes of built-in prompts vs user-modified prompts
- **Version-based choice tracking**: Store user's choice (overwrite/keep) per application version
- **One-time prompt per version**: Only prompt user once per version, remember choice for subsequent runs
- **User guidance**: If user chooses "keep", show message explaining how to upgrade later (delete files and restart)

### 2. Complete Messages Array Persistence
- **Session structure extension**: Add `RawMessages []json.RawMessage` field to Session for storing complete messages array
- **Full context preservation**: Save all messages including system, user, assistant (with tool_calls), and tool messages (with results)
- **Backward compatibility**: Keep legacy `Messages []Message` field for backward compatibility
- **JSON serialization**: Serialize entire messages array as JSON for flexible schema evolution

### 3. Message Content Normalization
- **Content field type safety**: Ensure all message content fields are strings (not objects) for LLM API compatibility
- **Normalization layers**: 
  - When loading from Session: Normalize content fields
  - When building messages: Ensure content is string type
  - Before sending to LLM API: Final normalization check
- **Null content handling**: Handle null content fields properly (assistant messages with tool_calls may have null content)

### 4. Reduce Redundant Output
- **Remove result summary from tool result**: Don't include `result_summary` in tool result when results are already displayed
- **Skip summary display**: When `finalResponse` is empty and `queryResult` exists, don't display summary (results already shown)
- **Prompt instruction**: Strengthen prompt to tell LLM not to repeat results when `displayed: true` in tool result

## Capabilities

### New Capabilities
- `prompt-version-detection`: Detect when user has modified prompt files and handle upgrade gracefully with version-based choice tracking
- `complete-message-persistence`: Persist full conversation context including tool calls and results in Session

### Modified Capabilities
- `sql-interactive-mode`: 
  - Load complete messages array from Session for full context
  - Save complete messages array after each turn
  - Normalize message content fields for LLM API compatibility
  - Reduce redundant output when results are already displayed
- `multi-turn-conversation`: Full context preservation across turns with tool execution details

## Impact

- **Affected code**:
  - `internal/prompt/loader.go`: 
    - Add content hash comparison (`checkPromptContentMismatch`)
    - Add version-based choice tracking (`checkAndHandleVersionMismatch`)
    - Integrate with `internal/config` and `internal/version` packages
  - `internal/session/session.go`: Add `RawMessages []json.RawMessage` field
  - `internal/session/history.go`: Add `GetRawMessages()`, `SetRawMessages()` methods
  - `internal/sql/mode.go`: 
    - Load complete messages array from Session
    - Save complete messages array after each turn
    - Normalize content fields when loading
    - Skip summary display when results already shown
  - `internal/sql/tool_handler.go`: 
    - Accept `rawMessages` parameter
    - Return complete messages array
    - Ensure assistant message content is always string (not null)
    - Remove `result_summary` from tool result when displayed
  - `internal/llm/client.go`: 
    - Normalize all messages before sending to LLM API
    - Ensure content fields are strings
    - Convert ChatMessage structs to maps for consistent serialization
  - `internal/prompt/loader.go`: Strengthen prompt instructions about not repeating displayed results

- **Architecture**:
  - **Session persistence**: Complete messages array serialized as JSON, allowing flexible schema evolution
  - **Message normalization**: Multiple normalization layers ensure LLM API compatibility
  - **Version tracking**: User choices stored per version, preventing repeated prompts
  - **Content hash comparison**: More reliable than version number comparison for detecting modifications

- **User experience**: 
  - Smooth prompt upgrade experience with one-time prompt per version
  - Full conversation context preserved across turns
  - Reduced redundant output for better readability
  - Clear guidance on how to upgrade prompts later if needed
