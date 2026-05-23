# GODNSLOG 2.0 Sprint E 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-22-godnslog-2-sprint-e-package.md`
- `frontend-next/src/app/dashboard/evidence/page.tsx`
- `frontend-next/src/app/dashboard/audit/page.tsx`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `server/v2_api.go`
- `server/v2_api_test.go`
- `internal/auth/service.go`
- `internal/models/audit.go`
- `frontend-next/e2e/evidence.spec.ts`
- `frontend-next/e2e/audit.spec.ts`

## 验收结论

**结论：有条件通过。**

从实现范围、后端契约、前端接线和提交流程看，Sprint E 的主目标已经完成，可以进入下一阶段。  
但本次前端 E2E 没有在当前环境完成“绿色验证”：

- Playwright 缺少 Chromium 浏览器二进制
- `next build` 在当前沙箱环境触发 Turbopack 运行时限制

因此我给出的结论不是“完全通过”，而是**有条件通过**：功能范围和代码走向可以放行，但需要 Windsurf 在可运行浏览器和可构建前端的环境里补一次前端验证记录。

## 本次完成点

### 1. Evidence 页面已切换到统一 Evidence 契约

当前 `/dashboard/evidence` 已不再走旧的 `interactionApi.export(...)`，而是改为调用：

- [frontend-next/src/app/dashboard/evidence/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/evidence/page.tsx:4)
- [frontend-next/src/lib/api-client.ts](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/lib/api-client.ts:113)

Evidence 页面现在直接消费：

- `data.evidence`
- `data.content`

已不再暴露：

- `csv`
- `include_raw`

这满足了 Sprint E 对 Evidence 页面主链路的要求。

### 2. Evidence 页面具备 MVP 级摘要、时间线、导出

当前页面已展示：

- `evidence_strength`
- `confidence`
- `interaction_count`
- `unique_sources`
- `explainability`
- timeline
- JSON / Markdown 下载

见：

- [frontend-next/src/app/dashboard/evidence/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/evidence/page.tsx:101)

### 3. `/api/v2/audit/logs` 后端契约已真实落地

当前 `/api/v2` 已注册 audit 路由：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:128)

`v2ListAuditLogs` 已支持：

- `page`
- `page_size`
- `user_id`
- `action`
- `resource_type`
- `start_time`
- `end_time`

见：

- [server/v2_api.go](/data/dev/github.com/chennqqi/godnslog/server/v2_api.go:3035)

并复用了已有的：

- [internal/auth/service.go](/data/dev/github.com/chennqqi/godnslog/internal/auth/service.go:220)

### 4. Audit 页面已对齐真实后端契约

当前 `/dashboard/audit` 已不再直接调用裸 `api.get('/audit/logs')`，而是通过：

- [frontend-next/src/app/dashboard/audit/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/audit/page.tsx:13)
- [frontend-next/src/lib/api-client.ts](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/lib/api-client.ts:117)

同时页面已显式展示错误态，不再把“后端缺失时静默吞掉”作为主路径设计：

- [frontend-next/src/app/dashboard/audit/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/audit/page.tsx:172)

### 5. Windsurf 已按任务完成本地 commit

本次可见的 Sprint E 提交有：

- `fd7c47b sprint-e: add /api/v2/audit/logs endpoint and tests`
- `f5007cf sprint-e: migrate Evidence page to unified Evidence API`
- `69601d3 sprint-e: migrate Audit page to real backend API`
- `01d7a28 sprint-e: add E2E tests for Evidence and Audit pages`

这满足了本 Sprint 额外增加的过程合规要求。

## 测试与验证

### 已通过

已执行：

```bash
GOCACHE=/tmp/gocache go test ./server ./internal/auth ./internal/models
GOCACHE=/tmp/gocache go test ./...
```

结果通过。

### 已补齐但当前环境未能完成

已尝试执行：

```bash
cd frontend-next && npx playwright test e2e/evidence.spec.ts e2e/audit.spec.ts
```

结果失败，原因不是业务断言失败，而是环境缺少 Playwright 浏览器：

- `Executable doesn't exist ... chrome-headless-shell`
- Playwright 明确要求执行 `npx playwright install`

另外我也尝试了：

```bash
cd frontend-next && npm run build
```

结果失败，但错误来自当前沙箱/Turbopack 环境限制：

- `Operation not permitted (os error 1)`
- 触发点在 Turbopack 处理 `globals.css` 时创建新进程/绑定端口

因此这两项当前不能作为代码不通过的直接证据。

## 对 6 个验收问题的判断

1. Evidence 页面是否真正消费统一 Evidence 契约，而不是旧 interaction export：**是**
2. Evidence 页面是否已经具备 MVP 级摘要、时间线和导出能力：**是**
3. Audit API 是否形成稳定、可分页、可筛选的 `/api/v2` 契约：**是**
4. Audit 页面是否真正消费后端数据，而不是以静默兜底为空为主路径：**是**
5. Evidence / Audit 页面和后端测试是否足够证明闭环成立：**后端是；前端 E2E 在当前环境未完成**
6. Windsurf 是否按任务完成本地 commit：**是**

## 结论建议

Sprint E 可按 **有条件通过** 处理，并进入下一阶段：

- `Sprint F：首批控制面页面收口`

但需要 Windsurf 补一条收口动作：

1. 在具备 Playwright 浏览器的环境执行：
   - `cd frontend-next && npx playwright install`
   - `cd frontend-next && yarn test:e2e evidence.spec.ts audit.spec.ts`
2. 回传一次前端验证结果

这条作为 Sprint E 的验证遗留项挂账，不阻塞 Sprint F 规划启动。
