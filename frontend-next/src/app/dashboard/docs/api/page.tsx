'use client'

import { DocPageLayout } from '../doc-page-layout'

const content = `## Authentication

All API requests require a JWT token via the \`Access-Token\` header:

\`\`\`bash
POST /api/v2/auth/login
Content-Type: application/json

{"username": "admin", "password": "your-password"}
\`\`\`

Response:

\`\`\`json
{
  "code": 0,
  "message": "OK",
  "data": {
    "token": "eyJ...",
    "user": { "id": 1, "username": "admin", "role": 0 }
  }
}
\`\`\`

## Core Endpoints

### Cases

- \`GET /api/v2/cases\` — List cases (paginated)
- \`POST /api/v2/cases\` — Create a case
- \`GET /api/v2/cases/:id\` — Get case detail
- \`PUT /api/v2/cases/:id\` — Update case
- \`DELETE /api/v2/cases/:id\` — Delete case
- \`GET /api/v2/cases/stats\` — Case statistics
- \`GET /api/v2/cases/:id/payloads\` — Payloads in case
- \`GET /api/v2/cases/:id/interactions\` — Interactions in case

### Payloads

- \`GET /api/v2/payloads\` — List payloads
- \`POST /api/v2/payloads\` — Create payload
- \`POST /api/v2/payloads/batch\` — Batch create
- \`GET /api/v2/payloads/:id\` — Get payload
- \`PUT /api/v2/payloads/:id\` — Update payload
- \`POST /api/v2/payloads/:id/revoke\` — Revoke payload
- \`POST /api/v2/payloads/:id/preview\` — Preview rendered payload

### Interactions

- \`GET /api/v2/interactions\` — List interactions (filter by type, token, IP, time)
- \`GET /api/v2/interactions/stats\` — Aggregated statistics
- \`GET /api/v2/interactions/timeline\` — Timeline view
- \`GET /api/v2/interactions/stream\` — SSE stream for real-time updates
- \`POST /api/v2/interactions/export\` — Export (JSON/Markdown/CSV)

### Other Modules

- \`/api/v2/apikeys\` — API Key management
- \`/api/v2/users\` — User management (admin)
- \`/api/v2/canary\` — Canary tokens
- \`/api/v2/rebinding\` — DNS rebinding rules and scenarios
- \`/api/v2/listeners\` — Protocol listeners (SMTP/LDAP/SMB/FTP)
- \`/api/v2/notifications/channels\` — Notification channels
- \`/api/v2/retention/policies\` — Data retention
- \`/api/v2/scanner-hub/adapters\` — Scanner integration
- \`/api/v2/agent-runs\` — Agent run tracking
- \`/api/v2/evidence/generate\` — Evidence generation
- \`/api/v2/settings\` — System settings
- \`/api/v2/dns/records\` — DNS record management
- \`/api/v2/cluster\` — HA cluster management

### MCP (JSON-RPC 2.0)

\`POST /api/v2/mcp\` — Streamable HTTP transport for AI Agent integration.

Available tools: \`create_oast_probe\`, \`create_case\`, \`create_payload\`, \`list_interactions\`, \`wait_for_interaction\`, \`summarize_evidence\`, \`export_report\`, \`get_evidence_summary\`, \`explain_evidence\`, \`list_agent_runs\`, \`get_agent_run\`, \`complete_agent_run\`, \`revoke_token\`.

## Swagger UI

Full interactive API documentation is available at \`/swagger/index.html\` when the server is started with the \`-swagger\` flag.`

export default function ApiDoc() {
  return <DocPageLayout titleKey="docs.api" md={content} />
}
