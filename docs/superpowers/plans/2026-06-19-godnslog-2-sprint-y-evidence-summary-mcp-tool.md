# GODNSLOG 2.0 Sprint Y Evidence Summary MCP Tool Implementation Plan

**Goal:** Add an MCP tool `get_evidence_summary` that calls `POST /api/v2/evidence/summary` so AI Agents can retrieve a structured evidence bundle via MCP.

**Architecture:** Reuse the existing Sprint X `POST /api/v2/evidence/summary` API and the MCP server's existing permission/audit framework. Add a new tool handler in `internal/mcp/server.go`, register it in the tool list and permission map, and add tests.

**Tech Stack:** Go 1.25, `internal/mcp`, `internal/agentpolicy`, `internal/evidencehub`.

---

## Sprint Boundary

### In Scope

- Add `get_evidence_summary` MCP tool with permission metadata (scope: `agent:summarize_evidence`, risk: low).
- Tool handler calls `POST /api/v2/evidence/summary` with `case_id`, `payload_id`, or `scanner_run_id`.
- Support optional `agent_run_id` for operation logging.
- Add MCP server tests covering success, missing-parameter, and permission-denied paths.
- Update `docs/MCP_SERVER_USAGE.md` with the new tool documentation.
- Update `docs/verification.md` and create acceptance document.

### Out of Scope

- No new API endpoints.
- No frontend UI changes.
- No new database tables or migrations.
- No LLM summarization.

## Verification Commands

```bash
GOCACHE=/tmp/gocache go test ./internal/mcp
GOCACHE=/tmp/gocache go test ./...
git diff --check
```
