# GODNSLOG 2.0 Sprint D 修正包验收结论

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

**结论：有进展，但仍不通过。**

本轮修正已经补上了两件重要事情：

- 对外响应中加入了结构化 `evidence`
- JSON 导出不再手工拼字符串，timeline 也开始按正序构建

但 Sprint D 修正包的两个核心 blocker 仍未关闭：

1. MCP 侧仍然强制 `case_id`，`payload_id` 单独生成路径没有真正打通
2. API 真实行为测试并未补齐，当前没有证据证明 `/api/v2/evidence/generate` 的核心契约已经成立

因此 Sprint D 仍不能关闭，当前不能进入 Sprint E。

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

## 本次完成点

### 1. EvidenceResponse 已显式承载结构化 Evidence

当前响应模型已经从“只有 `content` 字符串”调整为：

- [internal/interaction/evidence.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence.go:54)

已新增：

- `evidence`
- `format`
- `content`
- `metadata`

这比上一轮更接近 Sprint D 修正包要求的主契约。

### 2. Evidence 构建时开始显式按时间正序排序

当前 `GenerateEvidence()` 在构建 timeline 前增加了排序：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:56)

同时 JSON 导出改为 `json.MarshalIndent`：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:269)

这两个方向是对的。

### 3. `summarize_evidence` 已改为优先返回 `data.evidence`

MCP `summarize_evidence` 不再解析 `content` JSON 字符串，而是直接读取：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:298)

这一点比上一轮更干净。

## 仍然不通过的原因

### P1：MCP 仍然强制 `case_id`，payload-only 路径并没有真正打通

修正包明确要求：

- API 支持 `case_id` 或 `payload_id`
- MCP `summarize_evidence` / `export_report` 也必须同步支持 payload-only

但当前 MCP 仍然是：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:277)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:312)

`summarizeEvidence()` 和 `exportReport()` 都仍然把 `case_id` 作为必填。  
这意味着：

- 仅传 `payload_id` 的 MCP 路径仍然会被本地参数校验直接拒绝
- 修正包第 1 条和第 5 条未满足
- 端到端的 payload-only 证据生成仍未成立

这是当前最直接的 blocker。

### P1：API 真实行为测试并没有补齐，只留下了“未覆盖说明”

修正包明确要求 API 测试至少覆盖：

- `case_id` 生成
- `payload_id` 生成
- 空入参 bad request
- no evidence / not found
- `json` 结构化响应
- `markdown` 导出文本

但当前 `server/v2_api_test.go` 并没有新增这些真实测试，末尾只是补了一段说明注释：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:734)

这不能替代真实 API 测试。  
当前仍然没有仓库内证据证明 `/api/v2/evidence/generate` 的关键契约已被验证。

### P1：Evidence service 的 “no evidence” 测试没有测到真实目标行为

修正包要求 service 层补：

- no evidence 返回 `ErrEvidenceNotFound`

但当前新增测试并没有调用 `GenerateEvidence()`，而只是再次验证空交互下 `calculateEvidenceStrength()` 返回 `low/0`：

- [internal/interaction/evidence_service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service_test.go:261)

这没有证明：

- `GenerateEvidence()` 在真实空结果场景下会返回 `ErrEvidenceNotFound`

因此该项验收证据仍然缺失。

### P1：修正包要求的 MCP payload-only 测试也没有补齐

修正包明确要求 MCP 测试覆盖 payload-only。

当前 `internal/mcp/server_test.go` 新增的 Evidence 测试仍然只走：

- `case_id`

见：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:339)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:393)

没有看到：

- 仅 `payload_id` 的 `summarize_evidence`
- 仅 `payload_id` 的 `export_report`

因此修正包第 5 条仍未被证明成立。

### P2：EvidenceRequest 结构仍保留 `case_id required`，契约边界没有完全收口

虽然 API handler 本身已经放开 `case_id` 必填，但 `EvidenceRequest` 结构仍然保留：

- [internal/interaction/evidence.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence.go:46)

这里还是：

- `CaseID string ... binding:"required"`

这不一定直接影响当前运行路径，但说明契约定义还没有完全一致，后续继续使用该结构时会再次引入歧义。

### P2：当前新增的部分 service 测试仍是“辅助性测试”，不是主路径测试

例如 timeline 测试当前做法是：

1. 手工对切片排序
2. 再调用 `buildTimeline()`

见：

- [internal/interaction/evidence_service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service_test.go:286)

这证明了“排序后的输入能生成正序 timeline”，但没有直接证明 `GenerateEvidence()` 主路径整体保证了该语义。  
这类测试有价值，但仍弱于真实主路径测试。

## 验收判断

### 已达成

- 结构化 Evidence 主响应开始成形
- JSON 导出语义比上一轮完整
- timeline 正序处理已进入实现
- `summarize_evidence` 已不再解析 `content` JSON 字符串
- 全量测试继续通过

### 未达成

- MCP payload-only 路径未打通
- Evidence API 真实行为测试未补齐
- no evidence 的 service 主路径测试未补齐
- MCP payload-only 测试未补齐
- 契约定义仍有残留不一致

## 结论建议

Sprint D 继续停留在修正阶段，仍**不得进入 Sprint E**。

建议 Windsurf 再进行一次 **Sprint D 最后一轮收口修正**，只补以下 4 项：

1. `summarize_evidence` / `export_report` 去掉 `case_id` 强制要求，支持 payload-only
2. 补齐 `/api/v2/evidence/generate` 的真实 API 行为测试
3. 补齐 `GenerateEvidence()` 空结果返回 `ErrEvidenceNotFound` 的主路径测试
4. 收口 `EvidenceRequest` 等残留契约定义，避免再次出现 `case_id required`

这 4 项补齐后，我再做 Sprint D 关闭验收。
