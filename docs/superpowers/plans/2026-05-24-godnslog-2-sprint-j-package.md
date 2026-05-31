# GODNSLOG 2.0 Sprint J Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint J
- **Sprint 主题**：Agent Run MVP & MCP Audit Binding
- **所属阶段**：Phase 5 - Agent Governance and Replay

## Sprint 背景

Sprint H / I 已把 Scanner Hub 从 Nuclei JSONL 分发材料推进到 Scanner Run 持久化、历史、详情、状态更新和审计。Scanner 侧已经具备：

`Scanner Hub -> Scanner Run -> Payload -> Interactions / Evidence -> Audit`

下一条产品主线应进入 Agent-Native，而不是继续扩大 Scanner Hub。当前项目已有：

- `docs/agent-native-specification.md`
- `docs/unified-control-plane.md`
- `docs/MCP_SERVER_USAGE.md`
- `internal/agentrun/model.go`
- `internal/mcp` 的基础工具

但现状仍有明显断点：

- `internal/agentrun` 只是薄模型，未进入 `/api/v2`
- MCP 的 `agent_run_id` 只是拼接字符串，不是持久化实体
- MCP 工具调用没有绑定真实 Agent Run
- Audit 不能按 `agent_id` / `agent_run_id` 回溯一次 Agent 任务
- 前端没有 Agent Runs 页面

Sprint J 的目标是补齐最小 Agent Run 闭环，让企业可以回答：

- 哪个 Agent 在什么时候发起了一次 OAST 任务
- 这个任务创建了哪个 Case / Payload
- 等待了哪些 Interaction
- 生成或查看了哪些 Evidence
- 所有动作是否能在 Audit 中按 Agent Run 查询

本 Sprint 不做完整 Agent 管理平台，不做 Workspace / RBAC 大改，不做真实 LLM 调用，不做 Agent 自动决策。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 建立持久化 Agent Run 模型和最小 API
2. MCP `create_oast_probe` 创建或绑定真实 Agent Run
3. MCP `wait_for_interaction` / `summarize_evidence` / `export_report` 能携带 Agent Run 上下文
4. Audit 记录 Agent Run 上下文，可按 `agent_run_id` 查询
5. Web 控制面提供 Agent Runs 列表和详情页，能回到 Case / Payload / Interactions / Evidence

## 输入文档

Windsurf 实施前必须完整阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/agent-native-specification.md`
- `docs/unified-control-plane.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/product-positioning.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-i-package.md`
- `docs/superpowers/acceptance/2026-05-24-godnslog-2-sprint-i-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/agentrun/model.go` 已有非常轻量的 `AgentRun`
- `internal/agentrun/store.go` 已有基础 store
- `internal/mcp` 已有 `create_oast_probe`、`wait_for_interaction`、`summarize_evidence`、`export_report`
- `AuditLog` 模型和 `/api/v2/audit/logs` 已存在
- Sprint I 已证明新增运行实体可以接入 schema sync、API、前端和 audit

### 主要缺口

- Agent Run 未进入 `internal/models`
- Agent Run 未进入 `db/init.go` schema sync
- Agent Run 无 `/api/v2/agent-runs`
- MCP 工具没有真实创建 Agent Run
- Agent Run 状态没有生命周期更新
- AuditLog 没有稳定 `agent_id` / `agent_run_id` 上下文
- Web 没有 Agent Runs 页面

## 术语边界

### Agent

Sprint J 不实现完整 Agent 管理实体。

本 Sprint 中的 `agent_id` 是外部传入或从 Agent API Key 派生的字符串标识，用于最小归因。完整 Agent 管理、Agent 创建、Agent 策略、Agent API Key 生命周期留给后续 Sprint。

### Agent Run

Agent Run 表示一次 Agent 任务执行生命周期。

最小状态：

- `created`
- `running`
- `waiting`
- `completed`
- `failed`
- `cancelled`
- `timed_out`

Sprint J 只要求支持：

- create
- list
- get detail
- append operation
- update status

不要求实现自动超时调度、取消执行、后台 worker 或任务恢复。

### Agent Operation

Agent Operation 是 Agent Run 内的一次动作记录，例如：

- `create_oast_probe`
- `wait_for_interaction`
- `summarize_evidence`
- `export_report`

可以作为 Agent Run 的 JSON 字段、独立表，或 audit 派生视图。优先选择与当前代码成本匹配的简单实现，但必须能在详情页看到 operation timeline。

## 数据契约

### AgentRun 最小字段

建议字段：

- `id`
- `agent_id`
- `operator_id`
- `case_id`
- `payload_id`
- `target`
- `title`
- `status`
- `started_at`
- `ended_at`
- `created_at`
- `updated_at`

详情可附加：

- `operations`
- `interaction_count`
- `last_interaction_at`
- `case_url`
- `payload_url`
- `interactions_url`
- `evidence_url`

### AgentOperation 最小字段

建议字段：

- `id`
- `agent_run_id`
- `agent_id`
- `action`
- `risk_level`
- `request`
- `result`
- `error`
- `started_at`
- `ended_at`
- `created_at`

如果不新建表，可用 `AgentRun.Operations` JSON 字段承载，但测试必须证明 operation 会被保存和返回。

### Audit 扩展

Audit 记录必须能表达：

- `agent_id`
- `agent_run_id`
- `action`
- `resource_type`
- `resource_id`
- `risk_level`
- `details`

如果 `models.AuditLog` 当前没有专用字段，可先放入 `details`，但必须在 `GET /api/v2/audit/logs` 的响应中可见，并支持按 `agent_run_id` 过滤或在 Agent Run detail 中聚合。

## 实施范围

### 1. Agent Run 后端模型与服务

目标是把 Agent Run 从概念变为可持久化实体。

建议新增或改造：

- `internal/models/agent_run.go`
- `internal/agentrun/service.go`
- `internal/agentrun/service_test.go`
- `db/init.go`

至少覆盖：

- 创建 Agent Run
- 获取 Agent Run
- 列表过滤 `agent_id` / `case_id` / `payload_id` / `status`
- 更新状态
- 追加 operation
- 派生 Interactions / Evidence 回链 URL

要求：

- Agent Run 必须进入生产 schema sync
- 不继续使用仅时间戳生成 ID 的弱 ID
- 状态流转必须有校验
- operation 保存失败不能静默吞掉

### 2. Agent Run API

建议接口：

```text
POST /api/v2/agent-runs
GET /api/v2/agent-runs
GET /api/v2/agent-runs/:id
PUT /api/v2/agent-runs/:id/status
POST /api/v2/agent-runs/:id/operations
```

API 验收点：

- 未认证请求必须失败
- 创建 run 后能 list / get
- list 支持 `agent_id` / `case_id` / `payload_id` / `status`
- status update 会写入 operation 或 audit
- append operation 后 detail 可见
- detail 包含 Case / Payload / Interactions / Evidence 回链

### 3. MCP 绑定真实 Agent Run

目标是让 MCP 不再返回拼接字符串 `agent_run_id`。

至少覆盖：

- `create_oast_probe`：
  - 接收 `agent_id`
  - 创建 Agent Run
  - 创建 Case
  - 创建 Payload
  - 回写 Agent Run 的 `case_id` / `payload_id`
  - 写入 operation：`create_oast_probe`
  - 返回真实 `agent_run_id`

- `wait_for_interaction`：
  - 接收可选 `agent_run_id`
  - 如果提供，写入 operation：`wait_for_interaction`
  - 将 Agent Run 状态从 `running` 更新为 `waiting`，命中后更新回 `running` 或 `completed`

- `summarize_evidence` / `export_report`：
  - 接收可选 `agent_run_id`
  - 写入对应 operation

要求：

- MCP 工具失败也要写入 operation 或 audit
- 不引入真实 LLM 调用
- 不让 MCP 绕过 `/api/v2` 的统一契约，除非是内部 service 测试场景

### 4. Audit Binding

目标是每个 Agent 关键动作都能回溯。

至少记录：

- `agent_run.created`
- `agent_run.status_updated`
- `agent_operation.create_oast_probe`
- `agent_operation.wait_for_interaction`
- `agent_operation.summarize_evidence`
- `agent_operation.export_report`

审计要求：

- 包含 `agent_id`
- 包含 `agent_run_id`
- 包含 `case_id` / `payload_id`，如已生成
- 包含 `risk_level`
- 不记录 API Key 明文
- 不记录敏感 token 全量，必要时只记录 payload_id / token hash / token prefix

### 5. 前端 Agent Runs 控制面

建议新增：

- `frontend-next/src/app/dashboard/agent-runs/page.tsx`
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`

至少覆盖：

- Agent Runs 列表
- 过滤 `agent_id` / `status`
- Detail 展示：
  - agent id
  - status
  - target
  - case id / payload id
  - operations timeline
  - Interactions / Evidence 链接
- Detail 可从 Case / Payload / Interactions / Evidence 回链

要求：

- 页面是运维/审计工作台，不是营销页
- 不做完整 Agent 管理页面
- 不做 Agent 创建表单
- 不做 Workspace 权限 UI

## 明确禁止越界

Sprint J 不允许实现：

- 完整 Agent 管理平台
- Agent 创建 / Agent 策略编辑 / Agent Marketplace
- Workspace / RBAC 大改
- 真实 LLM 调用或 Agent 自动决策
- 后台任务队列 / worker / 自动重试
- 删除 payload / 修改配置等高风险 Agent 动作
- Rebinding C2 Agent 化
- Scanner 调度、SARIF、多扫描器深度适配
- Webhook 平台
- 全量前端 lint 历史债清理

如实现中发现必须触碰上述内容，必须停下回传，不允许自行扩范围。

## 建议文件清单

后端：

- `internal/models/agent_run.go`
- `internal/agentrun/service.go`
- `internal/agentrun/service_test.go`
- `internal/agentrun/migration.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `db/init.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

前端：

- `frontend-next/src/app/dashboard/agent-runs/page.tsx`
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`

文档：

- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

## 测试要求

后端测试至少覆盖：

- Agent Run create / list / get
- status transition 校验
- append operation 后 detail 可见
- audit 写入 `agent_run.created`
- audit 写入 `agent_operation.create_oast_probe`
- schema sync 包含 Agent Run
- MCP `create_oast_probe` 返回真实 `agent_run_id`
- MCP `wait_for_interaction` 带 `agent_run_id` 时写入 operation

前端 E2E 至少覆盖：

- `/dashboard/agent-runs` 加载列表并请求 `GET /api/v2/agent-runs`
- 按 `agent_id` 或 `status` 过滤时真实进入 API query
- 打开 detail 请求 `GET /api/v2/agent-runs/:id`
- detail 展示 operations timeline
- detail 的 Interactions / Evidence 链接携带 `payload_id`
- 没有 `test.skip` / `test.only`

## 验证命令

Windsurf 完成后必须回传实际执行命令和结果。建议至少执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

注意：

- E2E 必须使用一次性、非交互式命令
- 不得执行 `npx playwright show-report`
- 不得触发 Playwright HTML report 常驻服务
- 如果 `webServer` 不稳定，继续使用 `docs/verification.md` 的两步法

## 完成定义

Sprint J 只有同时满足以下条件才算完成：

1. Agent Run 已持久化并进入生产 schema sync
2. Agent Run API 支持 create / list / get / status update / append operation
3. MCP `create_oast_probe` 返回真实 Agent Run ID
4. MCP 工具调用能写入 Agent Run operation
5. Audit 中能回溯 `agent_id` / `agent_run_id`
6. Agent Runs 页面能展示历史 run
7. Agent Run Detail 能展示 operation timeline
8. Detail 能回到 Case / Payload / Interactions / Evidence
9. 后端测试覆盖 audit 和 operation 持久化
10. E2E 覆盖列表、过滤、详情、operation timeline、回链
11. 没有 skip / only / 静态空测
12. 未越界到完整 Agent 管理、Workspace/RBAC、真实 LLM、任务队列或高风险动作
13. `docs/verification.md` 记录 Sprint J 实际验证命令和结果

## Windsurf 回传要求

Windsurf 完成 Sprint J 后，请回传：

- 本地 commit hash
- 修改文件列表
- Agent Run 数据字段说明
- Agent Operation 数据字段说明
- API 路由清单
- MCP 工具变更清单
- Audit event 名称与触发点
- 已执行验证命令和结果
- 未完成项或刻意延后的能力

