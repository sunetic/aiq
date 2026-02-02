package tool

import (
	"github.com/aiq/aiq/internal/db"
	"github.com/aiq/aiq/internal/llm"
	"github.com/aiq/aiq/internal/tool/builtin"
)

// GetLLMFunctions converts tool definitions to LLM Function format
func GetLLMFunctions() []llm.Function {
	return GetLLMFunctionsWithBuiltin(nil)
}

// GetLLMFunctionsWithBuiltin returns LLM functions including built-in tools
// If dbConn is nil (free mode), execute_sql tool is excluded
func GetLLMFunctionsWithBuiltin(dbConn *db.Connection) []llm.Function {
	tools := []llm.Function{}

	// Only include execute_sql if database connection exists
	if dbConn != nil {
		tools = append(tools, llm.Function{
			Name:        "execute_sql",
			Description: "**MANDATORY TOOL CALL**: Execute a SQL query against the database and return the results. Available ONLY in database mode when a database source is selected. **CRITICAL**: When the user requests database operations (SELECT, INSERT, UPDATE, DELETE, CREATE, DROP, ALTER, SHOW, etc.), you MUST call this tool. Do NOT describe what you will do in text - actually call the tool. Do NOT say 'I will execute' or 'Stand by while I execute' - just call the tool directly.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sql": map[string]interface{}{
						"type":        "string",
						"description": "The SQL query to execute",
					},
					"risk_level": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"low", "medium", "high"},
						"description": "Optional: Risk level assessment for this operation. 'low' = safe to execute automatically (e.g., SELECT, SHOW), 'medium'/'high' = requires user confirmation (e.g., DROP, TRUNCATE). If not provided, system will assess risk conservatively.",
					},
					"task_type": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"definitive", "exploratory"},
						"description": "Optional: Task type classification. 'definitive' = task is clear and complete, 'exploratory' = task requires information gathering or multi-step process. If not provided, system will infer from context.",
					},
					"output_mode": map[string]interface{}{
						"type":        "string",
						"enum":        []string{"full", "streaming"},
						"description": "**REQUIRED**: Output display mode. Classify tool call as process-oriented or result-oriented: 'full' = result-oriented (tool output IS the final goal, e.g., 'show tables'), 'streaming' = process-oriented (tool output is intermediate step, e.g., 'analyze sales trends'). **MANDATORY**: Always explicitly set this parameter.",
					},
				},
				"required": []string{"sql"},
			},
		})
	}

	// Add render_table and render_chart (available in both modes)
	tools = append(tools, llm.Function{
		Name:        "render_table",
		Description: "Format query results as a table string. Use this when you want to show data in a tabular format. **IMPORTANT**: If recent query results are available in conversation history, use that data directly. Only generate new SQL queries if the user explicitly requests different data.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"columns": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Column names",
				},
				"rows": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"description": "Row data, each row is an array of string values",
				},
			},
			"required": []string{"columns", "rows"},
		},
	})

	tools = append(tools, llm.Function{
		Name:        "render_chart",
		Description: "**MANDATORY TOOL CALL**: When the user requests chart visualization (pie chart, bar chart, line chart, etc.), you MUST call this tool. Do NOT return text descriptions or JSON data. The chart will be automatically displayed in the terminal. **CRITICAL**: Check conversation history for recent query results first. Extract columns and rows from the result_summary or previous execute_sql results. Only generate new SQL queries if the user explicitly requests different data or no recent results are available.",
		Parameters: map[string]interface{}{
			"type": "object",
			"properties": map[string]interface{}{
				"columns": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "string"},
					"description": "Column names from query results (e.g., [\"category\", \"total_revenue\"])",
				},
				"rows": map[string]interface{}{
					"type":        "array",
					"items":       map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}},
					"description": "Row data from query results, each row is an array of string values (e.g., [[\"Appliances\", \"159.98\"], [\"Electronics\", \"2699.95\"]])",
				},
				"chart_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"bar", "line", "pie", "scatter"},
					"description": "Type of chart: 'pie' for pie charts, 'bar' for bar charts, 'line' for line charts, 'scatter' for scatter plots",
				},
			},
			"required": []string{"columns", "rows", "chart_type"},
		},
	})

	// Add built-in tools
	builtinDefs := builtin.GetBuiltinToolDefinitions(dbConn)
	for _, bt := range builtinDefs {
		if fn, ok := bt["function"].(map[string]interface{}); ok {
			tools = append(tools, llm.Function{
				Name:        fn["name"].(string),
				Description: fn["description"].(string),
				Parameters:  fn["parameters"].(map[string]interface{}),
			})
		}
	}

	return tools
}
