# Design: Refactor Chart Display to LLM-Based Decision Making

## Context

Currently, the system uses code-level string matching to detect view switching commands (e.g., "display as pie chart") in `internal/sql/mode.go` (lines 525-569). This approach:

1. **Violates AI-first architecture**: Decision-making should be handled by LLM, not hardcoded patterns
2. **Fails on typos**: Users with typos (e.g., "disply" instead of "display") cannot use the feature
3. **Bypasses LLM context**: The LLM never sees the request, missing opportunity to understand intent from conversation history
4. **Maintenance burden**: Requires maintaining keyword lists and pattern matching logic

The current flow:
- User query → String matching check → If matched, direct rendering → Otherwise, LLM processing
- Query results are simplified for LLM (only status/row_count, not actual data)
- Query results stored in `lastQueryResult` but not included in conversation history

## Goals / Non-Goals

**Goals:**
- Remove all code-level view switching detection logic
- Ensure LLM receives full context including query results in conversation history
- Let LLM decide when to call `render_chart` or `render_table` tools based on user intent
- Maintain backward compatibility: users can still request charts naturally, LLM will understand
- Improve robustness: handle typos, variations, and natural language better

**Non-Goals:**
- We are NOT adding fuzzy matching or spell correction at code level (that's still hardcoding)
- We are NOT maintaining keyword lists as fallback
- We are NOT optimizing for speed by bypassing LLM (AI-first means LLM handles decisions)

## Decisions

### Decision 1: Include Query Results in Conversation History

**Decision**: When a SQL query executes successfully, include a formatted summary of the query result in the assistant's response message that gets added to conversation history.

**Rationale**: 
- LLM needs to see what data is available to make informed decisions
- Current implementation simplifies results too much (only status/row_count)
- LLM can understand "show this as a chart" if it knows what "this" refers to

**Alternatives Considered**:
- **Option A**: Include full result data in conversation history
  - ❌ Rejected: Too verbose, increases token usage significantly
- **Option B**: Include structured summary (columns, sample rows, row count)
  - ✅ Selected: Provides enough context without excessive tokens
- **Option C**: Keep current simplified approach
  - ❌ Rejected: LLM doesn't know what data is available

**Implementation**:
- After `execute_sql` succeeds, format result summary: columns, row count, and first 2-3 sample rows
- Add this summary to assistant message in conversation history
- Format: "Query executed successfully. Returned X rows with columns: [col1, col2, ...]. Sample data: ..."

### Decision 2: Remove View Switching Detection Code

**Decision**: Delete the entire view switching detection block (lines 525-569 in `mode.go`) and the `detectChartTypeFromQuery` function.

**Rationale**:
- Aligns with AI-first architecture: LLM should interpret user intent
- Eliminates maintenance burden of keyword lists
- Handles typos and variations naturally through LLM understanding

**Alternatives Considered**:
- **Option A**: Keep detection as fallback, use LLM as primary
  - ❌ Rejected: Still violates AI-first principle, adds complexity
- **Option B**: Improve detection with fuzzy matching
  - ❌ Rejected: Still hardcoding, doesn't solve fundamental issue
- **Option C**: Complete removal
  - ✅ Selected: Clean, aligns with architecture

**Implementation**:
- Remove `isViewSwitch` detection logic
- Remove `detectChartTypeFromQuery` function
- Remove the entire `if isViewSwitch && lastQueryResult != nil` block
- All requests now go through LLM processing

### Decision 3: Enhance Tool Descriptions for LLM Guidance

**Decision**: Update `render_chart` and `render_table` tool descriptions to explicitly guide LLM on when and how to use them with existing query results.

**Rationale**:
- LLM needs clear guidance on tool usage patterns
- Current descriptions don't emphasize using existing results
- Better descriptions reduce hallucination and improve decision quality

**Alternatives Considered**:
- **Option A**: Rely on LLM to figure it out from context
  - ❌ Rejected: May lead to unnecessary SQL queries when results already available
- **Option B**: Add explicit instructions in system prompt
  - ✅ Selected: Clear, maintainable, guides LLM behavior
- **Option C**: Both tool descriptions and system prompt
  - ✅ Selected: Redundant guidance reduces errors

**Implementation**:
- Update `render_chart` description: "Format query results as a chart. **IMPORTANT**: If recent query results are available in conversation history, use that data directly. Only generate new SQL if user explicitly requests different data."
- Update `render_table` description similarly
- Add guidance in system prompt about prioritizing existing results

### Decision 4: Keep Query Result Display Logic

**Decision**: Continue displaying query results immediately after execution (table format), but also include summary in conversation history.

**Rationale**:
- Users expect immediate feedback when query executes
- Displaying results doesn't conflict with LLM-based chart rendering
- Summary in history helps LLM understand context for follow-up requests

**Alternatives Considered**:
- **Option A**: Don't display immediately, let LLM decide
  - ❌ Rejected: Poor UX, users expect immediate results
- **Option B**: Display and include in history (current approach)
  - ✅ Selected: Best UX, provides context for LLM

**Implementation**:
- Keep current display logic in `tool_handler.go` (lines 895-903)
- Add result summary to conversation history after display
- LLM sees summary in history, can reference it for chart requests

## Risks / Trade-offs

### Risk 1: LLM May Generate Unnecessary SQL Queries
**Risk**: LLM might not recognize that query results are available in conversation history and generate new SQL queries when user asks for chart.

**Mitigation**:
- Enhance tool descriptions to emphasize using existing results
- Add explicit guidance in system prompt
- Include clear result summaries in conversation history
- Monitor and iterate on prompt engineering if needed

**Trade-off**: Acceptable - LLM may occasionally generate unnecessary queries, but this is better than hardcoded pattern matching that fails on typos.

### Risk 2: Slight Latency Increase
**Risk**: All chart requests now go through LLM, adding ~100-500ms latency compared to direct code path.

**Mitigation**:
- LLM latency is acceptable for better UX (handles typos, variations)
- Most users won't notice the difference
- Can optimize LLM calls if needed (caching, batching)

**Trade-off**: Acceptable - Better robustness and alignment with architecture is worth small latency cost.

### Risk 3: LLM May Misinterpret User Intent
**Risk**: LLM might misunderstand when user wants chart vs. new query.

**Mitigation**:
- Clear tool descriptions guide LLM behavior
- Result summaries in history provide context
- System prompt emphasizes checking conversation history first
- If hallucinations occur frequently, can add minimal code-level guardrails (only as last resort)

**Trade-off**: Acceptable - LLM understanding is generally better than pattern matching. Can add minimal guards if needed.

### Risk 4: Token Usage Increase
**Risk**: Including result summaries in conversation history increases token usage per conversation.

**Mitigation**:
- Summaries are concise (columns, row count, 2-3 sample rows)
- Context compression already implemented can handle this
- Trade-off is worth it for better LLM understanding

**Trade-off**: Acceptable - Small token increase is worth better functionality.

## Migration Plan

### Phase 1: Implementation
1. Update `tool_handler.go` to include result summaries in conversation history
2. Remove view switching detection code from `mode.go`
3. Remove `detectChartTypeFromQuery` function
4. Update tool descriptions in `llm_functions.go`
5. Enhance system prompt in `tool_handler.go`

### Phase 2: Testing
1. Test with various user inputs (with/without typos)
2. Test with natural language variations
3. Verify LLM correctly uses existing results vs. generating new queries
4. Monitor for any regression in chart rendering functionality

### Phase 3: Rollout
- No breaking changes for end users (same natural language works)
- Users may notice better handling of typos/variations
- If issues arise, can rollback by reverting commits

### Rollback Strategy
- Keep old code in git history
- If critical issues arise, can temporarily restore view switching detection
- But prefer fixing LLM prompts/descriptions over rollback

## Open Questions

1. **Result Summary Format**: What's the optimal format for result summaries in conversation history?
   - **Decision Needed**: Columns + row count + sample rows format
   - **Resolution**: Use concise format: "Query returned X rows. Columns: [col1, col2]. Sample: row1, row2"

2. **System Prompt Wording**: How explicit should we be about using existing results?
   - **Decision Needed**: Balance between clarity and verbosity
   - **Resolution**: Add clear but concise guidance in system prompt

3. **Error Handling**: What if LLM generates SQL when it should use existing results?
   - **Decision Needed**: Accept as-is or add minimal guardrails?
   - **Resolution**: Monitor first, add minimal guards only if hallucinations are frequent

4. **Performance**: Should we optimize LLM calls for chart requests?
   - **Decision Needed**: Is current latency acceptable?
   - **Resolution**: Monitor user feedback, optimize if needed
