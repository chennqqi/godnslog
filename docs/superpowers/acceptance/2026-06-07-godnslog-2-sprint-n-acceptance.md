# GODNSLOG 2.0 Sprint N Acceptance

## 结论

未通过。

Sprint N 的主体实现已经落地：

- 后端新增 Review Queue service / API。
- 后端新增 Follow-up History service / API。
- Agent Runs 页面新增 Review Queue 视图。
- Agent Run Detail 新增 Follow-up History 区块。
- Audit 页面开始读取 `resource_type` / `resource_id` query params。
- E2E 数量从 17 增至 19，新增 Review Queue 和 Follow-up History 用例。

但当前实现仍不能关闭 Sprint N，主要阻塞如下：

1. 前端生产构建失败。
2. Review Queue / Follow-up History 使用的 audit action 名称与 Sprint L/M 真实契约不一致。
3. Audit API 没有真实支持 `resource_id` filter。
4. Review Queue 派生数据存在跨 run 统计错误。
5. E2E 未证明 Sprint N 要求的完整闭环。

## 已完成项

### 1. Review Queue API / Service

已新增：

- `internal/agentrun/review_queue.go`
- `internal/agentrun/review_queue_test.go`
- `GET /api/v2/agent-runs/review-queue`
- `frontend-next/src/app/dashboard/agent-runs/page.tsx` Review Queue 视图

当前支持：

- `review_state`
- `status`
- `evidence_strength`
- `agent_id`
- `case_id`
- `payload_id`
- `page`
- `page_size`

### 2. Follow-up History API / Service

已新增：

- `internal/agentrun/followup_history_test.go`
- `GET /api/v2/agent-runs/:id/followups`
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` Follow-up History 区块

当前会筛选 `followup.*` operation，并解析 reason / review packet id。

### 3. 前端闭环入口

已新增：

- Agent Runs 页面 Review Queue tab。
- Review Queue summary。
- Review Queue filters。
- Detail 页面 Follow-up History。
- Follow-up History audit link。
- Audit 页面 query params 初始化。

## 阻塞问题

### 1. 前端生产构建失败

`npm run build` 失败：

```text
useSearchParams() should be wrapped in a suspense boundary at page "/dashboard/audit".
Error occurred prerendering page "/dashboard/audit".
```

原因：

- `frontend-next/src/app/dashboard/audit/page.tsx` 直接使用 `useSearchParams()`。
- Next 16 要求使用 `useSearchParams()` 的 client subtree 包在 Suspense boundary 内，或拆出 client child component 后由 page 包裹 `<Suspense>`。

返修要求：

- 修复 `/dashboard/audit` 生产构建。
- `cd frontend-next && npm run build` 必须通过。

### 2. Audit action 名称与真实契约不一致

Sprint L/M 的真实 audit action 是：

```text
agent_run.review_generated
agent_run.followup_created
```

现有真实写入位置：

- `internal/agentrun/review.go` 写入 `agent_run.review_generated`
- `internal/agentrun/service.go` 写入 `agent_run.followup_created`

但 Sprint N 新增逻辑使用短名：

- `internal/agentrun/review_queue.go` 查 `review_generated`
- `internal/agentrun/review_queue.go` 查 `followup_created`
- `internal/agentrun/followup_history_test.go` 也插入 `followup_created`

影响：

- Review Queue 无法识别真实 Review Packet audit。
- Follow-up History 无法关联真实 `agent_run.followup_created` audit ref。
- E2E 中 audit ref 是 mock 出来的，不能证明真实后端闭环。

返修要求：

- 统一使用 `agent_run.review_generated`。
- 统一使用 `agent_run.followup_created`。
- 后端测试必须插入真实 action 名称，并断言返回结果。

### 3. Audit API 没有真实支持 `resource_id`

前端 `auditApi.list` 会传：

```ts
resource_type
resource_id
```

但后端 `server/v2_api.go` 的 `v2ListAuditLogs` 只读取：

```go
resourceType := c.Query("resource_type")
```

`internal/auth/service.go` 的 `ListAuditLogs` 也只按 `resource_type` 过滤，没有按 `resource_id` 过滤。

影响：

- Follow-up History 的 audit link 虽然带了 `resource_id=agent-run-1`，但后端不会使用它。
- Sprint N 要求的 Follow-up History -> Audit ref 真实过滤闭环没有成立。

返修要求：

- `v2ListAuditLogs` 读取 `resource_id`。
- `auth.Service.ListAuditLogs` 增加 `resourceID` 参数并过滤 `resource_id`。
- `server/v2_api_test.go` 增加 resource_id filter 测试。
- E2E 点击 audit ref 后必须等待 audit API request，并断言 query 中同时包含 `resource_type=agent_run` 和 `resource_id=agent-run-1`。

### 4. Review Queue 统计跨 run 污染

`internal/agentrun/review_queue.go` 当前计算 operation count 时使用：

```go
opCount, err = s.engine.Table("agent_operations").Count()
```

这会返回全局 operation 总数，不是当前 Agent Run 的 operation count。

影响：

- Review Queue item 的 `operation_count` 对多 Agent Run 场景不可信。
- Summary / row 上的派生数据不能准确代表单个 run。

返修要求：

- `operation_count` 必须按 `agent_run_id = run.ID` 过滤。
- 增加至少两个 Agent Run 的 service 测试，证明 operation count 不会互相污染。

### 5. E2E 覆盖不足

当前新增 E2E 通过，但没有覆盖 Sprint N 的关键验收点：

- Review Queue 用例没有切换 `review_state=needs_attention`，也没有断言 API query 真实变化。
- Summary 只检查静态文字和数字，没有证明 summary 随 filter/mock response 变化。
- Follow-up History 用例没有点击 audit ref 进入 Audit 页面。
- 没有等待 `/api/v2/audit/logs` request。
- 没有断言 audit request 中包含 `resource_type=agent_run` 和 `resource_id=agent-run-1`。
- 没有覆盖“创建新的 follow-up 后，Follow-up History 和 Operation timeline 都刷新”。

返修要求：

1. Review Queue E2E 切换 `review_state=needs_attention`，等待并断言 `/review-queue?review_state=needs_attention`。
2. Mock needs_attention response 改变 summary，断言 summary 随 response 变化。
3. Follow-up History E2E 点击 audit ref，等待 `/api/v2/audit/logs` request，断言 query。
4. 创建 follow-up 后，断言 Follow-up History 和 Operation timeline 都出现新增 action。
5. 不得使用 `test.skip` / `test.only` / 只检查静态文案的空测。

## Scope Boundary

本轮未发现明显越界到：

- Scanner Hub 扩展
- 生命周期治理
- 批量操作
- 真实 replay engine
- 完整 Agent 管理平台

## 验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS

cd frontend-next && npm run build
# FAIL
# /dashboard/audit useSearchParams() requires Suspense boundary

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
# 19 passed (42.2s)
```

Playwright 说明：

- 使用一次性非交互式 `--reporter=line`。
- 未执行 `npx playwright show-report`。
- 当前环境仍采用手动启动 `npm run dev` 后运行 Playwright 的方式。

## 返修清单

1. 修复 `/dashboard/audit` 的 `useSearchParams()` Suspense 构建错误。
2. Review Queue / Follow-up History 统一使用真实 audit action：
   - `agent_run.review_generated`
   - `agent_run.followup_created`
3. Audit API 增加真实 `resource_id` filter，并补后端测试。
4. Review Queue 的 `operation_count` 改为按当前 `agent_run_id` 统计。
5. E2E 补齐 Review Queue filter query、summary 变化、Follow-up History -> Audit API request、创建 follow-up 后 history/timeline 双刷新。
6. `docs/verification.md` 中 Sprint 名称应标为 Sprint N，而不是 Sprint L。

下一轮复验重点只需聚焦上述 6 项，并重新运行：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

---

## 四次返修复验（2026-06-07）

结论：**仍未通过验收**。

本轮重新核对 `frontend-next/e2e/agent-runs.spec.ts` 与 `server/v2_api_test.go` 后，三次返修清单中的核心缺口仍未补齐。验证命令仍全部通过，但当前测试仍不能证明 Sprint N 要求的 Review Queue scope、Audit 回链和 `resource_id` filter 结果闭环。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 仍未通过的硬性验收点

### 1. Review Queue `needs_attention` 未真实进入 API 请求

当前 E2E 仍只执行：

- 进入 `/dashboard/agent-runs`
- 点击 `Review Queue` tab
- 断言 `Review Queue Summary`、`Total`、`Not Reviewed`、`Test Agent Run`

mock 中虽然有 `review_state=needs_attention` 分支，但测试没有点击 `Needs Attention` filter，也没有 `page.waitForRequest` 断言：

```text
/api/v2/agent-runs/review-queue?review_state=needs_attention
```

因此 URL scope 是否真实进入 API 请求仍未成立。

### 2. Review Queue stats 未证明随 scope 变化

mock 中 `needs_attention` 分支会返回：

```text
not_reviewed: 0
needs_attention: 1
```

但测试没有触发该分支，也没有断言 summary 从 Not Reviewed scope 变化到 Needs Attention scope。当前 `getByText('1')` 是歧义断言，无法证明 stats 变化。

### 3. Follow-up History -> Audit 未形成浏览器闭环

当前 Follow-up History E2E 仍只检查 `audit-123` link 的 `href` 包含：

```text
/dashboard/audit
resource_type=agent_run
resource_id=agent-run-1
```

但没有：

- 点击 `audit-123`
- 等待 `/api/v2/audit/logs`
- 断言 request query 包含 `resource_type=agent_run`
- 断言 request query 包含 `resource_id=agent-run-1`
- 断言 Audit 页面显示对应 audit entry

### 4. 后端 `resource_id` filter 仍缺结果断言

`server/v2_api_test.go` 当前只新增了 `case-456` fixture，但测试仍只覆盖：

- list
- action filter
- resource_type filter
- unauthenticated

没有发起：

```text
GET /api/v2/audit/logs?resource_type=case&resource_id=case-123
```

也没有断言只返回 `case-123`、不返回 `case-456`。

## 四次返修清单

1. Review Queue E2E 点击 `Needs Attention` filter。
2. Review Queue E2E 使用 `page.waitForRequest` 断言 `review_state=needs_attention` 进入 API query。
3. Review Queue E2E 断言 summary/stats 随 scope 变化，避免 `getByText('1')` 这类歧义断言。
4. Follow-up History E2E 点击 `audit-123` 并等待 `/api/v2/audit/logs`。
5. Follow-up History -> Audit E2E 断言 `resource_type=agent_run` 与 `resource_id=agent-run-1`。
6. `server/v2_api_test.go` 增加 `resource_id=case-123` filter 结果断言，确认不返回 `case-456`。

---

## 五次返修复验（2026-06-07）

结论：**仍未通过验收**。

本轮复验确认：运行层面继续稳定，但四次返修清单中的硬性覆盖仍未补齐。当前修改没有证明 Review Queue scope 真实进入 API、stats 随 scope 变化、Follow-up History audit ref 到 Audit API 的浏览器闭环，也没有补后端 `resource_id` filter 的结果断言。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 仍未通过的验收点

### 1. Review Queue `needs_attention` scope 仍未被真实触发

`frontend-next/e2e/agent-runs.spec.ts` 当前 Review Queue 用例仍没有点击 `Needs Attention` filter。`needs_attention` 只存在于 mock response 分支中，测试未触发该分支。

缺失：

- `page.getBy...('Needs Attention').click()` 或等价 UI 操作。
- `page.waitForRequest` 捕获 review queue request。
- request URL query 断言 `review_state=needs_attention`。

### 2. stats / summary 仍未证明随 scope 变化

当前测试仍只断言 `Review Queue Summary`、`Total`、`Not Reviewed`、`Reviewed` 和歧义的 `getByText('1')`。这不能证明 summary 从 Not Reviewed scope 变为 Needs Attention scope。

缺失：

- 触发 `needs_attention` 后断言 `not_reviewed` 从 1 变 0。
- 触发 `needs_attention` 后断言 `needs_attention` 从 0 变 1。
- 避免使用单独 `getByText('1')` 作为 stats 断言。

### 3. Follow-up History -> Audit API 仍未形成闭环

当前 Follow-up History 用例仍只读取 `audit-123` 的 `href`，没有点击链接，也没有进入 Audit 页面验证请求。

缺失：

- 点击 `audit-123`。
- 等待 `/api/v2/audit/logs` request。
- 断言 query 包含 `resource_type=agent_run`。
- 断言 query 包含 `resource_id=agent-run-1`。
- 断言页面显示对应 audit entry。

### 4. 后端 `resource_id` filter 仍缺结果断言

`server/v2_api_test.go` 仍只创建了 `case-456` fixture，但未发起 `resource_id=case-123` 查询，也未验证结果集。

缺失：

```text
GET /api/v2/audit/logs?resource_type=case&resource_id=case-123
```

并断言：

- 返回 total/items 只包含 `case-123`。
- 不返回 `case-456`。

## 五次返修清单

1. Review Queue E2E 必须点击 `Needs Attention` filter。
2. Review Queue E2E 必须 `waitForRequest` 并断言 `review_state=needs_attention`。
3. Review Queue E2E 必须用非歧义方式断言 stats 随 scope 变化。
4. Follow-up History E2E 必须点击 `audit-123` 并等待 `/api/v2/audit/logs`。
5. Follow-up History -> Audit E2E 必须断言 `resource_type=agent_run` 和 `resource_id=agent-run-1`。
6. `server/v2_api_test.go` 必须补 `resource_id=case-123` 结果过滤断言，并确认不返回 `case-456`。

---

## 六次返修复验（2026-06-07）

结论：**仍未通过验收**。

本轮有实质进展：Review Queue E2E 已开始点击 `Needs Attention`，并使用 `page.waitForRequest` 捕获 `review_state=needs_attention` 请求；Follow-up History E2E 也开始点击 `audit-123` 并跳转到 Audit 页面。但 Audit API 请求闭环和后端 `resource_id` filter 结果测试仍未补齐，Review Queue stats 断言仍偏弱。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 已修复 / 有进展

### 1. Review Queue API scope 已被 E2E 触发

`frontend-next/e2e/agent-runs.spec.ts` 当前已新增：

- 点击 Review Queue tab。
- 点击 `All States`。
- 选择 `Needs Attention`。
- `page.waitForRequest` 捕获 `/api/v2/agent-runs/review-queue`。
- 断言 URL 包含 `review_state=needs_attention`。

这一项可从返修清单移除。

### 2. Follow-up History audit ref 已开始点击

Follow-up History E2E 当前已点击 `audit-123`，并断言 URL 进入 `/dashboard/audit`。

## 仍未通过的验收点

### 1. Review Queue stats 变化断言仍不足

当前测试只在切换后断言 `summarySection.getByText('Needs Attention')` 可见。它没有明确断言：

- `not_reviewed` 从 1 变 0。
- `needs_attention` 从 0 变 1。
- summary 数值对应切换后的 mock response。

因此“stats 是否随 scope 变化”仍未充分证明。

返修要求：

- 使用明确的 summary card / label 容器断言 Needs Attention 数值为 1。
- 同时断言 Not Reviewed 数值为 0，或用更结构化 locator 区分 label 与 value。

### 2. Follow-up History -> Audit API 请求仍未被 E2E 等待和断言

当前测试点击 audit link 后只断言 URL 到 `/dashboard/audit`。没有：

- `page.waitForRequest` 捕获 `/api/v2/audit/logs`。
- 断言 request query 包含 `resource_type=agent_run`。
- 断言 request query 包含 `resource_id=agent-run-1`。
- mock audit logs response。
- 断言页面显示 `agent_run.followup_created` 或对应 audit entry。

本轮 E2E 虽然 19 passed，但 dev server 日志出现：

```text
[browser] Failed to load audit log: AxiosError: Network Error
```

这说明测试没有覆盖 Audit API 成功闭环，只覆盖了页面跳转。

### 3. 后端 `resource_id` filter 仍缺结果断言

`server/v2_api_test.go` 当前仍未发起：

```text
GET /api/v2/audit/logs?resource_type=case&resource_id=case-123
```

也没有创建 `case-123` / `case-456` 两条可区分 fixture 并断言只返回 `case-123`。当前 Audit Logs 测试仍只覆盖 list、action filter、resource_type filter、unauthenticated。

## 六次返修清单

1. Review Queue E2E 补明确的 stats 数值变化断言，证明 `needs_attention=1` 且 `not_reviewed=0`。
2. Follow-up History E2E 点击 `audit-123` 时必须等待 `/api/v2/audit/logs` request。
3. Follow-up History -> Audit E2E 必须断言 audit request query 包含 `resource_type=agent_run` 与 `resource_id=agent-run-1`。
4. Follow-up History -> Audit E2E 必须 mock audit logs response 并断言页面显示对应 audit entry。
5. `server/v2_api_test.go` 必须补 `resource_id=case-123` 结果过滤断言，并确认不返回 `case-456`。

---

## 七次返修复验（2026-06-07）

结论：**仍未通过验收**。

本轮继续有进展：后端测试新增了 `case-123` / `case-456` fixture；Follow-up History E2E 新增了 audit logs route mock 和 `waitForRequest`；但实现仍没有证明两个关键闭环：

- 后端 `resource_id` filter 仍未通过 API response 结果断言验证。
- 前端 Audit 页面仍实际报 `Network Error`，E2E 没有捕获失败。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 已修复 / 有进展

### 1. Audit E2E 已开始等待 audit logs request

`agent-runs.spec.ts` 当前已新增：

- `page.route('**/api/v2/audit/logs', ...)`
- 点击 `audit-123`
- `page.waitForRequest` 等待 `/api/v2/audit/logs`
- request URL 条件包含 `resource_type=agent_run` 和 `resource_id=agent-run-1`

### 2. 后端测试已新增 case-123 / case-456 fixture

`server/v2_api_test.go` 当前会插入不同 `resource_id` 的 audit log fixture。

## 仍未通过的验收点

### 1. Audit E2E 仍未证明页面成功加载 audit entry

E2E 虽然 19 passed，但 dev server 日志仍出现：

```text
[browser] Failed to load audit log: AxiosError: Network Error
```

当前测试没有断言 Audit 页面显示 `agent_run.followup_created` 或 `audit-123`，所以即使 Audit API 加载失败也能通过。

另外当前断言：

```ts
expect(consoleErrors).not.toContain('Network Error')
```

无法捕获包含 `Failed to load audit log: AxiosError: Network Error` 的长字符串，因为 `toContain` 对数组检查的是完整元素匹配，不是子串匹配。

返修要求：

- 修正 route pattern，确保带 query 的 `/api/v2/audit/logs?...` 被 mock 命中。
- 使用 `consoleErrors.some(error => error.includes('Network Error'))` 断言没有网络错误。
- 断言 Audit 页面显示 `agent_run.followup_created` 或 `audit-123`。

### 2. 后端 `resource_id` filter 仍未通过 API response 验证

当前 `server/v2_api_test.go` 对 `resource_id` 的检查是直接 SQL：

```go
SELECT COUNT(*) FROM audit_logs WHERE resource_id = ?
```

这只能证明 fixture 插入成功，不能证明：

```text
GET /api/v2/audit/logs?resource_type=case&resource_id=case-123
```

会只返回 `case-123`、不返回 `case-456`。

返修要求：

- 发起真实 HTTP request：
  - `/api/v2/audit/logs?resource_type=case&resource_id=case-123`
- 解析 API response。
- 断言 `items` 只包含 `resource_id=case-123`。
- 断言不包含 `case-456`。

### 3. Review Queue stats 断言仍偏弱

当前 Review Queue E2E 仍只断言 summary text 包含 `Needs Attention`，没有明确断言数值变化。

返修要求：

- 明确断言 Needs Attention summary 数值为 `1`。
- 明确断言 Not Reviewed summary 数值为 `0`。

## 七次返修清单

1. Audit E2E 修正 network error 捕获：使用子串检查，而不是 `not.toContain('Network Error')`。
2. Audit E2E 必须断言页面显示 `agent_run.followup_created` 或 `audit-123`。
3. Audit E2E 确保 `/api/v2/audit/logs?...` route mock 命中，运行时不再出现 `Failed to load audit log`。
4. `server/v2_api_test.go` 必须通过真实 HTTP API request 验证 `resource_id=case-123` 过滤结果。
5. 后端测试必须断言 API response 不包含 `case-456`。
6. Review Queue E2E 补 `needs_attention=1` 和 `not_reviewed=0` 的明确数值断言。

---

## 八次返修复验（2026-06-07）

结论：**仍未通过验收，但剩余问题已收敛到前端 E2E 断言质量**。

本轮修复解决了两个关键问题：

- 后端 `resource_id` filter 已通过真实 HTTP API response 验证。
- Audit E2E route mock 已能命中带 query 的 `/api/v2/audit/logs?...`，本轮 E2E 运行日志未再出现 `Failed to load audit log: AxiosError: Network Error`。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 已修复 / 可关闭

### 1. 后端 `resource_id` filter API response 测试已补齐

`server/v2_api_test.go` 当前已通过真实 HTTP request 验证：

```text
GET /api/v2/audit/logs?resource_id=case-123
GET /api/v2/audit/logs?resource_id=case-456
```

并解析 API response，断言：

- `case-123` 查询返回的 items 均为 `resource_id=case-123`。
- `case-456` 查询返回的 items 均为 `resource_id=case-456`。
- `case-123` 查询结果不包含 `case-456`。

### 2. Audit API mock 已命中，运行日志无 Network Error

`agent-runs.spec.ts` 的 audit route pattern 已改为：

```ts
**/api/v2/audit/logs**
```

本轮 E2E dev server 日志未再出现：

```text
Failed to load audit log: AxiosError: Network Error
```

## 仍未通过的验收点

### 1. Audit 页面仍未断言显示 audit entry

当前 Follow-up History -> Audit E2E 已等待 audit request，也 mock 了 `agent_run.followup_created`，但没有断言 Audit 页面实际显示：

- `agent_run.followup_created`
- 或 `audit-123`
- 或对应 `resource_id=agent-run-1` 的 audit row

因此浏览器端“点击 audit ref -> Audit 页面展示对应 audit entry”的闭环仍未完全证明。

返修要求：

- 在点击 `audit-123` 并等待 `/api/v2/audit/logs` 后，断言页面出现 `agent_run.followup_created` 或 `audit-123`。

### 2. Review Queue stats 数值变化仍未明确断言

当前 Review Queue E2E 已证明 `review_state=needs_attention` 进入 API query，但仍只断言 summary 中出现 `Needs Attention` 文案。没有明确断言：

- `needs_attention = 1`
- `not_reviewed = 0`

因此“stats 是否随 scope 变化”仍未完全满足验收标准。

返修要求：

- 使用稳定 locator 找到 Needs Attention summary card / row，断言数值为 `1`。
- 使用稳定 locator 找到 Not Reviewed summary card / row，断言数值为 `0`。

### 3. Network Error 捕获断言仍建议修正

当前仍是：

```ts
expect(consoleErrors).not.toContain('Network Error')
```

这对数组执行完整元素匹配，不能捕获包含 `Network Error` 的长错误字符串。

返修要求：

```ts
expect(consoleErrors.some(error => error.includes('Network Error'))).toBe(false)
```

## 八次返修清单

1. Follow-up History -> Audit E2E 断言页面显示 `agent_run.followup_created` 或 `audit-123`。
2. Review Queue E2E 明确断言 `needs_attention=1`。
3. Review Queue E2E 明确断言 `not_reviewed=0`。
4. 修正 Network Error 捕获断言为子串匹配。

---

## 九次返修复验（2026-06-07）

结论：**通过验收**。

本轮返修补齐了八次返修清单中的剩余前端 E2E 断言，Sprint N 的 Review Queue scope、stats 变化、Follow-up History -> Audit API 回链、后端 `resource_id` filter 均已形成可验证闭环。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 验收点核对

### 1. URL scope 真实进入 API 请求

通过。Review Queue E2E 点击 `Needs Attention` filter，并使用 `page.waitForRequest` 断言 request URL 包含：

```text
review_state=needs_attention
```

### 2. stats 随 scope 变化

通过。Review Queue E2E 在 `needs_attention` scope 下断言 summary text：

```text
1.*Needs Attention
0.*Not Reviewed
```

### 3. Follow-up History -> Audit 闭环

通过。Follow-up History E2E：

- 点击 `audit-123`。
- 等待 `/api/v2/audit/logs` request。
- 断言 request query 包含 `resource_type=agent_run` 和 `resource_id=agent-run-1`。
- mock 返回 `agent_run.followup_created`。
- 断言 Audit 页面显示 `agent_run.followup_created`。

### 4. Network Error 捕获

通过。E2E 已将 Network Error 检查改为子串匹配：

```ts
consoleErrors.some(error => error.includes('Network Error'))
```

本轮 dev server 日志未再出现 `Failed to load audit log: AxiosError: Network Error`。

### 5. 后端 `resource_id` filter

通过。`server/v2_api_test.go` 已通过真实 HTTP API request 验证：

```text
GET /api/v2/audit/logs?resource_id=case-123
GET /api/v2/audit/logs?resource_id=case-456
```

并断言 `case-123` 查询结果不包含 `case-456`。

### 6. 无越界

通过。复验范围仍限定在 Review Queue、Follow-up History、Audit filter、Agent Runs E2E 和相关后端 audit filter 测试；未发现越界到 Scanner Hub、生命周期治理或批量操作。

## 最终结论

Sprint N 可以验收通过。

---

## 十次返修复验（2026-06-07）

结论：**通过验收**。

本轮 Windsurf 继续补强前端 E2E 断言后，Sprint N 验收状态维持通过。剩余前端断言质量问题已补齐：

- Review Queue E2E 明确断言 `needs_attention=1`。
- Review Queue E2E 明确断言 `not_reviewed=0`。
- Follow-up History -> Audit E2E 断言 Audit 页面显示 `agent_run.followup_created`。
- Network Error 检查已改为子串匹配。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 最终核对

- URL scope 真实进入 API 请求：通过。
- stats 随 scope 变化：通过。
- Follow-up History -> Audit API -> Audit 页面展示闭环：通过。
- Evidence / Audit 契约未破坏：通过。
- E2E 无 skip / 无纯静态空测：通过。
- 未越界到 Scanner Hub、生命周期治理或批量操作：通过。

Sprint N 维持验收通过。

---

## 二次返修复验（2026-06-07）

结论：**仍未通过验收**。

这轮返修后，命令层面的稳定性已经恢复，Follow-up 创建后的 History / Timeline 双刷新 E2E 也已补强；但 Sprint N 的两个关键闭环仍只停留在 mock / href 层面，没有被 E2E 和后端测试真实证明。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内首次运行因 Turbopack 需要绑定端口触发 `Operation not permitted`，在非沙箱环境重跑同一命令后通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 已修复 / 已改善

### 1. 生产构建已通过

`frontend-next/src/app/dashboard/audit/page.tsx` 已拆出 `audit-page-content.tsx`，解决 `useSearchParams()` Suspense 相关构建问题。生产构建通过，Audit 页面可正常预渲染。

### 2. 创建 follow-up 后 History / Timeline 双刷新已补强

`frontend-next/e2e/agent-runs.spec.ts` 的 follow-up 创建用例现在会等待：

- `/api/v2/agent-runs/agent-run-1` detail refresh
- `/api/v2/agent-runs/agent-run-1/followups` history refresh

并断言：

- Operation timeline 出现 `followup.recheck_evidence`
- `Follow-up History (1)` 出现
- Follow-up action 与 reason 出现

这一项可从剩余返修清单移除。

## 仍未通过的验收点

### 1. Review Queue E2E 仍未真实进入 `needs_attention` scope

当前 `agent-runs.spec.ts` mock 中已经准备了 `review_state=needs_attention` 分支，并会返回不同 summary：

```text
not_reviewed: 0
needs_attention: 1
```

但测试流程仍只点击 `Review Queue` tab，没有操作 UI 切换到 `Needs Attention`，也没有等待或断言真实请求：

```text
/api/v2/agent-runs/review-queue?review_state=needs_attention
```

因此“URL scope 是否真实进入 API 请求”和“stats 是否随 scope 变化”仍未被 E2E 证明。

返修要求：

- 在 Review Queue E2E 中点击 `Needs Attention` filter。
- 使用 `page.waitForRequest` 捕获 review queue API。
- 断言 request query 包含 `review_state=needs_attention`。
- 断言 summary 从 `not_reviewed=1 / needs_attention=0` 变化到 `not_reviewed=0 / needs_attention=1`。

### 2. Follow-up History -> Audit 仍未形成真实 API 闭环

当前 Follow-up History E2E 仍只检查 audit ref link 的 href：

```text
/dashboard/audit
resource_type=agent_run
resource_id=agent-run-1
```

但没有点击 audit ref，没有进入 Audit 页面，也没有等待：

```text
/api/v2/audit/logs
```

因此还不能证明 “Payload / Agent Run Detail -> Interactions / Follow-up History -> Evidence / Audit” 的回链在浏览器行为层面成立。

返修要求：

- 点击 `audit-123` link。
- 等待 `/api/v2/audit/logs` request。
- 断言 query 同时包含：
  - `resource_type=agent_run`
  - `resource_id=agent-run-1`
- 断言 Audit 页面显示对应 `agent_run.followup_created` audit entry。

### 3. `resource_id` 后端测试仍不足

`server/v2_api_test.go` 当前 Audit Logs 测试仍只覆盖：

- 普通 list
- action filter
- resource_type filter
- unauthenticated

虽然 fixture 已创建不同 `resource_id` 的 audit log，但测试没有发起：

```text
GET /api/v2/audit/logs?resource_type=case&resource_id=case-123
```

也没有断言只返回 `case-123`、不返回 `case-456`。

返修要求：

- 增加 `resource_id` filter 结果断言。
- 同时断言 `total/items` 数量和 item 的 `resource_id`。

### 4. `docs/verification.md` 仍有 Sprint 名称残留

`docs/verification.md` 的 Sprint N 段落末尾仍写着：

```text
Sprint L establishes the Review Queue and Follow-up History system:
```

应改为 Sprint N，避免验证文档和实际 Sprint 对不上。

## 二次返修清单

1. Review Queue E2E 必须真实触发 `review_state=needs_attention` request。
2. Review Queue E2E 必须断言 summary/stats 随 scope 响应变化。
3. Follow-up History audit ref E2E 必须点击进入 Audit，并等待 `/api/v2/audit/logs`。
4. Audit request E2E 必须断言 `resource_type=agent_run` 与 `resource_id=agent-run-1`。
5. `server/v2_api_test.go` 必须补 `resource_id` filter 结果断言。
6. `docs/verification.md` 必须修正 Sprint L 残留文案。

---

## 三次返修复验（2026-06-07）

结论：**仍未通过验收**。

本轮返修修正了 `docs/verification.md` 中 Sprint 名称残留，但 Review Queue scope、Audit 回链和后端 `resource_id` filter 结果测试仍未补齐。命令层面继续通过，问题仍是验收覆盖不足。

## 本轮验证命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

说明：沙箱内仍会因 Turbopack 绑定端口权限触发 `Operation not permitted`，非沙箱重跑同一命令通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```

结果：19 passed。

## 本轮已修复

### `docs/verification.md` Sprint 名称残留已修正

`docs/verification.md` 当前已显示：

```text
Sprint N establishes the Review Queue and Follow-up History system:
```

这一项可从返修清单移除。

## 仍未通过的验收点

### 1. Review Queue `needs_attention` scope 仍未被 E2E 真实触发

`frontend-next/e2e/agent-runs.spec.ts` 中仍只点击 `Review Queue` tab，并断言静态 summary / item 文案。虽然 mock route 支持 `review_state=needs_attention` 分支，但测试没有：

- 点击 `Needs Attention` filter。
- 使用 `page.waitForRequest` 捕获 review queue request。
- 断言 request query 包含 `review_state=needs_attention`。
- 断言 summary 从 Not Reviewed scope 变化到 Needs Attention scope。

因此“URL scope 是否真实进入 API 请求”和“stats 是否随 scope 变化”仍未通过。

### 2. Follow-up History -> Audit API 回链仍未被 E2E 证明

Follow-up History 用例仍只读取 `audit-123` link 的 `href`，没有点击 link，也没有等待 `/api/v2/audit/logs` request。

缺失断言仍是：

- 点击 audit ref 进入 `/dashboard/audit`。
- 捕获 `/api/v2/audit/logs` request。
- 断言 query 包含 `resource_type=agent_run`。
- 断言 query 包含 `resource_id=agent-run-1`。
- 断言 Audit 页面显示对应 audit entry。

### 3. 后端 Audit Logs `resource_id` filter 仍缺结果测试

`server/v2_api_test.go` 当前仍只覆盖 list、action filter、resource_type filter 和 unauthenticated。没有新增：

```text
GET /api/v2/audit/logs?resource_type=case&resource_id=case-123
```

也没有断言只返回 `case-123`、不返回其他 resource。

## 三次返修清单

1. Review Queue E2E 补真实 `needs_attention` filter 操作和 request query 断言。
2. Review Queue E2E 补 summary/stats 随 scope 变化的断言。
3. Follow-up History E2E 点击 audit ref，并等待 `/api/v2/audit/logs` request。
4. Follow-up History -> Audit E2E 断言 `resource_type=agent_run` 与 `resource_id=agent-run-1`。
5. `server/v2_api_test.go` 补 `resource_id` filter 的结果断言。

## 返修复验（2026-06-07）

结论：仍未通过。

本轮返修已修复 4 个主要实现问题：

1. `/dashboard/audit` 已拆为 Suspense wrapper + client content，生产构建恢复通过。
2. Review Queue / Follow-up History 已改用真实 audit action：
   - `agent_run.review_generated`
   - `agent_run.followup_created`
3. Audit API 后端已读取并传递 `resource_id`，`internal/auth.Service.ListAuditLogs` 也已按 `resource_id` 过滤。
4. Review Queue 的 `operation_count` 已改为按 `agent_run_id = run.ID` 统计。

本轮验证命令：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
# PASS

GOCACHE=/tmp/gocache go test ./...
# PASS

cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
# PASS

cd frontend-next && npm run build
# PASS

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
# 初次失败：Playwright Chromium headless shell 缺失

cd frontend-next && npx playwright install chromium
# PASS，安装 chromium / ffmpeg / chromium_headless_shell

cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
# 19 passed (42.4s)
```

Playwright 说明：

- 使用一次性非交互式 `--reporter=line`。
- 未执行 `npx playwright show-report`。
- 当前环境仍采用手动启动 `npm run dev` 后运行 Playwright 的方式。

仍未通过项：

### 1. E2E 仍未证明 Review Queue filter query

`frontend-next/e2e/agent-runs.spec.ts` 的 mock 已支持：

```ts
review_state=needs_attention
```

但测试没有实际操作 UI 切换到 `Needs Attention`，也没有使用 `page.waitForRequest` 断言 `/api/v2/agent-runs/review-queue?review_state=needs_attention`。

当前只是进入 Review Queue tab 并断言静态 summary / item 文案，不能证明 filter query 真实进入 API。

返修要求：

- 在 E2E 中切换 Review State filter 到 `Needs Attention`。
- 使用 `page.waitForRequest` 捕获 review queue request。
- 断言 request URL query 包含 `review_state=needs_attention`。

### 2. E2E 仍未证明 summary 随 response 变化

mock 已为 `needs_attention` 准备了不同 summary：

```ts
needs_attention: 1
not_reviewed: 0
```

但测试没有触发该分支，也没有断言 UI 从 Not Reviewed summary 变化到 Needs Attention summary。

返修要求：

- 触发 `needs_attention` filter 后断言 summary 更新。
- 不要只检查 `Review Queue Summary` / `Total` / `1` 这类静态或歧义文本。

### 3. E2E 仍未证明 Follow-up History -> Audit API 回链

当前 Follow-up History 用例只检查 audit link 的 `href` 包含：

```text
resource_type=agent_run
resource_id=agent-run-1
```

但没有点击 link，没有进入 Audit 页面，没有等待 `/api/v2/audit/logs` 请求，也没有断言请求 query。

返修要求：

- 点击 audit ref。
- 等待 `/api/v2/audit/logs` request。
- 断言 query 同时包含：
  - `resource_type=agent_run`
  - `resource_id=agent-run-1`
- 断言页面显示 `agent_run.followup_created` 或对应 audit entry。

### 4. E2E 仍未证明创建 follow-up 后 History 和 Timeline 双刷新

Sprint N 要求：

- 创建新的 follow-up 后，Follow-up History 和 Operation timeline 都刷新。

当前 follow-up 创建用例仍只证明 Operation timeline 出现 `followup.recheck_evidence`，Follow-up History 用例是独立 mock 页面加载，不证明创建后的刷新。

返修要求：

- 在创建 follow-up 的同一条 E2E 中，让 `/agent-runs/:id/followups` mock 在 POST 后返回新增 history。
- 断言：
  - Follow-up History count / item 更新。
  - Operation timeline 出现新增 `followup.recheck_evidence`。

### 5. `resource_id` 后端测试仍不足

后端实现已支持 `resource_id`，但 `server/v2_api_test.go` 当前看到的 Audit Logs 测试仍只覆盖：

- list
- action filter
- resource_type filter
- unauthenticated

没有断言 `resource_id=case-123` 只返回对应资源，不返回 `case-456`。

返修要求：

- 增加 `GET /api/v2/audit/logs?resource_type=case&resource_id=case-123` 测试。
- 断言 total/items 只包含 `case-123`。

## 返修清单（剩余）

1. Agent Runs E2E 补 `review_state=needs_attention` filter request 断言。
2. Agent Runs E2E 补 summary 随 mock response 变化的断言。
3. Agent Runs E2E 补 Follow-up History audit ref 点击到 Audit API 的真实请求断言。
4. Agent Runs E2E 补创建 follow-up 后 Follow-up History 和 Operation timeline 双刷新。
5. `server/v2_api_test.go` 补 `resource_id` filter 的结果断言。
6. `docs/verification.md` 中 Sprint 标题仍显示 `Sprint L: Review Queue & Follow-up History`，应改为 Sprint N。

下一轮复验继续运行：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/evidence.spec.ts
```
