# GODNSLOG 2.0 Sprint X Acceptance

Date: 2026-06-19

## Scope

Sprint X validates the Agent-friendly Evidence Summary API increment:

- Added `internal/evidencehub` service.
- Added `POST /api/v2/evidence/summary`.
- Summary requests accept `case_id`, `payload_id`, or `scanner_run_id`.
- `scanner_run_id` resolves to Case/Payload scope and returns the related Scanner Run.
- Response includes `scope`, structured `evidence`, related `scanner_runs`, `package_hashes`, deterministic `summary_hash`, `next_actions`, `generated_at`, and metadata.
- Frontend API types and `evidenceApi.summary()` are available for future UI/MCP use.

## Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/evidencehub
# PASS
```

```bash
GOCACHE=/tmp/gocache go test ./internal/evidencehub ./server -run 'TestBuildSummary|TestV2EvidenceSummaryWithScannerRun'
# PASS
```

### Full Verification (2026-06-19)

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS — all packages ok
```

```bash
cd frontend-next && npx eslint src/lib/api-client.ts src/types/index.ts
# PASS — no errors
```

```bash
cd frontend-next && npm run build
# PASS — Compiled successfully, 22/22 static pages generated
```

```bash
git diff --check
# PASS
```

## Findings

- The summary service reuses the existing on-demand EvidenceService and does not introduce a persistent Evidence table.
- Scanner Run lookup uses the current XORM `scanner_runs` column names (`case_i_d`, `payload_i_d`) to remain compatible with the existing schema.
- No scanner execution, scheduling, plugin build, or LLM summarization was added.

## Decision

Accepted. Sprint X provides the read-only evidence bundle needed by AI Agent and CI consumers. Full verification (go test ./..., frontend lint, frontend build, git diff --check) passed on 2026-06-19.
