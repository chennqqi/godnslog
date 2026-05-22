# GODNSLOG 2.0 Sprint E Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint E
- **Sprint 主题**：Evidence Web 展示与 Audit 收口
- **所属阶段**：Phase 3 - First Control Plane Pages

## Sprint 目标

把 Sprint D 已完成的结构化 Evidence 后端能力真正接进控制面，同时把 Audit 从“占位页面/零散模型”收口成可查询、可筛选、可验收的 `/api/v2` 契约。

本 Sprint 只聚焦 4 件事：

1. 把 `/dashboard/evidence` 从旧的 interaction export 语义切换到统一 Evidence 契约
2. 把 Evidence 页面做成 MVP 可用页面，而不是简单文本导出面板
3. 建立 `/api/v2/audit` 的最小可用查询契约，支持 Audit 页面真实读取
4. 让 `/dashboard/audit` 与后端 Audit 契约对齐，不再依赖“后端不存在时静默兜底”作为主路径

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/unified-control-plane.md`
- `docs/superpowers/acceptance/2026-05-22-godnslog-2-sprint-d-final-acceptance.md`

## 当前现状判断

当前已有基础，但还没有完成 Sprint E：

### Evidence 页面现状

- 文件：`frontend-next/src/app/dashboard/evidence/page.tsx`
- 当前仍通过 `interactionApi.export(...)` 生成报告
- 还在暴露 `csv` / `include_raw` 这类属于旧 export 语义的入口
- 页面目标更接近“交互导出面板”，不是“Evidence 时间线与摘要页面”

### Audit 页面现状

- 文件：`frontend-next/src/app/dashboard/audit/page.tsx`
- 当前调用 `/audit/logs`
- 但后端仓库里尚未看到真正注册到 `/api/v2` 的 audit 查询路由
- 页面主路径仍包含“后端可能不存在，UI 静默兜底为空”的逻辑

### 后端 Audit 现状

- `internal/models/audit.go` 已有统一 `AuditLog` / `AuditLogListResponse`
- `internal/auth/service.go` 已有 `CreateAuditLog` / `ListAuditLogs`
- 但 `/api/v2/audit` 路由和对应 handler 仍未形成清晰、可测试的主契约

## 实施范围

本 Sprint 只允许覆盖以下 4 个主题。

### 1. 收口 Evidence Web 页面到统一 Evidence 契约

目标是让 `/dashboard/evidence` 真正消费 Sprint D 的 `/api/v2/evidence/generate` 结果。

至少覆盖：

- case 选择
- 基于 case 生成 Evidence
- Evidence summary
  - `evidence_strength`
  - `confidence`
  - `interaction_count`
  - `unique_sources`
  - `explainability`
- timeline 展示
- JSON / Markdown 导出按钮

要求：

- 前端 Evidence 页面不得继续把 `interaction export` 伪装成 Evidence
- 不再暴露 `csv`
- 不再暴露 `include_raw`
- 主页面优先展示结构化 `data.evidence`
- 导出按钮只对应 `json` / `markdown`

### 2. 收口 Evidence 页面信息结构

目标不是做一个大而全的新页面，而是做 MVP 可用、和控制面风格一致的 Evidence 页面。

至少覆盖：

- 顶部摘要区
- timeline 区
- 导出操作区
- no evidence 空状态
- API 错误态

建议页面结构：

1. 顶部标题和 case 选择器
2. Evidence summary 卡片
3. Explainability 文本
4. Timeline 列表或时间线组件
5. Export 操作

要求：

- 复用 `frontend-next` 现有 UI 组件
- 视觉上与 `cases` / `interactions` / `audit` 页保持同一套控制面风格
- 不新增营销式布局，不做 landing page 风格

### 3. 建立 `/api/v2/audit` 最小可用查询契约

本 Sprint 的 Audit 收口范围是“页面+后端契约收口”，不是全链路治理大改。

至少覆盖：

- `GET /api/v2/audit/logs`

建议查询参数：

- `page`
- `page_size`
- `user_id`
- `action`
- `resource_type`
- `start_time`
- `end_time`

建议返回结构：

- `items`
- `total`
- `page`
- `page_size`
- `total_pages`

要求：

- 复用 `internal/auth/service.go` 中已有的 `ListAuditLogs`
- `/api/v2` 返回口径必须统一为 `code / message / data`
- 不能只做文档约定，必须有真实 handler 和真实测试

### 4. 对齐 `/dashboard/audit` 与 Audit 契约

目标是让 Audit 页面从“有 UI 但把缺失后端吞掉”升级到“真实消费后端审计数据”。

至少覆盖：

- 列表加载
- 分页或最小分页契约接入
- 结果筛选
- 空状态
- 错误态

要求：

- 页面应消费真实 `/api/v2/audit/logs`
- 当前“404 时静默兜底为空列表”的逻辑不能再作为主路径设计
- 可保留温和错误处理，但不能掩盖主链路缺失

## 禁止越界项

Windsurf 在本 Sprint 中不得进入以下内容：

- 不实现 CLI `audit-log`
- 不扩展 Agent Run 页面
- 不实现 Audit 导出
- 不开始 Workspace/Audit retention 治理增强
- 不开始 Evidence 编辑
- 不开始 PDF / SARIF 导出
- 不开始 Scanner Hub / Burp / Yakit / Nuclei 集成
- 不对前端做大规模导航重构

如果为了完成 Sprint E 需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作：

### 前端

- `frontend-next/src/app/dashboard/evidence/page.tsx`
- `frontend-next/src/app/dashboard/audit/page.tsx`
- `frontend-next/src/lib/api.ts`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/components/timeline.tsx`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/evidence.spec.ts`
- `frontend-next/e2e/*audit*.spec.ts` 或新增 audit 专项 e2e

### 后端

- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/auth/service.go`
- `internal/models/audit.go`

如确有必要，可补充：

- `frontend-next/src/features/*`

但只允许为 Evidence / Audit 页面收口服务，不允许借机扩展其他控制面能力。

## 建议实施顺序

Windsurf 应按以下顺序推进：

1. 先补 `/api/v2/audit/logs` 的后端 handler 和测试
2. 再收口前端 Evidence 页面到统一 Evidence 契约
3. 再收口 Audit 页面到真实后端契约
4. 最后补 E2E 验证

## 必须补齐的测试

### 1. Audit API 测试

至少覆盖：

- `/api/v2/audit/logs` 成功返回分页结构
- `action` / `resource_type` 过滤生效
- 时间范围过滤生效
- 未登录或无权限时返回稳定错误

### 2. Evidence 页面前端行为测试

至少覆盖：

- 页面可加载 case 列表
- 选择 case 后可生成 Evidence
- 页面展示 strength / confidence / explainability
- JSON / Markdown 导出按钮可触发正确请求或正确下载行为
- no evidence 空状态可见

### 3. Audit 页面前端行为测试

至少覆盖：

- 页面可加载真实 audit 数据
- 过滤条件会影响列表显示
- 空状态可见
- 接口失败时显示稳定错误提示，不再静默吞掉主链路错误

### 4. 全量验证

至少执行：

- `GOCACHE=/tmp/gocache go test ./server ./internal/auth ./internal/models`
- `GOCACHE=/tmp/gocache go test ./...`
- `cd frontend-next && yarn test:e2e evidence.spec.ts`
- `cd frontend-next && yarn test:e2e` 或等价可证明 Evidence/Audit 页面未回归的命令

## Commit 规则

这是本 Sprint 的强制协作规则，Windsurf 必须遵守：

1. **每完成一个任务，必须对本次新增/修改代码执行一次本地 commit**
2. **只需要 commit，不需要 push**
3. commit 必须是小步提交，和实施顺序对应
4. 回传结果时必须列出：
   - 本 Sprint 产生了哪些 commit
   - 每个 commit 对应哪个任务

Codex 验收时会把“是否按任务完成本地 commit”作为过程合规项检查。

## 完成定义

只有同时满足以下条件，Sprint E 才能视为完成：

1. `/dashboard/evidence` 已真实消费统一 Evidence 契约
2. Evidence 页面能展示摘要、explainability、timeline，并支持 JSON / Markdown 导出
3. `/api/v2/audit/logs` 已形成最小可用查询契约
4. `/dashboard/audit` 已真实消费 Audit 后端数据，而不是以静默兜底为空列表为主路径
5. Evidence / Audit 相关测试通过
6. `GOCACHE=/tmp/gocache go test ./...` 继续通过
7. Windsurf 已按任务完成本地 commit

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式向 Codex 回传：

### 1. 实际修改范围

- 修改了哪些目录
- 修改了哪些关键文件
- 哪些计划内文件未动，原因是什么

### 2. 实际实现内容

- Evidence 页面如何切换到统一 Evidence 契约
- Audit API 如何设计与实现
- Audit 页面如何对齐后端契约

### 3. 实际验证命令

至少包含实际执行过的命令，例如：

- `GOCACHE=/tmp/gocache go test ./server ./internal/auth ./internal/models`
- `GOCACHE=/tmp/gocache go test ./...`
- `cd frontend-next && yarn test:e2e evidence.spec.ts`
- `cd frontend-next && yarn test:e2e`

### 4. Commit 列表

必须列出：

- commit hash
- commit message
- 对应任务编号

### 5. 测试结果

- 新增了哪些测试
- 修改了哪些测试
- 哪些预期测试仍未覆盖

### 6. 风险与偏差

- 哪些地方与原规划不完全一致
- 哪些点被明确留给 Sprint F
- 哪些旧页面/旧契约桥接仍然存在

## Codex 验收问题

Codex 在验收 Sprint E 时只围绕以下问题判断：

1. Evidence 页面是否真正消费统一 Evidence 契约，而不是旧 interaction export
2. Evidence 页面是否已经具备 MVP 级摘要、时间线和导出能力
3. Audit API 是否形成稳定、可分页、可筛选的 `/api/v2` 契约
4. Audit 页面是否真正消费后端数据，而不是以静默兜底为空为主路径
5. Evidence / Audit 页面和后端测试是否足够证明闭环成立
6. Windsurf 是否按任务完成本地 commit

## 验收结论类型

Codex 对 Sprint E 的验收只会给出三种结果：

- **通过**：可进入 Sprint F
- **有条件通过**：允许进入 Sprint F，但必须挂明遗留项
- **不通过**：必须继续停留在 Sprint E 修正

## Sprint E 完成后的下一步

只有 Sprint E 被 Codex 验收通过后，才进入：

- `Sprint F：首批控制面页面收口`
