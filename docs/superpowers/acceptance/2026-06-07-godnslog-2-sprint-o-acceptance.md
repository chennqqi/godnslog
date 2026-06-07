# GODNSLOG 2.0 Sprint O Acceptance

## 结论

通过。

Sprint O 的主体实现已经落地：

- 后端新增 `POST /api/v2/agent-runs/:id/review-decision`。
- `internal/agentrun.Service.RecordReviewDecision` 会写入 `review_decision.<decision>` operation。
- Review Decision 会写入 `agent_run.review_decision_recorded` audit。
- Agent Run Detail 页面新增 Review Decision 记录入口。
- Review Queue item 开始暴露最近 review decision 字段。

但当前还不能关闭 Sprint O。主要问题不是构建或现有回归测试失败，而是 Sprint O 的核心验收闭环没有被 E2E/API 测试证明。

## 已完成项

### 1. Review Decision API

已新增路由：

```go
agentRuns.POST("/:id/review-decision", self.v2RecordReviewDecision)
```

处理函数会：

- 绑定 `AgentRunReviewDecisionRequest`。
- 校验当前登录用户。
- 调用 `RecordReviewDecision`。
- 对 not found / invalid decision / reason too long / review_packet_id 错误返回明确 HTTP 状态。

### 2. Review Decision Service

`RecordReviewDecision` 当前支持：

- `accepted`
- `false_positive`
- `needs_manual_followup`
- `insufficient_evidence`

并会：

- 校验 reason 长度不超过 500。
- 校验 `review_packet_id` 必须匹配当前 Agent Run。
- 追加 `review_decision.<decision>` operation。
- 写入 `agent_run.review_decision_recorded` audit。
- 返回 `operation_id` / `audit_ref_id`。

### 3. 前端入口

Agent Run Detail 页面已新增：

- “记录 Review Decision”按钮。
- Decision select。
- Reason textarea。
- `agentRunApi.createReviewDecision()` 调用。
- 提交成功后刷新 Agent Run Detail。

## 阻塞问题

### 1. Sprint O 主链路没有 E2E 覆盖

当前 `frontend-next/e2e/agent-runs.spec.ts` 仍只有 7 个用例：

- Agent Runs list。
- Agent Run detail。
- status update。
- Review Packet。
- Follow-up Action。
- Review Queue summary/filter。
- Follow-up History。

没有任何用例覆盖：

- 点击“记录 Review Decision”。
- 选择 `accepted` / `false_positive` / `needs_manual_followup` / `insufficient_evidence`。
- 等待并断言 `POST /api/v2/agent-runs/:id/review-decision`。
- 断言 request body 中包含 `decision`、`reason`、`review_packet_id`。
- 提交后 Operation timeline 出现 `review_decision.<decision>`。
- Audit 链接或 Audit API 中出现 `agent_run.review_decision_recorded`。
- 回到 Review Queue 后 summary / row state 随 decision 改变。

这使得 Sprint O 的核心闭环“Review Queue -> Detail -> Decision -> Operation -> Audit -> Queue closure”没有被验收测试证明。

返修要求：

1. 在 `agent-runs.spec.ts` 增加 Review Decision E2E。
2. 测试必须触发真实 UI 交互，而不是只检查静态文字。
3. 测试必须等待并断言 `POST /review-decision` 的 request body。
4. 测试必须验证 timeline 刷新后出现 `review_decision.accepted` 或对应 decision action。
5. 测试必须验证 Audit query 或 audit ref 闭环。
6. 测试必须验证 Review Queue summary / row state 在 decision 后变化。
7. 不得使用 `test.skip` / `test.only`。

### 2. Review Decision API 缺 server 层测试

`internal/agentrun/service_test.go` 已覆盖 service 层，但 `server/v2_api_test.go` 未覆盖：

- 未登录访问 `POST /api/v2/agent-runs/:id/review-decision` 返回 401。
- 登录后提交合法 decision 返回 200。
- invalid decision 返回 400。
- 不存在的 Agent Run 返回 404。
- 成功请求真实写入 operation 与 audit。

返修要求：

- 增加 `TestV2RecordReviewDecision` 或等价 server API 测试。
- 测试不能只 mock service；需要通过 HTTP handler 证明路由、鉴权、JSON binding、错误映射成立。

### 3. Review Queue closure 缺后端测试证明

`review_queue.go` 已尝试根据 `review_decision.*` 派生状态：

- `accepted` / `false_positive` -> `reviewed`
- `needs_manual_followup` / `insufficient_evidence` -> `needs_attention`

但 `internal/agentrun/review_queue_test.go` 没有覆盖 `review_decision.*` 场景。

返修要求：

- 增加 Review Queue service 测试：
  - `review_decision.accepted` 后 item `review_state=reviewed`，summary `reviewed=1`。
  - `review_decision.false_positive` 后 item `review_state=reviewed`。
  - `review_decision.needs_manual_followup` 后 item `review_state=needs_attention`。
  - 最新 decision 覆盖旧 follow-up / 旧 review 状态。

### 4. 前端提交后没有刷新 Follow-up History / Audit 状态

当前 Review Decision 提交成功后只刷新 Agent Run Detail：

```ts
const response = await agentRunApi.get(agentRun.id)
setAgentRun(response.data.data)
```

这可以让 operation timeline 刷新，但没有显式刷新 audit/follow-up history，也没有在 UI 上呈现 audit ref。若 Sprint O 要求形成“Decision -> Operation -> Audit”闭环，前端需要有可见路径或 E2E 可验证路径。

返修要求：

- 至少保证 Detail 页面可见新 `review_decision.<decision>` operation。
- 提供 Audit 跳转或 Audit ref，可验证 `resource_type=agent_run&resource_id=<id>`。
- E2E 必须覆盖这个路径。

## Scope Boundary

本轮检查未发现明显越界到：

- Scanner Hub。
- 生命周期治理 / retention。
- 批量操作。
- replay engine。
- 完整 Agent 管理平台。

## 验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS

cd frontend-next && npm run build
# FAIL in sandbox: Turbopack binding to a port denied by sandbox

cd frontend-next && npm run build
# PASS after approved rerun outside sandbox

cd frontend-next && npm run dev
# PASS: http://localhost:3000 ready

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 7 passed
```

## 第三次返修复验

结论：仍未通过。

本轮 Windsurf 新增了一个前端 E2E：

```text
should record review decision via UI and verify closure loop
```

该用例已证明：

- Agent Run Detail 页面可打开“记录 Review Decision”dialog。
- 可以选择 `Accepted`。
- 可以填写 reason。
- 会触发 `POST /api/v2/agent-runs/agent-run-1/review-decision`。
- request body 中包含 `accepted` 和 reason。

但该用例仍不足以关闭 Sprint O，因为它没有验证实际 closure loop：

1. 没有断言 request body 中的 `review_packet_id=agent-run-1` 或当前 Review Packet id。
2. 没有在提交后 mock detail refresh 为包含 `review_decision.accepted` 的 operation。
3. 没有断言 Operation timeline 出现 `review_decision.accepted`。
4. 没有断言 `agent_run.review_decision_recorded` audit 或 Audit API query。
5. 没有进入 / 返回 Review Queue。
6. 没有断言 Review Queue row / summary 从 not reviewed 变为 reviewed。
7. 测试名称写了 “verify closure loop”，但实际只验证了 UI submit 和部分 request body。

仍需返修：

- 扩展当前 E2E，而不是只保留提交检查。
- 提交后 detail refresh 必须返回新增 operation：

```text
review_decision.accepted
```

- E2E 必须断言 timeline 可见该 operation。
- E2E 必须断言 audit ref 或 Audit 页面/API 出现：

```text
agent_run.review_decision_recorded
resource_type=agent_run
resource_id=agent-run-1
```

- E2E 必须断言 Review Queue closure：

```text
not_reviewed -> reviewed
summary.reviewed += 1
summary.not_reviewed -= 1
```

### 第三次返修复验命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS with 1 warning: src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e

cd frontend-next && npm run build
# PASS after approved run outside sandbox

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# First run failed because Chromium headless shell was missing

cd frontend-next && npx playwright install chromium
# PASS

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 8 passed
```

## 第六次复验

结论：通过，Sprint O 验收结论不变。

本轮 Windsurf 修复后再次验证：

- Go 相关包通过。
- Go 全量通过。
- 前端生产构建通过。
- Agent Runs E2E 仍为 8 个用例并全部通过。
- Review Decision closure loop 覆盖仍然包含 `review_decision.accepted`、`agent_run.review_decision_recorded`、Review Queue `Reviewed=1` / `Not Reviewed=0`、`Decision: accepted`。

剩余非阻塞项仍存在：

- ESLint warning：`frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e`。

### 第六次复验命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS with 1 warning: src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e

cd frontend-next && npm run build
# PASS after approved run outside sandbox

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 8 passed
```

## 第五次返修复验

结论：通过。

本轮 Windsurf 修复了最后一个 Review Queue closure E2E 缺口：

- Review Queue mock URL 已改为真实路径：

```text
/api/v2/agent-runs/review-queue
```

- Mock response 已包含真实 `summary` 契约。
- E2E 已断言 summary 中 `Reviewed=1`、`Not Reviewed=0`。
- E2E 已断言 row 上可见 `Decision: accepted`。

Sprint O 主链路现在由测试覆盖：

```text
Agent Run Detail
-> 记录 Review Decision
-> POST /api/v2/agent-runs/:id/review-decision
-> request body 包含 decision / reason / review_packet_id
-> Operation timeline 出现 review_decision.accepted
-> Audit 页面出现 agent_run.review_decision_recorded
-> Review Queue summary 进入 reviewed
-> Review Queue row 显示 Decision: accepted
```

本轮没有发现越界到：

- Scanner Hub。
- 生命周期治理 / retention。
- 批量操作。
- replay engine。
- 完整 Agent 管理平台。

剩余非阻塞项：

- ESLint 有 1 个 warning：`frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e`。

### 第五次返修复验命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS with 1 warning: src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e

cd frontend-next && npm run build
# PASS after approved run outside sandbox

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 8 passed
```

## 第四次返修复验

结论：仍未通过。

本轮 Windsurf 扩展了 Review Decision E2E，新增覆盖已经能证明：

- `review_packet_id` 被包含在 request body 中。
- 提交后 Agent Run Detail refresh 可显示 `review_decision.accepted` operation。
- operation result 中的 `audit_ref_id` 会渲染 Audit 链接。
- Audit 页面能显示 `agent_run.review_decision_recorded`。

但 Review Queue closure 仍未被真实证明：

1. 新增用例 mock 的 URL 是：

```ts
await page.route('**/api/v2/review-queue**', ...)
```

实际前端 API client 使用的是：

```ts
api.get('/agent-runs/review-queue', params)
```

真实请求路径为：

```text
/api/v2/agent-runs/review-queue
```

因此新增用例里的 `reviewQueueCallCount` mock 不会匹配实际 Review Queue 请求。

2. 新增用例最后只断言：

```ts
await expect(page.getByText('Reviewed', { exact: true })).toBeVisible()
await expect(page.getByText('Test Agent Run')).toBeVisible()
```

`Reviewed` 是 Review Queue summary 的静态标签，不能证明 summary 数值从 `0` 变为 `1`，也不能证明 row 的 `review_state` 变为 `reviewed`。

3. 新增 mock 返回结构缺少 Sprint N/O Review Queue 页面实际依赖的 `summary` 字段；即使 URL 修正，也需要按现有 API 契约返回 `summary.reviewed` / `summary.not_reviewed`，并断言数值变化。

剩余返修要求：

- 把 mock URL 改为：

```ts
await page.route('**/api/v2/agent-runs/review-queue**', ...)
```

- Mock response 必须包含真实契约：

```json
{
  "items": [
    {
      "id": "agent-run-1",
      "review_state": "reviewed",
      "last_review_decision": "accepted"
    }
  ],
  "summary": {
    "total": 1,
    "not_reviewed": 0,
    "reviewed": 1,
    "followup_created": 0,
    "needs_attention": 0
  }
}
```

- E2E 必须等待 `/api/v2/agent-runs/review-queue` request。
- E2E 必须断言 summary text 里 `Reviewed` 对应值为 `1`，`Not Reviewed` 对应值为 `0`。
- E2E 必须断言 row 上可见 `Decision: accepted` 或等价的 reviewed/accepted 状态。

### 第四次返修复验命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS with 1 warning: src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e

cd frontend-next && npm run build
# PASS after approved run outside sandbox

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 8 passed
```

## 第二次返修复验

结论：仍未通过。

本轮复验重点检查上次唯一阻塞项：前端 Review Decision E2E。结果：

- `frontend-next/e2e/agent-runs.spec.ts` 没有变更。
- 当前仍只有 7 个用例。
- 搜索 `review-decision` / `review_decision` / `Review Decision` 在 `agent-runs.spec.ts` 中无命中。
- Playwright 仍只运行旧 7 个用例，全部通过，但没有覆盖 Sprint O 主链路。

因此 Sprint O 仍不能关闭。返修要求不变：必须新增 Review Decision E2E，真实触发 Detail 页 UI，等待并断言 `POST /api/v2/agent-runs/:id/review-decision`，验证 request body、operation timeline、audit 闭环和 Review Queue closure。

### 第二次返修复验命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS with 1 warning: src/app/dashboard/agent-runs/[id]/page.tsx:412:36 unused variable e

cd frontend-next && npm run build
# PASS after approved run outside sandbox

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 7 passed
```

Playwright 说明：

- 使用一次性非交互式 `--reporter=line`。
- 未使用 `npx playwright show-report`。
- 未启动 HTML report 常驻服务。

## 第一次返修复验

结论：仍未通过。

Windsurf 本轮返修关闭了部分后端覆盖缺口：

- `server/v2_api_test.go` 新增 `TestV2RecordReviewDecision`，覆盖成功提交、401、404、invalid decision 400、reason too long 400，并断言 operation / audit 写入。
- `internal/agentrun/review_queue_test.go` 新增 `review_decision.accepted`、`review_decision.false_positive`、`review_decision.needs_manual_followup`、decision 覆盖 follow-up 的 Review Queue closure 测试。
- 相关 Go 测试与全量 Go 测试通过。
- 前端生产构建通过。

但 Sprint O 仍缺最关键的前端 E2E 闭环：

- `frontend-next/e2e/agent-runs.spec.ts` 仍只有 7 个用例。
- 没有新增 Review Decision E2E。
- 搜索 `review-decision` / `review_decision` / `Review Decision` 在 `agent-runs.spec.ts` 中无命中。
- Playwright 通过的是旧 7 个用例，仍不能证明 `POST /api/v2/agent-runs/:id/review-decision` 被真实 UI 触发。

仍需返修：

1. 在 `agent-runs.spec.ts` 增加 Review Decision E2E。
2. 用例必须从 Agent Run Detail 的真实按钮打开 dialog。
3. 选择至少一个 closure decision，例如 `accepted`。
4. 填写 reason。
5. 等待并断言 `POST /api/v2/agent-runs/agent-run-1/review-decision`。
6. 断言 request body 包含 `decision=accepted`、`reason`、`review_packet_id=agent-run-1`。
7. Mock detail refresh 后断言 Operation timeline 出现 `review_decision.accepted`。
8. 断言 Audit 链接或 Audit API query 能到 `agent_run.review_decision_recorded`。
9. 返回或切换到 Review Queue 后，断言 row / summary 从 not reviewed 变为 reviewed。

### 第一次返修复验命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS with 1 warning: src/app/dashboard/agent-runs/[id]/page.tsx:408:36 unused variable e

cd frontend-next && npm run build
# FAIL in sandbox: Turbopack binding to a port denied by sandbox

cd frontend-next && npm run build
# PASS after approved rerun outside sandbox

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# First run failed because Chromium headless shell was missing

cd frontend-next && npx playwright install chromium
# PASS

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
# PASS: 7 passed
```
