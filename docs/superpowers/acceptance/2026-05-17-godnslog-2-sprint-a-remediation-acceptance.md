# GODNSLOG 2.0 Sprint A 修正包验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-a-remediation-package.md`
- Windsurf 本次修正后的仓库实现状态

## 验收结论

**结论：有明显进展，但仍不通过。**

本次修正已经补上了一部分 Sprint A 缺口：

- `authenticateAPIKey` 不再是空实现
- `internal/models` 已新增统一 `APIKey` 与 `AuditLog` 主定义
- `docs/openapi.yaml` 已补充 2.0 响应结构说明
- 全量测试 `GOCACHE=/tmp/gocache go test ./...` 继续通过

但仍存在 3 个未完成的契约级问题，Sprint A 不能关闭，仍需保持在 Sprint A。

## 本次修正完成点

### 1. API Key 鉴权入口已接入真实服务

[server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:50) 中的 `authenticateAPIKey` 已调用 `authService.ValidateAPIKey`，不再是之前的 TODO。

### 2. 统一模型方向继续推进

`APIKey` 与 `AuditLog` 的统一主定义已经进入 `internal/models`：

- [internal/models/apikey.go](/data/dev/github.com/chennqqi/godnslog/internal/models/apikey.go:32)
- [internal/models/audit.go](/data/dev/github.com/chennqqi/godnslog/internal/models/audit.go:29)

这说明模型收敛方向是对的。

### 3. OpenAPI 仓库文档已部分对齐

[docs/openapi.yaml](/data/dev/github.com/chennqqi/godnslog/docs/openapi.yaml:988) 已把登录响应改为 `code / message / data` 结构，和当前 `/api/v2` 返回风格更接近。

## 仍然不通过的原因

### P1：API Key 校验仍然只按 prefix 校验，不是完整 key 校验

当前 `authenticateAPIKey` 先截取前 8 位 prefix：
[server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:57)

随后 `ValidateAPIKey` 也是按 `key_prefix` 查询：
[internal/auth/service.go](/data/dev/github.com/chennqqi/godnslog/internal/auth/service.go:101)
[internal/auth/service.go](/data/dev/github.com/chennqqi/godnslog/internal/auth/service.go:177)

这意味着当前认证链路本质上仍然不是“完整 API Key 校验”，而是“prefix 命中即通过”。这在安全上不成立，也不满足“真实校验逻辑”的修正要求。

### P1：`IsAgent` 没有真正穿透到 server 侧身份模型

虽然 `internal/auth` 中间件测试已经覆盖 `IsAgent`，但 `server` 实际链路里：

- `authenticateAPIKey` 返回的是旧的 `models.TblAPIKey`
- `TblAPIKey` 本身没有 `IsAgent` 字段：
  [models/v2.go](/data/dev/github.com/chennqqi/godnslog/models/v2.go:66)
- `server/webui.go` 只做了 `c.Set("api_key", key)`，没有统一注入 agent 身份上下文

这说明 `IsAgent` 目前只在 `internal/auth` 的独立测试链路中成立，还没有真正进入 `server` 的实际鉴权身份链路。

### P1：运行时 Swagger 仍然指向旧 `server/docs`

虽然 `docs/openapi.yaml` 已更新，但运行时 Swagger 仍然使用：
[server/webserver.go](/data/dev/github.com/chennqqi/godnslog/server/webserver.go:255)

而被暴露出来的 `server/docs/swagger.yaml` 仍然主要是旧 `/user/*`、`/admin/*` 契约：
[server/docs/swagger.yaml](/data/dev/github.com/chennqqi/godnslog/server/docs/swagger.yaml:18)

这意味着：

- 仓库 OpenAPI 文档一套
- 运行时 Swagger 文档一套
- 实际 `/api/v2` 实现又是一套

三者仍未统一。

### P2：核心 `/api/v2` 返回结构仍未彻底统一到统一响应模型

当前核心路径虽大多采用 `code/message/data` 风格，但仍存在不一致的错误码口径，例如：
[server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:398)

`v2CreateCase` 参数错误时仍返回：

- HTTP 400
- 业务 `code: 2`

这与 Sprint A 所要求的“统一成功/失败/分页返回结构”仍有差距。当前更像“风格接近”，还不是“统一契约”。

## 验收判断

### 已达成

- 修正方向正确
- 风险点已显著减少
- 代码和文档正在往统一契约收敛

### 未达成

- 完整 API Key 校验
- `IsAgent` 穿透到实际 `server` 身份链路
- 运行时 Swagger 与 `/api/v2` 契约统一
- 核心错误码口径统一

## 结论建议

Sprint A 仍**不得进入 Sprint B**。

建议 Windsurf 再进行一次 **Sprint A 最后一轮收口修正**，只补以下 4 项：

1. API Key 改为完整 key 校验，不只查 prefix
2. `IsAgent` 进入 `server` 实际鉴权上下文
3. 运行时 Swagger 切换或对齐到 2.0 契约
4. 核心 `/api/v2` 错误返回码口径统一

这 4 项补齐后，Codex 再进行 Sprint A 关闭验收。
