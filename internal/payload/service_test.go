package payload

import (
	"testing"

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
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
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
