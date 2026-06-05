package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestServer(handler func(*http.Request) string) *Server {
	server := NewServer("http://godnslog.test", "test-key")
	server.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// First call custom handler to see if it wants to handle this request
		customResponse := handler(req)
		if customResponse != "" {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(customResponse)),
				Request:    req,
			}, nil
		}

		// Default handlers for auth/info and audit/logs (only if custom handler didn't handle)
		// Handle auth/info request for permission checking
		if strings.Contains(req.URL.Path, "/api/v2/auth/info") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success","data":{"user_id":"test-user","api_key_id":"key-1","api_key_prefix":"key_","scopes":["admin:all"],"is_agent":false,"risk_tolerance":"high"}}`)),
				Request:    req,
			}, nil
		}
		// Handle audit/logs POST request for permission denied logging
		if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/api/v2/audit/logs") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success"}`)),
				Request:    req,
			}, nil
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success"}`)),
			Request:    req,
		}, nil
	})
	return server
}

// TestNewServer tests server creation
func TestNewServer(t *testing.T) {
	apiURL := "http://localhost:8080"
	apiKey := "test-api-key"

	server := NewServer(apiURL, apiKey)

	if server == nil {
		t.Fatal("Server should not be nil")
	}

	if server.apiURL != apiURL {
		t.Fatalf("Expected API URL '%s', got '%s'", apiURL, server.apiURL)
	}

	if server.apiKey != apiKey {
		t.Fatalf("Expected API key '%s', got '%s'", apiKey, server.apiKey)
	}
}

// TestToolStruct tests tool structure
func TestToolStruct(t *testing.T) {
	tool := Tool{
		Name:        "test-tool",
		Description: "Test tool description",
		Execute: func(ctx context.Context, args map[string]interface{}) (interface{}, error) {
			return map[string]interface{}{"result": "success"}, nil
		},
	}

	if tool.Name != "test-tool" {
		t.Fatalf("Expected name 'test-tool', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Fatal("Description should not be empty")
	}

	if tool.Execute == nil {
		t.Fatal("Execute function should not be nil")
	}
}

// TestToolResultStruct tests tool result structure
func TestToolResultStruct(t *testing.T) {
	result := ToolResult{
		Success: true,
		Data:    map[string]interface{}{"key": "value"},
		Error:   "",
	}

	if !result.Success {
		t.Fatal("Success should be true")
	}

	if result.Data == nil {
		t.Fatal("Data should not be nil")
	}
}

// TestCreateCaseTool tests createCase tool
func TestCreateCaseTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/cases" {
			if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
				t.Fatalf("expected bearer auth header, got %q", got)
			}
			return `{"data":{"id":"case-1"}}`
		}
		return ""
	})

	args := map[string]interface{}{
		"title":       "Test Case",
		"description": "Test description",
		"target":      "example.com",
		"tags":        []string{"test"},
	}

	// This is a simplified test - in production, use proper context
	result, err := server.createCase(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}
}

// TestCreatePayloadTool tests createPayload tool
func TestCreatePayloadTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method != http.MethodPost || r.URL.Path != "/api/v2/payloads" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{"data":{"id":"payload-1","token":"tok-1"}}`
	})

	args := map[string]interface{}{
		"template_id": "ssrf-basic",
		"case_id":     "case-1",
		"expires_in":  "24h",
	}

	result, err := server.createPayload(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}
}

// TestListInteractionsTool tests listInteractions tool
func TestListInteractionsTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodGet && r.URL.Path == "/api/v2/interactions" {
			if got := r.URL.Query().Get("case_id"); got != "case-1" {
				t.Fatalf("expected case_id case-1, got %q", got)
			}
			// Return realistic response with semantic fields
			return `{"code":0,"message":"success","data":{"items":[{"id":"interaction-1","type":"dns","token":"test-token","case_id":"case-1","payload_id":"payload-1","timestamp":"2026-05-17T00:00:00Z","source_ip":"127.0.0.1"}],"total":1,"page":1,"page_size":50,"total_pages":1}}`
		}
		return ""
	})

	args := map[string]interface{}{
		"case_id": "case-1",
		"limit":   float64(50),
	}

	result, err := server.listInteractions(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	// Verify the data contains the raw API response
	if toolResult.Data == nil {
		t.Fatal("Expected data to be present")
	}

	// The data should be the raw API response structure
	data, ok := toolResult.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}

	// Verify the response has the expected structure from v2 API
	if data["code"] == nil {
		t.Error("Expected code field")
	}
	if data["data"] == nil {
		t.Error("Expected data field")
	}

	// Verify the nested data structure
	responseData, ok := data["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data.data to be a map")
	}

	items, ok := responseData["items"].([]interface{})
	if !ok {
		t.Fatal("Expected items to be a list")
	}

	if len(items) == 0 {
		t.Fatal("Expected at least one item")
	}

	item, ok := items[0].(map[string]interface{})
	if !ok {
		t.Fatal("Expected item to be a map")
	}

	// Verify required semantic fields
	if item["id"] == nil {
		t.Error("Expected id field")
	}
	if item["type"] == nil {
		t.Error("Expected type field")
	}
	if item["token"] == nil {
		t.Error("Expected token field")
	}
	if item["case_id"] == nil {
		t.Error("Expected case_id field")
	}
	if item["payload_id"] == nil {
		t.Error("Expected payload_id field")
	}
}

// TestWaitForInteractionTool tests waitForInteraction tool
func TestWaitForInteractionTool(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		server := newTestServer(func(r *http.Request) string {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v2/interactions" {
				return `{"code":0,"message":"success","data":{"items":[{"id":"interaction-1","type":"dns","token":"test-token","case_id":"case-1","payload_id":"payload-1"}],"total":1}}`
			}
			return ""
		})

		args := map[string]interface{}{
			"token":   "test-token",
			"timeout": float64(1),
		}

		result, err := server.waitForInteraction(nil, args)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		toolResult, ok := result.(ToolResult)
		if !ok {
			t.Fatal("Result should be ToolResult type")
		}

		if !toolResult.Success {
			t.Fatalf("Expected success, got error: %s", toolResult.Error)
		}

		// Verify data is present (semantic assertion that interaction data is returned)
		if toolResult.Data == nil {
			t.Fatal("Expected data to be present")
		}
	})

	t.Run("timeout", func(t *testing.T) {
		server := newTestServer(func(r *http.Request) string {
			if r.Method == http.MethodGet && r.URL.Path == "/api/v2/interactions" {
				return `{"code":0,"message":"success","data":{"items":[],"total":0}}`
			}
			return ""
		})

		args := map[string]interface{}{
			"token":   "test-token",
			"timeout": float64(0.1),
		}

		result, err := server.waitForInteraction(nil, args)

		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		toolResult, ok := result.(ToolResult)
		if !ok {
			t.Fatal("Result should be ToolResult type")
		}

		// Timeout should return success with timeout message (timeout path assertion)
		if !toolResult.Success {
			t.Fatalf("Expected success even on timeout, got error: %s", toolResult.Error)
		}

		if toolResult.Data == nil {
			t.Fatal("Expected data to be present")
		}

		data, ok := toolResult.Data.(map[string]interface{})
		if !ok {
			t.Fatal("Expected data to be a map")
		}

		// Verify timeout message is present
		if data["message"] == nil {
			t.Error("Expected message field on timeout")
		}
	})
}

// TestSummarizeEvidenceTool tests summarizeEvidence tool
func TestSummarizeEvidenceTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/evidence/generate" {
			return `{"code":0,"message":"success","data":{"evidence":{"id":"evidence-1","case_id":"case-1","evidence_strength":"medium","confidence":75,"interaction_count":4,"unique_sources":2,"explainability":"Captured 4 interactions from 2 unique sources. Evidence strength: medium. Confidence: 75%.","timeline":[{"type":"interaction","description":"DNS interaction","timestamp":"2026-05-18T00:00:00Z"}]},"format":"json","content":"{\"id\":\"evidence-1\"}","metadata":{"interaction_count":4}}}`
		}
		return ""
	})

	args := map[string]interface{}{
		"case_id": "case-1",
	}

	result, err := server.summarizeEvidence(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	// Verify that the tool returns structured evidence data
	if toolResult.Data == nil {
		t.Fatal("Expected data to be present")
	}

	// The data should be the evidence field (structured data)
	evidence, ok := toolResult.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map (structured evidence)")
	}

	// Verify evidence fields are present
	if evidence["evidence_strength"] == nil {
		t.Error("Expected evidence_strength field")
	}
	if evidence["confidence"] == nil {
		t.Error("Expected confidence field")
	}
	if evidence["interaction_count"] == nil {
		t.Error("Expected interaction_count field")
	}
	if evidence["explainability"] == nil {
		t.Error("Expected explainability field")
	}
	if evidence["timeline"] == nil {
		t.Error("Expected timeline field")
	}
}

// TestSummarizeEvidenceToolPayloadOnly tests summarizeEvidence tool with payload_id only
func TestSummarizeEvidenceToolPayloadOnly(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/evidence/generate" {
			return `{"code":0,"message":"success","data":{"evidence":{"id":"evidence-1","payload_id":"payload-1","evidence_strength":"medium","confidence":75,"interaction_count":4,"unique_sources":2,"explainability":"Captured 4 interactions from 2 unique sources. Evidence strength: medium. Confidence: 75%.","timeline":[{"type":"interaction","description":"DNS interaction","timestamp":"2026-05-18T00:00:00Z"}]},"format":"json","content":"{\"id\":\"evidence-1\"}","metadata":{"interaction_count":4}}}`
		}
		return ""
	})

	args := map[string]interface{}{
		"payload_id": "payload-1",
	}

	result, err := server.summarizeEvidence(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	if toolResult.Data == nil {
		t.Fatal("Expected data to be present")
	}

	evidence, ok := toolResult.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map (structured evidence)")
	}

	if evidence["evidence_strength"] == nil {
		t.Error("Expected evidence_strength field")
	}
	if evidence["confidence"] == nil {
		t.Error("Expected confidence field")
	}
}

// TestExportReportToolPayloadOnly tests exportReport tool with payload_id only
func TestExportReportToolPayloadOnly(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/evidence/generate" {
			return `{"code":0,"message":"success","data":{"format":"markdown","content":"# Evidence Report\n\n**Payload ID**: payload-1\n**Evidence Strength**: medium\n**Confidence**: 75%\n\n## Interactions\n\n- DNS interaction from 192.168.1.1\n- HTTP interaction from 192.168.1.2\n","metadata":{"interaction_count":2}}}`
		}
		// Let default handler handle auth/info and other requests
		return ""
	})

	args := map[string]interface{}{
		"payload_id": "payload-1",
		"format":     "markdown",
	}

	result, err := server.exportReport(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	if toolResult.Data == nil {
		t.Fatal("Expected data to be present")
	}

	content, ok := toolResult.Data.(string)
	if !ok {
		t.Fatal("Expected data to be a string (content)")
	}

	if content == "" {
		t.Error("Expected non-empty markdown content")
	}
}

// TestExportReportToolWithAgentRunID tests exportReport tool with agent_run_id
func TestExportReportToolWithAgentRunID(t *testing.T) {
	agentRunID := "agent-run-123"
	var reviewAPICalled bool
	var operationPosted bool
	var operationBody string

	server := newTestServer(func(r *http.Request) string {
		// Handle auth/info request for permission checking
		if strings.Contains(r.URL.Path, "/api/v2/auth/info") {
			return `{"code":0,"message":"success","data":{"user_id":"test-user","api_key_id":"key-1","api_key_prefix":"key_","scopes":["agent:export_report"],"is_agent":true,"risk_tolerance":"medium"}}`
		}
		if r.Method == "GET" && r.URL.Path == fmt.Sprintf("/api/v2/agent-runs/%s/review", agentRunID) {
			reviewAPICalled = true
			// Verify format query parameter
			format := r.URL.Query().Get("format")
			if format != "json" && format != "markdown" {
				t.Errorf("Expected format query parameter, got %s", format)
			}
			return `{"code":0,"message":"success","data":{"id":"` + agentRunID + `","interaction_summary":{"total":5,"dns_count":3,"http_count":2,"unique_sources":2},"evidence":{"evidence_strength":"high","confidence":85,"interaction_count":5,"unique_sources":2},"generated_at":"2026-05-31T10:00:00Z","format":"json"}}`
		}
		if r.Method == "POST" && r.URL.Path == fmt.Sprintf("/api/v2/agent-runs/%s/operations", agentRunID) {
			operationPosted = true
			body, _ := io.ReadAll(r.Body)
			operationBody = string(body)
			return `{"code":0,"message":"success"}`
		}
		t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		return ""
	})

	args := map[string]interface{}{
		"agent_run_id": agentRunID,
		"format":       "json",
	}

	result, err := server.exportReport(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	if !reviewAPICalled {
		t.Fatal("Expected Review API to be called with agent_run_id")
	}

	if !operationPosted {
		t.Fatal("Expected agent operation to be posted")
	}

	// Verify operation body contains review summary
	if operationBody == "" {
		t.Fatal("Operation body should not be empty")
	}

	var op map[string]interface{}
	if err := json.Unmarshal([]byte(operationBody), &op); err != nil {
		t.Fatalf("Failed to parse operation body: %v", err)
	}

	if op["action"] != "export_report" {
		t.Errorf("Expected action export_report, got %v", op["action"])
	}

	resultMap, ok := op["result"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected result to be a map")
	}

	if resultMap["agent_run_id"] != agentRunID {
		t.Errorf("Expected agent_run_id %s in result, got %v", agentRunID, resultMap["agent_run_id"])
	}

	if resultMap["format"] != "json" {
		t.Errorf("Expected format json in result, got %v", resultMap["format"])
	}

	if resultMap["interaction_count"] == nil {
		t.Error("Expected interaction_count in result")
	}

	if resultMap["evidence_strength"] == nil {
		t.Error("Expected evidence_strength in result")
	}
}

// TestExportReportToolWithoutScope tests exportReport tool without required scope
func TestExportReportToolWithoutScope(t *testing.T) {
	var reviewAPICalled bool
	var auditLogPosted bool

	server := newTestServer(func(r *http.Request) string {
		// Override auth/info to return limited scope
		if strings.Contains(r.URL.Path, "/api/v2/auth/info") {
			return `{"code":0,"message":"success","data":{"user_id":"test-user","api_key_id":"key-1","api_key_prefix":"key_","scopes":["agent:read_interactions"],"is_agent":true,"risk_tolerance":"medium"}}`
		}
		if r.Method == "GET" && strings.Contains(r.URL.Path, "/api/v2/agent-runs/") {
			reviewAPICalled = true
		}
		if r.Method == "POST" && r.URL.Path == "/api/v2/audit/logs" {
			auditLogPosted = true
		}
		return ""
	})

	args := map[string]interface{}{
		"agent_run_id": "agent-run-123",
		"format":       "json",
	}

	result, err := server.exportReport(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	// Should have permission denied error
	if toolResult.Success {
		t.Fatal("Expected permission denied, but tool succeeded")
	}

	if reviewAPICalled {
		t.Fatal("Review API should not be called without required scope")
	}

	if !auditLogPosted {
		t.Fatal("Audit log should be posted when permission is denied")
	}
}

// TestSummarizeEvidenceToolEmptyParams tests summarizeEvidence tool with empty params
func TestSummarizeEvidenceToolEmptyParams(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/evidence/generate" {
			return `{"code":1,"message":"Either case_id or payload_id is required"}`
		}
		return ""
	})

	args := map[string]interface{}{}

	result, err := server.summarizeEvidence(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if toolResult.Success {
		t.Fatal("Expected failure with empty params")
	}

	if toolResult.Error == "" {
		t.Error("Expected error message")
	}
}

// TestExportReportToolEmptyParams tests exportReport tool with empty params
func TestExportReportToolEmptyParams(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/evidence/generate" {
			return `{"code":1,"message":"Either case_id or payload_id is required"}`
		}
		// Let default handler handle auth/info and other requests
		return ""
	})

	args := map[string]interface{}{
		"format": "markdown",
	}

	result, err := server.exportReport(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if toolResult.Success {
		t.Fatal("Expected failure with empty params")
	}

	if toolResult.Error == "" {
		t.Error("Expected error message")
	}
}

// TestExportReportTool tests exportReport tool
func TestExportReportTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodPost && r.URL.Path == "/api/v2/evidence/generate" {
			return `{"code":0,"message":"success","data":{"format":"markdown","content":"# Evidence Report\n\n**Case ID**: case-1\n**Evidence Strength**: medium\n**Confidence**: 75%\n\n## Interactions\n\n- DNS interaction from 192.168.1.1\n- HTTP interaction from 192.168.1.2\n","metadata":{"interaction_count":2}}}`
		}
		// Let default handler handle auth/info and other requests
		return ""
	})

	args := map[string]interface{}{
		"case_id": "case-1",
		"format":  "markdown",
	}

	result, err := server.exportReport(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	// Verify that the tool returns the content field from API response
	if toolResult.Data == nil {
		t.Fatal("Expected data to be present")
	}

	content, ok := toolResult.Data.(string)
	if !ok {
		t.Fatal("Expected data to be a string (content)")
	}

	// Verify markdown content is present
	if content == "" {
		t.Error("Expected non-empty markdown content")
	}
}

// TestRevokeTokenTool tests revokeToken tool
func TestRevokeTokenTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method == http.MethodDelete && r.URL.Path == "/api/v2/apikeys/key-1" {
			return `{"data":{"revoked":true}}`
		}
		return ""
	})

	args := map[string]interface{}{
		"key_id": "key-1",
	}

	result, err := server.revokeToken(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}
}

// TestCreateOASTProbeToolCreatesCaseThenPayload tests the agent-native high level probe flow.
func TestCreateOASTProbeToolCreatesCaseThenPayload(t *testing.T) {
	var createdPayloadCaseID string
	var payloadRequestBody map[string]interface{}

	server := newTestServer(func(r *http.Request) string {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/cases":
			return `{"data":{"id":"case-1","title":"SSRF probe"}}`
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/payloads":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode payload request: %v", err)
			}
			createdPayloadCaseID, _ = body["case_id"].(string)
			payloadRequestBody = body
			return `{"data":{"id":"payload-1","token":"tok-1","value":"https://tok-1.example.com/callback"}}`
		case r.Method == http.MethodPost && r.URL.Path == "/api/v2/agent-runs":
			return `{"data":{"id":"agent-run-1","agent_id":"agent-1","status":"created"}}`
		case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/api/v2/agent-runs/") && strings.HasSuffix(r.URL.Path, "/operations"):
			return `{"data":{"id":"op-1","action":"create_oast_probe"}}`
		case r.Method == http.MethodPut && strings.HasPrefix(r.URL.Path, "/api/v2/agent-runs/") && strings.HasSuffix(r.URL.Path, "/status"):
			return `{"data":{"id":"agent-run-1","status":"running"}}`
		default:
			return ""
		}
	})

	result, err := server.createOASTProbe(context.Background(), map[string]interface{}{
		"title":              "SSRF probe",
		"target":             "https://target.example",
		"template_id":        "ssrf-basic",
		"expected_protocols": []interface{}{"dns", "http"},
		"variables":          map[string]interface{}{"path": "/callback"},
		"agent_id":           "agent-1",
		"expires_in":         "24h",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}
	if !toolResult.Success {
		t.Fatalf("Expected success, got error: %s", toolResult.Error)
	}

	data, ok := toolResult.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
	}

	if agentRunID, exists := data["agent_run_id"]; !exists || agentRunID == "" {
		t.Fatal("agent_run_id should be populated when agent_id is provided")
	}
	if createdPayloadCaseID != "case-1" {
		t.Fatalf("expected payload to be created for case-1, got %q", createdPayloadCaseID)
	}
	if data["probe_id"] == "" {
		t.Fatal("probe_id should be populated")
	}
	if data["case_id"] != "case-1" {
		t.Fatalf("expected case_id case-1, got %v", data["case_id"])
	}
	if data["payload_id"] != "payload-1" {
		t.Fatalf("expected payload_id payload-1, got %v", data["payload_id"])
	}

	// Strong assertion: verify request body conversion
	// Check that expires_in was converted to expires_at (RFC3339 format)
	if payloadRequestExpiresIn, exists := payloadRequestBody["expires_in"]; exists {
		t.Errorf("payload request should not contain expires_in, got %v", payloadRequestExpiresIn)
	}
	if payloadRequestExpiresAt, exists := payloadRequestBody["expires_at"]; exists {
		// If expires_in was not provided, expires_at may be empty, which is acceptable
		if payloadRequestExpiresAt.(string) != "" {
			// Verify expires_at is a valid RFC3339 timestamp when present
			if _, err := time.Parse(time.RFC3339, payloadRequestExpiresAt.(string)); err != nil {
				t.Errorf("expires_at should be valid RFC3339 format, got %v: %v", payloadRequestExpiresAt, err)
			}
		}
	}
	// Check that expected_protocols array was converted to expected_protocol single value
	if payloadRequestExpectedProtocols, exists := payloadRequestBody["expected_protocols"]; exists {
		t.Errorf("payload request should not contain expected_protocols array, got %v", payloadRequestExpectedProtocols)
	}
	if payloadRequestExpectedProtocol, exists := payloadRequestBody["expected_protocol"]; !exists {
		t.Error("payload request should contain expected_protocol after conversion")
	} else {
		// Verify expected_protocol is a string, not an array
		if _, ok := payloadRequestExpectedProtocol.(string); !ok {
			t.Errorf("expected_protocol should be a string, got %T", payloadRequestExpectedProtocol)
		}
		// Verify it's the first value from the original array
		if payloadRequestExpectedProtocol != "dns" {
			t.Errorf("expected_protocol should be 'dns' (first value), got %v", payloadRequestExpectedProtocol)
		}
	}
}

// TestPermissionGateMissingScope tests that tools are blocked when API key lacks required scope
func TestPermissionGateMissingScope(t *testing.T) {
	var businessAPICalled bool

	// Override to capture auth/info and business API calls
	server := NewServer("http://godnslog.test", "test-key")
	server.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Handle auth/info request for permission checking
		if strings.Contains(req.URL.Path, "/api/v2/auth/info") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success","data":{"user_id":"test-user","api_key_id":"key-123","api_key_prefix":"gdl_abc","scopes":["case:read"],"is_agent":true,"risk_tolerance":"medium"}}`)),
				Request:    req,
			}, nil
		}
		// Track if business API was called
		if strings.Contains(req.URL.Path, "/api/v2/cases") || strings.Contains(req.URL.Path, "/api/v2/payloads") {
			businessAPICalled = true
		}
		// Handle audit/logs POST request
		if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/api/v2/audit/logs") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success"}`)),
				Request:    req,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":{"id":"test-id"}}`)),
			Request:    req,
		}, nil
	})

	// Try to call createCase which requires case:create scope
	args := map[string]interface{}{
		"title":       "Test Case",
		"description": "Test description",
		"target":      "example.com",
	}

	result, err := server.createCase(nil, args)

	// Should succeed with permission denied error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	// Should have permission denied error
	if toolResult.Success {
		t.Fatal("Expected permission denied, but tool succeeded")
	}

	// Business API should NOT have been called
	if businessAPICalled {
		t.Fatal("Business API should not be called when scope is missing")
	}
}

// TestPermissionGateExceedsRiskTolerance tests that tools are blocked when tool risk exceeds API key tolerance
func TestPermissionGateExceedsRiskTolerance(t *testing.T) {
	var businessAPICalled bool

	// Override to capture auth/info and business API calls
	server := NewServer("http://godnslog.test", "test-key")
	server.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Handle auth/info request for permission checking
		if strings.Contains(req.URL.Path, "/api/v2/auth/info") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success","data":{"user_id":"test-user","api_key_id":"key-123","api_key_prefix":"gdl_abc","scopes":["agent:revoke_token"],"is_agent":true,"risk_tolerance":"low"}}`)),
				Request:    req,
			}, nil
		}
		// Track if business API was called
		if strings.Contains(req.URL.Path, "/api/v2/apikeys") {
			businessAPICalled = true
		}
		// Handle audit/logs POST request
		if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/api/v2/audit/logs") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success"}`)),
				Request:    req,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":{"revoked":true}}`)),
			Request:    req,
		}, nil
	})

	// Try to call revokeToken which has high risk
	args := map[string]interface{}{
		"key_id": "key-1",
	}

	result, err := server.revokeToken(nil, args)

	// Should succeed with permission denied error
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	// Should have permission denied error
	if toolResult.Success {
		t.Fatal("Expected permission denied, but tool succeeded")
	}

	// Business API should NOT have been called
	if businessAPICalled {
		t.Fatal("Business API should not be called when risk tolerance is exceeded")
	}
}

// TestPermissionDeniedAuditLog tests that permission denied events are logged to audit
func TestPermissionDeniedAuditLog(t *testing.T) {
	var auditLogPosted bool
	var auditLogBody string

	// Override to capture auth/info and audit log POST
	server := NewServer("http://godnslog.test", "test-key")
	server.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Handle auth/info request for permission checking
		if strings.Contains(req.URL.Path, "/api/v2/auth/info") {
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success","data":{"user_id":"test-user","api_key_id":"key-123","api_key_prefix":"gdl_abc","scopes":["case:read"],"is_agent":true,"risk_tolerance":"medium"}}`)),
				Request:    req,
			}, nil
		}
		// Capture audit/log POST
		if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/api/v2/audit/logs") {
			auditLogPosted = true
			bodyBytes, _ := io.ReadAll(req.Body)
			auditLogBody = string(bodyBytes)
			return &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Body:       io.NopCloser(strings.NewReader(`{"code":0,"message":"success"}`)),
				Request:    req,
			}, nil
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(`{"data":{"id":"test-id"}}`)),
			Request:    req,
		}, nil
	})

	// Try to call createCase which requires case:create scope
	args := map[string]interface{}{
		"title":       "Test Case",
		"description": "Test description",
		"target":      "example.com",
	}

	result, err := server.createCase(nil, args)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	toolResult, ok := result.(ToolResult)
	if !ok {
		t.Fatal("Result should be ToolResult type")
	}

	// Should have permission denied error
	if toolResult.Success {
		t.Fatal("Expected permission denied, but tool succeeded")
	}

	// Audit log should have been posted
	if !auditLogPosted {
		t.Fatal("Audit log should be posted when permission is denied")
	}

	// Verify audit log contains permission denied information
	if auditLogBody == "" {
		t.Fatal("Audit log body should not be empty")
	}

	var auditLog map[string]interface{}
	if err := json.Unmarshal([]byte(auditLogBody), &auditLog); err != nil {
		t.Fatalf("Failed to parse audit log body: %v", err)
	}

	if action, ok := auditLog["action"].(string); !ok || action != "agent_permission.denied" {
		t.Fatalf("Expected action 'agent_permission.denied', got %v", auditLog["action"])
	}
}
