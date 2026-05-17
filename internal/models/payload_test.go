package models

import (
	"strings"
	"testing"
)

// TestRenderTemplate tests the unified template rendering function
func TestRenderTemplate(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		variables map[string]string
		token     string
		domain    string
		wantErr   bool
	}{
		{
			name:      "Simple template",
			template:  "http://{token}.{domain}/",
			variables: map[string]string{},
			token:     "abc123",
			domain:    "example.com",
			wantErr:   false,
		},
		{
			name:      "Template with custom variable",
			template:  "http://{token}.{domain}/path?param={custom}",
			variables: map[string]string{"custom": "value"},
			token:     "abc123",
			domain:    "example.com",
			wantErr:   false,
		},
		{
			name:      "Template with callback_url variable",
			template:  "Callback: {callback_url}",
			variables: map[string]string{},
			token:     "abc123",
			domain:    "example.com",
			wantErr:   false,
		},
		{
			name:      "Template with unmatched variable",
			template:  "http://{token}.{domain}/path?param={unmatched}",
			variables: map[string]string{},
			token:     "abc123",
			domain:    "example.com",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTemplate(tt.template, tt.variables, tt.token, tt.domain)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				// Check that token and domain are substituted
				if tt.token != "" && !strings.Contains(result, tt.token) {
					t.Errorf("RenderTemplate() result should contain token %s, got %s", tt.token, result)
				}
				if tt.domain != "" && !strings.Contains(result, tt.domain) {
					t.Errorf("RenderTemplate() result should contain domain %s, got %s", tt.domain, result)
				}
			}
		})
	}
}

// TestRenderTemplateWithCase tests the template rendering with case variable
func TestRenderTemplateWithCase(t *testing.T) {
	tests := []struct {
		name      string
		template  string
		variables map[string]string
		token     string
		domain    string
		caseID    string
		wantErr   bool
	}{
		{
			name:      "Template with case variable",
			template:  "http://{token}.{domain}/case/{case}",
			variables: map[string]string{},
			token:     "abc123",
			domain:    "example.com",
			caseID:    "case-456",
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := RenderTemplateWithCase(tt.template, tt.variables, tt.token, tt.domain, tt.caseID)
			if (err != nil) != tt.wantErr {
				t.Errorf("RenderTemplateWithCase() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if !strings.Contains(result, tt.caseID) {
					t.Errorf("RenderTemplateWithCase() result should contain caseID %s, got %s", tt.caseID, result)
				}
			}
		})
	}
}

// TestPayloadTemplates tests that the payload templates are defined
func TestPayloadTemplates(t *testing.T) {
	if len(PayloadTemplates) == 0 {
		t.Fatal("PayloadTemplates should not be empty")
	}

	// Check that core templates exist
	coreTemplates := []string{"ssrf-basic", "xxe-basic", "rce-basic", "blind-sqli", "dns-rebinding"}
	for _, tmpl := range coreTemplates {
		if _, ok := PayloadTemplates[tmpl]; !ok {
			t.Errorf("PayloadTemplates should contain %s", tmpl)
		}
	}
}

// TestGenerateToken tests token generation
func TestGenerateToken(t *testing.T) {
	token1 := GenerateToken()
	token2 := GenerateToken()

	if token1 == "" {
		t.Fatal("GenerateToken() should not return empty string")
	}

	if token1 == token2 {
		t.Error("GenerateToken() should generate unique tokens")
	}

	if len(token1) < 8 {
		t.Errorf("GenerateToken() should generate tokens with at least 8 characters, got %d", len(token1))
	}
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	id1 := GenerateID()
	id2 := GenerateID()

	if id1 == "" {
		t.Fatal("GenerateID() should not return empty string")
	}

	if id1 == id2 {
		t.Error("GenerateID() should generate unique IDs")
	}

	if len(id1) < 16 {
		t.Errorf("GenerateID() should generate IDs with at least 16 characters, got %d", len(id1))
	}
}
