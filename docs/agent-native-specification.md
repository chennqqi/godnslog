# Agent-Native Product Specification

## Design Goal

Elevate MCP from "callable tools" to "governable agent operation surface". The core competitive advantage is not "MCP integration" but "making enterprises feel safe giving MCP to agents."

## Agent Identity Model

### Entity Relationships

```
Operator (Human)
  └── creates ──> Agent (AI/LLM)
        └── operates in ──> Workspace
              └── creates ──> Agent Run
                    └── generates ──> Case
                          └── contains ──> Payload
                                └── captures ──> Interaction
                                      └── aggregates to ──> Evidence
```

### Entity Definitions

**Operator**: Human user who owns and manages agents
- Has API keys with agent creation permissions
- Can view all Agent Runs created by their agents
- Can revoke agent permissions
- Is responsible for agent actions

**Agent**: AI/LLM entity with scoped permissions
- Identified by Agent ID (UUID)
- Has Agent-specific API key with `IsAgent` flag
- Has defined scopes (AgentScopes)
- Has risk tolerance level (default: medium)

**Workspace**: Isolation boundary for resources
- Contains Cases, Payloads, Interactions
- Has workspace-specific retention policies
- Has workspace-specific audit settings
- Agents can be assigned to specific workspaces

**Agent Run**: Single task execution lifecycle
- Created when agent initiates an OAST operation
- Tracked from creation to completion
- Contains all operations within the task
- Has start time, end time, and status
- Can be exported for analysis

## Agent Run Lifecycle

### States

1. **Created**: Agent Run initialized, awaiting first operation
2. **Running**: Agent is actively performing operations
3. **Waiting**: Agent is waiting for interactions (polling)
4. **Completed**: Agent Run finished successfully
5. **Failed**: Agent Run encountered an error
6. **Cancelled**: Agent Run was cancelled by operator
7. **Timed Out**: Agent Run exceeded time limit

### Lifecycle Events

```
[Created]
  ↓
[Running] ──→ [Waiting] ──→ [Running] ──→ ... ──→ [Completed]
  ↓                ↓
[Failed]         [Timed Out]
  ↓
[Cancelled] (by operator)
```

### Persistence

All Agent Run state changes are persisted to database:
- State transitions with timestamps
- Operation logs with parameters
- Interaction captures with attribution
- Final evidence summary

## Permission Scope System

### AgentScopes Definition

```go
type AgentScope string

const (
    AgentScopeCreateProbe      AgentScope = "create_probe"      // Create OAST probes
    AgentScopeWaitInteraction  AgentScope = "wait_interaction"  // Wait for interactions
    AgentScopeExportEvidence   AgentScope = "export_evidence"   // Export evidence
    AgentScopeDeletePayload    AgentScope = "delete_payload"    // Delete payloads
    AgentScopeModifyConfig     AgentScope = "modify_config"     // Modify sensitive configuration
)
```

### Scope Validation

**Create Probe**:
- Required scope: `create_probe`
- Risk level: Medium
- Audit: Full (title, template, target, variables)
- Default: Allowed

**Wait for Interaction**:
- Required scope: `wait_interaction`
- Risk level: Low
- Audit: Full (token, timeout, expected_count)
- Default: Allowed

**Export Evidence**:
- Required scope: `export_evidence`
- Risk level: Low
- Audit: Full (probe_id, format, include_raw)
- Default: Allowed

**Delete Payload**:
- Required scope: `delete_payload`
- Risk level: High
- Audit: Full (payload_id, reason)
- Default: Denied (requires explicit grant)

**Modify Configuration**:
- Required scope: `modify_config`
- Risk level: Critical
- Audit: Full (config_key, old_value, new_value)
- Default: Denied (requires explicit grant)

### Workspace-Based Scoping

Agents can be scoped to specific workspaces:
- `workspace:all` - Access to all workspaces
- `workspace:{id}` - Access to specific workspace only
- Default: `workspace:all` (can be restricted)

### Time-Based Scoping

Agent API keys have expiration:
- Default: 24 hours
- Maximum: 30 days
- Renewable by operator

## Risk Classification

### Risk Levels

**Low Risk**:
- Read operations (list, get)
- Wait for interaction (passive monitoring)
- Export evidence (data retrieval)

**Medium Risk**:
- Create probe (resource creation)
- Update payload (resource modification)

**High Risk**:
- Delete payload (resource destruction)
- Revoke token (access revocation)

**Critical Risk**:
- Modify configuration (system-wide changes)
- Delete case (engagement destruction)
- Modify workspace settings (tenant-level changes)

### Risk-Based Controls

**Low Risk Operations**:
- Always allowed with appropriate scope
- Standard audit logging
- No additional approval required

**Medium Risk Operations**:
- Allowed with appropriate scope
- Enhanced audit logging
- Rate limiting applies

**High Risk Operations**:
- Requires explicit scope grant
- Critical audit logging
- Operator notification on execution
- Rate limiting applies (stricter)

**Critical Risk Operations**:
- Requires explicit scope grant + operator approval
- Critical audit logging with alerting
- Operator notification before execution
- Strict rate limiting
- May require multi-factor approval

### Default Agent Policy

**Default Allowed**:
- `create_probe`
- `wait_interaction`
- `export_evidence`

**Default Denied**:
- `delete_payload`
- `modify_config`

Operators can override defaults by granting explicit scopes.

## Audit Standards

### Audit Event Structure

```go
type AuditEvent struct {
    ID          string                 `json:"id"`
    Timestamp   time.Time              `json:"timestamp"`
    AgentID     string                 `json:"agent_id"`
    AgentRunID  string                 `json:"agent_run_id"`
    OperatorID  string                 `json:"operator_id"`
    Action      string                 `json:"action"`
    Resource    string                 `json:"resource"`
    RiskLevel   string                 `json:"risk_level"`
    Parameters  map[string]interface{} `json:"parameters"`
    Result      string                 `json:"result"`
    Error       string                 `json:"error,omitempty"`
    IPAddress   string                 `json:"ip_address"`
    UserAgent   string                 `json:"user_agent"`
}
```

### Mandatory Audit Fields

**Who** (Identity):
- Agent ID
- Agent Run ID
- Operator ID (owner of agent)

**What** (Action):
- Action type (create_probe, wait_interaction, etc.)
- Resource type (case, payload, interaction)
- Resource ID

**When** (Time):
- Timestamp (UTC)
- Agent Run start time
- Operation duration

**Where** (Context):
- Workspace ID
- Target (if applicable)
- Case ID (if applicable)

**Why** (Parameters):
- Full operation parameters
- Variable values
- Configuration settings

**How** (Result):
- Success/Failure
- Error details (if failed)
- Evidence generated (if applicable)

### Audit Retention

- Low risk events: 90 days
- Medium risk events: 180 days
- High risk events: 365 days
- Critical risk events: 730 days (2 years)

### Audit Query

Operators can query audit logs by:
- Agent ID
- Agent Run ID
- Operator ID
- Time range
- Risk level
- Action type
- Resource type

### Audit Alerting

**Real-time Alerts** (Critical Risk):
- Configuration modification
- Case deletion
- Workspace settings change

**Daily Digest** (High Risk):
- Payload deletion
- Token revocation
- Scope modification

**Weekly Summary** (Medium Risk):
- Agent Run statistics
- Probe creation count
- Evidence export count

## MCP Tool Governance

### Tool Registration

All MCP tools must be registered with:
- Tool name and version
- Required scopes
- Risk level
- Parameter schema
- Audit requirements

### Tool Invocation Validation

Before executing any MCP tool:
1. Validate agent API key (exists, not expired, not revoked)
2. Validate agent has required scopes
3. Validate risk level against agent policy
4. Validate workspace access (if workspace-scoped)
5. Check rate limits
6. Log audit event (before execution)
7. Execute tool
8. Log audit event result (after execution)

### Tool Result Sanitization

Tool results are sanitized based on:
- Agent scope (what agent is allowed to see)
- Risk level (what data to include)
- Operator preferences (what to redact)

### Tool Timeout

All MCP tools have timeout limits:
- Low risk tools: 30 seconds
- Medium risk tools: 60 seconds
- High risk tools: 120 seconds
- Critical risk tools: 300 seconds

Timeout triggers:
- Operation cancellation
- Audit event with timeout status
- Agent notification (if configured)

## Agent-Specific Features

### Agent Run Dashboard

Operators can view:
- All Agent Runs by their agents
- Agent Run status and timeline
- Operations performed within each run
- Evidence generated
- Audit logs for each operation

### Agent Run Export

Operators can export Agent Runs in:
- JSON (full detail)
- Markdown (summary report)
- PDF (formatted report)

Export includes:
- Agent metadata
- Agent Run timeline
- All operations with parameters
- Evidence generated
- Audit log entries

### Agent Revocation

Operators can:
- Revoke agent API key (immediate)
- Cancel active Agent Runs
- Delete agent (and all associated runs)
- Modify agent scopes (add/remove)

Revocation triggers:
- Immediate audit event
- Agent notification (if configured)
- Cleanup of in-progress operations

### Agent Analytics

Operators can view:
- Agent Run frequency
- Success/failure rate
- Most used tools
- Risk level distribution
- Resource consumption

## Security Best Practices

### For Operators

1. **Principle of Least Privilege**: Grant only necessary scopes
2. **Regular Key Rotation**: Rotate agent API keys periodically
3. **Monitor Audit Logs**: Review audit logs regularly
4. **Set Time Limits**: Use appropriate key expiration
5. **Workspace Isolation**: Use workspaces to isolate agent activities
6. **Rate Limiting**: Configure appropriate rate limits
7. **Alert Configuration**: Set up alerts for critical operations

### For Agent Developers

1. **Scope Declaration**: Declare required scopes in tool registration
2. **Error Handling**: Handle errors gracefully and log appropriately
3. **Timeout Awareness**: Design tools to work within timeout limits
4. **Parameter Validation**: Validate all parameters before execution
5. **Result Sanitization**: Sanitize results based on agent scope
6. **Audit Compliance**: Ensure all operations are auditable

## Implementation Checklist

- [ ] Agent identity model implemented (Agent, Operator, Workspace, Agent Run)
- [ ] Agent Run lifecycle management implemented
- [ ] AgentScopes defined and validation implemented
- [ ] Risk classification system implemented
- [ ] Audit event structure defined and logging implemented
- [ ] MCP tool registration and validation implemented
- [ ] Tool timeout system implemented
- [ ] Agent Run dashboard implemented
- [ ] Agent Run export implemented
- [ ] Agent revocation implemented
- [ ] Agent analytics implemented
- [ ] Security best practices documented

## Conclusion

The Agent-Native specification transforms MCP from a simple API into a governed, auditable, and enterprise-ready agent operation surface. By defining clear identity models, permission scopes, risk classifications, and audit standards, GODNSLOG 2.0 enables enterprises to safely grant OAST capabilities to AI agents while maintaining full control and visibility.
