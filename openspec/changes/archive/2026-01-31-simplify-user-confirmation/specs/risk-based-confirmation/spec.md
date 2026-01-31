## ADDED Requirements

### Requirement: LLM-provided risk level in tool calls

The system SHALL allow LLM to provide optional `risk_level` field in tool call arguments to indicate risk assessment for the operation.

#### Scenario: LLM provides low risk level
- **WHEN** LLM calls a tool with `risk_level: "low"` in arguments
- **THEN** system executes the tool automatically without user confirmation
- **AND** system trusts LLM's risk assessment for this operation

#### Scenario: LLM provides high risk level
- **WHEN** LLM calls a tool with `risk_level: "high"` or `risk_level: "medium"` in arguments
- **THEN** system requires user confirmation before executing the tool
- **AND** system displays the operation details and asks user to confirm

#### Scenario: LLM does not provide risk level
- **WHEN** LLM calls a tool without `risk_level` field in arguments
- **THEN** system falls back to code-level risk assessment
- **AND** system checks code whitelist for common safe operations
- **AND** if operation is in whitelist, execute automatically
- **AND** if operation is not in whitelist, require confirmation (conservative default)

#### Scenario: Tool definitions include risk_level parameter
- **WHEN** system loads tool definitions for LLM
- **THEN** all tool definitions include optional `risk_level` parameter in their parameters schema
- **AND** `risk_level` is defined as enum: ["low", "medium", "high"]
- **AND** description explains that "low" = execute automatically, "medium"/"high" = require confirmation

### Requirement: Code-level whitelist fallback

The system SHALL maintain code-level whitelist of common safe operations as fallback when LLM does not provide risk_level.

#### Scenario: SQL whitelist fallback
- **WHEN** LLM calls execute_sql without risk_level
- **THEN** system checks SQL statement against whitelist (SELECT, SHOW, DESCRIBE, EXPLAIN)
- **AND** if SQL matches whitelist pattern, execute automatically
- **AND** if SQL does not match whitelist, require confirmation

#### Scenario: Command whitelist fallback
- **WHEN** LLM calls execute_command without risk_level
- **THEN** system checks command name against whitelist (ls, cat, pwd, echo, grep, etc.)
- **AND** if command is in whitelist, execute automatically
- **AND** if command is not in whitelist, require confirmation

#### Scenario: File operation whitelist fallback
- **WHEN** LLM calls file_operations without risk_level
- **THEN** system checks operation type against whitelist (read, list, exists)
- **AND** if operation is in whitelist, execute automatically
- **AND** if operation is not in whitelist (write), require confirmation

#### Scenario: HTTP request whitelist fallback
- **WHEN** LLM calls http_request without risk_level
- **THEN** system checks HTTP method against whitelist (GET, HEAD, OPTIONS)
- **AND** if method is in whitelist, execute automatically
- **AND** if method is not in whitelist (POST, PUT, DELETE, PATCH), require confirmation

### Requirement: Risk assessment priority

The system SHALL prioritize LLM-provided risk_level over code-level rules, with conservative default for unknown operations.

#### Scenario: LLM risk_level takes priority
- **WHEN** LLM provides risk_level="low" for an operation
- **THEN** system executes automatically even if operation is not in code whitelist
- **AND** system trusts LLM's assessment

#### Scenario: Conservative default for unknown operations
- **WHEN** LLM does not provide risk_level and operation is not in code whitelist
- **THEN** system requires user confirmation by default
- **AND** system does not attempt to enumerate all dangerous operations

#### Scenario: LLM can handle unknown commands intelligently
- **WHEN** LLM encounters unknown command (e.g., init, reboot, custom script)
- **THEN** LLM can assess risk and set risk_level="high" in tool call
- **AND** system will require confirmation based on LLM's assessment
- **AND** no need to enumerate all dangerous commands in code

### Requirement: LLM uncertainty handling

The system SHALL allow LLM to handle uncertain operations by either setting risk_level="high" or returning text to ask user.

#### Scenario: LLM sets high risk for uncertain operation
- **WHEN** LLM is uncertain about operation risk (e.g., init, reboot)
- **THEN** LLM can call tool with risk_level="high"
- **AND** system will require confirmation before executing

#### Scenario: LLM asks user before calling tool
- **WHEN** LLM is uncertain about operation risk
- **THEN** LLM can return text asking user: "This operation (init system) may be risky. Should I proceed?"
- **AND** system displays LLM's question to user
- **AND** user can confirm or deny
- **AND** if user confirms, LLM can then call the tool
- **AND** this is allowed as exception to "must call tools" rule for uncertain operations

### Requirement: Extensible risk assessor architecture

The system SHALL provide extensible risk assessor interface that allows easy addition of risk assessment for new tools.

#### Scenario: New tool can implement risk assessor
- **WHEN** a new tool is added to the system
- **THEN** tool can add risk_level parameter to its definition
- **AND** tool can implement risk assessor with whitelist fallback
- **AND** risk assessment follows same priority: LLM risk_level → code whitelist → require confirmation

#### Scenario: Risk assessor interface
- **WHEN** system needs to assess risk for a tool call
- **THEN** system uses RiskAssessor interface
- **AND** each tool type can have its own risk assessor implementation
- **AND** risk assessor checks LLM-provided risk_level first, then code whitelist, then defaults to requiring confirmation
