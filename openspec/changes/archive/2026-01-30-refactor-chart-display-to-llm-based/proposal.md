# Proposal: Refactor Chart Display to LLM-Based Decision Making

## Why

The current implementation uses code-level string matching to detect view switching commands (e.g., "display as pie chart"), which violates the core architecture principle of this AI project. As an AI-first application, decision-making should be primarily handled by the LLM based on conversation context, not through hardcoded pattern matching. The current approach causes failures when users have typos (e.g., "disply" instead of "display") and creates a maintenance burden as we need to maintain keyword lists. More importantly, it bypasses the LLM's ability to understand user intent from conversation context, including previous query results that are already available in the conversation history.

## What Changes

- **Remove code-level view switching detection**: Eliminate the string matching logic in `internal/sql/mode.go` (lines 525-569) that detects view switching commands
- **Ensure query results are in conversation history**: Modify conversation history building to include formatted query results, so LLM can see what data is available
- **Let LLM decide tool usage**: Rely on LLM to interpret user intent and call `render_chart` or `render_table` tools when appropriate, based on conversation context
- **Enhance tool descriptions**: Update `render_chart` and `render_table` tool descriptions to guide LLM on when to use them with existing query results
- **Remove `detectChartTypeFromQuery` function**: No longer needed as LLM will determine chart type from user input

**BREAKING**: The direct view switching shortcut (bypassing LLM) will be removed. All chart display requests will go through LLM decision-making.

## Capabilities

### New Capabilities
- `llm-based-chart-display`: LLM interprets user requests for chart visualization based on conversation context, including available query results, and decides when to call chart rendering tools

### Modified Capabilities
- `sql-interactive-mode`: Modified to remove code-level view switching detection and ensure query results are properly included in conversation history for LLM context
- `chart-visualization`: Modified to rely on LLM tool calling instead of code-level command detection. LLM determines chart type from user input rather than code extraction

## Impact

**Affected Code:**
- `internal/sql/mode.go`: Remove view switching detection logic (lines 525-569), remove `detectChartTypeFromQuery` function, modify conversation history building to include query results
- `internal/sql/tool_handler.go`: Enhance system prompt to guide LLM on using existing query results for visualization
- `internal/tool/llm_functions.go`: Update `render_chart` and `render_table` tool descriptions to emphasize using existing query results

**Behavior Changes:**
- Users can no longer rely on exact keyword matching for view switching
- All chart display requests will go through LLM, which may add slight latency but provides better understanding of user intent
- Better handling of typos and variations in user input
- LLM can make more intelligent decisions about when to use existing results vs. generating new queries

**Benefits:**
- Aligns with AI-first architecture principles
- Better handling of typos and natural language variations
- LLM can understand context better (e.g., "show this as a chart" referring to previous result)
- Reduced maintenance burden (no keyword lists to maintain)
- More flexible and extensible (LLM can handle new patterns without code changes)
