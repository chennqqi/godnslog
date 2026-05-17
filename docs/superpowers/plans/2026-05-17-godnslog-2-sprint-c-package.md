# GODNSLOG 2.0 Sprint C Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint C
- **Sprint 主题**：DNS/HTTP Interaction 捕获与归因
- **所属阶段**：Phase 2 - Probe → Interaction → Evidence → Export → Audit Closed Loop

## Sprint 目标

建立 MVP 的第二段真实业务闭环，把 “Probe 被投递后，系统能否稳定看到并归因回 Payload / Case” 这一段做实。

本 Sprint 只聚焦 4 件事：

1. DNS / HTTP 交互能进入统一 `Interaction` 模型
2. 交互能按 `token -> payload -> case` 稳定归因
3. `/api/v2/interactions` 与 `wait_for_interaction` 能基于统一存储工作
4. 为 Sprint D 的 Evidence 聚合提供稳定输入

本 Sprint 的重点不是“证据解释”或“结果导出”，而是把 **交互捕获与归因链路** 做成可依赖的后端事实层。

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/agent-native-specification.md`
- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-b-package.md`
- `docs/superpowers/acceptance/2026-05-17-godnslog-2-sprint-b-final-acceptance.md`

## 实施范围

本 Sprint 只允许覆盖以下 4 个主题。

### 1. 统一 DNS / HTTP Interaction 写入模型

目标是让 DNS 与 HTTP 捕获都进入统一 `internal/models.Interaction` 存储，不再把 `/api/v2/interactions` 继续建立在旧 `TblInteraction` 读写语义上。

至少覆盖：

- `id`
- `type`
- `token`
- `payload_id`
- `case_id`
- `timestamp`
- `source_ip`
- DNS 的 `domain` / `dns_type`
- HTTP 的 `method` / `path` / `headers` / `body`

要求：

- 统一写入主模型以 `internal/models.Interaction` 为准
- DNS 与 HTTP 捕获字段口径与 `docs/unified-terminology.md` 一致
- 旧表桥接只允许作为兼容来源，不允许继续扩大旧 `TblInteraction` 的职责

### 2. 建立稳定的 token 归因链路

目标是让所有进入统一表的 Interaction 都能尽可能稳定地回填：

- `payload_id`
- `case_id`

要求：

- 归因规则唯一，以 Payload token 关联为准
- DNS 与 HTTP 必须复用同一套归因逻辑，不允许复制两份近似实现继续漂移
- 找不到归因对象时允许保留空值，但行为必须一致、可测试

### 3. 收口 `/api/v2/interactions` 查询与明细语义

目标是让 API 读取链路真正基于统一 `Interaction` 模型工作。

至少覆盖：

- `GET /api/v2/interactions`
- `GET /api/v2/interactions/:id`
- `GET /api/v2/cases/:id/interactions`

要求：

- 列表与明细查询使用统一模型
- 返回字段口径与统一术语一致
- not found 返回标准 404
- 至少支持按 `case_id`、`token`、`type` 查询

### 4. 对齐 MCP `wait_for_interaction` 与 `list_interactions`

目标是让 Agent 侧读取交互结果时，也复用同一套后端事实数据。

至少覆盖：

- `list_interactions`
- `wait_for_interaction`

要求：

- MCP 查询结果来自统一 `/api/v2/interactions`
- `wait_for_interaction` 至少能基于 token 轮询统一表中的 DNS / HTTP 交互
- MCP 不得自己拼装另一套 Interaction 语义

## 禁止越界项

Windsurf 在本 Sprint 中不得进入以下内容：

- 不实现 Evidence 聚合、评分、解释
- 不实现 Evidence 导出（JSON/Markdown/PDF）
- 不开始 Audit 全链路治理增强
- 不开始前端 Interaction Timeline 页面开发
- 不开始 CLI 大规模扩展
- 不开始 Burp / Yakit / Nuclei 工具集成
- 不扩展 SMTP / LDAP / SMB / FTP 协议捕获
- 不引入 Agent Run 生命周期或 Agent Dashboard 新能力

如果为了完成 Sprint C 需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作：

### Interaction 模型与服务

- `internal/models/interaction.go`
- `internal/interaction/`
- 相关测试文件

### 监听与写入桥接

- `server/webserver.go`
- `server/webapi.go`

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

1. 先统一 Interaction 写入与归因逻辑
2. 再收口 `/api/v2/interactions` 与 `/cases/:id/interactions` 的读取链路
3. 再对齐 MCP `list_interactions` / `wait_for_interaction`
4. 最后补齐 DNS / HTTP 双协议回归测试

## 必须补齐的测试

### 1. Interaction 归因测试

至少覆盖：

- 已知 token 的 DNS 交互能回填 `payload_id` 与 `case_id`
- 已知 token 的 HTTP 交互能回填 `payload_id` 与 `case_id`
- 未知 token 的交互保留空归因但写入成功
- DNS 与 HTTP 使用同一套归因规则，不出现字段口径分叉

### 2. Interaction API 测试

至少覆盖：

- `/api/v2/interactions` 能列出统一表数据
- 按 `case_id` 查询可返回关联交互
- 按 `token` 查询可返回关联交互
- `/api/v2/interactions/:id` 不存在时返回标准 404
- `/api/v2/cases/:id/interactions` 与统一表结果一致

### 3. MCP 读取对齐测试

至少覆盖：

- `list_interactions` 能读取统一 API 返回
- `wait_for_interaction` 能基于 token 等到交互结果
- `wait_for_interaction` 超时路径返回一致错误语义
- MCP 返回的 Interaction 字段与 API 响应语义一致

### 4. Sprint B 遗留测试补强

作为 Sprint C 入场清单，本 Sprint 需顺手补齐：

- 一个真实鉴权通过的 preview 接口行为测试
- MCP `create_payload` / `create_oast_probe` 请求体转换断言测试

这两项不属于 Sprint C 主目标，但必须在本 Sprint 里补平，避免遗留继续滚动。

## 完成定义

只有同时满足以下条件，Sprint C 才能视为完成：

1. DNS / HTTP 交互都能进入统一 `Interaction` 模型
2. token 归因链路能稳定回填 `payload_id` / `case_id`
3. `/api/v2/interactions` 与 case 维度查询都基于统一模型
4. MCP `list_interactions` / `wait_for_interaction` 复用统一后端语义
5. Sprint B 遗留测试补强已完成
6. 相关测试通过，且 `GOCACHE=/tmp/gocache go test ./...` 继续通过

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式向 Codex 回传：

### 1. 实际修改范围

- 修改了哪些目录
- 修改了哪些关键文件
- 哪些计划内文件未动，原因是什么

### 2. 实际实现内容

- DNS / HTTP 如何统一进入 `Interaction` 主模型
- token 归因如何收口
- `/api/v2/interactions` 读取链路如何统一
- MCP `wait_for_interaction` / `list_interactions` 如何与 API 对齐

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
- 哪些点被明确留给 Sprint D
- 哪些旧模型桥接仍然存在

## Codex 验收问题

Codex 在验收 Sprint C 时只围绕以下问题判断：

1. DNS / HTTP 是否真正进入统一 `Interaction` 模型
2. token -> payload -> case 的归因链路是否唯一且稳定
3. `/api/v2/interactions` 是否真正基于统一模型，而不是旧表拼装
4. MCP 是否复用后端真实 Interaction 契约
5. 当前输出是否足够作为 Sprint D 的 Evidence 输入
6. 是否严格没有越界到 Evidence / Export / Frontend / 工具集成

## 验收结论类型

Codex 对 Sprint C 的验收只会给出三种结果：

- **通过**：可进入 Sprint D
- **有条件通过**：允许进入 Sprint D，但必须挂明遗留项
- **不通过**：必须继续停留在 Sprint C 修正

## Sprint C 完成后的下一步

只有 Sprint C 被 Codex 验收通过后，才进入 `Sprint D：Evidence 聚合、评分与导出`。
