package scannerhub

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateScannerRunsFromSearch verifies creating scanner runs from search results.
func TestCreateScannerRunsFromSearch(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-search")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-search")

	req := &models.ScannerRunCreateFromSearchRequest{
		CaseID:    caseItem.ID,
		PayloadID: payload.ID,
		Source:    "shodan",
		Results: []models.ScanTargetItem{
			{IP: "10.0.0.1", Port: 8080, Protocol: "http"},
			{IP: "10.0.0.2", Port: 443, Protocol: "https"},
			{Hostname: "example.com", Port: 80},
		},
	}

	runs, err := service.CreateScannerRunsFromSearch(req, "1", "http://godnslog.local")
	assert.NoError(t, err)
	require.Len(t, runs, 3)
	assert.Equal(t, "10.0.0.1:8080", runs[0].Target)
	assert.Equal(t, "10.0.0.2:443", runs[1].Target)
	assert.Equal(t, "example.com:80", runs[2].Target)
}

// TestCreateScannerRunsFromSearch_EmptyResults verifies error when no results provided.
func TestCreateScannerRunsFromSearch_EmptyResults(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	req := &models.ScannerRunCreateFromSearchRequest{
		CaseID:    "case-x",
		PayloadID: "payload-x",
		Source:    "shodan",
		Results:   []models.ScanTargetItem{},
	}

	_, err := service.CreateScannerRunsFromSearch(req, "1", "http://godnslog.local")
	assert.Error(t, err)
	assert.Equal(t, ErrNoResults, err)
}

// TestCreateScannerRunsFromSearch_Defaults verifies default scanner/delivery values.
func TestCreateScannerRunsFromSearch_Defaults(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-defaults")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-defaults")

	req := &models.ScannerRunCreateFromSearchRequest{
		CaseID:    caseItem.ID,
		PayloadID: payload.ID,
		Source:    "fofa",
		Results:   []models.ScanTargetItem{{IP: "192.168.1.1"}},
	}

	runs, err := service.CreateScannerRunsFromSearch(req, "1", "http://godnslog.local")
	assert.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, models.ScannerNuclei, runs[0].Scanner)
	assert.Equal(t, models.DeliveryMethodNucleiJsonl, runs[0].DeliveryMethod)
	assert.Equal(t, "ssrf-basic", runs[0].Template)
}

// TestCreateScannerRunsFromSearch_InvalidCase verifies error for non-existent case.
func TestCreateScannerRunsFromSearch_InvalidCase(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	req := &models.ScannerRunCreateFromSearchRequest{
		CaseID:    "nonexistent",
		PayloadID: "payload-x",
		Source:    "shodan",
		Results:   []models.ScanTargetItem{{IP: "10.0.0.1"}},
	}

	_, err := service.CreateScannerRunsFromSearch(req, "1", "http://godnslog.local")
	assert.Error(t, err)
	assert.Equal(t, ErrInvalidCase, err)
}

// TestBuildTargetFromResult verifies target string construction from search result items.
func TestBuildTargetFromResult(t *testing.T) {
	tests := []struct {
		name string
		item models.ScanTargetItem
		want string
	}{
		{"hostname with port", models.ScanTargetItem{Hostname: "example.com", Port: 80}, "example.com:80"},
		{"hostname without port", models.ScanTargetItem{Hostname: "example.com"}, "example.com"},
		{"ip with port", models.ScanTargetItem{IP: "10.0.0.1", Port: 8080}, "10.0.0.1:8080"},
		{"ip without port", models.ScanTargetItem{IP: "10.0.0.1"}, "10.0.0.1"},
		{"hostname takes priority over ip", models.ScanTargetItem{Hostname: "example.com", IP: "10.0.0.1", Port: 80}, "example.com:80"},
		{"empty item", models.ScanTargetItem{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTargetFromResult(tt.item)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestMigrateScannerHub verifies database migration for scanner hub tables.
func TestMigrateScannerHub(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	err := MigrateScannerHub(engine)
	assert.NoError(t, err)
}

// TestBackfillResults_BurpJSON verifies backfill with Burp JSON format.
func TestBackfillResults_BurpJSON(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-burp")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-burp")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerBurp,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodBurpExtension,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	require.NoError(t, err)

	burpJSON := `{"issues":[{"name":"SSRF","severity":"High","host":"http://example.com","path":"/fetch"}]}`
	backfillReq := &BackfillResultsRequest{
		Format:       "burp-json",
		RawResults:   burpJSON,
		ScannerRunID: scannerRun.ID,
	}
	result, err := service.BackfillResults(backfillReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Equal(t, "SSRF", result.Findings[0].Name)
}

// TestBackfillResults_XrayJSON verifies backfill with xray JSON format.
func TestBackfillResults_XrayJSON(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-xray")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-xray")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerXray,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodXrayWebhook,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	require.NoError(t, err)

	xrayJSON := `[{"vuln_id":"xray-ssrf","plugin":"ssrf","severity":"high","url":"http://example.com/fetch","detail":"SSRF detected"}]`
	backfillReq := &BackfillResultsRequest{
		Format:       "xray-json",
		RawResults:   xrayJSON,
		ScannerRunID: scannerRun.ID,
	}
	result, err := service.BackfillResults(backfillReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Equal(t, "xray-ssrf", result.Findings[0].RuleID)
}

// TestBurpSeverityToStandard verifies Burp severity conversion.
func TestBurpSeverityToStandard(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"High", "High"},
		{"Medium", "Medium"},
		{"Low", "Low"},
		{"info", "info"},
		{"Certain", "medium"},
		{"Firm", "medium"},
		{"unknown", "info"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := burpSeverityToStandard(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// TestListScannerRuns_Pagination verifies listing with pagination.
func TestListScannerRuns_Pagination(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-list")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-list")

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
		require.NoError(t, err)
	}

	// Test pagination - list all
	resp, err := service.ListScannerRuns("", "", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), resp.Total)
	assert.Len(t, resp.Items, 5)

	// Test pagination - page 1 with size 2
	resp, err = service.ListScannerRuns("", "", "", "", 1, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), resp.Total)
	assert.Len(t, resp.Items, 2)
	assert.Equal(t, 3, resp.TotalPages)
}

// TestUpdateScannerRunStatus_ValidTransitions verifies allowed status transitions.
func TestUpdateScannerRunStatus_ValidTransitions(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-transition")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-transition")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	require.NoError(t, err)

	err = service.UpdateScannerRunStatus(scannerRun.ID, &models.ScannerRunUpdateStatusRequest{
		Status: models.ScannerRunStatusDistributed,
	}, "1")
	assert.NoError(t, err)

	err = service.UpdateScannerRunStatus(scannerRun.ID, &models.ScannerRunUpdateStatusRequest{
		Status: models.ScannerRunStatusObserved,
	}, "1")
	assert.NoError(t, err)

	err = service.UpdateScannerRunStatus(scannerRun.ID, &models.ScannerRunUpdateStatusRequest{
		Status: models.ScannerRunStatusEvidenced,
	}, "1")
	assert.NoError(t, err)
}

// TestUpdateScannerRunStatus_InvalidTransition verifies disallowed status transitions.
func TestUpdateScannerRunStatus_InvalidTransition(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-invalid")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-invalid")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	require.NoError(t, err)

	err = service.UpdateScannerRunStatus(scannerRun.ID, &models.ScannerRunUpdateStatusRequest{
		Status: models.ScannerRunStatusObserved,
	}, "1")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

// TestGetScannerRunDetail_WithInteractions verifies detail with interaction data.
func TestGetScannerRunDetail_WithInteractions(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-detail")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-detail")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://godnslog.local")
	require.NoError(t, err)

	interaction := &models.Interaction{
		ID:        "interaction-1",
		PayloadID: &payload.ID,
		Token:     &payload.Token,
		Type:      "dns",
	}
	_, err = engine.Insert(interaction)
	require.NoError(t, err)

	detail, err := service.GetScannerRunDetail(scannerRun.ID, "http://godnslog.local")
	assert.NoError(t, err)
	assert.Equal(t, 1, detail.InteractionCount)
	assert.NotNil(t, detail.LastInteractionAt)
	assert.Contains(t, detail.InteractionsURL, "payload-detail")
}
