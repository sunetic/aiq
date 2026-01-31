## Context

The AIQ application allows users to customize prompt files in `~/.aiq/prompts` for easier modification. However, when upgrading the application, there's uncertainty about whether to use existing user-modified prompts or overwrite them with new built-in versions. Additionally, the conversation history system was losing context across turns because tool execution details (commands, parameters, outputs) were not being persisted, causing LLM to make decisions without full context.

The current Session structure only stores simplified `Messages []Message` with `role` and `content` fields, losing rich information like tool calls and tool results. This causes context loss when LLM needs to reference previous tool executions.

## Goals / Non-Goals

**Goals:**
- Detect when user has modified prompt files and handle upgrades gracefully
- Store user's choice (overwrite/keep) per application version to avoid repeated prompts
- Persist complete conversation context including tool calls and results in Session
- Ensure message content fields are properly normalized for LLM API compatibility
- Reduce redundant output when results are already displayed to users

**Non-Goals:**
- Automatic prompt file migration or transformation
- Complex version comparison logic (using content hash instead)
- Real-time prompt file watching or hot-reload
- Backward compatibility with very old Session formats (only recent formats supported)

## Decisions

### 1. Content Hash Comparison vs Version Number Comparison

**Decision**: Use SHA256 content hash comparison instead of version number comparison for detecting prompt modifications.

**Rationale**:
- More reliable: Detects actual content changes, not just version bumps
- User-friendly: Works even if user doesn't update version field in prompt files
- Simpler: No need to maintain version fields in prompt files

**Alternatives Considered**:
- Version number in prompt frontmatter: Requires users to manually update versions, error-prone
- Timestamp comparison: Doesn't detect if user reverted changes
- File modification time: Unreliable across different systems

### 2. Version-Based Choice Storage

**Decision**: Store user's choice (overwrite/keep) per application version in config file.

**Rationale**:
- One-time prompt per version: User makes decision once, system remembers it
- Version-specific: New version = new prompt, so user gets chance to decide again
- Persistent: Choice survives restarts

**Alternatives Considered**:
- Per-file choice: Too granular, user would be prompted for each file
- Global choice: Doesn't account for version changes
- Session-based choice: Lost on restart

### 3. JSON RawMessage Array for Complete Messages

**Decision**: Store complete messages array as `[]json.RawMessage` in Session structure.

**Rationale**:
- Flexible schema: JSON format allows easy evolution without breaking changes
- Complete context: Preserves all message types (system, user, assistant with tool_calls, tool messages)
- Backward compatible: Can coexist with legacy `Messages []Message` field

**Alternatives Considered**:
- Extend `Message` struct: Would require schema migration, less flexible
- Separate tool calls/results storage: More complex, harder to maintain consistency
- Plain JSON string: Less type-safe, harder to work with in Go

### 4. Multiple Normalization Layers

**Decision**: Normalize message content fields at three layers: loading from Session, building messages, and before API call.

**Rationale**:
- Defense in depth: Ensures content is always string type regardless of source
- LLM API requirement: Content must be string or array of objects, not plain object
- Handles edge cases: Null content, object content, etc.

**Alternatives Considered**:
- Single normalization point: Less robust, might miss edge cases
- Only normalize at API call: Too late, might have already caused issues
- Type-safe structs only: Too restrictive, loses flexibility

### 5. Skip Summary Display When Results Already Shown

**Decision**: Don't display `formatQueryResultSummary` when `finalResponse` is empty and `queryResult` exists.

**Rationale**:
- Results already displayed: Table output already shown to user
- Redundant: Summary would repeat information user can already see
- Cleaner UX: Less noise, better readability

**Alternatives Considered**:
- Always show summary: Too verbose, redundant
- Show summary only if no table: More complex logic
- Let LLM decide: LLM might still repeat results

## Risks / Trade-offs

### [Risk] Content hash comparison might be slow for large prompt files
**Mitigation**: SHA256 is fast even for large files. Hash calculation happens once per startup, acceptable performance impact.

### [Risk] User might forget their choice and want to change it
**Mitigation**: Clear message when choosing "keep": "To upgrade prompt files later, delete files in ~/.aiq/prompts/ and restart aiq". User can manually delete files to trigger recreation.

### [Risk] Complete messages array might grow very large
**Mitigation**: Implement trimming logic (keep last N messages). Current limit: `DefaultHistoryLimit * 10` (200 messages). Can be adjusted if needed.

### [Risk] JSON serialization might fail for complex message structures
**Mitigation**: Multiple normalization layers ensure content is always serializable. Invalid messages are skipped rather than causing failures.

### [Risk] Backward compatibility with old Session files
**Mitigation**: System checks for `RawMessages` field first, falls back to legacy `Messages` field if not present. Both formats are maintained during transition.

### [Trade-off] More complex Session structure vs simpler but incomplete
**Chosen**: More complex structure to preserve full context. The complexity is manageable and provides significant value.

### [Trade-off] Multiple normalization passes vs single pass
**Chosen**: Multiple passes for robustness. Performance impact is minimal (only on message processing, not execution).

## Migration Plan

### Phase 1: Implementation
1. Add `RawMessages []json.RawMessage` field to Session struct
2. Implement content hash comparison in prompt loader
3. Add version-based choice storage in config
4. Update tool handler to return complete messages array
5. Update mode.go to save/load complete messages
6. Add normalization logic at multiple layers

### Phase 2: Testing
1. Test prompt upgrade detection with modified files
2. Test version-based choice persistence
3. Test complete message loading/saving
4. Test content normalization with various message types
5. Test backward compatibility with legacy Session files

### Phase 3: Deployment
1. Deploy with backward compatibility enabled
2. Monitor for any serialization issues
3. Monitor Session file sizes
4. Collect user feedback on prompt upgrade experience

### Rollback Strategy
- If issues occur, can disable `RawMessages` loading and fall back to legacy format
- Prompt upgrade detection is non-fatal (logs warning, continues)
- Can revert to always showing summaries if needed

## Open Questions

1. **Session file size limits**: Should we implement automatic compression for very large message arrays? (Currently trimmed to 200 messages)
2. **Prompt file conflict resolution**: Should we support merging user changes with new built-in prompts? (Currently: overwrite or keep, no merge)
3. **Message format evolution**: How should we handle future changes to message structure? (Currently: JSON allows flexible evolution)
