package interaction

import (
	"errors"
	"time"
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
	var interactions []Interaction

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

	// Build evidence
	evidence := &Evidence{
		ID:           generateID(),
		CaseID:       caseID,
		PayloadID:    payloadID,
		Interactions: interactions,
		Timeline:     s.buildTimeline(interactions),
		CreatedAt:    time.Now(),
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

// buildTimeline builds a timeline from interactions
func (s *EvidenceService) buildTimeline(interactions []Interaction) []TimelineItem {
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
  "interaction_count": ` + string(rune(len(evidence.Interactions))) + `,
  "created_at": "` + evidence.CreatedAt.Format(time.RFC3339) + `"
}`
}

// generateMarkdownEvidence generates Markdown format evidence
func (s *EvidenceService) generateMarkdownEvidence(evidence *Evidence) string {
	md := "# Evidence Report\n\n"
	md += "**Case ID**: " + evidence.CaseID + "\n"
	md += "**Payload ID**: " + evidence.PayloadID + "\n"
	md += "**Generated**: " + evidence.CreatedAt.Format(time.RFC3339) + "\n\n"
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
