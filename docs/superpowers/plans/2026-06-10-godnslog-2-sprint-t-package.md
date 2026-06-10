# GODNSLOG 2.0 Sprint T Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a read-only Package Hash Trace view for Agent Run Review Evidence packages, so an operator can paste a `package_hash` and verify which Agent Run export, delivery attempts, delivery history entries, and audit records reference the same sanitized evidence package.

**Architecture:** Reuse Sprint S `package_hash` values already stored in Agent Operations, Audit details, Delivery responses, Delivery History, and UI. Add a read-only trace API plus a focused UI entry point. Do not introduce report storage, package storage, signatures, PKI, saved connectors, retry queues, batch delivery, Scanner Hub expansion, lifecycle governance, workflow engine, or MCP auto-delivery.

**Tech Stack:** Go, xorm, Gin routes in `server/v2_api.go`, `internal/agentrun`, `internal/auth`, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint T
- **Sprint 主题**：Review Package Hash Trace Lookup
- **所属阶段**：Phase 5 - Agent Governance and Review Operations

## Sprint 背景

Sprint S 已完成单个 Review Evidence Package 的完整性标识：

```text
Export Result -> Delivery Receipt -> Delivery History -> Audit details
```

现在 operator 可以在单个 Agent Run Detail 页面里看到同一个 `package_hash`。但复盘时经常是反向问题：

- 我手里有一个 webhook payload / audit 截图里的 `package_hash`，它属于哪个 Agent Run？
- 这份 package 是否确实被导出过？
- 它被 delivery 到哪些 destination host？
- 成功、失败、timeout 分别有哪些？
- 对应的 export operation、delivery operation、audit ref 是什么？

Sprint T 的目标是做一个只读 trace lookup，不创建 package 仓库，不存原始 package，不生成报告版本。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 新增 read-only Package Hash Trace API。
2. 从已有 `AgentOperation.result`、`AuditLog.details` 和 delivery history 派生 trace，不新增 package 实体表。
3. 返回 export refs、delivery refs、audit refs、Agent Run refs、format、result、destination host、status code 等 sanitized 信息。
4. 在 Audit 页面或 Agent Run Detail 中提供 Package Hash trace 入口，支持粘贴 hash 查询。
5. E2E 证明：输入同一个 `package_hash` 后，页面展示 Export -> Delivery -> History -> Audit 的只读追踪闭环。

## 明确不做

本 Sprint 严格不做：

- Report center、报告实体、报告版本、长期归档。
- 存储完整 package content、markdown content 或 webhook payload body。
- 数字签名、私钥管理、PKI、JWS、证书链。
- PDF、DOCX、ZIP、SARIF。
- Batch export / batch delivery / Case 级 package。
- Saved webhook connector、notification center、retry queue、后台任务。
- Scanner Hub 扩展、扫描器调度、真实扫描任务。
- 生命周期治理、retention、删除、归档策略。
- Workflow engine、SOAR playbook、ticket 自动创建。
- MCP 新工具或 Agent 自动投递。
- 真实 LLM 调用。
- 高风险动作，例如删除、撤销、revoke token、修改生产配置。

## 输入文档

Windsurf 实施前必须阅读：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/agent-native-specification.md`
- `docs/MCP_SERVER_USAGE.md`
- `docs/superpowers/plans/2026-06-08-godnslog-2-sprint-r-package.md`
- `docs/superpowers/acceptance/2026-06-08-godnslog-2-sprint-r-acceptance.md`
- `docs/superpowers/plans/2026-06-08-godnslog-2-sprint-s-package.md`
- `docs/superpowers/acceptance/2026-06-08-godnslog-2-sprint-s-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- Export operation result 已包含 `package_hash`。
- Export audit details 已包含 `package_hash`。
- Delivery webhook payload / response / operation / audit 已包含 `package_hash`。
- Delivery History item 已包含 `package_hash`。
- Agent Run Detail 已展示 Export Result、Delivery Receipt、Delivery History hash。
- Audit 页面已展示 Package Hash。
- E2E 已证明单个 Agent Run 内的 hash 闭环。

### 主要缺口

- Operator 无法从 `package_hash` 反查 Agent Run。
- Audit 页面只能浏览现有列表，不能专门按 hash 做 trace。
- Delivery History 是单 Agent Run 视角，不能从 hash 反向聚合。
- E2E 尚未证明“已知 hash -> trace view -> refs/audit”的反向追踪闭环。

## 术语边界

### Package Hash Trace

Package Hash Trace 是一个 read-only 派生视图，用于根据 `package_hash` 聚合已有 operation、audit、delivery history 记录。

它不是 package registry，不是报告仓库，不是长期归档系统。

### Trace Item

Trace Item 是一个引用同一 `package_hash` 的 export、delivery 或 audit 记录。

允许的 item type：

```text
export
delivery
audit
```

## 数据契约

### AgentRunReviewPackageTraceResponse

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunReviewPackageTraceResponse struct {
	PackageHash string                            `json:"package_hash"`
	Summary     AgentRunReviewPackageTraceSummary `json:"summary"`
	AgentRuns   []AgentRunReviewPackageTraceRun   `json:"agent_runs"`
	Exports     []AgentRunReviewPackageTraceExport `json:"exports"`
	Deliveries  []AgentRunReviewPackageTraceDelivery `json:"deliveries"`
	Audits      []AgentRunReviewPackageTraceAudit  `json:"audits"`
}
```

### Summary

```go
type AgentRunReviewPackageTraceSummary struct {
	AgentRunCount int `json:"agent_run_count"`
	ExportCount   int `json:"export_count"`
	DeliveryCount int `json:"delivery_count"`
	AuditCount    int `json:"audit_count"`
	Delivered     int `json:"delivered"`
	Failed        int `json:"failed"`
	Timeout       int `json:"timeout"`
}
```

### Agent Run Ref

```go
type AgentRunReviewPackageTraceRun struct {
	AgentRunID string `json:"agent_run_id"`
	Title      string `json:"title,omitempty"`
	Status     string `json:"status,omitempty"`
	CaseID     string `json:"case_id,omitempty"`
	PayloadID  string `json:"payload_id,omitempty"`
	Target     string `json:"target,omitempty"`
	URL        string `json:"url,omitempty"`
}
```

### Export Ref

```go
type AgentRunReviewPackageTraceExport struct {
	AgentRunID     string    `json:"agent_run_id"`
	OperationID    string    `json:"operation_id"`
	AuditRefID     string    `json:"audit_ref_id,omitempty"`
	ReviewPacketID string    `json:"review_packet_id,omitempty"`
	Format         string    `json:"format"`
	CreatedAt      time.Time `json:"created_at"`
}
```

### Delivery Ref

```go
type AgentRunReviewPackageTraceDelivery struct {
	AgentRunID           string    `json:"agent_run_id"`
	DeliveryID           string    `json:"delivery_id,omitempty"`
	DeliveryOperationID  string    `json:"delivery_operation_id"`
	ExportOperationID    string    `json:"export_operation_id,omitempty"`
	AuditRefID           string    `json:"audit_ref_id,omitempty"`
	Format               string    `json:"format"`
	Result               string    `json:"result"`
	DestinationHost      string    `json:"destination_host,omitempty"`
	StatusCode           int       `json:"status_code,omitempty"`
	ErrorSummary         string    `json:"error_summary,omitempty"`
	CreatedAt            time.Time `json:"created_at"`
	DeliveredAt          time.Time `json:"delivered_at,omitempty"`
}
```

### Audit Ref

```go
type AgentRunReviewPackageTraceAudit struct {
	AuditRefID   string    `json:"audit_ref_id"`
	AgentRunID   string    `json:"agent_run_id,omitempty"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
	URL          string    `json:"url,omitempty"`
}
```

## API 范围

新增 read-only API：

```text
GET /api/v2/agent-runs/review-package-trace?package_hash=<sha256>
```

行为要求：

- 必须鉴权。
- `package_hash` 必填。
- `package_hash` 必须是 64 位 hex string；非法返回 400。
- 查询结果为空时返回 200，summary 全 0，数组为空。
- 不返回 package content、markdown content、webhook URL、header values、response body、API key、Authorization、Cookie。
- API 只读，不创建 operation，不写 audit，不触发 delivery/export。

示例响应：

```json
{
  "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678",
  "summary": {
    "agent_run_count": 1,
    "export_count": 1,
    "delivery_count": 2,
    "audit_count": 3,
    "delivered": 1,
    "failed": 1,
    "timeout": 0
  },
  "agent_runs": [
    {
      "agent_run_id": "agent-run-1",
      "title": "SSRF probe review",
      "status": "completed",
      "case_id": "case-1",
      "payload_id": "payload-1",
      "target": "https://target.example",
      "url": "/dashboard/agent-runs/agent-run-1"
    }
  ],
  "exports": [
    {
      "agent_run_id": "agent-run-1",
      "operation_id": "op-export-1",
      "audit_ref_id": "audit-export-1",
      "review_packet_id": "review-packet-1",
      "format": "json",
      "created_at": "2026-06-10T00:00:00Z"
    }
  ],
  "deliveries": [
    {
      "agent_run_id": "agent-run-1",
      "delivery_id": "delivery-1",
      "delivery_operation_id": "op-delivery-1",
      "export_operation_id": "op-export-1",
      "audit_ref_id": "audit-delivery-1",
      "format": "json",
      "result": "delivered",
      "destination_host": "hooks.example.com",
      "status_code": 200,
      "created_at": "2026-06-10T00:01:00Z",
      "delivered_at": "2026-06-10T00:01:01Z"
    }
  ],
  "audits": [
    {
      "audit_ref_id": "audit-export-1",
      "agent_run_id": "agent-run-1",
      "action": "agent_run.review_exported",
      "resource_type": "agent_run",
      "resource_id": "agent-run-1",
      "timestamp": "2026-06-10T00:00:00Z",
      "url": "/dashboard/audit?resource_type=agent_run&resource_id=agent-run-1"
    }
  ]
}
```

## Backend 实施要求

### Trace Service

在 `internal/agentrun` 中实现：

```go
func (s *ReviewService) TraceReviewPackageByHash(packageHash string) (*models.AgentRunReviewPackageTraceResponse, error)
```

实现原则：

- 只读查询。
- 从 `AgentOperation` 中查找 result 包含 `package_hash` 的记录：
  - `review_export.json`
  - `review_export.markdown`
  - `review_delivery.webhook`
- 从 `AuditLog` 中查找 details 包含 `package_hash` 的记录：
  - `agent_run.review_exported`
  - `agent_run.review_delivered`
  - `agent_run.review_delivery_failed`
- 能关联 Agent Run 时补充 Agent Run title/status/case/payload/target。
- Delivery result 仍按 Sprint R/S 规则派生为 `delivered` / `failed` / `timeout`。
- 对重复引用去重：同一个 operation 或 audit 不应重复出现。

### API Handler

在 `server/v2_api.go` 增加：

```go
GET /api/v2/agent-runs/review-package-trace
```

测试覆盖：

- 未登录返回 401。
- 缺少 `package_hash` 返回 400。
- 非 64 hex 返回 400。
- 无匹配返回 200 empty summary。
- 匹配 export operation + export audit。
- 匹配 delivery success / failure / timeout。
- 响应不包含 full webhook URL、header values、response body、API key、Authorization、Cookie。

## Frontend 实施要求

### API Client / Types

在 `frontend-next/src/types/index.ts` 增加 trace response 类型。

在 `frontend-next/src/lib/api-client.ts` 或现有 agent run API client 增加：

```ts
traceReviewPackage(packageHash: string)
```

### UI 入口

优先在 `/dashboard/audit` 页面增加一个轻量 Package Hash Trace 控制区：

- 输入框：`Package Hash`
- 查询按钮
- 清除按钮
- 64 hex 校验提示
- loading/error/empty states

显示区域：

- Summary：Agent Runs / Exports / Deliveries / Audits / Delivered / Failed / Timeout
- Agent Run refs：title、status、target、case/payload、进入 Agent Run Detail 链接
- Export refs：format、operation id、audit ref、created_at
- Delivery refs：result、destination host、status code、format、operation/audit refs、error summary
- Audit refs：action、timestamp、resource、进入 audit filtered view 链接

UI 必须保持运维工具风格，紧凑、可扫描，不做 marketing hero，不做卡片套卡片。

### Hash Click-through

在以下已有 hash 显示旁边增加 trace 入口：

- Agent Run Detail Export Result Package Hash
- Delivery Receipt Package Hash
- Delivery History Package Hash
- Audit 页面 Package Hash

允许实现为：

```text
Trace
```

或 hash 文本点击后跳转：

```text
/dashboard/audit?package_hash=<hash>
```

若实现 query param，Audit 页面应自动触发 trace 查询。

## E2E 要求

更新 `frontend-next/e2e/agent-runs.spec.ts`，或新增 `frontend-next/e2e/audit.spec.ts`。

必须使用一次性非交互式命令：

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

不得使用 `npx playwright show-report`，不得触发 HTML report server 常驻。

### E2E 1：Package Hash Trace from Audit Page

1. Mock `GET /api/v2/agent-runs/review-package-trace?package_hash=<hash>`。
2. 打开 `/dashboard/audit`。
3. 输入 64 hex hash。
4. 点击查询。
5. 断言请求真实包含 `package_hash` query。
6. 断言 Summary 数字展示正确。
7. 断言 Agent Run ref、Export ref、Delivery ref、Audit ref 展示正确。
8. 断言页面不包含 full webhook URL、header value、response body、Authorization、Cookie。

### E2E 2：Package Hash Click-through from Agent Run Detail

1. Mock Agent Run Detail，包含 Export Result / Delivery History hash。
2. 点击 hash 旁边 Trace 入口。
3. 跳转或打开 Audit trace 区。
4. 断言 trace API 被调用。
5. 断言同一个 compact hash、Agent Run ref、delivery result、audit action 可见。

### E2E 3：Invalid / Empty Trace

1. 输入非法 hash，断言不会调用 API，并显示校验错误。
2. 输入合法但无匹配 hash，断言 empty state，不显示假数据。

## 安全与隐私要求

不得返回或展示：

- full webhook URL。
- header values。
- response body。
- API key。
- Authorization header。
- Cookie。
- raw package content。
- markdown content。
- full webhook payload。

允许展示：

- package hash。
- destination host。
- header names。
- status code。
- operation ID。
- audit ref ID。
- agent run/case/payload IDs。

## 验收标准

Sprint T 只有在以下条件都满足时才算完成：

- `GET /api/v2/agent-runs/review-package-trace?package_hash=...` 存在且只读。
- 非法 hash 返回 400，空结果返回 200 empty。
- Trace response 可聚合 export、delivery、audit 和 Agent Run refs。
- Delivery delivered / failed / timeout summary 正确。
- UI 可输入或从 hash click-through 进入 trace。
- E2E 证明 hash trace API 请求真实发生，不是静态文字检查。
- E2E 证明 full webhook URL / header values / response body / secrets 不出现。
- 不新增 report center、package storage、signature/PKI、saved connector、retry queue、Scanner Hub、workflow engine、MCP auto-delivery。

## 必跑验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

如果新增 `audit.spec.ts`，还必须运行：

```bash
cd frontend-next && npx playwright test --reporter=line e2e/audit.spec.ts
```

验证结果必须写入 `docs/verification.md`。

## 实施任务清单

- [ ] Add package trace response models.
- [ ] Implement read-only package hash trace service.
- [ ] Add package hash validation helper and tests.
- [ ] Add `GET /api/v2/agent-runs/review-package-trace` route and handler.
- [ ] Add server tests for auth, validation, empty result, export trace, delivery trace, audit trace, and sanitization.
- [ ] Add frontend trace types and API client method.
- [ ] Add Audit page Package Hash Trace UI.
- [ ] Add trace click-through from existing package hash displays.
- [ ] Add E2E for audit page trace lookup.
- [ ] Add E2E for click-through from Agent Run Detail.
- [ ] Add E2E for invalid and empty trace states.
- [ ] Update `docs/verification.md` with actual command results.
