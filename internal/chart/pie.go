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
	radius := 10 // Radius of the pie chart
	centerX := radius + 2
	centerY := radius + 1

	// Create a grid for the pie chart
	grid := make([][]rune, radius*2+3)
	for i := range grid {
		grid[i] = make([]rune, radius*2+5)
		for j := range grid[i] {
			grid[i][j] = ' '
		}
	}

	// Calculate angles for each segment
	cumulativeAngle := 0.0
	colors := []string{"█", "▓", "▒", "░", "▄", "▀", "▌", "▐", "▔", "▕"}
	
	for idx, item := range displayItems {
		angle := (item.Percentage / 100.0) * 360.0
		endAngle := cumulativeAngle + angle
		
		// Draw pie slice
		char := colors[idx%len(colors)]
		if !config.UseUnicode {
			char = "#"
		}
		
		drawPieSlice(grid, centerX, centerY, radius, cumulativeAngle, endAngle, []rune(char)[0])
		
		cumulativeAngle = endAngle
	}

	// Render the pie chart grid
	pieStyle := lipgloss.NewStyle().Foreground(config.Color)
	for y := 0; y < len(grid); y++ {
		line := string(grid[y])
		builder.WriteString(pieStyle.Render(line))
		builder.WriteString("\n")
	}

	// Legend
	builder.WriteString("\n")
	builder.WriteString("Legend:\n")
	for idx, item := range displayItems {
		char := colors[idx%len(colors)]
		if !config.UseUnicode {
			char = "#"
		}
		category := truncateString(item.Category, 20)
		legendStyle := lipgloss.NewStyle().Foreground(config.Color)
		legendChar := legendStyle.Render(char)
		line := fmt.Sprintf("  %s %-20s │ %5.1f%% │ %s\n", 
			legendChar, category, item.Percentage, formatNumber(item.Value))
		builder.WriteString(line)
	}

	// Summary
	builder.WriteString("\n")
	builder.WriteString(fmt.Sprintf("Total: %s (100.0%%)\n", formatNumber(total)))
	if len(items) > maxDisplayItems {
		builder.WriteString(fmt.Sprintf("Items: %d total (showing top %d + others)\n", len(items), maxDisplayItems))
	}

	return builder.String(), nil
}

// drawPieSlice draws a pie slice in the grid
func drawPieSlice(grid [][]rune, centerX, centerY, radius int, startAngle, endAngle float64, char rune) {
	// Convert angles to radians
	startRad := startAngle * math.Pi / 180.0
	endRad := endAngle * math.Pi / 180.0

	// Draw the slice by checking each point in the grid
	for y := 0; y < len(grid); y++ {
		for x := 0; x < len(grid[y]); x++ {
			// Calculate distance from center
			dx := float64(x - centerX)
			dy := float64(y - centerY)
			distance := math.Sqrt(dx*dx + dy*dy)

			// Check if point is within radius
			if distance <= float64(radius) {
				// Calculate angle of this point
				angle := math.Atan2(dy, dx)
				// Normalize angle to 0-2π
				if angle < 0 {
					angle += 2 * math.Pi
				}
				// Normalize startRad and endRad to 0-2π
				startNorm := startRad
				if startNorm < 0 {
					startNorm += 2 * math.Pi
				}
				endNorm := endRad
				if endNorm < 0 {
					endNorm += 2 * math.Pi
				}

				// Check if angle is within slice range
				if startNorm <= endNorm {
					if angle >= startNorm && angle <= endNorm {
						grid[y][x] = char
					}
				} else {
					// Handle wrap-around case
					if angle >= startNorm || angle <= endNorm {
						grid[y][x] = char
					}
				}
			}
		}
	}
}
