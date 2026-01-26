package chart

import (
	"fmt"

	"github.com/aiq/aiq/internal/ui"
	"github.com/charmbracelet/lipgloss"
)

// ColorPalette represents a predefined color scheme
type ColorPalette struct {
	Name  string
	Color lipgloss.Color
}

// PredefinedColorPalettes returns available color palettes
func PredefinedColorPalettes() []ColorPalette {
	return []ColorPalette{
		{Name: "Blue", Color: lipgloss.Color("39")},
		{Name: "Green", Color: lipgloss.Color("2")},
		{Name: "Yellow", Color: lipgloss.Color("3")},
		{Name: "Red", Color: lipgloss.Color("1")},
		{Name: "Magenta", Color: lipgloss.Color("5")},
		{Name: "Cyan", Color: lipgloss.Color("6")},
		{Name: "White", Color: lipgloss.Color("7")},
		{Name: "Gray", Color: lipgloss.Color("8")},
	}
}

// CustomizeChart prompts the user to customize chart settings
func CustomizeChart(defaultConfig *Config, availableTypes []ChartType, columns []string) (*Config, ChartType, error) {
	config := *defaultConfig // Copy default config

	// Ask if user wants to customize
	customizeItems := []ui.MenuItem{
		{Label: "yes - Customize chart settings", Value: "yes"},
		{Label: "no  - Use default settings", Value: "no"},
	}
	customize, err := ui.ShowMenu("Customize chart settings?", customizeItems)
	if err != nil {
		return nil, "", fmt.Errorf("customization cancelled: %w", err)
	}

	if customize == "no" {
		// Use default chart type (first available)
		if len(availableTypes) > 0 {
			return &config, availableTypes[0], nil
		}
		return &config, ChartTypeTable, nil
	}

	// 1. Chart type selection
	typeItems := make([]ui.MenuItem, len(availableTypes))
	for i, ct := range availableTypes {
		typeItems[i] = ui.MenuItem{
			Label: fmt.Sprintf("%s - %s chart", ct, getChartTypeDescription(ct)),
			Value: string(ct),
		}
	}
	selectedType, err := ui.ShowMenu("Select chart type", typeItems)
	if err != nil {
		return nil, "", fmt.Errorf("chart type selection cancelled: %w", err)
	}
	config.Type = ChartType(selectedType)

	// 2. Color scheme selection
	colorItems := make([]ui.MenuItem, len(PredefinedColorPalettes()))
	for i, palette := range PredefinedColorPalettes() {
		// Create a colored block character for preview
		colorPreview := lipgloss.NewStyle().Foreground(palette.Color).Render("█")
		colorItems[i] = ui.MenuItem{
			Label: fmt.Sprintf("%s - %s", palette.Name, colorPreview),
			Value: string(palette.Color),
		}
	}
	selectedColor, err := ui.ShowMenu("Select color scheme", colorItems)
	if err != nil {
		return nil, "", fmt.Errorf("color selection cancelled: %w", err)
	}
	config.Color = lipgloss.Color(selectedColor)

	// 3. Chart title
	title, err := ui.ShowInput("Chart title (press Enter for default)", "")
	if err != nil {
		return nil, "", fmt.Errorf("title input cancelled: %w", err)
	}
	if title != "" {
		config.Title = title
	}

	// 4. X-axis label (if applicable)
	if config.Type != ChartTypePie {
		xLabel, err := ui.ShowInput(fmt.Sprintf("X-axis label (press Enter for default: %s)", columns[0]), "")
		if err != nil {
			return nil, "", fmt.Errorf("X-axis label input cancelled: %w", err)
		}
		if xLabel != "" {
			config.XLabel = xLabel
		} else if len(columns) > 0 {
			config.XLabel = columns[0]
		}
	}

	// 5. Y-axis label (if applicable)
	if config.Type != ChartTypePie {
		yLabel, err := ui.ShowInput(fmt.Sprintf("Y-axis label (press Enter for default: %s)", columns[1]), "")
		if err != nil {
			return nil, "", fmt.Errorf("Y-axis label input cancelled: %w", err)
		}
		if yLabel != "" {
			config.YLabel = yLabel
		} else if len(columns) > 1 {
			config.YLabel = columns[1]
		}
	}

	return &config, config.Type, nil
}

// getChartTypeDescription returns a description for the chart type
func getChartTypeDescription(ct ChartType) string {
	switch ct {
	case ChartTypeBar:
		return "Bar chart for categorical vs numerical data"
	case ChartTypeLine:
		return "Line chart for time series or sequential data"
	case ChartTypePie:
		return "Pie chart for proportional categorical data"
	case ChartTypeScatter:
		return "Scatter plot for numerical vs numerical data"
	default:
		return "Unknown chart type"
	}
}
