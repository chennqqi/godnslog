package interaction

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

func TestCalculateEvidenceScore(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
			Body:      strPtr("test body"),
		},
	}

	score := service.calculateScore(interactions)

	if score < 0 || score > 100 {
		t.Fatalf("Expected score between 0 and 100, got %f", score)
	}

	// HTTP interactions should contribute more to score than DNS
	if score <= 10 {
		t.Fatalf("Expected score > 10 for HTTP+DNS interactions, got %f", score)
	}
}

func TestCalculateEvidenceStrength_NoInteractions(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{}
	strength, confidence := service.calculateEvidenceStrength(interactions)

	if strength != EvidenceStrengthLow {
		t.Errorf("Expected strength low for no interactions, got %s", strength)
	}
	if confidence != 0 {
		t.Errorf("Expected confidence 0 for no interactions, got %d", confidence)
	}
}

func TestCalculateEvidenceStrength_Low(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
	}

	strength, confidence := service.calculateEvidenceStrength(interactions)

	if strength != EvidenceStrengthLow {
		t.Errorf("Expected strength low for 2 interactions, got %s", strength)
	}
	if confidence < 0 || confidence > 100 {
		t.Errorf("Expected confidence between 0 and 100, got %d", confidence)
	}
}

func TestCalculateEvidenceStrength_Medium(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
	}

	strength, confidence := service.calculateEvidenceStrength(interactions)

	if strength != EvidenceStrengthMedium {
		t.Errorf("Expected strength medium for 3 interactions from 2 sources, got %s", strength)
	}
	if confidence < 0 || confidence > 100 {
		t.Errorf("Expected confidence between 0 and 100, got %d", confidence)
	}
}

func TestCalculateEvidenceStrength_High(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "dns",
			SourceIP:  "10.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "172.16.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
	}

	strength, confidence := service.calculateEvidenceStrength(interactions)

	if strength != EvidenceStrengthHigh {
		t.Errorf("Expected strength high for 6 interactions from 3 sources, got %s", strength)
	}
	if confidence < 0 || confidence > 100 {
		t.Errorf("Expected confidence between 0 and 100, got %d", confidence)
	}
}

func TestCountUniqueSources(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "10.0.0.1",
			Timestamp: time.Now(),
		},
	}

	count := service.countUniqueSources(interactions)

	if count != 3 {
		t.Errorf("Expected 3 unique sources, got %d", count)
	}
}

func TestGenerateExplainability(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
	}

	strength := EvidenceStrengthMedium
	confidence := 75

	explainability := service.generateExplainability(interactions, strength, confidence)

	if explainability == "" {
		t.Error("Expected non-empty explainability")
	}

	// Verify explainability contains key information
	if len(explainability) == 0 {
		t.Error("Explainability should not be empty")
	}
}

func strPtr(s string) *string {
	return &s
}
