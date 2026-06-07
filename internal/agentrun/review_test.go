package agentrun

import (
	"crypto/tls"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/models"
	_ "github.com/mattn/go-sqlite3"
	"xorm.io/xorm"
)

// TestBuildReviewPacket tests the BuildReviewPacket function
func TestBuildReviewPacket(t *testing.T) {
	// Setup in-memory database
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Sync tables
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

	// Create services
	authService := auth.NewService(engine)
	agentRunService := NewService(engine, authService)
	interactionService := interaction.NewService(engine)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := NewReviewService(engine, agentRunService, authService, evidenceService, interactionService)

	// Create test agent run
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
		EndedAt:    func() *time.Time { t := time.Now(); return &t }(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := engine.Insert(agentRun); err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Create test interactions
	token1 := "test-token"
	payloadID1 := "payload-1"
	caseID1 := "case-1"
	token2 := "test-token"
	payloadID2 := "payload-1"
	caseID2 := "case-1"
	interactions := []models.Interaction{
		{
			ID:        "interaction-1",
			Type:      "dns",
			Token:     &token1,
			PayloadID: &payloadID1,
			CaseID:    &caseID1,
			SourceIP:  "192.168.1.1",
			Timestamp: time.Now(),
			CreatedAt: time.Now(),
		},
		{
			ID:        "interaction-2",
			Type:      "http",
			Token:     &token2,
			PayloadID: &payloadID2,
			CaseID:    &caseID2,
			SourceIP:  "192.168.1.2",
			Timestamp: time.Now(),
			CreatedAt: time.Now(),
		},
	}
	for _, interaction := range interactions {
		if _, err := engine.Insert(&interaction); err != nil {
			t.Fatalf("Failed to insert interaction: %v", err)
		}
	}

	// Create test operation
	operation := &models.AgentOperation{
		ID:         "operation-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-1",
		Action:     "create_oast_probe",
		RiskLevel:  "medium",
		Request:    `{"title":"Test"}`,
		Result:     `{"case_id":"case-1"}`,
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	if _, err := engine.Insert(operation); err != nil {
		t.Fatalf("Failed to insert operation: %v", err)
	}

	// Create audit log with agent_run_id in details
	auditLog := &models.AuditLog{
		ID:           "audit-1",
		UserID:       func() *string { s := "user-1"; return &s }(),
		Action:       "agent_operation.create_oast_probe",
		ResourceType: "agent_run",
		ResourceID:   func() *string { s := "agent-run-1"; return &s }(),
		Details: models.AuditDetails{
			"agent_run_id": "agent-run-1",
			"action":       "create_oast_probe",
		},
		Timestamp: time.Now(),
		CreatedAt: time.Now(),
	}
	if _, err := engine.Insert(auditLog); err != nil {
		t.Fatalf("Failed to insert audit log: %v", err)
	}

	// Test JSON format
	t.Run("JSON format with interactions", func(t *testing.T) {
		packet, err := reviewService.BuildReviewPacket("agent-run-1", "json", "http://localhost:3000")
		if err != nil {
			t.Fatalf("Failed to build review packet: %v", err)
		}

		if packet.ID != "agent-run-1" {
			t.Errorf("Expected ID 'agent-run-1', got '%s'", packet.ID)
		}

		if packet.AgentRun.ID != "agent-run-1" {
			t.Errorf("Expected agent run ID 'agent-run-1', got '%s'", packet.AgentRun.ID)
		}

		if packet.CaseID != "case-1" {
			t.Errorf("Expected case ID 'case-1', got '%s'", packet.CaseID)
		}

		if packet.PayloadID != "payload-1" {
			t.Errorf("Expected payload ID 'payload-1', got '%s'", packet.PayloadID)
		}

		if packet.Target != "https://example.com" {
			t.Errorf("Expected target 'https://example.com', got '%s'", packet.Target)
		}

		if packet.InteractionSummary.Total != 2 {
			t.Errorf("Expected 2 interactions, got %d", packet.InteractionSummary.Total)
		}

		if packet.InteractionSummary.DNSCount != 1 {
			t.Errorf("Expected 1 DNS interaction, got %d", packet.InteractionSummary.DNSCount)
		}

		if packet.InteractionSummary.HTTPCount != 1 {
			t.Errorf("Expected 1 HTTP interaction, got %d", packet.InteractionSummary.HTTPCount)
		}

		if packet.InteractionSummary.UniqueSources != 2 {
			t.Errorf("Expected 2 unique sources, got %d", packet.InteractionSummary.UniqueSources)
		}

		if packet.Evidence == nil {
			t.Error("Expected evidence to be present")
		} else {
			if packet.Evidence.InteractionCount != 2 {
				t.Errorf("Expected evidence interaction count 2, got %d", packet.Evidence.InteractionCount)
			}
		}

		if len(packet.AuditRefs) == 0 {
			t.Error("Expected audit refs to be present")
		}

		if packet.Format != "json" {
			t.Errorf("Expected format 'json', got '%s'", packet.Format)
		}

		if packet.Content == "" {
			t.Error("Expected JSON content to be present")
		}
	})

	// Test Markdown format
	t.Run("Markdown format with interactions", func(t *testing.T) {
		packet, err := reviewService.BuildReviewPacket("agent-run-1", "markdown", "http://localhost:3000")
		if err != nil {
			t.Fatalf("Failed to build review packet: %v", err)
		}

		if packet.Format != "markdown" {
			t.Errorf("Expected format 'markdown', got '%s'", packet.Format)
		}

		if packet.Content == "" {
			t.Error("Expected markdown content to be present")
		}

		// Verify markdown contains key sections
		content := packet.Content
		if len(content) < 100 {
			t.Error("Expected markdown content to be substantial")
		}
	})

	// Test with no interactions
	t.Run("No interactions", func(t *testing.T) {
		// Delete existing interactions from previous test
		if _, err := engine.Exec("DELETE FROM interactions"); err != nil {
			t.Fatalf("Failed to delete interactions: %v", err)
		}

		// Create agent run without interactions (use different case_id to avoid conflict)
		caseID2 := "case-2"
		payloadID2 := "payload-2"
		agentRun2 := &models.AgentRun{
			ID:         "agent-run-2",
			AgentID:    "agent-1",
			OperatorID: "user-1",
			CaseID:     caseID2,
			PayloadID:  payloadID2,
			Target:     "https://example2.com",
			Title:      "Test Agent Run 2",
			Status:     models.AgentRunStatusCompleted,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if _, err := engine.Insert(agentRun2); err != nil {
			t.Fatalf("Failed to insert agent run 2: %v", err)
		}

		packet, err := reviewService.BuildReviewPacket("agent-run-2", "json", "http://localhost:3000")
		if err != nil {
			t.Fatalf("Failed to build review packet: %v", err)
		}

		if packet.InteractionSummary.Total != 0 {
			t.Errorf("Expected 0 interactions, got %d", packet.InteractionSummary.Total)
		}

		// Evidence can be nil when no interactions - not an error
		if packet.Evidence != nil {
			t.Error("Expected evidence to be nil when no interactions")
		}
	})

	// Test invalid format
	t.Run("Invalid format", func(t *testing.T) {
		_, err := reviewService.BuildReviewPacket("agent-run-1", "pdf", "http://localhost:3000")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
	})

	// Test agent run not found
	t.Run("Agent run not found", func(t *testing.T) {
		_, err := reviewService.BuildReviewPacket("non-existent", "json", "http://localhost:3000")
		if err == nil {
			t.Error("Expected error for non-existent agent run")
		}
	})
}

// TestDeliverReviewPackage tests the DeliverReviewPackage function
func TestDeliverReviewPackage(t *testing.T) {
	// Setup in-memory database
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Sync tables
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

	// Create services
	authService := auth.NewService(engine)
	agentRunService := NewService(engine, authService)
	interactionService := interaction.NewService(engine)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := NewReviewService(engine, agentRunService, authService, evidenceService, interactionService)

	// Create test agent run
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-1",
		OperatorID: "user-1",
		CaseID:     "case-1",
		PayloadID:  "payload-1",
		Target:     "https://example.com",
		Title:      "Test Agent Run",
		Status:     models.AgentRunStatusCompleted,
		StartedAt:  func() *time.Time { t := time.Now(); return &t }(),
		EndedAt:    func() *time.Time { t := time.Now(); return &t }(),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := engine.Insert(agentRun); err != nil {
		t.Fatalf("Failed to insert agent run: %v", err)
	}

	// Test invalid format
	t.Run("Invalid format", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "pdf",
			WebhookURL: "https://hooks.example.com/review",
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for invalid format")
		}
		if !strings.Contains(err.Error(), "invalid format") {
			t.Errorf("Expected invalid format error, got: %v", err)
		}
	})

	// Test blocked localhost URL
	t.Run("Blocked localhost URL", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: "https://localhost:8080/hook",
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for localhost URL")
		}
		if !strings.Contains(err.Error(), "localhost") {
			t.Errorf("Expected localhost error, got: %v", err)
		}
	})

	// Test blocked private IP
	t.Run("Blocked private IP", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: "https://192.168.1.1/hook",
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for private IP")
		}
		if !strings.Contains(err.Error(), "private") {
			t.Errorf("Expected private IP error, got: %v", err)
		}
	})

	// Test blocked metadata IP
	t.Run("Blocked metadata IP", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: "https://169.254.169.254/hook",
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for metadata IP")
		}
		if !strings.Contains(err.Error(), "metadata") {
			t.Errorf("Expected metadata IP error, got: %v", err)
		}
	})

	// Test forbidden header
	t.Run("Forbidden header", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: "https://hooks.example.com/review",
			Headers: map[string]string{
				"Authorization": "Bearer token",
			},
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for forbidden header")
		}
		if !strings.Contains(err.Error(), "header") {
			t.Errorf("Expected header error, got: %v", err)
		}
	})

	// Test unknown Agent Run
	t.Run("Unknown Agent Run", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: "https://hooks.example.com/review",
		}
		_, err := reviewService.DeliverReviewPackage("non-existent", req, "user-1")
		if err == nil {
			t.Error("Expected error for non-existent agent run")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})

	// Test review_packet_id mismatch
	t.Run("Review packet ID mismatch", func(t *testing.T) {
		req := &models.AgentRunReviewDeliveryRequest{
			Format:         "markdown",
			WebhookURL:     "https://hooks.example.com/review",
			ReviewPacketID: "different-id",
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for review_packet_id mismatch")
		}
		if !strings.Contains(err.Error(), "review_packet_id") {
			t.Errorf("Expected review_packet_id error, got: %v", err)
		}
	})

	// Test successful delivery with mock webhook server
	t.Run("Successful markdown delivery", func(t *testing.T) {
		// Create mock webhook server with TLS
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"received"}`))
		}))
		defer server.Close()

		// Inject custom URL validator that allows test server
		reviewService.urlValidator = func(url string) error {
			if url == server.URL {
				return nil
			}
			return ValidateWebhookURL(url)
		}

		// Inject custom HTTP client that skips TLS verification for testing
		reviewService.httpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: server.URL,
			Headers: map[string]string{
				"X-GODNSLOG-Source": "operator",
			},
		}
		resp, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err != nil {
			t.Fatalf("Expected successful delivery, got error: %v", err)
		}

		if resp.AgentRunID != "agent-run-1" {
			t.Errorf("Expected agent_run_id 'agent-run-1', got '%s'", resp.AgentRunID)
		}

		if resp.Format != "markdown" {
			t.Errorf("Expected format 'markdown', got '%s'", resp.Format)
		}

		if resp.Result != "delivered" {
			t.Errorf("Expected result 'delivered', got '%s'", resp.Result)
		}

		if resp.StatusCode != 200 {
			t.Errorf("Expected status code 200, got %d", resp.StatusCode)
		}

		if resp.DeliveryID == "" {
			t.Error("Expected delivery_id to be set")
		}

		if resp.DeliveryOperation == "" {
			t.Error("Expected delivery_operation_id to be set")
		}

		if resp.AuditRefID == "" {
			t.Error("Expected audit_ref_id to be set")
		}

		// Verify operation was created
		var operation models.AgentOperation
		has, err := engine.Where("action = ?", "review_delivery.webhook").Get(&operation)
		if err != nil || !has {
			t.Error("Expected delivery operation to be created")
		}

		// Verify operation does not contain full webhook URL
		if strings.Contains(operation.Request, server.URL) {
			t.Error("Operation request should not contain full webhook URL")
		}

		// Verify audit was created
		var auditLog models.AuditLog
		has, err = engine.Where("action = ?", "agent_run.review_delivered").Get(&auditLog)
		if err != nil || !has {
			t.Error("Expected delivery audit log to be created")
		}

		// Verify audit does not contain full webhook URL
		auditJSON, _ := json.Marshal(auditLog.Details)
		if strings.Contains(string(auditJSON), server.URL) {
			t.Error("Audit details should not contain full webhook URL")
		}
	})

	// Test webhook non-2xx response
	t.Run("Webhook non-2xx response", func(t *testing.T) {
		// Create mock webhook server with TLS that returns 500
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal error"}`))
		}))
		defer server.Close()

		// Inject custom URL validator that allows test server
		reviewService.urlValidator = func(url string) error {
			if url == server.URL {
				return nil
			}
			return ValidateWebhookURL(url)
		}

		// Inject custom HTTP client that skips TLS verification for testing
		reviewService.httpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: server.URL,
		}
		_, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for non-2xx response")
		}
		if !strings.Contains(err.Error(), "delivery failed") {
			t.Errorf("Expected delivery failed error, got: %v", err)
		}

		// Verify failure operation was created
		var operation models.AgentOperation
		has, err := engine.Where("action = ?", "review_delivery.webhook").OrderBy("created_at DESC").Limit(1).Get(&operation)
		if err != nil || !has {
			t.Error("Expected delivery operation to be created")
		}

		// Verify failure audit was created
		var auditLog models.AuditLog
		has, err = engine.Where("action = ?", "agent_run.review_delivery_failed").OrderBy("timestamp DESC").Limit(1).Get(&auditLog)
		if err != nil || !has {
			t.Error("Expected delivery failure audit log to be created")
		}
	})
}
