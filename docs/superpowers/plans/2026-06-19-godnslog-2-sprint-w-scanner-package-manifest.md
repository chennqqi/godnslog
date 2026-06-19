# GODNSLOG 2.0 Sprint W Scanner Package Manifest Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make Scanner Hub packages machine-readable and verifiable for AI Agent, CI, and scanner automation consumers.

**Architecture:** Keep Scanner Hub as a distribution and evidence-correlation layer. Extend `ScannerRun` with a structured integration package manifest and deterministic SHA-256 hash derived from scanner, delivery method, command, JSONL, generated files, next actions, and correlation URLs. Reuse existing Scanner Run APIs and UI; do not execute scanners or generate compiled plugin binaries.

**Tech Stack:** Go 1.25, XORM, `internal/models`, `internal/scannerhub`, Gin `/api/v2`, Next.js 16, TypeScript, Playwright.

---

## Sprint Boundary

### In Scope

- Add structured Scanner Hub package manifest fields:
  - schema version;
  - scanner and delivery method;
  - generated file list;
  - package hash and hash algorithm;
  - interactions/evidence URLs;
  - next actions for operator or Agent automation.
- Persist `package_manifest` and `package_hash` on `ScannerRun`.
- Include manifest/hash in create, list, and detail responses.
- Add deterministic hash tests proving stable hash for equivalent package data.
- Add frontend types and display package hash/manifest on Scanner Hub create and detail pages.
- Update Scanner Hub docs, verification log, and Sprint W acceptance.

### Out of Scope

- No scanner process execution.
- No scheduling, queue worker, plugin compilation, binary download, or live bidirectional scanner streaming.
- No destructive scanner actions.
- No new top-level navigation.

## Tasks

### Task 1: Backend Manifest Model and Generation

**Files:**
- Modify: `internal/models/scanner_run.go`
- Modify: `internal/scannerhub/service.go`
- Modify: `internal/scannerhub/service_test.go`

- [ ] Add failing tests for `package_manifest` and `package_hash` on generated Nuclei and Burp packages.
- [ ] Add scanner package manifest structs and fields to `ScannerRun`.
- [ ] Generate a stable manifest with files and next actions.
- [ ] Compute deterministic package hash using `models.ComputeDeterministicHash`.
- [ ] Persist manifest/hash in `CreateScannerRun`.

### Task 2: API and UI Contract

**Files:**
- Modify: `server/v2_api_test.go`
- Modify: `frontend-next/src/types/index.ts`
- Modify: `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- Modify: `frontend-next/src/app/dashboard/scanner-hub/[id]/page.tsx`
- Modify: `frontend-next/e2e/scanner-hub.spec.ts`

- [ ] Add API tests proving create/detail responses include package hash and manifest.
- [ ] Add TypeScript types for scanner package manifest.
- [ ] Show package hash and manifest summary after creating a scanner run.
- [ ] Show package hash and manifest summary on scanner run detail.
- [ ] Add E2E assertions for package hash and generated file names.

### Task 3: Documentation and Acceptance

**Files:**
- Modify: `docs/scanner-hub.md`
- Modify: `docs/verification.md`
- Create: `docs/superpowers/acceptance/2026-06-19-godnslog-2-sprint-w-acceptance.md`

- [ ] Document the manifest/hash response contract.
- [ ] Record actual verification commands and results.
- [ ] Write Sprint W acceptance with scope, findings, and decision.

## Verification Commands

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/app/dashboard/scanner-hub/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
git diff --check
```
