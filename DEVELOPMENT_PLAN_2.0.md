# GODNSLOG 2.0 Development Plan

## 目标

本文档面向工程执行，承接 `ROADMAP_2.0.md` 的产品路线。目标是在不破坏 1.0 可用性的前提下，逐步实现 2.0 的 OAST、扫描器协同和 Agent API/MCP 能力。

## 总体策略

- 新代码优先放入 2.0 独立目录或分支，避免直接破坏旧版。
- API 先行，前端、CLI、插件、MCP Server 都复用同一套后端接口。
- 先完成 DNS/HTTP 核心闭环，再扩展 SMTP、LDAP、Canary、Workflow。
- 前端新建 `frontend-next/`，旧 `frontend/` 保留到 2.0 稳定后再移除。
- 先支持 SQLite 单机模式，再完善 MySQL/PostgreSQL 生产模式。

## 建议目录结构

```text
cmd/
  godnslog/              # 主服务入口
  godnslog-cli/          # CLI：创建 Case、生成 Payload、轮询命中
  godnslog-mcp-server/   # MCP Server
internal/
  auth/                  # 用户、角色、APIKey、审计
  case/                  # Case、目标、协作上下文
  payload/               # Payload 模板、Token、变量渲染
  interaction/           # 统一事件模型、查询、导出
  listener/              # DNS/HTTP/SMTP/LDAP 监听
  workflow/              # 规则、动作、队列、重放
  notify/                # 飞书、企业微信、Webhook、Email 等
  integration/           # Nuclei、Burp/ZAP、YApi、CI/CD
  canary/                # 长期诱饵 Token
  config/                # 配置加载、校验、环境变量
frontend-next/           # Next.js + shadcn/ui 前端
docs/                    # 2.0 API、部署、插件、MCP 文档
```

## 技术选型

### 后端

- Go 作为主语言，延续当前项目基础。
- Gin 可继续使用，也可在 2.0 重构时评估更清晰的路由分层。
- 数据库优先支持 SQLite，生产模式支持 MySQL/PostgreSQL。
- OpenAPI 作为接口契约，生成 Swagger UI 和 SDK 基础类型。

### 前端

- Next.js + React + TypeScript。
- shadcn/ui + Tailwind CSS。
- TanStack Query 管理服务端状态。
- Zod 处理表单和接口数据校验。
- Playwright 覆盖核心 E2E 流程。

### 对外工具

- `godnslog-cli`：服务 Nuclei、CI/CD 和脚本化使用。
- `godnslog-mcp-server`：服务 AI Agent。
- Burp/ZAP 插件先做最小可用版本，再完善 UI。

## 阶段计划

### Phase 0：整理与兼容

- 固化 1.0 当前能力清单：DNSLOG、HTTPLOG、Rebinding、Callback、多用户、SDK、标准 DNS 解析、xip。
- 整理旧 TODO：管理员查看所有记录、匿名文档访问、Rebinding 第二阶段配置、登录提示问题、CDN 依赖问题。
- 定义 2.0 配置格式和数据迁移边界。
- 建立 OpenAPI 文档骨架。

### Phase 1：核心数据模型与 API

- 实现 `Case`、`Payload`、`Interaction`、`Evidence`、`APIKey` 数据模型。
- 实现 APIKey 作用域、过期、撤销、最后使用时间。
- 实现 DNS/HTTP Interaction 写入、查询、详情、删除、导出。
- 实现 Payload Token 生成、变量渲染和 Case 绑定。
- 实现基础审计日志。

### Phase 2：新前端 MVP

- 新建 `frontend-next/`。
- 实现登录、布局、Command Center。
- 实现 Case Board、Payload Studio、Interaction Timeline。
- 实现记录筛选、详情、标签、备注、导出。
- 实现系统设置中的域名、监听地址、通知和 Token 管理。

### Phase 3：扫描器协同

- 实现 `godnslog-cli`：
  - 创建 Case。
  - 生成 Payload。
  - 等待或轮询 Interaction。
  - 导出 JSON/Markdown 报告。
- 提供 Nuclei 模板示例和 JSONL 输出。
- 支持 OpenAPI/YApi 导入后的批量 Payload 注入。
- 提供 GitHub Actions/GitLab CI 示例。
- 设计 Burp/ZAP 插件 API，先实现最小插件或脚本扩展。

### Phase 4：Workflow 与通知

- 实现规则条件：协议、Token、来源 IP、路径、Header、Body、关键词、Case、风险等级。
- 实现动作：通知、打标签、转发 Webhook、创建报告、丢弃噪声。
- 支持飞书、企业微信、钉钉、Slack、Discord、Telegram、Email、Webhook。
- 支持异步队列、失败重试和历史命中重放。
- 支持自定义 HTTP 响应控制。

### Phase 5：MCP 与 Agent 赋能

- 实现 `godnslog-mcp-server`。
- 暴露工具：`create_case`、`create_payload`、`list_interactions`、`wait_for_interaction`、`summarize_evidence`、`export_report`、`revoke_token`。
- 为 MCP Client 提供独立 APIKey、作用域和过期策略。
- 所有 Agent 操作写入审计日志。
- 高风险能力默认禁用，例如长期 Canary、响应修改、DNS C2。

### Phase 6：高级能力

- 实现 SMTP/LDAP Listener。
- 实现 Canary Token 与持续监测。
- 实现 Rebinding Lab。
- 增加多工作空间、多域名、多 Listener 节点。
- 评估 SMB/FTP/TCP Raw Listener。
- 增加 AI 摘要、聚类和报告初稿插件。

## MVP 验收标准

- 新前端不依赖旧 Vue 代码，可独立构建。
- 用户可创建 Case，生成 SSRF/XXE/RCE/Blind SQLi Payload。
- DNS/HTTP 回连可自动关联到 Case 和 Payload。
- Interaction Timeline 可展示完整命中时间线。
- 支持 Markdown/JSON 导出证据。
- APIKey 支持作用域、过期和撤销。
- CLI 可创建 Case、生成 Payload、等待命中。
- 至少支持 Webhook、企业微信或飞书中的两种通知。
- Docker Compose 可从空环境启动可用服务。

## 测试计划

- 后端单元测试覆盖 Token 生成、变量渲染、权限校验、Interaction 归因。
- API 集成测试覆盖 Case、Payload、Interaction、APIKey、导出。
- Listener 测试覆盖 DNS 查询、HTTP 请求、异常请求和大 Body。
- 前端 Playwright 覆盖登录、创建 Case、生成 Payload、查看命中、导出报告。
- CLI 测试覆盖命令参数、JSON 输出和错误码。
- MCP 测试覆盖工具权限、超时等待、审计日志和撤销 Token。

## 迁移计划

- 2.0 初期兼容旧 DNS/HTTPLOG 基础能力。
- 旧数据库迁移到 Interaction 模型时保留原始字段。
- 旧用户迁移到新 auth 模型，管理员权限显式化。
- 旧前端在 2.0 beta 前保留，稳定后标记废弃。
- README 中旧 Roadmap 合并到 2.0 文档，避免维护多份冲突计划。

## 风险与控制

- 多协议 Listener 容易扩大攻击面：默认只启用 DNS/HTTP，其余按配置开启。
- Agent/MCP 容易被滥用：必须最小权限、过期、审计和高风险动作禁用。
- Workflow 可能造成 SSRF/转发风险：出站请求需要 allowlist、超时和大小限制。
- DNS C2 容易产生误用：仅保留授权实验模式，默认关闭并明确安全提示。
- 前端重构范围大：先做核心闭环，避免一次性迁移所有历史页面。

## 交付物

- `ROADMAP_2.0.md`：产品路线图。
- `DEVELOPMENT_PLAN_2.0.md`：工程开发计划。
- OpenAPI 文档。
- Docker Compose 示例。
- CLI 使用文档。
- MCP Server 使用文档。
- Nuclei、Burp/ZAP、CI/CD 集成示例。
