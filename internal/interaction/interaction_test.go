package interaction

import (
	"testing"
	"time"
)

// TestInteractionTableName tests table name
func TestInteractionTableName(t *testing.T) {
	i := Interaction{}
	tableName := i.TableName()
	if tableName != "interactions" {
		t.Fatalf("Expected table name 'interactions', got '%s'", tableName)
	}
}

// TestInteractionModel tests interaction model
func TestInteractionModel(t *testing.T) {
	now := time.Now()
	i := Interaction{
		ID:        "test-interaction-1",
		Type:      "dns",
		SourceIP:  "192.168.1.1",
		Timestamp: now,
	}

	if i.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if i.Type != "dns" {
		t.Fatalf("Expected type 'dns', got '%s'", i.Type)
	}

	if i.SourceIP == "" {
		t.Fatal("SourceIP should not be empty")
	}
}

// TestHeaders tests Headers type
func TestHeaders(t *testing.T) {
	h := Headers{
		"Content-Type": "application/json",
		"User-Agent":   "test",
	}

	if len(h) != 2 {
		t.Fatalf("Expected 2 headers, got %d", len(h))
	}

	if h["Content-Type"] != "application/json" {
		t.Fatalf("Expected 'application/json', got '%s'", h["Content-Type"])
	}
}

// TestInteractionListResponse tests list response model
func TestInteractionListResponse(t *testing.T) {
	resp := InteractionListResponse{
		Items:      []Interaction{{ID: "inter-1", Type: "dns"}},
		Total:      1,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	if resp.Total != 1 {
		t.Fatalf("Expected total 1, got %d", resp.Total)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(resp.Items))
	}
}

// TestExportRequest tests export request model
func TestExportRequest(t *testing.T) {
	req := ExportRequest{
		Format:     "markdown",
		IncludeRaw: true,
	}

	if req.Format != "markdown" {
		t.Fatalf("Expected format 'markdown', got '%s'", req.Format)
	}
}

// TestDeleteRequest tests delete request model
func TestDeleteRequest(t *testing.T) {
	req := DeleteRequest{
		IDs: []string{"id-1", "id-2"},
	}

	if len(req.IDs) != 2 {
		t.Fatalf("Expected 2 IDs, got %d", len(req.IDs))
	}
}
