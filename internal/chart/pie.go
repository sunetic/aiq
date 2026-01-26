package chart

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/aiq/aiq/internal/db"
	"github.com/charmbracelet/lipgloss"
)

// PieChartItem represents a single pie chart segment
type PieChartItem struct {
	Category   string
	Value      float64
	Percentage float64
	AngleStart float64 // Start angle in radians
	AngleEnd   float64 // End angle in radians
	Color      lipgloss.Color
}

// PieChartRenderer renders pie charts
type PieChartRenderer struct{}

// Render renders a pie chart
func (r *PieChartRenderer) Render(result *db.QueryResult, config *Config, xColIndex, yColIndex int) (string, error) {
	if len(result.Columns) < 2 {
		return "", fmt.Errorf("pie chart requires at least 2 columns")
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

	// Calculate total
	total := 0.0
	for _, v := range values {
		total += v
	}
	if total == 0 {
		return "", fmt.Errorf("total value is zero")
	}

	// Create pie chart items with percentages
	items := make([]PieChartItem, len(categories))
	for i := 0; i < len(categories); i++ {
		items[i] = PieChartItem{
			Category:   categories[i],
			Value:      values[i],
			Percentage: (values[i] / total) * 100.0,
		}
	}

	// Sort by percentage (descending)
	sort.Slice(items, func(i, j int) bool {
		return items[i].Percentage > items[j].Percentage
	})

	// Limit to top N items, group rest as "Others"
	maxDisplayItems := 8 // Limit for better visualization
	var displayItems []PieChartItem
	var othersTotal float64
	var othersPercentage float64

	if len(items) > maxDisplayItems {
		displayItems = items[:maxDisplayItems]
		for i := maxDisplayItems; i < len(items); i++ {
			othersTotal += items[i].Value
			othersPercentage += items[i].Percentage
		}
		if othersTotal > 0 {
			displayItems = append(displayItems, PieChartItem{
				Category:   fmt.Sprintf("Others (%d)", len(items)-maxDisplayItems),
				Value:      othersTotal,
				Percentage: othersPercentage,
			})
		}
	} else {
		displayItems = items
	}

	// Build chart
	var builder strings.Builder

	// Title
	if config.Title != "" {
		titleStyle := lipgloss.NewStyle().Bold(true).Foreground(config.Color)
		builder.WriteString(titleStyle.Render(config.Title))
		builder.WriteString("\n\n")
	}

	// Render circular pie chart using ASCII art
	chartSize := int(math.Min(float64(config.Height), float64(config.Width/2))) // Make it roughly square
	if chartSize < 10 {
		chartSize = 10
	}
	if chartSize > 20 {
		chartSize = 20 // Limit maximum size
	}
	radius := float64(chartSize / 2)
	centerX := radius
	centerY := radius

	// Create grid for the pie chart
	grid := make([][]rune, chartSize)
	for i := range grid {
		grid[i] = make([]rune, chartSize)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Color palette for different segments
	colorPalette := []lipgloss.Color{
		lipgloss.Color("69"), lipgloss.Color("105"), lipgloss.Color("141"), lipgloss.Color("177"),
		lipgloss.Color("213"), lipgloss.Color("225"), lipgloss.Color("195"), lipgloss.Color("159"),
		lipgloss.Color("123"), lipgloss.Color("87"),
	}

	// Calculate angles for each segment
	currentAngle := 0.0
	for i := range displayItems {
		displayItems[i].AngleStart = currentAngle
		displayItems[i].AngleEnd = currentAngle + (displayItems[i].Percentage / 100.0 * 2 * math.Pi)
		displayItems[i].Color = colorPalette[i%len(colorPalette)]
		currentAngle = displayItems[i].AngleEnd
	}

	// Draw pie slices
	for y := 0; y < chartSize; y++ {
		for x := 0; x < chartSize; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance <= radius {
				angle := math.Atan2(dy, dx)
				if angle < 0 {
					angle += 2 * math.Pi // Normalize angle to 0 to 2*Pi
				}

				// Find which segment this point belongs to
				for _, item := range displayItems {
					if angle >= item.AngleStart && angle < item.AngleEnd {
						grid[y][x] = '█'
						break
					}
				}
			}
		}
	}

	// Render grid with colors
	for y := 0; y < chartSize; y++ {
		for x := 0; x < chartSize; x++ {
			dx := float64(x) - centerX
			dy := float64(y) - centerY
			distance := math.Sqrt(dx*dx + dy*dy)

			if distance <= radius {
				angle := math.Atan2(dy, dx)
				if angle < 0 {
					angle += 2 * math.Pi
				}

				found := false
				for _, item := range displayItems {
					if angle >= item.AngleStart && angle < item.AngleEnd {
						builder.WriteString(lipgloss.NewStyle().Foreground(item.Color).Render(string(grid[y][x])))
						found = true
						break
					}
				}
				if !found {
					builder.WriteRune(grid[y][x])
				}
			} else {
				builder.WriteRune(' ')
			}
		}
		builder.WriteRune('\n')
	}

	// Legend
	builder.WriteString("\n")
	builder.WriteString(lipgloss.NewStyle().Bold(true).Render("Legend:\n"))
	for _, item := range displayItems {
		builder.WriteString(lipgloss.NewStyle().Foreground(item.Color).Render("█ "))
		builder.WriteString(fmt.Sprintf("%-20s %5.1f%% (%s)\n",
			truncateString(item.Category, 20), item.Percentage, formatNumber(item.Value)))
	}

	// Summary
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Total: %s (100.0%%)\n", formatNumber(total)))
	builder.WriteString(fmt.Sprintf("Items: %d (showing top %d)\n", len(items), len(displayItems)))

	return builder.String(), nil
}
