package auth

import (
	"testing"
	"time"
)

// TestGenerateAPIKey tests API key generation
func TestGenerateAPIKey(t *testing.T) {
	key, prefix, err := generateAPIKey()
	if err != nil {
		t.Fatalf("generateAPIKey failed: %v", err)
	}

	if key == "" {
		t.Fatal("Generated key is empty")
	}

	if prefix == "" {
		t.Fatal("Generated prefix is empty")
	}

	if len(key) < 64 {
		t.Fatalf("Generated key is too short: %d", len(key))
	}

	if len(prefix) != 8 {
		t.Fatalf("Expected prefix length 8, got %d", len(prefix))
	}

	if key[:8] != prefix {
		t.Fatal("Prefix should be first 8 characters of key")
	}
}

// TestGenerateID tests ID generation
func TestGenerateID(t *testing.T) {
	id := generateID()
	if id == "" {
		t.Fatal("Generated ID is empty")
	}

	if len(id) < 10 {
		t.Fatalf("Generated ID is too short: %d", len(id))
	}
}

// TestValidScopes tests valid scope validation
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
				t.Errorf("Expected valid=%v for scope '%s', got %v", tt.valid, tt.scope, valid)
			}
		})
	}
}

// TestAPIKeyIsValid tests API key validation
func TestAPIKeyIsValid(t *testing.T) {
	tests := []struct {
		name     string
		apiKey   APIKey
		expected bool
	}{
		{"valid key", APIKey{IsRevoked: false, ExpiresAt: nil}, true},
		{"revoked key", APIKey{IsRevoked: true, ExpiresAt: nil}, false},
		{"expired key", APIKey{IsRevoked: false, ExpiresAt: timePtr(time.Now().Add(-1 * time.Hour))}, false},
		{"future expiry", APIKey{IsRevoked: false, ExpiresAt: timePtr(time.Now().Add(1 * time.Hour))}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.apiKey.IsValid()
			if result != tt.expected {
				t.Errorf("Expected IsValid()=%v, got %v", tt.expected, result)
			}
		})
	}
}

// timePtr is a helper to get a pointer to time.Time
func timePtr(t time.Time) *time.Time {
	return &t
}
