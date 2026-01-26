package chart

import (
	"fmt"
	"strings"

	"github.com/aiq/aiq/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// LineChartRenderer renders line charts
type LineChartRenderer struct{}

// Render renders a line chart
func (r *LineChartRenderer) Render(result *db.QueryResult, config *Config, xColIndex, yColIndex int) (string, error) {
	if len(result.Columns) < 2 {
		return "", fmt.Errorf("line chart requires at least 2 columns")
	}

	if len(result.Rows) == 0 {
		return "", fmt.Errorf("no data to render")
	}

	// Extract data using detected column indices
	xLabels := extractCategoricalColumn(result.Rows, xColIndex)
	values, err := extractNumericalColumn(result.Rows, yColIndex)
	if err != nil {
		return "", fmt.Errorf("failed to extract numerical values: %w", err)
	}

	if len(xLabels) != len(values) {
		return "", fmt.Errorf("mismatched data length")
	}

	// Limit data points
	maxPoints := config.Width - 10
	if len(xLabels) > maxPoints {
		// Sample data
		step := len(xLabels) / maxPoints
		sampledLabels := make([]string, 0, maxPoints)
		sampledValues := make([]float64, 0, maxPoints)
		for i := 0; i < len(xLabels); i += step {
			sampledLabels = append(sampledLabels, xLabels[i])
			sampledValues = append(sampledValues, values[i])
		}
		xLabels = sampledLabels
		values = sampledValues
	}

	// Find min/max for scaling
	maxValue := findMax(values)
	minValue := findMin(values)
	rangeValue := maxValue - minValue
	if rangeValue == 0 {
		rangeValue = 1 // Avoid division by zero
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
	lineChar := '─'
	if !config.UseUnicode {
		pointChar = '*'
		lineChar = '-'
	}

	for i := 0; i < len(values); i++ {
		x := int(float64(i) / float64(len(values)-1) * float64(chartWidth-1))
		if len(values) == 1 {
			x = chartWidth / 2
		}

		normalizedValue := (values[i] - minValue) / rangeValue
		y := chartHeight - 1 - int(normalizedValue*float64(chartHeight-1))

		if y < 0 {
			y = 0
		}
		if y >= chartHeight {
			y = chartHeight - 1
		}

		// Draw point
		if x < chartWidth && y < chartHeight {
			grid[y][x] = pointChar

			// Draw line to next point
			if i < len(values)-1 {
				nextX := int(float64(i+1) / float64(len(values)-1) * float64(chartWidth-1))
				if len(values) == 1 {
					nextX = chartWidth / 2
				}
				nextNormalizedValue := (values[i+1] - minValue) / rangeValue
				nextY := chartHeight - 1 - int(nextNormalizedValue*float64(chartHeight-1))

				// Simple line drawing
				drawLine(grid, x, y, nextX, nextY, lineChar)
			}
		}
	}

	// Render grid
	lineStyle := lipgloss.NewStyle().Foreground(config.Color)
	for i := chartHeight - 1; i >= 0; i-- {
		line := string(grid[i])
		builder.WriteString(lineStyle.Render(line))
		builder.WriteString("\n")
	}

	// X-axis labels (sample)
	builder.WriteString("\n")
	if len(xLabels) > 0 {
		firstLabel := truncateString(xLabels[0], 10)
		lastLabel := truncateString(xLabels[len(xLabels)-1], 10)
		builder.WriteString(fmt.Sprintf("%s ... %s\n", firstLabel, lastLabel))
	}

	// Value range
	builder.WriteString(fmt.Sprintf("Range: %s - %s\n", formatNumber(minValue), formatNumber(maxValue)))

	return builder.String(), nil
}

// drawLine draws a simple line between two points
func drawLine(grid [][]rune, x1, y1, x2, y2 int, char rune) {
	dx := x2 - x1
	dy := y2 - y1
	steps := abs(dx)
	if abs(dy) > steps {
		steps = abs(dy)
	}
	if steps == 0 {
		return
	}

	for i := 0; i <= steps; i++ {
		x := x1 + (dx*i)/steps
		y := y1 + (dy*i)/steps
		if x >= 0 && x < len(grid[0]) && y >= 0 && y < len(grid) {
			if grid[y][x] == ' ' {
				grid[y][x] = char
			}
		}
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
