# GODNSLOG 2.0 Sprint Z Evidence Summary UI Page Acceptance

**Date:** 2026-06-19  
**Sprint:** Z — Evidence Summary UI Page

## Scope

Add a frontend UI page for the Evidence Summary API (`POST /api/v2/evidence/summary`), allowing operators to query structured evidence bundles from the web interface.

### In Scope

- New page at `/dashboard/evidence-summary` with scope selection (case/payload/scanner_run)
- Evidence stats display (strength, confidence, interaction count, unique sources)
- Scanner runs, package hashes, next actions, and metadata panels
- Sidebar navigation entry and page title mapping
- URL param support for `case_id`, `payload_id`, `scanner_run_id`

### Out of Scope

- No backend changes
- No new API endpoints
- No E2E tests (page is read-only query UI)

## Verification Steps

```bash
cd frontend-next && npx eslint src/app/dashboard/evidence-summary/page.tsx src/components/app-shell/sidebar.tsx
# PASS — no errors
```

```bash
cd frontend-next && npm run build
# PASS — Compiled successfully, 23/23 static pages generated
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

- The page reuses the existing `evidenceApi.summary` API client method and `EvidenceSummaryResponse` types from Sprint X.
- URL params (`case_id`, `payload_id`, `scanner_run_id`) enable deep linking from other pages.
- Sidebar navigation added under OAST CORE group with a clipboard-check icon.
- Pre-existing lint error in `app-shell/index.tsx` (set-state-in-effect) is unrelated to this sprint.

## Decision

Accepted. Sprint Z provides a functional Evidence Summary UI page with full scope selection, evidence display, and navigation integration.
