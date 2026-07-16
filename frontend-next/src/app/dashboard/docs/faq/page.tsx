'use client'

import { DocPageLayout } from '../doc-page-layout'

const content = `## Q: How do I change the admin password?

Use the \`resetpw\` subcommand:

\`\`\`bash
godnslog resetpw -driver sqlite -dsn "file:godnslog.db"
\`\`\`

Or change it from the **Settings** page in the web UI.

## Q: Why am I not seeing any interactions?

Check the following:

- Ensure your domain's NS records point to the GODNSLOG server IP.
- Verify the DNS server is running on port 53 (\`dig @your-server-ip test.your-domain.com\`).
- Confirm the payload token matches the one generated in the UI.
- Check that the target application actually makes outbound DNS/HTTP requests.
- Review server logs for errors.

## Q: What protocols are supported?

- **DNS**: Full support (A, AAAA, CNAME, TXT, MX, NS, wildcard, xip)
- **HTTP**: Full support (all methods, headers, body capture)
- **SMTP**: Listener-based (start a listener on port 25)
- **LDAP**: Listener-based (start a listener on port 389)
- **SMB**: Listener-based (start a listener on port 445)
- **FTP**: Listener-based (start a listener on port 21)

## Q: How does token attribution work?

Each payload gets a unique token embedded in the domain (e.g., \`abc123.your-domain.com\`). When a DNS query or HTTP request hits the server, the token is extracted from the subdomain and used to look up the associated Payload and Case automatically.

## Q: Can I use GODNSLOG with Nuclei?

Yes. Configure Nuclei to use your GODNSLOG instance as the OAST server:

\`\`\`bash
nuclei -t templates/ -interactsh-url your-domain.com -interactsh-token your-token
\`\`\`

You can also use the **Scanner Hub** page to create scanner runs and backfill results for automatic correlation.

## Q: How do I export evidence for a report?

Navigate to the **Evidence** page, select a Case or Payload ID, choose format (Markdown or JSON), and click Generate. You can also use the CLI:

\`\`\`bash
godnslog-cli report export --case-id <ID> --format markdown
\`\`\`

## Q: How do I set up HA (High Availability)?

1. Deploy multiple GODNSLOG instances behind a load balancer.
2. Configure Redis for session sharing (\`-redis-addr\`).
3. Use a shared MySQL database.
4. Register nodes via the cluster API (\`/api/v2/cluster/nodes\`).

## Q: What is the MCP Server and how do I use it?

The MCP (Model Context Protocol) server provides 13 tools for AI Agent integration, including \`create_oast_probe\`, \`wait_for_interaction\`, and \`summarize_evidence\`. Start it with:

\`\`\`bash
godnslog mcp-server
\`\`\`

Agents connect via JSON-RPC 2.0 at \`POST /api/v2/mcp\` using APIKey auth.

## Q: How do I migrate from 1.0 to 2.0?

Use the migration tool to sync 1.0 DNS/HTTP records to the 2.0 Interaction model:

\`\`\`bash
godnslog migrate -driver sqlite -dsn "file:godnslog.db"
\`\`\``

export default function FaqDoc() {
  return <DocPageLayout titleKey="docs.faq" md={content} />
}
