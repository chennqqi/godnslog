# Test Coverage Improvement Design

- **Date**: 2026-06-21
- **Goal**: Push GODNSLOG 2.0 to production-ready through comprehensive testing

## Current State

### Backend (Go)
- 142 source files, 47 test files, 31% total coverage
- 7 packages with 0% coverage: notification, openapi, rebinding, workspace, cmd/mcp-server, cmd/migrate, server/docs
- Key 0% modules: marketplace/service+store, canary/, rebinding/, ha/service+store, notification/, workspace/
- Low coverage: server/v2_api.go (16.9%), server/webui.go (13.1%), rule/engine.go (11.1%)

### Frontend (Next.js)
- 0 unit tests (no Jest/Vitest config)
- 15 E2E files, 117 tests, covering 15 pages
- 6 pages without E2E: docs, evidence-summary, export, listeners, retention, users
- Weak E2E coverage: marketplace (2), canary (2), rebinding (2), settings (2), workflow (3)

## Approach: Comprehensive (Option C)

All layers in parallel, committed in rounds.

### Round 1: Backend 0% core modules + Frontend missing E2E

**Backend unit tests** (target: 0% → 60%+):
- `internal/marketplace/service_test.go` + `store_test.go`
- `internal/rebinding/service_test.go` + `store_test.go`
- `internal/canary/detector_test.go` + `service_test.go`
- `internal/notification/service_test.go`
- `internal/workspace/store_test.go` + `model_test.go`

**Frontend E2E** (6 missing pages):
- `e2e/users.spec.ts`
- `e2e/listeners.spec.ts`
- `e2e/retention.spec.ts`
- `e2e/evidence-summary.spec.ts`
- `e2e/export.spec.ts`
- `e2e/docs.spec.ts`

### Round 2: Backend API integration tests + Frontend E2E hardening

**Backend integration tests** (extend `server/v2_api_test.go`):
- Auth: login, APIKey CRUD, permission check
- Cases: CRUD, stats, detail
- Payloads: create, list, detail, revoke
- Interactions: list, detail, stats
- Evidence: generate, summary
- Scanner-runs: create, list, detail, status update
- Marketplace: plugin install, list, templates

**Frontend E2E hardening** (weak pages):
- marketplace: plugin install flow, template creation
- canary: token creation, hit viewing
- settings: language switch, save
- workflow: rule create, edit, delete

### Round 3: Edge cases + regression + final verification

- Backend: rule/engine.go, interaction/noise_filter.go, listener/handler.go low-coverage modules
- Frontend: error states, empty states, permission boundaries E2E
- Full `go test ./...` + `npm run lint` + `CI=1 npx playwright test`
- Update `docs/verification.md`

## Commit Strategy

Each round: implement → verify → commit → continue next round.
