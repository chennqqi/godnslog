# GODNSLOG 2.0 Unified Terminology and Domain Model

**Purpose**: This document defines the standard terminology for GODNSLOG 2.0. All documentation, API specifications, CLI commands, MCP tools, and implementation code must use these definitions consistently.

**Principle**: Single source of truth for core concepts. No concept drift across interfaces.

## Core Entities

### Case (案例)
**Definition**: A security engagement or testing session that groups related OAST activities.

**Attributes**:
- `id`: Unique identifier (UUID)
- `title`: Human-readable title
- `description`: Engagement description
- `status`: Current state (active, completed, archived)
- `created_by`: User who created the case
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp
- `tags`: Optional tags for organization

**Relationships**:
- Contains: Payloads
- Attributed to: Interactions
- Aggregated to: Evidence
- Owned by: User/Operator

**Usage Context**: Security Engineers organize testing work by case. Automation platforms create cases per engagement. Agents create cases per task.

### Payload (载荷)
**Definition**: A generated OAST probe template with rendered values, ready for delivery to targets.

**Attributes**:
- `id`: Unique identifier (UUID)
- `case_id`: Parent Case ID
- `token`: Unique token for interaction correlation
- `template_id`: Template used for generation
- `template_rendered`: Rendered payload with variables substituted
- `variables`: Variable values used in rendering
- `expires_at`: Expiration timestamp
- `created_at`: Creation timestamp
- `status`: Current state (active, expired, revoked)

**Relationships**:
- Belongs to: Case
- Captures: Interactions
- Referenced by: Evidence

**Usage Context**: Security Engineers generate payloads from templates. Automation platforms create payloads via API. Agents create payloads via MCP.

### Probe (探针)
**Definition**: The combination of a Payload and its delivery context (target, method, timing). A Probe is the actionable unit sent to a target.

**Attributes**:
- `probe_id`: Composite identifier (case_id:payload_id)
- `payload_id`: Underlying Payload ID
- `target`: Target system (domain, IP, URL)
- `delivery_method`: How probe was delivered (manual, scanner, agent)
- `delivered_at`: When probe was delivered
- `expected_protocols`: Protocols expected to trigger interaction

**Relationships**:
- References: Payload
- Triggers: Interactions
- Part of: Agent Run (if created by agent)

**Usage Context**: Probe represents the "moment of delivery" - the payload in context of target and delivery method. Used in audit logs and evidence attribution.

### Interaction (交互)
**Definition**: A captured out-of-band event (DNS, HTTP, SMTP, LDAP, SMB, FTP) triggered by a Probe.

**Attributes**:
- `id`: Unique identifier (UUID)
- `type`: Protocol type (dns, http, smtp, ldap, smb, ftp)
- `token`: Token from triggering payload
- `payload_id`: Payload that triggered this interaction
- `case_id`: Case associated with payload
- `source_ip`: Source IP address of request
- `timestamp`: When interaction occurred
- `data`: Protocol-specific data (query string, headers, body, etc.)
- `created_at`: Capture timestamp

**Relationships**:
- Triggered by: Payload
- Attributed to: Case
- Aggregated into: Evidence
- Logged in: Audit Event

**Usage Context**: The fundamental unit of captured evidence. All interactions are stored and attributed to their source payload and case.

### Evidence (证据)
**Definition**: Aggregated, scored, and explained summary of Interactions for a Case, providing explainable conclusions.

**Attributes**:
- `id`: Unique identifier (UUID)
- `case_id`: Parent Case ID
- `evidence_strength`: Qualitative assessment (low, medium, high, critical)
- `confidence`: Numerical confidence score (0-100)
- `timeline`: Chronological sequence of interactions
- `clustering`: Grouped interactions by pattern
- `explainability`: Human-readable explanation of findings
- `created_at`: Generation timestamp
- `updated_at`: Last update timestamp

**Relationships**:
- Aggregates: Interactions
- Belongs to: Case
- Exported via: API/CLI/Web

**Usage Context**: The output of the OAST process - not raw logs, but explainable conclusions for stakeholders.

### Agent (代理)
**Definition**: An AI/LLM entity with scoped permissions for OAST operations.

**Attributes**:
- `id`: Unique identifier (UUID)
- `name`: Agent name
- `api_key`: Agent-specific API key with `IsAgent` flag
- `scopes`: Permission scopes (create_probe, wait_interaction, export_evidence, delete_payload, modify_config)
- `workspace_id`: Workspace constraint (optional)
- `risk_tolerance`: Risk level tolerance (low, medium, high)
- `created_by`: Operator who created the agent
- `created_at`: Creation timestamp
- `expires_at`: API key expiration

**Relationships**:
- Owned by: Operator
- Creates: Agent Runs
- Operates in: Workspace (optional)

**Usage Context**: AI agents use MCP to interact with GODNSLOG. Agents have scoped permissions and are fully auditable.

### Agent Run (代理运行)
**Definition**: A single task execution lifecycle for an Agent, tracking all operations within the task.

**Attributes**:
- `id`: Unique identifier (UUID)
- `agent_id`: Agent executing the run
- `operator_id`: Operator who owns the agent
- `status`: Current state (created, running, waiting, completed, failed, cancelled, timed_out)
- `started_at`: Start timestamp
- `ended_at`: End timestamp (if completed/failed/cancelled)
- `operations`: List of operations performed
- `evidence_generated`: Evidence summary from this run
- `error`: Error message (if failed)

**Relationships**:
- Belongs to: Agent
- Owned by: Operator
- Contains: Operations
- Generates: Evidence

**Usage Context**: The unit of agent auditability. Every agent operation is logged within an Agent Run context.

### Workspace (工作空间)
**Definition**: An isolation boundary for resources, enabling multi-tenant or team-based organization.

**Attributes**:
- `id`: Unique identifier (UUID)
- `name`: Workspace name
- `description`: Workspace description
- `retention_policy`: Data retention settings
- `audit_settings`: Audit logging configuration
- `created_at`: Creation timestamp
- `updated_at`: Last update timestamp

**Relationships**:
- Contains: Cases, Payloads, Interactions
- Constrains: Agents (optional)
- Managed by: Users

**Usage Context**: Multi-tenant isolation or team-based organization. Resources are scoped to workspaces.

### Audit Event (审计事件)
**Definition**: A log entry recording a specific operation for compliance and security auditing.

**Attributes**:
- `id`: Unique identifier (UUID)
- `timestamp`: When the event occurred
- `agent_id`: Agent ID (if operation by agent)
- `agent_run_id`: Agent Run ID (if operation by agent)
- `operator_id`: Operator ID (human owner)
- `action`: Action performed (create_probe, wait_interaction, export_evidence, etc.)
- `resource`: Resource type and ID affected
- `risk_level`: Risk classification (low, medium, high, critical)
- `parameters`: Full operation parameters
- `result`: Operation result (success/failure)
- `error`: Error message (if failed)
- `ip_address`: Source IP address
- `user_agent`: User agent string

**Relationships**:
- Logs: Agent Run operations
- Logs: User operations
- References: Resources (Case, Payload, Interaction, etc.)

**Usage Context**: Complete audit trail for compliance. All agent and user operations are logged with full context.

## Concept Relationships

### Primary Flow
```
Operator/Agent
  └── creates ──> Case
        └── contains ──> Payload
              └── delivered as ──> Probe
                    └── triggers ──> Interaction
                          └── aggregated into ──> Evidence
```

### Agent Flow
```
Operator
  └── creates ──> Agent
        └── executes ──> Agent Run
              └── performs ──> Operations
                    └── logged in ──> Audit Event
```

### Workspace Flow
```
Workspace
  └── contains ──> Cases
        └── contains ──> Payloads
              └── captures ──> Interactions
                    └── aggregated into ──> Evidence
```

## Interface Mapping

### API Endpoints
- `/api/v2/cases` - Case CRUD
- `/api/v2/payloads` - Payload CRUD
- `/api/v2/interactions` - Interaction query and export
- `/api/v2/evidence` - Evidence generation and export
- `/api/v2/agents` - Agent management
- `/api/v2/agent-runs` - Agent Run tracking
- `/api/v2/audit` - Audit log query

### CLI Commands
- `godnslog create-case` - Create Case
- `godnslog create-payload` - Create Payload
- `godnslog wait-interaction` - Wait for Interaction
- `godnslog export-evidence` - Export Evidence
- `godnslog create-agent` - Create Agent
- `godnslog list-agent-runs` - List Agent Runs
- `godnslog audit-log` - Query Audit Log

### MCP Tools
- `create_oast_probe` - Create Probe (creates Case + Payload)
- `wait_for_interaction` - Wait for Interaction
- `export_evidence` - Export Evidence
- (Additional tools for Agent management - Operator only)

### Web UI Pages
- `/dashboard/cases` - Case management
- `/dashboard/payloads` - Payload studio
- `/dashboard/interactions` - Interaction timeline
- `/dashboard/evidence` - Evidence timeline and export
- `/dashboard/agents` - Agent management (Operator only)
- `/dashboard/agent-runs` - Agent Run tracking (Operator only)
- `/dashboard/audit` - Audit log viewer (Operator only)

## Terminology Rules

### Rule 1: Consistent Naming
- Use these exact terms in all documentation, API specs, and code comments
- Do not create synonyms (e.g., don't use "Engagement" as synonym for "Case")
- Use English terms in code, Chinese translations in Chinese documentation

### Rule 2: Entity vs. Action
- Entities are nouns: Case, Payload, Interaction, Evidence, Agent, Agent Run
- Actions are verbs: create, update, delete, export, wait
- API endpoints follow REST conventions: GET /cases, POST /cases, etc.

### Rule 3: Scope Boundaries
- Case: Highest-level grouping
- Payload: Unit of probe generation
- Probe: Payload in delivery context
- Interaction: Unit of capture
- Evidence: Aggregated output
- Agent: AI entity with permissions
- Agent Run: Single task lifecycle
- Workspace: Isolation boundary
- Audit Event: Compliance logging

### Rule 4: Attribution Chain
- Every Interaction must be attributed to a Payload
- Every Payload must belong to a Case
- Every Agent operation must be logged in an Agent Run
- Every Agent Run must belong to an Agent
- Every Agent must belong to an Operator

## Cross-Reference to Design Documents

This terminology unifies the following design documents:
- `docs/product-positioning.md` - Uses these terms for product narrative
- `docs/scanner-hub-adapter-design.md` - Uses these terms for adapter contracts
- `docs/agent-native-specification.md` - Uses these terms for agent governance
- `docs/unified-control-plane.md` - Uses these terms for UI/UX design

Any future documentation must reference this terminology document as the source of truth.

## Glossary

| Term | Chinese | Definition |
|------|---------|------------|
| Case | 案例 | Security engagement grouping related OAST activities |
| Payload | 载荷 | Generated OAST probe template with rendered values |
| Probe | 探针 | Payload in delivery context (target, method, timing) |
| Interaction | 交互 | Captured out-of-band event triggered by Probe |
| Evidence | 证据 | Aggregated, scored, explained summary of Interactions |
| Agent | 代理 | AI/LLM entity with scoped permissions |
| Agent Run | 代理运行 | Single task execution lifecycle for Agent |
| Workspace | 工作空间 | Isolation boundary for resources |
| Audit Event | 审计事件 | Log entry recording operation for compliance |

## Version History

- v1.0 (2026-05-17): Initial unified terminology for GODNSLOG 2.0
