package scannerhub

import (
	"strings"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupTestEngine(t *testing.T) *xorm.Engine {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Drop tables first to ensure clean schema
	engine.Exec("DROP TABLE IF EXISTS audit_logs")
	engine.Exec("DROP TABLE IF EXISTS scanner_runs")
	engine.Exec("DROP TABLE IF EXISTS interactions")
	engine.Exec("DROP TABLE IF EXISTS payloads")
	engine.Exec("DROP TABLE IF EXISTS cases")

	// Sync tables individually to ensure proper schema
	err = engine.Sync2(new(models.Case))
	if err != nil {
		t.Fatalf("Failed to sync Case table: %v", err)
	}
	err = engine.Sync2(new(models.Payload))
	if err != nil {
		t.Fatalf("Failed to sync Payload table: %v", err)
	}
	err = engine.Sync2(new(models.ScannerRun))
	if err != nil {
		t.Fatalf("Failed to sync ScannerRun table: %v", err)
	}
	err = engine.Sync2(new(models.Interaction))
	if err != nil {
		t.Fatalf("Failed to sync Interaction table: %v", err)
	}
	err = engine.Sync2(new(models.AuditLog))
	if err != nil {
		t.Fatalf("Failed to sync AuditLog table: %v", err)
	}

	return engine
}

func createTestCase(t *testing.T, engine *xorm.Engine, id string) *models.Case {
	caseItem := &models.Case{
		ID:        id,
		Title:     "Test Case",
		Status:    "active",
		CreatedBy: "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(caseItem)
	if err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}
	return caseItem
}

func createTestPayload(t *testing.T, engine *xorm.Engine, caseID, id string) *models.Payload {
	payload := &models.Payload{
		ID:               id,
		CaseID:           caseID,
		Token:            "test-token-" + id,
		TemplateID:       "ssrf-basic",
		TemplateRendered: "http://test-token-" + id + ".example.com/",
		Variables:        models.Variables{},
		Status:           "active",
		CreatedBy:        "1",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	_, err := engine.Insert(payload)
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}
	return payload
}

func TestCreateScannerRun(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	// Create test case and payload
	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}
	if scannerRun == nil {
		t.Fatal("Expected scanner run, got nil")
	}

	// Verify basic fields
	if scannerRun.CaseID != caseItem.ID {
		t.Errorf("Expected case ID %s, got %s", caseItem.ID, scannerRun.CaseID)
	}
	if scannerRun.PayloadID != payload.ID {
		t.Errorf("Expected payload ID %s, got %s", payload.ID, scannerRun.PayloadID)
	}
	if scannerRun.Scanner != models.ScannerNuclei {
		t.Errorf("Expected scanner %s, got %s", models.ScannerNuclei, scannerRun.Scanner)
	}
	if scannerRun.Status != models.ScannerRunStatusCreated {
		t.Errorf("Expected status %s, got %s", models.ScannerRunStatusCreated, scannerRun.Status)
	}

	// Verify command is generated
	if scannerRun.Command == "" {
		t.Error("Expected command to be generated")
	}
	if !strings.Contains(scannerRun.Command, "nuclei") {
		t.Error("Expected command to contain 'nuclei'")
	}

	// Verify JSONL is generated
	if scannerRun.Jsonl == "" {
		t.Error("Expected JSONL to be generated")
	}
	if !strings.Contains(scannerRun.Jsonl, "nuclei") {
		t.Error("Expected JSONL to contain 'nuclei'")
	}

	// Verify audit log was created for scanner_run.created
	var auditLog models.AuditLog
	has, err := engine.Where("action = ? AND resource_type = ?", "scanner_run.created", "scanner_run").Get(&auditLog)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if !has {
		t.Error("Expected audit log entry for scanner_run.created")
	}
	if auditLog.UserID == nil || *auditLog.UserID != "1" {
		t.Error("Expected audit log to have user_id = 1")
	}
	if auditLog.Details == nil {
		t.Error("Expected audit log to have details")
	}
}

func TestCreateScannerRunWithInvalidCase(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	req := &models.ScannerRunCreateRequest{
		CaseID:         "non-existent-case",
		PayloadID:      "payload-123",
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	_, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != ErrInvalidCase {
		t.Fatalf("Expected ErrInvalidCase, got %v", err)
	}
}

func TestCreateScannerRunWithInvalidPayload(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	// Create test case but not payload
	caseItem := createTestCase(t, engine, "case-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      "non-existent-payload",
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	_, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != ErrInvalidPayload {
		t.Fatalf("Expected ErrInvalidPayload, got %v", err)
	}
}

func TestCreateScannerRunWithPayloadNotInCase(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	// Create test case and payload in different cases
	caseItem1 := createTestCase(t, engine, "case-123")
	caseItem2 := createTestCase(t, engine, "case-456")
	payload := createTestPayload(t, engine, caseItem1.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem2.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	_, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != ErrPayloadNotInCase {
		t.Fatalf("Expected ErrPayloadNotInCase, got %v", err)
	}
}

func TestCreateScannerRunWithInvalidScanner(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        "invalid-scanner",
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	_, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != ErrInvalidScanner {
		t.Fatalf("Expected ErrInvalidScanner, got %v", err)
	}
}

func TestCreateScannerRunWithInvalidDeliveryMethod(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: "invalid-delivery",
	}

	_, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != ErrInvalidDelivery {
		t.Fatalf("Expected ErrInvalidDelivery, got %v", err)
	}
}

func TestGetScannerRunByID(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	created, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}

	retrieved, err := service.GetScannerRunByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get scanner run: %v", err)
	}
	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestGetScannerRunByIDNotFound(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	_, err := service.GetScannerRunByID("non-existent-id")
	if err != ErrScannerRunNotFound {
		t.Fatalf("Expected ErrScannerRunNotFound, got %v", err)
	}
}

func TestGetScannerRunDetail(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	created, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}

	detail, err := service.GetScannerRunDetail(created.ID, "http://example.com")
	if err != nil {
		t.Fatalf("Failed to get scanner run detail: %v", err)
	}

	// Verify basic fields are populated
	if detail.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, detail.ID)
	}

	// Verify URLs
	if !strings.Contains(detail.InteractionsURL, payload.ID) {
		t.Error("Expected interactions URL to contain payload ID")
	}
	if !strings.Contains(detail.EvidenceURL, payload.ID) {
		t.Error("Expected evidence URL to contain payload ID")
	}
}

func TestListScannerRuns(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	// Create multiple scanner runs
	for i := 0; i < 5; i++ {
		req := &models.ScannerRunCreateRequest{
			CaseID:         caseItem.ID,
			PayloadID:      payload.ID,
			Scanner:        models.ScannerNuclei,
			Target:         "http://example.com",
			Template:       "ssrf-basic",
			DeliveryMethod: models.DeliveryMethodNucleiJsonl,
		}
		_, err := service.CreateScannerRun(req, "1", "http://example.com")
		if err != nil {
			t.Fatalf("Failed to create scanner run: %v", err)
		}
	}

	resp, err := service.ListScannerRuns("", "", "", "", 1, 10)
	if err != nil {
		t.Fatalf("Failed to list scanner runs: %v", err)
	}
	if len(resp.Items) != 5 {
		t.Errorf("Expected 5 scanner runs, got %d", len(resp.Items))
	}
	if resp.Total != 5 {
		t.Errorf("Expected total 5, got %d", resp.Total)
	}
}

func TestListScannerRunsWithCaseFilter(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem1 := createTestCase(t, engine, "case-123")
	caseItem2 := createTestCase(t, engine, "case-456")
	payload1 := createTestPayload(t, engine, caseItem1.ID, "payload-123")
	payload2 := createTestPayload(t, engine, caseItem2.ID, "payload-456")

	// Create scanner runs for both cases
	req1 := &models.ScannerRunCreateRequest{
		CaseID:         caseItem1.ID,
		PayloadID:      payload1.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	_, err := service.CreateScannerRun(req1, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}

	req2 := &models.ScannerRunCreateRequest{
		CaseID:         caseItem2.ID,
		PayloadID:      payload2.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	_, err = service.CreateScannerRun(req2, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}

	// List all scanner runs without filters
	resp, err := service.ListScannerRuns("", "", "", "", 1, 10)
	if err != nil {
		t.Fatalf("Failed to list scanner runs: %v", err)
	}
	if len(resp.Items) != 2 {
		t.Errorf("Expected 2 scanner runs, got %d", len(resp.Items))
	}
}

func TestUpdateScannerRunStatus(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	created, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}

	updateReq := &models.ScannerRunUpdateStatusRequest{
		Status: models.ScannerRunStatusDistributed,
	}

	err = service.UpdateScannerRunStatus(created.ID, updateReq, "1")
	if err != nil {
		t.Fatalf("Failed to update scanner run status: %v", err)
	}

	retrieved, err := service.GetScannerRunByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated scanner run: %v", err)
	}
	if retrieved.Status != models.ScannerRunStatusDistributed {
		t.Errorf("Expected status %s, got %s", models.ScannerRunStatusDistributed, retrieved.Status)
	}

	// Verify audit log was created for scanner_run.status_updated
	var auditLog models.AuditLog
	has, err := engine.Where("action = ? AND resource_type = ?", "scanner_run.status_updated", "scanner_run").Get(&auditLog)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if !has {
		t.Error("Expected audit log entry for scanner_run.status_updated")
	}
	if auditLog.UserID == nil || *auditLog.UserID != "1" {
		t.Error("Expected audit log to have user_id = 1")
	}
	if auditLog.Details == nil {
		t.Error("Expected audit log to have details")
	}
}

func TestUpdateScannerRunStatusInvalidTransition(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-123")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-123")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}

	created, err := service.CreateScannerRun(req, "1", "http://example.com")
	if err != nil {
		t.Fatalf("Failed to create scanner run: %v", err)
	}

	// Try to skip status transition from created to evidenced
	updateReq := &models.ScannerRunUpdateStatusRequest{
		Status: models.ScannerRunStatusEvidenced,
	}

	err = service.UpdateScannerRunStatus(created.ID, updateReq, "1")
	if err == nil {
		t.Fatal("Expected error for invalid status transition")
	}
}
