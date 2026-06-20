package workflow

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

// MockEngine creates a mock xorm engine for testing
func MockEngine() (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	// Sync tables
	err = engine.Sync2(
		new(models.Workflow),
	)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func TestNewService(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	service := NewService(engine)
	assert.NotNil(t, service)
}

func TestService_CreateWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)
	assert.NotEqual(t, "", workflow.ID)

	// Verify workflow was created
	var retrieved models.Workflow
	_, err = engine.ID(workflow.ID).Get(&retrieved)
	assert.NoError(t, err)
	assert.Equal(t, "Test Workflow", retrieved.Name)
}

func TestService_GetWorkflowByID(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Get workflow
	retrieved, err := service.GetWorkflowByID(workflow.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test Workflow", retrieved.Name)
}

func TestService_ListWorkflows(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	// Create multiple workflows
	for i := 0; i < 3; i++ {
		workflow := &models.Workflow{
			ID:          models.GenerateID(),
			Name:        "Test Workflow",
			Description: "Test Description",
			Enabled:     true,
			Actions:     models.Actions{},
			CreatedBy:   "test-user",
		}
		err = service.CreateWorkflow(workflow)
		assert.NoError(t, err)
	}

	// List workflows
	response, err := service.ListWorkflows("", nil, 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.GreaterOrEqual(t, len(response.Items), 3)
}

func TestService_UpdateWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Update workflow
	workflow.Name = "Updated Workflow"
	workflow.Description = "Updated Description"
	err = service.UpdateWorkflow(workflow)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := service.GetWorkflowByID(workflow.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Workflow", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
}

func TestService_DeleteWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Delete workflow
	err = service.DeleteWorkflow(workflow.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := service.GetWorkflowByID(workflow.ID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
}

func TestService_ExecuteWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions: models.Actions{
			{
				Type: "notify",
				Config: map[string]interface{}{
					"message": "Test notification",
				},
			},
		},
		CreatedBy: "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Execute workflow
	token := "test-token"
	interaction := &models.Interaction{
		ID:       models.GenerateID(),
		Type:     "dns",
		SourceIP: "192.168.1.1",
		Token:    &token,
	}
	err = service.ExecuteWorkflow(workflow.ID, interaction)
	assert.NoError(t, err)
}

// --- HTTP Action Executor Tests ---

func TestService_ExecuteHTTPAction_Success(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"127.0.0.1"}))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	action := models.Action{
		Type: models.ActionTypeHTTP, Enabled: true,
		Config: map[string]interface{}{"url": srv.URL, "method": "POST", "body": `{"test":true}`},
	}
	token := "test-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", SourceIP: "192.168.1.1", Token: &token}

	err = service.executeHTTPAction(action, interaction)
	assert.NoError(t, err)
}

func TestService_ExecuteHTTPAction_URLNotAllowed(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"allowed.com"}))

	action := models.Action{
		Type: models.ActionTypeHTTP, Enabled: true,
		Config: map[string]interface{}{"url": "https://evil.com/exfil", "method": "GET"},
	}
	token := "test-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeHTTPAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in allowlist")
}

func TestService_ExecuteHTTPAction_NoSecurityConfigured(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	action := models.Action{
		Type: models.ActionTypeHTTP, Enabled: true,
		Config: map[string]interface{}{"url": "https://example.com/hook"},
	}
	token := "test-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeHTTPAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not in allowlist")
}

// --- Webhook Action Executor Tests ---

func TestService_ExecuteWebhookAction_Success(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"127.0.0.1"}))

	var receivedBody string
	var receivedHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		receivedHeaders = r.Header
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	action := models.Action{
		Type: models.ActionTypeWebhook, Enabled: true,
		Config: map[string]interface{}{
			"url":     srv.URL,
			"method":  "POST",
			"headers": map[string]interface{}{"X-Custom": "testval"},
			"body":    `{"interaction_id":"{{.id}}","source_ip":"{{.source_ip}}"}`,
		},
	}
	token := "webhook-token"
	interaction := &models.Interaction{ID: "interaction-123", Type: "dns", SourceIP: "10.0.0.5", Token: &token}

	err = service.executeWebhookAction(action, interaction)
	assert.NoError(t, err)
	assert.Contains(t, receivedBody, "interaction-123")
	assert.Contains(t, receivedBody, "10.0.0.5")
	assert.Equal(t, "testval", receivedHeaders.Get("X-Custom"))
}

func TestService_ExecuteWebhookAction_SSRFBlocked(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"*"}))

	action := models.Action{
		Type: models.ActionTypeWebhook, Enabled: true,
		Config: map[string]interface{}{"url": "http://169.254.169.254/latest/meta-data"},
	}
	token := "test-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeWebhookAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "blocked")
}

// --- DNS Action Executor Tests ---

func TestService_ExecuteDNSAction_Success(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	action := models.Action{
		Type: models.ActionTypeDNS, Enabled: true,
		Config: map[string]interface{}{"domain": "localhost", "type": "A"},
	}
	token := "dns-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeDNSAction(action, interaction)
	assert.NoError(t, err)
}

func TestService_ExecuteDNSAction_MissingDomain(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	action := models.Action{Type: models.ActionTypeDNS, Enabled: true, Config: map[string]interface{}{}}
	token := "dns-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeDNSAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing domain")
}

func TestService_ExecuteDNSAction_UnsupportedType(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	action := models.Action{
		Type: models.ActionTypeDNS, Enabled: true,
		Config: map[string]interface{}{"domain": "localhost", "type": "SRV"},
	}
	token := "dns-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeDNSAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}

// --- Notify Action Executor Tests ---

func TestService_ExecuteNotifyAction_WebhookChannel(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)
	service.SetOutboundSecurity(NewOutboundSecurity([]string{"127.0.0.1"}))

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	action := models.Action{
		Type: models.ActionTypeNotify, Enabled: true,
		Config: map[string]interface{}{
			"channel":     "webhook",
			"webhook_url": srv.URL,
			"message":     "Interaction {{.id}} from {{.source_ip}}",
		},
	}
	token := "notify-token"
	interaction := &models.Interaction{ID: "inter-1", Type: "dns", SourceIP: "10.0.0.1", Token: &token}

	err = service.executeNotifyAction(action, interaction)
	assert.NoError(t, err)
}

func TestService_ExecuteNotifyAction_MissingChannel(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	action := models.Action{Type: models.ActionTypeNotify, Enabled: true, Config: map[string]interface{}{}}
	token := "notify-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeNotifyAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "missing channel")
}

func TestService_ExecuteNotifyAction_UnsupportedChannel(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	service := NewService(engine)

	action := models.Action{
		Type: models.ActionTypeNotify, Enabled: true,
		Config: map[string]interface{}{"channel": "email"},
	}
	token := "notify-token"
	interaction := &models.Interaction{ID: models.GenerateID(), Type: "dns", Token: &token}

	err = service.executeNotifyAction(action, interaction)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported")
}
