# GODNSLOG 2.0 Sprint A Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint A
- **Sprint 主题**：统一领域模型与 API 契约基线
- **所属阶段**：Phase 1 - Unified Domain Model and API Contract

## Sprint 目标

为后续所有功能建立统一的基础契约，确保：

1. 核心实体命名和字段口径统一
2. `/api/v2` 响应格式统一
3. 鉴权和权限入口统一
4. OpenAPI/Swagger 与实际 API 对齐

本 Sprint 完成后，后续 Sprint 可以在稳定契约上继续开发，而不需要反复返工基础模型。

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/agent-native-specification.md`
- `docs/superpowers/plans/2026-05-17-godnslog-2-round-4-codex-windsurf-plan.md`

## 实施范围

本 Sprint 只允许覆盖以下四个主题。

### 1. 统一核心实体口径

目标是让当前实现与统一术语文档收敛，至少覆盖：

- Case
- Payload
- Interaction
- APIKey
- AgentRun
- AuditLog / AuditEvent

允许的动作：

- 统一字段名和结构体语义
- 补充缺失字段
- 删除明显重复或冲突的定义
- 统一模型所在目录与职责边界

### 2. 统一 `/api/v2` 响应格式

目标是让 `/api/v2` 下的核心接口返回一致的响应结构。

本 Sprint 只要求统一以下内容：

- 成功响应包装格式
- 失败响应包装格式
- 常用分页响应格式
- 认证失败/权限失败的错误返回

不要求本 Sprint 内把所有业务接口都补完，只要求先把基础口径统一并在核心路径上落地。

### 3. 统一鉴权与权限入口

目标是明确 `/api/v2` 的认证和权限边界，至少覆盖：

- JWT 用户认证
- API Key 认证
- `IsAgent` 标记识别
- Gin Context 中的统一身份注入
- 基础角色/权限判断入口

本 Sprint 关注的是“入口统一”，不是完整实现 Agent 治理。

### 4. OpenAPI/Swagger 对齐

目标是让 API 文档与当前真实接口口径对齐，至少覆盖：

- `/api/v2` 核心资源的基本说明
- 统一响应格式说明
- 鉴权方式说明
- 关键错误码说明

本 Sprint 只要求打通契约一致性，不要求完整覆盖所有未来接口。

## 禁止越界项

Windsurf 在本 Sprint 中不得进入以下内容：

- 不实现 Probe → Interaction → Evidence 的完整闭环
- 不开始前端页面开发
- 不开始 Nuclei / Burp / Yakit 集成
- 不扩展 SMTP / LDAP / SMB / FTP 等额外协议
- 不新增产品定位层文档
- 不开始 Agent Run Dashboard 或 Agent UI
- 不把 Swagger 对齐扩展成大规模接口重写

如果实施过程中发现某项工作超出上述边界，必须暂停并回传 Codex 重新裁剪范围。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作：

### 模型与领域定义

- `internal/models/`
- `internal/auth/`
- `internal/agentrun/`
- `internal/audit/`

### API 与认证入口

- `server/v2_api.go`
- `server/middleware.go`
- `server/router.go`
- `server/webui.go`

### API 文档

- `server/docs/`
- `docs/openapi.yaml`
- 与 Swagger/OpenAPI 生成相关的入口文件

### 测试

- `server/v2_api_test.go`
- 与 `internal/auth/`、`internal/agentrun/`、`internal/audit/` 相关的测试文件

## 建议实施顺序

Windsurf 应按以下顺序推进，避免返工：

1. 先梳理当前重复模型与字段冲突
2. 再定义统一 API 响应包装
3. 再收敛认证/权限入口
4. 最后更新 OpenAPI/Swagger 和测试

## 必须补齐的测试

本 Sprint 至少需要补以下类型的验证。

### 1. 模型一致性测试

验证核心模型不会因为统一字段而破坏现有存取行为。

### 2. `/api/v2` 响应格式测试

至少覆盖：

- 成功返回
- 参数错误
- 未认证
- 无权限

### 3. 鉴权入口测试

至少覆盖：

- JWT 可通过
- API Key 可通过
- 无凭证被拒绝
- Agent API Key 能被识别为 Agent 身份

### 4. Swagger/OpenAPI 对齐验证

至少验证：

- Swagger 文档可生成或可读取
- `/api/v2` 基础路径存在
- 鉴权方式与统一响应结构有描述

## 完成定义

只有同时满足以下条件，Sprint A 才能视为完成：

1. 统一术语文档涉及的核心实体，在当前实现里有明确映射关系
2. `/api/v2` 核心路径的响应格式已经统一
3. 认证入口不再分散为多套互相冲突的判断逻辑
4. `IsAgent` 至少进入鉴权识别链路
5. OpenAPI/Swagger 能反映新的基础契约
6. 相关测试通过，且 `GOCACHE=/tmp/gocache go test ./...` 仍保持通过

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式向 Codex 回传：

### 1. 实际修改范围

- 修改了哪些目录
- 修改了哪些关键文件
- 哪些计划内文件没有改，为什么

### 2. 实际实现内容

- 核心实体统一了哪些点
- API 响应格式如何统一
- 鉴权入口统一了哪些点
- Swagger/OpenAPI 更新了哪些点

### 3. 实际验证命令

必须列出实际执行过的命令，例如：

- `GOCACHE=/tmp/gocache go test ./server`
- `GOCACHE=/tmp/gocache go test ./internal/auth ./internal/agentrun ./internal/audit`
- `GOCACHE=/tmp/gocache go test ./...`

### 4. 测试结果

- 哪些测试新增
- 哪些测试修改
- 哪些测试仍然缺失

### 5. 风险与偏差

- 哪些地方与原规划不完全一致
- 哪些问题被延后到 Sprint B 或后续
- 哪些潜在技术债需要 Codex 在下一 Sprint 处理

## Codex 验收问题

Codex 在验收 Sprint A 时只围绕以下问题判断：

1. 是否真的建立了统一实体口径，而不是局部修补
2. `/api/v2` 是否已有统一响应结构，而不是少数接口示例化处理
3. 鉴权入口是否收敛到统一身份模型
4. `IsAgent` 是否已经被识别，而不是只存在于字段定义里
5. Swagger/OpenAPI 是否与实现口径一致
6. 是否严格没有越界到闭环业务、前端页面、工具集成

## 验收结论类型

Codex 对 Sprint A 的验收只会给出三种结果：

- **通过**：可进入 Sprint B
- **有条件通过**：允许进入 Sprint B，但必须挂明遗留项
- **不通过**：必须继续停留在 Sprint A 修正

## Sprint A 完成后的下一步

只有 Sprint A 被 Codex 验收通过后，才进入 `Sprint B：Probe 创建与 Payload 渲染`。
