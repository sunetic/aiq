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

### 🎬 Demo

<video src="docs/basic.webm" controls width="100%"></video>

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

#### Option 1: One-Click Install (Recommended)

**Unix/Linux/macOS:**
```bash
curl -fsSL https://raw.githubusercontent.com/sunetic/aiq/main/scripts/install.sh | bash
```

**Windows:**
```powershell
# Download and run install.bat
# Or run in PowerShell:
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/sunetic/aiq/main/scripts/install.bat" -OutFile "install.bat"
.\install.bat
```

The installation script will:
- Automatically detect the latest version from GitHub Releases
- Detect your system architecture (darwin-amd64, darwin-arm64, linux-amd64, linux-arm64, windows-amd64)
- Download the binary to `~/.aiq/bin` (Unix/Linux/macOS) or `%USERPROFILE%\.aiq\bin` (Windows)
- Print PATH command for you to add manually
- Verify the installation

**After installation, add to PATH:**

**Unix/Linux/macOS (zsh):**
```bash
echo 'export PATH="$HOME/.aiq/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc
```

**Unix/Linux/macOS (bash):**
```bash
echo 'export PATH="$HOME/.aiq/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

**Windows:**
```cmd
setx PATH "%PATH%;%USERPROFILE%\.aiq\bin"
```
(Then open a new terminal window)

#### Option 2: Manual Installation

```bash
# Clone and build
git clone https://github.com/sunetic/aiq.git
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

**Database Mode** (with source selected): Full SQL query capabilities with chart visualization  
**Free Mode** (no source selected): General conversation and Skills operations

**Commands:** `/history` - View history | `/clear` - Clear history | `exit`/`back` - Exit (auto-saved)

**Session restore:** `aiq -s ~/.aiq/sessions/session_20260126100000.json`

**Version:** `aiq -v` or `aiq --version` - Display version and commit ID

### Chart Visualization

Auto-detects chart types: Categorical+Numerical → Bar/Pie | Temporal+Numerical → Line | Numerical+Numerical → Scatter

## 🎯 Skills - Extending AI Capabilities

Skills extend AIQ's capabilities by providing custom instructions and context. Automatically matched and loaded using **LLM-based semantic matching**.

<video src="docs/skill.webm" controls width="100%"></video>

### Quick Start

1. **Create Skill:** `mkdir -p ~/.aiq/skills/my-skill`
2. **Create SKILL.md** with YAML frontmatter:
```markdown
---
name: my-skill
description: Domain-specific guidance for metrics and SQL patterns
---

# My Custom Skill
[Your instructions and examples here]
```
3. **Restart AIQ** - Skills auto-load on startup

### How It Works

- **Matching**: LLM-based semantic matching (falls back to keyword matching)
- **Loading**: Top 3 most relevant Skills loaded per query
- **Management**: Auto-evicts unused Skills, tracks usage, manages priority
- **Compression**: LLM-based semantic compression at 80%/90%/95% thresholds

### Built-in Tools

Skills guide AI on using: `execute_sql`, `http_request`, `execute_command`, `file_operations`

### Recommended Skills

- **[seekdb Skill](https://github.com/oceanbase/seekdb-ecology-plugins/blob/main/claudecode-plugin/skills/seekdb/SKILL.md)** - SeekDB documentation and usage guidance

## ⚙️ Configuration

Config files in `~/.aiq/`: `config/config.yaml` (LLM), `config/sources.yaml` (databases), `sessions/`, `skills/`, `bin/`

## 🛠️ Development

**Build:** `go build -o aiq cmd/aiq/main.go`  
**Test:** `go test ./...`

## 📝 License

This project is licensed under the [Apache License 2.0](LICENSE).

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

---

<div align="center">

**Made with ❤️ using Go**

[Report Bug](https://github.com/sunetic/aiq/issues) · [Request Feature](https://github.com/sunetic/aiq/issues) · [View Documentation](https://github.com/sunetic/aiq)

</div>
