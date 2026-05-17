# GODNSLOG 2.0 Sprint A Closing Patch 验收结论

## 验收对象

- Windsurf 最新一轮 `Sprint A closing patch`
- 当前仓库实现状态

## 验收结论

**结论：仍不通过，但只剩 1 个阻断项。**

这轮 patch 已经把运行时 Swagger 切到了 2.0 契约，说明 Sprint A 的文档契约部分已经闭合。当前唯一剩余的阻断项，是 `/api/v2` 仍然存在旧业务错误码口径残留。

## 本轮确认通过的点

### 1. 运行时 Swagger 已明确切换到 2.0 契约

[server/webserver.go](/data/dev/github.com/chennqqi/godnslog/server/webserver.go:255) 现在使用：

- `ginSwagger.URL("/docs/openapi.yaml")`
- `r.StaticFile("/docs/openapi.yaml", "docs/openapi.yaml")`

这意味着 Swagger UI 已经明确指向 `docs/openapi.yaml`，不再默认依赖旧的 `server/docs/swagger.yaml`。

这一项相较上轮已达标。

### 2. 核心参数错误口径已有局部修复

例如 [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:398) 中的 `v2CreateCase`，现在参数错误已从旧的 `code: 2` 调整为 `code: 400`。

说明 Windsurf 已开始真正清理旧业务码。

### 3. 全量测试仍然通过

重新执行 `GOCACHE=/tmp/gocache go test ./...`，结果继续通过。

## 仍然不通过的原因

### P1：`/api/v2` 旧业务错误码仍有残留，且范围不小

当前仓库中仍然存在多处旧口径错误码：

- [server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:125) 仍返回 `code: 5`
- [server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:141) 和 [server/middleware.go](/data/dev/github.com/chennqqi/godnslog/server/middleware.go:152) 仍返回 `code: 4`
- `server/v2_api.go` 中仍有多处 `code: 4`，例如：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:670)
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1172)
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1678)
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:2095)

这说明当前不是“漏了 1-2 处”，而是权限/禁止类错误仍然沿用旧业务码体系。

Sprint A 要求的是“统一基础响应契约”，因此这些残留仍然构成阻断。

## 验收判断

### 已完成

- 完整 API Key 校验
- `IsAgent` 进入 server 侧桥接模型
- 运行时 Swagger 切到 2.0 契约
- 大量 500/400 类返回开始收敛

### 未完成

- 旧业务错误码 `4/5` 在核心 API 和中间件中仍有残留

## 结论建议

Sprint A 仍然**不得关闭**，但已经只剩最后一个收口动作：

1. 统一 `server/middleware.go` 与 `server/v2_api.go` 中所有旧业务错误码残留
2. 至少清掉 `code: 4`、`code: 5`
3. 对应补 401/403/500 场景的响应格式测试

这一轮完成后，Codex 可以直接进行 Sprint A 最终关闭验收。
