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

// TestValidatePackageHash tests the ValidatePackageHash function
func TestValidatePackageHash(t *testing.T) {
	tests := []struct {
		name        string
		packageHash string
		wantErr     error
	}{
		{
			name:        "valid 64-char hex string",
			packageHash: "abc123def4567890123456789012345678901234567890123456789012345678",
			wantErr:     nil,
		},
		{
			name:        "valid 64-char hex string uppercase",
			packageHash: "ABC123DEF4567890123456789012345678901234567890123456789012345678",
			wantErr:     nil,
		},
		{
			name:        "valid 64-char hex string mixed case",
			packageHash: "AbC123DeF4567890123456789012345678901234567890123456789012345678",
			wantErr:     nil,
		},
		{
			name:        "empty string",
			packageHash: "",
			wantErr:     ErrInvalidPackageHash,
		},
		{
			name:        "too short",
			packageHash: "abc123",
			wantErr:     ErrInvalidPackageHash,
		},
		{
			name:        "too long",
			packageHash: "abc123def4567890123456789012345678901234567890123456789012345678extra",
			wantErr:     ErrInvalidPackageHash,
		},
		{
			name:        "contains non-hex characters",
			packageHash: "abc123def456789012345678901234567890123456789012345678901234567g",
			wantErr:     ErrInvalidPackageHash,
		},
		{
			name:        "contains spaces",
			packageHash: "abc123def456789012345678901234567890123456789012345678901234567 ",
			wantErr:     ErrInvalidPackageHash,
		},
		{
			name:        "63 characters",
			packageHash: "abc123def456789012345678901234567890123456789012345678901234567",
			wantErr:     ErrInvalidPackageHash,
		},
		{
			name:        "65 characters",
			packageHash: "abc123def45678901234567890123456789012345678901234567890123456789",
			wantErr:     ErrInvalidPackageHash,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePackageHash(tt.packageHash)
			if err != tt.wantErr {
				t.Errorf("ValidatePackageHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestListReviewDeliveries tests the ListReviewDeliveries function
func TestListReviewDeliveries(t *testing.T) {
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
	startedAt := time.Now()
	agentRun := &models.AgentRun{
		ID:         "agent-run-1",
		AgentID:    "agent-123",
		OperatorID: "user-1",
		Status:     "completed",
		StartedAt:  &startedAt,
	}
	if _, err := engine.Insert(agentRun); err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	// Test non-existent agent run
	t.Run("Non-existent agent run", func(t *testing.T) {
		_, err := reviewService.ListReviewDeliveries("non-existent")
		if err == nil {
			t.Error("Expected error for non-existent agent run")
		}
		if !strings.Contains(err.Error(), "not found") {
			t.Errorf("Expected not found error, got: %v", err)
		}
	})

	// Test empty delivery history
	t.Run("Empty delivery history", func(t *testing.T) {
		resp, err := reviewService.ListReviewDeliveries("agent-run-1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resp.AgentRunID != "agent-run-1" {
			t.Error("AgentRunID mismatch")
		}
		if resp.Summary.Total != 0 {
			t.Error("Expected total to be 0")
		}
		if len(resp.Items) != 0 {
			t.Error("Expected no items")
		}
	})

	// Create a delivered operation
	deliveredRequest := map[string]interface{}{
		"format":           "markdown",
		"destination_host": "hooks.example.com",
		"header_names":     []string{"X-Custom-Header"},
	}
	deliveredRequestJSON, _ := json.Marshal(deliveredRequest)
	deliveredResult := map[string]interface{}{
		"delivery_id":         "delivery-123",
		"export_operation_id": "export-123",
		"status_code":         200,
		"result":              "delivered",
		"delivered_at":        time.Now().Format(time.RFC3339),
	}
	deliveredResultJSON, _ := json.Marshal(deliveredResult)
	deliveredOp := &models.AgentOperation{
		ID:         "op-delivery-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Request:    string(deliveredRequestJSON),
		Result:     string(deliveredResultJSON),
		StartedAt:  time.Now(),
	}
	if _, err := engine.Insert(deliveredOp); err != nil {
		t.Fatalf("Failed to create delivered operation: %v", err)
	}

	// Create audit log for delivered
	userID := "user-1"
	agentRunID := agentRun.ID
	auditLog := &models.AuditLog{
		ID:           "audit-delivery-1",
		UserID:       &userID,
		Action:       "agent_run.review_delivered",
		ResourceType: "agent_run",
		ResourceID:   &agentRunID,
		Details: models.AuditDetails{
			"delivery_operation_id": "op-delivery-1",
		},
		Timestamp: time.Now(),
	}
	if _, err := engine.Insert(auditLog); err != nil {
		t.Fatalf("Failed to create audit log: %v", err)
	}

	// Test delivered item
	t.Run("Delivered item", func(t *testing.T) {
		resp, err := reviewService.ListReviewDeliveries("agent-run-1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resp.Summary.Total != 1 {
			t.Errorf("Expected total 1, got %d", resp.Summary.Total)
		}
		if resp.Summary.Delivered != 1 {
			t.Errorf("Expected delivered 1, got %d", resp.Summary.Delivered)
		}
		if len(resp.Items) != 1 {
			t.Fatalf("Expected 1 item, got %d", len(resp.Items))
		}
		item := resp.Items[0]
		if item.DeliveryOperationID != "op-delivery-1" {
			t.Error("DeliveryOperationID mismatch")
		}
		if item.DeliveryID != "delivery-123" {
			t.Error("DeliveryID mismatch")
		}
		if item.ExportOperationID != "export-123" {
			t.Error("ExportOperationID mismatch")
		}
		if item.Format != "markdown" {
			t.Error("Format mismatch")
		}
		if item.Result != "delivered" {
			t.Error("Result mismatch")
		}
		if item.DestinationHost != "hooks.example.com" {
			t.Error("DestinationHost mismatch")
		}
		if item.StatusCode != 200 {
			t.Error("StatusCode mismatch")
		}
		if len(item.HeaderNames) != 1 || item.HeaderNames[0] != "X-Custom-Header" {
			t.Error("HeaderNames mismatch")
		}
		if item.AuditRefID != "audit-delivery-1" {
			t.Error("AuditRefID mismatch")
		}
		// Verify no full webhook URL or header values in response
		responseJSON, _ := json.Marshal(resp)
		responseStr := string(responseJSON)
		if strings.Contains(responseStr, "https://") || strings.Contains(responseStr, "http://") {
			t.Error("Response should not contain full webhook URL")
		}
		if strings.Contains(responseStr, "test-value") {
			t.Error("Response should not contain header values")
		}
	})

	// Create a failed operation
	failedRequest := map[string]interface{}{
		"format":           "json",
		"destination_host": "hooks.example.com",
	}
	failedRequestJSON, _ := json.Marshal(failedRequest)
	failedResult := map[string]interface{}{
		"delivery_id": "delivery-456",
		"status_code": 500,
		"result":      "failed",
		"error":       "internal server error",
	}
	failedResultJSON, _ := json.Marshal(failedResult)
	failedOp := &models.AgentOperation{
		ID:         "op-delivery-2",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Request:    string(failedRequestJSON),
		Result:     string(failedResultJSON),
		StartedAt:  time.Now(),
	}
	if _, err := engine.Insert(failedOp); err != nil {
		t.Fatalf("Failed to create failed operation: %v", err)
	}

	// Test failed item
	t.Run("Failed item", func(t *testing.T) {
		resp, err := reviewService.ListReviewDeliveries("agent-run-1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resp.Summary.Total != 2 {
			t.Errorf("Expected total 2, got %d", resp.Summary.Total)
		}
		if resp.Summary.Failed != 1 {
			t.Errorf("Expected failed 1, got %d", resp.Summary.Failed)
		}
		// Find failed item
		var failedItem *models.AgentRunReviewDeliveryHistoryItem
		for _, item := range resp.Items {
			if item.DeliveryID == "delivery-456" {
				failedItem = &item
				break
			}
		}
		if failedItem == nil {
			t.Fatal("Failed item not found")
		}
		if failedItem.Result != "failed" {
			t.Error("Result should be failed")
		}
		if failedItem.StatusCode != 500 {
			t.Error("StatusCode should be 500")
		}
		if failedItem.ErrorSummary != "internal server error" {
			t.Error("ErrorSummary mismatch")
		}
	})

	// Create a timeout operation (real Sprint Q pattern: result="failed" with timeout error)
	timeoutRequest := map[string]interface{}{
		"format":           "markdown",
		"destination_host": "hooks.example.com",
	}
	timeoutRequestJSON, _ := json.Marshal(timeoutRequest)
	timeoutResult := map[string]interface{}{
		"delivery_id": "delivery-789",
		"result":      "failed",
		"error":       "request timed out",
	}
	timeoutResultJSON, _ := json.Marshal(timeoutResult)
	timeoutOp := &models.AgentOperation{
		ID:         "op-delivery-3",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Request:    string(timeoutRequestJSON),
		Result:     string(timeoutResultJSON),
		StartedAt:  time.Now(),
	}
	if _, err := engine.Insert(timeoutOp); err != nil {
		t.Fatalf("Failed to create timeout operation: %v", err)
	}

	// Test timeout item
	t.Run("Timeout item", func(t *testing.T) {
		resp, err := reviewService.ListReviewDeliveries("agent-run-1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if resp.Summary.Total != 3 {
			t.Errorf("Expected total 3, got %d", resp.Summary.Total)
		}
		if resp.Summary.Timeout != 1 {
			t.Errorf("Expected timeout 1, got %d", resp.Summary.Timeout)
		}
		// Find timeout item
		var timeoutItem *models.AgentRunReviewDeliveryHistoryItem
		for _, item := range resp.Items {
			if item.DeliveryID == "delivery-789" {
				timeoutItem = &item
				break
			}
		}
		if timeoutItem == nil {
			t.Fatal("Timeout item not found")
		}
		if timeoutItem.Result != "timeout" {
			t.Error("Result should be timeout")
		}
		if timeoutItem.ErrorSummary != "request timed out" {
			t.Error("ErrorSummary mismatch")
		}
	})
}

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
		// Inject custom URL validator that skips DNS resolution for test
		reviewService.urlValidator = func(url string) error {
			// Skip DNS resolution for test URLs
			if strings.Contains(url, "hooks.example.com") {
				return nil
			}
			return ValidateWebhookURL(url)
		}

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
		// Inject custom URL validator that skips DNS resolution for test
		reviewService.urlValidator = func(url string) error {
			// Skip DNS resolution for test URLs
			if strings.Contains(url, "hooks.example.com") {
				return nil
			}
			return ValidateWebhookURL(url)
		}

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
		// Inject custom URL validator that skips DNS resolution for test
		reviewService.urlValidator = func(url string) error {
			// Skip DNS resolution for test URLs
			if strings.Contains(url, "hooks.example.com") {
				return nil
			}
			return ValidateWebhookURL(url)
		}

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

	// Test webhook timeout
	t.Run("Webhook timeout", func(t *testing.T) {
		// Create mock webhook server with TLS that delays response
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(6 * time.Second) // Exceed the 5s timeout
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

		// Inject custom HTTP client with short timeout
		reviewService.httpClient = &http.Client{
			Timeout: 1 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		req := &models.AgentRunReviewDeliveryRequest{
			Format:     "markdown",
			WebhookURL: server.URL,
		}

		resp, err := reviewService.DeliverReviewPackage("agent-run-1", req, "user-1")
		if err == nil {
			t.Error("Expected error for webhook timeout")
		}
		if !strings.Contains(err.Error(), "timed out") {
			t.Errorf("Expected timeout error, got: %v", err)
		}

		// Verify response indicates failure
		if resp != nil && resp.Result != "failed" {
			t.Error("Expected failed result in response")
		}

		// Verify failure operation was created
		var operation models.AgentOperation
		has, err := engine.Where("action = ?", "review_delivery.webhook").Get(&operation)
		if err != nil || !has {
			t.Error("Expected delivery operation to be created")
		}

		// Verify operation result contains timeout error
		if !strings.Contains(operation.Result, "timed out") {
			t.Error("Expected operation result to contain timeout error")
		}

		// Verify failure audit was created
		var auditLog models.AuditLog
		has, err = engine.Where("action = ?", "agent_run.review_delivery_failed").Get(&auditLog)
		if err != nil || !has {
			t.Error("Expected delivery failure audit log to be created")
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

// TestExportPackageHash tests that export package includes package_hash and manifest
func TestExportPackageHash(t *testing.T) {
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

	// Test JSON export with package_hash
	t.Run("JSON export includes package_hash and manifest", func(t *testing.T) {
		req := &models.AgentRunReviewExportRequest{
			Format:       "json",
			IncludeAudit: false,
		}
		resp, err := reviewService.ExportReviewPackage("agent-run-1", req, "user-1")
		if err != nil {
			t.Fatalf("Failed to export review package: %v", err)
		}

		// Verify package_hash is present
		if resp.PackageHash == "" {
			t.Error("Expected package_hash to be set in export response")
		}

		// Verify package_hash is 64 characters (SHA-256 hex)
		if len(resp.PackageHash) != 64 {
			t.Errorf("Expected package_hash to be 64 characters, got %d", len(resp.PackageHash))
		}

		// Verify manifest is present
		if resp.Manifest == nil {
			t.Error("Expected manifest to be set in export response")
		} else {
			// Verify manifest fields
			if resp.Manifest.SchemaVersion != "review-package-manifest/v1" {
				t.Errorf("Expected schema_version 'review-package-manifest/v1', got '%s'", resp.Manifest.SchemaVersion)
			}
			if resp.Manifest.AgentRunID != "agent-run-1" {
				t.Errorf("Expected agent_run_id 'agent-run-1', got '%s'", resp.Manifest.AgentRunID)
			}
			if resp.Manifest.Format != "json" {
				t.Errorf("Expected format 'json', got '%s'", resp.Manifest.Format)
			}
			if resp.Manifest.HashAlgorithm != "sha256" {
				t.Errorf("Expected hash_algorithm 'sha256', got '%s'", resp.Manifest.HashAlgorithm)
			}
			if resp.Manifest.PackageHash != resp.PackageHash {
				t.Error("Manifest package_hash should match response package_hash")
			}
			if resp.Manifest.Refs["operation_id"] == "" {
				t.Error("Expected manifest refs to contain operation_id")
			}
			if resp.Manifest.Refs["audit_ref_id"] == "" {
				t.Error("Expected manifest refs to contain audit_ref_id")
			}
		}

		// Verify operation result contains package_hash
		var operation models.AgentOperation
		has, err := engine.Where("action = ?", "review_export.json").Get(&operation)
		if err != nil || !has {
			t.Fatal("Expected export operation to be created")
		}
		var opResult map[string]interface{}
		if err := json.Unmarshal([]byte(operation.Result), &opResult); err != nil {
			t.Fatalf("Failed to unmarshal operation result: %v", err)
		}
		if opResult["package_hash"] == nil {
			t.Error("Expected operation result to contain package_hash")
		}
		if opResult["package_hash"].(string) != resp.PackageHash {
			t.Error("Operation result package_hash should match response package_hash")
		}

		// Verify audit details contain package_hash
		var auditLog models.AuditLog
		has, err = engine.Where("action = ?", "agent_run.review_exported").Get(&auditLog)
		if err != nil || !has {
			t.Fatal("Expected export audit log to be created")
		}
		if auditLog.Details["package_hash"] == nil {
			t.Error("Expected audit details to contain package_hash")
		}
		if auditLog.Details["package_hash"].(string) != resp.PackageHash {
			t.Error("Audit package_hash should match response package_hash")
		}
	})

	// Test Markdown export - package_hash should be computed from Markdown content
	t.Run("Markdown export has package_hash", func(t *testing.T) {
		req := &models.AgentRunReviewExportRequest{
			Format:       "markdown",
			IncludeAudit: false,
		}
		resp, err := reviewService.ExportReviewPackage("agent-run-1", req, "user-1")
		if err != nil {
			t.Fatalf("Failed to export review package: %v", err)
		}

		// Markdown format should have package_hash and manifest
		if resp.PackageHash == "" {
			t.Error("Expected package_hash to be non-empty for markdown format")
		}
		if len(resp.PackageHash) != 64 {
			t.Errorf("Expected package_hash to be 64 characters, got %d", len(resp.PackageHash))
		}
		if resp.Manifest == nil {
			t.Error("Expected manifest to be non-nil for markdown format")
		}
		if resp.Manifest != nil {
			if resp.Manifest.Format != "markdown" {
				t.Errorf("Expected manifest format to be markdown, got %s", resp.Manifest.Format)
			}
			if resp.Manifest.PackageHash != resp.PackageHash {
				t.Error("Expected manifest package_hash to match response package_hash")
			}
			if resp.Manifest.HashAlgorithm != "sha256" {
				t.Errorf("Expected hash_algorithm to be sha256, got %s", resp.Manifest.HashAlgorithm)
			}
		}
	})
}

// TestDeliveryPackageHash tests that delivery includes package_hash from export
func TestDeliveryPackageHash(t *testing.T) {
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

	// Test successful delivery with package_hash
	t.Run("Successful delivery includes package_hash", func(t *testing.T) {
		// First export to get package_hash
		exportReq := &models.AgentRunReviewExportRequest{
			Format: "json",
		}
		exportResp, err := reviewService.ExportReviewPackage("agent-run-1", exportReq, "user-1")
		if err != nil {
			t.Fatalf("Failed to export review package: %v", err)
		}
		exportPackageHash := exportResp.PackageHash
		if exportPackageHash == "" {
			t.Fatal("Expected export package_hash to be set")
		}

		// Create mock webhook server
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify webhook payload contains package_hash
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Errorf("Failed to decode webhook payload: %v", err)
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if payload["package_hash"] == nil {
				t.Error("Expected webhook payload to contain package_hash")
			}
			if payload["package_hash"].(string) != exportPackageHash {
				t.Error("Webhook payload package_hash should match export package_hash")
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"received"}`))
		}))
		defer server.Close()

		// Inject custom URL validator and HTTP client
		reviewService.urlValidator = func(url string) error {
			if url == server.URL {
				return nil
			}
			return ValidateWebhookURL(url)
		}
		reviewService.httpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// Deliver the package
		deliveryReq := &models.AgentRunReviewDeliveryRequest{
			Format:     "json",
			WebhookURL: server.URL,
		}
		deliveryResp, err := reviewService.DeliverReviewPackage("agent-run-1", deliveryReq, "user-1")
		if err != nil {
			t.Fatalf("Failed to deliver review package: %v", err)
		}

		// Verify delivery response contains package_hash
		if deliveryResp.PackageHash != exportPackageHash {
			t.Error("Delivery response package_hash should match export package_hash")
		}

		// Verify delivery operation result contains package_hash
		var deliveryOp models.AgentOperation
		has, err := engine.Where("action = ?", "review_delivery.webhook").OrderBy("created_at DESC").Limit(1).Get(&deliveryOp)
		if err != nil || !has {
			t.Fatal("Expected delivery operation to be created")
		}
		var deliveryOpResult map[string]interface{}
		if err := json.Unmarshal([]byte(deliveryOp.Result), &deliveryOpResult); err != nil {
			t.Fatalf("Failed to unmarshal delivery operation result: %v", err)
		}
		if deliveryOpResult["package_hash"] == nil {
			t.Error("Expected delivery operation result to contain package_hash")
		}
		if deliveryOpResult["package_hash"].(string) != exportPackageHash {
			t.Error("Delivery operation result package_hash should match export package_hash")
		}

		// Verify delivery audit details contain package_hash
		var deliveryAudit models.AuditLog
		has, err = engine.Where("action = ?", "agent_run.review_delivered").Get(&deliveryAudit)
		if err != nil || !has {
			t.Fatal("Expected delivery audit log to be created")
		}
		if deliveryAudit.Details["package_hash"] == nil {
			t.Error("Expected delivery audit details to contain package_hash")
		}
		if deliveryAudit.Details["package_hash"].(string) != exportPackageHash {
			t.Error("Delivery audit package_hash should match export package_hash")
		}
	})

	// Test failed delivery with package_hash
	t.Run("Failed delivery includes package_hash", func(t *testing.T) {
		// Create mock webhook server that returns 500
		server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"internal error"}`))
		}))
		defer server.Close()

		// Inject custom URL validator and HTTP client
		reviewService.urlValidator = func(url string) error {
			if url == server.URL {
				return nil
			}
			return ValidateWebhookURL(url)
		}
		reviewService.httpClient = &http.Client{
			Timeout: 5 * time.Second,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}

		// Get export package_hash
		exportReq := &models.AgentRunReviewExportRequest{
			Format: "json",
		}
		exportResp, err := reviewService.ExportReviewPackage("agent-run-1", exportReq, "user-1")
		if err != nil {
			t.Fatalf("Failed to export review package: %v", err)
		}
		exportPackageHash := exportResp.PackageHash

		// Deliver the package (should fail)
		deliveryReq := &models.AgentRunReviewDeliveryRequest{
			Format:     "json",
			WebhookURL: server.URL,
		}
		_, err = reviewService.DeliverReviewPackage("agent-run-1", deliveryReq, "user-1")
		if err == nil {
			t.Error("Expected delivery to fail")
		}

		// Verify failed delivery operation result contains package_hash
		var deliveryOp models.AgentOperation
		has, err := engine.Where("action = ?", "review_delivery.webhook").OrderBy("created_at DESC").Limit(1).Get(&deliveryOp)
		if err != nil || !has {
			t.Fatal("Expected delivery operation to be created")
		}
		var deliveryOpResult map[string]interface{}
		if err := json.Unmarshal([]byte(deliveryOp.Result), &deliveryOpResult); err != nil {
			t.Fatalf("Failed to unmarshal delivery operation result: %v", err)
		}
		if deliveryOpResult["package_hash"] == nil {
			t.Error("Expected failed delivery operation result to contain package_hash")
		}
		if deliveryOpResult["package_hash"].(string) != exportPackageHash {
			t.Error("Failed delivery operation result package_hash should match export package_hash")
		}

		// Verify failed delivery audit details contain package_hash
		var deliveryAudit models.AuditLog
		has, err = engine.Where("action = ?", "agent_run.review_delivery_failed").Get(&deliveryAudit)
		if err != nil || !has {
			t.Fatal("Expected delivery failure audit log to be created")
		}
		if deliveryAudit.Details["package_hash"] == nil {
			t.Error("Expected failed delivery audit details to contain package_hash")
		}
		if deliveryAudit.Details["package_hash"].(string) != exportPackageHash {
			t.Error("Failed delivery audit package_hash should match export package_hash")
		}
	})
}

// TestDeliveryHistoryPackageHash tests that delivery history includes package_hash
func TestDeliveryHistoryPackageHash(t *testing.T) {
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
		Status:     "completed",
	}
	if _, err := engine.Insert(agentRun); err != nil {
		t.Fatalf("Failed to create agent run: %v", err)
	}

	// Create delivery operation with package_hash
	deliveryRequest := map[string]interface{}{
		"format":           "json",
		"destination_host": "hooks.example.com",
	}
	deliveryRequestJSON, _ := json.Marshal(deliveryRequest)
	deliveryResult := map[string]interface{}{
		"delivery_id":         "delivery-123",
		"export_operation_id": "export-123",
		"status_code":         200,
		"result":              "delivered",
		"package_hash":        "abc123def4567890123456789012345678901234567890123456789012345678",
	}
	deliveryResultJSON, _ := json.Marshal(deliveryResult)
	deliveryOp := &models.AgentOperation{
		ID:         "op-delivery-1",
		AgentRunID: "agent-run-1",
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Request:    string(deliveryRequestJSON),
		Result:     string(deliveryResultJSON),
		StartedAt:  time.Now(),
	}
	if _, err := engine.Insert(deliveryOp); err != nil {
		t.Fatalf("Failed to create delivered operation: %v", err)
	}

	// Test delivery history includes package_hash
	t.Run("Delivery history includes package_hash", func(t *testing.T) {
		resp, err := reviewService.ListReviewDeliveries("agent-run-1")
		if err != nil {
			t.Fatalf("Failed to list deliveries: %v", err)
		}

		if len(resp.Items) != 1 {
			t.Fatalf("Expected 1 delivery item, got %d", len(resp.Items))
		}

		item := resp.Items[0]
		if item.PackageHash == "" {
			t.Error("Expected delivery history item to contain package_hash")
		}
		if item.PackageHash != "abc123def4567890123456789012345678901234567890123456789012345678" {
			t.Error("Delivery history item package_hash should match operation result")
		}
	})
}
