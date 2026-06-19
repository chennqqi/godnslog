# GODNSLOG 2.0 Sprint V Acceptance

Date: 2026-06-19

## Scope

Sprint V validates the Agent Risk Controls increment:

- Shared Agent policy catalog in `internal/agentpolicy`.
- Authenticated policy endpoint `GET /api/v2/agent-policy/scopes`.
- Agent API Key creation defaults to safe Agent scopes when none are supplied.
- High-risk scopes remain explicit: `agent:revoke_token`, `agent:delete_payload`, `agent:modify_config`.
- MCP permission metadata uses Agent scopes, including `create_case` and `create_payload` through `agent:create_probe`.
- API Key UI loads the policy catalog and separates default scopes from high-risk scopes.
- Documentation records the policy endpoint, UI behavior, MCP mapping, and verification evidence.

## Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/agentpolicy ./internal/models
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./server -run 'TestV2AgentPolicyScopes|TestV2CreateAgentAPIKeyUsesSafeDefaults'
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp -run 'TestToolPermissionsUseAgentPolicyCatalog|TestPermissionDeniedAuditLog|TestPermissionGateMissingScope|TestPermissionGateExceedsRiskTolerance'
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./internal/agentpolicy ./internal/auth ./internal/mcp ./server
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS
```

```bash
cd frontend-next && npx eslint src/app/dashboard/apikeys/page.tsx src/lib/api-client.ts src/types/index.ts e2e/apikeys.spec.ts
# PASS
```

```bash
cd frontend-next && npm run build
# PASS
```

```bash
cd frontend-next && npx playwright test --reporter=line e2e/apikeys.spec.ts
# PASS - 5 passed
```

```bash
git diff --check
# PASS
```

## Findings

- The shared catalog closes the previous split between model constants, MCP tool permissions, and UI scope choices.
- The API Key UI has a fallback policy for endpoint failure, but the canonical source is now `/api/v2/agent-policy/scopes`.
- `agent:delete_payload` and `agent:modify_config` are policy entries only; no destructive Agent implementation was added in this sprint.
- E2E verification used a non-interactive Playwright reporter and did not start the HTML report server.

## Decision

Accepted. Sprint V meets the planned risk-control scope and should be the stopping point for this Codex run. Remaining work should be delegated to another AI Agent using the handoff prompt in the final response.
