package agentrun

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// TestCompleteAgentRun tests the CompleteAgentRun orchestration method.
func TestCompleteAgentRun(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if err := engine.Sync2(new(models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent_runs table: %v", err)
	}
	if err := engine.Sync2(new(models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent_operations table: %v", err)
	}
	if err := engine.Sync2(new(models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}
	if err := engine.Sync2(new(models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit_logs table: %v", err)
	}

	authService := auth.NewService(engine)
	agentRunService := NewService(engine, authService)
	interactionService := interaction.NewService(engine, nil, nil, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := NewReviewService(engine, agentRunService, authService, evidenceService, interactionService)

	// Create a running agent run
	agentRun := &models.AgentRun{
		ID:         "agent-run-complete-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-complete-1",
		PayloadID:  "payload-complete-1",
		Target:     "https://example.com",
		Title:      "Test Complete Agent Run",
		Status:     models.AgentRunStatusRunning,
		StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := engine.Insert(agentRun); err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	t.Run("Successful completion with JSON format", func(t *testing.T) {
		req := &CompleteAgentRunRequest{
			Format:       "json",
			IncludeAudit: true,
		}

		resp, err := reviewService.CompleteAgentRun("agent-run-complete-1", req, "user-1")
		if err != nil {
			t.Fatalf("Failed to complete agent run: %v", err)
		}

		if resp.AgentRunID != "agent-run-complete-1" {
			t.Errorf("Expected agent_run_id 'agent-run-complete-1', got '%s'", resp.AgentRunID)
		}

		if resp.Status != string(models.AgentRunStatusCompleted) {
			t.Errorf("Expected status 'completed', got '%s'", resp.Status)
		}

		if resp.PackageHash == "" {
			t.Error("Expected package_hash to be set")
		}

		if resp.OperationID == "" {
			t.Error("Expected operation_id to be set")
		}

		if resp.AuditRefID == "" {
			t.Error("Expected audit_ref_id to be set")
		}

		// Verify agent run status is now completed
		updated, err := agentRunService.GetAgentRunByID("agent-run-complete-1")
		if err != nil {
			t.Fatalf("Failed to get updated agent run: %v", err)
		}
		if updated.Status != models.AgentRunStatusCompleted {
			t.Errorf("Expected status 'completed', got '%s'", updated.Status)
		}
		if updated.EndedAt == nil {
			t.Error("Expected ended_at to be set")
		}

		// Verify audit log for completion
		var auditLog models.AuditLog
		has, err := engine.Where("action = ?", "agent_run.completed").Get(&auditLog)
		if err != nil {
			t.Fatalf("Failed to query audit log: %v", err)
		}
		if !has {
			t.Error("Expected audit log entry for agent_run.completed")
		}
	})

	t.Run("Already terminal status", func(t *testing.T) {
		// The agent run was completed in the previous subtest
		req := &CompleteAgentRunRequest{
			Format: "json",
		}

		_, err := reviewService.CompleteAgentRun("agent-run-complete-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for already terminal agent run")
		}
	})

	t.Run("Not found", func(t *testing.T) {
		req := &CompleteAgentRunRequest{
			Format: "json",
		}

		_, err := reviewService.CompleteAgentRun("non-existent", req, "user-1")
		if err == nil {
			t.Error("Expected error for non-existent agent run")
		}
	})

	t.Run("Invalid format", func(t *testing.T) {
		// Create another running agent run
		ar := &models.AgentRun{
			ID:         "agent-run-complete-2",
			AgentID:    "agent-2",
			OperatorID: "user-1",
			Target:     "https://example2.com",
			Title:      "Test Complete Agent Run 2",
			Status:     models.AgentRunStatusRunning,
			StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if _, err := engine.Insert(ar); err != nil {
			t.Fatalf("Failed to insert agent run 2: %v", err)
		}

		req := &CompleteAgentRunRequest{
			Format: "pdf",
		}

		_, err := reviewService.CompleteAgentRun("agent-run-complete-2", req, "user-1")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	t.Run("With decision", func(t *testing.T) {
		// Create another running agent run
		ar := &models.AgentRun{
			ID:         "agent-run-complete-3",
			AgentID:    "agent-3",
			OperatorID: "user-1",
			Target:     "https://example3.com",
			Title:      "Test Complete Agent Run 3",
			Status:     models.AgentRunStatusRunning,
			StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if _, err := engine.Insert(ar); err != nil {
			t.Fatalf("Failed to insert agent run 3: %v", err)
		}

		req := &CompleteAgentRunRequest{
			Format:         "markdown",
			Decision:       "accepted",
			DecisionReason: "Evidence confirmed vulnerability",
			IncludeAudit:   true,
		}

		resp, err := reviewService.CompleteAgentRun("agent-run-complete-3", req, "user-1")
		if err != nil {
			t.Fatalf("Failed to complete agent run with decision: %v", err)
		}

		if resp.Decision != "accepted" {
			t.Errorf("Expected decision 'accepted', got '%s'", resp.Decision)
		}

		// Verify review decision operation was created
		var op models.AgentOperation
		has, err := engine.Where("agent_run_i_d = ? AND action = ?", "agent-run-complete-3", "review_decision.accepted").Get(&op)
		if err != nil || !has {
			t.Error("Expected review_decision.accepted operation to be created")
		}
	})

	t.Run("Invalid decision", func(t *testing.T) {
		// Create another running agent run
		ar := &models.AgentRun{
			ID:         "agent-run-complete-4",
			AgentID:    "agent-4",
			OperatorID: "user-1",
			Target:     "https://example4.com",
			Title:      "Test Complete Agent Run 4",
			Status:     models.AgentRunStatusRunning,
			StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if _, err := engine.Insert(ar); err != nil {
			t.Fatalf("Failed to insert agent run 4: %v", err)
		}

		req := &CompleteAgentRunRequest{
			Format:   "json",
			Decision: "invalid_decision",
		}

		_, err := reviewService.CompleteAgentRun("agent-run-complete-4", req, "user-1")
		if err == nil {
			t.Error("Expected error for invalid decision")
		}
	})
}

// TestCompleteAgentRunJSON tests the CompleteAgentRunJSON convenience method.
func TestCompleteAgentRunJSON(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	if err := engine.Sync2(new(models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent_runs table: %v", err)
	}
	if err := engine.Sync2(new(models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent_operations table: %v", err)
	}
	if err := engine.Sync2(new(models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}
	if err := engine.Sync2(new(models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit_logs table: %v", err)
	}

	authService := auth.NewService(engine)
	agentRunService := NewService(engine, authService)
	interactionService := interaction.NewService(engine, nil, nil, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := NewReviewService(engine, agentRunService, authService, evidenceService, interactionService)

	ar := &models.AgentRun{
		ID:         "agent-run-json-1",
		AgentID:    "agent-json",
		OperatorID: "user-1",
		Target:     "https://json-example.com",
		Title:      "Test JSON Complete",
		Status:     models.AgentRunStatusRunning,
		StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := engine.Insert(ar); err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	req := &CompleteAgentRunRequest{
		Format: "json",
	}

	jsonBytes, err := reviewService.CompleteAgentRunJSON("agent-run-json-1", req, "user-1")
	if err != nil {
		t.Fatalf("Failed to complete agent run JSON: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("Expected non-empty JSON bytes")
	}

	var result map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	if result["agent_run_id"] != "agent-run-json-1" {
		t.Errorf("Expected agent_run_id 'agent-run-json-1', got '%v'", result["agent_run_id"])
	}

	if result["status"] != "completed" {
		t.Errorf("Expected status 'completed', got '%v'", result["status"])
	}
}
