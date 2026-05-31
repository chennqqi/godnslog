# GODNSLOG 2.0 Sprint L 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-31-godnslog-2-sprint-l-package.md`
- `internal/agentrun/review.go`
- `internal/agentrun/review_test.go`
- `internal/agentrun/service.go`
- `internal/mcp/server.go`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`
- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

## 验收结论

**结论：Sprint L 未通过验收，需要返修。**

本轮已经补入 Agent Run Review Packet 的后端 service、`GET /api/v2/agent-runs/:id/review`、Agent Run 详情页 Review Packet 区块，以及 MCP `export_report` 对 `agent_run_id` 的 Review API 路径。

但是 Sprint L 的完成定义要求“Agent Run detail -> Review -> Evidence 形成真实闭环，E2E 必须真实等待 review API request，MCP 必须证明 `export_report(agent_run_id=...)` 走 Review API”。当前实现和测试还没有满足这些硬门槛，并且相关 E2E 已真实失败。

## 已执行验证

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过，0 error；但有 1 个 warning：

- `frontend-next/e2e/agent-runs.spec.ts:1:29`：`Page` is defined but never used

```bash
cd frontend-next && npm run build
```

结果：通过。

备注：按当前环境约束，在非沙箱环境运行。

```bash
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：失败。

- 总计：`14 passed, 1 failed`
- 失败用例：
  - `frontend-next/e2e/agent-runs.spec.ts:179`
  - `Agent Runs › should display agent run detail with operations timeline and backlinks`

失败原因：

```text
strict mode violation: getByText('查看证据') resolved to 2 elements
```

页面现在同时存在：

- Review Packet 区块里的 `查看证据` button
- 快速链接里的 `查看证据` link

旧测试用 `getByText('查看证据')` 命中两个元素，因此 Playwright strict mode 失败。

本次 E2E 使用的 dev server 已停止。

## 已通过项

### 1. Review Packet service 已补入

部分通过。

`internal/agentrun/review.go` 已新增 `ReviewService` 和 `BuildReviewPacket`，能聚合：

- Agent Run detail
- Interaction summary
- Evidence
- Audit refs
- JSON / Markdown content

`internal/agentrun/review_test.go` 覆盖了：

- JSON packet
- Markdown packet
- 无 interaction 时不 500
- invalid format
- run not found

### 2. Review Packet 复用了统一 Evidence service

部分通过。

实现通过 `interaction.EvidenceService.GenerateEvidence` 生成 `*interaction.Evidence`，方向符合 Sprint L 的统一 Evidence 契约要求。

### 3. 前端 Review Packet 区块已补入

部分通过。

`frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已新增：

- `Review Packet` 区块
- `生成 JSON Review`
- `生成 Markdown Review`
- Evidence Strength / Confidence / Interaction Count / Unique Sources / Generated At 展示
- Markdown Preview
- `查看证据` 按钮

## 阻塞问题

### 1. Review API 的 run not found 契约错误

未通过。

Sprint L 要求：

- run 不存在返回 404

当前 `server/v2_api.go` 的 `v2GetAgentRunReview` 会把 `BuildReviewPacket` 返回的 `agent run not found` 当作普通错误，返回 500。

更关键的是，`server/v2_api_test.go` 当前测试也在断言 500：

```go
if w.Code != http.StatusInternalServerError {
    t.Errorf("Expected 500, got %d", w.Code)
}
```

这说明测试锁定了错误行为，而不是 Sprint L 计划里的契约。

返修要求：

- `BuildReviewPacket` 或 API 层提供可识别的 not found 错误。
- `GET /api/v2/agent-runs/:id/review` 对不存在 run 返回 404。
- `server/v2_api_test.go` 改为断言 404。

### 2. E2E 没有新增 Review Packet API 真实请求断言

未通过。

Sprint L 要求 `frontend-next/e2e/agent-runs.spec.ts` 至少覆盖：

- `page.waitForRequest` 捕获 `/api/v2/agent-runs/agent-run-1/review?format=json`
- JSON review 返回后页面展示 evidence strength / confidence / interaction count
- `page.waitForRequest` 捕获 `/api/v2/agent-runs/agent-run-1/review?format=markdown`
- Markdown review 返回后页面展示 markdown preview

当前 `agent-runs.spec.ts` 仍只有原有 3 条测试：

- list + filter
- detail + operation timeline + backlinks
- update status

没有 Review Packet 按钮点击测试，也没有 `waitForRequest` 断言 Review API。

返修要求：

- 在 `agent-runs.spec.ts` 增加 Review Packet E2E。
- E2E 必须点击 `生成 JSON Review` / `生成 Markdown Review`。
- E2E 必须等待并断言 `/api/v2/agent-runs/agent-run-1/review?format=...` 请求。
- 不允许只检查静态文字。

### 3. Agent Runs E2E 当前真实失败

未通过。

新增 Review 区块后，页面上有两个 `查看证据` 元素，导致原有测试 strict mode 失败。

返修要求：

- 将旧测试中的 `page.getByText('查看证据')` 改为更精确的 role locator，例如：

```ts
page.getByRole('link', { name: '查看证据' })
```

- Review 区块里的按钮也应使用 role locator 单独断言：

```ts
page.getByRole('button', { name: '查看证据' })
```

### 4. MCP `export_report(agent_run_id=...)` 缺少专门测试

未通过。

代码层已经看到 `internal/mcp/server.go` 在 `agent_run_id` 存在时调用：

```text
GET /api/v2/agent-runs/:id/review?format=<format>
```

但 `internal/mcp/server_test.go` 没有新增对应测试证明：

- 带 `agent_run_id` 时走 Review API
- 缺少 `agent:export_report` scope 时不调用 Review API
- operation result 包含 `agent_run_id`、`format`、`interaction_count`、`evidence_strength`

返修要求：

- 增加 `export_report` + `agent_run_id` 的 MCP 测试。
- 增加缺 scope 拒绝时不调用 Review API 的测试。
- 验证 operation result 的 review summary。

### 5. Review API 正向路径测试不足

未通过。

`server/v2_api_test.go` 当前只覆盖：

- unauthenticated 401
- invalid format 400
- run not found，但断言为 500

缺少 Sprint L 要求的：

- json review 200 且有 `evidence`
- markdown review 200 且有 `content`
- response 不包含完整 API Key / Authorization header / raw secret

返修要求：

- 在 API test 中插入真实 AgentRun + Interaction + Operation。
- 验证 JSON response 的 `data.evidence` 和 `data.interaction_summary`。
- 验证 Markdown response 的 `data.content`。
- 验证敏感字段不泄露。

## 返修清单

1. 修复 `GET /api/v2/agent-runs/:id/review` 不存在 run 返回 404。
2. 修正 `server/v2_api_test.go`，不要把 500 当成预期行为。
3. 补 Review API 的 json / markdown 正向测试。
4. 补 MCP `export_report(agent_run_id=...)` 走 Review API 的测试。
5. 补 MCP 缺 scope 时不调用 Review API 的测试。
6. 补 Agent Runs Review Packet E2E，必须真实等待 review API request。
7. 修复 `查看证据` locator strict mode 冲突。
8. 清理 `frontend-next/e2e/agent-runs.spec.ts` 未使用的 `Page` import。
9. 重新运行并记录：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

## 复验重点

下一轮验收只要聚焦以下问题：

1. Review API not found 是否返回 404。
2. Review API json / markdown 正向路径是否有真实测试。
3. MCP `export_report(agent_run_id=...)` 是否被测试证明走 Review API。
4. 缺 scope 时是否不调用 Review API。
5. Agent Runs E2E 是否真实点击 Review 按钮并等待 request。
6. `agent-runs.spec.ts` 是否无 `test.skip` / `test.only` / `waitForTimeout` / 静态空测。
7. 是否没有越界到 Scanner Hub、生命周期治理、批量操作或完整 Agent 管理。
