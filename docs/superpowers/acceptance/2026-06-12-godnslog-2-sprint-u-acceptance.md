# Sprint U Acceptance: Scanner Hub Multi-Tool Adapter Packages

**Date**: 2026-06-12
**Sprint**: U - Scanner Hub Multi-Tool Adapter Packages
**Status**: Accepted

## Summary

Sprint U is accepted.

Codex implemented the next product-competitiveness slice directly: Scanner Hub is no longer Nuclei-only. It now has a backend adapter catalog, scanner/delivery compatibility validation, persisted Scanner Run package generation for the primary tool matrix, frontend adapter selection, and request-level E2E coverage for Nuclei, Burp Suite, and Yakit/Yak.

## Implemented Scope

- Added `GET /api/v2/scanner-hub/adapters`.
- Added official adapter catalog entries for:
  - Nuclei
  - Burp Suite
  - Yakit/Yak
  - ZAP
  - xray
  - rad
  - Postman
  - Apifox
- Extended scanner/delivery validation:
  - `nuclei -> nuclei-jsonl, nuclei-var`
  - `burp -> burp-extension`
  - `yakit -> yakit-script`
  - `zap -> zap-script`
  - `xray -> xray-webhook`
  - `rad -> rad-webhook`
  - `postman -> postman-env`
  - `apifox -> apifox-env`
- Extended Scanner Run artifact generation so each supported adapter gets:
  - `command` package text;
  - single-line `jsonl` metadata;
  - Case/Payload/token/rendered payload references;
  - Interactions and Evidence URLs.
- Updated Scanner Hub UI from a Nuclei-only workspace to a multi-tool adapter workspace.
- Fixed frontend Scanner Run create response handling to match the backend's single-layer `data` response.
- Updated `docs/scanner-hub.md` to document Sprint U support boundaries.

## Verification Run

### 2026-06-19 Re-Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
```

Result: PASS

```bash
GOCACHE=/tmp/gocache go test ./...
```

Result: PASS

```bash
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
```

Result: PASS

```bash
cd frontend-next && npm run build
```

Result: PASS

```bash
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts e2e/audit.spec.ts
```

Result: PASS - 21 passed

```bash
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx e2e/scanner-hub.spec.ts
```

Result: PASS

```bash
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
```

Result: PASS - 15 passed

```bash
git diff --check
```

Result: PASS

Notes:

- `npx playwright install chromium` was required because the resumed environment was missing the Playwright Chromium headless shell.
- Playwright was run only with `--reporter=line`; no HTML report server was started.
- The 2026-06-19 patch de-duplicates the created Payload before updating the Scanner Hub payload selector, preventing duplicate React keys when the API or E2E mock returns an already-listed payload ID.

### 2026-06-12 Original Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
```

Result: PASS

```bash
GOCACHE=/tmp/gocache go test ./...
```

Result: PASS

```bash
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
```

Result: PASS

```bash
cd frontend-next && npm run build
```

Result: PASS

```bash
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
```

Result: PASS - 15 passed

## Acceptance Findings

### 1. Backend Adapter Catalog

`GET /api/v2/scanner-hub/adapters` returns the official catalog. Server tests verify the route is registered and the response includes Burp Suite, Yakit/Yak, ZAP, xray, rad, Postman, and Apifox.

### 2. Compatibility Validation

`internal/scannerhub` now validates scanner/delivery pairs against an explicit compatibility matrix. Unit tests cover all valid pairs and invalid scanner / invalid pair failures.

### 3. Artifact Generation

Backend tests verify generated artifacts for Nuclei, Burp Suite, Yakit/Yak, ZAP, xray, rad, Postman, and Apifox. The generated JSONL records include scanner, delivery method, Case ID, Payload ID, token, target, template, rendered payload, interactions URL, and evidence URL.

### 4. Frontend Multi-Tool Workspace

The Scanner Hub page now loads adapter metadata from the API, renders a scanner selector, shows the full adapter matrix, and generates packages using the selected scanner and delivery method.

E2E proves:

- the page is no longer Nuclei-only;
- Burp Suite, Yakit/Yak, ZAP, xray, rad, Postman, and Apifox appear in the UI;
- Nuclei generation still works;
- Burp Suite sends `scanner=burp` and `delivery_method=burp-extension`;
- Yakit/Yak sends `scanner=yakit` and `delivery_method=yakit-script`;
- generated package labels and output text are visible.

### 5. No Prohibited Scope Creep

No scanner execution, scheduling, worker queue, native plugin binary, plugin marketplace, bidirectional event stream, Workflow expansion, or MCP auto-delivery was introduced.

## Decision

Sprint U is accepted. Scanner Hub now has a credible multi-tool support surface while preserving the product boundary: GODNSLOG generates audited OAST integration packages and evidence links; it does not execute external scanners.
