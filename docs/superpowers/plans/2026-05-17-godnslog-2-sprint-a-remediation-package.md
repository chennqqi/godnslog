# GODNSLOG 2.0 Sprint A Remediation Package

> **协作模式**
>
> - **Codex**：负责修正包规划与验收
> - **Windsurf**：负责按修正包补齐 Sprint A 缺口

## 修正目标

本修正包不是新 Sprint，而是 **Sprint A 补验收包**。目标只有一个：把上次验收不通过的基础契约缺口补齐，使 Sprint A 达到可通过状态。

## 仅修正以下缺口

### 1. API Key 鉴权必须真正落地

Windsurf 必须补齐：

- `X-API-Key` 的真实校验逻辑
- API Key 的有效性判断
- API Key 身份注入 Gin Context
- `IsAgent` 标记识别进入鉴权链路

### 2. 核心模型重复定义必须继续收敛

至少要消除或明确收敛以下重复定义的边界：

- `APIKey`
- `AuditLog`

要求不是“大重构”，而是让核心 API 契约只有一个明确的主定义来源。

### 3. `/api/v2` 响应包装必须在核心路径统一

至少覆盖以下核心路径：

- `/api/v2/auth/login`
- `/api/v2/cases`
- `/api/v2/payloads`
- `/api/v2/interactions`
- `/api/v2/apikeys`

要求：

- 成功返回格式一致
- 错误返回格式一致
- 分页返回结构一致

### 4. Swagger/OpenAPI 必须和真实 `/api/v2` 对齐

至少要满足：

- 运行时 Swagger 不再主要指向旧 `/user/*` 接口
- `docs/openapi.yaml` 与真实 `/api/v2` 基础认证方式一致
- 文档能体现统一响应包装

## 禁止越界项

本修正包仍然禁止进入：

- Sprint B 的 Probe/Payload 业务扩展
- Evidence 聚合与导出逻辑
- 前端页面开发
- Scanner 集成
- Agent UI / Agent Dashboard

## 必须新增或修正的测试

### 鉴权测试

- API Key 成功通过
- 无效 API Key 被拒绝
- Agent API Key 被识别为 Agent

### 响应格式测试

- 核心 `/api/v2` 路径成功返回统一结构
- 认证失败返回统一结构
- 参数错误返回统一结构

### 文档对齐验证

- Swagger/OpenAPI 至少存在 `/api/v2` 核心路径
- 鉴权方式描述与实现一致

## 完成定义

只有同时满足以下条件，修正包才算完成：

1. `authenticateAPIKey` 不再是 TODO
2. `IsAgent` 进入实际鉴权识别链路
3. `APIKey`/`AuditLog` 的主定义来源清晰
4. 核心 `/api/v2` 路径返回统一包装格式
5. Swagger/OpenAPI 与 `/api/v2` 实现基础契约一致
6. `GOCACHE=/tmp/gocache go test ./...` 继续通过

## Windsurf 回传要求

Windsurf 必须特别说明：

- 哪个模型是最终主定义来源
- 哪些重复定义被删除、保留或桥接
- API Key 校验具体接入了哪条链路
- 哪些 `/api/v2` 路径已完成统一返回
- Swagger/OpenAPI 实际如何验证

## 验收结果

本修正包通过后，Codex 才会重新判定 Sprint A 是否关闭。
