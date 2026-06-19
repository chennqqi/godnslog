# Sprint T Acceptance: Package Hash Trace Lookup

**Date**: 2026-06-12
**Sprint**: T - Review Package Hash Trace Lookup
**Status**: Accepted

## Summary

Sprint T is accepted.

Codex completed the remaining implementation and verification work directly. The final fixes address the prior acceptance blockers:

- Audit table Package Hash click-through now traces the clicked hash directly instead of relying on stale React state.
- Frontend trace response handling now matches the backend API shape: `GET /api/v2/agent-runs/review-package-trace` returns the trace object in `data`, not nested under `data.data`.
- The Trace UI renders concrete export, delivery, and audit refs that E2E can verify.
- Empty trace results show an explicit empty state.
- Audit E2E now proves trace request query strings, rendered aggregate counts, rendered concrete refs, empty results, invalid-hash client-side validation, click-through API calls, and sensitive delivery field non-rendering.
- Playwright defaults to the non-interactive `line` reporter, and verification used `--reporter=line`.

## Verification Run

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```

Result: PASS

```bash
GOCACHE=/tmp/gocache go test ./...
```

Result: PASS

```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/app/dashboard/audit/audit-page-content.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts e2e/audit.spec.ts
```

Result: PASS

```bash
cd frontend-next && npm run build
```

Result: PASS

```bash
cd frontend-next && npx playwright test --reporter=line e2e/audit.spec.ts
```

Result: PASS - 6 passed

```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts e2e/audit.spec.ts
```

Result: PASS - 20 passed

## Acceptance Findings

### 1. Backend API contract is present and verified

`GET /api/v2/agent-runs/review-package-trace?package_hash=...` is registered before `/:id`, validates missing and invalid hashes, returns 400 for invalid input, returns 200 with empty arrays for no matches, and aggregates Agent Run, export, delivery, and audit refs for matched package hashes.

Backend tests cover unauthenticated access, missing hash, invalid hash, empty result, successful trace, route order, and delivered/failed/timeout summary counts.

### 2. Frontend trace flow is now correct

The Audit page now passes a clicked Package Hash directly into `handleTracePackage(packageHash)`, updates the input to the traced hash, and calls the trace API with that exact value.

The API client and page now read the backend response shape correctly. This fixes the previous issue where a successful backend response did not render because the UI expected `resp.data.data`.

### 3. E2E proof is acceptance-grade

`frontend-next/e2e/audit.spec.ts` now:

- waits for the trace API request and asserts `package_hash` query values;
- proves invalid hashes do not call the trace API;
- asserts visible summary counts for Agent Runs, Exports, Deliveries, Audits, Delivered, Failed, and Timeout;
- asserts concrete refs and safe fields including `Test Agent Run`, `op-export-1`, `op-delivery-1`, `audit-1`, and `hooks.example.com`;
- asserts empty trace results render zero counts and a real empty-state message;
- asserts audit table click-through sends the clicked hash and populates the input;
- injects sensitive fields into the mocked response and proves webhook URL, headers, bearer token, API key, cookie, response body secret, and other secret values are not rendered.

### 4. No prohibited scope creep observed

No intentional scope creep into report center, package storage, signatures/PKI, saved connectors, retry queues, Scanner Hub expansion, workflow engine, or MCP auto-delivery was observed.

## Decision

Sprint T is accepted. The Package Hash Trace Lookup contract is implemented, user-facing trace rendering is observable, and the E2E suite now proves the critical request, rendering, empty-state, click-through, and sanitization behavior.
