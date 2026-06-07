# GODNSLOG 2.0 Sprint N Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the single Agent Run review/follow-up loop from Sprint L/M into a small operator-facing review queue, so security operators can find runs that need attention, inspect their review packet, see follow-up history, and confirm audit traceability.

**Architecture:** Reuse the existing Agent Run, Agent Operation, Review Packet, Follow-up Action, Evidence, and Audit contracts. Add derived review metadata and queue APIs around existing records. Do not introduce a new workflow engine, replay backend, scanner scheduler, or bulk operation system.

**Tech Stack:** Go, xorm, Gin server routes in `server/v2_api.go`, existing `internal/agentrun`, `internal/auth` audit service, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint N
- **Sprint 主题**：Agent Review Queue & Follow-up Traceability
- **所属阶段**：Phase 5 - Agent Governance and Replay

## Sprint 背景

Sprint L 已完成单个 Agent Run 的 Review Packet：

`Agent Run -> Review Packet -> Evidence Summary / Markdown -> Export Audit`

Sprint M 已完成最小 follow-up action：

`Review Packet -> Follow-up Action -> Agent Operation -> Audit`

现在单个 Agent Run 已经可以复核和记录后续动作，但 operator 还缺一个“从列表里找出需要复核的 Agent Run”的入口。当前 Agent Runs 列表主要是运行记录列表，无法直接回答：

- 哪些 Agent Run 已生成 review？
- 哪些 Agent Run 已创建 follow-up？
- 哪些 Agent Run 仍没有任何 follow-up？
- 哪些 Agent Run 有高置信 Evidence，需要 operator 二次复核？
- 单个 Run 的 follow-up 是否能从列表、详情、Audit 三处互相印证？

Sprint N 的目标是补齐这个 operator review queue，而不是实现完整 Agent 管理平台或 Agent 生命周期治理。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 为 Agent Runs 增加 review queue 视图/API，支持按 review / follow-up / evidence strength / status 做最小筛选。
2. 在 Agent Run 列表和详情中展示 review/follow-up 派生状态。
3. 在 Agent Run detail 增加 follow-up history 区块，复用 Agent Operation timeline 中的 `followup.*` operation。
4. 在 Audit 页面或 Agent Run detail 中提供 `agent_run.followup_created` / `agent_run.review_generated` 的最小回链证明。
5. 用后端测试和 E2E 证明：Review Queue -> Agent Run Detail -> Review Packet -> Follow-up History -> Audit 形成闭环。

## 明确不做

本 Sprint 严格不做：

- 完整 Agent 管理平台。
- Agent 创建、删除、启停或策略管理。
- Agent replay 引擎、后台任务队列或自动重放。
- 批量 review、批量 follow-up、批量状态修改。
- Scanner Hub 扩展或扫描器调度。
- 生命周期治理、retention、归档、审批流。
- 真实 LLM 调用。
- 删除、撤销、修改配置、revoke token 等高风险动作。
- PDF/DOCX/ZIP 报告中心。

## 输入文档

Windsurf 实施前必须阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/unified-control-plane.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-j-package.md`
- `docs/superpowers/acceptance/2026-05-24-godnslog-2-sprint-j-acceptance.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-k-package.md`
- `docs/superpowers/acceptance/2026-05-25-godnslog-2-sprint-k-acceptance.md`
- `docs/superpowers/plans/2026-05-31-godnslog-2-sprint-l-package.md`
- `docs/superpowers/acceptance/2026-05-31-godnslog-2-sprint-l-acceptance.md`
- `docs/superpowers/plans/2026-06-06-godnslog-2-sprint-m-package.md`
- `docs/superpowers/acceptance/2026-06-06-godnslog-2-sprint-m-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/agentrun/service.go` 已有 Agent Run create / list / get / status update / append operation。
- `internal/agentrun/review.go` 已有单 Agent Run Review Packet。
- `server/v2_api.go` 已有 Agent Run list/detail、review、follow-up API。
- `internal/models/agent_run.go` 已有 Agent Run、Agent Operation、Follow-up request/response。
- `frontend-next/src/app/dashboard/agent-runs/page.tsx` 已有 Agent Runs 列表。
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已有 detail、Review Packet、Follow-up Action、Operation timeline。
- `frontend-next/e2e/agent-runs.spec.ts` 已覆盖 list/detail/review/follow-up 主链路。

### 主要缺口

- Agent Runs 列表不能按 review/follow-up 状态筛选。
- Agent Runs 列表没有体现一个 run 是否已经被 review 或 follow-up。
- Follow-up history 只混在 Operation timeline 中，没有 operator 可快速扫描的 review action 摘要。
- Audit 与 Agent Run detail 缺少明确回链，operator 不能从 follow-up 快速确认对应 audit entry。
- E2E 还没有证明从 review queue 到 detail、follow-up history、audit 的完整 traceability。

## 术语边界

### Review Queue

Review Queue 是 Agent Run 列表的一个复核视图。它展示和筛选单个 Agent Run 的复核状态，不是新的任务队列，也不执行后台作业。

### Review State

Review State 是从现有 Agent Operation / Audit / Review Packet 派生出来的 UI/API 字段，不要求新增独立持久化表。

最小状态：

- `not_reviewed`：没有 `agent_run.review_generated` audit，也没有 review export operation。
- `reviewed`：已生成 Review Packet 或 review export。
- `followup_created`：存在 `followup.*` Agent Operation 或 `agent_run.followup_created` audit。
- `needs_attention`：高置信 Evidence 且没有 follow-up，或最近一次 follow-up action 是 `recheck_evidence`。

### Follow-up History

Follow-up History 是 `AgentOperation.Action` 以 `followup.` 开头的 operations 的 operator 摘要视图。它不是新的模型，也不允许直接编辑历史。

最小字段：

- `operation_id`
- `action_type`
- `reason`
- `review_packet_id`
- `created_at`
- `audit_ref`

## 数据契约

### AgentRunReviewQueueItem

建议新增到 `internal/models/agent_run.go`，或放在 `internal/agentrun` 中作为 API response DTO：

```go
type AgentRunReviewQueueItem struct {
	ID                 string     `json:"id"`
	AgentID            string     `json:"agent_id,omitempty"`
	OperatorID         string     `json:"operator_id,omitempty"`
	CaseID             string     `json:"case_id,omitempty"`
	PayloadID          string     `json:"payload_id,omitempty"`
	Target             string     `json:"target,omitempty"`
	Status             string     `json:"status"`
	ReviewState        string     `json:"review_state"`
	EvidenceStrength   string     `json:"evidence_strength,omitempty"`
	InteractionCount   int        `json:"interaction_count"`
	OperationCount     int        `json:"operation_count"`
	FollowupCount      int        `json:"followup_count"`
	LastFollowupAction  string     `json:"last_followup_action,omitempty"`
	LastReviewedAt      *time.Time `json:"last_reviewed_at,omitempty"`
	LastFollowupAt      *time.Time `json:"last_followup_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DetailURL          string     `json:"detail_url"`
	EvidenceURL         string     `json:"evidence_url,omitempty"`
}
```

### AgentRunFollowupHistoryItem

```go
type AgentRunFollowupHistoryItem struct {
	OperationID    string     `json:"operation_id"`
	ActionType     string     `json:"action_type"`
	Reason         string     `json:"reason,omitempty"`
	ReviewPacketID string     `json:"review_packet_id,omitempty"`
	RiskLevel      string     `json:"risk_level"`
	AuditRefID     string     `json:"audit_ref_id,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}
```

### AgentRunReviewQueueResponse

```go
type AgentRunReviewQueueResponse struct {
	Items      []AgentRunReviewQueueItem `json:"items"`
	Total      int64                     `json:"total"`
	Page       int                       `json:"page"`
	PageSize   int                       `json:"page_size"`
	TotalPages int                       `json:"total_pages"`
	Summary    AgentRunReviewQueueSummary `json:"summary"`
}
```

### AgentRunReviewQueueSummary

```go
type AgentRunReviewQueueSummary struct {
	Total             int64 `json:"total"`
	NotReviewed       int64 `json:"not_reviewed"`
	Reviewed          int64 `json:"reviewed"`
	FollowupCreated   int64 `json:"followup_created"`
	NeedsAttention    int64 `json:"needs_attention"`
}
```

## API 范围

新增：

```text
GET /api/v2/agent-runs/review-queue
```

Query params：

- `review_state=not_reviewed|reviewed|followup_created|needs_attention`
- `status=running|completed|failed|cancelled|timed_out`
- `evidence_strength=high|medium|low|none`
- `agent_id=...`
- `case_id=...`
- `payload_id=...`
- `page=1`
- `page_size=20`

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [],
    "total": 0,
    "page": 1,
    "page_size": 20,
    "total_pages": 0,
    "summary": {
      "total": 0,
      "not_reviewed": 0,
      "reviewed": 0,
      "followup_created": 0,
      "needs_attention": 0
    }
  }
}
```

新增：

```text
GET /api/v2/agent-runs/:id/followups
```

响应：

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "operation_id": "op-followup-1",
        "action_type": "recheck_evidence",
        "reason": "Evidence needs second review",
        "review_packet_id": "agent-run-1",
        "risk_level": "low",
        "audit_ref_id": "audit-1",
        "created_at": "2026-06-07T10:00:00Z"
      }
    ]
  }
}
```

允许改造：

- `GET /api/v2/agent-runs/:id` 可以在 detail response 中增加 `followup_history` 或 `review_state`，但不要破坏现有字段。
- 如果实现单独 `GET /followups` 更简单清晰，detail 页面可额外调用该 API。

错误语义：

- 未认证请求：401。
- unknown Agent Run：404。
- invalid `review_state` / `evidence_strength`：400。
- `page_size` 超过 100 时 clamp 到 100 或返回 400，必须有测试锁定。

安全要求：

- 响应不得包含完整 API Key、Authorization header、token secret 或敏感回连配置。
- Follow-up reason 可以返回，但不得包含 raw request header / token secret。

## 后端实施范围

### 1. Review Queue Service

建议修改：

- `internal/agentrun/service.go`
- `internal/agentrun/review_queue.go`
- `internal/agentrun/review_queue_test.go`
- `internal/models/agent_run.go`

要求：

- [ ] 实现 `ListReviewQueue(filters)`。
- [ ] 从 Agent Run + Operations + Audit 派生 `review_state`。
- [ ] `followup_count` 必须从 `followup.*` operations 聚合。
- [ ] `last_followup_action` / `last_followup_at` 必须来自最新 follow-up operation。
- [ ] `reviewed` 状态必须能通过 `agent_run.review_generated` audit 或 review export operation 判断。
- [ ] `needs_attention` 至少覆盖：
  - high evidence 且 `followup_count = 0`
  - 最新 follow-up action 是 `followup.recheck_evidence`
- [ ] 如果当前 Evidence strength 只能通过 Review Packet 动态生成，允许先以 Review Packet 结果计算，不新增 Evidence 持久化表。
- [ ] 不存在 Interaction/Evidence 时仍返回 queue item，`evidence_strength` 可为 `none`。

### 2. Follow-up History Service

建议修改：

- `internal/agentrun/service.go`
- `internal/agentrun/followup_history.go`
- `internal/agentrun/service_test.go`

要求：

- [ ] 实现 `ListFollowupHistory(agentRunID)`。
- [ ] 只返回 `Action` 前缀为 `followup.` 的 operations。
- [ ] 从 operation result JSON 解析 `reason`、`review_packet_id`、`action_type`。
- [ ] 关联 `agent_run.followup_created` audit，返回 `audit_ref_id`。
- [ ] JSON 解析失败时不 500，返回该 operation 的最小信息并在测试中覆盖。

### 3. API Handler

建议修改：

- `server/v2_api.go`
- `server/v2_api_test.go`

要求：

- [ ] 注册 `GET /api/v2/agent-runs/review-queue`。
- [ ] 注册 `GET /api/v2/agent-runs/:id/followups`。
- [ ] 覆盖成功路径。
- [ ] 覆盖 unknown Agent Run -> 404。
- [ ] 覆盖 invalid filter -> 400。
- [ ] 覆盖鉴权。
- [ ] 测试响应不包含 `Authorization`、完整 API key、token secret。

## 前端实施范围

### 1. Types and API Client

建议修改：

- `frontend-next/src/types/index.ts`
- `frontend-next/src/lib/api-client.ts`

要求：

- [ ] 增加 `AgentRunReviewQueueItem`。
- [ ] 增加 `AgentRunReviewQueueSummary`。
- [ ] 增加 `AgentRunFollowupHistoryItem`。
- [ ] 增加 `agentRunApi.getReviewQueue(params)`。
- [ ] 增加 `agentRunApi.getFollowups(agentRunID)`。

### 2. Agent Runs List Review Queue

建议修改：

- `frontend-next/src/app/dashboard/agent-runs/page.tsx`

要求：

- [ ] 页面增加 Review Queue tab 或 segmented control，不新增大范围导航。
- [ ] Review Queue 视图调用 `/api/v2/agent-runs/review-queue`，不能只复用静态 mock。
- [ ] 展示 summary：Total / Not Reviewed / Reviewed / Follow-up / Needs Attention。
- [ ] 支持筛选：
  - review state
  - status
  - evidence strength
- [ ] 筛选变化必须进入 API query。
- [ ] 每一行展示：
  - Agent Run ID / title
  - status
  - review state
  - evidence strength
  - interaction count
  - followup count
  - last follow-up action
  - detail link
- [ ] 空状态、loading、error 状态完整。

### 3. Agent Run Detail Follow-up History

建议修改：

- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`

要求：

- [ ] 在 Review Packet / Follow-up Action 附近增加 Follow-up History 区块。
- [ ] Follow-up History 调用真实 API 或来自 detail response 的真实字段。
- [ ] 展示 action type、reason、review packet id、created_at、audit ref。
- [ ] audit ref 可以链接到 `/dashboard/audit?resource_type=agent_run&resource_id=<id>` 或等价 filter URL。
- [ ] 创建 follow-up 成功后，history 和 operation timeline 都刷新。
- [ ] 不新增批量 action。

### 4. Audit 回链

建议修改：

- `frontend-next/src/app/dashboard/audit/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/e2e/agent-runs.spec.ts` 或新增 `agent-review-queue.spec.ts`

要求：

- [ ] Audit 页面能读取 query params 并进入对应过滤条件，至少支持 `resource_type` 和 `resource_id`。
- [ ] 从 Follow-up History 的 audit link 进入 Audit 页面后，能看到 `agent_run.followup_created`。
- [ ] 不能只断言静态文字；E2E 必须等待 audit API request，并断言 query params。

## E2E 验收要求

新增或修改：

- `frontend-next/e2e/agent-runs.spec.ts`
- 或新增 `frontend-next/e2e/agent-review-queue.spec.ts`

必须覆盖：

1. Review Queue tab/视图进入后，真实请求 `/api/v2/agent-runs/review-queue`。
2. 切换 `review_state=needs_attention` 时，API query 真实变化。
3. Summary 随 mock response 变化，不是静态文字。
4. 点击 queue item 进入 Agent Run detail。
5. Detail 中 Follow-up History 真实渲染 `followup.recheck_evidence`、reason、audit ref。
6. 创建新的 follow-up 后，Follow-up History 和 Operation timeline 都刷新。
7. 点击 audit ref 后进入 Audit 页面，并真实请求带 `resource_type=agent_run` / `resource_id=<id>` 的 audit API。
8. E2E 中不得出现 `test.skip`、`test.only`。
9. 不允许只检查标题、按钮、静态说明文案。

## 完成定义

Sprint N 只有同时满足以下条件才可关闭：

- [ ] 后端 Review Queue API 存在，并从真实 Agent Run / Operation / Audit 派生状态。
- [ ] Follow-up History API 存在，并只返回 `followup.*` operations。
- [ ] Agent Runs 页面能以 Review Queue 方式筛选和展示 review/follow-up 状态。
- [ ] Agent Run detail 能展示 Follow-up History，并和 Operation timeline / Audit 互相回链。
- [ ] Audit 页面能从 query params 进入资源过滤。
- [ ] E2E 证明 Review Queue -> Detail -> Follow-up History -> Audit 闭环。
- [ ] E2E 没有 skip / only / 空测。
- [ ] 未越界到 Scanner Hub、生命周期治理、批量操作、真实 replay engine、完整 Agent 管理平台。
- [ ] `docs/verification.md` 记录实际运行过的命令和结果。

## 验证命令

Windsurf 完成后必须至少运行并回传：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

如果新增 `agent-review-queue.spec.ts`，Playwright 命令改为：

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-review-queue.spec.ts e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

执行前端 E2E 时禁止使用会触发 HTML report 常驻服务的流程：

- 不执行 `npx playwright show-report`
- 不使用打开式 report
- 只使用一次性 `--reporter=line` 或 `--reporter=list`

## 验收重点

Codex 验收 Sprint N 时重点检查：

1. Review Queue API 是否真实从 Agent Run / Operation / Audit 派生状态。
2. `review_state` / `evidence_strength` / `status` filter 是否真实进入 API 请求。
3. Summary 是否随 scope/filter/mock response 变化。
4. Follow-up History 是否只来自 `followup.*` operations。
5. Follow-up History -> Audit ref 是否形成真实回链。
6. 创建 follow-up 后，History 和 Operation timeline 是否都刷新。
7. Evidence 是否继续使用统一 Evidence 契约，没有新增不兼容 evidence shape。
8. E2E 是否没有 skip / only / 只检查静态文字的空测。
9. 是否严格没有越界到 Scanner Hub、生命周期治理、批量操作、真实 replay engine 或完整 Agent 管理平台。

## 推荐实施顺序

1. 后端测试先行：写 Review Queue / Follow-up History service 测试。
2. 实现 service 和 DTO。
3. 增加 API handler 和 API 测试。
4. 增加前端 types/api-client。
5. 改 Agent Runs list，增加 Review Queue 视图。
6. 改 Agent Run detail，增加 Follow-up History。
7. 改 Audit 页面 query param filter。
8. 补 E2E 闭环。
9. 跑验证命令并更新 `docs/verification.md`。
