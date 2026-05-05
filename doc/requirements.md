# GODNSLOG 2.0 需求文档

## Phase 11: 测试框架开发

为项目开发完整的测试框架，包括后端单元测试和前端端到端测试。

### 后端单元测试

已创建以下测试文件：
- internal/auth/apikey_test.go
- internal/payload/payload_test.go
- internal/canary/model_test.go
- internal/listener/model_test.go

### 前端E2E测试

已创建以下测试文件：
- frontend-next/playwright.config.ts
- frontend-next/e2e/login.spec.ts
- frontend-next/e2e/dashboard.spec.ts

### 运行测试

```bash
# 后端测试
go test ./...

# 前端E2E测试
cd frontend-next
cnpm install
cnpm run test:e2e
```

## 2026-05-03
按照2.0的规划开始开发，每一轮开发完成后在开发计划中标记进度

- 按照 ROADMAP_2.0.md 和 DEVELOPMENT_PLAN_2.0.md 执行开发
- 从 Phase 0 开始，逐步推进各个阶段
- 每轮开发完成后在 DEVELOPMENT_PLAN_2.0.md 中标记进度

## 2026-05-04
Phase 12：测试框架扩展
- 扩展interaction模块测试
- 扩展mcp模块测试
- 扩展listener模块测试
- 添加CLI测试
- 修复所有编译错误
- 所有测试通过
- 创建CLI使用文档
- 创建MCP Server使用文档
- 更新开发计划标记交付物完成

Phase 13：文档完善
- 修复所有编译错误
- 扩展测试覆盖
- 创建CLI使用文档
- 创建MCP Server使用文档
- 更新开发计划标记交付物完成
- 验证MVP验收标准

Phase 14：SMB/FTP Listener
- 扩展数据模型（SMBRequest、FTPCommand）
- 实现SMB Listener逻辑
- 实现FTP Listener逻辑
- 扩展Store接口
- 添加Handler端点
- 添加模型测试
- 更新文档
- 所有测试通过

Phase 15：企业级功能
- 实现数据保留策略
- 实现数据归档功能
- 实现高可用配置
- 添加企业级测试
- 创建文档
- 所有测试通过

Phase 16：插件市场/模板市场
- 实现插件模型
- 实现模板模型
- 实现插件服务
- 实现模板服务
- 添加插件/模板测试
- 创建文档
- 所有测试通过

## 2026-05-05
用户反馈2.0版本虽然标记完成，但存在严重问题：
1. 很多1.0的核心功能在2.0都丢失了：
   - 日志记录
   - API自动化操作
   - 多用户支持
   - 一些payload可见性
   - 一些文档
2. 2.0的功能实现都很草率，根本没有深入理解意思，去认真实现
3. 前端实现很粗糙，跟demo html没有区别，完全看不出来是一个生产可用的项目

根据ROADMAP_2.0.md开始完整的开发流程，分步骤以保证质量：
- 需求梳理 ✅ 已完成（doc/requirements-analysis-2026-05-05.md）
- 程序设计 ✅ 已完成（doc/design-plan-2026-05-05.md）
- 程序开发
- 测试验收
- 程序修复
- 完成验收

## 2026-05-05 (下午)
实现v2_api.go中的TODO标记API端点业务逻辑：

### 创建Service层
- internal/canary/service.go - Canary服务层，提供Canary的CRUD操作
- internal/rebinding/service.go - Rebinding服务层，提供Rebinding Rule的CRUD操作
- internal/listener/service.go - Listener服务层，提供Listener的CRUD操作

### 实现Canary API端点
- v2ListCanaries - 列出所有Canary Token
- v2CreateCanary - 创建新的Canary Token
- v2GetCanary - 获取指定Canary Token
- v2UpdateCanary - 更新Canary Token
- v2DeleteCanary - 删除Canary Token
- v2ListCanaryHits - 列出Canary Token的命中记录

### 实现Rebinding API端点
- v2ListRebindingRules - 列出所有Rebinding规则
- v2CreateRebindingRule - 创建新的Rebinding规则
- v2GetRebindingRule - 获取指定Rebinding规则
- v2UpdateRebindingRule - 更新Rebinding规则
- v2DeleteRebindingRule - 删除Rebinding规则
- v2ListRebindingSessions - 列出Rebinding会话

### 实现Listener API端点
- v2ListListeners - 列出所有Listener
- v2CreateListener - 创建新的Listener
- v2GetListener - 获取指定Listener
- v2UpdateListener - 更新Listener
- v2DeleteListener - 删除Listener
- v2ListListenerInteractions - 列出Listener的交互记录

### 实现Evidence API端点
- v2GetEvidence - 获取Evidence报告（添加说明：Evidence报告按需生成，不持久化存储）

## 2026-05-05 (晚间)
用户确认Q1-Q7决策项，输出统一数据模型设计和前端UI/UX规范。

### Q1: Payload模板变量设计 — 结论
- base32_context暂时不明确，预留，程序设计时考虑清楚再明确
- 需要用户定义任意变量名
- Payload模板是纯文本替换（字符串替换）

### Q2: 通知渠道优先级 — 结论
- 支持企微/飞书/钉钉和通用webhook
- 不需要支持SMTP通知（SMTP用于邮件嗅探，不是通知）
- 通知触发通过workflow触发
- 默认关闭逐条通知

### Q3: 1.0数据模型兼容策略 — 结论
- 1.0前端直接废弃，但需要提供数据迁移工具
- Roadmap明确要求1.0中的用户管理、文档、API等核心功能需要在2.0前端中具备

### Q4: MCP Server技术选型 — 结论
- MCP采用Streamable HTTP协议

### Q5: 开源协议与商业模式 — 结论
- 继续开源，保持原有协议，以后考虑OpenCore模式

### Q6: Rebinding Lab安全边界 — 结论
- Rebinding作为全局配置，仅Super/Admin可修改
- 所有Rebinding Interaction标记类型
- 为此功能增加遥测功能，默认打开，启动时通过环境变量关闭（防止产品被恶意滥用）

### Q7: 前端技术栈确认 — 结论
- Next.js App Router + TanStack Query + Zustand + React Hook Form
- shadcn/ui深度使用
- 目录结构采用features/按业务模块组织
- TypeScript严格模式 + ESLint + Prettier
- 数据获取策略：Server Components + TanStack Query混合

### 输出文档
- `doc/2.0-Requirement.md` — 重新梳理的需求规格说明书（含P0/P1/P2需求清单、7个待决策问题）
- `doc/2.0-data-model-design.md` — 统一数据模型设计（数据库Schema、迁移策略、Go模型规范）
- `doc/2.0-frontend-uiux-spec.md` — 前端UI/UX规范（目录结构、组件规范、页面设计、技术栈）

### 补充原则（API模型共享）
后端接口的请求/响应结构体必须使用 Go `struct` 声明，放置于**独立可导出的包**（如 `pkg/apimodels/` 或 `internal/models/`）。这些结构体是服务端 Gin handler 的绑定目标，也是客户端（Next.js/CLI）TypeScript 类型的唯一来源，确保前后端契约一致，禁止前端独立编造类型。

---

## 2026-05-05 Phase 1 开发完成

### 1. 数据库迁移 + 双写逻辑
- `server/webui.go`: `initDatabase` 新增 `v2models.Interaction` 表同步（创建 `interactions` 统一表）
- `server/webserver.go`: DNS 记录双写（`models.TblDns` + `v2models.Interaction`）
- `server/webapi.go`: HTTP 记录双写（`models.TblHttp` + `v2models.Interaction`）

### 2. 模型统一清理
- `internal/models/case.go`: 新增 `Tags` 自定义类型（`sql.Scanner`/`driver.Valuer`），支持 JSON 数组在 API 中序列化为数组、数据库中存储为 JSON 字符串
- `models/v2.go`: 添加废弃注释，明确 API 类型与 `internal/models` 冗余；保留 `Tbl*` 数据库模型确保现有表兼容

### 3. 前端 features/ 目录重构
- 创建 `features/auth/store.ts`（Zustand auth store）、`features/auth/hooks/use-auth.ts`
- 创建 `features/cases/hooks/use-cases.ts`（TanStack Query: list/get/create/update/delete）
- 创建 `features/payloads/hooks/use-payloads.ts`（list/get/create/revoke）
- 创建 `features/interactions/hooks/use-interactions.ts`（list/get/delete/export）
- 初始化 shadcn/ui 组件（新增17个）：badge, dialog, select, tooltip, toast, tabs, separator, scroll-area, dropdown-menu, label, textarea, checkbox, switch, skeleton, pagination, popover, accordion
- 安装依赖：`zustand`, `react-hook-form`, `@hookform/resolvers`，所有 `@radix-ui/*` 依赖

### 4. middleware.ts + lib/api.ts
- `src/middleware.ts`: Next.js 中间件认证守卫，检查 cookie/token，未认证重定向到 `/login`
- `src/lib/api.ts`: SSR-safe Axios 拦截器，`typeof window` 检查避免 SSR 崩溃，`ApiError` 统一错误类型，401 自动清除认证并跳转

### 编译状态
- `go build ./...` ✅ 通过
- `cd frontend-next && npx next lint` ✅ 通过（仅 useEffect 依赖警告，不影响功能）
