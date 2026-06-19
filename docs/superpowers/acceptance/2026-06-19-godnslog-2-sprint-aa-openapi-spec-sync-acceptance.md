# GODNSLOG 2.0 Sprint AA OpenAPI Spec Sync Acceptance

**Date:** 2026-06-19  
**Sprint:** AA — OpenAPI Spec Sync

## Scope

Sync `docs/openapi.yaml` with all v2 API endpoints introduced in Sprints U–Y.

### In Scope

- Add paths for `/evidence/generate`, `/evidence/summary`
- Add paths for `/agent-policy/scopes`
- Add paths for `/scanner-hub/adapters`, `/scanner-runs` (CRUD + status)
- Add paths for all `/agent-runs` endpoints (list, create, get, review, status, operations, followups, review-decision, review-export, review-delivery, review-deliveries, review-queue, review-package-trace)
- Add 20+ new schemas for the above endpoints
- Add missing tags (agent-policy, scanner-hub, scanner-runs, agent-runs)

### Out of Scope

- No backend code changes
- No frontend changes
- No new tests (existing OpenAPI tests validate structure)

## Verification Steps

```bash
GOCACHE=/tmp/gocache go test ./docs -v
# PASS — 7 tests
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

- All 7 existing OpenAPI tests pass with the updated spec.
- New paths and schemas follow the existing OpenAPI 3.0.3 convention.
- Tags section updated to include agent-policy, scanner-hub, scanner-runs, and agent-runs.

## Decision

Accepted. Sprint AA brings the OpenAPI spec in sync with all implemented v2 API endpoints from Sprints U through Y.
