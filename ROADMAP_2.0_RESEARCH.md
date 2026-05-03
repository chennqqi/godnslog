# GODNSLOG 2.0 调研与创新能力规划

## 一句话定位

GODNSLOG 2.0 不应只是“新版 DNSLOG/HTTPLOG”，而应升级为 **OAST 交互验证与安全证据平台**：统一生成 Payload、捕获多协议回连、自动归因、沉淀证据、联动扫描器与通知工作流，帮助安全团队更快确认盲区漏洞和异步风险。

## 调研结论

### 行业方向

PortSwigger 将 OAST 定义为 DAST 的重要增强，用外部可控服务发现传统响应中不可见的漏洞，例如盲注、盲 XSS、盲命令执行和异步触发问题。Burp Collaborator 的关键价值不是“收到 DNS 请求”，而是能把交互精准关联到触发它的请求与 Payload。

ProjectDiscovery Interactsh 已经把 OOB 能力扩展到 DNS、HTTP(S)、SMTP(S)、LDAP，并提供客户端、服务端、自托管、加密通信、动态响应和多域名能力。Nuclei 也通过 `{{interactsh-url}}` 把 OOB 交互自动关联到模板和请求。

Webhook.site 的价值点则偏向工作流：收到请求后可执行条件判断、变量提取、转发、修改响应、异步队列、重放和自动化动作。这说明 2.0 不能只做“记录列表”，还要做“请求触发后的自动处理”。

Canarytokens 的 DNS Token 说明另一类场景：安全团队需要长期布设轻量级诱饵与触发器，触发后带上可读上下文，用于暴露访问、泄露、扫描或横向移动迹象。

## 1.0 能力再定位

1.0 已经具备 DNSLOG、HTTPLOG、Rebinding、Callback、多用户、SDK、标准 DNS 解析、xip 等基础能力。它的优势是“轻量、自托管、贴近漏洞验证”。2.0 应保留这些优势，但把能力抽象成更高层的五个核心对象：

- **Interaction**：任何外部回连事件，包含 DNS、HTTP、SMTP、LDAP、SMB、FTP 等。
- **Payload**：用于触发 Interaction 的可追踪载荷。
- **Case**：一次测试任务、漏洞验证或演练场景。
- **Evidence**：可复现、可导出、可审计的证据链。
- **Workflow**：命中后的自动通知、转发、标记、响应或二次探测。

## 2.0 核心创新能力

### 1. OAST 交互中枢

将 DNSLOG/HTTPLOG 升级为统一 OAST 事件总线。

- 支持 DNS、HTTP、HTTPS、SMTP、LDAP 作为优先级第一批协议。
- 后续扩展 SMB、FTP、Redis、TCP Raw Listener，用于内网探测和协议型回连识别。
- 每次事件统一记录来源 IP、协议、Token、原始报文、解析结果、时间线、关联 Case。
- 支持“延迟回连”场景，事件发生在数小时或数天后仍能关联到原始 Payload。
- 提供交互风险判定：仅 DNS 查询、DNS+HTTP 请求、带敏感 Header、带云元数据路径、疑似代理访问等。

### 2. Payload Studio

把“复制一个域名去测”升级为可编排的 Payload 工作台。

- 内置 SSRF、XXE、RCE、SSTI、反序列化、模板注入、PDF/HTML 渲染、Webhook、CI/CD、云元数据探测等场景模板。
- 每个 Payload 自动生成独立 Token、用途说明、期望交互、风险等级和过期时间。
- 支持变量：`{token}`、`{case}`、`{domain}`、`{callback_url}`、`{base32_context}`。
- 支持批量生成：同一 Case 下为多个目标、多个参数、多个协议生成不同 Payload。
- 支持 Payload 生命周期：草稿、已投放、已命中、已归档、已过期。

### 3. 证据链与自动归因

2.0 的核心吸引力应是“收到回连后直接告诉你它意味着什么”。

- 自动把 Interaction 关联到 Payload、Case、目标、测试人、投放时间。
- 自动生成证据时间线：Payload 创建、投放、DNS 查询、HTTP 请求、通知发送、人工备注。
- 对命中结果进行分类：可疑出网、服务端请求伪造、异步任务执行、外部资源加载、扫描器噪声。
- 生成一键报告：Markdown、JSON、CSV，包含请求详情、DNS 详情、截图占位、复现 Payload 和修复建议占位。
- 支持证据脱敏，避免导出真实 Token、Cookie、Authorization、内网 IP。

### 4. Workflow 自动化

参考 Webhook.site 的工作流思路，命中事件后不只“展示”，还要“处理”。

- 条件规则：协议、Token、来源 IP、路径、Header、Body、关键词、Case、严重级别。
- 动作：通知、打标签、转发 Webhook、修改响应、保存附件、丢弃噪声、创建报告、调用外部 API。
- 支持同步响应控制：为特定路径返回自定义状态码、Header、Body、重定向或文件。
- 支持异步队列：慢操作不阻塞回连响应。
- 支持重放：对历史命中重新执行 Workflow，便于修复失败通知或补充分析。

### 5. Canary 与持续监测

把一次性漏洞验证扩展到长期安全监测。

- 创建 DNS、HTTP、文档、配置文件、CI 变量、对象存储、邮件地址等 Canary Token。
- Token 支持备注和编码上下文，例如项目、资产、投放位置、负责人。
- 支持过期策略和静默窗口，降低长期噪声。
- 命中后展示“这是哪个诱饵在哪里被触发”，而不是只显示一次请求。
- 可用于供应链泄露、配置泄露、离职账号访问、扫描器误触、内网横向移动提示。

### 6. Scanner Hub

GODNSLOG 2.0 应成为扫描器和人工测试之间的桥。

- 提供 Burp/ZAP 插件或轻量扩展：一键生成 Payload、拉取命中、回填备注。
- 提供 Nuclei 集成：兼容类似 `{{godnslog-url}}` 的占位符思路，支持模板执行后的交互关联。
- 提供 CLI：创建 Case、生成 Payload、轮询命中、导出报告。
- 提供 GitHub Actions/GitLab CI 示例，在测试环境中验证 SSRF/Webhook/回调能力。
- 提供 OpenAPI，方便企业内部扫描平台接入。

### 7. Rebinding Lab

1.0 已有 Rebinding 能力，2.0 可以将其产品化为实验室。

- 可视化阶段配置：首次解析、后续解析、TTL、目标 IP、命中条件。
- 内置场景：浏览器 DNS Rebinding、云元数据访问、内网管理面探测、IoT/路由器场景。
- 展示浏览器访问链路、DNS 解析变化和 HTTP 请求时间线。
- 提供安全防护提示：Host 校验、DNS Pinning、Metadata 防护、内网访问限制。

### 8. AI 辅助分析，但不作为核心依赖

AI 能力用于降噪和解释，不应阻塞离线、自托管和隐私场景。

- 对命中事件生成摘要：发生了什么、可能风险、下一步验证建议。
- 自动提取关键字段：来源、目标、Payload 类型、疑似漏洞类别。
- 对大量命中进行聚类，识别同源噪声、扫描器行为和真实业务触发。
- 为报告生成初稿，但必须保留人工确认状态。
- 默认本地规则优先，AI 分析作为可选插件。

## 产品形态规划

### 模式一：个人安全测试模式

面向渗透测试、SRC、Bug Bounty 和研发自测。

- 单用户或轻量多用户。
- 快速生成 Payload。
- 实时命中提醒。
- 一键导出证据。
- Docker Compose 快速启动。

### 模式二：团队协作模式

面向企业安全团队。

- 工作空间、项目、Case、角色权限。
- 审计日志和数据保留策略。
- 扫描器集成和统一报告。
- 企业通知渠道。
- 多域名、多租户隔离。

### 模式三：持续监测模式

面向长期 Canary 和外部交互监控。

- 长期 Token。
- 命中分级。
- 噪声压制。
- 自动升级通知。
- 周报/月报。

## 前端体验升级

技术栈仍建议使用 Next.js + TypeScript + shadcn/ui + Tailwind CSS，但界面重点不应是“管理后台换皮”，而是围绕安全测试流程重构：

- **Command Center**：展示活跃 Case、最近命中、高风险交互、系统状态。
- **Interaction Timeline**：按时间线查看一次命中的 DNS、HTTP、后续动作和备注。
- **Payload Studio**：像 IDE 一样编辑、预览、复制和批量生成 Payload。
- **Case Board**：按目标或漏洞类型管理验证任务。
- **Workflow Builder**：用条件和动作搭建命中后的自动化处理。
- **Evidence Report**：从命中记录直接生成报告草稿。
- **Rebinding Lab**：可视化 DNS 阶段与浏览器访问链路。

## 后端架构建议

### 核心模块

- `interaction`：统一事件接入、解析、存储。
- `payload`：Token、模板、变量、生命周期。
- `case`：任务、目标、证据、协作。
- `workflow`：规则、动作、队列、重放。
- `listener`：DNS、HTTP、SMTP、LDAP 等协议监听。
- `integrations`：Burp、ZAP、Nuclei、Webhook、通知平台。
- `canary`：长期诱饵 Token 和触发策略。
- `auth`：用户、角色、API Token、审计。

### 数据能力

- 事件原始数据和结构化字段分层存储。
- 热数据用于快速检索，冷数据用于归档和报告。
- 支持按工作空间、Case、Token、协议、来源 IP 分区查询。
- 大字段如 Body、附件、原始报文可单独存储，避免拖慢列表查询。

## 优先级路线

### P0：建立 2.0 核心闭环

- 新前端骨架：Next.js + shadcn/ui。
- 统一 Interaction 模型。
- DNS/HTTP 监听兼容旧能力。
- Payload Studio 最小版本。
- Case 与 Token 关联。
- 命中后自动归因和时间线。

### P1：形成差异化

- Workflow 自动化。
- Canary Token。
- 报告导出。
- 通知规则。
- Nuclei/CLI 集成。

### P2：增强竞争力

- SMTP/LDAP 监听。
- Rebinding Lab。
- Burp/ZAP 扩展。
- AI 辅助摘要与聚类。
- 多工作空间和审计。

### P3：平台化

- 插件市场或模板市场。
- 多节点 Listener。
- 企业级数据保留和归档。
- 高可用部署。
- SaaS/私有化双形态准备。

## 2.0 最小可发布版本

MVP 不建议做成“大而全后台”，而是聚焦一个有传播力的闭环：

1. 创建 Case。
2. 在 Payload Studio 选择漏洞场景并生成 Payload。
3. 目标触发 DNS/HTTP 回连。
4. Interaction Timeline 自动归因并解释命中。
5. 一键生成证据报告。
6. 命中通过飞书/企业微信/Webhook 推送。

这个闭环比“新版记录列表”更容易体现 2.0 价值。

## 与现有 ROADMAP 的关系

`ROADMAP_2.0.md` 可以作为工程实施路线，本文作为产品能力和创新方向补充。建议后续将两者合并为：

- `ROADMAP_2.0.md`：面向开发执行。
- `PRODUCT_2.0.md`：面向产品定位、能力边界和版本价值。
- `ARCHITECTURE_2.0.md`：面向模块设计、数据模型和接口规范。

## 参考资料

- PortSwigger OAST 介绍：<https://portswigger.net/burp/application-security-testing/oast>
- Burp Collaborator 文档：<https://portswigger.net/burp/documentation/collaborator>
- Burp Collaborator 典型用途：<https://portswigger.net/burp/documentation/collaborator/uses>
- ProjectDiscovery Interactsh 文档：<https://docs.projectdiscovery.io/tools/interactsh>
- Nuclei OOB Testing 文档：<https://docs.projectdiscovery.io/templates/reference/oob-testing>
- Nuclei 运行与 Interactsh 参数：<https://docs.projectdiscovery.io/tools/nuclei/running>
- Webhook.site Custom Actions：<https://docs.webhook.site/custom-actions.html>
- Canarytokens DNS Token：<https://docs.canarytokens.org/guide/dns-token.html>
