package payload

import (
	"testing"
	"time"
)

// TestPayloadTableName tests table name
func TestPayloadTableName(t *testing.T) {
	p := Payload{}
	tableName := p.TableName()
	if tableName != "payloads" {
		t.Fatalf("Expected table name 'payloads', got '%s'", tableName)
	}
}

// TestVariablesScan tests Variables Scan method
func TestVariablesScan(t *testing.T) {
	var v Variables
	err := v.Scan(nil)
	if err != nil {
		t.Fatalf("Scan nil failed: %v", err)
	}

	if v != nil {
		t.Fatal("Expected nil after scanning nil")
	}
}

// TestVariablesValue tests Variables Value method
func TestVariablesValue(t *testing.T) {
	v := Variables{"key1": "value1", "key2": "value2"}
	_, err := v.Value()
	if err != nil {
		t.Fatalf("Value failed: %v", err)
	}
}

// TestPayloadCreateRequest tests payload creation request
func TestPayloadCreateRequest(t *testing.T) {
	req := &PayloadCreateRequest{
		CaseID:           "test-case-1",
		Template:         "ssrf",
		Variables:        map[string]string{"url": "http://example.com"},
		ExpectedProtocol: "http",
	}

	if req.CaseID == "" {
		t.Fatal("CaseID should not be empty")
	}

	if req.Template == "" {
		t.Fatal("Template should not be empty")
	}

	if len(req.Variables) != 1 {
		t.Fatalf("Expected 1 variable, got %d", len(req.Variables))
	}
}

// TestPayloadUpdateRequest tests payload update request
func TestPayloadUpdateRequest(t *testing.T) {
	req := &PayloadUpdateRequest{
		Status:           "deployed",
		ExpectedProtocol: "http",
	}

	if req.Status != "deployed" {
		t.Fatalf("Expected status 'deployed', got '%s'", req.Status)
	}
}

// TestPayloadListResponse tests payload list response
func TestPayloadListResponse(t *testing.T) {
	resp := &PayloadListResponse{
		Items:      []Payload{{Token: "token1"}, {Token: "token2"}},
		Total:      2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	if len(resp.Items) != 2 {
		t.Fatalf("Expected 2 items, got %d", len(resp.Items))
	}

	if resp.Total != 2 {
		t.Fatalf("Expected total 2, got %d", resp.Total)
	}
}

// TestPayloadModel tests payload model
func TestPayloadModel(t *testing.T) {
	now := time.Now()
	p := Payload{
		ID:               "test-id-1",
		CaseID:           "test-case-1",
		Token:            "test-token-abc123",
		Template:         "ssrf",
		Status:           "draft",
		ExpectedProtocol: "http",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if p.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if p.Token == "" {
		t.Fatal("Token should not be empty")
	}

	if p.Status != "draft" {
		t.Fatalf("Expected status 'draft', got '%s'", p.Status)
	}
}
