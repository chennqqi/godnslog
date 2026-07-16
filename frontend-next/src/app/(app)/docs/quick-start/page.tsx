'use client'

import { DocPageLayout } from '../doc-page-layout'

const zh = `## 1. 启动服务

使用您的域名和 IP 地址运行 GODNSLOG：

\`\`\`bash
go run . serve -domain example.com -4 127.0.0.1
\`\`\`

## 2. 登录

访问 \`http://localhost:80\`，使用控制台打印的自动生成的管理员凭据登录。登录后请立即在**设置**页面修改密码。

## 3. 创建 Case

进入**案例**页面，点击**新建案例**。输入标题和描述，开始追踪安全测试任务。

## 4. 生成 Payload

进入**Payload**页面，点击**新建 Payload**。选择模板（SSRF、XXE、RCE 等），关联 Case，系统自动生成基于唯一 Token 的 Payload URL。

## 5. 注入并监控

将生成的 Payload URL 复制到目标应用中进行注入。在**交互**页面实时查看 DNS/HTTP 回连。每次交互会自动关联到对应的 Case 和 Payload。

## 6. 导出证据

进入**证据**页面，选择 Case 或 Payload，选择格式（JSON、CSV、Markdown）下载带有归因元数据的交互日志。

## 7. 证据摘要

使用**证据摘要**页面从更高维度查看交互聚合，按技术类型、协议、置信度分组。摘要生成的证据包可直接供 AI 使用。`

const en = `## 1. Start the Server

Run GODNSLOG with your domain and IP address:

\`\`\`bash
go run . serve -domain example.com -4 127.0.0.1
\`\`\`

## 2. Log In

Navigate to \`http://localhost:80\` and log in with the auto-generated admin credentials printed in the server console. Change the password immediately via the **Settings** page.

## 3. Create a Case

Go to **Cases** and click **New Case**. Enter a title and description for your testing task.

## 4. Generate a Payload

Navigate to **Payloads** and click **New Payload**. Select a template (SSRF, XXE, RCE, etc.), choose your Case, and the system generates a unique token-based payload URL.

## 5. Inject and Monitor

Copy the generated payload URL and inject it into your target application. Watch the **Interactions** page for incoming DNS/HTTP callbacks. Each interaction is automatically attributed to your Case and Payload.

## 6. Export Evidence

Go to **Evidence** and select your Case or Payload. Choose a format (JSON, CSV, Markdown) to download the interaction log with attribution metadata.

## 7. Evidence Summary

For a higher-level view, use the **Evidence Summary** page to aggregate interactions by technique, protocol, and confidence level. The summary generates AI-ready evidence bundles.`

export default function QuickStartDoc() {
  return <DocPageLayout titleKey="docs.quick_start" mdEn={en} mdZh={zh} />
}
