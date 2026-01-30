package sql

import (
	"context"
	"errors"
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

// ErrReturnToMenu is returned when user exits chat mode to return to main menu
var ErrReturnToMenu = errors.New("return to main menu")

// readlineWithHint reads a line from readline and shows hint when user types "/"
// Since readline doesn't support real-time hints, we show hint when user presses Enter with incomplete command
func readlineWithHint(rl *readline.Instance, buildPrompt func() string, commands []string, descriptions map[string]string) (string, error) {
	line, err := rl.Readline()
	if err != nil {
		return "", err
	}

	// If user typed just "/" or a partial command starting with "/", show hint
	trimmed := strings.TrimSpace(line)
	if trimmed == "/" {
		// User typed just "/", show all available commands as hint
		var hints []string
		for _, cmd := range commands {
			desc := descriptions[cmd]
			if desc != "" {
				hints = append(hints, fmt.Sprintf("%s - %s", cmd, desc))
			} else {
				hints = append(hints, cmd)
			}
		}
		if len(hints) > 0 {
			fmt.Print(ui.HintText("  Available commands: " + strings.Join(hints, ", ") + "\n"))
			fmt.Print(buildPrompt())
			// Read the rest of the input
			rest, err := rl.Readline()
			if err != nil {
				return "", err
			}
			return "/" + rest, nil
		}
	} else if strings.HasPrefix(trimmed, "/") && len(trimmed) > 1 {
		// User typed partial command, check if it matches any command
		matched := false
		for _, cmd := range commands {
			if trimmed == cmd {
				matched = true
				break
			}
		}
		if !matched {
			// Show matching commands as hint
			var hints []string
			for _, cmd := range commands {
				if strings.HasPrefix(cmd, trimmed) {
					desc := descriptions[cmd]
					if desc != "" {
						hints = append(hints, fmt.Sprintf("%s - %s", cmd, desc))
					} else {
						hints = append(hints, cmd)
					}
				}
			}
			if len(hints) > 0 {
				fmt.Print(ui.HintText("  Did you mean: " + strings.Join(hints, ", ") + "\n"))
			}
		}
	}

	return line, nil
}

// readMultiLineInput reads multi-line input until completion
// Supports:
// - System commands (single line only, immediate execution)
// - Natural language queries (single line, immediate execution on Enter)
// - SQL queries (multi-line until semicolon or Ctrl+D)
// - Pasting multi-line SQL (detected by checking if next line arrives quickly)
func readMultiLineInput(rl *readline.Instance, buildPrompt func() string, commands []string, descriptions map[string]string) (string, error) {
	// Read first line
	firstLine, err := readlineWithHint(rl, buildPrompt, commands, descriptions)
	if err != nil {
		return "", err
	}

	trimmed := strings.TrimSpace(firstLine)

	// System commands are always single line, execute immediately
	if strings.HasPrefix(trimmed, "/") {
		return firstLine, nil
	}

	// Check if first line ends with semicolon (SQL query complete)
	// If so, return immediately (supports single-line SQL)
	if strings.HasSuffix(trimmed, ";") {
		return firstLine, nil
	}

	// Check if input contains SQL-like patterns (SELECT, FROM, etc.)
	// Only enter multi-line mode for SQL queries
	upperLine := strings.ToUpper(trimmed)
	isSQLLike := strings.Contains(upperLine, "SELECT") ||
		strings.Contains(upperLine, "INSERT") ||
		strings.Contains(upperLine, "UPDATE") ||
		strings.Contains(upperLine, "DELETE") ||
		strings.Contains(upperLine, "CREATE") ||
		strings.Contains(upperLine, "ALTER") ||
		strings.Contains(upperLine, "DROP") ||
		strings.Contains(upperLine, "FROM") ||
		strings.Contains(upperLine, "WHERE") ||
		strings.Contains(upperLine, "JOIN")

	if isSQLLike {
		// SQL-like input: continue reading until semicolon or Ctrl+D
		lines := []string{firstLine}
		for {
			// Show continuation prompt
			fmt.Print(ui.HintText("    -> "))
			line, err := rl.Readline()
			if err != nil {
				if err == readline.ErrInterrupt {
					// Ctrl+C - cancel multi-line input
					fmt.Println()
					return "", readline.ErrInterrupt
				}
				// EOF (Ctrl+D) - submit what we have
				return strings.Join(lines, "\n"), nil
			}

			lines = append(lines, line)
			trimmedLine := strings.TrimSpace(line)

			// If line ends with semicolon, input is complete
			if strings.HasSuffix(trimmedLine, ";") {
				return strings.Join(lines, "\n"), nil
			}

			// If empty line, also submit (allows submitting SQL without semicolon)
			if trimmedLine == "" {
				return strings.Join(lines[:len(lines)-1], "\n"), nil
			}
		}
	}

	// Natural language input: single line, execute immediately on Enter
	// This is the default behavior - user presses Enter and query executes
	return firstLine, nil
}

// RunSQLMode runs the SQL interactive mode
// sessionFile is optional path to a session file to restore
func RunSQLMode(sessionFile string) error {
	return RunSQLModeWithSource("", sessionFile, "")
}

// RunSQLModeWithSource runs the SQL interactive mode with a specific source
// providedSourceName is the name of the source to use (empty string means prompt for selection)
// sessionFile is optional path to a session file to restore
// overrideDatabase is optional database name to override source's database for this session only
func RunSQLModeWithSource(providedSourceName string, sessionFile string, overrideDatabase string) error {
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

	// Use provided source name if available (from CLI args)
	if providedSourceName != "" {
		sourceName = providedSourceName
	}

	// Select source (if not restored from session and not provided) - now optional for free mode
	if sess == nil && sourceName == "" {
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
		// If overrideDatabase is provided, create a temporary source copy with overridden database
		actualSource := src
		if overrideDatabase != "" {
			// Create a copy of the source with overridden database
			tempSource := *src
			tempSource.Database = overrideDatabase
			actualSource = &tempSource
		}
		conn, err = db.NewConnection(actualSource.DSN(), string(actualSource.Type))
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		// Fetch schema for context (use actualSource.Database which may be overridden)
		schema, err = conn.GetSchema(ctx, actualSource.Database)
		if err != nil {
			ui.ShowWarning(fmt.Sprintf("Failed to fetch schema: %v. Continuing without schema context.", err))
			schema = &db.Schema{}
		}
	}

	// Initialize Skills manager
	skillsManager := skills.NewManager()
	if err := skillsManager.Initialize(); err != nil {
		ui.ShowWarning(fmt.Sprintf("Failed to initialize Skills manager: %v. Continuing without Skills.", err))
	}

	// Create LLM client
	llmClient := llm.NewClient(cfg.LLM.URL, cfg.LLM.APIKey, cfg.LLM.Model)

	// Show mode info
	if src != nil {
		// Use actualDatabase which may be overridden by -D parameter
		dbDisplay := ""
		if overrideDatabase != "" {
			dbDisplay = overrideDatabase
		} else if src.Database != "" {
			dbDisplay = src.Database
		}
		if dbDisplay != "" {
			ui.ShowInfo(fmt.Sprintf("Entering chat mode. Source: %s | Database: %s", ui.HighlightText(src.Name), ui.SuccessText(dbDisplay)))
		} else {
			ui.ShowInfo(fmt.Sprintf("Entering chat mode. Source: %s", ui.HighlightText(src.Name)))
		}
	} else {
		ui.ShowInfo("Entering free mode (general conversation and Skills only, no SQL execution)")
	}
	if len(sess.Messages) > 0 {
		ui.ShowInfo(fmt.Sprintf("Conversation history: %d messages", len(sess.Messages)))
	}
	// Show Skills status - simplified, only show names
	metadata := skillsManager.GetMetadata()
	if len(metadata) > 0 {
		skillNames := make([]string, 0, len(metadata))
		for _, md := range metadata {
			skillNames = append(skillNames, ui.HighlightText(md.Name))
		}
		fmt.Print(ui.InfoText("Skills: "))
		fmt.Print(strings.Join(skillNames, ", "))
		fmt.Print(ui.HintText(" (dynamic loading enabled)"))
		fmt.Println()
	} else {
		fmt.Print(ui.InfoText("Skills: "))
		fmt.Print(ui.HintText("No skills found"))
		fmt.Println()
	}
	fmt.Print(ui.HintText("Tip: Use '/help' for commands, ask questions in natural language"))
	fmt.Println()
	fmt.Println()

	// Store last generated SQL for execute command
	var lastGeneratedSQL string

	// Determine actual database being used (may be overridden)
	actualDatabase := ""
	if src != nil {
		if overrideDatabase != "" {
			actualDatabase = overrideDatabase
		} else {
			actualDatabase = src.Database
		}
	}

	// Build dynamic prompt based on source availability and actual database
	// Use different separators/colors to distinguish source and database
	var buildPrompt func() string
	if src != nil {
		if actualDatabase != "" {
			buildPrompt = func() string {
				// Use @ to separate source and database for better distinction
				return ui.InfoText(fmt.Sprintf("aiq[%s@%s]> ", src.Name, actualDatabase))
			}
		} else {
			buildPrompt = func() string {
				return ui.InfoText(fmt.Sprintf("aiq[%s]> ", src.Name))
			}
		}
	} else {
		buildPrompt = func() string {
			return ui.InfoText("aiq> ")
		}
	}

	// Define available commands for hint display
	commands := []string{"/exit", "/help", "/history", "/clear", "/paste"}
	commandDescriptions := map[string]string{
		"/exit":    "Exit chat mode",
		"/help":    "Show help",
		"/history": "View history",
		"/clear":   "Clear history",
		"/paste":   "Enter paste mode for multi-line SQL",
	}

	// Define command completer for Tab completion (only for / commands)
	// Create completer with descriptions for better UX
	completer := readline.NewPrefixCompleter()
	for _, cmd := range commands {
		desc := commandDescriptions[cmd]
		if desc != "" {
			// Add description as a child item for better hint display
			completer.Children = append(completer.Children, readline.PcItem(cmd, readline.PcItem(desc)))
		} else {
			completer.Children = append(completer.Children, readline.PcItem(cmd))
		}
	}

	// Use readline for better Unicode/Chinese character support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          buildPrompt(),
		HistoryFile:     "",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    completer,
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	// Update prompt dynamically (readline doesn't support dynamic prompts directly,
	// but we can recreate it if source changes in future)
	// For now, prompt is set once at initialization

	for {
		// Read multi-line input (supports SQL queries and natural language)
		query, err := readMultiLineInput(rl, buildPrompt, commands, commandDescriptions)
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl+C - continue to next prompt
				fmt.Println()
				continue
			}
			// EOF (Ctrl+D) - exit chat mode (only if no input collected)
			if query == "" {
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
				return ErrReturnToMenu
			}
			// If we have input, submit it (Ctrl+D after typing submits the input)
		}

		query = strings.TrimSpace(query)

		// Handle empty input
		if query == "" {
			continue
		}

		// Handle special commands - check /exit and /help first, then text commands
		if strings.HasPrefix(query, "/") {
			// Handle /exit command
			if strings.ToLower(query) == "/exit" {
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
				return ErrReturnToMenu
			}
			// Handle /help command
			if strings.ToLower(query) == "/help" {
				fmt.Println()
				ui.ShowInfo("Available Commands:")
				fmt.Println()
				fmt.Println("  /exit     - Exit chat mode and return to main menu")
				fmt.Println("  /help     - Show this help message")
				fmt.Println("  /history  - View conversation history")
				fmt.Println("  /clear    - Clear conversation history")
				fmt.Println("  /paste    - Enter paste mode for multi-line SQL (press Ctrl+D to finish)")
				fmt.Println()
				fmt.Println("You can also ask questions in natural language to query the database.")
				fmt.Println()
				continue
			}

			// Handle /paste command - enter multi-line paste mode
			if strings.ToLower(query) == "/paste" {
				fmt.Println()
				ui.ShowInfo("Paste mode: Paste your multi-line SQL, then press Ctrl+D to submit (or Enter on empty line)")
				fmt.Println()
				var pasteLines []string
				for {
					fmt.Print(ui.HintText("paste> "))
					line, err := rl.Readline()
					if err != nil {
						if err == readline.ErrInterrupt {
							// Ctrl+C - cancel paste mode
							fmt.Println()
							ui.ShowWarning("Paste mode cancelled.")
							fmt.Println()
							break
						}
						// EOF (Ctrl+D) - submit pasted content
						if len(pasteLines) > 0 {
							query = strings.Join(pasteLines, "\n")
							// Remove the /paste command from query processing
							query = strings.TrimSpace(query)
							if query != "" {
								// Process the pasted SQL as a normal query
								break
							}
						}
						fmt.Println()
						ui.ShowWarning("No content pasted.")
						fmt.Println()
						break
					}

					trimmedLine := strings.TrimSpace(line)
					// Empty line signals end of paste (alternative to Ctrl+D)
					if trimmedLine == "" && len(pasteLines) > 0 {
						query = strings.Join(pasteLines, "\n")
						query = strings.TrimSpace(query)
						if query != "" {
							break
						}
					}

					if trimmedLine != "" {
						pasteLines = append(pasteLines, line)
					}
				}

				// If we collected paste content, continue to process it
				if len(pasteLines) > 0 {
					// query is already set above, continue to process it
				} else {
					continue
				}
			}
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

		// Add query result summary to conversation history
		if queryResult != nil {
			// Format result summary for conversation history
			resultSummary := formatQueryResultSummary(queryResult)
			// Append summary to final response so it's included in conversation history
			if finalResponse != "" {
				finalResponse = finalResponse + "\n\n" + resultSummary
			} else {
				finalResponse = resultSummary
			}
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
