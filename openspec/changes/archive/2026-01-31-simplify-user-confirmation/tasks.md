## 1. Risk Assessor Interface and Core Module

- [x] 1.1 Create `internal/tool/risk_assessor.go` with RiskAssessor interface:
  - Define `RiskLevel` type (RiskLow, RiskHigh)
  - Define `RiskAssessor` interface with `AssessRisk(toolName string, args map[string]interface{}) RiskLevel` method
  - Implement priority logic: LLM risk_level → code whitelist → require confirmation

- [x] 1.2 Create SQL risk assessor:
  - Check for LLM-provided risk_level in args
  - Implement whitelist check for SQL statements (SELECT, SHOW, DESCRIBE, EXPLAIN)
  - Return RiskLow for whitelisted operations, RiskHigh for others

- [x] 1.3 Create command risk assessor:
  - Check for LLM-provided risk_level in args
  - Implement whitelist check for command names (ls, cat, pwd, echo, grep, etc.)
  - Extract first command from command string (handle env vars)
  - Return RiskLow for whitelisted commands, RiskHigh for others

- [x] 1.4 Create file operation risk assessor:
  - Check for LLM-provided risk_level in args
  - Check operation type (read, list, exists vs write)
  - Return RiskLow for read/list/exists, RiskHigh for write

- [x] 1.5 Create HTTP request risk assessor:
  - Check for LLM-provided risk_level in args
  - Check HTTP method (GET, HEAD, OPTIONS vs POST, PUT, DELETE, PATCH)
  - Return RiskLow for GET/HEAD/OPTIONS, RiskHigh for others

## 2. Tool Definition Updates

- [x] 2.1 Update `internal/tool/llm_functions.go`:
  - Add optional `risk_level` parameter to execute_sql tool definition
  - Add optional `risk_level` parameter to execute_command tool definition
  - Add optional `risk_level` parameter to file_operations tool definition
  - Add optional `risk_level` parameter to http_request tool definition
  - Parameter should be enum: ["low", "medium", "high"] with description

- [x] 2.2 Update tool definitions in `internal/tool/builtin/`:
  - Ensure all built-in tool definitions include risk_level parameter
  - Update GetDefinition() methods to include risk_level in parameters schema

## 3. Tool Execution Flow Updates

- [x] 3.1 Update `internal/sql/tool_handler.go`:
  - Add risk assessment call before tool execution in HandleToolCallLoop
  - For execute_sql: Check risk level, skip confirmation if RiskLow, require confirmation if RiskHigh
  - Extract risk_level from tool call arguments
  - Display SQL and ask confirmation only for high-risk operations

- [x] 3.2 Update command execution flow:
  - Add risk assessment for execute_command tool calls
  - Skip confirmation for low-risk commands
  - Require confirmation for high-risk commands

- [x] 3.3 Update file operation execution flow:
  - Add risk assessment for file_operations tool calls
  - Skip confirmation for low-risk operations (read, list, exists)
  - Require confirmation for high-risk operations (write)

- [x] 3.4 Update HTTP request execution flow:
  - Add risk assessment for http_request tool calls
  - Skip confirmation for low-risk methods (GET, HEAD, OPTIONS)
  - Require confirmation for high-risk methods (POST, PUT, DELETE, PATCH)

## 4. Prompt Guidance Updates

- [x] 4.1 Update `internal/prompt/loader.go`:
  - Add `<RISK_ASSESSMENT>` section to database mode prompt
  - Add `<RISK_ASSESSMENT>` section to free mode prompt
  - Explain when to set risk_level="low" vs "high"
  - Provide examples of low-risk vs high-risk operations
  - Explain that LLM can return text to ask user if uncertain

- [x] 4.2 Add risk assessment guidance:
  - SQL operations: SELECT/SHOW = low, DROP/TRUNCATE = high
  - Commands: ls/cat/pwd = low, rm/sudo = high
  - File operations: read/list = low, write = high
  - HTTP requests: GET = low, POST/DELETE = high
  - Unknown operations: set risk_level="high" or ask user

## 5. LLM Uncertainty Handling

- [ ] 5.1 Update tool_handler.go to handle LLM text responses for uncertain operations:
  - Allow LLM to return text asking user for confirmation (exception to "must call tools" rule)
  - Detect risk confirmation questions in LLM text responses
  - Display question to user and wait for confirmation
  - After user confirms, allow LLM to continue with tool call

- [ ] 5.2 Update prompt to explain uncertainty handling:
  - LLM can set risk_level="high" for uncertain operations
  - LLM can return text asking user: "This operation may be risky. Should I proceed?"
  - Both approaches are acceptable

## 6. Testing and Validation

- [ ] 6.1 Test low-risk operations execute automatically:
  - Test SELECT, SHOW, DESCRIBE SQL queries
  - Test ls, cat, pwd commands
  - Test file read operations
  - Test GET HTTP requests

- [ ] 6.2 Test high-risk operations require confirmation:
  - Test DROP, TRUNCATE SQL queries
  - Test rm, sudo commands
  - Test file write operations
  - Test POST/DELETE HTTP requests

- [ ] 6.3 Test LLM-provided risk_level:
  - Test LLM sets risk_level="low" for safe operations
  - Test LLM sets risk_level="high" for dangerous operations
  - Test LLM does not provide risk_level (fallback to whitelist)

- [ ] 6.4 Test unknown operations:
  - Test init, reboot commands (not in whitelist, require confirmation)
  - Test custom scripts (not in whitelist, require confirmation)
  - Test LLM can set risk_level="high" for unknown operations

- [ ] 6.5 Test LLM uncertainty handling:
  - Test LLM returns text asking user for confirmation
  - Test user confirms, LLM continues with tool call
  - Test user denies, operation is cancelled

## 7. Documentation and Cleanup

- [ ] 7.1 Add code comments explaining risk assessment logic
- [ ] 7.2 Add code comments explaining risk_level parameter usage
- [ ] 7.3 Verify all tests pass
- [ ] 7.4 Code review and cleanup
