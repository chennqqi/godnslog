package agentrun

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// Service provides agent run business logic
type Service struct {
	engine      *xorm.Engine
	authService *auth.Service
}

// NewService creates a new agent run service
func NewService(engine *xorm.Engine, authService *auth.Service) *Service {
	return &Service{
		engine:      engine,
		authService: authService,
	}
}

// CreateAgentRun creates a new agent run
func (s *Service) CreateAgentRun(req *models.AgentRunCreateRequest, userID string) (*models.AgentRun, error) {
	agentRun := &models.AgentRun{
		ID:         generateID(),
		AgentID:    req.AgentID,
		OperatorID: req.OperatorID,
		CaseID:     req.CaseID,
		PayloadID:  req.PayloadID,
		Target:     req.Target,
		Title:      req.Title,
		Status:     models.AgentRunStatusCreated,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if _, err := s.engine.Insert(agentRun); err != nil {
		return nil, err
	}

	// Create audit log for agent_run.created
	userIDPtr := &userID
	resourceIDPtr := &agentRun.ID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.created",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"agent_id":    agentRun.AgentID,
			"operator_id": agentRun.OperatorID,
			"target":      agentRun.Target,
			"title":       agentRun.Title,
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return agentRun, nil
}

// GetAgentRunByID retrieves an agent run by its ID
func (s *Service) GetAgentRunByID(id string) (*models.AgentRun, error) {
	var agentRun models.AgentRun
	has, err := s.engine.ID(id).Get(&agentRun)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, nil
	}
	return &agentRun, nil
}

// GetAgentRunDetail retrieves an agent run with computed fields
func (s *Service) GetAgentRunDetail(id string, baseURL string) (*models.AgentRunDetail, error) {
	agentRun, err := s.GetAgentRunByID(id)
	if err != nil {
		return nil, err
	}
	if agentRun == nil {
		return nil, nil
	}

	detail := &models.AgentRunDetail{
		AgentRun: *agentRun,
	}

	// Count interactions
	interactionCount, err := s.engine.Where("case_id = ?", agentRun.CaseID).Count(&models.Interaction{})
	if err == nil {
		detail.InteractionCount = int(interactionCount)
	}

	// Get last interaction
	var lastInteraction models.Interaction
	has, err := s.engine.Where("case_id = ?", agentRun.CaseID).OrderBy("created_at DESC").Limit(1).Get(&lastInteraction)
	if err == nil && has {
		detail.LastInteractionAt = &lastInteraction.CreatedAt
	}

	// Load operations
	var operations []models.AgentOperation
	err = s.engine.Where("agent_run_i_d = ?", id).OrderBy("started_at ASC").Find(&operations)
	if err == nil {
		detail.Operations = operations
	}

	// Generate URLs
	if agentRun.CaseID != "" {
		detail.CaseURL = fmt.Sprintf("%s/dashboard/cases/%s", baseURL, agentRun.CaseID)
	}
	if agentRun.PayloadID != "" {
		detail.PayloadURL = fmt.Sprintf("%s/dashboard/payloads/%s", baseURL, agentRun.PayloadID)
		detail.InteractionsURL = fmt.Sprintf("%s/api/v2/interactions?payload_id=%s", baseURL, agentRun.PayloadID)
		detail.EvidenceURL = fmt.Sprintf("%s/dashboard/evidence?payload_id=%s", baseURL, agentRun.PayloadID)
	}

	return detail, nil
}

// ListAgentRuns lists agent runs with filters
func (s *Service) ListAgentRuns(req *models.AgentRunListRequest) (*models.AgentRunListResponse, error) {
	session := s.engine.NewSession()

	if req.AgentID != "" {
		session = session.Where("agent_i_d = ?", req.AgentID)
	}
	if req.CaseID != "" {
		session = session.Where("case_i_d = ?", req.CaseID)
	}
	if req.PayloadID != "" {
		session = session.Where("payload_i_d = ?", req.PayloadID)
	}
	if req.Status != "" {
		session = session.Where("status = ?", req.Status)
	}

	// Count total
	total, err := session.Count(&models.AgentRun{})
	if err != nil {
		return nil, err
	}

	// Pagination
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	offset := (page - 1) * pageSize

	// Query items
	var agentRuns []models.AgentRun
	err = session.OrderBy("created_at DESC").Limit(pageSize, offset).Find(&agentRuns)
	if err != nil {
		return nil, err
	}

	// Convert to details
	items := make([]models.AgentRunDetail, len(agentRuns))
	for i, run := range agentRuns {
		detail, err := s.GetAgentRunDetail(run.ID, "")
		if err != nil {
			return nil, err
		}
		items[i] = *detail
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.AgentRunListResponse{
		Items:      items,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateAgentRunStatus updates the status of an agent run
func (s *Service) UpdateAgentRunStatus(id string, req *models.AgentRunUpdateStatusRequest, userID string) error {
	// Get existing run
	existingRun, err := s.GetAgentRunByID(id)
	if err != nil {
		return err
	}
	if existingRun == nil {
		return errors.New("agent run not found")
	}

	// Validate status transition
	if !isValidAgentRunStatusTransition(existingRun.Status, req.Status) {
		return fmt.Errorf("invalid status transition from %s to %s", existingRun.Status, req.Status)
	}

	// Save old status for audit log before updating
	oldStatus := existingRun.Status

	// Update status
	existingRun.Status = req.Status
	existingRun.UpdatedAt = time.Now()

	// Update started_at if transitioning to running
	if req.Status == models.AgentRunStatusRunning && existingRun.StartedAt == nil {
		now := time.Now()
		existingRun.StartedAt = &now
	}

	// Update ended_at if transitioning to terminal state
	if isTerminalStatus(req.Status) && existingRun.EndedAt == nil {
		now := time.Now()
		existingRun.EndedAt = &now
	}

	_, err = s.engine.ID(id).Cols("status", "started_at", "ended_at", "updated_at").Update(existingRun)
	if err != nil {
		return err
	}

	// Create audit log for agent_run.status_updated
	userIDPtr := &userID
	resourceIDPtr := &id
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.status_updated",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"from_status": string(oldStatus),
			"to_status":   string(req.Status),
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// AppendAgentOperation appends an operation to an agent run
func (s *Service) AppendAgentOperation(agentRunID string, req *models.AgentOperationCreateRequest, userID string) error {
	// Serialize request and result
	requestJSON, _ := json.Marshal(req.Request)
	resultJSON, _ := json.Marshal(req.Result)

	operation := &models.AgentOperation{
		ID:         generateID(),
		AgentRunID: agentRunID,
		AgentID:    "", // Will be filled from agent run
		Action:     req.Action,
		RiskLevel:  req.RiskLevel,
		Request:    string(requestJSON),
		Result:     string(resultJSON),
		Error:      req.Error,
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}

	// Get agent run to fill agent_id
	agentRun, err := s.GetAgentRunByID(agentRunID)
	if err != nil {
		return err
	}
	if agentRun == nil {
		return errors.New("agent run not found")
	}
	operation.AgentID = agentRun.AgentID

	if _, err := s.engine.Insert(operation); err != nil {
		return err
	}

	// Create audit log for agent operation
	userIDPtr := &userID
	resourceIDPtr := &agentRunID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       fmt.Sprintf("agent_operation.%s", req.Action),
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"agent_run_id": agentRunID,
			"action":       req.Action,
			"risk_level":   req.RiskLevel,
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// CreateFollowupAction creates a follow-up action for an agent run
func (s *Service) CreateFollowupAction(agentRunID string, req *models.AgentRunFollowupRequest, userID string) (*models.AgentRunFollowupResponse, error) {
	agentRun, err := s.GetAgentRunByID(agentRunID)
	if err != nil {
		return nil, err
	}
	if agentRun == nil {
		return nil, errors.New("agent run not found")
	}
	if req == nil {
		return nil, errors.New("request is required")
	}
	if !models.IsAllowedAgentRunFollowupAction(req.ActionType) {
		return nil, fmt.Errorf("invalid followup action type: %s", req.ActionType)
	}
	reason := strings.TrimSpace(req.Reason)
	if reason == "" {
		return nil, errors.New("reason is required")
	}
	if len(reason) > 500 {
		return nil, errors.New("reason must be 500 characters or less")
	}

	result := map[string]interface{}{
		"success":             true,
		"source_agent_run_id": agentRun.ID,
		"action_type":         req.ActionType,
		"reason":              reason,
		"review_packet_id":    req.ReviewPacketID,
		"case_id":             agentRun.CaseID,
		"payload_id":          agentRun.PayloadID,
	}

	now := time.Now()
	opReq := &models.AgentOperationCreateRequest{
		Action:    "followup." + req.ActionType,
		RiskLevel: "low",
		Request:   map[string]interface{}{"reason": reason, "review_packet_id": req.ReviewPacketID},
		Result:    result,
	}

	if err := s.AppendAgentOperation(agentRun.ID, opReq, userID); err != nil {
		return nil, err
	}

	// Get the created operation
	var operation models.AgentOperation
	has, err := s.engine.Where("agent_run_i_d = ? AND action = ?", agentRun.ID, "followup."+req.ActionType).OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve created operation: has=%v err=%v", has, err)
	}

	userIDPtr := &userID
	resourceIDPtr := &agentRun.ID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.followup_created",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"agent_run_id":     agentRun.ID,
			"operation_id":     operation.ID,
			"action_type":      req.ActionType,
			"reason":           reason,
			"review_packet_id": req.ReviewPacketID,
			"case_id":          agentRun.CaseID,
			"payload_id":       agentRun.PayloadID,
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create followup audit log: %w", err)
	}

	return &models.AgentRunFollowupResponse{
		AgentRunID:     agentRun.ID,
		OperationID:    operation.ID,
		ActionType:     req.ActionType,
		Reason:         reason,
		ReviewPacketID: req.ReviewPacketID,
		Operation:      operation,
		CreatedAt:      now,
	}, nil
}

// RecordReviewDecision records a review decision for an agent run
func (s *Service) RecordReviewDecision(agentRunID string, req *models.AgentRunReviewDecisionRequest, userID string) (*models.AgentRunReviewDecisionResponse, error) {
	// Validate decision
	validDecisions := map[string]bool{
		"accepted":              true,
		"false_positive":        true,
		"needs_manual_followup": true,
		"insufficient_evidence": true,
	}
	if !validDecisions[req.Decision] {
		return nil, errors.New("invalid decision: must be one of accepted, false_positive, needs_manual_followup, insufficient_evidence")
	}

	// Validate reason length
	if len(req.Reason) > 500 {
		return nil, errors.New("reason too long: maximum 500 characters")
	}

	// Get agent run
	var agentRun models.AgentRun
	has, err := s.engine.ID(agentRunID).Get(&agentRun)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent run: %w", err)
	}
	if !has {
		return nil, errors.New("agent run not found")
	}

	// Validate review_packet_id if provided
	if req.ReviewPacketID != "" && req.ReviewPacketID != agentRunID {
		return nil, errors.New("review_packet_id must match the current agent run")
	}

	now := time.Now()

	// Build operation result
	result := map[string]interface{}{
		"decision":         req.Decision,
		"reason":           req.Reason,
		"review_packet_id": req.ReviewPacketID,
		"evidence_id":      req.EvidenceID,
		"audit_action":     "agent_run.review_decision_recorded",
	}

	// Create operation
	opReq := &models.AgentOperationCreateRequest{
		Action:    "review_decision." + req.Decision,
		RiskLevel: "low",
		Request: map[string]interface{}{
			"decision":         req.Decision,
			"reason":           req.Reason,
			"review_packet_id": req.ReviewPacketID,
			"evidence_id":      req.EvidenceID,
		},
		Result: result,
	}

	if err := s.AppendAgentOperation(agentRunID, opReq, userID); err != nil {
		return nil, fmt.Errorf("failed to append review decision operation: %w", err)
	}

	// Get the created operation
	var operation models.AgentOperation
	has, err = s.engine.Where("agent_run_i_d = ? AND action = ?", agentRunID, "review_decision."+req.Decision).OrderBy("created_at DESC").Limit(1).Get(&operation)
	if err != nil || !has {
		return nil, fmt.Errorf("failed to retrieve created operation: has=%v err=%v", has, err)
	}

	// Create audit log
	userIDPtr := &userID
	resourceIDPtr := &agentRunID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "agent_run.review_decision_recorded",
		ResourceType: "agent_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"decision":         req.Decision,
			"reason":           req.Reason,
			"review_packet_id": req.ReviewPacketID,
			"operation_id":     operation.ID,
			"case_id":          agentRun.CaseID,
			"payload_id":       agentRun.PayloadID,
		},
		Timestamp: now,
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		return nil, fmt.Errorf("failed to create review decision audit log: %w", err)
	}

	return &models.AgentRunReviewDecisionResponse{
		AgentRunID:     agentRunID,
		OperationID:    operation.ID,
		Decision:       req.Decision,
		ReviewPacketID: req.ReviewPacketID,
		AuditRefID:     auditLog.ID,
		Operation:      &operation,
		Audit: map[string]interface{}{
			"id":            auditLog.ID,
			"action":        auditLog.Action,
			"resource_type": auditLog.ResourceType,
			"resource_id":   auditLog.ResourceID,
			"details":       auditLog.Details,
			"timestamp":     auditLog.Timestamp,
		},
	}, nil
}

// isValidAgentRunStatusTransition validates if a status transition is allowed
func isValidAgentRunStatusTransition(from, to models.AgentRunStatus) bool {
	validTransitions := map[models.AgentRunStatus][]models.AgentRunStatus{
		models.AgentRunStatusCreated:   {models.AgentRunStatusRunning, models.AgentRunStatusCancelled},
		models.AgentRunStatusRunning:   {models.AgentRunStatusWaiting, models.AgentRunStatusCompleted, models.AgentRunStatusFailed, models.AgentRunStatusCancelled, models.AgentRunStatusTimedOut},
		models.AgentRunStatusWaiting:   {models.AgentRunStatusRunning, models.AgentRunStatusCompleted, models.AgentRunStatusFailed, models.AgentRunStatusTimedOut},
		models.AgentRunStatusCompleted: {},
		models.AgentRunStatusFailed:    {},
		models.AgentRunStatusCancelled: {},
		models.AgentRunStatusTimedOut:  {},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, validTo := range allowed {
		if validTo == to {
			return true
		}
	}

	return false
}

// isTerminalStatus checks if a status is terminal
func isTerminalStatus(status models.AgentRunStatus) bool {
	return status == models.AgentRunStatusCompleted ||
		status == models.AgentRunStatusFailed ||
		status == models.AgentRunStatusCancelled ||
		status == models.AgentRunStatusTimedOut
}

// generateID generates a unique ID using base32 encoding
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}
