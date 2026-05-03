package canary

import (
	"testing"
	"time"
)

// TestCanaryTableName tests table name
func TestCanaryTableName(t *testing.T) {
	c := Canary{}
	tableName := c.TableName()
	if tableName != "canaries" {
		t.Fatalf("Expected table name 'canaries', got '%s'", tableName)
	}
}

// TestCanaryHitTableName tests table name
func TestCanaryHitTableName(t *testing.T) {
	h := CanaryHit{}
	tableName := h.TableName()
	if tableName != "canary_hits" {
		t.Fatalf("Expected table name 'canary_hits', got '%s'", tableName)
	}
}

// TestCanaryType constants
func TestCanaryType(t *testing.T) {
	types := []CanaryType{
		CanaryTypeDNS,
		CanaryTypeHTTP,
		CanaryTypeDocument,
		CanaryTypeConfig,
		CanaryTypeCI,
		CanaryTypeStorage,
		CanaryTypeEmail,
	}

	for _, ct := range types {
		if ct == "" {
			t.Fatal("Canary type should not be empty")
		}
	}
}

// TestCanaryModel tests canary model
func TestCanaryModel(t *testing.T) {
	now := time.Now()
	c := Canary{
		ID:          "test-canary-1",
		Type:        string(CanaryTypeDNS),
		Token:       "test-token-abc123",
		Context:     "encoded-context",
		ExpiresAt:   now,
		IsEnabled:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if c.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if c.Token == "" {
		t.Fatal("Token should not be empty")
	}

	if c.Type != string(CanaryTypeDNS) {
		t.Fatalf("Expected type '%s', got '%s'", CanaryTypeDNS, c.Type)
	}
}

// TestCanaryHitModel tests canary hit model
func TestCanaryHitModel(t *testing.T) {
	now := time.Now()
	h := CanaryHit{
		ID:          "test-hit-1",
		CanaryID:    "test-canary-1",
		SourceIP:    "192.168.1.1",
		UserAgent:   "test-agent",
		Timestamp:   now,
		IsCompressed: false,
	}

	if h.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if h.SourceIP == "" {
		t.Fatal("SourceIP should not be empty")
	}
}

// TestCanaryConfig tests canary config
func TestCanaryConfig(t *testing.T) {
	config := CanaryConfig{
		MaxRetentionDays: 90,
		DefaultExpiry:    "90d",
		SilentWindow:     300,
		CompressionThreshold: 10,
		NotificationLevels: []string{"low", "medium", "high"},
	}

	if config.MaxRetentionDays != 90 {
		t.Fatalf("Expected MaxRetentionDays 90, got %d", config.MaxRetentionDays)
	}

	if len(config.NotificationLevels) != 3 {
		t.Fatalf("Expected 3 notification levels, got %d", len(config.NotificationLevels))
	}
}
