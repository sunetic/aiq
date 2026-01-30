## Context

当前系统在处理 LLM 返回的工具调用参数时，`ParseArguments()` 函数尝试解析 JSON 字符串。虽然已经实现了对双重编码 JSON 字符串的处理（检测并 unquote 外层引号），但在某些情况下仍然会失败。

**当前实现的问题：**
1. `ParseArguments()` 只处理一层引号包裹的情况，如果有多层嵌套的 JSON 字符串，可能无法正确处理
2. 错误处理逻辑分散在两个地方：`ParseArguments()` 和 `tool_handler.go`，导致代码重复
3. 错误信息不够清晰，难以诊断问题

**错误场景示例：**
- LLM 返回：`"{\"command\":\"brew services list | grep mysql\"}"`
- `ParseArguments()` 尝试 unquote 后得到：`{\"command\":\"brew services list | grep mysql\"}`
- 但解析时可能因为转义字符处理不当而失败

## Goals / Non-Goals

**Goals:**
- 改进 `ParseArguments()` 函数，能够正确处理各种格式的 JSON 参数（包括多重编码、转义字符等）
- 统一错误处理逻辑，减少代码重复
- 提供更清晰的错误信息，便于调试
- 添加全面的测试用例，覆盖各种边界情况

**Non-Goals:**
- 不改变工具调用的 API 接口
- 不修改 LLM 客户端的其他功能
- 不改变工具定义的结构

## Decisions

### 1. 改进 JSON 解析策略

**决策：** 使用递归方式处理多重编码的 JSON 字符串，直到无法再 unquote 为止。

**理由：**
- 当前实现只处理一层引号，无法处理多层嵌套的情况
- 递归处理可以确保无论有多少层嵌套，都能正确解析

**实现方式：**
```go
func (tc *ToolCall) ParseArguments() (map[string]interface{}, error) {
    argsStr := tc.Function.Arguments
    
    // 递归 unquote，直到无法再 unquote
    for {
        trimmed := strings.TrimSpace(argsStr)
        if len(trimmed) < 2 || trimmed[0] != '"' || trimmed[len(trimmed)-1] != '"' {
            break
        }
        var unquoted string
        if err := json.Unmarshal([]byte(trimmed), &unquoted); err != nil {
            break
        }
        argsStr = unquoted
    }
    
    // 尝试解析为 JSON 对象
    var args map[string]interface{}
    if err := json.Unmarshal([]byte(argsStr), &args); err != nil {
        return nil, fmt.Errorf("failed to parse arguments: %w", err)
    }
    return args, nil
}
```

**替代方案考虑：**
- **方案 A：** 只处理一层引号（当前方案）- 无法处理多层嵌套
- **方案 B：** 使用正则表达式匹配 - 不够健壮，可能误匹配
- **方案 C：** 递归 unquote（选择）- 最健壮，能处理任意层数

### 2. 简化错误处理逻辑

**决策：** 将错误处理逻辑集中在 `ParseArguments()` 中，移除 `tool_handler.go` 中的重复处理。

**理由：**
- 当前错误处理逻辑分散在两个地方，导致代码重复和维护困难
- 统一处理逻辑可以提高代码可维护性

**实现方式：**
- 在 `ParseArguments()` 中处理所有 JSON 解析相关的错误
- `tool_handler.go` 中只处理工具执行相关的错误

**替代方案考虑：**
- **方案 A：** 保持当前的双重处理 - 代码重复，维护困难
- **方案 B：** 统一到 `ParseArguments()`（选择）- 更清晰，易于维护

### 3. 改进错误信息

**决策：** 在错误信息中包含原始参数和解析步骤，便于调试。

**理由：**
- 当前的错误信息不够详细，难以诊断问题
- 包含更多上下文信息可以帮助快速定位问题

**实现方式：**
```go
return nil, fmt.Errorf("failed to parse arguments after unquoting: %w (original: %s)", err, truncateString(tc.Function.Arguments, 100))
```

## Risks / Trade-offs

**风险 1：** 递归 unquote 可能导致无限循环
- **缓解措施：** 添加最大递归深度限制（例如 10 层）

**风险 2：** 过度处理可能导致性能问题
- **缓解措施：** 递归深度限制可以防止性能问题，且大多数情况下只需要处理 1-2 层

**风险 3：** 可能误解析某些特殊格式的字符串
- **缓解措施：** 在 unquote 后验证是否为有效的 JSON 对象，如果不是则停止处理

**权衡：**
- **健壮性 vs 性能：** 选择健壮性，因为工具调用失败的影响远大于解析性能的微小开销
- **简单性 vs 完整性：** 选择完整性，确保能处理各种边界情况

## Migration Plan

**部署步骤：**
1. 实现改进后的 `ParseArguments()` 函数
2. 添加测试用例覆盖各种边界情况
3. 运行现有测试确保没有回归
4. 提交代码并合并到主分支

**回滚策略：**
- 如果发现问题，可以快速回滚到之前的版本
- 改进是向后兼容的，不会影响现有功能

**测试策略：**
- 单元测试：测试各种 JSON 格式的解析
- 集成测试：测试实际的工具调用场景
- 边界测试：测试空字符串、无效 JSON、多层嵌套等情况

## Open Questions

1. **是否需要支持其他格式的参数？** 例如 YAML、TOML 等
   - **决定：** 暂不支持，只支持 JSON 格式

2. **是否需要记录解析失败的日志？** 用于监控和调试
   - **决定：** 暂不记录，错误信息已足够详细

3. **是否需要添加配置选项？** 例如最大递归深度、是否允许某些格式等
   - **决定：** 暂不需要，使用合理的默认值即可
