# GODNSLOG 2.0 Phase 4/5 审计发现与遗留问题

- **审计日期**：2026-06-21
- **审计基线**：`da91391`
- **审计范围**：Phase 4（Agent & Scanner Integration）+ Phase 5（Platform — Enterprise-Grade）
- **审计结论**：Phase 4 通过；Phase 5 功能实现完成，但遗留以下问题需在上线前修复或验证。

---

## 1. 安全类（必须修复 / 上线前验证）

### 1.1 多协议监听器安全审计未完成

- **问题**：`internal/listener/smtp.go`、`ldap.go`、`smb.go`、`ftp.go` 已实现真实 TCP listener，但 Phase 5a 验收标准中明确要求“Security review passed for all protocol listeners”。
- **风险**：监听器默认运行在主进程内，开放额外端口，可能被扫描/利用；SMB/LDAP 协议存在已知欺骗/反射攻击风险。
- **建议**：
  - 默认全部禁用，仅在配置中显式启用。
  - 生产环境将协议监听器放到独立进程或网络命名空间（如 sidecar、容器）。
  - 增加端口访问控制（iptables/安全组）、连接速率限制、TLS 选项。
  - 完成一次专门的安全审计并输出报告。

### 1.2 Telegram / Email 通知外发无安全沙箱

- **问题**：
  - `internal/workflow/service.go:541` 的 `sendNotifyTelegram` 直接访问 `https://api.telegram.org`。
  - `sendNotifyEmail` 使用 `smtp.SendMail`，未强制 TLS，且配置中的密码以明文形式存储在 action config。
- **风险**：外发请求可能被 SSRF 利用；邮件凭证泄露风险。
- **建议**：
  - 复用 `validateNotifyURL` 对 Telegram URL 进行白名单校验。
  - Email 默认要求 TLS/STARTTLS，凭证存储改为只保存 token 或引用密钥管理。

### 1.3 MCP 会话与 APIKey 权限

- **问题**：`internal/mcp/server.go` 的 session store 为内存实现，未与 Redis 等共享存储对接。
- **风险**：多实例部署时 MCP session 无法共享，导致负载均衡后 session 丢失。
- **建议**：提供基于 Redis 的 session store 实现，或在 HA 部署时启用 sticky session。

---

## 2. 架构与运维类

### 2.1 容器内使用 `&` 后台启动前后端

- **问题**：`Dockerfile` 的 `CMD` 使用 `sh -c "/app/godnslog serve ... & cd /app/frontend && npx next start ..."`。
- **风险**：任一进程崩溃不会被容器编排系统感知；无法优雅退出；日志混在一处。
- **建议**：
  - 方案 A：拆分为两个容器（backend + frontend），使用 `docker-compose` 分别管理。
  - 方案 B：单容器内引入进程监管（supervisor/s6/tini），确保僵尸进程回收与崩溃重启。

### 2.2 HEALTHCHECK 在 OCI 格式下被忽略

- **问题**：`docker build` 时提示 `HEALTHCHECK is not supported for OCI image format and will be ignored`。
- **风险**：生产环境无法依赖容器镜像内置健康检查。
- **建议**：使用 `docker` 镜像格式构建，或在编排层（compose/K8s）配置 liveness/readiness probe。

### 2.3 HA 服务尚未接入主服务启动流程

- **问题**：`internal/ha/service.go` 与 `store.go` 已实现，但需确认 `WebServer` 启动时是否初始化 HA 状态、是否注册节点心跳。
- **建议**：在 `server/webserver.go` 的 `Run()` 中添加 HA 初始化，提供 `/health/ha` 或类似端点；在 K8s 中验证 leader 选举。

### 2.4 Workflow Queue 无持久化

- **问题**：`internal/workflow/queue.go` 使用内存 channel，进程重启后未执行 job 丢失。
- **建议**：增加基于数据库/Redis 的持久化队列，或至少记录未执行 job 到 `TblWorkflowActionLog` 以便恢复。

---

## 3. 前端质量类

### 3.1 lint warnings 共 25 个

- **来源**：`npm --prefix ./frontend-next run lint` 返回 0 errors，25 warnings。
- **主要问题**：
  - `src/app/dashboard/retention/page.tsx` 未使用的 `Select` 组件导入。
  - `src/app/dashboard/users/page.tsx` 未使用的 `useCallback`。
  - `src/app/dashboard/settings/page.tsx` 未使用的 `setApiKeys`。
  - `src/app/dashboard/workflow/page.tsx` 未使用的 `conditions` / `actions`。
  - `src/app/dashboard/payloads/page.tsx` 缺少 `useEffect` 依赖 `updatePreview`。
  - `src/app/dashboard/payloads/new/page.tsx` 使用 `form.watch()` 触发 React Compiler 不兼容警告。
  - `src/features/interactions/hooks/use-interaction-stream.ts` 不必要的 `enabled` 依赖。
  - UI 组件（dropdown-menu/select）未使用的图标导入。
- **建议**：批量清理，优先处理 missing dependency 与不必要依赖，避免运行时 bug。

### 3.2 Phase 5 页面缺少 i18n 统一

- **问题**：
  - `src/app/dashboard/marketplace/page.tsx` 使用硬编码英文（`Plugins`、`Templates`、`Installed`、`Marketplace`）。
  - `src/app/dashboard/rebinding/page.tsx` 使用硬编码英文（`Rebinding Lab`、`Predefined Scenarios` 等）。
  - `src/app/dashboard/agent-runs/[id]/page.tsx` 中英文混用（`创建 Follow-up Action`、`Export JSON`、`Deliver to Webhook`、`操作历史`、`快速链接`）。
- **建议**：将 Phase 5 页面接入 `useI18n`，补充 `en-US` / `zh-CN` 翻译 key，确保与 Phase 3 统一。

### 3.3 前端测试与页面标签不匹配（已修复）

- **修复记录**：
  - `e2e/marketplace.spec.ts` 中文 tab 断言改为英文。
  - `e2e/rebinding.spec.ts` “添加阶段”按钮断言改为 `Predefined Scenarios` 卡片。
  - `e2e/agent-runs.spec.ts` follow-up 刷新请求 waiter 提前注册，避免竞态。
- **建议**：后续新增页面时，E2E 应与页面实际文案保持一致，避免硬编码语言断言。

### 3.4 Dialog 组件缺少可访问性描述

- **问题**：E2E 运行中多次出现 `Warning: Missing `Description` or `aria-describedby={undefined}` for {DialogContent}`。
- **建议**：为所有 `DialogContent` 补充 `DialogDescription` 或 `aria-describedby`，提升可访问性。

---

## 4. 测试与覆盖率类

### 4.1 `server` 包覆盖率偏低

- **问题**：`server` 包集成测试覆盖不足，Phase 4/5 新增的核心端点（MCP `/mcp`、scanner backfill、workflow trigger、agent-run lifecycle）缺少端到端 API 测试。
- **建议**：在 `server/v2_api_test.go` 或独立测试包中补充：
  - MCP initialize / tools/list / tools/call 流程。
  - workflow trigger 后 action 执行与通知日志。
  - scanner run backfill 解析 JSONL/SARIF 并关联 interaction。
  - agent-run follow-up / review decision / export / delivery。

### 4.2 多协议监听器缺少真实协议 E2E

- **问题**：当前 listener 测试集中在 `manager_test.go`、`model_test.go`、`ratelimit_test.go`、`security_context_test.go`，缺少真实 SMTP/LDAP/FTP/SMB 客户端交互。
- **建议**：补充集成测试，使用对应协议客户端发送请求，验证 interaction 写入与 attribution。

### 4.3 HA 与 Marketplace 缺少集成测试

- **问题**：`internal/ha` 与 `internal/marketplace` 有单元测试，但缺少与 WebServer/v2 API 的集成验证。
- **建议**：为 HA 节点注册/心跳、Marketplace 插件/模板安装流程补充 API 级测试。

---

## 5. 功能完善类

### 5.1 Custom HTTP Response 仅覆盖 HTTP 命中

- **问题**：`server/webapi.go:278` 在 HTTP record 存储后检查 `payload.CustomResponse`，但 DNS 命中路径未实现相同逻辑。
- **建议**：如果需求包含 DNS 响应定制（如特定 TXT/CNAME 记录），在 `server/webserver.go` 的 DNS 处理流程中补充。

### 5.2 Scanner Hub 创建 payload 报错（mock 环境下）

- **问题**：E2E 中 `scanner-hub/page.tsx:133` 报 `Failed to create payload: { code: 1, message: 'failed' }`，但测试仍通过。说明该错误路径被 swallow，未影响断言。
- **建议**：确认是 mock 数据缺失还是真实 API 缺陷；若是真实缺陷，修复 `payloadApi.create` 调用参数；若是 mock 问题，完善测试 mock。

### 5.3 Marketplace 安装语义不清晰

- **问题**：`src/app/dashboard/marketplace/page.tsx` 的 `installPlugin` 仅调用 `marketplaceApi.getPlugin(id)` 并将本地状态设为 installed，未真正执行安装动作。
- **建议**：后端提供 `install`/`uninstall` API，前端调用后刷新状态，并记录审计日志。

---

## 6. 优先级汇总

| 优先级 | 问题 | 负责人建议 |
|---|---|---|
| **P0** | 协议监听器安全审计 | 安全团队 |
| **P0** | 容器 `&` 启动与 HEALTHCHECK 问题 | DevOps |
| **P1** | 清理前端 lint warnings | 前端 |
| **P1** | Phase 5 页面 i18n 统一 | 前端 |
| **P1** | server 层集成测试补充 | 后端 |
| **P2** | Workflow Queue 持久化 | 后端 |
| **P2** | MCP session Redis 共享 | 后端 |
| **P2** | HA 接入主服务流程 | 后端 |
| **P2** | Marketplace 真实安装 API | 后端+前端 |
| **P3** | Dialog 可访问性描述 | 前端 |

---

## 7. 参考文档

- 验收报告：`docs/superpowers/acceptance/phase4-5-acceptance-report.md`
- 需求记录：`doc/requirements.md`、`doc/requirements-analysis.md`
- 设计文档：`docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §6/§7
