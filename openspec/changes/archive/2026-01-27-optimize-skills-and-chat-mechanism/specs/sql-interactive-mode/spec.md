## MODIFIED Requirements

### Requirement: SQL interactive mode entry
The system SHALL enter SQL interactive mode when user selects `chat` from main menu, with optional source selection.

#### Scenario: Enter chat mode without sources configured
- **WHEN** user selects chat and no data sources are configured
- **THEN** system enters free chat mode without requiring source selection

#### Scenario: Enter chat mode with source selection prompt
- **WHEN** user selects chat and sources are available
- **THEN** system prompts user to select a source or skip to enter free mode

#### Scenario: Enter chat mode with selected source
- **WHEN** user selects a source from available sources
- **THEN** system enters chat mode with selected source and database connection

#### Scenario: Enter chat mode in free mode
- **WHEN** user chooses to skip source selection
- **THEN** system enters free chat mode without database connection

### Requirement: Natural language to SQL translation
The system SHALL translate user's natural language questions into SQL queries using LLM, with Skills-enhanced prompts, supporting both database mode and free mode.

#### Scenario: Submit natural language query in database mode
- **WHEN** user enters a natural language question in chat mode with database source
- **THEN** system sends question to LLM API with database schema context, matched Skills content, and available tools (including execute_sql)

#### Scenario: Submit natural language query in free mode
- **WHEN** user enters a natural language question in free chat mode
- **THEN** system sends question to LLM API with matched Skills content and available tools (excluding execute_sql)

#### Scenario: Match Skills to query
- **WHEN** user submits a natural language query
- **THEN** system matches query against Skills metadata and loads relevant Skills content (filtered by minimum score threshold)

#### Scenario: Build prompt with Skills
- **WHEN** system prepares prompt for LLM translation
- **THEN** system includes matched Skills content in system prompt section, ordered by priority

#### Scenario: Manage prompt length
- **WHEN** prompt token count exceeds thresholds
- **THEN** system compresses conversation history and evicts low-priority Skills to stay within token limits

#### Scenario: Receive SQL translation
- **WHEN** LLM returns SQL query
- **THEN** system displays translated SQL query to user

#### Scenario: Confirm SQL execution
- **WHEN** SQL is translated
- **THEN** system prompts user to confirm execution or modify query

### Requirement: SQL mode interface
The system SHALL provide an interactive interface for SQL queries with prompt and command handling, displaying current source information.

#### Scenario: Display SQL prompt with source
- **WHEN** SQL mode is active with a database source
- **THEN** system displays prompt as `aiq[source-name]> ` indicating active source

#### Scenario: Display prompt in free mode
- **WHEN** SQL mode is active without database source (free mode)
- **THEN** system displays prompt as `aiq> ` (without source name)

#### Scenario: Accept multi-line input
- **WHEN** user enters SQL query or general query
- **THEN** system accepts multi-line input until user submits (e.g., Ctrl+D or special command)

#### Scenario: Exit SQL mode
- **WHEN** user types `exit` or `back` in chat mode
- **THEN** system returns to main menu

### Requirement: Database schema context
The system SHALL provide database schema information to LLM for accurate SQL generation when a database source is selected.

#### Scenario: Fetch schema on source selection
- **WHEN** user selects a data source in chat mode
- **THEN** system optionally fetches schema information (tables, columns) for context

#### Scenario: Include schema in LLM request
- **WHEN** translating natural language to SQL in database mode
- **THEN** system includes relevant schema information in LLM API request

#### Scenario: No schema in free mode
- **WHEN** system is in free chat mode
- **THEN** system does not include database schema information in LLM requests
