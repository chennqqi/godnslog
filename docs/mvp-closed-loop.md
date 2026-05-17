# GODNSLOG 2.0 MVP Closed-Loop Scope

**Purpose**: Define the minimum viable closed-loop for the first MVP delivery. This scope is intentionally minimal to ensure the first release delivers a complete, working OAST evidence chain.

**Principle**: Complete workflow over feature breadth. The first MVP must deliver the full Probe → Interaction → Evidence → Export → Audit loop with minimum viable pages, APIs, and tools.

## The Closed-Loop Workflow

### Step 1: Create Probe
**Goal**: Generate an OAST probe ready for delivery to a target.

**Minimum Viable Implementation**:
- **Entity**: Case + Payload (Probe is implicit delivery context)
- **API**: `POST /api/v2/cases` (creates case) + `POST /api/v2/payloads` (creates payload)
- **CLI**: `godnslog create-case` + `godnslog create-payload`
- **MCP**: `create_oast_probe` (creates case + payload in one call)
- **Web**: Case creation page + Payload studio page

**Minimum Data**:
- Case: title, description (optional), status (default: active)
- Payload: template_id, variables (optional), expires_in (default: 24h)
- Output: case_id, payload_id, token, rendered_payload

**Acceptance Criteria**:
- [ ] Can create a Case via API
- [ ] Can create a Payload via API
- [ ] Can create a Probe via MCP (single call)
- [ ] Payload renders template with variables
- [ ] Token is unique and correlation-ready

### Step 2: Distribute to Scanner or Agent
**Goal**: Deliver the probe to a target via scanner integration or agent action.

**Minimum Viable Implementation**:
- **Entity**: Probe (payload_id + target context)
- **Scanner**: Nuclei JSONL export (Primary only)
- **Agent**: MCP `create_oast_probe` returns probe details
- **Web**: Copy probe to clipboard (manual distribution)

**Minimum Data**:
- Probe ID: case_id:payload_id
- Target: domain/IP/URL
- Delivery method: manual/scanner/agent
- Expected protocols: dns, http

**Acceptance Criteria**:
- [ ] Can export probe details in JSON format
- [ ] Can copy probe payload to clipboard
- [ ] Nuclei can use probe in template (documented)
- [ ] Agent receives probe details via MCP

**Note**: Step 2 is primarily documentation and export in MVP. Deep scanner integration comes later.

### Step 3: Capture Interaction
**Goal**: Capture out-of-band events triggered by the probe.

**Minimum Viable Implementation**:
- **Entity**: Interaction
- **Protocols**: DNS, HTTP only (SMTP/LDAP/SMB/FTP in later phases)
- **API**: `GET /api/v2/interactions?token={token}` (query by token)
- **CLI**: `godnslog wait-interaction --token {token}`
- **MCP**: `wait_for_interaction` (polling)
- **Web**: Interaction timeline page (filtered by case or token)

**Minimum Data**:
- Interaction: id, type (dns/http), token, payload_id, case_id, source_ip, timestamp, data
- DNS data: query_type, query_name, answer
- HTTP data: method, path, headers, body (first 1KB)

**Acceptance Criteria**:
- [ ] DNS queries are captured and stored
- [ ] HTTP requests are captured and stored
- [ ] Interactions are attributed to payload and case
- [ ] Can query interactions by token via API
- [ ] Can wait for interactions via CLI/MCP
- [ ] Web UI shows interaction timeline

### Step 4: Form Evidence
**Goal**: Aggregate interactions into explainable evidence.

**Minimum Viable Implementation**:
- **Entity**: Evidence
- **API**: `GET /api/v2/evidence/{case_id}` (generate evidence for case)
- **CLI**: `godnslog export-evidence --case-id {case_id}`
- **MCP**: `export_evidence` (export evidence for case)
- **Web**: Evidence timeline page (auto-generated when viewing case)

**Minimum Data**:
- Evidence: id, case_id, evidence_strength (low/medium/high), confidence (0-100), timeline, explainability
- Timeline: chronological list of interactions with timestamps
- Explainability: "Captured X interactions from Y unique sources. Evidence strength: Z."

**Acceptance Criteria**:
- [ ] Can generate evidence for a case via API
- [ ] Evidence includes interaction timeline
- [ ] Evidence includes strength and confidence scores
- [ ] Evidence includes human-readable explanation
- [ ] Can export evidence via CLI/MCP
- [ ] Web UI auto-generates evidence for case

**Scoring Logic (MVP)**:
- Low: 0-2 interactions
- Medium: 3-5 interactions from ≥2 unique sources
- High: 6+ interactions from ≥3 unique sources
- Confidence: Based on interaction count and source diversity

### Step 5: Export Results and Audit
**Goal**: Export evidence in standard format and log all operations for audit.

**Minimum Viable Implementation**:
- **Entity**: Audit Event
- **Export Formats**: JSON, Markdown (PDF in later phases)
- **API**: `GET /api/v2/evidence/{case_id}?format=json|md`
- **CLI**: `godnslog export-evidence --format json|md`
- **MCP**: `export_evidence` with format parameter
- **Web**: Evidence export buttons (JSON, Markdown)

**Minimum Audit Fields**:
- Audit Event: id, timestamp, operator_id (or agent_id/agent_run_id), action, resource, risk_level, parameters, result
- Risk Levels: low (read/export), medium (create/update), high (delete), critical (config change)
- Actions: create_case, create_payload, wait_interaction, export_evidence, delete_payload

**Acceptance Criteria**:
- [ ] Can export evidence in JSON format
- [ ] Can export evidence in Markdown format
- [ ] All create_probe operations are logged in audit
- [ ] All wait_interaction operations are logged in audit
- [ ] All export_evidence operations are logged in audit
- [ ] Audit logs are queryable via API
- [ ] Audit logs include agent_id and agent_run_id for agent operations

## Minimum Viable Pages (Web UI)

### 1. Case List Page
**URL**: `/dashboard/cases`
**Purpose**: View and manage cases

**Minimum Features**:
- Table showing: title, status, created_at, interaction_count
- Filters: status (active/completed), date range
- Search: by title
- Actions: Create Case button, View Case button
- Pagination: 20 items per page

**Not in MVP**:
- Bulk actions
- Tags
- Advanced filtering

### 2. Case Detail Page
**URL**: `/dashboard/cases/{case_id}`
**Purpose**: View case details and associated payloads

**Minimum Features**:
- Case metadata: title, description, status, created_at
- Payloads list: token, template, created_at, interaction_count
- Actions: Create Payload button, Export Evidence button
- Evidence summary: evidence_strength, confidence, interaction_count

**Not in MVP**:
- Case editing
- Case deletion
- Tags
- Collaboration features

### 3. Payload Studio Page
**URL**: `/dashboard/payloads`
**Purpose**: Create and manage payloads

**Minimum Features**:
- Template selection dropdown (5 core templates: ssrf-basic, xxe-basic, rce-basic, blind-sqli, dns-rebinding)
- Variable input form (dynamic based on template)
- Live preview of rendered payload
- Payload list: token, template, created_at, status
- Actions: Create Payload button, Copy Token button

**Not in MVP**:
- Template center
- Custom template creation
- Batch generation
- Payload editing

### 4. Interaction Timeline Page
**URL**: `/dashboard/interactions`
**Purpose**: View captured interactions

**Minimum Features**:
- Timeline view of interactions (chronological cards)
- Filters: token, case_id, protocol (dns/http), date range
- Search: by token or source_ip
- Interaction detail drawer (click to expand)
- Actions: Export Interactions button

**Not in MVP**:
- Clustering view
- Noise reduction
- Advanced filtering

### 5. Evidence Timeline Page
**URL**: `/dashboard/evidence/{case_id}`
**Purpose**: View evidence for a case

**Minimum Features**:
- Evidence summary: strength, confidence, explainability
- Interaction timeline (same as Interaction Timeline but scoped to case)
- Export buttons: JSON, Markdown

**Not in MVP**:
- PDF export
- Evidence editing
- SARIF export

### 6. Audit Log Page
**URL**: `/dashboard/audit` (Admin/Operator only)
**Purpose**: View audit logs

**Minimum Features**:
- Audit log table: timestamp, operator/agent, action, resource, risk_level, result
- Filters: operator_id, agent_id, risk_level, date range
- Search: by action or resource
- Pagination: 50 items per page

**Not in MVP**:
- Alert configuration
- Audit log export
- Advanced filtering

## Minimum Viable API Endpoints

### Cases
- `POST /api/v2/cases` - Create case
- `GET /api/v2/cases` - List cases (with pagination and filters)
- `GET /api/v2/cases/{id}` - Get case details

### Payloads
- `POST /api/v2/payloads` - Create payload
- `GET /api/v2/payloads` - List payloads (with pagination and filters)
- `GET /api/v2/payloads/{id}` - Get payload details

### Interactions
- `GET /api/v2/interactions` - List interactions (with filters: token, case_id, protocol, date_range)
- `GET /api/v2/interactions/{id}` - Get interaction details

### Evidence
- `GET /api/v2/evidence/{case_id}` - Generate evidence for case
- `GET /api/v2/evidence/{case_id}?format=json|md` - Export evidence

### Audit
- `GET /api/v2/audit` - List audit logs (with filters: operator_id, agent_id, risk_level, date_range)

## Minimum Viable CLI Commands

- `godnslog create-case --title "SSRF Test" --description "Testing SSRF vulnerability"` - Create case
- `godnslog create-payload --case-id {id} --template ssrf-basic --target example.com` - Create payload
- `godnslog wait-interaction --token {token} --timeout 300` - Wait for interaction
- `godnslog export-evidence --case-id {id} --format json` - Export evidence
- `godnslog audit-log --operator-id {id} --risk-level high` - Query audit log

## Minimum Viable MCP Tools

- `create_oast_probe(title, template, target, variables)` - Create probe (case + payload)
- `wait_for_interaction(token, timeout, expected_count)` - Wait for interaction
- `export_evidence(case_id, format)` - Export evidence

**Note**: Agent management tools (create_agent, list_agent_runs) are not in MVP closed-loop but are in Agent governance scope (separate track).

## Minimum Viable Audit Fields

### All Audit Events Must Include:
- `id`: UUID
- `timestamp`: UTC timestamp
- `operator_id`: User ID (or null for agent)
- `agent_id`: Agent ID (or null for user)
- `agent_run_id`: Agent Run ID (or null for user)
- `action`: Action type
- `resource`: Resource type and ID
- `risk_level`: low/medium/high/critical
- `parameters`: Full operation parameters
- `result`: success/failure
- `error`: Error message (if failed)
- `ip_address`: Source IP
- `user_agent`: User agent string

### MVP Audited Actions:
- `create_case` (medium risk)
- `create_payload` (medium risk)
- `wait_interaction` (low risk)
- `export_evidence` (low risk)
- `delete_payload` (high risk) - optional in MVP

## MVP Exclusions (Deliberate Scope Boundaries)

### Not in MVP Closed-Loop:
- Multi-protocol listeners (SMTP/LDAP/SMB/FTP) - Phase 2
- Canary tokens - Separate feature track
- DNS Rebinding Lab - Separate feature track
- Workflow automation - Phase 2
- Scanner Hub adapters (beyond Nuclei JSONL) - Phase 2
- Agent management UI - Phase 2
- Workspace isolation - Phase 2
- Data retention policies - Phase 2
- Notification channels - Phase 2
- Template center - Phase 2
- Clustering and noise reduction - Phase 2
- SARIF export - Phase 2
- PDF export - Phase 2

### Rationale for Exclusions:
- Focus on core OAST evidence chain first
- Ensure complete workflow before expanding features
- Prevent scope creep in first delivery
- Allow for iterative enhancement based on MVP feedback

## MVP Success Criteria

The MVP is considered successful when:
1. A user can create a case and payload via Web UI
2. A user can copy the payload and trigger an interaction
3. The interaction appears in the Web UI timeline
4. The user can generate and export evidence for the case
5. All operations are logged in the audit trail
6. An agent can perform the same workflow via MCP
7. The audit trail shows all agent operations with full attribution

## Next Phase Scope (Post-MVP)

After MVP validation, the next phase should add:
1. Multi-protocol listeners (SMTP/LDAP/SMB/FTP)
2. Scanner Hub adapters (Burp, ZAP, Yakit)
3. Agent management UI
4. Workspace isolation
5. Workflow automation
6. Canary tokens
7. DNS Rebinding Lab
8. Advanced evidence features (clustering, SARIF)

This phased approach ensures the MVP delivers a complete, working OAST evidence chain before expanding feature breadth.
