package docs

import (
	"os"
	"strings"
	"testing"
)

// TestOpenAPIFileExists verifies the OpenAPI YAML file exists
func TestOpenAPIFileExists(t *testing.T) {
	_, err := os.Stat("openapi.yaml")
	if os.IsNotExist(err) {
		t.Fatal("openapi.yaml file does not exist")
	}
	if err != nil {
		t.Fatalf("Error checking openapi.yaml: %v", err)
	}
}

// TestOpenAPIContainsRequiredSections verifies the OpenAPI file has required sections
func TestOpenAPIContainsRequiredSections(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("Error reading openapi.yaml: %v", err)
	}

	contentStr := string(content)

	requiredSections := []string{
		"openapi:",
		"info:",
		"servers:",
		"paths:",
		"components:",
		"securitySchemes:",
		"schemas:",
	}

	for _, section := range requiredSections {
		if !strings.Contains(contentStr, section) {
			t.Errorf("Missing required section: %s", section)
		}
	}
}

// TestOpenAPIContainsV2Paths verifies /api/v2 base path exists
func TestOpenAPIContainsV2Paths(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("Error reading openapi.yaml: %v", err)
	}

	contentStr := string(content)

	// Check for v2 base path in server URLs
	if !strings.Contains(contentStr, "/api/v2") {
		t.Error("OpenAPI file does not contain /api/v2 base path")
	}

	// Check for v2 API endpoints
	requiredEndpoints := []string{
		"/auth/login",
		"/cases",
		"/payloads",
		"/interactions",
		"/apikeys",
	}

	for _, endpoint := range requiredEndpoints {
		if !strings.Contains(contentStr, endpoint) {
			t.Errorf("Missing required endpoint: %s", endpoint)
		}
	}
}

// TestOpenAPIAuthenticationMethods verifies authentication methods are documented
func TestOpenAPIAuthenticationMethods(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("Error reading openapi.yaml: %v", err)
	}

	contentStr := string(content)

	// Check for JWT authentication
	if !strings.Contains(contentStr, "bearerAuth") {
		t.Error("OpenAPI file does not document bearerAuth (JWT) authentication")
	}

	// Check for API Key authentication
	if !strings.Contains(contentStr, "apiKeyAuth") {
		t.Error("OpenAPI file does not document apiKeyAuth authentication")
	}

	// Check for X-API-Key header
	if !strings.Contains(contentStr, "X-API-Key") {
		t.Error("OpenAPI file does not document X-API-Key header")
	}

	// Check for IsAgent flag description
	if !strings.Contains(contentStr, "is_agent") {
		t.Error("OpenAPI file does not document is_agent flag")
	}
}

// TestOpenAPIResponseStructure verifies unified response format is documented
func TestOpenAPIResponseStructure(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("Error reading openapi.yaml: %v", err)
	}

	contentStr := string(content)

	// Check for Error schema
	if !strings.Contains(contentStr, "Error:") {
		t.Error("OpenAPI file does not document Error schema")
	}

	// Check for SuccessResponse schema
	if !strings.Contains(contentStr, "SuccessResponse:") {
		t.Error("OpenAPI file does not document SuccessResponse schema")
	}

	// Check for code field
	if !strings.Contains(contentStr, "code:") {
		t.Error("OpenAPI file does not document code field in response")
	}

	// Check for message field
	if !strings.Contains(contentStr, "message:") {
		t.Error("OpenAPI file does not document message field in response")
	}

	// Check for data field
	if !strings.Contains(contentStr, "data:") {
		t.Error("OpenAPI file does not document data field in response")
	}
}

// TestOpenAPIErrorCodes verifies error codes are documented
func TestOpenAPIErrorCodes(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("Error reading openapi.yaml: %v", err)
	}

	contentStr := string(content)

	// Check for error codes
	errorCodes := []string{
		"400",
		"401",
		"403",
		"404",
		"500",
		"502",
	}

	for _, code := range errorCodes {
		if !strings.Contains(contentStr, code) {
			t.Errorf("OpenAPI file does not document error code %s", code)
		}
	}
}

// TestOpenAPIModelUpdates verifies model changes are reflected
func TestOpenAPIModelUpdates(t *testing.T) {
	content, err := os.ReadFile("openapi.yaml")
	if err != nil {
		t.Fatalf("Error reading openapi.yaml: %v", err)
	}

	contentStr := string(content)

	// Check for updated Payload fields
	if !strings.Contains(contentStr, "template_id") {
		t.Error("OpenAPI file does not document template_id field")
	}
	if !strings.Contains(contentStr, "template_rendered") {
		t.Error("OpenAPI file does not document template_rendered field")
	}

	// Check for updated APIKey fields
	if !strings.Contains(contentStr, "workspace_id") {
		t.Error("OpenAPI file does not document workspace_id field")
	}
	if !strings.Contains(contentStr, "risk_tolerance") {
		t.Error("OpenAPI file does not document risk_tolerance field")
	}
}
