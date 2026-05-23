# GODNSLOG 2.0 Sprint F 验收结论

## 验收对象

- `docs/superpowers/specs/2026-05-23-godnslog-2-sprint-f-design.md`
- `frontend-next/src/app/dashboard/cases/page.tsx`
- `frontend-next/src/app/dashboard/cases/[id]/page.tsx`
- `frontend-next/src/app/dashboard/payloads/new/page.tsx`
- `frontend-next/src/app/dashboard/payloads/[id]/page.tsx`
- `frontend-next/e2e/cases.spec.ts`

## 验收结论

**结论：不通过，需回炉修正。**

Sprint F 的页面收口方向已经基本对上设计稿，主链路的页面骨架也已经补齐：

- `Cases Board`
- `Case Detail`
- `New Payload`
- `Payload Detail`

但当前实现还不能进入“通过”状态，原因不是范围没做，而是本轮代码本身仍有可确认的质量阻断项：

1. Sprint F 相关页面存在新的 `eslint` 错误
2. `Cases Board` 的搜索 / 状态筛选实现存在状态滞后风险
3. 新增 E2E 当前未形成有效绿色证据

因此本轮应视为“功能轮廓已完成，但未收口”。

## 本次完成点

### 1. Cases Board 已被收口为轻量入口

当前 `Cases Board` 已明显移除上一版的批量操作、编辑、删除等重动作，保留为：

- 搜索
- 状态筛选
- 进入详情
- 新建 Case

见：

- [frontend-next/src/app/dashboard/cases/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/page.tsx:92)

### 2. Case Detail 已成为主工作页

当前 `Case Detail` 已补齐：

- Case 基本信息
- `payload_count`
- `interaction_count`
- `hit_payload_count`
- Payload 列表
- `Create Payload`
- `View Evidence`
- `View Interactions`

见：

- [frontend-next/src/app/dashboard/cases/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/[id]/page.tsx:70)
- [frontend-next/src/app/dashboard/cases/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/[id]/page.tsx:94)
- [frontend-next/src/app/dashboard/cases/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/[id]/page.tsx:109)
- [frontend-next/src/app/dashboard/cases/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/[id]/page.tsx:145)

### 3. New Payload 已补上 Case 上下文

当前 `New Payload` 页面已经在 `case_id` 存在时展示当前关联 Case，并在列表加载后回填所选 Case：

- [frontend-next/src/app/dashboard/payloads/new/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/new/page.tsx:350)
- [frontend-next/src/app/dashboard/payloads/new/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/new/page.tsx:412)

这满足了 Sprint F 对“从 Case 进入执行面创建向导”的设计要求。

### 4. Payload Detail 已补上关联与后续闭环入口

当前 `Payload Detail` 已新增：

- 关联 Case
- 最近交互
- 跳转到 `Interactions`
- 跳转到 `Evidence`

同时已移除原有的 `revoke` 操作，不再把该页做成生命周期管理台。

见：

- [frontend-next/src/app/dashboard/payloads/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/[id]/page.tsx:141)
- [frontend-next/src/app/dashboard/payloads/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/[id]/page.tsx:160)
- [frontend-next/src/app/dashboard/payloads/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/[id]/page.tsx:191)

### 5. Windsurf 已按轮次提交本地 commit

本轮可见提交：

- `2fb32f5 sprint-f:收口 Cases Board、Case Detail、New Payload、Payload Detail 主链路`
- `260cd90 sprint-f: 添加 E2E 测试验证主链路`

这满足了“每次任务完成后本地 commit，不需要 push”的过程要求。

## 阻断项

### 1. Sprint F 页面本身存在新的 lint 错误

我针对本轮相关文件单独执行了：

```bash
cd frontend-next && npx eslint src/app/dashboard/cases/page.tsx src/app/dashboard/cases/[id]/page.tsx src/app/dashboard/payloads/new/page.tsx src/app/dashboard/payloads/[id]/page.tsx e2e/cases.spec.ts
```

结果失败，且错误直接落在 Sprint F 相关文件：

- `frontend-next/src/app/dashboard/cases/page.tsx`
- `frontend-next/src/app/dashboard/cases/[id]/page.tsx`
- `frontend-next/src/app/dashboard/payloads/[id]/page.tsx`

主要问题：

1. `useEffect` 中调用了后声明的函数，触发 `Cannot access variable before it is declared`
2. `Payload Detail` 继续通过 `(payload as any).scenario` 访问未建模字段

见：

- [frontend-next/src/app/dashboard/cases/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/page.tsx:44)
- [frontend-next/src/app/dashboard/cases/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/[id]/page.tsx:16)
- [frontend-next/src/app/dashboard/payloads/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/[id]/page.tsx:16)
- [frontend-next/src/app/dashboard/payloads/[id]/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/payloads/[id]/page.tsx:106)

这是本轮代码质量阻断，不应带着这些错误继续进入下一 Sprint。

### 2. Cases Board 的搜索与筛选实现存在状态滞后风险

当前 `Cases Board` 在输入搜索词和切换状态筛选时，采用的是：

- 先 `setSearchTerm(...)` / `setStatusFilter(...)`
- 再立刻调用 `loadCases()`

但 `loadCases()` 读取的是闭包里的旧状态，这会导致请求参数相对用户当前输入至少落后一拍。

见：

- [frontend-next/src/app/dashboard/cases/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/page.tsx:107)
- [frontend-next/src/app/dashboard/cases/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/page.tsx:113)

这不是代码风格问题，而是直接影响 `Cases Board` 主路径可用性的行为缺陷。

### 3. 新增 E2E 目前不能作为通过证据

我执行了：

```bash
cd frontend-next && npm run test:e2e -- --grep 'Cases Board|Case Detail|New Payload|Payload Detail'
```

Playwright 已开始跑，但当前环境缺少 Chromium 二进制，27 个用例全部在浏览器启动阶段失败：

- `Executable doesn't exist ... chrome-headless-shell`
- Playwright 明确要求执行 `npx playwright install`

因此，这批 E2E 目前只能证明“测试已编写”，不能证明“主链路已验证通过”。

另外，测试代码本身还有一处明显的契约偏差风险：

- 页面用的是 Radix `Select`
- 测试却按原生 `select` 元素查找

见：

- [frontend-next/src/app/dashboard/cases/page.tsx](/data/dev/github.com/chennqqi/godnslog/frontend-next/src/app/dashboard/cases/page.tsx:113)
- [frontend-next/e2e/cases.spec.ts](/data/dev/github.com/chennqqi/godnslog/frontend-next/e2e/cases.spec.ts:36)

这意味着即使浏览器补齐，测试也未必会绿。

## 对 Sprint F 5 个验收问题的判断

1. 用户能否从 `Cases Board` 顺滑进入 `Case Detail`：**实现已补，但未完成有效前端验证**
2. 用户能否从 `Case Detail` 发起 `Create Payload`：**实现已补，是**
3. `New Payload` 是否清楚绑定当前 Case：**是**
4. `Payload Detail` 是否从静态展示页变成闭环连接页：**是**
5. 当前收口是否没有越界到批量操作、生命周期治理、模板平台：**是**

## 本次验证

已执行：

```bash
git log --oneline -8
git diff --name-only 01d7a28..260cd90
cd frontend-next && npm run lint
cd frontend-next && npx eslint src/app/dashboard/cases/page.tsx src/app/dashboard/cases/[id]/page.tsx src/app/dashboard/payloads/new/page.tsx src/app/dashboard/payloads/[id]/page.tsx e2e/cases.spec.ts
cd frontend-next && npm run test:e2e -- --grep 'Cases Board|Case Detail|New Payload|Payload Detail'
```

结果：

- `git` 证据确认 Sprint F 的两个提交存在
- 全量 `lint` 失败，仓库仍有大量存量问题
- 定向 `eslint` 失败，且失败点直接命中 Sprint F 相关文件
- E2E 失败，原因是当前环境缺少 Playwright 浏览器

## Windsurf 修正清单

Windsurf 需要先完成以下收口后，再回传验收：

1. 修复 Sprint F 相关页面的新增 `eslint` 错误
2. 修正 `Cases Board` 搜索 / 状态筛选的状态滞后问题
3. 校正 `cases.spec.ts` 中对状态筛选控件的断言方式，使其匹配 Radix `Select`
4. 在可运行 Playwright 的环境补齐：
   - `cd frontend-next && npx playwright install`
   - `cd frontend-next && npm run test:e2e -- --grep 'Cases Board|Case Detail|New Payload|Payload Detail'`
5. 按规则新增本地 commit，并回传：
   - `commit hash`
   - `commit message`
   - 对应修正任务

## 下一步建议

本轮不要进入 Sprint G。

先把 Sprint F 收口成“代码质量通过 + 主链路验证通过”的状态，再进入下一轮规划，否则会把控制面主链路的不稳定性继续向后传。
