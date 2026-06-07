package agentrun

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
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

	// Create review_export operation
	result := map[string]interface{}{
		"format":           req.Format,
		"agent_run_id":     agentRunID,
		"review_packet_id": req.ReviewPacketID,
		"decision":         decision,
		"audit_action":     "agent_run.review_exported",
		"exported_at":      now,
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
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create export audit log: %w", err)
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

	// Build webhook payload
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
			"delivery_operation_id": deliveryID,
			"audit_ref_id":          exportResp.AuditRefID,
		},
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

// createDeliverySuccess creates success operation and audit records
func (s *ReviewService) createDeliverySuccess(agentRunID string, req *models.AgentRunReviewDeliveryRequest, deliveryID, destinationHost, exportOperationID string, statusCode int, userID string) (*models.AgentRunReviewDeliveryResponse, error) {
	now := time.Now()

	// Extract header names for operation request (sanitized)
	headerNames := make([]string, 0, len(req.Headers))
	for headerName := range req.Headers {
		headerNames = append(headerNames, headerName)
	}

	// Create review_delivery.webhook operation
	result := map[string]interface{}{
		"result":              "delivered",
		"status_code":         statusCode,
		"delivery_id":         deliveryID,
		"export_operation_id": exportOperationID,
		"audit_action":        "agent_run.review_delivered",
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
		return nil, fmt.Errorf("failed to append delivery operation: %w", err)
	}

	// Get the created operation
	var operation models.AgentOperation
	has, err := s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_delivery.webhook").
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
	}, nil
}

// createDeliveryFailure creates failure operation and audit records
func (s *ReviewService) createDeliveryFailure(agentRunID string, req *models.AgentRunReviewDeliveryRequest, deliveryID, destinationHost, exportOperationID string, deliveryError error, userID string) (*models.AgentRunReviewDeliveryResponse, error) {
	now := time.Now()

	// Extract header names for operation request (sanitized)
	headerNames := make([]string, 0, len(req.Headers))
	for headerName := range req.Headers {
		headerNames = append(headerNames, headerName)
	}

	// Create review_delivery.webhook operation with failure result
	result := map[string]interface{}{
		"result":      "failed",
		"status_code": 0,
		"error":       deliveryError.Error(),
		"delivery_id": deliveryID,
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
		return nil, fmt.Errorf("failed to append delivery operation: %w", err)
	}

	// Get the created operation
	var operation models.AgentOperation
	has, err := s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_delivery.webhook").
		OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve created operation: has=%v err=%v", has, err)
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
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create delivery failure audit log: %w", err)
	}

	return nil, fmt.Errorf("delivery failed: %w", deliveryError)
}
