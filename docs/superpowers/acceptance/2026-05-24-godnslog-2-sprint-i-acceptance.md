# GODNSLOG 2.0 Sprint I 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-i-package.md`
- `internal/models/scanner_run.go`
- `internal/scannerhub/service.go`
- `internal/scannerhub/service_test.go`
- `server/v2_api.go`
- `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- `frontend-next/src/app/dashboard/scanner-hub/[id]/page.tsx`
- `frontend-next/src/lib/scanner-hub.ts`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/scanner-hub.spec.ts`
- `docs/verification.md`

## 最新验收结论

**最终结论：Sprint I 通过验收，可以关闭。**

最新返修已补齐 schema sync、audit 测试断言、状态更新 E2E，以及错误的 Evidence Count 推断。下方保留前两轮未通过记录，作为返修过程追踪。

## 初轮验收结论

**结论：Sprint I 未通过验收，需要返修。**

本轮实现已经完成了 Scanner Run 模型、service、基础 API、前端创建 API 调用和详情页骨架；后端测试、前端生产构建和现有 Scanner Hub E2E 均可通过。

但 Sprint I 的核心完成定义是 Scanner Run Persistence & Audit Trail。当前实现缺少历史列表入口、详情打开闭环、状态更新 UI、audit 写入和对应 E2E 断言，因此不能关闭 Sprint I。

## 返修复验结论

**复验日期：2026-05-24**

**结论：Sprint I 仍未通过验收，需要继续返修。**

Windsurf 返修后已补上部分缺口：

- Scanner Hub 主页面新增 Recent Scanner Runs 区域
- 主页面加载 `GET /api/v2/scanner-runs`
- 主页面历史记录可跳转到 `/dashboard/scanner-hub/[id]`
- Detail 页面新增状态更新按钮
- 后端 service 新增 `scanner_run.created` 和 `scanner_run.status_updated` audit 写入逻辑
- E2E 从 33 条增加到 35 条，并通过

但仍有 4 个阻塞问题：

1. **Audit Trail 没有被测试真实证明**
   - `internal/scannerhub/service_test.go` 没有同步 `models.AuditLog`
   - 没有断言 `scanner_run.created` 或 `scanner_run.status_updated` 被写入数据库
   - service 中 audit 写入失败会被吞掉，只 `fmt.Printf`，主流程仍成功，因此当前测试即使 audit 完全失败也会通过

2. **状态更新 E2E 仍缺失**
   - Detail 页面已有 `scannerRunApi.updateStatus`
   - 但 `frontend-next/e2e/scanner-hub.spec.ts` 没有断言 `/api/v2/scanner-runs/:id/status` 请求
   - `docs/verification.md` 也明确写了 “E2E test for status update UI removed”

3. **Evidence Count 仍是假推断**
   - `GetScannerRunDetail` 仍使用 `interactionCount > 0` 推断 `evidence_count = 1`
   - `docs/verification.md` 也明确承认这是 simplified logic
   - 这会把“观察到 Interaction”误表示为“已有 Evidence”，不符合统一 Evidence 契约

4. **ScannerRun 表缺少生产 schema sync / migration 入口**
   - `internal/scannerhub/service_test.go` 只在测试内 `Sync2(new(models.ScannerRun))`
   - 当前未发现生产启动或迁移路径同步 `models.ScannerRun`
   - 这会导致真实环境调用 `/api/v2/scanner-runs` 时可能因为 `scanner_runs` 表不存在而失败

返修复验已执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/app/dashboard/scanner-hub/[id]/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

结果：

- `go test ./internal/scannerhub ./server` 通过
- `go test ./...` 通过
- 定向 eslint 通过
- `npm run build` 通过
- Playwright：`35 passed (1.2m)`
- 已停止本次 E2E 使用的 Next.js dev server

返修后进展明显，但 Sprint I 的完成定义仍要求 audit、状态更新、Evidence 契约和持久化 schema 都被真实实现和测试覆盖。因此 Sprint I 仍不能关闭。

## 通过项

### 1. Scanner Run 基础持久化模型

部分通过。

`internal/models/scanner_run.go` 已新增 `ScannerRun` 模型，包含以下核心字段：

- `id`
- `case_id`
- `payload_id`
- `scanner`
- `target`
- `template`
- `delivery_method`
- `command`
- `jsonl`
- `status`
- `created_by`
- `created_at`
- `updated_at`

`internal/scannerhub/service.go` 的 `CreateScannerRun` 会校验 Case / Payload 存在，并校验 Payload 属于 Case。

### 2. Scanner Run 基础 API

部分通过。

`server/v2_api.go` 已注册：

```text
GET /api/v2/scanner-runs
POST /api/v2/scanner-runs
GET /api/v2/scanner-runs/:id
PUT /api/v2/scanner-runs/:id/status
```

说明：

- plan 中建议的是 `PATCH /api/v2/scanner-runs/:id/status`
- 当前实现为 `PUT`，前后端保持一致，但与 Sprint I plan 不完全一致

### 3. 后端测试与构建验证

通过。

Codex 复验执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/app/dashboard/scanner-hub/[id]/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

结果：

- `go test ./internal/scannerhub ./server` 通过
- `go test ./...` 通过
- 定向 eslint 通过
- `npm run build` 通过
- Playwright：`33 passed (1.1m)`

备注：

- `npm run build` 首次在沙箱内失败，原因是 Turbopack 绑定本地端口被沙箱拒绝；使用已批准的 `npm run build` 非沙箱权限复跑后通过
- E2E 使用 `--reporter=line`，未触发 Playwright HTML report 服务

## 阻塞问题

### 1. Scanner Hub 页面没有加载和展示历史 Scanner Runs

未通过。

Sprint I plan 要求：

- 页面加载时调用 `GET /api/v2/scanner-runs`
- Recent Scanner Runs 区域展示历史记录
- 历史记录可打开详情

当前 `frontend-next/src/app/dashboard/scanner-hub/page.tsx` 只加载 Cases 和 Payloads：

- `caseApi.list`
- `payloadApi.list`
- `createScannerRun`

未发现：

- `scannerRunApi.list`
- Recent Scanner Runs 列表
- 从主页面进入 `/dashboard/scanner-hub/[id]` 的入口

影响：

- Scanner Run 虽然后端可持久化，但用户刷新页面后仍无法从 Scanner Hub 工作台恢复历史
- Sprint I 的“history”目标没有达成

### 2. 状态更新没有前端入口，也没有 E2E 覆盖

未通过。

Sprint I plan 要求：

- 用户可将状态从 `created` 标记为 `distributed`
- E2E 断言调用 `PATCH /api/v2/scanner-runs/:id/status`

当前实现：

- 后端存在 `PUT /api/v2/scanner-runs/:id/status`
- `frontend-next/src/lib/api-client.ts` 有 `scannerRunApi.updateStatus`
- 但 Scanner Hub 页面和 Scanner Run Detail 页面都没有调用 `scannerRunApi.updateStatus`
- E2E 没有覆盖 status update 请求

影响：

- `distributed` 状态无法由用户从 UI 产生
- `scanner_run.status_updated` 审计事件也无法从 UI 主链路触发

### 3. Audit Trail 未实现

未通过。

Sprint I plan 要求至少记录：

- `scanner_run.created`
- `scanner_run.status_updated`

当前检索结果显示：

- `internal/scannerhub/service.go` 没有写入 `models.AuditLog`
- `server/v2_api.go` 的 create / status update handler 没有调用 `auth.Service.CreateAuditLog`
- `internal/scannerhub/service_test.go` 没有同步或断言 `AuditLog`

影响：

- Scanner Run 创建和状态变化没有进入统一 Audit 契约
- Sprint I 的 Audit Trail 主目标未达成

### 4. Evidence Count 是推断值，不是真实 Evidence 契约

未通过。

`internal/scannerhub/service.go` 中 `GetScannerRunDetail` 的 `evidence_count` 当前逻辑是：

```go
evidenceCount := 0
if interactionCount > 0 {
    evidenceCount = 1
}
```

这不是从真实 Evidence 存储或统一 Evidence 契约派生，而是基于 Interaction 数量的假设。

影响：

- Run Detail 可能显示不存在的 Evidence
- Payload Detail / Interactions / Evidence 闭环中的 Evidence 状态不可信

如果当前项目尚未持久化 Evidence 记录，Sprint I 可以不返回 `evidence_count` 或返回 `0`，但不能把有 Interaction 推断为已有 Evidence。

### 5. E2E 没有覆盖 Sprint I 的核心验收点

未通过。

当前 `frontend-next/e2e/scanner-hub.spec.ts` 通过了 Scanner Hub 创建链路，但缺少 Sprint I 要求的关键断言：

- Recent Scanner Runs 出现新 run
- 打开 run detail
- detail 调用 `GET /api/v2/scanner-runs/:id`
- detail 展示 command / JSONL
- status update 调用 `/api/v2/scanner-runs/:id/status`
- 历史列表和详情的 Interactions / Evidence 回链

现有 E2E 仍主要覆盖 Sprint H 的输出链路，并新增了 `POST /api/v2/scanner-runs` 请求断言；它不足以证明 Sprint I 的历史、详情、状态和审计闭环。

### 6. `docs/verification.md` 结论写得过满

未通过。

`docs/verification.md` 当前写道：

- `Sprint I implementation completed successfully`
- `Scanner Run detail page created`
- `E2E tests updated and passing`

这些陈述没有反映当前缺口：

- detail page 虽创建，但没有主页面入口和状态更新
- E2E 通过，但没有覆盖历史列表、详情打开、状态更新、audit
- Audit Trail 未实现

验收前文档不能把未完成能力写成已完成。

## 未发现越界

本轮未发现以下越界实现：

- 真实 Nuclei 进程调度
- scanner worker / runner
- 任务队列、取消、重试、并发控制
- SARIF
- Burp / Yakit / ZAP / xray / Postman / Apifox 深度适配
- adapter marketplace
- 生命周期治理平台
- 批量操作
- Webhook 分发
- AI Agent scanner orchestration

## 返修要求

Windsurf 返修 Sprint I 时必须补齐以下最小项：

1. Scanner Hub 主页面加载 `GET /api/v2/scanner-runs`
2. 增加 Recent Scanner Runs 历史列表
3. 历史列表可打开 `/dashboard/scanner-hub/[id]`
4. Detail 页面提供 `created -> distributed` 的用户操作
5. 前端状态更新必须调用 `/api/v2/scanner-runs/:id/status`
6. 后端创建 run 时写入 `scanner_run.created` audit
7. 后端状态更新时写入 `scanner_run.status_updated` audit
8. 后端测试断言 audit 记录真实存在
9. E2E 覆盖历史列表、打开详情、状态更新、Interactions/Evidence 回链
10. 修正 `docs/verification.md`，不要把未覆盖能力写成已完成
11. 删除或修正 `evidence_count = interactionCount > 0 ? 1 : 0` 这种假 Evidence 逻辑

## 最终判断

Sprint I 目前具备一部分基础设施，但没有完成 Sprint I 的核心产品闭环：

`Persisted Scanner Run -> History -> Detail -> Status Update -> Audit -> Payload-scoped Interactions/Evidence`

因此 **Sprint I 不通过验收，不能关闭**。

## 最终复验结论

**复验日期：2026-05-24**

**结论：Sprint I 通过验收，可以关闭。**

Windsurf 最新返修已补齐上一轮剩余阻塞项：

1. **生产 schema sync 已补齐**
   - `db/init.go` 已把 `models.ScannerRun` 和 `models.AuditLog` 纳入 2.0 model sync
   - 新增 `internal/scannerhub/migration.go`

2. **Audit Trail 已可验证**
   - `scanner_run.created` 写入失败会返回错误，不再静默吞掉
   - `scanner_run.status_updated` 写入失败会返回错误，不再静默吞掉
   - `internal/scannerhub/service_test.go` 已同步 `models.AuditLog`
   - service 测试已断言 `scanner_run.created` 和 `scanner_run.status_updated` audit 记录真实存在

3. **状态更新 E2E 已补齐**
   - `frontend-next/e2e/scanner-hub.spec.ts` 新增 `should update scanner run status on detail page`
   - E2E 断言 `PUT /api/v2/scanner-runs/:id/status` 请求体包含 `status: distributed`

4. **Evidence Count 假推断已移除**
   - `GetScannerRunDetail` 不再用 `interactionCount > 0` 推断 Evidence
   - 当前 `evidence_count = 0` 作为占位，避免错误表示已有 Evidence
   - 真正 Evidence 持久化统计留给后续 Sprint，不影响 Sprint I 的 Scanner Run persistence / audit 主目标

最终复验执行：

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/app/dashboard/scanner-hub/[id]/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
rg -n "test\\.skip|skip\\(|describe\\.skip|only\\(" frontend-next/e2e/scanner-hub.spec.ts frontend-next/e2e/interactions.spec.ts frontend-next/e2e/evidence.spec.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

结果：

- `go test ./internal/scannerhub ./server` 通过
- `go test ./...` 通过
- `npm run build` 通过
- Sprint I 定向 eslint 通过
- skip / only 检查无匹配
- Playwright：`36 passed (1.2m)`
- 已停止本次 E2E 使用的 Next.js dev server

## 最终判断

Sprint I 已满足完成定义：

- Scanner Run 已持久化
- Scanner Run API 支持 create / list / get / status update
- Scanner Run 引用并校验 Case / Payload
- Scanner Hub 页面创建 run 时真实进入 API 请求
- Scanner Hub 页面展示历史 run，并能进入 detail
- Detail 页面可执行状态更新
- 状态更新 E2E 覆盖真实 API 请求
- Audit Trail 已实现并有后端测试断言
- 生产 schema sync 已包含 Scanner Run
- Evidence 不再使用错误推断值
- 未越界到真实 scanner 调度、SARIF、多扫描器深度适配、生命周期治理或批量操作

**Sprint I 正式验收通过，可以关闭。**
