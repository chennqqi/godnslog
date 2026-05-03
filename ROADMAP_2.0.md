# GODNSLOG 2.0 开发规划

## 目标定位

GODNSLOG 2.0 应从“可用的 DNS/HTTP Log 工具”升级为“面向安全测试、资产验证和协作分析的现代化交互平台”。核心目标是降低部署和使用成本、提升记录分析能力、增强团队协作，并为后续扩展更多协议与自动化能力预留架构空间。

## 现状问题

1. 前端基于较旧的 Vue 2 技术栈，依赖重、维护成本高，组件和状态管理方式已经落后。
2. 现有能力集中在 DNSLOG、HTTPLOG、Rebinding 和基础回调，缺少面向现代安全测试场景的自动化、检索、关联分析和团队能力。
3. API、权限、审计、配置管理和部署体验需要系统化梳理，避免继续在旧结构上堆功能。

## 前端 2.0 重构规划

### 技术选型

- 使用 Next.js + React + TypeScript 作为新前端基础。
- 使用 shadcn/ui + Tailwind CSS 构建轻量、可维护的组件体系。
- 使用 TanStack Query 管理服务端状态，减少手写请求状态逻辑。
- 使用 Zod 统一表单校验、接口入参校验和类型推导。
- 使用 Playwright 补充关键流程端到端测试。

### 页面结构

- 登录与初始化页面：支持管理员初始化、密码重置提示和部署状态检查。
- 工作台：展示 DNS、HTTP、回调、Rebinding、Token、系统状态的摘要。
- 记录中心：统一查看 DNS/HTTP/Callback 记录，支持高级筛选、全文搜索、标签、备注、批量导出。
- Payload 中心：按漏洞类型生成和管理 payload，例如 SSRF、XXE、RCE、SSTI、反序列化、PDF/HTML 渲染探测。
- Rebinding 管理：可视化配置阶段、解析策略、TTL、目标地址和测试结果。
- 用户与团队：支持角色、Token、空间隔离、操作审计。
- 系统设置：域名、监听地址、数据库、通知、回调、安全策略等配置集中管理。
- 文档中心：保留内置使用文档，但改为可版本化、可搜索的 Markdown 文档。

### 迁移策略

1. 新建 `frontend-next/`，不直接破坏旧 `frontend/`。
2. 先实现登录、布局、记录列表、详情页和基础 API 封装。
3. 再逐步替换配置、用户、文档和 Rebinding 页面。
4. 稳定后将构建产物接入 Go 后端静态资源服务。
5. 完成迁移后再移除旧 Vue 前端。

## 2.0 新能力规划

### 记录分析能力

- DNS、HTTP、Callback 记录统一事件模型。
- 支持关键词、来源 IP、User-Agent、路径、Token、时间范围、记录类型组合筛选。
- 支持标签、备注、收藏、误报标记和批量删除。
- 支持 CSV、JSON、Markdown 报告导出。
- 支持相同目标、相同 Token、相同来源 IP 的自动关联。

### Payload 与自动化能力

- 内置常见漏洞场景 payload 模板：SSRF、XXE、RFI、RCE、SSTI、Log4Shell、Webhook、PDF 渲染探测等。
- Payload 支持变量插值，例如 `{token}`、`{domain}`、`{callback_url}`。
- 支持一次生成多协议 payload：DNS、HTTP、HTTPS、TXT、CNAME。
- 为每个 payload 生成独立追踪 Token，便于归因。

### 协作与权限

- 支持多工作空间或项目空间。
- 支持管理员、成员、只读观察者等角色。
- API Token 支持作用域、过期时间、最后使用时间和撤销。
- 增加操作审计日志，记录登录、配置变更、Token 操作和数据导出。

### 通知与集成

- 支持 Webhook 通知。
- 支持企业微信、钉钉、飞书、Slack、Telegram 等通知通道。
- 支持按规则触发通知，例如命中特定 Token、路径、来源 IP 或关键词。
- 提供稳定 REST API，便于接入扫描器、CI、漏洞验证脚本。

### 部署与运维

- 提供 Docker Compose 一键部署模板。
- 支持 SQLite 单机模式和 MySQL/PostgreSQL 生产模式。
- 增加健康检查、指标接口和基础运行状态页面。
- 配置优先使用环境变量和配置文件，避免硬编码。
- 支持日志级别、日志格式和数据保留周期配置。

### 安全增强

- 登录加入速率限制和失败锁定策略。
- Cookie、JWT、API Token 增加统一安全策略。
- 管理后台支持可信代理、来源限制和可选双因素认证。
- 敏感配置和 Token 仅展示一次或脱敏展示。
- 对外 API 增加权限校验、审计和错误返回规范。

## 后端重构方向

### 模块拆分

- `internal/server`：HTTP 管理后台与 API。
- `internal/dns`：DNS 服务、解析策略、Rebinding。
- `internal/record`：统一事件写入、查询、导出。
- `internal/auth`：用户、角色、Token、会话。
- `internal/notify`：通知通道与规则。
- `internal/config`：配置加载、校验和热更新能力。

### 数据模型

统一记录模型建议包含：

- `id`
- `workspace_id`
- `token`
- `type`
- `source_ip`
- `method`
- `host`
- `path`
- `query`
- `headers`
- `body`
- `dns_question`
- `dns_answer`
- `tags`
- `note`
- `created_at`

## 开发里程碑

### M1：基础架构与兼容层

- 建立 2.0 分支或目录结构。
- 明确 API v2 路由规范。
- 新建 Next.js 前端骨架。
- 梳理旧数据模型，设计迁移方案。
- 保持旧版核心 DNS/HTTPLOG 能力可运行。

### M2：核心记录闭环

- 实现统一记录模型。
- 完成 DNS、HTTP 记录写入和查询 API。
- 完成新前端登录、布局、记录列表和记录详情。
- 支持基础筛选、分页、删除和导出。

### M3：Payload、Token 与协作

- 实现 Payload 模板中心。
- 实现独立追踪 Token。
- 实现用户、角色、工作空间和 API Token。
- 增加审计日志。

### M4：通知、Rebinding 与自动化

- 重构 Rebinding 配置与展示。
- 增加通知规则和多通道推送。
- 提供稳定 API 文档。
- 增加常见扫描器和脚本接入示例。

### M5：生产化与发布

- 完善 Docker Compose、配置模板和升级文档。
- 增加健康检查、日志、数据保留和备份说明。
- 补齐单元测试、接口测试和关键 E2E 测试。
- 发布 2.0 beta，再根据反馈发布稳定版。

## 验收标准

- 新前端不依赖旧 Vue 代码，核心页面可独立构建和运行。
- DNSLOG、HTTPLOG、Rebinding、回调能力保持兼容或提供迁移说明。
- 记录中心支持统一查询、筛选、详情、标签、导出。
- Payload 可按场景生成并自动关联命中记录。
- 管理后台具备基础权限、Token 管理和审计能力。
- Docker Compose 能完成从空环境到可用服务的部署。

## 建议优先级

优先完成“统一记录模型 + 新前端记录中心 + Payload 追踪 Token”。这是 2.0 的核心价值闭环。通知、团队协作、Rebinding 可视化和高级报表可以在核心闭环稳定后逐步增强。
