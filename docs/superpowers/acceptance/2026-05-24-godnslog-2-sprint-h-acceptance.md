# GODNSLOG 2.0 Sprint H 验收结论

## 验收对象

- `docs/superpowers/plans/2026-05-24-godnslog-2-sprint-h-package.md`
- `docs/scanner-hub.md`
- `examples/nuclei/README.md`
- `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- `frontend-next/src/lib/scanner-hub.ts`
- `frontend-next/src/lib/api-client.ts`
- `frontend-next/src/types/index.ts`
- `frontend-next/src/features/users/api.ts`
- `frontend-next/e2e/scanner-hub.spec.ts`
- `frontend-next/e2e/interactions.spec.ts`
- `frontend-next/e2e/evidence.spec.ts`

## 验收结论

**结论：Sprint H 通过验收，可以关闭。**

Windsurf 本轮以未提交工作区形式完成 Sprint H，Codex 复验时做了以下补充修正：

- 移除 `scanner-hub.spec.ts` 中大量 `test.skip`，改为真实执行 Scanner Hub 主链路 E2E
- 补强 E2E 断言，覆盖 JSONL 字段、Nuclei command、统一 Payload API 请求、payload-scoped Interactions/Evidence 链接、统一 Evidence 契约请求
- 修正 `scanner-hub.ts` 的 Nuclei command shell quoting
- 将 JSONL `evidence_url` 调整为 payload-scoped Evidence URL
- 将 `docs/scanner-hub.md` 中旧 `expires_in` 口径收口到当前统一 Payload 契约的 `expires_at` / `expected_protocol`
- 小范围收口 `api-client.ts` / `types/index.ts` 既有类型问题，使 Sprint H 定向 lint 和生产构建通过

## 验收点判断

### 1. JSONL 是否合法且字段和统一契约一致

通过。

`frontend-next/src/lib/scanner-hub.ts` 生成单行 JSONL，E2E 已解析 JSON 并断言最小字段：

- `scanner: "nuclei"`
- `case_id`
- `payload_id`
- `token`
- `target`
- `template`
- `rendered_payload`
- `interactions_url`
- `evidence_url`
- `created_at`

### 2. Nuclei command 是否真实包含 payload/token 变量

通过。

E2E 已断言 command 包含：

- `nuclei -u 'https://target.example'`
- `-t godnslog-ssrf-basic.yaml`
- `godnslog_payload=http://tok-abc123.example.com/callback`

Command 参数已做 shell quoting，避免 target 或 payload 中的特殊字符破坏命令结构。

### 3. Scanner Hub 是否复用 Case/Payload API

通过。

Scanner Hub 页面使用：

- `caseApi.list`
- `payloadApi.list`
- `payloadApi.create`

E2E 已通过 `waitForRequest` 断言创建 Payload 时请求：

```text
POST /api/v2/payloads
```

且请求体包含：

- `case_id: "case-1"`
- `template: "ssrf-basic"`

未发现 scanner-only payload 或平行实体模型。

### 4. Interactions/Evidence 链接是否携带 `payload_id`

通过。

Scanner Hub 输出链接已指向：

- `/dashboard/interactions?payload_id=payload-1`
- `/dashboard/evidence?payload_id=payload-1`

E2E 已覆盖点击跳转。

### 5. Evidence 请求是否继续走统一契约

通过。

从 Scanner Hub 进入 Evidence 后，仍由 Sprint G 的 Evidence 页面调用：

```text
POST /api/v2/evidence/generate
```

E2E 已断言请求体包含：

- `payload_id: "payload-1"`
- `format: "markdown"`

### 6. E2E 是否不是静态标题空测

通过。

初始实现存在大量 `test.skip`，Codex 已移除并重写为真实链路测试。当前：

```bash
rg -n "test\\.skip|skip\\(|describe\\.skip|only\\(" frontend-next/e2e/scanner-hub.spec.ts frontend-next/e2e/interactions.spec.ts frontend-next/e2e/evidence.spec.ts
```

无匹配。

### 7. 是否严格没有越界

通过。

本轮未实现：

- Burp/Yakit/ZAP/xray/Postman/Apifox 实际适配
- Nuclei 进程启动、调度、取消、并发队列
- 插件市场或 adapter marketplace
- SARIF 导出
- Scanner Run 复杂生命周期治理

Sprint H 保持在 Nuclei JSONL / template variable / 控制面分发材料 / payload-scoped Interactions/Evidence 闭环内。

## 已执行验证

```bash
git status --short --untracked-files=all
git log --oneline -12
rg -n "test\\.skip|skip\\(|describe\\.skip|only\\(" frontend-next/e2e/scanner-hub.spec.ts frontend-next/e2e/interactions.spec.ts frontend-next/e2e/evidence.spec.ts
rg -n "expires_in|expected_protocols|tool|SARIF|Burp|Yakit|ZAP|xray|Postman|Apifox|plugin|marketplace|调度|队列|进程" docs/scanner-hub.md examples/nuclei/README.md frontend-next/src/app/dashboard/scanner-hub/page.tsx frontend-next/src/lib/scanner-hub.ts frontend-next/e2e/scanner-hub.spec.ts
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npm run build
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
cd frontend-next && npm run dev
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/interactions.spec.ts e2e/evidence.spec.ts
rm -rf frontend-next/test-results
```

结果：

- `go test ./...` 通过
- `npm run build` 通过
- Sprint H 定向 `eslint` 通过
- 首轮 Playwright：`27 passed, 4 failed`，失败均为 E2E 断言写法问题
- 修正 E2E 断言后复跑 Playwright：`32 passed (1.1m)`
- 已停止手动启动的 dev server
- 已清理 `frontend-next/test-results`

## 遗留说明

- Sprint H 未实现真实 Nuclei 进程调度、队列、取消、SARIF 或多扫描器深度适配，这符合 Sprint H plan 的禁止越界项。
- Scanner Run 当前为前端 in-memory 对象，未持久化。后续如需要历史记录、审计或运行状态治理，应单独开 Sprint。
- Playwright `webServer` 单命令自动启动在 Codex 环境仍作为既有技术债处理；Sprint H 验收按 `docs/verification.md` 两步法执行。
- 全量前端 `npm run lint` 的历史问题不属于 Sprint H 范围；本轮 Sprint H 定向 lint 已通过。

## 最终判断

Sprint H 已满足完成定义：

1. Scanner Hub MVP 契约收口到 Nuclei JSONL / template variable
2. JSONL 与 command 输出可复制、字段稳定、可测试
3. Scanner Hub 页面能基于 Case/Payload 生成 Nuclei 分发材料
4. Scanner Hub 输出明确关联 Case 和 Payload
5. Scanner Hub 能跳转到 payload-scoped Interactions / Evidence
6. Evidence 继续复用统一 Evidence 契约
7. E2E 覆盖 Scanner Hub 到 Sprint G 的闭环且无 skip
8. 未越界到多扫描器深度集成、插件市场、真实 scanner 调度或 SARIF
9. 后端测试、前端 build、Sprint H 定向 lint、Sprint H E2E 均通过

**Sprint H 正式验收通过，可以关闭。**
