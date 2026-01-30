package sql

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aiq/aiq/internal/db"
	"github.com/aiq/aiq/internal/llm"
	"github.com/aiq/aiq/internal/prompt"
	"github.com/aiq/aiq/internal/skills"
	"github.com/aiq/aiq/internal/tool"
	"github.com/aiq/aiq/internal/tool/builtin"
	"github.com/aiq/aiq/internal/ui"
)

// ToolHandler handles tool execution and manages tool calling loop
type ToolHandler struct {
	conn          *db.Connection
	skillsManager *skills.Manager
	matcher       *skills.Matcher
	promptBuilder *prompt.Builder
	compressor    *prompt.Compressor
	promptLoader  *prompt.Loader
}

// NewToolHandler creates a new tool handler
func NewToolHandler(conn *db.Connection, skillsManager *skills.Manager, llmClient *llm.Client) *ToolHandler {
	matcher := skills.NewMatcher()
	if llmClient != nil {
		matcher.SetLLMClient(llmClient)
	}
	compressor := prompt.NewCompressor(prompt.DefaultContextWindow)
	if llmClient != nil {
		compressor.SetLLMClient(llmClient)
	}
	// Initialize prompt loader (loads prompts from ~/.aiqconfig/prompts)
	promptLoader, err := prompt.NewLoader()
	if err != nil {
		// Log error but continue with default prompts (fallback behavior)
		// This allows the system to work even if prompt files can't be loaded
		fmt.Printf("Warning: Failed to load prompts: %v. Using default prompts.\n", err)
		promptLoader = nil
	}
	return &ToolHandler{
		conn:          conn,
		skillsManager: skillsManager,
		matcher:       matcher,
		promptBuilder: prompt.NewBuilder(""), // Will be set in HandleToolCallLoop
		compressor:    compressor,
		promptLoader:  promptLoader,
	}
}

// formatToolCall formats a tool call for display, truncating long arguments
func (h *ToolHandler) formatToolCall(toolCall llm.ToolCall) string {
	toolName := toolCall.Function.Name
	argsStr := toolCall.Function.Arguments

	// Try to parse arguments to format them nicely
	args, err := toolCall.ParseArguments()
	if err != nil {
		// If parsing fails, just show the raw arguments (truncated)
		return h.formatToolCallWithRawArgs(toolName, argsStr)
	}

	// Format arguments based on tool type
	switch toolName {
	case "execute_sql":
		if sql, ok := args["sql"].(string); ok {
			return fmt.Sprintf("Calling tool [%s] with SQL: %s", toolName, h.truncateString(sql, 80))
		}
	case "execute_command":
		if cmd, ok := args["command"].(string); ok {
			return fmt.Sprintf("Calling tool [%s] with command: %s", toolName, h.truncateString(cmd, 80))
		}
	case "http_request":
		if url, ok := args["url"].(string); ok {
			method := "GET"
			if m, ok := args["method"].(string); ok {
				method = m
			}
			return fmt.Sprintf("Calling tool [%s] %s %s", toolName, method, h.truncateString(url, 60))
		}
	case "file_operations":
		if op, ok := args["operation"].(string); ok {
			if path, ok := args["path"].(string); ok {
				return fmt.Sprintf("Calling tool [%s] %s: %s", toolName, op, h.truncateString(path, 60))
			}
			return fmt.Sprintf("Calling tool [%s] %s", toolName, op)
		}
	case "render_table", "render_chart":
		if rows, ok := args["rows"].([]interface{}); ok {
			rowCount := len(rows)
			return fmt.Sprintf("Calling tool [%s] with %d row(s)", toolName, rowCount)
		}
	}

	// Default: show tool name with truncated arguments (cleaned up for display)
	displayStr := argsStr

	// If it's a quoted JSON string, unquote it for cleaner display
	if len(displayStr) >= 2 && displayStr[0] == '"' && displayStr[len(displayStr)-1] == '"' {
		var unquoted string
		if err := json.Unmarshal([]byte(displayStr), &unquoted); err == nil {
			displayStr = unquoted
		}
	}

	return fmt.Sprintf("Calling tool [%s] with args: %s", toolName, h.truncateString(displayStr, 60))
}

// formatToolCallWithRawArgs formats tool call when arguments can't be parsed
func (h *ToolHandler) formatToolCallWithRawArgs(toolName, argsStr string) string {
	// Clean up the JSON string for display - remove outer quotes and unescape
	displayStr := argsStr

	// If it's a quoted JSON string, unquote it
	if len(displayStr) >= 2 && displayStr[0] == '"' && displayStr[len(displayStr)-1] == '"' {
		// Use json.Unmarshal to properly unquote and unescape
		var unquoted string
		if err := json.Unmarshal([]byte(displayStr), &unquoted); err == nil {
			displayStr = unquoted
		}
	}

	return fmt.Sprintf("Calling tool [%s] with args: %s", toolName, h.truncateString(displayStr, 60))
}

// truncateString truncates a string to maxLen, adding "..." if truncated
func (h *ToolHandler) truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// ExecuteTool executes a tool call and returns the result
func (h *ToolHandler) ExecuteTool(ctx context.Context, toolCall llm.ToolCall) (json.RawMessage, error) {
	toolName := toolCall.Function.Name

	// Check if execute_sql is called in free mode (no connection)
	if toolName == "execute_sql" && h.conn == nil {
		errorJSON := map[string]interface{}{
			"status": "error",
			"error":  "SQL execution is not available in free mode. Please select a database source to enable SQL queries.",
		}
		jsonData, err := json.Marshal(errorJSON)
		if err != nil {
			errorMsg := `{"status":"error","error":"SQL execution is not available in free mode. Please select a database source to enable SQL queries."}`
			return json.RawMessage(errorMsg), nil
		}
		return json.RawMessage(jsonData), nil
	}

	// Parse arguments from JSON string
	// All JSON parsing errors are handled in ParseArguments()
	args, err := toolCall.ParseArguments()
	if err != nil {
		return nil, err
	}

	switch toolName {
	case "execute_sql":
		sql, ok := args["sql"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid sql parameter")
		}

		// Execute SQL - this does NOT print anything, only returns data
		result, err := tool.ExecuteSQL(ctx, h.conn, sql)
		if err != nil {
			// Properly encode error message as JSON
			// Include hint that LLM can check Skills for guidance and retry
			errorMessage := err.Error()
			errorJSON := map[string]interface{}{
				"status": "error",
				"error":  errorMessage,
				"hint":   "Check available Skills for correct syntax or required parameters. You may retry with corrections (max 2 attempts).",
			}
			jsonData, jsonErr := json.Marshal(errorJSON)
			if jsonErr != nil {
				// Fallback if JSON encoding fails
				errorMsg := fmt.Sprintf(`{"status":"error","error":"%s","hint":"Check available Skills for guidance and retry (max 2 attempts)"}`, strings.ReplaceAll(errorMessage, `"`, `\"`))
				return json.RawMessage(errorMsg), nil
			}
			return json.RawMessage(jsonData), nil
		}

		// Convert result to JSON and return to LLM
		// LLM will decide how to display this (via render_table or text description)
		resultJSON := map[string]interface{}{
			"status":    "success",
			"columns":   result.Columns,
			"rows":      result.Rows,
			"row_count": len(result.Rows),
		}

		// For DDL/DML operations (no data returned), add explicit completion message
		if len(result.Rows) == 0 {
			sqlUpper := strings.ToUpper(strings.TrimSpace(sql))
			isDDL := strings.HasPrefix(sqlUpper, "CREATE") || strings.HasPrefix(sqlUpper, "ALTER") ||
				strings.HasPrefix(sqlUpper, "DROP") || strings.HasPrefix(sqlUpper, "INSERT") ||
				strings.HasPrefix(sqlUpper, "UPDATE") || strings.HasPrefix(sqlUpper, "DELETE") ||
				strings.HasPrefix(sqlUpper, "CALL") || strings.HasPrefix(sqlUpper, "TRUNCATE")

			if isDDL {
				resultJSON["message"] = "SQL executed successfully. Task completed."
			}
		}

		jsonData, err := json.Marshal(resultJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		return json.RawMessage(jsonData), nil

	case "render_table":
		columnsInterface, ok := args["columns"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid columns parameter")
		}
		rowsInterface, ok := args["rows"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid rows parameter")
		}

		// Convert to string slices
		columns := make([]string, len(columnsInterface))
		for i, col := range columnsInterface {
			columns[i] = fmt.Sprintf("%v", col)
		}

		rows := make([][]string, len(rowsInterface))
		for i, rowInterface := range rowsInterface {
			rowArray, ok := rowInterface.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid row format")
			}
			rows[i] = make([]string, len(rowArray))
			for j, val := range rowArray {
				rows[i][j] = fmt.Sprintf("%v", val)
			}
		}

		// Format the table as a string (do not print)
		tableOutput, err := tool.RenderTableString(columns, rows)
		if err != nil {
			errorMsg := fmt.Sprintf(`{"error": "%s"}`, err.Error())
			return json.RawMessage(errorMsg), nil
		}

		// Return formatted table to LLM
		resultJSON := map[string]interface{}{
			"status":    "success",
			"format":    "table",
			"output":    tableOutput,
			"row_count": len(rows),
		}
		jsonData, err := json.Marshal(resultJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		return json.RawMessage(jsonData), nil

	case "render_chart":
		columnsInterface, ok := args["columns"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid columns parameter")
		}
		rowsInterface, ok := args["rows"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid rows parameter")
		}
		chartTypeStr, ok := args["chart_type"].(string)
		if !ok {
			return nil, fmt.Errorf("invalid chart_type parameter")
		}

		// Convert to string slices
		columns := make([]string, len(columnsInterface))
		for i, col := range columnsInterface {
			columns[i] = fmt.Sprintf("%v", col)
		}

		rows := make([][]string, len(rowsInterface))
		for i, rowInterface := range rowsInterface {
			rowArray, ok := rowInterface.([]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid row format")
			}
			rows[i] = make([]string, len(rowArray))
			for j, val := range rowArray {
				rows[i][j] = fmt.Sprintf("%v", val)
			}
		}

		// Create QueryResult
		result := &db.QueryResult{
			Columns: columns,
			Rows:    rows,
		}

		chartOutput, err := tool.RenderChartString(result, chartTypeStr)
		if err != nil {
			errorMsg := fmt.Sprintf(`{"error": "%s"}`, err.Error())
			return json.RawMessage(errorMsg), nil
		}

		// Return formatted chart to LLM
		resultJSON := map[string]interface{}{
			"status":     "success",
			"format":     "chart",
			"output":     chartOutput,
			"chart_type": chartTypeStr,
			"row_count":  len(result.Rows),
		}
		jsonData, err := json.Marshal(resultJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		return json.RawMessage(jsonData), nil

	default:
		// Try built-in tools - use ExecuteBuiltinTool (no callback support yet)
		result, err := builtin.ExecuteBuiltinTool(ctx, toolName, args, h.conn)
		if err != nil {
			// Check if it's truly an unknown tool or an execution error
			if strings.Contains(err.Error(), "unknown built-in tool") {
				return nil, fmt.Errorf("unknown tool: %s", toolName)
			}
			// For execution errors, encode as JSON similar to execute_sql errors
			errorMessage := err.Error()
			errorJSON := map[string]interface{}{
				"status": "error",
				"error":  errorMessage,
			}
			jsonData, jsonErr := json.Marshal(errorJSON)
			if jsonErr != nil {
				// Fallback if JSON encoding fails
				errorMsg := fmt.Sprintf(`{"status":"error","error":"%s"}`, strings.ReplaceAll(errorMessage, `"`, `\"`))
				return json.RawMessage(errorMsg), nil
			}
			return json.RawMessage(jsonData), nil
		}
		// Convert result to JSON
		// For execute_command, use truncated output for LLM and add status field
		if toolName == "execute_command" {
			// First marshal to JSON, then unmarshal to map to add status and use truncated output
			jsonBytes, err := json.Marshal(result)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}

			var resultMap map[string]interface{}
			if err := json.Unmarshal(jsonBytes, &resultMap); err != nil {
				return nil, fmt.Errorf("failed to unmarshal result: %w", err)
			}

			// Check exit_code to determine status (simple rule: 0 = success, non-zero = error)
			exitCode := 0
			if ec, ok := resultMap["exit_code"].(float64); ok {
				exitCode = int(ec)
			} else if ec, ok := resultMap["exit_code"].(int); ok {
				exitCode = ec
			}

			// Save full output for display (before truncation)
			fullStdout, _ := resultMap["stdout"].(string)
			fullStderr, _ := resultMap["stderr"].(string)

			// Use truncated output for LLM (if available)
			if truncatedStdout, ok := resultMap["truncated_stdout"].(string); ok && truncatedStdout != "" {
				// Keep full output in a separate field for display
				resultMap["_full_stdout"] = fullStdout
				resultMap["stdout"] = truncatedStdout
			}
			if truncatedStderr, ok := resultMap["truncated_stderr"].(string); ok && truncatedStderr != "" {
				// Keep full output in a separate field for display
				resultMap["_full_stderr"] = fullStderr
				resultMap["stderr"] = truncatedStderr
			}

			// Remove truncated fields from JSON (they're only for internal use)
			delete(resultMap, "truncated_stdout")
			delete(resultMap, "truncated_stderr")

			// Add explicit status field based on exit_code
			if exitCode == 0 {
				resultMap["status"] = "success"
			} else {
				resultMap["status"] = "error"
				resultMap["error"] = fmt.Sprintf("Command exited with code %d", exitCode)
			}

			// Note: Keep _full_stdout and _full_stderr for display in tool_handler
			// They will be removed before sending to LLM in the tool call loop

			// Marshal back to JSON
			jsonData, err := json.Marshal(resultMap)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal result: %w", err)
			}
			return json.RawMessage(jsonData), nil
		}

		// For other tools, convert result to JSON as-is
		resultJSON, err := json.Marshal(result)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal result: %w", err)
		}
		return json.RawMessage(resultJSON), nil
	}
}

// HandleToolCallLoop handles the complete tool calling loop
// Returns the final response content and any query result
// If schemaContext is empty, runs in free mode (no database connection)
func (h *ToolHandler) HandleToolCallLoop(ctx context.Context, llmClient *llm.Client, userInput string, schemaContext string, databaseType string, conversationHistory []llm.ChatMessage, tools []llm.Function) (string, *db.QueryResult, error) {
	// Determine mode: free mode or database mode
	isFreeMode := schemaContext == "" || h.conn == nil

	// Load prompts from files or use defaults
	var baseSystemPrompt string
	var commonPrompt string

	if h.promptLoader != nil {
		// Use prompts loaded from ~/.aiqconfig/prompts
		if isFreeMode {
			baseSystemPrompt = h.promptLoader.GetFreeModeBasePrompt()
		} else {
			// Get database base prompt + database-specific patch
			baseSystemPrompt = h.promptLoader.GetDatabaseModeBasePrompt(databaseType, schemaContext)
		}
		commonPrompt = h.promptLoader.GetCommonPrompt()
	} else {
		// Fallback to hardcoded defaults if prompt loader failed
		if isFreeMode {
			baseSystemPrompt = `<MODE>
FREE MODE - No database connection available. SQL execution is not available.
</MODE>

<ROLE>
You are a helpful AI assistant. You can have natural conversations and help with system operations using available tools.
</ROLE>

<TOOLS>
- execute_command: System operations (install, setup, configuration). Not for database queries.
- http_request: Make HTTP requests.
- file_operations: Read/write files.
</TOOLS>

<POLICY>
- If the user asks for database operations, explain that no database is connected and ask whether they want to select a source.
- Do not guess database commands or run mysql/psql in free mode.
- If the request is ambiguous for the current mode, ask a clarifying question before acting.
</POLICY>`
		} else {
			baseSystemPrompt = fmt.Sprintf(`<MODE>
DATABASE MODE - Connected to a database.
</MODE>

<ROLE>
You are a helpful AI assistant for database queries and related tasks.
</ROLE>

<CONTEXT>
- Database engine type: %s
- Database connection and schema information:
%s
</CONTEXT>

<POLICY>
- Use execute_sql for database queries. Do not use execute_command to run mysql/psql.
- Respect engine-specific syntax. If unsure, ask a clarifying question or rely on schema context.
- If a request is not a database query, use the appropriate non-SQL tools.
</POLICY>

<TOOLS>
- execute_sql: Execute SQL queries against the database.
- render_table: Format query results as a table.
- render_chart: Format query results as a chart when user explicitly asks for visualization.
- execute_command: System operations (install, setup, configuration). Not for database queries.
- http_request: Make HTTP requests.
- file_operations: Read/write files.
</TOOLS>
`, databaseType, schemaContext)
		}

		// Fallback common prompt
		commonPrompt = `<EXECUTION>
- For system operations, use execute_command with explicit commands.
- If a command requires elevated privileges or interactive input, ask the user to run it manually and explain why.
- Do not fabricate command outputs. Use tool results to decide the next step.
</EXECUTION>`

	}

	// Combine base prompt with common sections
	baseSystemPrompt = baseSystemPrompt + commonPrompt

	// Match Skills to user query and manage dynamic loading/eviction
	var loadedSkills []*skills.Skill
	if h.skillsManager != nil {
		metadataList := h.skillsManager.GetMetadata()
		if len(metadataList) > 0 {
			matchedMetadata := h.matcher.Match(userInput, metadataList)

			// Evict Skills not matched in recent queries before loading new ones
			evicted := h.skillsManager.EvictUnusedSkills(skills.DefaultEvictionQueries)
			if len(evicted) > 0 {
				ui.ShowInfo(fmt.Sprintf("Evicted %d unused skill(s): %v", len(evicted), evicted))
			}

			if len(matchedMetadata) > 0 {
				// Track usage for matched Skills
				skillNames := make([]string, len(matchedMetadata))
				skillMetadataMap := make(map[string]*skills.Metadata) // Map for quick lookup
				for i, md := range matchedMetadata {
					skillNames[i] = md.Name
					skillMetadataMap[md.Name] = md
					// Track usage
					h.skillsManager.TrackUsage(md.Name, userInput)
					// Set priority: matched Skills are relevant
					h.skillsManager.SetPriority(md.Name, skills.PriorityRelevant)
				}

				loaded, err := h.skillsManager.LoadSkills(skillNames)
				if err == nil {
					loadedSkills = loaded
					// Set priority to active for loaded Skills
					for _, skill := range loadedSkills {
						h.skillsManager.SetPriority(skill.Name, skills.PriorityActive)
					}
					// Show which Skills were loaded with descriptions
					if len(loadedSkills) > 0 {
						fmt.Print(ui.InfoText("Loaded "))
						fmt.Print(ui.HighlightText(fmt.Sprintf("%d skill(s)", len(loadedSkills))))
						fmt.Print(ui.InfoText(": "))
						skillDisplays := make([]string, 0, len(loadedSkills))
						for _, skill := range loadedSkills {
							if md, exists := skillMetadataMap[skill.Name]; exists && md.Description != "" {
								skillDisplays = append(skillDisplays, fmt.Sprintf("%s - %s", ui.HighlightText(skill.Name), ui.HintText(md.Description)))
							} else {
								skillDisplays = append(skillDisplays, ui.HighlightText(skill.Name))
							}
						}
						fmt.Print(strings.Join(skillDisplays, ", "))
						fmt.Println()
					}
				} else {
					ui.ShowWarning(fmt.Sprintf("Failed to load some skills: %v", err))
				}
			}
		}
	}

	// Build system prompt with Skills
	h.promptBuilder = prompt.NewBuilder(baseSystemPrompt)
	systemPrompt := h.promptBuilder.BuildSystemPrompt(loadedSkills)

	// Convert conversation history to strings for compression
	historyStrings := make([]string, len(conversationHistory))
	for i, msg := range conversationHistory {
		historyStrings[i] = fmt.Sprintf("%s: %s", msg.Role, msg.Content)
	}

	// Compress prompt if needed
	compressionResult, err := h.compressor.Compress(historyStrings, loadedSkills, systemPrompt, userInput)
	if err == nil && compressionResult.Compressed {
		// Rebuild conversation history from compressed version
		conversationHistory = make([]llm.ChatMessage, 0, len(compressionResult.CompressedHistory))
		for _, histStr := range compressionResult.CompressedHistory {
			// Parse "role: content" format
			parts := strings.SplitN(histStr, ": ", 2)
			if len(parts) == 2 {
				conversationHistory = append(conversationHistory, llm.ChatMessage{
					Role:    parts[0],
					Content: parts[1],
				})
			}
		}
		loadedSkills = compressionResult.RemainingSkills
		// Rebuild system prompt with remaining Skills
		systemPrompt = h.promptBuilder.BuildSystemPrompt(loadedSkills)
	}

	// Build initial messages
	messages := []interface{}{
		llm.ChatMessage{
			Role:    "system",
			Content: systemPrompt,
		},
	}

	// Add conversation history
	for _, msg := range conversationHistory {
		messages = append(messages, msg)
	}

	// Add user input
	messages = append(messages, llm.ChatMessage{
		Role:    "user",
		Content: userInput,
	})

	var lastQueryResult *db.QueryResult
	var lastFormattedOutput string
	maxIterations := 10 // Prevent infinite loops

	for i := 0; i < maxIterations; i++ {
		// Call LLM - show "Thinking..." while LLM is processing
		stopThinking := ui.ShowLoading("Thinking...")
		response, err := llmClient.ChatWithTools(ctx, messages, tools)
		stopThinking()
		if err != nil {
			return "", nil, fmt.Errorf("LLM call failed: %w", err)
		}

		if len(response.Choices) == 0 {
			return "", nil, fmt.Errorf("no choices in response")
		}

		choice := response.Choices[0]
		message := choice.Message

		// Add assistant message to history
		// If there are tool_calls, content might be empty or null
		assistantMsg := map[string]interface{}{
			"role": "assistant",
		}
		// Only add content if it's not empty
		if message.Content != "" {
			assistantMsg["content"] = message.Content
		}
		if len(message.ToolCalls) > 0 {
			assistantMsg["tool_calls"] = message.ToolCalls
		}
		messages = append(messages, assistantMsg)

		// If no tool calls, return the final response
		if len(message.ToolCalls) == 0 {
			if message.Content == "" {
				if lastFormattedOutput != "" {
					return lastFormattedOutput, lastQueryResult, nil
				}
				return "", lastQueryResult, fmt.Errorf("empty response from LLM")
			}
			return message.Content, lastQueryResult, nil
		}

		// Execute tool calls (show status only)
		for _, toolCall := range message.ToolCalls {
			// For execute_sql, show SQL and ask for confirmation before executing
			if toolCall.Function.Name == "execute_sql" {
				args, err := toolCall.ParseArguments()
				if err != nil {
					ui.ShowError(fmt.Sprintf("Tool [%s] failed: %v", toolCall.Function.Name, err))
					errorMsg := fmt.Sprintf(`{"error": "%s"}`, err.Error())
					toolResult := json.RawMessage(errorMsg)
					toolMsg := map[string]interface{}{
						"role":         "tool",
						"content":      string(toolResult),
						"tool_call_id": toolCall.ID,
					}
					messages = append(messages, toolMsg)
					continue
				}

				sql, ok := args["sql"].(string)
				if !ok {
					err := fmt.Errorf("invalid sql parameter")
					ui.ShowError(fmt.Sprintf("Tool [%s] failed: %v", toolCall.Function.Name, err))
					errorMsg := fmt.Sprintf(`{"error": "%s"}`, err.Error())
					toolResult := json.RawMessage(errorMsg)
					toolMsg := map[string]interface{}{
						"role":         "tool",
						"content":      string(toolResult),
						"tool_call_id": toolCall.ID,
					}
					messages = append(messages, toolMsg)
					continue
				}

				fmt.Println()
				ui.ShowInfo("Generated SQL:")
				fmt.Println(ui.HighlightSQL(sql))
				fmt.Println()

				confirm, err := ui.ShowConfirm("Execute this query?")
				if err != nil {
					fmt.Println()
					// Treat as cancelled
					ui.ShowWarning("Query execution cancelled.")
					toolResult := json.RawMessage(`{"status":"cancelled","message":"query execution cancelled by user"}`)
					toolMsg := map[string]interface{}{
						"role":         "tool",
						"content":      string(toolResult),
						"tool_call_id": toolCall.ID,
					}
					messages = append(messages, toolMsg)
					continue
				}
				if !confirm {
					ui.ShowWarning("Query execution cancelled.")
					toolResult := json.RawMessage(`{"status":"cancelled","message":"query execution cancelled by user"}`)
					toolMsg := map[string]interface{}{
						"role":         "tool",
						"content":      string(toolResult),
						"tool_call_id": toolCall.ID,
					}
					messages = append(messages, toolMsg)
					continue
				}
			}

			// Format and display tool call with arguments
			toolCallDisplay := h.formatToolCall(toolCall)
			var startTime time.Time
			var toolResult json.RawMessage
			var err error

			// Special handling for execute_command: track time and streaming output
			if toolCall.Function.Name == "execute_command" {
				// Display tool call with loading icon (normal color for main command)
				fmt.Println("⏳ " + toolCallDisplay)
				startTime = time.Now()

				// Create rolling output for real-time display
				rollingOutput := ui.NewRollingOutput(3)

				// Parse arguments
				args, parseErr := toolCall.ParseArguments()
				if parseErr != nil {
					err = parseErr
					toolResult = nil
				} else {
					// Execute with callback for streaming output - rolling window display
					result, execErr := builtin.ExecuteBuiltinToolWithCallback(ctx, "execute_command", args, h.conn, func(line string) {
						// AddLine handles the rolling display (clears old lines, prints new ones)
						rollingOutput.AddLine(line)
					})

					// Show summary after command completes
					rollingOutput.Finish()

					if execErr != nil {
						err = execErr
						toolResult = nil
					} else {
						// Convert result to JSON
						jsonBytes, marshalErr := json.Marshal(result)
						if marshalErr != nil {
							err = fmt.Errorf("failed to marshal result: %w", marshalErr)
							toolResult = nil
						} else {
							// Process result for LLM (truncation, status, etc.)
							var resultMap map[string]interface{}
							if unmarshalErr := json.Unmarshal(jsonBytes, &resultMap); unmarshalErr == nil {
								exitCode := 0
								if ec, ok := resultMap["exit_code"].(float64); ok {
									exitCode = int(ec)
								} else if ec, ok := resultMap["exit_code"].(int); ok {
									exitCode = ec
								}

								// Save full output for display
								fullStdout, _ := resultMap["stdout"].(string)
								fullStderr, _ := resultMap["stderr"].(string)

								// Use truncated output for LLM
								if truncatedStdout, ok := resultMap["truncated_stdout"].(string); ok && truncatedStdout != "" {
									resultMap["_full_stdout"] = fullStdout
									resultMap["stdout"] = truncatedStdout
								}
								if truncatedStderr, ok := resultMap["truncated_stderr"].(string); ok && truncatedStderr != "" {
									resultMap["_full_stderr"] = fullStderr
									resultMap["stderr"] = truncatedStderr
								}

								delete(resultMap, "truncated_stdout")
								delete(resultMap, "truncated_stderr")

								// Add status
								if exitCode == 0 {
									resultMap["status"] = "success"
								} else {
									resultMap["status"] = "error"
									resultMap["error"] = fmt.Sprintf("Command exited with code %d", exitCode)
								}

								jsonData, _ := json.Marshal(resultMap)
								toolResult = json.RawMessage(jsonData)
							}
						}
					}
				}
			} else {
				// For other tools, use normal display
				ui.ShowInfo(toolCallDisplay)

				waitingMsg := "Waiting..."
				if toolCall.Function.Name == "execute_sql" {
					waitingMsg = "Executing SQL..."
				} else if toolCall.Function.Name == "http_request" {
					waitingMsg = "Waiting for HTTP response..."
				}
				stopWaiting := ui.ShowLoading(waitingMsg)
				toolResult, err = h.ExecuteTool(ctx, toolCall)
				stopWaiting()
			}
			if err != nil {
				// Format error message for LLM
				errorMsg := fmt.Sprintf(`{"error": "%s"}`, strings.ReplaceAll(err.Error(), `"`, `\"`))
				toolResult = json.RawMessage(errorMsg)
				// Show user-friendly error message with duration for execute_command
				if toolCall.Function.Name == "execute_command" {
					duration := time.Since(startTime)
					ui.ShowError(fmt.Sprintf("Tool [execute_command] failed: %s (%.1fs)", err.Error(), duration.Seconds()))
				} else {
					ui.ShowError(fmt.Sprintf("Tool [%s] failed: %s", toolCall.Function.Name, err.Error()))
				}
			} else {
				// Check if tool result contains an error (even if ExecuteTool returned nil error)
				var resultData map[string]interface{}
				if jsonErr := json.Unmarshal(toolResult, &resultData); jsonErr == nil {
					// Special handling for execute_command: display status (output already printed in real-time)
					if toolCall.Function.Name == "execute_command" {
						exitCode := 0
						if ec, ok := resultData["exit_code"].(float64); ok {
							exitCode = int(ec)
						} else if ec, ok := resultData["exit_code"].(int); ok {
							exitCode = ec
						}

						// Calculate execution time
						duration := time.Since(startTime)

						// Display status with icon and duration
						if exitCode == 0 {
							ui.ShowSuccess(fmt.Sprintf("Tool [execute_command] completed (%.1fs)", duration.Seconds()))
						} else {
							ui.ShowError(fmt.Sprintf("Tool [execute_command] failed with exit code %d (%.1fs)", exitCode, duration.Seconds()))
						}
						fmt.Println()

						// Remove internal fields before sending to LLM
						delete(resultData, "_full_stdout")
						delete(resultData, "_full_stderr")
						// Re-marshal toolResult without internal fields
						jsonData, _ := json.Marshal(resultData)
						toolResult = json.RawMessage(jsonData)
					} else {
						// For other tools, use existing display logic
						if errorMsg, hasError := resultData["error"].(string); hasError && errorMsg != "" {
							ui.ShowError(fmt.Sprintf("Tool [%s] failed: %s", toolCall.Function.Name, errorMsg))
						} else {
							ui.ShowSuccess(fmt.Sprintf("Tool [%s] executed successfully", toolCall.Function.Name))
						}
					}
				} else {
					ui.ShowSuccess(fmt.Sprintf("Tool [%s] executed successfully", toolCall.Function.Name))
				}
			}

			// For execute_sql: directly render table output (mysql client style)
			// and simplify the result sent to LLM
			if toolCall.Function.Name == "execute_sql" && err == nil {
				var resultData map[string]interface{}
				if err := json.Unmarshal(toolResult, &resultData); err == nil {
					if columns, ok := resultData["columns"].([]interface{}); ok {
						if rows, ok := resultData["rows"].([]interface{}); ok {
							// Convert to QueryResult
							cols := make([]string, len(columns))
							for i, col := range columns {
								cols[i] = fmt.Sprintf("%v", col)
							}
							rowsData := make([][]string, len(rows))
							for i, rowInterface := range rows {
								rowArray, ok := rowInterface.([]interface{})
								if ok {
									rowsData[i] = make([]string, len(rowArray))
									for j, val := range rowArray {
										rowsData[i][j] = fmt.Sprintf("%v", val)
									}
								}
							}
							lastQueryResult = &db.QueryResult{
								Columns: cols,
								Rows:    rowsData,
							}

							// Directly render table output (mysql client style)
							if len(rowsData) > 0 {
								fmt.Println()
								tableOutput, tableErr := tool.RenderTableString(cols, rowsData)
								if tableErr == nil {
									fmt.Println(tableOutput)
								}
								fmt.Printf("%d row(s) in set\n", len(rowsData))
							}

							// Simplify result for LLM - don't send raw data
							// Explicitly instruct LLM not to repeat the data
							simplifiedResult := map[string]interface{}{
								"status":      "success",
								"row_count":   len(rowsData),
								"displayed":   true,
								"instruction": "Results already displayed to user in table format. Do NOT list, repeat, or summarize the data. Just confirm completion or ask if user needs anything else.",
							}
							simplifiedJSON, _ := json.Marshal(simplifiedResult)
							toolResult = json.RawMessage(simplifiedJSON)
						}
					}
				}
			}

			// Store last formatted output if tool returned it
			if err == nil {
				var outputData map[string]interface{}
				if jsonErr := json.Unmarshal(toolResult, &outputData); jsonErr == nil {
					if output, ok := outputData["output"].(string); ok && output != "" {
						lastFormattedOutput = output
					}
				}
			}

			// Step 5: Add tool result message back to LLM
			// LLM will decide next step based on tool results
			toolMsg := map[string]interface{}{
				"role":         "tool",
				"content":      string(toolResult), // Convert json.RawMessage to string
				"tool_call_id": toolCall.ID,
			}
			messages = append(messages, toolMsg)
		}
		// Loop continues: LLM will process tool results and decide next action
	}

	return "", nil, fmt.Errorf("max iterations reached")
}
