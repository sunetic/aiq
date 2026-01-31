## MODIFIED Requirements

### Requirement: Handle query errors
The system SHALL handle query errors with structured error information and context-aware retry guidance.

#### Scenario: Handle query errors
- **WHEN** query execution fails
- **THEN** system displays clear error message and allows user to retry or modify
- **AND** system includes structured error information (error_code, error_type, affected_resources, dependencies) in error response
- **AND** system provides tool execution summary highlighting recent state changes that may affect retry decisions

#### Scenario: LLM receives structured error information
- **WHEN** query execution fails
- **THEN** LLM receives error response with structured fields (error_code, error_type, affected_resources, dependencies, suggested_actions)
- **AND** LLM can analyze error structure to understand failure cause and retry feasibility

#### Scenario: LLM receives tool execution summary
- **WHEN** LLM processes user query after tool execution failures
- **THEN** LLM receives summary of recent tool executions highlighting successes, failures, and state changes
- **AND** LLM can identify when dependencies have been resolved and retry is appropriate

#### Scenario: LLM makes intelligent retry decisions
- **WHEN** LLM receives error indicating dependency issue (e.g., foreign key constraint)
- **AND** tool execution summary shows dependency has been resolved (e.g., dependent table deleted)
- **THEN** LLM automatically retries the failed operation without requiring explicit user intervention
