# GODNSLOG 2.0 Sprint D 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-d-package.md`
- `internal/interaction/evidence.go`
- `internal/interaction/evidence_service.go`
- `internal/interaction/evidence_service_test.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

## 验收结论

**结论：不通过。**

Sprint D 已经把 Evidence 主模型和基础评分逻辑搭起来了，全量测试也继续通过；但当前实现还没有达到 Sprint D 的完成定义，尤其是 `payload_id` 路径、结构化 Evidence 契约、时间线/评分语义、以及 API/MCP 对齐证明这四个点，仍存在契约级缺口。

Sprint D 不能关闭，当前不能进入 Sprint E。

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

## 本次已完成部分

### 1. Evidence 主模型已显式落地

当前 `Evidence` 已经承载 Sprint D 所需的大部分核心字段：

- [internal/interaction/evidence.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence.go:12)

已包含：

- `id`
- `case_id`
- `payload_id`
- `interaction_count`
- `unique_sources`
- `timeline`
- `confidence`
- `evidence_strength`
- `explainability`
- `generated_at`

### 2. MVP 基础评分规则已初步实现

当前强度与置信度规则已经存在，并且是确定性的：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:94)

`low / medium / high` 的基础门槛与 `confidence 0-100` 口径已经具备雏形。

### 3. `/api/v2/evidence/generate` 与 MCP 已建立调用链

Evidence API 入口已接到真实 `EvidenceService`：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:2155)

MCP 的 `summarize_evidence` / `export_report` 也已复用该 API：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:275)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:316)

## 不通过原因

### P1：`/api/v2/evidence/generate` 仍不支持 payload-only 生成

Sprint D 明确要求 `generate` 支持 `case_id` 或 `payload_id`。

但当前 API 绑定仍然把 `case_id` 设为必填：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:2157)

这意味着：

- 仅传 `payload_id` 会在 API 层直接被拒绝
- Sprint D 完成定义第 3 条未满足
- MCP 也无法真正走 payload-only 路径

这是一票否决项。

### P1：对外主契约仍然是 `content string`，不是统一结构化 Evidence 结果

虽然内部有 `Evidence` 结构，但对外响应仍是：

- [internal/interaction/evidence.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence.go:52)

当前 `EvidenceResponse` 只有：

- `format`
- `content`
- `metadata`

并没有把结构化 `Evidence` 作为主语义直接暴露出来。

进一步看，`GenerateEvidence()` 返回的也是导出包装，而不是统一结构化对象：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:29)

这与 Sprint D 的核心要求“Evidence 不再退化成临时字符串”仍有差距。

### P1：JSON 导出没有输出完整 Evidence 语义，`timeline` 等关键字段缺失

当前 JSON 导出是手工拼接字符串：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:247)

导出内容里只有：

- `id`
- `case_id`
- `payload_id`
- `evidence_strength`
- `confidence`
- `interaction_count`
- `unique_sources`
- `explainability`
- `generated_at`

但缺少 Sprint D 明确要求的：

- `timeline`

也没有把 `interactions` 或其他结构化上下文一并稳定暴露。

这说明当前 JSON 导出仍偏“摘要字符串”，还不足以作为后续 Web / Audit 的统一消费出口。

### P1：时间线语义与 MVP 文档不一致，当前实现不是按时间正序构建

`buildTimeline()` 本身不排序，只按输入顺序追加：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:220)

而 `GenerateEvidence()` 读取交互时调用 `ListInteractions()`，该方法默认按 `timestamp DESC` 返回：

- [internal/interaction/service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/service.go:83)
- [internal/interaction/service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/service.go:110)

这和 `docs/mvp-closed-loop.md` 中 “Timeline: chronological list of interactions with timestamps” 不一致。当前更接近“倒序列表”，不是“按时间线形成证据”。

### P1：评分逻辑没有落实 DNS / HTTP 差异化权重

Sprint D 实施包要求“DNS / HTTP 的基本差异化权重”。

但当前实际用于 Evidence 的 `calculateEvidenceStrength()` 只使用：

- 交互数量
- 来源数量

见：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:100)

协议差异化权重只存在于已废弃的 `calculateScore()`：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:165)

也就是说，DNS / HTTP 在当前主评分链路中没有任何差异化影响，未满足 Sprint D 范围要求。

### P1：MCP 仍然通过解析 `content` JSON 字符串获取结构化数据

`summarize_evidence` 当前不是直接消费后端结构化 Evidence，而是：

1. 调 `/api/v2/evidence/generate`
2. 读取 `data.content`
3. 再把这个 JSON 字符串 `Unmarshal`

见：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:287)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:298)

这说明 MCP 只是“复用导出字符串”，还不是“复用统一后端 Evidence 契约”。

### P1：API 层缺少 Evidence 真实行为测试，无法证明核心契约已完成

当前 `server/v2_api_test.go` 对 Evidence 的覆盖，只能看到路由存在性：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:21)

但没有看到以下必须的真实 API 行为测试：

- `case_id` 生成
- `payload_id` 生成
- `json` 导出结构
- `markdown` 导出内容
- no evidence / not found

这意味着即使包测试全绿，也无法证明 Sprint D 的 API 契约已经被真实验证。

### P2：Evidence 服务测试覆盖仍明显不足

当前 `internal/interaction/evidence_service_test.go` 主要覆盖：

- 强度 low / medium / high
- `confidence` 范围
- `unique_sources`
- `explainability` 非空

见：

- [internal/interaction/evidence_service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service_test.go:46)
- [internal/interaction/evidence_service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service_test.go:226)

但仍未覆盖 Sprint D 包要求的关键路径：

- no evidence / not found
- timeline 排序
- JSON 导出结构
- Markdown 导出明细
- `payload_id` 生成

### P2：`csv` 仍保留在服务层，但边界没有说明清楚

当前 `EvidenceService` 仍然支持 `csv`：

- [internal/interaction/evidence_service.go](/data/dev/github.com/chennqqi/godnslog/internal/interaction/evidence_service.go:72)

而 API 层只允许：

- `json`
- `markdown`

见：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:2160)

这不一定是 blocker，但当前没有明确说明 `csv` 只是内部兼容还是未来正式能力，边界仍不够清晰。

## 对 6 个验收问题的判断

1. Evidence 是否真正形成统一结构化结果，而不是临时字符串：**否**
2. 评分、强度和 explainability 是否稳定且符合 MVP 口径：**部分是，但未完全达标**
3. `/api/v2/evidence/generate` 是否真正基于统一 Interaction 输入工作：**部分是，但 payload-only 不成立**
4. MCP 是否复用后端真实 Evidence 契约：**否，当前仍在解析 content 字符串**
5. 当前输出是否足够给 Sprint E 的 Web / Audit 消费：**否**
6. 是否严格没有越界到前端 / Audit 治理 / 工具集成：**是**

## 结论建议

Sprint D 继续停留在当前阶段，不进入 Sprint E。

下一步不应直接写 `Sprint E` 包，而应先下发一份 **Sprint D 修正包**，只收口以下缺口：

1. `case_id` / `payload_id` 二选一生成契约
2. 统一结构化 Evidence 主响应
3. timeline 正序与 JSON / Markdown 导出语义
4. DNS / HTTP 差异化评分落地
5. API / MCP 对齐测试补齐
