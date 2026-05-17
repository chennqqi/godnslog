# GODNSLOG 2.0 Sprint C 修正后复验

## 验收对象

- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `internal/models/interaction.go`

## 验收结论

**结论：不通过。**

本轮复验中，测试已恢复通过，但上轮明确退回的核心问题并没有被实际修复，因此 Sprint C 仍不能结束。

## 未修复问题

### 1. `v2GetInteraction` 仍是旧 ID 语义

当前实现仍然：

- 把 `id` 解析为整数：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1316)
- 用整数主键方式读取统一 `Interaction`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1331)
- not found 仍返回旧业务码 `code: 6`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1340)

这与统一模型的字符串 ID 主键不一致：

- [internal/models/interaction.go](/data/dev/github.com/chennqqi/godnslog/internal/models/interaction.go:38)

因此，Sprint C 第 3 条验收要求仍未满足。

### 2. 归因测试仍然没有强断言

`TestInteractionTokenAttributionChain` 依然保留 TODO，并明确写着 attribution deferred：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:495)

当前测试没有直接验证：

- `payload_id`
- `case_id`

是否被正确回填。  
也就是说，Sprint C 最核心的 `token -> payload -> case` 归因链路仍未被证明成立。

### 3. MCP 读取测试仍未补足验收级断言

`list_interactions` 与 `wait_for_interaction` 测试依旧主要验证“调用成功”：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:159)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:192)

仍缺少：

- 返回字段语义断言
- 超时路径断言
- 与统一 Interaction API 输出的一致性断言

## 本轮变化说明

本轮相比上次唯一变化是：

- `go test` 已恢复为通过

但这是“测试状态恢复”，不是“需求收口完成”。验收基于 Sprint C 包定义的目标，而不是只看测试是否为绿。

## 对 6 个验收问题的判断

1. DNS / HTTP 是否真正进入统一 `Interaction` 模型：**部分满足**
2. token -> payload -> case 的归因链路是否唯一且稳定：**否**
3. `/api/v2/interactions` 是否真正基于统一模型，而不是旧表拼装：**部分满足，但 `:id` 明细不通过**
4. MCP 是否复用后端真实 Interaction 契约：**部分满足**
5. 当前输出是否足够作为 Sprint D 的 Evidence 输入：**否**
6. 是否严格没有越界到 Evidence / Export / Frontend / 工具集成：**是**

## 修正要求

Windsurf 继续停留在 Sprint C，必须完成以下 3 个硬修正后再回传：

1. `v2GetInteraction` 改为统一字符串 ID 读取，并收口到标准 `404 / code: 404`
2. `TestInteractionTokenAttributionChain` 改为强断言，直接验证 `payload_id` / `case_id`
3. 补强 MCP `list_interactions` / `wait_for_interaction` 的语义与超时断言

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果通过，但**Sprint C 复验仍不通过**。
