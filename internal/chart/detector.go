package chart

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// ChartDetectionResult contains the detected chart type and column indices
type ChartDetectionResult struct {
	Type      ChartType
	XColIndex int // Column index for X-axis
	YColIndex int // Column index for Y-axis
}

// DetectChartType detects the appropriate chart type based on query result structure
func DetectChartType(columns []string, rows [][]string) (ChartType, error) {
	result, err := DetectChartTypeWithColumns(columns, rows)
	if err != nil {
		return ChartTypeTable, err
	}
	return result.Type, nil
}

// DetectChartTypeWithColumns detects chart type and returns column indices
func DetectChartTypeWithColumns(columns []string, rows [][]string) (*ChartDetectionResult, error) {
	if len(columns) == 0 {
		return &ChartDetectionResult{Type: ChartTypeTable}, fmt.Errorf("no columns in result")
	}

	if len(rows) == 0 {
		return &ChartDetectionResult{Type: ChartTypeTable}, fmt.Errorf("no rows in result")
	}

	// Single column - table only
	if len(columns) == 1 {
		return &ChartDetectionResult{Type: ChartTypeTable}, nil
	}

	// Two columns - check for bar, line, or pie chart
	if len(columns) == 2 {
		return detectTwoColumnChart(columns, rows)
	}

	// Three or more columns - try to find suitable column combination
	return detectMultiColumnChart(columns, rows)
}

// detectTwoColumnChart detects chart type for 2-column results
func detectTwoColumnChart(columns []string, rows [][]string) (*ChartDetectionResult, error) {
	firstColType := DetectColumnType(columns[0], rows, 0)
	secondColType := DetectColumnType(columns[1], rows, 1)

	// Bar chart: categorical (string) + numerical
	if firstColType == ColumnTypeCategorical && secondColType == ColumnTypeNumerical {
		// Check if suitable for pie chart (< 10 categories)
		uniqueCategories := countUniqueValues(rows, 0)
		if uniqueCategories < 10 && uniqueCategories > 1 {
			return &ChartDetectionResult{Type: ChartTypePie, XColIndex: 0, YColIndex: 1}, nil
		}
		return &ChartDetectionResult{Type: ChartTypeBar, XColIndex: 0, YColIndex: 1}, nil
	}

	// Bar chart: numerical + categorical (reverse order)
	if firstColType == ColumnTypeNumerical && secondColType == ColumnTypeCategorical {
		uniqueCategories := countUniqueValues(rows, 1)
		if uniqueCategories < 10 && uniqueCategories > 1 {
			return &ChartDetectionResult{Type: ChartTypePie, XColIndex: 1, YColIndex: 0}, nil
		}
		return &ChartDetectionResult{Type: ChartTypeBar, XColIndex: 1, YColIndex: 0}, nil
	}

	// Line chart: temporal/sequential + numerical
	if (firstColType == ColumnTypeTemporal || firstColType == ColumnTypeSequential) && secondColType == ColumnTypeNumerical {
		return &ChartDetectionResult{Type: ChartTypeLine, XColIndex: 0, YColIndex: 1}, nil
	}

	// Scatter plot: numerical + numerical
	if firstColType == ColumnTypeNumerical && secondColType == ColumnTypeNumerical {
		return &ChartDetectionResult{Type: ChartTypeScatter, XColIndex: 0, YColIndex: 1}, nil
	}

	return &ChartDetectionResult{Type: ChartTypeTable}, nil
}

// detectMultiColumnChart detects chart type for 3+ column results
func detectMultiColumnChart(columns []string, rows [][]string) (*ChartDetectionResult, error) {
	// Find first categorical column and first numerical column
	var catColIdx, numColIdx = -1, -1
	var numColIdx2 = -1 // For scatter plot
	var temporalColIdx = -1

		for i := 0; i < len(columns); i++ {
			colType := DetectColumnType(columns[i], rows, i)
		if colType == ColumnTypeCategorical && catColIdx == -1 {
			catColIdx = i
		}
		if colType == ColumnTypeNumerical {
			if numColIdx == -1 {
				numColIdx = i
			} else if numColIdx2 == -1 {
				numColIdx2 = i
			}
		}
		if (colType == ColumnTypeTemporal || colType == ColumnTypeSequential) && temporalColIdx == -1 {
			temporalColIdx = i
		}
	}

	// Bar/Pie chart: categorical + numerical (any order)
	if catColIdx != -1 && numColIdx != -1 {
		uniqueCategories := countUniqueValues(rows, catColIdx)
		if uniqueCategories < 10 && uniqueCategories > 1 {
			return &ChartDetectionResult{Type: ChartTypePie, XColIndex: catColIdx, YColIndex: numColIdx}, nil
		}
		return &ChartDetectionResult{Type: ChartTypeBar, XColIndex: catColIdx, YColIndex: numColIdx}, nil
	}

	// Line chart: temporal/sequential + numerical
	if temporalColIdx != -1 && numColIdx != -1 {
		return &ChartDetectionResult{Type: ChartTypeLine, XColIndex: temporalColIdx, YColIndex: numColIdx}, nil
	}

	// Scatter plot: two numerical columns
	if numColIdx != -1 && numColIdx2 != -1 {
		return &ChartDetectionResult{Type: ChartTypeScatter, XColIndex: numColIdx, YColIndex: numColIdx2}, nil
	}

	return &ChartDetectionResult{Type: ChartTypeTable}, nil
}

// ColumnType represents the type of a column
type ColumnType string

const (
	ColumnTypeCategorical ColumnType = "categorical" // String values
	ColumnTypeNumerical  ColumnType = "numerical"    // Numbers
	ColumnTypeTemporal   ColumnType = "temporal"     // Dates/times
	ColumnTypeSequential ColumnType = "sequential"   // Sequential numbers
)

// DetectColumnType detects the type of a column based on sample data
// This is exported so it can be used by other packages to determine available chart types
func DetectColumnType(columnName string, rows [][]string, colIndex int) ColumnType {
	if len(rows) == 0 {
		return ColumnTypeCategorical
	}

	// Check column name for aggregation functions (COUNT, SUM, AVG, MAX, MIN, etc.)
	columnNameLower := strings.ToLower(columnName)
	if strings.Contains(columnNameLower, "count") || 
	   strings.Contains(columnNameLower, "sum") ||
	   strings.Contains(columnNameLower, "avg") ||
	   strings.Contains(columnNameLower, "average") ||
	   strings.Contains(columnNameLower, "max") ||
	   strings.Contains(columnNameLower, "min") ||
	   strings.Contains(columnNameLower, "total") {
		// If column name suggests it's an aggregation result, check if values are numeric
		if len(rows) > 0 {
			// Check first few values to confirm
			allNumeric := true
			checkSize := 5
			if len(rows) < checkSize {
				checkSize = len(rows)
			}
			for i := 0; i < checkSize; i++ {
				value := strings.TrimSpace(rows[i][colIndex])
				if value != "" && value != "NULL" && !isNumerical(value) {
					allNumeric = false
					break
				}
			}
			if allNumeric {
				return ColumnTypeNumerical
			}
		}
	}

	// Sample first 10 rows (or all if less than 10)
	sampleSize := 10
	if len(rows) < sampleSize {
		sampleSize = len(rows)
	}

	// Check for temporal patterns
	temporalCount := 0
	numericalCount := 0
	sequentialCount := 0

	for i := 0; i < sampleSize; i++ {
		value := strings.TrimSpace(rows[i][colIndex])
		if value == "" || value == "NULL" {
			continue
		}

		// Check if it's a date/time
		if isTemporal(value) {
			temporalCount++
			continue
		}

		// Check if it's a number
		if isNumerical(value) {
			numericalCount++
			// Check if it's sequential (increasing)
			if i > 0 && isSequential(rows[i-1][colIndex], value) {
				sequentialCount++
			}
			continue
		}
	}

	// Determine type based on counts
	// Lower threshold for numerical detection (at least 30% should be numeric)
	if temporalCount > sampleSize/2 {
		return ColumnTypeTemporal
	}

	if numericalCount >= sampleSize*3/10 || numericalCount >= 2 {
		// If we have at least 30% numeric or at least 2 numeric values, consider it numerical
		if sequentialCount > sampleSize/3 {
			return ColumnTypeSequential
		}
		return ColumnTypeNumerical
	}

	return ColumnTypeCategorical
}

// isTemporal checks if a string represents a date/time
func isTemporal(value string) bool {
	// Common date/time formats
	formats := []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05.000000",
	}

	for _, format := range formats {
		if _, err := time.Parse(format, value); err == nil {
			return true
		}
	}

	return false
}

// isNumerical checks if a string represents a number
func isNumerical(value string) bool {
	// Try integer
	if _, err := strconv.ParseInt(value, 10, 64); err == nil {
		return true
	}

	// Try float
	if _, err := strconv.ParseFloat(value, 64); err == nil {
		return true
	}

	return false
}

// isSequential checks if two values form a sequential pattern
func isSequential(prev, curr string) bool {
	prevNum, err1 := strconv.ParseFloat(strings.TrimSpace(prev), 64)
	currNum, err2 := strconv.ParseFloat(strings.TrimSpace(curr), 64)

	if err1 != nil || err2 != nil {
		return false
	}

	// Check if current is greater than previous
	return currNum > prevNum
}

// countUniqueValues counts unique values in a column
func countUniqueValues(rows [][]string, colIndex int) int {
	unique := make(map[string]bool)
	for _, row := range rows {
		if colIndex < len(row) {
			value := strings.TrimSpace(row[colIndex])
			if value != "" && value != "NULL" {
				unique[value] = true
			}
		}
	}
	return len(unique)
}
