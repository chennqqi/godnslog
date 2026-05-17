# GODNSLOG 2.0 Sprint B 最终验收

## 验收对象

- `internal/mcp/server.go`
- `internal/payload/service.go`
- `internal/payload/service_test.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `models/v2.go`

## 验收结论

**结论：有条件通过。**

Sprint B 上一轮卡住的核心契约问题已经修复，可以进入 Sprint C。  
仍有少量测试表达不够直接的问题，需要在 Sprint C 初始阶段顺手补强，但不再阻塞阶段推进。

## 本轮通过项

### 1. MCP 已收敛到后端真实契约

`create_oast_probe` 不再把 `expires_in` 直接下发给 API，而是在 MCP 层转换成 `expires_at`：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:92)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:125)

`create_payload` 也是同样处理：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:203)

API 请求模型中的旁路字段 `expires_in` 已移除，保持为统一契约：

- [models/v2.go](/data/dev/github.com/chennqqi/godnslog/models/v2.go:134)
- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:927)

### 2. `expected_protocols` 已收敛为后端单值字段

MCP 现在明确把 `expected_protocols` 数组收敛成单值 `expected_protocol`：

- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:101)
- [internal/mcp/server.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server.go:130)

后端统一模型仍保持单值语义，契约一致：

- [internal/models/payload.go](/data/dev/github.com/chennqqi/godnslog/internal/models/payload.go:139)

### 3. Payload 读取与预览链路保持统一模型

- `v2CreatePayload` 继续走统一服务：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:925)
- `v2GetPayload` 不存在时返回标准 404：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:984)
- `v2PreviewPayload` 不存在时返回标准 404：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1087)

### 4. `case` 变量渲染已补强为强断言

Payload 服务测试现在对 `case` 变量做了实际断言，不再只是检查非空：

- [internal/payload/service_test.go](/data/dev/github.com/chennqqi/godnslog/internal/payload/service_test.go:125)

## 保留问题

### 1. 预览测试仍偏“服务级模拟”，不是完整接口行为验证

当前 [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:329) 已经从原来的“未鉴权占位测试”前进了一步，但仍主要通过 service 读取结果来间接证明预览逻辑，而不是直接验证鉴权通过后的 `/api/v2/payloads/:id/preview` HTTP 响应体。

这属于测试力度问题，不再是契约错误。

### 2. MCP 测试对请求体转换的断言还可以更强

[internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:127) 目前仍以工具成功为主，没有把 `expires_at` 和 `expected_protocol` 的请求体内容直接断言到位。

这同样属于测试补强项，不阻塞 Sprint B 关闭。

## 对 6 个验收问题的判断

1. Payload 契约是否真正向统一模型收敛：**是**
2. 模板渲染是否只有一套主逻辑：**是**
3. 创建和预览是否复用同一渲染规则：**是**
4. Probe 最小输出是否足够给后续 Interaction 使用：**是**
5. MCP 是否复用后端真实契约：**是**
6. 是否严格没有越界：**是**

## 最终判断

Sprint B 可以结束，进入：

- `Sprint C：DNS/HTTP Interaction 捕获与归因`

建议将以下两项挂入 Sprint C 的入场清单，但不单独阻塞：

1. 补一个真实鉴权通过的 preview 接口行为测试
2. 补 MCP 请求体转换断言测试

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/payload ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果通过。
