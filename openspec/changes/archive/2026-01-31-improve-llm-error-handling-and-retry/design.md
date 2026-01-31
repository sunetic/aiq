## Context

### Current State

**Error Handling:**
- Tool errors are returned as simple JSON with `status: "error"` and `error: "<message>"`
- Error messages are plain text strings extracted from database/command errors
- LLM receives full conversation history but no structured summaries of recent tool executions
- No mechanism to highlight state changes that might affect retry decisions

**Example Problem:**
When dropping tables fails due to foreign key constraints:
1. `DROP TABLE customers` fails: "Cannot drop table 'customers' referenced by foreign key constraint 'sales_OBFK_...' on table 'sales'"
2. `DROP TABLE products` fails: "Cannot drop table 'products' referenced by foreign key constraint 'sales_OBFK_...' on table 'sales'"
3. `DROP TABLE sales` succeeds
4. LLM doesn't understand that `customers` and `products` can now be dropped, requires user to explicitly ask again

**Conversation History:**
- Full message history is sent to LLM (up to 20 message pairs)
- No summarization or highlighting of tool execution outcomes
- LLM must parse through entire history to understand state changes

**System Prompts:**
- Basic error handling guidance exists but doesn't emphasize analyzing state changes
- No explicit instructions for intelligent retry based on conversation context

### Constraints

- Must maintain backward compatibility with existing tool response format
- Cannot break existing error handling flows
- Must work with all database types (MySQL, PostgreSQL, SeekDB)
- Should not significantly increase token usage (summaries must be concise)
- Must follow AI-first architecture principles (LLM makes decisions, not hardcoded logic)

## Goals / Non-Goals

**Goals:**
- Enhance error responses with structured information (error codes, types, affected resources, dependencies)
- Provide concise summaries of recent tool executions highlighting state changes
- Guide LLM to analyze errors and conversation history for intelligent retry decisions
- Enable LLM to automatically retry operations when dependencies are resolved
- Improve error recovery without requiring explicit user intervention

**Non-Goals:**
- Hardcoded retry logic (all retry decisions must be LLM-based)
- Automatic retry without LLM decision (violates AI-first principle)
- Complex state tracking system (keep it simple, rely on LLM understanding)
- Breaking changes to existing tool response format (add fields, don't remove)
- Database-specific error parsing (use generic patterns that work across databases)

## Decisions

### 1. Structured Error Information: Add Fields to Error JSON

**Decision**: Enhance error JSON responses with structured fields while maintaining backward compatibility.

**Format:**
```json
{
  "status": "error",
  "error": "<original error message>",
  "error_code": "<database error code if available>",
  "error_type": "<categorized error type>",
  "affected_resources": ["<resource1>", "<resource2>"],
  "dependencies": ["<dependency1>", "<dependency2>"],
  "suggested_actions": ["<action1>", "<action2>"]
}
```

**Rationale:**
- Structured information helps LLM understand error context without parsing text
- Backward compatible: existing `error` field remains, new fields are additive
- Error type categorization helps LLM make retry decisions (e.g., "foreign_key_constraint" vs "syntax_error")
- Dependencies field explicitly lists what must be resolved before retry

**Error Type Categories:**
- `foreign_key_constraint`: Table/row referenced by foreign key
- `syntax_error`: SQL syntax issues
- `permission_denied`: Insufficient privileges
- `resource_not_found`: Table/column doesn't exist
- `resource_exists`: Table/column already exists
- `connection_error`: Database connection issues
- `timeout`: Operation timeout
- `unknown`: Unclassified errors

**Alternative Considered:**
- **Separate error response format**: Rejected - breaks backward compatibility, requires more changes

### 2. Error Information Extraction: Pattern-Based Parsing

**Decision**: Extract structured information from error messages using pattern matching, not database-specific parsing.

**Approach:**
- Parse common error patterns (foreign key constraints, syntax errors, etc.)
- Extract resource names (tables, columns) from error messages
- Identify dependencies from constraint error messages
- Use regex patterns that work across MySQL, PostgreSQL, and SeekDB

**Rationale:**
- Database-agnostic approach works with all supported databases
- Pattern matching is simpler than database-specific error code mapping
- LLM can still use original error message if extraction fails

**Example Patterns:**
- Foreign key: `Cannot drop table 'X' referenced by foreign key constraint 'Y' on table 'Z'`
  - Extract: affected_resource="X", dependency="Z", constraint="Y"
- Syntax error: `You have an error in your SQL syntax near 'X'`
  - Extract: error_type="syntax_error", affected_resource="<query>"

**Alternative Considered:**
- **Database-specific error code mapping**: Rejected - requires maintaining mappings for each database, more complex

### 3. Conversation History Summarization: Tool Execution Summary

**Decision**: Generate concise summaries of recent tool executions (last 3-5 tool calls) highlighting successes, failures, and state changes.

**Format:**
Add a summary section before user query in conversation:
```
<TOOL_EXECUTION_SUMMARY>
Recent tool executions:
- execute_sql: DROP TABLE sales → SUCCESS (table deleted)
- execute_sql: DROP TABLE customers → FAILED (foreign key constraint from 'sales')
- execute_sql: DROP TABLE products → FAILED (foreign key constraint from 'sales')

State changes: Table 'sales' has been deleted, foreign key constraints removed.
</TOOL_EXECUTION_SUMMARY>
```

**Rationale:**
- Highlights recent state changes that affect retry decisions
- Concise (3-5 recent executions) to avoid token bloat
- Explicitly calls out state changes (dependencies resolved, resources created/deleted)
- Helps LLM quickly understand what has changed since last error

**Implementation:**
- Track last N tool executions in `HandleToolCallLoop`
- Generate summary before sending to LLM
- Include in system message or as separate message before user query

**Alternative Considered:**
- **Full conversation history analysis**: Rejected - too verbose, increases token usage significantly
- **No summarization, rely on full history**: Rejected - LLM struggles to identify relevant state changes in long histories

### 4. Prompt Enhancement: Error Analysis and Retry Guidance

**Decision**: Add explicit instructions in system prompts for error analysis and intelligent retry.

**New Prompt Sections:**
```
<ERROR_HANDLING>
When a tool execution fails:
1. Analyze the error structure (error_type, dependencies, affected_resources)
2. Review recent tool execution summary to identify state changes
3. If dependencies mentioned in error have been resolved (e.g., dependent table deleted), automatically retry the failed operation
4. If error suggests a fix (e.g., drop constraint first), propose and execute the fix
5. Only ask user for guidance if error cannot be resolved automatically

Example: If dropping table 'customers' failed due to foreign key from 'sales', and 'sales' was later dropped successfully, automatically retry dropping 'customers'.
</ERROR_HANDLING>
```

**Rationale:**
- Explicit guidance helps LLM understand when retry is appropriate
- Emphasizes analyzing structured error information and state changes
- Provides concrete example of intelligent retry scenario
- Maintains AI-first approach (LLM decides, not hardcoded logic)

**Placement:**
- Add to `database-base.md` prompt (database mode)
- Add to `common.md` prompt (applies to all modes)

**Alternative Considered:**
- **Hardcoded retry logic**: Rejected - violates AI-first architecture principle

### 5. Summary Generation: Lightweight State Tracking

**Decision**: Track tool execution outcomes in memory during `HandleToolCallLoop`, generate summary on-the-fly.

**Approach:**
- Maintain a slice of recent tool executions (last 5) in `ToolHandler`
- Each entry: `{tool: "execute_sql", args: {...}, status: "success|error", error_info: {...}, state_changes: [...]}`
- Generate summary string before each LLM call
- Reset summary when new user query starts (fresh context)

**Rationale:**
- Simple in-memory tracking, no persistent storage needed
- Lightweight, doesn't add significant overhead
- Summary regenerated each time, always current
- No need to parse full conversation history

**State Change Detection:**
- DDL operations (CREATE/DROP/ALTER) → state change
- Successful operations → may resolve dependencies
- Failed operations → dependencies still exist

**Alternative Considered:**
- **Parse conversation history**: Rejected - more complex, requires parsing message format
- **Persistent state tracking**: Rejected - unnecessary complexity, in-memory is sufficient

## Risks / Trade-offs

### Risk 1: Token Usage Increase
**Risk**: Summaries and structured error fields increase token usage per LLM call.

**Mitigation**: 
- Keep summaries concise (3-5 recent executions)
- Only include essential information in structured fields
- Monitor token usage and adjust summary length if needed

### Risk 2: LLM May Over-Retry
**Risk**: LLM might retry operations inappropriately based on error analysis.

**Mitigation**:
- Clear guidance in prompts: "only retry if dependencies are resolved"
- Structured error information helps LLM understand when retry is safe
- User can always interrupt or correct LLM behavior

### Risk 3: Error Pattern Matching May Fail
**Risk**: Pattern-based error extraction might miss edge cases or database-specific formats.

**Mitigation**:
- Fallback to original error message if extraction fails
- LLM can still parse original error message
- Start with common patterns, expand based on real-world errors

### Risk 4: Summary May Miss Important Context
**Risk**: Summarizing only last 3-5 executions might miss relevant earlier state changes.

**Mitigation**:
- Full conversation history still available to LLM
- Summary is additive, not replacement
- Can adjust summary length based on feedback

### Trade-off: Simplicity vs. Completeness
**Trade-off**: Simple pattern matching vs. comprehensive error parsing.

**Decision**: Start with simple patterns, expand based on real-world needs. Better to have working simple solution than complex incomplete one.

## Migration Plan

### Phase 1: Structured Error Information
1. Add error extraction functions to `internal/tool/sql_tool.go` and `internal/tool/builtin/command_tool.go`
2. Update `internal/sql/tool_handler.go` to use structured error format
3. Test with various error types (foreign key, syntax, permission, etc.)

### Phase 2: Conversation History Summarization
1. Add tool execution tracking to `ToolHandler` struct
2. Implement summary generation function
3. Integrate summary into `HandleToolCallLoop` before LLM calls
4. Test with multi-step operations (drop tables, create dependencies, etc.)

### Phase 3: Prompt Enhancement
1. Update `internal/prompt/loader.go` to add error handling sections
2. Update `database-base.md` and `common.md` prompt templates
3. Test LLM behavior with new prompts

### Phase 4: Testing and Refinement
1. Test with real-world scenarios (foreign key constraints, multi-step operations)
2. Monitor LLM retry behavior
3. Refine error patterns and summary format based on feedback
4. Adjust prompt guidance if needed

### Rollback Strategy
- All changes are additive (new fields, new summaries, new prompt sections)
- If issues arise, can disable summarization or revert prompt changes
- Structured error fields are optional, LLM can ignore them

## Open Questions

1. **Summary Length**: Should summary include last 3, 5, or 7 tool executions? Start with 5, adjust based on testing.

2. **Error Pattern Coverage**: Which error patterns are most common? Start with foreign key constraints and syntax errors, expand based on real-world usage.

3. **Retry Limits**: Should there be explicit retry limits in prompts? Or rely on LLM's judgment? Start without limits, add if needed.

4. **State Change Detection**: How to detect state changes reliably? Start with DDL operation detection, refine based on testing.
