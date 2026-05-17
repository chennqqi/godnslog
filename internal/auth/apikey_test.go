package auth

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// TestGenerateAPIKey tests API key generation
func TestGenerateAPIKey(t *testing.T) {
	key, prefix, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generateAPIKey failed: %v", err)
	}
	if key == "" {
		t.Error("key should not be empty")
	}
	if prefix == "" {
		t.Error("prefix should not be empty")
	}
	if len(prefix) != 8 {
		t.Errorf("prefix should be 8 characters, got %d", len(prefix))
	}
	if key[:8] != prefix {
		t.Error("prefix should match first 8 characters of key")
	}
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	id := generateID()
	if id == "" {
		t.Error("id should not be empty")
	}
	// Base32 encoding of 16 bytes produces 26 characters
	if len(id) != 32 {
		t.Errorf("id should be 32 characters, got %d", len(id))
	}
}

// TestValidScopes tests scope validation
func TestValidScopes(t *testing.T) {
	tests := []struct {
		name  string
		scope string
		valid bool
	}{
		{"case read", "case:read", true},
		{"case write", "case:write", true},
		{"payload read", "payload:read", true},
		{"admin all", "admin:all", true},
		{"invalid scope", "invalid:scope", false},
		{"empty scope", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := ValidScopes[tt.scope]
			if valid != tt.valid {
				t.Errorf("scope %s validity should be %v, got %v", tt.scope, tt.valid, valid)
			}
		})
	}
}

// TestAPIKeyIsValid tests API key validation
func TestAPIKeyIsValid(t *testing.T) {
	timePtr := func(t time.Time) *time.Time {
		return &t
	}

	tests := []struct {
		name   string
		apiKey models.APIKey
		valid  bool
	}{
		{"valid key", models.APIKey{IsRevoked: false, ExpiresAt: nil}, true},
		{"revoked key", models.APIKey{IsRevoked: true, ExpiresAt: nil}, false},
		{"expired key", models.APIKey{IsRevoked: false, ExpiresAt: timePtr(time.Now().Add(-1 * time.Hour))}, false},
		{"future expiry", models.APIKey{IsRevoked: false, ExpiresAt: timePtr(time.Now().Add(1 * time.Hour))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiKey.IsValid()
			if result != tt.valid {
				t.Errorf("IsValid() should return %v, got %v", tt.valid, result)
			}
		})
	}
}

// timePtr is a helper to get a pointer to time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
