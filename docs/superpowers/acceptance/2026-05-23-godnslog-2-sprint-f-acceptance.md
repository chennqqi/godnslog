# GODNSLOG 2.0 Sprint F 复验结论

## 验收对象

- `docs/superpowers/specs/2026-05-23-godnslog-2-sprint-f-design.md`
- `frontend-next/src/app/dashboard/cases/page.tsx`
- `frontend-next/src/app/dashboard/cases/[id]/page.tsx`
- `frontend-next/src/app/dashboard/payloads/new/page.tsx`
- `frontend-next/src/app/dashboard/payloads/[id]/page.tsx`
- `frontend-next/src/types/index.ts`
- `frontend-next/e2e/cases.spec.ts`

## 验收结论

**结论：仍不通过。**

Windsurf 这一轮已经修掉了上一版复验里指出的代码问题：

- Sprint F 相关页面的定向 `eslint` 错误已消失
- `Cases Board` 搜索 / 状态筛选的状态滞后问题已修正
- `Payload.scenario` 已进入类型定义，不再依赖 `(payload as any)`
- `cases.spec.ts` 对状态筛选控件的断言已改为匹配 Radix `Select`

但 Sprint F 仍然不能关闭，原因不是页面代码还在报错，而是 **主链路 E2E 仍然没有形成可放行的通过证据**，并且 Windsurf 本轮提交里还带入了失败的 Playwright 产物。

## 本轮确认通过的部分

### 1. 定向 lint 已通过

我执行了：

```bash
cd frontend-next && npx eslint src/app/dashboard/cases/page.tsx src/app/dashboard/cases/[id]/page.tsx src/app/dashboard/payloads/new/page.tsx src/app/dashboard/payloads/[id]/page.tsx e2e/cases.spec.ts
```

结果通过，没有报错。

这说明上一轮复验指出的三个代码质量问题已经被处理：

- `useEffect` 中调用后声明函数
- `Cases Board` 闭包状态滞后
- `Payload Detail` 的 `any` 访问

### 2. Sprint F 修复提交存在

本轮新增提交：

- `dda24e3 sprint-f: 修复 eslint 错误和状态滞后问题`

这满足了 Windsurf 每轮修复必须本地 `commit` 的流程要求。

### 3. 代码修复方向与设计一致

本轮修复没有越界到 Sprint F 以外的能力扩展，仍然围绕：

- `Cases Board`
- `Case Detail`
- `New Payload`
- `Payload Detail`

这点是合格的。

## 当前阻断项

### 1. Sprint F 主链路 E2E 仍未形成通过证据

我执行了两条命令：

```bash
cd frontend-next && npx playwright test --reporter=line --grep 'Cases Board|Case Detail|New Payload|Payload Detail'
cd frontend-next && npx playwright test --reporter=line e2e/cases.spec.ts
```

两次都失败了。当前环境的失败原因一致：

- Chromium 能启动
- 但在当前沙箱环境里立即触发：
- `sandbox_host_linux.cc:41`
- `Operation not permitted (1)`

所以我不能在当前环境里给出“E2E 已验证通过”的结论。

这比上一轮前进了一步：

- 上一轮是缺少浏览器二进制
- 这一次浏览器已经安装，但运行环境仍然阻断浏览器用例

因此，**E2E 仍然没有新鲜绿色证据**。

### 2. Windsurf 提交里带入了失败的 Playwright 产物

`dda24e3` 中直接包含：

- `frontend-next/test-results/.last-run.json`
- `frontend-next/playwright-report/index.html`

而 `frontend-next/test-results/.last-run.json` 当前明确记录：

- `status: "failed"`

这意味着 Windsurf 这轮提交并不是基于一组绿色 E2E 结果形成的。

这与 Sprint F 当前需要的“主链路验证通过后再收口”目标不一致。

### 3. 已提交产物里还出现过功能性失败证据

我抽查了已提交的 Playwright 报告数据，能看到至少一类非环境级失败：

- `Payload Detail` 用例期望出现 `ssrf_http`
- 实际页面拿到的是登录页标题 `GODNSLOG 2.0`

见：

- [frontend-next/playwright-report/data/09098f3264d40dd7c7c43eee9560c0b8da23b539.md](/data/dev/github.com/chennqqi/godnslog/frontend-next/playwright-report/data/09098f3264d40dd7c7c43eee9560c0b8da23b539.md:1)

这说明 Windsurf 即使在可启动浏览器的环境里，也没有证明整条 Sprint F 主链路已经跑绿。

## 对 Sprint F 5 个验收问题的判断

1. 用户能否从 `Cases Board` 顺滑进入 `Case Detail`：**代码实现已完成，但缺少可放行的 E2E 通过证据**
2. 用户能否从 `Case Detail` 发起 `Create Payload`：**代码实现已完成，但缺少可放行的 E2E 通过证据**
3. `New Payload` 是否清楚绑定当前 Case：**代码实现是**
4. `Payload Detail` 是否从静态展示页变成闭环连接页：**代码实现是**
5. 当前收口是否没有越界到批量操作、生命周期治理、模板平台：**是**

## 本次验证

已执行：

```bash
git log --oneline -8
git diff --name-only 260cd90..HEAD
git show --stat --name-only --oneline dda24e3
cd frontend-next && npx eslint src/app/dashboard/cases/page.tsx src/app/dashboard/cases/[id]/page.tsx src/app/dashboard/payloads/new/page.tsx src/app/dashboard/payloads/[id]/page.tsx e2e/cases.spec.ts
cd frontend-next && npx playwright test --reporter=line --grep 'Cases Board|Case Detail|New Payload|Payload Detail'
cd frontend-next && npx playwright test --reporter=line e2e/cases.spec.ts
```

结果：

- `eslint` 定向通过
- Playwright 在当前环境仍然全部失败，失败点在浏览器沙箱启动阶段
- 已提交的 `.last-run.json` 显示 Windsurf 自己的最近一次 Playwright 运行也是 `failed`

## 给 Windsurf 的下一步要求

Sprint F 现在只剩验证收口，不需要再扩页面功能。

Windsurf 下一轮只做三件事：

1. 在可运行 Playwright Chromium 的环境里，真正跑绿 `e2e/cases.spec.ts`
2. 不要提交 `playwright-report/` 和 `test-results/` 这类失败产物
3. 回传明确证据：
   - 执行命令
   - 通过用例数
   - 失败用例数
   - 本轮 `commit hash / message`

在这三件事完成前，Sprint F 不能进入下一 Sprint。
