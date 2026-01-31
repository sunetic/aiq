## 1. Prompt Version Detection Infrastructure

- [ ] 1.1 Add content hash calculation function (`hashContent`) in `internal/prompt/loader.go`
- [ ] 1.2 Implement `getBuiltInPromptHashes()` to calculate hashes for all built-in prompts
- [ ] 1.3 Implement `getUserPromptHashes()` to calculate hashes for user prompt files
- [ ] 1.4 Implement `checkPromptContentMismatch()` to compare hashes and return modified files
- [ ] 1.5 Implement `hasContentMismatch()` helper function

## 2. Version-Based Choice Storage

- [ ] 2.1 Add `GetChoiceForVersion(version string)` function in `internal/config` package
- [ ] 2.2 Add `SetChoiceForVersion(version string, choice string)` function in `internal/config` package
- [ ] 2.3 Implement storage mechanism (config file or database) for version choices
- [ ] 2.4 Integrate with `internal/version` package to get current application version

## 3. Prompt Upgrade Detection and User Prompting

- [ ] 3.1 Implement `checkAndHandleVersionMismatch()` in `internal/prompt/loader.go`
- [ ] 3.2 Add logic to check if choice exists for current version
- [ ] 3.3 Implement `promptForVersionUpgrade()` function to ask user for choice
- [ ] 3.4 Store user choice after prompting
- [ ] 3.5 Display guidance message when user chooses "keep"
- [ ] 3.6 Integrate version check into `NewLoader()` before `initializeDefaults()`

## 4. Respect User Choice in Prompt Initialization

- [ ] 4.1 Modify `initializeDefaults()` to check user choice for current version
- [ ] 4.2 Implement logic: if choice is "overwrite", overwrite existing files
- [ ] 4.3 Implement logic: if choice is "keep", skip file creation if files exist
- [ ] 4.4 Ensure files are still created if they don't exist (regardless of choice)

## 5. Session Structure Extension

- [ ] 5.1 Add `RawMessages []json.RawMessage` field to `Session` struct in `internal/session/session.go`
- [ ] 5.2 Add `GetRawMessages()` method to return raw messages array
- [ ] 5.3 Add `SetRawMessages()` method to set raw messages array with trimming
- [ ] 5.4 Ensure backward compatibility: keep `Messages []Message` field

## 6. Complete Messages Array Persistence

- [ ] 6.1 Modify `HandleToolCallLoop` to accept `rawMessages []interface{}` parameter
- [ ] 6.2 Modify `HandleToolCallLoop` to return complete messages array as third return value
- [ ] 6.3 Update `mode.go` to load `RawMessages` from Session before calling tool handler
- [ ] 6.4 Update `mode.go` to save complete messages array to Session after tool execution
- [ ] 6.5 Convert `json.RawMessage` to `[]interface{}` when loading from Session
- [ ] 6.6 Convert `[]interface{}` to `[]json.RawMessage` when saving to Session

## 7. Message Content Normalization - Loading Layer

- [ ] 7.1 Add normalization logic in `mode.go` when loading rawMessages from Session
- [ ] 7.2 Handle content field type conversion (object/array → JSON string)
- [ ] 7.3 Handle null content fields (remove for tool messages, set empty string for assistant)
- [ ] 7.4 Ensure assistant messages always have content field (empty string if null)

## 8. Message Content Normalization - Building Layer

- [ ] 8.1 Convert `llm.ChatMessage` structs to `map[string]interface{}` in `tool_handler.go`
- [ ] 8.2 Ensure system message uses map format with string content
- [ ] 8.3 Ensure user message uses map format with string content
- [ ] 8.4 Ensure conversation history messages use map format
- [ ] 8.5 Normalize content fields when processing rawMessages from Session

## 9. Message Content Normalization - API Layer

- [ ] 9.1 Add normalization logic in `ChatWithTools` before sending to LLM API
- [ ] 9.2 Convert all messages to map format if needed
- [ ] 9.3 Ensure content fields are strings (convert objects/arrays to JSON strings)
- [ ] 9.4 Handle ChatMessage structs by converting to maps
- [ ] 9.5 Add final safety check for first message content

## 10. Assistant Message Content Handling

- [ ] 10.1 Modify assistant message creation to always include content field
- [ ] 10.2 Set content to empty string (not null) when LLM returns empty content
- [ ] 10.3 Preserve tool_calls field when present
- [ ] 10.4 Ensure proper serialization format: `{"role": "assistant", "content": "", "tool_calls": [...]}`

## 11. Reduce Redundant Output - Tool Result Simplification

- [ ] 11.1 Remove `result_summary` field from tool result when results are displayed
- [ ] 11.2 Update tool result instruction to be more explicit about not repeating results
- [ ] 11.3 Ensure `displayed: true` flag is set when table output is shown

## 12. Reduce Redundant Output - Display Logic

- [ ] 12.1 Modify `mode.go` to skip summary display when `finalResponse` is empty and `queryResult` exists
- [ ] 12.2 Remove logic that displays `formatQueryResultSummary` for already-displayed results
- [ ] 12.3 Ensure summary is still saved to legacy Messages for backward compatibility

## 13. Prompt Instruction Strengthening

- [ ] 13.1 Update prompt in `internal/prompt/loader.go` to include guidance about `displayed: true`
- [ ] 13.2 Add explicit instruction: "If tool result contains 'displayed': true, do NOT repeat results"
- [ ] 13.3 Add examples of correct vs incorrect behavior in prompt

## 14. Testing and Validation

- [ ] 14.1 Test prompt upgrade detection with modified files
- [ ] 14.2 Test version-based choice persistence across restarts
- [ ] 14.3 Test complete message loading/saving with various message types
- [ ] 14.4 Test content normalization with edge cases (null, objects, arrays)
- [ ] 14.5 Test backward compatibility with legacy Session files
- [ ] 14.6 Test redundant output reduction (verify no duplicate summaries)
- [ ] 14.7 Test message trimming when array exceeds limit

## 15. Cleanup and Documentation

- [ ] 15.1 Remove any unused code or redundant normalization logic
- [ ] 15.2 Add code comments explaining normalization layers
- [ ] 15.3 Verify all error cases are handled gracefully
- [ ] 15.4 Update any relevant documentation
