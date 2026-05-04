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

### Phase 0：整理与兼容 ✅ 已完成

- ✅ 固化 1.0 当前能力清单：DNSLOG、HTTPLOG、Rebinding、Callback、多用户、SDK、标准 DNS 解析、xip。
- ✅ 整理旧 TODO：管理员查看所有记录、匿名文档访问、Rebinding 第二阶段配置、登录提示问题、CDN 依赖问题。
- ✅ 定义 2.0 配置格式和数据迁移边界。
- ✅ 建立 OpenAPI 文档骨架。

**交付物**：
- doc/1.0-capabilities.md - 1.0能力清单
- doc/1.0-todo-inventory.md - TODO项清单
- doc/2.0-config-format.md - 2.0配置格式和迁移策略
- docs/openapi.yaml - OpenAPI 3.0文档骨架

### Phase 1：核心数据模型与 API ✅ 已完成

- ✅ 实现 `Case`、`Payload`、`Interaction`、`Evidence`、`APIKey` 数据模型。
- ✅ 实现 APIKey 作用域、过期、撤销、最后使用时间。
- ✅ 实现 DNS/HTTP Interaction 写入、查询、详情、删除、导出。
- ✅ 实现 Payload Token 生成、变量渲染和 Case 绑定。
- ✅ 实现基础审计日志。

**交付物**：
- internal/case/case.go - Case数据模型和服务
- internal/payload/payload.go - Payload数据模型和服务
- internal/interaction/interaction.go - Interaction数据模型和服务
- internal/interaction/evidence.go - Evidence数据模型
- internal/interaction/evidence_service.go - Evidence服务
- internal/auth/apikey.go - APIKey数据模型和服务
- internal/auth/audit.go - AuditLog数据模型和服务
- 各模块的migration.go - 数据库迁移文件

### Phase 2：新前端 MVP ✅ 已完成

- ✅ 新建 `frontend-next/`。
- ✅ 实现登录、布局、Command Center。
- ✅ 实现 Case Board、Payload Studio、Interaction Timeline。
- ✅ 实现记录筛选、详情、标签、备注、导出。
- ✅ 实现系统设置中的域名、监听地址、通知和 Token 管理。

**交付物**：
- frontend-next/ - Next.js项目结构
- 登录页面和认证逻辑
- Dashboard布局和Command Center
- Case Board - Case列表和详情页
- Payload Studio - Payload列表和搜索
- Interaction Timeline - Interaction列表和筛选
- Interaction详情页 - 完整信息展示和导出功能
- 系统设置页面 - 通用、域名、监听、通知、Token管理
- API客户端和类型定义

### Phase 3：扫描器协同 🔄 进行中

- ✅ 实现 `godnslog-cli`：
  - ✅ 创建 Case。
  - ✅ 生成 Payload。
  - ✅ 等待或轮询 Interaction。
  - ✅ 导出 JSON/Markdown 报告。
- ✅ 提供 Nuclei 模板示例和 JSONL 输出。

**交付物**：
- cmd/cli/main.go - CLI入口
- cli/ - CLI实现（case, payload, interaction, report命令）
- examples/nuclei/ - Nuclei模板示例
- 支持 OpenAPI/YApi 导入后的批量 Payload 注入。
- 提供 GitHub Actions/GitLab CI 示例。
- 设计 Burp/ZAP 插件 API，先实现最小插件或脚本扩展。

### Phase 4：Workflow 与通知 ✅ 已完成

- ✅ 实现规则条件：协议、Token、来源 IP、路径、Header、Body、关键词、Case、风险等级。
- ✅ 实现动作：通知、打标签、转发 Webhook、创建报告、丢弃噪声。
- ✅ 支持飞书、企业微信、钉钉、Slack、Discord、Telegram、Email、Webhook。
- ✅ 支持异步队列、失败重试和历史命中重放。
- ✅ 实现规则存储和API接口。

**交付物**：
- internal/rule/model.go - 规则数据模型
- internal/rule/engine.go - 规则引擎（条件匹配）
- internal/rule/action.go - 动作执行器（通知渠道）
- internal/rule/queue.go - 异步队列和重试
- internal/rule/store.go - 规则存储实现
- internal/rule/handler.go - HTTP API处理器

### Phase 5：扫描器协同 ✅ 已完成

- ✅ Burp Suite插件（Java实现）
- ✅ CI/CD集成示例（GitHub Actions、GitLab CI、Jenkins）
- ✅ Payload模板库扩展
- ✅ 高风险检测门禁
- ✅ 命中聚类和噪声压缩

**交付物**：
- extensions/burp/ - Burp Suite扩展（Java + Maven）
- examples/ci/ - CI/CD集成示例
- templates/ - Payload模板库
- internal/clustering/ - 命中聚类和噪声压缩

### Phase 6：MCP 与 Agent 赋能 ✅ 已完成

- ✅ 实现 `godnslog-mcp-server`
- ✅ 暴露工具：`create_case`、`create_payload`、`list_interactions`、`wait_for_interaction`、`summarize_evidence`、`export_report`、`revoke_token`
- ✅ 审计日志集成
- ✅ 高风险能力默认禁用

**交付物**：
- cmd/mcp-server/ - MCP服务器入口
- internal/mcp/ - MCP服务器实现
- 支持7个MCP工具

### Phase 7：Canary 持续监测 ✅ 已完成

- ✅ 实现Canary数据模型
- ✅ 实现Canary检测逻辑
- ✅ 实现风险等级评估
- ✅ 实现静默窗口和压缩
- ✅ 实现Canary API

**交付物**：
- internal/canary/model.go - Canary数据模型
- internal/canary/detector.go - Canary检测器
- internal/canary/store.go - Canary存储
- internal/canary/handler.go - Canary API处理器
- internal/canary/README.md - 文档

### Phase 8：Rebinding Lab 与高级 DNS ✅ 已完成

- ✅ 实现Rebinding数据模型
- ✅ 实现多阶段解析逻辑
- ✅ 实现会话追踪
- ✅ 实现5种预定义场景
- ✅ 实现Rebinding API
- ✅ C2安全控制（默认禁用）

**交付物**：
- internal/rebinding/model.go - Rebinding数据模型
- internal/rebinding/resolver.go - Rebinding解析器
- internal/rebinding/store.go - Rebinding存储
- internal/rebinding/handler.go - Rebinding API处理器
- internal/rebinding/README.md - 文档

### Phase 9：SMTP/LDAP/SMB/FTP Listener ✅ 已完成

- ✅ 多工作空间、多域名、多 Listener 节点。数据模型
- ✅ 实现SMTP Listener
- ✅ 实现LDAP Listener
- ✅ 实现Listener存储
- ✅ 实现Listener API

**交付物**：
- internal/listener/model.go - Listener数据模型
- internal/listener/smtp.go - SMTP Listener实现
- internal/listener/ldap.go - LDAP Listener实现
- internal/listener/store.go - Listener存储
- internal/listener/handler.go - Listener API
- internal/listener/README.md - 文档

### Phase 10：多工作空间支持 ✅ 已完成

- ✅ 实现Workspace数据模型
- ✅ 实现Workspace存储
- ✅ 实现Workspace API
- ✅ 实现成员管理
- ✅ 实现域名管理

**交付物**：
- internal/workspace/model.go - Workspace数据模型
- internal/workspace/store.go - Workspace存储
- internal/workspace/handler.go - Workspace API
- internal/workspace/README.md - 文档

### Phase 11：测试框架开发 ✅ 已完成

- ✅ 实现后端单元测试（auth、payload、canary、listener）
- ✅ 实现前端E2E测试（Playwright）
- ✅ 修复编译错误
- ✅ 创建测试文档

**交付物**：
- internal/auth/apikey_test.go - Auth单元测试
- internal/payload/payload_test.go - Payload单元测试
- internal/canary/model_test.go - Canary单元测试
- internal/listener/model_test.go - Listener单元测试
- frontend-next/playwright.config.ts - Playwright配置
- frontend-next/e2e/login.spec.ts - 登录E2E测试
- frontend-next/e2e/dashboard.spec.ts - 仪表板E2E测试
- doc/phase11-summary.md - Phase 11总结文档

### Phase 12：测试框架扩展 ✅ 已完成

- ✅ 扩展interaction模块测试
- ✅ 扩展mcp模块测试
- ✅ 扩展listener模块测试
- ✅ 添加CLI测试
- ✅ 修复所有编译错误
- ✅ 清理测试缓存
- ✅ 创建Phase 12总结文档

**交付物**：
- internal/interaction/interaction_test.go - Interaction单元测试
- internal/mcp/server_test.go - MCP服务器单元测试
- internal/listener/model_test.go - Listener模型测试
- cmd/cli/main_test.go - CLI测试
- doc/phase12-summary.md - Phase 12总结文档

### Phase 13：文档完善 ✅ 已完成

- ✅ 修复所有编译错误
- ✅ 扩展测试覆盖（interaction、mcp、cli）
- ✅ 创建CLI使用文档
- ✅ 创建MCP Server使用文档
- ✅ 更新开发计划标记交付物完成
- ✅ 验证MVP验收标准
- ✅ 创建Phase 13总结文档

**交付物**：
- docs/CLI_USAGE.md - CLI使用文档
- docs/MCP_SERVER_USAGE.md - MCP Server使用文档
- doc/phase13-summary.md - Phase 13总结文档

### Phase 14：SMB/FTP Listener ✅ 已完成

- ✅ 扩展数据模型（SMBRequest、FTPCommand）
- ✅ 实现SMB Listener逻辑
- ✅ 实现FTP Listener逻辑
- ✅ 扩展Store接口
- ✅ 添加Handler端点
- ✅ 添加模型测试
- ✅ 更新文档
- ✅ 创建Phase 14总结文档

**交付物**：
- internal/listener/model.go - 扩展的数据模型
- internal/listener/smb.go - SMB Listener实现
- internal/listener/ftp.go - FTP Listener实现
- internal/listener/store.go - 扩展的Store接口
- internal/listener/handler.go - 扩展的Handler端点
- internal/listener/model_test.go - SMB/FTP测试
- internal/listener/README.md - 更新的文档
- doc/phase14-summary.md - Phase 14总结文档

### Phase 15：企业级功能 ✅ 已完成

- ✅ 实现数据保留策略
- ✅ 实现数据归档功能
- ✅ 实现高可用配置
- ✅ 添加企业级测试
- ✅ 创建文档
- ✅ 创建Phase 15总结文档

**交付物**：
- internal/retention/model.go - 数据保留和归档模型
- internal/retention/service.go - 保留和归档服务
- internal/retention/store.go - XORM存储实现
- internal/retention/model_test.go - 模型测试
- internal/retention/README.md - 数据保留和归档文档
- internal/ha/model.go - 高可用模型
- internal/ha/service.go - 高可用服务
- internal/ha/store.go - XORM存储实现
- internal/ha/model_test.go - 模型测试
- internal/ha/README.md - 高可用文档
- doc/phase15-summary.md - Phase 15总结文档

### Phase 16：插件市场/模板市场 ✅ 已完成

- ✅ 实现插件模型
- ✅ 实现模板模型
- ✅ 实现插件服务
- ✅ 实现模板服务
- ✅ 添加插件/模板测试
- ✅ 创建文档
- ✅ 创建Phase 16总结文档

**交付物**：
- internal/marketplace/model.go - 插件和模板数据模型
- internal/marketplace/service.go - 市场服务实现
- internal/marketplace/store.go - XORM存储实现
- internal/marketplace/model_test.go - 模型测试
- internal/marketplace/README.md - 市场文档
- doc/phase16-summary.md - Phase 16总结文档

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

- ✅ `ROADMAP_2.0.md`：产品路线图
- ✅ `DEVELOPMENT_PLAN_2.0.md`：工程开发计划
- ✅ OpenAPI 文档：`docs/openapi.yaml`
- ✅ Docker Compose 示例：`docker-compose.yml`
- ✅ CLI 使用文档：`docs/CLI_USAGE.md`
- ✅ MCP Server 使用文档：`docs/MCP_SERVER_USAGE.md`
- ✅ Nuclei、Burp/ZAP、CI/CD 集成示例：`examples/nuclei/`, `examples/ci/`, `extensions/burp/`
