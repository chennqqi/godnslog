# GODNSLOG 2.0 Sprint D Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint D
- **Sprint 主题**：Evidence 聚合、评分与导出
- **所属阶段**：Phase 2 - Probe → Interaction → Evidence → Export → Audit Closed Loop

## Sprint 目标

建立 MVP 的第三段真实业务闭环，把 “交互数据如何变成可解释、可导出的证据结果” 这一段做实。

本 Sprint 只聚焦 4 件事：

1. 基于 `Case` / `Payload` 聚合统一 `Interaction`
2. 生成 MVP 级别的 `Evidence` 评分、强度和说明
3. 输出稳定的 JSON / Markdown 证据导出
4. 让 `/api/v2/evidence` 与 MCP `summarize_evidence` / `export_report` 复用同一后端事实结果

本 Sprint 的重点不是“审计治理”或“证据页面”，而是把 **Evidence 结果模型** 做成后续 UI、CLI、审计都能复用的稳定后端出口。

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/agent-native-specification.md`
- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-c-package.md`
- `docs/superpowers/acceptance/2026-05-17-godnslog-2-sprint-c-final-acceptance.md`

## 实施范围

本 Sprint 只允许覆盖以下 4 个主题。

### 1. 建立统一 Evidence 输出模型

目标是让 Evidence 不再只是“临时拼装字符串”，而是形成统一后端结果语义。

至少覆盖：

- `id`
- `case_id`
- `payload_id`
- `interaction_count`
- `timeline`
- `confidence`
- `evidence_strength`
- `explainability`
- `generated_at`

要求：

- 后端主语义以 `docs/unified-terminology.md` 的 Evidence 定义为准
- `internal/interaction/evidence.go` 中的结构应能承载 MVP 所需字段
- 不允许继续把 Evidence 退化成“只有 content 字符串，没有结构化元数据”

### 2. 收口 MVP 评分与说明逻辑

目标是把 MVP 的证据评分逻辑做成稳定、可测、可解释的规则。

至少覆盖：

- `confidence`：0-100
- `evidence_strength`：`low | medium | high`
- `interaction_count`
- `unique_sources`
- DNS / HTTP 的基本差异化权重

要求：

- 评分逻辑与 `docs/mvp-closed-loop.md` 的 MVP 口径一致
- 算法必须是确定性的，不能依赖随机行为
- 必须输出简洁、稳定的人类可读说明，例如：
  `Captured 4 interactions from 2 unique sources. Evidence strength: medium.`

### 3. 收口 `/api/v2/evidence` 生成与导出语义

目标是让 Evidence API 真正返回统一结果，而不是临时占位。

至少覆盖：

- `POST /api/v2/evidence/generate`
- `GET /api/v2/evidence/:id` 的边界说明

要求：

- `generate` 支持 `case_id` 或 `payload_id`
- `format` 至少支持 `json`、`markdown`
- 当没有交互数据时返回稳定、标准的 not found / no evidence 语义
- `GET /api/v2/evidence/:id` 如果本 Sprint 仍不做持久化，必须保持边界清晰，不伪装成已支持读取

### 4. 对齐 MCP `summarize_evidence` 与 `export_report`

目标是让 Agent 侧证据读取与导出也复用后端真实证据结果。

至少覆盖：

- `summarize_evidence`
- `export_report`

要求：

- MCP 不得自己拼接另一套 Evidence 结构
- MCP 返回内容必须与 `/api/v2/evidence` 的后端语义一致
- `summarize_evidence` 与 `export_report` 的结果边界必须清晰：
  一个偏结构化摘要，一个偏格式化导出

## 禁止越界项

Windsurf 在本 Sprint 中不得进入以下内容：

- 不实现 Audit 事件全链路增强
- 不实现 Evidence 持久化存储设计大改
- 不开始前端 Evidence Timeline 页面开发
- 不开始 CLI 大规模扩展
- 不开始 Burp / Yakit / Nuclei 集成
- 不引入 AI 总结或 LLM explainability
- 不扩展 PDF / SARIF / CSV 之外的新导出形态
- 不开始 Agent Run / Workspace 治理扩展

如果为了完成 Sprint D 需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作：

### Evidence 模型与服务

- `internal/interaction/evidence.go`
- `internal/interaction/evidence_service.go`
- `internal/interaction/evidence_service_test.go`

### Interaction 服务复用

- `internal/interaction/service.go`

### API 入口

- `server/v2_api.go`
- `server/v2_api_test.go`

### MCP 对接

- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

### 如确有必要的旧模型桥接

- `models/v2.go`

但只允许为兼容和收口服务，不允许继续扩大旧模型职责。

## 建议实施顺序

Windsurf 应按以下顺序推进：

1. 先统一 Evidence 结构模型
2. 再收口评分、强度和 explainability 规则
3. 再接入 `/api/v2/evidence/generate`
4. 最后对齐 MCP `summarize_evidence` / `export_report`

## 必须补齐的测试

### 1. Evidence 评分测试

至少覆盖：

- 无交互时返回 no evidence / not found
- 仅少量交互时得到 `low`
- 多交互且多源时得到 `medium`
- 更高数量与更多源时得到 `high`
- `confidence` 稳定落在 `0-100`

### 2. Evidence 说明与时间线测试

至少覆盖：

- 返回的 `timeline` 为按时间排序的交互序列
- `interaction_count` 正确
- `unique_sources` 统计正确
- `explainability` 文本包含数量、来源数和强度

### 3. Evidence 导出测试

至少覆盖：

- `json` 导出包含结构化字段
- `markdown` 导出包含摘要与交互明细
- `csv` 如果继续保留，必须明确是内部兼容输出还是正式支持；若不是正式支持，不得混入对外主契约

### 4. API / MCP 对齐测试

至少覆盖：

- `/api/v2/evidence/generate` 支持 `case_id`
- `/api/v2/evidence/generate` 支持 `payload_id`
- `summarize_evidence` 结果与 API 摘要字段一致
- `export_report` 结果与 API 导出内容一致

## 完成定义

只有同时满足以下条件，Sprint D 才能视为完成：

1. Evidence 已形成稳定的结构化输出模型
2. 评分、强度和 explainability 逻辑稳定可测
3. `/api/v2/evidence/generate` 可基于 case 或 payload 输出 JSON / Markdown
4. MCP `summarize_evidence` / `export_report` 复用统一后端语义
5. 相关测试通过，且 `GOCACHE=/tmp/gocache go test ./...` 继续通过

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式向 Codex 回传：

### 1. 实际修改范围

- 修改了哪些目录
- 修改了哪些关键文件
- 哪些计划内文件未动，原因是什么

### 2. 实际实现内容

- Evidence 主模型如何统一
- 评分 / 强度 / explainability 如何定义
- `/api/v2/evidence/generate` 如何输出统一结果
- MCP `summarize_evidence` / `export_report` 如何与 API 对齐

### 3. 实际验证命令

至少包含实际执行过的命令，例如：

- `GOCACHE=/tmp/gocache go test ./internal/interaction`
- `GOCACHE=/tmp/gocache go test ./internal/mcp ./server`
- `GOCACHE=/tmp/gocache go test ./...`

### 4. 测试结果

- 新增了哪些测试
- 修改了哪些测试
- 哪些预期测试仍未覆盖

### 5. 风险与偏差

- 哪些地方与原规划不完全一致
- 哪些点被明确留给 Sprint E
- 哪些旧模型桥接仍然存在

## Codex 验收问题

Codex 在验收 Sprint D 时只围绕以下问题判断：

1. Evidence 是否真正形成统一结构化结果，而不是临时字符串
2. 评分、强度和 explainability 是否稳定且符合 MVP 口径
3. `/api/v2/evidence/generate` 是否真正基于统一 Interaction 输入工作
4. MCP 是否复用后端真实 Evidence 契约
5. 当前输出是否足够给 Sprint E 的 Web / Audit 消费
6. 是否严格没有越界到前端 / Audit 治理 / 工具集成

## 验收结论类型

Codex 对 Sprint D 的验收只会给出三种结果：

- **通过**：可进入 Sprint E
- **有条件通过**：允许进入 Sprint E，但必须挂明遗留项
- **不通过**：必须继续停留在 Sprint D 修正

## Sprint D 完成后的下一步

只有 Sprint D 被 Codex 验收通过后，才进入 `Sprint E：Evidence Web 展示与 Audit 收口`。
