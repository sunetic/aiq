## ADDED Requirements

### Requirement: LLM-based chart display decision
The system SHALL use LLM to interpret user requests for chart visualization based on conversation context, including available query results, and decide when to call chart rendering tools.

#### Scenario: LLM interprets chart request from conversation context
- **WHEN** user requests chart visualization (e.g., "display as pie chart", "show as bar chart")
- **THEN** LLM receives full conversation history including previous query results and interprets user intent

#### Scenario: LLM uses existing query results for chart rendering
- **WHEN** user requests chart visualization and previous query results are available in conversation history
- **THEN** LLM calls render_chart tool with existing result data instead of generating new SQL query

#### Scenario: LLM determines chart type from user input
- **WHEN** user specifies chart type (e.g., "pie chart", "bar chart", "line chart")
- **THEN** LLM extracts chart type from user input and passes it to render_chart tool

#### Scenario: LLM handles typos and variations
- **WHEN** user requests chart with typos or natural language variations (e.g., "disply as pie chart", "show me a chart")
- **THEN** LLM correctly interprets intent and calls appropriate chart rendering tool

#### Scenario: LLM decides when to use existing results vs. new query
- **WHEN** user requests chart visualization
- **THEN** LLM checks conversation history for available query results and uses them if appropriate, or generates new SQL query if user explicitly requests different data
