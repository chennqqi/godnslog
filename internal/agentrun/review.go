package agentrun

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// ErrAgentRunNotFound is returned when an agent run is not found
var ErrAgentRunNotFound = errors.New("agent run not found")

// ErrInvalidPackageHash is returned when a package hash is invalid
var ErrInvalidPackageHash = errors.New("invalid package hash: must be 64-character hex string")

// packageHashRegex validates SHA-256 hex strings (64 hex characters)
var packageHashRegex = regexp.MustCompile(`^[a-fA-F0-9]{64}$`)

// ValidatePackageHash validates that a string is a valid 64-character hex string (SHA-256 hash)
func ValidatePackageHash(packageHash string) error {
	if packageHash == "" {
		return ErrInvalidPackageHash
	}
	if !packageHashRegex.MatchString(packageHash) {
		return ErrInvalidPackageHash
	}
	return nil
}

// ReviewService provides agent run review packet generation services
type ReviewService struct {
	engine             *xorm.Engine
	agentRunService    *Service
	authService        *auth.Service
	evidenceService    *interaction.EvidenceService
	interactionService *interaction.Service
	urlValidator       func(string) error // For testing: inject custom URL validator
	httpClient         *http.Client       // For testing: inject custom HTTP client
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
		ID:                 agentRunID,
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

// ExportReviewPackage exports a review evidence package for an agent run
func (s *ReviewService) ExportReviewPackage(agentRunID string, req *models.AgentRunReviewExportRequest, userID string) (*models.AgentRunReviewExportResponse, error) {
	// Validate format
	if req.Format != "json" && req.Format != "markdown" {
		return nil, errors.New("invalid format: must be json or markdown")
	}

	// Get agent run
	agentRun, err := s.agentRunService.GetAgentRunByID(agentRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent run: %w", err)
	}
	if agentRun == nil {
		return nil, ErrAgentRunNotFound
	}

	// Validate review_packet_id if provided
	if req.ReviewPacketID != "" && req.ReviewPacketID != agentRunID {
		return nil, errors.New("review_packet_id must match the current agent run")
	}

	// Build review packet
	packet, err := s.BuildReviewPacket(agentRunID, req.Format, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build review packet: %w", err)
	}

	// Get most recent review decision operation
	var recentDecisionOp models.AgentOperation
	has, err := s.engine.Where("agent_run_i_d = ? AND action LIKE ?", agentRunID, "review_decision.%").
		OrderBy("created_at DESC").Limit(1).Get(&recentDecisionOp)
	var decision string
	if err == nil && has {
		// Extract decision from action name
		parts := strings.Split(recentDecisionOp.Action, ".")
		if len(parts) == 2 {
			decision = parts[1]
		}
	}

	now := time.Now()

	// Build export package
	var exportPackage map[string]interface{}
	var content string

	if req.Format == "json" {
		exportPackage = map[string]interface{}{
			"agent_run": map[string]interface{}{
				"id":         packet.AgentRun.ID,
				"agent_id":   packet.AgentRun.AgentID,
				"case_id":    packet.AgentRun.CaseID,
				"payload_id": packet.AgentRun.PayloadID,
				"target":     packet.Target,
				"status":     packet.AgentRun.Status,
			},
			"review_packet": map[string]interface{}{
				"id": packet.ID,
				"evidence_strength": func() string {
					if packet.Evidence != nil {
						return string(packet.Evidence.EvidenceStrength)
					}
					return "none"
				}(),
				"confidence": func() int {
					if packet.Evidence != nil {
						return packet.Evidence.Confidence
					}
					return 0
				}(),
				"interaction_count": packet.InteractionSummary.Total,
				"unique_sources":    packet.InteractionSummary.UniqueSources,
			},
			"review_decision": func() map[string]interface{} {
				if decision != "" {
					resultMap := make(map[string]interface{})
					json.Unmarshal([]byte(recentDecisionOp.Result), &resultMap)
					return map[string]interface{}{
						"decision":     decision,
						"reason":       resultMap["reason"],
						"operation_id": recentDecisionOp.ID,
						"audit_ref_id": resultMap["audit_ref_id"],
					}
				}
				return nil
			}(),
			"links": map[string]interface{}{
				"case_url":     packet.AgentRun.CaseURL,
				"payload_url":  packet.AgentRun.PayloadURL,
				"evidence_url": packet.AgentRun.EvidenceURL,
				"audit_url": func() string {
					if req.IncludeAudit {
						return fmt.Sprintf("/dashboard/audit?resource_type=agent_run&resource_id=%s", agentRunID)
					}
					return ""
				}(),
			},
		}
	} else {
		content = s.generateExportMarkdownContent(packet, decision, recentDecisionOp)
	}

	// Compute package hash for both JSON and Markdown formats
	var packageHash string
	var manifest *models.AgentRunReviewPackageManifest
	if req.Format == "json" && exportPackage != nil {
		hash, err := models.ComputeDeterministicHash(exportPackage)
		if err != nil {
			return nil, fmt.Errorf("failed to compute package hash: %w", err)
		}
		packageHash = hash

		// Build manifest
		manifest = &models.AgentRunReviewPackageManifest{
			SchemaVersion:  "review-package-manifest/v1",
			AgentRunID:     agentRunID,
			ReviewPacketID: req.ReviewPacketID,
			Format:         req.Format,
			PackageHash:    packageHash,
			HashAlgorithm:  "sha256",
			GeneratedAt:    now,
			Refs: map[string]string{
				"operation_id": "", // Will be filled after operation creation
				"audit_ref_id": "", // Will be filled after audit creation
			},
		}
	} else if req.Format == "markdown" && content != "" {
		// Compute hash of exact Markdown content bytes
		hash := sha256.Sum256([]byte(content))
		packageHash = hex.EncodeToString(hash[:])

		// Build manifest for Markdown
		manifest = &models.AgentRunReviewPackageManifest{
			SchemaVersion:  "review-package-manifest/v1",
			AgentRunID:     agentRunID,
			ReviewPacketID: req.ReviewPacketID,
			Format:         req.Format,
			PackageHash:    packageHash,
			HashAlgorithm:  "sha256",
			GeneratedAt:    now,
			Refs: map[string]string{
				"operation_id": "", // Will be filled after operation creation
				"audit_ref_id": "", // Will be filled after audit creation
			},
		}
	}

	// Create review_export operation
	result := map[string]interface{}{
		"format":           req.Format,
		"agent_run_id":     agentRunID,
		"review_packet_id": req.ReviewPacketID,
		"decision":         decision,
		"audit_action":     "agent_run.review_exported",
		"exported_at":      now,
		"package_hash":     packageHash,
	}

	opReq := &models.AgentOperationCreateRequest{
		Action:    "review_export." + req.Format,
		RiskLevel: "low",
		Request: map[string]interface{}{
			"format":           req.Format,
			"review_packet_id": req.ReviewPacketID,
			"include_audit":    req.IncludeAudit,
		},
		Result: result,
	}

	if err := s.agentRunService.AppendAgentOperation(agentRunID, opReq, userID); err != nil {
		return nil, fmt.Errorf("failed to append export operation: %w", err)
	}

	// Get the created operation
	var operation models.AgentOperation
	has, err = s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_export."+req.Format).
		OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve created operation: has=%v err=%v", has, err)
	}

	// Create audit log
	userIDPtr := &userID
	resourceIDPtr := &agentRunID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.review_exported",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"format":           req.Format,
			"review_packet_id": req.ReviewPacketID,
			"decision":         decision,
			"operation_id":     operation.ID,
			"package_hash":     packageHash,
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create export audit log: %w", err)
	}

	// Update manifest refs
	if manifest != nil {
		manifest.Refs["operation_id"] = operation.ID
		manifest.Refs["audit_ref_id"] = auditLog.ID
	}

	return &models.AgentRunReviewExportResponse{
		AgentRunID:     agentRunID,
		Format:         req.Format,
		OperationID:    operation.ID,
		AuditRefID:     auditLog.ID,
		ReviewPacketID: req.ReviewPacketID,
		Decision:       decision,
		Content:        content,
		Package:        exportPackage,
		Manifest:       manifest,
		PackageHash:    packageHash,
		GeneratedAt:    now,
	}, nil
}

// generateExportMarkdownContent generates markdown content for export package
func (s *ReviewService) generateExportMarkdownContent(packet *AgentRunReviewPacket, decision string, decisionOp models.AgentOperation) string {
	md := fmt.Sprintf("# Agent Run Review Evidence Package\n\n")

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

	md += fmt.Sprintf("\n## Evidence Summary\n\n")
	md += fmt.Sprintf("- **Total Interactions**: %d\n", packet.InteractionSummary.Total)
	md += fmt.Sprintf("- **DNS Count**: %d\n", packet.InteractionSummary.DNSCount)
	md += fmt.Sprintf("- **HTTP Count**: %d\n", packet.InteractionSummary.HTTPCount)
	md += fmt.Sprintf("- **Unique Sources**: %d\n", packet.InteractionSummary.UniqueSources)
	if packet.Evidence != nil {
		md += fmt.Sprintf("- **Evidence Strength**: %s\n", packet.Evidence.EvidenceStrength)
		md += fmt.Sprintf("- **Confidence**: %d%%\n", packet.Evidence.Confidence)
	}

	if decision != "" {
		md += fmt.Sprintf("\n## Review Decision\n\n")
		md += fmt.Sprintf("- **Decision**: %s\n", decision)
		resultMap := make(map[string]interface{})
		json.Unmarshal([]byte(decisionOp.Result), &resultMap)
		if reason, ok := resultMap["reason"].(string); ok && reason != "" {
			md += fmt.Sprintf("- **Reason**: %s\n", reason)
		}
		md += fmt.Sprintf("- **Operation ID**: %s\n", decisionOp.ID)
		if auditRefID, ok := resultMap["audit_ref_id"].(string); ok && auditRefID != "" {
			md += fmt.Sprintf("- **Audit Ref ID**: %s\n", auditRefID)
		}
	}

	md += fmt.Sprintf("\n## Timeline References\n\n")
	for _, op := range packet.AgentRun.Operations {
		md += fmt.Sprintf("- **%s**: %s\n", op.StartedAt.Format(time.RFC3339), op.Action)
	}

	md += fmt.Sprintf("\n## Audit References\n\n")
	for _, ref := range packet.AuditRefs {
		md += fmt.Sprintf("- **%s**: %s (%s)\n", ref.Timestamp.Format(time.RFC3339), ref.Action, ref.ResourceType)
	}

	md += fmt.Sprintf("\n## Links\n\n")
	if packet.AgentRun.CaseURL != "" {
		md += fmt.Sprintf("- **Case**: %s\n", packet.AgentRun.CaseURL)
	}
	if packet.AgentRun.PayloadURL != "" {
		md += fmt.Sprintf("- **Payload**: %s\n", packet.AgentRun.PayloadURL)
	}
	if packet.AgentRun.EvidenceURL != "" {
		md += fmt.Sprintf("- **Evidence**: %s\n", packet.AgentRun.EvidenceURL)
	}
	md += fmt.Sprintf("- **Audit**: /dashboard/audit?resource_type=agent_run&resource_id=%s\n", packet.AgentRun.ID)

	md += fmt.Sprintf("\n---\n\n")
	md += fmt.Sprintf("Generated at: %s\n", time.Now().Format(time.RFC3339))

	return md
}

// DeliverReviewPackage delivers a review evidence package to a webhook
func (s *ReviewService) DeliverReviewPackage(agentRunID string, req *models.AgentRunReviewDeliveryRequest, userID string) (*models.AgentRunReviewDeliveryResponse, error) {
	// Validate format
	if req.Format != "json" && req.Format != "markdown" {
		return nil, errors.New("invalid format: must be json or markdown")
	}

	// Validate webhook URL
	validator := s.urlValidator
	if validator == nil {
		validator = ValidateWebhookURL
	}
	if err := validator(req.WebhookURL); err != nil {
		return nil, fmt.Errorf("invalid webhook URL: %w", err)
	}

	// Validate headers
	if err := ValidateWebhookHeaders(req.Headers); err != nil {
		return nil, fmt.Errorf("invalid headers: %w", err)
	}

	// Get agent run
	agentRun, err := s.agentRunService.GetAgentRunByID(agentRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent run: %w", err)
	}
	if agentRun == nil {
		return nil, ErrAgentRunNotFound
	}

	// Validate review_packet_id if provided
	if req.ReviewPacketID != "" && req.ReviewPacketID != agentRunID {
		return nil, errors.New("review_packet_id must match the current agent run")
	}

	// First, export the review package to get the content
	exportReq := &models.AgentRunReviewExportRequest{
		Format:         req.Format,
		ReviewPacketID: req.ReviewPacketID,
		IncludeAudit:   req.IncludeAudit,
	}
	exportResp, err := s.ExportReviewPackage(agentRunID, exportReq, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to export review package: %w", err)
	}

	// Generate delivery ID
	deliveryID := generateID()
	now := time.Now()

	// Parse webhook URL to get destination host
	parsedURL, err := url.Parse(req.WebhookURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook URL: %w", err)
	}
	destinationHost := parsedURL.Hostname()

	// Create pending delivery operation first to get real operation ID
	pendingOperationID, err := s.createPendingDeliveryOperation(agentRunID, req, destinationHost, exportResp.OperationID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to create pending delivery operation: %w", err)
	}

	// Build webhook payload with real operation ID
	webhookPayload := map[string]interface{}{
		"event":        "agent_run.review_evidence_delivered",
		"agent_run_id": agentRunID,
		"format":       req.Format,
		"delivery_id":  deliveryID,
		"generated_at": now.Format(time.RFC3339),
		"package": map[string]interface{}{
			"content": exportResp.Content,
		},
		"refs": map[string]interface{}{
			"export_operation_id":   exportResp.OperationID,
			"delivery_operation_id": pendingOperationID,
			"audit_ref_id":          exportResp.AuditRefID,
			"package_hash":          exportResp.PackageHash,
		},
		"package_hash": exportResp.PackageHash,
	}

	// For JSON format, include the package in the payload
	if req.Format == "json" && exportResp.Package != nil {
		webhookPayload["package"] = exportResp.Package
	}

	// Marshal payload
	payloadBytes, err := json.Marshal(webhookPayload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	// Create HTTP request with timeout
	client := s.httpClient
	if client == nil {
		client = &http.Client{
			Timeout: 5 * time.Second,
			// Do not follow redirects
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
	}

	httpReq, err := http.NewRequest("POST", req.WebhookURL, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set headers
	httpReq.Header.Set("Content-Type", "application/json")
	for key, value := range req.Headers {
		httpReq.Header.Set(key, value)
	}

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		// Check if error is a timeout
		if netErr, ok := err.(interface{ Timeout() bool }); ok && netErr.Timeout() {
			timeoutErr := fmt.Errorf("webhook request timed out")
			return s.createDeliveryFailure(agentRunID, req, deliveryID, destinationHost, exportResp.OperationID, timeoutErr, userID)
		}
		// Create failure operation and audit
		return s.createDeliveryFailure(agentRunID, req, deliveryID, destinationHost, exportResp.OperationID, err, userID)
	}
	defer resp.Body.Close()

	// Read response body (but don't store it)
	_, _ = io.Copy(io.Discard, resp.Body)

	// Check status code
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		err := fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
		return s.createDeliveryFailure(agentRunID, req, deliveryID, destinationHost, exportResp.OperationID, err, userID)
	}

	// Create success operation and audit
	return s.createDeliverySuccess(agentRunID, req, deliveryID, destinationHost, exportResp.OperationID, resp.StatusCode, userID)
}

// createPendingDeliveryOperation creates a pending delivery operation and returns its ID
func (s *ReviewService) createPendingDeliveryOperation(agentRunID string, req *models.AgentRunReviewDeliveryRequest, destinationHost, exportOperationID string, userID string) (string, error) {
	// Extract header names for operation request (sanitized)
	headerNames := make([]string, 0, len(req.Headers))
	for headerName := range req.Headers {
		headerNames = append(headerNames, headerName)
	}

	// Create review_delivery.webhook operation with pending status
	result := map[string]interface{}{
		"result":              "pending",
		"status_code":         0,
		"export_operation_id": exportOperationID,
	}

	opReq := &models.AgentOperationCreateRequest{
		Action:    "review_delivery.webhook",
		RiskLevel: "low",
		Request: map[string]interface{}{
			"format":           req.Format,
			"review_packet_id": req.ReviewPacketID,
			"destination_host": destinationHost,
			"include_audit":    req.IncludeAudit,
			"header_names":     headerNames,
		},
		Result: result,
	}

	if err := s.agentRunService.AppendAgentOperation(agentRunID, opReq, userID); err != nil {
		return "", fmt.Errorf("failed to append pending delivery operation: %w", err)
	}

	// Get the created operation
	var operation models.AgentOperation
	has, err := s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_delivery.webhook").
		OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return "", fmt.Errorf("failed to retrieve created operation: has=%v err=%v", has, err)
	}

	return operation.ID, nil
}

// createDeliverySuccess creates success operation and audit records
func (s *ReviewService) createDeliverySuccess(agentRunID string, req *models.AgentRunReviewDeliveryRequest, deliveryID, destinationHost, exportOperationID string, statusCode int, userID string) (*models.AgentRunReviewDeliveryResponse, error) {
	now := time.Now()

	// Get the export response to retrieve package_hash
	var exportOperation models.AgentOperation
	has, err := s.engine.ID(exportOperationID).Get(&exportOperation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve export operation: has=%v err=%v", has, err)
	}

	var packageHash string
	var exportResult map[string]interface{}
	if exportOperation.Result != "" {
		if err := json.Unmarshal([]byte(exportOperation.Result), &exportResult); err == nil {
			if hash, ok := exportResult["package_hash"].(string); ok {
				packageHash = hash
			}
		}
	}

	// Get the pending delivery operation
	var operation models.AgentOperation
	has, err = s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_delivery.webhook").
		OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve pending operation: has=%v err=%v", has, err)
	}

	// Update operation result to delivered
	resultMap := map[string]interface{}{
		"result":              "delivered",
		"status_code":         statusCode,
		"delivery_id":         deliveryID,
		"export_operation_id": exportOperationID,
		"audit_action":        "agent_run.review_delivered",
		"package_hash":        packageHash,
	}
	resultJSON, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operation result: %w", err)
	}
	operation.Result = string(resultJSON)
	if _, err := s.engine.ID(operation.ID).Cols("result").Update(&operation); err != nil {
		return nil, fmt.Errorf("failed to update delivery operation: %w", err)
	}

	// Create audit log
	userIDPtr := &userID
	resourceIDPtr := &agentRunID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.review_delivered",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"format":                req.Format,
			"delivery_id":           deliveryID,
			"delivery_operation_id": operation.ID,
			"export_operation_id":   exportOperationID,
			"destination_host":      destinationHost,
			"status_code":           statusCode,
			"package_hash":          packageHash,
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create delivery audit log: %w", err)
	}

	return &models.AgentRunReviewDeliveryResponse{
		AgentRunID:        agentRunID,
		Format:            req.Format,
		DeliveryID:        deliveryID,
		DeliveryOperation: operation.ID,
		ExportOperationID: exportOperationID,
		AuditRefID:        auditLog.ID,
		DestinationHost:   destinationHost,
		StatusCode:        statusCode,
		Result:            "delivered",
		DeliveredAt:       now,
		PackageHash:       packageHash,
	}, nil
}

// createDeliveryFailure creates failure operation and audit records
func (s *ReviewService) createDeliveryFailure(agentRunID string, req *models.AgentRunReviewDeliveryRequest, deliveryID, destinationHost, exportOperationID string, deliveryError error, userID string) (*models.AgentRunReviewDeliveryResponse, error) {
	now := time.Now()

	// Get the export response to retrieve package_hash
	var exportOperation models.AgentOperation
	has, err := s.engine.ID(exportOperationID).Get(&exportOperation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve export operation: has=%v err=%v", has, err)
	}

	var packageHash string
	var exportResult map[string]interface{}
	if exportOperation.Result != "" {
		if err := json.Unmarshal([]byte(exportOperation.Result), &exportResult); err == nil {
			if hash, ok := exportResult["package_hash"].(string); ok {
				packageHash = hash
			}
		}
	}

	// Get the pending delivery operation
	var operation models.AgentOperation
	has, err = s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_delivery.webhook").
		OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve pending operation: has=%v err=%v", has, err)
	}

	// Update operation result to failed
	resultMap := map[string]interface{}{
		"result":       "failed",
		"status_code":  0,
		"error":        deliveryError.Error(),
		"delivery_id":  deliveryID,
		"package_hash": packageHash,
	}
	resultJSON, err := json.Marshal(resultMap)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal operation result: %w", err)
	}
	operation.Result = string(resultJSON)
	if _, err := s.engine.ID(operation.ID).Cols("result").Update(&operation); err != nil {
		return nil, fmt.Errorf("failed to update delivery operation: %w", err)
	}

	// Create audit log for failure
	userIDPtr := &userID
	resourceIDPtr := &agentRunID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.review_delivery_failed",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"format":                req.Format,
			"delivery_id":           deliveryID,
			"delivery_operation_id": operation.ID,
			"destination_host":      destinationHost,
			"error":                 deliveryError.Error(),
			"package_hash":          packageHash,
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create delivery failure audit log: %w", err)
	}

	return nil, fmt.Errorf("delivery failed: %w", deliveryError)
}

// ListReviewDeliveries returns the delivery history for an agent run
func (s *ReviewService) ListReviewDeliveries(agentRunID string) (*models.AgentRunReviewDeliveryHistoryResponse, error) {
	// Validate the Agent Run exists
	var agentRun models.AgentRun
	has, err := s.engine.ID(agentRunID).Get(&agentRun)
	if err != nil {
		return nil, fmt.Errorf("failed to query agent run: %w", err)
	}
	if !has {
		return nil, ErrAgentRunNotFound
	}

	// Query review_delivery.webhook operations
	var operations []models.AgentOperation
	err = s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_delivery.webhook").
		OrderBy("created_at DESC").
		Find(&operations)
	if err != nil {
		return nil, fmt.Errorf("failed to query delivery operations: %w", err)
	}

	// Build history items
	items := make([]models.AgentRunReviewDeliveryHistoryItem, 0, len(operations))
	summary := models.AgentRunReviewDeliverySummary{
		Total:     len(operations),
		Delivered: 0,
		Failed:    0,
		Timeout:   0,
	}

	for _, op := range operations {
		item := models.AgentRunReviewDeliveryHistoryItem{
			DeliveryOperationID: op.ID,
			CreatedAt:           op.StartedAt,
		}

		// Parse operation request
		var request map[string]interface{}
		if op.Request != "" {
			if err := json.Unmarshal([]byte(op.Request), &request); err == nil {
				if format, ok := request["format"].(string); ok {
					item.Format = format
				}
				if destHost, ok := request["destination_host"].(string); ok {
					item.DestinationHost = destHost
				}
				if headerNames, ok := request["header_names"].([]interface{}); ok {
					names := make([]string, 0, len(headerNames))
					for _, h := range headerNames {
						if name, ok := h.(string); ok {
							names = append(names, name)
						}
					}
					item.HeaderNames = names
				}
			}
		}

		// Parse operation result
		var result map[string]interface{}
		if op.Result != "" {
			if err := json.Unmarshal([]byte(op.Result), &result); err == nil {
				if deliveryID, ok := result["delivery_id"].(string); ok {
					item.DeliveryID = deliveryID
				}
				if exportOpID, ok := result["export_operation_id"].(string); ok {
					item.ExportOperationID = exportOpID
				}
				if statusCode, ok := result["status_code"].(float64); ok {
					item.StatusCode = int(statusCode)
				}
				if res, ok := result["result"].(string); ok {
					item.Result = res
				}
				if errorSummary, ok := result["error"].(string); ok {
					item.ErrorSummary = errorSummary
				}
				if deliveredAtStr, ok := result["delivered_at"].(string); ok {
					if deliveredAt, err := time.Parse(time.RFC3339, deliveredAtStr); err == nil {
						item.DeliveredAt = deliveredAt
					}
				}
				if packageHash, ok := result["package_hash"].(string); ok {
					item.PackageHash = packageHash
				}
			}
		}

		// Derive result from error summary if result is failed
		// Real timeout operations save result="failed" with timeout error message
		if item.Result == "failed" && item.ErrorSummary != "" {
			if strings.Contains(strings.ToLower(item.ErrorSummary), "timeout") || strings.Contains(strings.ToLower(item.ErrorSummary), "timed out") {
				item.Result = "timeout"
			}
		}

		// Derive result from error summary if not set
		if item.Result == "" && item.ErrorSummary != "" {
			if strings.Contains(strings.ToLower(item.ErrorSummary), "timeout") || strings.Contains(strings.ToLower(item.ErrorSummary), "timed out") {
				item.Result = "timeout"
			} else {
				item.Result = "failed"
			}
		}

		// Resolve audit_ref_id by matching audit details
		var auditLog models.AuditLog
		has, err := s.engine.Where("action = ? AND details LIKE ?", "agent_run.review_delivered", "%"+op.ID+"%").
			Or("action = ? AND details LIKE ?", "agent_run.review_delivery_failed", "%"+op.ID+"%").
			OrderBy("timestamp DESC").
			Limit(1).
			Get(&auditLog)
		if err == nil && has && auditLog.ID != "" {
			item.AuditRefID = auditLog.ID
		}

		// Update summary counts
		switch item.Result {
		case "delivered":
			summary.Delivered++
		case "failed":
			summary.Failed++
		case "timeout":
			summary.Timeout++
		}

		items = append(items, item)
	}

	return &models.AgentRunReviewDeliveryHistoryResponse{
		AgentRunID: agentRunID,
		Summary:    summary,
		Items:      items,
	}, nil
}

// TraceReviewPackageByHash traces a review package by its hash
func (s *ReviewService) TraceReviewPackageByHash(packageHash string) (*models.AgentRunReviewPackageTraceResponse, error) {
	response := &models.AgentRunReviewPackageTraceResponse{
		PackageHash: packageHash,
		Summary: models.AgentRunReviewPackageTraceSummary{
			AgentRunCount: 0,
			ExportCount:   0,
			DeliveryCount: 0,
			AuditCount:    0,
			Delivered:     0,
			Failed:        0,
			Timeout:       0,
		},
		AgentRuns:  make([]models.AgentRunReviewPackageTraceRun, 0),
		Exports:    make([]models.AgentRunReviewPackageTraceExport, 0),
		Deliveries: make([]models.AgentRunReviewPackageTraceDelivery, 0),
		Audits:     make([]models.AgentRunReviewPackageTraceAudit, 0),
	}

	// Track seen operation IDs and audit IDs for deduplication
	seenOperationIDs := make(map[string]bool)
	seenAuditIDs := make(map[string]bool)
	seenAgentRunIDs := make(map[string]bool)

	// Query AgentOperations with package_hash in result
	var operations []models.AgentOperation
	err := s.engine.Where("result LIKE ?", "%"+packageHash+"%").
		Find(&operations)
	if err != nil {
		return nil, fmt.Errorf("failed to query operations with package_hash: %w", err)
	}

	for _, op := range operations {
		if seenOperationIDs[op.ID] {
			continue
		}
		seenOperationIDs[op.ID] = true

		// Parse result to verify package_hash is actually present
		var result map[string]interface{}
		if op.Result != "" {
			if err := json.Unmarshal([]byte(op.Result), &result); err == nil {
				if hash, ok := result["package_hash"].(string); ok && hash == packageHash {
					// This operation actually contains the package_hash
					switch op.Action {
					case "review_export.json", "review_export.markdown":
						response.Summary.ExportCount++
						response.Exports = append(response.Exports, s.buildTraceExportFromOperation(op, result))
					case "review_delivery.webhook":
						response.Summary.DeliveryCount++
						delivery := s.buildTraceDeliveryFromOperation(op, result)
						response.Deliveries = append(response.Deliveries, delivery)
						// Update summary counts based on result
						if delivery.Result == "delivered" {
							response.Summary.Delivered++
						} else if delivery.Result == "failed" {
							response.Summary.Failed++
						} else if delivery.Result == "timeout" {
							response.Summary.Timeout++
						}
					}

					// Track Agent Run
					if !seenAgentRunIDs[op.AgentRunID] {
						seenAgentRunIDs[op.AgentRunID] = true
						response.Summary.AgentRunCount++
						agentRunRef := s.buildTraceRunFromAgentRunID(op.AgentRunID)
						if agentRunRef != nil {
							response.AgentRuns = append(response.AgentRuns, *agentRunRef)
						}
					}
				}
			}
		}
	}

	// Query AuditLogs with package_hash in details
	var auditLogs []models.AuditLog
	err = s.engine.Where("details LIKE ?", "%"+packageHash+"%").
		Find(&auditLogs)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs with package_hash: %w", err)
	}

	for _, audit := range auditLogs {
		if seenAuditIDs[audit.ID] {
			continue
		}
		seenAuditIDs[audit.ID] = true

		// Verify package_hash is actually in details
		if hash, ok := audit.Details["package_hash"].(string); ok && hash == packageHash {
			response.Summary.AuditCount++
			response.Audits = append(response.Audits, s.buildTraceAuditFromAuditLog(audit))

			// Track Agent Run if resource_type is agent_run
			if audit.ResourceType == "agent_run" && audit.ResourceID != nil && !seenAgentRunIDs[*audit.ResourceID] {
				seenAgentRunIDs[*audit.ResourceID] = true
				response.Summary.AgentRunCount++
				agentRunRef := s.buildTraceRunFromAgentRunID(*audit.ResourceID)
				if agentRunRef != nil {
					response.AgentRuns = append(response.AgentRuns, *agentRunRef)
				}
			}
		}
	}

	return response, nil
}

// buildTraceExportFromOperation builds a trace export from an operation
func (s *ReviewService) buildTraceExportFromOperation(op models.AgentOperation, result map[string]interface{}) models.AgentRunReviewPackageTraceExport {
	export := models.AgentRunReviewPackageTraceExport{
		AgentRunID:  op.AgentRunID,
		OperationID: op.ID,
		Format:      op.Action, // "review_export.json" or "review_export.markdown"
		CreatedAt:   op.StartedAt,
	}

	// Extract audit_ref_id from result
	if auditRefID, ok := result["audit_ref_id"].(string); ok {
		export.AuditRefID = auditRefID
	}

	// Extract review_packet_id from result
	if reviewPacketID, ok := result["review_packet_id"].(string); ok {
		export.ReviewPacketID = reviewPacketID
	}

	// Format field should be "json" or "markdown"
	if op.Action == "review_export.json" {
		export.Format = "json"
	} else if op.Action == "review_export.markdown" {
		export.Format = "markdown"
	}

	return export
}

// buildTraceDeliveryFromOperation builds a trace delivery from an operation
func (s *ReviewService) buildTraceDeliveryFromOperation(op models.AgentOperation, result map[string]interface{}) models.AgentRunReviewPackageTraceDelivery {
	delivery := models.AgentRunReviewPackageTraceDelivery{
		AgentRunID:          op.AgentRunID,
		DeliveryOperationID: op.ID,
		Format:              "json", // Default format, can be extracted from result if needed
		Result:              "unknown",
		CreatedAt:           op.StartedAt,
	}

	// Extract result field
	if res, ok := result["result"].(string); ok {
		delivery.Result = res
	}

	// Extract delivery_id
	if deliveryID, ok := result["delivery_id"].(string); ok {
		delivery.DeliveryID = deliveryID
	}

	// Extract export_operation_id
	if exportOpID, ok := result["export_operation_id"].(string); ok {
		delivery.ExportOperationID = exportOpID
	}

	// Extract audit_ref_id
	if auditRefID, ok := result["audit_ref_id"].(string); ok {
		delivery.AuditRefID = auditRefID
	}

	// Extract destination_host from request
	var request map[string]interface{}
	if op.Request != "" {
		if err := json.Unmarshal([]byte(op.Request), &request); err == nil {
			if webhookURL, ok := request["webhook_url"].(string); ok {
				// Extract host from URL for sanitization
				if u, err := url.Parse(webhookURL); err == nil {
					delivery.DestinationHost = u.Hostname()
				}
			}
		}
	}

	// Extract status_code
	if statusCode, ok := result["status_code"].(float64); ok {
		delivery.StatusCode = int(statusCode)
	}

	// Extract error summary
	if err, ok := result["error"].(string); ok {
		delivery.ErrorSummary = err
	}

	// Extract delivered_at
	if deliveredAt, ok := result["delivered_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, deliveredAt); err == nil {
			delivery.DeliveredAt = t
		}
	}

	return delivery
}

// buildTraceAuditFromAuditLog builds a trace audit from an audit log
func (s *ReviewService) buildTraceAuditFromAuditLog(audit models.AuditLog) models.AgentRunReviewPackageTraceAudit {
	auditRef := models.AgentRunReviewPackageTraceAudit{
		AuditRefID:   audit.ID,
		Action:       audit.Action,
		ResourceType: audit.ResourceType,
		Timestamp:    audit.Timestamp,
	}

	if audit.ResourceID != nil {
		auditRef.ResourceID = *audit.ResourceID
	}

	if audit.ResourceType == "agent_run" && audit.ResourceID != nil {
		auditRef.AgentRunID = *audit.ResourceID
	}

	// Build URL for audit filtered view
	if audit.ResourceType != "" && audit.ResourceID != nil {
		auditRef.URL = fmt.Sprintf("/dashboard/audit?resource_type=%s&resource_id=%s", audit.ResourceType, *audit.ResourceID)
	}

	return auditRef
}

// buildTraceRunFromAgentRunID builds a trace run from an agent run ID
func (s *ReviewService) buildTraceRunFromAgentRunID(agentRunID string) *models.AgentRunReviewPackageTraceRun {
	var agentRun models.AgentRun
	has, err := s.engine.ID(agentRunID).Get(&agentRun)
	if err != nil || !has {
		return nil
	}

	return &models.AgentRunReviewPackageTraceRun{
		AgentRunID: agentRun.ID,
		Title:      agentRun.Title,
		Status:     string(agentRun.Status),
		CaseID:     agentRun.CaseID,
		PayloadID:  agentRun.PayloadID,
		Target:     agentRun.Target,
		URL:        fmt.Sprintf("/dashboard/agent-runs/%s", agentRun.ID),
	}
}
