package chart

import (
	"fmt"
	"strings"

	"github.com/aiq/aiq/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// BarChartRenderer renders bar charts
type BarChartRenderer struct{}

// Render renders a bar chart
func (r *BarChartRenderer) Render(result *db.QueryResult, config *Config, xColIndex, yColIndex int) (string, error) {
	if len(result.Columns) < 2 {
		return "", fmt.Errorf("bar chart requires at least 2 columns")
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("no data to render")
	}

	// Extract data using detected column indices
	categories := extractCategoricalColumn(result.Rows, xColIndex)
	values, err := extractNumericalColumn(result.Rows, yColIndex)
	if err != nil {
		return "", fmt.Errorf("failed to extract numerical values: %w", err)
	}

	if len(categories) != len(values) {
		return "", fmt.Errorf("mismatched data length")
	}

	// Limit data points for display
	maxBars := config.Width / 3 // Approximate space per bar
	if len(categories) > maxBars {
		categories = categories[:maxBars]
		values = values[:maxBars]
	}

	// Find max value for scaling
	maxValue := findMax(values)
	if maxValue == 0 {
		return "", fmt.Errorf("all values are zero")
	}

	// Build chart
	var builder strings.Builder

	// Title
	if config.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(config.Color)
		builder.WriteString(titleStyle.Render(config.Title))
		builder.WriteString("\n\n")
	}

	// Chart bars (horizontal)
	chartWidth := config.Width - 20 // Reserve space for labels
	if chartWidth < 10 {
		chartWidth = 40 // Minimum width
	}

	for i := 0; i < len(categories); i++ {
		category := truncateString(categories[i], 15)
		value := values[i]
		barLength := int((value / maxValue) * float64(chartWidth))

		// Bar character
		barChar := "█"
		if !config.UseUnicode {
			barChar = "#"
		}

		bar := strings.Repeat(barChar, barLength)
		if barLength == 0 && value > 0 {
			bar = "▏" // Small bar for tiny values
			if !config.UseUnicode {
				bar = "|"
			}
		}

		barStyle := lipgloss.NewStyle().Foreground(config.Color)
		formattedBar := barStyle.Render(bar)

		// Format line: category | bar | value
		line := fmt.Sprintf("%-15s │ %s %s\n", category, formattedBar, formatNumber(value))
		builder.WriteString(line)
	}

	// Labels
	if config.XLabel != "" || config.YLabel != "" {
		builder.WriteString("\n")
		if config.XLabel != "" {
			builder.WriteString(fmt.Sprintf("X: %s", config.XLabel))
		}
		if config.YLabel != "" {
			if config.XLabel != "" {
				builder.WriteString("  ")
			}
			builder.WriteString(fmt.Sprintf("Y: %s", config.YLabel))
		}
		builder.WriteString("\n")
	}

	return builder.String(), nil
}
