# GODNSLOG 2.0 Roadmap

## 一句话定位

GODNSLOG 2.0 不再只是 DNSLOG/HTTPLOG 工具，而是面向安全测试、扫描器协同和 AI Agent 的 **OAST 交互验证与证据平台**。它应负责生成可追踪 Payload、捕获多协议回连、自动归因、沉淀证据、触发工作流，并通过 API/MCP 为外部工具赋能。

## 设计原则

- **自托管优先**：安全团队可私有部署，敏感交互数据不离开组织。
- **API 先行**：前端、CLI、SDK、扫描器插件和 MCP Server 共用同一套 API。
- **证据优先**：所有功能最终服务于确认漏洞、解释命中、导出证据。
- **协同优先**：不替代 Nuclei、Burp、ZAP、YApi、CI/CD，而是增强它们。
- **权限优先**：用户、扫描器、APIKey、Agent 都必须有最小权限和审计。

## 核心对象

- **Case**：一次测试任务、漏洞验证、项目或演练。
- **Payload**：带 Token、场景、变量和生命周期的可追踪载荷。
- **Interaction**：DNS、HTTP、SMTP、LDAP 等外部回连事件。
- **Evidence**：由 Payload、Interaction、时间线和备注组成的证据链。
- **Workflow**：命中后的通知、转发、打标、响应控制和自动处理。
- **Canary**：长期布设的诱饵 Token，用于泄露、访问和横向移动监测。

## 核心能力

### 0. 1.0的核心功能需要保留
- 保留用户管理功能
- 保留API交互功能
- 保留日志功能
- 保留在平台可以直接看到使用文档，API部分可以改进为swagger，使用文档保留markdown渲染，并增加2.0的文档

### 1. OAST 交互中枢

- 统一接入 DNS、HTTP、HTTPS，后续扩展 SMTP、LDAP、SMB、FTP、TCP Raw Listener。
- 支持延迟回连，数小时或数天后仍可关联到原始 Payload。
- 记录来源 IP、协议、Token、原始报文、解析结果、Case 和时间线。
- 自动识别仅 DNS 查询、DNS+HTTP、敏感 Header、云元数据路径、代理访问等风险特征。

### 2. Payload Studio

- 内置 SSRF、XXE、RFI、RCE、Blind SQLi、SSTI、反序列化、CORS/JSONP、SMTP injection、PDF/HTML 渲染、Webhook、CI/CD、云元数据探测等模板。
- 支持变量：`{token}`、`{case}`、`{domain}`、`{callback_url}`、`{base32_context}`。
- 支持批量生成、独立追踪 Token、过期时间、场景说明和期望回连协议。
- 支持 Payload 生命周期：草稿、已投放、已命中、已归档、已过期。

### 3. 证据链与自动归因

- 自动把 Interaction 关联到 Payload、Case、目标、测试人和投放时间。
- 生成证据时间线：Payload 创建、投放、DNS 查询、HTTP 请求、通知发送、人工备注。
- 自动分类命中类型：可疑出网、SSRF、外部资源加载、异步任务执行、扫描器噪声。
- 支持 Markdown、JSON、CSV 报告导出，并提供脱敏策略。

### 4. Scanner Hub

- Nuclei：提供 `godnslog-cli`、JSONL 输出、模板示例和类似 `{{godnslog-url}}` 的集成思路。
- Burp/ZAP：提供插件或扩展，支持右键生成 Payload、插入请求、拉取命中、回填备注。
- YApi/OpenAPI：导入接口定义，批量向参数、Header、Body、URL 注入 OAST Payload。
- CI/CD：提供 GitHub Actions、GitLab CI、Jenkins 示例；高危命中可作为流水线门禁。

### 5. Agent API 与 MCP

- 提供稳定 REST API 和 OpenAPI 文档，覆盖 Case、Payload、Interaction、Evidence、Workflow、Token。
- 提供 `godnslog-mcp-server`，为 AI Agent 暴露受控工具：`create_case`、`create_payload`、`list_interactions`、`wait_for_interaction`、`summarize_evidence`、`export_report`、`revoke_token`。
- MCP Client 使用独立 APIKey，必须支持作用域、过期时间、审计日志和高风险动作限制。
- GODNSLOG 在 Agent 工作流中定位为外部感知层和证据系统，而不是黑盒自动攻击器。

### 6. Workflow 自动化

- 条件：协议、Token、来源 IP、路径、Header、Body、关键词、Case、风险等级。
- 动作：通知、打标签、转发 Webhook、修改响应、保存附件、丢弃噪声、创建报告、调用外部 API。
- 支持同步响应控制、自定义状态码/Header/Body/重定向/文件。
- 支持异步队列和历史命中重放。

### 7. Canary 持续监测

- 创建 DNS、HTTP、文档、配置文件、CI 变量、对象存储、邮件地址等 Canary Token。
- Token 支持备注和编码上下文，例如项目、资产、投放位置、负责人。
- 支持过期、静默窗口、重复命中压缩和分级通知。
- 面向供应链泄露、配置泄露、离职账号访问、扫描器误触和内网横向移动提示。

### 8. Rebinding Lab 与高级 DNS

- 可视化配置首次解析、后续解析、TTL、目标 IP 和命中条件。
- 内置浏览器 DNS Rebinding、云元数据访问、内网管理面探测、IoT/路由器场景。
- DNS ReverseProxy 作为受控实验能力，用于域名映射、解析策略和 Rebinding 验证。
- DNS C2 仅作为授权实验能力，默认关闭并强制审计。

## 前端产品形态

- **Command Center**：活跃 Case、最近命中、高风险交互、系统状态。
- **Payload Studio**：像 IDE 一样编辑、预览、复制和批量生成 Payload。
- **Interaction Timeline**：按时间线展示 DNS、HTTP、后续动作和备注。
- **Case Board**：按目标、漏洞类型或项目管理验证任务。
- **Workflow Builder**：配置命中后的条件和动作。
- **Evidence Report**：从命中记录直接生成报告草稿。
- **Rebinding Lab**：可视化 DNS 阶段与访问链路。

## 版本路线

### 2.0 MVP：核心闭环 

- Next.js + TypeScript + shadcn/ui 新前端。
- APIKey、OpenAPI、Case/Payload/Interaction 核心 API。
- DNS/HTTP OAST 兼容旧能力。
- Payload Studio 支持 SSRF、XXE、RCE、Blind SQLi。
- 命中后自动归因、时间线和 Markdown/JSON 证据导出。
- Webhook、企业微信、飞书通知。
- Nuclei/CLI 最小集成。

### 2.1：扫描器协同版 

- Burp/ZAP 插件。
- CI/CD 示例和门禁能力。
- 更完整的 Payload 模板库。
- 命中聚类、噪声压缩和报告增强。

### 2.2：Agent 赋能版 

- MCP Server。
- Agent 专用最小权限 APIKey。
- `wait_for_interaction` 等异步工具。
- Agent 操作审计。
- AI 摘要、证据解释和报告初稿，默认作为可选插件。

### 2.3：平台化版本

- Canary 长期监测。
- Rebinding Lab 完整版。
- SMTP/LDAP/SMB/FTP Listener。
- 多工作空间、多域名、多 Listener 节点。
- 企业级数据保留、归档和高可用部署。
- 插件市场或模板市场。

## 参考方向

- PortSwigger OAST / Burp Collaborator：强调不可见漏洞检测和请求归因。
- ProjectDiscovery Interactsh / Nuclei：强调多协议 OOB、模板化和扫描器集成。
- Webhook.site：强调请求触发后的工作流、变量、转发、重放和响应控制。
- Canarytokens：强调长期诱饵、上下文编码和触发告警。
