'use client'

import { DocPageLayout } from '../doc-page-layout'

export default function FaqDoc() {
  return (
    <DocPageLayout titleKey="docs.faq">
      <h2>Q: How do I change the admin password?</h2>
      <p>
        Use the <code>resetpw</code> subcommand:
      </p>
      <pre><code>{`godnslog resetpw -driver sqlite -dsn "file:godnslog.db"`}</code></pre>
      <p>Or change it from the Settings page in the web UI.</p>

      <h2>Q: Why am I not seeing any interactions?</h2>
      <p>Check the following:</p>
      <ul>
        <li>Ensure your domain&apos;s NS records point to the GODNSLOG server IP.</li>
        <li>Verify the DNS server is running on port 53 (<code>dig @your-server-ip test.your-domain.com</code>).</li>
        <li>Confirm the payload token matches the one generated in the UI.</li>
        <li>Check that the target application actually makes outbound DNS/HTTP requests.</li>
        <li>Review server logs for errors.</li>
      </ul>

      <h2>Q: What protocols are supported?</h2>
      <ul>
        <li><strong>DNS</strong>: Full support (A, AAAA, CNAME, TXT, MX, NS, wildcard, xip)</li>
        <li><strong>HTTP</strong>: Full support (all methods, headers, body capture)</li>
        <li><strong>SMTP</strong>: Listener-based (start a listener on port 25)</li>
        <li><strong>LDAP</strong>: Listener-based (start a listener on port 389)</li>
        <li><strong>SMB</strong>: Listener-based (start a listener on port 445)</li>
        <li><strong>FTP</strong>: Listener-based (start a listener on port 21)</li>
      </ul>

      <h2>Q: How does token attribution work?</h2>
      <p>
        Each payload gets a unique token embedded in the domain (e.g., <code>abc123.your-domain.com</code>).
        When a DNS query or HTTP request hits the server, the token is extracted from the subdomain
        and used to look up the associated Payload and Case automatically.
      </p>

      <h2>Q: Can I use GODNSLOG with Nuclei?</h2>
      <p>
        Yes. Configure Nuclei to use your GODNSLOG instance as the OAST server:
      </p>
      <pre><code>{`nuclei -t templates/ -interactsh-url your-domain.com -interactsh-token your-token`}</code></pre>
      <p>
        You can also use the Scanner Hub page to create scanner runs and backfill results
        for automatic correlation.
      </p>

      <h2>Q: How do I export evidence for a report?</h2>
      <p>
        Navigate to the Evidence page, select a Case or Payload ID, choose format (Markdown or JSON),
        and click Generate. You can also use the CLI:
      </p>
      <pre><code>{`godnslog-cli report export --case-id <ID> --format markdown`}</code></pre>

      <h2>Q: How do I set up HA (High Availability)?</h2>
      <p>
        1. Deploy multiple GODNSLOG instances behind a load balancer.<br />
        2. Configure Redis for session sharing (<code>-redis-addr</code>).<br />
        3. Use a shared MySQL database.<br />
        4. Register nodes via the cluster API (<code>/api/v2/cluster/nodes</code>).
      </p>

      <h2>Q: What is the MCP Server and how do I use it?</h2>
      <p>
        The MCP (Model Context Protocol) server provides 13 tools for AI Agent integration,
        including <code>create_oast_probe</code>, <code>wait_for_interaction</code>, and
        <code>summarize_evidence</code>. Start it with:
      </p>
      <pre><code>{`godnslog mcp-server`}</code></pre>
      <p>Agents connect via JSON-RPC 2.0 at <code>POST /api/v2/mcp</code> using APIKey auth.</p>

      <h2>Q: How do I migrate from 1.0 to 2.0?</h2>
      <p>
        Use the migration tool to sync 1.0 DNS/HTTP records to the 2.0 Interaction model:
      </p>
      <pre><code>{`godnslog migrate -driver sqlite -dsn "file:godnslog.db"`}</code></pre>
    </DocPageLayout>
  )
}
