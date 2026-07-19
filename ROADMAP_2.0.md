# GODNSLOG 2.0 Roadmap

## 一句话定位

GODNSLOG 2.0 不再只是 DNSLOG/HTTPLOG 工具，而是面向安全团队、扫描器和 AI Agent 的 **自托管 OAST 证据中枢**。它应负责生成可追踪 Payload、捕获多协议回连、自动归因、沉淀可解释证据、触发工作流，并通过 API/MCP/插件为外部工具赋能。

## 设计原则

- **自托管优先**：安全团队可私有部署，敏感交互数据不离开组织。
- **API 先行**：前端、CLI、SDK、扫描器插件和 MCP Server 共用同一套 API。
- **证据优先**：所有功能最终服务于确认漏洞、解释命中、导出证据。
- **协同优先**：不替代 Nuclei、Burp、Yakit/Yak、ZAP、xray/rad、YApi、CI/CD，而是增强它们。
- **权限优先**：用户、扫描器、APIKey、Agent 都必须有最小权限和审计。
- **Agent 友好**：面向 AI Agent 提供异步等待、结构化证据、可恢复任务上下文和受控高层工具，而不是只暴露 CRUD API。

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
- 支持 **RMI 协议监听**，补齐 JNDI 注入检测场景的协议覆盖。
- 支持 **WebSocket 实时推送**，事件发生即推送至前端，提升交互式体验（参考 Hyuga / 商业平台）。
- HTTP 服务支持**请求回显**：在响应体中回显收到的完整请求（请求行+Header+Body），便于调试排查（参考商业平台）。

### 2. Payload Studio

- 内置 SSRF、XXE、RFI、RCE、Blind SQLi、SSTI、反序列化、CORS/JSONP、SMTP injection、PDF/HTML 渲染、Webhook、CI/CD、云元数据探测等模板。
- **反弹 Shell 命令生成器**：支持 bash/sh/nc/python/awk/telnet 等多种反弹 Shell 命令模板，自动填充 IP 和端口（参考 Alphalog）。
- 支持变量：`{token}`、`{case}`、`{domain}`、`{callback_url}`、`{base32_context}`。
- 支持批量生成、独立追踪 Token、过期时间、场景说明和期望回连协议。
- 支持 Payload 生命周期：草稿、已投放、已命中、已归档、已过期。

### 3. 证据链、增强分析与自动归因

- 自动把 Interaction 关联到 Payload、Case、目标、测试人和投放时间。
- **攻击链时间线**：按子域名 token 将 DNS/HTTP/LDAP/RMI 等多协议命中聚合成攻击链，时间轴展示完整攻击过程（DNS 解析 → HTTP 回连 → LDAP/JNDI 回连），每条命中附带利用类型标签和解码数据（参考商业平台）。
- **外带数据自动解码**：DNS 标签中的 base32/hex 自动识别并还原明文，HTTP body 中的 base64 自动解码并展示解码结果，消除手动解码环节（参考商业平台）。
- **利用类型自动标注**：根据 Payload 特征、路径、Header 等内容自动识别并标记 Log4Shell/JNDI、Fastjson、SSRF、XXE、SQLi 盲注等利用类型（参考商业平台/Alphalog）。
- **来源指纹归属**：自动判别请求来自扫描器、云厂商还是真实目标 IP，辅助判断漏洞真实性（参考商业平台）。
- 自动分类命中类型：可疑出网、SSRF、外部资源加载、异步任务执行、扫描器噪声。
- 提供 Evidence Score：按 DNS-only、DNS+HTTP、敏感 Header、云元数据路径、内部来源、延迟命中等维度评估证据强度。
- 生成 Explainable Evidence：解释命中意味着什么、置信度、误报线索和建议的下一步验证。
- 支持 Markdown、JSON、CSV 报告导出，并提供脱敏策略。

### 4. Scanner Hub：多工具 OAST 协同

- Nuclei/ProjectDiscovery：提供 `godnslog-cli`、JSONL 输出、模板示例、私有 OAST 后端集成和 CI 门禁。
- Burp Suite：提供扩展，支持右键生成 Payload、插入请求、拉取命中、回填备注和证据导出。
- Yakit/Yak：提供 Yak 脚本/插件，支持在 MITM、Web Fuzzer、PoC 执行和批量扫描中生成 OAST Payload、等待命中并回填结果。
- OWASP ZAP：提供脚本或 Add-on，支持主动扫描、请求重放和 OAST 命中轮询。
- xray/rad：提供 webhook/CLI/代理模式示例，用于爬虫、被动扫描和 OAST 结果关联。
- YApi/OpenAPI：导入接口定义，批量向参数、Header、Body、URL 注入 OAST Payload。
- Postman/Apifox/Hoppscotch：提供环境变量和 pre-request 脚本示例，便于 API 测试人员直接投放 Payload。
- CI/CD：提供 GitHub Actions、GitLab CI、Jenkins 示例；高危命中可作为流水线门禁。
- 通用集成协议：所有工具优先复用 REST API、OpenAPI、CLI、JSONL/SARIF 输出和 Webhook，插件只做体验增强。

### 5. Agent-Native API 与 MCP

- 提供稳定 REST API 和 OpenAPI 文档，覆盖 Case、Payload、Interaction、Evidence、Workflow、Token。
- 提供 `godnslog-mcp-server`，为 AI Agent 暴露受控工具：`create_oast_probe`、`create_case`、`create_payload`、`list_interactions`、`wait_for_interaction`、`summarize_evidence`、`export_report`、`revoke_token`。
- 引入 AgentRun/TaskRun 概念，记录 Agent 的测试目标、步骤、Payload、Interaction、决策日志和最终证据。
- 输出机器可读证据 Schema：漏洞类型、置信度、证据强度、命中详情、误报线索和下一步建议。
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

- **Command Center**：活跃 Case、最近命中、高风险交互、系统状态，以及 **Payload 速查表**（常用 Payload 复制即用）。
- **Payload Studio**：像 IDE 一样编辑、预览、复制和批量生成 Payload。
- **Interaction Timeline / 攻击链视图**：按时间线展示 DNS、HTTP、后续动作和备注；支持按 token **聚合为攻击链视图**，展示跨协议关联。
- **Case Board**：按目标、漏洞类型或项目管理验证任务。
- **Workflow Builder**：配置命中后的条件和动作。
- **Evidence Report**：从命中记录直接生成报告草稿。
- **Rebinding Lab**：可视化 DNS 阶段与访问链路。

## 版本路线

### 2.0 MVP：核心闭环 ✅

- ✅ Next.js + TypeScript + shadcn/ui 新前端。
- ✅ APIKey、OpenAPI、Case/Payload/Interaction 核心 API。
- ✅ DNS/HTTP OAST 兼容旧能力。
- ✅ Payload Studio 支持 SSRF、XXE、RCE、Blind SQLi。
- ✅ 命中后自动归因、时间线和 Markdown/JSON 证据导出。
- ✅ Webhook、企业微信、飞书通知。
- ✅ Nuclei/CLI 最小集成，并定义 Burp Suite、Yakit/Yak、ZAP、xray/rad 的集成协议和样例优先级。

### 2.1：扫描器协同版 ✅

- ✅ Burp Suite 插件。
- ✅ Yakit/Yak 插件或 Yak 脚本包。
- ✅ ZAP 脚本或 Add-on。
- ✅ xray/rad、Postman/Apifox 集成示例。
- ✅ CI/CD 示例和门禁能力。
- ✅ 更完整的 Payload 模板库。
- ✅ 命中聚类、噪声压缩和报告增强。

### 2.2：Agent 赋能版 ✅

- ✅ MCP Server。
- ✅ `create_oast_probe` 等高层 Agent 工具。
- ✅ AgentRun/TaskRun 和机器可读证据 Schema。
- ✅ Agent 专用最小权限 APIKey。
- ✅ `wait_for_interaction` 等异步工具。
- ✅ Agent 操作审计。
- ✅ AI 摘要、证据解释和报告初稿，默认作为可选插件。

### 2.3：平台化版本 ✅

- ✅ Canary 长期监测。
- ✅ Rebinding Lab 完整版。
- ✅ SMTP/LDAP/SMB/FTP Listener。
- ✅ 多工作空间、多域名、多 Listener 节点。
- ✅ 企业级数据保留、归档和高可用部署。
- ✅ 插件市场或模板市场。

### 2.4：智能增强版 ✅

- ✅ 攻击链时间线
- ✅ 外带数据自动解码
- ✅ 利用类型自动标注
- ✅ WebSocket 实时推送
- ✅ 反弹 Shell 命令生成器
- ✅ 补充通知渠道（Bark、Server酱）
- ✅ Payload 速查表

### 2.5：工具链深度集成 ✅

- ✅ 来源指纹归属
- ✅ Burp 风格轮询 API
- ✅ 自定义 HTTP Response
- ✅ 请求回显
- ✅ RMI 协议监听
- ⬜ 现有 Scanner Hub 集成完善（部分适配器可补充）

### 2.6：可扩展平台 ✅

- ✅ Template 插件化执行引擎（基础匹配引擎，待集成到 Interaction 管道）
- ✅ 匿名模式
- ✅ 搜索引擎集成（ZoomEye/Shodan/Fofa API 适配器）
- ✅ 任务化 Case 模型优化（新增批量字段 + stats API）

### 2.7：遗留补全（技术债与优化）✅

- **DNS 协议补全**：
  - ✅ IPv6 查询支持（AAAA 记录，已实现）
  - ✅ A 记录通配匹配（修复 `findResolves` 中忽略 store.Get 返回值 bug）
  - ✅ SRV 记录支持
  - ✅ NS 记录响应
- **测试覆盖提升**：
  - 🟡 `rule` 模块从 9% 提升至 51.7%（增加 engine/store/action 测试，handler 待补充）
  - ✅ `ha` 模块从 4% 提升至 75.2%
  - 🟡 `listener` 模块从 20% 提升至 31.4%（增加 store/service 测试，协议 handler 待补充）
  - ✅ `openapi` 添加基础测试（已有 6 个静态分析测试）
- **代码清理**：
  - ✅ 1.0 遗留路由端点实现（`server/router.go` 注释代码清理）
  - ✅ 中间件 `workspace_id` 集成（`auth/middleware.go`）
  - ✅ DNS 服务器 TODO 清理（IPv6 已实现，通配/SRV/NS 已修复）
  - ✅ `scannerhub` evidence 查询 TODO（改为基于已丰富交互的计数）

### 2.8：体验优化 ✅

- **通知系统增强**：
  - ✅ 通知发送 HTTP 超时配置
  - ✅ Telegram Markdown 转义处理
  - ✅ WebSocket Hub 优雅关闭
- **GeoIP 自动更新**：
  - ✅ 启动时自动检测/下载 GeoLite2-ASN.mmdb
- **前端 UI 组件补全**：
  - ✅ app-shell / sidebar / top-bar 布局组件（已有）
  - ✅ charts（折线图、环形图）可视化组件
  - ✅ kanban-board 看板组件
- **OpenAPI 文档完善**：
  - ❌ 新增端点的 API 文档同步（Phase 17-19 新端点无 swagger 注释）

### 2.9：部署模式与 Demo 环境 ⬜

- **双部署模式支持**：
  - ⬜ **Let's Encrypt 自动证书独立托管模式**：
    - 容器直接对外暴露，监听 443/80/53 端口
    - 自动申请并续期 Let's Encrypt 证书
    - 内置自签名证书兜底（证书申请失败时仍可启动）
    - DNSLog 验证功能在 443 和 80 上均兼容
    - Web 界面和 API 始终在 443 上提供服务
  - ⬜ **Nginx 反向代理模式**：
    - 不监听非特权端口（80/443），仅监听高位端口或 Unix Socket
    - 不再自动申请证书，TLS 由 Nginx 终结
    - 所有流量通过 Nginx 转发到后端
    - 提供标准 Nginx 配置指引（DNS over TLS、HTTP/HTTPS 反代、WebSocket 透传）
    - 提供 `deploy/nginx/` 参考配置模板
- **Demo 模式支持**：
  - ⬜ 自动创建若干 Demo 账号（含预置 Case、Payload、Interaction 样例数据）
  - ⬜ Demo 账号权限受限：禁止管理员操作、禁止删除/修改系统配置、禁止创建 APIKey
  - ⬜ Demo 账号数据隔离，定期自动重置
  - ⬜ 启动参数 `--demo` 或环境变量 `GODNSLOG_DEMO=true` 开启

## 参考方向

- PortSwigger OAST / Burp Collaborator：强调不可见漏洞检测和请求归因。
- ProjectDiscovery Interactsh / Nuclei：强调多协议 OOB、模板化和扫描器集成。
- Webhook.site：强调请求触发后的工作流、变量、转发、重放和响应控制。
- Canarytokens：强调长期诱饵、上下文编码和触发告警。
- [Hyuga](https://github.com/ac0d3r/Hyuga)：WebSocket 实时推送、第三方通知集成（Bark/Lark/钉钉/飞书/Server酱）、DNS Rebinding。
- [Antenna](https://github.com/wuba/Antenna)（58同城）：Template 插件化组件系统、任务驱动检测模型、OAST 全协议覆盖。
- [Alphalog](https://github.com/AlphabugX/Alphalog)：反弹 Shell 一键生成、匿名设计模式、纯 Redis 轻量部署。
- [Bridge](https://github.com/SPuerBRead/Bridge)：自定义 HTTP Response（状态码/Header/Body）、三层域名架构设计。
- [商业 dnslog 平台](https://mp.weixin.qq.com/s/8YovEBZq2VKNx4lRGfCccA)：攻击链时间线、外带数据自动解码、利用类型自动标注、来源指纹归属、Burp 风格轮询 API。
- https://github.com/ZackSecurity/Zack-AI-Scanner

## 参考项目特性映射

以下将参考资料分析中识别的高价值特性映射到现有路线图中，标注纳入位置：

| 参考来源 | 特性 | 纳入位置 | 状态 |
|----------|------|---------|------|
| 商业平台 | 攻击链时间线（跨协议聚合） | 核心能力 §3 | ✅ |
| 商业平台 | 外带数据自动解码（base32/hex/base64） | 核心能力 §3 | ✅ |
| 商业平台 | 利用类型自动标注 | 核心能力 §3 | ✅ |
| Hyuga | WebSocket 实时推送 | 核心能力 §1 | ✅ |
| Alphalog | 反弹 Shell 命令生成 | 核心能力 §2 | ✅ |
| 商业平台/Hyuga | 第三方通知渠道（Bark/Server酱） | 核心能力 §6（已有企微/飞书/钉钉，新增 Bark/Server酱） | ✅ |
| 商业平台 | 来源指纹归属 | 核心能力 §3 | ✅ |
| 商业平台 | Burp 风格轮询 API | 核心能力 §4 | ✅ |
| Bridge | 自定义 HTTP Response | 核心能力 §6（已有 Payload 级 + Workflow 级） | ✅ |
| 商业平台 | 请求回显 | 核心能力 §1 | ✅ |
| 商业平台 | Payload 速查表 | 前端 Command Center | ✅ |
| RMI | RMI 协议监听 | 核心能力 §1 | ✅ |
| Antenna | Template 插件化组件系统 | 核心能力 §6（基础执行引擎，待集成管道） | ⬜ |
| Antenna | 任务驱动模型 | Case 模块（新增批量字段 + stats API） | ✅ |
| Alphalog | 匿名模式 | 系统设置 | ✅ |
| POC-S | 搜索引擎集成（ZoomEye/Shodan/Fofa） | 核心能力 §4 | ✅ |
| 商业平台/Hyuga | 第三方通知补充（Telegram/Slack） | 核心能力 §6 | ✅ |
| 商业平台 | 通知 HTTP 超时配置 | 2.8 体验优化 | ❌ |
| 商业平台 | GeoIP ASN 自动更新 | 2.8 体验优化 | ❌ |
| — | DNS 协议补全（IPv6/通配/SRV/NS） | 2.7 遗留补全 | ✅ |
| — | 低模块测试覆盖（rule:51.7%/ha:75.2%/listener:31.4%） | 2.7 遗留补全 | ✅ |
| — | 代码清理（router/middleware/DNS TODO/scannerhub） | 2.7 遗留补全 | ✅ |
| — | WebSocket Hub 优雅关闭 | 2.8 体验优化 | ❌ |
| — | 前端 UI 组件补全（sidebar/charts/kanban） | 2.8 体验优化 | ❌ |

> P1=高优先级（2.4），P2=中优先级（2.5），P3=低优先级（2.6）

注：WEB-INF/web.xml 等自定义映射、参考项目通过 DNS 传递大文件等技术细节作为能力备选，不纳入正式路线图。
