package sql

import (
	"context"
	"fmt"
	"strings"

	"github.com/chzyer/readline"

	"github.com/aiq/aiq/internal/chart"
	"github.com/aiq/aiq/internal/config"
	"github.com/aiq/aiq/internal/db"
	"github.com/aiq/aiq/internal/llm"
	"github.com/aiq/aiq/internal/session"
	"github.com/aiq/aiq/internal/skills"
	"github.com/aiq/aiq/internal/source"
	"github.com/aiq/aiq/internal/tool"
	"github.com/aiq/aiq/internal/ui"
)

// RunSQLMode runs the SQL interactive mode
// sessionFile is optional path to a session file to restore
func RunSQLMode(sessionFile string) error {
	var sess *session.Session
	var src *source.Source
	var sourceName string

	// Restore session if provided
	if sessionFile != "" {
		loadedSession, err := session.LoadSession(sessionFile)
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to load session: %v", err))
			ui.ShowInfo("Starting with a new session.")
		} else {
			sess = loadedSession
			sourceName = sess.Metadata.DataSource
			ui.ShowInfo(fmt.Sprintf("Restored session from %s", sessionFile))
			ui.ShowInfo(fmt.Sprintf("Conversation history: %d messages", len(sess.Messages)))
		}
	}

	// Select source (if not restored from session) - now optional for free mode
	if sess == nil || sourceName == "" {
		sources, err := source.LoadSources()
		if err != nil {
			return fmt.Errorf("failed to load sources: %w", err)
		}

		// If no sources configured, enter free mode automatically
		if len(sources) == 0 {
			ui.ShowInfo("No data sources configured. Entering free mode (general conversation and Skills only, no SQL).")
			src = nil
			sourceName = ""
		} else {
			// Build menu items with sources and skip option
			items := make([]ui.MenuItem, 0, len(sources)+1)
			for _, s := range sources {
				label := fmt.Sprintf("%s (%s/%s:%d/%s)", s.Name, s.Type, s.Host, s.Port, s.Database)
				items = append(items, ui.MenuItem{Label: label, Value: s.Name})
			}
			items = append(items, ui.MenuItem{Label: "Skip (free mode) - General conversation and Skills only", Value: "__free_mode__"})

			sourceName, err = ui.ShowMenu("Select Data Source", items)
			if err != nil {
				return err
			}

			// Check if user chose free mode
			if sourceName == "__free_mode__" {
				src = nil
				sourceName = ""
			} else {
				// Load selected source
				src, err = source.GetSource(sourceName)
				if err != nil {
					return fmt.Errorf("failed to load source: %w", err)
				}
			}
		}
	} else {
		// Load source from session
		var err error
		src, err = source.GetSource(sourceName)
		if err != nil {
			// If source from session doesn't exist, prompt for new one or free mode
			if sess != nil {
				ui.ShowWarning(fmt.Sprintf("Data source '%s' from session no longer exists.", sourceName))
				sources, loadErr := source.LoadSources()
				if loadErr != nil {
					return fmt.Errorf("failed to load sources: %w", loadErr)
				}
				if len(sources) == 0 {
					ui.ShowInfo("No data sources available. Entering free mode.")
					src = nil
					sourceName = ""
				} else {
					items := make([]ui.MenuItem, 0, len(sources)+1)
					for _, s := range sources {
						label := fmt.Sprintf("%s (%s/%s:%d/%s)", s.Name, s.Type, s.Host, s.Port, s.Database)
						items = append(items, ui.MenuItem{Label: label, Value: s.Name})
					}
					items = append(items, ui.MenuItem{Label: "Skip (free mode) - General conversation and Skills only", Value: "__free_mode__"})
					sourceName, err = ui.ShowMenu("Select Data Source", items)
					if err != nil {
						return err
					}
					if sourceName == "__free_mode__" {
						src = nil
						sourceName = ""
					} else {
						src, err = source.GetSource(sourceName)
						if err != nil {
							return fmt.Errorf("failed to load source: %w", err)
						}
					}
				}
			} else {
				return fmt.Errorf("failed to load source: %w", err)
			}
		}
	}

	// Create new session if not restored
	if sess == nil {
		if src != nil {
			sess = session.NewSession(sourceName, string(src.Type))
		} else {
			// Free mode session (no source)
			sess = session.NewSession("", "")
		}
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create database connection only if source exists
	var conn *db.Connection
	var schema *db.Schema
	ctx := context.Background() // Create context for use throughout the function
	if src != nil {
		var err error
		conn, err = db.NewConnection(src.DSN(), string(src.Type))
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		// Fetch schema for context
		schema, err = conn.GetSchema(ctx, src.Database)
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to fetch schema: %v. Continuing without schema context.", err))
			schema = &db.Schema{}
		}
	}

	// Initialize Skills manager
	skillsManager := skills.NewManager()
	if err := skillsManager.Initialize(); err != nil {
		ui.ShowWarning(fmt.Sprintf("Failed to initialize Skills manager: %v. Continuing without Skills.", err))
	} else {
		// Show Skills loading status
		metadata := skillsManager.GetMetadata()
		if len(metadata) > 0 {
			ui.ShowInfo(fmt.Sprintf("Loaded %d skill(s) metadata (progressive loading enabled)", len(metadata)))
			for _, md := range metadata {
				ui.ShowInfo(fmt.Sprintf("  - %s: %s", md.Name, md.Description))
			}
		}
	}

	// Create LLM client
	llmClient := llm.NewClient(cfg.LLM.URL, cfg.LLM.APIKey, cfg.LLM.Model)

	// Show mode info
	if src != nil {
		ui.ShowInfo(fmt.Sprintf("Entering chat mode. Source: %s (%s/%s:%d/%s)", src.Name, src.Type, src.Host, src.Port, src.Database))
	} else {
		ui.ShowInfo("Entering free mode (general conversation and Skills only, no SQL execution)")
	}
	if len(sess.Messages) > 0 {
		ui.ShowInfo(fmt.Sprintf("Conversation history: %d messages", len(sess.Messages)))
	}
	// Show Skills status
	metadata := skillsManager.GetMetadata()
	if len(metadata) > 0 {
		ui.ShowInfo(fmt.Sprintf("Skills: %d skill(s) available (progressive loading enabled)", len(metadata)))
		for _, md := range metadata {
			ui.ShowInfo(fmt.Sprintf("  - %s: %s", md.Name, md.Description))
		}
	} else {
		ui.ShowInfo("Skills: No skills found in ~/.aiqconfig/skills/")
	}
	ui.ShowInfo("Chat freely or ask database questions. Use '/history' to view conversation, '/clear' to clear history, or 'exit' to return to main menu.")
	fmt.Println()

	// Store last query result for view switching
	var lastQueryResult *db.QueryResult
	// Store last generated SQL for execute command
	var lastGeneratedSQL string

	// Build dynamic prompt based on source availability
	var buildPrompt func() string
	if src != nil {
		buildPrompt = func() string {
			return ui.InfoText(fmt.Sprintf("aiq[%s]> ", src.Name))
		}
	} else {
		buildPrompt = func() string {
			return ui.InfoText("aiq> ")
		}
	}

	// Use readline for better Unicode/Chinese character support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          buildPrompt(),
		HistoryFile:     "",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	// Update prompt dynamically (readline doesn't support dynamic prompts directly,
	// but we can recreate it if source changes in future)
	// For now, prompt is set once at initialization

	for {
		// Read input line (readline handles Unicode properly)
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl+C - continue to next prompt
				fmt.Println()
				continue
			}
			// EOF (Ctrl+D) - exit chat mode
			fmt.Println()
			// Save session before exiting
			timestamp := session.GetTimestamp()
			sessionPath, err := session.GetSessionFilePath(timestamp)
			if err != nil {
				ui.ShowWarning(fmt.Sprintf("Failed to generate session path: %v", err))
			} else {
				if err := session.SaveSession(sess, sessionPath); err != nil {
					ui.ShowWarning(fmt.Sprintf("Failed to save session: %v", err))
				} else {
					ui.ShowInfo(fmt.Sprintf("Current session saved to %s", sessionPath))
					ui.ShowInfo(fmt.Sprintf("Run 'aiq -s %s' to continue.", sessionPath))
				}
			}
			ui.ShowInfo("Exiting chat mode (EOF).")
			return nil
		}

		line = strings.TrimSpace(line)

		// Handle empty input
		if line == "" {
			continue
		}

		// Use the line directly as query (readline handles multi-byte characters correctly)
		query := line

		// Handle special commands
		if strings.ToLower(query) == "exit" || strings.ToLower(query) == "back" {
			// Save session before exiting
			timestamp := session.GetTimestamp()
			sessionPath, err := session.GetSessionFilePath(timestamp)
			if err != nil {
				ui.ShowWarning(fmt.Sprintf("Failed to generate session path: %v", err))
			} else {
				if err := session.SaveSession(sess, sessionPath); err != nil {
					ui.ShowWarning(fmt.Sprintf("Failed to save session: %v", err))
				} else {
					ui.ShowInfo(fmt.Sprintf("Current session saved to %s", sessionPath))
					ui.ShowInfo(fmt.Sprintf("Run 'aiq -s %s' to continue.", sessionPath))
				}
			}
			return nil
		}

		// Handle /history command
		if strings.ToLower(query) == "/history" {
			history := sess.GetHistory()
			if len(history) == 0 {
				ui.ShowInfo("No conversation history.")
			} else {
				fmt.Println()
				ui.ShowInfo("Conversation History:")
				fmt.Println()
				for i, msg := range history {
					roleLabel := "User"
					if msg.Role == "assistant" {
						roleLabel = "Assistant"
					}
					fmt.Printf("[%d] %s (%s):\n", i+1, roleLabel, msg.Timestamp.Format("15:04:05"))
					if msg.Role == "assistant" {
						fmt.Println(ui.HighlightSQL(msg.Content))
					} else {
						fmt.Println(msg.Content)
					}
					fmt.Println()
				}
			}
			fmt.Println()
			continue
		}

		// Handle /clear command
		if strings.ToLower(query) == "/clear" {
			confirm, err := ui.ShowConfirm("Clear conversation history?")
			if err != nil {
				fmt.Println()
				continue
			}
			if confirm {
				sess.ClearHistory()
				ui.ShowInfo("Conversation history cleared.")
			}
			fmt.Println()
			continue
		}

		// Handle execute command - execute last generated SQL
		if strings.ToLower(query) == "execute" || strings.ToLower(query) == "/execute" {
			if lastGeneratedSQL == "" {
				ui.ShowWarning("No SQL query to execute. Please generate a query first.")
				fmt.Println()
				continue
			}

			// Execute the last generated SQL
			stopLoading := ui.ShowLoading("Calling tool [execute_sql]...")
			result, err := tool.ExecuteSQL(ctx, conn, lastGeneratedSQL)
			stopLoading()

			if err != nil {
				// Error already displayed by tool
				ui.ShowInfo("You can modify the query and try again.")
				fmt.Println()
				continue
			}
			// Tool success message is displayed by tool.ExecuteSQL

			// Store result for potential view switching
			lastQueryResult = result

			// Display results
			fmt.Println()
			if len(result.Rows) == 0 {
				ui.ShowInfo("Query executed successfully. No rows returned.")
				fmt.Println()
				continue
			}

			ui.ShowSuccess(fmt.Sprintf("Query executed successfully. %d row(s) returned.", len(result.Rows)))
			fmt.Println()

			// Automatically display as table for execute command
			ui.PrintTable(result.Columns, result.Rows)
			fmt.Println()
			continue
		}

		// Check if this is a view switching command (e.g., "display as pie chart", "show as bar")
		queryLower := strings.ToLower(query)
		isViewSwitch := (strings.Contains(queryLower, "display as") || strings.Contains(queryLower, "show as") ||
			strings.Contains(queryLower, "view as") || strings.Contains(queryLower, "render as")) &&
			(strings.Contains(queryLower, "chart") || strings.Contains(queryLower, "pie") ||
				strings.Contains(queryLower, "bar") || strings.Contains(queryLower, "line") ||
				strings.Contains(queryLower, "scatter") || strings.Contains(queryLower, "table"))

		// If it's a view switch command and we have last query result, use it directly with tools
		if isViewSwitch && lastQueryResult != nil {
			// Extract chart type from query
			chartType := detectChartTypeFromQuery(queryLower)
			if chartType != "" {
				// Display the last result with the requested chart type
				if chartType == "table" {
					ui.PrintTable(lastQueryResult.Columns, lastQueryResult.Rows)
					fmt.Println()
				} else {
					// Basic guard: chart types need at least 2 columns
					if len(lastQueryResult.Columns) < 2 {
						ui.ShowWarning(fmt.Sprintf("Cannot render %s chart: requires at least 2 columns. Showing table instead.", chartType))
						ui.PrintTable(lastQueryResult.Columns, lastQueryResult.Rows)
						fmt.Println()
						// Add to conversation history
						sess.AddMessage("user", query)
						sess.AddMessage("assistant", fmt.Sprintf("Displayed results as table (chart requires at least 2 columns)."))
						continue
					}
					chartOutput, err := tool.RenderChartString(lastQueryResult, chartType)
					if err != nil {
						ui.ShowWarning(fmt.Sprintf("Cannot render %s chart: %v. Showing table instead.", chartType, err))
						ui.ShowInfo("Displaying table view instead.")
						ui.PrintTable(lastQueryResult.Columns, lastQueryResult.Rows)
					} else {
						title := fmt.Sprintf("Chart (%d rows)", len(lastQueryResult.Rows))
						ui.DisplayChart(chartOutput, chartType, title)
					}
					fmt.Println()
				}
				// Add to conversation history
				sess.AddMessage("user", query)
				sess.AddMessage("assistant", fmt.Sprintf("Displaying results as %s chart.", chartType))
				continue
			}
		}

		// Convert existing session messages to LLM chat messages (before adding current query)
		conversationHistory := make([]llm.ChatMessage, 0)
		for _, msg := range sess.GetHistory() {
			conversationHistory = append(conversationHistory, llm.ChatMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}

		// Prepare schema context (empty for free mode)
		var schemaContext string
		var databaseType string
		if src != nil && schema != nil {
			schemaContext = schema.FormatSchema()
			if schemaContext == "" {
				schemaContext = fmt.Sprintf("Currently connected to database: %s\nNo schema information available yet.", src.Database)
			} else {
				schemaContext = fmt.Sprintf("Currently connected to database: %s\n\n%s", src.Database, schemaContext)
			}
			databaseType = src.GetDatabaseType()
		} else {
			// Free mode: no schema context
			schemaContext = ""
			databaseType = ""
		}

		// Get tool definitions (including built-in tools)
		tools := tool.GetLLMFunctionsWithBuiltin(conn)

		// Create tool handler
		toolHandler := NewToolHandler(conn, skillsManager, llmClient)

		// Use tool calling loop - LLM decides which tools to call
		// Note: "Thinking..." and "Waiting..." messages are handled inside HandleToolCallLoop
		finalResponse, queryResult, err := toolHandler.HandleToolCallLoop(ctx, llmClient, query, schemaContext, databaseType, conversationHistory, tools)

		if err != nil {
			ui.ShowError(fmt.Sprintf("Failed to process request: %v", err))
			ui.ShowInfo("Please check your LLM configuration and try again.")
			fmt.Println()
			continue
		}

		// Add user message to history
		sess.AddMessage("user", query)

		// Store query result for potential view switching
		if queryResult != nil {
			lastQueryResult = queryResult
		}

		// Display final response
		if finalResponse != "" {
			sess.AddMessage("assistant", finalResponse)
			fmt.Println()
			fmt.Println(finalResponse)
			fmt.Println()
		}
	}
}

// displayChart displays query results as a chart
func displayChart(result *db.QueryResult) error {
	// Check for single column result
	if len(result.Columns) == 1 {
		return fmt.Errorf("single column results cannot be visualized as charts")
	}

	// Detect chart type
	detection, err := chart.DetectChartTypeWithColumns(result.Columns, result.Rows)
	if err != nil {
		return fmt.Errorf("chart detection failed: %w", err)
	}

	// Check if chartable
	if detection.Type == chart.ChartTypeTable {
		return fmt.Errorf("data structure not suitable for chart visualization (no numerical data detected)")
	}

	// Check dataset size
	if len(result.Rows) > 1000 {
		ui.ShowWarning(fmt.Sprintf("Large dataset (%d rows). Chart may be slow to render.", len(result.Rows)))
		proceed, _ := ui.ShowConfirm("Continue with chart rendering?")
		if !proceed {
			return fmt.Errorf("chart rendering cancelled")
		}
	}

	// Get available chart types using detector
	availableTypes := chart.GetAvailableChartTypes(result.Columns, result.Rows)
	if len(availableTypes) == 0 {
		return fmt.Errorf("no suitable chart types available for this data")
	}

	// Convert to menu items for display
	availableTypesMenu := make([]ui.MenuItem, len(availableTypes))
	for i, ct := range availableTypes {
		availableTypesMenu[i] = ui.MenuItem{
			Label: fmt.Sprintf("%s - %s", ct, getChartTypeLabel(ct)),
			Value: string(ct),
		}
	}

	// Let user select chart type
	var chartType chart.ChartType
	if len(availableTypesMenu) == 1 {
		// Only one option, use it
		chartType = availableTypes[0]
		ui.ShowInfo(fmt.Sprintf("Using chart type: %s", chartType))
	} else {
		// Multiple options, let user choose
		selected, err := ui.ShowMenu("Select chart type", availableTypesMenu)
		if err != nil {
			return fmt.Errorf("chart type selection cancelled")
		}
		chartType = chart.ChartType(selected)
	}

	// Render chart string and display
	chartOutput, err := tool.RenderChartString(result, string(chartType))
	if err != nil {
		return fmt.Errorf("chart rendering failed: %w", err)
	}
	title := fmt.Sprintf("Chart (%d rows)", len(result.Rows))
	ui.DisplayChart(chartOutput, string(chartType), title)

	return nil
}

// getChartTypeLabel returns a descriptive label for chart type
func getChartTypeLabel(ct chart.ChartType) string {
	switch ct {
	case chart.ChartTypeBar:
		return "Bar chart (categorical vs numerical)"
	case chart.ChartTypeLine:
		return "Line chart (time series)"
	case chart.ChartTypePie:
		return "Pie chart (distribution)"
	case chart.ChartTypeScatter:
		return "Scatter plot (numerical vs numerical)"
	default:
		return "Unknown chart type"
	}
}

// detectChartTypeFromQuery extracts chart type from user query
func detectChartTypeFromQuery(queryLower string) string {
	if strings.Contains(queryLower, "pie") {
		return "pie"
	}
	if strings.Contains(queryLower, "bar") {
		return "bar"
	}
	if strings.Contains(queryLower, "line") {
		return "line"
	}
	if strings.Contains(queryLower, "scatter") {
		return "scatter"
	}
	if strings.Contains(queryLower, "table") {
		return "table"
	}
	return ""
}
