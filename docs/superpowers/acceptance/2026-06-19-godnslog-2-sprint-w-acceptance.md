# GODNSLOG 2.0 Sprint W Acceptance

Date: 2026-06-19

## Scope

Sprint W validates the Scanner Package Manifest increment:

- `ScannerRun` now includes `package_manifest` and `package_hash`.
- Scanner Hub generates a machine-readable manifest with schema version, scanner, delivery method, generated file list, hash algorithm, correlation URLs, and next actions.
- Package hash is deterministic for equivalent package contents and excludes the JSONL timestamp from hash input.
- Create/list/detail API responses expose the manifest and hash.
- Scanner Hub create and detail pages display package hash and manifest files.
- E2E verifies package hash and manifest display without using Playwright HTML report server.

## Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub -run 'TestGenerateScannerPackageIncludesManifestAndHash|TestGenerateScannerPackageHashIsDeterministic|TestGenerateScannerArtifactsForPrimaryAdapters'
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server -run 'TestGenerateScannerPackage|TestGenerateScannerArtifactsForPrimaryAdapters|TestCreateScannerRun|TestV2CreateScannerRunSupportsBurpAndRejectsWrongDelivery'
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS, run unsandboxed because internal/agentrun httptest listeners require local socket binding
```

```bash
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/app/dashboard/scanner-hub/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
# PASS
```

```bash
cd frontend-next && npm run build
# PASS
```

```bash
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts -g "package manifest"
# PASS - 2 passed
```

```bash
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
# PASS - 17 passed
```

## Findings

- Sprint W improves Scanner Hub competitiveness for AI Agent and CI consumers by replacing string-only package output with a stable manifest/hash contract.
- The implementation still does not execute scanners, schedule scans, compile plugins, or stream live scanner events.
- Existing dev-server logs can show mocked navigation/API errors from E2E paths that intentionally do not mock unrelated destination pages; the Scanner Hub assertions pass.

## Decision

Accepted. Sprint W meets the planned manifest/hash scope and is ready for the next product increment.
