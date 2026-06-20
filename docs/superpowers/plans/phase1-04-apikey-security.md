# Phase 1.4: APIKey Security Hardening

> **Spec ref:** `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §3.5

**Goal:** Migrate APIKey storage from plaintext to bcrypt hash with auto-migration on auth, add key rotation endpoint.

**Acceptance:**
- New API keys stored as bcrypt hash in `key_hash`, plaintext returned only once
- Legacy plaintext keys auto-migrated to bcrypt on first authentication
- `POST /api/v2/apikeys/:id/rotate` generates new key
- `go test ./internal/auth/ ./server/ -run "TestAPIKey|TestV2Rotate" -v` passes

---

## Task 1: Add KeyHash Field to APIKey Model

**Files:** Modify `internal/models/apikey.go`

- [ ] **Step 1: Write failing test** — Add to `internal/auth/service_test.go`:
  - `TestAPIKeyBcryptMigration_NewKey`: create key, verify `KeyHash` is non-empty and not equal to plaintext `Key`
  - `TestAPIKeyBcryptMigration_LegacyPlaintext`: insert legacy key with `KeyHash=""`, validate, verify `KeyHash` backfilled and `Key` cleared
  - `TestAPIKeyRotation`: create key, rotate, old key fails validation, new key works

- [ ] **Step 2: Verify failure** — `go test ./internal/auth/ -run "TestAPIKeyBcrypt|TestAPIKeyRotation" -v` → FAIL (KeyHash field doesn't exist, RotateAPIKey undefined)

- [ ] **Step 3: Add field** — In `internal/models/apikey.go`, add after `Key` field (line 38):
  ```go
  KeyHash string `json:"-" xorm:"varchar(128)"` // Bcrypt hash; empty for legacy keys
  ```

- [ ] **Step 4: Implement bcrypt in CreateAPIKey** — In `internal/auth/service.go`:
  - Add import: `"golang.org/x/crypto/bcrypt"`, `"fmt"`
  - In `CreateAPIKey`: after generating key, compute `bcrypt.GenerateFromPassword([]byte(key), bcrypt.DefaultCost)`, store hash in `KeyHash`, keep plaintext in `Key` for return value (caller receives it once)

- [ ] **Step 5: Implement ValidateAPIKey migration** — Replace `ValidateAPIKey` (lines 212-238):
  ```
  1. Extract prefix (first 8 chars) from fullKey
  2. Lookup by prefix via GetAPIKeyByPrefix
  3. Check validity (revoked/expired)
  4. If KeyHash non-empty: bcrypt.CompareHashAndPassword → return key
  5. If KeyHash empty (legacy): compare plaintext Key
     - If match: compute bcrypt hash, store in KeyHash, clear Key, update DB
     - If no match: return ErrAPIKeyNotFound
  ```

- [ ] **Step 6: Implement RotateAPIKey** — Add method:
  ```go
  func (s *Service) RotateAPIKey(id string) (string, error)
  ```
  - Fetch key by ID, generate new key+prefix, bcrypt hash, update `KeyHash`+`KeyPrefix`, clear `Key`, return plaintext

- [ ] **Step 7: Fix UpdateLastUsed** — Currently queries by `key = ?`, change to query by `key_prefix = ?` (first 8 chars of full key)

- [ ] **Step 8: Verify pass** — `go test ./internal/auth/ -run "TestAPIKeyBcrypt|TestAPIKeyRotation" -v`

- [ ] **Step 9: Commit** — `git add internal/models/apikey.go internal/auth/service.go internal/auth/service_test.go && git commit -m "feat(auth): add bcrypt key_hash with auto-migration and key rotation"`

---

## Task 2: V2 API — APIKey Rotate Endpoint

**Files:** Modify `server/v2_api.go`, `server/v2_api_test.go`

- [ ] **Step 1: Write failing test** — Add `TestV2RotateAPIKey`:
  - Setup: admin user, login, create API key
  - POST `/api/v2/apikeys/:id/rotate` → 200, response data contains new `key` (non-empty)
  - Verify new key is different from original (if original was returned during creation)

- [ ] **Step 2: Verify failure** — `go test ./server/ -run TestV2RotateAPIKey -v` → FAIL (404)

- [ ] **Step 3: Add route and handler** — In `server/v2_api.go`:
  - Add to apikeys route group:
    ```go
    apikeys.POST("/:id/rotate", self.v2RotateAPIKey)
    ```
  - `v2RotateAPIKey`:
    - Get identity from context
    - Fetch API key, verify ownership (creator or admin)
    - Call `authService.RotateAPIKey(keyID)`
    - Return `{id, key, key_prefix}`

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2RotateAPIKey -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(api): add POST /apikeys/:id/rotate endpoint"`

---

## Task 3: Update V2 APIKey List — Hide Full Key

**Files:** Modify `server/v2_api.go` (`v2ListAPIKeys`, `v2GetAPIKey`)

- [ ] **Step 1: Write failing test** — Add `TestV2ListAPIKeys_HidesFullKey`:
  - Create API key, list keys, verify response items show `key_prefix` but not full `key` or `key_hash`

- [ ] **Step 2: Verify failure** — Check current `v2ListAPIKeys` response includes full key

- [ ] **Step 3: Fix** — In `v2ListAPIKeys` and `v2GetAPIKey` handlers:
  - Replace `key` field with `key_prefix + "..."` in response
  - Never include `key_hash` in response

- [ ] **Step 4: Verify pass** — `go test ./server/ -run TestV2ListAPIKeys_HidesFullKey -v`

- [ ] **Step 5: Commit** — `git commit -m "fix(api): hide full API key in list/get responses, show only prefix"`

---

## Final Verification

```bash
go build ./internal/auth/... ./internal/models/... ./server/...
go test ./internal/auth/ -v
go test ./server/ -run "TestV2Rotate|TestV2ListAPIKeys" -v
```
