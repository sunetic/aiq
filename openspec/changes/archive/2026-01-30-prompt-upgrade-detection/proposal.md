## Why

为了方便用户修改 prompt 文件，系统将 prompt 文件输出到 `~/.aiq/prompts` 目录。但在升级时，系统不确定应该使用用户修改过的 prompt 文件，还是用新版本覆盖。这导致升级后可能使用过时的 prompt，或者意外覆盖用户的定制化修改。

需要添加一个检测机制，在启动时检测 prompt 文件版本与内置版本是否一致，如果不一致则询问用户是否覆盖，并记录用户的选择以避免重复询问。

## What Changes

- **添加应用版本号管理**：
  - 运行时动态获取应用版本号（优先从构建信息获取，如果没有则使用 git describe，最后使用默认值）
  - 获取 commit id（通过 git rev-parse 或构建时注入）
  - 用于记录用户选择，避免同一版本重复询问
  - 支持 `-v` / `--version` 命令行参数，打印版本号和 commit id
- **Prompt 内容检测机制**：
  - 计算代码中内置 prompt 字符串的内容哈希（SHA256）
  - 计算 `~/.aiq/prompts/` 目录中 prompt 文件的内容哈希（SHA256）
  - 启动时比较内置内容哈希与用户文件内容哈希，检测用户是否修改过文件
- **用户交互提示**：当检测到用户修改过 prompt 文件时，询问用户是否覆盖现有 prompt 文件
- **版本选择记录**：记录用户对每个应用版本的选择（覆盖/保留），存储在 `~/.aiq/config/prompt-version-choices.yaml`，相同应用版本仅询问一次
- **提示信息**：如果用户选择保留，显示提示信息告知如何手动删除文件以触发重建

## Capabilities

### New Capabilities
- `prompt-version-detection`: 检测 prompt 文件版本与内置版本的一致性，并在不一致时提示用户
- `application-version-management`: 管理应用版本号，用于版本检测和升级提示

### Modified Capabilities
- `configuration-management`: 需要添加版本选择记录的存储机制（可能存储在 `~/.aiq/config/` 目录下的新文件）

## Impact

- **代码变更**：
  - `internal/prompt/loader.go`: 
    - 添加内容哈希计算函数（SHA256）
    - 添加内容检测逻辑，比较内置内容哈希与用户文件内容哈希
    - 修改 `initializeDefaults()` 和 `NewLoader()` 方法
  - `internal/config/`: 添加版本选择记录的存储和读取功能
  - `internal/cli/root.go`: 在启动时调用版本检测逻辑
  - `cmd/aiq/main.go`: 添加 `-v` / `--version` 参数处理，打印版本号和 commit id 后退出
  - 新增文件：`internal/version/version.go` 提供运行时获取版本信息的函数：
    - `GetVersion()`: 优先从构建时注入的版本号获取（通过 `-ldflags` 注入），如果没有则尝试使用 `git describe`，最后返回默认值 "dev"
    - `GetCommitID()`: 优先从构建时注入的 commit id 获取，如果没有则尝试使用 `git rev-parse HEAD`，最后返回默认值 "unknown"
    - `GetVersionInfo()`: 返回格式化的版本信息字符串（如 "aiq v1.0.0 (commit: abc1234)"）

- **配置变更**：
  - 新增版本选择记录文件（如 `~/.aiq/config/prompt-version-choices.yaml`），记录用户对每个版本的选择

- **构建变更**：
  - CI/CD 构建时通过 `-ldflags` 注入版本号和 commit id：
    - `-X github.com/aiq/aiq/internal/version.Version=${{ github.ref_name }}`（从 git tag 获取）
    - `-X github.com/aiq/aiq/internal/version.CommitID=${{ github.sha }}`（从 GitHub Actions 获取 commit SHA）
  - 本地开发构建时如果没有注入版本号，会尝试从 git describe 和 git rev-parse 获取
  - 如果都没有，使用默认值 "dev" 和 "unknown"

- **用户体验**：
  - 启动时不打印版本号（避免干扰正常使用）
  - 支持 `aiq -v` 或 `aiq --version` 查看版本号和 commit id
  - 启动时可能显示版本检测提示（仅在版本不一致且用户未选择过时）
  - 升级后首次启动会询问是否覆盖 prompt 文件
