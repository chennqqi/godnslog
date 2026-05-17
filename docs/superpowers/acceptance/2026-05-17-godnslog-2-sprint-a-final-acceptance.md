# GODNSLOG 2.0 Sprint A 最终关闭验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-a-package.md`
- `docs/superpowers/plans/2026-05-17-godnslog-2-sprint-a-remediation-package.md`
- Windsurf 最新一轮 Sprint A 收口结果

## 验收结论

**结论：仍不通过。**

这次收口后，Sprint A 已经完成了大部分基础契约工作，但仍有 2 个阻断级缺口没有闭合，因此 Sprint A 不能正式关闭，也不能进入 Sprint B。

## 本轮确认通过的点

### 1. API Key 已从 prefix 校验提升为完整 key 校验

[internal/auth/service.go](/data/dev/github.com/chennqqi/godnslog/internal/auth/service.go:177) 现在接收完整 key，并在 prefix 命中后继续校验 `apiKey.Key != fullKey`。

[server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:63) 也已改为把完整 `X-API-Key` 传入 `ValidateAPIKey`。

这一项相较上轮已达标。

### 2. `IsAgent` 已进入 server 侧 API Key 对象

[models/v2.go](/data/dev/github.com/chennqqi/godnslog/models/v2.go:66) 的 `TblAPIKey` 已新增 `IsAgent` 字段。

[server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:72) 在向旧模型桥接时也把 `key.IsAgent` 传递到了 `tblKey.IsAgent`。

这说明 `IsAgent` 已不再只存在于 `internal/auth` 独立链路。

### 3. 全量测试仍然通过

本轮重新执行 `GOCACHE=/tmp/gocache go test ./...`，结果继续通过。

## 仍然不通过的原因

### P1：运行时 Swagger 仍未真正切换到 2.0 OpenAPI 契约

虽然 [server/webserver.go](/data/dev/github.com/chennqqi/godnslog/server/webserver.go:255) 做了改动，但当前实现只是：

- 保留 `ginSwagger.WrapHandler(swaggerFiles.Handler)`
- 额外挂出 `/docs/openapi.yaml`

这并不等于 Swagger UI 已经使用 `docs/openapi.yaml`。

仓库里运行时 Swagger 生成产物仍然是旧的：
[server/docs/swagger.yaml](/data/dev/github.com/chennqqi/godnslog/server/docs/swagger.yaml:1)

也就是说现在仍然是：

- 运行时 Swagger UI 背后还是旧 `server/docs`
- 仓库文档另有一份 `docs/openapi.yaml`

两套契约并存，未真正统一。

### P1：`/api/v2` 错误码口径仍未收齐

虽然大量 `500` 风格已经统一，但核心接口仍存在旧业务码残留，例如：
[server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:398)

`v2CreateCase` 参数错误时仍返回：

- HTTP `400`
- 业务 `code: 2`

这与 Sprint A 要求的“统一成功/失败/分页响应结构与错误口径”仍不一致。

同类问题仍存在于若干核心路径，说明现在是“局部统一”，还不是“契约基线统一完成”。

## 验收判断

### 已完成

- 完整 API Key 校验
- `IsAgent` 进入 server 侧桥接模型
- 大量 500 类错误码收敛
- OpenAPI 仓库文档继续完善

### 未完成

- 运行时 Swagger 与 2.0 OpenAPI 契约真正统一
- 核心 `/api/v2` 错误码口径完全统一

## 结论建议

Sprint A 仍然**不得关闭**。

但当前剩余问题已经非常集中，建议 Windsurf 不再做“大修正包”，只做一轮 **Sprint A closing patch**，范围严格限制为：

1. 让运行时 Swagger UI 明确使用 2.0 契约，或明确替换/同步 `server/docs`
2. 统一核心 `/api/v2` 路径中的旧业务错误码残留，至少清掉 `code: 2/4/5` 这类旧口径

这两项完成后，Codex 再做 Sprint A 关闭验收。
