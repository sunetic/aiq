## Context

AIQ currently displays query results in tabular format using `ui.PrintTable()`. Users need to analyze numerical data, trends, and distributions, which would benefit from visual chart representation. This change adds chart visualization capabilities while maintaining the CLI-only constraint.

**Constraints:**
- Must render charts in terminal/CLI (no GUI)
- Should work with standard terminal emulators
- Must support common chart types: bar, line, pie, scatter
- Should automatically detect appropriate chart type based on data structure
- Must integrate seamlessly with existing SQL mode flow

**Stakeholders:**
- End users: Database users who need to analyze query results visually
- Developers: Future contributors maintaining chart rendering code

## Goals / Non-Goals

**Goals:**
- Render query results as ASCII/Unicode charts in terminal
- Support multiple chart types (bar, line, pie, scatter)
- Automatically detect suitable chart type based on data structure
- Provide interactive chart customization options
- Maintain consistent visual style with existing UI components
- Support both automatic and manual chart type selection

**Non-Goals:**
- Export charts to image files (future feature)
- Interactive chart manipulation (zoom, pan) - CLI limitation
- 3D charts or complex visualizations
- Real-time chart updates
- Chart history or saved chart configurations

## Decisions

### Chart Rendering Library: Custom Implementation with `charmbracelet/lipgloss`

**Decision**: Build custom chart rendering using `charmbracelet/lipgloss` and ASCII/Unicode characters.

**Rationale**:
- `lipgloss` is already a dependency, providing consistent styling
- Full control over rendering and customization
- No additional heavy dependencies
- Better integration with existing UI components
- Can optimize for terminal display constraints

**Alternatives Considered**:
- `go-echarts`: Designed for web, not suitable for CLI
- `termui`: Good but adds significant dependency overhead
- `bubbletea`: Overkill for static chart rendering
- External ASCII chart libraries: Limited customization, potential compatibility issues

### Chart Type Detection Strategy

**Decision**: Implement heuristic-based automatic detection with manual override option.

**Detection Rules**:
- **Bar Chart**: 2 columns, first is categorical (string), second is numerical
- **Line Chart**: 2 columns, first is temporal (date/time) or sequential (numeric), second is numerical
- **Pie Chart**: 2 columns, first is categorical, second is numerical, small number of categories (< 10)
- **Scatter Plot**: 2+ numerical columns, multiple rows
- **Table**: Default fallback for complex or unrecognized structures

**Rationale**:
- Most queries follow predictable patterns
- Reduces user decision fatigue
- Can be overridden if detection is incorrect
- Simple heuristics are fast and reliable

### Chart Display Flow

**Decision**: Add chart rendering as optional step after query execution, with prompt to choose view type.

**Flow**:
1. Query executes successfully
2. System displays results count
3. System prompts: "View as: [Table] [Chart] [Both]"
4. If Chart/Both selected, detect chart type and render
5. User can customize chart if needed

**Rationale**:
- Non-intrusive - doesn't change existing table display
- Gives users choice
- Can be skipped for non-visualizable queries
- Maintains backward compatibility

### Chart Customization

**Decision**: Provide basic customization options through interactive prompts.

**Options**:
- Chart type override (if auto-detection is wrong)
- Color scheme selection (from predefined palettes)
- Axis labels and formatting
- Chart title

**Rationale**:
- Basic customization covers most use cases
- Advanced options can be added later
- Keeps UI simple and intuitive
- Predefined options ensure visual consistency

## Architecture

### New Components

- `internal/chart/`: Chart rendering package
  - `detector.go`: Chart type detection logic
  - `renderer.go`: Chart rendering engine
  - `bar.go`: Bar chart implementation
  - `line.go`: Line chart implementation
  - `pie.go`: Pie chart implementation
  - `scatter.go`: Scatter plot implementation
  - `config.go`: Chart configuration and customization

### Modified Components

- `internal/sql/mode.go`: Add chart rendering option after query execution
- `internal/ui/`: Add chart-related UI helpers if needed

## Risks

- **Terminal Compatibility**: Some terminals may not render Unicode characters correctly
  - **Mitigation**: Fallback to ASCII-only rendering, detect terminal capabilities

- **Performance**: Rendering large datasets as charts may be slow
  - **Mitigation**: Limit chart rendering to reasonable dataset sizes (< 1000 points), provide option to sample data

- **Chart Type Detection Accuracy**: Heuristics may misclassify some queries
  - **Mitigation**: Always allow manual override, provide clear feedback on detected type

## Open Questions

- Should we support multi-series charts (e.g., multiple lines on same chart)?
- What is the maximum dataset size for chart rendering?
- Should charts be scrollable for large datasets?
- Do we need chart export functionality in this iteration?
