'use client'

import { DocPageLayout } from '../doc-page-layout'

export default function ApiDoc() {
  return (
    <DocPageLayout titleKey="docs.api">
      <h2>Authentication</h2>
      <p>All API requests require a JWT token via the <code>Access-Token</code> header:</p>
      <pre><code>{`POST /api/v2/auth/login
Content-Type: application/json

{"username": "admin", "password": "your-password"}`}</code></pre>
      <p>Response:</p>
      <pre><code>{`{
  "code": 0,
  "message": "OK",
  "data": {
    "token": "eyJ...",
    "user": { "id": 1, "username": "admin", "role": 0 }
  }
}`}</code></pre>

      <h2>Core Endpoints</h2>
      <h3>Cases</h3>
      <ul>
        <li><code>GET /api/v2/cases</code> — List cases (paginated)</li>
        <li><code>POST /api/v2/cases</code> — Create a case</li>
        <li><code>GET /api/v2/cases/:id</code> — Get case detail</li>
        <li><code>PUT /api/v2/cases/:id</code> — Update case</li>
        <li><code>DELETE /api/v2/cases/:id</code> — Delete case</li>
        <li><code>GET /api/v2/cases/:id/stats</code> — Case statistics</li>
        <li><code>GET /api/v2/cases/:id/payloads</code> — Payloads in case</li>
        <li><code>GET /api/v2/cases/:id/interactions</code> — Interactions in case</li>
      </ul>

      <h3>Payloads</h3>
      <ul>
        <li><code>GET /api/v2/payloads</code> — List payloads</li>
        <li><code>POST /api/v2/payloads</code> — Create payload</li>
        <li><code>POST /api/v2/payloads/batch</code> — Batch create</li>
        <li><code>GET /api/v2/payloads/:id</code> — Get payload</li>
        <li><code>PUT /api/v2/payloads/:id</code> — Update payload</li>
        <li><code>POST /api/v2/payloads/:id/revoke</code> — Revoke payload</li>
        <li><code>POST /api/v2/payloads/:id/preview</code> — Preview rendered payload</li>
      </ul>

      <h3>Interactions</h3>
      <ul>
        <li><code>GET /api/v2/interactions</code> — List interactions (filter by type, token, IP, time)</li>
        <li><code>GET /api/v2/interactions/stats</code> — Aggregated statistics</li>
        <li><code>GET /api/v2/interactions/timeline</code> — Timeline view</li>
        <li><code>GET /api/v2/interactions/stream</code> — SSE stream for real-time updates</li>
        <li><code>POST /api/v2/interactions/export</code> — Export (JSON/Markdown/CSV)</li>
      </ul>

      <h3>Other Modules</h3>
      <ul>
        <li><code>/api/v2/apikeys</code> — API Key management</li>
        <li><code>/api/v2/users</code> — User management (admin)</li>
        <li><code>/api/v2/canary</code> — Canary tokens</li>
        <li><code>/api/v2/rebinding</code> — DNS rebinding rules and scenarios</li>
        <li><code>/api/v2/listeners</code> — Protocol listeners (SMTP/LDAP/SMB/FTP)</li>
        <li><code>/api/v2/notifications/channels</code> — Notification channels</li>
        <li><code>/api/v2/retention/policies</code> — Data retention</li>
        <li><code>/api/v2/scanner-hub/adapters</code> — Scanner integration</li>
        <li><code>/api/v2/agent-runs</code> — Agent run tracking</li>
        <li><code>/api/v2/evidence/generate</code> — Evidence generation</li>
        <li><code>/api/v2/settings</code> — System settings</li>
        <li><code>/api/v2/dns/records</code> — DNS record management</li>
        <li><code>/api/v2/cluster</code> — HA cluster management</li>
      </ul>

      <h3>MCP (JSON-RPC 2.0)</h3>
      <p><code>POST /api/v2/mcp</code> — Streamable HTTP transport for AI Agent integration</p>
      <p>Available tools: <code>create_oast_probe</code>, <code>create_case</code>, <code>create_payload</code>, <code>list_interactions</code>, <code>wait_for_interaction</code>, <code>summarize_evidence</code>, <code>export_report</code>, <code>get_evidence_summary</code>, <code>explain_evidence</code>, <code>list_agent_runs</code>, <code>get_agent_run</code>, <code>complete_agent_run</code>, <code>revoke_token</code></p>

      <h2>Swagger UI</h2>
      <p>Full interactive API documentation is available at <code>/swagger/index.html</code> when the server is started with the <code>-swagger</code> flag.</p>
    </DocPageLayout>
  )
}
