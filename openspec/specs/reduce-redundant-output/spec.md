## MODIFIED Requirements

### Requirement: Tool result simplification

The system SHALL simplify tool results when results are already displayed to avoid redundant output.

#### Scenario: SQL query with table output
- **WHEN** execute_sql tool executes successfully and displays table output
- **THEN** system simplifies tool result sent to LLM:
  - Removes `result_summary` field (not needed, causes redundancy)
  - Sets `displayed: true` flag
  - Includes clear instruction: "CRITICAL: Results are already displayed to the user in table format. Do NOT repeat the results in your response. Return finish_reason='stop' with empty content (no text output). The user can see the results above."
- **AND** LLM receives instruction to return empty content

#### Scenario: LLM response after displayed results
- **WHEN** LLM receives tool result with `displayed: true`
- **THEN** LLM returns `finish_reason="stop"` with empty content
- **AND** system doesn't display any additional text
- **AND** user only sees table output (no redundant summary)

### Requirement: Summary display logic

The system SHALL skip displaying summary when results are already shown.

#### Scenario: Results already displayed
- **WHEN** `finalResponse` is empty and `queryResult` exists
- **THEN** system doesn't display `formatQueryResultSummary` output
- **AND** system recognizes that results were already displayed (e.g., table format)
- **AND** user doesn't see redundant "Query executed successfully..." message

#### Scenario: LLM provides text response
- **WHEN** `finalResponse` is not empty
- **THEN** system displays LLM's text response
- **AND** system may append result summary if it adds value (not redundant)

### Requirement: Prompt instruction strengthening

The system SHALL strengthen prompt instructions to prevent LLM from repeating displayed results.

#### Scenario: Prompt update
- **WHEN** prompt includes guidance about tool results
- **THEN** prompt explicitly states:
  - If tool result contains `"displayed": true`, do NOT repeat results
  - Return `finish_reason="stop"` with empty content immediately
  - Do NOT include result summaries or descriptions
- **AND** prompt provides clear examples of correct vs incorrect behavior
