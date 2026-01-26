package chart

import "github.com/charmbracelet/lipgloss"

// ChartType represents the type of chart
type ChartType string

const (
	ChartTypeBar     ChartType = "bar"
	ChartTypeLine    ChartType = "line"
	ChartTypePie     ChartType = "pie"
	ChartTypeScatter ChartType = "scatter"
	ChartTypeTable   ChartType = "table" // Fallback to table view
)

// Config represents chart configuration
type Config struct {
	Type      ChartType
	Title     string
	XLabel    string
	YLabel    string
	Color     lipgloss.Color // Keep lipgloss.Color for compatibility
	Width     int
	Height    int
	UseUnicode bool
}


// DefaultConfig returns a default chart configuration
func DefaultConfig() *Config {
	return &Config{
		Type:       ChartTypeTable,
		Title:      "",
		XLabel:     "",
		YLabel:     "",
		Color:      lipgloss.Color("39"), // Default blue
		Width:      80,
		Height:     20,
		UseUnicode: true,
	}
}
