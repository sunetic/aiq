## MODIFIED Requirements

### Requirement: Chart rendering option
The system SHALL allow users to request chart visualization through natural language, and LLM SHALL interpret these requests and call appropriate chart rendering tools based on conversation context.

#### Scenario: User requests chart visualization
- **WHEN** user requests chart visualization using natural language (e.g., "display as pie chart", "show as bar chart", "render as line chart")
- **THEN** LLM interprets the request, checks conversation history for available query results, and calls render_chart tool with appropriate data

#### Scenario: LLM uses existing query results
- **WHEN** user requests chart visualization and previous query results are available in conversation history
- **THEN** LLM calls render_chart tool with existing result data instead of generating new SQL query

#### Scenario: LLM determines chart type from user input
- **WHEN** user specifies chart type in request (e.g., "pie chart", "bar chart")
- **THEN** LLM extracts chart type from user input and passes it to render_chart tool

#### Scenario: Handle typos and variations
- **WHEN** user requests chart with typos or natural language variations (e.g., "disply as pie chart", "show me a chart")
- **THEN** LLM correctly interprets intent and calls appropriate chart rendering tool

#### Scenario: Skip chart for empty results
- **WHEN** query executes successfully but returns no rows
- **THEN** system only displays table view (no chart option)

### Requirement: Chart rendering implementation
The system SHALL render charts using ASCII/Unicode characters in terminal, triggered by LLM tool calls.

#### Scenario: Render bar chart via LLM tool call
- **WHEN** LLM calls render_chart tool with bar chart type
- **THEN** system renders vertical or horizontal bar chart with proper scaling and labels

#### Scenario: Render line chart via LLM tool call
- **WHEN** LLM calls render_chart tool with line chart type
- **THEN** system renders line chart with data points and connecting lines

#### Scenario: Render pie chart via LLM tool call
- **WHEN** LLM calls render_chart tool with pie chart type
- **THEN** system renders pie chart showing proportional distribution

#### Scenario: Render scatter plot via LLM tool call
- **WHEN** LLM calls render_chart tool with scatter plot type
- **THEN** system renders scatter plot with data points

## REMOVED Requirements

### Requirement: Chart rendering option (prompt-based)
**Reason**: Replaced by LLM-based decision making. System no longer prompts user to choose view type after query execution. Instead, users request charts naturally and LLM interprets and handles the request.

**Migration**: Users can request charts at any time using natural language (e.g., "display as pie chart"). No need to wait for prompt after query execution. LLM will handle the request based on available query results in conversation history.

### Requirement: Chart display integration (prompt-based)
**Reason**: Removed prompt-based chart selection workflow. Chart display is now handled entirely through LLM tool calling based on user's natural language requests.

**Migration**: Users can request charts naturally after seeing query results. LLM will interpret the request and call appropriate tools. No breaking changes - same functionality, different implementation approach.
