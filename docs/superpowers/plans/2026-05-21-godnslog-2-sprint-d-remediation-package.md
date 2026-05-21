# GODNSLOG 2.0 Sprint D Remediation Package

> **协作模式**
>
> - **Codex**：负责修正包规划与验收
> - **Windsurf**：负责按修正包补齐 Sprint D 缺口

## 修正目标

本修正包不是新 Sprint，而是 **Sprint D 补验收包**。目标只有一个：把 Evidence 的对外契约、评分语义、API/MCP 对齐和验证证据补齐，使 Sprint D 达到可关闭状态。

## 仅修正以下缺口

### 1. `/api/v2/evidence/generate` 必须真正支持 `case_id` 或 `payload_id`

Windsurf 必须补齐：

- API 入参校验从“`case_id` 必填”改为“`case_id` 或 `payload_id` 至少一个”
- `payload_id` 单独生成 Evidence 的真实成功路径
- `case_id + payload_id` 同时提供时的边界说明
- no evidence / invalid input 的稳定响应语义

要求：

- 不能只在 service 层支持，必须从 API 层到测试层都成立
- MCP `summarize_evidence` / `export_report` 也必须同步支持 payload-only 路径，不能继续强制 `case_id`

### 2. 统一结构化 Evidence 主响应，不能继续以 `content string` 充当主语义

Windsurf 必须补齐：

- 对外主语义必须显式暴露结构化 `Evidence`
- `EvidenceResponse` 需要能同时承载：
  - 结构化 Evidence 结果
  - 导出格式信息
  - 导出文本内容（仅在需要导出时）
- `json` 路径不能再依赖手工字符串拼接

建议口径：

- `data.evidence`：统一结构化结果
- `data.format`：`json | markdown`
- `data.content`：导出内容；`json` 可为结构化 JSON 字符串，`markdown` 为 Markdown 文本
- `data.metadata`：仅保留辅助信息，不再承担主语义

要求：

- Web / MCP / 后续 Audit 应消费 `data.evidence`
- 导出只是结构化结果的一种投影，不应反过来主导后端语义

### 3. 收口 Evidence 语义完整性：timeline、排序、导出字段

Windsurf 必须补齐：

- `timeline` 必须进入结构化 Evidence 对外结果
- timeline 必须是时间正序（chronological），不能沿用倒序列表
- `json` 导出至少包含：
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
- `markdown` 导出必须同时包含摘要和交互明细

要求：

- 不允许继续使用手工字符串拼接 JSON
- 若需要排序，应在 Evidence 构建阶段显式排序，而不是依赖查询默认顺序

### 4. 评分逻辑必须纳入 DNS / HTTP 的基本差异化权重

Windsurf 必须补齐：

- 把 DNS / HTTP 差异化权重放进当前主评分链路
- 保持评分确定性与可解释性
- 明确“数量 + 来源数 + 协议差异”三者如何共同影响 `confidence` 或 `evidence_strength`

要求：

- 不做复杂算法
- 不引入 AI / LLM
- 不引入随机因素

建议最小口径：

- `evidence_strength` 仍以数量和来源数作为主门槛
- `confidence` 在此基础上，再对 HTTP 给予有限加权
- `explainability` 要能解释评分依据，而不只是输出结果

### 5. MCP 必须直接复用统一后端 Evidence 契约

Windsurf 必须补齐：

- `summarize_evidence` 直接返回 API 的结构化 Evidence 结果
- `export_report` 直接返回 API 的导出内容
- MCP 不得再通过解析 `data.content` 的 JSON 字符串来“二次拼结构”

要求：

- `summarize_evidence` 的结构化字段与 API `data.evidence` 一致
- `export_report` 的文本内容与 API `data.content` 一致
- MCP 的 payload-only 路径同样要有测试

### 6. 证据测试必须从“局部函数测试”补到“真实契约测试”

本修正包必须新增或补齐以下测试。

#### Evidence Service 测试

- no evidence 返回 `ErrEvidenceNotFound`
- timeline 为时间正序
- JSON 导出包含结构化字段，尤其是 `timeline`
- Markdown 导出包含摘要与交互明细
- DNS / HTTP 差异化权重对评分有可测影响

#### API 测试

- `/api/v2/evidence/generate` 支持 `case_id`
- `/api/v2/evidence/generate` 支持 `payload_id`
- 仅传空入参返回稳定 bad request
- 无交互时返回稳定 not found / no evidence
- `json` 响应包含结构化 `evidence`
- `markdown` 响应包含导出文本

#### MCP 测试

- `summarize_evidence` 与 API `data.evidence` 字段一致
- `export_report` 与 API `data.content` 一致
- `payload_id` 路径可工作

## 禁止越界项

本修正包仍然禁止进入：

- Sprint E 的前端 Evidence 页面开发
- Audit 链路增强
- CLI 扩展
- Burp / Yakit / Nuclei / ZAP / xray 集成
- Evidence 持久化存储设计大改
- Agent Run / Workspace 治理扩展

如果为了完成修正包需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作：

- `internal/interaction/evidence.go`
- `internal/interaction/evidence_service.go`
- `internal/interaction/evidence_service_test.go`
- `internal/interaction/service.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

如确有必要，可补充：

- `internal/models/`

但仅限于 Evidence 契约承载，不允许借机扩大领域模型范围。

## 完成定义

只有同时满足以下条件，修正包才算完成：

1. `/api/v2/evidence/generate` 可基于 `case_id` 或 `payload_id` 成功生成
2. 对外主响应已显式暴露结构化 `Evidence`，而不是只有 `content` 字符串
3. `timeline` 已进入统一 Evidence 结果，且为时间正序
4. DNS / HTTP 差异化权重已进入主评分链路
5. MCP `summarize_evidence` / `export_report` 已直接复用统一后端 Evidence 契约
6. API / Service / MCP 相关测试补齐
7. `GOCACHE=/tmp/gocache go test ./...` 继续通过

## Windsurf 回传要求

Windsurf 必须特别说明：

- Evidence 最终对外响应结构是什么
- `case_id` / `payload_id` 双入口如何校验
- timeline 如何保证正序
- DNS / HTTP 差异化权重具体如何落地
- `summarize_evidence` 如何不再解析 `content` 字符串
- 新增了哪些 API / MCP / service 测试

## 验收结果

本修正包通过后，Codex 才会重新判定 Sprint D 是否关闭；只有 Sprint D 被重新验收通过，才进入 `Sprint E：Evidence Web 展示与 Audit 收口`。
