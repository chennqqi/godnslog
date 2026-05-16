package mcp

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func newTestServer(handler func(*http.Request) string) *Server {
	server := NewServer("http://godnslog.test", "test-key")
	server.httpClient.Transport = roundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader(handler(req))),
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
		if r.Method != http.MethodPost || r.URL.Path != "/cases" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
			t.Fatalf("expected bearer auth header, got %q", got)
		}
		return `{"data":{"id":"case-1"}}`
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
		if r.Method != http.MethodPost || r.URL.Path != "/payloads" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{"data":{"id":"payload-1","token":"tok-1"}}`
	})

	args := map[string]interface{}{
		"template":   "http://{{.Token}}.example.com",
		"case_id":    "case-1",
		"expires_in": "24h",
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
		if r.Method != http.MethodGet || r.URL.Path != "/interactions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("case_id"); got != "case-1" {
			t.Fatalf("expected case_id case-1, got %q", got)
		}
		return `{"data":{"items":[]}}`
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
}

// TestWaitForInteractionTool tests waitForInteraction tool
func TestWaitForInteractionTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method != http.MethodGet || r.URL.Path != "/interactions" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{"data":{"items":[{"id":"interaction-1","token":"test-token"}]}}`
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
}

// TestSummarizeEvidenceTool tests summarizeEvidence tool
func TestSummarizeEvidenceTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method != http.MethodGet || r.URL.Path != "/evidence/case-1/summary" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{"data":{"confidence":"medium"}}`
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
}

// TestExportReportTool tests exportReport tool
func TestExportReportTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method != http.MethodPost || r.URL.Path != "/evidence/export" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{"data":{"format":"markdown"}}`
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
}

// TestRevokeTokenTool tests revokeToken tool
func TestRevokeTokenTool(t *testing.T) {
	server := newTestServer(func(r *http.Request) string {
		if r.Method != http.MethodDelete || r.URL.Path != "/apikeys/key-1" {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{"data":{"revoked":true}}`
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

	server := newTestServer(func(r *http.Request) string {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/cases":
			return `{"data":{"id":"case-1","title":"SSRF probe"}}`
		case r.Method == http.MethodPost && r.URL.Path == "/payloads":
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode payload request: %v", err)
			}
			createdPayloadCaseID, _ = body["case_id"].(string)
			return `{"data":{"id":"payload-1","token":"tok-1","value":"https://tok-1.example.com/callback"}}`
		default:
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		return `{}`
	})

	result, err := server.createOASTProbe(context.Background(), map[string]interface{}{
		"title":              "SSRF probe",
		"target":             "https://target.example",
		"template":           "ssrf-url",
		"expected_protocols": []interface{}{"dns", "http"},
		"variables":          map[string]interface{}{"path": "/callback"},
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
	if createdPayloadCaseID != "case-1" {
		t.Fatalf("expected payload to be created for case-1, got %q", createdPayloadCaseID)
	}

	data, ok := toolResult.Data.(map[string]interface{})
	if !ok {
		t.Fatal("Data should be a map")
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
	if data["agent_next_action"] == "" {
		t.Fatal("agent_next_action should guide the agent")
	}
}
