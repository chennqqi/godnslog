'use client'

import { DocPageLayout } from '../doc-page-layout'

const en = `## Cases

A Case represents a testing task, vulnerability verification, or exercise project. Each Case can contain multiple Payloads and captures all related Interactions.

- **Create**: Navigate to Cases → New Case, fill in title, description, and tags.
- **Detail View**: Click a case to see stats, payloads, interactions timeline.
- **Edit/Close**: Update status (active/closed/archived) from the detail page.

## Payloads

Payloads are trackable probes with unique tokens. Each payload is associated with a Case and supports template-based generation with variable substitution.

- **Templates**: Built-in templates for SSRF, XXE, RCE, Blind SQLi, and more.
- **Variables**: \`{{token}}\`, \`{{case_id}}\`, \`{{domain}}\` are auto-replaced.
- **Preview**: Preview the rendered payload before injection.
- **Revoke**: Revoke a payload to stop tracking new interactions.

## Interactions

Captured out-of-band events (DNS queries, HTTP requests, SMTP connections, etc.), automatically attributed to Cases and Payloads via token matching.

- **Timeline View**: Visual chronological display.
- **Filters**: By protocol type, token, source IP, time range.
- **Real-time**: SSE stream endpoint for live updates.
- **Export**: JSON, Markdown, or CSV.

## Canary Tokens

Long-lived decoy probes for detecting unauthorized access or data exfiltration. Types: DNS, HTTP, web bug.

## DNS Rebinding

Configure rebinding rules to return different IPs on subsequent queries. Useful for SSRF testing.

## Protocol Listeners

Start SMTP, LDAP, SMB, or FTP listeners for OAST beyond DNS and HTTP. All **disabled by default**.

## Workflow Rules

Define rules that trigger actions (HTTP, DNS, webhook, notifications) when interactions match criteria.

## Scanner Hub

Integrate with Nuclei, Burp, ZAP, xray. Create scanner runs, backfill results, correlate findings.

## Agent Runs

Track AI Agent testing sessions with full review workflow and evidence export.`

const zh = `## Cases（案例）

案例代表一个测试任务或漏洞验证项目。每个 Case 可以包含多个 Payload，并关联所有相关交互记录。

- **创建**：进入案例页 → 新建案例，填写标题和描述。
- **详情**：点击案例查看统计、Payload 列表、交互时间线。
- **编辑/关闭**：在详情页更新状态（活跃/已完成/已归档）。

## Payloads（载荷）

Payload 是带有唯一 Token 的可追踪探测。每个 Payload 关联一个 Case，支持模板生成和变量替换。

- **模板**：内置 SSRF、XXE、RCE、盲 SQLi 等模板。
- **变量**：\`{{token}}\`、\`{{case_id}}\`、\`{{domain}}\` 会自动替换。
- **预览**：注入前可预览渲染后的 Payload。
- **撤销**：撤销后停止追踪新的交互。

## Interactions（交互）

捕获的带外事件（DNS 查询、HTTP 请求、SMTP 连接等），通过 Token 自动关联到 Case 和 Payload。

- **时间线视图**：按时间顺序可视化展示。
- **过滤器**：按协议类型、Token、源 IP、时间范围筛选。
- **实时流**：SSE 端点支持实时更新。
- **导出**：JSON、Markdown 或 CSV 格式。

## Canary 令牌

用于检测未授权访问或数据泄露的长效诱饵探针。支持 DNS、HTTP、网页漏洞类型。

## DNS 重绑定

配置重绑定规则，使后续查询返回不同 IP。用于 SSRF 测试突破访问控制。

## 协议监听器

启动 SMTP、LDAP、SMB 或 FTP 监听器，捕获 DNS/HTTP 以外的 OAST 回连。**默认全部关闭**。

## 工作流规则

定义规则，在交互匹配条件时自动触发动作（HTTP 请求、DNS 查询、Webhook、多渠道通知）。

## Scanner Hub

集成 Nuclei、Burp、ZAP、xray 等扫描器。创建扫描运行、回填结果、自动关联发现。

## Agent Runs（代理运行）

追踪 AI Agent 测试会话，支持完整的审查工作流和证据导出。`

export default function UserGuideDoc() {
  return <DocPageLayout titleKey="docs.user_guide" mdEn={en} mdZh={zh} />
}
