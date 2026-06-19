# GODNSLOG 2.0 Sprint HH Data Retention & Archive Policy Acceptance

**Date:** 2026-06-19  
**Sprint:** HH — Data Retention & Archive Policy

## Scope

Implement actual database operations for the data retention and archive policy system, replacing stub methods with functional xorm-based logic.

### In Scope

- Implement `retainInteractions`, `retainCases`, and `retainPayloads` with actual xorm count and delete operations
- Add `InteractionRecord`, `CaseRecord`, and `PayloadRecord` lightweight table-mapping types
- Add comprehensive service tests with mock store and xorm-based integration test

### Out of Scope

- No API endpoint changes
- No frontend changes
- No scheduled job execution (manual `RunPolicy` only)

## Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/retention -v
# PASS — 10 tests
```

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS — all packages ok
```

```bash
git diff --check
# PASS
```

## Decision

Accepted. Sprint HH implements the data retention and archive policy system with actual database operations for interaction, case, and payload retention, verified by xorm-based integration testing.
