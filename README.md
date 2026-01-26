<div align="center">

# AIQ

**An intelligent SQL client that translates natural language into SQL queries**

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg?style=flat-square)](LICENSE)

*Ask questions in plain English, get precise SQL queries, visualize results as beautiful charts*

</div>

---

## 📖 Introduction

AIQ (AI Query) is an intelligent SQL client that enables you to interact with databases using natural language. No need to write SQL manually—just ask questions in natural language, and AIQ will automatically generate SQL queries, execute them, and visualize the results as beautiful charts.

### ✨ Key Features

- 🗣️ **Natural Language to SQL** - Ask questions in plain English or Chinese, get precise SQL queries
- 📊 **Chart Visualization** - Support for bar charts, line charts, pie charts, and scatter plots
- 🔌 **Multiple Database Support** - [seekdb](https://www.oceanbase.ai/), MySQL, and PostgreSQL
- 🎨 **Beautiful CLI Interface** - Smooth interactions and color-coded output
- ⚙️ **Easy Configuration** - Guided setup wizard
- 🔒 **Local Storage** - Configuration and connection info stored securely locally

## 🚀 Quick Start

### Installation

#### Build from Source

```bash
# Clone the repository
git clone https://github.com/aiq/aiq.git
cd aiq

# Build the binary
go build -o aiq cmd/aiq/main.go

# Install to system path (optional)
sudo mv aiq /usr/local/bin/
```

### First Run

1. **Start AIQ**
   ```bash
   aiq
   ```

2. **Configure LLM**
   - On first launch, a configuration wizard will start
   - Enter LLM API URL (e.g., `https://api.openai.com/v1`)
   - Enter API Key
   - Enter model name (e.g., `gpt-4`)
   - Configuration is saved to `~/.aiqconfig/config.yaml`

3. **Add Data Source**
   ```bash
   # Select 'source' from the main menu
   # Choose 'add' to add a new connection
   # Enter database connection details:
   #   - Database type: seekdb / MySQL / PostgreSQL
   #   - Host address
   #   - Port
   #   - Database name
   #   - Username and password
   ```

4. **Start Querying**
   ```bash
   # Select 'chat' from the main menu
   # Choose a data source
   # Enter your question in natural language, for example:
   #   "Show total sales for the last week"
   #   "Count products by category"
   #   "Show user registration trends"
   ```

## 💡 Features

### 1. Natural Language to SQL

No need to write SQL—just ask questions in natural language:

```
aiq> Show total orders for the last week
```

AIQ will automatically:
- Understand your question
- Generate the corresponding SQL query
- Execute the query and return results

### 2. Chart Visualization

Query results can be automatically visualized as various chart types:

#### 📊 Bar Chart
Perfect for categorical data comparison, for example:
```sql
SELECT category, COUNT(*) AS count FROM products GROUP BY category;
```

#### 📈 Line Chart
Ideal for time series data, for example:
```sql
SELECT date, SUM(amount) AS total FROM sales GROUP BY date ORDER BY date;
```

#### 🥧 Pie Chart
Great for proportional distribution, for example:
```sql
SELECT status, COUNT(*) AS count FROM orders GROUP BY status;
```

#### 📉 Scatter Plot
Useful for numerical relationship analysis, for example:
```sql
SELECT price, sales FROM products;
```

### 3. Intelligent Chart Detection

AIQ automatically detects the most suitable chart type based on query result structure:
- **Categorical + Numerical** → Bar chart or Pie chart
- **Temporal + Numerical** → Line chart
- **Numerical + Numerical** → Scatter plot

### 4. Multiple Database Support

Supports various database types:
- **[seekdb](https://www.oceanbase.ai/)** - The AI-Native Search Database that unifies relational, vector, text, JSON and GIS in a single engine
- **MySQL** - The most popular relational database
- **PostgreSQL** - Powerful open-source database

### 5. Interactive Interface

- 🎯 Clear menu navigation
- 🎨 Color-coded output and syntax highlighting
- ⌨️ Support for Chinese input
- 📋 Table and chart display

## 📚 Usage Guide

### Main Menu

AIQ provides a clean main menu:

```
AIQ - Main Menu
? config   - Manage LLM configuration
  source   - Manage database connections
  chat     - Query database with natural language
  exit     - Exit application
```

### Configuration Management

Select `config` from the main menu to:
- View current LLM configuration
- Update API URL
- Update API Key
- Update model name

### Data Source Management

Select `source` from the main menu to:
- Add new database connections
- View all configured data sources
- Remove data sources

### Chat Mode

Select `chat` from the main menu to enter query mode:

1. **Select Data Source** - Choose from configured data sources
2. **Enter Question** - Describe your query in natural language
3. **View Results** - Choose to display as table, chart, or both
4. **Select Chart Type** - If multiple chart types are available, choose your preferred one

### Example Queries

```
# Aggregation queries
aiq> Count employees by department

# Trend analysis
aiq> Show sales trends for the last month

# Distribution analysis
aiq> Show order status distribution

# Correlation analysis
aiq> Analyze the relationship between product price and sales
```

## ⚙️ Configuration

Configuration files are stored in `~/.aiqconfig/`:

- **config.yaml** - LLM configuration (URL, API Key, Model)
- **sources.yaml** - Database connection configurations

### Configuration Examples

**config.yaml**
```yaml
llm:
  url: https://api.openai.com/v1
  apiKey: sk-...
  model: gpt-4
```

**sources.yaml**
```yaml
sources:
  - name: local-mysql
    type: MySQL
    host: localhost
    port: 3306
    database: testdb
    username: root
    password: password
```

## 🛠️ Development

### Project Structure

```
aiq/
├── cmd/aiq/          # Main entry point
├── internal/
│   ├── cli/          # CLI commands and menu system
│   ├── config/       # Configuration management
│   ├── source/       # Data source management
│   ├── sql/          # SQL interactive mode (chat mode)
│   ├── llm/          # LLM client integration
│   ├── db/           # Database connection and query execution
│   ├── chart/        # Chart visualization (bar, line, pie, scatter)
│   └── ui/           # UI components (prompts, colors, loading, charts)
└── openspec/         # OpenSpec change management
```

### Building

```bash
go build -o aiq cmd/aiq/main.go
```

### Running Tests

```bash
go test ./...
```

## 📝 License

This project is licensed under the [Apache License 2.0](LICENSE).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

<div align="center">

**Made with ❤️ using Go**

[Report Bug](https://github.com/aiq/aiq/issues) · [Request Feature](https://github.com/aiq/aiq/issues) · [View Documentation](https://github.com/aiq/aiq)

</div>
