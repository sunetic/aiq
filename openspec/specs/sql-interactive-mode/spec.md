## ADDED Requirements

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

## MODIFIED Requirements

### Requirement: SQL interactive mode entry
The system SHALL enter SQL interactive mode when user selects `chat` from main menu, with optional source selection, or directly when MySQL or PostgreSQL CLI connection arguments are provided.

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

#### Scenario: Direct entry from MySQL CLI args
- **WHEN** user provides MySQL CLI connection arguments
- **THEN** system validates connection, creates source automatically (or reuses existing), and directly enters chat mode with the source (skipping source selection menu)

#### Scenario: Direct entry from PostgreSQL CLI args
- **WHEN** user provides PostgreSQL CLI connection arguments
- **THEN** system validates connection, creates source automatically (or reuses existing), and directly enters chat mode with the source (skipping source selection menu)

#### Scenario: LLM config check before direct entry
- **WHEN** user provides MySQL or PostgreSQL CLI args but LLM is not configured
- **THEN** system prompts user to configure LLM first before entering chat mode

### Requirement: Natural language to SQL translation
The system SHALL translate user's natural language questions into SQL queries using LLM.

#### Scenario: Submit natural language query
- **WHEN** user enters a natural language question in SQL mode
- **THEN** system sends question to LLM API with database schema context

#### Scenario: Receive SQL translation
- **WHEN** LLM returns SQL query (in tool call)
- **THEN** system checks risk level before displaying or executing
- **AND** if low-risk, execute automatically without displaying SQL first
- **AND** if high-risk or unknown, display SQL and ask for confirmation

#### Scenario: Confirm SQL execution
- **WHEN** SQL is translated and requires confirmation (high-risk)
- **THEN** system prompts user to confirm execution or modify query

### Requirement: SQL query execution
The system SHALL execute SQL queries against the selected database connection using risk-based confirmation.

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

#### Scenario: Execute confirmed query
- **WHEN** user confirms SQL execution (for high-risk operations)
- **THEN** system executes query against selected database
- **AND** system displays results after execution

#### Scenario: Display query results
- **WHEN** query executes successfully
- **THEN** system displays results in formatted table with syntax highlighting

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

### Requirement: SQL mode interface
The system SHALL provide an interactive interface for SQL queries with prompt, command handling, and tab completion support.

#### Scenario: Display SQL prompt with source
- **WHEN** SQL mode is active with a database source
- **THEN** system displays prompt as `aiq[source-name]> ` indicating active source

#### Scenario: Display SQL prompt with source and database
- **WHEN** SQL mode is active with a database source and database name is available (from source config or `-D` override)
- **THEN** system displays prompt as `aiq[source-name@database-name]> ` indicating active source and database being used

#### Scenario: Display prompt in free mode
- **WHEN** SQL mode is active without database source (free mode)
- **THEN** system displays prompt as `aiq> ` (without source name)

#### Scenario: Accept multi-line input
- **WHEN** user enters SQL query or general query
- **THEN** system accepts multi-line input until user submits (e.g., Ctrl+D or special command)

#### Scenario: Exit SQL mode with /exit command
- **WHEN** user types `/exit` in chat mode
- **THEN** system saves session and returns to main menu

#### Scenario: Display help with /help command
- **WHEN** user types `/help` in chat mode
- **THEN** system displays list of available commands and their descriptions

#### Scenario: Help command shows available commands
- **WHEN** user types `/help`
- **THEN** system displays:
  - `/exit` - Exit chat mode and return to main menu
  - `/help` - Show this help message
  - `/history` - View conversation history
  - `/clear` - Clear conversation history

#### Scenario: Tab completion for commands
- **WHEN** user types `/` and presses Tab
- **THEN** system displays available commands: `/exit`, `/help`, `/history`, `/clear`

#### Scenario: Tab completion completes command name
- **WHEN** user types `/ex` and presses Tab
- **THEN** system completes to `/exit`

#### Scenario: Tab completion only for commands
- **WHEN** user types natural language query (not starting with `/`) and presses Tab
- **THEN** system does not provide completion (to avoid interfering with natural language input)

#### Scenario: Return to main menu on exit
- **WHEN** user exits chat mode (via `/exit` or Ctrl+D)
- **THEN** system returns to main menu instead of exiting the application

### Requirement: SQL query result display
The system SHALL display SQL query results in mysql client standard table format.

#### Scenario: Display results in mysql client format
- **WHEN** SQL query returns data
- **THEN** system displays results in ASCII table format using `+`, `-`, `|` borders (mysql client style)

#### Scenario: Display row count after results
- **WHEN** SQL query executes successfully with results
- **THEN** system displays "N row(s) in set" after the table

#### Scenario: No LLM processing of displayed results
- **WHEN** SQL query results are displayed to user
- **THEN** LLM does NOT repeat, list, or summarize the already-displayed data

### Requirement: Query results in conversation history
The system SHALL include formatted query result summaries in conversation history so LLM can understand what data is available for visualization requests.

#### Scenario: Include result summary after successful query
- **WHEN** SQL query executes successfully and returns results
- **THEN** system adds formatted result summary to assistant message in conversation history, including columns, row count, and sample rows (2-3 rows)

#### Scenario: Result summary format
- **WHEN** query result is added to conversation history
- **THEN** summary format includes: "Query executed successfully. Returned X rows with columns: [col1, col2, ...]. Sample data: [first 2-3 rows]"

#### Scenario: LLM sees query results in context
- **WHEN** user requests chart visualization after executing a query
- **THEN** LLM receives conversation history including previous query result summary and can reference it for chart rendering

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

## MODIFIED Requirements

### Requirement: Natural language to SQL translation
The system SHALL translate user's natural language questions into SQL queries using LLM, with Skills-enhanced prompts, supporting both database mode and free mode, with enhanced mode awareness and tool usage understanding.

#### Scenario: Submit natural language query in database mode
- **WHEN** user enters a natural language question in chat mode with database source
- **THEN** system sends question to LLM API with database schema context, matched Skills content, and available tools (including execute_sql), with explicit guidance that execute_sql tool should be used for database queries

#### Scenario: Submit natural language query in free mode
- **WHEN** user enters a natural language question in free chat mode
- **THEN** system sends question to LLM API with matched Skills content and available tools (excluding execute_sql), with explicit guidance that database queries are invalid in free mode

#### Scenario: Match Skills to query
- **WHEN** user submits a natural language query
- **THEN** system matches query against Skills metadata using LLM semantic matching with precision filtering and loads relevant Skills content

#### Scenario: Build prompt with Skills
- **WHEN** system prepares prompt for LLM translation
- **THEN** system includes matched Skills content in system prompt section, ordered by priority

#### Scenario: Manage prompt length
- **WHEN** prompt token count exceeds thresholds
- **THEN** system uses LLM semantic compression for conversation history and evicts low-priority Skills to stay within token limits

#### Scenario: Receive SQL translation
- **WHEN** LLM returns SQL query
- **THEN** system displays translated SQL query to user

#### Scenario: Confirm SQL execution
- **WHEN** SQL is translated
- **THEN** system prompts user to confirm execution or modify query

#### Scenario: LLM understands tool availability
- **WHEN** system is in database mode and user asks database query
- **THEN** LLM uses execute_sql tool, NOT execute_command with mysql/psql commands

#### Scenario: LLM validates query appropriateness
- **WHEN** user submits query in free mode that requires database
- **THEN** LLM recognizes query as invalid for current mode and asks clarifying question instead of attempting tool execution

#### Scenario: LLM distinguishes SQL tool from shell commands
- **WHEN** user asks database query in database mode
- **THEN** LLM uses execute_sql tool for SQL queries, and only uses execute_command for system operations (not database queries)
