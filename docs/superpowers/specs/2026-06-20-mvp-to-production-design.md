# GODNSLOG 2.0: From MVP to Production — Multi-Phase Design

> Date: 2026-06-20
> Status: Revised — review issues addressed, pending final approval
> Approach: Bottom-up (Phase A) — backend first, then core loop, then frontend, then agent, then platform

---

## 1. Problem Statement

GODNSLOG 2.0 has severe engineering issues despite all phases being marked "complete":

- **Scope creep**: Jumped from MVP directly to platform-level features (Phase 0-16 all marked done)
- **Low implementation quality**: Workflow action executors are `TODO: Implement`; Rule engine CIDR matching is a placeholder; many features are stub/placeholder
- **1.0 feature loss**: Standard DNS resolution (A/CNAME/TXT/MX/NS) and xip IP encoding exist in 1.0 code but are not exposed in v2 API or frontend; multi-user role system only has APIKey
- **Frontend is demo-level**: No edit pages, no real-time updates, no ErrorBoundary, TanStack Query barely used, React Hook Form + Zod installed but unused, i18n coverage extremely shallow

### 1.1 Evidence from Code Audit

**Backend stubs found**:
- `internal/workflow/service.go:150-173` — 4 × `TODO: Implement` (HTTP/DNS/Webhook/Notify actions)
- `internal/rule/action.go:34,200,222,243` — placeholder for noise filtering, email notification, tag action, report generation
- `internal/rule/engine.go:266` — CIDR matching uses `strings.HasPrefix` instead of `net.ParseCIDR`
- `internal/listener/ftp.go:177` — `502 Command not implemented`
- `server/v2_api.go` — no DNS record management, no xip API, no rebinding v2 API

**1.0 features lost in v2**:
- `server/dnsserver.go:511` — `parseXip()` exists but not exposed via v2 API
- `server/webserver.go:304-306` — DNS resolve record management exists in v1 API only
- `server/webserver.go:310-316` — User management (add/delete/list) exists in v1 API only

**Frontend gaps**:
- No WebSocket/SSE (real-time updates) — 0 matches for `EventSource|useWebSocket`
- No React Hook Form usage — 0 matches for `useForm|zodResolver`
- No ErrorBoundary — 0 matches
- TanStack Query — only 3 hooks files use it (cases, payloads, interactions)
- `e2e/payloads.spec.ts:44` — `test.skip()` for payload detail page
- 260 `dark:` CSS classes but no systematic theme switching

### 1.2 Backend Status

- `go build ./...` — passes
- `go test ./...` — all packages pass (22 packages with tests)
- V2 API routes cover: Cases, Payloads, Interactions, APIKeys, Notifications, Users, Marketplace, Rules, Evidence, Audit, Canary, Rebinding, Listeners, Settings, Scanner Hub, Agent Runs
- Evidence generation (`v2GenerateEvidence`) — actually implemented (not stub)
- MCP Server — calls v2 API via HTTP (not mock), but does not implement MCP protocol (stdio/SSE/Streamable HTTP)

---

## 2. Strategy: Bottom-Up Phased Approach

**Rationale**: Backend stubs are the biggest risk — building frontend on fake APIs is wasted effort. 1.0 feature loss is P0 blocker — cannot replace 1.0 without them. Frontend overhaul is better done once after APIs are stable.

```
Phase 1: Backend Realization — Stub cleanup & API gap filling
Phase 2: Core Loop Completion — MVP feature gap filling
Phase 3: Frontend Production — Deep overhaul
Phase 4: Agent & Scanner Integration — Advanced capabilities
Phase 5: Platform — Enterprise-grade
```

### Dependency Graph

```
Phase 1 ──→ Phase 2 ──→ Phase 3
    │            │
    │            └──→ Phase 4
    │                      │
    └──────────────────────┴──→ Phase 5
```

Phase 4 depends on Phase 1-2 (stable APIs + core features). Phase 5 depends on all prior phases.

**Phase 4 vs Phase 3 sequencing**: Phase 4 backend/agent work (MCP protocol, notification channels, CLI) can start once Phase 2 is stable. Phase 4 frontend pages (Scanner Hub UI, Agent Run UI) should be implemented **after** Phase 3 patterns (TanStack Query, ErrorBoundary, RHF+Zod) are established, to avoid building new pages on the old data-fetching pattern. In practice, Phase 4 backend and Phase 3 can run in parallel; Phase 4 frontend follows Phase 3.

---

## 3. Phase 1: Backend Realization — Stub Cleanup & API Gap Filling

**Goal**: Make existing APIs truly usable, fill 1.0 feature gaps, provide reliable foundation for frontend.

### 3.1 Workflow Action Executor Implementation

**Current state**: `internal/workflow/service.go:150-173` has 4 `TODO: Implement` stubs.

**Implementation**:
- `executeHTTPAction` — execute custom HTTP request on hit, using `net/http.Client` with 30s timeout and domain allowlist
- `executeDNSAction` — trigger DNS operations on hit
- `executeWebhookAction` — forward to configured Webhook URL with custom Header/Body template rendering
- `executeNotifyAction` — trigger notification channels via `internal/notification` package

**Outbound Action Security Constraints**:
- **Allowlist configuration**: stored in system settings (`/api/v2/settings/security`), managed by Super/Admin only. Per-action URL must match allowlist; no per-action override.
- **Default behavior**: deny all outbound requests if allowlist is empty or not configured.
- **SSRF protection**: block private IP ranges (10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16, 127.0.0.0/8, 169.254.0.0/16), localhost, and cloud metadata endpoints (169.254.169.254).
- **Rate limiting**: max 10 outbound requests per action per minute.
- **All outbound requests**: 30s timeout, max 1MB response body.
- **All actions**: audit log entry with action type, target, result, duration.
- **Async queue**: exponential backoff retry (max 3 retries, base delay 1s, max delay 30s).

**Files affected**:
- `internal/workflow/service.go` — implement 4 action executors
- `internal/workflow/service_test.go` — add tests for each executor
- `internal/notification/` — verify/extend notification channel implementations

### 3.2 Rule Engine Completion

**Current state**: `internal/rule/action.go` and `internal/rule/engine.go` have placeholders.

**Implementation**:
- `matchCIDR` — replace `strings.HasPrefix` with `net.ParseCIDR` + `net.Contains`
- `executeTagAction` — update Interaction tags in database via `internal/interaction` service
- `executeReport` — integrate with Evidence generation system (`internal/evidencehub`)
- `sendEmailNotification` — **Won't do (by design)**: per `doc/2.0-Requirement.md` Q2, SMTP is not used for notifications. Keep returning `fmt.Errorf("email notification not implemented")`.
- `DiscardNoise` — integrate with `internal/interaction/noise_filter.go`

**Files affected**:
- `internal/rule/engine.go` — fix CIDR matching
- `internal/rule/action.go` — implement tag/report/noise actions
- `internal/rule/engine_test.go` — add CIDR matching tests
- `internal/rule/action_test.go` — add action execution tests

### 3.3 1.0 Feature v2 API Exposure

**Current state**: 1.0 core features exist in `server/dnsserver.go` and `server/webserver.go` but v2 API has no exposure.

**New v2 API endpoints**:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v2/dns/records` | GET | List DNS resolve records (A/CNAME/TXT/MX/NS) |
| `/api/v2/dns/records` | POST | Create DNS resolve record |
| `/api/v2/dns/records/:id` | PUT | Update DNS resolve record |
| `/api/v2/dns/records/:id` | DELETE | Delete DNS resolve record |
| `/api/v2/dns/xip/:ip` | GET | Query xip IP encoding (verify 3 encoding formats: dotted decimal, hex, binary) |
| `/api/v2/dns/rebinding/rules` | GET | List rebinding rules (bridge to existing `internal/rebinding` package) |
| `/api/v2/dns/rebinding/rules` | POST | Create rebinding rule |
| `/api/v2/dns/rebinding/rules/:id` | DELETE | Delete rebinding rule |
| `/api/v2/users` | POST | Create user (admin only) |
| `/api/v2/users/:id` | PUT | Update user (role assignment) |
| `/api/v2/users/:id` | DELETE | Delete user (admin only, cannot delete super) |

**Implementation notes**:
- DNS record management wraps existing `TblResolve` model and `getResolveRecord`/`setResolveRecord`/`delResolveRecord` handlers
- xip endpoint is read-only verification — takes IP, returns encoded domain forms
- Rebinding rules bridge to `internal/rebinding` package (already has data model and API)
- User management wraps existing `addUser`/`setUser`/`delUser` handlers with v2 response format

**Files affected**:
- `server/v2_api.go` — add new route registrations and handlers
- `server/v2_api_test.go` — add tests for new endpoints
- `internal/rebinding/` — verify service layer completeness

### 3.4 Data Model Unification (Design + Migration Script)

**Decision** (per requirement doc Q3): Approach A — dual-write synchronization.

**Phase 1 scope**: Design the dual-write contract and write the migration script. Actual dual-write landing in Phase 2 (§4.6).

**Design**:
- DNS handler in `server/dnsserver.go`: after writing to `TblDns`, also write to `Interaction` table
- HTTP handler in `server/webserver.go`: after writing to `TblHttp`, also write to `Interaction` table
- V2 API queries `Interaction` table exclusively for unified history

**Interaction write fields**:
- `type`: "dns" or "http"
- `token`: extracted from subdomain
- `source_ip`: from query/request
- `domain`: full queried domain
- `raw_data`: JSON of original record
- `timestamp`: event time
- `case_id` / `payload_id`: auto-attributed via token lookup

**Migration script** (`migration/sync.go`):
- One-time script to import historical `TblDns`/`TblHttp` records into `Interaction` table
- Supports `--dry-run` flag for testing on a copy of production data
- Batch processing (1000 records per batch) with progress reporting
- Idempotent: checks if Interaction records already exist for the same timestamp+token before inserting

**Files affected (Phase 1)**:
- `migration/sync.go` — new migration script with dry-run support
- `internal/interaction/service.go` — add batch import method
- `migration/sync_test.go` — tests for migration logic

### 3.5 APIKey Security Hardening

**Current state**: APIKey stores `key` field in plaintext.

**Implementation**:
- `CreateAPIKey`: generate plaintext key, return to user once, store bcrypt hash
- `GetAPIKey`/`ListAPIKeys`: return only key prefix (first 8 chars + `...`)
- `AuthenticateAPIKey`: bcrypt compare incoming key against stored hash

**Migration Algorithm** (concrete):
1. Add `key_hash` column to `tbl_api_keys` via XORM auto-migration (`internal/models/apikey.go`)
2. On authentication, lookup APIKey by key prefix (first 8 chars):
   - If `key_hash` is empty (legacy plaintext key): compare plaintext `key` field; if match, compute bcrypt hash, store in `key_hash`, clear `key` field (invalidate plaintext)
   - If `key_hash` is non-empty: bcrypt compare incoming key against `key_hash`
3. If no match by prefix, return 401 immediately (no full-table scan)
4. Key rotation: user can voluntarily rotate via `POST /api/v2/apikeys/:id/rotate` — generates new key, returns plaintext once, stores new bcrypt hash
5. Migration is complete when all rows have non-empty `key_hash`; a background job can be run to force-migrate remaining plaintext keys (optional)

**Files affected**:
- `internal/models/apikey.go` — add `KeyHash` field, update XORM tags
- `server/v2_api.go` — update `v2CreateAPIKey`, `v2ListAPIKeys`, add `v2RotateAPIKey` handler
- `internal/auth/middleware.go` — update API key authentication with migration logic

### 3.6 Phase 1 Acceptance Criteria

- [ ] All 4 workflow action executors implemented with tests
- [ ] Rule engine CIDR matching uses `net.ParseCIDR`
- [ ] Rule engine tag/report/noise actions implemented
- [ ] v2 API exposes DNS record CRUD
- [ ] v2 API exposes xip encoding query
- [ ] v2 API exposes user management CRUD
- [ ] DNS/HTTP handlers dual-write to Interaction table
- [ ] Migration script exists and is tested
- [ ] APIKey stored as bcrypt hash
- [ ] `go build ./...` passes
- [ ] `go test ./...` passes with new tests

---

## 4. Phase 2: Core Loop Completion — MVP Feature Gap Filling

**Goal**: Fill MVP-required frontend pages so the core OAST loop is usable end-to-end.

### 4.1 Case Management Complete Loop

**Current state**: `cases/page.tsx` (230 lines) list, `cases/[id]/page.tsx` (170 lines) detail, no edit.

**Additions**:
- **Case edit**: inline edit or dedicated edit route, supports title/description/target/tags/status modification
- **Case creation**: wire to React Hook Form + Zod validation (title required, target format, tags array)
- **Case stats**: display payload_count/interaction_count/hit_payload_count from `v2GetCaseStats`
- **Case associations**: detail page shows related Payloads and Interactions (API: `/:id/payloads`, `/:id/interactions`)

**Files affected**:
- `frontend-next/src/app/dashboard/cases/[id]/page.tsx` — add edit mode
- `frontend-next/src/app/dashboard/cases/page.tsx` — wire form validation
- `frontend-next/src/features/cases/hooks/use-cases.ts` — add update mutation
- `frontend-next/src/features/cases/schemas/case-schema.ts` — new Zod schema

### 4.2 Payload Management Complete Loop

**Current state**: `payloads/page.tsx` (302 lines) list, `payloads/[id]/page.tsx` (220 lines) exists but E2E skip, `payloads/new/page.tsx` (518 lines) wizard.

**Additions**:
- **Payload detail**: display Token, rendered payload, variables, status, related Interactions, copy button
- **Payload preview**: wire to `v2PreviewPayload` API for real-time template variable rendering
- **Payload revoke**: wire to `v2RevokePayload`, status changes to archived
- **Payload interactions**: new API `GET /api/v2/payloads/:id/interactions` (currently missing — add to Phase 1 §3.3 backend API list)
- **Remove E2E skip**: `e2e/payloads.spec.ts:44` — un-skip after detail page is functional

**Files affected**:
- `frontend-next/src/app/dashboard/payloads/[id]/page.tsx` — make functional
- `server/v2_api.go` — add `GET /payloads/:id/interactions` endpoint (backend work, can be done in Phase 1)
- `e2e/payloads.spec.ts` — remove `test.skip()`
- `frontend-next/src/features/payloads/hooks/use-payloads.ts` — add preview/revoke mutations

### 4.3 Interaction Complete Loop

**Current state**: `interactions/page.tsx` (539 lines) list, `interactions/[id]/page.tsx` (189 lines) detail page.

**Additions**:
- **Detail drawer**: change from page navigation to side sheet (shadcn/ui Sheet component)
- **Attribution display**: show related Case and Payload info in detail
- **Filter enhancement**: add time range, protocol, token search filters
- **Export**: wire to `v2ExportInteractions` API, support CSV/JSON export

**Files affected**:
- `frontend-next/src/app/dashboard/interactions/page.tsx` — add drawer, filters, export
- `frontend-next/src/app/dashboard/interactions/[id]/page.tsx` — keep as standalone route for direct URL access
- `frontend-next/src/features/interactions/hooks/use-interactions.ts` — add export mutation

### 4.4 Evidence Export

**Current state**: `evidence/page.tsx` (381 lines), `evidence-summary/page.tsx` (445 lines). Backend `v2GenerateEvidence` is implemented (not stub).

**Verification & additions**:
- Verify `v2GenerateEvidence` produces correct Markdown/JSON output
- Frontend: format selection (Markdown/JSON), redaction options UI
- Evidence timeline visualization: Payload creation → deployment → DNS query → HTTP request → notification

**Files affected**:
- `frontend-next/src/app/dashboard/evidence/page.tsx` — add format selection and redaction UI
- `frontend-next/src/app/dashboard/evidence-summary/page.tsx` — verify data accuracy

### 4.5 User Management & System Settings

**Current state**: `users/page.tsx` (215 lines), `settings/page.tsx` (273 lines).

**Additions**:
- **User management**: wire to v2 API `GET /api/v2/users` (existing) + new POST/PUT/DELETE (Phase 1)
- **Role assignment**: Super/Admin/Normal/Guest role management UI
- **System settings**: domain config, listen address, notification channel config wired to v2 Settings API

**Files affected**:
- `frontend-next/src/app/dashboard/users/page.tsx` — add create/edit/delete UI
- `frontend-next/src/app/dashboard/settings/page.tsx` — wire to v2 Settings API
- `frontend-next/src/features/settings/` — new feature module if needed

### 4.6 Data Model Unification Landing

Phase 1 (§3.4) designed the dual-write contract and wrote the migration script. This phase lands the actual dual-write in the DNS/HTTP handlers:

- `server/dnsserver.go`: add Interaction table write in DNS handler
- `server/webserver.go`: add Interaction table write in HTTP record handler
- Run `migration/sync.go` on a staging copy of production data, verify, then run on production

**Files affected**:
- `server/dnsserver.go` — add Interaction write
- `server/webserver.go` — add Interaction write

### 4.7 Phase 2 Acceptance Criteria

- [ ] Case edit page functional with form validation
- [ ] Case detail shows stats and associated Payloads/Interactions
- [ ] Payload detail page functional, E2E test un-skipped and passing
- [ ] Payload preview and revoke work end-to-end
- [ ] Interaction detail drawer works from list
- [ ] Interaction filters include time range, protocol, token search
- [ ] Interaction export (CSV/JSON) works
- [ ] Evidence export with format selection and redaction
- [ ] User management CRUD functional
- [ ] System settings wired to v2 API
- [ ] DNS/HTTP dual-write to Interaction table
- [ ] All E2E tests pass

---

## 5. Phase 3: Frontend Production — Deep Overhaul

**Goal**: Elevate frontend from "works" to "production-ready" with maintainable component architecture and data flow.

### 5.1 TanStack Query Full Integration

**Current state**: Only 3 hooks files use TanStack Query. Most pages use manual `useState` + `useEffect` + `fetch`.

**Changes**:
- **QueryClientProvider**: inject in root layout, configure global staleTime (30s lists, 60s detail), refetchOnWindowFocus
- **Unified hooks layer**: each `src/features/*/hooks/` directory has complete query/mutation hooks
- **Cache strategy**: mutations auto-invalidate related queries
- **Loading/error states**: unified handling replacing per-page manual state

**Files affected**:
- `frontend-next/src/app/layout.tsx` — add QueryClientProvider
- `frontend-next/src/features/*/hooks/` — all feature modules
- `frontend-next/src/lib/query-client.ts` — new QueryClient configuration

### 5.2 React Hook Form + Zod Form Validation

**Current state**: Dependencies installed, zero usage in code.

**Changes**:
- **Case create/edit form**: title required, target format, tags array validation
- **Payload create wizard**: template_id required, variables dynamic field validation
- **Login form**: username/password non-empty
- **Settings form**: domain format, port range
- **Unified Zod schemas**: `src/features/*/schemas/` matching backend API contracts

**Files affected**:
- `frontend-next/src/features/*/schemas/` — new schema files per feature
- All form components — replace manual validation with RHF + Zod

### 5.3 ErrorBoundary & Error Handling

**Current state**: No ErrorBoundary exists.

**Changes**:
- **Global ErrorBoundary**: wraps AppShell content area, catches render errors, shows fallback UI
- **API error handling**: TanStack Query global onError — 401 redirect to login, 500 show toast
- **Page-level ErrorBoundary**: critical pages have independent boundaries to prevent full-page white screen

**Files affected**:
- `frontend-next/src/components/error-boundary.tsx` — new
- `frontend-next/src/app/dashboard/layout.tsx` — wrap with ErrorBoundary
- `frontend-next/src/lib/query-client.ts` — global error handler

### 5.4 Real-Time Updates (SSE)

**Current state**: No WebSocket/SSE, pure static loading.

**Changes**:
- **Backend SSE endpoint**: `GET /api/v2/interactions/stream` — Server-Sent Events pushing new Interactions
- **Frontend EventSource hook**: `useInteractionStream()` subscribes to SSE, new data auto-appends to list
- **Dashboard real-time**: recent hits display updates without manual refresh
- **Interaction list real-time**: new hits auto-appear at top
- **Fallback**: if SSE connection fails, fall back to 30s polling

**Files affected**:
- `server/v2_api.go` — add SSE endpoint
- `frontend-next/src/features/interactions/hooks/use-interaction-stream.ts` — new
- `frontend-next/src/app/dashboard/page.tsx` — wire real-time updates
- `frontend-next/src/app/dashboard/interactions/page.tsx` — wire real-time updates

### 5.5 i18n Unification

**Current state**: i18n coverage extremely shallow, most pages mix English and Chinese hardcoded text.

**Changes**:
- **Translation files**: `src/i18n/locales/zh.json` and `en.json` cover all page text
- **Page text unified**: all hardcoded text replaced with `t('key')` calls
- **Language switcher**: TopBar switcher wired to Zustand store, persisted to localStorage
- **Backend i18n**: API error messages support Accept-Language header

**Files affected**:
- `frontend-next/src/i18n/locales/zh.json` — expand
- `frontend-next/src/i18n/locales/en.json` — expand
- All page components — replace hardcoded text
- `frontend-next/src/components/app-shell/topbar.tsx` — language switcher

### 5.6 Dark Mode Completion

**Current state**: 260 `dark:` CSS classes but no systematic theme switching.

**Changes**:
- **Theme Provider**: Zustand store manages theme state, persisted to localStorage
- **System theme detection**: `prefers-color-scheme` auto-detection
- **Theme toggle UI**: TopBar toggle button
- **CSS variables**: land the CSS variable system from `doc/2.0-frontend-implementation-gap.md`
- **Flash elimination**: inject theme class during SSR to avoid hydration mismatch

**Files affected**:
- `frontend-next/src/stores/theme-store.ts` — new
- `frontend-next/src/app/layout.tsx` — theme class injection
- `frontend-next/src/components/app-shell/topbar.tsx` — theme toggle
- `frontend-next/tailwind.config.ts` — CSS variable mapping

### 5.7 Component Library Deepening

**Current state**: DataTable, CopyButton, StatusBadge, EmptyState, Timeline exist but underused.

**Changes**:
- **DataTable enhancement**: support sorting, pagination, batch operations, column config
- **LoadingSkeleton**: page loading skeleton replacing spinner
- **ErrorBoundary component**: reusable error boundary
- **Toast/Notification**: unified operation feedback via shadcn/ui Toast
- **Usage**: all list pages use DataTable, all status displays use StatusBadge

**Files affected**:
- `frontend-next/src/components/data-table.tsx` — enhance
- `frontend-next/src/components/loading-skeleton.tsx` — new
- `frontend-next/src/components/error-boundary.tsx` — new
- All list pages — adopt DataTable

### 5.8 Phase 3 Acceptance Criteria

- [ ] QueryClientProvider in root layout, all data fetching via TanStack Query
- [ ] All forms use React Hook Form + Zod
- [ ] Global ErrorBoundary catches render errors
- [ ] API errors: 401 → login redirect, 500 → toast
- [ ] SSE endpoint pushes new Interactions
- [ ] Dashboard and Interaction list update in real-time
- [ ] All page text uses i18n keys
- [ ] Dark mode toggle works without flash
- [ ] DataTable supports sorting, pagination, batch operations
- [ ] All E2E tests pass

---

## 6. Phase 4: Agent & Scanner Integration — Advanced Capabilities

**Goal**: Make GODNSLOG a usable OAST backend for scanners and AI Agents, realizing existing stubs.

### 6.1 MCP Protocol Compliance

**Current state**: `internal/mcp/server.go` calls v2 API via HTTP but does not implement MCP protocol.

**Decomposed into 3 deliverables** (per review §3.10):

**Deliverable 6.1a — Transport Layer**:
- Implement `POST /mcp` endpoint with Streamable HTTP transport
- Support MCP methods: `initialize`, `notifications/initialized`
- JSON-RPC 2.0 request/response handling
- Files: `internal/mcp/transport.go` (new), `internal/mcp/transport_test.go`

**Deliverable 6.1b — Session Layer**:
- MCP session ID generation (UUID), state tracking, timeout cleanup (30min idle)
- Session store (in-memory for MVP, Redis for HA in Phase 5)
- Files: `internal/mcp/session.go` (new), `internal/mcp/session_test.go`

**Deliverable 6.1c — Tool Registration & Permission Gating**:
- Register existing 7 tools (createCase, createPayload, listInteractions, waitForInteraction, summarizeEvidence, exportReport, createOastProbe) as MCP tools via `tools/list` and `tools/call`
- Maintain existing APIKey scope check and risk tolerance logic
- Audit logging: all MCP tool calls recorded
- Files: `internal/mcp/server.go` — refactor to use new transport/session, `cmd/mcp-server/main.go`

**Note**: If Phase 4 runs long, Deliverable 6.1c can be deferred to early Phase 5 without blocking other Phase 4 work.

### 6.2 Workflow Action Executor Completion

Phase 1 implemented basic action executors; this phase completes them:

- **Notification channels**: WeChat Work/Feishu/DingTalk/Slack/Discord/Telegram/Webhook actual sending (via `internal/notification` package)
- **Async queue**: action execution in queue with exponential backoff retry
- **Custom HTTP response**: return custom status code/Header/Body/redirect on hit (SCA-02)
- **Action execution log**: each execution records result, duration, retry count

**Files affected**:
- `internal/workflow/service.go` — complete action executors
- `internal/notification/` — verify/extend channel implementations
- `internal/workflow/queue.go` — new async queue

### 6.3 Scanner Hub Realization

**Current state**: `scanner-hub/page.tsx` (560 lines) and `internal/scannerhub/service.go` exist but integration depth insufficient.

**Changes**:
- **Nuclei deep integration**: private OAST backend config, template examples, JSONL/SARIF output, scan result → Interaction auto-association
- **Scanner Run complete flow**: create Run → generate integration package → execute scan → backfill results → associate Interactions
- **Frontend Scanner Hub**: adapter list, Run creation wizard, Run detail, results display
- **CI/CD gate**: high-risk hits as gate, GitHub Actions/GitLab CI/Jenkins examples

**Files affected**:
- `internal/scannerhub/service.go` — complete integration flows
- `frontend-next/src/app/dashboard/scanner-hub/page.tsx` — make functional
- `frontend-next/src/app/dashboard/scanner-hub/[id]/page.tsx` — Run detail
- `examples/ci/` — update CI/CD examples

### 6.4 CLI Tool Completion

**Current state**: `cmd/cli/` and `cli/` have basic structure.

**Changes**:
- **CLI Case management**: create/list/detail/close
- **CLI Payload management**: generate/list/revoke/preview
- **CLI Interaction polling**: timeout, filter, JSON output
- **CLI report export**: Markdown/JSON
- **CLI Scanner Run**: initiate scan from CLI

**Files affected**:
- `cmd/cli/main.go` — add subcommands
- `cli/` — implement command logic

### 6.5 Agent Run Complete Loop

**Current state**: `agent-runs/page.tsx` (385 lines), `agent-runs/[id]/page.tsx` (1149 lines), backend API complete.

**Verification & additions**:
- **Agent Run lifecycle**: create → running → waiting → completed/failed, state transitions complete
- **Review Queue**: frontend shows pending review Agent Runs
- **Follow-up Action**: create follow-ups, view history
- **Evidence package export**: Review Package export and delivery

**Files affected**:
- `frontend-next/src/app/dashboard/agent-runs/page.tsx` — verify and fix
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx` — verify and fix

### 6.6 Phase 4 Acceptance Criteria

- [ ] MCP protocol Streamable HTTP transport implemented
- [ ] MCP tools callable via standard MCP client
- [ ] All notification channels (WeChat/Feishu/DingTalk/Slack/Discord/Telegram/Webhook) send actual messages
- [ ] Async action queue with retry works
- [ ] Custom HTTP response on hit works
- [ ] Nuclei integration: scan → Interaction association works end-to-end
- [ ] Scanner Run complete flow functional in frontend
- [ ] CLI supports Case/Payload/Interaction/Report/Scanner operations
- [ ] Agent Run lifecycle complete in frontend
- [ ] All E2E tests pass

---

## 7. Phase 5: Platform — Enterprise-Grade

**Goal**: Enterprise-level long-term monitoring, multi-protocol support, and production-grade deployment.

**Split into two sub-phases** (per review §3.5): Phase 5a (multi-protocol + security) and Phase 5b (HA + marketplace). This reflects the high risk and security review requirements of real protocol listeners.

### Phase 5a: Multi-Protocol Listeners & Security (3-4 weeks)

### 7.1 Multi-Protocol Listeners

**Current state**: `internal/listener/` has SMTP, LDAP, SMB, FTP code files but as data model stubs.

**Implementation**:
- **SMTP Listener**: `internal/listener/smtp.go` (6077 bytes) — actual SMTP server listening, records mail interactions (for email user sniffening, not notification)
- **LDAP Listener**: `internal/listener/ldap.go` (8687 bytes) — actual LDAP server listening, records query interactions
- **SMB Listener**: `internal/listener/smb.go` (4985 bytes) — SMB protocol listening
- **FTP Listener**: `internal/listener/ftp.go` (4617 bytes) — FTP protocol listening (currently has `502 Command not implemented` placeholder)
- **Unified Interaction write**: all protocol hits write to Interaction table with auto-attribution
- **Frontend Listener management**: start/stop/configure each protocol listener

**Files affected**:
- `internal/listener/smtp.go` — implement actual SMTP server
- `internal/listener/ldap.go` — implement actual LDAP server
- `internal/listener/smb.go` — implement actual SMB server
- `internal/listener/ftp.go` — implement actual FTP server
- `frontend-next/src/app/dashboard/settings/page.tsx` — listener management UI

### 7.2 Canary Complete Version

**Current state**: `internal/canary/` has data model and API, detection logic incomplete.

**Changes**:
- **Multi-type tokens**: DNS/HTTP/document/config/CI variable/object storage/email Canary token types
- **Detection logic**: token access triggers alert, records interaction
- **Supply chain leak monitoring**: document/config token detection after external leak
- **Frontend Canary management**: `canary/page.tsx` (326 lines) — create/deploy/alert viewing
- **Tiered notification**: Canary hits trigger different channels by risk level

**Files affected**:
- `internal/canary/service.go` — complete detection logic
- `frontend-next/src/app/dashboard/canary/page.tsx` — make functional

### 7.3 Rebinding Lab Visualization

**Current state**: `internal/rebinding/` has data model and API, `rebinding/page.tsx` (213 lines).

**Changes**:
- **Visual configuration**: frontend configures first/subsequent resolution IP, TTL, hit conditions
- **Session tracking**: `internal/rebinding/` session tracking logic realized
- **5 predefined scenarios**: one-click create common rebinding scenarios
- **Security controls**: Super/Admin only, telemetry on by default (per Q6)
- **Interaction marking**: rebinding requests marked in Interaction

**Files affected**:
- `internal/rebinding/service.go` — complete session tracking
- `frontend-next/src/app/dashboard/rebinding/page.tsx` — visual config UI

### 7.4 Data Retention & Archive

**Current state**: `internal/retention/` has basic code.

**Changes**:
- **Auto-cleanup**: auto-clean expired Interaction/Payload/Case by time/count
- **Archive**: closed Cases archived, not in main query
- **Compliance retention**: configurable retention period
- **Frontend config**: data retention policy in system settings

**Files affected**:
- `internal/retention/service.go` — complete cleanup logic
- `frontend-next/src/app/dashboard/settings/page.tsx` — retention config UI

### 7.5 High Availability Deployment

**Current state**: `internal/ha/` has basic code.

**Changes**:
- **Multi-instance**: stateless service, session/cache via Redis
- **Database cluster**: MySQL/PostgreSQL master-slave support
- **Load balancing**: health check endpoint, graceful shutdown
- **Deployment docs**: Docker Compose/K8s templates

**Files affected**:
- `internal/ha/` — complete HA logic
- `deploy/` — new deployment templates

### 7.6 Plugin/Template Marketplace

**Current state**: `internal/marketplace/` has basic code, `marketplace/page.tsx` (220 lines).

**Changes**:
- **Template sharing**: Payload template upload/download/rating
- **Plugin extension**: define plugin interface, support third-party extensions
- **Community-driven**: open-source community contributes templates and plugins
- **Frontend marketplace**: browse/search/install templates and plugins

**Files affected**:
- `internal/marketplace/service.go` — complete marketplace logic
- `frontend-next/src/app/dashboard/marketplace/page.tsx` — make functional

### Phase 5a Acceptance Criteria

- [ ] SMTP/LDAP/SMB/FTP listeners actually listen and record interactions
- [ ] Canary tokens detect access and trigger alerts
- [ ] Rebinding Lab visual configuration works
- [ ] Data retention auto-cleanup works
- [ ] Security review passed for all protocol listeners
- [ ] All tests pass

### Phase 5b: HA, Deployment & Marketplace (2-3 weeks)

### 7.7 Phase 5b Acceptance Criteria

- [ ] Multi-instance deployment with Redis cache works
- [ ] Marketplace browse/search/install works
- [ ] Docker Compose/K8s deployment templates validated
- [ ] All tests pass

---

## 8. Phase Summary

| Phase | Goal | Est. Effort | Dependencies |
|-------|------|-------------|--------------|
| **Phase 1** | Backend Realization — Stub cleanup & API gap filling | 2-3 weeks | None |
| **Phase 2** | Core Loop Completion — MVP feature gap filling | 3-4 weeks | Phase 1 |
| **Phase 3** | Frontend Production — Deep overhaul | 2-3 weeks | Phase 2 |
| **Phase 4** | Agent & Scanner Integration — Advanced capabilities | 3-4 weeks | Phase 1-2 |
| **Phase 5a** | Multi-Protocol Listeners & Security | 3-4 weeks | Phase 1-4 |
| **Phase 5b** | HA, Deployment & Marketplace | 2-3 weeks | Phase 5a |

**Total estimated effort**: 16-23 weeks

---

## 9. Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Data model unification breaks existing 1.0 data | Medium | High | Dual-write approach, migration script tested on copy of production data first; dry-run mode |
| MCP protocol implementation complexity | High | Medium | Decomposed into 3 deliverables (transport/session/tools); start with Streamable HTTP; 6.1c can defer to Phase 5 if needed |
| Frontend overhaul introduces regressions | High | Medium | E2E tests as safety net; **incremental feature-by-feature migration**, not big-bang rewrite |
| Multi-protocol listener security exposure | Medium | High | Default disabled, explicit config required, security audit; **all listeners run in sandboxed process or network namespace by default** |
| Scope creep during implementation | High | High | **Change control: no new endpoints or features mid-phase without updating this spec and acceptance criteria** |

---

## 10. 1.0 to 2.0 Upgrade Path

Existing 1.0 deployments need a clear migration path:

### 10.1 Database Schema Migration
- XORM auto-migration handles additive schema changes (new columns, new tables)
- `migration/sync.go` (Phase 1) handles historical data import from `TblDns`/`TblHttp` to `Interaction` table
- Run `migration/sync.go --dry-run` on a copy of production data first, verify row counts, then run on production with `--batch-size=1000`

### 10.2 API Consumer Migration
- V1 API (`/api/v1/*`, `/auth`, `/data`, `/setting`, `/admin`, `/payload`) remains available and functional throughout 2.0
- V2 API (`/api/v2/*`) runs in parallel; no v1 endpoint is removed until 2.0 is stable
- Deprecation timeline: v1 API marked deprecated in 2.0 release, removed in 2.1 release
- V1 API consumers can migrate at their own pace; no forced cutover

### 10.3 Frontend Migration
- 1.0 frontend (`frontend/`) is directly deprecated per requirement doc Q3 conclusion
- 2.0 frontend (`frontend-next/`) is the only supported UI from 2.0 release
- User management, documentation, and API features from 1.0 are available in v2 API and 2.0 frontend

### 10.4 DNS/HTTP Dual-Write Transition
1. **Phase 1**: Migration script written and tested on staging
2. **Phase 2**: Dual-write enabled in DNS/HTTP handlers — both `TblDns`/`TblHttp` and `Interaction` table written simultaneously
3. **Post-Phase 2**: Verify Interaction table has complete data, run migration script for historical data
4. **Phase 3+**: V2 API queries Interaction table exclusively; v1 API continues querying old tables for backward compatibility
5. **2.1 release**: Stop dual-write, deprecate old tables

---

## 11. Testing Strategy

### 11.1 Backend Testing
- **Unit tests**: all new Go functions must have unit tests in `*_test.go` files
- **Integration tests**: all new API endpoints must have integration tests in `server/v2_api_test.go`
- **Coverage target**: > 60% for core business logic (per `doc/2.0-Requirement.md` §6.4)
- **Run command**: `go test ./... -count=1`
- **Verification log**: record commands and results in `docs/verification.md`

### 11.2 Frontend Testing
- **E2E tests**: all new or significantly modified pages must have Playwright E2E tests
- **Component tests**: new reusable components (DataTable, ErrorBoundary, LoadingSkeleton) should have component tests
- **Run command**: `cd frontend-next && npx playwright test --reporter=line`
- **No blocking report server**: use one-shot non-interactive commands only (per AGENTS.md windsurf convention)

### 11.3 Acceptance Testing
- Each Phase has explicit acceptance criteria (checkboxes in spec)
- All acceptance criteria must be verified before Phase is marked complete
- Verification commands and results recorded in `docs/verification.md`

---

## 12. Decisions Table

| Decision | Rationale | Reference |
|----------|-----------|-----------|
| Bottom-up phasing (backend first) | Building frontend on fake APIs is wasted effort | This spec §2 |
| Dual-write for data model unification | Compatibility with 1.0, gradual migration | `doc/2.0-Requirement.md` Q3 |
| SMTP notification not implemented | SMTP is for email user sniffening, not notifications | `doc/2.0-Requirement.md` Q2 |
| MCP Streamable HTTP protocol | Per user decision | `doc/2.0-Requirement.md` Q4 |
| Rebinding telemetry on by default | Anti-abuse measure | `doc/2.0-Requirement.md` Q6 |
| 1.0 frontend directly deprecated | Per user decision, data migration tool provided | `doc/2.0-Requirement.md` Q3 |
| APIKey bcrypt hash storage | Security requirement | `doc/2.0-Requirement.md` §6.2 |
| Phase 5 split into 5a/5b | Protocol listeners are high-risk, need security review | Review §3.5 |

---

## 13. References

- `doc/2.0-Requirement.md` — Product requirements and Q1-Q7 decisions
- `doc/2.0-frontend-implementation-gap.md` — Frontend gap analysis
- `doc/2.0-completion-summary.md` — Previous completion summary (inaccurate)
- `server/v2_api.go` — Current v2 API routes (3888 lines)
- `internal/workflow/service.go` — Workflow stubs
- `internal/rule/action.go` — Rule engine stubs
- `internal/mcp/server.go` — MCP server (HTTP-based, not protocol-compliant)
