# GODNSLOG 2.0 Sprint C 最终复验

## 验收对象

- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/mcp/server_test.go`
- `internal/models/interaction.go`

## 验收结论

**结论：不通过。**

本轮只剩最后 1 个问题，但它正好是 Sprint C 的核心要求，因此仍不能进入 Sprint D。

## 已完成修正

### 1. `v2GetInteraction` 已完成统一 ID / 404 收口

- 字符串 ID 查询：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1316)
- 标准 `404 / code: 404`：
  [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:1332)

### 2. MCP 读取语义测试已明显补强

- `list_interactions` 已开始断言返回结构和关键字段：
  [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:159)
- `wait_for_interaction` 已增加 timeout 子用例：
  [internal/mcp/server_test.go](/data/dev/github.com/chennqqi/godnslog/internal/mcp/server_test.go:249)

这一部分我认为已达到 Sprint C 要求。

## 仍未通过的问题

### 1. 归因链路测试不是在验证真实归因逻辑，而是在手工塞值

`TestInteractionTokenAttributionChain` 里，在调用 `FromTblDnsWithAttribution` 之后，测试手工把：

- `interaction.PayloadID = &payload.ID`
- `interaction.CaseID = &caseID`

写回对象：

- [server/v2_api_test.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api_test.go:473)

这意味着当前强断言验证的是“手工写进去的值能被查出来”，而不是：

- `FromTblDnsWithAttribution`
- 或统一归因逻辑

真的基于 token 自动完成了 `payload_id` / `case_id` 回填。

而真正的归因逻辑在：

- [internal/models/interaction.go](/data/dev/github.com/chennqqi/godnslog/internal/models/interaction.go:146)

Sprint C 的核心目标是“token -> payload -> case 的归因链路稳定成立”，不是“测试里可以模拟出一个已归因结果”。  
所以这一步还不能算通过。

## 对 6 个验收问题的判断

1. DNS / HTTP 是否真正进入统一 `Interaction` 模型：**部分满足**
2. token -> payload -> case 的归因链路是否唯一且稳定：**否**
3. `/api/v2/interactions` 是否真正基于统一模型，而不是旧表拼装：**是**
4. MCP 是否复用后端真实 Interaction 契约：**是**
5. 当前输出是否足够作为 Sprint D 的 Evidence 输入：**否**
6. 是否严格没有越界到 Evidence / Export / Frontend / 工具集成：**是**

## 最后修正要求

Windsurf 只剩最后 1 个硬修正：

1. 去掉 `TestInteractionTokenAttributionChain` 中对 `PayloadID` / `CaseID` 的手工赋值，改为直接断言真实归因逻辑输出

如果现有归因逻辑因为 session/engine 可见性问题导致测不通，就应修归因实现或测试构造方式，而不是在测试里手工补值。

## 本次验证

已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./internal/mcp ./server ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果通过，但**Sprint C 最终复验仍不通过**。
