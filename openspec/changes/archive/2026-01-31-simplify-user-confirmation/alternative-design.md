# Alternative Design: LLM-Assisted Risk Assessment

## 方案C：LLM在Tool Call中提供Risk Level（推荐）

### 核心思路
- LLM在调用工具时，在arguments中添加可选的`risk_level`字段
- 代码层优先使用LLM提供的risk_level，如果没有则做保守判断
- 结合代码层的基础规则和LLM的智能判断

### 实现方式

#### 1. Tool Definition扩展
在所有工具的parameters中添加可选的`risk_level`字段：

```json
{
  "type": "object",
  "properties": {
    "sql": {...},
    "risk_level": {
      "type": "string",
      "enum": ["low", "medium", "high"],
      "description": "Optional: Risk level assessment for this operation. 'low' = safe to execute automatically, 'medium'/'high' = requires user confirmation. If not provided, system will assess risk conservatively."
    }
  }
}
```

#### 2. Prompt指导
在prompt中指导LLM：
- 对于明显安全的操作（SELECT, SHOW, ls, read），设置`risk_level: "low"`
- 对于明显危险的操作（DROP, TRUNCATE, rm, write），设置`risk_level: "high"`
- 对于不确定的操作（init, reboot, 自定义脚本），设置`risk_level: "high"`或询问用户

#### 3. 代码层处理逻辑
```go
func assessRisk(toolCall, args) RiskDecision {
    // 1. 检查LLM提供的risk_level
    if riskLevel, ok := args["risk_level"].(string); ok {
        switch riskLevel {
        case "low":
            return RiskLow  // 直接执行
        case "medium", "high":
            return RiskHigh  // 需要确认
        }
    }
    
    // 2. LLM没有提供risk_level，代码层做保守判断
    // 只对明确安全的操作（白名单）直接执行
    if isWhitelisted(toolCall, args) {
        return RiskLow
    }
    
    // 3. 其他情况默认需要确认（保守策略）
    return RiskHigh
}
```

### 优势
- ✅ LLM可以智能判断未知命令的风险
- ✅ 代码层有基础规则保障（白名单）
- ✅ 不需要穷举所有危险命令
- ✅ LLM判断错误时，代码层保守策略保证安全
- ✅ 符合Agent流程：LLM做决策，代码层执行

### 劣势
- ⚠️ 需要修改所有tool definitions
- ⚠️ LLM可能不总是提供risk_level（需要fallback）

---

## 方案D：两阶段Tool Call（LLM先评估风险）

### 核心思路
- LLM在调用实际工具前，先调用一个`assess_risk`工具
- `assess_risk`工具返回风险级别
- 代码层根据风险级别决定是否确认

### 实现方式

#### 1. 新增assess_risk工具
```go
{
  "name": "assess_risk",
  "description": "Assess the risk level of a tool operation before execution",
  "parameters": {
    "tool_name": "string",
    "tool_args": "object"
  }
}
```

#### 2. 流程
```
1. LLM调用assess_risk工具
2. assess_risk返回risk_level
3. 代码层根据risk_level决定：
   - low → 直接执行原工具
   - high → 要求确认后执行
4. LLM调用实际工具
```

### 优势
- ✅ 分离关注点：风险评估和执行分离
- ✅ LLM可以充分评估风险

### 劣势
- ⚠️ 增加一次LLM调用，影响性能
- ⚠️ 增加复杂度
- ⚠️ 可能被LLM跳过（不调用assess_risk）

---

## 方案E：代码层异步询问LLM（不推荐）

### 核心思路
- 代码层先做基础判断
- 如果无法确定，异步调用LLM快速评估风险
- 根据LLM评估结果决定是否确认

### 劣势
- ❌ 需要异步调用LLM，影响性能
- ❌ 增加延迟
- ❌ 实现复杂

---

## 方案F：LLM返回文本询问用户（符合Agent流程）

### 核心思路
- LLM在调用工具前，如果不确定风险，可以返回文本询问用户
- 用户确认后，LLM再调用工具
- 代码层只处理明确的情况（白名单直接执行，其他需要确认）

### 实现方式

#### Prompt指导
```
<RISK_ASSESSMENT>
When calling tools, assess the risk level:
- **Low-risk operations** (SELECT, SHOW, ls, read, GET): Call tool directly with risk_level="low"
- **High-risk operations** (DROP, TRUNCATE, rm, write, DELETE): Call tool with risk_level="high" or ask user first
- **Uncertain operations** (init, reboot, custom scripts): 
  - Option 1: Return text asking user for confirmation before calling tool
  - Option 2: Call tool with risk_level="high" (system will ask for confirmation)
  
If you're uncertain about an operation's risk, you can:
1. Return text asking user: "This operation (init system) may be risky. Proceed?"
2. Or call tool with risk_level="high" and let system handle confirmation
</RISK_ASSESSMENT>
```

#### 代码层处理
```go
// 如果LLM返回文本询问用户（没有tool_calls）
if len(message.ToolCalls) == 0 && message.Content != "" {
    // 检查是否是风险询问
    if isRiskConfirmationQuestion(message.Content) {
        // 显示给用户，等待用户确认
        // 用户确认后，LLM可以继续调用工具
        return message.Content, nil, nil
    }
}
```

### 优势
- ✅ 符合Agent流程：LLM可以做决策
- ✅ 灵活：LLM可以选择询问或直接调用
- ✅ 不需要修改tool definitions
- ✅ 代码层保持简单

### 劣势
- ⚠️ 可能违反"必须调用工具"的原则（但可以允许在不确定时询问）

---

## 推荐方案：方案C（LLM在Tool Call中提供Risk Level）

结合方案C和方案F的优点：
1. **主要方式**：LLM在tool call的arguments中提供`risk_level`
2. **备选方式**：如果LLM不确定，可以返回文本询问用户（允许这种例外）
3. **代码层保障**：代码层有基础白名单，即使LLM不提供risk_level也能工作

### 最终流程

```
1. LLM决定调用工具
   ↓
2. LLM在arguments中添加risk_level（可选）
   ↓
3. 代码层评估：
   - 如果LLM提供了risk_level="low" → 直接执行
   - 如果LLM提供了risk_level="high" → 要求确认
   - 如果LLM没有提供risk_level：
     - 代码层白名单 → 直接执行
     - 其他 → 要求确认（保守策略）
   ↓
4. 如果需要确认：
   - 显示操作内容
   - 询问用户确认
   - 用户确认后执行
```

### 特殊情况处理

如果LLM不确定操作风险，可以：
- **方式1**：调用工具时设置`risk_level="high"`，让系统要求确认
- **方式2**：返回文本询问用户："This operation (init system) may be risky. Should I proceed?"
  - 用户确认后，LLM再调用工具
  - 这是Agent流程的一部分，允许LLM在不确定时询问用户
