package ai

import (
	"testing"
)

// TestGenerateSummary tests basic summary generation
func TestGenerateSummary(t *testing.T) {
	service := NewSummaryService()

	interactions := []interface{}{
		map[string]interface{}{
			"type":     "dns",
			"raw_data": "query for token.example.com",
		},
		map[string]interface{}{
			"type":     "http",
			"raw_data": "GET /admin HTTP/1.1",
		},
	}

	summary, err := service.GenerateSummary("case-1", "Test Case", interactions)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if summary == nil {
		t.Fatal("Expected summary, got nil")
	}
	if summary.CaseID != "case-1" {
		t.Errorf("Expected case_id 'case-1', got '%s'", summary.CaseID)
	}
	if summary.Title != "Test Case" {
		t.Errorf("Expected title 'Test Case', got '%s'", summary.Title)
	}
	if len(summary.Findings) != 2 {
		t.Errorf("Expected 2 findings, got %d", len(summary.Findings))
	}
}

// TestGenerateSummaryEmpty tests summary with no interactions
func TestGenerateSummaryEmpty(t *testing.T) {
	service := NewSummaryService()

	summary, err := service.GenerateSummary("case-2", "Empty Case", []interface{}{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if summary.RiskLevel != "low" {
		t.Errorf("Expected risk level 'low' for empty, got '%s'", summary.RiskLevel)
	}
}

// TestAssessRiskLevelCritical tests critical risk detection
func TestAssessRiskLevelCritical(t *testing.T) {
	service := NewSummaryService()

	interactions := []interface{}{
		map[string]interface{}{
			"raw_data": "accessing 169.254.169.254 metadata endpoint",
		},
	}

	risk := service.assessRiskLevel(interactions)
	if risk != "critical" {
		t.Errorf("Expected 'critical', got '%s'", risk)
	}
}

// TestAssessRiskLevelHigh tests high risk detection
func TestAssessRiskLevelHigh(t *testing.T) {
	service := NewSummaryService()

	interactions := make([]interface{}, 15)
	for i := range interactions {
		interactions[i] = map[string]interface{}{
			"type": "dns",
		}
	}

	risk := service.assessRiskLevel(interactions)
	if risk != "high" {
		t.Errorf("Expected 'high', got '%s'", risk)
	}
}

// TestExplainEvidence tests structured evidence explanation
func TestExplainEvidence(t *testing.T) {
	service := NewSummaryService()

	req := &ExplainEvidenceRequest{
		CaseID:           "case-1",
		EvidenceID:       "evidence-1",
		EvidenceStrength: "high",
		Confidence:       85,
		Interactions: []map[string]interface{}{
			{
				"type":     "http",
				"raw_data": "GET /admin HTTP/1.1",
			},
			{
				"type":     "dns",
				"raw_data": "query for token.example.com",
			},
		},
	}

	resp, err := service.ExplainEvidence(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if resp.CaseID != "case-1" {
		t.Errorf("Expected case_id 'case-1', got '%s'", resp.CaseID)
	}
	if resp.EvidenceID != "evidence-1" {
		t.Errorf("Expected evidence_id 'evidence-1', got '%s'", resp.EvidenceID)
	}
	if resp.Explanation == "" {
		t.Error("Expected non-empty explanation")
	}
	if len(resp.Findings) != 2 {
		t.Errorf("Expected 2 findings, got %d", len(resp.Findings))
	}
	if len(resp.Remediation) == 0 {
		t.Error("Expected non-empty remediation")
	}
	if resp.Metadata["analysis_method"] != "rule_based" {
		t.Errorf("Expected analysis_method 'rule_based', got '%s'", resp.Metadata["analysis_method"])
	}
}

// TestExplainEvidenceNilRequest tests nil request handling
func TestExplainEvidenceNilRequest(t *testing.T) {
	service := NewSummaryService()

	_, err := service.ExplainEvidence(nil)
	if err == nil {
		t.Fatal("Expected error for nil request")
	}
}

// TestExplainEvidenceCritical tests critical finding detection in explanation
func TestExplainEvidenceCritical(t *testing.T) {
	service := NewSummaryService()

	req := &ExplainEvidenceRequest{
		CaseID:     "case-critical",
		EvidenceID: "evidence-critical",
		Interactions: []map[string]interface{}{
			{
				"type":     "http",
				"raw_data": "accessing 169.254.169.254 metadata",
			},
		},
	}

	resp, err := service.ExplainEvidence(req)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if resp.RiskLevel != "critical" {
		t.Errorf("Expected risk 'critical', got '%s'", resp.RiskLevel)
	}

	// Check that critical findings are mentioned in explanation
	foundCritical := false
	for _, f := range resp.Findings {
		if f.Severity == "critical" {
			foundCritical = true
			break
		}
	}
	if !foundCritical {
		t.Error("Expected at least one critical finding")
	}
}

// TestGenerateRecommendations tests recommendation generation
func TestGenerateRecommendations(t *testing.T) {
	service := NewSummaryService()

	// Critical risk
	recs := service.generateRecommendations("critical", []Finding{})
	if len(recs) < 3 {
		t.Errorf("Expected at least 3 recommendations for critical, got %d", len(recs))
	}

	// Low risk
	recs = service.generateRecommendations("low", []Finding{})
	if len(recs) < 1 {
		t.Errorf("Expected at least 1 recommendation for low, got %d", len(recs))
	}

	// With SSRF finding
	recs = service.generateRecommendations("high", []Finding{{Type: "ssrf"}})
	foundSSRF := false
	for _, r := range recs {
		if r == "Review SSRF protections in affected endpoints" {
			foundSSRF = true
			break
		}
	}
	if !foundSSRF {
		t.Error("Expected SSRF-specific recommendation")
	}
}
