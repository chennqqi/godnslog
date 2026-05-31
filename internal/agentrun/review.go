package agentrun

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// ErrAgentRunNotFound is returned when an agent run is not found
var ErrAgentRunNotFound = errors.New("agent run not found")

// ReviewService provides agent run review packet generation services
type ReviewService struct {
	engine             *xorm.Engine
	agentRunService    *Service
	authService        *auth.Service
	evidenceService    *interaction.EvidenceService
	interactionService *interaction.Service
}

// NewReviewService creates a new review service
func NewReviewService(engine *xorm.Engine, agentRunService *Service, authService *auth.Service, evidenceService *interaction.EvidenceService, interactionService *interaction.Service) *ReviewService {
	return &ReviewService{
		engine:             engine,
		agentRunService:    agentRunService,
		authService:        authService,
		evidenceService:    evidenceService,
		interactionService: interactionService,
	}
}

// AgentRunReviewPacket represents a review packet for a single agent run
type AgentRunReviewPacket struct {
	ID                 string                     `json:"id"`
	AgentRun           models.AgentRunDetail      `json:"agent_run"`
	CaseID             string                     `json:"case_id,omitempty"`
	PayloadID          string                     `json:"payload_id,omitempty"`
	Target             string                     `json:"target,omitempty"`
	InteractionSummary AgentRunInteractionSummary `json:"interaction_summary"`
	Evidence           *interaction.Evidence      `json:"evidence,omitempty"`
	AuditRefs          []AgentRunAuditRef         `json:"audit_refs"`
	GeneratedAt        time.Time                  `json:"generated_at"`
	Format             string                     `json:"format"`
	Content            string                     `json:"content,omitempty"`
}

// AgentRunInteractionSummary represents a summary of interactions for an agent run
type AgentRunInteractionSummary struct {
	Total             int        `json:"total"`
	DNSCount          int        `json:"dns_count"`
	HTTPCount         int        `json:"http_count"`
	UniqueSources     int        `json:"unique_sources"`
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"`
}

// AgentRunAuditRef represents an audit reference for an agent run
type AgentRunAuditRef struct {
	ID           string    `json:"id"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   string    `json:"resource_id,omitempty"`
	Timestamp    time.Time `json:"timestamp"`
}

// BuildReviewPacket generates a review packet for an agent run
func (s *ReviewService) BuildReviewPacket(agentRunID, format, baseURL string) (*AgentRunReviewPacket, error) {
	// Get agent run detail
	agentRunDetail, err := s.agentRunService.GetAgentRunDetail(agentRunID, baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent run detail: %w", err)
	}
	if agentRunDetail == nil {
		return nil, ErrAgentRunNotFound
	}

	// Build interaction summary from real interactions
	interactionSummary, err := s.buildInteractionSummary(agentRunDetail.CaseID, agentRunDetail.PayloadID)
	if err != nil {
		return nil, fmt.Errorf("failed to build interaction summary: %w", err)
	}

	// Generate evidence using unified Evidence service
	var evidence *interaction.Evidence
	if agentRunDetail.PayloadID != "" {
		evidenceResp, err := s.evidenceService.GenerateEvidence("", agentRunDetail.PayloadID, "json")
		if err == nil {
			evidence = evidenceResp.Evidence
		}
		// If no interactions, evidence can be nil - not an error
	} else if agentRunDetail.CaseID != "" {
		evidenceResp, err := s.evidenceService.GenerateEvidence(agentRunDetail.CaseID, "", "json")
		if err == nil {
			evidence = evidenceResp.Evidence
		}
		// If no interactions, evidence can be nil - not an error
	}

	// Get audit references
	auditRefs, err := s.getAuditRefs(agentRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit refs: %w", err)
	}

	// Build review packet
	packet := &AgentRunReviewPacket{
		ID:                 fmt.Sprintf("review-%s", agentRunID),
		AgentRun:           *agentRunDetail,
		CaseID:             agentRunDetail.CaseID,
		PayloadID:          agentRunDetail.PayloadID,
		Target:             agentRunDetail.Target,
		InteractionSummary: *interactionSummary,
		Evidence:           evidence,
		AuditRefs:          auditRefs,
		GeneratedAt:        time.Now(),
		Format:             format,
	}

	// Generate content based on format
	switch format {
	case "json":
		packet.Content = s.generateJSONContent(packet)
	case "markdown":
		packet.Content = s.generateMarkdownContent(packet)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}

	// Write audit log for review generation
	userID := agentRunDetail.OperatorID
	userIDPtr := &userID
	resourceIDPtr := &agentRunID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.review_generated",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"format":            format,
			"case_id":           agentRunDetail.CaseID,
			"payload_id":        agentRunDetail.PayloadID,
			"interaction_count": interactionSummary.Total,
			"evidence_strength": func() string {
				if evidence != nil {
					return string(evidence.EvidenceStrength)
				}
				return "none"
			}(),
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		// Log error but don't fail the review generation
		fmt.Printf("Warning: failed to create audit log: %v\n", err)
	}

	return packet, nil
}

// buildInteractionSummary builds interaction summary from real interactions
func (s *ReviewService) buildInteractionSummary(caseID, payloadID string) (*AgentRunInteractionSummary, error) {
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
	}

	summary := &AgentRunInteractionSummary{
		Total: len(interactions),
	}

	// Count by protocol and unique sources
	sourceSet := make(map[string]bool)
	for _, interaction := range interactions {
		switch interaction.Type {
		case "dns":
			summary.DNSCount++
		case "http":
			summary.HTTPCount++
		}
		if interaction.SourceIP != "" {
			sourceSet[interaction.SourceIP] = true
		}
	}
	summary.UniqueSources = len(sourceSet)

	// Get last interaction
	if len(interactions) > 0 {
		summary.LastInteractionAt = &interactions[len(interactions)-1].Timestamp
	}

	return summary, nil
}

// getAuditRefs gets audit references for an agent run
func (s *ReviewService) getAuditRefs(agentRunID string) ([]AgentRunAuditRef, error) {
	// Query audit logs for this agent run
	var auditLogs []models.AuditLog
	err := s.engine.Where("details LIKE ?", fmt.Sprintf("%%\"agent_run_id\":\"%s\"%%", agentRunID)).
		OrderBy("timestamp ASC").
		Limit(100).
		Find(&auditLogs)
	if err != nil {
		return nil, err
	}

	refs := make([]AgentRunAuditRef, 0, len(auditLogs))
	for _, log := range auditLogs {
		ref := AgentRunAuditRef{
			ID:           log.ID,
			Action:       log.Action,
			ResourceType: log.ResourceType,
			ResourceID: func() string {
				if log.ResourceID != nil {
					return *log.ResourceID
				}
				return ""
			}(),
			Timestamp: log.Timestamp,
		}
		refs = append(refs, ref)
	}

	return refs, nil
}

// generateJSONContent generates JSON content for the review packet
func (s *ReviewService) generateJSONContent(packet *AgentRunReviewPacket) string {
	data, _ := json.MarshalIndent(packet, "", "  ")
	return string(data)
}

// generateMarkdownContent generates Markdown content for the review packet
func (s *ReviewService) generateMarkdownContent(packet *AgentRunReviewPacket) string {
	md := fmt.Sprintf("# Agent Run Review Packet\n\n")
	md += fmt.Sprintf("## Agent Run\n\n")
	md += fmt.Sprintf("- **ID**: %s\n", packet.AgentRun.ID)
	md += fmt.Sprintf("- **Title**: %s\n", packet.AgentRun.Title)
	md += fmt.Sprintf("- **Status**: %s\n", packet.AgentRun.Status)
	md += fmt.Sprintf("- **Target**: %s\n", packet.Target)
	if packet.CaseID != "" {
		md += fmt.Sprintf("- **Case ID**: %s\n", packet.CaseID)
	}
	if packet.PayloadID != "" {
		md += fmt.Sprintf("- **Payload ID**: %s\n", packet.PayloadID)
	}
	md += fmt.Sprintf("- **Started At**: %s\n", func() string {
		if packet.AgentRun.StartedAt != nil {
			return packet.AgentRun.StartedAt.Format(time.RFC3339)
		}
		return "N/A"
	}())
	md += fmt.Sprintf("- **Ended At**: %s\n", func() string {
		if packet.AgentRun.EndedAt != nil {
			return packet.AgentRun.EndedAt.Format(time.RFC3339)
		}
		return "N/A"
	}())

	md += fmt.Sprintf("\n## Interaction Summary\n\n")
	md += fmt.Sprintf("- **Total**: %d\n", packet.InteractionSummary.Total)
	md += fmt.Sprintf("- **DNS Count**: %d\n", packet.InteractionSummary.DNSCount)
	md += fmt.Sprintf("- **HTTP Count**: %d\n", packet.InteractionSummary.HTTPCount)
	md += fmt.Sprintf("- **Unique Sources**: %d\n", packet.InteractionSummary.UniqueSources)
	md += fmt.Sprintf("- **Last Interaction At**: %s\n", func() string {
		if packet.InteractionSummary.LastInteractionAt != nil {
			return packet.InteractionSummary.LastInteractionAt.Format(time.RFC3339)
		}
		return "N/A"
	}())

	if packet.Evidence != nil {
		md += fmt.Sprintf("\n## Evidence\n\n")
		md += fmt.Sprintf("- **Evidence Strength**: %s\n", packet.Evidence.EvidenceStrength)
		md += fmt.Sprintf("- **Confidence**: %d%%\n", packet.Evidence.Confidence)
		md += fmt.Sprintf("- **Explainability**: %s\n", packet.Evidence.Explainability)
	}

	md += fmt.Sprintf("\n## Operations Timeline\n\n")
	for _, op := range packet.AgentRun.Operations {
		md += fmt.Sprintf("- **%s** (%s): %s\n", op.StartedAt.Format(time.RFC3339), op.Action, func() string {
			if op.Error != "" {
				return "Error: " + op.Error
			}
			return "Success"
		}())
	}

	md += fmt.Sprintf("\n## Audit References\n\n")
	for _, ref := range packet.AuditRefs {
		md += fmt.Sprintf("- **%s**: %s (%s)\n", ref.Timestamp.Format(time.RFC3339), ref.Action, ref.ResourceType)
	}

	md += fmt.Sprintf("\n---\n\n")
	md += fmt.Sprintf("Generated at: %s\n", packet.GeneratedAt.Format(time.RFC3339))

	return md
}
