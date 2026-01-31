## Why

Currently, every SQL query and command execution requires user confirmation, which creates friction in the workflow. For low-risk operations like SELECT queries, SHOW commands, or read-only file operations, this confirmation step is unnecessary and slows down productivity. Users want a smarter confirmation system that only prompts for potentially dangerous operations.

## What Changes

- **Tool-level risk assessment system**: Implement risk assessment at tool execution layer (not chat/user input layer) where tool parameters are explicit and unambiguous
- **Extensible risk assessor architecture**: Design a pluggable risk assessment system where each tool type can have its own risk assessor
- **Low-risk operations execute automatically**: Safe operations (SELECT, SHOW, DESCRIBE, ls, read, GET requests, etc.) execute without confirmation
- **High-risk operations require confirmation**: Dangerous operations (DROP, TRUNCATE, DELETE, rm, write, POST/DELETE requests, etc.) still require explicit user confirmation
- **LLM-assisted risk assessment**: 
  - LLM provides optional `risk_level` field in tool call arguments ("low", "medium", "high")
  - Code layer prioritizes LLM's risk_level assessment
  - Code layer maintains basic whitelist for common safe operations (fallback when LLM doesn't provide risk_level)
  - Unknown operations default to requiring confirmation (conservative safety-first approach)
  - LLM can intelligently assess risk for unknown commands (init, reboot, custom scripts) without code-level enumeration
  - If LLM is uncertain, it can either set risk_level="high" or return text asking user for confirmation
- **Configurable safety levels**: Allow users to configure confirmation behavior (strict mode vs. smart mode)

## Capabilities

### New Capabilities
- `risk-based-confirmation`: Tool-level risk assessment system that determines whether tool executions require user confirmation based on tool type and explicit tool parameters. Risk assessment happens at tool execution layer (not chat/user input layer) where parameters are structured and unambiguous.

### Modified Capabilities
- `sql-interactive-mode`: Modify confirmation requirements to use risk-based assessment instead of requiring confirmation for all operations
- `free-chat-mode`: Apply risk-based confirmation to command and file operations in free mode

## Impact

- **Affected code**:
  - `internal/tool/risk_assessor.go`: New module providing extensible risk assessment interface that prioritizes LLM-provided risk_level, with code-based whitelist fallback
  - `internal/sql/tool_handler.go`: Add risk assessment call before tool execution, check LLM-provided risk_level first, then code whitelist, default to requiring confirmation
  - `internal/tool/llm_functions.go`: Add optional `risk_level` parameter to all tool definitions
  - `internal/tool/builtin/command_tool.go`: Add command risk assessor with code whitelist (common safe commands) as fallback
  - `internal/tool/builtin/file_tool.go`: Add file operation risk assessor with code whitelist (read/list/exists) as fallback
  - `internal/tool/builtin/http_tool.go`: Add HTTP request risk assessor with code whitelist (GET/HEAD/OPTIONS) as fallback
  - `internal/prompt/loader.go`: Add guidance to prompts explaining how to set risk_level in tool calls, and when to ask user for confirmation
  - Future tools: Each new tool can add risk_level parameter and implement risk assessor with whitelist fallback

- **Architecture**: 
  - Risk assessment happens at tool execution layer (where parameters are explicit)
  - No risk assessment at chat/user input layer (where SQL text vs natural language is ambiguous)
  - Three-tier risk assessment (priority order):
    1. **LLM-provided risk_level**: If LLM sets risk_level="low" → execute automatically; if "medium"/"high" → require confirmation
    2. **Code whitelist** (fallback): If LLM doesn't provide risk_level, check code whitelist (SELECT, SHOW, ls, read, GET) → execute automatically
    3. **Unknown operations**: If not in whitelist and no LLM risk_level → require confirmation by default (conservative safety-first)
  - LLM role: LLM intelligently assesses risk and provides risk_level in tool calls. Can handle unknown commands (init, reboot) without code enumeration.
  - Code layer role: Provides whitelist fallback and enforces conservative default for unknown operations
  - Extensible design allows easy addition of risk assessors for new tools

- **User experience**: Faster workflow for safe operations, maintained safety for dangerous operations
