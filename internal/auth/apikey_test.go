package auth

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
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

// setupAuthTestEngine creates an in-memory SQLite engine for auth tests
func setupAuthTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	if err := engine.Sync2(new(models.APIKey)); err != nil {
		t.Fatalf("failed to sync APIKey table: %v", err)
	}
	return engine
}

func TestAPIKeyBcryptMigration_NewKey(t *testing.T) {
	engine := setupAuthTestEngine(t)
	service := NewService(engine)

	req := &models.APIKeyCreateRequest{
		Name:   "test-bcrypt-key",
		Scopes: []string{"case:read"},
	}
	apiKey, err := service.CreateAPIKey(req, "test-user")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	if len(apiKey.KeyHash) == 0 {
		t.Fatal("KeyHash should be non-empty for new keys")
	}
	if apiKey.KeyHash == apiKey.Key {
		t.Fatal("KeyHash should not equal plaintext key")
	}

	// Validate the key works
	validated, err := service.ValidateAPIKey(apiKey.Key)
	if err != nil {
		t.Fatalf("ValidateAPIKey failed: %v", err)
	}
	if validated.ID != apiKey.ID {
		t.Fatal("validated key ID mismatch")
	}
}

func TestAPIKeyBcryptMigration_LegacyPlaintext(t *testing.T) {
	engine := setupAuthTestEngine(t)
	service := NewService(engine)

	// Insert a legacy key with empty KeyHash
	legacyKey := &models.APIKey{
		ID:        generateID(),
		Key:       "legacykey1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		KeyHash:   "",
		KeyPrefix: "legacyke",
		Name:      "legacy-key",
		Scopes:    models.Scopes{"case:read"},
		CreatedBy: "test-user",
	}
	_, err := engine.Insert(legacyKey)
	if err != nil {
		t.Fatalf("failed to insert legacy key: %v", err)
	}

	// Validate — should succeed and auto-migrate
	validated, err := service.ValidateAPIKey(legacyKey.Key)
	if err != nil {
		t.Fatalf("ValidateAPIKey for legacy key failed: %v", err)
	}
	if validated.ID != legacyKey.ID {
		t.Fatal("validated key ID mismatch")
	}

	// Verify KeyHash was backfilled
	var updated models.APIKey
	engine.ID(legacyKey.ID).Get(&updated)
	if len(updated.KeyHash) == 0 {
		t.Fatal("KeyHash should be backfilled after migration")
	}
	if updated.Key != "" {
		t.Fatal("plaintext Key should be cleared after migration")
	}

	// Validate again using bcrypt — should still work
	_, err = service.ValidateAPIKey(legacyKey.Key)
	if err != nil {
		t.Fatalf("ValidateAPIKey after migration failed: %v", err)
	}
}

func TestAPIKeyRotation(t *testing.T) {
	engine := setupAuthTestEngine(t)
	service := NewService(engine)

	req := &models.APIKeyCreateRequest{
		Name:   "test-rotate-key",
		Scopes: []string{"case:read"},
	}
	apiKey, err := service.CreateAPIKey(req, "test-user")
	if err != nil {
		t.Fatalf("CreateAPIKey failed: %v", err)
	}

	originalKey := apiKey.Key

	// Rotate
	newKey, err := service.RotateAPIKey(apiKey.ID)
	if err != nil {
		t.Fatalf("RotateAPIKey failed: %v", err)
	}
	if newKey == "" {
		t.Fatal("rotated key should not be empty")
	}
	if newKey == originalKey {
		t.Fatal("new key should differ from original")
	}

	// Old key should fail validation
	_, err = service.ValidateAPIKey(originalKey)
	if err == nil {
		t.Fatal("old key should fail validation after rotation")
	}

	// New key should pass validation
	validated, err := service.ValidateAPIKey(newKey)
	if err != nil {
		t.Fatalf("new key validation failed: %v", err)
	}
	if validated.ID != apiKey.ID {
		t.Fatal("rotated key ID should match original")
	}
}
