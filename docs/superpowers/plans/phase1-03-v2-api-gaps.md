# Phase 1.3: V2 API Gap Filling — DNS Records, XIP, User Management

> **Spec ref:** `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §3.3

**Goal:** Add missing v2 API endpoints for DNS record CRUD, xip encoding query, and user management CRUD.

**Acceptance:**
- `GET/POST/PUT/DELETE /api/v2/dns/records` works
- `GET /api/v2/dns/xip/:ip` returns 3 encoding formats
- `POST/PUT/DELETE /api/v2/users` works (admin only, cannot delete super)
- `go test ./server/ -run "TestV2DNS|TestV2User|TestV2Xip" -v` passes

---

## Task 1: DNS Record CRUD — List & Create

**Files:** Modify `server/v2_api.go` (routes + handlers), `server/v2_api_test.go`

- [ ] **Step 1: Write failing test** — Add `TestV2DNSRecordsCRUD`:
  - Setup: in-memory SQLite, create admin user, login, get token
  - POST `/api/v2/dns/records` with `{"host":"www","type":"A","value":"192.168.1.1","ttl":300}` → 200
  - GET `/api/v2/dns/records` → 200, items length >= 1
  - Verify response format: `{code, message, data: {items, total, page, page_size, total_pages}}`

- [ ] **Step 2: Verify failure** — `go test ./server/ -run TestV2DNSRecordsCRUD -v` → FAIL (404)

- [ ] **Step 3: Add routes and handlers** — In `server/v2_api.go`:
  - Add route group after rebinding block:
    ```go
    dnsRecords := v2.Group("/dns/records", self.authHandler)
    { dnsRecords.GET("", self.v2ListDNSRecords); dnsRecords.POST("", self.v2CreateDNSRecord) }
    ```
  - `v2ListDNSRecords`: paginate `models.TblResolve`, return items with id/host/type/value/ttl/created_at/updated_at
  - `v2CreateDNSRecord`: bind JSON `{host, type, value, ttl}`, default ttl=300, insert `TblResolve`, return created record
  - Add `"net"` and `"fmt"` to imports if not present

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2DNSRecordsCRUD -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(api): add v2 DNS record list and create endpoints"`

---

## Task 2: DNS Record CRUD — Update & Delete

**Files:** Modify `server/v2_api.go`, `server/v2_api_test.go`

- [ ] **Step 1: Write failing test** — Extend `TestV2DNSRecordsCRUD`:
  - PUT `/api/v2/dns/records/:id` with `{"value":"10.0.0.1"}` → 200
  - GET list → verify updated value
  - DELETE `/api/v2/dns/records/:id` → 200
  - GET list → verify deleted (length decreased)
  - PUT `/api/v2/dns/records/99999` → 404
  - DELETE `/api/v2/dns/records/99999` → 404 (or 200 with no-op)

- [ ] **Step 2: Verify failure** — `go test ./server/ -run TestV2DNSRecordsCRUD -v` → FAIL (404 on PUT/DELETE)

- [ ] **Step 3: Add routes and handlers** — Add to route group:
  ```go
  dnsRecords.PUT("/:id", self.v2UpdateDNSRecord)
  dnsRecords.DELETE("/:id", self.v2DeleteDNSRecord)
  ```
  - `v2UpdateDNSRecord`: parse `:id`, bind JSON, fetch existing (404 if not found), update non-empty fields, save
  - `v2DeleteDNSRecord`: parse `:id`, delete by ID, return success

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2DNSRecordsCRUD -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(api): add v2 DNS record update and delete endpoints"`

---

## Task 3: XIP Encoding Query

**Files:** Modify `server/v2_api.go`, `server/v2_api_test.go`

- [ ] **Step 1: Write failing test** — Add `TestV2QueryXip`:
  - GET `/api/v2/dns/xip/192.168.1.1` → 200
  - Verify response data contains: `dotted` = "192.168.1.1", `hex` = "c0a80101", `binary` = "0b11000000101010000000000100000001"
  - Verify `examples` contains dotted_decimal, hex, binary with ".example.com" suffix
  - GET `/api/v2/dns/xip/notanip` → 400
  - GET `/api/v2/dns/xip/::1` → 400 (IPv6 not supported)

- [ ] **Step 2: Verify failure** — `go test ./server/ -run TestV2QueryXip -v` → FAIL (404)

- [ ] **Step 3: Add route and handler** — Add route:
  ```go
  v2.GET("/dns/xip/:ip", self.authHandler, self.v2QueryXip)
  ```
  - `v2QueryXip`: parse `:ip`, validate with `net.ParseIP`, reject non-IPv4
  - Generate 3 formats: dotted decimal, hex (`%02x%02x%02x%02x`), binary (`0b%08b%08b%08b%08b`)
  - Return with example domains

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2QueryXip -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(api): add v2 xip encoding query endpoint"`

---

## Task 4: User Management — Create

**Files:** Modify `server/v2_api.go` (routes + handlers), `server/v2_api_test.go`

- [ ] **Step 1: Write failing test** — Add `TestV2UserManagement`:
  - Setup: admin user (role=0), login, get token
  - POST `/api/v2/users` with `{"username":"newuser","email":"new@test.com","password":"StrongPass123!","role":1}` → 200
  - GET `/api/v2/users` → 200, items >= 2
  - POST with weak password → 400
  - POST with duplicate username → 409

- [ ] **Step 2: Verify failure** — `go test ./server/ -run TestV2UserManagement -v` → FAIL (404 on POST)

- [ ] **Step 3: Add route and handler** — Update users route block:
  ```go
  users := v2.Group("/users", self.authHandler)
  { users.GET("", self.v2ListUsers); users.POST("", self.v2CreateUser) }
  ```
  - `v2CreateUser`: bind JSON `{username, email, password, role, lang}`, validate password >= 8 chars, bcrypt hash, generate token+shortid, insert `TblUser`, handle duplicate → 409

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2UserManagement -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(api): add v2 user create endpoint"`

---

## Task 5: User Management — Update & Delete

**Files:** Modify `server/v2_api.go`, `server/v2_api_test.go`

- [ ] **Step 1: Write failing test** — Extend `TestV2UserManagement`:
  - PUT `/api/v2/users/:id` with `{"role":2}` → 200
  - PUT `/api/v2/users/:id` with `{"password":"NewStrongPass456!"}` → 200
  - DELETE `/api/v2/users/:id` (non-super user) → 200
  - DELETE super user (role=0) → 403
  - PUT super user role to 1 → 403

- [ ] **Step 2: Verify failure** — `go test ./server/ -run TestV2UserManagement -v` → FAIL (404 on PUT/DELETE)

- [ ] **Step 3: Add routes and handlers** — Add to users route group:
  ```go
  users.PUT("/:id", self.v2UpdateUser)
  users.DELETE("/:id", self.v2DeleteUser)
  ```
  - `v2UpdateUser`: parse `:id`, fetch user (404 if not found), prevent demoting super (role 0 → non-0 = 403), update email/role/lang/password (bcrypt if provided), save
  - `v2DeleteUser`: parse `:id`, fetch user (404 if not found), prevent deleting super (role=0 → 403), delete

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2UserManagement -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(api): add v2 user update and delete endpoints with super-user protection"`

---

## Final Verification

```bash
go build ./server/...
go test ./server/ -run "TestV2DNS|TestV2User|TestV2Xip" -v
```
