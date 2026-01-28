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
- 💬 **Multi-Turn Conversation** - Maintain conversation context for refined queries and follow-up questions
- 🆓 **Free Chat Mode** - General conversation and Skills operations without database connection
- 📊 **Chart Visualization** - Automatic chart detection and rendering (bar, line, pie, scatter plots)
- 🔌 **Multiple Database Support** - [seekdb](https://www.oceanbase.ai/), MySQL, and PostgreSQL
- 🎯 **Skills System** - Extend AI capabilities with custom domain knowledge (LLM-based semantic matching)
- 🧠 **Intelligent Context Management** - Dynamic Skills loading/eviction and LLM-based compression
- 🎨 **Beautiful CLI Interface** - Smooth interactions and color-coded output
- 💾 **Session Persistence** - Save and restore conversation sessions

## 🚀 Quick Start

### Installation

```bash
# Clone and build
git clone https://github.com/aiq/aiq.git
cd aiq
go build -o aiq cmd/aiq/main.go

# Install (optional)
sudo mv aiq /usr/local/bin/
```

### First Run

1. **Start AIQ**: `aiq`
2. **Configure LLM**: Enter API URL, API Key, and model name (wizard runs on first launch)
3. **Add Data Source**: Select `source` → `add` → Enter database connection details
4. **Start Querying**: Select `chat` → Choose data source → Ask questions in natural language

**Example queries:**
```
aiq> Show total sales for the last week
aiq> Count products by category
aiq> Show user registration trends
```

## 📚 Usage

### Main Menu

```
AIQ - Main Menu
? config   - Manage LLM configuration
  source   - Manage database connections
  chat     - Query database with natural language
  exit     - Exit application
```

### Chat Mode

AIQ supports two modes:

**Database Mode** (with source selected):
- Full SQL query capabilities
- Database schema context available
- Chart visualization support

**Free Mode** (no source selected):
- General conversation with AI
- Skills-based operations (execute_command, http_request, file_operations)
- No SQL execution (select a source to enable SQL queries)

**Multi-turn conversation:**
```
aiq[source-name]> Show total sales for last week
[Generated SQL and results...]

aiq[source-name]> Modify to show only last 3 days
[AIQ understands context and generates updated SQL...]
```

**Commands:**
- `/history` - View conversation history
- `/clear` - Clear conversation history
- `exit` or `back` - Exit chat mode (session auto-saved)

**Entering Free Mode:**
- When no sources are configured, AIQ automatically enters free mode
- When sources exist, select "Skip (free mode)" option in source selection menu

**Session restore:**
```bash
aiq -s ~/.aiqconfig/sessions/session_20260126100000.json
```

### Chart Visualization

AIQ automatically detects suitable chart types based on query results:
- **Categorical + Numerical** → Bar chart or Pie chart
- **Temporal + Numerical** → Line chart
- **Numerical + Numerical** → Scatter plot

## 🎯 Skills - Extending AI Capabilities

Skills allow you to extend AIQ's capabilities by providing custom instructions and context to the AI agent. Skills are automatically matched and loaded based on your queries using **LLM-based semantic matching** for improved accuracy.

### Quick Start

1. **Create Skill directory:**
```bash
mkdir -p ~/.aiqconfig/skills/my-skill
```

2. **Create SKILL.md file:**
```markdown
---
name: my-skill
description: Domain-specific guidance for metrics, dashboards, and SQL patterns
---

# My Custom Skill

This skill provides guidance for analytics workflows and common SQL patterns.

## Key Concepts

- Naming conventions for metrics and dimensions
- KPI calculation patterns and caveats
- Time-based aggregations and cohort analysis

## Usage Examples

### Weekly KPI Summary
```sql
SELECT DATE_TRUNC('week', created_at) AS week,
       COUNT(*) AS orders,
       SUM(amount) AS revenue
FROM orders
GROUP BY week
ORDER BY week;
```

3. **Restart AIQ** - Skills are loaded automatically on startup

4. **Use it** - When you query about topics matching your skill's description, it will be automatically loaded

### Skill File Format

Each Skill must have:

- **YAML Frontmatter** (required):
  - `name`: Skill name (lowercase, use hyphens, e.g., `my-skill`)
  - `description`: Skill description (max 200 chars, used for query matching)

- **Markdown content**: Instructions, examples, and guidance

### How It Works

1. **On Startup**: AIQ loads metadata (name, description) from all Skills in `~/.aiqconfig/skills/`
2. **On Query**: System uses **LLM-based semantic matching** to find relevant Skills (falls back to keyword matching if LLM unavailable)
3. **Auto-Load**: Top 3 most relevant Skills are loaded into the prompt
4. **Dynamic Management**: System tracks Skills usage and evicts unused Skills during conversation
5. **Smart Compression**: System automatically manages prompt length using LLM semantic compression (preserves key context while reducing tokens)

### Matching Rules

Skills are matched using **LLM-based semantic understanding** for improved accuracy:
- **Semantic Matching**: LLM understands query intent and matches Skills based on meaning, not just keywords
- **Example**: "install mysql" matches installation Skills, not database documentation Skills (even if they mention MySQL)
- **Fallback**: If LLM matching fails, system falls back to keyword-based matching:
  - Exact name match (highest priority)
  - Partial name match
  - Description keyword match

### Dynamic Skills Management

During multi-turn conversations:
- **Usage Tracking**: System tracks when each Skill was last matched/used
- **Automatic Eviction**: Skills not matched in last 3 queries are evicted to free up tokens
- **Context Relevance**: Skills relevant to current conversation context are kept even if not matched in current query
- **Priority Management**: Active Skills (used recently) > Relevant Skills (matched) > Inactive Skills (not matched)

### Recommended Skills

- **[seekdb Skill](https://github.com/oceanbase/seekdb-ecology-plugins/blob/main/claudecode-plugin/skills/seekdb/SKILL.md)** - Documentation catalog and usage guidance for SeekDB

### Built-in Tools

Skills can use these built-in tools in their instructions:

- **`execute_sql`** - Execute SQL queries against the database
- **`http_request`** - Make HTTP requests (GET, POST, PUT, DELETE)
- **`execute_command`** - Execute shell commands (with security allowlist)
- **`file_operations`** - Read/write files (restricted to safe directories)

**Note**: Skills are context information, not tools themselves. They guide the AI on how to use the built-in tools.

### Prompt Management & LLM Compression

System automatically manages prompt length using **LLM-based semantic compression**:

- **80% threshold**: Start LLM compression (moderate compression, ~50% reduction)
  - LLM summarizes conversation history while preserving key decisions, results, and user preferences
  - Falls back to simple truncation if LLM compression fails
  
- **90% threshold**: Aggressive LLM compression (~70% reduction) + evict low-priority Skills
  - More aggressive summarization while maintaining essential context
  
- **95% threshold**: Maximum LLM compression (~80% reduction, keep only essential context)
  - Compresses both conversation history and Skills content
  - Keeps only active Skills and essential context

**Benefits of LLM Compression:**
- Preserves important context better than simple truncation
- Maintains conversation continuity
- Reduces token usage while keeping essential information
- Caches compression results to avoid re-compressing same content

### Directory Structure

Skills are stored in `~/.aiqconfig/skills/<skill-name>/SKILL.md`:

```
~/.aiqconfig/
└── skills/
    ├── my-skill/
    │   └── SKILL.md
    └── data-analysis/
        └── SKILL.md
```

**Note**: Each Skill directory contains only one `SKILL.md` file. If you need multiple files, merge content into one file or split into multiple smaller Skills.

### Troubleshooting

**Skills not loaded:**
- Check directory structure: `~/.aiqconfig/skills/<skill-name>/SKILL.md`
- Verify YAML frontmatter format (must start/end with `---`)
- Ensure `name` and `description` fields exist
- Check startup logs for errors

**Skills not matched:**
- Include relevant keywords in Skill `description`
- Try using Skill name in your query
- Check if multiple Skills are competing (only top 3 are selected)

## ⚙️ Configuration

Configuration files are stored in `~/.aiqconfig/`:

- **config/config.yaml** - LLM configuration (URL, API Key, Model)
- **config/sources.yaml** - Database connection configurations
- **sessions/** - Conversation session files (auto-generated)
- **skills/** - Custom Skills (see Skills section above)

**Example config.yaml:**
```yaml
llm:
  url: https://api.openai.com/v1
  apiKey: sk-...
  model: gpt-4
```

**Example sources.yaml:**
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
│   ├── skills/       # Skills system (matching, loading, management)
│   ├── prompt/       # Prompt building and compression
│   ├── llm/          # LLM client integration
│   ├── db/           # Database connection and query execution
│   ├── chart/        # Chart visualization
│   ├── tool/         # Tool system (built-in tools)
│   └── ui/           # UI components
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
