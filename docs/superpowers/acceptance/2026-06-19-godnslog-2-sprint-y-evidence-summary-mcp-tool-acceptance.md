# GODNSLOG 2.0 Sprint Y Evidence Summary MCP Tool Acceptance

**Date:** 2026-06-19  
**Sprint:** Y — Evidence Summary MCP Tool

## Scope

Add an MCP tool `get_evidence_summary` that calls `POST /api/v2/evidence/summary` so AI Agents can retrieve a structured evidence bundle via MCP.

### In Scope

- `get_evidence_summary` MCP tool with permission metadata (scope: `agent:summarize_evidence`, risk: low)
- Tool handler calls `POST /api/v2/evidence/summary` with `case_id`, `payload_id`, or `scanner_run_id`
- Optional `agent_run_id` for operation logging
- MCP server tests covering success, scanner_run_id, missing params, and permission denied paths
- Documentation updates: `docs/MCP_SERVER_USAGE.md`, `docs/verification.md`

### Out of Scope

- No new API endpoints
- No frontend UI changes
- No new database tables or migrations
- No LLM summarization

## Verification Steps

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp -v -run 'TestGetEvidenceSummary'
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

- The tool reuses the existing Sprint X `POST /api/v2/evidence/summary` API endpoint.
- Permission metadata is aligned with the shared agent policy catalog (`agent:summarize_evidence`, risk: low).
- The tool extracts and returns the `data` field from the API response for Agent convenience.
- Optional `agent_run_id` triggers operation logging via `POST /api/v2/agent-runs/:id/operations`.

## Decision

Accepted. Sprint Y adds the `get_evidence_summary` MCP tool, making the Sprint X Evidence Summary API directly accessible to AI Agents via MCP with proper permission gating and audit logging.
