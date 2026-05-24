# GODNSLOG 2.0 Sprint H Implementation Package

> **协作模式**
>
> - **Codex**：负责本实施包的规划、边界控制、验收
> - **Windsurf**：负责按本实施包进行具体开发、自测、提交本地 commit、回传结果

## Sprint 标识

- **Sprint 名称**：Sprint H
- **Sprint 主题**：Scanner Hub MVP - Nuclei JSONL Closed Loop
- **所属阶段**：Phase 4 - Primary Scanner Integration

## Sprint 背景

Sprint F 已完成控制面 `Case -> Payload` 主链路，Sprint G 已完成 `Payload Detail -> Interactions -> Evidence` 的上下文贯通。当前产品已经能完成手工 OAST 验证闭环：

`Case -> Payload -> Interaction -> Evidence`

Sprint H 开始进入 Scanner Hub，但必须从最小可验收场景开始：**Nuclei / JSONL / CLI 文档化闭环**。本 Sprint 不做完整扫描器平台，不做多工具插件市场，不做 Burp/Yakit/ZAP 深度集成。

Sprint H 的核心问题是：

安全团队能否把 GODNSLOG 生成的可追踪 Payload 分发给 Nuclei 场景，并把扫描侧结果以统一 Case/Payload/Interaction/Evidence 语义回到 GODNSLOG。

## Sprint 目标

本 Sprint 只聚焦 4 件事：

1. 建立 Scanner Hub MVP 数据契约，描述一次 scanner run 如何关联 Case、Payload、Interaction、Evidence
2. 提供 Nuclei JSONL / template variable 最小可执行集成路径
3. 提供控制面 Scanner Hub MVP 页面或现有页面入口，让用户能生成、复制、查看 Nuclei 集成命令和 JSONL 输出
4. 建立 `Scanner Hub -> Payload -> Interactions -> Evidence` 的可验收闭环

本 Sprint 的目标不是“执行真实 Nuclei 进程”。真实外部 scanner 执行可以作为后续 Sprint；Sprint H 先把统一契约、输出格式、页面入口和测试闭环做实。

## 输入文档

Windsurf 实施前必须完整阅读以下文档：

- `docs/unified-terminology.md`
- `docs/mvp-closed-loop.md`
- `docs/implementation-dependencies.md`
- `docs/scanner-hub.md`
- `docs/scanner-hub-adapter-design.md`
- `docs/official-support-boundary.md`
- `docs/CLI_USAGE.md`
- `examples/nuclei/README.md`
- `docs/superpowers/plans/2026-05-23-godnslog-2-sprint-g-package.md`
- `docs/superpowers/acceptance/2026-05-23-godnslog-2-sprint-g-acceptance.md`

## 当前现状判断

### Scanner Hub 文档现状

- `docs/scanner-hub.md` 已定义基础 Create Probe / Wait For Result / Result Formats
- `docs/scanner-hub-adapter-design.md` 已定义成熟度模型
- `docs/official-support-boundary.md` 已将 Nuclei 定为 Primary，并明确 MVP 目标是 JSONL / official script
- `examples/nuclei/README.md` 已存在基础说明

问题：

- 文档仍偏静态说明，缺少“从 Case 生成 scanner payload、复制 Nuclei 命令、返回 GODNSLOG 证据”的具体闭环
- `docs/scanner-hub.md` 中仍有旧字段口径，例如 `expires_in` / `expected_protocols`，需要和当前统一 Payload 契约核对，避免继续固化双轨契约
- 当前控制面没有 Scanner Hub 专用工作台

### 前端现状

- `/dashboard/marketplace` 是插件/模板市场占位，不适合承载 Scanner Hub MVP 主路径
- `/dashboard/payloads/new` 可创建 Payload，但不是 scanner 视角的集成入口
- `/dashboard/interactions` 和 `/dashboard/evidence` 已能接住 scanner payload 后续闭环

### 后端现状

已有基础实体和 API：

- `Case`
- `Payload`
- `Interaction`
- `Evidence`
- `Audit`
- `/api/v2/payloads`
- `/api/v2/interactions`
- `/api/v2/evidence/generate`

Sprint H 优先复用这些契约。只有为了表达 scanner run / JSONL 输出确有缺口时，才允许新增小型后端模型和 API。

## 术语边界

Sprint H 中的 Scanner Hub MVP 只引入一个轻量概念：

### Scanner Run

Scanner Run 表示一次面向外部 scanner 的分发上下文，不替代 Case，也不替代 Payload。

建议最小字段：

- `id`
- `case_id`
- `payload_id`
- `scanner`：固定支持 `nuclei`
- `target`
- `template`
- `delivery_method`：`nuclei-jsonl` 或 `nuclei-var`
- `command`
- `jsonl`
- `status`：`created | distributed | observed | evidenced`
- `created_at`
- `updated_at`

要求：

- Scanner Run 必须引用已有 Case 和 Payload
- Interaction 仍归因到 Payload / Case
- Evidence 仍通过统一 Evidence 契约生成
- 不建立与 Case/Payload/Interaction 平行的新证据模型

如果实现成本过高，Sprint H 可先不持久化 Scanner Run，但必须在文档和前端状态中保持字段语义稳定，并明确后续持久化边界。

## 实施范围

本 Sprint 只允许覆盖以下 4 个主题。

### 1. 收口 Scanner Hub MVP 契约与文档

目标是把 Nuclei 集成从“散落说明”收口为可执行的最小流程。

至少覆盖：

- 更新 `docs/scanner-hub.md`
- 更新或补充 `examples/nuclei/README.md`
- 明确 Nuclei MVP 支持两种分发方式：
  - template variable：`-var godnslog_payload=<rendered_payload>`
  - JSONL：一行一个 scanner probe record
- 明确 JSONL 最小字段：
  - `scanner`
  - `case_id`
  - `payload_id`
  - `token`
  - `target`
  - `template`
  - `rendered_payload`
  - `interactions_url`
  - `evidence_url`
  - `created_at`
- 明确 Wait / Evidence 读取方式：
  - `/api/v2/interactions?payload_id=...`
  - `/dashboard/interactions?payload_id=...`
  - `/dashboard/evidence?payload_id=...`

要求：

- 文档不得继续把未实现能力写成已完成
- 字段必须和当前统一契约一致，避免重新引入已废弃或旁路字段
- Nuclei 之外的 Burp/Yakit/ZAP/xray/Postman 只允许保留为支持矩阵，不进入实现细节

### 2. 提供 Scanner Hub MVP 控制面入口

目标是让用户不需要手工拼接命令，就能从控制面生成 Nuclei 分发材料。

建议实现为新增页面：

- `frontend-next/src/app/dashboard/scanner-hub/page.tsx`

至少覆盖：

- 选择或输入 Case
- 输入 target
- 选择 Nuclei template 类型：
  - `ssrf-basic`
  - `xxe-basic`
  - `rce-callback`
- 创建或选择 Payload
- 展示：
  - token
  - rendered payload
  - Nuclei command
  - JSONL preview
  - Interactions link
  - Evidence link

要求：

- 页面是工作台，不是插件市场，不是营销页
- 必须复用已有 Case/Payload API，不造一套 scanner-only payload
- 生成的 Interactions/Evidence 链接必须携带 `payload_id`
- Copy 按钮必须针对 command、JSONL、rendered payload
- 空态和错误态可验收

如果新增 sidebar 导航代价过高，可先通过 `/dashboard/scanner-hub` 路由和文档入口验收；不强制做全局导航重排。

### 3. 提供 Nuclei JSONL / command 生成逻辑

目标是形成稳定、可测、可复用的前端或后端 helper。

至少覆盖：

- 根据 Case / Payload / target 生成 Nuclei command
- 根据 Case / Payload / target 生成 JSONL record
- JSONL 必须是一行合法 JSON
- command 中必须包含 rendered payload 或 token 变量
- 输出中必须包含回到 GODNSLOG 的 Interactions / Evidence URL

建议优先实现为前端纯函数或轻量后端 service：

- 前端：`frontend-next/src/lib/scanner-hub.ts`
- 或后端：`internal/scannerhub/`

选择原则：

- 如果只生成 command/JSONL，可先放前端 helper，并配前端单元/E2E 验证
- 如果需要持久化 Scanner Run，则放后端 `internal/scannerhub/` 并配 Go 测试

### 4. 建立 Scanner Hub 到 Sprint G 闭环

目标是让 scanner 入口不是孤立页面，而是能回到已验收的 Interactions / Evidence 工作流。

至少覆盖：

- Scanner Hub 生成 Payload 后，能直接打开 `/dashboard/interactions?payload_id=...`
- Scanner Hub 生成 Payload 后，能直接打开 `/dashboard/evidence?payload_id=...`
- 如果 mock 有 Interaction，Evidence 能基于 payload scope 生成
- 页面明确当前 scanner output 对应的 Case / Payload

要求：

- 不实现真实 Nuclei 进程管理
- 不轮询真实外部 scanner
- 不实现任务队列
- 不实现 scanner result 上传大平台

## 禁止越界项

Windsurf 在 Sprint H 中不得进入以下内容：

- 不实现 Burp Suite、Yakit/Yak、ZAP、xray/rad、Postman/Apifox 的实际适配
- 不实现 Nuclei 进程启动、调度、取消、并发队列
- 不实现插件市场或 adapter marketplace
- 不实现 SARIF 导出
- 不实现 Scanner Run 的复杂生命周期治理
- 不实现多 workspace 权限模型
- 不重构全局 dashboard shell
- 不修复全量前端 lint 历史债
- 不把 Playwright webServer 自动启动兼容性作为 Sprint H 功能任务
- 不改变 Sprint F/G 已验收的 Case/Payload/Interactions/Evidence 主链路

如果为了完成 Sprint H 需要触碰上述内容，必须先回传 Codex 重新裁剪。

## 建议修改范围

Windsurf 优先在以下文件和目录内工作。

### 文档与示例

- `docs/scanner-hub.md`
- `docs/official-support-boundary.md`
- `docs/CLI_USAGE.md`
- `examples/nuclei/README.md`
- `examples/nuclei/*`

### 前端

- `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- `frontend-next/src/lib/scanner-hub.ts`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/scanner-hub.spec.ts`

如确有必要，可轻触：

- `frontend-next/src/components/app-shell/index.tsx`
- `frontend-next/src/lib/i18n.ts`

但只允许增加 Scanner Hub 入口，不允许借机重构全局导航。

### 后端可选

仅当需要持久化 Scanner Run 或提供服务端生成接口时，才允许新增：

- `internal/scannerhub/`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/models/`

后端新增必须有 Go 测试。

## 建议实施顺序

Windsurf 应按以下顺序推进：

1. 先收口 `docs/scanner-hub.md` 和 Nuclei JSONL 字段契约
2. 再实现 command / JSONL 生成 helper
3. 再新增 Scanner Hub MVP 页面
4. 再把页面链接接到 `/dashboard/interactions?payload_id=...` 和 `/dashboard/evidence?payload_id=...`
5. 最后补 E2E 和必要单元/后端测试

不要先做真实 scanner 执行，也不要先做多扫描器 UI。

## 必须补齐的测试

### 1. Scanner Hub helper 测试

至少覆盖：

- 生成的 JSONL 是合法单行 JSON
- JSONL 包含 `scanner=nuclei`
- JSONL 包含 `case_id`、`payload_id`、`token`、`target`、`template`、`rendered_payload`
- JSONL 包含 `interactions_url` 和 `evidence_url`
- Nuclei command 包含 payload/token 变量

如果 helper 在前端实现，可用现有前端测试工具或 E2E 请求/页面断言覆盖；如果 helper 在后端实现，必须有 Go 测试。

### 2. Scanner Hub 页面 E2E

至少覆盖：

- 页面可加载
- 用户能选择 Case 或输入 Case scope
- 用户能输入 target
- 用户能生成或选择 Payload
- 页面显示 token、rendered payload、Nuclei command、JSONL preview
- Copy 按钮存在并针对 command / JSONL / payload
- Interactions 链接包含 `payload_id`
- Evidence 链接包含 `payload_id`
- 空态和 API 错误态稳定可见

### 3. Scanner Hub 到 Sprint G 闭环 E2E

至少覆盖：

- 从 Scanner Hub 点击 Interactions 后进入 `/dashboard/interactions?payload_id=...`
- 从 Scanner Hub 点击 Evidence 后进入 `/dashboard/evidence?payload_id=...`
- Evidence 页面发送 `/api/v2/evidence/generate` 请求体时包含 `payload_id`

### 4. 文档验收

至少覆盖：

- `docs/scanner-hub.md` 不再宣称未实现的多工具深度集成已完成
- `examples/nuclei/README.md` 给出可复制命令
- 字段口径和当前 Payload/Evidence API 一致

## 验证命令

完成 Sprint H 前必须记录实际运行命令和结果。最低要求：

```bash
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
```

UI 行为验证按 `docs/verification.md` 当前约定执行 Sprint H 相关 Playwright 测试。由于 Playwright webServer 自动启动在不同环境中仍有兼容性技术债，Sprint H 验收允许采用两步法：

```bash
cd frontend-next
npm run dev
npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
```

要求：

- 禁止使用 `npx playwright show-report`
- 禁止使用会常驻 HTML report server 的流程
- 测试结束后停止 dev server
- 清理 `frontend-next/test-results`

如果 Sprint H 未新增后端代码，可不新增后端专项包测试，但仍必须跑 `go test ./...`。

## 完成定义

只有同时满足以下条件，Sprint H 才能视为完成：

1. Scanner Hub MVP 契约已收口到 Nuclei JSONL / template variable
2. JSONL 和 command 输出可复制、字段稳定、可测试
3. Scanner Hub 页面能基于 Case/Payload 生成 Nuclei 分发材料
4. Scanner Hub 输出明确关联 Case 和 Payload
5. Scanner Hub 能跳转到 payload-scoped Interactions / Evidence
6. Evidence 仍复用统一 Evidence 契约
7. E2E 覆盖 Scanner Hub 到 Sprint G 的闭环且无 `test.skip`
8. 未越界到多扫描器深度集成、插件市场、真实 scanner 调度或 SARIF
9. `go test ./...`、前端 build、Sprint H 定向 lint 通过

## Windsurf 回传模板

Windsurf 完成实施后，必须按以下格式回传：

```text
Sprint H 完成回传

本地 commits:
- <hash> <title>
- <hash> <title>

实现摘要:
- Scanner Hub contract:
- Nuclei JSONL/command:
- Scanner Hub page:
- Interactions/Evidence closed loop:

测试结果:
- GOCACHE=/tmp/gocache go test ./...: <pass/fail>
- cd frontend-next && npm run build: <pass/fail>
- Sprint H 定向 eslint: <pass/fail>
- Playwright Sprint H specs: <pass/fail, passed/failed 数量>

已知遗留:
- <如无，写 无>

需要 Codex 复验的点:
- <列出最重要的 3-5 个验收点>
```

## Codex 验收重点

Codex 复验 Sprint H 时重点检查：

1. JSONL 是否是合法单行 JSON，且字段和统一契约一致
2. Nuclei command 是否真实包含可用 payload/token 变量
3. Scanner Hub 是否复用 Case/Payload API，而不是造 scanner-only payload
4. Interactions/Evidence 链接是否携带 `payload_id`
5. Evidence 请求是否继续走 `/api/v2/evidence/generate`
6. E2E 是否断言 command/JSONL/request/link，而不是只检查静态标题
7. 是否严格没有越界到 Burp/Yakit/ZAP/xray、插件市场、真实 scanner 调度、SARIF 或生命周期治理

只有上述全部成立，Sprint H 才能关闭。
