# GODNSLOG 2.0 Sprint G Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint G
- **Sprint 主题**：Interaction Triage 与 Evidence 贯通
- **所属阶段**：Phase 3 - First Control Plane Pages

## Sprint 背景

Sprint F 已完成控制面主链路收口：

`Cases Board -> Case Detail -> New Payload -> Payload Detail`

Sprint F 的价值是让用户能从 Case 创建 Payload，并从 Payload Detail 自然进入后续验证闭环。Sprint G 不应立刻跳到 Scanner Hub 或批量治理，而应把 Sprint F 已经指向的两个目标页做实：

- `/dashboard/interactions`
- `/dashboard/evidence`

Sprint G 的核心问题是：

用户从 `Case Detail` 或 `Payload Detail` 进入 `Interactions / Evidence` 后，能否带着上下文完成查看、筛选、定位、详情确认和证据生成。

## Sprint 目标

本 Sprint 只聚焦 4 件事：

1. 让 Interaction 页面真正支持 `case_id` / `payload_id` 上下文过滤
2. 把 Interaction 详情从“原始记录查看”收口成可用于验证判断的 triage 面板
3. 让 Evidence 页面支持从 URL 上下文自动生成对应 Case 或 Payload 的 Evidence
4. 建立 `Payload Detail -> Interactions -> Evidence` 的可验收闭环

本 Sprint 不是更多页面开发，也不是 Scanner Hub 集成。目标是把已有闭环中的 Interaction / Evidence 目标页变成可用工作台。

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/unified-control-plane.md`
- `docs/implementation-dependencies.md`
- `docs/superpowers/plans/2026-05-22-godnslog-2-sprint-e-package.md`
- `docs/superpowers/specs/2026-05-23-godnslog-2-sprint-f-design.md`
- `docs/superpowers/acceptance/2026-05-23-godnslog-2-sprint-f-acceptance.md`

## 当前现状判断

### Interactions 页面现状

- 文件：`frontend-next/src/app/dashboard/interactions/page.tsx`
- 当前已能列表展示 Interaction
- 当前已支持类型筛选、搜索、表格/时间线视图和详情弹窗
- 当前主加载逻辑仍固定调用 `interactionApi.list({ page: 1, page_size: 100 })`
- 当前未把 URL 中的 `case_id` / `payload_id` 纳入主查询条件
- 当前 stats 未随上下文过滤，容易让用户误读范围
- 当前详情弹窗展示基础字段，但缺少 Case/Payload 归因、证据动作、复制关键值等 triage 需要的信息结构

### Evidence 页面现状

- 文件：`frontend-next/src/app/dashboard/evidence/page.tsx`
- 当前已切换到统一 Evidence 契约
- 当前支持选择 Case 后生成 Evidence
- 当前展示 summary、explainability、timeline 和报告预览
- 当前未从 URL 自动读取 `case_id` / `payload_id`
- 当前从 Payload Detail 或 Case Detail 跳转过来后，用户仍需要重新选择上下文

### 后端 API 现状

已有基础接口：

- `GET /api/v2/interactions`
- `GET /api/v2/interactions/stats`
- `GET /api/v2/interactions/timeline`
- `GET /api/v2/interactions/:id`
- `POST /api/v2/interactions/export`
- `POST /api/v2/evidence/generate`

Sprint G 优先复用现有 API。只有当前端上下文过滤或 stats 过滤存在真实后端缺口时，才允许补小型后端修正和测试。

## 实施范围

本 Sprint 只允许覆盖以下 4 个主题。

### 1. Interactions 页面支持上下文过滤

目标是让从 `Case Detail` / `Payload Detail` 进入 Interactions 时，页面天然处于对应对象的交互范围内。

至少覆盖：

- 读取 URL query：
  - `case_id`
  - `payload_id`
  - `type`
- 将 `case_id` / `payload_id` 传入 `interactionApi.list`
- 将 `case_id` / `payload_id` 传入 `interactionApi.stats`
- 页面顶部显示当前作用域：
  - All Interactions
  - Case scoped
  - Payload scoped
- 提供清除当前 scope 的入口，回到全局 Interactions

要求：

- URL 是可分享状态，刷新后过滤仍然生效
- 页面文案必须明确当前列表范围，避免用户误以为在看全局数据
- 搜索框和类型筛选只在当前 scope 内继续过滤
- 空状态必须区分“全局无数据”和“当前 Case/Payload 暂无交互”

### 2. Interaction 详情升级为 triage 面板

目标是让详情弹窗不只是字段堆叠，而是能支持安全验证人员判断这条回连是否可信、归因到哪里、下一步去哪里。

至少覆盖：

- 基本信息：
  - type
  - timestamp
  - source_ip
  - token
- 归因信息：
  - case_id
  - payload_id
- 协议细节：
  - DNS domain
  - HTTP method / path / user_agent
  - headers
  - body
- 快捷动作：
  - 复制 token
  - 复制 domain / path
  - 跳转到关联 Payload Detail
  - 跳转到关联 Case Detail
  - 用当前 Case 或 Payload 生成 Evidence

要求：

- 不在 Sprint G 中实现“人工标记噪声”“删除”“批量操作”
- 对长 headers/body 使用可读的 code block，不撑破弹窗
- 对缺失字段显示稳定空态，不显示 `undefined`
- 快捷动作必须只在有对应 ID 或字段时出现

### 3. Evidence 页面支持上下文自动生成

目标是让用户从 `Case Detail` 或 `Payload Detail` 进入 Evidence 后，不需要重新选择上下文即可看到证据结果。

至少覆盖：

- 读取 URL query：
  - `case_id`
  - `payload_id`
  - `format`
- 如果有 `case_id`，默认生成该 Case 的 Evidence
- 如果有 `payload_id`，默认生成该 Payload 的 Evidence
- 页面顶部显示当前 Evidence scope
- 保留手动选择 Case 的能力，但不能覆盖 URL scope 的主路径体验
- JSON / Markdown 导出继续可用

要求：

- `payload_id` Evidence 生成必须调用 `POST /api/v2/evidence/generate` 的 `payload_id` 参数
- `case_id` Evidence 生成必须调用 `POST /api/v2/evidence/generate` 的 `case_id` 参数
- 无 evidence 时显示稳定 no evidence 空态
- API 错误态要明确，不允许静默吞掉主链路错误

### 4. 串联 Sprint F 到 Sprint G 的闭环导航

目标是补齐 Sprint F 中已经出现的 Interactions / Evidence 入口，让用户进入目标页时不会丢失上下文。

至少覆盖：

- `Case Detail -> View Interactions` 携带 `case_id`
- `Case Detail -> View Evidence` 携带 `case_id`
- `Payload Detail -> View Interactions` 携带 `payload_id`，如页面已有 case_id 也可同时携带
- `Payload Detail -> View Evidence` 携带 `payload_id`，如页面已有 case_id 也可同时携带
- Interactions triage 面板中的 Evidence 动作能跳转到 `/dashboard/evidence?...`

要求：

- 不改变 Sprint F 已验收的主链路语义
- 不把批量操作、生命周期治理、模板平台带回本 Sprint
- 所有新增导航必须可被 E2E 覆盖

## 禁止越界项

Windsurf 在 Sprint G 中不得进入以下内容：

- 不实现 Scanner Hub / Nuclei / Burp / Yakit / ZAP / xray 集成
- 不实现 Interaction 批量删除或批量标记
- 不实现噪声规则管理页面
- 不实现 Evidence 编辑、持久化大改或报告中心
- 不实现 PDF / SARIF 导出
- 不开始 Payload 生命周期管理，例如 revoke / expire / 手动状态编辑
- 不做 dashboard 全局导航重构
- 不修复全量前端 lint 历史债
- 不把 Playwright webServer 自动启动兼容性作为 Sprint G 功能任务

如果为了完成 Sprint G 需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作。

### 前端页面

- `frontend-next/src/app/dashboard/interactions/page.tsx`
- `frontend-next/src/app/dashboard/evidence/page.tsx`
- `frontend-next/src/app/dashboard/cases/[id]/page.tsx`
- `frontend-next/src/app/dashboard/payloads/[id]/page.tsx`

### 前端 API 与类型

- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`

### E2E

- `frontend-next/e2e/cases.spec.ts`
- `frontend-next/e2e/interactions.spec.ts`
- `frontend-next/e2e/evidence.spec.ts`

如当前没有 `interactions.spec.ts`，允许新增。新增测试必须使用非交互式 Playwright 命令运行。

### 后端可选小修

仅当现有 API 无法支撑 Sprint G 主路径时，才允许修改：

- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/interaction/service.go`
- `internal/interaction/service_test.go`

后端修改必须有对应 Go 测试。

## 建议实施顺序

Windsurf 应按以下顺序推进：

1. 先修正 Sprint F 页面到 Interactions / Evidence 的导航参数
2. 再让 Interactions 页面读取 URL scope 并传给 list/stats
3. 再升级 Interaction 详情弹窗为 triage 面板
4. 再让 Evidence 页面读取 URL scope 并自动生成
5. 最后补 E2E 覆盖完整闭环

不要先做大 UI 重排。先保证 URL scope、API 参数和 E2E 主链路成立。

## 必须补齐的测试

### 1. Interactions 页面 E2E

至少覆盖：

- `/dashboard/interactions?case_id=case-1` 会按 case scope 加载列表
- `/dashboard/interactions?payload_id=payload-1` 会按 payload scope 加载列表
- 页面可见当前 scope 标识
- 清除 scope 后回到全局列表
- 类型筛选只影响当前 scope 内数据
- 点击 Interaction 打开 triage 面板
- triage 面板展示 token、case_id、payload_id、source_ip
- triage 面板可跳转到关联 Case / Payload / Evidence

### 2. Evidence 页面 E2E

至少覆盖：

- `/dashboard/evidence?case_id=case-1` 自动生成 Case Evidence
- `/dashboard/evidence?payload_id=payload-1` 自动生成 Payload Evidence
- 页面展示 `evidence_strength`、`confidence`、`interaction_count`、`unique_sources`
- timeline 可见
- Markdown / JSON 导出仍可触发
- no evidence 空态稳定可见
- API 错误态稳定可见

### 3. Sprint F 到 Sprint G 链路 E2E

至少覆盖：

- 从 `Case Detail` 点击 View Interactions 后进入带 `case_id` 的 Interactions 页面
- 从 `Case Detail` 点击 View Evidence 后进入带 `case_id` 的 Evidence 页面
- 从 `Payload Detail` 点击 View Interactions 后进入带 `payload_id` 的 Interactions 页面
- 从 `Payload Detail` 点击 View Evidence 后进入带 `payload_id` 的 Evidence 页面

### 4. 后端测试

如果 Sprint G 修改了后端 API，至少覆盖：

- `GET /api/v2/interactions?case_id=...` 过滤有效
- `GET /api/v2/interactions?payload_id=...` 过滤有效
- `GET /api/v2/interactions/stats?case_id=...` 统计随 case scope 变化
- `GET /api/v2/interactions/stats?payload_id=...` 统计随 payload scope 变化
- `POST /api/v2/evidence/generate` 的 `payload_id` 分支稳定返回结构化 Evidence

## 验证命令

完成 Sprint G 前必须记录实际运行命令和结果。最低要求：

```bash
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/interactions/page.tsx src/app/dashboard/evidence/page.tsx src/app/dashboard/cases/[id]/page.tsx src/app/dashboard/payloads/[id]/page.tsx e2e/cases.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

UI 行为验证按 `docs/verification.md` 当前约定执行 Sprint G 相关 Playwright 测试。由于 Playwright webServer 自动启动在不同环境中仍有兼容性技术债，Sprint G 验收允许采用两步法：

```bash
cd frontend-next
npm run dev
npx playwright test --reporter=line e2e/cases.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

要求：

- 禁止使用 `npx playwright show-report`
- 禁止使用会常驻 HTML report server 的流程
- 测试结束后停止 dev server
- 清理 `frontend-next/test-results`

## 完成定义

只有同时满足以下条件，Sprint G 才能视为完成：

1. Case / Payload 进入 Interactions 时不会丢失上下文
2. Interactions 页面 list 与 stats 都支持当前 scope
3. Interaction triage 面板能支持判断、复制和跳转
4. Case / Payload 进入 Evidence 时能自动生成对应 Evidence
5. Evidence 页面继续展示结构化 summary、timeline、export
6. Sprint F 主链路没有回归
7. E2E 覆盖 Sprint G 主链路且无 `test.skip`
8. `go test ./...`、前端 build、Sprint G 定向 lint 通过

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式回传：

```text
Sprint G 完成回传

本地 commits:
- <hash> <title>
- <hash> <title>

实现摘要:
- Interactions:
- Interaction triage:
- Evidence:
- Sprint F navigation:

测试结果:
- GOCACHE=/tmp/gocache go test ./...: <pass/fail>
- cd frontend-next && npm run build: <pass/fail>
- Sprint G 定向 eslint: <pass/fail>
- Playwright Sprint G specs: <pass/fail, passed/failed 数量>

已知遗留:
- <如无，写 无>

需要 Codex 复验的点:
- <列出最重要的 3-5 个验收点>
```

## Codex 验收重点

Codex 复验 Sprint G 时重点检查：

1. URL scope 是否真实进入 API 请求，而不是只在前端本地过滤
2. stats 是否随 scope 变化
3. `Payload Detail -> Interactions -> Evidence` 是否形成闭环
4. Evidence 是否继续使用统一 Evidence 契约
5. E2E 是否没有 skip 或“只检查静态文字”的空测
6. 是否严格没有越界到 Scanner Hub、生命周期治理或批量操作

只有上述全部成立，Sprint G 才能关闭。
