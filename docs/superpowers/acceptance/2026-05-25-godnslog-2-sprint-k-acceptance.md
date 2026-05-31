# GODNSLOG 2.0 Sprint K 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-k-package.md`
- `internal/models/apikey.go`
- `internal/auth/service.go`
- `internal/auth/middleware_test.go`
- `internal/mcp/permissions.go`
- `internal/mcp/server.go`
- `internal/mcp/server_test.go`
- `server/v2_api.go`
- `frontend-next/src/app/dashboard/apikeys/page.tsx`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/apikeys.spec.ts`
- `docs/MCP_SERVER_USAGE.md`
- `docs/agent-native-specification.md`
- `docs/verification.md`

## 验收结论

**结论：Sprint K 未通过验收，需要返修。**

本轮实现已经补入 Agent scope 命名、`internal/mcp/permissions.go`、API Keys 页面 Agent Key 模式、`/api/v2/apikeys` 切向 `internal/auth.Service` 的部分实现，以及相关文档更新。后端聚焦测试、后端全量测试和前端生产构建均可通过。

但 Sprint K 的核心完成定义是“Agent API Key Permission Gate & MCP Safety Controls”。当前 MCP scope/risk gate 在真实路径中没有读取真实 API Key 权限信息，API Keys E2E 失败，目标 ESLint 失败，高风险 scope 设计也与完成定义冲突，因此不能关闭 Sprint K。

## 返修复验（2026-05-31）

**结论：Sprint K 返修后仍不建议关闭。**

Windsurf 本轮返修已经解决多项后端阻塞：

- `server/middleware.go` 已将完整 API Key 放入 Gin context 的 `api_key_full`。
- `server/v2_api.go` 的 `/api/v2/auth/info` 已返回真实 API Key 的 `scopes`、`is_agent`、`risk_tolerance`、`workspace_id`。
- `internal/mcp/server.go` 的 `getAPIKeyInfo` 已从 `/api/v2/auth/info` 解析真实 API Key 权限信息，不再对 API Key 路径默认 `admin:all`。
- `internal/mcp/server.go` 的 `writePermissionDeniedAudit` 已改为 `POST /api/v2/audit/logs`。
- `server/v2_api.go` 已补 `POST /api/v2/audit/logs` 和 `v2CreateAuditLog`。
- `internal/models/apikey.go` 的 `ValidateAgentScopes` 已允许 `HighRiskAgentScopes`，`agent:revoke_token` 可作为显式授权 scope 创建。
- API Keys 页面已补创建成功后一次性明文 key 展示弹窗，目标 ESLint 从 error 修复为仅 warning。

本次复验执行结果：

```bash
GOCACHE=/tmp/gocache go test ./internal/auth ./internal/mcp ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
```

结果：通过，但仍有 warning：

- `frontend-next/src/app/dashboard/apikeys/page.tsx:45:5` unused eslint-disable directive
- `frontend-next/src/app/dashboard/apikeys/page.tsx:47:6` `useEffect` missing dependency `loadAPIKeys`

```bash
cd frontend-next && npm run build
```

结果：通过。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts e2e/agent-runs.spec.ts
```

结果：未能进入浏览器执行，7 条均因本机 Playwright Chromium 未安装失败：

- 缺失路径：`/home/chenq/.cache/ms-playwright/chromium_headless_shell-1223/chrome-headless-shell-linux64/chrome-headless-shell`
- Playwright 提示需执行：`npx playwright install`

因此，本环境不能把 E2E 失败归因为业务逻辑，但也不能确认 Sprint K 的 API Keys E2E 已通过。

剩余验收阻塞：

1. `internal/mcp/server_test.go` 仍没有看到针对 `checkToolPermission` / `revoke_token` 的真实门禁测试。当前测试 helper 的 `/api/v2/auth/info` 返回 `{"user_id":"test-user"}`，实际走 JWT fallback `admin:all`，不能证明缺 scope、超 risk 时下游业务 API 不会被调用。
2. `agent_permission.denied` 已改为真实路由，但缺少测试证明拒绝时会调用 `POST /api/v2/audit/logs`，也缺少后端路由持久化测试覆盖。
3. `frontend-next/e2e/apikeys.spec.ts` 已修复登录 token，但 `should create agent API key` 只断言 `is_agent` 和 `risk_tolerance`，未断言创建请求体包含 Agent scopes 和 `expires_at`。
4. API Keys E2E 未断言创建成功后一次性明文 key 弹窗真实出现并展示返回的 full key。
5. API Keys E2E 未覆盖 Agent-safe scopes 与普通 human scopes 的模式切换约束。
6. API Keys E2E 当前仍有 `waitForTimeout(2000)` 和调试 `console.log`，不够稳定，不应作为最终验收用例。

返修后端主链路方向正确，但 Sprint K 的完成定义包含“门禁可证明”和“E2E 覆盖真实动态行为”。在上述测试补齐并在安装 Playwright browser 的环境跑通前，Sprint K 不应标记完成。

## 已执行验证

```bash
GOCACHE=/tmp/gocache go test ./internal/auth ./internal/mcp ./server
```

结果：通过。

```bash
GOCACHE=/tmp/gocache go test ./...
```

结果：通过。

```bash
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
```

结果：失败。

- `frontend-next/src/app/dashboard/apikeys/page.tsx:30`
- `react-hooks/immutability`
- `loadAPIKeys` 在声明前被 `useEffect` 访问
- 另有 warning：`frontend-next/e2e/apikeys.spec.ts:129` 未使用变量 `url`

```bash
cd frontend-next && npm run build
```

结果：通过。

备注：按既有环境约束，在非沙箱环境执行。

```bash
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts e2e/agent-runs.spec.ts
```

结果：失败。

- 总计：`3 passed, 4 failed`
- 失败均来自 `frontend-next/e2e/apikeys.spec.ts`
- `agent-runs.spec.ts` 3 条通过
- API Keys 失败项：
  - `should display API keys list`
  - `should create agent API key`
  - `should revoke API key`
  - `should not leak full API key in list`

失败原因：

- `apikeys.spec.ts` 没有像其他 E2E 一样设置 `localStorage.token`
- 页面访问 `/dashboard/apikeys` 后实际停留在登录页
- 断言看到的是登录页标题 `GODNSLOG 2.0`，不是 `API Keys 管理`

已停止本次 E2E 使用的 Next.js dev server。

## 通过项

### 1. Agent scope 命名已补入

部分通过。

`internal/models/apikey.go` 已新增 Sprint K scope：

- `agent:create_probe`
- `agent:wait_interaction`
- `agent:read_interactions`
- `agent:summarize_evidence`
- `agent:export_report`
- `agent:read_runs`

并新增高风险 scope 列表：

- `agent:revoke_token`
- `agent:delete_payload`
- `agent:modify_config`

### 2. MCP tool -> scope -> risk 映射已补入

部分通过。

`internal/mcp/permissions.go` 已定义：

- `create_oast_probe` -> `agent:create_probe` -> medium
- `wait_for_interaction` -> `agent:wait_interaction` -> low
- `list_interactions` -> `agent:read_interactions` -> low
- `summarize_evidence` -> `agent:summarize_evidence` -> low
- `export_report` -> `agent:export_report` -> low
- `revoke_token` -> `agent:revoke_token` -> high

### 3. 后端测试和构建门禁

部分通过。

- 后端聚焦测试通过
- 后端全量测试通过
- 前端生产构建通过

## 阻塞问题

### 1. MCP scope/risk gate 没有读取真实 API Key 权限，生产路径形同虚设

未通过。

`internal/mcp/server.go` 的 `getAPIKeyInfo` 调用了 `/api/v2/auth/info`，但没有解析 API Key scopes / `is_agent` / `risk_tolerance`。当前实现直接返回：

```go
Scopes:        []string{"admin:all"},
IsAgent:       false,
RiskTolerance: "high",
```

影响：

- 任意 MCP API Key 都会被当作非 Agent admin
- 缺少 scope 不会被真实拒绝
- risk 超过 tolerance 不会被真实拒绝
- `revoke_token` 对 Agent 默认禁用无法在真实路径成立
- `agent_permission.denied` audit 只可能在构造出来的测试路径成立，不能证明生产链路成立

这直接违反 Sprint K 完成定义：

- MCP 缺 scope 或超 risk 时拒绝执行
- `revoke_token` 对 Agent 默认禁用
- 权限拒绝写入 audit

### 2. `agent:revoke_token` 无法作为合法 Agent scope 创建

未通过。

Sprint K 计划要求：

- Agent Key 缺少 `agent:revoke_token` 时不能执行
- Agent Key 有 `agent:revoke_token` 但 `risk_tolerance` 低于 high 时不能执行
- 只有显式 high scope + high risk tolerance 才允许

当前 `internal/auth/service.go` 的 `ValidScopes` 包含 `agent:revoke_token`，但 `internal/models.ValidateAgentScopes` 只允许 `AgentScopes`，不包含 `HighRiskAgentScopes`。

影响：

- Agent Key 无法合法创建带 `agent:revoke_token` 的显式授权
- 无法测试“有 high scope 但 risk_tolerance 不足被拒绝”和“high scope + high tolerance 允许”的完整门禁

### 3. 权限拒绝 audit 调用的 API 路径不成立

未通过。

`writePermissionDeniedAudit` 调用：

```text
POST /api/v2/audit
```

但现有 v2 API 中已知审计查询路径是：

```text
GET /api/v2/audit/logs
```

本轮未发现 `POST /api/v2/audit` 的真实路由实现。因此 `agent_permission.denied` audit 很可能不会被持久化。

### 4. API Keys E2E 全部失败

未通过。

`frontend-next/e2e/apikeys.spec.ts` 没有设置认证 token，页面被 middleware / auth 流程重定向到 `/login`，因此 4 条 API Keys E2E 全部失败。

失败表现：

- 期望 `API Keys 管理`
- 实际页面标题 `GODNSLOG 2.0`
- 创建按钮、删除按钮、`Key:` 文本均找不到

这直接违反 Sprint K 完成定义：

- E2E 覆盖 Agent Key 创建、展示、撤销和 no full key leakage
- E2E 无静态空测且必须真实通过

### 5. 目标 ESLint 失败

未通过。

`frontend-next/src/app/dashboard/apikeys/page.tsx` 存在 React Hooks 规则错误：

```text
Cannot access variable before it is declared
loadAPIKeys is accessed before it is declared
```

### 6. API Keys 页面没有展示一次性明文 key

未通过。

Sprint K 要求：

- 创建成功后只显示一次明文 key
- 页面不得显示历史完整明文 key

当前 `handleCreateKey` 成功后直接关闭 modal 并刷新列表，没有把创建响应中的 `key` 以一次性方式展示给用户。

影响：

- 用户创建 key 后无法复制明文 key
- “只显示一次”这条安全 UX 契约没有实现

### 7. E2E 覆盖不足

未通过。

当前 `apikeys.spec.ts` 即使修复登录问题，仍缺少关键断言：

- 未断言创建请求包含 `expires_at`
- 未断言 Agent Key 模式只展示 Agent-safe scopes
- 未断言 Human Key 模式不会展示 Agent-only 默认 scope
- 未断言创建成功后只显示一次明文 key
- 未断言列表展示 `risk_tolerance` / `expires_at`
- 未断言撤销后列表刷新
- 未断言没有历史完整明文 key 泄漏

### 8. `docs/verification.md` 记录过度乐观

未通过。

`docs/verification.md` 当前记录 Sprint K 已完成并通过多项能力，但本次复验结果是：

- 目标 ESLint 失败
- API Keys E2E 失败
- MCP gate 没有真实 API Key 权限来源
- 权限拒绝 audit 路由不成立

返修时需要按真实执行结果修正文档。

## 返修要求

1. 修复 MCP `getAPIKeyInfo`，必须从真实认证上下文或有效 API 返回 scopes / `is_agent` / `risk_tolerance`，不能默认 `admin:all`。
2. 补真实测试证明缺 scope / 超 risk 时 MCP 不调用下游业务 API。
3. 明确支持 `agent:revoke_token` 的显式授权路径，并测试：
   - 无 scope 拒绝
   - 有 scope 但 risk_tolerance 低于 high 拒绝
   - 有 scope 且 high tolerance 才允许
4. 修复 `agent_permission.denied` audit 写入，使用真实存在的 service 或 API 路径，并测试持久化。
5. 修复 `apikeys/page.tsx` 目标 ESLint 错误。
6. 修复 `apikeys.spec.ts` 登录 setup，让 E2E 真实进入 `/dashboard/apikeys`。
7. 补 E2E 断言创建请求体包含 `is_agent`、Agent scopes、`risk_tolerance`、`expires_at`。
8. 实现并测试创建成功后一次性明文 key 展示。
9. 清理或忽略本次失败产生的 `frontend-next/test-results`，不要把失败截图作为功能提交内容混入。
10. 修正 `docs/verification.md`，只记录真实通过的命令。

## 复验建议命令

```bash
GOCACHE=/tmp/gocache go test ./internal/auth ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts e2e/agent-runs.spec.ts
```

复验前必须确认：

- `apikeys.spec.ts` 没有 `test.skip` / `test.only`
- API Keys E2E 不停留在登录页
- MCP permission tests 能证明下游业务 API 未被调用
- `agent_permission.denied` audit 真实写入
