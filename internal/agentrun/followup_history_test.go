package agentrun

import (
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/models"
)

func setupFollowupHistoryTest(t *testing.T) (*xorm.Engine, *Service) {
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

func TestListFollowupHistory_Empty(t *testing.T) {
	engine, service := setupFollowupHistoryTest(t)
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

	// Test ListFollowupHistory
	history, err := service.ListFollowupHistory("agent-run-1")
	if err != nil {
		t.Fatalf("ListFollowupHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history, got %d items", len(history))
	}
}

func TestListFollowupHistory_WithFollowup(t *testing.T) {
	engine, service := setupFollowupHistoryTest(t)
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

	// Add followup operation with request containing reason and review_packet_id
	operation := &models.AgentOperation{
		ID:         "op-followup-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "followup.mark_resolved",
		RiskLevel:  "low",
		Request:    `{"reason": "Issue resolved", "review_packet_id": "packet-1"}`,
		CreatedAt:  time.Now(),
	}
	_, err = engine.Insert(operation)
	if err != nil {
		t.Fatalf("Failed to insert operation: %v", err)
	}

	// Add audit log for reference
	runID := "agent-run-1"
	userID := "user-1"
	auditLog := &models.AuditLog{
		ID:           "audit-1",
		ResourceType: "agent_run",
		ResourceID:   &runID,
		Action:       "followup_created",
		UserID:       &userID,
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}
	_, err = engine.Insert(auditLog)
	if err != nil {
		t.Fatalf("Failed to insert audit log: %v", err)
	}

	// Test ListFollowupHistory
	history, err := service.ListFollowupHistory("agent-run-1")
	if err != nil {
		t.Fatalf("ListFollowupHistory failed: %v", err)
	}

	if len(history) != 1 {
		t.Fatalf("Expected 1 history item, got %d", len(history))
	}

	item := history[0]
	if item.OperationID != "op-followup-1" {
		t.Errorf("Expected OperationID op-followup-1, got %s", item.OperationID)
	}
	if item.ActionType != "mark_resolved" {
		t.Errorf("Expected ActionType mark_resolved, got %s", item.ActionType)
	}
	if item.Reason != "Issue resolved" {
		t.Errorf("Expected Reason 'Issue resolved', got %s", item.Reason)
	}
	if item.ReviewPacketID != "packet-1" {
		t.Errorf("Expected ReviewPacketID packet-1, got %s", item.ReviewPacketID)
	}
	if item.AuditRefID != "audit-1" {
		t.Errorf("Expected AuditRefID audit-1, got %s", item.AuditRefID)
	}
}

func TestListFollowupHistory_MultipleFollowups(t *testing.T) {
	engine, service := setupFollowupHistoryTest(t)
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
	operation1 := &models.AgentOperation{
		ID:         "op-followup-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "followup.recheck_evidence",
		RiskLevel:  "high",
		Request:    `{"reason": "Need to recheck"}`,
		CreatedAt:  time.Now(),
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
		Request:    `{"reason": "Resolved"}`,
		CreatedAt:  time.Now(),
	}
	_, err = engine.Insert(operation2)
	if err != nil {
		t.Fatalf("Failed to insert operation 2: %v", err)
	}

	// Add a non-followup operation (should be filtered out)
	operation3 := &models.AgentOperation{
		ID:         "op-scan-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "scan.dns",
		RiskLevel:  "medium",
		CreatedAt:  time.Now(),
	}
	_, err = engine.Insert(operation3)
	if err != nil {
		t.Fatalf("Failed to insert operation 3: %v", err)
	}

	// Test ListFollowupHistory
	history, err := service.ListFollowupHistory("agent-run-1")
	if err != nil {
		t.Fatalf("ListFollowupHistory failed: %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("Expected 2 history items, got %d", len(history))
	}
}

func TestListFollowupHistory_AgentRunNotFound(t *testing.T) {
	engine, service := setupFollowupHistoryTest(t)
	defer engine.Close()

	// Test with non-existent agent run
	history, err := service.ListFollowupHistory("non-existent")
	if err != nil {
		t.Fatalf("ListFollowupHistory failed: %v", err)
	}

	if len(history) != 0 {
		t.Errorf("Expected empty history for non-existent agent run, got %d items", len(history))
	}
}
