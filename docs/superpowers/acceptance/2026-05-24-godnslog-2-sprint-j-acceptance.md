# GODNSLOG 2.0 Sprint J 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-j-package.md`
- `internal/models/agent_run.go`
- `internal/agentrun/service.go`
- `internal/agentrun/service_test.go`
- `internal/agentrun/migration.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `server/v2_api.go`
- `db/init.go`
- `frontend-next/src/app/dashboard/agent-runs/page.tsx`
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/agent-runs.spec.ts`
- `docs/verification.md`

## 验收结论

**最终结论：Sprint J 通过验收，可以关闭。**

最新返修已补齐 Agent Runs E2E 的真实链路断言，并让 `wait_for_interaction` poll 失败路径写入 failed operation。后端聚焦测试、后端全量测试、前端生产构建、目标 ESLint 和关键 E2E 均已复验通过。下方保留前两轮未通过记录，作为返修过程追踪。

## 最终复验结论

**复验日期：2026-05-24**

**结论：Sprint J 通过验收，可以关闭。**

本轮返修补齐了上一轮剩余的两个阻塞点：

1. `frontend-next/e2e/agent-runs.spec.ts` 已从 URL 空测升级为真实链路断言：
   - 断言 `GET /api/v2/agent-runs` 被调用
   - 断言列表数据展示
   - 断言 status filter 后结果变化
   - 断言 detail 数据展示
   - 断言 operation timeline 展示 `create_oast_probe` / `wait_for_interaction`
   - 断言 Interactions / Evidence 回链携带 `payload_id`
   - 断言 Case / Payload detail 链接
   - 断言 status update 触发 `PUT /agent-runs/:id/status`

2. `internal/mcp/server.go` 的 `wait_for_interaction` poll 失败路径已补充 failed operation 写入：
   - poll 失败后先尝试将 Agent Run 更新为 `failed`
   - 即使 failed status update 失败，也会继续尝试追加 `wait_for_interaction` failed operation
   - failed operation 记录 `success: false` 和 `error`

### 最终复验已执行

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

结果：

- `go test ./internal/agentrun ./internal/mcp ./server` 通过
- `go test ./...` 通过
- 目标 ESLint：0 errors，1 warning
  - warning：`frontend-next/e2e/agent-runs.spec.ts` 未使用 `Page` import
  - 不影响验收关闭
- `npm run build` 通过
- Playwright：`26 passed (58.3s)`
- `agent-runs.spec.ts` / `interactions.spec.ts` / `evidence.spec.ts` 未发现 `test.skip` / `test.only`
- 已停止本次 E2E 使用的 Next.js dev server

### 剩余非阻塞项

- `frontend-next/e2e/agent-runs.spec.ts` 有未使用 `Page` import warning，可在后续清理。
- `wait_for_interaction` 失败路径已经尝试写 failed operation；如果 failed operation 写入 API 自身也失败，当前只能日志化。考虑到工具已处于失败返回路径，且 Sprint J 要求的失败 operation 主链路已补齐，本项不阻塞 Sprint J 关闭。

## 初轮验收结论

**结论：Sprint J 未通过验收，需要返修。**

本轮实现已经补入 Agent Run / Agent Operation 数据模型、基础 service、`/api/v2/agent-runs` API、MCP 侧真实 Agent Run 创建调用、Agent Runs 前端列表和详情页。前端生产构建可通过。

但 Sprint J 的完成定义要求 MCP 绑定真实 Agent Run、operation/audit 失败不能静默吞掉、后端测试通过、E2E 无 skip 且覆盖真实链路。当前存在后端测试失败、目标 ESLint 失败、Agent Runs E2E 整体 skip、MCP operation/status 失败被日志吞掉等阻塞问题，因此不能关闭 Sprint J。

## 返修复验结论

**复验日期：2026-05-24**

**结论：Sprint J 仍未通过验收，需要继续返修。**

Windsurf 本轮返修已经解决多项工程门禁问题：

- `internal/mcp/server_test.go` 已适配 `/api/v2/agent-runs`
- `GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server` 通过
- `GOCACHE=/tmp/gocache go test ./...` 通过
- `UpdateAgentRunStatus` 已正确保存旧状态用于 audit `from_status`
- `AppendAgentOperation` 已对不存在的 Agent Run 返回错误
- MCP 成功路径上的 append operation / status update 失败已改为返回错误
- `frontend-next/e2e/agent-runs.spec.ts` 已取消 `test.skip`
- `cd frontend-next && npm run build` 在非沙箱环境通过
- 目标 ESLint 无错误，仅剩 1 个未使用 import warning
- `cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts` 通过，`25 passed`

但仍有 2 个阻塞问题：

1. **Agent Runs E2E 仍是静态空测**
   - `frontend-next/e2e/agent-runs.spec.ts` 只有两个测试：
     - `should display agent runs list page`
     - `should display agent run detail page`
   - 两个测试只断言 `page.url()` 包含目标路径
   - 未断言 `GET /api/v2/agent-runs`
   - 未断言 `agent_id` / `status` 过滤参数进入 API query
   - 未断言 `GET /api/v2/agent-runs/:id`
   - 未展示和断言 operation timeline 的真实数据
   - 未断言 Interactions / Evidence 链接携带 `payload_id`
   - `docs/verification.md` 也承认当前是 “Basic page load verification”

2. **MCP 失败路径仍未完整形成 operation 或 audit 闭环**
   - `wait_for_interaction` 的 poll 失败路径尝试把 Agent Run 更新为 `failed`
   - 如果该 status update 失败，当前仍只 `log.Printf("Failed to update agent run status to failed: %v", err2)`
   - Sprint J plan 要求 “MCP 工具失败也要写入 operation 或 audit”
   - 当前失败路径没有可靠写入失败 operation，也没有返回 status update 失败信息

因此，本轮返修已经把测试门禁从失败推进到通过，但产品完成定义中的 E2E 真实性和 Agent 失败审计闭环仍未达标。

### 返修复验已执行

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

结果：

- `go test ./internal/agentrun ./internal/mcp ./server` 通过
- `go test ./...` 通过
- 目标 ESLint：0 errors，1 warning
- `npm run build` 通过
- Playwright：`25 passed (57.0s)`
- 已停止本次 E2E 使用的 Next.js dev server

## 已执行验证

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
```

结果：失败。

- `internal/agentrun` 通过
- `server` 通过
- `internal/mcp` 失败：
  - `TestCreateOASTProbeToolCreatesCaseThenPayload`
  - `unexpected request POST /api/v2/agent-runs`

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：失败。

- 失败点同样在 `internal/mcp`
- `TestCreateOASTProbeToolCreatesCaseThenPayload` 未适配 Sprint J 新增的 `/api/v2/agent-runs` 调用

```bash
cd frontend-next && npm run build
```

结果：通过。

备注：沙箱内首次因 Turbopack 创建进程和绑定端口被拒绝失败；按验收需要使用非沙箱权限复跑后通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
```

结果：失败。

- `frontend-next/src/app/dashboard/agent-runs/page.tsx:55`
- `react-hooks/set-state-in-effect`
- `loadAgentRuns()` 在 effect 内同步触发 setState

```bash
rg -n "test\\.skip|describe\\.skip|\\.only\\(|test\\.only|it\\.only|skip\\(" frontend-next/e2e
```

结果：发现 Sprint J 新增 E2E 整体 skip。

- `frontend-next/e2e/agent-runs.spec.ts:4`
- `test.skip(true, 'Requires authentication setup - pages redirect to /login without token')`

## 通过项

### 1. Agent Run 基础模型和 schema sync

部分通过。

已新增：

- `internal/models/agent_run.go`
- `AgentRun`
- `AgentOperation`
- `internal/agentrun/migration.go`
- `db/init.go` 中同步 `AgentRun` / `AgentOperation`

### 2. Agent Run API

部分通过。

`server/v2_api.go` 已注册：

```text
GET /api/v2/agent-runs
POST /api/v2/agent-runs
GET /api/v2/agent-runs/:id
PUT /api/v2/agent-runs/:id/status
POST /api/v2/agent-runs/:id/operations
```

后端 service 测试覆盖了 create / get / list / update status / append operation / invalid transition 的一部分基础行为。

### 3. Agent Runs 前端页面

部分通过。

已新增：

- `/dashboard/agent-runs`
- `/dashboard/agent-runs/[id]`
- `agentRunApi`
- 前端 Agent Run 类型

页面具备列表、`agent_id` / `status` 过滤控件、详情、operation timeline、Case / Payload / Interactions / Evidence 链接的基础展示。

### 4. 生产构建

通过。

`cd frontend-next && npm run build` 在非沙箱环境通过，并生成：

- `/dashboard/agent-runs`
- `/dashboard/agent-runs/[id]`

## 阻塞问题

### 1. MCP 测试失败，仓库后端测试门禁未通过

未通过。

`internal/mcp/server.go` 的 `create_oast_probe` 在提供 `agent_id` 时新增调用：

```text
POST /api/v2/agent-runs
```

但 `internal/mcp/server_test.go` 的 `TestCreateOASTProbeToolCreatesCaseThenPayload` mock 没有适配该请求，导致：

```text
unexpected request POST /api/v2/agent-runs
```

影响：

- `GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server` 失败
- `GOCACHE=/tmp/gocache go test ./...` 失败
- Sprint J 不能关闭

### 2. Agent Runs E2E 整体 skip，且属于静态空测

未通过。

`frontend-next/e2e/agent-runs.spec.ts` 顶层直接跳过：

```ts
test.skip(true, 'Requires authentication setup - pages redirect to /login without token')
```

文件内两个测试只检查 URL 是否包含目标路径，没有断言：

- `GET /api/v2/agent-runs`
- 过滤参数是否进入 query
- `GET /api/v2/agent-runs/:id`
- operations timeline 是否渲染真实数据
- Interactions / Evidence 链接是否携带 `payload_id`

这直接违反 Sprint J 完成定义：

- E2E 覆盖列表、过滤、详情、operation timeline、回链
- 没有 skip / only / 静态空测

### 3. MCP operation/status 写入失败被静默吞掉

未通过。

Sprint J plan 明确要求：

- operation 保存失败不能静默吞掉
- MCP 工具失败也要写入 operation 或 audit

当前 `internal/mcp/server.go` 多处只记录日志，不让工具失败，也没有补偿 audit：

- `create_oast_probe` append operation 失败只 `log.Printf`
- `create_oast_probe` status update 失败只 `log.Printf`
- `wait_for_interaction` status update 失败只 `log.Printf`
- `wait_for_interaction` append operation 失败只 `log.Printf`
- `summarize_evidence` append operation 失败只 `log.Printf`
- `export_report` append operation 失败只 `log.Printf`

影响：

- Agent Run timeline 可能丢 operation 但 MCP 仍返回成功
- Audit 不能可靠证明 Agent 做过什么
- 不满足 Agent-Native 审计闭环

### 4. `UpdateAgentRunStatus` audit 的 `from_status` 记录错误

未通过。

`internal/agentrun/service.go` 在写 audit 前已经把 `existingRun.Status` 改成新状态：

```go
existingRun.Status = req.Status
```

随后 audit details 使用：

```go
"from_status": string(existingRun.Status),
"to_status": string(req.Status),
```

因此 `from_status` 和 `to_status` 会同时记录为新状态，不能真实回溯状态变化。

### 5. `AppendAgentOperation` 允许写入不存在的 Agent Run

未通过。

`AppendAgentOperation` 查询 Agent Run 后，如果不存在只是不填 `AgentID`，仍继续插入 operation 并写 audit。

影响：

- 可以产生悬空 `agent_run_id`
- operation timeline 和 audit 失去实体约束
- `/api/v2/agent-runs/:id/operations` 对不存在 run 不应成功

### 6. 目标 ESLint 失败

未通过。

目标 lint 命令失败：

```text
frontend-next/src/app/dashboard/agent-runs/page.tsx:55
react-hooks/set-state-in-effect
```

### 7. `docs/verification.md` 对 Sprint J 的结果记录过度乐观

未通过。

`docs/verification.md` 当前写明：

- `go test ./...` all tests passed
- Sprint J completed with all high-priority tasks
- E2E skipped is acceptable

但本次验收复跑结果是：

- `go test ./...` 失败
- E2E skip 不符合 Sprint J 完成定义
- MCP operation/status 失败处理不符合 plan

返修时需要按实际结果修正该文档。

## 返修要求

1. 修复 `internal/mcp/server_test.go`，覆盖真实 Agent Run 创建、append operation、status update 调用。
2. 取消 `frontend-next/e2e/agent-runs.spec.ts` 的 `test.skip`，改为带 auth/mock 的真实链路测试。
3. Agent Runs E2E 至少断言列表 API、过滤 query、详情 API、operation timeline、Interactions / Evidence 链接。
4. MCP append operation / status update 失败不能只 log，需要失败返回或写入可靠 audit。
5. 修复 `UpdateAgentRunStatus` 的 `from_status` audit 记录。
6. `AppendAgentOperation` 对不存在 Agent Run 必须返回错误，并补测试。
7. 修复目标 ESLint 问题。
8. 更新 `docs/verification.md`，只记录真实执行过且真实通过的命令。

## 复验建议命令

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

复验前必须确认 `frontend-next/e2e/agent-runs.spec.ts` 没有 `test.skip`、`test.only` 或只检查静态 URL 的空测。
