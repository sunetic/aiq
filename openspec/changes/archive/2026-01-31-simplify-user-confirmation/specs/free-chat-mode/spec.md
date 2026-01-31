## MODIFIED Requirements

### Requirement: Command and file operation execution confirmation

The system SHALL use risk-based confirmation for command and file operations in free chat mode, instead of requiring confirmation for all operations.

#### Scenario: Low-risk commands execute automatically
- **WHEN** LLM calls execute_command tool with risk_level="low" or command matches whitelist (ls, cat, pwd, echo, grep, etc.)
- **THEN** system executes command automatically without user confirmation
- **AND** system displays command output directly to user

#### Scenario: High-risk commands require confirmation
- **WHEN** LLM calls execute_command tool with risk_level="high" or risk_level="medium", or command does not match whitelist
- **THEN** system displays command to user
- **AND** system prompts user to confirm execution: "Execute this command? [y/N]"
- **AND** system waits for user confirmation before executing

#### Scenario: LLM provides risk level for commands
- **WHEN** LLM calls execute_command tool
- **THEN** LLM can optionally include risk_level field in arguments ("low", "medium", "high")
- **AND** system prioritizes LLM-provided risk_level over code-level rules
- **AND** if risk_level="low", execute automatically
- **AND** if risk_level="medium" or "high", require confirmation

#### Scenario: Code whitelist fallback for commands
- **WHEN** LLM calls execute_command tool without risk_level field
- **THEN** system checks command name against code whitelist (ls, cat, pwd, echo, grep, etc.)
- **AND** if command is in whitelist, execute automatically
- **AND** if command is not in whitelist, require confirmation (conservative default)

#### Scenario: Low-risk file operations execute automatically
- **WHEN** LLM calls file_operations tool with risk_level="low" or operation is read/list/exists
- **THEN** system executes file operation automatically without user confirmation
- **AND** system displays results directly to user

#### Scenario: High-risk file operations require confirmation
- **WHEN** LLM calls file_operations tool with risk_level="high" or operation is write
- **THEN** system displays file operation details to user
- **AND** system prompts user to confirm execution: "Execute this file operation? [y/N]"
- **AND** system waits for user confirmation before executing

#### Scenario: LLM provides risk level for file operations
- **WHEN** LLM calls file_operations tool
- **THEN** LLM can optionally include risk_level field in arguments ("low", "medium", "high")
- **AND** system prioritizes LLM-provided risk_level over code-level rules
- **AND** if risk_level="low", execute automatically
- **AND** if risk_level="medium" or "high", require confirmation

#### Scenario: Code whitelist fallback for file operations
- **WHEN** LLM calls file_operations tool without risk_level field
- **THEN** system checks operation type against code whitelist (read, list, exists)
- **AND** if operation is in whitelist, execute automatically
- **AND** if operation is not in whitelist (write), require confirmation

#### Scenario: HTTP request risk assessment
- **WHEN** LLM calls http_request tool
- **THEN** system checks HTTP method and LLM-provided risk_level
- **AND** if method is GET/HEAD/OPTIONS or risk_level="low", execute automatically
- **AND** if method is POST/PUT/DELETE/PATCH or risk_level="high", require confirmation
