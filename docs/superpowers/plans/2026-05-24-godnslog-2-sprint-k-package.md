# GODNSLOG 2.0 Sprint K Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint K
- **Sprint 主题**：Agent API Key Permission Gate & MCP Safety Controls
- **所属阶段**：Phase 5 - Agent Governance and Replay

## Sprint 背景

Sprint J 已经完成 Agent Run MVP：

`MCP Tool -> Agent Run -> Agent Operation -> Case / Payload -> Interactions / Evidence -> Audit`

这解决了“Agent 做过什么”的可见性问题。下一步不应立刻做完整 Agent 管理平台，而应补齐 Agent-Native 的安全边界：

- Agent API Key 是否真的能被标记为 Agent 身份
- Agent Key 是否只能拥有允许的最小 scope
- MCP 工具是否根据 scope 执行权限校验
- 高风险工具是否默认对 Agent 禁用
- 权限拒绝是否写入 audit，方便安全团队回溯

当前仓库已有一些基础：

- `internal/models/apikey.go` 已有 `is_agent`、`workspace_id`、`risk_tolerance`、`AgentScopes`
- `internal/auth/service.go` 已有 `CreateAPIKey` 和 Agent scope 校验雏形
- `internal/auth/middleware.go` 已能识别 API Key identity
- `/api/v2/apikeys` 已存在
- `internal/mcp/server.go` 已有 MCP 工具和 Agent Run 绑定
- `frontend-next/src/app/dashboard/apikeys/page.tsx` 已有基础 API Key 页面

但仍有断点：

- `/api/v2/apikeys` 当前主要使用 legacy `models.TblAPIKey` 路径，Agent 字段和 scope 校验不稳定
- API Key 页面不能明确创建 Agent Key、选择 Agent-safe scopes、设置风险容忍度或过期时间
- MCP 工具没有统一的 scope gate
- `revoke_token` 等高风险能力对 Agent 缺少默认禁用和显式授权机制
- 权限拒绝没有稳定 audit 事件
- E2E 未证明 Agent Key 创建与 MCP scope 拒绝链路

Sprint K 的目标是补齐“可给 Agent 发 key，但默认安全”的最小闭环。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 统一 `/api/v2/apikeys` 与 `internal/models.APIKey` 的 Agent 字段契约
2. 支持创建 Agent API Key，默认只能选择 Agent-safe scopes
3. 为 MCP 工具增加统一 scope / risk gate
4. 对 Agent 高风险动作默认拒绝，并写入 audit
5. Web API Keys 页面支持 Agent Key 创建、展示、撤销和 E2E 覆盖

本 Sprint 不做完整 Agent 管理平台，不做 Agent Marketplace，不做 Workspace/RBAC 大改，不做真实 LLM，不做后台任务队列。

## 输入文档

Windsurf 实施前必须完整阅读：

- `docs/agent-native-specification.md`
- `docs/unified-control-plane.md`
- `docs/product-positioning.md`
- `docs/implementation-dependencies.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-j-package.md`
- `docs/superpowers/acceptance/2026-05-24-godnslog-2-sprint-j-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/models.APIKey` 已有：
  - `scopes`
  - `is_agent`
  - `workspace_id`
  - `risk_tolerance`
  - `expires_at`
  - `last_used_at`
  - `is_revoked`
- `internal/models.AgentScopes` 已有最小 Agent scope 列表雏形
- `internal/auth.Service.CreateAPIKey` 已校验 Agent scopes
- `internal/auth.AuthMiddleware` 能从 Bearer / API key header 生成 identity
- `/dashboard/apikeys` 已有基础创建、编辑、删除 UI
- MCP tools 已通过 `/api/v2` 调用主系统

### 主要缺口

- `/api/v2/apikeys` 的 create / list / get / update / delete 未稳定使用 `internal/auth.Service`
- Agent Key 创建后是否真正写入 `is_agent`、`risk_tolerance`、`workspace_id` 不可靠
- API Key 页面缺少 Agent 专用模式和 Agent-safe scope 说明
- MCP server 没有统一的 `requireScope(toolName)` 或等价机制
- 高风险工具没有明确风险等级和默认拒绝策略
- 权限拒绝没有 `agent_permission.denied` audit

## 术语边界

### Agent API Key

Agent API Key 是给 AI Agent / MCP client 使用的最小权限 key。

最小字段：

- `id`
- `name`
- `key_prefix`
- `scopes`
- `is_agent`
- `risk_tolerance`
- `workspace_id`
- `expires_at`
- `last_used_at`
- `is_revoked`
- `created_by`
- `created_at`

创建响应中可以返回一次性明文 key；list / get 不得返回完整明文 key。

### Agent Scope

Sprint K 要求收敛并使用统一 Agent scope 命名。

建议最小 Agent-safe scopes：

- `agent:create_probe`
- `agent:wait_interaction`
- `agent:read_interactions`
- `agent:summarize_evidence`
- `agent:export_report`
- `agent:read_runs`

高风险 scopes 必须默认不出现在 Agent Key 创建 UI 的普通选择区：

- `agent:revoke_token`
- `agent:delete_payload`
- `agent:modify_config`

如果保留旧 scope 名称兼容，必须提供明确映射并有测试证明。

### Risk Gate

每个 MCP 工具必须定义风险等级：

- low：只读或等待，例如 `list_interactions`、`wait_for_interaction`
- medium：创建新资源，例如 `create_oast_probe`、`create_case`、`create_payload`
- high：撤销、删除、修改配置，例如 `revoke_token`

Agent Key 的 `risk_tolerance` 决定可执行的最高风险等级。默认应为 `medium` 或更低，不能默认允许 high。

## 实施范围

### 1. Agent API Key 后端契约

目标是让 `/api/v2/apikeys` 真实支持 Agent Key 字段和校验。

建议改造：

- `server/v2_api.go`
- `internal/auth/service.go`
- `internal/auth/apikey_test.go`
- `internal/auth/middleware_test.go`
- `internal/models/apikey.go`

要求：

- 创建 Agent Key 时必须校验 scope 均属于允许列表
- Agent Key 必须强制过期时间；如果请求未传，后端设置合理默认过期时间
- Agent Key 创建响应只在创建时返回明文 key
- list / get 只返回 `key_prefix`，不得返回完整明文 key
- revoke 必须写 audit
- create / update / revoke 的 audit details 必须包含：
  - `api_key_id`
  - `key_prefix`
  - `is_agent`
  - `scopes`
  - `risk_tolerance`
  - 不包含完整明文 key

### 2. MCP Scope / Risk Gate

目标是所有 MCP 工具执行前进入统一权限校验。

建议新增或改造：

- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- 可新增 `internal/mcp/permissions.go`

要求：

- MCP Server 能拿到当前 API Key identity 或至少可注入测试 identity
- 每个 MCP tool 声明 required scope 和 risk level
- 缺少 scope 时返回失败，不执行后续 API 调用
- risk 超过 Agent Key tolerance 时返回失败，不执行后续 API 调用
- 对非 Agent 管理员 key 可保持现有行为，但必须有测试说明
- 权限拒绝写 audit：
  - action: `agent_permission.denied`
  - resource_type: `mcp_tool`
  - details 包含 `tool_name`、`required_scope`、`risk_level`、`agent_id` 或 `api_key_id`

### 3. 高风险动作默认禁用

目标是让 Agent 默认不能执行破坏性动作。

Sprint K 至少覆盖：

- `revoke_token`

要求：

- Agent Key 缺少 `agent:revoke_token` 时不能执行
- Agent Key 有 `agent:revoke_token` 但 `risk_tolerance` 低于 high 时不能执行
- 拒绝时不能调用 `/api/v2/apikeys/:id`
- 拒绝要写 audit
- 管理员或非 Agent 高权限 key 的兼容路径必须明确测试

不要求本 Sprint 新增 `delete_payload` 或 `modify_config` 工具。

### 4. API Keys 前端 Agent Key 模式

目标是让安全团队能从 Web 创建和管理 Agent Key。

建议改造：

- `frontend-next/src/app/dashboard/apikeys/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/apikeys.spec.ts`

要求：

- 创建表单有 Agent Key 开关
- 开启 Agent Key 时：
  - 只展示 Agent-safe scopes
  - 展示 `risk_tolerance` 选择
  - 要求或默认设置 `expires_at`
- 列表展示：
  - Agent / Human 标识
  - key prefix
  - scopes
  - risk tolerance
  - expires_at / last_used_at
  - revoked 状态或撤销入口
- 创建成功后只显示一次明文 key
- 页面不得显示历史完整明文 key
- 撤销 Agent Key 后列表更新，并真实调用 revoke API

### 5. 文档与验证记录

目标是让安全边界可审计、可复验。

建议更新：

- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

要求：

- 文档列出 MCP tool -> required scope -> risk level 映射
- 文档说明 Agent Key 创建示例
- 文档说明高风险动作默认拒绝
- `docs/verification.md` 只记录真实执行过的命令和结果，不得预写通过

## 明确禁止越界

Sprint K 不允许实现：

- 完整 Agent 管理平台
- Agent 创建 / Agent 策略编辑 / Agent Marketplace
- Workspace / RBAC 大改
- 真实 LLM 调用或 Agent 自动决策
- 后台任务队列 / worker / 自动重试
- 新增删除 Payload / 修改系统配置等高风险 MCP 工具
- Scanner Hub / SARIF / 多扫描器扩展
- Webhook 平台
- 全量前端 lint 历史债清理
- API Key 加密存储的大迁移，除非当前改动无法安全实现；如必须触碰需先回传

如实现中发现必须触碰上述内容，必须停下回传，不允许自行扩范围。

## 建议文件清单

后端：

- `internal/models/apikey.go`
- `internal/auth/service.go`
- `internal/auth/apikey_test.go`
- `internal/auth/middleware.go`
- `internal/auth/middleware_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `server/v2_api.go`
- `server/v2_api_test.go`

前端：

- `frontend-next/src/app/dashboard/apikeys/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/apikeys.spec.ts`

文档：

- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

## 测试要求

后端测试至少覆盖：

- 创建 Agent Key 成功，返回一次性明文 key
- list / get 不返回完整明文 key
- Agent Key 使用非法 scope 创建失败
- Agent Key 默认或强制有过期时间
- revoke Agent Key 写 audit，且 audit 不包含完整明文 key
- MCP `create_oast_probe` 缺少 `agent:create_probe` 时被拒绝且不调用下游 API
- MCP `wait_for_interaction` 缺少 `agent:wait_interaction` 时被拒绝
- MCP `summarize_evidence` 缺少 `agent:summarize_evidence` 时被拒绝
- MCP `export_report` 缺少 `agent:export_report` 时被拒绝
- MCP `revoke_token` 对普通 Agent Key 默认拒绝
- MCP 权限拒绝写入 `agent_permission.denied` audit
- 非 Agent 高权限 key 或 admin scope 的兼容路径通过

前端 E2E 至少覆盖：

- `/dashboard/apikeys` 加载并请求 `GET /api/v2/apikeys`
- 创建 Agent Key 时请求体包含 `is_agent=true`、Agent scopes、`risk_tolerance`、`expires_at`
- 创建成功后只显示一次明文 key
- 列表展示 Agent 标识、prefix、scope、risk、expires_at
- Human Key 模式不会展示 Agent-only 默认 scope
- 撤销 Agent Key 调用 revoke API 并刷新列表
- 没有 `test.skip` / `test.only` / 只检查静态文字的空测

## 验证命令

Windsurf 完成后必须回传实际执行命令和结果。建议至少执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/auth ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts e2e/agent-runs.spec.ts
```

注意：

- E2E 必须使用一次性、非交互式命令
- 不得执行 `npx playwright show-report`
- 不得触发 Playwright HTML report 常驻服务
- 如果 `webServer` 不稳定，继续使用 `docs/verification.md` 的两步法

## 完成定义

Sprint K 只有同时满足以下条件才算完成：

1. `/api/v2/apikeys` 真实支持 `is_agent`、Agent scopes、`risk_tolerance`、`expires_at`
2. Agent Key scope 校验在后端生效
3. Agent Key list / get 不泄露完整明文 key
4. Agent Key create / revoke 写 audit，且 audit 不包含完整明文 key
5. MCP tools 统一声明 required scope 和 risk level
6. MCP 缺 scope 或超 risk 时拒绝执行，且不调用后续业务 API
7. `revoke_token` 对 Agent 默认禁用，只有显式 high scope + high risk tolerance 才允许
8. 权限拒绝写入 `agent_permission.denied` audit
9. API Keys 页面支持 Agent Key 创建、展示和撤销
10. E2E 覆盖 Agent Key 创建、展示、撤销和 no full key leakage
11. E2E 无 skip / only / 静态空测
12. 未越界到完整 Agent 管理、Workspace/RBAC、真实 LLM、任务队列、Scanner 扩展
13. `docs/verification.md` 记录 Sprint K 实际验证命令和结果

## Windsurf 回传要求

Windsurf 完成 Sprint K 后，请回传：

- 本地 commit hash
- 修改文件列表
- Agent scope 列表
- MCP tool -> scope -> risk 映射
- API Key 创建 / 列表 / 撤销契约说明
- Audit event 名称与触发点
- 已执行验证命令和结果
- 未完成项或刻意延后的能力

## Codex 验收重点

Codex 复验 Sprint K 时重点检查：

1. Agent Key 是否真的进入后端模型和 `/api/v2/apikeys` 响应，而不是只在前端模拟
2. 非法 Agent scope 是否被后端拒绝
3. list / get / audit 是否没有泄露完整明文 key
4. MCP 权限拒绝是否发生在业务 API 调用之前
5. `agent_permission.denied` audit 是否真实写入
6. `revoke_token` 是否默认对 Agent 禁用
7. E2E 是否断言 API 请求体、响应展示和撤销行为
8. 是否严格没有越界到完整 Agent 管理、Workspace/RBAC、LLM 或 Scanner 扩展

只有上述全部成立，Sprint K 才能关闭。
