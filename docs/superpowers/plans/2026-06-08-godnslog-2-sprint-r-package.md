# GODNSLOG 2.0 Sprint R Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn Sprint Q's one-shot webhook delivery into a reviewable delivery history: an operator can see every Review Evidence Delivery attempt for a single Agent Run, inspect the sanitized receipt, jump to the related Operation/Audit, and distinguish delivered / failed / timeout outcomes without opening a report center or notification system.

**Architecture:** Reuse Sprint Q `review_delivery.webhook` Agent Operations and `agent_run.review_delivered` / `agent_run.review_delivery_failed` Audit records as the source of truth. Add a read-only delivery history API and Agent Run Detail UI section. Do not introduce saved connectors, retry queues, notification center, batch delivery, report center, lifecycle governance, Scanner Hub expansion, workflow engine, or MCP auto-delivery.

**Tech Stack:** Go, xorm, Gin routes in `server/v2_api.go`, `internal/agentrun`, `internal/auth`, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint R
- **Sprint 主题**：Review Delivery History & Receipt Review
- **所属阶段**：Phase 5 - Agent Governance and Review Operations

## Sprint 背景

Sprint P 完成单个 Agent Run Review Evidence Export Package：

```text
Agent Run Detail -> Export JSON / Markdown -> Operation -> Audit
```

Sprint Q 完成单个 Review Evidence Package 的 webhook delivery：

```text
Agent Run Detail -> Deliver to Webhook -> Delivery Receipt -> Operation -> Audit
```

现在 operator 可以完成一次外部交付，并在弹窗中看到 receipt。但 receipt 是即时结果，后续回到 Agent Run Detail 时，operator 只能从通用 Operation timeline 或 Audit 页面间接查找。安全团队在复盘时需要更直接地回答：

- 这个 Agent Run 一共交付过几次？
- 哪些交付成功、失败或 timeout？
- 每次交付给哪个 destination host？
- 对应的 export operation、delivery operation、audit ref 是什么？
- 是否只展示 sanitized receipt，不泄露 full webhook URL 或 header values？

Sprint R 的目标是把 Q 的 receipt 从一次性弹窗提升为单个 Agent Run 内的可回查历史，不做连接器或通知中心。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 增加单个 Agent Run 的 Review Delivery History read API。
2. 从 `review_delivery.webhook` operations 和 delivery audit 派生 sanitized delivery receipt 列表。
3. Agent Run Detail 增加 Delivery History section，显示 delivered / failed / timeout、destination host、status code、refs。
4. Delivery History 中提供 Operation / Audit 回链，不展示 full webhook URL 或 header values。
5. E2E 证明：Delivery -> History refresh -> Receipt detail -> Operation timeline -> Audit 闭环。

## 明确不做

本 Sprint 严格不做：

- Saved webhook connector、connector management、endpoint vault。
- Notification center、订阅规则、Slack/Teams/Email/IM 集成。
- Batch delivery、Case 级 delivery、workspace 级 delivery。
- Retry queue、后台任务、调度、cron、dead-letter。
- Report center、报告版本管理、PDF/DOCX/ZIP/SARIF。
- Scanner Hub 扩展、扫描器调度、真实扫描任务。
- 生命周期治理、retention、归档、删除策略。
- Workflow engine、SOAR playbook、ticket 自动创建器。
- MCP 新工具或 Agent 自动投递。
- 真实 LLM 调用。
- 高风险动作，例如删除、撤销、revoke token、修改生产配置。

## 输入文档

Windsurf 实施前必须阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-p-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-p-acceptance.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-q-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-q-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- Sprint Q 已新增 `POST /api/v2/agent-runs/:id/review-delivery`。
- Delivery 成功会更新 `review_delivery.webhook` operation result 为 `delivered`。
- Delivery 失败 / timeout 会更新 `review_delivery.webhook` operation result 为 `failed`，并写 `agent_run.review_delivery_failed` audit。
- Operation request 已保存 sanitized `destination_host`、`format`、`review_packet_id`、`include_audit`、`header_names`。
- Audit details 已保存 `delivery_id`、`delivery_operation_id`、`export_operation_id`、`destination_host`、`status_code` 或 error summary。
- Agent Run Detail 已有 Delivery dialog、Delivery Receipt、Operation timeline、Audit link。
- `frontend-next/e2e/agent-runs.spec.ts` 已覆盖 delivery happy path、headers、blocked URL。

### 主要缺口

- Delivery Receipt 只在提交后的 dialog 中出现，刷新后没有专门 history view。
- Operator 需要从通用 operation timeline/audit 页面手动找 delivery 记录。
- 没有单独 API 返回 delivery attempts summary。
- E2E 还不能证明页面刷新后 delivery history 仍可回查。
- 缺少 delivered / failed / timeout 维度的 delivery count/stats。

## 术语边界

### Review Delivery History

Review Delivery History 是单个 Agent Run 的 delivery attempt 列表，由 `review_delivery.webhook` operations 派生，不是新实体表，也不是 delivery queue。

### Delivery Attempt

Delivery Attempt 是一次 webhook delivery 尝试。最小状态：

- `delivered`: webhook 返回 2xx。
- `failed`: webhook 返回非 2xx 或网络错误。
- `timeout`: webhook 请求超时。若当前 operation result 只有 error 字符串，可由 error summary 包含 `timed out` 派生。

### Delivery Receipt Detail

Delivery Receipt Detail 是一个 sanitized 展示视图，只展示 destination host、status code、format、header names、operation/audit refs、error summary，不展示 full webhook URL、header values 或 response body。

## 数据契约

### AgentRunReviewDeliveryHistoryResponse

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunReviewDeliveryHistoryResponse struct {
	AgentRunID string                           `json:"agent_run_id"`
	Summary    AgentRunReviewDeliverySummary    `json:"summary"`
	Items      []AgentRunReviewDeliveryHistoryItem `json:"items"`
}
```

### AgentRunReviewDeliverySummary

```go
type AgentRunReviewDeliverySummary struct {
	Total     int `json:"total"`
	Delivered int `json:"delivered"`
	Failed    int `json:"failed"`
	Timeout   int `json:"timeout"`
}
```

### AgentRunReviewDeliveryHistoryItem

```go
type AgentRunReviewDeliveryHistoryItem struct {
	DeliveryID          string    `json:"delivery_id,omitempty"`
	DeliveryOperationID string    `json:"delivery_operation_id"`
	ExportOperationID   string    `json:"export_operation_id,omitempty"`
	AuditRefID          string    `json:"audit_ref_id,omitempty"`
	Format              string    `json:"format"`
	Result              string    `json:"result"`
	DestinationHost     string    `json:"destination_host"`
	StatusCode          int       `json:"status_code,omitempty"`
	HeaderNames         []string  `json:"header_names,omitempty"`
	ErrorSummary        string    `json:"error_summary,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	DeliveredAt         time.Time `json:"delivered_at,omitempty"`
}
```

`result` 只允许：

```text
delivered
failed
timeout
pending
```

`pending` 只作为数据兼容兜底展示，不引入后台队列。

## API 范围

新增：

```text
GET /api/v2/agent-runs/:id/review-deliveries
```

Response:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "agent_run_id": "agent-run-1",
    "summary": {
      "total": 3,
      "delivered": 1,
      "failed": 1,
      "timeout": 1
    },
    "items": [
      {
        "delivery_id": "delivery-123",
        "delivery_operation_id": "op-delivery-1",
        "export_operation_id": "op-export-1",
        "audit_ref_id": "audit-delivery-1",
        "format": "markdown",
        "result": "delivered",
        "destination_host": "hooks.example.com",
        "status_code": 200,
        "header_names": ["X-GODNSLOG-Source"],
        "created_at": "2026-06-08T00:00:00Z",
        "delivered_at": "2026-06-08T00:00:01Z"
      }
    ]
  }
}
```

Error behavior:

- `404`: Agent Run not found.
- `500`: unexpected read/parse error.

This API is read-only and low-risk. It must not trigger webhook calls, exports, retries, or state changes.

## Backend 实施要求

### Service

在 `internal/agentrun/review.go` 或合适文件中新增：

```go
func (s *ReviewService) ListReviewDeliveries(agentRunID string) (*models.AgentRunReviewDeliveryHistoryResponse, error)
```

Implementation rules:

- Validate the Agent Run exists.
- Query `AgentOperation` where:
  - `agent_run_i_d = ?`
  - `action = "review_delivery.webhook"`
- Order by `created_at DESC`.
- Parse operation `request` and `result` JSON defensively.
- Derive:
  - `format` / `destination_host` / `header_names` from operation request.
  - `delivery_id` / `status_code` / `export_operation_id` / `error_summary` from operation result.
  - `result = timeout` when result is failed and error contains timeout / timed out.
- Resolve `audit_ref_id` by matching audit details `delivery_operation_id`, or by using audit action/resource/time fallback if exact details are unavailable.
- Never include full webhook URL or header values.
- Use small, bounded error summaries. Do not include response body.

### Tests

Add Go tests covering:

- Delivered item is listed with status code, destination host, operation ref, audit ref.
- Failed non-2xx item is listed as `failed` with bounded error summary.
- Timeout item is listed as `timeout`.
- Summary counts are correct.
- Header names are included but header values are not present in serialized response.
- Full webhook URL is not present in serialized response.
- Non-existent Agent Run returns not found.

## Frontend 实施要求

### API Client / Types

Add TypeScript types:

- `AgentRunReviewDeliveryHistoryResponse`
- `AgentRunReviewDeliverySummary`
- `AgentRunReviewDeliveryHistoryItem`

Add API client method:

```ts
agentRunApi.listReviewDeliveries(id)
```

### Agent Run Detail UI

In `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` add a `Delivery History` section near Review Export / Delivery:

- Summary chips:
  - Total
  - Delivered
  - Failed
  - Timeout
- Recent delivery table/list:
  - Result badge
  - Destination host
  - Format
  - Status code
  - Created time / delivered time
  - Header names
  - Operation ID
  - Audit ref link
- Empty state when no delivery attempts exist.
- After successful `Deliver to Webhook`, refresh delivery history automatically.
- Do not display full webhook URL or header values.

UI should remain compact and operational; no marketing/hero layout.

## E2E 范围

Update `frontend-next/e2e/agent-runs.spec.ts`:

### Happy path history loop

1. Mock Agent Run Detail with review packet.
2. Mock initial `GET /review-deliveries` as empty.
3. Submit `Deliver to Webhook`.
4. Mock refreshed `GET /review-deliveries` with a delivered item.
5. Assert Delivery History shows:
   - delivered result
   - destination host
   - status code
   - `review_delivery.webhook` operation ref
   - audit ref link
   - header name
6. Assert full webhook URL and header value are not visible.
7. Navigate to Audit via history link and assert `agent_run.review_delivered`.

### Failure / timeout display

1. Mock `GET /review-deliveries` with failed and timeout items.
2. Assert summary counts total/delivered/failed/timeout.
3. Assert failed/timeout badges and bounded error summary are visible.
4. Assert no retry button exists.

## Documentation

Update:

- `docs/verification.md`
  - Add Sprint R commands and results after implementation.
- `docs/agent-native-specification.md`
  - Add a short note under Agent Run Export / Delivery: delivery history is operator-visible and read-only; it is not a notification center or retry queue.
- Optional: `docs/MCP_SERVER_USAGE.md`
  - Clarify MCP does not auto-deliver review packages in Sprint R.

## 验收标准

- `GET /api/v2/agent-runs/:id/review-deliveries` returns sanitized delivery history derived from existing operations/audits.
- Summary counts delivered/failed/timeout are correct.
- Full webhook URL, header values, response bodies, API keys, cookies, authorization tokens never appear in response, operation/audit display, or E2E-visible UI.
- Agent Run Detail shows Delivery History and refreshes it after a successful delivery.
- Delivery History provides Operation/Audit refs and Audit navigation.
- E2E proves Delivery -> History refresh -> Receipt detail -> Operation/Audit loop.
- E2E includes failed/timeout display and proves no retry button / queue behavior exists.
- No scope creep into saved connectors, notification center, batch delivery, retry queue, report center, lifecycle governance, Scanner Hub, workflow engine, or MCP auto-delivery.

## 必跑验证

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

```bash
GOCACHE=/tmp/gocache go test ./...
```

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

```bash
cd frontend-next && npm run build
```

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

Playwright 必须使用一次性非交互式命令，不得执行 `npx playwright show-report`，不得触发 HTML report server 常驻。

## 实施 Checklist

- [ ] Read required input documents.
- [ ] Add backend models for delivery history response/summary/item.
- [ ] Implement `ReviewService.ListReviewDeliveries`.
- [ ] Add `GET /api/v2/agent-runs/:id/review-deliveries` route and handler.
- [ ] Add Go tests for delivered/failed/timeout/history sanitization/not found.
- [ ] Add frontend API client method and TypeScript types.
- [ ] Add Agent Run Detail Delivery History section.
- [ ] Refresh Delivery History after successful delivery.
- [ ] Add E2E happy path history loop.
- [ ] Add E2E failed/timeout history display.
- [ ] Update docs.
- [ ] Run required verification commands and record results.
