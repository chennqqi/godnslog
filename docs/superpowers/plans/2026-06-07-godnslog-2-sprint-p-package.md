# GODNSLOG 2.0 Sprint P Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the closed Agent Run review loop from Sprint O into a minimal exportable evidence package: an operator can export a single reviewed Agent Run as JSON or Markdown, and the export is traceable through Agent Operation and Audit.

**Architecture:** Reuse existing Agent Run Review Packet, Review Decision, Agent Operation, Audit, MCP `export_report`, and frontend Agent Run Detail contracts. Add one single-run export API and UI entry. Do not introduce report center, PDF/DOCX/ZIP, SARIF, batch export, retention/lifecycle governance, scanner scheduling, or workflow engines.

**Tech Stack:** Go, xorm, Gin routes in `server/v2_api.go`, `internal/agentrun`, `internal/auth`, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint P
- **Sprint 主题**：Agent Review Evidence Export Package
- **所属阶段**：Phase 5 - Agent Governance and Review Operations

## Sprint 背景

Sprint L 完成单个 Agent Run 的 Review Packet。

Sprint M 完成最小 Follow-up Action。

Sprint N 完成 Review Queue 和 Follow-up History。

Sprint O 完成 Review Decision 和 Queue Closure：

```text
Review Queue -> Agent Run Detail -> Review Decision -> Operation -> Audit -> Queue Closure
```

现在 operator 可以完成单个 Agent Run 的复核闭环，但闭环结果还不能被作为一个稳定的“证据交付包”导出。安全团队需要把复核完成后的内容交给漏洞报告、工单系统、审计系统或外部 Agent，但当前能力分散在 Review Packet、Operation timeline、Audit、Review Queue 中。

Sprint P 的目标是补齐最小导出包，不做完整报表中心。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 增加单个 Agent Run 的 Review Evidence Export API。
2. Export package 复用 Review Packet，并附带最近 Review Decision、Operation ref、Audit ref。
3. 导出动作写入 Agent Operation 和 Audit。
4. Agent Run Detail 提供 JSON / Markdown 导出入口。
5. E2E 证明：Detail -> Export JSON/Markdown -> Operation timeline -> Audit -> 不泄露敏感信息。

## 明确不做

本 Sprint 严格不做：

- 批量导出、Case 级批量 package、ZIP 导出。
- PDF、DOCX、SARIF。
- 报表中心、模板系统、报告版本管理。
- 生命周期治理、retention、归档、删除策略。
- Scanner Hub 扩展、扫描器调度、真实扫描任务。
- Agent replay engine、后台任务队列、自动重放。
- 多人审批流、SLA、assign、notification。
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
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-o-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-o-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- `internal/agentrun/review.go` 已能生成 Agent Run Review Packet。
- `GET /api/v2/agent-runs/:id/review?format=json|markdown` 已存在。
- `internal/agentrun/service.go` 已有 Review Decision，并写入 `review_decision.<decision>` operation。
- `agent_run.review_decision_recorded` audit 已存在。
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已有 Review Packet、Review Decision、Operation timeline、Audit link。
- `internal/mcp/server.go` 已有 `export_report`，并已有 Agent Run Review API 优先路径基础。
- `docs/MCP_SERVER_USAGE.md` 已描述 `export_report`。

### 主要缺口

- Web UI 上没有稳定的“导出已复核 Agent Run 证据包”入口。
- Review Packet 与 Review Decision 没有组合成一个导出契约。
- Web 导出动作没有 Agent Operation / Audit 可追踪记录。
- E2E 不能证明导出后的 Operation timeline 和 Audit 闭环。
- 当前导出语义容易和旧 Evidence export、MCP report export 混淆。

## 术语边界

### Review Evidence Package

Review Evidence Package 是单个 Agent Run 的复核证据包。它是导出视图，不是新实体表，也不是报告中心。

它由以下数据组合而成：

- Agent Run identity。
- Review Packet。
- 最近 Review Decision。
- 相关 Operation refs。
- 相关 Audit refs。
- Case / Payload / Evidence backlinks。

### Export Action

Export Action 是一次低风险、可审计操作：

```text
action = review_export.<format>
risk_level = low
audit_action = agent_run.review_exported
```

### Out of Scope: Report Center

本 Sprint 不提供报告列表、报告版本、模板管理、附件管理、PDF 渲染、批量打包或长期归档。

## 数据契约

### AgentRunReviewExportRequest

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunReviewExportRequest struct {
	Format         string `json:"format" binding:"required"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
	IncludeAudit   bool   `json:"include_audit,omitempty"`
}
```

`format` 只允许：

```text
json
markdown
```

### AgentRunReviewExportResponse

```go
type AgentRunReviewExportResponse struct {
	AgentRunID      string                 `json:"agent_run_id"`
	Format          string                 `json:"format"`
	OperationID     string                 `json:"operation_id"`
	AuditRefID      string                 `json:"audit_ref_id,omitempty"`
	ReviewPacketID  string                 `json:"review_packet_id,omitempty"`
	Decision        string                 `json:"decision,omitempty"`
	Content         string                 `json:"content,omitempty"`
	Package         map[string]interface{} `json:"package,omitempty"`
	GeneratedAt     time.Time              `json:"generated_at"`
}
```

### JSON Package Contract

JSON response 的 `package` 至少包含：

```json
{
  "agent_run": {
    "id": "agent-run-1",
    "agent_id": "agent-123",
    "case_id": "case-1",
    "payload_id": "payload-1",
    "target": "https://target.example",
    "status": "completed"
  },
  "review_packet": {
    "id": "agent-run-1",
    "evidence_strength": "high",
    "confidence": 85,
    "interaction_count": 5,
    "unique_sources": 2
  },
  "review_decision": {
    "decision": "accepted",
    "reason": "Evidence reviewed by operator",
    "operation_id": "op-decision-1",
    "audit_ref_id": "audit-decision-1"
  },
  "links": {
    "case_url": "/dashboard/cases/case-1",
    "payload_url": "/dashboard/payloads/payload-1",
    "evidence_url": "/dashboard/evidence?payload_id=payload-1",
    "audit_url": "/dashboard/audit?resource_type=agent_run&resource_id=agent-run-1"
  }
}
```

### Markdown Contract

Markdown response 的 `content` 必须包含：

```text
# Agent Run Review Evidence Package

## Agent Run
## Evidence Summary
## Review Decision
## Timeline References
## Audit References
## Links
```

Markdown 不要求文件下载；可在 UI 中预览并复制。

### Sanitization Contract

导出包不得包含：

- 完整 API Key。
- Authorization header。
- raw token secret。
- DNS callback secret。
- HTTP raw headers 中的敏感字段。
- 生产数据库连接串。

允许包含：

- Case ID / Payload ID。
- Token 的安全展示形式或 payload public identifier。
- Interaction count、source count、evidence strength、confidence。
- Operation ID、Audit ID。

## API 范围

新增：

```text
POST /api/v2/agent-runs/:id/review-export
```

Request:

```json
{
  "format": "markdown",
  "review_packet_id": "agent-run-1",
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
    "operation_id": "op-export-1",
    "audit_ref_id": "audit-export-1",
    "review_packet_id": "agent-run-1",
    "decision": "accepted",
    "content": "# Agent Run Review Evidence Package\n..."
  }
}
```

Validation:

- `format` 必须为 `json` 或 `markdown`。
- Agent Run 不存在返回 404。
- `review_packet_id` 可选，但如果提供必须和当前 Agent Run 相关。
- API 只能导出单个 Agent Run。
- 导出必须鉴权。

## Operation / Audit 契约

### AgentOperation

导出成功后追加 operation：

```text
action = review_export.<format>
risk_level = low
```

`result` 至少包含：

```json
{
  "format": "markdown",
  "agent_run_id": "agent-run-1",
  "review_packet_id": "agent-run-1",
  "decision": "accepted",
  "audit_action": "agent_run.review_exported",
  "exported_at": "2026-06-07T12:00:00Z"
}
```

### Audit

写入 audit：

```text
action = agent_run.review_exported
resource_type = agent_run
resource_id = <agent_run_id>
result = success
```

details 至少包含：

```json
{
  "format": "markdown",
  "review_packet_id": "agent-run-1",
  "decision": "accepted",
  "operation_id": "op-export-1"
}
```

## 前端范围

修改：

- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`

UI 要求：

- 在 Agent Run Detail 的 Review Packet / Review Decision 附近增加导出入口。
- 提供两个明确按钮：
  - `Export JSON`
  - `Export Markdown`
- 点击按钮必须调用 `POST /api/v2/agent-runs/:id/review-export`。
- JSON 导出可展示结构化预览。
- Markdown 导出可展示 Markdown preview。
- 成功后刷新 Agent Run Detail，让 Operation timeline 出现 `review_export.<format>`。
- 如果 response 有 `audit_ref_id`，页面必须提供 Audit 链接。
- 不做浏览器文件下载要求，避免跨环境 flakiness。

## MCP / CLI 边界

Sprint P 不要求重做 MCP `export_report`，但必须保持兼容：

- 不破坏 `export_report(agent_run_id=...)` 的现有路径。
- 不更改 APIKey scope。
- 不新增 CLI 命令。

可选但非必须：

- 文档说明 Web Review Export 与 MCP `export_report` 的关系。

## 测试要求

### 后端单测

新增或扩展：

- `internal/agentrun/service_test.go`
- `server/v2_api_test.go`

必须覆盖：

- JSON export 成功，返回 package。
- Markdown export 成功，返回 content。
- invalid format 返回 400。
- unknown Agent Run 返回 404。
- `review_packet_id` 不匹配返回 400。
- 成功导出写入 `review_export.json` / `review_export.markdown` operation。
- 成功导出写入 `agent_run.review_exported` audit。
- response 不泄露 Authorization / API key / secret。

### 前端 E2E

扩展：

- `frontend-next/e2e/agent-runs.spec.ts`

必须证明：

1. Detail 页面点击 `Export JSON`。
2. 等待并断言 `POST /api/v2/agent-runs/agent-run-1/review-export`。
3. 断言 request body 包含 `format=json` 和 `review_packet_id`。
4. Mock response 返回 `operation_id` / `audit_ref_id` / package。
5. Detail refresh 后 timeline 出现 `review_export.json`。
6. 点击或访问 Audit 链接后，Audit 页面/API 出现 `agent_run.review_exported`。
7. 点击 `Export Markdown` 后显示 Markdown preview，且包含 `# Agent Run Review Evidence Package`。
8. 不使用 `test.skip` / `test.only`。
9. 不只检查静态文案，必须断言 API request 和刷新后的 timeline/audit。

## 验证命令

必须运行并记录到 `docs/verification.md`：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

如果 Playwright 需要 dev server，必须使用非交互式两步法：

```bash
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

禁止：

- `npx playwright show-report`
- `npm run test:e2e:ui`
- 任何会触发 `Serving HTML report at http://localhost:9323. Press Ctrl+C to quit.` 的流程。

## 建议任务拆分

### Task 1：模型与 Service 契约

- [ ] 在 `internal/models/agent_run.go` 增加 Review Export request/response DTO。
- [ ] 在 `internal/agentrun/service.go` 增加 `ExportReviewPackage` 或等价 service 方法。
- [ ] 复用 `ReviewService.BuildReviewPacket`。
- [ ] 读取最近 `review_decision.*` operation。
- [ ] 生成 JSON package 和 Markdown content。
- [ ] 写入 `review_export.<format>` operation。
- [ ] 写入 `agent_run.review_exported` audit。

### Task 2：API

- [ ] 在 `server/v2_api.go` 增加 `POST /api/v2/agent-runs/:id/review-export`。
- [ ] 增加 request validation 和错误映射。
- [ ] 增加 `server/v2_api_test.go` 覆盖成功和失败场景。

### Task 3：前端 API / 类型

- [ ] 在 `frontend-next/src/types/index.ts` 增加 Review Export types。
- [ ] 在 `frontend-next/src/lib/api-client.ts` 增加 `agentRunApi.exportReview`。

### Task 4：前端 Detail UI

- [ ] 在 Agent Run Detail 增加 Export JSON / Export Markdown 按钮。
- [ ] 成功后展示 JSON 或 Markdown preview。
- [ ] 成功后刷新 Agent Run Detail。
- [ ] Operation timeline 显示 `review_export.<format>`。
- [ ] Audit ref 进入 Audit 页面。

### Task 5：E2E

- [ ] 扩展 `frontend-next/e2e/agent-runs.spec.ts`。
- [ ] 覆盖 JSON export API request / timeline / audit。
- [ ] 覆盖 Markdown export preview。
- [ ] 检查没有 `test.skip` / `test.only`。

### Task 6：文档与验证

- [ ] 更新 `docs/MCP_SERVER_USAGE.md` 或相关说明，明确 Web Review Export 与 MCP `export_report` 的关系。
- [ ] 更新 `docs/verification.md`，只记录真实运行过的命令和结果。

## 验收清单

- [ ] `POST /api/v2/agent-runs/:id/review-export` 存在并鉴权。
- [ ] JSON export 返回结构化 package。
- [ ] Markdown export 返回 Markdown content。
- [ ] 导出动作写入 `review_export.json` / `review_export.markdown` operation。
- [ ] 导出动作写入 `agent_run.review_exported` audit。
- [ ] Agent Run Detail 可触发 JSON / Markdown export。
- [ ] Agent Run Detail 导出后 timeline 刷新。
- [ ] Audit 页面可回查 `agent_run.review_exported`。
- [ ] E2E 不使用 skip/only。
- [ ] E2E 不只是静态文字检查。
- [ ] 无敏感信息泄露。
- [ ] 没有越界到 Scanner Hub、生命周期治理、批量操作、PDF/ZIP/SARIF、报表中心。

## 交付物

- 后端 service / API / tests。
- 前端 type / api-client / detail UI / E2E。
- 文档更新。
- 验证记录。

