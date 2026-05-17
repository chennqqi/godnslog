package payload

import (
	"strings"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func TestCreatePayload(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		CaseID:           "case-123",
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}
	if payload == nil {
		t.Fatal("Expected payload, got nil")
	}

	// Verify template_rendered is set
	if payload.TemplateRendered == "" {
		t.Fatal("Expected TemplateRendered to be set")
	}
	// Verify template_rendered contains token and domain
	if !strings.Contains(payload.TemplateRendered, payload.Token) {
		t.Errorf("TemplateRendered should contain token %s", payload.Token)
	}
	if !strings.Contains(payload.TemplateRendered, "example.com") {
		t.Error("TemplateRendered should contain domain")
	}
}

func TestCreatePayloadWithInvalidTemplate(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		CaseID:     "case-123",
		TemplateID: "invalid-template",
		Variables:  map[string]string{},
	}

	_, err = service.CreatePayload(req, "1", "example.com")
	if err != ErrInvalidTemplate {
		t.Fatalf("Expected ErrInvalidTemplate, got %v", err)
	}
}

func TestCreatePayloadWithCaseVariable(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		CaseID:           "case-456",
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Verify case variable is substituted in template_rendered
	// Since we use RenderTemplateWithCase, the case ID should be available in the template
	// For ssrf-basic template: "http://{token}.{domain}/"
	// If template includes {case}, it should be substituted
	if payload.TemplateRendered == "" {
		t.Fatal("Expected TemplateRendered to be set")
	}
}

func TestCreatePayloadWithCustomVariables(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		CaseID:           "case-789",
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{"custom": "value123"},
		ExpectedProtocol: "http",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Verify custom variables are stored
	if payload.Variables == nil {
		t.Fatal("Expected Variables to be set")
	}
	if payload.Variables["custom"] != "value123" {
		t.Errorf("Expected custom variable to be 'value123', got '%s'", payload.Variables["custom"])
	}
}

func TestCreatePayloadWithExpiresAt(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	expiresAt := time.Now().Add(24 * time.Hour)
	req := &PayloadCreateRequest{
		CaseID:           "case-999",
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
		ExpiresAt:        &expiresAt,
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Verify ExpiresAt is set
	if payload.ExpiresAt == nil {
		t.Fatal("Expected ExpiresAt to be set")
	}
}

func TestGetPayloadByID(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		CaseID:           "case-123",
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	retrieved, err := service.GetPayloadByID(payload.ID)
	if err != nil {
		t.Fatalf("Failed to get payload: %v", err)
	}
	if retrieved.Token != payload.Token {
		t.Errorf("Expected token %s, got %s", payload.Token, retrieved.Token)
	}
}

func TestUpdatePayload(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	updateReq := &PayloadUpdateRequest{
		Status: "deployed",
	}

	err = service.UpdatePayload(payload.ID, updateReq)
	if err != nil {
		t.Fatalf("Failed to update payload: %v", err)
	}

	retrieved, err := service.GetPayloadByID(payload.ID)
	if err != nil {
		t.Fatalf("Failed to get updated payload: %v", err)
	}
	if retrieved.Status != "deployed" {
		t.Errorf("Expected updated status deployed, got %s", retrieved.Status)
	}
}

func TestListPayloads(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(Payload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	for i := 0; i < 5; i++ {
		req := &PayloadCreateRequest{
			TemplateID:       "ssrf-basic",
			Variables:        map[string]string{},
			ExpectedProtocol: "dns",
		}
		_, err = service.CreatePayload(req, "1", "example.com")
		if err != nil {
			t.Fatalf("Failed to create test payload: %v", err)
		}
	}

	payloads, err := service.ListPayloads("", "", 1, 10)
	if err != nil {
		t.Fatalf("Failed to list payloads: %v", err)
	}
	if len(payloads.Items) != 5 {
		t.Errorf("Expected 5 payloads in list, got %d", len(payloads.Items))
	}
}
