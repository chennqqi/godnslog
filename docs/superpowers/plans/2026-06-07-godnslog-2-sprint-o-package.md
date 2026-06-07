# GODNSLOG 2.0 Sprint O Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn Sprint N's Review Queue from a traceability view into a minimal operator closure loop: an operator can record a final review decision for one Agent Run, the queue state updates from that decision, and Audit / Operation / Detail all prove the decision.

**Architecture:** Reuse existing Agent Run, Agent Operation, Review Packet, Follow-up History, Review Queue, and Audit contracts. Add a single-run review decision action and derived queue closure state. Do not introduce workflow engines, approval systems, bulk operations, retention/lifecycle governance, scanner scheduling, or new Agent management.

**Tech Stack:** Go, xorm, Gin routes in `server/v2_api.go`, `internal/agentrun`, `internal/auth` audit service, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint O
- **Sprint 主题**：Agent Review Decision & Queue Closure
- **所属阶段**：Phase 5 - Agent Governance and Review Operations

## Sprint 背景

Sprint L 完成单个 Agent Run 的 Review Packet。

Sprint M 完成最小 Follow-up Action。

Sprint N 完成 Review Queue 和 Follow-up History：

```text
Review Queue -> Agent Run Detail -> Follow-up History -> Audit
```

现在 operator 能找到需要复核的 Agent Run，也能确认 follow-up 和 audit 回链。但队列仍缺一个最小“复核结论”动作。安全团队完成复核后，无法明确记录：

- 这个 Agent Run 是否已接受复核？
- 是否判定为 false positive？
- 是否确认需要后续人工处理？
- 队列 item 为什么从 needs_attention 退出？
- 复核结论是否进入 Operation timeline 和 Audit？

Sprint O 的目标是补齐单个 Agent Run 的 review decision / queue closure，不做完整审批流。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 增加单个 Agent Run 的 Review Decision API。
2. Review Decision 写入 Agent Operation 和 Audit。
3. Review Queue 从 Review Decision 派生 `review_state` / closure 状态。
4. Agent Run detail 提供最小 review decision UI，并在 Operation timeline / Audit 中可回查。
5. E2E 证明：Review Queue needs_attention -> Detail -> Submit Decision -> Queue stats 更新 -> Audit entry 显示。

## 明确不做

本 Sprint 严格不做：

- 批量 review、批量 decision、批量 follow-up。
- 多人审批流、SLA、assign、comment thread、notification。
- Agent 创建、删除、启停或策略管理。
- Agent replay engine、后台任务队列、自动重放。
- Scanner Hub 扩展、扫描器调度、真实扫描任务。
- 生命周期治理、retention、归档、删除策略。
- PDF/DOCX/ZIP 报告中心。
- 真实 LLM 调用。
- 高风险动作，例如删除、撤销、revoke token、修改生产配置。

## 输入文档

Windsurf 实施前必须阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-05-31-godnslog-2-sprint-l-package.md`
- `docs/superpowers/acceptance/2026-05-31-godnslog-2-sprint-l-acceptance.md`
- `docs/superpowers/plans/2026-06-06-godnslog-2-sprint-m-package.md`
- `docs/superpowers/acceptance/2026-06-06-godnslog-2-sprint-m-acceptance.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-n-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-n-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/agentrun/service.go` 已有 Agent Run detail、status update、operation append、follow-up action。
- `internal/agentrun/review.go` 已有 Review Packet。
- `internal/agentrun/review_queue.go` 已有 Review Queue 派生逻辑。
- `server/v2_api.go` 已有 Agent Run list/detail/review/follow-up/review-queue API。
- `internal/models/agent_run.go` 已有 AgentOperation 和 Review Queue DTO。
- `frontend-next/src/app/dashboard/agent-runs/page.tsx` 已有 Review Queue tab、summary、filter。
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已有 Review Packet、Follow-up Action、Follow-up History、Operation timeline。
- `frontend-next/e2e/agent-runs.spec.ts` 已覆盖 Review Queue scope、stats、Follow-up History -> Audit 回链。

### 主要缺口

- Review Queue item 可以进入 `needs_attention`，但 operator 无法记录最终 review decision。
- 当前 `reviewed` 状态主要来自 review generation，而不是 operator decision。
- Follow-up Action 是“后续动作记录”，不是“复核结论”。
- 没有统一 audit action 表示 operator 已接受、驳回、确认误报或确认需要人工处理。
- E2E 不能证明 queue item 从 `needs_attention` 通过 operator decision 被关闭。

## 术语边界

### Review Decision

Review Decision 是 operator 对单个 Agent Run 的复核结论。它是审计和操作记录，不是审批流，也不是状态机引擎。

最小 decision 类型：

- `accepted`: 复核通过，证据足够，可关闭 needs_attention。
- `false_positive`: 复核认为无需继续处理。
- `needs_manual_followup`: 需要人工后续处理，但不自动创建任务。
- `insufficient_evidence`: 证据不足，建议后续补充。

### Queue Closure

Queue Closure 是 Review Queue 派生状态。它由最近一次 `review_decision.*` operation 或 `agent_run.review_decision_recorded` audit 决定，不新增独立队列表。

### Out of Scope: Assignment / Workflow

Review Decision 不包含 assign、owner、SLA、审批人链路、通知、批量操作或生命周期策略。

## 数据契约

### AgentRunReviewDecisionRequest

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunReviewDecisionRequest struct {
	Decision       string `json:"decision" binding:"required"`
	Reason         string `json:"reason,omitempty"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
	EvidenceID      string `json:"evidence_id,omitempty"`
}
```

`decision` 只允许：

```text
accepted
false_positive
needs_manual_followup
insufficient_evidence
```

### AgentRunReviewDecisionResponse

```go
type AgentRunReviewDecisionResponse struct {
	AgentRunID     string                 `json:"agent_run_id"`
	OperationID    string                 `json:"operation_id"`
	Decision       string                 `json:"decision"`
	ReviewPacketID string                 `json:"review_packet_id,omitempty"`
	AuditRefID     string                 `json:"audit_ref_id,omitempty"`
	Operation      *AgentOperation        `json:"operation,omitempty"`
	Audit          map[string]interface{} `json:"audit,omitempty"`
}
```

### AgentOperation Result Contract

Review Decision 写入 `AgentOperation`：

```text
action = review_decision.<decision>
risk_level = low
```

`result` JSON 至少包含：

```json
{
  "decision": "accepted",
  "reason": "Evidence reviewed by operator",
  "review_packet_id": "agent-run-1",
  "evidence_id": "optional",
  "audit_action": "agent_run.review_decision_recorded"
}
```

### Audit Contract

Review Decision 写入 audit：

```text
action = agent_run.review_decision_recorded
resource_type = agent_run
resource_id = <agent_run_id>
result = success
```

`parameters` 至少包含：

```json
{
  "decision": "accepted",
  "reason": "Evidence reviewed by operator",
  "review_packet_id": "agent-run-1",
  "operation_id": "op-..."
}
```

不得写入 token、raw request header、secret、完整 payload secret。

## API 范围

新增：

```text
POST /api/v2/agent-runs/:id/review-decision
```

Request:

```json
{
  "decision": "accepted",
  "reason": "Evidence reviewed by operator",
  "review_packet_id": "agent-run-1"
}
```

Response:

```json
{
  "code": 0,
  "data": {
    "agent_run_id": "agent-run-1",
    "operation_id": "op-...",
    "decision": "accepted",
    "review_packet_id": "agent-run-1",
    "audit_ref_id": "audit-..."
  }
}
```

Validation:

- `decision` 必须为允许枚举。
- `reason` 最大长度建议 500。
- `review_packet_id` 可选，但如果提供必须和当前 Agent Run 相关。
- API 只能操作单个 Agent Run。

## Review Queue 派生规则

Sprint O 只调整 Review Queue 的最小派生逻辑：

1. 如果存在最近的 `review_decision.accepted` 或 `review_decision.false_positive` operation，则 `review_state=reviewed`，`needs_attention=false`。
2. 如果存在最近的 `review_decision.needs_manual_followup` 或 `review_decision.insufficient_evidence`，则 `review_state=needs_attention`，但 item 必须展示 `last_review_decision`。
3. 如果没有 review decision，沿用 Sprint N 的 Review Queue 派生规则。
4. Follow-up history 不混入 review decision；Review Decision 可在 Operation timeline 和 Review Decision summary 中展示。

建议为 Review Queue item 增加可选字段：

```go
LastReviewDecision string     `json:"last_review_decision,omitempty"`
LastDecisionReason string     `json:"last_decision_reason,omitempty"`
LastDecisionAt     *time.Time `json:"last_decision_at,omitempty"`
```

## 后端任务

### Task 1: Service Tests First

- [ ] 在 `internal/agentrun` 增加 Review Decision service 测试。
- [ ] 覆盖 `accepted` decision 写入 operation。
- [ ] 覆盖 `false_positive` decision 写入 operation。
- [ ] 覆盖非法 decision 返回错误。
- [ ] 覆盖 reason 长度限制。
- [ ] 覆盖 audit 写入 `agent_run.review_decision_recorded`。
- [ ] 覆盖 Review Queue 在 `accepted` decision 后不再返回 `needs_attention`。
- [ ] 覆盖 Review Queue 在 `needs_manual_followup` decision 后仍为 `needs_attention` 且包含 `last_review_decision`。

### Task 2: Service Implementation

- [ ] 在 `internal/agentrun/service.go` 增加 `RecordReviewDecision`。
- [ ] 复用现有 Agent Run / Operation 写入模式。
- [ ] 写入 `review_decision.<decision>` Agent Operation。
- [ ] 写入 `agent_run.review_decision_recorded` Audit。
- [ ] 返回 operation id 和 audit ref。
- [ ] 不新增后台任务，不触发真实 replay，不创建 scanner run。

### Task 3: Review Queue Derivation

- [ ] 更新 `internal/agentrun/review_queue.go`，识别 `review_decision.*` operation。
- [ ] Queue item 增加 `last_review_decision` / `last_decision_reason` / `last_decision_at`。
- [ ] `accepted` / `false_positive` 能关闭 `needs_attention`。
- [ ] `needs_manual_followup` / `insufficient_evidence` 保持或进入 `needs_attention`。
- [ ] Summary 统计必须基于更新后的派生结果。

### Task 4: API Route

- [ ] 在 `server/v2_api.go` 注册 `POST /api/v2/agent-runs/:id/review-decision`。
- [ ] 增加 request validation。
- [ ] 返回统一 v2 response。
- [ ] 增加 `server/v2_api_test.go` API 测试：
  - [ ] success
  - [ ] invalid decision
  - [ ] unauthenticated
  - [ ] operation written
  - [ ] audit written
  - [ ] review queue state updates after decision

## 前端任务

### Task 5: API Client and Types

- [ ] 在 `frontend-next/src/types/index.ts` 增加 Review Decision request/response 类型。
- [ ] 在 `frontend-next/src/lib/api-client.ts` 增加 `createAgentRunReviewDecision`。
- [ ] 类型字段必须与后端 JSON 契约一致。

### Task 6: Agent Run Detail UI

在 `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`：

- [ ] 在 Review Packet / Follow-up Action 附近增加 Review Decision 控件。
- [ ] 使用 select/segmented control 选择 decision。
- [ ] 使用 textarea 输入 reason。
- [ ] Submit 后调用真实 API。
- [ ] 成功后刷新 Agent Run detail / Operation timeline。
- [ ] Operation timeline 显示 `review_decision.accepted` 等 action。
- [ ] 显示 audit ref link，进入 Audit 页面带 `resource_type=agent_run&resource_id=<id>`。
- [ ] 不新增批量 action。

### Task 7: Review Queue UI

在 `frontend-next/src/app/dashboard/agent-runs/page.tsx`：

- [ ] Review Queue item 展示 `last_review_decision`。
- [ ] Review Queue item 展示 `last_decision_at`。
- [ ] `accepted` / `false_positive` 后 item 不再出现在 `needs_attention` scope。
- [ ] `needs_manual_followup` / `insufficient_evidence` 后 item 仍可在 `needs_attention` scope 中定位。
- [ ] 不新增单独一级导航。

## E2E 要求

更新 `frontend-next/e2e/agent-runs.spec.ts`，不得 skip。

必须证明：

1. 从 Review Queue `needs_attention` filter 进入 detail。
2. 在 detail 提交 `accepted` Review Decision。
3. E2E 等待 `POST /api/v2/agent-runs/:id/review-decision`。
4. 断言 request body 包含 `decision=accepted` 和 reason。
5. 提交后 Operation timeline 出现 `review_decision.accepted`。
6. 提交后 Audit link 进入 Audit 页面，并等待 `/api/v2/audit/logs`。
7. Audit 页面显示 `agent_run.review_decision_recorded`。
8. 返回 Review Queue，切换 `needs_attention`，断言该 run 不再出现，summary `needs_attention` 数值减少。
9. 再覆盖一个 `needs_manual_followup` 分支，断言该 run 仍在 `needs_attention` scope 且显示 `last_review_decision`。

E2E 禁止：

- 只检查静态标题。
- 只检查 href 不点击。
- 只 mock response 不断言 request。
- `test.skip`。
- `npx playwright show-report`。

## 安全与审计要求

- Review Decision 是低风险审计动作。
- reason 不得包含 token、API key、cookie、raw header。
- Audit parameters 只能保存结构化 decision metadata。
- 不允许通过 Review Decision 触发 outbound request、scanner run、agent replay。
- 不允许修改 API key、user、scanner config、retention policy。

## 验证命令

Windsurf 完成后必须运行并记录：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

如果 Playwright 需要 dev server，必须使用一次性非交互式测试命令，不得打开 HTML report。

## 完成定义

Sprint O 只有同时满足以下条件才可关闭：

- [ ] `POST /api/v2/agent-runs/:id/review-decision` 存在并写入真实 operation。
- [ ] Review Decision 写入 `agent_run.review_decision_recorded` audit。
- [ ] Review Queue 能从 `review_decision.*` operation 派生 closure 状态。
- [ ] `accepted` / `false_positive` 能从 `needs_attention` scope 中移除对应 run。
- [ ] `needs_manual_followup` / `insufficient_evidence` 能继续在 `needs_attention` scope 中展示。
- [ ] Agent Run detail 能提交 review decision，并刷新 Operation timeline。
- [ ] Audit 页面能显示 `agent_run.review_decision_recorded`。
- [ ] E2E 证明 Review Queue -> Detail -> Decision -> Operation -> Audit -> Queue stats 的闭环。
- [ ] E2E 没有 skip，没有只检查静态文字的空测。
- [ ] 未越界到 Scanner Hub、生命周期治理、批量操作、真实 replay engine、完整 Agent 管理平台。
- [ ] `docs/verification.md` 记录 Sprint O 实际验证命令和结果。

## Windsurf 完成回传格式

Windsurf 完成 Sprint O 后，请回传：

```text
Sprint O completed.

Backend:
- Review decision service:
- Review queue derivation:
- API route:
- Audit action:
- Tests:

Frontend:
- Detail decision UI:
- Review Queue decision display:
- API client/types:
- E2E:

Verification:
- go test ./internal/agentrun ./server:
- go test ./...:
- targeted eslint:
- npm run build:
- playwright agent-runs:

Out of scope confirmation:
- No Scanner Hub expansion:
- No lifecycle / retention governance:
- No bulk operations:
- No replay engine:
```

## Codex 验收重点

Codex 验收 Sprint O 时必须逐项检查：

1. Review Decision 是否写入 `AgentOperation`。
2. Audit 是否使用 `agent_run.review_decision_recorded`。
3. Review Queue 是否真实从 operation/audit 派生 decision state。
4. `accepted` 是否真实影响 `needs_attention` API query 和 summary。
5. Detail 提交是否真实进入 API request body。
6. Operation timeline 是否刷新显示 `review_decision.*`。
7. Audit 页面是否显示对应 audit entry。
8. E2E 是否没有 skip 或静态空测。
9. 是否严格没有越界到 Scanner Hub、生命周期治理、批量操作或 replay engine。
