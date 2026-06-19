# GODNSLOG 2.0 Sprint V Agent Risk Controls Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn the existing Agent API Key and MCP permission work into a visible, testable Agent Risk Control surface that makes scoped agent operations safer and easier to audit.

**Architecture:** Keep the implementation inside the existing API Key, MCP, Audit, and Agent Run boundaries. Add a shared policy catalog for agent scopes and risk levels, expose it through `/api/v2/agent-policy`, use it to validate API keys and MCP tools, and surface the same policy in the API Key UI and audit/agent-run review flows. Do not create a general-purpose agent management platform, workflow engine, or approval system in this sprint.

**Tech Stack:** Go 1.25, Gin, XORM, `internal/models`, `internal/auth`, `internal/mcp`, `/api/v2`, Next.js 16, TypeScript, shadcn/ui/Tailwind, Playwright.

---

## Sprint Boundary

### In Scope

- Define a single Agent policy catalog with:
  - scope ID;
  - display name;
  - risk level;
  - default allowed/denied state;
  - MCP tool mappings;
  - whether the scope is high-risk and requires explicit selection.
- Add `GET /api/v2/agent-policy/scopes`.
- Validate Agent API Key creation and update against the catalog.
- Ensure default Agent keys include only safe default scopes:
  - `agent:create_probe`
  - `agent:wait_interaction`
  - `agent:read_interactions`
  - `agent:summarize_evidence`
  - `agent:export_report`
  - `agent:read_runs`
- Keep high-risk scopes explicit:
  - `agent:revoke_token`
  - `agent:delete_payload`
  - `agent:modify_config`
- Make MCP tool permission metadata derive from the same catalog.
- Add audit proof for denied MCP operations with `risk_level`, `required_scope`, `reason`, `api_key_id`, and `tool_name`.
- Update the API Key UI so Agent scopes are grouped by risk and high-risk scopes are visually separated.
- Add E2E coverage proving:
  - safe Agent key defaults;
  - high-risk scope visibility;
  - risk tolerance selector;
  - no hidden HTML report server flow.

### Out of Scope

- No new Agent entity CRUD.
- No multi-step approval workflow.
- No MFA or human approval gate.
- No workspace UI redesign.
- No new MCP tools beyond wiring existing scope/risk metadata.
- No destructive agent action implementation for `delete_payload` or `modify_config`.

## File Structure Map

- Create: `internal/agentpolicy/policy.go` for the shared scope/risk catalog.
- Create: `internal/agentpolicy/policy_test.go` for catalog and compatibility tests.
- Modify: `internal/models/apikey.go` to delegate agent-scope validation to catalog-compatible constants or helpers.
- Modify: `internal/auth/service.go` to enforce defaults, maximum expiration, and explicit high-risk selection semantics.
- Modify: `internal/mcp/permissions.go` to build tool permissions from the shared policy catalog.
- Modify: `internal/mcp/server_test.go` to prove denied operations are audited with risk context.
- Modify: `server/v2_api.go` to register and implement `GET /api/v2/agent-policy/scopes`.
- Modify: `server/v2_api_test.go` for endpoint and API key validation coverage.
- Modify: `frontend-next/src/types/index.ts` to add Agent policy types.
- Modify: `frontend-next/src/lib/api-client.ts` to add `agentPolicyApi.listScopes()`.
- Modify: `frontend-next/src/app/dashboard/apikeys/page.tsx` to load policy metadata and render risk-grouped Agent scopes.
- Modify: `frontend-next/e2e/apikeys.spec.ts` or create it if missing.
- Modify: `docs/MCP_SERVER_USAGE.md` and `docs/agent-native-specification.md` with implementation status and operator-facing behavior.
- Modify: `docs/verification.md` with Sprint V commands and results.

## Policy Contract

Backend response:

```json
{
  "items": [
    {
      "scope": "agent:create_probe",
      "name": "Create OAST Probe",
      "risk_level": "medium",
      "default_allowed": true,
      "high_risk": false,
      "tool_names": ["create_oast_probe"],
      "description": "Create a Case and Payload for OAST validation."
    }
  ],
  "default_scopes": [
    "agent:create_probe",
    "agent:wait_interaction",
    "agent:read_interactions",
    "agent:summarize_evidence",
    "agent:export_report",
    "agent:read_runs"
  ],
  "high_risk_scopes": [
    "agent:revoke_token",
    "agent:delete_payload",
    "agent:modify_config"
  ]
}
```

## Task 1: Shared Agent Policy Catalog

**Files:**
- Create: `internal/agentpolicy/policy.go`
- Create: `internal/agentpolicy/policy_test.go`
- Modify: `internal/models/apikey.go`

- [ ] **Step 1: Write failing catalog tests**

Add tests that assert:

```go
func TestCatalogContainsExpectedAgentScopes(t *testing.T) {
	catalog := agentpolicy.ListScopes()
	require.Contains(t, catalog.ByScope(), "agent:create_probe")
	require.Contains(t, catalog.ByScope(), "agent:wait_interaction")
	require.Contains(t, catalog.ByScope(), "agent:export_report")
	require.Contains(t, catalog.ByScope(), "agent:revoke_token")
	require.True(t, catalog.ByScope()["agent:revoke_token"].HighRisk)
	require.Equal(t, "high", catalog.ByScope()["agent:revoke_token"].RiskLevel)
	require.False(t, catalog.ByScope()["agent:delete_payload"].DefaultAllowed)
}
```

- [ ] **Step 2: Run failing tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/agentpolicy
```

Expected: FAIL because the package does not exist.

- [ ] **Step 3: Implement catalog**

Create `internal/agentpolicy/policy.go` with:

```go
type ScopePolicy struct {
	Scope          string   `json:"scope"`
	Name           string   `json:"name"`
	RiskLevel      string   `json:"risk_level"`
	DefaultAllowed bool     `json:"default_allowed"`
	HighRisk       bool     `json:"high_risk"`
	ToolNames      []string `json:"tool_names"`
	Description    string   `json:"description"`
}

type ScopeCatalog struct {
	Items          []ScopePolicy `json:"items"`
	DefaultScopes  []string      `json:"default_scopes"`
	HighRiskScopes []string      `json:"high_risk_scopes"`
}
```

Include all existing Agent scopes and high-risk scopes. Add helper methods:

```go
func ListScopes() ScopeCatalog
func ValidateScopes(scopes []string) bool
func DefaultScopes() []string
func HighRiskScopes() []string
func PolicyForScope(scope string) (ScopePolicy, bool)
```

- [ ] **Step 4: Wire model validation**

Change `internal/models/apikey.go` so `ValidateAgentScopes` delegates to `agentpolicy.ValidateScopes`, while keeping exported `AgentScopes` and `HighRiskAgentScopes` compatible for existing callers.

- [ ] **Step 5: Verify and commit**

```bash
GOCACHE=/tmp/gocache go test ./internal/agentpolicy ./internal/models
git add internal/agentpolicy internal/models/apikey.go
git commit -m "feat: add shared agent policy catalog"
```

## Task 2: API Key Enforcement and Policy Endpoint

**Files:**
- Modify: `internal/auth/service.go`
- Modify: `server/v2_api.go`
- Modify: `server/v2_api_test.go`
- Modify: `docs/openapi.yaml` if the OpenAPI source is manually maintained.

- [ ] **Step 1: Add failing API tests**

Add server tests for:

- `GET /api/v2/agent-policy/scopes` requires auth and returns default/high-risk scopes.
- Creating an Agent key without scopes uses `agentpolicy.DefaultScopes()`.
- Creating an Agent key with `agent:delete_payload` is allowed only when explicitly supplied.
- Creating an Agent key with unknown scope returns 400.
- Agent key expiration is still capped at 30 days.

- [ ] **Step 2: Run failing API tests**

```bash
GOCACHE=/tmp/gocache go test ./server -run 'TestV2AgentPolicy|TestV2CreateAPIKeyAgent'
```

Expected: FAIL until endpoint/default enforcement is implemented.

- [ ] **Step 3: Implement endpoint**

Register:

```go
agentPolicy := v2.Group("/agent-policy", self.authHandler)
agentPolicy.GET("/scopes", self.v2ListAgentPolicyScopes)
```

Handler response must be a normal v2 envelope:

```go
c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": agentpolicy.ListScopes()})
```

- [ ] **Step 4: Enforce defaults and validation in auth service**

In `internal/auth/service.go`, for `req.IsAgent`:

- if `len(req.Scopes) == 0`, set scopes to `agentpolicy.DefaultScopes()`;
- reject unknown scopes using `agentpolicy.ValidateScopes`;
- keep default risk tolerance at `medium`;
- keep maximum expiration at 30 days.

- [ ] **Step 5: Verify and commit**

```bash
GOCACHE=/tmp/gocache go test ./internal/auth ./server
git add internal/auth/service.go server/v2_api.go server/v2_api_test.go
git commit -m "feat: expose and enforce agent policy scopes"
```

## Task 3: MCP Permission Metadata and Denied Audit Proof

**Files:**
- Modify: `internal/mcp/permissions.go`
- Modify: `internal/mcp/server.go`
- Modify: `internal/mcp/server_test.go`

- [ ] **Step 1: Add failing MCP tests**

Add tests proving:

- every MCP tool permission references a scope in `agentpolicy.ListScopes`;
- denied missing-scope operation posts `agent_permission.denied` audit;
- denied risk-tolerance operation posts `risk_level`, `required_scope`, `reason`, `api_key_id`, and `tool_name`;
- non-agent admin key still passes `admin:all`.

- [ ] **Step 2: Run failing tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp -run 'TestToolPermissions|TestPermissionDeniedAudit'
```

Expected: FAIL until permissions use the shared catalog and audit test handlers see the expected payload.

- [ ] **Step 3: Build permissions from catalog**

Keep `ToolPermissions` available, but derive scope/risk from `agentpolicy.PolicyForScope` to prevent drift. Ensure `create_payload` is either added to `ToolPermissions` with `agent:create_probe` or removed from public MCP registration if it cannot be governed.

- [ ] **Step 4: Strengthen denied audit payload**

Ensure `writePermissionDeniedAudit` includes:

```json
{
  "action": "agent_permission.denied",
  "resource_type": "mcp_tool",
  "tool_name": "revoke_token",
  "required_scope": "agent:revoke_token",
  "risk_level": "high",
  "reason": "risk level exceeds tolerance",
  "api_key_id": "key-123",
  "key_prefix": "gdl_abc",
  "is_agent": true
}
```

- [ ] **Step 5: Verify and commit**

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp
git add internal/mcp/permissions.go internal/mcp/server.go internal/mcp/server_test.go
git commit -m "feat: bind mcp permissions to agent policy"
```

## Task 4: Agent Key UI Risk Grouping

**Files:**
- Modify: `frontend-next/src/types/index.ts`
- Modify: `frontend-next/src/lib/api-client.ts`
- Modify: `frontend-next/src/app/dashboard/apikeys/page.tsx`
- Test: `frontend-next/e2e/apikeys.spec.ts`

- [ ] **Step 1: Add failing E2E tests**

Create or extend `apikeys.spec.ts` to mock:

- `GET /api/v2/agent-policy/scopes`;
- `GET /api/v2/apikeys`;
- `POST /api/v2/apikeys`.

Assertions:

- switching to Agent Key shows safe scopes selected by default;
- high-risk scopes appear in a separate section labeled `高风险作用域`;
- selecting `agent:revoke_token` sends it in the create request;
- risk tolerance selector sends `risk_tolerance`.

- [ ] **Step 2: Run failing E2E**

```bash
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts
```

Expected: FAIL until the UI loads the policy endpoint.

- [ ] **Step 3: Add frontend types and API client**

Add:

```ts
export interface AgentScopePolicy {
  scope: string
  name: string
  risk_level: 'low' | 'medium' | 'high' | 'critical'
  default_allowed: boolean
  high_risk: boolean
  tool_names: string[]
  description: string
}

export interface AgentScopeCatalog {
  items: AgentScopePolicy[]
  default_scopes: string[]
  high_risk_scopes: string[]
}
```

Add `agentPolicyApi.listScopes()`.

- [ ] **Step 4: Update API Key UI**

Load the policy catalog on mount. When Agent Key is checked:

- select `catalog.default_scopes`;
- render default/low/medium scopes in one section;
- render high-risk scopes in a separate warning section;
- keep risk tolerance visible and required for Agent keys.

- [ ] **Step 5: Verify and commit**

```bash
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts
git add frontend-next/src/app/dashboard/apikeys/page.tsx frontend-next/src/lib/api-client.ts frontend-next/src/types/index.ts frontend-next/e2e/apikeys.spec.ts
git commit -m "feat: show agent key risk policy in ui"
```

## Task 5: Documentation, Full Verification, and Acceptance

**Files:**
- Modify: `docs/MCP_SERVER_USAGE.md`
- Modify: `docs/agent-native-specification.md`
- Modify: `docs/verification.md`
- Create: `docs/superpowers/acceptance/2026-06-19-godnslog-2-sprint-v-acceptance.md`

- [ ] **Step 1: Update docs**

Document:

- policy endpoint;
- safe default Agent scopes;
- high-risk explicit scope behavior;
- MCP denied audit payload;
- current non-goals.

- [ ] **Step 2: Run full verification**

```bash
GOCACHE=/tmp/gocache go test ./internal/agentpolicy ./internal/auth ./internal/mcp ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts
git diff --check
```

Expected: all pass.

- [ ] **Step 3: Confirm E2E command constraint**

Confirm no command used `npx playwright show-report`, `npm run test:e2e:ui`, or any flow that starts the Playwright HTML report server.

- [ ] **Step 4: Write acceptance**

Acceptance must explicitly state:

- whether the shared catalog removed scope/risk drift;
- whether API, MCP, UI, and docs all use the same policy;
- whether denied MCP operations are auditable;
- verification commands and results.

- [ ] **Step 5: Commit docs**

```bash
git add docs/MCP_SERVER_USAGE.md docs/agent-native-specification.md docs/verification.md docs/superpowers/acceptance/2026-06-19-godnslog-2-sprint-v-acceptance.md
git commit -m "docs: accept sprint v agent risk controls"
```

## Acceptance Checklist

- [ ] Agent policy endpoint returns the same scope/risk catalog used by backend validation and MCP permissions.
- [ ] Agent API Key defaults are safe and do not silently include high-risk scopes.
- [ ] High-risk scopes remain explicit and visible.
- [ ] MCP denied operations produce auditable risk context.
- [ ] API Key UI explains and groups scopes by risk without hiding dangerous capabilities.
- [ ] No Playwright HTML report server is triggered during verification.
- [ ] No scanner execution, workflow expansion, or general Agent management platform is added.
