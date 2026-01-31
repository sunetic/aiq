## MODIFIED Requirements

### Requirement: SQL query execution confirmation

The system SHALL use risk-based confirmation instead of requiring confirmation for all SQL operations.

#### Scenario: Low-risk SQL executes automatically
- **WHEN** LLM calls execute_sql tool with risk_level="low" or SQL matches whitelist (SELECT, SHOW, DESCRIBE, EXPLAIN)
- **THEN** system executes SQL automatically without user confirmation
- **AND** system displays results directly to user

#### Scenario: High-risk SQL requires confirmation
- **WHEN** LLM calls execute_sql tool with risk_level="high" or risk_level="medium", or SQL does not match whitelist
- **THEN** system displays SQL query to user
- **AND** system prompts user to confirm execution: "Execute this query? [y/N]"
- **AND** system waits for user confirmation before executing

#### Scenario: LLM provides risk level in tool call
- **WHEN** LLM calls execute_sql tool
- **THEN** LLM can optionally include risk_level field in arguments ("low", "medium", "high")
- **AND** system prioritizes LLM-provided risk_level over code-level rules
- **AND** if risk_level="low", execute automatically
- **AND** if risk_level="medium" or "high", require confirmation

#### Scenario: Code whitelist fallback for SQL
- **WHEN** LLM calls execute_sql tool without risk_level field
- **THEN** system checks SQL statement against code whitelist (SELECT, SHOW, DESCRIBE, EXPLAIN)
- **AND** if SQL matches whitelist pattern, execute automatically
- **AND** if SQL does not match whitelist, require confirmation (conservative default)

#### Scenario: Receive SQL translation
- **WHEN** LLM returns SQL query (in tool call)
- **THEN** system checks risk level before displaying or executing
- **AND** if low-risk, execute automatically without displaying SQL first
- **AND** if high-risk or unknown, display SQL and ask for confirmation

#### Scenario: Execute confirmed query
- **WHEN** user confirms SQL execution (for high-risk operations)
- **THEN** system executes query against selected database
- **AND** system displays results after execution
