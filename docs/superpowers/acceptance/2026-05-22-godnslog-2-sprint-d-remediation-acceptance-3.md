# GODNSLOG 2.0 Sprint D 修正包三次验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-21-godnslog-2-sprint-d-remediation-package.md`
- `internal/interaction/evidence.go`
- `internal/interaction/evidence_service_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `server/v2_api_test.go`

## 验收结论

**结论：仍不通过。**

这一次又补上了一个缺口：

- MCP payload-only 的测试已经进入仓库

但 Sprint D 修正包要求的核心验收证据仍然缺两类：

1. `/api/v2/evidence/generate` 的真实主路径 API 测试仍然不完整
2. `GenerateEvidence()` 空结果返回 `ErrEvidenceNotFound` 的主路径测试仍然没有

因此 Sprint D 仍不能关闭。

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

## 本次新增完成点

### 1. MCP payload-only 测试已补上

当前已新增：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:387)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:432)

分别覆盖：

- `summarize_evidence` 仅传 `payload_id`
- `export_report` 仅传 `payload_id`

这项现在可以视为通过。

### 2. MCP 空入参与错误边界测试已补上

当前还新增了：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:475)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:506)

这使 MCP 的参数边界比上一轮更完整。

## 仍然不通过的原因

### P1：Evidence API 真实行为测试仍然没有覆盖核心成功路径

当前 `server/v2_api_test.go` 新增的 Evidence 测试主要覆盖了：

- 空参数 bad request
- 非法 format bad request

见：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:859)
- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:968)

但修正包要求的真实主路径测试至少还缺：

- `case_id` 成功生成
- `payload_id` 成功生成
- `json` 成功响应且包含 `data.evidence`
- `markdown` 成功响应且包含 `data.content`
- no evidence / not found

也就是说，当前 API 侧只证明了“拦截错误参数”，还没有证明“正确输入时契约成立”。

### P1：`GenerateEvidence()` 空结果返回 `ErrEvidenceNotFound` 的主路径测试仍然缺失

当前 `internal/interaction/evidence_service_test.go` 里仍然没有看到：

- 调用 `GenerateEvidence(...)`
- 断言错误为 `ErrEvidenceNotFound`

现有测试仍主要是：

- timeline 排序
- JSON 导出结构
- Markdown 导出结构
- HTTP 加权

见：

- [internal/interaction/evidence_service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service_test.go:261)

这说明修正包要求的 no-evidence 主路径证明仍然不存在。

## 验收判断

### 已达成

- 结构化 Evidence 主响应已建立
- JSON 导出与 timeline 正序已收口
- MCP `case_id` / `payload_id` 双入口代码已打通
- MCP payload-only 测试已补上
- 全量测试继续通过

### 未达成

- Evidence API 成功路径真实测试未补齐
- Evidence API no-evidence 真实测试未补齐
- `GenerateEvidence()` 的 `ErrEvidenceNotFound` 主路径测试未补齐

## 结论建议

Sprint D 继续停留在修正阶段，仍**不得进入 Sprint E**。

Windsurf 下一轮只需要补 2 组测试，不需要再改业务设计：

1. `server/v2_api_test.go`
   - 补 `case_id` 成功
   - 补 `payload_id` 成功
   - 补 `json` / `markdown` 成功响应
   - 补 no-evidence 404
2. `internal/interaction/evidence_service_test.go`
   - 补 `GenerateEvidence()` 空结果返回 `ErrEvidenceNotFound`

这两组补完后，我再做 Sprint D 最终关闭验收。
