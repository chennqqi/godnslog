# GODNSLOG 2.0 Sprint D 修正包二次验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-21-godnslog-2-sprint-d-remediation-package.md`
- `internal/interaction/evidence.go`
- `internal/interaction/evidence_service.go`
- `internal/interaction/evidence_service_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `server/v2_api.go`
- `server/v2_api_test.go`

## 验收结论

**结论：仍不通过。**

相比上一轮，这次已经补上了一个关键缺口：

- MCP `summarize_evidence` / `export_report` 现在都支持 `case_id` 或 `payload_id`

但 Sprint D 修正包要求的不只是“代码看起来已支持”，还要求有足够的主路径验证证据。当前仍缺两类关键证明：

1. `/api/v2/evidence/generate` 的真实 API 行为测试
2. `GenerateEvidence()` 空结果返回 `ErrEvidenceNotFound` 的主路径测试

因此 Sprint D 仍不能关闭，当前仍不能进入 Sprint E。

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

## 本次新增完成点

### 1. MCP payload-only 路径已在代码层放开

当前 `summarizeEvidence()` 与 `exportReport()` 都不再强制 `case_id`：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:276)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:316)

当前口径已变为：

- `case_id`
- `payload_id`

二者至少一个必填。

这比昨天更接近修正包要求。

### 2. `EvidenceRequest` 的残留 `case_id required` 已去掉

当前 `EvidenceRequest` 已改为：

- [internal/interaction/evidence.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence.go:47)

这里不再保留 `case_id` 的 `binding:"required"`，结构定义与 handler 契约更一致。

## 仍然不通过的原因

### P1：Evidence API 真实行为测试仍然没有补齐

修正包要求 API 测试至少覆盖：

- `case_id` 生成
- `payload_id` 生成
- 空入参 bad request
- no evidence / not found
- `json` 结构化响应
- `markdown` 导出文本

但当前 `server/v2_api_test.go` 仍然没有新增这些测试。文件末尾依然只是说明注释：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:734)

这意味着当前仓库里仍然没有证据证明：

- `/api/v2/evidence/generate` 的 handler 契约已经被真实验证

这是本轮仍然不能放行的主因。

### P1：`GenerateEvidence()` 的 no-evidence 主路径测试仍然缺失

上一轮的问题是：

- 名为 `TestGenerateEvidence_NoInteractions`
- 实际没有调用 `GenerateEvidence()`

这一轮该测试被删除了，但没有换成真正的主路径测试。

当前 `internal/interaction/evidence_service_test.go` 里仍然看不到：

- 调用 `GenerateEvidence(...)`
- 断言返回 `ErrEvidenceNotFound`

因此修正包要求的这条仍未满足。

### P2：MCP payload-only 代码已放开，但测试仍没有证明 payload-only 主路径成立

当前 `internal/mcp/server_test.go` 的 Evidence 相关测试仍然主要使用：

- `case_id`

见：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:339)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:393)

没有看到：

- 仅 `payload_id` 的 `summarize_evidence`
- 仅 `payload_id` 的 `export_report`

这项现在已经从“代码 blocker”下降为“验证 blocker”，但仍然没有形成验收证据。

## 验收判断

### 已达成

- 结构化 Evidence 主响应已建立
- JSON 导出已改为结构化序列化
- timeline 正序处理已进入实现
- MCP 不再强制 `case_id`
- `EvidenceRequest` 契约残留已收口
- 全量测试继续通过

### 未达成

- Evidence API 真实行为测试未补齐
- `GenerateEvidence()` 空结果主路径测试未补齐
- MCP payload-only 主路径测试未补齐

## 结论建议

Sprint D 继续停留在修正阶段，仍**不得进入 Sprint E**。

建议 Windsurf 最后一轮只补 3 项，不再扩散范围：

1. 在 `server/v2_api_test.go` 补齐 `/api/v2/evidence/generate` 的真实 handler 测试
2. 在 `internal/interaction/evidence_service_test.go` 补齐 `GenerateEvidence()` 空结果返回 `ErrEvidenceNotFound` 的主路径测试
3. 在 `internal/mcp/server_test.go` 补齐 payload-only 的 `summarize_evidence` / `export_report` 测试

这 3 项补齐后，我再做 Sprint D 关闭验收。
