'use client'

import { DocPageLayout } from '../doc-page-layout'

const en = `## Authentication

All API requests require a JWT token via the \`Access-Token\` header:

\`\`\`bash
POST /api/v2/auth/login
Content-Type: application/json

{"username": "admin", "password": "your-password"}
\`\`\`

Response:

\`\`\`json
{
  "code": 0,
  "data": {
    "token": "eyJ...",
    "user": { "id": 1, "username": "admin", "role": 0 }
  },
  "message": "OK"
}
\`\`\`

## Core Endpoints

### Cases

- \`GET /api/v2/cases\` — List cases (paginated)
- \`POST /api/v2/cases\` — Create a case
- \`GET /api/v2/cases/:id\` — Get case detail
- \`PUT /api/v2/cases/:id\` — Update case
- \`DELETE /api/v2/cases/:id\` — Delete case
- \`GET /api/v2/cases/stats\` — Case statistics

### Payloads

- \`GET /api/v2/payloads\` — List payloads
- \`POST /api/v2/payloads\` — Create payload
- \`POST /api/v2/payloads/batch\` — Batch create
- \`GET /api/v2/payloads/:id\` — Get payload
- \`PUT /api/v2/payloads/:id\` — Update payload
- \`POST /api/v2/payloads/:id/revoke\` — Revoke payload

### Interactions

- \`GET /api/v2/interactions\` — List interactions
- \`GET /api/v2/interactions/stats\` — Aggregated stats
- \`GET /api/v2/interactions/stream\` — SSE real-time stream
- \`POST /api/v2/interactions/export\` — Export (JSON/Markdown/CSV)
- \`GET /api/v2/poll\` — Burp Collaborator-style cursor polling

### Scanner Hub & Search

- \`GET /api/v2/scanner-hub/adapters\` — List scanner adapters
- \`GET /api/v2/search/{zoomeye,shodan,fofa}\` — Search engines
- \`POST /api/v2/scanner-runs\` — Create scanner run
- \`POST /api/v2/scanner-runs/from-search\` — Create runs from search results

### Other

- \`/api/v2/apikeys\` — API Key management
- \`/api/v2/canary\` — Canary tokens
- \`/api/v2/rebinding\` — DNS rebinding
- \`/api/v2/listeners\` — Protocol listeners
- \`/api/v2/notifications/channels\` — Notification channels
- \`/api/v2/retention/policies\` — Data retention
- \`/api/v2/agent-runs\` — Agent run tracking
- \`/api/v2/evidence/generate\` — Evidence generation
- \`/api/v2/settings\` — System settings
- \`/api/v2/cluster\` — HA cluster management

### MCP (JSON-RPC 2.0)

\`POST /api/v2/mcp\` — Streamable HTTP transport for AI Agent integration. 13 tools available.

## Swagger UI

Full interactive API docs at \`/swagger/index.html\` (requires \`-swagger\` flag).`

const zh = `## 认证

所有 API 请求需要通过 \`Access-Token\` 头传递 JWT：

\`\`\`bash
POST /api/v2/auth/login
Content-Type: application/json

{"username": "admin", "password": "your-password"}
\`\`\`

响应：

\`\`\`json
{
  "code": 0,
  "data": {
    "token": "eyJ...",
    "user": { "id": 1, "username": "admin", "role": 0 }
  },
  "message": "OK"
}
\`\`\`

## 核心端点

### Cases

- \`GET /api/v2/cases\` — 案例列表（分页）
- \`POST /api/v2/cases\` — 创建案例
- \`GET /api/v2/cases/:id\` — 案例详情
- \`PUT /api/v2/cases/:id\` — 更新案例
- \`DELETE /api/v2/cases/:id\` — 删除案例

### Payloads

- \`GET /api/v2/payloads\` — 载荷列表
- \`POST /api/v2/payloads\` — 创建载荷
- \`POST /api/v2/payloads/batch\` — 批量创建
- \`GET /api/v2/payloads/:id\` — 载荷详情
- \`PUT /api/v2/payloads/:id\` — 更新载荷
- \`POST /api/v2/payloads/:id/revoke\` — 撤销载荷

### Interactions

- \`GET /api/v2/interactions\` — 交互列表
- \`GET /api/v2/interactions/stats\` — 聚合统计
- \`GET /api/v2/interactions/stream\` — SSE 实时流
- \`POST /api/v2/interactions/export\` — 导出
- \`GET /api/v2/poll\` — 游标轮询（Burp Collaborator 风格）

### Scanner Hub & 搜索引擎

- \`GET /api/v2/scanner-hub/adapters\` — 扫描器适配器列表
- \`GET /api/v2/search/{zoomeye,shodan,fofa}\` — 搜索引擎
- \`POST /api/v2/scanner-runs\` — 创建扫描运行
- \`POST /api/v2/scanner-runs/from-search\` — 从搜索结果创建

### 其他

- \`/api/v2/apikeys\` — API 密钥管理
- \`/api/v2/canary\` — Canary 令牌
- \`/api/v2/rebinding\` — DNS 重绑定
- \`/api/v2/listeners\` — 协议监听器
- \`/api/v2/notifications/channels\` — 通知渠道
- \`/api/v2/retention/policies\` — 数据保留
- \`/api/v2/agent-runs\` — Agent 运行追踪
- \`/api/v2/evidence/generate\` — 证据生成
- \`/api/v2/settings\` — 系统设置
- \`/api/v2/cluster\` — HA 集群管理

### MCP（JSON-RPC 2.0）

\`POST /api/v2/mcp\` — 面向 AI Agent 集成的 Streamable HTTP 传输，提供 13 个工具。

## Swagger UI

完整的交互式 API 文档在 \`/swagger/index.html\`（需 \`-swagger\` 启动参数）。`

export default function ApiDoc() {
  return <DocPageLayout titleKey="docs.api" mdEn={en} mdZh={zh} />
}
