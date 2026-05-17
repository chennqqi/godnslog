# GODNSLOG 2.0 Sprint B 验收

## 验收对象

- `internal/models/payload.go`
- `internal/payload/`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

## 验收结论

**结论：不通过。**

Sprint B 的测试虽然通过，但当前实现还没有满足实施包中定义的统一契约、统一渲染和 MCP 对齐要求，暂时不能进入 Sprint C。

## 关键问题

### 1. 创建与预览没有复用同一套渲染语义

- `CreatePayload` 只调用了不带 case 变量的渲染路径：
  [internal/payload/service.go](/data/dev/github.com/chennqqi/godnslog/internal/payload/service.go:38)
- 虽然存在 `RenderTemplateWithCase`，但创建流程并未使用：
  [internal/models/payload.go](/data/dev/github.com/chennqqi/godnslog/internal/models/payload.go:114)

这意味着 Sprint B 要求支持的 `case` 变量并没有真正进入主创建链路，`template_rendered` 的统一渲染能力不完整。

### 2. MCP 没有复用后端真实契约，而是在发送另一套字段语义

- `create_oast_probe` 向 `/api/v2/payloads` 发送的是 `expires_in` 和 `expected_protocols`：
  [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:110)
- `create_payload` 也是同样问题：
  [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:177)
- 但后端统一请求模型接收的是 `expires_at` 和 `expected_protocol`：
  [internal/models/payload.go](/data/dev/github.com/chennqqi/godnslog/internal/models/payload.go:139)

这不符合 Sprint B 第 5 条验收问题：MCP 应复用后端真实契约，而不是继续拼装一套旁路语义。

### 3. API 仍残留旧模型与旧错误码口径，Payload 契约没有真正收敛

- `v2GetPayload` 仍读取 `models.TblPayload` 并返回旧 `models.Payload`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:984)
- 不存在时仍返回旧业务码 `code: 6`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1009)

Sprint B 的目标是让 `/api/v2/payloads` 向统一模型收敛；当前只在创建和预览局部接入了新服务，读路径仍停留在旧口径。

### 4. 预览接口错误处理不符合 Sprint B 要求

- `v2PreviewPayload` 对 “payload 不存在” 没有转成标准 404，而是统一回 500：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1120)

实施包明确要求预览接口在对象不存在时返回标准 404，这一点尚未完成。

### 5. 测试没有覆盖本 Sprint 的关键验收点

- 预览测试实际上只验证了“路由存在且未通过鉴权”，没有验证统一渲染逻辑，也没有验证 404：
  [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:328)
- Payload 服务测试没有验证 `case_id`、`template_rendered`、非法模板拒绝、自定义变量覆盖等 Sprint B 必测项：
  [internal/payload/service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/payload/service_test.go:10)
- MCP 测试验证了 `probe_id`、`case_id`、`payload_id`，但没有验证请求字段已经与 API 统一契约对齐：
  [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:330)

## 对 6 个验收问题的判断

1. Payload 契约是否真正向统一模型收敛：**否**
2. 模板渲染是否只有一套主逻辑：**否**
3. 创建和预览是否复用同一渲染规则：**部分满足，但不足以通过**
4. Probe 最小输出是否足够给后续 Interaction 使用：**部分满足**
5. MCP 是否复用后端真实契约：**否**
6. 是否严格没有越界：**是**

## 修正要求

Windsurf 需继续停留在 Sprint B，完成以下修正后再回传验收：

1. 让 Payload 创建主链路真正支持 `case` 变量渲染，并明确唯一主渲染入口。
2. 统一 MCP 请求字段到后端真实契约，消除 `expires_in` / `expected_protocols` 这类旁路字段。
3. 将 `v2GetPayload` 与相关读路径继续向统一模型收敛，至少清理 Payload 读取链路中的旧响应口径。
4. 修复 `v2PreviewPayload` 的 not found 分支，返回标准 404。
5. 按 Sprint B 实施包补齐真实测试，不接受“路由存在型”占位测试。

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/payload ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果：测试通过，但**验收不通过**。
