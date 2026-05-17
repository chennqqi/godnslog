# GODNSLOG 2.0 Sprint C 二次复验

## 验收对象

- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server_test.go`

## 验收结论

**结论：不通过。**

这轮修正只完成了 `v2GetInteraction` 的字符串 ID / 404 收口；另外两项上轮硬要求仍未落地，因此 Sprint C 仍不能结束。

## 已完成修正

### 1. `v2GetInteraction` 已切到统一字符串 ID 语义

当前实现已改为：

- 直接按字符串 `id` 查询：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1316)
- not found 返回标准 `404 / code: 404`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1332)

这一项通过。

## 仍未通过的问题

### 1. 归因链路测试仍然没有强断言

`TestInteractionTokenAttributionChain` 依旧保留 TODO，并明确写着 attribution deferred：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:495)

当前仍然没有直接断言：

- `payload_id`
- `case_id`

是否被正确回填。  
因此，Sprint C 的核心目标“token -> payload -> case 稳定归因”仍未被证明。

### 2. MCP 读取测试仍然只是调用成功测试

以下测试仍旧没有补足语义断言：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:159)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:192)

缺失点仍然是：

- `list_interactions` 返回字段语义断言
- `wait_for_interaction` 超时路径断言
- 与统一 `/api/v2/interactions` 输出的一致性断言

## 对 6 个验收问题的判断

1. DNS / HTTP 是否真正进入统一 `Interaction` 模型：**部分满足**
2. token -> payload -> case 的归因链路是否唯一且稳定：**否**
3. `/api/v2/interactions` 是否真正基于统一模型，而不是旧表拼装：**是**
4. MCP 是否复用后端真实 Interaction 契约：**部分满足**
5. 当前输出是否足够作为 Sprint D 的 Evidence 输入：**否**
6. 是否严格没有越界到 Evidence / Export / Frontend / 工具集成：**是**

## 修正要求

Windsurf 继续停留在 Sprint C，只剩 2 个硬修正：

1. `TestInteractionTokenAttributionChain` 改为强断言，直接验证 `payload_id` / `case_id`
2. 补强 MCP `list_interactions` / `wait_for_interaction` 的语义与超时断言

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果通过，但**Sprint C 二次复验仍不通过**。
