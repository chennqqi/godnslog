# GODNSLOG 2.0 Sprint S Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use `superpowers:subagent-driven-development` or `superpowers:executing-plans` to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a deterministic integrity manifest to single Agent Run Review Evidence export and delivery flows, so an operator can prove that the package exported, delivered to webhook, recorded in Delivery History, and referenced in Audit is the same sanitized evidence package.

**Architecture:** Reuse Sprint P Review Evidence Export Package, Sprint Q Delivery Receipt, and Sprint R Delivery History. Add a small deterministic package manifest and SHA-256 content hash to export responses, delivery webhook payload refs, Agent Operations, Audit details, and Delivery History display. Do not introduce report storage, signed attestations, PKI, PDF/DOCX/ZIP, saved connectors, notification center, retry queue, lifecycle governance, Scanner Hub, workflow engine, or MCP auto-delivery.

**Tech Stack:** Go, xorm, Gin routes in `server/v2_api.go`, `internal/agentrun`, `internal/auth`, standard `crypto/sha256`, deterministic JSON encoding helper, Next.js/TypeScript, shadcn/ui, Playwright.

---

## Sprint 标识

- **Sprint 名称**：Sprint S
- **Sprint 主题**：Review Evidence Package Integrity Manifest
- **所属阶段**：Phase 5 - Agent Governance and Review Operations

## Sprint 背景

Sprint P 完成单个 Agent Run Review Evidence Export Package：

```text
Agent Run Detail -> Export JSON / Markdown -> Operation -> Audit
```

Sprint Q 完成单次 webhook delivery：

```text
Agent Run Detail -> Deliver to Webhook -> Delivery Receipt -> Operation -> Audit
```

Sprint R 完成 delivery history：

```text
Delivery -> History refresh -> Receipt detail -> Audit
```

现在 operator 已能导出、交付和回查历史。但复盘时仍缺一个轻量的完整性证据：如何确认 webhook 收到的 package 与 Export dialog 中展示的 package、Delivery History 中记录的 receipt、Audit 中的交付记录是同一份内容？

Sprint S 的目标是补齐“同一份 evidence package”的可验证标识，不做签名系统或报告归档。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 为 Review Evidence Export Package 生成 deterministic package manifest。
2. Export response、export operation、export audit 中记录 `package_hash` 和 manifest refs。
3. Delivery webhook payload、delivery operation、delivery audit、delivery history 中复用同一个 `package_hash`。
4. Agent Run Detail 在 Export Result、Delivery Receipt、Delivery History 中展示短 hash，并支持查看完整 hash。
5. E2E 证明：Export hash -> Delivery payload hash -> Delivery History hash -> Audit hash 一致，且不泄露敏感内容。

## 明确不做

本 Sprint 严格不做：

- 数字签名、私钥管理、PKI、JWS、证书链。
- Report center、报告实体、报告版本、长期归档。
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
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-p-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-p-acceptance.md`
- `docs/superpowers/plans/2026-06-07-godnslog-2-sprint-q-package.md`
- `docs/superpowers/acceptance/2026-06-07-godnslog-2-sprint-q-acceptance.md`
- `docs/superpowers/plans/2026-06-08-godnslog-2-sprint-r-package.md`
- `docs/superpowers/acceptance/2026-06-08-godnslog-2-sprint-r-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已有基础

- Sprint P `ExportReviewPackage` 已能生成 JSON / Markdown package。
- Export action 已写 `review_export.<format>` operation 和 `agent_run.review_exported` audit。
- Sprint Q `DeliverReviewPackage` 复用 export package 发送 webhook。
- Delivery action 已写 `review_delivery.webhook` operation 和 `agent_run.review_delivered` / `agent_run.review_delivery_failed` audit。
- Sprint R `ListReviewDeliveries` 已从 operation / audit 派生 delivery history。
- Agent Run Detail 已有 Export Result、Delivery Receipt、Delivery History、Audit navigation。

### 主要缺口

- Export response 没有稳定 package hash。
- Webhook payload refs 没有 package hash。
- Delivery History 无法证明某次 delivery 对应哪份 package 内容。
- Audit 只能证明动作发生，不能证明 export / delivery / history 引用同一 package。
- E2E 无法比较 Export、Delivery payload、History、Audit 的同一 hash。

## 术语边界

### Package Manifest

Package Manifest 是 Review Evidence Package 的轻量完整性摘要，不是新实体表。它由 package 内容和关键 refs 派生。

### Package Hash

Package Hash 是 canonical package bytes 的 SHA-256 hex string。

```text
package_hash = sha256(canonical_package_bytes)
```

### Out of Scope: Signature / Attestation

Sprint S 只生成 hash 和 manifest，不提供加密签名、外部公证、不可抵赖证明或证书验证。

## 数据契约

### AgentRunReviewPackageManifest

建议新增到 `internal/models/agent_run.go`：

```go
type AgentRunReviewPackageManifest struct {
	SchemaVersion string    `json:"schema_version"`
	AgentRunID    string    `json:"agent_run_id"`
	ReviewPacketID string   `json:"review_packet_id,omitempty"`
	Format        string    `json:"format"`
	PackageHash   string    `json:"package_hash"`
	HashAlgorithm string    `json:"hash_algorithm"`
	GeneratedAt   time.Time `json:"generated_at"`
	Refs          map[string]string `json:"refs,omitempty"`
}
```

Required values:

```text
schema_version = "review-package-manifest/v1"
hash_algorithm = "sha256"
```

### Export Response Extension

Extend `AgentRunReviewExportResponse`:

```go
Manifest    *AgentRunReviewPackageManifest `json:"manifest,omitempty"`
PackageHash string                         `json:"package_hash,omitempty"`
```

### Delivery Response Extension

Extend `AgentRunReviewDeliveryResponse`:

```go
PackageHash string `json:"package_hash,omitempty"`
```

### Delivery History Extension

Extend `AgentRunReviewDeliveryHistoryItem`:

```go
PackageHash string `json:"package_hash,omitempty"`
```

## Hashing Rules

### JSON format

For `format=json`, hash the deterministic JSON representation of `exportResp.Package`.

Requirements:

- Stable key ordering.
- No timestamp fields that make the hash nondeterministic unless they are part of the package contract and also stored consistently.
- Do not include webhook URL, request headers, delivery destination, status code, or audit-only metadata in package hash input.

Implementation guidance:

- Prefer a helper that normalizes maps recursively into stable JSON bytes.
- Go `encoding/json` sorts map keys for string-keyed maps, but tests should still assert determinism with differently ordered maps.

### Markdown format

For `format=markdown`, hash the exact markdown content string bytes.

### What Not To Hash

Do not hash or include:

- Full webhook URL.
- Header values.
- Authorization / Cookie values.
- API keys.
- Response body from webhook.
- Runtime-only delivery status.

## Backend 实施要求

### Manifest Helper

Add helper functions in `internal/agentrun`:

```go
func BuildReviewPackageManifest(agentRunID string, format string, reviewPacketID string, content string, pkg map[string]interface{}, refs map[string]string) (*models.AgentRunReviewPackageManifest, error)
func ComputeReviewPackageHash(format string, content string, pkg map[string]interface{}) (string, error)
```

Tests must cover:

- JSON hash deterministic regardless of map insertion order.
- Markdown hash deterministic for identical content.
- JSON and Markdown for same logical data produce different hashes.
- Hash input excludes webhook URL and headers.

### Export Flow

In `ExportReviewPackage`:

- Build manifest after package content is generated.
- Include `package_hash` and `manifest` in response.
- Add `package_hash` to `review_export.<format>` operation result.
- Add `package_hash` to `agent_run.review_exported` audit details.
- Do not include full package content in audit details.

### Delivery Flow

In `DeliverReviewPackage`:

- Reuse `ExportReviewPackage` response manifest/hash.
- Add `package_hash` to webhook payload:

```json
"refs": {
  "export_operation_id": "op-export-1",
  "delivery_operation_id": "op-delivery-1",
  "audit_ref_id": "audit-delivery-1",
  "package_hash": "sha256..."
}
```

- Add `manifest` or minimal manifest refs to webhook payload without sensitive data.
- Add `package_hash` to delivery operation result.
- Add `package_hash` to delivery audit details.
- Add `package_hash` to delivery response.
- Failure audit should include package hash if export succeeded before delivery failure.

### Delivery History

In `ListReviewDeliveries`:

- Include `package_hash` from operation result or audit details.
- Summary counts unchanged.
- Do not derive package hash from webhook payload or response body.

### API Tests

Add Go tests covering:

- Export JSON response includes 64-char SHA-256 hash and manifest.
- Export Markdown response includes stable hash and manifest.
- Export operation/audit include `package_hash`.
- Delivery webhook payload includes the same package hash as export response.
- Delivery response / operation / audit / history all expose the same `package_hash`.
- Failure delivery still records package hash.
- No full webhook URL or header values appear in manifest, operation, audit, or history.

## Frontend 实施要求

### Types / API Client

Update `frontend-next/src/types/index.ts`:

- Add `AgentRunReviewPackageManifest`.
- Add `package_hash` / `manifest` fields to export and delivery response types.
- Add `package_hash` to delivery history item type.

### Agent Run Detail UI

In `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`:

- Export Result:
  - Show `Package Hash` as shortened hash, e.g. first 12 + last 8 chars.
  - Provide copy full hash button.
  - Show manifest schema version and hash algorithm.
- Delivery Receipt:
  - Show same package hash.
  - Do not show full webhook URL or header values.
- Delivery History:
  - Show package hash for each delivery attempt.
  - The hash display should be compact and copyable.
- Avoid adding report-like layout or marketing sections.

## E2E 范围

Update `frontend-next/e2e/agent-runs.spec.ts`:

### Export hash loop

1. Export JSON.
2. Assert response includes `package_hash`.
3. Assert Export dialog shows shortened hash and manifest metadata.
4. Navigate to Audit and assert `agent_run.review_exported` includes or displays package hash reference.

### Delivery hash loop

1. Initial Delivery History is empty.
2. Submit Deliver to Webhook.
3. Intercept delivery request and response.
4. Assert webhook-delivery API response includes `package_hash`.
5. Assert Delivery Receipt shows package hash.
6. Refresh Delivery History and assert same package hash appears.
7. Navigate to Audit and assert `agent_run.review_delivered`.
8. Assert full webhook URL and header values are not visible.

### Failure hash loop

1. Mock delivery failure after export.
2. Assert failure receipt / history records the same `package_hash`.
3. Assert no retry button appears.

E2E must not use `test.skip`.

## Documentation

Update:

- `docs/verification.md`
  - Add Sprint S commands and real results.
- `docs/agent-native-specification.md`
  - Under Agent Run Export / Delivery, document package hash and manifest as integrity metadata.
- `docs/MCP_SERVER_USAGE.md`
  - Clarify MCP `export_report` remains read-only and may expose package hash only for review packets, not delivery automation.

## 验收标准

- Export JSON / Markdown responses include deterministic `package_hash` and manifest.
- Export operation and audit include `package_hash`.
- Delivery webhook payload, delivery response, delivery operation, delivery audit, and delivery history all carry the same `package_hash`.
- Failure delivery keeps package hash if export was generated.
- UI shows compact package hash in Export Result, Delivery Receipt, and Delivery History.
- E2E proves Export hash -> Delivery hash -> History hash -> Audit loop.
- No full webhook URL, header values, API keys, cookies, authorization tokens, or webhook response bodies appear in manifest, operation, audit, or UI.
- No scope creep into signatures/PKI, report center, saved connectors, notification center, retry queue, batch operations, lifecycle governance, Scanner Hub, workflow engine, or MCP auto-delivery.

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
- [ ] Add backend manifest model and hash fields.
- [ ] Add deterministic hash helper and unit tests.
- [ ] Extend export response / operation / audit with `package_hash`.
- [ ] Extend delivery webhook payload / response / operation / audit with `package_hash`.
- [ ] Extend delivery history item with `package_hash`.
- [ ] Add Go tests for export/delivery/history hash consistency and sanitization.
- [ ] Update frontend types and Agent Run Detail UI hash display.
- [ ] Add E2E export hash loop.
- [ ] Add E2E delivery hash loop.
- [ ] Add E2E failure hash loop.
- [ ] Update docs.
- [ ] Run required verification commands and record results.
