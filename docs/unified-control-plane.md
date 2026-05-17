# Unified Control Plane Planning

## Design Goal

Define a unified control plane information architecture that enables three user personas to achieve their workflows through Web, API, CLI, and MCP interfaces. The goal is not to display all capabilities, but to enable complete workflows for each user type.

## User Personas

### 1. Security Engineer
**Goals**:
- Quickly initiate OAST verification
- Filter and analyze evidence
- Export conclusions for stakeholders
- Collaborate with team members

**Primary Workflow**:
1. Create Case for engagement
2. Generate Payloads using templates
3. Deliver Payloads to targets
4. Monitor Interactions in real-time
5. Analyze Evidence timeline
6. Export report

**Key Needs**:
- Fast probe creation
- Real-time interaction monitoring
- Evidence filtering and clustering
- One-click report generation
- Team collaboration features

### 2. Automation Platform
**Goals**:
- Stable API/CLI integration
- Predictable behavior
- Standardized output formats
- Error handling and retry logic

**Primary Workflow**:
1. Authenticate with API key
2. Create Probe via API/CLI
3. Wait for interactions via polling or webhook
4. Export results in standard format
5. Process results in pipeline

**Key Needs**:
- RESTful API with clear contracts
- CLI tool for scripting
- Standard output formats (JSON, JSONL, SARIF)
- Webhook support for real-time events
- Comprehensive error codes and messages

### 3. AI Agent Operator
**Goals**:
- Safely grant permissions to agents
- Monitor agent activities
- Review agent run history
- Control risk exposure

**Primary Workflow**:
1. Create Agent with scoped permissions
2. Monitor Agent Runs in real-time
3. Review audit logs periodically
4. Adjust agent policies as needed
5. Revoke agent access if compromised

**Key Needs**:
- Agent management dashboard
- Agent Run timeline and replay
- Audit log query and filtering
- Risk-based alerting
- One-click revocation

## Information Architecture

### Primary Navigation Structure

```
GODNSLOG 2.0
├── Dashboard
│   └── Overview
├── OAST Core
│   ├── Cases
│   │   ├── Case List
│   │   ├── Case Detail
│   │   └── Case Settings
│   ├── Payloads
│   │   ├── Payload Studio
│   │   ├── Template Center
│   │   └── Payload Detail
│   └── Interactions
│       ├── Interaction Timeline
│       ├── Clustering View
│       └── Interaction Detail
├── Evidence
│   ├── Evidence Timeline
│   ├── Evidence Export
│   └── Evidence Reports
├── Monitor
│   ├── Canary Tokens
│   ├── Rebinding Lab
│   └── Workflow
├── Integrations
│   ├── Scanner Hub
│   ├── Webhooks
│   └── Notifications
├── Agent Operations
│   ├── Agents
│   ├── Agent Runs
│   └── Agent Analytics
└── System
    ├── Settings
    ├── Users
    ├── API Keys
    ├── Audit Log
    ├── Workspaces
    └── Docs
```

### Page-by-Page Specification

#### Dashboard / Overview
**Purpose**: High-level view of system activity and quick actions

**Content**:
- Statistics cards (Cases, Payloads, Interactions, Evidence)
- Recent Interactions timeline (last 24 hours)
- Active Cases summary
- Quick actions (New Case, New Payload, Create Agent)
- System health status

**User Personas**:
- Security Engineer: Monitor activity, quick access
- Automation Platform: System health check
- Agent Operator: Agent activity overview

#### Cases / Case List
**Purpose**: Manage security engagement cases

**Content**:
- Case table with columns: Title, Status, Created, Owner, Tags, Interaction Count
- Filters: Status, Owner, Tags, Date Range
- Search: By title or description
- Actions: Create Case, Bulk Actions (export, delete)
- Pagination

**User Personas**:
- Security Engineer: Primary workflow
- Automation Platform: API access only

#### Cases / Case Detail
**Purpose**: View and manage single case

**Content**:
- Case metadata (title, description, status, tags)
- Associated Payloads list
- Interaction timeline for this case
- Evidence summary
- Actions: Edit, Delete, Export Evidence, Share

**User Personas**:
- Security Engineer: Primary workflow

#### Payloads / Payload Studio
**Purpose**: Create and manage OAST payloads

**Content**:
- Payload creation wizard (3 steps: Template, Variables, Preview)
- Template selection with categories (SSRF, XXE, RCE, Blind SQLi, etc.)
- Variable input form
- Live preview of rendered payload
- Payload list with columns: Token, Template, Created, Status, Interaction Count
- Actions: Create Payload, Batch Generate, Revoke, Copy

**User Personas**:
- Security Engineer: Primary workflow
- Automation Platform: API/CLI access

#### Payloads / Template Center
**Purpose**: Browse and manage payload templates

**Content**:
- Template library with categories
- Template cards with: Name, Description, Category, Variables
- Template preview
- Custom template creation
- Template import/export

**User Personas**:
- Security Engineer: Browse and create templates
- Automation Platform: API access for template listing

#### Interactions / Interaction Timeline
**Purpose**: View captured interactions in chronological order

**Content**:
- Timeline view with interaction cards
- Filters: Protocol, Source IP, Token, Time Range
- Search: By token, IP, or pattern
- Clustering toggle (group by IP, token, pattern)
- Interaction detail drawer
- Actions: Export, Delete, Add to Evidence

**User Personas**:
- Security Engineer: Primary workflow
- Automation Platform: API access for filtering and export

#### Interactions / Clustering View
**Purpose**: Group interactions for pattern analysis

**Content**:
- Cluster groups (by IP, token, pattern)
- Cluster statistics (count, time range, protocols)
- Cluster drill-down to individual interactions
- Noise reduction controls
- Actions: Export cluster, Mark as noise, Add to Evidence

**User Personas**:
- Security Engineer: Pattern analysis

#### Evidence / Evidence Timeline
**Purpose**: View evidence chains for cases

**Content**:
- Evidence timeline by case
- Evidence strength indicators
- Confidence scores
- Explainability notes
- Actions: Export Evidence, Generate Report

**User Personas**:
- Security Engineer: Primary workflow
- Automation Platform: API access for export

#### Evidence / Evidence Export
**Purpose**: Export evidence in various formats

**Content**:
- Export format selection (Markdown, JSON, JSONL, SARIF, PDF)
- Export scope selection (Case, Time Range, Custom)
- Redaction options (sensitive data masking)
- Export history
- Actions: Export, Download, Share

**User Personas**:
- Security Engineer: Primary workflow
- Automation Platform: API access for export

#### Monitor / Canary Tokens
**Purpose**: Long-term monitoring with canary tokens

**Content**:
- Canary token list
- Token creation wizard (type, context, encoding)
- Token hit timeline
- Risk assessment
- Actions: Create Token, Revoke, View Hits

**User Personas**:
- Security Engineer: Long-term monitoring

#### Monitor / Rebinding Lab
**Purpose**: DNS rebinding testing

**Content**:
- Rebinding rule list
- Rule creation wizard (stages, targets, timing)
- Rebinding session tracking
- Session timeline
- Actions: Create Rule, Start Session, View Sessions

**User Personas**:
- Security Engineer: Rebinding testing

#### Monitor / Workflow
**Purpose**: Workflow automation rules

**Content**:
- Workflow rule list
- Rule builder (conditions, actions)
- Rule testing
- Rule execution history
- Actions: Create Rule, Enable/Disable, Test

**User Personas**:
- Security Engineer: Automation setup

#### Integrations / Scanner Hub
**Purpose**: Scanner integration management

**Content**:
- Supported scanners list with maturity levels
- Scanner integration guides
- Adapter configuration
- Integration status
- Actions: View Guide, Configure Integration

**User Personas**:
- Security Engineer: Integration setup
- Automation Platform: Reference for API/CLI usage

#### Integrations / Webhooks
**Purpose**: Webhook configuration and management

**Content**:
- Webhook list
- Webhook creation wizard (URL, events, headers)
- Webhook delivery history
- Retry configuration
- Actions: Create Webhook, Test Delivery, View History

**User Personas**:
- Security Engineer: Webhook setup
- Automation Platform: Webhook target

#### Integrations / Notifications
**Purpose**: Notification channel configuration

**Content**:
- Notification channel list (Webhook, Enterprise WeChat, Feishu, DingTalk)
- Channel configuration wizard
- Notification rules
- Delivery history
- Actions: Add Channel, Configure Rules, View History

**User Personas**:
- Security Engineer: Notification setup

#### Agent Operations / Agents
**Purpose**: AI agent management

**Content**:
- Agent list with status
- Agent creation wizard (name, scopes, workspace, expiration)
- Agent API key display
- Agent statistics (runs, success rate)
- Actions: Create Agent, View API Key, Modify Scopes, Revoke

**User Personas**:
- Agent Operator: Primary workflow

#### Agent Operations / Agent Runs
**Purpose**: Monitor agent execution history

**Content**:
- Agent Run list with filters (Agent, Status, Time Range)
- Agent Run detail with timeline
- Operation log
- Evidence generated
- Actions: View Detail, Export Run, Cancel (if running)

**User Personas**:
- Agent Operator: Primary workflow

#### Agent Operations / Agent Analytics
**Purpose**: Agent usage analytics

**Content**:
- Agent Run frequency chart
- Success/failure rate
- Most used tools
- Risk level distribution
- Resource consumption
- Actions: Export Analytics

**User Personas**:
- Agent Operator: Monitoring and optimization

#### System / Settings
**Purpose**: System-wide configuration

**Content**:
- General settings (domain, retention policies)
- Security settings (password policy, session timeout)
- Notification settings (email, alerts)
- Feature flags
- Actions: Save Settings

**User Personas**:
- System Administrator

#### System / Users
**Purpose**: User management

**Content**:
- User list with roles
- User creation wizard
- Role assignment
- User activity
- Actions: Create User, Edit User, Delete User

**User Personas**:
- System Administrator

#### System / API Keys
**Purpose**: API key management

**Content**:
- API Key list with scopes and expiration
- API Key creation wizard (name, scopes, expiration, workspace)
- API Key usage statistics
- Actions: Create Key, View Key, Revoke Key, Regenerate Key

**User Personas**:
- System Administrator
- Agent Operator: For agent keys

#### System / Audit Log
**Purpose**: Audit log viewer

**Content**:
- Audit log table with filters (User, Agent, Action, Risk Level, Time Range)
- Audit log detail view
- Export audit log
- Actions: View Detail, Export, Search

**User Personas**:
- System Administrator
- Agent Operator: Monitor agent activities

#### System / Workspaces
**Purpose**: Multi-tenant workspace management

**Content**:
- Workspace list
- Workspace creation wizard
- Workspace settings (retention, quotas)
- Member management
- Actions: Create Workspace, Edit Settings, Manage Members

**User Personas**:
- System Administrator

#### System / Docs
**Purpose**: Documentation hub

**Content**:
- Documentation navigation
- Quick links to key docs
- API reference
- CLI usage guide
- MCP Server usage guide
- Scanner Hub integration guides

**User Personas**:
- All users

## Interface Coordination

### Web → API → CLI → MCP Coordination

**Consistent Entity Model**:
- All interfaces use the same entity definitions (Case, Payload, Interaction, Evidence, Agent Run)
- API is the source of truth for data structures
- CLI and MCP use API contracts
- Web UI consumes API responses

**Consistent Authentication**:
- Web: Cookie-based + localStorage token
- API: Bearer token (API key or JWT)
- CLI: API key or interactive login
- MCP: Agent-specific API key with scopes

**Consistent Error Handling**:
- All interfaces use the same error code system
- Error messages are consistent across interfaces
- HTTP status codes follow REST conventions
- CLI and MCP map HTTP errors to local error formats

**Consistent Output Formats**:
- JSON for structured data
- JSONL for streaming data
- Markdown for human-readable reports
- SARIF for tool integration

### Workflow Cross-Interface Support

**Security Engineer Workflow**:
- Web UI for manual operations
- CLI for scripting repetitive tasks
- API for custom integrations

**Automation Platform Workflow**:
- API for primary integration
- CLI for local testing and debugging
- Web UI for monitoring and configuration

**AI Agent Operator Workflow**:
- Web UI for agent management and monitoring
- MCP for agent operations
- API for custom agent implementations

## Implementation Priority

### Phase 1: Core OAST Workflow (Security Engineer)
1. Cases / Case List
2. Cases / Case Detail
3. Payloads / Payload Studio
4. Interactions / Interaction Timeline
5. Evidence / Evidence Timeline

### Phase 2: Automation Support (Automation Platform)
1. Complete API coverage for Phase 1
2. CLI tool for Phase 1 operations
3. Evidence / Evidence Export
4. Integrations / Webhooks

### Phase 3: Agent Governance (Agent Operator)
1. Agent Operations / Agents
2. Agent Operations / Agent Runs
3. System / API Keys (agent-specific)
4. System / Audit Log

### Phase 4: Advanced Features
1. Payloads / Template Center
2. Interactions / Clustering View
3. Monitor / Canary Tokens
4. Monitor / Rebinding Lab
5. Monitor / Workflow
6. Integrations / Scanner Hub
7. Integrations / Notifications
8. Agent Operations / Agent Analytics

### Phase 5: System Administration
1. Dashboard / Overview
2. System / Settings
3. System / Users
4. System / Workspaces
5. System / Docs

## Success Metrics

- **Workflow Completion Rate**: Percentage of users who complete full workflows
- **Interface Usage Distribution**: Web vs API vs CLI vs MCP usage
- **Cross-Interface Consistency**: Error rate due to interface inconsistencies
- **Time to Value**: Time from first interaction to first successful workflow
- **User Satisfaction**: Feedback from each persona

## Conclusion

The unified control plane provides a coherent information architecture that enables three distinct user personas to achieve their workflows through Web, API, CLI, and MCP interfaces. By defining clear page specifications, interface coordination patterns, and implementation priorities, GODNSLOG 2.0 ensures that all users can efficiently complete their OAST workflows while maintaining consistency across all access methods.
