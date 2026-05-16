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

func strPtr(s string) *string {
	return &s
}
