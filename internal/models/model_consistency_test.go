package models

import (
	"testing"
	"time"
)

// TestModelConsistency verifies that unified model changes don't break existing behavior
func TestModelConsistency(t *testing.T) {
	t.Run("Case model has required fields", func(t *testing.T) {
		c := Case{
			ID:          "test-id",
			Title:       "Test Case",
			Description: "Test Description",
			Status:      CaseStatusActive,
			Tags:        Tags{"tag1", "tag2"},
			CreatedBy:   "user-1",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if c.ID == "" {
			t.Error("Case ID should not be empty")
		}
		if c.Title == "" {
			t.Error("Case Title should not be empty")
		}
		if c.Status == "" {
			t.Error("Case Status should not be empty")
		}
		if c.CreatedBy == "" {
			t.Error("Case CreatedBy should not be empty")
		}
	})

	t.Run("Payload model has required fields", func(t *testing.T) {
		expiresAt := time.Now().Add(24 * time.Hour)
		p := Payload{
			ID:               "test-id",
			CaseID:           "case-1",
			Token:            "test-token",
			TemplateID:       "ssrf",
			TemplateRendered: "http://{{token}}.example.com",
			Variables:        Variables{"domain": "example.com"},
			Status:           "active",
			ExpiresAt:        &expiresAt,
			CreatedAt:        time.Now(),
		}

		if p.ID == "" {
			t.Error("Payload ID should not be empty")
		}
		if p.CaseID == "" {
			t.Error("Payload CaseID should not be empty")
		}
		if p.Token == "" {
			t.Error("Payload Token should not be empty")
		}
		if p.TemplateID == "" {
			t.Error("Payload TemplateID should not be empty")
		}
		if p.TemplateRendered == "" {
			t.Error("Payload TemplateRendered should not be empty")
		}
		if p.Status == "" {
			t.Error("Payload Status should not be empty")
		}
	})

	t.Run("APIKey model has required fields", func(t *testing.T) {
		expiresAt := time.Now().Add(30 * 24 * time.Hour)
		workspaceID := "workspace-1"
		k := APIKey{
			ID:            "test-id",
			Key:           "test-api-key",
			KeyPrefix:     "test-api",
			Name:          "Test API Key",
			Scopes:        Scopes{"case:read", "payload:read"},
			WorkspaceID:   &workspaceID,
			RiskTolerance: "medium",
			IsAgent:       false,
			ExpiresAt:     &expiresAt,
			CreatedBy:     "user-1",
			CreatedAt:     time.Now(),
		}

		if k.ID == "" {
			t.Error("APIKey ID should not be empty")
		}
		if k.Key == "" {
			t.Error("APIKey Key should not be empty")
		}
		if k.KeyPrefix == "" {
			t.Error("APIKey KeyPrefix should not be empty")
		}
		if k.Name == "" {
			t.Error("APIKey Name should not be empty")
		}
		if k.CreatedBy == "" {
			t.Error("APIKey CreatedBy should not be empty")
		}
		if k.RiskTolerance == "" {
			t.Error("APIKey RiskTolerance should not be empty")
		}
	})

	t.Run("APIKey IsAgent flag is present", func(t *testing.T) {
		k := APIKey{
			ID:      "test-id",
			Key:     "test-api-key",
			Name:    "Agent API Key",
			Scopes:  Scopes{"cases:read", "payloads:read"},
			IsAgent: true,
		}

		if !k.IsAgent {
			t.Error("APIKey IsAgent should be true")
		}
	})

	t.Run("Response wrapper has correct structure", func(t *testing.T) {
		resp := SuccessResponse(map[string]string{"key": "value"})

		if resp.Code != CodeSuccess {
			t.Errorf("Response code should be %d, got %d", CodeSuccess, resp.Code)
		}
		if resp.Message != MessageSuccess {
			t.Errorf("Response message should be %s, got %s", MessageSuccess, resp.Message)
		}
		if resp.Data == nil {
			t.Error("Response data should not be nil")
		}
	})

	t.Run("Error response has correct structure", func(t *testing.T) {
		resp := BadRequestResponse("Invalid input")

		if resp.Code != CodeBadRequest {
			t.Errorf("Error code should be %d, got %d", CodeBadRequest, resp.Code)
		}
		if resp.Message != "Invalid input" {
			t.Errorf("Error message should be 'Invalid input', got %s", resp.Message)
		}
	})

	t.Run("ListResponse has correct structure", func(t *testing.T) {
		listResp := ListResponse{
			Items:      []Case{},
			Total:      10,
			Page:       1,
			PageSize:   20,
			TotalPages: 1,
		}

		if listResp.Total != 10 {
			t.Errorf("Total should be 10, got %d", listResp.Total)
		}
		if listResp.Page != 1 {
			t.Errorf("Page should be 1, got %d", listResp.Page)
		}
		if listResp.PageSize != 20 {
			t.Errorf("PageSize should be 20, got %d", listResp.PageSize)
		}
		if listResp.TotalPages != 1 {
			t.Errorf("TotalPages should be 1, got %d", listResp.TotalPages)
		}
	})
}

// TestResponseStatusCodes verifies response code constants
func TestResponseStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected int
	}{
		{"Success code", CodeSuccess, 0},
		{"Bad request code", CodeBadRequest, 400},
		{"Unauthorized code", CodeUnauthorized, 401},
		{"Forbidden code", CodeForbidden, 403},
		{"Not found code", CodeNotFound, 404},
		{"Internal server error code", CodeInternalServerError, 500},
		{"Service unavailable code", CodeServiceUnavailable, 502},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code != tt.expected {
				t.Errorf("Expected code %d, got %d", tt.expected, tt.code)
			}
		})
	}
}
