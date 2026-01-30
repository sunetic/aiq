## Why

当 LLM 调用 `execute_command` 工具时，有时会返回双重编码的 JSON 字符串（例如 `"{\"command\":\"brew services list | grep mysql\"}"`），而不是直接的 JSON 对象。当前的 `ParseArguments()` 函数虽然尝试处理这种情况，但在某些情况下仍然会失败，导致工具调用失败并显示错误信息 "json: cannot unmarshal string into Go value of type map[string]interface {}"。这会影响用户体验，特别是在 Skills 使用 `execute_command` 工具时。

## What Changes

- **修复 JSON 参数解析逻辑**：改进 `ParseArguments()` 函数，更健壮地处理双重编码的 JSON 字符串
- **增强错误处理**：在 `tool_handler.go` 中改进错误处理逻辑，提供更清晰的错误信息
- **添加测试用例**：添加针对双重编码 JSON 字符串的测试用例，确保修复有效

## Capabilities

### New Capabilities
<!-- 无新能力 -->

### Modified Capabilities
- `tool-execution`: 改进工具参数解析的健壮性，确保能正确处理 LLM 返回的各种 JSON 格式

## Impact

**受影响的代码：**
- `internal/llm/client.go`: `ParseArguments()` 函数需要改进 JSON 解析逻辑
- `internal/sql/tool_handler.go`: `ExecuteTool()` 函数中的错误处理逻辑可能需要调整

**测试：**
- 需要添加测试用例验证双重编码 JSON 字符串的解析
- 需要验证各种边界情况（空字符串、无效 JSON、嵌套引号等）

**用户体验：**
- 修复后，LLM 返回的 JSON 参数格式即使不够标准，也能被正确解析
- 减少工具调用失败的情况，提升系统稳定性
