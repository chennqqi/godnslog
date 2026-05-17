# GODNSLOG 2.0 Sprint C 验收

## 验收对象

- `internal/models/interaction.go`
- `internal/interaction/`
- `server/webserver.go`
- `server/webapi.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`

## 验收结论

**结论：不通过。**

Sprint C 当前既没有满足实施包的全部验收要求，也没有通过全量验证，不能进入 Sprint D。

## 关键问题

### 1. Sprint B 遗留测试补强没有真正完成，而且当前直接导致测试失败

`TestPayloadPreviewReturnsRenderedTemplate` 现在会稳定失败：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:575)

实际结果：

- 预期 `200`
- 实际 `401 unauthorized`

这说明本 Sprint 被明确挂入入场清单的“真实鉴权通过 preview 接口行为测试”并未做实，反而把 server 测试打红了。

### 2. `GET /api/v2/interactions/:id` 仍未收口到统一 Interaction 主模型语义

`v2GetInteraction` 还在按整数 ID 解析：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1314)

统一 `Interaction` 主模型的 ID 是字符串 UUID 风格主键，不是旧自增整数：

- [internal/models/interaction.go](/data/dev/github.com/chennqqi/godnslog/internal/models/interaction.go:38)

同时 not found 仍返回旧业务码 `code: 6`：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1338)

这直接违反了 Sprint C 第 3 条要求：读取链路必须真正基于统一模型，且标准 404 收口。

### 3. token 归因链路没有被验证通过，当前测试自己承认“归因被推迟”

最关键的归因测试 `TestInteractionTokenAttributionChain` 没有断言 `payload_id` / `case_id` 真正回填成功，反而写了明确的 TODO：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:493)

当前测试结论实际上是：

- 只验证 dual-write 成功
- 没有证明 `token -> payload -> case` 归因链路稳定成立

这使 Sprint C 的核心目标 2 没有被证实完成。

### 4. MCP 对齐测试仍然偏弱，没有证明 `wait_for_interaction` / `list_interactions` 与统一 API 契约完全对齐

当前 MCP 测试只验证了“调用成功”，没有对返回字段语义做足够强的断言：

- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:159)
- [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:192)

例如：

- `list_interactions` 没有验证返回项字段是否与 API 一致
- `wait_for_interaction` 没有验证超时路径语义
- 也没有证明轮询读取的是统一 Interaction 返回结构，而不是只要 `items` 非空就算通过

## 对 6 个验收问题的判断

1. DNS / HTTP 是否真正进入统一 `Interaction` 模型：**部分满足**
2. token -> payload -> case 的归因链路是否唯一且稳定：**否**
3. `/api/v2/interactions` 是否真正基于统一模型，而不是旧表拼装：**部分满足，但 `:id` 明细不通过**
4. MCP 是否复用后端真实 Interaction 契约：**部分满足**
5. 当前输出是否足够作为 Sprint D 的 Evidence 输入：**否**
6. 是否严格没有越界到 Evidence / Export / Frontend / 工具集成：**是**

## 修正要求

Windsurf 需继续停留在 Sprint C，至少完成以下修正后再回传验收：

1. 修复 preview 行为测试，使其在真实鉴权条件下稳定通过
2. 将 `v2GetInteraction` 改为按统一字符串 ID 读取，并收口 not found 为标准 404 / `code: 404`
3. 把归因测试补成强断言，必须直接验证 `payload_id` 与 `case_id` 被正确回填
4. 补齐 MCP `list_interactions` / `wait_for_interaction` 的返回语义断言与超时路径断言
5. 再次执行 `GOCACHE=/tmp/gocache go test ./...`，确保全量通过

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果：

- `./server` 失败
- `./...` 失败

当前不满足 Sprint C 完成定义。
