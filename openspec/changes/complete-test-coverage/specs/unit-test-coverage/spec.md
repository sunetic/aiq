## ADDED Requirements

### Requirement: Unit test coverage for prompt management

The system SHALL have comprehensive unit tests for prompt management functionality including loading, version detection, and upgrade flows.

#### Scenario: Test prompt loader initialization
- **WHEN** prompt loader is initialized
- **THEN** system creates default prompt files if they don't exist
- **AND** system loads prompts from user directory if they exist
- **AND** system falls back to built-in prompts if user files are invalid

#### Scenario: Test prompt version detection
- **WHEN** prompt files are checked for modifications
- **THEN** system calculates SHA256 hash of file content
- **AND** system compares user file hash with built-in prompt hash
- **AND** system detects modifications when hashes differ

#### Scenario: Test version-based choice persistence
- **WHEN** user makes a choice (overwrite/keep) for prompt upgrade
- **THEN** system stores choice associated with current application version
- **AND** system retrieves stored choice on subsequent runs
- **AND** system prompts user again when application version changes

#### Scenario: Test prompt upgrade flow
- **WHEN** modified prompts are detected and user chooses overwrite
- **THEN** system overwrites user files with built-in versions
- **AND** system preserves user files when user chooses keep
- **AND** system displays instruction message when user chooses keep

### Requirement: Unit test coverage for session management

The system SHALL have comprehensive unit tests for session management including message persistence, content normalization, and backward compatibility.

#### Scenario: Test complete message array storage
- **WHEN** messages are saved to session
- **THEN** system stores complete messages array including tool calls and results
- **AND** system serializes messages as JSON (`json.RawMessage` array)
- **AND** system updates legacy `Messages []Message` field for backward compatibility

#### Scenario: Test complete message array loading
- **WHEN** session is loaded from file
- **THEN** system loads complete messages array if available
- **AND** system converts `json.RawMessage` to `[]interface{}` for use
- **AND** system falls back to legacy format if complete messages don't exist

#### Scenario: Test content normalization
- **WHEN** messages are loaded from session
- **THEN** system normalizes content field to string type
- **AND** system converts object/array content to JSON string
- **AND** system handles null content appropriately (empty string for assistant, remove for tool messages)

#### Scenario: Test backward compatibility
- **WHEN** legacy session file is loaded
- **THEN** system loads legacy `Messages []Message` format
- **AND** system converts to conversation history format
- **AND** system works normally without complete messages array

#### Scenario: Test message trimming
- **WHEN** message array exceeds limit
- **THEN** system trims oldest messages
- **AND** system preserves most recent messages up to limit

### Requirement: Unit test coverage for risk assessment

The system SHALL have comprehensive unit tests for risk assessment and confirmation mechanisms.

#### Scenario: Test LLM-provided risk level
- **WHEN** LLM calls tool with `risk_level="low"`
- **THEN** system executes tool automatically without confirmation
- **AND** system trusts LLM's risk assessment

#### Scenario: Test code-level whitelist fallback
- **WHEN** LLM calls tool without `risk_level` field
- **THEN** system checks operation against code whitelist
- **AND** system executes automatically if operation is in whitelist
- **AND** system requires confirmation if operation is not in whitelist

#### Scenario: Test risk assessment priority
- **WHEN** LLM provides `risk_level="low"` for operation not in whitelist
- **THEN** system prioritizes LLM's assessment
- **AND** system executes automatically based on LLM's risk level

#### Scenario: Test high-risk operations
- **WHEN** LLM calls tool with `risk_level="high"` or `risk_level="medium"`
- **THEN** system requires user confirmation before executing
- **AND** system displays operation details to user

### Requirement: Unit test coverage for tool implementations

The system SHALL have comprehensive unit tests for all built-in tool implementations.

#### Scenario: Test command tool execution
- **WHEN** command tool executes a command
- **THEN** system executes command with specified timeout
- **AND** system captures stdout and stderr
- **AND** system returns exit code and output
- **AND** system handles command timeout appropriately

#### Scenario: Test file tool operations
- **WHEN** file tool performs read operation
- **THEN** system reads file content successfully
- **AND** system returns file content and metadata
- **AND** system handles file not found errors

#### Scenario: Test file tool write operations
- **WHEN** file tool performs write operation
- **THEN** system requires confirmation for write operations
- **AND** system writes content to file after confirmation
- **AND** system handles permission errors

#### Scenario: Test HTTP tool requests
- **WHEN** HTTP tool makes GET request
- **THEN** system executes automatically (low risk)
- **AND** system returns response status, headers, and body
- **AND** system handles network errors

#### Scenario: Test HTTP tool POST requests
- **WHEN** HTTP tool makes POST request
- **THEN** system requires confirmation (high risk)
- **AND** system executes after user confirmation
- **AND** system sends request body correctly

#### Scenario: Test SQL tool execution
- **WHEN** SQL tool executes SELECT query
- **THEN** system executes automatically (low risk)
- **AND** system returns query results
- **AND** system handles SQL syntax errors

#### Scenario: Test SQL tool DROP operations
- **WHEN** SQL tool executes DROP statement
- **THEN** system requires confirmation (high risk)
- **AND** system executes after user confirmation
- **AND** system handles foreign key constraint errors

### Requirement: Unit test coverage for error extraction

The system SHALL have comprehensive unit tests for error extraction and handling functionality.

#### Scenario: Test structured error extraction
- **WHEN** tool execution fails
- **THEN** system extracts error code, error type, and affected resources
- **AND** system identifies dependencies from error message
- **AND** system provides suggested actions

#### Scenario: Test error parsing for different database types
- **WHEN** MySQL error occurs
- **THEN** system extracts MySQL-specific error information
- **AND** system identifies foreign key constraints correctly

#### Scenario: Test error parsing for PostgreSQL
- **WHEN** PostgreSQL error occurs
- **THEN** system extracts PostgreSQL-specific error information
- **AND** system identifies constraint violations correctly

### Requirement: Unit test coverage for configuration management

The system SHALL have comprehensive unit tests for configuration loading, validation, and directory management.

#### Scenario: Test configuration loading
- **WHEN** configuration is loaded
- **THEN** system loads from correct file path
- **AND** system validates configuration structure
- **AND** system returns error for invalid configuration

#### Scenario: Test directory structure creation
- **WHEN** directory structure is ensured
- **THEN** system creates all required subdirectories
- **AND** system handles existing directories gracefully
- **AND** system returns error for permission issues

#### Scenario: Test version choices storage
- **WHEN** version choice is stored
- **THEN** system saves choice to version choices file
- **AND** system associates choice with application version
- **AND** system retrieves choice correctly

### Requirement: Unit test coverage for Skills system

The system SHALL have comprehensive unit tests for Skills loading, matching, and parsing.

#### Scenario: Test Skills metadata loading
- **WHEN** Skills metadata is loaded
- **THEN** system loads metadata from all Skills directories
- **AND** system parses frontmatter correctly
- **AND** system handles missing or invalid Skills gracefully

#### Scenario: Test Skills matching
- **WHEN** query is matched against Skills
- **THEN** system scores Skills based on relevance
- **AND** system filters Skills by precision threshold
- **AND** system returns matched Skills ordered by priority

#### Scenario: Test Skills parser
- **WHEN** Skill file is parsed
- **THEN** system parses YAML frontmatter correctly
- **AND** system extracts metadata fields
- **AND** system handles invalid YAML gracefully

### Requirement: Unit test coverage for database operations

The system SHALL have comprehensive unit tests for database connection, query execution, and schema fetching.

#### Scenario: Test database connection
- **WHEN** database connection is established
- **THEN** system connects to database successfully
- **AND** system handles connection errors
- **AND** system validates connection parameters

#### Scenario: Test query execution
- **WHEN** SQL query is executed
- **THEN** system executes query against database
- **AND** system returns query results
- **AND** system handles query errors

#### Scenario: Test schema fetching
- **WHEN** database schema is fetched
- **THEN** system retrieves table and column information
- **AND** system formats schema information correctly
- **AND** system handles database-specific schema queries

### Requirement: Unit test coverage for chart rendering

The system SHALL have comprehensive unit tests for chart detection, configuration, and rendering logic.

#### Scenario: Test chart type detection
- **WHEN** chart type is detected from query
- **THEN** system identifies chart type (bar, line, pie, scatter)
- **AND** system extracts chart configuration from query
- **AND** system handles ambiguous queries

#### Scenario: Test chart configuration
- **WHEN** chart is configured
- **THEN** system sets axis labels correctly
- **AND** system configures colors and styling
- **AND** system handles missing configuration gracefully
