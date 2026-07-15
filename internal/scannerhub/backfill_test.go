package scannerhub

import (
	"encoding/json"
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestParseNucleiJSONL(t *testing.T) {
	raw := `{"template-id":"ssrf-basic","info":{"name":"SSRF Vulnerability","severity":"high"},"matched-at":"http://example.com/vuln"}
{"template-id":"xxe-basic","info":{"name":"XXE Vulnerability","severity":"critical"},"matched-at":"http://example.com/xml"}`

	findings, err := parseNucleiJSONL(raw)
	assert.NoError(t, err)
	assert.Len(t, findings, 2)

	assert.Equal(t, "ssrf-basic", findings[0].RuleID)
	assert.Equal(t, "SSRF Vulnerability", findings[0].Name)
	assert.Equal(t, "high", findings[0].Severity)
	assert.Equal(t, "http://example.com/vuln", findings[0].URL)

	assert.Equal(t, "xxe-basic", findings[1].RuleID)
	assert.Equal(t, "XXE Vulnerability", findings[1].Name)
	assert.Equal(t, "critical", findings[1].Severity)
}

func TestParseNucleiJSONL_EmptyInput(t *testing.T) {
	findings, err := parseNucleiJSONL("")
	assert.NoError(t, err)
	assert.Len(t, findings, 0)
}

func TestParseNucleiJSONL_InvalidLines(t *testing.T) {
	raw := `invalid json line
{"template-id":"valid","info":{"name":"Valid Finding","severity":"medium"}}
another invalid line`

	findings, err := parseNucleiJSONL(raw)
	assert.NoError(t, err)
	assert.Len(t, findings, 1)
	assert.Equal(t, "valid", findings[0].RuleID)
}

func TestParseSARIF(t *testing.T) {
	sarifJSON := `{
		"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/main/Schemata/sarif-schema-2.1.0.json",
		"version": "2.1.0",
		"runs": [{
			"tool": {
				"driver": {
					"name": "nuclei",
					"rules": [{
						"id": "ssrf-basic",
						"shortDescription": {"text": "SSRF Detection"},
						"fullDescription": {"text": "Server-side request forgery detected"}
					}]
				}
			},
			"results": [{
				"ruleId": "ssrf-basic",
				"level": "error",
				"message": {"text": "SSRF found at /api/fetch"},
				"locations": [{
					"physicalLocation": {
						"artifactLocation": {"uri": "http://example.com/api/fetch"}
					}
				}]
			}]
		}]
	}`

	findings, err := parseSARIF(sarifJSON)
	assert.NoError(t, err)
	assert.Len(t, findings, 1)

	assert.Equal(t, "ssrf-basic", findings[0].RuleID)
	assert.Equal(t, "SSRF found at /api/fetch", findings[0].Name)
	assert.Equal(t, "high", findings[0].Severity)
	assert.Equal(t, "http://example.com/api/fetch", findings[0].URL)
	assert.Equal(t, "Server-side request forgery detected", findings[0].Description)
}

func TestParseSARIF_InvalidJSON(t *testing.T) {
	_, err := parseSARIF("invalid json")
	assert.Error(t, err)
}

func TestSarifLevelToSeverity(t *testing.T) {
	assert.Equal(t, "high", sarifLevelToSeverity("error"))
	assert.Equal(t, "medium", sarifLevelToSeverity("warning"))
	assert.Equal(t, "low", sarifLevelToSeverity("note"))
	assert.Equal(t, "info", sarifLevelToSeverity("none"))
	assert.Equal(t, "info", sarifLevelToSeverity("unknown"))
}

func TestSeverityToSARIFLevel(t *testing.T) {
	assert.Equal(t, "error", severityToSARIFLevel("high"))
	assert.Equal(t, "error", severityToSARIFLevel("critical"))
	assert.Equal(t, "warning", severityToSARIFLevel("medium"))
	assert.Equal(t, "note", severityToSARIFLevel("low"))
	assert.Equal(t, "none", severityToSARIFLevel("info"))
}

func TestGenerateSARIF(t *testing.T) {
	scannerRun := &models.ScannerRun{
		Scanner:  models.ScannerNuclei,
		Target:   "http://example.com",
		Template: "ssrf-basic",
	}
	findings := []ScanFinding{
		{RuleID: "ssrf-basic", Name: "SSRF Vulnerability", Severity: "high"},
		{RuleID: "xxe-basic", Name: "XXE Vulnerability", Severity: "critical"},
	}

	output, err := GenerateSARIF(scannerRun, findings)
	assert.NoError(t, err)
	assert.NotNil(t, output)

	var sarif map[string]interface{}
	err = json.Unmarshal(output, &sarif)
	assert.NoError(t, err)

	assert.Equal(t, "2.1.0", sarif["version"])

	runs, ok := sarif["runs"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, runs, 1)

	run := runs[0].(map[string]interface{})
	tool := run["tool"].(map[string]interface{})
	driver := tool["driver"].(map[string]interface{})
	assert.Equal(t, "godnslog-nuclei", driver["name"])

	rules, ok := driver["rules"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, rules, 2)

	results, ok := run["results"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, results, 2)
}

func TestBackfillResults_NucleiJSONL(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	// Create test case and payload
	caseItem := createTestCase(t, engine, "case-backfill")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-backfill")

	// Create scanner run
	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	assert.NoError(t, err)

	// Backfill results
	rawResults := `{"template-id":"ssrf-basic","info":{"name":"SSRF Found","severity":"high"},"matched-at":"http://example.com/vuln"}`
	backfillReq := &BackfillResultsRequest{
		Format:       "jsonl",
		RawResults:   rawResults,
		ScannerRunID: scannerRun.ID,
	}

	result, err := service.BackfillResults(backfillReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Equal(t, scannerRun.ID, result.ScannerRunID)
	assert.Len(t, result.Findings, 1)
	assert.Equal(t, "ssrf-basic", result.Findings[0].RuleID)
	assert.Equal(t, caseItem.ID, result.Findings[0].CaseID)
	assert.Equal(t, payload.ID, result.Findings[0].PayloadID)
}

func TestBackfillResults_SARIF(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-sarif")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-sarif")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	assert.NoError(t, err)

	sarifInput := `{"version":"2.1.0","runs":[{"tool":{"driver":{"name":"nuclei","rules":[]}},"results":[{"ruleId":"xss-reflected","level":"warning","message":{"text":"XSS found"}}]}]}`
	backfillReq := &BackfillResultsRequest{
		Format:       "sarif",
		RawResults:   sarifInput,
		ScannerRunID: scannerRun.ID,
	}

	result, err := service.BackfillResults(backfillReq)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, 1, result.FindingsCount)
	assert.Equal(t, "xss-reflected", result.Findings[0].RuleID)
}

func TestBackfillResults_UnsupportedFormat(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	caseItem := createTestCase(t, engine, "case-fmt")
	payload := createTestPayload(t, engine, caseItem.ID, "payload-fmt")

	req := &models.ScannerRunCreateRequest{
		CaseID:         caseItem.ID,
		PayloadID:      payload.ID,
		Scanner:        models.ScannerNuclei,
		Target:         "http://example.com",
		Template:       "ssrf-basic",
		DeliveryMethod: models.DeliveryMethodNucleiJsonl,
	}
	scannerRun, err := service.CreateScannerRun(req, "1", "http://example.com")
	assert.NoError(t, err)

	backfillReq := &BackfillResultsRequest{
		Format:       "xml",
		RawResults:   "<results/>",
		ScannerRunID: scannerRun.ID,
	}

	_, err = service.BackfillResults(backfillReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported results format")
}

func TestBackfillResults_ScannerRunNotFound(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	service := NewService(engine)

	backfillReq := &BackfillResultsRequest{
		Format:       "jsonl",
		RawResults:   "{}",
		ScannerRunID: "nonexistent-id",
	}

	_, err := service.BackfillResults(backfillReq)
	assert.Error(t, err)
	assert.Equal(t, ErrScannerRunNotFound, err)
}

func TestGetString(t *testing.T) {
	m := map[string]interface{}{
		"foo": "bar",
		"nested": map[string]interface{}{
			"key": "value",
		},
	}

	assert.Equal(t, "bar", getString(m, "foo"))
	assert.Equal(t, "value", getString(m, "nested", "key"))
	assert.Equal(t, "", getString(m, "nonexistent"))
	assert.Equal(t, "", getString(m, "nested", "nonexistent"))
}

func TestParseBurpJSON(t *testing.T) {
	raw := `{"issues":[
		{"name":"SQL Injection","severity":"high","confidence":"certain","host":"https://example.com","path":"/api/users"},
		{"name":"XSS","severity":"medium","confidence":"firm","host":"https://example.com","path":"/search"},
		{"name":"","severity":"info","confidence":"tentative","host":"https://example.com","path":"/empty"}
	]}`

	findings, err := parseBurpJSON(raw)
	assert.NoError(t, err)
	assert.Len(t, findings, 2)

	assert.Equal(t, "SQL Injection", findings[0].Name)
	assert.Equal(t, "high", findings[0].Severity)
	assert.Equal(t, "https://example.com/api/users", findings[0].URL)
	assert.Equal(t, "burp-1", findings[0].ID)

	assert.Equal(t, "XSS", findings[1].Name)
	assert.Equal(t, "medium", findings[1].Severity)
}

func TestParseBurpJSON_Invalid(t *testing.T) {
	_, err := parseBurpJSON("invalid json")
	assert.Error(t, err)
}

func TestParseBurpJSON_Empty(t *testing.T) {
	raw := `{"issues":[]}`
	findings, err := parseBurpJSON(raw)
	assert.NoError(t, err)
	assert.Len(t, findings, 0)
}

func TestParseXrayJSON(t *testing.T) {
	raw := `[
		{"vuln_id":"xss","plugin":"xss/reflected","severity":"high","url":"http://example.com/search?q=test","payload":"<script>","detail":"Reflected XSS in search"},
		{"vuln_id":"","plugin":"sqli/error","severity":"medium","url":"http://example.com/api","payload":"' OR 1=1--","detail":""},
		{"vuln_id":"","plugin":"","severity":"low","url":"http://example.com/other","payload":""}
	]`

	findings, err := parseXrayJSON(raw)
	assert.NoError(t, err)
	assert.Len(t, findings, 2)

	assert.Equal(t, "xss", findings[0].RuleID)
	assert.Equal(t, "xss/reflected", findings[0].Name)
	assert.Equal(t, "high", findings[0].Severity)
	assert.Equal(t, "http://example.com/search?q=test", findings[0].URL)
	assert.Equal(t, "Reflected XSS in search", findings[0].Description)

	assert.Equal(t, "sqli/error", findings[1].RuleID)
	assert.Equal(t, "sqli/error", findings[1].Name)
	assert.Equal(t, "Payload: ' OR 1=1--", findings[1].Description)
}

func TestParseXrayJSON_Invalid(t *testing.T) {
	_, err := parseXrayJSON("not json at all")
	assert.Error(t, err)
}

func TestParseXrayJSON_Empty(t *testing.T) {
	findings, err := parseXrayJSON(`[]`)
	assert.NoError(t, err)
	assert.Len(t, findings, 0)
}
