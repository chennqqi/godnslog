# GODNSLOG 2.0 融合规划：OAST、扫描器协同与 Agent 赋能

## 结论

README 中作者原有 Roadmap 可以和 2.0 规划融合，而且应当作为 2.0 的核心来源之一。原 Roadmap 已经指出了几个正确方向：前端重构、APIKey/Swagger、更多 SDK、通知、多漏洞检测、DNS ReverseProxy、Nuclei/Burp/YApi 集成，以及 MCP/A2A 连接大模型和 Agent。

2.0 不应只是“更现代的 DNSLOG”，而应定位为：**面向扫描器、AI Agent 和安全团队的 OAST 交互验证基础设施**。

## 从 1.0 Roadmap 到 2.0 能力域

### 易用性升级

原计划中的“前端重构、APIKey 认证、Swagger UI、更多 SDK、类 Firebase SaaS 服务”应整合为 2.0 的开发者体验能力：

- 前端使用 Next.js + shadcn/ui + TypeScript，替换旧 Vue 2。
- 提供 OpenAPI/Swagger 文档，所有核心能力先 API 化。
- APIKey 升级为作用域 Token：支持过期时间、权限范围、最后使用时间和撤销。
- SDK 覆盖 Go、Python、JavaScript/TypeScript，优先服务扫描器、CI 和 Agent。
- 保留自托管优先，同时预留 SaaS/托管版形态。

### 通知能力升级

原计划的邮件、微信、企业微信、Slack、Discord、Webhook 通知应升级为规则化通知系统：

- 按协议、Payload、Case、来源 IP、Header、Body 关键词触发通知。
- 支持通知降噪、聚合、静默窗口和重复命中压缩。
- 支持飞书、企业微信、钉钉、Slack、Discord、Telegram、Email、Webhook。
- 通知内容包含可读证据摘要，而不只是“收到一条记录”。

### 检测能力升级

原计划中的 SSRF、XXE、RFI、RCE、Blind SQL injection、反序列化、CORS/JSONP、SMTP injection，应沉淀为 Payload Studio：

- 每类漏洞提供 Payload 模板、使用说明、期望回连协议和风险解释。
- 每个 Payload 自动绑定 Token、Case 和投放上下文。
- 支持 DNS、HTTP、HTTPS、SMTP、LDAP 等多协议 OAST 验证。
- 命中后自动归因到漏洞类型，并生成证据时间线。

### 工具能力升级

原计划中的 DNS ReverseProxy、DNS C2、EMAIL、SMS 不宜简单做成零散工具，应拆成可控模块：

- DNS ReverseProxy：用于内网探测、域名映射、Rebinding Lab 和高级解析策略。
- EMAIL/SMTP Listener：用于 SMTP injection、邮件回调、邮件型 Canary。
- SMS：作为通知通道或外部集成，低优先级。
- DNS C2：应谨慎定位为受控实验能力，默认关闭，强调授权测试和审计。

## 作为既有扫描器的协同扩展

这是 2.0 最值得强化的方向之一。GODNSLOG 不需要替代扫描器，而应成为扫描器的 OAST 后端、Payload 生成器和证据归因中心。

### Nuclei 协同

- 提供 `godnslog-cli`，支持创建 Case、生成 OAST URL、轮询命中和导出结果。
- 提供 Nuclei helper 模式，类似 `{{interactsh-url}}` 的思路生成 `{{godnslog-url}}`。
- 支持把 Nuclei template、目标 URL、请求编号与回连事件自动关联。
- 提供模板示例：SSRF、XXE、Blind SQLi、RCE、SMTP injection。
- 输出 JSONL 结果，便于被 Nuclei、CI 或内部平台消费。

### Burp Suite / ZAP 协同

- 提供 Burp/ZAP 插件：右键生成 Payload、插入请求、拉取命中、回填备注。
- 支持从 Proxy/Repeater/Intruder/Scanner 请求创建 Case。
- 命中后展示“哪个请求、哪个参数、哪个 Payload 触发了回连”。
- 支持私有 OAST Server，满足企业内网和隐私要求。

### YApi / OpenAPI / API 测试平台协同

- 支持导入 OpenAPI/YApi 接口定义，批量生成测试 Payload。
- 对接口参数、Header、Body、URL 参数自动插入 OAST Payload。
- 将命中结果回写到接口测试报告或导出为独立证据。

### CI/CD 协同

- GitHub Actions、GitLab CI、Jenkins 插件或脚本示例。
- 在测试环境中验证 Webhook、SSRF、防火墙出网和回调能力。
- 支持失败门禁：命中特定高危 OAST 事件时阻断流水线。

## 通过 AI Agent API / MCP 对外赋能

完全可行，而且很适合 GODNSLOG 2.0。AI Agent 在安全测试中经常缺少“真实世界反馈通道”，而 GODNSLOG 可以提供这个反馈通道：Agent 生成 Payload，投放到目标，然后通过 API/MCP 获取是否触发、触发了什么、证据是什么。

### API 优先设计

2.0 的核心对象都应可通过 API 操作：

- `Case`：创建测试任务，绑定目标、操作者、场景。
- `Payload`：按漏洞类型生成 Payload。
- `Interaction`：查询 DNS/HTTP/SMTP/LDAP 回连。
- `Evidence`：生成证据摘要和报告。
- `Workflow`：配置命中后的通知、转发、标记和响应控制。
- `Token`：创建、撤销、限定作用域。

### MCP Server 能力

可以提供 `godnslog-mcp-server`，向 AI Agent 暴露受控工具：

- `create_case`：创建一次测试任务。
- `create_payload`：生成指定类型的 OAST Payload。
- `list_interactions`：查询某个 Case 或 Token 的回连记录。
- `wait_for_interaction`：等待一段时间，看 Payload 是否被触发。
- `summarize_evidence`：生成命中摘要。
- `export_report`：导出 Markdown/JSON 证据。
- `revoke_token`：撤销测试 Token。

这些工具必须有权限边界，避免 Agent 滥用：

- 每个 MCP Client 绑定独立 APIKey。
- APIKey 必须有最小权限和过期时间。
- 高风险动作如创建长期 Canary、修改响应、启用 DNS C2 默认禁止。
- 所有 Agent 操作进入审计日志。

### A2A / Agent 工作流

2.0 可以支持多个 Agent 或工具协作：

- Recon Agent 发现目标接口。
- Payload Agent 生成 OAST Payload。
- Scanner Agent 投放 Payload。
- GODNSLOG 监听回连并归因。
- Report Agent 生成证据报告。
- Human Reviewer 最终确认。

GODNSLOG 在这里不是“AI 扫描器”，而是 Agent 的外部感知层和证据系统。

## 建议版本拆分

### 2.0 MVP

- Next.js 新前端。
- APIKey + OpenAPI。
- Case / Payload / Interaction 三个核心对象。
- DNS/HTTP OAST 兼容旧能力。
- Payload Studio 支持 SSRF、XXE、RCE、Blind SQLi。
- Nuclei/CLI 最小集成。
- Webhook、企业微信、飞书通知。
- Markdown/JSON 证据导出。

### 2.1 扫描器协同版

- Burp/ZAP 插件。
- YApi/OpenAPI 导入。
- CI/CD 示例和门禁能力。
- 更完整的 Payload 模板库。
- 回连自动归因和命中聚类。

### 2.2 Agent 赋能版

- MCP Server。
- Agent 专用最小权限 APIKey。
- `wait_for_interaction` 等异步工具。
- Agent 操作审计。
- AI 摘要、证据解释和报告初稿。

### 2.3 平台化版本

- 多工作空间、多域名、多 Listener 节点。
- Canary 长期监测。
- Rebinding Lab。
- SMTP/LDAP/SMB/FTP Listener。
- 模板市场或插件体系。

## 实施原则

- API 先行：前端、CLI、扫描器插件、MCP 都使用同一套 API。
- 自托管优先：安全团队能私有部署，避免敏感数据出域。
- 证据优先：所有能力最终服务于“确认漏洞、证明触发、导出证据”。
- 集成优先：不重复造扫描器，而是增强 Nuclei、Burp、ZAP、YApi、CI 和 Agent。
- 权限优先：Token、Agent、扫描器、用户都必须有可审计的最小权限。

## 与已有文档的关系

- `ROADMAP_2.0.md`：工程实施路线。
- `ROADMAP_2.0_RESEARCH.md`：调研与创新能力池。
- 本文档：融合 README 原 Roadmap 后的产品主线，重点明确扫描器协同和 AI Agent/MCP 赋能。
