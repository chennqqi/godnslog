# GODNSLOG 2.0 Sprint I Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint I
- **Sprint 主题**：Scanner Run Persistence & Audit Trail
- **所属阶段**：Phase 4 - Primary Scanner Integration

## Sprint 背景

Sprint H 已完成 Scanner Hub MVP：用户可以在控制面基于 Case/Payload 生成 Nuclei command、JSONL record，并跳转到 payload-scoped Interactions / Evidence，形成：

`Scanner Hub -> Payload -> Interactions -> Evidence`

但 Sprint H 验收中明确留下一个产品缺口：

- Scanner Run 当前是前端 in-memory 对象
- 页面刷新后无法恢复 scanner 分发历史
- 无法对 scanner run 的创建、查看、状态变化形成审计
- 不能从历史 run 回到 Case/Payload/Interaction/Evidence 上下文

Sprint I 的目标是补齐 Scanner Run 的**持久化、历史、详情和审计**。本 Sprint 仍然不执行真实 Nuclei 进程，不做扫描任务队列，不扩展多扫描器深度适配。

## Sprint 目标

本 Sprint 只聚焦 5 件事：

1. 建立持久化 Scanner Run 模型，稳定引用已有 Case 和 Payload
2. 提供最小 Scanner Run API：create / list / get / status update
3. Scanner Hub 页面从 in-memory 状态升级为真实 API 数据源
4. Scanner Run Detail 能回到 Payload、Interactions、Evidence 闭环
5. 记录关键 Scanner Run 审计事件，并补齐后端测试和 E2E

完成后，安全团队应能回答：

- 某个 Case 下曾经分发过哪些 scanner run
- 每个 run 使用了哪个 Payload、目标、template、command、JSONL
- 这个 run 是否已经观察到 Interaction，是否已经进入 Evidence
- 这个 run 的创建或状态变化是否有审计记录

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/scanner-hub.md`
- `docs/scanner-hub-adapter-design.md`
- `docs/official-support-boundary.md`
- `examples/nuclei/README.md`
- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-h-package.md`
- `docs/superpowers/acceptance/2026-05-24-godnslog-2-sprint-h-acceptance.md`
- `docs/verification.md`

## 当前现状判断

### 已完成能力

Sprint H 已具备以下能力：

- `/dashboard/scanner-hub` 工作台
- Nuclei template variable command 生成
- Nuclei JSONL record 生成
- Scanner Hub 创建或复用 Payload
- Scanner Hub 输出 payload-scoped Interactions / Evidence 链接
- E2E 覆盖 Scanner Hub 到 Evidence 的主链路

### 主要缺口

- Scanner Run 未持久化
- Scanner Run 无后端 ID、无数据库记录、无 API 列表
- Scanner Hub 页面刷新后无法恢复历史
- 无法从一个历史 run 进入详情页
- 无法审计 scanner run 创建和状态变化
- 无法用 API 测试证明 scanner run 真实绑定 Case/Payload

## 术语边界

### Scanner Run

Scanner Run 表示一次面向外部 scanner 的分发上下文。

Scanner Run 不替代：

- Case
- Payload
- Interaction
- Evidence
- Scanner adapter
- 真实 scanner 任务

Scanner Run 必须引用已有 Case 和 Payload。Interaction 仍通过 Payload / Case 自动归因，Evidence 仍通过统一 Evidence 契约生成。

### 状态模型

Sprint I 只允许使用最小状态：

- `created`：run 已创建，分发材料已生成
- `distributed`：用户已确认 command / JSONL 被分发到外部 scanner
- `observed`：系统观察到该 Payload 的 Interaction
- `evidenced`：该 Payload 已生成 Evidence

要求：

- `created` 必须由创建 API 产生
- `distributed` 可由用户显式操作产生
- `observed` / `evidenced` 优先从已有 Interaction / Evidence 数据派生；如果实现成本过高，可在详情 API 中以统计字段表达，不强制落库
- 不允许引入复杂生命周期、队列状态、执行状态或失败重试状态

## 数据契约

建议最小字段：

- `id`
- `case_id`
- `payload_id`
- `scanner`：Sprint I 只支持 `nuclei`
- `target`
- `template`
- `delivery_method`：`nuclei-jsonl` 或 `nuclei-var`
- `command`
- `jsonl`
- `status`
- `created_by`
- `created_at`
- `updated_at`

详情接口可附加派生字段：

- `interaction_count`
- `last_interaction_at`
- `evidence_count`
- `latest_evidence_id`
- `interactions_url`
- `evidence_url`

要求：

- `case_id` 和 `payload_id` 必须真实存在
- `payload_id` 必须属于 `case_id`
- `scanner` 仅允许 `nuclei`
- `delivery_method` 仅允许 `nuclei-jsonl` / `nuclei-var`
- `command` 和 `jsonl` 必须来自与 Sprint H 一致的生成逻辑
- JSONL 必须保持一行合法 JSON

## 实施范围

本 Sprint 只允许覆盖以下 5 个主题。

### 1. 后端 Scanner Run 模型与服务

目标是把 Scanner Run 从前端临时对象变成后端持久化实体。

建议新增或扩展：

- `internal/scannerhub/`
- `internal/models/scanner_run.go`
- `server/v2_api.go`
- `server/v2_api_test.go`

至少覆盖：

- Scanner Run 数据模型
- 创建时校验 Case / Payload 关系
- 生成或保存 command / JSONL
- 列表查询支持 `case_id` / `payload_id` / `scanner` / `status`
- 详情查询返回回链 URL 和统计字段

要求：

- 不复制 Case/Payload/Interaction/Evidence 的业务模型
- 不新增 scanner-only payload
- 不让 Scanner Run 创建 Interaction 或 Evidence
- 不把 Scanner Run 设计成真实任务调度系统

### 2. Scanner Run API

目标是提供最小、可测、可被前端真实调用的 API。

建议接口：

```text
POST /api/v2/scanner-runs
GET /api/v2/scanner-runs
GET /api/v2/scanner-runs/:id
PATCH /api/v2/scanner-runs/:id/status
```

请求和响应必须沿用项目既有 API 风格，例如：

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

创建请求至少包含：

```json
{
  "case_id": "case-1",
  "payload_id": "payload-1",
  "scanner": "nuclei",
  "target": "https://target.example",
  "template": "ssrf-basic",
  "delivery_method": "nuclei-jsonl"
}
```

API 验收点：

- 未认证请求必须失败
- 无效 `case_id` 必须失败
- 无效 `payload_id` 必须失败
- `payload_id` 不属于 `case_id` 必须失败
- 非 `nuclei` scanner 必须失败
- 非允许值的 `delivery_method` 必须失败
- list 能按 `case_id` 和 `payload_id` 过滤
- get 能返回 command、JSONL、Interactions URL、Evidence URL

### 3. Scanner Hub 前端持久化改造

目标是让用户在 UI 中创建、恢复和查看历史 Scanner Run。

建议修改：

- `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- `frontend-next/src/lib/scanner-hub.ts`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`

可新增：

- `frontend-next/src/app/dashboard/scanner-hub/[id]/page.tsx`
- 或在现有 Scanner Hub 页面内提供 detail drawer/panel

至少覆盖：

- 创建 Scanner Run 时真实调用 `POST /api/v2/scanner-runs`
- 页面加载时调用 `GET /api/v2/scanner-runs`
- Recent Scanner Runs 区域展示历史记录
- 历史记录可打开详情
- 详情展示 command、JSONL、Case、Payload、target、template、status
- 详情提供 Interactions / Evidence 链接，且链接必须携带 `payload_id`
- 用户可将状态从 `created` 标记为 `distributed`

要求：

- 不再把 Scanner Run 仅存在 React state 中作为唯一来源
- 空态、加载态、错误态必须可验收
- Copy command / JSONL / rendered payload 的 Sprint H 能力不能回退
- 页面仍是工作台，不做营销页，不迁移到 Marketplace

### 4. Audit Trail

目标是把 scanner 分发行为纳入安全审计。

至少记录：

- `scanner_run.created`
- `scanner_run.status_updated`

审计内容建议包含：

- `scanner_run_id`
- `case_id`
- `payload_id`
- `scanner`
- `target`
- `template`
- `delivery_method`
- `from_status`
- `to_status`
- `actor`
- `created_at`

要求：

- 复用现有 Audit 机制
- 不新增平行 audit 系统
- 不记录敏感 API token
- 如果 command / JSONL 中包含完整 payload，审计中优先记录 run id、payload id、target、template，不强制记录完整 command

### 5. 测试与验收

目标是证明 Sprint I 不是 UI 静态展示，而是真实 API + 持久化 + 回链闭环。

后端测试至少覆盖：

- 创建 Scanner Run 成功
- 创建时校验 Case / Payload 存在
- 创建时校验 Payload 属于 Case
- 创建时拒绝非 `nuclei` scanner
- 列表过滤 `case_id` / `payload_id`
- 详情返回 command / JSONL / links
- 状态更新产生 audit

前端 E2E 至少覆盖：

- 进入 `/dashboard/scanner-hub`
- 创建 Scanner Run，并通过 request assertion 确认进入 `POST /api/v2/scanner-runs`
- 请求体真实包含 `case_id`、`payload_id`、`scanner`、`target`、`template`、`delivery_method`
- Recent Scanner Runs 出现新 run
- 打开 run detail 后可看到 command 和 JSONL
- 点击 Interactions 链接进入 `/dashboard/interactions?payload_id=...`
- 点击 Evidence 链接进入 `/dashboard/evidence?payload_id=...`
- 标记为 `distributed` 后调用 `PATCH /api/v2/scanner-runs/:id/status`

测试要求：

- 不允许 `test.skip`
- 不允许 `test.only`
- 不允许只检查静态标题或静态文字
- E2E 必须断言真实 API 请求和响应驱动的 UI 变化

## 明确禁止越界

Sprint I 不允许实现以下内容：

- 启动真实 Nuclei 进程
- scanner 任务队列、取消、重试、并发控制
- scanner worker / runner
- SARIF 导入或导出
- Burp / Yakit / ZAP / xray / Postman / Apifox 深度适配
- 插件市场或 adapter marketplace
- 生命周期治理平台
- 批量操作
- Workspace / RBAC 大改
- Webhook 分发
- AI Agent scanner orchestration
- 全量前端 lint 历史债清理
- Playwright `webServer` 架构重做

如实现中发现必须碰到上述内容，必须停下并回传，不允许自行扩范围。

## 建议文件清单

后端：

- `internal/models/scanner_run.go`
- `internal/scannerhub/service.go`
- `internal/scannerhub/service_test.go`
- `server/v2_api.go`
- `server/v2_api_test.go`

前端：

- `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- `frontend-next/src/app/dashboard/scanner-hub/[id]/page.tsx`
- `frontend-next/src/lib/scanner-hub.ts`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/scanner-hub.spec.ts`

文档：

- `docs/scanner-hub.md`
- `docs/verification.md`

实际文件可按项目现有结构微调，但不得绕开既有 API、类型和测试组织方式。

## 验证命令

Windsurf 完成后必须回传实际执行过的命令和结果。建议至少执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

注意：

- E2E 必须使用一次性、非交互式命令
- 不得执行 `npx playwright show-report`
- 不得触发 `Serving HTML report at http://localhost:9323. Press Ctrl+C to quit.`
- 如果 Playwright `webServer` 在环境中不稳定，继续使用 `docs/verification.md` 里的两步法：先 `npm run dev`，再单独跑 `npx playwright test --reporter=line`

## 完成定义

Sprint I 只有在同时满足以下条件时才算完成：

1. Scanner Run 已持久化，不再只是前端 in-memory 状态
2. Scanner Run API 支持 create / list / get / status update
3. Scanner Run 必须真实引用 Case 和 Payload，并校验归属关系
4. Scanner Hub 页面创建 run 时真实进入 API 请求
5. Scanner Hub 页面能展示历史 run 并打开详情
6. Run detail 能回到 payload-scoped Interactions / Evidence
7. command / JSONL 输出保持 Sprint H 契约，不发生字段回退
8. `scanner_run.created` 和 `scanner_run.status_updated` 有审计记录
9. 后端测试覆盖模型、API、校验和 audit
10. E2E 覆盖 API 请求、历史列表、详情、Interactions/Evidence 回链
11. E2E 没有 skip / only / 静态空测
12. 未越界到真实 scanner 调度、SARIF、多扫描器适配、生命周期治理或批量操作
13. `docs/verification.md` 记录 Sprint I 实际验证命令和结果

## Windsurf 回传要求

Windsurf 完成 Sprint I 后，请回传：

- 本地 commit hash
- 修改文件列表
- API 路由清单
- Scanner Run 数据字段说明
- Audit event 名称与触发点
- 已执行验证命令和结果
- 如有未完成项，必须明确列出，不得写成已完成能力

