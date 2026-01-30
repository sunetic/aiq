## 1. Add Query Results to Conversation History

- [x] 1.1 Create helper function to format query result summary (columns, row count, sample rows)
- [x] 1.2 Modify `tool_handler.go` to include result summary in assistant message after successful SQL execution
- [x] 1.3 Ensure result summary is added to conversation history when query executes successfully

## 2. Remove View Switching Detection Code

- [x] 2.1 Remove `isViewSwitch` detection logic from `mode.go` (lines 527-531)
- [x] 2.2 Remove the entire `if isViewSwitch && lastQueryResult != nil` block (lines 534-569)
- [x] 2.3 Remove `detectChartTypeFromQuery` function from `mode.go` (lines 716-734)
- [x] 2.4 Remove `lastQueryResult` variable usage for view switching (keep for other purposes if needed)

## 3. Enhance Tool Descriptions

- [x] 3.1 Update `render_chart` tool description in `llm_functions.go` to emphasize using existing query results
- [x] 3.2 Update `render_table` tool description similarly
- [x] 3.3 Ensure tool descriptions guide LLM to check conversation history first

## 4. Enhance System Prompt

- [x] 4.1 Add guidance in system prompt about prioritizing existing query results for visualization requests
- [x] 4.2 Ensure system prompt instructs LLM to check conversation history before generating new SQL queries

## 5. Testing and Verification

- [ ] 5.1 Test chart rendering with natural language requests (e.g., "display as pie chart")
- [ ] 5.2 Test with typos (e.g., "disply as pie chart")
- [ ] 5.3 Verify LLM uses existing query results instead of generating new SQL
- [ ] 5.4 Verify chart rendering still works correctly
