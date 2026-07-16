'use client'

import { DocPageLayout } from '../doc-page-layout'

const en = `## Q: How do I change the admin password?

Use the \`resetpw\` subcommand:

\`\`\`bash
godnslog resetpw -driver sqlite -dsn "file:godnslog.db"
\`\`\`

Or change it from the **Settings** page in the web UI.

## Q: Why am I not seeing any interactions?

Check the following:

- Ensure your domain's NS records point to the GODNSLOG server IP.
- Verify DNS server is running on port 53 (\`dig @your-server-ip test.your-domain.com\`).
- Confirm the payload token matches the one generated in the UI.
- Check that the target application makes outbound DNS/HTTP requests.

## Q: What protocols are supported?

- **DNS**: Full support (A, AAAA, CNAME, TXT, MX, NS, wildcard, xip)
- **HTTP**: Full support (all methods, headers, body capture)
- **SMTP/LDAP/SMB/FTP**: Listener-based OAST protocols

## Q: How does token attribution work?

Each payload gets a unique token embedded in the domain (e.g., \`abc123.your-domain.com\`). When a DNS query or HTTP request hits the server, the token is extracted from the subdomain and used to look up the associated Payload and Case automatically.

## Q: Can I use GODNSLOG with Nuclei?

Yes. Configure Nuclei to use your GODNSLOG instance as the OAST server:

\`\`\`bash
nuclei -t templates/ -interactsh-url your-domain.com -interactsh-token your-token
\`\`\`

Also use the **Scanner Hub** to create scanner runs and backfill results.

## Q: How do I export evidence?

Navigate to the **Evidence** page, select a Case or Payload, choose format (Markdown/JSON), and click Generate.

\`\`\`bash
godnslog-cli report export --case-id <ID> --format markdown
\`\`\`

## Q: How do I set up HA?

1. Deploy multiple instances behind a load balancer.
2. Configure Redis for session sharing (\`-redis-addr\`).
3. Use a shared MySQL database.
4. Register nodes via \`/api/v2/cluster/nodes\`.

## Q: What is the MCP Server?

The MCP server provides 13 tools for AI Agent integration. Start with \`godnslog mcp-server\` or connect directly to \`POST /api/v2/mcp\` with an API key.`

const zh = `## 问：如何修改管理员密码？

使用 \`resetpw\` 子命令：

\`\`\`bash
godnslog resetpw -driver sqlite -dsn "file:godnslog.db"
\`\`\`

或在 Web UI 的**设置**页面修改。

## 问：为什么看不到交互记录？

请检查：

- 域名的 NS 记录是否指向 GODNSLOG 服务器 IP。
- DNS 服务器是否在 53 端口运行（\`dig @your-server-ip test.your-domain.com\`）。
- Payload Token 是否与 UI 中生成的一致。
- 目标应用是否确实发出了 DNS/HTTP 请求。

## 问：支持哪些协议？

- **DNS**：完整支持（A、AAAA、CNAME、TXT、MX、NS、通配符、xip）
- **HTTP**：完整支持（所有方法、头、正文捕获）
- **SMTP/LDAP/SMB/FTP**：基于监听器的 OAST 协议

## 问：Token 归因如何工作？

每个 Payload 在域名中嵌入唯一 Token（如 \`abc123.your-domain.com\`）。当 DNS 查询或 HTTP 请求到达服务器时，从子域名中提取 Token 并自动查找关联的 Payload 和 Case。

## 问：能否与 Nuclei 一起使用？

可以。配置 Nuclei 使用您的 GODNSLOG 实例作为 OAST 服务器：

\`\`\`bash
nuclei -t templates/ -interactsh-url your-domain.com -interactsh-token your-token
\`\`\`

也可使用 **Scanner Hub** 创建扫描运行并回填结果。

## 问：如何导出证据？

进入**证据**页面，选择 Case 或 Payload，选择格式（Markdown/JSON），点击生成。

\`\`\`bash
godnslog-cli report export --case-id <ID> --format markdown
\`\`\`

## 问：如何配置高可用（HA）？

1. 部署多个实例在负载均衡后面。
2. 配置 Redis 共享会话（\`-redis-addr\`）。
3. 使用共享 MySQL 数据库。
4. 通过 \`/api/v2/cluster/nodes\` 注册节点。

## 问：MCP 服务器是什么？

MCP 服务器为 AI Agent 集成提供 13 个工具。使用 \`godnslog mcp-server\` 启动，或通过 API Key 直接连接 \`POST /api/v2/mcp\`。`

export default function FaqDoc() {
  return <DocPageLayout titleKey="docs.faq" mdEn={en} mdZh={zh} />
}
