# Phase 1.2: Rule Engine Completion

> **Spec ref:** `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §3.2

**Goal:** Fix CIDR matching placeholder, implement tag/report/noise actions in rule engine.

**Acceptance:**
- `matchCIDR` uses `net.ParseCIDR` + `network.Contains`
- Tag action adds/removes tags on interaction
- Report action generates report data
- DiscardNoise sets noise flag
- `go test ./internal/rule/ -v` passes

---

## Task 1: Fix CIDR Matching

**Files:** Modify `internal/rule/engine.go:263-268`, create `internal/rule/engine_test.go`

- [ ] **Step 1: Write failing test** — Create `internal/rule/engine_test.go` with tests:
  - `TestMatchCIDR_Exact32`: "192.168.1.1/32" matches "192.168.1.1", not "192.168.1.2"
  - `TestMatchCIDR_24`: "192.168.1.0/24" matches .50 and .255, not 192.168.2.1
  - `TestMatchCIDR_10`: "10.0.0.0/8" matches 10.255.255.255 and 10.1.2.3, not 11.0.0.1
  - `TestMatchCIDR_InvalidCIDR`: "invalid" returns false
  - `TestMatchCIDR_InvalidIP`: "not-an-ip" returns false

- [ ] **Step 2: Verify failure** — `go test ./internal/rule/ -run TestMatchCIDR -v` → FAIL (prefix matching can't handle /24 ranges)

- [ ] **Step 3: Fix** — Add `"net"` to imports. Replace `matchCIDR` (lines 263-268):
  ```go
  func matchCIDR(cidr, ip string) bool {
      _, network, err := net.ParseCIDR(cidr)
      if err != nil { return false }
      parsedIP := net.ParseIP(ip)
      if parsedIP == nil { return false }
      return network.Contains(parsedIP)
  }
  ```

- [ ] **Step 4: Verify pass** — `go test ./internal/rule/ -run TestMatchCIDR -v`

- [ ] **Step 5: Commit** — `git add internal/rule/engine.go internal/rule/engine_test.go && git commit -m "fix(rule): replace placeholder CIDR matching with net.ParseCIDR"`

---

## Task 2: Implement Tag Action

**Files:** Modify `internal/rule/action.go:219-225`, create `internal/rule/action_test.go`

- [ ] **Step 1: Write failing test** — Create `internal/rule/action_test.go`:
  - `TestExecuteTagAction_AddTags`: add "malicious", "confirmed" → verify `inter["_tags"]` contains both
  - `TestExecuteTagAction_RemoveTags`: existing tags ["noise", "suspicious"], remove "noise" → only "suspicious" remains
  - `TestExecuteTagAction_AddDuplicate`: add "tag1" when already present → no duplicate

- [ ] **Step 2: Verify failure** — `go test ./internal/rule/ -run TestExecuteTagAction -v` → FAIL (returns nil without setting `_tags`)

- [ ] **Step 3: Implement** — Replace `executeTagAction` (lines 219-225):
  - Read existing tags from `inter["_tags"]`
  - Append new tags from `tag.Add` (deduplicate)
  - Remove tags in `tag.Remove` using a set
  - Write back to `inter["_tags"]`

- [ ] **Step 4: Verify pass** — `go test ./internal/rule/ -run TestExecuteTagAction -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(rule): implement tag action with add/remove support"`

---

## Task 3: Implement Report Action

**Files:** Modify `internal/rule/action.go:240-245`, `internal/rule/action_test.go`

- [ ] **Step 1: Write failing test** — Add `TestExecuteReport_GeneratesReport`:
  - Report{Format: "json", Title: "Test Report"}
  - Interaction has id, type, source_ip
  - After execution, `inter["_report"]` is not nil and contains title, format, interaction_id

- [ ] **Step 2: Verify failure** — `go test ./internal/rule/ -run TestExecuteReport -v` → FAIL (returns nil without setting `_report`)

- [ ] **Step 3: Implement** — Replace `executeReport` (lines 240-245):
  - Build report data map with title, format, interaction_id, source_ip, type
  - Store in `inter["_report"]`

- [ ] **Step 4: Verify pass** — `go test ./internal/rule/ -run TestExecuteReport -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(rule): implement report action"`

---

## Task 4: Implement DiscardNoise Action

**Files:** Modify `internal/rule/action.go:31-35`, `internal/rule/action_test.go`

- [ ] **Step 1: Write failing test** — Add `TestExecute_DiscardNoise`:
  - Rule with `Actions.DiscardNoise = true`
  - After `Execute()`, `inter["_noise"]` is true

- [ ] **Step 2: Verify failure** — `go test ./internal/rule/ -run TestExecute_DiscardNoise -v` → FAIL (doesn't set `_noise`)

- [ ] **Step 3: Implement** — Replace DiscardNoise block (lines 31-35):
  ```go
  if actions.DiscardNoise {
      inter["_noise"] = true
  }
  ```

- [ ] **Step 4: Verify pass** — `go test ./internal/rule/ -run TestExecute_DiscardNoise -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(rule): implement discard-noise action flag"`

---

## Final Verification

```bash
go build ./internal/rule/...
go test ./internal/rule/ -v
```
