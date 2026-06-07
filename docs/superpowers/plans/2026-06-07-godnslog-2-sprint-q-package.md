# GODNSLOG 2.0 Sprint Q Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Extend Sprint P's single Agent Run Review Evidence Package from "exportable" to "deliverable": an operator can send one reviewed Agent Run export package to a configured external webhook, and GODNSLOG records a low-risk Agent Operation plus Audit receipt proving what was delivered, where, and with which sanitized package reference.

**Architecture:** Reuse Sprint P `POST /api/v2/agent-runs/:id/review-export`, Agent Operation, Audit, and Agent Run Detail UI. Add a single-run, foreground webhook delivery action with strict URL safety checks and no background queue. Do not introduce report center, notification center, workflow engine, saved connectors, batch delivery, retries, lifecycle governance, Scanner Hub expansion, or PDF/DOCX/ZIP.

**Tech Stack:** Go, xorm, Gin routes in `server/v2_api.go`, `internal/agentrun`, `internal/auth`, standard `net/http` with bounded timeout, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint Q
- **Sprint 主题**：Review Evidence Delivery Receipt
- **所属阶段**：Phase 5 - Agent Governance and Review Operations

## Sprint 背景

Sprint L 完成单个 Agent Run Review Packet。

Sprint M 完成最小 Follow-up Action。

Sprint N 完成 Review Queue 和 Follow-up History。

Sprint O 完成 Review Decision 和 Queue Closure。

Sprint P 完成单个 Review Evidence Export Package：

```text
Agent Run Detail -> Export JSON / Markdown -> Operation -> Audit
```

现在 operator 可以把已复核 Agent Run 导出为 JSON / Markdown，并能在 Operation timeline 和 Audit 中证明导出动作。但实际安全团队还需要把这个 evidence package 交给外部工单、SOAR、知识库、漏洞管理平台或自建 Webhook receiver。当前只能人工复制，缺少一个最小、可审计、可回查的“交付回执”。

Sprint Q 的目标是补齐单条导出包的外部交付动作，不做通知中心或自动化工作流。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 增加单个 Agent Run Review Evidence Package 的 Webhook delivery API。
2. Delivery 复用 Sprint P export package，支持 `json` / `markdown`，不重新定义报告格式。
3. Delivery 写入 `review_delivery.webhook` Agent Operation 和 `agent_run.review_delivered` Audit。
4. Agent Run Detail 提供最小 Deliver to Webhook UI，并展示 delivery receipt。
5. E2E 证明：Detail -> Export/Deliver -> Operation timeline -> Audit -> webhook payload sanitized。

## 明确不做

本 Sprint 严格不做：

- 批量 delivery、Case 级 delivery、workspace 级 delivery。
- 报表中心、通知中心、连接器中心、保存 webhook 配置。
- 后台任务队列、重试队列、调度、cron、dead-letter。
- PDF、DOCX、ZIP、SARIF。
- Scanner Hub 扩展、扫描器调度、真实扫描任务。
- 生命周期治理、retention、归档、删除策略。
- Workflow engine、SOAR playbook、自动 ticket 创建器。
- MCP 新工具或 Agent 自动投递。
- 真实 LLM 调用。
- 高风险动作，例如删除、撤销、revoke token、修改生产配置。

## 输入文档

Windsurf 实施前必须阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-n-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-n-acceptance.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-o-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-o-acceptance.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-p-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-p-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- Sprint P 已新增 `POST /api/v2/agent-runs/:id/review-export`。
- `internal/agentrun.ReviewService.ExportReviewPackage` 已能生成 JSON / Markdown export response。
- Export 已写入 `review_export.<format>` operation 和 `agent_run.review_exported` audit。
- Agent Run Detail 已有 Export JSON / Markdown UI、Export Result、Audit link。
- `server/v2_api.go` 已有 Agent Run 路由和 Audit 路由。
- `frontend-next/e2e/agent-runs.spec.ts` 已覆盖 Review Export 的 UI、operation、audit。
- 仓库已有 webhook 相关历史代码，但它们不应被扩展成通知中心或 workflow engine。

### 主要缺口

- Operator 无法从 Agent Run Detail 将已复核证据包直接交付给外部 Webhook。
- Delivery 没有独立 operation / audit receipt。
- 当前外部交付如果靠手工复制，无法在 GODNSLOG 中证明交付目标、时间、格式、状态码和 package reference。
- 需要避免把任意 webhook URL 变成 SSRF 能力。
- E2E 还不能证明 export package -> external delivery -> receipt -> audit 的闭环。

## 术语边界

### Review Evidence Delivery

Review Evidence Delivery 是对单个 Agent Run Review Evidence Package 的一次外部交付动作。它不是报告实体，也不是通知规则。

### Delivery Receipt

Delivery Receipt 是 operation / audit 中记录的交付结果摘要，至少包含：

- delivery format。
- destination host 的安全展示。
- HTTP status code。
- package operation ref。
- audit ref。
- delivered_at。

### Out of Scope: Connector / Notification Center

本 Sprint 不保存 webhook endpoint，不提供 connector 列表，不做重试策略，也不把 delivery 升级成 workflow action。

## 数据契约

### AgentRunReviewDeliveryRequest

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunReviewDeliveryRequest struct {
	Format         string            `json:"format" binding:"required"`
	ReviewPacketID string            `json:"review_packet_id,omitempty"`
	WebhookURL     string            `json:"webhook_url" binding:"required"`
	Headers        map[string]string `json:"headers,omitempty"`
	IncludeAudit   bool              `json:"include_audit,omitempty"`
}
```

`format` 只允许：

```text
json
markdown
```

### AgentRunReviewDeliveryResponse

```go
type AgentRunReviewDeliveryResponse struct {
	AgentRunID        string    `json:"agent_run_id"`
	Format            string    `json:"format"`
	DeliveryID        string    `json:"delivery_id"`
	DeliveryOperation string    `json:"delivery_operation_id"`
	ExportOperationID string    `json:"export_operation_id,omitempty"`
	AuditRefID        string    `json:"audit_ref_id,omitempty"`
	DestinationHost   string    `json:"destination_host"`
	StatusCode        int       `json:"status_code"`
	Result            string    `json:"result"`
	DeliveredAt       time.Time `json:"delivered_at"`
}
```

### Webhook Payload Contract

Webhook body 必须是 JSON envelope，即使 `format=markdown`：

```json
{
  "event": "agent_run.review_evidence_delivered",
  "agent_run_id": "agent-run-1",
  "format": "markdown",
  "delivery_id": "delivery-...",
  "generated_at": "2026-06-07T00:00:00Z",
  "package": {
    "content": "# Agent Run Review Evidence Package\n..."
  },
  "refs": {
    "export_operation_id": "op-export-1",
    "delivery_operation_id": "op-delivery-1",
    "audit_ref_id": "audit-delivery-1"
  }
}
```

For `format=json`, `package` contains Sprint P JSON package.

For `format=markdown`, `package.content` contains Sprint P Markdown content.

### Sanitization Contract

Delivery request, operation result, audit details, and webhook payload must not include:

- Full API key.
- Authorization header value.
- Cookie header value.
- Raw token secret.
- DNS callback secret.
- HTTP raw sensitive headers/body.
- Production database connection string.
- Full webhook URL in audit or operation result.

Allowed:

- Destination host only, e.g. `hooks.example.com`.
- HTTP status code.
- Header names, but not sensitive header values.
- Case ID / Payload ID / Agent Run ID.
- Operation ID / Audit ID.
- Evidence strength, confidence, interaction count.

## Security Contract

Webhook delivery must be guarded:

- Only `https://` is allowed by default.
- `http://127.0.0.1`, `localhost`, link-local, private RFC1918 ranges, metadata IPs, and Unix sockets must be rejected.
- Redirects must not be followed, or must be blocked if they leave the validated host.
- Timeout must be short and explicit, recommended 5 seconds.
- Request body size must be bounded.
- Response body must not be stored; only status code and small error summary are recorded.
- Header allowlist: permit `Content-Type` and `X-*` headers; reject `Authorization`, `Cookie`, `Set-Cookie`, `Proxy-*`, and hop-by-hop headers.

Local test servers may be used in tests via an explicit test-only path or injected HTTP client; do not weaken production URL validation for tests.

## API 范围

新增：

```text
POST /api/v2/agent-runs/:id/review-delivery
```

Request:

```json
{
  "format": "markdown",
  "review_packet_id": "agent-run-1",
  "webhook_url": "https://hooks.example.com/godnslog/review",
  "headers": {
    "X-GODNSLOG-Source": "operator"
  },
  "include_audit": true
}
```

Response:

```json
{
  "code": 0,
  "data": {
    "agent_run_id": "agent-run-1",
    "format": "markdown",
    "delivery_id": "delivery-1",
    "delivery_operation_id": "op-delivery-1",
    "export_operation_id": "op-export-1",
    "audit_ref_id": "audit-delivery-1",
    "destination_host": "hooks.example.com",
    "status_code": 200,
    "result": "delivered",
    "delivered_at": "2026-06-07T00:00:00Z"
  }
}
```

错误映射：

- 400: invalid format、invalid URL、blocked destination、forbidden header、invalid review_packet_id。
- 401: unauthenticated。
- 404: Agent Run not found。
- 502: webhook returned non-2xx or transport failed。
- 504: webhook timeout。

## Agent Operation Contract

Delivery 成功或失败都必须写入 Agent Operation：

```text
action = review_delivery.webhook
risk_level = low
```

`request` JSON：

```json
{
  "format": "markdown",
  "review_packet_id": "agent-run-1",
  "destination_host": "hooks.example.com",
  "include_audit": true,
  "header_names": ["X-GODNSLOG-Source"]
}
```

`result` JSON：

```json
{
  "result": "delivered",
  "status_code": 200,
  "delivery_id": "delivery-1",
  "export_operation_id": "op-export-1",
  "audit_action": "agent_run.review_delivered"
}
```

失败时：

```json
{
  "result": "failed",
  "status_code": 502,
  "error": "webhook returned non-2xx",
  "delivery_id": "delivery-1"
}
```

Do not store full webhook URL, secret header values, or response body.

## Audit Contract

Delivery 成功写入：

```text
action = agent_run.review_delivered
resource_type = agent_run
resource_id = <agent_run_id>
result = success
```

Delivery 失败写入：

```text
action = agent_run.review_delivery_failed
resource_type = agent_run
resource_id = <agent_run_id>
result = failure
```

Details:

```json
{
  "format": "markdown",
  "delivery_id": "delivery-1",
  "delivery_operation_id": "op-delivery-1",
  "export_operation_id": "op-export-1",
  "destination_host": "hooks.example.com",
  "status_code": 200
}
```

## 前端范围

更新 `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`：

- 在 Review Packet / Export 区域增加 `Deliver to Webhook` 按钮。
- Dialog 字段：
  - Format: JSON / Markdown segmented control。
  - Webhook URL input。
  - Optional header key/value rows，默认只支持 `X-*`。
  - Include Audit checkbox。
- Submit 后展示 Delivery Receipt：
  - delivery result。
  - destination host。
  - status code。
  - delivery operation id。
  - audit ref link。
- 错误状态必须展示 blocked URL / forbidden header / timeout / non-2xx。

UI 不应描述实现细节或安全规则长文案；错误信息可以短句说明。

## E2E 验收范围

更新 `frontend-next/e2e/agent-runs.spec.ts`：

必须新增真实行为测试，不允许 `test.skip` / `test.only`。

E2E 必须证明：

1. Agent Run Detail 点击 `Deliver to Webhook`。
2. 选择 `Markdown`，输入 `https://hooks.example.test/review`。
3. 请求命中 `POST /api/v2/agent-runs/:id/review-delivery`。
4. 请求体包含：
   - `format=markdown`
   - `review_packet_id=agent-run-1`
   - sanitized headers
   - 不包含 secret value
5. Dialog 中显示 Delivery Receipt。
6. Operation timeline 显示 `review_delivery.webhook`。
7. Audit 页面显示 `agent_run.review_delivered`。
8. E2E 不能只断言静态文字；必须检查 API request body、post-delivery detail refresh、audit page API mock。

Error E2E:

- 输入 blocked URL，例如 `http://127.0.0.1:8080/hook`，UI 显示错误且不写成功 receipt。

## 后端测试范围

新增或更新 Go 测试：

- `internal/agentrun` delivery service tests：
  - successful markdown delivery。
  - successful json delivery。
  - invalid format。
  - unknown Agent Run。
  - review_packet_id mismatch。
  - blocked localhost / private IP / metadata IP。
  - forbidden sensitive header。
  - webhook non-2xx writes failure operation/audit。
  - webhook timeout maps to timeout error。
  - operation/audit do not store full webhook URL or sensitive header values。
- `server/v2_api_test.go`：
  - authenticated success。
  - unauthenticated 401。
  - bad URL 400。
  - non-2xx 502。

Tests should use `httptest.Server` or injected HTTP client. Do not weaken production URL validation to make tests pass.

## 文档更新

Update:

- `docs/MCP_SERVER_USAGE.md`
  - Clarify Sprint Q Web Review Delivery is operator Web API only.
  - MCP `export_report` remains read-only and does not deliver to webhook.
- `docs/verification.md`
  - Record exact commands and results.

Optional:

- Add a short note in `docs/agent-native-specification.md` under Agent Run Export / Delivery that external delivery is operator-driven and auditable.

## 实施步骤

- [ ] Read Sprint P plan and acceptance before editing.
- [ ] Add delivery request/response DTOs.
- [ ] Add URL/header validation helpers with unit tests.
- [ ] Add delivery service method reusing Sprint P export package.
- [ ] Add route `POST /api/v2/agent-runs/:id/review-delivery`.
- [ ] Write success and failure operation/audit records.
- [ ] Add backend tests for success, failure, security validation, and sanitization.
- [ ] Add frontend API client/types.
- [ ] Add Agent Run Detail delivery dialog and receipt display.
- [ ] Add E2E happy path and blocked URL path.
- [ ] Update docs and verification log.
- [ ] Run required verification commands.

## 验收标准

Sprint Q 只有在以下全部满足时才可通过：

- Single Agent Run delivery API exists and is authenticated.
- Delivery reuses Sprint P export package; no duplicate package format is introduced.
- Successful delivery writes `review_delivery.webhook` operation and `agent_run.review_delivered` audit.
- Failed delivery writes failure operation/audit without pretending success.
- URL/header safety checks block obvious SSRF and sensitive header cases.
- Operation/audit never store full webhook URL, authorization/cookie values, raw response body, or secrets.
- Agent Run Detail UI exposes delivery action and receipt.
- E2E proves Detail -> Delivery -> Operation timeline -> Audit loop.
- E2E includes blocked URL error behavior.
- No `test.skip` / `test.only`.
- No scope creep into report center, notification center, saved connectors, batch delivery, retries, lifecycle governance, Scanner Hub, workflow engine, or MCP auto-delivery.

## 必跑验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

Playwright 必须使用一次性非交互式 reporter，不得使用 `npx playwright show-report`，不得触发常驻 HTML report server。

## Windsurf 交付说明

交付时请明确列出：

- 新增/修改的文件。
- Delivery API request/response 示例。
- URL/header 安全规则。
- Operation / Audit action 名称。
- Go 测试命令和结果。
- Frontend lint/build/E2E 命令和结果。
- 未做事项确认，尤其是未做 batch、report center、notification center、workflow engine、Scanner Hub、MCP auto-delivery。
