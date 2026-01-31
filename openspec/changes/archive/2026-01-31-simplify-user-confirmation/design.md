## Context

Currently, every tool execution (SQL queries, commands, file operations) requires user confirmation, which creates friction in the workflow. Users want a smarter confirmation system that only prompts for potentially dangerous operations while automatically executing safe operations.

The system uses tool calling where LLM calls tools with structured parameters. Tool parameters are explicit and unambiguous at the tool execution layer, making it suitable for risk assessment. However, at the chat/user input layer, SQL text and natural language are ambiguous, so risk assessment should NOT happen there.

## Goals / Non-Goals

**Goals:**
- Implement risk-based confirmation system at tool execution layer
- Allow LLM to provide risk assessment via optional `risk_level` field in tool calls
- Maintain code-level whitelist as fallback for common safe operations
- Support all existing tools (execute_sql, execute_command, file_operations, http_request) and future tools
- Provide conservative default: unknown operations require confirmation

**Non-Goals:**
- Risk assessment at chat/user input layer (where SQL text vs natural language is ambiguous)
- Exhaustive enumeration of all dangerous commands/operations
- LLM-based risk assessment that requires separate API calls (would add latency)
- Changing the fundamental tool calling flow or Agent architecture

## Decisions

### Decision 1: LLM-provided risk_level in tool arguments

**Decision**: LLM provides optional `risk_level` field ("low", "medium", "high") in tool call arguments.

**Rationale**:
- LLM can intelligently assess risk for unknown commands (init, reboot, custom scripts) without code enumeration
- Tool parameters are structured and unambiguous, suitable for risk assessment
- Optional field allows backward compatibility (if LLM doesn't provide, fallback to code rules)
- Single tool call, no additional latency

**Alternatives considered**:
- Separate `assess_risk` tool: Rejected - adds latency and complexity
- Code-level enumeration only: Rejected - cannot cover all dangerous commands
- LLM returns text asking user: Considered but less efficient than risk_level field

### Decision 2: Three-tier risk assessment priority

**Decision**: Priority order: (1) LLM-provided risk_level, (2) Code whitelist, (3) Conservative default (require confirmation).

**Rationale**:
- LLM's assessment takes priority (most intelligent)
- Code whitelist provides fast fallback for common operations
- Conservative default ensures safety for unknown operations
- No need to enumerate all dangerous operations

**Alternatives considered**:
- Code blacklist for dangerous operations: Rejected - cannot enumerate all dangerous commands
- Always require LLM risk_level: Rejected - adds friction, LLM may not always provide

### Decision 3: Code whitelist for common safe operations

**Decision**: Maintain code-level whitelist for common safe operations as fallback.

**Rationale**:
- Provides fast path for common operations (SELECT, SHOW, ls, read, GET)
- Works even if LLM doesn't provide risk_level
- Minimal maintenance (only safe operations, not dangerous ones)
- Conservative: only whitelist operations that are clearly safe

**Whitelist examples**:
- SQL: SELECT, SHOW, DESCRIBE, EXPLAIN
- Commands: ls, cat, pwd, echo, grep (read-only)
- File operations: read, list, exists
- HTTP: GET, HEAD, OPTIONS

### Decision 4: Extensible risk assessor architecture

**Decision**: Create `RiskAssessor` interface that each tool type can implement.

**Rationale**:
- Allows easy addition of risk assessment for new tools
- Each tool type can have its own risk assessment logic
- Follows same priority: LLM risk_level → code whitelist → require confirmation
- Maintains separation of concerns

**Interface design**:
```go
type RiskAssessor interface {
    AssessRisk(toolName string, args map[string]interface{}) RiskLevel
}

type RiskLevel int
const (
    RiskLow RiskLevel = iota  // Execute automatically
    RiskHigh                  // Require confirmation
)
```

### Decision 5: LLM uncertainty handling

**Decision**: Allow LLM to handle uncertain operations by either (1) setting risk_level="high" or (2) returning text to ask user.

**Rationale**:
- Provides flexibility for LLM to handle edge cases
- Option 1 (risk_level="high") is more efficient (single tool call)
- Option 2 (return text) allows LLM to ask clarifying questions
- Both options maintain safety (require confirmation or user input)

**Implementation**:
- Option 1: LLM sets risk_level="high" → system requires confirmation
- Option 2: LLM returns text asking user → system displays question, user confirms, LLM calls tool

## Risks / Trade-offs

**Risk**: LLM may incorrectly assess risk (mark dangerous operation as "low")
- **Mitigation**: Code whitelist only includes clearly safe operations. Unknown operations default to requiring confirmation. User can always cancel.

**Risk**: LLM may not always provide risk_level field
- **Mitigation**: Code whitelist fallback ensures common operations still work efficiently. Unknown operations require confirmation (conservative).

**Risk**: Adding risk_level parameter to all tools increases complexity
- **Mitigation**: Parameter is optional, backward compatible. Tools that don't need it can ignore it.

**Risk**: Code whitelist may need maintenance as new safe operations are identified
- **Mitigation**: Whitelist is minimal (only clearly safe operations). LLM can handle most cases via risk_level.

**Trade-off**: LLM-provided risk_level vs code rules
- **Chosen**: LLM takes priority (more intelligent), code whitelist as fallback (faster, safer)
- **Reason**: Balances intelligence (LLM) with safety (code rules)

**Trade-off**: Automatic execution vs always requiring confirmation
- **Chosen**: Low-risk operations execute automatically, high-risk require confirmation
- **Reason**: Improves user experience while maintaining safety

## Migration Plan

1. **Phase 1**: Add risk_level parameter to all tool definitions
   - Update `internal/tool/llm_functions.go` to add optional risk_level parameter
   - Update prompt to guide LLM on setting risk_level

2. **Phase 2**: Implement risk assessor interface
   - Create `internal/tool/risk_assessor.go` with RiskAssessor interface
   - Implement risk assessors for each tool type (SQL, command, file, HTTP)

3. **Phase 3**: Update tool execution flow
   - Modify `internal/sql/tool_handler.go` to call risk assessor before execution
   - Skip confirmation for low-risk operations
   - Require confirmation for high-risk and unknown operations

4. **Phase 4**: Add prompt guidance
   - Update `internal/prompt/loader.go` to include risk assessment guidance
   - Explain when to set risk_level="low" vs "high"
   - Explain when to ask user vs set risk_level="high"

5. **Phase 5**: Testing and refinement
   - Test with various operations (safe, dangerous, unknown)
   - Refine whitelist based on usage patterns
   - Monitor LLM's risk_level accuracy

**Rollback**: If issues arise, can temporarily disable automatic execution and require confirmation for all operations.

## Open Questions

1. Should there be a configuration option to disable automatic execution (strict mode)?
   - **Decision needed**: Yes, allow users to configure confirmation behavior

2. Should risk_level="medium" be treated differently from "high"?
   - **Decision needed**: Treat both as requiring confirmation (simpler, safer)

3. How to handle user cancellation of high-risk operations?
   - **Decision needed**: Return cancelled status to LLM, let LLM decide next action

4. Should there be a "yesAll" mode for batch operations?
   - **Decision**: No, not needed. The risk-based confirmation system is sufficient - low-risk operations execute automatically, high-risk operations require confirmation. This provides good balance without needing batch confirmation modes.
