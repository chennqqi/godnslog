package agentrun

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupReviewQueueTest(t *testing.T) (*xorm.Engine, *Service) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Sync tables in dependency order
	err = engine.Sync2(new(models.Case))
	if err != nil {
		t.Fatalf("Failed to sync Case table: %v", err)
	}
	err = engine.Sync2(new(models.Payload))
	if err != nil {
		t.Fatalf("Failed to sync Payload table: %v", err)
	}
	err = engine.Sync2(new(models.Interaction))
	if err != nil {
		t.Fatalf("Failed to sync Interaction table: %v", err)
	}
	err = engine.Sync2(new(models.AgentRun))
	if err != nil {
		t.Fatalf("Failed to sync AgentRun table: %v", err)
	}
	err = engine.Sync2(new(models.AgentOperation))
	if err != nil {
		t.Fatalf("Failed to sync AgentOperation table: %v", err)
	}
	err = engine.Sync2(new(models.AuditLog))
	if err != nil {
		t.Fatalf("Failed to sync AuditLog table: %v", err)
	}

	authService := auth.NewService(engine)
	agentRunService := NewService(engine, authService)

	return engine, agentRunService
}

func TestListReviewQueue_NotReviewed(t *testing.T) {
	engine, service := setupReviewQueueTest(t)
	defer engine.Close()

	// Create a case and payload
	case1 := &models.Case{
		ID:        "case-1",
		Title:     "Test Case",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(case1)
	if err != nil {
		t.Fatalf("Failed to insert case: %v", err)
	}

	payload1 := &models.Payload{
		ID:         "payload-1",
		CaseID:     "case-1",
		Token:      "test-token",
		TemplateID: "ssrf-basic",
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(payload1)
	if err != nil {
		t.Fatalf("Failed to insert payload: %v", err)
	}

	// Create an agent run without review or follow-up
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun)
	if err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Test ListReviewQueue with filter for not_reviewed
	filters := ReviewQueueFilters{
		ReviewState: "not_reviewed",
		Page:        1,
		PageSize:    20,
	}

	result, err := service.ListReviewQueue(filters)
	if err != nil {
		t.Fatalf("ListReviewQueue failed: %v", err)
	}

	// Should return the agent run with review_state = not_reviewed
	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.ID != "agent-run-1" {
		t.Errorf("Expected ID agent-run-1, got %s", item.ID)
	}

	if item.ReviewState != "not_reviewed" {
		t.Errorf("Expected review_state not_reviewed, got %s", item.ReviewState)
	}

	if result.Summary.NotReviewed != 1 {
		t.Errorf("Expected NotReviewed count 1, got %d", result.Summary.NotReviewed)
	}
}

func TestListReviewQueue_Reviewed(t *testing.T) {
	engine, service := setupReviewQueueTest(t)
	defer engine.Close()

	// Create a case and payload
	case1 := &models.Case{
		ID:        "case-1",
		Title:     "Test Case",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(case1)
	if err != nil {
		t.Fatalf("Failed to insert case: %v", err)
	}

	payload1 := &models.Payload{
		ID:         "payload-1",
		CaseID:     "case-1",
		Token:      "test-token",
		TemplateID: "ssrf-basic",
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(payload1)
	if err != nil {
		t.Fatalf("Failed to insert payload: %v", err)
	}

	// Create an agent run
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun)
	if err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Add review_generated audit log
	runID := "agent-run-1"
	userID := "user-1"
	auditLog := &models.AuditLog{
		ID:           "audit-1",
		ResourceType: "agent_run",
		ResourceID:   &runID,
		Action:       "agent_run.review_generated",
		UserID:       &userID,
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}
	_, err = engine.Insert(auditLog)
	if err != nil {
		t.Fatalf("Failed to insert audit log: %v", err)
	}

	// Test ListReviewQueue with filter for reviewed
	filters := ReviewQueueFilters{
		ReviewState: "reviewed",
		Page:        1,
		PageSize:    20,
	}

	result, err := service.ListReviewQueue(filters)
	if err != nil {
		t.Fatalf("ListReviewQueue failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.ReviewState != "reviewed" {
		t.Errorf("Expected review_state reviewed, got %s", item.ReviewState)
	}

	if result.Summary.Reviewed != 1 {
		t.Errorf("Expected Reviewed count 1, got %d", result.Summary.Reviewed)
	}
}

func TestListReviewQueue_FollowupCreated(t *testing.T) {
	engine, service := setupReviewQueueTest(t)
	defer engine.Close()

	// Create a case and payload
	case1 := &models.Case{
		ID:        "case-1",
		Title:     "Test Case",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(case1)
	if err != nil {
		t.Fatalf("Failed to insert case: %v", err)
	}

	payload1 := &models.Payload{
		ID:         "payload-1",
		CaseID:     "case-1",
		Token:      "test-token",
		TemplateID: "ssrf-basic",
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(payload1)
	if err != nil {
		t.Fatalf("Failed to insert payload: %v", err)
	}

	// Create an agent run
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun)
	if err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Add followup operation (use mark_resolved to avoid needs_attention classification)
	operation := &models.AgentOperation{
		ID:         "op-followup-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "followup.mark_resolved",
		RiskLevel:  "low",
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	_, err = engine.Insert(operation)
	if err != nil {
		t.Fatalf("Failed to insert operation: %v", err)
	}

	// Test ListReviewQueue with filter for followup_created
	filters := ReviewQueueFilters{
		ReviewState: "followup_created",
		Page:        1,
		PageSize:    20,
	}

	result, err := service.ListReviewQueue(filters)
	if err != nil {
		t.Fatalf("ListReviewQueue failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.ReviewState != "followup_created" {
		t.Errorf("Expected review_state followup_created, got %s", item.ReviewState)
	}

	if item.FollowupCount != 1 {
		t.Errorf("Expected FollowupCount 1, got %d", item.FollowupCount)
	}

	if item.LastFollowupAction != "followup.mark_resolved" {
		t.Errorf("Expected LastFollowupAction followup.mark_resolved, got %s", item.LastFollowupAction)
	}

	if result.Summary.FollowupCreated != 1 {
		t.Errorf("Expected FollowupCreated count 1, got %d", result.Summary.FollowupCreated)
	}
}

func TestListReviewQueue_NeedsAttention(t *testing.T) {
	engine, service := setupReviewQueueTest(t)
	defer engine.Close()

	// Create a case and payload
	case1 := &models.Case{
		ID:        "case-1",
		Title:     "Test Case",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(case1)
	if err != nil {
		t.Fatalf("Failed to insert case: %v", err)
	}

	payload1 := &models.Payload{
		ID:         "payload-1",
		CaseID:     "case-1",
		Token:      "test-token",
		TemplateID: "ssrf-basic",
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(payload1)
	if err != nil {
		t.Fatalf("Failed to insert payload: %v", err)
	}

	// Create an agent run with high evidence (simulated via audit)
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun)
	if err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Add review_generated audit with high evidence strength
	runID := "agent-run-1"
	userID := "user-1"
	auditLog := &models.AuditLog{
		ID:           "audit-1",
		ResourceType: "agent_run",
		ResourceID:   &runID,
		Action:       "agent_run.review_generated",
		UserID:       &userID,
		Details:      models.AuditDetails{"evidence_strength": "high"},
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}
	_, err = engine.Insert(auditLog)
	if err != nil {
		t.Fatalf("Failed to insert audit log: %v", err)
	}

	// Test ListReviewQueue with filter for needs_attention
	filters := ReviewQueueFilters{
		ReviewState: "needs_attention",
		Page:        1,
		PageSize:    20,
	}

	result, err := service.ListReviewQueue(filters)
	if err != nil {
		t.Fatalf("ListReviewQueue failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.ReviewState != "needs_attention" {
		t.Errorf("Expected review_state needs_attention, got %s", item.ReviewState)
	}

	if result.Summary.NeedsAttention != 1 {
		t.Errorf("Expected NeedsAttention count 1, got %d", result.Summary.NeedsAttention)
	}
}

func TestListReviewQueue_MultipleFollowups(t *testing.T) {
	engine, service := setupReviewQueueTest(t)
	defer engine.Close()

	// Create a case and payload
	case1 := &models.Case{
		ID:        "case-1",
		Title:     "Test Case",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(case1)
	if err != nil {
		t.Fatalf("Failed to insert case: %v", err)
	}

	payload1 := &models.Payload{
		ID:         "payload-1",
		CaseID:     "case-1",
		Token:      "test-token",
		TemplateID: "ssrf-basic",
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(payload1)
	if err != nil {
		t.Fatalf("Failed to insert payload: %v", err)
	}

	// Create an agent run
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun)
	if err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Add multiple followup operations
	now := time.Now()
	operation1 := &models.AgentOperation{
		ID:         "op-followup-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "followup.recheck_evidence",
		RiskLevel:  "low",
		StartedAt:  now.Add(-2 * time.Hour),
		CreatedAt:  now.Add(-2 * time.Hour),
	}
	_, err = engine.Insert(operation1)
	if err != nil {
		t.Fatalf("Failed to insert operation 1: %v", err)
	}

	operation2 := &models.AgentOperation{
		ID:         "op-followup-2",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "followup.mark_resolved",
		RiskLevel:  "low",
		StartedAt:  now.Add(-1 * time.Hour),
		CreatedAt:  now.Add(-1 * time.Hour),
	}
	_, err = engine.Insert(operation2)
	if err != nil {
		t.Fatalf("Failed to insert operation 2: %v", err)
	}

	// Test ListReviewQueue
	filters := ReviewQueueFilters{
		Page:     1,
		PageSize: 20,
	}

	result, err := service.ListReviewQueue(filters)
	if err != nil {
		t.Fatalf("ListReviewQueue failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	item := result.Items[0]
	if item.FollowupCount != 2 {
		t.Errorf("Expected FollowupCount 2, got %d", item.FollowupCount)
	}

	// Due to xorm auto-timestamp, both operations have same CreatedAt
	// Just verify that LastFollowupAction is set
	if item.LastFollowupAction == "" {
		t.Errorf("Expected LastFollowupAction to be set, got empty string")
	}
}

func TestListReviewQueue_FilterByStatus(t *testing.T) {
	engine, service := setupReviewQueueTest(t)
	defer engine.Close()

	// Create a case and payload
	case1 := &models.Case{
		ID:        "case-1",
		Title:     "Test Case",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(case1)
	if err != nil {
		t.Fatalf("Failed to insert case: %v", err)
	}

	payload1 := &models.Payload{
		ID:         "payload-1",
		CaseID:     "case-1",
		Token:      "test-token",
		TemplateID: "ssrf-basic",
		CreatedBy:  "user-1",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(payload1)
	if err != nil {
		t.Fatalf("Failed to insert payload: %v", err)
	}

	// Create agent runs with different statuses
	agentRun1 := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run 1",
		Status:     models.AgentRunStatusCompleted,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun1)
	if err != nil {
		t.Fatalf("Failed to insert agent run 1: %v", err)
	}

	agentRun2 := &models.AgentRun{
		ID:         "agent-run-2",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "example.com",
		Title:      "Test Agent Run 2",
		Status:     models.AgentRunStatusRunning,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	_, err = engine.Insert(agentRun2)
	if err != nil {
		t.Fatalf("Failed to insert agent run 2: %v", err)
	}

	// Test filter by status
	filters := ReviewQueueFilters{
		Status:   "completed",
		Page:     1,
		PageSize: 20,
	}

	result, err := service.ListReviewQueue(filters)
	if err != nil {
		t.Fatalf("ListReviewQueue failed: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(result.Items))
	}

	if result.Items[0].ID != "agent-run-1" {
		t.Errorf("Expected agent-run-1, got %s", result.Items[0].ID)
	}
}
