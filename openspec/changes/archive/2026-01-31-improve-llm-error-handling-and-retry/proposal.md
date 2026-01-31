## Why

Current LLM error handling lacks context awareness and intelligent retry capabilities. When operations fail due to dependencies (e.g., foreign key constraints), the LLM receives error messages but doesn't understand that subsequent operations may have resolved the original issue. For example, when dropping tables fails due to foreign key constraints, and the dependent table is later dropped, the LLM should automatically retry dropping the original tables without requiring explicit user intervention. This leads to inefficient workflows where users must manually guide the LLM through error recovery, even when the system state has changed in ways that make retry safe and appropriate.

## What Changes

- **Agent basic flow**: Establish fundamental Agent flow where LLM makes all decisions about task classification (exploratory vs definitive) and continuation (when to finish vs continue). Code layer removes all task classification logic and simply executes tools and returns all results to LLM for decision-making. This ensures consistent behavior across all tool types (SQL, commands, files, HTTP requests).
- **Enhanced error context in tool responses**: Include structured error information (error type, affected resources, dependencies) in tool error responses to help LLM understand failure causes
- **Conversation history summarization**: Provide LLM with concise summaries of recent tool executions and their outcomes, highlighting state changes that may affect retry decisions
- **Improved prompt guidance for Agent flow**: Add `<AGENT_FLOW>` section to prompts with clear guidance on task classification, tool execution result handling, and when to continue vs finish
- **Improved prompt guidance for error handling**: Update system prompts to guide LLM on analyzing errors, understanding system state changes, and making intelligent retry decisions
- **Structured error information**: Enhance error JSON responses with structured fields (error_code, error_type, affected_resources, suggested_actions) instead of just error messages
- **Context-aware retry guidance**: Add explicit instructions in prompts for LLM to analyze conversation history, identify resolved dependencies, and automatically retry previously failed operations when appropriate

## Capabilities

### New Capabilities
- `agent-basic-flow`: Fundamental Agent flow where LLM makes all decisions about task classification (exploratory vs definitive) and continuation (when to finish vs continue), while code layer simply executes tools and returns all results to LLM. This ensures consistent behavior across all tool types (SQL, commands, files, HTTP) and prevents code-level heuristics from interfering with LLM's decision-making.
- `llm-error-context-enhancement`: Enhanced error context in tool responses with structured error information (error codes, types, affected resources, dependencies) to help LLM understand failures and make intelligent retry decisions
- `conversation-history-summarization`: Summarization of recent tool executions and outcomes in conversation history, highlighting state changes that may affect retry decisions

### Modified Capabilities
- `sql-interactive-mode`: Enhance error handling requirements to include structured error information and context-aware retry guidance
- `multi-turn-conversation`: Enhance conversation history management to include tool execution summaries and state change tracking

## Impact

- **Affected code**:
  - `internal/sql/tool_handler.go`: Remove code-level exploratory/definitive classification, return all tool results to LLM, enhance error JSON responses with structured error information
  - `internal/prompt/loader.go`: Add `<AGENT_FLOW>` section to prompts with clear guidance on task classification and continuation decisions, update system prompts with error handling and retry guidance
  - `internal/sql/mode.go`: Enhance conversation history management with tool execution summaries
  - `internal/tool/sql_tool.go`: Extract structured error information from database errors
  - `internal/tool/builtin/command_tool.go`: Extract structured error information from command execution errors

- **APIs**: Tool error responses will include additional structured fields (error_code, error_type, affected_resources, suggested_actions)

- **User experience**: More intelligent error recovery, fewer manual interventions needed, smoother multi-step operations
