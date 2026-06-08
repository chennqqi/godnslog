# GODNSLOG 2.0 Sprint R Acceptance

## 验收结论

**已通过验收。**

Sprint R 已实现 Review Delivery History & Receipt Review 的目标能力：单个 Agent Run 的 webhook delivery attempts 可以通过 read-only API 和 Agent Run Detail UI 回查，历史记录是 sanitized receipt，能区分 delivered / failed / timeout，并能从 Delivery History 回链到 Audit。

本轮返修已关闭此前全部阻断：

- Agent Runs E2E 不再 404。
- Sprint R 核心测试不再 `skip`。
- Timeout history 按真实 Sprint Q 数据形态 `result="failed"` + timeout error 派生为 `timeout`。
- Happy path E2E 已证明 empty history -> delivery -> history refresh -> receipt display -> audit navigation -> `agent_run.review_delivered`。
- `docs/verification.md` 已记录实际 `14 passed`。

## 已确认完成

- 已新增 `AgentRunReviewDeliveryHistoryResponse` / `AgentRunReviewDeliverySummary` / `AgentRunReviewDeliveryHistoryItem`。
- 已新增 `ReviewService.ListReviewDeliveries`。
- 已新增 `GET /api/v2/agent-runs/:id/review-deliveries` 路由和 handler。
- Agent Run Detail 已加载 delivery history，并在 delivery 成功后刷新。
- UI 已展示 Delivery History summary、items、result、destination host、status code、header names、operation/audit refs。
- Go 测试覆盖 delivered / failed / timeout / empty / not found / sanitization 基础场景。
- Timeout 单测已使用真实形态：

```json
{"result":"failed","error":"request timed out"}
```

- `ListReviewDeliveries` 已在 `result == "failed"` 且 error 包含 `timeout` / `timed out` 时派生 `timeout`。
- E2E happy path 使用 call count：第一次 `GET /review-deliveries` 返回 empty，delivery 成功后第二次返回 delivered item。
- E2E 点击 `Deliver to Webhook`、提交 delivery、等待 Delivery Receipt、关闭 dialog、验证 history refresh。
- E2E 点击 history 中的 audit ref 并进入 Audit 页面，验证 `agent_run.review_delivered`。
- E2E failed / timeout path 验证 summary counts、failed/timeout 展示、error summary，以及不存在 retry button。
- E2E 继续验证 full webhook URL 和 header value 不可见。
- 未发现新增 saved connector、notification center、batch delivery、retry queue、report center、Scanner Hub、workflow engine 或 MCP auto-delivery。

## 验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过，0 errors，0 warnings。

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过。

```text
14 passed
```

## 最终判定

Sprint R 达到 package 验收标准，可以进入提交或规划下一轮。
