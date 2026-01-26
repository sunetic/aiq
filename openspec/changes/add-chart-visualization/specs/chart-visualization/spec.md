## ADDED Requirements

### Requirement: Chart type detection
The system SHALL automatically detect appropriate chart type based on query result structure.

#### Scenario: Detect bar chart
- **WHEN** query result has 2 columns where first column is categorical (string) and second is numerical
- **THEN** system detects bar chart as appropriate type

#### Scenario: Detect line chart
- **WHEN** query result has 2 columns where first column is temporal (date/time) or sequential numeric and second is numerical
- **THEN** system detects line chart as appropriate type

#### Scenario: Detect pie chart
- **WHEN** query result has 2 columns where first is categorical, second is numerical, and number of categories is small (< 10)
- **THEN** system detects pie chart as appropriate type

#### Scenario: Detect scatter plot
- **WHEN** query result has 2+ numerical columns with multiple rows
- **THEN** system detects scatter plot as appropriate type

#### Scenario: Fallback to table
- **WHEN** query result structure doesn't match any chart type pattern
- **THEN** system defaults to table view

### Requirement: Chart rendering option
The system SHALL provide option to render query results as charts after successful query execution.

#### Scenario: Display chart option
- **WHEN** query executes successfully and returns results
- **THEN** system prompts user to choose view type: [Table] [Chart] [Both]

#### Scenario: Render chart
- **WHEN** user selects Chart or Both option
- **THEN** system renders query results as detected chart type using ASCII/Unicode characters

#### Scenario: Skip chart for empty results
- **WHEN** query executes successfully but returns no rows
- **THEN** system only displays table view option (no chart option)

### Requirement: Chart rendering implementation
The system SHALL render charts using ASCII/Unicode characters in terminal.

#### Scenario: Render bar chart
- **WHEN** bar chart type is detected or selected
- **THEN** system renders vertical or horizontal bar chart with proper scaling and labels

#### Scenario: Render line chart
- **WHEN** line chart type is detected or selected
- **THEN** system renders line chart with data points and connecting lines

#### Scenario: Render pie chart
- **WHEN** pie chart type is detected or selected
- **THEN** system renders pie chart showing proportional distribution

#### Scenario: Render scatter plot
- **WHEN** scatter plot type is detected or selected
- **THEN** system renders scatter plot with data points

### Requirement: Chart customization
The system SHALL allow users to customize chart appearance and settings.

#### Scenario: Override chart type
- **WHEN** user wants different chart type than auto-detected
- **THEN** system provides option to manually select chart type

#### Scenario: Customize chart colors
- **WHEN** user selects chart customization
- **THEN** system provides predefined color palette options

#### Scenario: Set chart title and labels
- **WHEN** user customizes chart
- **THEN** system allows setting chart title and axis labels

### Requirement: Chart display integration
The system SHALL integrate chart rendering seamlessly into SQL mode workflow.

#### Scenario: Chart after query execution
- **WHEN** query executes successfully in SQL mode
- **THEN** system displays results count and prompts for view type selection

#### Scenario: Display both table and chart
- **WHEN** user selects "Both" option
- **THEN** system displays table first, then chart below

#### Scenario: Return to SQL prompt
- **WHEN** chart is displayed
- **THEN** system returns to SQL prompt after user acknowledges (Enter key)

### Requirement: Chart rendering constraints
The system SHALL handle edge cases and limitations gracefully.

#### Scenario: Large dataset handling
- **WHEN** query result has more than 1000 data points
- **THEN** system warns user and offers to sample data or render subset

#### Scenario: Terminal compatibility
- **WHEN** terminal doesn't support Unicode characters
- **THEN** system falls back to ASCII-only chart rendering

#### Scenario: Single column result
- **WHEN** query result has only one column
- **THEN** system suggests table view only (no chart option)

#### Scenario: Non-numerical data
- **WHEN** query result columns don't contain numerical data suitable for charts
- **THEN** system defaults to table view with appropriate message
