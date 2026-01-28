<div align="center">

# AIQ

**一个将自然语言转换为 SQL 查询的智能 SQL 客户端**

[![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg?style=flat-square)](LICENSE)

*用自然语言提问，获得精确的 SQL 查询，将结果可视化为精美的图表*

</div>

---

## 📖 简介

AIQ (AI Query) 是一个智能 SQL 客户端，通过自然语言与数据库交互。无需编写 SQL，只需用自然语言提问，AIQ 会自动生成 SQL 查询并执行，还能将结果可视化为精美的图表。

### ✨ 核心特性

- 🗣️ **自然语言查询** - 用中文或英文提问，自动生成 SQL
- 💬 **多轮对话** - 保持对话上下文，支持查询优化和后续问题
- 🆓 **自由聊天模式** - 无需数据库连接即可进行通用对话和 Skills 操作
- 📊 **图表可视化** - 自动检测并渲染图表（柱状图、折线图、饼图、散点图）
- 🔌 **多数据库支持** - [seekdb](https://www.oceanbase.ai/)、MySQL、PostgreSQL
- 🎯 **Skills 系统** - 通过自定义领域知识扩展 AI 能力（基于 LLM 的语义匹配）
- 🧠 **智能上下文管理** - 动态 Skills 加载/淘汰和基于 LLM 的压缩
- 🎨 **美观的 CLI 界面** - 流畅的交互体验和彩色输出
- 💾 **会话持久化** - 保存和恢复对话会话

## 🚀 快速开始

### 安装

```bash
# 克隆并构建
git clone https://github.com/aiq/aiq.git
cd aiq
go build -o aiq cmd/aiq/main.go

# 安装（可选）
sudo mv aiq /usr/local/bin/
```

### 首次使用

1. **启动 AIQ**: `aiq`
2. **配置 LLM**: 输入 API URL、API Key 和模型名称（首次运行会启动配置向导）
3. **添加数据源**: 选择 `source` → `add` → 输入数据库连接信息
4. **开始查询**: 选择 `chat` → 选择数据源 → 用自然语言提问

**示例查询:**
```
aiq> 显示最近一周的销售额
aiq> 统计每个类别的商品数量
aiq> 查看用户注册趋势
```

## 📚 使用指南

### 主菜单

```
AIQ - Main Menu
? config   - Manage LLM configuration
  source   - Manage database connections
  chat     - Query database with natural language
  exit     - Exit application
```

### 聊天模式

AIQ 支持两种模式：

**数据库模式**（已选择数据源）：
- 完整的 SQL 查询功能
- 数据库模式上下文可用
- 图表可视化支持

**自由模式**（未选择数据源）：
- 与 AI 进行通用对话
- Skills 操作（execute_command、http_request、file_operations）
- 不支持 SQL 执行（选择数据源后可启用 SQL 查询）

**多轮对话:**
```
aiq[source-name]> 显示上周的总销售额
[生成 SQL 和结果...]

aiq[source-name]> 修改为只显示最近 3 天
[AIQ 理解上下文并生成更新的 SQL...]
```

**命令:**
- `/history` - 查看对话历史
- `/clear` - 清除对话历史
- `exit` 或 `back` - 退出聊天模式（会话自动保存）

**进入自由模式:**
- 当没有配置数据源时，AIQ 自动进入自由模式
- 当存在数据源时，在源选择菜单中选择"Skip (free mode)"选项

**恢复会话:**
```bash
aiq -s ~/.aiqconfig/sessions/session_20260126100000.json
```

### 图表可视化

AIQ 会根据查询结果自动检测合适的图表类型：
- **分类 + 数值** → 柱状图或饼图
- **时间 + 数值** → 折线图
- **数值 + 数值** → 散点图

## 🎯 Skills - 扩展 AI 能力

Skills 允许你通过提供自定义指令和上下文来扩展 AIQ 的能力。Skills 会根据你的查询自动匹配和加载，使用**基于 LLM 的语义匹配**以提高准确性。

### 快速开始

1. **创建 Skill 目录:**
```bash
mkdir -p ~/.aiqconfig/skills/my-skill
```

2. **创建 SKILL.md 文件:**
```markdown
---
name: my-skill
description: 针对指标、仪表板和 SQL 模式的领域特定指导
---

# My Custom Skill

此 Skill 提供分析工作流和常见 SQL 模式的指导。

## 核心概念

- 指标和维度的命名规范
- KPI 计算模式和注意事项
- 基于时间的聚合和队列分析

## 使用示例

### 周度 KPI 汇总
```sql
SELECT DATE_TRUNC('week', created_at) AS week,
       COUNT(*) AS orders,
       SUM(amount) AS revenue
FROM orders
GROUP BY week
ORDER BY week;
```
```

3. **重启 AIQ** - Skills 会在启动时自动加载

4. **使用** - 当你查询与 Skill 描述匹配的主题时，它会自动加载

### Skill 文件格式

每个 Skill 必须包含：

- **YAML Frontmatter**（必需）：
  - `name`: Skill 名称（小写，使用连字符，如 `my-skill`）
  - `description`: Skill 描述（最多 200 字符，用于查询匹配）

- **Markdown 内容**: 指令、示例和指导

### 工作原理

1. **启动时**: AIQ 从 `~/.aiqconfig/skills/` 加载所有 Skills 的元数据（name, description）
2. **查询时**: 系统使用**基于 LLM 的语义匹配**来查找相关 Skills（如果 LLM 不可用则回退到关键词匹配）
3. **自动加载**: 最相关的 Top 3 Skills 会被加载到 prompt 中
4. **动态管理**: 系统跟踪 Skills 使用情况，在对话过程中淘汰未使用的 Skills
5. **智能压缩**: 系统使用 LLM 语义压缩自动管理 prompt 长度（保留关键上下文的同时减少 token 使用）

### 匹配规则

Skills 使用**基于 LLM 的语义理解**进行匹配，以提高准确性：
- **语义匹配**: LLM 理解查询意图，基于含义匹配 Skills，而不仅仅是关键词
- **示例**: "install mysql" 匹配安装类 Skills，而不是数据库文档 Skills（即使它们提到 MySQL）
- **回退机制**: 如果 LLM 匹配失败，系统回退到基于关键词的匹配：
  - 精确名称匹配（最高优先级）
  - 部分名称匹配
  - 描述关键词匹配

### 动态 Skills 管理

在多轮对话过程中：
- **使用跟踪**: 系统跟踪每个 Skill 最后匹配/使用的时间
- **自动淘汰**: 在最近 3 次查询中未匹配的 Skills 会被淘汰以释放 token
- **上下文相关性**: 与当前对话上下文相关的 Skills 会被保留，即使在当前查询中未匹配
- **优先级管理**: 活跃 Skills（最近使用）> 相关 Skills（已匹配）> 非活跃 Skills（未匹配）

### 推荐 Skills

- **[seekdb Skill](https://github.com/oceanbase/seekdb-ecology-plugins/blob/main/claudecode-plugin/skills/seekdb/SKILL.md)** - SeekDB 文档目录和使用指导

### 内置工具

Skills 可以在其指令中使用以下内置工具：

- **`execute_sql`** - 执行数据库 SQL 查询
- **`http_request`** - 发起 HTTP 请求（GET, POST, PUT, DELETE）
- **`execute_command`** - 执行 shell 命令（有安全白名单限制）
- **`file_operations`** - 读写文件（限制在安全目录内）

**注意**: Skills 是上下文信息，不是工具本身。它们指导 AI 如何使用内置工具。

### Prompt 管理与 LLM 压缩

系统使用**基于 LLM 的语义压缩**自动管理 prompt 长度：

- **80% 阈值**: 开始 LLM 压缩（中等压缩，约 50% 减少）
  - LLM 总结对话历史，同时保留关键决策、结果和用户偏好
  - 如果 LLM 压缩失败，回退到简单截断
  
- **90% 阈值**: 激进 LLM 压缩（约 70% 减少）+ 淘汰低优先级 Skills
  - 更激进的总结，同时保持基本上下文
  
- **95% 阈值**: 最大 LLM 压缩（约 80% 减少，只保留基本上下文）
  - 压缩对话历史和 Skills 内容
  - 只保留活跃 Skills 和基本上下文

**LLM 压缩的优势:**
- 比简单截断更好地保留重要上下文
- 保持对话连续性
- 在保持基本信息的同时减少 token 使用
- 缓存压缩结果以避免重复压缩相同内容

### 目录结构

Skills 存储在 `~/.aiqconfig/skills/<skill-name>/SKILL.md`:

```
~/.aiqconfig/
└── skills/
    ├── my-skill/
    │   └── SKILL.md
    └── data-analysis/
        └── SKILL.md
```

**注意**: 每个 Skill 目录只包含一个 `SKILL.md` 文件。如果需要多个文件，将内容合并到一个文件中，或拆分为多个更小的 Skills。

### 故障排除

**Skills 未加载:**
- 检查目录结构: `~/.aiqconfig/skills/<skill-name>/SKILL.md`
- 验证 YAML frontmatter 格式（必须以 `---` 开始和结束）
- 确保 `name` 和 `description` 字段存在
- 查看启动日志中的错误

**Skills 未匹配:**
- 在 Skill `description` 中包含相关关键词
- 尝试在查询中使用 Skill 名称
- 检查是否有多个 Skills 竞争（只选择 Top 3）

## ⚙️ 配置

配置文件存储在 `~/.aiqconfig/`:

- **config/config.yaml** - LLM 配置（URL、API Key、模型）
- **config/sources.yaml** - 数据库连接配置
- **sessions/** - 对话会话文件（自动生成）
- **skills/** - 自定义 Skills（见上方 Skills 部分）

**示例 config.yaml:**
```yaml
llm:
  url: https://api.openai.com/v1
  apiKey: sk-...
  model: gpt-4
```

**示例 sources.yaml:**
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

## 🛠️ 开发

### 项目结构

```
aiq/
├── cmd/aiq/          # 主程序入口
├── internal/
│   ├── cli/          # CLI 命令和菜单系统
│   ├── config/       # 配置管理
│   ├── source/       # 数据源管理
│   ├── sql/          # SQL 交互模式（chat 模式）
│   ├── skills/       # Skills 系统（匹配、加载、管理）
│   ├── prompt/       # Prompt 构建和压缩
│   ├── llm/          # LLM 客户端集成
│   ├── db/           # 数据库连接和查询执行
│   ├── chart/        # 图表可视化
│   ├── tool/         # 工具系统（内置工具）
│   └── ui/           # UI 组件
└── openspec/         # OpenSpec 变更管理
```

### 构建

```bash
go build -o aiq cmd/aiq/main.go
```

### 运行测试

```bash
go test ./...
```

## 📝 许可证

本项目采用 [Apache License 2.0](LICENSE) 许可证。

## 🤝 贡献

欢迎贡献！请随时提交 Pull Request。

---

<div align="center">

**Made with ❤️ using Go**

[报告问题](https://github.com/aiq/aiq/issues) · [提交功能请求](https://github.com/aiq/aiq/issues) · [查看文档](https://github.com/aiq/aiq)

</div>
