# Phase 1: Backend Realization — Implementation Plans

> **Spec:** `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §3

## Overview

Phase 1 is split into 5 independent sub-plans. Each can be executed in parallel (no cross-dependencies between sub-plans), except all should complete before Phase 2.

## Sub-Plans

| # | Plan File | Scope | Est. Tasks |
|---|-----------|-------|------------|
| 1.1 | `phase1-01-workflow-action-executors.md` | Outbound security module + 4 action executors (HTTP, DNS, Webhook, Notify) | 5 tasks |
| 1.2 | `phase1-02-rule-engine.md` | Fix CIDR matching + implement tag/report/noise actions | 4 tasks |
| 1.3 | `phase1-03-v2-api-gaps.md` | DNS record CRUD, xip query, user management CRUD | 5 tasks |
| 1.4 | `phase1-04-apikey-security.md` | Bcrypt key_hash migration, auto-migration on auth, key rotation | 3 tasks |
| 1.5 | `phase1-05-data-model-migration.md` | Idempotent migration script with dry-run, batch import | 3 tasks |

## Execution Order

Recommended sequence (can be parallelized with subagent-driven-development):

1. **1.1** (workflow) — no dependencies
2. **1.2** (rule engine) — no dependencies
3. **1.3** (v2 API) — no dependencies
4. **1.4** (apikey security) — no dependencies
5. **1.5** (migration) — no dependencies

All 5 sub-plans are independent and can be executed in parallel.

## Phase 1 Acceptance Criteria (from spec §3.6)

- [ ] All 4 workflow action executors implemented with tests
- [ ] Rule engine CIDR matching uses `net.ParseCIDR`
- [ ] Rule engine tag/report/noise actions implemented
- [ ] v2 API exposes DNS record CRUD
- [ ] v2 API exposes xip encoding query
- [ ] v2 API exposes user management CRUD
- [ ] Migration script exists and is tested
- [ ] APIKey stored as bcrypt hash
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes with new tests

## Final Verification

After all sub-plans complete:

```bash
go build ./...
go test ./internal/workflow/ ./internal/rule/ ./internal/auth/ ./internal/interaction/ ./migration/ ./server/ -v
```
