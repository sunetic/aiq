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
			Description: "Execute a SQL query against the database and return the results. Use this when the user wants to query the database.",
			Parameters: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"sql": map[string]interface{}{
						"type":        "string",
						"description": "The SQL query to execute",
					},
				},
				"required": []string{"sql"},
			},
		})
	}

	// Add render_table and render_chart (available in both modes)
	tools = append(tools, llm.Function{
		Name:        "render_table",
		Description: "Format query results as a table string. Use this when you want to show data in a tabular format.",
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
		Description: "Format query results as a chart string (bar, line, pie, or scatter). Use this when the user wants to visualize data.",
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
				"chart_type": map[string]interface{}{
					"type":        "string",
					"enum":        []string{"bar", "line", "pie", "scatter"},
					"description": "Type of chart to render",
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
