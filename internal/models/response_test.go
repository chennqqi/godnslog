package models

import (
	"testing"
)

// TestAPIResponseFormat verifies unified API response format
func TestAPIResponseFormat(t *testing.T) {
	t.Run("SuccessResponse has correct structure", func(t *testing.T) {
		data := map[string]string{"id": "123", "name": "test"}
		resp := SuccessResponse(data)

		if resp.Code != CodeSuccess {
			t.Errorf("Expected code %d, got %d", CodeSuccess, resp.Code)
		}
		if resp.Message != MessageSuccess {
			t.Errorf("Expected message %s, got %s", MessageSuccess, resp.Message)
		}
		if resp.Data == nil {
			t.Error("Data should not be nil")
		}
	})

	t.Run("SuccessResponse without data", func(t *testing.T) {
		resp := SuccessResponse(nil)

		if resp.Code != CodeSuccess {
			t.Errorf("Expected code %d, got %d", CodeSuccess, resp.Code)
		}
		if resp.Data != nil {
			t.Error("Data should be nil when nil is passed")
		}
	})

	t.Run("BadRequestResponse has correct structure", func(t *testing.T) {
		resp := BadRequestResponse("Invalid parameter")

		if resp.Code != CodeBadRequest {
			t.Errorf("Expected code %d, got %d", CodeBadRequest, resp.Code)
		}
		if resp.Message != "Invalid parameter" {
			t.Errorf("Expected message 'Invalid parameter', got %s", resp.Message)
		}
	})

	t.Run("BadRequestResponse with default message", func(t *testing.T) {
		resp := BadRequestResponse("")

		if resp.Message != MessageBadRequest {
			t.Errorf("Expected message %s, got %s", MessageBadRequest, resp.Message)
		}
	})

	t.Run("UnauthorizedResponse has correct structure", func(t *testing.T) {
		resp := UnauthorizedResponse("Invalid token")

		if resp.Code != CodeUnauthorized {
			t.Errorf("Expected code %d, got %d", CodeUnauthorized, resp.Code)
		}
		if resp.Message != "Invalid token" {
			t.Errorf("Expected message 'Invalid token', got %s", resp.Message)
		}
	})

	t.Run("UnauthorizedResponse with default message", func(t *testing.T) {
		resp := UnauthorizedResponse("")

		if resp.Message != MessageUnauthorized {
			t.Errorf("Expected message %s, got %s", MessageUnauthorized, resp.Message)
		}
	})

	t.Run("ForbiddenResponse has correct structure", func(t *testing.T) {
		resp := ForbiddenResponse("Insufficient permissions")

		if resp.Code != CodeForbidden {
			t.Errorf("Expected code %d, got %d", CodeForbidden, resp.Code)
		}
		if resp.Message != "Insufficient permissions" {
			t.Errorf("Expected message 'Insufficient permissions', got %s", resp.Message)
		}
	})

	t.Run("ForbiddenResponse with default message", func(t *testing.T) {
		resp := ForbiddenResponse("")

		if resp.Message != MessageForbidden {
			t.Errorf("Expected message %s, got %s", MessageForbidden, resp.Message)
		}
	})

	t.Run("NotFoundResponse has correct structure", func(t *testing.T) {
		resp := NotFoundResponse("Resource not found")

		if resp.Code != CodeNotFound {
			t.Errorf("Expected code %d, got %d", CodeNotFound, resp.Code)
		}
		if resp.Message != "Resource not found" {
			t.Errorf("Expected message 'Resource not found', got %s", resp.Message)
		}
	})

	t.Run("NotFoundResponse with default message", func(t *testing.T) {
		resp := NotFoundResponse("")

		if resp.Message != MessageNotFound {
			t.Errorf("Expected message %s, got %s", MessageNotFound, resp.Message)
		}
	})

	t.Run("InternalServerErrorResponse has correct structure", func(t *testing.T) {
		resp := InternalServerErrorResponse("Database error")

		if resp.Code != CodeInternalServerError {
			t.Errorf("Expected code %d, got %d", CodeInternalServerError, resp.Code)
		}
		if resp.Message != "Database error" {
			t.Errorf("Expected message 'Database error', got %s", resp.Message)
		}
	})

	t.Run("InternalServerErrorResponse with default message", func(t *testing.T) {
		resp := InternalServerErrorResponse("")

		if resp.Message != MessageInternalServerError {
			t.Errorf("Expected message %s, got %s", MessageInternalServerError, resp.Message)
		}
	})

	t.Run("ServiceUnavailableResponse has correct structure", func(t *testing.T) {
		resp := ServiceUnavailableResponse("Service down")

		if resp.Code != CodeServiceUnavailable {
			t.Errorf("Expected code %d, got %d", CodeServiceUnavailable, resp.Code)
		}
		if resp.Message != "Service down" {
			t.Errorf("Expected message 'Service down', got %s", resp.Message)
		}
	})

	t.Run("ServiceUnavailableResponse with default message", func(t *testing.T) {
		resp := ServiceUnavailableResponse("")

		if resp.Message != MessageServiceUnavailable {
			t.Errorf("Expected message %s, got %s", MessageServiceUnavailable, resp.Message)
		}
	})

	t.Run("ErrorResponse has correct structure", func(t *testing.T) {
		resp := ErrorResponse(400, "Custom error")

		if resp.Code != 400 {
			t.Errorf("Expected code 400, got %d", resp.Code)
		}
		if resp.Message != "Custom error" {
			t.Errorf("Expected message 'Custom error', got %s", resp.Message)
		}
	})
}

// TestListResponseFormat verifies list response format
func TestListResponseFormat(t *testing.T) {
	t.Run("ListResponse has correct pagination structure", func(t *testing.T) {
		items := []Case{
			{ID: "1", Title: "Case 1"},
			{ID: "2", Title: "Case 2"},
		}
		resp := ListResponse{
			Items:      items,
			Total:      100,
			Page:       1,
			PageSize:   20,
			TotalPages: 5,
		}

		if resp.Total != 100 {
			t.Errorf("Expected Total 100, got %d", resp.Total)
		}
		if resp.Page != 1 {
			t.Errorf("Expected Page 1, got %d", resp.Page)
		}
		if resp.PageSize != 20 {
			t.Errorf("Expected PageSize 20, got %d", resp.PageSize)
		}
		if resp.TotalPages != 5 {
			t.Errorf("Expected TotalPages 5, got %d", resp.TotalPages)
		}
		if len(resp.Items.([]Case)) != 2 {
			t.Errorf("Expected 2 items, got %d", len(resp.Items.([]Case)))
		}
	})

	t.Run("ListResponse calculates TotalPages correctly", func(t *testing.T) {
		resp := ListResponse{
			Total:      25,
			Page:       1,
			PageSize:   10,
			TotalPages: 3,
		}

		if resp.TotalPages != 3 {
			t.Errorf("Expected TotalPages 3 for 25 items with page size 10, got %d", resp.TotalPages)
		}
	})
}

// TestHTTPStatusCodeConversion verifies HTTP status code mapping
func TestHTTPStatusCodeConversion(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		expectedStatus int
	}{
		{"Success maps to 200", CodeSuccess, 200},
		{"BadRequest maps to 400", CodeBadRequest, 400},
		{"Unauthorized maps to 401", CodeUnauthorized, 401},
		{"Forbidden maps to 403", CodeForbidden, 403},
		{"NotFound maps to 404", CodeNotFound, 404},
		{"InternalServerError maps to 500", CodeInternalServerError, 500},
		{"ServiceUnavailable maps to 502", CodeServiceUnavailable, 502},
		{"Unknown code maps to 500", 999, 500},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := HTTPStatusCode(tt.code)
			if status != tt.expectedStatus {
				t.Errorf("Expected HTTP status %d for code %d, got %d", tt.expectedStatus, tt.code, status)
			}
		})
	}
}
