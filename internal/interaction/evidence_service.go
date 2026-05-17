package interaction

import (
	"errors"
	"fmt"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrEvidenceNotFound = errors.New("evidence not found")
)

// EvidenceService provides evidence chain generation services
type EvidenceService struct {
	interactionService *Service
	payloadService     interface{} // Will be payload.Service
	caseService        interface{} // Will be case.Service
}

// NewEvidenceService creates a new evidence service
func NewEvidenceService(interactionService *Service) *EvidenceService {
	return &EvidenceService{
		interactionService: interactionService,
	}
}

// GenerateEvidence generates an evidence chain for a case or payload
func (s *EvidenceService) GenerateEvidence(caseID, payloadID string, format string) (*EvidenceResponse, error) {
	// Get interactions
	var interactions []models.Interaction

	if payloadID != "" {
		resp, err := s.interactionService.ListInteractions("", payloadID, "", nil, nil, 1, 1000)
		if err != nil {
			return nil, err
		}
		interactions = resp.Items
	} else if caseID != "" {
		resp, err := s.interactionService.ListInteractions(caseID, "", "", nil, nil, 1, 1000)
		if err != nil {
			return nil, err
		}
		interactions = resp.Items
	} else {
		return nil, errors.New("either case_id or payload_id must be specified")
	}

	if len(interactions) == 0 {
		return nil, ErrEvidenceNotFound
	}

	// Build evidence with MVP scoring logic
	strength, confidence := s.calculateEvidenceStrength(interactions)
	evidence := &Evidence{
		ID:               models.GenerateID(),
		CaseID:           caseID,
		PayloadID:        payloadID,
		Interactions:     interactions,
		Timeline:         s.buildTimeline(interactions),
		EvidenceStrength: strength,
		Confidence:       confidence,
		InteractionCount: len(interactions),
		UniqueSources:    s.countUniqueSources(interactions),
		Explainability:   s.generateExplainability(interactions, strength, confidence),
		GeneratedAt:      time.Now(),
	}

	// Generate content based on format
	var content string
	switch format {
	case "json":
		content = s.generateJSONEvidence(evidence)
	case "markdown":
		content = s.generateMarkdownEvidence(evidence)
	case "csv":
		content = s.generateCSVEvidence(evidence)
	default:
		return nil, errors.New("unsupported format")
	}

	return &EvidenceResponse{
		Format:  format,
		Content: content,
		Metadata: map[string]interface{}{
			"interaction_count": len(interactions),
			"case_id":           caseID,
			"payload_id":        payloadID,
		},
	}, nil
}

// calculateEvidenceStrength calculates evidence strength and confidence based on MVP scoring logic
// According to docs/mvp-closed-loop.md:
// - Low: 0-2 interactions
// - Medium: 3-5 interactions from ≥2 unique sources
// - High: 6+ interactions from ≥3 unique sources
// - Confidence: Based on interaction count and source diversity (0-100)
func (s *EvidenceService) calculateEvidenceStrength(interactions []models.Interaction) (EvidenceStrength, int) {
	if len(interactions) == 0 {
		return EvidenceStrengthLow, 0
	}

	uniqueSources := s.countUniqueSources(interactions)
	count := len(interactions)

	var strength EvidenceStrength
	var confidence int

	// Determine evidence strength
	if count <= 2 {
		strength = EvidenceStrengthLow
	} else if count >= 3 && count <= 5 && uniqueSources >= 2 {
		strength = EvidenceStrengthMedium
	} else if count >= 6 && uniqueSources >= 3 {
		strength = EvidenceStrengthHigh
	} else {
		// Default to low if conditions not met
		strength = EvidenceStrengthLow
	}

	// Calculate confidence (0-100) based on interaction count and source diversity
	// Base confidence from interaction count (max 60 points)
	countScore := float64(count) * 10.0
	if countScore > 60.0 {
		countScore = 60.0
	}

	// Source diversity bonus (max 40 points)
	sourceScore := float64(uniqueSources-1) * 20.0
	if sourceScore > 40.0 {
		sourceScore = 40.0
	}

	confidence = int(countScore + sourceScore)
	if confidence > 100 {
		confidence = 100
	}
	if confidence < 0 {
		confidence = 0
	}

	return strength, confidence
}

// countUniqueSources counts unique source IPs from interactions
func (s *EvidenceService) countUniqueSources(interactions []models.Interaction) int {
	sources := make(map[string]bool)
	for _, i := range interactions {
		sources[i.SourceIP] = true
	}
	return len(sources)
}

// generateExplainability generates human-readable explanation of findings
// Format: "Captured X interactions from Y unique sources. Evidence strength: Z."
func (s *EvidenceService) generateExplainability(interactions []models.Interaction, strength EvidenceStrength, confidence int) string {
	count := len(interactions)
	uniqueSources := s.countUniqueSources(interactions)

	return fmt.Sprintf("Captured %d interaction%s from %d unique source%s. Evidence strength: %s. Confidence: %d%%.",
		count,
		pluralize(count),
		uniqueSources,
		pluralize(uniqueSources),
		strength,
		confidence,
	)
}

// pluralize returns "s" if count != 1, otherwise empty string
func pluralize(count int) string {
	if count != 1 {
		return "s"
	}
	return ""
}

// calculateScore calculates an evidence score based on interactions (deprecated, replaced by calculateEvidenceStrength)
func (s *EvidenceService) calculateScore(interactions []models.Interaction) float64 {
	if len(interactions) == 0 {
		return 0.0
	}

	score := 0.0
	for _, interaction := range interactions {
		// Base score for any interaction
		score += 10.0

		// Bonus for HTTP interactions (more valuable than DNS)
		if interaction.Type == "http" {
			score += 20.0
		}

		// Bonus for interactions with body content
		if interaction.Body != nil && len(*interaction.Body) > 0 {
			score += 20.0
		}
	}

	// Normalize score to 0-100 range
	maxScore := float64(len(interactions)) * 30.0
	if maxScore > 0 {
		score = (score / maxScore) * 100.0
	}

	if score > 100.0 {
		score = 100.0
	}

	return score
}

// buildTimeline builds a timeline from interactions
func (s *EvidenceService) buildTimeline(interactions []models.Interaction) []TimelineItem {
	timeline := make([]TimelineItem, 0, len(interactions))

	for _, i := range interactions {
		item := TimelineItem{
			Type:        "interaction",
			Description: i.Type + " interaction from " + i.SourceIP,
			Timestamp:   i.Timestamp,
			Metadata: map[string]interface{}{
				"interaction_id": i.ID,
				"type":           i.Type,
				"source_ip":      i.SourceIP,
			},
		}

		if i.Domain != nil {
			item.Metadata["domain"] = *i.Domain
		}
		if i.Path != nil {
			item.Metadata["path"] = *i.Path
		}
		if i.Method != nil {
			item.Metadata["method"] = *i.Method
		}

		timeline = append(timeline, item)
	}

	return timeline
}

// generateJSONEvidence generates JSON format evidence
func (s *EvidenceService) generateJSONEvidence(evidence *Evidence) string {
	// Simplified JSON generation
	return `{
  "id": "` + evidence.ID + `",
  "case_id": "` + evidence.CaseID + `",
  "payload_id": "` + evidence.PayloadID + `",
  "evidence_strength": "` + string(evidence.EvidenceStrength) + `",
  "confidence": ` + fmt.Sprintf("%d", evidence.Confidence) + `,
  "interaction_count": ` + fmt.Sprintf("%d", evidence.InteractionCount) + `,
  "unique_sources": ` + fmt.Sprintf("%d", evidence.UniqueSources) + `,
  "explainability": "` + evidence.Explainability + `",
  "generated_at": "` + evidence.GeneratedAt.Format(time.RFC3339) + `"
}`
}

// generateMarkdownEvidence generates Markdown format evidence
func (s *EvidenceService) generateMarkdownEvidence(evidence *Evidence) string {
	md := "# Evidence Report\n\n"
	md += "**Case ID**: " + evidence.CaseID + "\n"
	md += "**Payload ID**: " + evidence.PayloadID + "\n"
	md += "**Evidence Strength**: " + string(evidence.EvidenceStrength) + "\n"
	md += "**Confidence**: " + fmt.Sprintf("%d%%", evidence.Confidence) + "\n"
	md += "**Interaction Count**: " + fmt.Sprintf("%d", evidence.InteractionCount) + "\n"
	md += "**Unique Sources**: " + fmt.Sprintf("%d", evidence.UniqueSources) + "\n"
	md += "**Generated**: " + evidence.GeneratedAt.Format(time.RFC3339) + "\n\n"
	md += "## Explainability\n\n"
	md += evidence.Explainability + "\n\n"
	md += "## Interactions\n\n"

	for _, i := range evidence.Interactions {
		md += "### " + i.Type + " - " + i.Timestamp.Format(time.RFC3339) + "\n"
		md += "- Source IP: " + i.SourceIP + "\n"
		if i.Domain != nil {
			md += "- Domain: " + *i.Domain + "\n"
		}
		if i.Path != nil {
			md += "- Path: " + *i.Path + "\n"
		}
		md += "\n"
	}

	return md
}

// generateCSVEvidence generates CSV format evidence
func (s *EvidenceService) generateCSVEvidence(evidence *Evidence) string {
	csv := "ID,Type,Timestamp,SourceIP\n"
	for _, i := range evidence.Interactions {
		csv += i.ID + "," + i.Type + "," + i.Timestamp.Format(time.RFC3339) + "," + i.SourceIP + "\n"
	}
	return csv
}
