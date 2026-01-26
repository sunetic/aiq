## Why

Currently, AIQ displays query results only in tabular format. While tables are effective for structured data, they don't provide intuitive visual insights for numerical trends, distributions, or comparisons. Users often need to analyze data patterns, trends over time, or compare values across categories - tasks that are much easier with visual charts than reading through rows of numbers.

This change adds data visualization capabilities to AIQ, enabling users to render query results as ASCII/Unicode charts directly in the CLI. This enhances the analytical power of AIQ without requiring users to export data to external tools.

## What Changes

- **Chart Rendering Library**: Integrate a Go library for ASCII/Unicode chart generation (e.g., `go-echarts` or custom implementation using `termui`/`bubbletea`)
- **Chart Type Detection**: Automatically detect suitable chart types based on query result structure (bar charts for categorical data, line charts for time series, pie charts for distributions)
- **Chart Display Integration**: Add chart rendering option in SQL mode after query execution, allowing users to choose between table view and chart view
- **Chart Configuration**: Provide options for chart customization (colors, labels, axis formatting) through interactive prompts
- **Multi-Chart Support**: Support rendering multiple charts for complex queries with multiple data series

## Capabilities

### New Capabilities

- `chart-visualization`: Render query results as ASCII/Unicode charts in CLI, supporting multiple chart types (bar, line, pie, scatter) with automatic type detection and customization options

### Modified Capabilities

- `sql-interactive-mode`: Extend to support chart rendering option after query execution, allowing users to visualize results as charts in addition to tables

## Impact

- **New Dependencies**: Chart rendering library (e.g., `github.com/go-echarts/charts` or similar ASCII chart library)
- **User Experience**: Enhanced data analysis capabilities with visual representation of query results
- **SQL Mode Flow**: Extended query result display flow to include chart rendering option
- **Performance**: Minimal impact - chart rendering happens client-side after query execution
- **Documentation**: Update README and usage examples to demonstrate chart visualization features
