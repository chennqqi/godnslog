# Phase 1.5: Data Model Migration Script

> **Spec ref:** `docs/superpowers/specs/2026-06-20-mvp-to-production-design.md` §3.4

**Goal:** Write idempotent migration script with `--dry-run` support to import historical `TblDns`/`TblHttp` records into `Interaction` table. Add batch import method to interaction service.

**Acceptance:**
- `migration/sync.go` CLI entry point with `--dry-run` flag
- Migration is idempotent (skips already-migrated records)
- Batch processing (1000/batch) with progress logging
- `go test ./migration/ -v` passes

---

## Task 1: Add Batch Import to Interaction Service

**Files:** Modify `internal/interaction/service.go`, create/modify `internal/interaction/service_test.go`

- [ ] **Step 1: Write failing test** — Add test `TestBatchImportInteractions`:
  - Setup: in-memory SQLite, sync Interaction schema
  - Create 3 `Interaction` records with distinct timestamp+token
  - Call `BatchImport(interactions)` → verify 3 inserted
  - Call again with same records → verify 0 inserted (idempotent)
  - Call with 2 new + 1 duplicate → verify 2 inserted

- [ ] **Step 2: Verify failure** — `go test ./internal/interaction/ -run TestBatchImport -v` → FAIL (method doesn't exist)

- [ ] **Step 3: Implement** — Add to `internal/interaction/service.go`:
  ```go
  // BatchImport imports interactions idempotently.
  // Skips records where an interaction with the same timestamp+token already exists.
  // Returns the number of records actually inserted.
  func (s *Service) BatchImport(interactions []*models.Interaction) (int, error)
  ```
  - For each interaction: check if `timestamp = ? AND token = ?` already exists
  - If not exists, insert; if exists, skip
  - Return count of inserted records

- [ ] **Step 4: Verify pass** — `go test ./internal/interaction/ -run TestBatchImport -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(interaction): add idempotent batch import method"`

---

## Task 2: Add Dry-Run and Idempotency to Migrator

**Files:** Modify `migration/migrate.go`, `migration/migrate_test.go`

- [ ] **Step 1: Write failing test** — Add tests:
  - `TestMigrateDNS_DryRun`: insert TblDns records, run `MigrateDNSWithFlags(batchSize, dryRun=true)`, verify no Interaction records inserted, verify stats count is correct
  - `TestMigrateDNS_Idempotent`: run migration twice, verify second run inserts 0 records
  - `TestMigrateHTTP_DryRun`: same pattern for HTTP
  - `TestMigrateStats`: verify `MigrationStats` struct with DNSCount, HTTPCount, DNSMigrated, HTTPMigrated, DNSSkipped, HTTPSkipped

- [ ] **Step 2: Verify failure** — `go test ./migration/ -run "TestMigrateDNS_DryRun|TestMigrateDNS_Idempotent|TestMigrateStats" -v` → FAIL

- [ ] **Step 3: Implement** — In `migration/migrate.go`:
  - Add `MigrationStats` struct:
    ```go
    type MigrationStats struct {
        DNSCount     int64
        HTTPCount    int64
        DNSMigrated  int64
        HTTPMigrated int64
        DNSSkipped   int64
        HTTPSkipped  int64
    }
    ```
  - Add `MigrateDNSWithFlags(batchSize int, dryRun bool) (*MigrationStats, error)`:
    - Batch fetch TblDns records
    - For each: convert to Interaction via `models.FromTblDns`
    - Check if Interaction with same timestamp+token exists → skip (increment DNSSkipped)
    - If dryRun: don't insert, just count
    - If not dryRun: insert, increment DNSMigrated
  - Add `MigrateHTTPWithFlags(batchSize int, dryRun bool) (*MigrationStats, error)` — same pattern
  - Update `MigrateAll` to call new methods with `dryRun=false`

- [ ] **Step 4: Verify pass** — `go test ./migration/ -run "TestMigrateDNS_DryRun|TestMigrateDNS_Idempotent|TestMigrateStats" -v`

- [ ] **Step 5: Commit** — `git commit -m "feat(migration): add dry-run support and idempotent migration"`

---

## Task 3: CLI Entry Point — sync.go

**Files:** Create `migration/sync.go`, `migration/sync_test.go`

- [ ] **Step 1: Write failing test** — Create `migration/sync_test.go`:
  - `TestSyncCommand_DryRun`: setup engine with TblDns records, call `RunSync(engine, dryRun=true)`, verify no interactions inserted, verify stats returned with correct counts
  - `TestSyncCommand_RealRun`: call `RunSync(engine, dryRun=false)`, verify interactions inserted
  - `TestSyncCommand_Idempotent`: call twice, verify second call inserts 0

- [ ] **Step 2: Verify failure** — `go test ./migration/ -run TestSyncCommand -v` → FAIL (RunSync undefined)

- [ ] **Step 3: Implement** — Create `migration/sync.go`:
  ```go
  package migration

  import (
      "flag"
      "fmt"
      "log"
      "os"
      "xorm.io/xorm"
      _ "modernc.org/sqlite"
  )

  // RunSync executes the migration with the given options.
  func RunSync(engine *xorm.Engine, dryRun bool) (*MigrationStats, error) {
      migrator := NewMigrator(engine)
      dnsStats, err := migrator.MigrateDNSWithFlags(1000, dryRun)
      if err != nil { return nil, fmt.Errorf("DNS migration failed: %w", err) }
      httpStats, err := migrator.MigrateHTTPWithFlags(1000, dryRun)
      if err != nil { return nil, fmt.Errorf("HTTP migration failed: %w", err) }
      combined := &MigrationStats{
          DNSCount: dnsStats.DNSCount, HTTPCount: httpStats.HTTPCount,
          DNSMigrated: dnsStats.DNSMigrated, HTTPMigrated: httpStats.HTTPMigrated,
          DNSSkipped: dnsStats.DNSSkipped, HTTPSkipped: httpStats.HTTPSkipped,
      }
      return combined, nil
  }

  // Main is the CLI entry point for the sync command.
  func Main() {
      driver := flag.String("driver", "sqlite", "database driver")
      dsn := flag.String("dsn", "", "database DSN")
      dryRun := flag.Bool("dry-run", false, "dry run mode (no actual writes)")
      flag.Parse()
      if *dsn == "" { log.Fatal("dsn is required") }
      engine, err := xorm.NewEngine(*driver, *dsn)
      if err != nil { log.Fatalf("failed to create engine: %v", err) }
      defer engine.Close()
      stats, err := RunSync(engine, *dryRun)
      if err != nil { log.Fatalf("migration failed: %v", err) }
      log.Printf("Migration complete: DNS=%d migrated, %d skipped; HTTP=%d migrated, %d skipped",
          stats.DNSMigrated, stats.DNSSkipped, stats.HTTPMigrated, stats.HTTPSkipped)
      _ = os.Stdout.Sync()
  }
  ```

- [ ] **Step 4: Verify pass** — `go test ./migration/ -run TestSyncCommand -v`

- [ ] **Step 5: Commit** — `git add migration/sync.go migration/sync_test.go && git commit -m "feat(migration): add sync CLI with dry-run support"`

---

## Final Verification

```bash
go build ./migration/... ./internal/interaction/...
go test ./migration/ -v
go test ./internal/interaction/ -run TestBatchImport -v
```
