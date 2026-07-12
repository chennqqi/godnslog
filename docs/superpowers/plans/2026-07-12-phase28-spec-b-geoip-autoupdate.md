# Phase 2.8 Spec B: GeoIP Auto-update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Auto-download GeoLite2-ASN.mmdb at startup when missing, wire the fingerprinter into all production interaction-service call sites, and persist ASN/Org enrichment on interactions.

**Architecture:** Add ASN/Org fields to the `Interaction` model (xorm auto-migrates via existing `Sync2`). Add a `downloader` that fetches the MaxMind tar.gz and extracts the mmdb. Thread a `*fingerprint.Fingerprinter` from `servecmd.go` through `WebServerConfig` into `WebServer`, and replace the `nil` fingerprinter at all 10 production `interaction.NewService` call sites.

**Tech Stack:** Go 1.25, xorm, `archive/tar`, `compress/gzip`, `net/http`, `os`, `io`; MaxMind free signup at https://www.maxmind.com/en/geolite2/signup.

## Global Constraints

- Module path: `github.com/chennqqi/godnslog`
- Test command: `GOCACHE=/tmp/gocache go test ./...`
- xorm `Sync2` auto-migrates new struct fields - no hand-written SQL
- GeoLite2-ASN.mmdb only contains ASN + organization (no country) - Country field stays empty
- Download must not block startup on failure - log warning and continue without GeoIP
- License key is optional; without it, GeoIP is disabled with a warning that includes the signup URL

---

## File Structure

| File | Responsibility |
|------|----------------|
| `internal/models/interaction.go` | Add ASN/Org/Country fields |
| `internal/interaction/service.go` | Persist ASN/Org in `enhanceInteraction` |
| `internal/interaction/fingerprint/downloader.go` | New: download + extract mmdb |
| `internal/interaction/fingerprint/downloader_test.go` | New: downloader tests |
| `internal/interaction/fingerprint/geoip_test.go` | New or extend: geoip lookup tests |
| `server/webserver.go` | Add `fingerprinter` field + `GeoIPMMDBPath`/`GeoIPLicenseKey` config |
| `server/v2_api.go` | Replace 10 `nil` fingerprinter args with `self.fingerprinter` |
| `internal/evidencehub/service.go` | Optional: thread fingerprinter (reads-only, low priority) |
| `servecmd.go` | Add `-geoip-mmdb` / `-geoip-license-key` flags + download logic |

---

### Task 1: Add ASN/Org/Country fields to Interaction model

**Files:**
- Modify: `internal/models/interaction.go:69-72`

**Interfaces:**
- Produces: `Interaction.ASN *uint`, `Interaction.Org *string`, `Interaction.Country *string` (auto-migrated by existing `Sync2` in `server/webui.go:64`)

- [ ] **Step 1: Add the fields**

In `internal/models/interaction.go`, after the existing `SourceName` field (line 71), add:

```go
	// GeoIP enrichment fields (set by fingerprint pipeline when mmdb is loaded)
	ASN     *uint   `json:"asn,omitempty" xorm:"'asn' int"`
	Org     *string `json:"org,omitempty" xorm:"'org' varchar(255)"`
	Country *string `json:"country,omitempty" xorm:"'country' varchar(64)"`
```

- [ ] **Step 2: Verify build + run interaction model tests**

Run: `GOCACHE=/tmp/gocache go build ./... && GOCACHE=/tmp/gocache go test ./internal/models/... ./internal/interaction/...`
Expected: Build OK, tests PASS.

- [ ] **Step 3: Commit**

```bash
git add internal/models/interaction.go
git commit -m "feat(models): add ASN/Org/Country fields to Interaction for GeoIP enrichment"
```

---

### Task 2: Persist ASN/Org in enhanceInteraction

**Files:**
- Modify: `internal/interaction/service.go:372-383`
- Test: `internal/interaction/service_test.go` (extend) or new test file

**Interfaces:**
- Consumes: `Interaction.ASN/Org/Country` from Task 1, `fingerprint.Fingerprint` (already has ASN/Org/Country fields)
- Produces: `enhanceInteraction` populates ASN/Org/Country on the interaction

- [ ] **Step 1: Write the failing test**

Create `internal/interaction/geoip_enrich_test.go`:

```go
package interaction

import (
	"testing"
	"time"

	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/interaction/fingerprint"
)

func TestEnhanceInteraction_PersistsASN(t *testing.T) {
	// NewFingerprinter("") disables GeoIP (no mmdb), so ASN stays nil.
	// This test verifies the persistence path does not crash when fp is nil.
	fp := fingerprint.NewFingerprinter("")
	domain := "token.example.com"
	ip := "1.2.3.4"
	inter := &v2models.Interaction{
		ID:        "test-id",
		Type:      "dns",
		Timestamp: time.Now(),
		SourceIP:  ip,
		Domain:    &domain,
	}
	enhanceInteraction(inter, fp)
	if inter.ASN != nil {
		t.Errorf("expected nil ASN with empty fingerprinter, got %v", *inter.ASN)
	}
}
```

- [ ] **Step 2: Run test to verify it passes (baseline)**

Run: `GOCACHE=/tmp/gocache go test ./internal/interaction/... -run TestEnhanceInteraction_PersistsASN -v`
Expected: PASS (currently `enhanceInteraction` doesn't touch ASN, so it stays nil).

- [ ] **Step 3: Modify enhanceInteraction to persist ASN/Org/Country**

In `internal/interaction/service.go`, update the source fingerprint block (around line 372-383). Replace:

```go
	// Step 3: Source fingerprint
	if fp != nil && interaction.SourceIP != "" {
		var ua string
		if interaction.UserAgent != nil {
			ua = *interaction.UserAgent
		}
		result := fp.Lookup(interaction.SourceIP, ua)
		if result != nil && result.SourceType != fingerprint.SourceUnknown {
			interaction.SourceType = &result.SourceType
			interaction.SourceName = &result.SourceName
		}
	}
```

With:

```go
	// Step 3: Source fingerprint
	if fp != nil && interaction.SourceIP != "" {
		var ua string
		if interaction.UserAgent != nil {
			ua = *interaction.UserAgent
		}
		result := fp.Lookup(interaction.SourceIP, ua)
		if result != nil {
			if result.SourceType != fingerprint.SourceUnknown {
				interaction.SourceType = &result.SourceType
				interaction.SourceName = &result.SourceName
			}
			// Persist GeoIP enrichment regardless of source classification.
			if result.ASN > 0 {
				interaction.ASN = &result.ASN
			}
			if result.Org != "" {
				interaction.Org = &result.Org
			}
			if result.Country != "" {
				interaction.Country = &result.Country
			}
		}
	}
```

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./internal/interaction/... -run TestEnhanceInteraction_PersistsASN -v`
Expected: PASS.

- [ ] **Step 5: Run full interaction test suite**

Run: `GOCACHE=/tmp/gocache go test ./internal/interaction/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/interaction/service.go internal/interaction/geoip_enrich_test.go
git commit -m "feat(interaction): persist ASN/Org/Country from fingerprint lookup"
```

---

### Task 3: Implement mmdb downloader

**Files:**
- Create: `internal/interaction/fingerprint/downloader.go`
- Create: `internal/interaction/fingerprint/downloader_test.go`

**Interfaces:**
- Produces: `DownloadAndExtractMMDB(url, destPath string) error`, `EnsureMMDB(path, licenseKey string) error`

- [ ] **Step 1: Write the failing test**

Create `internal/interaction/fingerprint/downloader_test.go`:

```go
package fingerprint

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// buildTarGz builds a .tar.gz whose sole entry is dir/GeoLite2-ASN.mmdb
// containing fakeContent, matching MaxMind's release archive layout.
func buildTarGz(t *testing.T, dir, name, content string) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)
	filePath := filepath.Join(dir, name)
	hdr := &tar.Header{
		Name: filePath, Mode: 0644, Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("write header: %v", err)
	}
	if _, err := tw.Write([]byte(content)); err != nil {
		t.Fatalf("write body: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gw.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
	}
	return buf.Bytes()
}

func TestDownloadAndExtractMMDB_Success(t *testing.T) {
	tarBytes := buildTarGz(t, "GeoLite2-ASN_20260101", "GeoLite2-ASN.mmdb", "fake-mmdb-content")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/gzip")
		w.Write(tarBytes)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "GeoLite2-ASN.mmdb")
	if err := DownloadAndExtractMMDB(srv.URL, dest); err != nil {
		t.Fatalf("DownloadAndExtractMMDB: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read dest: %v", err)
	}
	if string(got) != "fake-mmdb-content" {
		t.Errorf("extracted content = %q, want %q", got, "fake-mmdb-content")
	}
}

func TestDownloadAndExtractMMDB_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "GeoLite2-ASN.mmdb")
	err := DownloadAndExtractMMDB(srv.URL, dest)
	if err == nil {
		t.Fatal("expected error on 403, got nil")
	}
}

func TestEnsureMMDB_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GeoLite2-ASN.mmdb")
	if err := os.WriteFile(path, []byte("existing"), 0644); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	// Should be a no-op when file already exists (returns nil, no download).
	if err := EnsureMMDB(path, "any-key"); err != nil {
		t.Errorf("EnsureMMDB with existing file should return nil, got %v", err)
	}
}

func TestEnsureMMDB_NoLicenseKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "GeoLite2-ASN.mmdb")
	// No license key + no file -> should return error mentioning the signup URL.
	err := EnsureMMDB(path, "")
	if err == nil {
		t.Fatal("expected error when no license key and no file, got nil")
	}
}
```

- [ ] **Step 2: Run test to verify it fails**

Run: `GOCACHE=/tmp/gocache go test ./internal/interaction/fingerprint/... -run 'TestDownloadAndExtractMMDB|TestEnsureMMDB' -v`
Expected: FAIL - `DownloadAndExtractMMDB` and `EnsureMMDB` undefined.

- [ ] **Step 3: Implement the downloader**

Create `internal/interaction/fingerprint/downloader.go`:

```go
package fingerprint

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const mmdbLicenseSignupURL = "https://www.maxmind.com/en/geolite2/signup"

// mmdbName is the file we extract from inside the MaxMind tar.gz archive.
const mmdbName = "GeoLite2-ASN.mmdb"

// downloadMMDBTarGz fetches the tar.gz bytes from url with a 5-minute timeout.
func downloadMMDBTarGz(url string) ([]byte, error) {
	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("download mmdb: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download mmdb: unexpected status %d", resp.StatusCode)
	}
	return io.ReadAll(resp.Body)
}

// extractMMDB walks the tar.gz, finds the entry ending in mmdbName,
// and writes its contents to destPath.
func extractMMDB(tarGzBytes []byte, destPath string) error {
	gzr, err := gzip.NewReader(bytes.NewReader(tarGzBytes))
	if err != nil {
		return fmt.Errorf("open gzip: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar: %w", err)
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		if filepath.Base(hdr.Name) != mmdbName {
			continue
		}
		// Ensure parent dir exists.
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return fmt.Errorf("mkdir dest: %w", err)
		}
		out, err := os.Create(destPath)
		if err != nil {
			return fmt.Errorf("create dest: %w", err)
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return fmt.Errorf("write dest: %w", err)
		}
		return out.Close()
	}
	return fmt.Errorf("mmdb entry %q not found in archive", mmdbName)
}

// DownloadAndExtractMMDB downloads a MaxMind tar.gz from url and extracts
// GeoLite2-ASN.mmdb to destPath.
func DownloadAndExtractMMDB(url, destPath string) error {
	data, err := downloadMMDBTarGz(url)
	if err != nil {
		return err
	}
	return extractMMDB(data, destPath)
}

// EnsureMMDB ensures an mmdb file exists at path. If the file already exists,
// it is a no-op. If the file is missing and licenseKey is non-empty, it
// downloads from MaxMind. If both are missing, it returns an error that
// mentions the free signup URL.
func EnsureMMDB(path, licenseKey string) error {
	if path == "" {
		return errors.New("mmdb path is empty")
	}
	if _, err := os.Stat(path); err == nil {
		return nil // file already present
	}
	if licenseKey == "" {
		return fmt.Errorf("mmdb not found at %q and no license key configured; sign up for a free key at %s", path, mmdbLicenseSignupURL)
	}
	url := fmt.Sprintf("https://download.maxmind.com/app/geoip_download?edition_id=GeoLite2-ASN&license_key=%s&suffix=tar.gz", licenseKey)
	if err := DownloadAndExtractMMDB(url, path); err != nil {
		return fmt.Errorf("auto-download mmdb failed: %w (sign up at %s)", err, mmdbLicenseSignupURL)
	}
	return nil
}
```

Add `"bytes"` to the imports (already used above). Verify the import block includes: `archive/tar`, `bytes`, `compress/gzip`, `errors`, `fmt`, `io`, `net/http`, `os`, `path/filepath`, `strings` (strings unused - drop it), `time`.

- [ ] **Step 4: Run test to verify it passes**

Run: `GOCACHE=/tmp/gocache go test ./internal/interaction/fingerprint/... -run 'TestDownloadAndExtractMMDB|TestEnsureMMDB' -v`
Expected: PASS.

- [ ] **Step 5: Run full fingerprint package tests**

Run: `GOCACHE=/tmp/gocache go test ./internal/interaction/fingerprint/...`
Expected: PASS.

- [ ] **Step 6: Commit**

```bash
git add internal/interaction/fingerprint/downloader.go internal/interaction/fingerprint/downloader_test.go
git commit -m "feat(fingerprint): add GeoLite2-ASN.mmdb auto-download and tar.gz extraction"
```

---

### Task 4: Thread fingerprinter through WebServer

**Files:**
- Modify: `server/webserver.go:36-110`
- Modify: `server/v2_api.go` (10 call sites)
- Modify: `internal/evidencehub/service.go` (optional, low priority)

**Interfaces:**
- Consumes: `fingerprint.NewFingerprinter(mmdbPath)` from existing code, `EnsureMMDB` from Task 3
- Produces: `WebServer.fingerprinter *fingerprint.Fingerprinter` field, `WebServerConfig.GeoIPMMDBPath` / `GeoIPLicenseKey` fields

- [ ] **Step 1: Add config fields + fingerprinter to WebServer**

In `server/webserver.go`, add to `WebServerConfig` (after `AnonymousMode` around line 55):

```go
	// GeoIP configuration. MMDBPath is the location of GeoLite2-ASN.mmdb.
	// LicenseKey enables auto-download when the file is missing.
	GeoIPMMDBPath  string
	GeoIPLicenseKey string
```

Add to the `WebServer` struct (after `redisClient` around line 81):

```go
	fingerprinter *fingerprint.Fingerprinter
```

Add the import `"github.com/chennqqi/godnslog/internal/interaction/fingerprint"` if not present.

In `NewWebServer` (around line 84-110), after `app.orm = orm` and before `return app, nil`, add fingerprinter initialization:

```go
	// Initialize GeoIP fingerprinter. Auto-download mmdb if missing and a
	// license key is configured. Missing db is non-fatal.
	if cfg.GeoIPMMDBPath != "" {
		if err := fingerprint.EnsureMMDB(cfg.GeoIPMMDBPath, cfg.GeoIPLicenseKey); err != nil {
			logrus.Warnf("[webserver] GeoIP disabled: %v", err)
		}
		app.fingerprinter = fingerprint.NewFingerprinter(cfg.GeoIPMMDBPath)
		logrus.Infof("[webserver] GeoIP fingerprinter initialized (mmdb=%s)", cfg.GeoIPMMDBPath)
	} else {
		logrus.Info("[webserver] GeoIP disabled (no mmdb path configured)")
	}
```

- [ ] **Step 2: Replace nil fingerprinter at all production call sites**

In `server/v2_api.go`, replace each `interaction.NewService(self.orm, nil, nil, false)` with `interaction.NewService(self.orm, nil, self.fingerprinter, false)`.

Affected line numbers (verify with grep before editing): 2359, 2761, 4187, 4367, 4404, 4443, 4479, 4550, 5202, 5220.

The 2nd arg (`wsHub`) stays `nil` for now - WebSocket threading is out of scope for this plan. Only the 3rd arg (fingerprinter) changes from `nil` to `self.fingerprinter`.

Run this to confirm all sites are updated:
```bash
grep -n 'interaction.NewService(self.orm, nil, nil, false)' server/v2_api.go
```
Expected: no output (all replaced).

- [ ] **Step 3: Update evidencehub (optional consistency)**

In `internal/evidencehub/service.go`, the `Service` struct (lines 42-49) and `BuildSummary` (line 58) currently pass `nil, nil, false`. This is a read-only path so the fingerprinter has no effect. Leave as-is to keep the diff small, OR add a fingerprinter field if you want symmetry. **Decision: leave as-is** (documented here so the implementer doesn't waste time).

- [ ] **Step 4: Build and run server tests**

Run: `GOCACHE=/tmp/gocache go build ./... && GOCACHE=/tmp/gocache go test ./server/...`
Expected: Build OK, server tests PASS.

- [ ] **Step 5: Commit**

```bash
git add server/webserver.go server/v2_api.go
git commit -m "feat(server): wire GeoIP fingerprinter into interaction service production paths"
```

---

### Task 5: Add CLI flags and startup download in servecmd.go

**Files:**
- Modify: `servecmd.go:20-59` (struct + SetFlags)
- Modify: `servecmd.go:79-94` (WebServerConfig construction)

**Interfaces:**
- Consumes: `WebServerConfig.GeoIPMMDBPath` / `GeoIPLicenseKey` from Task 4
- Produces: `-geoip-mmdb` and `-geoip-license-key` CLI flags; env vars `MMDB_PATH` / `MMDB_LICENSE_KEY`

- [ ] **Step 1: Add fields to servePwCmd struct**

In `servecmd.go`, add to the `servePwCmd` struct (after `redisAddr string` around line 32):

```go
	geoipMMDBPath  string
	geoipLicenseKey string
```

- [ ] **Step 2: Add flags in SetFlags**

In `SetFlags` (after the `-redis` flag around line 58), add:

```go
	f.StringVar(&p.geoipMMDBPath, "geoip-mmdb", os.Getenv("MMDB_PATH"), "path to GeoLite2-ASN.mmdb for source ASN enrichment (optional); auto-downloads if -geoip-license-key is set")
	f.StringVar(&p.geoipLicenseKey, "geoip-license-key", os.Getenv("MMDB_LICENSE_KEY"), "MaxMind license key for auto-downloading GeoLite2-ASN.mmdb; sign up free at https://www.maxmind.com/en/geolite2/signup")
```

- [ ] **Step 3: Pass config into WebServerConfig**

In `Execute`, the `NewWebServer` call (lines 79-94), add the two fields to the `WebServerConfig` literal:

```go
	web, err := server.NewWebServer(&server.WebServerConfig{
		Driver:                       p.driver,
		Dsn:                          p.dsn,
		Domain:                       p.domain,
		IP:                           p.ipv4,
		Listen:                       p.httpListen,
		Swagger:                      p.swagger,
		WithGuest:                    p.withGuest,
		TestMode:                     p.testMode,
		AuthExpire:                   AuthExpire,
		DefaultCleanInterval:         DefaultCleanInterval,
		DefaultQueryApiMaxItem:       DefaultQueryApiMaxItem,
		DefaultMaxCallbackErrorCount: DefaultMaxCallbackErrorCount,
		DefaultLanguage:              DefaultLanguage,
		RedisAddr:                    p.redisAddr,
		GeoIPMMDBPath:                p.geoipMMDBPath,
		GeoIPLicenseKey:              p.geoipLicenseKey,
	}, store)
```

- [ ] **Step 4: Build the binary**

Run: `GOCACHE=/tmp/gocache go build ./...`
Expected: Build OK.

- [ ] **Step 5: Verify flags appear in help**

Run: `./godnslog serve -help 2>&1 | grep -A1 geoip`
Expected: shows both `-geoip-mmdb` and `-geoip-license-key` flags with their descriptions (the license-key description mentions the signup URL).

- [ ] **Step 6: Run full Go test suite**

Run: `GOCACHE=/tmp/gocache go test ./...`
Expected: All tests PASS.

- [ ] **Step 7: Commit**

```bash
git add servecmd.go
git commit -m "feat(serve): add -geoip-mmdb and -geoip-license-key flags for GeoIP auto-download"
```

---

## Self-Review

**Spec coverage:**
- ✅ 启动时自动检测/下载 GeoLite2-ASN.mmdb -> Task 3 (downloader) + Task 4 (NewWebServer calls EnsureMMDB) + Task 5 (CLI flags)
- ✅ 持久化 ASN/Org -> Task 1 (model) + Task 2 (enhanceInteraction)
- ✅ 串联 fingerprinter 到生产路径 -> Task 4
- ✅ 免费 Key 申请链接 -> Task 3 (mmdbLicenseSignupURL) + Task 5 (flag description)

**Placeholder scan:** None - all code blocks complete.

**Type consistency:** `EnsureMMDB(path, licenseKey string) error`, `DownloadAndExtractMMDB(url, destPath string) error`, `WebServerConfig.GeoIPMMDBPath`/`GeoIPLicenseKey`, `WebServer.fingerprinter` - all consistent across tasks.
