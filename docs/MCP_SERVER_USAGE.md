# GODNSLOG MCP Server 使用文档

## 概述

`godnslog-mcp-server` 是 GODNSLOG 2.0 的 Model Context Protocol (MCP) 服务器，为 AI Agent 提供受控的 OAST 交互验证和证据收集能力。

## 安装

```bash
# 从源码构建
go build -o godnslog-mcp-server ./cmd/mcp-server

# 或使用 Go install
go install github.com/chennqqi/godnslog/cmd/mcp-server@latest
```

## 配置

MCP Server 通过环境变量配置：

```bash
export GODNSLOG_API_URL="http://localhost:8080"
export GODNSLOG_API_KEY="your-mcp-api-key"
```

## 启动

```bash
godnslog-mcp-server
```

服务器默认监听 `:8081` 端口。

## 可用工具

### 1. create_oast_probe

为 Agent 创建一次完整的 OAST 探针。该工具会先创建 Case，再创建绑定该 Case 的 Payload，并返回 `probe_id`、`case_id`、`payload_id`、`token` 和下一步动作提示。Agent 优先使用这个高层工具，只有需要精细控制时才分别调用 `create_case` 和 `create_payload`。

**参数**：
- `title` (string, required)：探针标题
- `template` (string, required)：Payload 模板
- `description` (string, optional)：Case 描述
- `target` (string, optional)：目标系统
- `variables` (object, optional)：模板变量
- `expected_protocols` (array of strings, optional)：期望回连协议，如 `dns`、`http`
- `expires_in` (string, optional)：过期时间（如 "1h"）

**示例**：
```json
{
  "title": "SSRF 探针",
  "target": "https://example.com",
  "template": "ssrf-url",
  "variables": {"path": "/callback"},
  "expected_protocols": ["dns", "http"],
  "expires_in": "1h"
}
```

### 2. create_case

创建一个新的测试 Case。

**参数**：
- `title` (string, required)：Case 标题
- `description` (string, optional)：Case 描述
- `target` (string, optional)：目标系统
- `tags` (array of strings, optional)：标签

**示例**：
```json
{
  "title": "SSRF 漏洞验证",
  "description": "验证目标系统的SSRF漏洞",
  "target": "example.com",
  "tags": ["ssrf", "oast"]
}
```

### 3. create_payload

生成一个可追踪的 Payload。

**参数**：
- `template` (string, required)：Payload 模板
- `case_id` (string, optional)：关联的 Case ID
- `variables` (object, optional)：模板变量
- `expires_in` (string, optional)：过期时间（如 "24h"）

**示例**：
```json
{
  "template": "http://{{.Token}}.example.com",
  "case_id": "case-123",
  "variables": {"target": "example.com"},
  "expires_in": "24h"
}
```

### 4. list_interactions

列出 Interaction 记录。

**参数**：
- `case_id` (string, optional)：按 Case ID 筛选
- `limit` (number, optional)：返回数量限制，默认 50

**示例**：
```json
{
  "case_id": "case-123",
  "limit": 100
}
```

### 5. wait_for_interaction

等待特定的 Interaction 发生。

**参数**：
- `token` (string, required)：要等待的 Token
- `timeout` (number, optional)：超时时间（秒），默认 300

**示例**：
```json
{
  "token": "your-token-here",
  "timeout": 600
}
```

### 6. summarize_evidence

汇总 Case 的证据。

**参数**：
- `case_id` (string, required)：Case ID

**示例**：
```json
{
  "case_id": "case-123"
}
```

### 7. export_report

导出 Case 报告。

**参数**：
- `case_id` (string, required)：Case ID
- `format` (string, optional)：报告格式（markdown/json/csv），默认 markdown

**示例**：
```json
{
  "case_id": "case-123",
  "format": "markdown"
}
```

### 8. revoke_token

撤销 API Key。

**参数**：
- `key_id` (string, required)：要撤销的 API Key ID

**示例**：
```json
{
  "key_id": "key-123"
}
```

## 安全性

### Agent API Key 权限控制

Sprint K 引入了 Agent API Key 权限控制系统，为 MCP 工具提供细粒度的权限和风险控制。

#### 权限 Scope

每个 MCP 工具都需要特定的权限 scope：

| 工具 | 必需 Scope | 风险等级 |
|------|-----------|---------|
| create_oast_probe | agent:create_probe | Medium |
| wait_for_interaction | agent:wait_interaction | Low |
| list_interactions | agent:read_interactions | Low |
| summarize_evidence | agent:summarize_evidence | Low |
| export_report | agent:export_report | Low |
| list_agent_runs | agent:read_runs | Low |
| get_agent_run | agent:read_runs | Low |
| revoke_token | agent:revoke_token | High |

#### 风险容忍度

Agent API Key 支持三种风险容忍度级别：

- **low**：仅允许低风险操作（list_interactions, wait_for_interaction, summarize_evidence, export_report, list_agent_runs, get_agent_run）
- **medium**：允许低风险和中风险操作（包括 create_oast_probe）
- **high**：允许所有操作（包括高风险的 revoke_token）

默认风险容忍度为 `medium`。

#### 权限拒绝

当 Agent 尝试执行超出其权限或风险容忍度的操作时：
- 操作会被拒绝
- 拒绝事件会记录到审计日志（action: `agent_permission.denied`）
- 返回明确的错误信息说明拒绝原因

#### 创建 Agent API Key

通过 Web UI 创建 Agent API Key 时：
- 必须选择 Agent Key 类型
- Scope 限制为 Agent-safe scopes
- 必须设置过期时间（默认 24 小时）
- 必须设置风险容忍度（默认 medium）

### 最小权限

MCP Server 使用的 API Key 应该具有最小权限：
- 只能创建和查看自己创建的资源
- 不能访问其他用户的 Case 和 Interaction
- 不能修改系统配置

### 过期时间

MCP API Key 应该设置合理的过期时间（如 24 小时）。Agent API Key 强制要求设置过期时间。

### 审计日志

所有 MCP 工具调用都会记录审计日志，包括：
- 调用时间
- 调用的工具
- 使用的参数
- 调用结果
- 权限拒绝事件（action: `agent_permission.denied`）

### 高风险动作限制

以下高风险动作默认禁用（需要 `high` 风险容忍度）：
- revoke_token（撤销 API Key）
- agent:delete_payload（删除 Payload）
- agent:modify_config（修改配置）

## 使用示例

### Claude AI 集成

在 Claude Desktop 的配置文件中添加：

```json
{
  "mcpServers": {
    "godnslog": {
      "command": "godnslog-mcp-server",
      "env": {
        "GODNSLOG_API_URL": "http://localhost:8080",
        "GODNSLOG_API_KEY": "your-mcp-api-key"
      }
    }
  }
}
```

### 典型工作流

```python
# Agent 使用 MCP 工具的典型工作流

# 1. 创建 OAST 探针
probe = await mcp.call_tool("create_oast_probe", {
    "title": "SSRF 漏洞验证",
    "target": "example.com",
    "template": "ssrf-url",
    "expected_protocols": ["dns", "http"],
    "expires_in": "1h"
})

token = probe["token"]

# 2. 在测试中使用 Payload
# ... 执行测试，将 Payload 注入到目标系统 ...

# 3. 等待 Interaction
interaction = await mcp.call_tool("wait_for_interaction", {
    "token": token,
    "timeout": 300
})

# 4. 汇总证据
evidence = await mcp.call_tool("summarize_evidence", {
    "case_id": probe["case_id"]
})

# 5. 导出报告
report = await mcp.call_tool("export_report", {
    "case_id": probe["case_id"],
    "format": "markdown"
})
```

## HTTP API

MCP Server 同时提供 HTTP API 端点：

### 列出可用工具

```bash
curl http://localhost:8081/
```

### 执行工具

```bash
curl -X POST http://localhost:8081/tool/create_case \
  -H "Content-Type: application/json" \
  -d '{
    "title": "SSRF 测试",
    "target": "example.com"
  }'
```

## 错误处理

工具执行失败时返回：

```json
{
  "success": false,
  "error": "错误描述"
}
```

常见错误：
- `API key is required`：未配置 API Key
- `title is required`：缺少必需参数
- `Tool not found`：工具不存在
- `Case not found`：Case 不存在

## 最佳实践

1. **使用专用 API Key**：为 MCP Server 创建专用的 API Key，设置合理的过期时间和权限
2. **设置合理超时**：`wait_for_interaction` 工具应该设置合理的超时时间，避免无限等待
3. **定期撤销 Token**：测试完成后，使用 `revoke_token` 撤销不再需要的 API Key
4. **审计日志**：定期检查审计日志，监控 MCP Server 的使用情况
5. **限制访问**：在防火墙层面限制 MCP Server 的访问，只允许受信任的 Agent 访问

## 故障排查

### 连接失败

检查环境变量是否正确设置：
```bash
echo $GODNSLOG_API_URL
echo $GODNSLOG_API_KEY
```

### 工具执行失败

查看 MCP Server 日志，检查：
- API Key 是否有效
- 参数是否正确
- 后端服务是否正常运行

### 超时问题

增加 `wait_for_interaction` 的超时时间，或检查网络连接。
