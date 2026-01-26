package chart

import (
	"fmt"
	"strings"

	"github.com/aiq/aiq/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// ScatterChartRenderer renders scatter plots
type ScatterChartRenderer struct{}

// Render renders a scatter plot
func (r *ScatterChartRenderer) Render(result *db.QueryResult, config *Config, xColIndex, yColIndex int) (string, error) {
	if len(result.Columns) < 2 {
		return "", fmt.Errorf("scatter plot requires at least 2 columns")
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("no data to render")
	}

	// Extract data using detected column indices
	xValues, err := extractNumericalColumn(result.Rows, xColIndex)
	if err != nil {
		return "", fmt.Errorf("failed to extract X values: %w", err)
	}

	yValues, err := extractNumericalColumn(result.Rows, yColIndex)
	if err != nil {
		return "", fmt.Errorf("failed to extract Y values: %w", err)
	}

	if len(xValues) != len(yValues) {
		return "", fmt.Errorf("mismatched data length")
	}

	// Find ranges for scaling
	maxX := findMax(xValues)
	minX := findMin(xValues)
	rangeX := maxX - minX
	if rangeX == 0 {
		rangeX = 1
	}

	maxY := findMax(yValues)
	minY := findMin(yValues)
	rangeY := maxY - minY
	if rangeY == 0 {
		rangeY = 1
	}

	// Build chart
	var builder strings.Builder

	// Title
	if config.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(config.Color)
		builder.WriteString(titleStyle.Render(config.Title))
		builder.WriteString("\n\n")
	}

	// Chart area
	chartHeight := config.Height
	chartWidth := config.Width - 20
	if chartWidth < 20 {
		chartWidth = 40
	}

	// Create grid
	grid := make([][]rune, chartHeight)
	for i := range grid {
		grid[i] = make([]rune, chartWidth)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Plot points
	pointChar := '●'
	if !config.UseUnicode {
		pointChar = '*'
	}

	for i := 0; i < len(xValues); i++ {
		normalizedX := (xValues[i] - minX) / rangeX
		normalizedY := (yValues[i] - minY) / rangeY

		x := int(normalizedX * float64(chartWidth-1))
		y := chartHeight - 1 - int(normalizedY*float64(chartHeight-1))

		if x >= 0 && x < chartWidth && y >= 0 && y < chartHeight {
			grid[y][x] = pointChar
		}
	}

	// Render grid
	pointStyle := lipgloss.NewStyle().Foreground(config.Color)
	for i := chartHeight - 1; i >= 0; i-- {
		line := string(grid[i])
		builder.WriteString(pointStyle.Render(line))
		builder.WriteString("\n")
	}

	// Value ranges
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("X Range: %s - %s\n", formatNumber(minX), formatNumber(maxX)))
	builder.WriteString(fmt.Sprintf("Y Range: %s - %s\n", formatNumber(minY), formatNumber(maxY)))

	return builder.String(), nil
}
