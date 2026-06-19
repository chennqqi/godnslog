# GODNSLOG 2.0 Sprint BB Agent Run MCP Tools Acceptance

**Date:** 2026-06-19  
**Sprint:** BB — Agent Run MCP Tools

## Scope

Add `list_agent_runs` and `get_agent_run` MCP tools to allow AI Agents to query their run history and details via the MCP.

### In Scope

- Register `list_agent_runs` and `get_agent_run` tools in `server.go`
- Implement `listAgentRuns` handler calling `GET /api/v2/agent-runs` with optional filters (status, agent_type, page, page_size)
- Implement `getAgentRun` handler calling `GET /api/v2/agent-runs/{id}`
- Add 4 test cases: list success, get success, missing id, permission denied
- Update `MCP_SERVER_USAGE.md` with tool documentation

### Out of Scope

- No backend API changes (endpoints already exist from Sprint U)
- No frontend changes
- Permission metadata already registered in Sprint V

## Verification Steps

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp -v -run 'TestListAgentRuns|TestGetAgentRun'
# PASS — 4 tests
```

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp -v -run 'TestToolPermissions'
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS — all packages ok
```

```bash
git diff --check
# PASS
```

## Findings

- Permission metadata (scope: `agent:read_runs`, risk: low) was already registered in `permissions.go` from Sprint V.
- Both handlers follow the established pattern: check permission → validate params → call API → extract data → return ToolResult.
- Documentation updated with tool descriptions, parameters, and examples.

## Decision

Accepted. Sprint BB provides `list_agent_runs` and `get_agent_run` MCP tools, completing the agent run visibility surface for AI Agents.
