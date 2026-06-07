package agentrun

import (
	"encoding/json"
	"sort"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// ReviewQueueFilters represents filters for the review queue
type ReviewQueueFilters struct {
	ReviewState      string `json:"review_state"`
	Status           string `json:"status"`
	EvidenceStrength string `json:"evidence_strength"`
	AgentID          string `json:"agent_id"`
	CaseID           string `json:"case_id"`
	PayloadID        string `json:"payload_id"`
	Page             int    `json:"page"`
	PageSize         int    `json:"page_size"`
}

// ListReviewQueue returns a paginated list of agent runs for the review queue
func (s *Service) ListReviewQueue(filters ReviewQueueFilters) (*models.AgentRunReviewQueueResponse, error) {
	// Set default pagination
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PageSize <= 0 {
		filters.PageSize = 20
	}
	if filters.PageSize > 100 {
		filters.PageSize = 100
	}

	// Build base query for agent runs
	query := s.engine.NewSession()
	defer query.Close()

	// Apply database-level filters
	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}
	if filters.AgentID != "" {
		query = query.Where("agent_id = ?", filters.AgentID)
	}
	if filters.CaseID != "" {
		query = query.Where("case_id = ?", filters.CaseID)
	}
	if filters.PayloadID != "" {
		query = query.Where("payload_id = ?", filters.PayloadID)
	}

	// Get all agent runs (we'll filter in memory for review_state and evidence_strength)
	var agentRuns []models.AgentRun
	err := query.OrderBy("created_at DESC").Find(&agentRuns)
	if err != nil {
		return nil, err
	}

	// Build review queue items with in-memory filtering
	allItems := make([]models.AgentRunReviewQueueItem, 0, len(agentRuns))
	for _, run := range agentRuns {
		item, err := s.buildReviewQueueItem(&run)
		if err != nil {
			return nil, err
		}

		// Apply review_state filter
		if filters.ReviewState != "" && item.ReviewState != filters.ReviewState {
			continue
		}

		// Apply evidence_strength filter
		if filters.EvidenceStrength != "" && item.EvidenceStrength != filters.EvidenceStrength {
			continue
		}

		allItems = append(allItems, *item)
	}

	// Calculate pagination after filtering
	total := int64(len(allItems))
	totalPages := int(total) / filters.PageSize
	if int(total)%filters.PageSize != 0 {
		totalPages++
	}

	// Apply pagination
	start := (filters.Page - 1) * filters.PageSize
	end := start + filters.PageSize
	if start >= len(allItems) {
		end = start
	} else if end > len(allItems) {
		end = len(allItems)
	}

	items := allItems[start:end]

	// Calculate summary from all filtered items (not just current page)
	summary := models.AgentRunReviewQueueSummary{
		Total: total,
	}
	for _, item := range allItems {
		switch item.ReviewState {
		case "not_reviewed":
			summary.NotReviewed++
		case "reviewed":
			summary.Reviewed++
		case "followup_created":
			summary.FollowupCreated++
		case "needs_attention":
			summary.NeedsAttention++
		}
	}

	return &models.AgentRunReviewQueueResponse{
		Items:      items,
		Total:      total,
		Page:       filters.Page,
		PageSize:   filters.PageSize,
		TotalPages: totalPages,
		Summary:    summary,
	}, nil
}

// buildReviewQueueItem builds a review queue item from an agent run
func (s *Service) buildReviewQueueItem(run *models.AgentRun) (*models.AgentRunReviewQueueItem, error) {
	item := &models.AgentRunReviewQueueItem{
		ID:          run.ID,
		AgentID:     run.AgentID,
		OperatorID:  run.OperatorID,
		CaseID:      run.CaseID,
		PayloadID:   run.PayloadID,
		Target:      run.Target,
		Status:      string(run.Status),
		CreatedAt:   run.CreatedAt,
		UpdatedAt:   run.UpdatedAt,
		DetailURL:   "/dashboard/agent-runs/" + run.ID,
		EvidenceURL: "/dashboard/evidence?case_id=" + run.CaseID,
	}

	// Get interaction count
	interactionCount, err := s.engine.Table("interactions").Where("case_id = ?", run.CaseID).Count()
	if err != nil {
		interactionCount = 0
	}
	item.InteractionCount = int(interactionCount)

	// Get operation count
	var opCount int64
	opCount, err = s.engine.Table("agent_operations").Count()
	if err != nil {
		opCount = 0
	}
	item.OperationCount = int(opCount)

	// Get all operations and filter in Go
	var allOps []models.AgentOperation
	err = s.engine.Find(&allOps)
	if err != nil {
		allOps = []models.AgentOperation{}
	}

	// Filter for this agent run's follow-up operations
	var followupOps []models.AgentOperation
	for _, op := range allOps {
		if op.AgentRunID == run.ID && len(op.Action) > 9 && op.Action[:9] == "followup." {
			followupOps = append(followupOps, op)
		}
	}

	// Sort by created_at descending (most recent first)
	sort.Slice(followupOps, func(i, j int) bool {
		return followupOps[i].CreatedAt.After(followupOps[j].CreatedAt)
	})

	item.FollowupCount = len(followupOps)

	if len(followupOps) > 0 {
		item.LastFollowupAction = followupOps[0].Action
		item.LastFollowupAt = &followupOps[0].CreatedAt
	}

	// Get review state and evidence strength
	reviewState, evidenceStrength, lastReviewedAt, err := s.deriveReviewState(run, followupOps)
	if err != nil {
		return nil, err
	}
	item.ReviewState = reviewState
	item.EvidenceStrength = evidenceStrength
	item.LastReviewedAt = lastReviewedAt

	return item, nil
}

// ListFollowupHistory returns the follow-up history for an agent run
func (s *Service) ListFollowupHistory(agentRunID string) ([]models.AgentRunFollowupHistoryItem, error) {
	// Get all operations
	var allOps []models.AgentOperation
	err := s.engine.Find(&allOps)
	if err != nil {
		return nil, err
	}

	// Get all audits for linking
	var allAudits []models.AuditLog
	err = s.engine.Find(&allAudits)
	if err != nil {
		return nil, err
	}

	// Build audit ref map
	auditRefMap := make(map[string]string)
	for _, audit := range allAudits {
		if audit.ResourceType == "agent_run" && audit.ResourceID != nil && *audit.ResourceID == agentRunID {
			auditRefMap[audit.Action] = audit.ID
		}
	}

	// Filter and build history items
	history := make([]models.AgentRunFollowupHistoryItem, 0)
	for _, op := range allOps {
		if op.AgentRunID == agentRunID && len(op.Action) > 9 && op.Action[:9] == "followup." {
			item := models.AgentRunFollowupHistoryItem{
				OperationID: op.ID,
				ActionType:  op.Action[9:], // Remove "followup." prefix
				RiskLevel:   op.RiskLevel,
				CreatedAt:   op.CreatedAt,
			}

			// Parse request JSON for reason and review_packet_id
			if op.Request != "" {
				var requestData map[string]interface{}
				if err := json.Unmarshal([]byte(op.Request), &requestData); err == nil {
					if reason, ok := requestData["reason"].(string); ok {
						item.Reason = reason
					}
					if packetID, ok := requestData["review_packet_id"].(string); ok {
						item.ReviewPacketID = packetID
					}
				}
			}

			// Link audit ref
			if auditID, ok := auditRefMap["followup_created"]; ok {
				item.AuditRefID = auditID
			}

			history = append(history, item)
		}
	}

	// Sort by created_at descending
	sort.Slice(history, func(i, j int) bool {
		return history[i].CreatedAt.After(history[j].CreatedAt)
	})

	return history, nil
}

// deriveReviewState derives the review state from existing data
func (s *Service) deriveReviewState(run *models.AgentRun, followupOps []models.AgentOperation) (string, string, *time.Time, error) {
	// Check for follow-up operations
	if len(followupOps) > 0 {
		// Check if latest follow-up is recheck_evidence
		latestAction := followupOps[0].Action
		if latestAction == "followup.recheck_evidence" {
			return "needs_attention", "high", nil, nil
		}
		return "followup_created", "medium", nil, nil
	}

	// Check for review_generated audit
	var allAudits []models.AuditLog
	err := s.engine.Find(&allAudits)
	if err != nil {
		allAudits = []models.AuditLog{}
	}

	for _, audit := range allAudits {
		if audit.ResourceType == "agent_run" && audit.ResourceID != nil && *audit.ResourceID == run.ID && audit.Action == "review_generated" {
			// Extract evidence strength from audit details
			evidenceStrength := "medium"
			if audit.Details != nil {
				if strength, ok := audit.Details["evidence_strength"].(string); ok {
					evidenceStrength = strength
				}
			}

			// Check if needs attention (high evidence without follow-up)
			if evidenceStrength == "high" {
				return "needs_attention", evidenceStrength, &audit.Timestamp, nil
			}

			return "reviewed", evidenceStrength, &audit.Timestamp, nil
		}
	}

	// Check for review export operation
	var allOps []models.AgentOperation
	err = s.engine.Find(&allOps)
	if err != nil {
		allOps = []models.AgentOperation{}
	}

	for _, op := range allOps {
		if op.AgentRunID == run.ID && op.Action == "review.export" {
			return "reviewed", "medium", &op.CreatedAt, nil
		}
	}

	// Default: not reviewed
	return "not_reviewed", "none", nil, nil
}
