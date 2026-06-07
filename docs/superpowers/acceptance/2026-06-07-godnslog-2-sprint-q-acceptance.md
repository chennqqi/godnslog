# GODNSLOG 2.0 Sprint Q Acceptance

## 验收结论

**已通过验收。**

Sprint Q 已实现 Review Evidence Delivery Receipt 的目标能力：单个 Agent Run Review Evidence Package 可交付到 webhook，交付结果形成 Delivery Receipt，并闭环到 Agent Operation 与 Audit。上一轮阻断问题已关闭：DNS resolver 失败改为 fail-closed，Delivery UI 已补充 optional header key/value rows，E2E 已覆盖 header 提交路径。

## 已确认完成

- 已新增 `POST /api/v2/agent-runs/:id/review-delivery`。
- `DeliverReviewPackage` 复用 Sprint P `ExportReviewPackage`，支持 `json` / `markdown`。
- delivery 成功写入 `review_delivery.webhook` operation 和 `agent_run.review_delivered` audit。
- webhook 非 2xx / timeout 会写入失败 operation 和 `agent_run.review_delivery_failed` audit。
- webhook payload 中 `refs.delivery_operation_id` 使用真实 Agent Operation ID。
- API handler 将 timeout 映射为 HTTP 504。
- URL 校验覆盖 `https`、localhost、私网字面 IP、link-local、metadata、域名解析到 loopback/private/link-local/metadata IP。
- DNS resolver 失败时 fail-closed，不再在无法确认目的地址安全时放行 webhook。
- Header 后端校验拒绝 `Authorization`、`Cookie`、`Set-Cookie`、`Proxy-*`、hop-by-hop headers，只允许 `Content-Type` 和 `X-*`。
- Agent Run Detail 增加 `Deliver to Webhook` dialog、Delivery Receipt、optional header key/value rows。
- E2E 覆盖 Delivery Receipt、`review_delivery.webhook` Operation timeline、`agent_run.review_delivered` Audit 页面闭环。
- E2E 覆盖 header UI 提交路径和 blocked URL path。
- 未发现越界实现 report center、notification center、saved connector、batch delivery、retry queue、workflow engine、Scanner Hub 或 MCP auto-delivery。

## 返修关闭项

### 1. Resolver fail-open

已关闭。`ValidateWebhookURLWithResolver` 现在对域名执行 resolver；resolver 返回错误时返回 `failed to resolve hostname`，测试用例 `resolver fails - fail-closed` 期望错误。

### 2. Header UI / E2E

已关闭。`frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已添加 header rows，`handleDeliverReview` 会把 header rows 转成 `headers` map 提交。`frontend-next/e2e/agent-runs.spec.ts` 已新增 `should deliver with sanitized headers`，验证 UI 添加 `X-Test-Header` 后 request body 携带 headers。

## 验收备注

- 后端实现已将 operation request 中的 header 信息限制为 `header_names`，audit details 不包含 header values。
- 当前单测覆盖 forbidden headers 和 URL/operation/audit 不泄露 full webhook URL。后续可补一条更直接的单测，显式断言 `X-*` header value 不出现在 operation/audit JSON 中；这不是本轮阻断项。

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

结果：通过，12 passed。命令为一次性非交互式运行，未打开 HTML report。

## 最终判定

Sprint Q 达到 package 验收标准，可以进入提交或规划下一轮。
