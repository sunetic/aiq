## ADDED Requirements

### Requirement: Structured error information in tool responses
The system SHALL include structured error information in tool error responses to help LLM understand failure causes and make intelligent retry decisions.

#### Scenario: Include error code in error response
- **WHEN** a tool execution fails and error code is available (e.g., database error code)
- **THEN** system includes `error_code` field in error JSON response with the error code value

#### Scenario: Include error type in error response
- **WHEN** a tool execution fails
- **THEN** system includes `error_type` field in error JSON response with categorized error type (e.g., "foreign_key_constraint", "syntax_error", "permission_denied")

#### Scenario: Include affected resources in error response
- **WHEN** a tool execution fails and error message mentions specific resources (tables, columns, etc.)
- **THEN** system includes `affected_resources` field in error JSON response as an array of resource names

#### Scenario: Include dependencies in error response
- **WHEN** a tool execution fails due to dependencies (e.g., foreign key constraints)
- **THEN** system includes `dependencies` field in error JSON response as an array of dependency resource names

#### Scenario: Include suggested actions in error response
- **WHEN** a tool execution fails and error suggests possible fixes
- **THEN** system includes `suggested_actions` field in error JSON response as an array of suggested action descriptions

#### Scenario: Maintain backward compatibility
- **WHEN** a tool execution fails
- **THEN** system includes original `error` field with error message string
- **AND** structured fields are additive (do not replace existing fields)

#### Scenario: Handle missing error information gracefully
- **WHEN** error information cannot be extracted (e.g., unknown error format)
- **THEN** system includes only `status: "error"` and `error` fields with original error message
- **AND** structured fields are omitted (not included as empty arrays)

### Requirement: Error information extraction
The system SHALL extract structured error information from error messages using pattern matching.

#### Scenario: Extract foreign key constraint information
- **WHEN** error message contains foreign key constraint information (e.g., "Cannot drop table 'X' referenced by foreign key constraint 'Y' on table 'Z'")
- **THEN** system extracts affected resource (table X), dependency (table Z), and constraint name (Y)
- **AND** sets error_type to "foreign_key_constraint"

#### Scenario: Extract syntax error information
- **WHEN** error message indicates SQL syntax error
- **THEN** system sets error_type to "syntax_error"
- **AND** extracts affected resource as the SQL query or query portion mentioned in error

#### Scenario: Extract permission error information
- **WHEN** error message indicates permission denied
- **THEN** system sets error_type to "permission_denied"
- **AND** extracts affected resource from error message if mentioned

#### Scenario: Extract resource not found information
- **WHEN** error message indicates table or column does not exist
- **THEN** system sets error_type to "resource_not_found"
- **AND** extracts affected resource (table/column name) from error message

#### Scenario: Extract resource exists information
- **WHEN** error message indicates table or column already exists
- **THEN** system sets error_type to "resource_exists"
- **AND** extracts affected resource (table/column name) from error message

#### Scenario: Use database-agnostic patterns
- **WHEN** extracting error information
- **THEN** system uses pattern matching that works across MySQL, PostgreSQL, and SeekDB
- **AND** does not rely on database-specific error code mappings

### Requirement: Error type categorization
The system SHALL categorize errors into standard types to help LLM understand error nature and retry feasibility.

#### Scenario: Categorize foreign key constraint errors
- **WHEN** error is related to foreign key constraints
- **THEN** system categorizes as "foreign_key_constraint" error type

#### Scenario: Categorize syntax errors
- **WHEN** error is related to SQL syntax issues
- **THEN** system categorizes as "syntax_error" error type

#### Scenario: Categorize permission errors
- **WHEN** error is related to insufficient privileges
- **THEN** system categorizes as "permission_denied" error type

#### Scenario: Categorize connection errors
- **WHEN** error is related to database connection issues
- **THEN** system categorizes as "connection_error" error type

#### Scenario: Categorize timeout errors
- **WHEN** error is related to operation timeout
- **THEN** system categorizes as "timeout" error type

#### Scenario: Categorize unknown errors
- **WHEN** error does not match any known pattern
- **THEN** system categorizes as "unknown" error type
- **AND** includes original error message for LLM analysis
