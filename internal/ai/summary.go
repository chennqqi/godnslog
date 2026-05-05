package ai

import (
	"fmt"
	"strings"
	"time"
)

// SummaryService handles AI-powered evidence summarization
type SummaryService struct {
	// In production, this would have an AI client (OpenAI, Claude, etc.)
}

// NewSummaryService creates a new summary service
func NewSummaryService() *SummaryService {
	return &SummaryService{}
}

// EvidenceSummary represents a summary of evidence for a case
type EvidenceSummary struct {
	CaseID          string    `json:"case_id"`
	Title           string    `json:"title"`
	Summary         string    `json:"summary"`
	RiskLevel       string    `json:"risk_level"` // low, medium, high, critical
	Findings        []Finding `json:"findings"`
	Recommendations []string  `json:"recommendations"`
	GeneratedAt     time.Time `json:"generated_at"`
}

// Finding represents a security finding
type Finding struct {
	Type        string `json:"type"`        // ssrf, xxe, rce, etc.
	Severity    string `json:"severity"`    // low, medium, high, critical
	Description string `json:"description"`
	Evidence    string `json:"evidence"`
}

// GenerateSummary generates an AI-powered summary of evidence
func (s *SummaryService) GenerateSummary(caseID, title string, interactions []interface{}) (*EvidenceSummary, error) {
	// In production, this would call an AI API to analyze interactions
	// For now, implement rule-based analysis
	
	summary := &EvidenceSummary{
		CaseID:      caseID,
		Title:       title,
		Summary:     s.generateBasicSummary(interactions),
		RiskLevel:   s.assessRiskLevel(interactions),
		Findings:    s.extractFindings(interactions),
		GeneratedAt: time.Now(),
	}
	
	summary.Recommendations = s.generateRecommendations(summary.RiskLevel, summary.Findings)
	
	return summary, nil
}

// generateBasicSummary generates a basic summary without AI
func (s *SummaryService) generateBasicSummary(interactions []interface{}) string {
	count := len(interactions)
	if count == 0 {
		return "No interactions detected in this case."
	}
	
	return fmt.Sprintf("Detected %d interaction(s) in this case. Analysis shows potential security testing activity.", count)
}

// assessRiskLevel assesses the overall risk level based on interactions
func (s *SummaryService) assessRiskLevel(interactions []interface{}) string {
	// In production, this would analyze interaction patterns
	// For now, use simple heuristics
	
	if len(interactions) == 0 {
		return "low"
	}
	
	// Check for high-risk indicators
	for _, interaction := range interactions {
		if m, ok := interaction.(map[string]interface{}); ok {
			if raw, ok := m["raw_data"].(string); ok {
				if strings.Contains(strings.ToLower(raw), "metadata") ||
					strings.Contains(strings.ToLower(raw), "169.254") ||
					strings.Contains(strings.ToLower(raw), "admin") ||
					strings.Contains(strings.ToLower(raw), "password") {
					return "critical"
				}
			}
		}
	}
	
	if len(interactions) > 10 {
		return "high"
	}
	
	return "medium"
}

// extractFindings extracts security findings from interactions
func (s *SummaryService) extractFindings(interactions []interface{}) []Finding {
	findings := []Finding{}
	
	for _, interaction := range interactions {
		if m, ok := interaction.(map[string]interface{}); ok {
			// Analyze interaction type
			if type_, ok := m["type"].(string); ok {
				finding := Finding{
					Type:        type_,
					Severity:    "medium",
					Description: fmt.Sprintf("Detected %s interaction", type_),
					Evidence:    fmt.Sprintf("%v", m),
				}
				
				// Adjust severity based on patterns
				if raw, ok := m["raw_data"].(string); ok {
					lowerRaw := strings.ToLower(raw)
					if strings.Contains(lowerRaw, "metadata") ||
						strings.Contains(lowerRaw, "169.254") {
						finding.Severity = "critical"
						finding.Description = "Potential cloud metadata access detected"
					}
				}
				
				findings = append(findings, finding)
			}
		}
	}
	
	return findings
}

// generateRecommendations generates security recommendations
func (s *SummaryService) generateRecommendations(riskLevel string, findings []Finding) []string {
	recommendations := []string{}
	
	switch riskLevel {
	case "critical":
		recommendations = append(recommendations,
			"Immediate security review required",
			"Verify if cloud metadata access was successful",
			"Review access controls and permissions",
			"Consider rotating credentials if exposure confirmed")
	case "high":
		recommendations = append(recommendations,
			"Security review recommended",
			"Investigate the source of interactions",
			"Review firewall and network rules")
	case "medium":
		recommendations = append(recommendations,
			"Monitor for additional interactions",
			"Review the security context of the test")
	case "low":
		recommendations = append(recommendations,
			"No immediate action required",
			"Continue monitoring")
	}
	
	// Add specific recommendations based on findings
	for _, finding := range findings {
		if finding.Type == "ssrf" {
			recommendations = append(recommendations,
				"Review SSRF protections in affected endpoints")
		}
		if finding.Type == "xxe" {
			recommendations = append(recommendations,
				"Review XML parser configurations")
		}
	}
	
	return recommendations
}

// ExplainEvidence provides a detailed explanation of specific evidence
func (s *SummaryService) ExplainEvidence(evidenceID string) (string, error) {
	// In production, this would retrieve the evidence and use AI to explain it
	// For now, return a placeholder
	return fmt.Sprintf("Evidence %s explanation: This interaction indicates external connectivity from the target system.", evidenceID), nil
}
