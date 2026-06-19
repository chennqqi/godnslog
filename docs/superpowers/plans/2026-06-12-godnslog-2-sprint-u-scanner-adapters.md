# GODNSLOG 2.0 Sprint U Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Turn Scanner Hub from a Nuclei-only workspace into a verifiable multi-tool adapter catalog that can generate official integration packages for Nuclei, Burp Suite, Yakit/Yak, ZAP, xray/rad, Postman, and Apifox.

**Architecture:** Keep Scanner Hub as a distribution and evidence-correlation layer, not a scanner scheduler. Add a backend adapter catalog with scanner/delivery compatibility rules, use the existing `ScannerRun` persistence model for generated integration packages, and render safe command/script/webhook/env snippets in the Web UI. All scanner outputs must still point back to Case, Payload, Interaction, Evidence, and Audit contracts.

**Tech Stack:** Go 1.25, Gin, XORM, `internal/scannerhub`, `internal/models`, `/api/v2`, Next.js 16, TypeScript, shadcn/ui, Playwright.

---

## Sprint Boundary

### In Scope

- Add a first-class Scanner Hub adapter catalog.
- Support scanner names: `nuclei`, `burp`, `yakit`, `zap`, `xray`, `rad`, `postman`, `apifox`.
- Support delivery methods:
  - `nuclei-jsonl`
  - `nuclei-var`
  - `burp-extension`
  - `yakit-script`
  - `zap-script`
  - `xray-webhook`
  - `rad-webhook`
  - `postman-env`
  - `apifox-env`
- Generate persisted Scanner Run artifacts for each supported scanner/delivery pair:
  - `command`: operator-facing command, script, webhook contract, or environment setup.
  - `jsonl`: one-line structured adapter package metadata.
- Add `GET /api/v2/scanner-hub/adapters` for UI/API discovery.
- Update Scanner Hub UI from "Nuclei workspace" to "multi-tool adapter workspace".
- Add E2E proof for at least Burp Suite and Yakit/Yak generation in addition to Nuclei.
- Keep audit creation for generated scanner runs.

### Out of Scope

- No scanner execution.
- No remote plugin build, upload, marketplace, or update channel.
- No long-running scan scheduling or queue workers.
- No native Burp/Yakit/ZAP plugin binaries.
- No bidirectional live scanner event streaming.
- No SARIF exporter unless already available through existing evidence export.
- No Workflow engine expansion.
- No MCP auto-delivery expansion.

## File Structure Map

- Modify `internal/models/scanner_run.go`: scanner and delivery constants, request validation tags, comments.
- Modify `internal/scannerhub/service.go`: adapter catalog, compatibility validation, artifact generation.
- Modify `server/v2_api.go`: register and implement adapter catalog endpoint.
- Modify `server/v2_api_test.go`: API coverage for adapter catalog and non-Nuclei scanner run creation.
- Modify `internal/scannerhub/service_test.go`: unit coverage for supported pairs and rejected invalid pairs.
- Modify `frontend-next/src/types/index.ts`: scanner and delivery union types, adapter catalog response types.
- Modify `frontend-next/src/lib/api-client.ts`: `scannerRunApi.listAdapters()`.
- Modify `frontend-next/src/lib/scanner-hub.ts`: scanner-aware request builder and artifact helper types.
- Modify `frontend-next/src/app/dashboard/scanner-hub/page.tsx`: adapter selector, matrix display, generated package output labels.
- Modify `frontend-next/e2e/scanner-hub.spec.ts`: request-level and rendering assertions for Nuclei, Burp Suite, and Yakit/Yak.
- Modify `docs/scanner-hub.md`: update support status from future-only to Sprint U adapter package support.
- Modify `docs/verification.md`: add Sprint U verification results after implementation.

## Adapter Contract

### Backend Response Shape

Add these Go structs in `internal/models/scanner_run.go`:

```go
type ScannerAdapter struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	Category        string   `json:"category"`
	Maturity        string   `json:"maturity"`
	SupportedMethods []string `json:"supported_methods"`
	DefaultMethod   string   `json:"default_method"`
	Description     string   `json:"description"`
}

type ScannerAdapterListResponse struct {
	Items []ScannerAdapter `json:"items"`
}
```

### Supported Compatibility Matrix

```text
nuclei  -> nuclei-jsonl, nuclei-var
burp    -> burp-extension
yakit   -> yakit-script
zap     -> zap-script
xray    -> xray-webhook
rad     -> rad-webhook
postman -> postman-env
apifox  -> apifox-env
```

Invalid scanner names and invalid scanner/delivery combinations must return 400 from `POST /api/v2/scanner-runs`.

## Task 1: Backend Adapter Catalog

**Files:**
- Modify: `internal/models/scanner_run.go`
- Modify: `internal/scannerhub/service.go`
- Test: `internal/scannerhub/service_test.go`

- [ ] **Step 1: Add failing service tests**

Add tests:

```go
func TestListScannerAdaptersIncludesPrimaryTools(t *testing.T) {
	service := NewService(nil)
	resp := service.ListAdapters()
	require.NotNil(t, resp)

	byID := map[string]models.ScannerAdapter{}
	for _, adapter := range resp.Items {
		byID[adapter.ID] = adapter
	}

	require.Contains(t, byID, models.ScannerNuclei)
	require.Contains(t, byID, models.ScannerBurp)
	require.Contains(t, byID, models.ScannerYakit)
	require.Contains(t, byID, models.ScannerZap)
	require.Contains(t, byID, models.ScannerXray)
	require.Contains(t, byID, models.ScannerRad)
	require.Contains(t, byID, models.ScannerPostman)
	require.Contains(t, byID, models.ScannerApifox)

	assert.Equal(t, models.DeliveryMethodBurpExtension, byID[models.ScannerBurp].DefaultMethod)
	assert.Contains(t, byID[models.ScannerYakit].SupportedMethods, models.DeliveryMethodYakitScript)
}

func TestValidateScannerDeliveryPair(t *testing.T) {
	tests := []struct {
		name     string
		scanner  string
		method   string
		wantErr  error
	}{
		{"nuclei jsonl", models.ScannerNuclei, models.DeliveryMethodNucleiJsonl, nil},
		{"burp extension", models.ScannerBurp, models.DeliveryMethodBurpExtension, nil},
		{"yakit script", models.ScannerYakit, models.DeliveryMethodYakitScript, nil},
		{"zap script", models.ScannerZap, models.DeliveryMethodZapScript, nil},
		{"xray webhook", models.ScannerXray, models.DeliveryMethodXrayWebhook, nil},
		{"rad webhook", models.ScannerRad, models.DeliveryMethodRadWebhook, nil},
		{"postman env", models.ScannerPostman, models.DeliveryMethodPostmanEnv, nil},
		{"apifox env", models.ScannerApifox, models.DeliveryMethodApifoxEnv, nil},
		{"unknown scanner", "unknown", models.DeliveryMethodNucleiJsonl, ErrInvalidScanner},
		{"wrong pair", models.ScannerBurp, models.DeliveryMethodNucleiJsonl, ErrInvalidDelivery},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScannerDelivery(tt.scanner, tt.method)
			if tt.wantErr == nil {
				require.NoError(t, err)
				return
			}
			require.ErrorIs(t, err, tt.wantErr)
		})
	}
}
```

- [ ] **Step 2: Run the failing tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub -run 'TestListScannerAdaptersIncludesPrimaryTools|TestValidateScannerDeliveryPair'
```

Expected: FAIL because the adapter catalog and non-Nuclei constants do not exist yet.

- [ ] **Step 3: Add constants and catalog**

In `internal/models/scanner_run.go`, add scanner/delivery constants for every supported tool and the adapter response structs.

In `internal/scannerhub/service.go`, add:

```go
func (s *Service) ListAdapters() *models.ScannerAdapterListResponse
func validateScannerDelivery(scanner, delivery string) error
```

The catalog order must be stable: Nuclei, Burp Suite, Yakit/Yak, ZAP, xray, rad, Postman, Apifox.

- [ ] **Step 4: Replace Nuclei-only validation**

In `CreateScannerRun`, replace the direct `req.Scanner != models.ScannerNuclei` and Nuclei delivery checks with `validateScannerDelivery(req.Scanner, req.DeliveryMethod)`.

- [ ] **Step 5: Run focused tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub
```

Expected: PASS.

## Task 2: Multi-Tool Artifact Generation

**Files:**
- Modify: `internal/scannerhub/service.go`
- Test: `internal/scannerhub/service_test.go`

- [ ] **Step 1: Add failing artifact tests**

Add table-driven tests that create a fake request and payload, then call the artifact generation helper for:

- Burp Suite: command contains `Burp Suite Extension`, `/api/v2/payloads`, `/api/v2/interactions?payload_id=payload-1`.
- Yakit/Yak: command contains `yak`, `CreateHTTPFlow`, and the rendered payload.
- ZAP: command contains `ZAP Script` and `zap.script`.
- xray/rad: command contains `webhook` and `/api/v2/interactions?payload_id=payload-1`.
- Postman/Apifox: command contains environment variable names `GODNSLOG_PAYLOAD`, `GODNSLOG_INTERACTIONS_URL`, `GODNSLOG_EVIDENCE_URL`.

For every generated artifact, parse `jsonl` and assert:

```go
assert.Equal(t, tt.scanner, record["scanner"])
assert.Equal(t, "case-1", record["case_id"])
assert.Equal(t, "payload-1", record["payload_id"])
assert.Equal(t, "tok-abc123", record["token"])
assert.Equal(t, tt.method, record["delivery_method"])
assert.Equal(t, "http://tok-abc123.example.com/callback", record["rendered_payload"])
```

- [ ] **Step 2: Run failing artifact tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub -run TestGenerateScannerArtifactsForPrimaryAdapters
```

Expected: FAIL because only Nuclei artifact generation exists.

- [ ] **Step 3: Implement scanner-aware artifact generation**

Replace `generateNucleiCommandAndJsonl` with:

```go
func (s *Service) generateScannerArtifacts(req *models.ScannerRunCreateRequest, payload *models.Payload, baseURL string) (string, string, error)
```

Keep Nuclei output backward compatible. Add deterministic command/script text for other tools. The JSONL metadata must include `scanner`, `delivery_method`, `case_id`, `payload_id`, `token`, `target`, `template`, `rendered_payload`, `interactions_url`, `evidence_url`, and `created_at`.

- [ ] **Step 4: Wire `CreateScannerRun` to the new helper**

`CreateScannerRun` must call `generateScannerArtifacts` for all supported scanners.

- [ ] **Step 5: Run scannerhub tests**

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub
```

Expected: PASS.

## Task 3: V2 API Endpoint and Server Coverage

**Files:**
- Modify: `server/v2_api.go`
- Modify: `server/v2_api_test.go`

- [ ] **Step 1: Add failing API tests**

Add tests that prove:

- `GET /api/v2/scanner-hub/adapters` returns 200 and includes `burp`, `yakit`, `zap`, `xray`, `rad`, `postman`, `apifox`.
- `POST /api/v2/scanner-runs` accepts `scanner=burp`, `delivery_method=burp-extension`.
- `POST /api/v2/scanner-runs` rejects `scanner=burp`, `delivery_method=nuclei-jsonl` with 400.

- [ ] **Step 2: Run failing server tests**

```bash
GOCACHE=/tmp/gocache go test ./server -run 'TestV2ScannerAdapters|TestV2CreateScannerRun'
```

Expected: FAIL because the adapter endpoint is not registered and non-Nuclei requests are rejected.

- [ ] **Step 3: Add route**

Register:

```go
v2.GET("/scanner-hub/adapters", self.v2ListScannerAdapters)
```

Use the existing auth middleware rules consistently with other Scanner Hub endpoints.

- [ ] **Step 4: Implement handler**

Add:

```go
func (self *WebServer) v2ListScannerAdapters(c *gin.Context) {
	scannerHubService := scannerhub.NewService(self.orm)
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": scannerHubService.ListAdapters()})
}
```

- [ ] **Step 5: Run server tests**

```bash
GOCACHE=/tmp/gocache go test ./server
```

Expected: PASS.

## Task 4: Frontend Types and API Client

**Files:**
- Modify: `frontend-next/src/types/index.ts`
- Modify: `frontend-next/src/lib/api-client.ts`
- Modify: `frontend-next/src/lib/scanner-hub.ts`

- [ ] **Step 1: Add TypeScript scanner unions**

Add:

```ts
export type ScannerKind = 'nuclei' | 'burp' | 'yakit' | 'zap' | 'xray' | 'rad' | 'postman' | 'apifox'
export type ScannerDeliveryMethod =
  | 'nuclei-jsonl'
  | 'nuclei-var'
  | 'burp-extension'
  | 'yakit-script'
  | 'zap-script'
  | 'xray-webhook'
  | 'rad-webhook'
  | 'postman-env'
  | 'apifox-env'
```

Use these types in `ScannerRun` and `ScannerRunCreateRequest`.

- [ ] **Step 2: Add adapter response types**

```ts
export interface ScannerAdapter {
  id: ScannerKind
  name: string
  category: string
  maturity: string
  supported_methods: ScannerDeliveryMethod[]
  default_method: ScannerDeliveryMethod
  description: string
}

export interface ScannerAdapterListResponse {
  items: ScannerAdapter[]
}
```

- [ ] **Step 3: Add API client**

In `scannerRunApi`, add:

```ts
listAdapters: () => api.get<ScannerAdapterListResponse>('/scanner-hub/adapters')
```

- [ ] **Step 4: Update `createScannerRun` helper**

Let `createScannerRun(input, scanner, deliveryMethod)` pass selected scanner and method instead of hard-coding Nuclei.

- [ ] **Step 5: Run lint**

```bash
cd frontend-next && npx eslint src/types/index.ts src/lib/api-client.ts src/lib/scanner-hub.ts
```

Expected: PASS.

## Task 5: Scanner Hub UI

**Files:**
- Modify: `frontend-next/src/app/dashboard/scanner-hub/page.tsx`
- Test: `frontend-next/e2e/scanner-hub.spec.ts`

- [ ] **Step 1: Add failing E2E assertions**

Update mocks to include `GET /api/v2/scanner-hub/adapters`. Add tests that prove:

- The page title no longer says only "Nuclei 集成工作台".
- Adapter cards or options include Burp Suite, Yakit/Yak, ZAP, xray/rad, Postman, Apifox.
- Selecting Burp Suite sends `scanner=burp` and `delivery_method=burp-extension` to `POST /api/v2/scanner-runs`.
- Selecting Yakit/Yak sends `scanner=yakit` and `delivery_method=yakit-script`.
- Generated output renders scanner-specific labels: `Burp Suite Extension Package`, `Yakit/Yak Script Package`, and `JSONL Preview`.

- [ ] **Step 2: Run failing E2E**

```bash
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
```

Expected: FAIL because the page is currently Nuclei-only.

- [ ] **Step 3: Load adapters in UI**

Call `scannerRunApi.listAdapters()` on page load. Store adapter list in state. Default selection should be `nuclei` and `nuclei-jsonl`.

- [ ] **Step 4: Add scanner selector**

Render a scanner selector before target/template/payload selection. When scanner changes, set delivery method to adapter `default_method`.

- [ ] **Step 5: Update generate handler**

Pass selected scanner and selected delivery method to `createScannerRun`.

- [ ] **Step 6: Update generated output labels**

Use scanner-specific labels:

- Nuclei: `Nuclei Command`
- Burp: `Burp Suite Extension Package`
- Yakit: `Yakit/Yak Script Package`
- ZAP: `ZAP Script Package`
- xray/rad: `Webhook Bridge Package`
- Postman/Apifox: `Environment Package`

- [ ] **Step 7: Run frontend checks**

```bash
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
```

Expected: PASS.

## Task 6: Docs and Verification

**Files:**
- Modify: `docs/scanner-hub.md`
- Modify: `docs/verification.md`
- Create: `docs/superpowers/acceptance/2026-06-12-godnslog-2-sprint-u-acceptance.md`

- [ ] **Step 1: Update scanner hub docs**

Document Sprint U support as "official package generation" for the supported tools, while clearly saying native plugin binaries and scan execution remain out of scope.

- [ ] **Step 2: Run full verification**

```bash
GOCACHE=/tmp/gocache go test ./internal/scannerhub ./server
GOCACHE=/tmp/gocache go test ./...
cd frontend-next && npx eslint src/app/dashboard/scanner-hub/page.tsx src/lib/scanner-hub.ts src/lib/api-client.ts src/types/index.ts e2e/scanner-hub.spec.ts
cd frontend-next && npm run build
cd frontend-next && npx playwright test --reporter=line e2e/scanner-hub.spec.ts
```

Expected: all PASS.

- [ ] **Step 3: Record results**

Append Sprint U verification results to `docs/verification.md`.

- [ ] **Step 4: Write acceptance**

Create Sprint U acceptance with:

- implementation summary;
- verification commands and exact outcomes;
- acceptance matrix against all supported tools;
- no-scope-creep audit;
- final accepted/not-accepted decision.

## Acceptance Criteria

- `GET /api/v2/scanner-hub/adapters` returns the official adapter catalog.
- Backend accepts all supported scanner/delivery pairs and rejects invalid pairs.
- Generated Scanner Run artifacts include Case, Payload, token, rendered payload, interactions URL, evidence URL, scanner, and delivery method.
- Scanner Hub UI is no longer Nuclei-only.
- UI can generate at least Nuclei, Burp Suite, and Yakit/Yak packages through real mocked API requests in E2E.
- E2E asserts request body values and visible generated package labels, not only static text.
- Existing Scanner Run list/detail/status behavior remains intact.
- Full verification is recorded in `docs/verification.md`.
- No scan execution, scheduler, plugin marketplace, or MCP auto-delivery is introduced.

## Next Sprint Candidate

After Sprint U, the strongest next candidate is Sprint V: Agent Scope & Risk Controls. It should take the Agent-Native specification and enforce scoped API keys, risk labels, and visible audit/risk summaries for MCP/Agent Run operations.
