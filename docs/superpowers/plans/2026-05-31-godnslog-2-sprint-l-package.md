# GODNSLOG 2.0 Sprint L Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint L
- **Sprint 主题**：Agent Run Review Packet & Evidence Export Closure
- **所属阶段**：Phase 5 - Agent Governance and Replay

## Sprint 背景

Sprint J 已建立 Agent Run MVP：

`MCP Tool -> Agent Run -> Agent Operation -> Case / Payload -> Interactions / Evidence -> Audit`

Sprint K 已补齐 Agent API Key 与 MCP scope / risk gate：

`Agent API Key -> MCP Permission Gate -> Agent Run / Operation -> Audit`

现在 Agent-Native 主链路已经能回答“Agent 做过什么”和“是否有权限做”。下一步不应扩大到完整 Agent 管理平台，也不应切回 Scanner Hub，而应补齐安全团队真正复核一次 Agent 任务时需要的最小证据包：

- 这次 Agent Run 的 Case / Payload / Interaction / Evidence 是否能聚合为一个稳定 review packet
- packet 是否能导出为 JSON / Markdown，且继续使用统一 Evidence 契约
- Agent Run detail 是否能一键查看证据摘要、导出复核包、回到 Evidence 页面
- 导出动作是否进入 Agent Operation 和 Audit，便于审计

Sprint L 的目标是把 Agent Run 从“运行记录”推进到“可复盘证据包”，但不实现完整报告中心、批量导出、生命周期治理或 Scanner 扩展。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 为单个 Agent Run 生成 review packet，聚合 run、operations、case、payload、interactions summary、structured evidence、audit references。
2. 新增最小 `/api/v2/agent-runs/:id/review` API，支持 `format=json|markdown`。
3. MCP `export_report` 使用同一 review packet 契约，并把导出动作记录到 Agent Operation 和 Audit。
4. Agent Runs 详情页增加 review packet 区块和导出入口，保持 Evidence 回链闭环。
5. 补后端测试和 E2E，证明 review packet 使用统一 Evidence 契约、没有越界到 Scanner Hub / 生命周期治理 / 批量操作。

本 Sprint 不做完整 Agent 管理平台，不做真实 LLM，不做后台任务队列，不做批量导出，不做 Scanner Hub 新集成。

## 输入文档

Windsurf 实施前必须完整阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/unified-control-plane.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-j-package.md`
- `docs/superpowers/acceptance/2026-05-24-godnslog-2-sprint-j-acceptance.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-k-package.md`
- `docs/superpowers/acceptance/2026-05-25-godnslog-2-sprint-k-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/agentrun/service.go` 已有 Agent Run create / list / get / status update / append operation。
- `server/v2_api.go` 已有 `/api/v2/agent-runs` 系列接口。
- `internal/mcp/server.go` 已有 `summarize_evidence` 和 `export_report`，并会写 Agent Operation。
- `internal/interaction/evidence_service.go` 已有统一 Evidence 结构和 JSON / Markdown / CSV 生成。
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已有 Agent Run 详情、operation timeline、Case / Payload / Interactions / Evidence 回链。
- `frontend-next/e2e/agent-runs.spec.ts` 已覆盖列表、详情、状态更新和回链。

### 主要缺口

- Agent Run 详情页只能看运行和 operation，不能直接看到可复核证据包。
- MCP `export_report` 当前主要调用 `/api/v2/evidence/generate`，没有形成 Agent Run 级别 review packet。
- Agent Run API 没有单 run review / export 契约。
- 导出动作没有稳定表达 `agent_run_id`、`format`、`evidence_strength`、`interaction_count` 等复核字段。
- E2E 未证明 Agent Run detail -> Evidence -> Review Packet -> Export 的闭环。

## 术语边界

### Agent Run Review Packet

Review Packet 是针对单个 Agent Run 的复核包。它不是完整报告系统，也不是批量导出。

最小语义：

- `agent_run`
- `operations`
- `case_id`
- `payload_id`
- `target`
- `interaction_summary`
- `evidence`
- `audit_refs`
- `generated_at`
- `format`

`evidence` 必须复用 `internal/interaction.Evidence` 语义，不允许重新造一个不兼容的证据结构。

### Review Export

Review Export 是把单个 Agent Run 的 Review Packet 输出为 JSON 或 Markdown。

允许格式：

- `json`
- `markdown`

本 Sprint 不实现 PDF、DOCX、SARIF、ZIP、批量导出或长期归档。

### Audit Reference

Audit Reference 是 review packet 中用于回溯的审计摘要，不是新的审计模型。

最小字段：

- `id`
- `action`
- `resource_type`
- `resource_id`
- `timestamp`

如果当前 audit 查询按 `agent_run_id` 过滤能力不足，可以在 service 层按 `details.agent_run_id` 或 `resource_id` 做最小聚合，但必须有测试证明 packet 能包含相关审计引用。

## 数据契约

### AgentRunReviewPacket

建议新增到 `internal/models/agent_run.go` 或与 Agent Run 模型同文件：

```go
type AgentRunReviewPacket struct {
	ID                 string                 `json:"id"`
	AgentRun          AgentRunDetail          `json:"agent_run"`
	CaseID             string                 `json:"case_id,omitempty"`
	PayloadID          string                 `json:"payload_id,omitempty"`
	Target             string                 `json:"target,omitempty"`
	InteractionSummary AgentRunInteractionSummary `json:"interaction_summary"`
	Evidence           *interaction.Evidence  `json:"evidence,omitempty"`
	AuditRefs          []AgentRunAuditRef      `json:"audit_refs"`
	GeneratedAt        time.Time              `json:"generated_at"`
	Format             string                 `json:"format"`
	Content            string                 `json:"content,omitempty"`
}
```

如果直接 import `internal/interaction` 会造成包循环，应把 Review Packet 放在 `internal/agentrun` 或定义仅用于 JSON 的 Evidence payload wrapper，但响应中的 `evidence` 字段必须来自现有 EvidenceService 的结果。

### AgentRunInteractionSummary

```go
type AgentRunInteractionSummary struct {
	Total             int        `json:"total"`
	DNSCount          int        `json:"dns_count"`
	HTTPCount         int        `json:"http_count"`
	UniqueSources     int        `json:"unique_sources"`
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"`
}
```

### AgentRunAuditRef

```go
type AgentRunAuditRef struct {
	ID           string    `json:"id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}
```

### API Response

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "id": "review-agent-run-1",
    "agent_run": {},
    "case_id": "case-1",
    "payload_id": "payload-1",
    "target": "https://target.example",
    "interaction_summary": {
      "total": 2,
      "dns_count": 1,
      "http_count": 1,
      "unique_sources": 1,
      "last_interaction_at": "2026-05-31T10:00:00Z"
    },
    "evidence": {},
    "audit_refs": [],
    "generated_at": "2026-05-31T10:01:00Z",
    "format": "json",
    "content": ""
  }
}
```

For `format=markdown`, `content` must contain a Markdown report and `evidence` must still be present as structured data.

## 实施范围

### 1. Agent Run Review Service

目标是用已有 Agent Run / Evidence / Audit 契约生成单 run review packet。

建议新增或改造：

- `internal/agentrun/review.go`
- `internal/agentrun/review_test.go`
- `internal/agentrun/service.go`
- `internal/models/agent_run.go`
- `internal/interaction/evidence_service.go` 仅在必要时补可复用方法

要求：

- `BuildReviewPacket(agentRunID, format, baseURL)` 根据 Agent Run 找到 Case / Payload 范围。
- 有 `payload_id` 时优先按 payload 生成 evidence；无 `payload_id` 但有 `case_id` 时按 case 生成 evidence。
- 没有 Interaction 时返回 packet，但 `evidence` 可以为空，不能 500。
- Interaction summary 必须从真实 `models.Interaction` 聚合，不能只读 AgentRun detail 的 count。
- Markdown content 必须包含 Agent Run ID、Case ID、Payload ID、Target、Evidence Strength、Interaction Count、Operations timeline。
- JSON packet 必须包含 structured `evidence` 字段。
- 生成 review packet 时写 audit：
  - action: `agent_run.review_generated`
  - resource_type: `agent_run`
  - resource_id: Agent Run ID
  - details 包含 `format`、`case_id`、`payload_id`、`interaction_count`、`evidence_strength`

### 2. Agent Run Review API

目标是提供单 run review endpoint。

建议改造：

- `server/v2_api.go`
- `server/v2_api_test.go`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`

接口：

```text
GET /api/v2/agent-runs/:id/review?format=json
GET /api/v2/agent-runs/:id/review?format=markdown
```

API 要求：

- 未认证请求返回 401。
- run 不存在返回 404。
- `format` 缺省为 `json`。
- 非 `json|markdown` 返回 400。
- `format=json` 返回 structured packet。
- `format=markdown` 返回 structured packet，并在 `content` 返回 Markdown。
- 响应不得包含完整 API Key、Authorization header、raw token secret 或敏感回连配置。

### 3. MCP `export_report` 复用 Review Packet

目标是让 Agent 通过 MCP 导出的报告与 Web Review Packet 使用同一契约。

建议改造：

- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `docs/MCP_SERVER_USAGE.md`

要求：

- `export_report` 入参支持 `agent_run_id` 和 `format`。
- 当传入 `agent_run_id` 时，优先调用：

```text
GET /api/v2/agent-runs/:id/review?format=<format>
```

- 当未传 `agent_run_id`，保持当前基于 `case_id` / `payload_id` 的 evidence export 兼容路径。
- `export_report` 的 Agent Operation result 必须记录 review packet 摘要：
  - `format`
  - `agent_run_id`
  - `case_id`
  - `payload_id`
  - `interaction_count`
  - `evidence_strength`
- `export_report` 权限仍使用 Sprint K 的 `agent:export_report` / low risk gate。
- 缺少 scope 或 risk 超限时不能调用 review API。

### 4. Agent Runs Detail 前端 Review 区块

目标是让安全团队在 Agent Run detail 上直接完成复核。

建议改造：

- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`

页面要求：

- 在 Agent Run 详情页新增 “Review Packet” 区块。
- 显示：
  - Evidence Strength
  - Confidence
  - Interaction Count
  - Unique Sources
  - Last Interaction
  - Generated At
- 提供按钮：
  - “生成 JSON Review”
  - “生成 Markdown Review”
  - “查看证据”
- 点击生成按钮必须真实请求 `/api/v2/agent-runs/:id/review?format=...`。
- Markdown review 显示简短预览，不要求实现文件下载。
- “查看证据” 保持跳转到 `/dashboard/evidence?payload_id=...` 或现有 Evidence URL。
- 页面不得展示完整 API Key 或敏感 token。

### 5. 文档与验证记录

建议更新：

- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

要求：

- MCP 文档说明 `export_report` 的 `agent_run_id` 优先路径。
- Agent Native 文档说明 Review Packet 契约。
- `docs/verification.md` 只记录真实执行过的命令和结果。

## 明确禁止越界

Sprint L 不允许实现：

- 完整 Agent 管理平台
- Agent 创建 / Agent 策略编辑 / Agent Marketplace
- Workspace / RBAC 大改
- 真实 LLM 调用或 Agent 自动决策
- 后台任务队列、worker、自动重试、自动超时治理
- 生命周期治理、保留策略、归档策略
- 批量 review / 批量 export / ZIP 导出
- PDF / DOCX / SARIF 导出
- Scanner Hub / 多扫描器扩展 / 新 scanner run 能力
- Webhook 平台
- API Key 加密存储迁移
- 全量 UI 重设计或无关 lint 清理

如实现中发现必须触碰上述内容，必须停下回传，不允许自行扩范围。

## 建议文件清单

后端：

- `internal/agentrun/review.go`
- `internal/agentrun/review_test.go`
- `internal/agentrun/service.go`
- `internal/agentrun/service_test.go`
- `internal/models/agent_run.go`
- `server/v2_api.go`
- `server/v2_api_test.go`

MCP：

- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

前端：

- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`

文档：

- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

## 推荐实施顺序

### Task 1：后端 Review Packet service

1. 先在 `internal/agentrun/review_test.go` 写失败测试：
   - 有 run + payload + interactions 时生成 packet。
   - packet 包含 structured evidence。
   - markdown content 包含 run / case / payload / evidence / operations。
   - 无 interactions 时返回空 evidence packet，不 500。
2. 实现 `BuildReviewPacket`。
3. 运行：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun
```

4. 预期：通过。

### Task 2：Agent Run Review API

1. 在 `server/v2_api_test.go` 增加 API 测试：
   - 未认证 401。
   - 不存在 run 404。
   - invalid format 400。
   - json review 200 且有 `evidence`。
   - markdown review 200 且有 `content`。
2. 在 `server/v2_api.go` 增加路由：

```text
GET /api/v2/agent-runs/:id/review
```

3. 运行：

```bash
GOCACHE=/tmp/gocache go test ./server ./internal/agentrun
```

4. 预期：通过。

### Task 3：MCP `export_report` 接入 Review API

1. 在 `internal/mcp/server_test.go` 增加测试：
   - `export_report` 带 `agent_run_id` 时请求 `/api/v2/agent-runs/:id/review?format=markdown`。
   - `export_report` 缺少 `agent:export_report` scope 时不调用 review API。
   - operation result 包含 `agent_run_id`、`format`、`interaction_count`、`evidence_strength`。
2. 修改 `internal/mcp/server.go`。
3. 运行：

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp
```

4. 预期：通过。

### Task 4：Agent Run detail UI review 区块

1. 在 `frontend-next/e2e/agent-runs.spec.ts` 增加失败测试：
   - 详情页点击 “生成 JSON Review” 后请求 `/api/v2/agent-runs/agent-run-1/review?format=json`。
   - 页面展示 evidence strength、confidence、interaction count。
   - 点击 “生成 Markdown Review” 后请求 markdown review，并展示 markdown preview。
   - “查看证据” 仍跳到 `/dashboard/evidence?payload_id=payload-1`。
2. 修改类型和 API client：
   - `frontend-next/src/types/index.ts`
   - `frontend-next/src/lib/api-client.ts`
3. 修改详情页：
   - `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
4. 运行：

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
```

5. 预期：通过。

### Task 5：文档与全量验证

1. 更新 `docs/MCP_SERVER_USAGE.md`：
   - `export_report` 支持 `agent_run_id`。
   - Review Packet JSON / Markdown 示例。
2. 更新 `docs/agent-native-specification.md`：
   - Review Packet 契约。
   - Agent Run 复核闭环。
3. 更新 `docs/verification.md`：
   - 只记录真实执行命令和结果。
4. 运行最终验证：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

5. 预期：
   - Go 测试通过。
   - 前端目标 ESLint 0 error。
   - 生产构建通过。
   - E2E 无 `test.skip` / `test.only`，且真实点击 review 按钮触发 API 请求。

## 后端验收要求

后端必须满足：

- `GET /api/v2/agent-runs/:id/review?format=json` 返回 structured review packet。
- `GET /api/v2/agent-runs/:id/review?format=markdown` 返回 structured review packet + markdown `content`。
- `format` 非法返回 400。
- run 不存在返回 404。
- 未认证返回 401。
- packet 的 `evidence` 使用统一 Evidence 契约。
- packet 的 interaction summary 来自真实 interactions。
- review 生成写 `agent_run.review_generated` audit。
- audit details 不包含完整 API Key、Authorization header、raw secret。

## MCP 验收要求

MCP 必须满足：

- `export_report` 传 `agent_run_id` 时优先走 Agent Run Review API。
- `export_report` 未传 `agent_run_id` 时保持当前 evidence export 兼容路径。
- `export_report` 权限仍受 Sprint K scope / risk gate 控制。
- 缺 scope 时不调用 review API。
- 成功导出后写 Agent Operation，operation result 包含 review summary。

## 前端验收要求

前端必须满足：

- Agent Run detail 有 Review Packet 区块。
- JSON review 和 Markdown review 按钮真实调用 API。
- 页面展示 evidence strength、confidence、interaction count、unique sources、generated_at。
- Markdown review 有可读预览。
- Evidence 回链继续使用现有 `/dashboard/evidence?...`。
- 页面不展示完整 API Key、Authorization header 或 raw secret。
- E2E 不能只检查静态文字，必须等待 review API request。

## E2E 验收要求

`frontend-next/e2e/agent-runs.spec.ts` 至少新增或更新以下断言：

- `page.waitForRequest` 捕获 `/api/v2/agent-runs/agent-run-1/review?format=json`。
- JSON review 返回后页面展示 `high` / `90%` / interaction count。
- `page.waitForRequest` 捕获 `/api/v2/agent-runs/agent-run-1/review?format=markdown`。
- Markdown review 返回后页面展示 markdown preview。
- Evidence 链接仍跳转到 `/dashboard/evidence?payload_id=payload-1`。
- 没有 `test.skip`、`test.only`、`waitForTimeout`、只检查静态文字的空测。

## 验收完成定义

Sprint L 只有同时满足以下条件才算完成：

1. Agent Run Review Packet 后端 service 有测试覆盖。
2. `/api/v2/agent-runs/:id/review` 支持 JSON / Markdown。
3. Review Packet 复用统一 Evidence 契约。
4. Interaction summary 来自真实 Interaction 聚合。
5. Review 生成写 audit。
6. MCP `export_report(agent_run_id=...)` 走 Review API。
7. MCP 权限拒绝时不调用 Review API。
8. Agent Run detail 页面能生成并展示 review packet。
9. Evidence 回链仍闭合。
10. E2E 真实触发 review API 请求并断言结果。
11. 无 `test.skip` / `test.only` / `waitForTimeout` / 静态空测。
12. 未越界到 Agent 管理、Scanner Hub、生命周期治理、批量操作或新导出平台。
13. `docs/verification.md` 记录 Sprint L 实际验证命令和结果。

## Windsurf 回传要求

Windsurf 完成 Sprint L 后，请回传：

- 修改文件清单。
- Review Packet 响应示例。
- MCP `export_report` 带 `agent_run_id` 的请求路径说明。
- 新增 / 更新测试清单。
- 实际运行的验证命令和结果。
- 如有未完成项，必须明确标为遗留，不得写“已完成”。

## Codex 验收重点

Codex 复验 Sprint L 时重点检查：

1. Review Packet 是否真实来自 Agent Run / Operation / Interaction / Evidence / Audit，而不是前端 mock。
2. Evidence 是否继续使用统一 Evidence 契约。
3. Markdown 和 JSON 是否来自同一个 service 契约。
4. MCP `export_report` 是否走 Review API，且缺 scope 时不调用。
5. Agent Run detail -> Review -> Evidence 是否形成闭环。
6. E2E 是否真实等待 API request，不是只看静态文字。
7. 是否严格没有越界到 Scanner Hub、生命周期治理、批量操作或完整 Agent 管理。

只有上述全部成立，Sprint L 才能关闭。
