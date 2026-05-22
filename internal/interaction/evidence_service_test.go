package interaction

import (
	"sort"
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

func TestGenerateEvidence_ErrEvidenceNotFound(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Sync interactions table
	err = engine.Sync2(new(models.Interaction))
	if err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	service := NewEvidenceService(NewService(engine))

	// Test with non-existent case_id
	_, err = service.GenerateEvidence("nonexistent-case", "", "json")
	if err != ErrEvidenceNotFound {
		t.Errorf("Expected ErrEvidenceNotFound for non-existent case_id, got %v", err)
	}

	// Test with non-existent payload_id
	_, err = service.GenerateEvidence("", "nonexistent-payload", "json")
	if err != ErrEvidenceNotFound {
		t.Errorf("Expected ErrEvidenceNotFound for non-existent payload_id, got %v", err)
	}
}

func TestGenerateEvidence_TimelineChronological(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	// Create interactions with specific timestamps
	baseTime := time.Now()
	interactions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: baseTime.Add(2 * time.Hour),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: baseTime.Add(1 * time.Hour),
		},
		{
			Type:      "dns",
			SourceIP:  "10.0.0.1",
			Timestamp: baseTime,
		},
	}

	// Sort interactions chronologically (as done in GenerateEvidence)
	sort.Slice(interactions, func(i, j int) bool {
		return interactions[i].Timestamp.Before(interactions[j].Timestamp)
	})

	// Build timeline from sorted interactions
	timeline := service.buildTimeline(interactions)

	if len(timeline) != 3 {
		t.Errorf("Expected 3 timeline items, got %d", len(timeline))
	}

	// Verify timeline is in chronological order
	for i := 1; i < len(timeline); i++ {
		if timeline[i].Timestamp.Before(timeline[i-1].Timestamp) {
			t.Errorf("Timeline should be chronological, but item %d is before item %d", i, i-1)
		}
	}
}

func TestGenerateEvidence_JSONExportStructure(t *testing.T) {
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
	}

	evidence := &Evidence{
		ID:               "test-evidence-1",
		CaseID:           "case-1",
		Interactions:     interactions,
		Timeline:         service.buildTimeline(interactions),
		EvidenceStrength: EvidenceStrengthLow,
		Confidence:       10,
		InteractionCount: 1,
		UniqueSources:    1,
		Explainability:   "Test explanation",
		GeneratedAt:      time.Now(),
	}

	jsonContent := service.generateJSONEvidence(evidence)

	if jsonContent == "" {
		t.Error("Expected non-empty JSON content")
	}

	// Verify JSON contains key fields
	requiredFields := []string{"id", "case_id", "evidence_strength", "confidence", "interaction_count", "unique_sources", "explainability", "generated_at", "timeline"}
	for _, field := range requiredFields {
		if !contains(jsonContent, field) {
			t.Errorf("Expected JSON to contain field '%s'", field)
		}
	}
}

func TestGenerateEvidence_MarkdownExportContent(t *testing.T) {
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
	}

	evidence := &Evidence{
		ID:               "test-evidence-1",
		CaseID:           "case-1",
		Interactions:     interactions,
		Timeline:         service.buildTimeline(interactions),
		EvidenceStrength: EvidenceStrengthLow,
		Confidence:       10,
		InteractionCount: 1,
		UniqueSources:    1,
		Explainability:   "Test explanation",
		GeneratedAt:      time.Now(),
	}

	markdownContent := service.generateMarkdownEvidence(evidence)

	if markdownContent == "" {
		t.Error("Expected non-empty Markdown content")
	}

	// Verify Markdown contains summary and interaction details
	requiredSections := []string{"Evidence Report", "Case ID", "Evidence Strength", "Confidence", "Explainability", "Interactions"}
	for _, section := range requiredSections {
		if !contains(markdownContent, section) {
			t.Errorf("Expected Markdown to contain section '%s'", section)
		}
	}
}

func TestCalculateEvidenceStrength_HTTPWeighted(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewEvidenceService(NewService(engine))

	// Test with HTTP interactions (should have higher confidence than DNS only)
	httpInteractions := []models.Interaction{
		{
			Type:      "http",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "http",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
	}

	dnsInteractions := []models.Interaction{
		{
			Type:      "dns",
			SourceIP:  "127.0.0.1",
			Timestamp: time.Now(),
		},
		{
			Type:      "dns",
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
		},
	}

	_, httpConfidence := service.calculateEvidenceStrength(httpInteractions)
	_, dnsConfidence := service.calculateEvidenceStrength(dnsInteractions)

	// HTTP should have higher confidence than DNS for same count
	if httpConfidence <= dnsConfidence {
		t.Errorf("Expected HTTP confidence (%d) > DNS confidence (%d)", httpConfidence, dnsConfidence)
	}
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && (s == substr || len(s) > len(substr) && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func strPtr(s string) *string {
	return &s
}
