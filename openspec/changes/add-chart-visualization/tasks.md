## 1. Chart Package Foundation

- [x] 1.1 Create `internal/chart/` directory structure
- [x] 1.2 Create `internal/chart/config.go` with chart configuration struct (type, colors, labels, title)
- [x] 1.3 Create `internal/chart/detector.go` for chart type detection logic
- [x] 1.4 Implement detection heuristics for bar chart (2 cols: categorical + numerical)
- [x] 1.5 Implement detection heuristics for line chart (2 cols: temporal/sequential + numerical)
- [x] 1.6 Implement detection heuristics for pie chart (2 cols: categorical + numerical, < 10 categories)
- [x] 1.7 Implement detection heuristics for scatter plot (2+ numerical columns)
- [x] 1.8 Add fallback to table view for unrecognized structures

## 2. Chart Rendering Implementation

- [x] 2.1 Create `internal/chart/renderer.go` with base chart rendering interface
- [x] 2.2 Create `internal/chart/bar.go` for bar chart rendering (vertical and horizontal)
- [x] 2.3 Implement bar chart scaling and label formatting
- [x] 2.4 Create `internal/chart/line.go` for line chart rendering
- [x] 2.5 Implement line chart data point plotting and line connection
- [x] 2.6 Create `internal/chart/pie.go` for pie chart rendering
- [x] 2.7 Implement pie chart proportional calculation and Unicode rendering
- [x] 2.8 Create `internal/chart/scatter.go` for scatter plot rendering
- [x] 2.9 Implement scatter plot point plotting with proper scaling

## 3. Chart Rendering Utilities

- [x] 3.1 Create utility functions for data scaling and normalization
- [x] 3.2 Implement Unicode character detection and ASCII fallback
- [x] 3.3 Create color palette management using lipgloss
- [x] 3.4 Implement axis label formatting (numbers, dates, strings)
- [x] 3.5 Add chart title and legend rendering
- [x] 3.6 Implement chart width/height calculation based on terminal size

## 4. SQL Mode Integration

- [x] 4.1 Modify `internal/sql/mode.go` to add chart rendering option after query execution
- [x] 4.2 Add prompt for view type selection: [Table] [Chart] [Both]
- [x] 4.3 Integrate chart type detection with query results
- [x] 4.4 Add chart rendering call when Chart or Both is selected
- [x] 4.5 Handle empty result case (no chart option)
- [x] 4.6 Add return to SQL prompt after chart display

## 5. Chart Customization

- [x] 5.1 Create `internal/chart/customizer.go` for chart customization prompts
- [x] 5.2 Add chart type override option (manual selection)
- [x] 5.3 Implement color scheme selection from predefined palettes
- [x] 5.4 Add chart title input prompt
- [x] 5.5 Add axis label customization prompts
- [x] 5.6 Integrate customization flow into SQL mode

## 6. Edge Cases and Constraints

- [x] 6.1 Add dataset size check (> 1000 points warning)
- [x] 6.2 Implement data sampling option for large datasets
- [x] 6.3 Add terminal Unicode support detection
- [x] 6.4 Implement ASCII-only fallback rendering
- [x] 6.5 Handle single column result (table only)
- [x] 6.6 Validate numerical data availability for charts
- [x] 6.7 Add appropriate error messages for non-chartable data

## 7. UI Enhancements

- [x] 7.1 Create `internal/ui/chart.go` for chart display helpers
- [x] 7.2 Add chart display formatting and spacing
- [x] 7.3 Integrate chart colors with existing color scheme
- [x] 7.4 Add loading indicator for chart rendering (if needed)
- [x] 7.5 Ensure chart display is consistent with table display style

## 8. Testing and Documentation

- [x] 8.1 Test bar chart rendering with various data types
- [x] 8.2 Test line chart rendering with time series data
- [x] 8.3 Test pie chart rendering with categorical data
- [x] 8.4 Test scatter plot rendering with numerical data
- [x] 8.5 Test chart type detection accuracy
- [x] 8.6 Test edge cases (empty results, large datasets, non-numerical data)
- [x] 8.7 Test terminal compatibility (Unicode vs ASCII)
- [x] 8.8 Update README with chart visualization examples
- [x] 8.9 Add usage examples for different chart types
