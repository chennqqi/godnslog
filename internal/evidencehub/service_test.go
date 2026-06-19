package evidencehub

import (
	"strings"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestBuildSummaryFromScannerRun(t *testing.T) {
	engine := setupEvidenceSummaryEngine(t)
	defer engine.Close()

	caseID, payloadID, scannerRunID := seedEvidenceSummaryData(t, engine)
	service := NewService(engine)

	summary, err := service.BuildSummary(&SummaryRequest{ScannerRunID: scannerRunID}, "http://godnslog.local")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary.Scope.ScannerRunID != scannerRunID {
		t.Fatalf("expected scanner_run_id %q, got %q", scannerRunID, summary.Scope.ScannerRunID)
	}
	if summary.Scope.CaseID != caseID {
		t.Fatalf("expected case_id %q, got %q", caseID, summary.Scope.CaseID)
	}
	if summary.Scope.PayloadID != payloadID {
		t.Fatalf("expected payload_id %q, got %q", payloadID, summary.Scope.PayloadID)
	}
	if summary.Evidence == nil {
		t.Fatal("expected evidence")
	}
	if summary.Evidence.InteractionCount != 2 {
		t.Fatalf("expected 2 interactions, got %d", summary.Evidence.InteractionCount)
	}
	if len(summary.ScannerRuns) != 1 {
		t.Fatalf("expected one related scanner run, got %d", len(summary.ScannerRuns))
	}
	if len(summary.PackageHashes) != 1 || summary.PackageHashes[0] == "" {
		t.Fatalf("expected package hash list, got %#v", summary.PackageHashes)
	}
	if summary.SummaryHash == "" {
		t.Fatal("expected summary hash")
	}
	if len(summary.SummaryHash) != 64 {
		t.Fatalf("expected sha256 hex hash length 64, got %d", len(summary.SummaryHash))
	}
	if len(summary.NextActions) == 0 {
		t.Fatal("expected next actions")
	}
}

func TestBuildSummaryHashIsDeterministic(t *testing.T) {
	engine := setupEvidenceSummaryEngine(t)
	defer engine.Close()

	_, payloadID, _ := seedEvidenceSummaryData(t, engine)
	service := NewService(engine)

	summary1, err := service.BuildSummary(&SummaryRequest{PayloadID: payloadID}, "http://godnslog.local")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	summary2, err := service.BuildSummary(&SummaryRequest{PayloadID: payloadID}, "http://godnslog.local")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if summary1.SummaryHash != summary2.SummaryHash {
		t.Fatalf("expected deterministic summary hash, got %q and %q", summary1.SummaryHash, summary2.SummaryHash)
	}
}

func TestBuildSummaryRequiresScope(t *testing.T) {
	engine := setupEvidenceSummaryEngine(t)
	defer engine.Close()

	service := NewService(engine)
	_, err := service.BuildSummary(&SummaryRequest{}, "http://godnslog.local")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "case_id, payload_id, or scanner_run_id") {
		t.Fatalf("expected scope error, got %v", err)
	}
}

func setupEvidenceSummaryEngine(t *testing.T) *xorm.Engine {
	t.Helper()

	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	if err := engine.Sync2(new(models.Case), new(models.Payload), new(models.ScannerRun), new(models.Interaction), new(models.AuditLog)); err != nil {
		t.Fatalf("failed to sync schema: %v", err)
	}
	return engine
}

func seedEvidenceSummaryData(t *testing.T, engine *xorm.Engine) (string, string, string) {
	t.Helper()

	now := time.Date(2026, 6, 19, 12, 0, 0, 0, time.UTC)
	caseID := "case-summary-1"
	payloadID := "payload-summary-1"
	scannerRunID := "scanner-summary-1"
	token := "tok-summary"
	domain := "tok-summary.example.com"
	path := "/callback"
	method := "GET"

	if _, err := engine.Insert(&models.Case{
		ID:        caseID,
		Title:     "Evidence Summary Case",
		Status:    "active",
		CreatedBy: "1",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("failed to insert case: %v", err)
	}
	if _, err := engine.Insert(&models.Payload{
		ID:               payloadID,
		CaseID:           caseID,
		Token:            token,
		TemplateID:       "ssrf-basic",
		TemplateRendered: "http://tok-summary.example.com/callback",
		Variables:        models.Variables{},
		Status:           "active",
		CreatedBy:        "1",
		CreatedAt:        now,
		UpdatedAt:        now,
	}); err != nil {
		t.Fatalf("failed to insert payload: %v", err)
	}
	if _, err := engine.Insert(&models.ScannerRun{
		ID:             scannerRunID,
		CaseID:         caseID,
		PayloadID:      payloadID,
		Scanner:        models.ScannerBurp,
		Target:         "https://target.example",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodBurpExtension,
		Command:        "Burp Suite Extension Package",
		Jsonl:          `{"scanner":"burp","payload_id":"payload-summary-1"}`,
		PackageHash:    "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		PackageManifest: models.ScannerPackageManifest{
			SchemaVersion:  "scanner-package.v1",
			Scanner:        models.ScannerBurp,
			DeliveryMethod: models.DeliveryMethodBurpExtension,
			PackageHash:    "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			HashAlgorithm:  "sha256",
		},
		Status:    models.ScannerRunStatusObserved,
		CreatedBy: "1",
		CreatedAt: now,
		UpdatedAt: now,
	}); err != nil {
		t.Fatalf("failed to insert scanner run: %v", err)
	}

	interactions := []models.Interaction{
		{
			ID:        "interaction-summary-1",
			Type:      models.InteractionTypeDNS,
			CaseID:    &caseID,
			PayloadID: &payloadID,
			Token:     &token,
			Timestamp: now.Add(1 * time.Minute),
			SourceIP:  "198.51.100.10",
			Domain:    &domain,
			RawData:   "dns callback",
			CreatedAt: now.Add(1 * time.Minute),
		},
		{
			ID:        "interaction-summary-2",
			Type:      models.InteractionTypeHTTP,
			CaseID:    &caseID,
			PayloadID: &payloadID,
			Token:     &token,
			Timestamp: now.Add(2 * time.Minute),
			SourceIP:  "198.51.100.11",
			Method:    &method,
			Path:      &path,
			RawData:   "http callback",
			CreatedAt: now.Add(2 * time.Minute),
		},
	}
	if _, err := engine.Insert(&interactions); err != nil {
		t.Fatalf("failed to insert interactions: %v", err)
	}

	return caseID, payloadID, scannerRunID
}
