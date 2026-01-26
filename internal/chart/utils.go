package chart

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unicode/utf8"
)

// normalizeData normalizes numerical data to fit within chart dimensions
func normalizeData(values []float64, maxValue float64) []float64 {
	if maxValue == 0 {
		return make([]float64, len(values))
	}

	normalized := make([]float64, len(values))
	for i, v := range values {
		normalized[i] = (v / maxValue) * 100.0 // Scale to 0-100
	}
	return normalized
}

// parseFloat64 safely parses a string to float64
func parseFloat64(s string) (float64, error) {
	s = strings.TrimSpace(s)
	if s == "" || s == "NULL" {
		return 0, fmt.Errorf("empty or NULL value")
	}
	return strconv.ParseFloat(s, 64)
}

// extractNumericalColumn extracts numerical values from a column
func extractNumericalColumn(rows [][]string, colIndex int) ([]float64, error) {
	values := make([]float64, 0, len(rows))
	for _, row := range rows {
		if colIndex >= len(row) {
			continue
		}
		val, err := parseFloat64(row[colIndex])
		if err != nil {
			continue // Skip non-numerical values
		}
		values = append(values, val)
	}
	return values, nil
}

// extractCategoricalColumn extracts categorical values from a column
func extractCategoricalColumn(rows [][]string, colIndex int) []string {
	values := make([]string, 0, len(rows))
	for _, row := range rows {
		if colIndex < len(row) {
			value := strings.TrimSpace(row[colIndex])
			if value != "" && value != "NULL" {
				values = append(values, value)
			}
		}
	}
	return values
}

// findMax finds the maximum value in a slice
func findMax(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	max := values[0]
	for _, v := range values {
		if v > max {
			max = v
		}
	}
	return max
}

// findMin finds the minimum value in a slice
func findMin(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	min := values[0]
	for _, v := range values {
		if v < min {
			min = v
		}
	}
	return min
}

// formatNumber formats a number for display
func formatNumber(n float64) string {
	if n == math.Trunc(n) {
		return fmt.Sprintf("%.0f", n)
	}
	return fmt.Sprintf("%.2f", n)
}

// truncateString truncates a string to max length
func truncateString(s string, maxLen int) string {
	if utf8.RuneCountInString(s) <= maxLen {
		return s
	}
	runes := []rune(s)
	return string(runes[:maxLen-3]) + "..."
}

// detectUnicodeSupport checks if terminal supports Unicode
func detectUnicodeSupport() bool {
	// Check environment variable
	// Most modern terminals support Unicode
	// Can be enhanced with actual terminal capability detection
	return true
}

// sampleData samples data for large datasets
func sampleData[T any](data []T, maxSize int) []T {
	if len(data) <= maxSize {
		return data
	}

	sampled := make([]T, 0, maxSize)
	step := len(data) / maxSize
	if step < 1 {
		step = 1
	}

	for i := 0; i < len(data); i += step {
		sampled = append(sampled, data[i])
		if len(sampled) >= maxSize {
			break
		}
	}

	return sampled
}
