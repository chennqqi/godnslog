package agentrun

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupTestEngine(t *testing.T) *xorm.Engine {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}

	// Drop tables first to ensure clean schema
	engine.Exec("DROP TABLE IF EXISTS audit_logs")
	engine.Exec("DROP TABLE IF EXISTS agent_operations")
	engine.Exec("DROP TABLE IF EXISTS agent_runs")
	engine.Exec("DROP TABLE IF EXISTS interactions")
	engine.Exec("DROP TABLE IF EXISTS payloads")
	engine.Exec("DROP TABLE IF EXISTS cases")

	// Sync tables individually to ensure proper schema
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

	// Verify tables exist by checking schema
	tables, err := engine.DBMetas()
	if err != nil {
		t.Fatalf("Failed to get table metadata: %v", err)
	}

	tableNames := make(map[string]bool)
	for _, table := range tables {
		tableNames[table.Name] = true
	}

	if !tableNames["agent_runs"] && !tableNames["agent_run"] {
		t.Fatalf("agent_runs table not found, available tables: %v", tableNames)
	}
	if !tableNames["agent_operations"] && !tableNames["agent_operation"] {
		t.Fatalf("agent_operations table not found, available tables: %v", tableNames)
	}

	return engine
}

func createTestCase(t *testing.T, engine *xorm.Engine, id string) *models.Case {
	caseItem := &models.Case{
		ID:        id,
		Title:     "Test Case",
		Status:    "active",
		CreatedBy: "1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	_, err := engine.Insert(caseItem)
	if err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}
	return caseItem
}

func createTestPayload(t *testing.T, engine *xorm.Engine, caseID, id string) *models.Payload {
	payload := &models.Payload{
		ID:               id,
		CaseID:           caseID,
		Token:            "test-token",
		TemplateRendered: "http://{{.Token}}.example.com",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	_, err := engine.Insert(payload)
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}
	return payload
}

func TestCreateAgentRun(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	req := &models.AgentRunCreateRequest{
		AgentID:    "agent-123",
		OperatorID: "operator-1",
		CaseID:     "case-123",
		PayloadID:  "payload-123",
		Target:     "http://example.com",
		Title:      "Test Agent Run",
	}

	agentRun, err := service.CreateAgentRun(req, "1")
	if err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}
	if agentRun == nil {
		t.Fatal("Expected agent run, got nil")
	}

	// Verify basic fields
	if agentRun.AgentID != req.AgentID {
		t.Errorf("Expected agent ID %s, got %s", req.AgentID, agentRun.AgentID)
	}
	if agentRun.OperatorID != req.OperatorID {
		t.Errorf("Expected operator ID %s, got %s", req.OperatorID, agentRun.OperatorID)
	}
	if agentRun.Status != models.AgentRunStatusCreated {
		t.Errorf("Expected status %s, got %s", models.AgentRunStatusCreated, agentRun.Status)
	}

	// Verify audit log was created for agent_run.created
	var auditLog models.AuditLog
	has, err := engine.Where("action = ? AND resource_type = ?", "agent_run.created", "agent_run").Get(&auditLog)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if !has {
		t.Error("Expected audit log entry for agent_run.created")
	}
	if auditLog.UserID == nil || *auditLog.UserID != "1" {
		t.Error("Expected audit log to have user_id = 1")
	}
	if auditLog.Details == nil {
		t.Error("Expected audit log to have details")
	}
}

func TestGetAgentRunByID(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	req := &models.AgentRunCreateRequest{
		AgentID:    "agent-123",
		OperatorID: "operator-1",
		Target:     "http://example.com",
		Title:      "Test Agent Run",
	}

	created, err := service.CreateAgentRun(req, "1")
	if err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	retrieved, err := service.GetAgentRunByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get agent run: %v", err)
	}
	if retrieved == nil {
		t.Fatal("Expected agent run, got nil")
	}
	if retrieved.ID != created.ID {
		t.Errorf("Expected ID %s, got %s", created.ID, retrieved.ID)
	}
}

func TestListAgentRuns(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	// Create multiple agent runs
	for i := 0; i < 3; i++ {
		req := &models.AgentRunCreateRequest{
			AgentID:    "agent-123",
			OperatorID: "operator-1",
			Target:     "http://example.com",
			Title:      "Test Agent Run",
		}
		_, err := service.CreateAgentRun(req, "1")
		if err != nil {
			t.Fatalf("Failed to create agent run: %v", err)
		}
	}

	req := &models.AgentRunListRequest{
		AgentID:  "agent-123",
		Page:     1,
		PageSize: 10,
	}

	resp, err := service.ListAgentRuns(req)
	if err != nil {
		t.Fatalf("Failed to list agent runs: %v", err)
	}
	if resp == nil {
		t.Fatal("Expected response, got nil")
	}
	if len(resp.Items) != 3 {
		t.Errorf("Expected 3 agent runs, got %d", len(resp.Items))
	}
	if resp.Total != 3 {
		t.Errorf("Expected total 3, got %d", resp.Total)
	}
}

func TestUpdateAgentRunStatus(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	req := &models.AgentRunCreateRequest{
		AgentID:    "agent-123",
		OperatorID: "operator-1",
		Target:     "http://example.com",
		Title:      "Test Agent Run",
	}

	created, err := service.CreateAgentRun(req, "1")
	if err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	updateReq := &models.AgentRunUpdateStatusRequest{
		Status: models.AgentRunStatusRunning,
	}

	err = service.UpdateAgentRunStatus(created.ID, updateReq, "1")
	if err != nil {
		t.Fatalf("Failed to update agent run status: %v", err)
	}

	retrieved, err := service.GetAgentRunByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated agent run: %v", err)
	}
	if retrieved.Status != models.AgentRunStatusRunning {
		t.Errorf("Expected status %s, got %s", models.AgentRunStatusRunning, retrieved.Status)
	}
	if retrieved.StartedAt == nil {
		t.Error("Expected started_at to be set")
	}

	// Verify audit log was created for agent_run.status_updated
	var auditLog models.AuditLog
	has, err := engine.Where("action = ? AND resource_type = ?", "agent_run.status_updated", "agent_run").Get(&auditLog)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if !has {
		t.Error("Expected audit log entry for agent_run.status_updated")
	}
}

func TestAppendAgentOperation(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	req := &models.AgentRunCreateRequest{
		AgentID:    "agent-123",
		OperatorID: "operator-1",
		Target:     "http://example.com",
		Title:      "Test Agent Run",
	}

	created, err := service.CreateAgentRun(req, "1")
	if err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	opReq := &models.AgentOperationCreateRequest{
		Action:    "create_oast_probe",
		RiskLevel: "medium",
		Request:   map[string]interface{}{"target": "http://example.com"},
		Result:    map[string]interface{}{"success": true},
	}

	err = service.AppendAgentOperation(created.ID, opReq, "1")
	if err != nil {
		t.Fatalf("Failed to append agent operation: %v", err)
	}

	// Verify operation was created
	var operation models.AgentOperation
	has, err := engine.Where("agent_run_i_d = ? AND action = ?", created.ID, "create_oast_probe").Get(&operation)
	if err != nil {
		t.Fatalf("Failed to query operation: %v", err)
	}
	if !has {
		t.Error("Expected operation to be created")
	}

	// Verify audit log was created for agent operation
	var auditLog models.AuditLog
	has, err = engine.Where("action = ? AND resource_type = ?", "agent_operation.create_oast_probe", "agent_run").Get(&auditLog)
	if err != nil {
		t.Fatalf("Failed to query audit log: %v", err)
	}
	if !has {
		t.Error("Expected audit log entry for agent_operation.create_oast_probe")
	}
}

func TestInvalidStatusTransition(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	req := &models.AgentRunCreateRequest{
		AgentID:    "agent-123",
		OperatorID: "operator-1",
		Target:     "http://example.com",
		Title:      "Test Agent Run",
	}

	created, err := service.CreateAgentRun(req, "1")
	if err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	// Try invalid transition: created -> completed (should fail)
	updateReq := &models.AgentRunUpdateStatusRequest{
		Status: models.AgentRunStatusCompleted,
	}

	err = service.UpdateAgentRunStatus(created.ID, updateReq, "1")
	if err == nil {
		t.Error("Expected error for invalid status transition")
	}
}

func TestCreateFollowupAction(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	created, err := service.CreateAgentRun(&models.AgentRunCreateRequest{
		AgentID:    "agent-1",
		OperatorID: "operator-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://target.example",
		Title:      "Review target",
	}, "1")
	if err != nil {
		t.Fatalf("create agent run: %v", err)
	}

	resp, err := service.CreateFollowupAction(created.ID, &models.AgentRunFollowupRequest{
		ActionType:     "recheck_evidence",
		Reason:         "Evidence is high confidence and needs second review",
		ReviewPacketID: created.ID,
	}, "1")
	if err != nil {
		t.Fatalf("create followup: %v", err)
	}

	if resp.AgentRunID != created.ID {
		t.Fatalf("expected agent run id %s, got %s", created.ID, resp.AgentRunID)
	}
	if resp.OperationID == "" {
		t.Fatal("expected operation id")
	}
	if resp.ActionType != "recheck_evidence" {
		t.Fatalf("unexpected action type %s", resp.ActionType)
	}

	var op models.AgentOperation
	has, err := engine.ID(resp.OperationID).Get(&op)
	if err != nil || !has {
		t.Fatalf("expected operation row, has=%v err=%v", has, err)
	}
	if op.Action != "followup.recheck_evidence" {
		t.Fatalf("unexpected operation action %s", op.Action)
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(op.Result), &result); err != nil {
		t.Fatalf("parse operation result: %v", err)
	}
	if result["source_agent_run_id"] != created.ID {
		t.Fatalf("operation result missing source_agent_run_id")
	}
	if result["action_type"] != "recheck_evidence" {
		t.Fatalf("operation result missing action_type")
	}

	var auditLog models.AuditLog
	has, err = engine.Where("action = ? AND resource_type = ?", "agent_run.followup_created", "agent_run").Get(&auditLog)
	if err != nil || !has {
		t.Fatalf("expected followup audit, has=%v err=%v", has, err)
	}
	if auditLog.Details["agent_run_id"] != created.ID {
		t.Fatalf("audit missing agent_run_id")
	}
}

func TestCreateFollowupActionValidation(t *testing.T) {
	engine := setupTestEngine(t)
	defer engine.Close()

	authService := auth.NewService(engine)
	service := NewService(engine, authService)

	created, err := service.CreateAgentRun(&models.AgentRunCreateRequest{
		AgentID:    "agent-1",
		OperatorID: "operator-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://target.example",
		Title:      "Review target",
	}, "1")
	if err != nil {
		t.Fatalf("create agent run: %v", err)
	}

	cases := []struct {
		name string
		id   string
		req  *models.AgentRunFollowupRequest
	}{
		{
			name: "unknown run",
			id:   "missing",
			req:  &models.AgentRunFollowupRequest{ActionType: "recheck_evidence", Reason: "valid reason"},
		},
		{
			name: "invalid action",
			id:   created.ID,
			req:  &models.AgentRunFollowupRequest{ActionType: "invalid_action", Reason: "valid reason"},
		},
		{
			name: "empty reason",
			id:   created.ID,
			req:  &models.AgentRunFollowupRequest{ActionType: "recheck_evidence", Reason: ""},
		},
		{
			name: "too long reason",
			id:   created.ID,
			req:  &models.AgentRunFollowupRequest{ActionType: "recheck_evidence", Reason: string(make([]byte, 501))},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := service.CreateFollowupAction(tc.id, tc.req, "1"); err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
