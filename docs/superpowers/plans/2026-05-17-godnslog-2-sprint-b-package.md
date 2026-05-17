# GODNSLOG 2.0 Sprint B Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint B
- **Sprint 主题**：Probe 创建与 Payload 渲染
- **所属阶段**：Phase 2 - Probe → Interaction → Evidence → Export → Audit Closed Loop

## Sprint 目标

建立 MVP 的第一段真实业务闭环：

1. 可以基于 Case 创建 Payload
2. Payload 能稳定生成唯一 token
3. Payload 模板能按统一变量口径完成渲染
4. MCP `create_oast_probe` 能复用同一套后端契约

本 Sprint 的重点不是“交互捕获”，而是把 **创建 Probe** 这一段做实，并为后续 Interaction/Evidence 阶段准备稳定输入。

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/agent-native-specification.md`
- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-a-package.md`
- `docs/superpowers/acceptance/2026-05-17-godnslog-2-sprint-a-final-close.md`

## 实施范围

本 Sprint 只允许覆盖以下 4 个主题。

### 1. 统一 Payload 创建契约

目标是让 `/api/v2/payloads` 真正按统一模型工作，至少覆盖：

- `case_id`
- `template_id`
- `template_rendered`
- `variables`
- `token`
- `status`
- `expires_at`

要求：

- 后端主契约以 `internal/models.Payload` 为准
- 请求/响应字段口径与 `docs/unified-terminology.md` 一致
- 不再继续放大旧 `TblPayload` 与统一模型之间的语义漂移

### 2. 建立统一模板渲染逻辑

目标是让 Payload 生成与预览走同一套渲染规则。

至少需要支持：

- `token`
- `domain`
- `case`
- `callback_url`
- 用户自定义 `variables`

要求：

- 创建 Payload 时使用统一渲染器
- 预览 Payload 时使用同一渲染器
- MCP `create_payload` / `create_oast_probe` 间接复用同一后端结果

### 3. 建立 Probe 最小输出模型

Sprint B 不单独创建 `Probe` 持久化实体，但必须形成一致的 Probe 输出语义：

- `probe_id = case_id:payload_id`
- `payload_id`
- `case_id`
- `token`
- `template_rendered`
- `expected_protocols`

要求：

- `/api/v2/payloads` 创建响应中能提供创建后交付所需信息
- MCP `create_oast_probe` 返回的字段与 API 语义对齐

### 4. 统一 Payload 预览与批量创建边界

本 Sprint 只要求：

- `preview` 走统一渲染逻辑
- `batch create` 不破坏统一契约

不要求本 Sprint 做复杂批量策略，只要求不成为后续返工点。

## 禁止越界项

Windsurf 在本 Sprint 中不得进入以下内容：

- 不实现 DNS/HTTP Interaction 捕获
- 不实现 Evidence 汇总、评分、导出
- 不开始前端 Payload Studio 页面开发
- 不开始 Scanner 集成
- 不开始 CLI 大规模扩展
- 不开始 Agent Run 治理或 Agent Dashboard
- 不扩展 SMTP/LDAP/SMB/FTP 协议逻辑

如果为了完成 Sprint B 需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作：

### Payload 模型与服务

- `internal/models/payload.go`
- `internal/payload/`
- 相关测试文件

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

1. 先统一 `Payload` 主模型与请求/响应字段口径
2. 再抽出统一模板渲染函数
3. 再接入 `v2CreatePayload` 和 `v2PreviewPayload`
4. 最后对齐 MCP `create_payload` / `create_oast_probe`

## 必须补齐的测试

### 1. Payload 创建测试

至少覆盖：

- 有效 `case_id + template_id` 可创建成功
- token 唯一且非空
- 创建结果包含 `template_rendered`
- 非法模板被拒绝

### 2. 模板渲染测试

至少覆盖：

- `token` 变量渲染
- `domain` 变量渲染
- `case` 变量渲染
- `callback_url` 渲染
- 自定义变量覆盖

### 3. 预览测试

至少覆盖：

- 预览接口使用和创建相同的渲染逻辑
- 不存在的 payload 返回标准 404

### 4. MCP 对齐测试

至少覆盖：

- `create_oast_probe` 能返回 `case_id`、`payload_id`、`token`
- `probe_id` 符合 `case_id:payload_id`
- MCP 返回的 payload 字段与 API 响应语义一致

## 完成定义

只有同时满足以下条件，Sprint B 才能视为完成：

1. `POST /api/v2/payloads` 能基于统一契约创建 Payload
2. Payload 创建和预览走同一套渲染逻辑
3. 返回结果包含 Probe 交付所需最小字段
4. MCP `create_oast_probe` 与后端 Payload 语义对齐
5. 相关测试通过，且 `GOCACHE=/tmp/gocache go test ./...` 继续通过

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式向 Codex 回传：

### 1. 实际修改范围

- 修改了哪些目录
- 修改了哪些关键文件
- 哪些计划内文件未动，原因是什么

### 2. 实际实现内容

- Payload 主契约如何统一
- 模板渲染如何统一
- 创建/预览接口如何复用同一逻辑
- MCP `create_oast_probe` 如何与 API 对齐

### 3. 实际验证命令

至少包含实际执行过的命令，例如：

- `GOCACHE=/tmp/gocache go test ./internal/payload`
- `GOCACHE=/tmp/gocache go test ./internal/mcp ./server`
- `GOCACHE=/tmp/gocache go test ./...`

### 4. 测试结果

- 新增了哪些测试
- 修改了哪些测试
- 哪些预期测试仍未覆盖

### 5. 风险与偏差

- 哪些地方与原规划不完全一致
- 哪些点被明确留给 Sprint C
- 哪些旧模型桥接仍然存在

## Codex 验收问题

Codex 在验收 Sprint B 时只围绕以下问题判断：

1. Payload 契约是否真正向统一模型收敛
2. 模板渲染是否只有一套主逻辑
3. 创建和预览是否复用同一渲染规则
4. Probe 最小输出是否足够给后续 Interaction 阶段使用
5. MCP 是否复用后端真实契约，而不是自己拼装另一套语义
6. 是否严格没有越界到 Interaction/Evidence/前端/工具集成

## 验收结论类型

Codex 对 Sprint B 的验收只会给出三种结果：

- **通过**：可进入 Sprint C
- **有条件通过**：允许进入 Sprint C，但必须挂明遗留项
- **不通过**：必须继续停留在 Sprint B 修正

## Sprint B 完成后的下一步

只有 Sprint B 被 Codex 验收通过后，才进入 `Sprint C：DNS/HTTP Interaction 捕获与归因`。
