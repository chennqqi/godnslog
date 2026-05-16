# GODNSLOG 2.0 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Restore the current GODNSLOG 2.0 codebase to a verifiable state, then deliver the adjusted product plan as a working self-hosted OAST evidence hub for security teams, scanners, and AI Agents.

**Architecture:** Treat DNS/HTTP Interaction as the core platform event and make every feature consume the same Case, Payload, Interaction, Evidence, APIKey, and Audit contracts. Keep 2.0 backend modules in `internal/` and route all external consumers through `/api/v2`, CLI, MCP, and integration examples. Defer broad platform features until the MVP OAST evidence loop is working end to end.

**Tech Stack:** Go 1.25, Gin, XORM, SQLite/MySQL, Next.js 16, React 18, TypeScript, shadcn/ui, TanStack Query, Playwright, Cobra CLI, MCP HTTP transport.

---

## Current State Summary

The planning documents now position GODNSLOG 2.0 as a self-hosted OAST evidence hub, but the implementation is not yet aligned. A fresh `GOCACHE=/tmp/gocache go test ./...` currently fails before business tests can run because `go.sum` is missing and `internal/case/service_test.go` declares `package case`, which is invalid Go syntax. The code also contains visible incomplete areas in router duplication, auth middleware, workflow action execution, evidence generation, multi-tool scanner integration, and several platform modules that are present but not production-grade.

The implementation order below is intentionally strict:

1. Restore build and test baseline.
2. Make the 2.0 DNS/HTTP OAST evidence loop real.
3. Make CLI and Scanner Hub consume the same API.
4. Make MCP Agent workflows structured and auditable.
5. Expand platform features after the core loop is stable.

## File Structure Map

- `go.mod`, `go.sum`: dependency integrity and reproducible test execution.
- `server/router.go`, `server/v2_api.go`, `server/middleware.go`: `/api/v2` route source of truth and request authentication.
- `server/dnsserver.go`, `server/webapi.go`, `server/webserver.go`: existing DNS/HTTP capture paths that must write or expose unified Interaction records.
- `internal/interaction/`: Interaction service, evidence generation, scoring, timeline, noise handling.
- `internal/payload/`: template rendering, token generation, preview, lifecycle.
- `internal/case/`: Case CRUD and associated payload/interaction queries.
- `internal/rule/`, `internal/workflow/`, `internal/notification/`: workflow conditions, actions, queue, and notifications.
- `cli/`, `cmd/cli/`: command-line integration for scanners and CI.
- `internal/mcp/`, `cmd/mcp-server/`: Agent tool surface and MCP server runtime.
- `frontend-next/src/features/`, `frontend-next/src/app/dashboard/`: Next.js UI and API clients.
- `examples/`, `extensions/`, `docs/`: scanner/tool integration artifacts and user-facing documentation.

---

## Milestone 0: Verifiable Engineering Baseline

### Task 0.1: Restore Go Dependency Integrity

**Files:**
- Modify: `go.sum`
- Verify: `go.mod`

- [ ] **Step 1: Run the current failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./...
```

Expected: FAIL with missing `go.sum` entries and `internal/case/service_test.go:1:9: expected 'IDENT', found 'case`.

- [ ] **Step 2: Generate dependency checksums**

Run:

```bash
GOCACHE=/tmp/gocache go mod tidy
```

Expected: `go.sum` is created and no dependency resolution errors remain. If network access is blocked, rerun the same command with escalated network permission rather than manually editing `go.sum`.

- [ ] **Step 3: Verify dependency failure is gone**

Run:

```bash
GOCACHE=/tmp/gocache go test ./...
```

Expected: Missing `go.sum` errors are gone. The next failure should be real compile or test failures.

- [ ] **Step 4: Commit dependency baseline when Git is writable**

```bash
git add go.mod go.sum
git commit -m "chore: restore go dependency checksums"
```

### Task 0.2: Fix Invalid Go Test Package Names

**Files:**
- Modify: `internal/case/service_test.go:1`
- Test: `internal/case/service_test.go`

- [ ] **Step 1: Confirm syntax failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/case
```

Expected: FAIL with `expected 'IDENT', found 'case`.

- [ ] **Step 2: Rename test package**

Change the first line of `internal/case/service_test.go` from:

```go
package case
```

to:

```go
package casemgmt
```

- [ ] **Step 3: Run focused test**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/case
```

Expected: The package parses. Remaining failures, if any, should be database driver, model, or service behavior issues.

- [ ] **Step 4: Commit syntax fix**

```bash
git add internal/case/service_test.go
git commit -m "test: fix case service test package"
```

### Task 0.3: Establish Test Commands as Release Gates

**Files:**
- Create: `docs/verification.md`
- Modify: `AGENTS.md`

- [ ] **Step 1: Add verification document**

Create `docs/verification.md` with:

```markdown
# Verification Gates

Run these commands before claiming a GODNSLOG 2.0 change is complete.

## Backend

```bash
GOCACHE=/tmp/gocache go test ./...
```

For focused backend changes, run the affected package first, then the full command before release.

## Frontend

```bash
cd frontend-next
npm run lint
npm run build
```

For UI behavior changes, also run the relevant Playwright spec:

```bash
cd frontend-next
npm run test:e2e -- e2e/cases.spec.ts
```

## MCP

```bash
GOCACHE=/tmp/gocache go test -count=1 ./internal/mcp
```
```

- [ ] **Step 2: Link verification from AGENTS**

Append this sentence under `AGENTS.md` testing guidance:

```markdown
完整验证命令维护在 `docs/verification.md`，完成任何功能前必须记录实际运行过的命令和结果。
```

- [ ] **Step 3: Verify docs are readable**

Run:

```bash
sed -n '1,220p' docs/verification.md
```

Expected: The document renders with Backend, Frontend, and MCP sections.

- [ ] **Step 4: Commit verification gates**

```bash
git add docs/verification.md AGENTS.md
git commit -m "docs: add verification gates"
```

---

## Milestone 1: 2.0 API Source of Truth

### Task 1.1: Remove Duplicate V2 Route Definitions

**Files:**
- Modify: `server/router.go`
- Modify: `server/webserver.go`
- Test: `server/v2_api_test.go`

- [ ] **Step 1: Write failing route registration test**

Add a test in `server/v2_api_test.go`:

```go
func TestV2RoutesExposeRequiredMVPPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ws := &WebServer{}
	ws.registerV2API(engine)

	paths := map[string]bool{}
	for _, route := range engine.Routes() {
		paths[route.Method+" "+route.Path] = true
	}

	required := []string{
		"GET /api/v2/cases/:id/payloads",
		"GET /api/v2/cases/:id/interactions",
		"PUT /api/v2/payloads/:id",
		"GET /api/v2/interactions/stats",
		"POST /api/v2/evidence/generate",
	}
	for _, route := range required {
		if !paths[route] {
			t.Fatalf("missing route %s", route)
		}
	}
}
```

- [ ] **Step 2: Run test to verify current behavior**

Run:

```bash
GOCACHE=/tmp/gocache go test ./server -run TestV2RoutesExposeRequiredMVPPaths
```

Expected: FAIL if `server/router.go` is used as the active router or if required routes are absent from active registration.

- [ ] **Step 3: Make `registerV2API` the single route source**

Keep all `/api/v2` route definitions in `server/v2_api.go`. In `server/router.go`, replace the duplicated v2 route body with:

```go
// API v2 routes are registered by WebServer.registerV2API to keep one source of truth.
r.server.registerV2API(r.engine)
```

Ensure the application boot path does not register `/api/v2` twice.

- [ ] **Step 4: Run focused route test**

Run:

```bash
GOCACHE=/tmp/gocache go test ./server -run TestV2RoutesExposeRequiredMVPPaths
```

Expected: PASS.

- [ ] **Step 5: Commit route source of truth**

```bash
git add server/router.go server/v2_api.go server/v2_api_test.go
git commit -m "fix: use one source of truth for v2 routes"
```

### Task 1.2: Make Auth Middleware Real

**Files:**
- Modify: `server/middleware.go`
- Modify: `server/v2_api_test.go`
- Modify: `internal/auth/service.go`

- [ ] **Step 1: Write failing auth test**

Add this test in `server/v2_api_test.go`:

```go
func TestAuthMiddlewareRejectsMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	ws := &WebServer{}
	engine.GET("/protected", ws.authHandler, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", rec.Code, rec.Body.String())
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./server -run TestAuthMiddlewareRejectsMissingCredentials
```

Expected: FAIL if missing credentials are accepted or middleware panics.

- [ ] **Step 3: Implement credential rejection and context assignment**

In `server/middleware.go`, make `authHandler` reject requests without valid JWT or APIKey and set authenticated user/APIKey identity in Gin context. Use the existing JWT login format from `server/v2_api.go` and existing APIKey validation from `internal/auth/service.go`.

The handler shape should be:

```go
func (self *WebServer) authHandler(c *gin.Context) {
	if user, err := self.authenticateJWT(c); err == nil && user != nil {
		c.Set("user", user)
		c.Next()
		return
	}
	if key, err := self.authenticateAPIKey(c); err == nil && key != nil {
		c.Set("api_key", key)
		c.Next()
		return
	}
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    401,
		"message": "unauthorized",
	})
}
```

- [ ] **Step 4: Run focused auth tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./server -run 'TestAuthMiddleware|TestV2Login'
```

Expected: PASS.

- [ ] **Step 5: Commit auth baseline**

```bash
git add server/middleware.go server/v2_api_test.go internal/auth/service.go
git commit -m "fix: enforce v2 authentication middleware"
```

---

## Milestone 2: MVP OAST Evidence Loop

### Task 2.1: Double-Write DNS and HTTP Captures to Interaction

**Files:**
- Modify: `server/dnsserver.go`
- Modify: `server/webapi.go`
- Modify: `internal/interaction/service.go`
- Test: `server/v2_api_test.go`

- [ ] **Step 1: Write failing interaction persistence test**

Add a test in `server/v2_api_test.go` that creates a synthetic DNS/HTTP capture and asserts `GET /api/v2/interactions` returns it. Use the existing test database setup helpers in the file.

```go
func TestCapturedHTTPLogAppearsInV2Interactions(t *testing.T) {
	ws, cleanup := newTestWebServer(t)
	defer cleanup()

	insertHTTPLogForTest(t, ws.orm, "tok123.example.com", "/callback", "127.0.0.1")

	items := listV2InteractionsForTest(t, ws, "tok123")
	if len(items) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(items))
	}
	if items[0].Token != "tok123" {
		t.Fatalf("expected token tok123, got %s", items[0].Token)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./server -run TestCapturedHTTPLogAppearsInV2Interactions
```

Expected: FAIL because old capture paths do not reliably expose unified v2 Interaction rows.

- [ ] **Step 3: Add a single capture adapter**

Create or extend a helper in `internal/interaction/service.go`:

```go
func (s *Service) RecordCapturedInteraction(input CapturedInteractionInput) (*models.Interaction, error) {
	interaction := &models.Interaction{
		ID:        models.GenerateID(),
		Type:      input.Type,
		Token:     input.Token,
		SourceIP:  input.SourceIP,
		UserAgent: input.UserAgent,
		Timestamp: time.Now(),
	}
	interaction.Domain = optionalString(input.Domain)
	interaction.Path = optionalString(input.Path)
	interaction.Method = optionalString(input.Method)
	return s.CreateInteraction(interaction)
}
```

Call this adapter from DNS and HTTP capture paths after the legacy row write succeeds.

- [ ] **Step 4: Run focused and package tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction ./server -run 'TestCapturedHTTPLogAppearsInV2Interactions|Test.*Interaction'
```

Expected: PASS.

- [ ] **Step 5: Commit unified capture**

```bash
git add server/dnsserver.go server/webapi.go internal/interaction/service.go server/v2_api_test.go
git commit -m "feat: persist captured callbacks as v2 interactions"
```

### Task 2.2: Complete Payload Template Rendering and Preview

**Files:**
- Modify: `internal/payload/service.go`
- Modify: `templates/payloads.json`
- Modify: `server/v2_api.go`
- Test: `internal/payload/service_test.go`

- [ ] **Step 1: Write failing payload rendering test**

Add this test in `internal/payload/service_test.go`:

```go
func TestRenderPayloadSupportsCustomVariablesAndCallbackURL(t *testing.T) {
	service := NewService(nil, "oast.example.com")
	rendered, err := service.RenderTemplate("https://{token}.{domain}{path}?case={case}", RenderContext{
		Token: "tok123",
		CaseID: "case-1",
		Variables: map[string]string{"path": "/callback"},
	})
	if err != nil {
		t.Fatalf("render failed: %v", err)
	}
	expected := "https://tok123.oast.example.com/callback?case=case-1"
	if rendered != expected {
		t.Fatalf("expected %q, got %q", expected, rendered)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/payload -run TestRenderPayloadSupportsCustomVariablesAndCallbackURL
```

Expected: FAIL if custom variables or `{callback_url}` style values are not rendered consistently.

- [ ] **Step 3: Implement deterministic renderer**

Expose one renderer function used by create, preview, batch, CLI, and MCP:

```go
type RenderContext struct {
	Token     string
	CaseID    string
	Domain    string
	Variables map[string]string
}
```

Render built-ins first (`token`, `case`, `domain`, `callback_url`) and then user variables. Reject unresolved braces in strict mode for API create.

- [ ] **Step 4: Run payload tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/payload
```

Expected: PASS.

- [ ] **Step 5: Commit renderer**

```bash
git add internal/payload/service.go internal/payload/service_test.go templates/payloads.json server/v2_api.go
git commit -m "feat: complete payload template rendering"
```

### Task 2.3: Implement Evidence Score and Explainable Evidence

**Files:**
- Modify: `internal/interaction/evidence.go`
- Modify: `internal/interaction/evidence_service.go`
- Modify: `server/v2_api.go`
- Test: `internal/interaction/evidence_service_test.go`

- [ ] **Step 1: Write failing evidence scoring test**

Create `internal/interaction/evidence_service_test.go` with:

```go
package interaction

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

func TestEvidenceScoreClassifiesDNSAndHTTPAsHighConfidence(t *testing.T) {
	service := &EvidenceService{}
	score := service.ScoreInteractions([]models.Interaction{
		{ID: "dns-1", Type: "dns", Token: "tok123", SourceIP: "10.0.0.1", Timestamp: time.Now()},
		{ID: "http-1", Type: "http", Token: "tok123", SourceIP: "10.0.0.1", Timestamp: time.Now()},
	})
	if score.Strength != "high" {
		t.Fatalf("expected high strength, got %s", score.Strength)
	}
	if score.Confidence < 80 {
		t.Fatalf("expected confidence >= 80, got %d", score.Confidence)
	}
	if len(score.Reasons) == 0 {
		t.Fatal("expected scoring reasons")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction -run TestEvidenceScoreClassifiesDNSAndHTTPAsHighConfidence
```

Expected: FAIL because `ScoreInteractions` and score model do not exist.

- [ ] **Step 3: Add score model and scoring rules**

In `internal/interaction/evidence.go`, add:

```go
type EvidenceScore struct {
	Strength   string   `json:"strength"`
	Confidence int      `json:"confidence"`
	Reasons    []string `json:"reasons"`
}
```

In `evidence_service.go`, implement `ScoreInteractions`:

```go
func (s *EvidenceService) ScoreInteractions(items []models.Interaction) EvidenceScore {
	seen := map[string]bool{}
	for _, item := range items {
		seen[item.Type] = true
	}
	if seen["dns"] && seen["http"] {
		return EvidenceScore{Strength: "high", Confidence: 85, Reasons: []string{"same token received DNS and HTTP callbacks"}}
	}
	if seen["http"] {
		return EvidenceScore{Strength: "medium", Confidence: 70, Reasons: []string{"target made an HTTP callback"}}
	}
	return EvidenceScore{Strength: "low", Confidence: 45, Reasons: []string{"only DNS resolution was observed"}}
}
```

Add score metadata to JSON and Markdown evidence responses.

- [ ] **Step 4: Run interaction tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/interaction
```

Expected: PASS.

- [ ] **Step 5: Commit evidence scoring**

```bash
git add internal/interaction/evidence.go internal/interaction/evidence_service.go internal/interaction/evidence_service_test.go server/v2_api.go
git commit -m "feat: add explainable evidence scoring"
```

---

## Milestone 3: Scanner Hub Baseline

### Task 3.1: Define Shared Scanner Integration Contract

**Files:**
- Create: `docs/scanner-hub.md`
- Create: `examples/scanner-hub/payload-request.json`
- Create: `examples/scanner-hub/interaction-result.json`
- Modify: `docs/CLI_USAGE.md`

- [ ] **Step 1: Write scanner contract**

Create `docs/scanner-hub.md` with exact sections:

```markdown
# Scanner Hub Integration Contract

GODNSLOG exposes one integration contract for scanners and proxy tools.

## Create Probe

POST `/api/v2/payloads`

Required fields: `template`, `case_id`.
Optional fields: `variables`, `expires_in`, `expected_protocols`, `tool`.

## Wait For Result

GET `/api/v2/interactions?token=<token>&page_size=10`

## Result Formats

CLI integrations should support JSONL and SARIF. Webhook integrations should POST one JSON object per confirmed evidence event.

## Supported Tool Paths

- Nuclei: CLI wrapper and template variables.
- Burp Suite: extension calls the REST API.
- Yakit/Yak: Yak script calls REST API and polls token.
- ZAP: script or add-on calls REST API and polls token.
- xray/rad: CLI or webhook bridge maps scanner events to Case and Payload.
- Postman/Apifox: environment variables and pre-request scripts.
```

- [ ] **Step 2: Add example request and response**

Create `examples/scanner-hub/payload-request.json`:

```json
{
  "template": "ssrf-url",
  "case_id": "case-123",
  "variables": {
    "path": "/callback"
  },
  "expected_protocols": ["dns", "http"],
  "tool": "burp-suite"
}
```

Create `examples/scanner-hub/interaction-result.json`:

```json
{
  "token": "tok123",
  "evidence_strength": "high",
  "confidence": 85,
  "interactions": [
    {"type": "dns", "source_ip": "10.0.0.1"},
    {"type": "http", "source_ip": "10.0.0.1", "path": "/callback"}
  ]
}
```

- [ ] **Step 3: Verify docs**

Run:

```bash
sed -n '1,220p' docs/scanner-hub.md
```

Expected: The supported tool paths include Nuclei, Burp Suite, Yakit/Yak, ZAP, xray/rad, and Postman/Apifox.

- [ ] **Step 4: Commit scanner contract**

```bash
git add docs/scanner-hub.md examples/scanner-hub docs/CLI_USAGE.md
git commit -m "docs: define scanner hub integration contract"
```

### Task 3.2: Add Yakit/Yak and ZAP Minimal Script Examples

**Files:**
- Create: `examples/yak/godnslog-oast.yak`
- Create: `examples/zap/godnslog-oast.js`
- Create: `examples/xray-rad/README.md`

- [ ] **Step 1: Add Yak script skeleton with real API calls**

Create `examples/yak/godnslog-oast.yak`:

```text
// GODNSLOG OAST helper for Yakit/Yak.
// Configure GODNSLOG_API_URL and GODNSLOG_API_KEY in the Yak runtime environment.
apiURL = getenv("GODNSLOG_API_URL")
apiKey = getenv("GODNSLOG_API_KEY")

createPayload = func(caseID, templateName) {
    body = sprintf(`{"case_id":"%s","template":"%s","tool":"yakit"}`, caseID, templateName)
    return http.Post(apiURL + "/payloads", "application/json", body, {"Authorization": "Bearer " + apiKey})
}

waitInteraction = func(token) {
    return http.Get(apiURL + "/interactions?token=" + token + "&page_size=10", {"Authorization": "Bearer " + apiKey})
}
```

- [ ] **Step 2: Add ZAP script example**

Create `examples/zap/godnslog-oast.js`:

```javascript
// GODNSLOG OAST helper for OWASP ZAP scripting.
var apiUrl = java.lang.System.getenv("GODNSLOG_API_URL");
var apiKey = java.lang.System.getenv("GODNSLOG_API_KEY");

function authHeaders() {
  var headers = new java.util.HashMap();
  headers.put("Authorization", "Bearer " + apiKey);
  headers.put("Content-Type", "application/json");
  return headers;
}

function createPayload(caseId, template) {
  var body = JSON.stringify({ case_id: caseId, template: template, tool: "zap" });
  return org.parosproxy.paros.network.HttpSender().sendAndReceive(apiUrl + "/payloads", "POST", authHeaders(), body);
}
```

- [ ] **Step 3: Add xray/rad bridge doc**

Create `examples/xray-rad/README.md` with:

```markdown
# xray/rad Integration

Use GODNSLOG as the private OAST evidence backend for crawler and passive scan workflows.

1. Create a Case for the scan target with `godnslog-cli case create`.
2. Generate payloads with `godnslog-cli payload create --tool xray`.
3. Inject payloads through xray/rad configuration or proxy rules.
4. Poll `godnslog-cli interaction wait --token <token>`.
5. Export evidence with `godnslog-cli report export --case-id <case> --format markdown`.
```

- [ ] **Step 4: Verify files exist**

Run:

```bash
test -f examples/yak/godnslog-oast.yak
test -f examples/zap/godnslog-oast.js
test -f examples/xray-rad/README.md
```

Expected: All commands exit 0.

- [ ] **Step 5: Commit examples**

```bash
git add examples/yak examples/zap examples/xray-rad
git commit -m "docs: add scanner hub script examples"
```

---

## Milestone 4: Agent-Native OAST

### Task 4.1: Add AgentRun Model and Store

**Files:**
- Create: `internal/agentrun/model.go`
- Create: `internal/agentrun/store.go`
- Create: `internal/agentrun/model_test.go`
- Modify: `internal/models/doc.go`

- [ ] **Step 1: Write model test**

Create `internal/agentrun/model_test.go`:

```go
package agentrun

import "testing"

func TestAgentRunDefaultsToRunning(t *testing.T) {
	run := NewAgentRun("agent-1", "case-1", "https://target.example")
	if run.Status != "running" {
		t.Fatalf("expected running status, got %s", run.Status)
	}
	if run.AgentID != "agent-1" || run.CaseID != "case-1" {
		t.Fatalf("unexpected agent run identity: %#v", run)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun
```

Expected: FAIL because the package and constructor do not exist.

- [ ] **Step 3: Implement model**

Create `internal/agentrun/model.go`:

```go
package agentrun

import "time"

type AgentRun struct {
	ID        string    `json:"id" xorm:"pk varchar(64)"`
	AgentID   string    `json:"agent_id" xorm:"varchar(128) index"`
	CaseID    string    `json:"case_id" xorm:"varchar(64) index"`
	Target    string    `json:"target" xorm:"text"`
	Status    string    `json:"status" xorm:"varchar(32) index"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
	UpdatedAt time.Time `json:"updated_at" xorm:"updated"`
}

func NewAgentRun(agentID, caseID, target string) *AgentRun {
	return &AgentRun{
		AgentID: agentID,
		CaseID:  caseID,
		Target:  target,
		Status:  "running",
	}
}
```

Create `internal/agentrun/store.go`:

```go
package agentrun

import "xorm.io/xorm"

type Store struct {
	engine *xorm.Engine
}

func NewStore(engine *xorm.Engine) *Store {
	return &Store{engine: engine}
}

func (s *Store) Create(run *AgentRun) error {
	_, err := s.engine.Insert(run)
	return err
}

func (s *Store) Get(id string) (*AgentRun, error) {
	run := new(AgentRun)
	found, err := s.engine.ID(id).Get(run)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return run, nil
}
```

- [ ] **Step 4: Run agentrun tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun
```

Expected: PASS.

- [ ] **Step 5: Commit AgentRun model**

```bash
git add internal/agentrun internal/models/doc.go
git commit -m "feat: add agent run model"
```

### Task 4.2: Persist MCP create_oast_probe as AgentRun

**Files:**
- Modify: `internal/mcp/server.go`
- Modify: `internal/mcp/server_test.go`
- Modify: `docs/MCP_SERVER_USAGE.md`

- [ ] **Step 1: Extend MCP test**

Extend `TestCreateOASTProbeToolCreatesCaseThenPayload` in `internal/mcp/server_test.go` to assert the returned map includes `agent_run_id` when `agent_id` is provided:

```go
if data["agent_run_id"] == "" {
	t.Fatal("agent_run_id should be populated when agent_id is provided")
}
```

Call the tool with:

```go
"agent_id": "agent-1",
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test -count=1 ./internal/mcp -run TestCreateOASTProbeToolCreatesCaseThenPayload
```

Expected: FAIL because `agent_run_id` is not returned yet.

- [ ] **Step 3: Implement agent run response field**

In `internal/mcp/server.go`, when `agent_id` is present, include:

```go
agentRunID := ""
if agentID, _ := args["agent_id"].(string); agentID != "" {
	agentRunID = agentID + ":" + caseID + ":" + payloadID
}
```

Add `"agent_run_id": agentRunID` to the response map. If `Server` has an AgentRun store configured, persist the run before returning the tool result:

```go
if s.agentRunStore != nil && agentRunID != "" {
	run := agentrun.NewAgentRun(agentID, caseID, target)
	run.ID = agentRunID
	if err := s.agentRunStore.Create(run); err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}
}
```

- [ ] **Step 4: Run MCP tests**

Run:

```bash
GOCACHE=/tmp/gocache go test -count=1 ./internal/mcp
```

Expected: PASS.

- [ ] **Step 5: Commit MCP AgentRun response**

```bash
git add internal/mcp/server.go internal/mcp/server_test.go docs/MCP_SERVER_USAGE.md
git commit -m "feat: return agent run context from oast probe"
```

---

## Milestone 5: Frontend MVP Completion

### Task 5.1: Make Dashboard Data Real

**Files:**
- Modify: `frontend-next/src/app/dashboard/page.tsx`
- Modify: `frontend-next/src/features/interactions/api.ts`
- Modify: `frontend-next/e2e/dashboard.spec.ts`

- [ ] **Step 1: Update E2E expectation**

In `frontend-next/e2e/dashboard.spec.ts`, assert that dashboard stats are loaded from `/api/v2/interactions/stats` rather than hardcoded text:

```ts
await page.route('**/api/v2/interactions/stats', async route => {
  await route.fulfill({
    json: { code: 0, data: { today: 3, total: 10, high_risk: 1 } }
  })
})
await expect(page.getByText('3')).toBeVisible()
```

- [ ] **Step 2: Run spec to verify failure**

Run:

```bash
cd frontend-next
npm run test:e2e -- e2e/dashboard.spec.ts
```

Expected: FAIL if dashboard does not consume the stats API.

- [ ] **Step 3: Implement stats API client and UI binding**

Add an API function:

```ts
export async function getInteractionStats() {
  return apiClient.get('/interactions/stats')
}
```

Use it from the dashboard page and render loading, error, and empty states with existing components.

- [ ] **Step 4: Run frontend checks**

Run:

```bash
cd frontend-next
npm run lint
npm run test:e2e -- e2e/dashboard.spec.ts
```

Expected: PASS.

- [ ] **Step 5: Commit dashboard binding**

```bash
git add frontend-next/src/app/dashboard/page.tsx frontend-next/src/features/interactions/api.ts frontend-next/e2e/dashboard.spec.ts
git commit -m "feat: bind dashboard to interaction stats"
```

### Task 5.2: Complete Case, Payload, Interaction Detail Workflows

**Files:**
- Modify: `frontend-next/src/app/dashboard/cases/[id]/page.tsx`
- Modify: `frontend-next/src/app/dashboard/payloads/[id]/page.tsx`
- Modify: `frontend-next/src/app/dashboard/interactions/[id]/page.tsx`
- Modify: `frontend-next/e2e/cases.spec.ts`
- Modify: `frontend-next/e2e/payloads.spec.ts`
- Modify: `frontend-next/e2e/interactions.spec.ts`

- [ ] **Step 1: Add E2E detail assertions**

Each spec should route a detail API response and assert the page renders a stable field:

```ts
await page.route('**/api/v2/cases/case-1', route => route.fulfill({
  json: { code: 0, data: { id: 'case-1', title: 'SSRF case', status: 'active' } }
}))
await page.goto('/dashboard/cases/case-1')
await expect(page.getByText('SSRF case')).toBeVisible()
```

- [ ] **Step 2: Run detail specs to verify failure**

Run:

```bash
cd frontend-next
npm run test:e2e -- e2e/cases.spec.ts e2e/payloads.spec.ts e2e/interactions.spec.ts
```

Expected: FAIL for any page still using incomplete API logic.

- [ ] **Step 3: Implement details using feature API clients**

Use the existing `frontend-next/src/features/*/api.ts` files to fetch detail endpoints. Render copy buttons for tokens, timeline for interaction history, and status badges for lifecycle fields.

- [ ] **Step 4: Run frontend checks**

Run:

```bash
cd frontend-next
npm run lint
npm run test:e2e -- e2e/cases.spec.ts e2e/payloads.spec.ts e2e/interactions.spec.ts
```

Expected: PASS.

- [ ] **Step 5: Commit detail workflows**

```bash
git add frontend-next/src/app/dashboard/cases frontend-next/src/app/dashboard/payloads frontend-next/src/app/dashboard/interactions frontend-next/e2e
git commit -m "feat: complete core detail workflows"
```

---

## Milestone 6: Workflow and Notifications

### Task 6.1: Replace Workflow Action Placeholders with Executable Actions

**Files:**
- Modify: `internal/workflow/service.go`
- Modify: `internal/workflow/service_test.go`
- Modify: `internal/notification/service.go`

- [ ] **Step 1: Write failing webhook action test**

Add this test to `internal/workflow/service_test.go`:

```go
func TestExecuteWebhookActionPostsInteraction(t *testing.T) {
	var received bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = true
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	service := NewService(nil)
	err := service.ExecuteAction(context.Background(), Action{
		Type: "webhook",
		Config: map[string]string{"url": server.URL},
	}, map[string]interface{}{"token": "tok123"})
	if err != nil {
		t.Fatalf("execute action failed: %v", err)
	}
	if !received {
		t.Fatal("expected webhook server to receive request")
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/workflow -run TestExecuteWebhookActionPostsInteraction
```

Expected: FAIL because webhook execution currently returns without sending a request.

- [ ] **Step 3: Implement HTTP and webhook execution**

Implement action execution with method, URL, headers, JSON body, timeout, and response code validation. Restrict outbound hosts through an allowlist loaded from settings before enabling production use.

- [ ] **Step 4: Run workflow tests**

Run:

```bash
GOCACHE=/tmp/gocache go test ./internal/workflow ./internal/notification
```

Expected: PASS.

- [ ] **Step 5: Commit workflow actions**

```bash
git add internal/workflow internal/notification
git commit -m "feat: execute workflow webhook actions"
```

---

## Milestone 7: Release Readiness

### Task 7.1: Align Documentation with Actual State

**Files:**
- Modify: `DEVELOPMENT_PLAN_2.0.md`
- Modify: `README.md`
- Modify: `README_CN.md`
- Modify: `docs/CLI_USAGE.md`
- Modify: `docs/MCP_SERVER_USAGE.md`

- [ ] **Step 1: Remove inaccurate completion claims**

Change phase labels from broad “completed” status to one of:

```text
Ready
In progress
Planned
Experimental
```

Use “Ready” only when the code has passing tests and documented usage.

- [ ] **Step 2: Add release checklist**

Add this checklist to `DEVELOPMENT_PLAN_2.0.md`:

```markdown
## 2.0 MVP Release Checklist

- Go tests pass with `GOCACHE=/tmp/gocache go test ./...`.
- Frontend lint and build pass.
- User can create Case, generate Payload, trigger HTTP callback, view Interaction, export Evidence.
- CLI can create payload and wait for interaction.
- MCP `create_oast_probe` and `wait_for_interaction` pass focused tests.
- Scanner Hub docs include Nuclei, Burp Suite, Yakit/Yak, ZAP, xray/rad, and CI/CD.
```

- [ ] **Step 3: Verify docs**

Run:

```bash
rg -n "✅ 已完成|伪实现|stub|占位" DEVELOPMENT_PLAN_2.0.md README.md README_CN.md docs
```

Expected: No release-facing claim says a feature is complete without a matching verification note.

- [ ] **Step 4: Commit documentation alignment**

```bash
git add DEVELOPMENT_PLAN_2.0.md README.md README_CN.md docs/CLI_USAGE.md docs/MCP_SERVER_USAGE.md
git commit -m "docs: align 2.0 plan with implementation state"
```

### Task 7.2: Final Full Verification

**Files:**
- Verify entire repository

- [ ] **Step 1: Backend verification**

Run:

```bash
GOCACHE=/tmp/gocache go test ./...
```

Expected: PASS.

- [ ] **Step 2: Frontend verification**

Run:

```bash
cd frontend-next
npm run lint
npm run build
```

Expected: PASS.

- [ ] **Step 3: Focused E2E verification**

Run:

```bash
cd frontend-next
npm run test:e2e -- e2e/login.spec.ts e2e/dashboard.spec.ts e2e/cases.spec.ts e2e/payloads.spec.ts e2e/interactions.spec.ts
```

Expected: PASS.

- [ ] **Step 4: Record verification evidence**

Append the exact command output summary to the release PR description or implementation summary. Include any skipped command and the reason.

---

## Self-Review

- Spec coverage: This plan covers the adjusted positioning, MVP OAST evidence loop, multi-tool Scanner Hub, Agent-native MCP, frontend completion, workflow execution, and documentation truthfulness.
- Placeholder scan: Plan steps avoid indefinite work items; each task has exact files, commands, expected outcomes, and a commit boundary.
- Type consistency: Backend terms use Case, Payload, Interaction, Evidence, APIKey, AgentRun, and Workflow consistently with current repository modules.
- Scope control: SMTP/LDAP/SMB/FTP, HA, marketplace, long-term Canary, and full platform features remain outside the MVP release path until the core DNS/HTTP evidence loop is verifiably stable.
