# GODNSLOG 2.0 Sprint X Evidence Summary API Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an Agent-friendly evidence summary API that returns Case/Payload evidence, Scanner Run package metadata, summary hash, and next actions in one call.

**Architecture:** Reuse the existing on-demand `interaction.EvidenceService` and `scannerhub` persistence. Add a small `internal/evidencehub` service that resolves `case_id`, `payload_id`, or `scanner_run_id` into a single read-only evidence summary bundle. Do not persist Evidence, run scanners, or add a new UI page in this sprint.

**Tech Stack:** Go 1.25, XORM, Gin `/api/v2`, `internal/interaction`, `internal/scannerhub`, `internal/models`, Next.js TypeScript types.

---

## Sprint Boundary

### In Scope

- Add `POST /api/v2/evidence/summary`.
- Accept one of:
  - `case_id`;
  - `payload_id`;
  - `scanner_run_id`.
- Resolve `scanner_run_id` to its Case/Payload and include the Scanner Run.
- Return:
  - normalized `scope`;
  - structured `evidence`;
  - related `scanner_runs`;
  - `package_hashes`;
  - deterministic `summary_hash`;
  - `next_actions`;
  - `generated_at`.
- Add service and API tests.
- Add frontend API client/types for future UI/MCP use.
- Document the endpoint and verification results.

### Out of Scope

- No persistent Evidence table.
- No new top-level UI page.
- No scanner execution or scheduling.
- No LLM-generated summaries.
- No changes to MCP tools in this sprint.

## Verification Commands

```bash
GOCACHE=/tmp/gocache go test ./internal/evidencehub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/lib/api-client.ts src/types/index.ts
cd frontend-next && npm run build
git diff --check
```
