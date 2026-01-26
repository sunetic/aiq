package ui

import (
	"fmt"
)

// DisplayChart displays a chart with proper formatting and spacing
// chartType and title are passed as strings to avoid import cycle
func DisplayChart(chartOutput string, chartType string, title string) {
	fmt.Println()
	
	// Show chart type and title
	if title != "" {
		fmt.Println(InfoText(fmt.Sprintf("Chart Type: %s | Title: %s", chartType, title)))
	} else {
		fmt.Println(InfoText(fmt.Sprintf("Chart Type: %s", chartType)))
	}
	
	fmt.Println()
	
	// Display the chart
	fmt.Println(chartOutput)
	
	fmt.Println()
}

// FormatChartTitle formats a chart title with consistent styling
func FormatChartTitle(title string) string {
	return InfoText("📊 " + title)
}

// FormatChartLegend formats chart legend items
func FormatChartLegend(items []string) string {
	var result string
	for i, item := range items {
		if i > 0 {
			result += " | "
		}
		result += item
	}
	return Secondary.Render(result)
}
