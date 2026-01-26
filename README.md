# AIQ

AIQ (AI Query): An intelligent SQL client that translates your natural language questions into precise SQL queries for MySQL, SeekDB, and other databases.

## Features

- 🎯 **Natural Language to SQL**: Ask questions in plain English, get precise SQL queries
- 📊 **Chart Visualization**: View query results as beautiful charts (bar, line, pie, scatter)
- 🔌 **Multiple Database Support**: Supports MySQL, PostgreSQL, and SeekDB
- ⚙️ **Easy Configuration**: Guided setup wizard for LLM and database connections
- 🎨 **Beautiful CLI**: Interactive menus with smooth transitions and color-coded output
- 🎨 **Customizable Charts**: Customize chart types, colors, titles, and axis labels
- 🔒 **Secure**: Local configuration storage, no cloud sync required

## Installation

### Prerequisites

- Go 1.21 or later
- MySQL database (for database queries)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/aiq/aiq.git
cd aiq

# Build the binary
go build -o aiq cmd/aiq/main.go

# Install (optional)
sudo mv aiq /usr/local/bin/
```

## Quick Start

1. **Run AIQ**:
   ```bash
   aiq
   ```

2. **First Run Setup**:
   - On first launch, you'll be guided through LLM configuration
   - Enter your LLM API URL and API Key
   - Configuration is saved to `~/.config/aiq/config.yaml`

3. **Add a Database Source**:
   - Select `source` from the main menu
   - Choose `add` to add a new MySQL connection
   - Enter connection details (host, port, database, username, password)

4. **Query Your Database**:
   - Select `sql` from the main menu
   - Choose a data source
   - Enter your question in natural language
   - Review the generated SQL and confirm execution
   - Choose to view results as table, chart, or both
   - Customize chart settings (type, color, title, labels)

## Usage

### Main Menu

- **config**: Manage LLM and tool configuration
- **source**: Manage database connection configurations
- **sql**: Enter SQL interactive mode (requires a selected source)
- **exit**: Exit the application

### Configuration Management

- View current LLM configuration
- Update LLM API URL
- Update LLM API Key

### Data Source Management

- Add new database connections (MySQL, PostgreSQL, SeekDB)
- List all configured sources
- Remove database connections

### Chart Visualization

AIQ supports multiple chart types for visualizing query results:

- **Bar Chart**: For categorical vs numerical data (e.g., sales by region)
- **Line Chart**: For time series or sequential data (e.g., sales over time)
- **Pie Chart**: For proportional categorical data (e.g., market share)
- **Scatter Plot**: For numerical vs numerical data (e.g., correlation analysis)

**Chart Customization**:
- Select chart type manually or use auto-detection
- Choose from predefined color palettes
- Customize chart title and axis labels
- Automatic detection of suitable chart types based on data structure

**Example Queries for Charts**:
```sql
-- Bar chart: Count by category
SELECT category, COUNT(*) AS count FROM products GROUP BY category;

-- Line chart: Sales over time
SELECT date, SUM(amount) AS total FROM sales GROUP BY date ORDER BY date;

-- Pie chart: Distribution
SELECT status, COUNT(*) AS count FROM orders GROUP BY status;

-- Scatter plot: Correlation
SELECT price, sales FROM products;
```

## Configuration

Configuration files are stored in `~/.config/aiq/`:

- `config.yaml`: LLM configuration (URL, API Key)
- `sources.yaml`: Database connection configurations

## Development

### Project Structure

```
cmd/aiq/          # Main entry point
internal/
  cli/            # CLI commands and menu system
  config/         # Configuration management
  source/         # Data source management
  sql/            # SQL interactive mode
  llm/            # LLM client integration
  db/             # Database connection and query execution
  chart/          # Chart visualization (bar, line, pie, scatter)
  ui/             # UI components (prompts, colors, loading, charts)
```

### Building

```bash
go build -o aiq cmd/aiq/main.go
```

### Running Tests

```bash
go test ./...
```

## License

See [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
