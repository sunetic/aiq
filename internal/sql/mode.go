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
	"github.com/aiq/aiq/internal/source"
	"github.com/aiq/aiq/internal/ui"
)

// RunSQLMode runs the SQL interactive mode
func RunSQLMode() error {
	// Select source
	sources, err := source.LoadSources()
	if err != nil {
		return fmt.Errorf("failed to load sources: %w", err)
	}

	if len(sources) == 0 {
		return fmt.Errorf("no data sources configured. Please add a source first")
	}

	items := make([]ui.MenuItem, 0, len(sources))
	for _, s := range sources {
		label := fmt.Sprintf("%s (%s/%s:%d/%s)", s.Name, s.Type, s.Host, s.Port, s.Database)
		items = append(items, ui.MenuItem{Label: label, Value: s.Name})
	}

	sourceName, err := ui.ShowMenu("Select Data Source", items)
	if err != nil {
		return err
	}

	// Load source
	src, err := source.GetSource(sourceName)
	if err != nil {
		return fmt.Errorf("failed to load source: %w", err)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create database connection
	conn, err := db.NewConnection(src.DSN(), string(src.Type))
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}
	defer conn.Close()

	// Create LLM client
	llmClient := llm.NewClient(cfg.LLM.URL, cfg.LLM.APIKey, cfg.LLM.Model)

	// Fetch schema for context
	ctx := context.Background()
	schema, err := conn.GetSchema(ctx, src.Database)
	if err != nil {
		ui.ShowWarning(fmt.Sprintf("Failed to fetch schema: %v. Continuing without schema context.", err))
		schema = &db.Schema{}
	}

	ui.ShowInfo(fmt.Sprintf("Entering SQL mode. Source: %s (%s/%s:%d/%s)", src.Name, src.Type, src.Host, src.Port, src.Database))
	ui.ShowInfo("Type your question in natural language, or 'exit' to return to main menu.")
	fmt.Println()

	// Use readline for better Unicode/Chinese character support
	rl, err := readline.NewEx(&readline.Config{
		Prompt:          ui.InfoText("aiq> "),
		HistoryFile:     "",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
	})
	if err != nil {
		return fmt.Errorf("failed to initialize readline: %w", err)
	}
	defer rl.Close()

	for {
		// Read input line (readline handles Unicode properly)
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// Ctrl+C - continue to next prompt
				fmt.Println()
				continue
			}
			// EOF (Ctrl+D) - exit SQL mode
			fmt.Println()
			ui.ShowInfo("Exiting SQL mode (EOF).")
			return nil
		}

		line = strings.TrimSpace(line)

		// Handle empty input
		if line == "" {
			continue
		}

		// Use the line directly as query (readline handles multi-byte characters correctly)
		query := line

		// Handle exit command
		if strings.ToLower(query) == "exit" || strings.ToLower(query) == "back" {
			return nil
		}

		// Translate natural language to SQL
		stopLoading := ui.ShowLoading("Translating to SQL...")
		sqlQuery, err := llmClient.TranslateToSQL(ctx, query, schema.FormatSchema(), src.GetDatabaseType())
		stopLoading()

		if err != nil {
			ui.ShowError(fmt.Sprintf("Failed to translate to SQL: %v", err))
			ui.ShowInfo("Please check your LLM configuration and try again.")
			fmt.Println()
			continue
		}

		// Display translated SQL
		fmt.Println()
		ui.ShowInfo("Generated SQL:")
		fmt.Println(ui.HighlightSQL(sqlQuery))
		fmt.Println()

		// Confirm execution
		confirm, err := ui.ShowConfirm("Execute this query?")
		if err != nil {
			fmt.Println()
			continue
		}

		if !confirm {
			ui.ShowInfo("Query cancelled.")
			fmt.Println()
			continue
		}

		// Execute query
		stopLoading = ui.ShowLoading("Executing query...")
		result, err := conn.ExecuteQuery(ctx, sqlQuery)
		stopLoading()

		if err != nil {
			ui.ShowError(fmt.Sprintf("Query execution failed: %v", err))
			ui.ShowInfo("You can modify the query and try again.")
			fmt.Println()
			continue
		}

		// Display results
		fmt.Println()
		if len(result.Rows) == 0 {
			ui.ShowInfo("Query executed successfully. No rows returned.")
			fmt.Println()
			continue
		}

		ui.ShowSuccess(fmt.Sprintf("Query executed successfully. %d row(s) returned.", len(result.Rows)))
		fmt.Println()

		// Prompt for view type
		viewItems := []ui.MenuItem{
			{Label: "table - View as table", Value: "table"},
			{Label: "chart - View as chart", Value: "chart"},
			{Label: "both  - View both table and chart", Value: "both"},
		}

		viewChoice, err := ui.ShowMenu("Select view type", viewItems)
		if err != nil {
			fmt.Println()
			continue
		}

		// Display table if requested
		if viewChoice == "table" || viewChoice == "both" {
			ui.PrintTable(result.Columns, result.Rows)
			fmt.Println()
		}

		// Display chart if requested
		if viewChoice == "chart" || viewChoice == "both" {
			if err := displayChart(result); err != nil {
				ui.ShowWarning(fmt.Sprintf("Failed to render chart: %v", err))
				ui.ShowInfo("Displaying table view instead.")
				ui.PrintTable(result.Columns, result.Rows)
			}
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

	// Create default config
	config := chart.DefaultConfig()
	config.Width = 80
	config.Height = 20
	config.Title = fmt.Sprintf("Query Results (%d rows)", len(result.Rows))

	// Render chart
	chartOutput, err := chart.RenderChart(result, chartType, config)
	if err != nil {
		return fmt.Errorf("chart rendering failed: %w", err)
	}

	// Display chart using UI helper
	ui.DisplayChart(chartOutput, string(chartType), config.Title)

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
