# Changelog

All notable changes to GODNSLOG will be documented in this file.

## [2.9.0] - 2026-07-19

### Added

#### 部署与运维
- TLS/ACME 支持：Let's Encrypt 自动证书管理、静态证书、自签名证书三种模式
- HA 高可用：领导者选举（基于 Redis + MySQL）、节点注册、故障转移
- Redis 会话共享支持（-redis 标志）
- K8s 完整部署清单：Deployment/Service/Ingress/PDB/健康检查/资源限制
- Nginx 反向代理参考配置（TLS 1.2/1.3、HSTS、WebSocket 支持）
- Docker 多模式部署：独立 TLS 模式、Nginx 反向代理模式、HA 模式
- 健康检查端点：`/api/v2/health`、`/api/v2/ready`
- Dockerfile HEALTHCHECK 和就绪探针
- Demo 模式：预置演示账户和样本数据，定时自动重置

#### 安全
- 登录滑块验证码（slide captcha）
- API Key 管理：作用域隔离、bcrypt 哈希、遗留密钥自动迁移
- 匿名模式：交互记录中隐藏源 IP
- 出站安全模块：IP 白名单、SSRF 防护、速率限制

#### 交互捕获与协议
- 多协议监听器：SMTP/LDAP/SMB/FTP/RMI（JNDI 注入检测）
- WebSocket Hub：实时交互推送
- EventSource/SSE 实时交互流
- Burp Collaborator 风格轮询 API（`GET /api/v2/poll`）
- HTTP 请求回显（Echo）
- 数据解码引擎：base64/base32/hex 自动检测和解码
- 利用分类器：基于规则匹配的攻击类型识别
- 源指纹检测：扫描器/云 CIDR + User-Agent 规则
- GeoIP 增强：MaxMind GeoLite2-ASN 自动下载，ASN/Org/Country 富化

#### 扫描器集成
- Scanner Hub：8 个扫描器适配器（Nuclei/Burp Suite/ZAP/Yakit/xray/rad/Postman/Apifox）
- 扫描结果回填：JSONL + SARIF 格式解析器
- 搜索引擎适配器：ZoomEye、Shodan、Fofa
- from-search 端点：从搜索结果批量创建扫描任务
- Burp Suite 扩展原型

#### 案例与工作流
- 看板视图：表格/看板双模式、拖拽状态更新
- 攻击链：数据模型、聚合逻辑、API 和前端组件
- 工作流自动化：HTTP/DNS/Webhook/Notify 动作执行器、异步队列与重试
- 自定义 HTTP 响应控制
- 模板执行引擎：检测插件模板

#### 仪表盘与可视化
- 协议分布环形图、7 天趋势折线图（recharts）
- 每日统计 API：`GET /api/v2/interactions/stats/daily`
- Payload Cheat Sheet
- 新用户欢迎横幅

#### 通知
- 通知渠道：Bark、ServerChan、Telegram、Slack、Discord、邮件（SMTP）
- 企业微信/飞书/钉钉 Webhook
- 可配置 HTTP 客户端（默认 30s 超时）

#### MCP 与 AI
- MCP Streamable HTTP 传输、会话管理、JSON-RPC 2.0 协议兼容
- Agent Run 全链路：创建→操作→状态更新→完成
- AI 证据摘要工具和 UI 页面
- 证据时间线 + 评分

#### CLI
- `resetpw` 子命令：重置管理员密码
- 子命令：scanner run/list/get、case close、payload revoke/preview
- JSON 输出格式支持

#### 数据管理
- 数据保留策略和归档
- 数据迁移工具：幂等批量导入、dry-run 支持
- 交互聚类与噪声压缩
- 审计日志 API（`GET /api/v2/audit/logs`）

### Changed
- Go 版本升级至 1.25（go.mod）
- 前端迁移至 Next.js 16 + TypeScript + Ant Design
- API v2 端点全面覆盖所有核心功能
- 认证简化：纯 JWT + localStorage，移除冗余 RSC 重定向逻辑
- 移除 `/dashboard` 路由前缀，路由结构扁平化
- 文档系统迁移至中英文双语 Markdown 渲染
- Swagger UI 默认启用
- 由 `google/subcommands` 框架管理 CLI 子命令

### Fixed
- 登录认证流程重构，修复 RSC 拦截和重定向问题
- 交互统计 SQL 错误（XORM 表名映射）
- 前端路由守卫和状态管理
- Telegram 通知 Markdown 特殊字符转义
- 搜索适配器数据提取（Fofa 总数修正）
- API 响应格式一致性
- 生产就绪项：gofmt（26 文件）、go vet（3 问题）、Dockerfile 路径、信号处理、不可达代码
- HA 领导者选举竞态条件和数据竞争
- E2E 测试套件适配
- i18n 切换和容器配置
- 工作流持久化和死信状态
- CI/CD 流水线稳定性

### Removed
- 旧版 Vue 前端（已被 Next.js 替换）
- 冗余 `router.push`/`router.replace` 用法，统一替换为 `window.location.href`
- 遗留的 `/dashboard` 路由前缀

**Docker Hub 发布：** 待确认（TODO）

## [2.0.0] - 2026-05-17

### Added
- Complete rewrite with enterprise-grade architecture
- OAST Evidence Platform with Case/Payload/Interaction tracking
- Agent-Native MCP Server for AI/LLM integration
- Scanner Hub with Nuclei, Burp Suite, ZAP, Yakit/Yak, xray/rad integration
- Workflow Automation with rule-based notification triggers
- Webhook notifications (Enterprise WeChat, Feishu, DingTalk)
- Canary Tokens for long-term monitoring (DNS/HTTP/SMTP)
- DNS Rebinding Lab with multi-stage rebinding
- Multi-protocol listeners (SMTP/LDAP/SMB/FTP)
- Multi-workspace isolation
- Data retention policies
- Audit logging
- CLI tool for OAST probe creation
- 30+ payload templates (SSRF/XXE/RCE/Blind SQLi/deserialization/LDAP/SMB/FTP/DNS Rebinding/Log4j JNDI/cloud metadata SSRF)
- Interaction clustering and noise reduction
- Evidence timeline with scoring
- Agent-specific API keys with scopes
- AI evidence summarization

### Changed
- Frontend migrated to Next.js 16 with TypeScript
- Backend refactored with internal/ directory structure
- Unified data models in internal/models/
- API v2 endpoints for all core features
- Docker build updated to use frontend-next
- Go version updated to 1.22
- Node version updated to 24.13.0

### Fixed
- Login authentication flow
- API response format consistency
- E2E test framework with Playwright
- Frontend routing and state management

### Removed
- Old Vue frontend (replaced by Next.js)
- Legacy API v1 endpoints (maintained for compatibility)

## [1.0.0] - Earlier

- Initial DNS/HTTP log server
- Basic user management
- DNS rebinding support
- Canary tokens
- Web UI
