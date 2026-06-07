# GODNSLOG 2.0 Sprint P Acceptance

## 验收结论

**通过。**

Sprint P 已实现单个 Agent Run Review Evidence Package 的 Web/API 主体能力，生产构建、Go 测试和 Agent Runs E2E 均通过。第五轮返修后，后端导出契约测试、导出请求的 `review_packet_id` 契约、基础敏感字段断言、Markdown UI 内容断言和 lint warning 均已处理。当前可以关闭 Sprint P。

## 返修复验记录

2026-06-07 Windsurf 返修后复验：

- 已新增 `server/v2_api_test.go::TestV2ExportReviewPackage`，覆盖 JSON / Markdown 成功导出、401、404、invalid format、invalid `review_packet_id`、operation、audit 和基础敏感字段检查。
- 已清理 `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 的 ESLint unused warning。
- Go、lint、build、Playwright 均通过。
- 仍未关闭验收：E2E 的真实请求体仍断言 `review_packet_id=review-packet-1`，与后端 `review_packet_id == agentRunID` 契约冲突；Markdown 导出仍只断言 API 被调用，没有断言 UI 预览内容。

2026-06-07 Windsurf 第二轮返修后复验：

- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` 已改为导出时发送 `review_packet_id: agentRun.id`，对齐后端 `review_packet_id == agentRunID` 契约。
- `frontend-next/e2e/agent-runs.spec.ts` 的导出响应与请求体断言已改为 `agent-run-1`，不再放过真实后端会拒绝的导出请求。
- `server/v2_api_test.go::TestV2ExportReviewPackage` 已扩展敏感关键字检查，包含 `api_key`、`apikey`、`api-key`、`token`、`secret`、`password`、`cookie`、`header`、`private_key`、`private-key`、`credentials`。
- Go、lint、build、Playwright 均通过。
- 仍未关闭验收：Markdown 导出 E2E 仍只断言 `markdownExportCalled === true`，没有断言 UI 中出现 Markdown contract 内容。

2026-06-07 Windsurf 第三轮返修后复验：

- E2E 新增了 `markdownResponseBody` 断言，但该变量由测试代码直接赋值，不来自页面 DOM、textarea、pre/code、copy 区域或真实响应解析后的 UI 展示。
- 该断言只能证明 mock 字符串包含 Markdown 标题，不能证明 Agent Run Detail UI 展示了 Markdown 导出内容。
- E2E 新增 `requestBody` 变量但未使用，`npx eslint ...` 报 1 个 warning。
- Go、build、Playwright 均通过；lint 为 0 errors / 1 warning。
- 仍未关闭验收：必须改为断言页面中可见的 Markdown 导出内容。

2026-06-07 Windsurf 第四轮返修后复验：

- 已删除第三轮新增的 unused `requestBody` warning，lint 恢复 0 errors / 0 warnings。
- E2E 增加了 dialog 内 `pre` 的 Markdown 内容断言。
- 但 E2E 同时保留 fallback：如果 `Export Result` 不可见，会关闭 dialog 并跳转 Audit 页，仅断言 `agent_run.review_exported`。
- 本次 Playwright 用时约 26.9s，比前几轮约 16s 多出约 10s，符合 fallback 等待 20 次 `500ms` 后继续的行为特征；即 Markdown UI 内容断言很可能未被执行。
- Go、lint、build、Playwright 均通过。
- 仍未关闭验收：必须移除 fallback，让 Markdown UI 内容不可见时测试失败。

2026-06-07 Windsurf 第五轮返修后复验：

- E2E 已移除 Audit fallback；`Export Result` 不可见时测试会失败。
- E2E 已断言 dialog 内 `pre` 展示 Markdown 内容，包括 `# Agent Run Review Evidence Package`、`## Agent Run`、`## Evidence Summary`。
- Playwright 用时回到约 16.6s，符合不再进入 fallback 等待分支的表现。
- Go、lint、build、Playwright 均通过。
- Sprint P 验收通过。

## 已确认完成

- `POST /api/v2/agent-runs/:id/review-export` 已接入 v2 Agent Runs 路由。
- `internal/agentrun.ReviewService.ExportReviewPackage` 已支持 `json` / `markdown` 两种格式。
- 导出动作会追加 `review_export.<format>` Agent Operation，`risk_level=low`。
- 导出动作会写入 `agent_run.review_exported` Audit。
- Agent Run Detail 已提供 `Export JSON` / `Export Markdown` 入口。
- E2E 覆盖了 Detail 导出 JSON 后的 Operation timeline 与 Audit 页面闭环。
- 未发现越界实现 PDF/DOCX/ZIP、报告中心、批量导出、生命周期治理、Scanner Hub、replay engine 或高风险动作。

## 阻断问题

### 1. 后端导出 API / Service 测试覆盖

返修状态：**已基本修复**。新增的 `server/v2_api_test.go::TestV2ExportReviewPackage` 已覆盖 JSON / Markdown 成功导出、401、404、invalid format、invalid `review_packet_id`、operation、audit 和基础敏感字段检查。

剩余建议：敏感字段断言仍可再扩展，见问题 4。

### 2. E2E 使用的 `review_packet_id` 与后端契约不一致

后端实现要求：

```go
if req.ReviewPacketID != "" && req.ReviewPacketID != agentRunID {
    return nil, errors.New("review_packet_id must match the current agent run")
}
```

返修状态：**已修复**。前端导出请求已改为发送 `agentRun.id`，E2E 导出响应与请求体断言已改为 `agent-run-1`。虽然测试 fixture 的非导出 timeline 文本中仍有历史 `review-packet-1`，但导出请求体已不再使用该值。

剩余建议：将导出请求体断言从 `toContain('agent-run-1')` 改为 `JSON.parse(capturedBody)` 后精确断言 `format === 'json'`、`review_packet_id === 'agent-run-1'`、`include_audit === true`，避免字符串包含误判。

### 3. Markdown 导出 E2E 验证导出内容

E2E mock 返回了：

```text
# Agent Run Review Evidence Package
```

返修状态：**已修复**。第五轮返修后，测试已移除 Audit fallback，并断言 dialog 内 `pre` 节点展示 Markdown 内容。当前覆盖：

- `# Agent Run Review Evidence Package`
- `## Agent Run`
- `## Evidence Summary`

剩余建议：后续可继续补齐 `## Review Decision`、`## Timeline References`、`## Audit References`、`## Links` 的 UI 断言，但这不再作为 Sprint P 阻断项。

### 4. Sanitization contract 缺少自动化断言

计划要求导出包不得包含 API key、secret、token、password、cookie、authorization header、原始敏感 header/body 等字段。当前未看到后端测试或 E2E 对敏感字段泄露做断言。

返修状态：**已基本修复**。新增后端测试已检查 `api_key`、`apikey`、`api-key`、`token`、`secret`、`password`、`cookie`、`header`、`private_key`、`private-key`、`credentials` 等关键字未出现在 JSON / Markdown 响应中。

剩余建议：后续可以用带敏感 header/body 的 fixture 做更真实的结构化泄露测试，但不再作为本轮阻断。

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

结果：返修前为 0 errors，1 warning；返修复验为 0 errors，0 warnings。

```bash
cd frontend-next && npm run build
```

结果：通过。首次 sandbox 内运行失败于 Turbopack 本地进程/端口权限，已按仓库协作约定使用 escalated rerun。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过，9 passed。命令为一次性非交互式运行，未打开 HTML report。

返修复验补跑结果：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过，0 errors，0 warnings。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过，9 passed。

第四轮返修复验补跑结果：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过，0 errors，0 warnings。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过，9 passed；但当前测试仍包含 Markdown UI 断言 fallback，不能作为该验收点通过依据。

第五轮返修复验补跑结果：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过，0 errors，0 warnings。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过，9 passed。

第三轮返修复验补跑结果：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：0 errors，1 warning：

```text
frontend-next/e2e/agent-runs.spec.ts
  1261:13  warning  'requestBody' is assigned a value but never used  @typescript-eslint/no-unused-vars
```

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过，9 passed。

第二轮返修复验补跑结果：

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：通过，0 errors，0 warnings。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```

结果：通过，9 passed。

## 后续建议

- 将导出请求体断言从 `toContain` 升级为 `JSON.parse(capturedBody)` 后精确断言。
- 补齐 Markdown UI 对 `## Review Decision`、`## Timeline References`、`## Audit References`、`## Links` 的断言。
- 后续可用带敏感 header/body 的 fixture 做更真实的结构化泄露测试。
