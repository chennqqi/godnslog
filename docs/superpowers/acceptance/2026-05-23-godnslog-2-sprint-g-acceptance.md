# GODNSLOG 2.0 Sprint G 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-23-godnslog-2-sprint-g-package.md`
- `frontend-next/src/app/dashboard/interactions/page.tsx`
- `frontend-next/src/app/dashboard/evidence/page.tsx`
- `frontend-next/src/app/dashboard/cases/[id]/page.tsx`
- `frontend-next/src/app/dashboard/payloads/[id]/page.tsx`
- `frontend-next/e2e/cases.spec.ts`
- `frontend-next/e2e/interactions.spec.ts`
- `frontend-next/e2e/evidence.spec.ts`

## 验收结论

**结论：Sprint G 通过验收，可以关闭。**

Windsurf 最新相关提交：

- `ed4591e sprint-g: 修复 lint 错误并完成验证`
- `6c6748f sprint-g: 实现 Interaction Triage 与 Evidence 贯通`

Codex 复验补充：

- `frontend-next/e2e/interactions.spec.ts`：补充请求级断言，确认 `case_id` / `payload_id` 真实进入 `/api/v2/interactions` 和 `/api/v2/interactions/stats`
- `frontend-next/e2e/evidence.spec.ts`：补充请求级断言，确认 `case_id` / `payload_id` 和 `format` 真实进入 `/api/v2/evidence/generate` 请求体

## 验收点判断

### 1. URL scope 是否真实进入 API 请求

通过。

`Interactions` 页面读取 URL query 后传入：

- `interactionApi.list({ case_id, payload_id, page, page_size })`
- `interactionApi.stats({ case_id, payload_id })`

Codex 已补充 E2E 请求断言，验证：

- `/dashboard/interactions?case_id=case-1` 会请求 `/api/v2/interactions?...case_id=case-1`
- `/dashboard/interactions?case_id=case-1` 会请求 `/api/v2/interactions/stats?...case_id=case-1`
- `/dashboard/interactions?payload_id=payload-1` 会请求 `/api/v2/interactions?...payload_id=payload-1`
- `/dashboard/interactions?payload_id=payload-1` 会请求 `/api/v2/interactions/stats?...payload_id=payload-1`

### 2. stats 是否随 scope 变化

通过。

`Interactions` 页面 stats 请求已携带当前 scope。Codex 补充 E2E mock，让 `payload_id=payload-1` 的 stats 返回 `total=2`，并断言页面统计区域渲染该 scoped total。

### 3. `Payload Detail -> Interactions -> Evidence` 是否形成闭环

通过。

已验证：

- `Payload Detail -> 查看交互` 跳转到 `/dashboard/interactions?payload_id=payload-1`
- `Payload Detail -> 查看证据` 跳转到 `/dashboard/evidence?payload_id=payload-1`
- Interaction triage 面板展示 `case_id` / `payload_id`
- Interaction triage 面板提供 `Generate Evidence (Case)` / `Generate Evidence (Payload)` 动作

### 4. Evidence 是否继续使用统一 Evidence 契约

通过。

Evidence 页面继续调用：

```text
POST /api/v2/evidence/generate
```

并消费统一响应中的：

- `data.evidence`
- `data.content`
- `evidence_strength`
- `confidence`
- `interaction_count`
- `unique_sources`
- `explainability`
- `timeline`

Codex 已补充 E2E 请求体断言：

- `case_id` scope 发送 `{ case_id: "case-1", format: "markdown" }`
- `payload_id` scope 发送 `{ payload_id: "payload-1", format: "markdown" }`
- URL `format=json` 会发送 `{ case_id: "case-1", format: "json" }`

### 5. E2E 是否没有 skip 或“只检查静态文字”的空测

通过。

`rg -n "test\\.skip|skip\\(|describe\\.skip|only\\(" frontend-next/e2e/cases.spec.ts frontend-next/e2e/interactions.spec.ts frontend-next/e2e/evidence.spec.ts` 无匹配。

Codex 复验时发现 Windsurf 原始新增 E2E 偏静态文字断言，因此已补充请求级断言，避免只检查页面文字。

### 6. 是否严格没有越界到 Scanner Hub、生命周期治理或批量操作

通过。

本轮代码改动集中在：

- Interactions 页面 URL scope、stats scope、triage 面板
- Evidence 页面 URL scope 自动生成
- Payload Detail 到 Evidence 的 payload scope 导航
- E2E 覆盖

未发现 Sprint G 改动进入 Scanner Hub、Payload 生命周期治理、Interaction 批量删除/标记或批量操作。

## 已执行验证

```bash
git status --short --untracked-files=all
git log --oneline -12
git show --stat --name-only --oneline HEAD
git show --stat --name-only --oneline 6c6748f
git diff ae713f2..HEAD -- frontend-next/src/app/dashboard/interactions/page.tsx frontend-next/src/app/dashboard/evidence/page.tsx frontend-next/src/app/dashboard/cases/[id]/page.tsx frontend-next/src/app/dashboard/payloads/[id]/page.tsx frontend-next/src/lib/api-client.ts frontend-next/src/types/index.ts
rg -n "test\\.skip|skip\\(|describe\\.skip|only\\(" frontend-next/e2e/cases.spec.ts frontend-next/e2e/interactions.spec.ts frontend-next/e2e/evidence.spec.ts
rg -n "Scanner|Nuclei|Burp|Yakit|ZAP|xray|revoke|expire|delete|batch|lifecycle|生命周期|批量|删除|scanner" frontend-next/src/app/dashboard frontend-next/e2e docs/superpowers/acceptance docs/superpowers/plans/2026-05-23-godnslog-2-sprint-g-package.md
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/interactions/page.tsx src/app/dashboard/evidence/page.tsx src/app/dashboard/cases/[id]/page.tsx src/app/dashboard/payloads/[id]/page.tsx e2e/cases.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts playwright.config.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/cases.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
rm -rf frontend-next/test-results
```

结果：

- `go test ./...` 通过
- `npm run build` 通过
- Sprint G 定向 `eslint` 通过
- 首轮 Playwright：`50 passed (2.0m)`
- 补充请求级 E2E 断言后复跑 Playwright：`52 passed (2.0m)`
- 已停止手动启动的 dev server
- 已清理 `frontend-next/test-results`

## 遗留说明

- Playwright `webServer` 单命令自动启动在 Codex 环境仍作为 Sprint F 已记录技术债处理；Sprint G 验收按 `docs/verification.md` 的两步法执行。
- 全量前端 `npm run lint` 的历史问题不属于 Sprint G 范围；本轮 Sprint G 定向 lint 已通过。

## 最终判断

Sprint G 已满足完成定义：

1. Case / Payload 进入 Interactions 不丢失上下文
2. Interactions list 与 stats 都支持 scope
3. Interaction triage 面板支持判断、复制和跳转
4. Case / Payload 进入 Evidence 能自动生成对应 Evidence
5. Evidence 页面继续使用统一 Evidence 契约
6. Sprint F 主链路未回归
7. E2E 无 skip，并已补充请求级断言
8. 后端测试、前端 build、Sprint G 定向 lint、Sprint G E2E 均通过

**Sprint G 正式验收通过，可以关闭。**
