## ADDED Requirements

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

## REMOVED Requirements

### Requirement: Code-level view switching detection
**Reason**: Violates AI-first architecture principle. Decision-making should be handled by LLM based on conversation context, not hardcoded pattern matching. Removed to align with project's core design philosophy of using LLM + tools for capabilities.

**Migration**: Users can still request charts using natural language (e.g., "display as pie chart", "show as bar chart"). LLM will interpret these requests based on conversation context. No user-facing changes - same natural language works, but now handled by LLM instead of code matching.
