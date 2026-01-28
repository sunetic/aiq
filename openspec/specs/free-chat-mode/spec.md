## ADDED Requirements

### Requirement: Free chat mode without database source
The system SHALL support chat mode operation without requiring a database source to be selected.

#### Scenario: Enter chat mode without sources
- **WHEN** user selects chat mode and no data sources are configured
- **THEN** system enters free chat mode without requiring source selection

#### Scenario: Enter chat mode with source selection skipped
- **WHEN** user selects chat mode and chooses to skip source selection
- **THEN** system enters free chat mode

#### Scenario: Free mode capabilities
- **WHEN** system is in free chat mode
- **THEN** system supports general conversation, Skills-based operations (execute_command, http_request, file_operations), but NOT SQL execution (execute_sql tool unavailable)

#### Scenario: Free mode prompt indication
- **WHEN** system is in free chat mode
- **THEN** system displays prompt as `aiq> ` (without source name)

#### Scenario: SQL execution attempt in free mode
- **WHEN** user attempts to execute SQL in free chat mode
- **THEN** system displays error message indicating that a database source is required for SQL operations

#### Scenario: Switch from free mode to database mode
- **WHEN** user is in free chat mode and wants to use database features
- **THEN** user must exit chat mode and re-enter with source selection (no mid-session switching)

### Requirement: Optional source selection in chat mode
The system SHALL make source selection optional when entering chat mode.

#### Scenario: Source selection prompt
- **WHEN** user enters chat mode and sources are available
- **THEN** system prompts user to select a source or skip (enter free mode)

#### Scenario: Source selection from menu
- **WHEN** user selects a source from available sources
- **THEN** system enters chat mode with selected source and database connection

#### Scenario: Skip source selection
- **WHEN** user chooses to skip source selection
- **THEN** system enters free chat mode without database connection
