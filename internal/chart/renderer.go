package chart

import (
	"fmt"

	"github.com/aiq/aiq/internal/db"
)

// Renderer is the interface for chart rendering
type Renderer interface {
	Render(result *db.QueryResult, config *Config, xColIndex, yColIndex int) (string, error)
}

// RenderChart renders a chart based on the specified type
func RenderChart(result *db.QueryResult, chartType ChartType, config *Config) (string, error) {
	// Detect chart type with column indices to get column mapping
	detection, err := DetectChartTypeWithColumns(result.Columns, result.Rows)
	if err != nil {
		return "", err
	}

	// Use user-specified chart type, but keep the detected column indices
	// This allows users to override the auto-detected type while using correct columns
	if config == nil {
		config = DefaultConfig()
	}
	config.Type = chartType

	// For pie chart, we still need categorical + numerical columns
	// Use the detected column indices which should work for pie chart too
	xColIndex := detection.XColIndex
	yColIndex := detection.YColIndex

	// If user selected a different type, we may need to adjust column indices
	// For now, use the detected indices as they should work for most cases
	var renderer Renderer
	switch chartType {
	case ChartTypeBar:
		renderer = &BarChartRenderer{}
	case ChartTypeLine:
		renderer = &LineChartRenderer{}
	case ChartTypePie:
		renderer = &PieChartRenderer{}
	case ChartTypeScatter:
		renderer = &ScatterChartRenderer{}
	default:
		return "", fmt.Errorf("unsupported chart type: %s", chartType)
	}

	return renderer.Render(result, config, xColIndex, yColIndex)
}
