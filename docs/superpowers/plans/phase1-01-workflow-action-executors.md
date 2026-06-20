# Phase 1.1: Workflow Action Executors

> **Spec ref:** `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §3.1

**Goal:** Implement all 4 workflow action executor stubs (HTTP, DNS, Webhook, Notify) with outbound security constraints.

**Acceptance:**
- All 4 executors implemented with unit tests
- Outbound security: allowlist, SSRF protection, rate limiting
- `go test ./internal/workflow/ -v` passes

---

## Task 1: Outbound Security Module

**Files:** Create `internal/workflow/security.go`, `internal/workflow/security_test.go`

- [ ] **Step 1: Write failing test** — Create `internal/workflow/security_test.go` with tests for `IsURLAllowed` (empty allowlist denies all, matching allowlist, non-matching denied), `IsSSRF` (private IPs blocked, public allowed), `CheckRate` (under limit allowed, over limit blocked, window reset).

- [ ] **Step 2: Verify failure** — `go test ./internal/workflow/ -run "TestIsURL|TestIsSSRF|TestRateLimiter" -v` → FAIL (undefined: NewOutboundSecurity)

- [ ] **Step 3: Implement** — Create `internal/workflow/security.go` with:
  - `OutboundSecurity` struct: allowlist, rateCounts map, rateWindow, rateLimit
  - `NewOutboundSecurity(allowlist []string)` — empty = deny all
  - `IsURLAllowed(rawURL)` — parse URL, match hostname against allowlist (support `"*"` wildcard)
  - `IsSSRF(rawURL)` — block localhost, private ranges (10/8, 172.16/12, 192.168/16, 127/8, 169.254/16) via `net.ParseCIDR`
  - `CheckRate(actionID)` — sliding window counter, 10 req/min default
  - `ValidateURL(actionID, rawURL)` — combine all 3 checks, return error on failure

- [ ] **Step 4: Verify pass** — `go test ./internal/workflow/ -run "TestIsURL|TestIsSSRF|TestRateLimiter" -v`

- [ ] **Step 5: Commit** — `git add internal/workflow/security.go internal/workflow/security_test.go && git commit -m "feat(workflow): add outbound security module with allowlist, SSRF protection, rate limiting"`

---

## Task 2: HTTP Action Executor

**Files:** Modify `internal/workflow/service.go` (imports, struct, `executeHTTPAction`), `internal/workflow/service_test.go`

- [ ] **Step 1: Write failing test** — Add tests: `TestService_ExecuteHTTPAction_Success` (httptest server, verify request sent), `TestService_ExecuteHTTPAction_URLNotAllowed` (evil.com denied), `TestService_ExecuteHTTPAction_NoSecurityConfigured` (nil security = deny all).

- [ ] **Step 2: Verify failure** — `go test ./internal/workflow/ -run TestService_ExecuteHTTPAction -v` → FAIL (SetOutboundSecurity undefined)

- [ ] **Step 3: Implement** — In `service.go`:
  - Add `security *OutboundSecurity` field to `Service` struct
  - Add `SetOutboundSecurity(sec)` method
  - Replace `executeHTTPAction` stub: extract url/method/body/headers from `action.Config`, run `security.ValidateURL`, build `http.Request` with 30s timeout context, send via `http.Client`, limit response to 1MB, return error on status >= 400
  - Add imports: `"bytes"`, `"context"`, `"io"`, `"net/http"`

- [ ] **Step 4: Verify pass** — `go test ./internal/workflow/ -run TestService_ExecuteHTTPAction -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(workflow): implement HTTP action executor with outbound security"`

---

## Task 3: Webhook Action Executor

**Files:** Modify `internal/workflow/service.go` (`executeWebhookAction`), `internal/workflow/service_test.go`

- [ ] **Step 1: Write failing test** — Add tests: `TestService_ExecuteWebhookAction_Success` (verify template rendering `{{.id}}` → interaction ID, `{{.source_ip}}` → IP, custom header forwarded), `TestService_ExecuteWebhookAction_SSRFBlocked` (169.254.169.254 denied even with `"*"` allowlist).

- [ ] **Step 2: Verify failure** — `go test ./internal/workflow/ -run TestService_ExecuteWebhookAction -v` → FAIL (returns nil without making request)

- [ ] **Step 3: Implement** — Replace `executeWebhookAction` stub:
  - Extract url/method/headers/body from config
  - Security check via `ValidateURL`
  - Render body template: replace `{{.id}}`, `{{.type}}`, `{{.source_ip}}`, `{{.token}}`, `{{.domain}}`, `{{.path}}` with interaction fields
  - Send HTTP request with rendered body and headers
  - Add `renderActionTemplate(template, interaction)` helper using `strings.ReplaceAll`
  - Add `"strings"` and `"fmt"` to imports

- [ ] **Step 4: Verify pass** — `go test ./internal/workflow/ -run TestService_ExecuteWebhookAction -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(workflow): implement webhook action executor with template rendering"`

---

## Task 4: DNS Action Executor

**Files:** Modify `internal/workflow/service.go` (`executeDNSAction`), `internal/workflow/service_test.go`

- [ ] **Step 1: Write failing test** — Add tests: `TestService_ExecuteDNSAction_Success` (domain "localhost" type "A"), `TestService_ExecuteDNSAction_MissingDomain` (empty config → error "missing domain"), `TestService_ExecuteDNSAction_UnsupportedType` (type "SRV" → error "unsupported").

- [ ] **Step 2: Verify failure** — `go test ./internal/workflow/ -run TestService_ExecuteDNSAction -v` → FAIL (returns nil)

- [ ] **Step 3: Implement** — Replace `executeDNSAction` stub:
  - Extract domain and type from config (default type "A")
  - Use `net.Resolver{}` with 10s timeout context
  - Support A (LookupHost), TXT (LookupTXT), MX (LookupMX), NS (LookupNS)
  - Return error for unsupported types
  - Add `"net"` to imports

- [ ] **Step 4: Verify pass** — `go test ./internal/workflow/ -run TestService_ExecuteDNSAction -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(workflow): implement DNS action executor"`

---

## Task 5: Notify Action Executor

**Files:** Modify `internal/workflow/service.go` (`executeNotifyAction`), `internal/workflow/service_test.go`

- [ ] **Step 1: Write failing test** — Add tests: `TestService_ExecuteNotifyAction_WebhookChannel` (httptest server, verify message rendered and POST sent), `TestService_ExecuteNotifyAction_MissingChannel` (error), `TestService_ExecuteNotifyAction_UnsupportedChannel` (channel "email" → error "only 'webhook' supported").

- [ ] **Step 2: Verify failure** — `go test ./internal/workflow/ -run TestService_ExecuteNotifyAction -v` → FAIL (returns nil)

- [ ] **Step 3: Implement** — Replace `executeNotifyAction` stub:
  - Extract channel and message template from config
  - Render message with `renderActionTemplate`
  - For channel "webhook": extract `webhook_url`, security check, POST JSON `{"message": "..."}`, 30s timeout
  - For other channels: return error (Phase 1 only supports webhook)
  - Add `"encoding/json"` to imports

- [ ] **Step 4: Verify pass** — `go test ./internal/workflow/ -run TestService_ExecuteNotifyAction -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(workflow): implement notify action executor (webhook channel)"`

---

## Final Verification

```bash
go build ./internal/workflow/...
go test ./internal/workflow/ -v
```
