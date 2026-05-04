package mcp

import (
	"context"
	"testing"
)

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
	server := NewServer("http://localhost:8080", "test-key")

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
	server := NewServer("http://localhost:8080", "test-key")

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
	server := NewServer("http://localhost:8080", "test-key")

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
	server := NewServer("http://localhost:8080", "test-key")

	args := map[string]interface{}{
		"token":   "test-token",
		"timeout": float64(300),
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
	server := NewServer("http://localhost:8080", "test-key")

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
	server := NewServer("http://localhost:8080", "test-key")

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
	server := NewServer("http://localhost:8080", "test-key")

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
