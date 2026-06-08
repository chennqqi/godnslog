# GODNSLOG 2.0 Sprint R Acceptance

## 验收结论

**未通过验收，需要返修。**

Sprint R 已实现部分后端和前端骨架：新增 delivery history API、模型、服务方法、Agent Run Detail 的 Delivery History UI、以及相关 Go 测试。Go、lint、build 均通过。

但仍有 3 个阻断问题：

1. Sprint R 的两个核心 E2E 被 `test.skip` 跳过，Playwright 实际结果是 `12 passed, 2 skipped`。
2. Timeout history 派生逻辑不符合真实 Sprint Q 数据：真实 timeout operation 会保存 `result: "failed"` 和 timeout error，当前 Sprint R 只在 `result` 为空时派生 `timeout`，会把真实 timeout 统计为 failed。
3. `docs/verification.md` 记录 E2E history 测试通过，但实际测试被 skip，验证记录不可信。

## 已确认完成

- 已新增 `AgentRunReviewDeliveryHistoryResponse` / `AgentRunReviewDeliverySummary` / `AgentRunReviewDeliveryHistoryItem`。
- 已新增 `ReviewService.ListReviewDeliveries`。
- 已新增 `GET /api/v2/agent-runs/:id/review-deliveries` 路由和 handler。
- Agent Run Detail 已加载 delivery history，并在 delivery 成功后刷新。
- UI 已展示 Delivery History summary、items、result、destination host、status code、header names、operation/audit refs。
- Go 测试覆盖 delivered / failed / timeout / empty / not found / sanitization 基础场景。
- 未发现新增 saved connector、notification center、batch delivery、retry queue、report center、Scanner Hub、workflow engine 或 MCP auto-delivery。

## 阻断问题

### 1. 核心 E2E 被 skip

`frontend-next/e2e/agent-runs.spec.ts` 中 Sprint R 两个核心测试被跳过：

```ts
test.skip('should display delivery history with happy path loop', async ({ page }) => {
```

```ts
test.skip('should display failed and timeout delivery history', async ({ page }) => {
```

这两个测试正对应 Sprint R package 的关键验收范围：

- Delivery -> History refresh -> Receipt detail -> Operation/Audit loop。
- failed / timeout display。
- no retry button。
- full webhook URL / header value 不可见。

实际 Playwright 结果：

```text
Running 14 tests using 1 worker
2 skipped
12 passed
```

因此不能认为 Sprint R 的前端闭环已被 E2E 验证。

返修要求：

- 移除 `test.skip`。
- 修正 mock / UI / selector，使两个测试真实运行并通过。
- Playwright 最终结果必须没有 Sprint R 相关 skip。

### 2. 真实 timeout 会被统计为 failed

Sprint Q 的 `createDeliveryFailure` 对 timeout 也写入：

```json
{
  "result": "failed",
  "error": "webhook request timed out"
}
```

Sprint R 当前 `ListReviewDeliveries` 逻辑是：

```go
if res, ok := result["result"].(string); ok {
    item.Result = res
}

if item.Result == "" && item.ErrorSummary != "" {
    if strings.Contains(..., "timeout") || strings.Contains(..., "timed out") {
        item.Result = "timeout"
    }
}
```

由于真实 timeout 已有 `result = failed`，后续派生不会执行，summary 会把 timeout 算到 failed。

当前单测构造 timeout operation 时没有设置 `result`，所以没有覆盖真实数据形态：

```go
timeoutResult := map[string]interface{}{
    "delivery_id": "delivery-789",
    "error":       "request timed out",
}
```

返修要求：

- 当 `result == "failed"` 且 `error` 包含 `timeout` / `timed out` 时，也应派生为 `timeout`。
- 单测必须使用真实形态覆盖：

```json
{"result":"failed","error":"webhook request timed out"}
```

- Summary 中 `timeout` 加 1，`failed` 不加。

### 3. `docs/verification.md` 与实际验证结果不一致

`docs/verification.md` 当前写了：

- `should display delivery history with happy path loop`
- `should display failed and timeout delivery history`
- `No retry button verification`

但这两个 E2E 是 `test.skip`，并没有真实运行。验证文档不能把 skipped test 记录为通过。

返修要求：

- 修复并真实运行 E2E 后，再更新 `docs/verification.md`。
- 若仍有 skip，必须在 verification 中明确记录为 skip，不得写成通过。

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

结果：未达到验收要求。命令完成但存在 skipped tests：

```text
12 passed, 2 skipped
```

## 返修验收标准

- Sprint R 两个 Delivery History E2E 不得 skip，必须真实运行通过。
- Timeout history 必须按真实 Sprint Q operation result 派生为 `timeout`。
- Summary counts 必须正确区分 delivered / failed / timeout。
- `docs/verification.md` 必须与实际验证结果一致。
- 返修后重新运行本文件列出的 Go、lint、build、Playwright 验证命令。
