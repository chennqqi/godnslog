# Sprint T Acceptance: Package Hash Trace Lookup

**Date**: 2026-06-10
**Sprint**: T - Review Package Hash Trace Lookup
**Status**: Not Accepted

## Summary

Sprint T is not accepted.

The implementation adds the expected high-level pieces: trace response models, `TraceReviewPackageByHash`, `GET /api/v2/agent-runs/review-package-trace`, Audit page trace UI, frontend types/client, and a new `frontend-next/e2e/audit.spec.ts`.

However, the required verification fails:

- `server` package tests do not compile.
- The new Audit E2E suite fails before reaching the Sprint T trace assertions.
- The click-through implementation from an Audit table Package Hash is likely broken because it sets React state and immediately reads the old state.
- `docs/verification.md` originally claimed incomplete/pass-like Sprint T results; it has been corrected with the actual failure.

## Blocking Findings

### 1. Server tests do not compile

Command:

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

Result: FAIL.

Observed compile errors include:

- `undefined: setupTestServer`
- `undefined: generateTestToken`
- `unknown field Password in struct literal of type models.TblUser`
- `undefined: models.AgentRun`
- `undefined: models.AgentRunStatusCompleted`
- `undefined: models.AgentOperation`
- `undefined: models.AuditLog`
- `undefined: models.AuditDetails`
- `undefined: models.AgentRunReviewPackageTraceResponse`

Root cause: the new `TestV2TraceReviewPackage` in `server/v2_api_test.go` uses helper names that do not exist in this test file and references the legacy `github.com/chennqqi/godnslog/models` package for 2.0 models that live under `internal/models` as `v2models`.

### 2. Audit E2E fails before testing Sprint T behavior

Command:

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/audit.spec.ts
```

Result: FAIL - 14 passed, 16 failed.

Every `audit.spec.ts` test fails in `beforeEach`:

```text
page.waitForURL: Timeout 10000ms exceeded.
waiting for navigation to "/dashboard"
```

The new `audit.spec.ts` uses a real login flow:

```ts
await page.goto('/login')
await page.fill('input[name="username"]', 'admin')
await page.fill('input[name="password"]', 'test123')
await page.click('button[type="submit"]')
await page.waitForURL('/dashboard', { timeout: 10000 })
```

This does not match the project's existing E2E pattern, where specs set `localStorage.token` and mock `/api/v2/auth/info`. As a result, the Sprint T trace assertions never run.

### 3. Trace click-through likely reads stale state

In `frontend-next/src/app/dashboard/audit/audit-page-content.tsx`, the Package Hash click handler does:

```ts
setPackageHashInput(packageHash)
setShowTrace(true)
handleTracePackage()
```

`handleTracePackage` reads `packageHashInput`, but React state has not necessarily updated yet. The click-through path can therefore trace the previous/empty hash. Sprint T required hash click-through from existing hash displays, so this needs a direct `handleTracePackage(packageHash)` style path or equivalent.

### 4. Route order needs explicit verification after tests compile

`server/v2_api.go` registers:

```go
agentRuns.GET("/:id", self.v2GetAgentRun)
...
agentRuns.GET("/review-package-trace", self.v2TraceReviewPackage)
agentRuns.GET("/review-queue", self.v2ListReviewQueue)
```

After server tests are fixed, the route must be proven to hit `v2TraceReviewPackage` rather than being swallowed by `/:id`. The current non-compiling test suite does not prove this.

## Partial Positives

- Frontend ESLint passed for the touched files.
- `npm run build` passed.
- Existing `agent-runs.spec.ts` tests passed as part of the combined Playwright run.
- The trace API handler and UI are present in code, but not yet accepted because tests fail.

## Verification Run

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

Result: FAIL - `server` test package does not compile.

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts e2e/audit.spec.ts
```

Result: PASS

```bash
cd frontend-next && npm run build
```

Result: PASS

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/audit.spec.ts
```

Result: FAIL - 14 passed, 16 failed.

## Required Fixes

1. Fix `server/v2_api_test.go` so `TestV2TraceReviewPackage` compiles using existing test setup patterns and `v2models` for 2.0 models.
2. Ensure server tests prove `GET /api/v2/agent-runs/review-package-trace?package_hash=...` reaches the trace handler and is not captured by `/:id`.
3. Add/repair backend tests for delivered, failed, and timeout trace summary counts.
4. Replace real-login `audit.spec.ts` setup with the repository's mocked-auth E2E pattern.
5. Make audit E2E assert the actual trace request URL includes the expected `package_hash` query.
6. Make audit E2E assert real trace data, not only generic labels or the first visible `1`.
7. Add sanitization E2E assertions that full webhook URL, header values, response body, Authorization, Cookie, and secrets are absent.
8. Fix Package Hash click-through so it traces the clicked hash, not stale state.
9. Add E2E proof for click-through from an existing Package Hash display.
10. Keep verification non-interactive: use `--reporter=line`, do not run `show-report`.

## Scope Check

No evidence was found that Sprint T intentionally added report center, package storage, signatures/PKI, saved connectors, retry queues, Scanner Hub expansion, workflow engine, or MCP auto-delivery.

## Decision

Sprint T remains blocked on compiling backend tests and functional E2E proof. The feature surface exists, but the required acceptance evidence is not valid yet.
