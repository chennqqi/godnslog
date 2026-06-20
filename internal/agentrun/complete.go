package agentrun

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// CompleteAgentRunRequest is the request body for completing an agent run.
type CompleteAgentRunRequest struct {
	Format         string `json:"format"`
	Decision       string `json:"decision,omitempty"`
	DecisionReason string `json:"decision_reason,omitempty"`
	IncludeAudit   bool   `json:"include_audit,omitempty"`
}

// CompleteAgentRunResponse is the response body for completing an agent run.
type CompleteAgentRunResponse struct {
	AgentRunID       string      `json:"agent_run_id"`
	Status           string      `json:"status"`
	ReviewPacket     interface{} `json:"review_packet,omitempty"`
	ExportPackage    interface{} `json:"export_package,omitempty"`
	PackageHash      string      `json:"package_hash,omitempty"`
	Decision         string      `json:"decision,omitempty"`
	OperationID      string      `json:"operation_id,omitempty"`
	AuditRefID       string      `json:"audit_ref_id,omitempty"`
	InteractionCount int         `json:"interaction_count"`
	EvidenceStrength string      `json:"evidence_strength,omitempty"`
	CompletedAt      time.Time   `json:"completed_at"`
}

// CompleteAgentRun orchestrates the full agent run completion loop:
// 1. Verify the agent run exists and is not already terminal
// 2. Build a review packet (with interactions and evidence)
// 3. Optionally record a review decision
// 4. Export the review package
// 5. Update status to completed
// 6. Return the complete response with all artifacts
func (s *ReviewService) CompleteAgentRun(agentRunID string, req *CompleteAgentRunRequest, userID string) (*CompleteAgentRunResponse, error) {
	if req == nil {
		req = &CompleteAgentRunRequest{}
	}
	format := req.Format
	if format == "" {
		format = "json"
	}
	if format != "json" && format != "markdown" {
		return nil, fmt.Errorf("invalid format: must be json or markdown")
	}

	// 1. Get agent run and verify it's not terminal
	agentRun, err := s.agentRunService.GetAgentRunByID(agentRunID)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent run: %w", err)
	}
	if agentRun == nil {
		return nil, ErrAgentRunNotFound
	}
	if isTerminalStatus(agentRun.Status) {
		return nil, fmt.Errorf("agent run is already in terminal status: %s", agentRun.Status)
	}

	// 2. Build review packet
	packet, err := s.BuildReviewPacket(agentRunID, format, "")
	if err != nil {
		return nil, fmt.Errorf("failed to build review packet: %w", err)
	}

	// 3. Optionally record review decision
	if req.Decision != "" {
		validDecisions := map[string]bool{
			"accepted":              true,
			"false_positive":        true,
			"needs_manual_followup": true,
			"insufficient_evidence": true,
		}
		if !validDecisions[req.Decision] {
			return nil, fmt.Errorf("invalid decision: must be one of accepted, false_positive, needs_manual_followup, insufficient_evidence")
		}

		decisionReq := &models.AgentRunReviewDecisionRequest{
			Decision:       req.Decision,
			Reason:         req.DecisionReason,
			ReviewPacketID: agentRunID,
		}
		if _, err := s.agentRunService.RecordReviewDecision(agentRunID, decisionReq, userID); err != nil {
			return nil, fmt.Errorf("failed to record review decision: %w", err)
		}
	}

	// 4. Export review package
	exportReq := &models.AgentRunReviewExportRequest{
		Format:         format,
		ReviewPacketID: agentRunID,
		IncludeAudit:   req.IncludeAudit,
	}
	exportResp, err := s.ExportReviewPackage(agentRunID, exportReq, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to export review package: %w", err)
	}

	// 5. Update status to completed
	statusReq := &models.AgentRunUpdateStatusRequest{
		Status: models.AgentRunStatusCompleted,
	}
	if err := s.agentRunService.UpdateAgentRunStatus(agentRunID, statusReq, userID); err != nil {
		return nil, fmt.Errorf("failed to update agent run status to completed: %w", err)
	}

	// 6. Build response
	resp := &CompleteAgentRunResponse{
		AgentRunID:       agentRunID,
		Status:           string(models.AgentRunStatusCompleted),
		ReviewPacket:     packet,
		ExportPackage:    exportResp.Package,
		PackageHash:      exportResp.PackageHash,
		Decision:         req.Decision,
		OperationID:      exportResp.OperationID,
		AuditRefID:       exportResp.AuditRefID,
		InteractionCount: packet.InteractionSummary.Total,
		CompletedAt:      time.Now(),
	}

	if packet.Evidence != nil {
		resp.EvidenceStrength = string(packet.Evidence.EvidenceStrength)
	}

	// 7. Create audit log for completion
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       &userID,
		Action:       "agent_run.completed",
		ResourceType: "agent_run",
		ResourceID:   &agentRunID,
		Details: models.AuditDetails{
			"format":              format,
			"interaction_count":   packet.InteractionSummary.Total,
			"evidence_strength":   resp.EvidenceStrength,
			"package_hash":        exportResp.PackageHash,
			"export_operation_id": exportResp.OperationID,
			"decision":            req.Decision,
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		// Log but don't fail the completion
		fmt.Printf("Warning: failed to create completion audit log: %v\n", err)
	}

	return resp, nil
}

// CompleteAgentRunJSON is a convenience method that returns the completion result as raw JSON bytes.
func (s *ReviewService) CompleteAgentRunJSON(agentRunID string, req *CompleteAgentRunRequest, userID string) ([]byte, error) {
	resp, err := s.CompleteAgentRun(agentRunID, req, userID)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(resp, "", "  ")
}
