package payload

import (
	"testing"

	"github.com/chennqqi/godnslog/models"
	"xorm.io/xorm"
)

func TestCreatePayload(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblPayload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		Template:         "ssrf",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Errorf("Failed to create payload: %v", err)
	}
	if payload == nil {
		t.Error("Expected payload, got nil")
	}
}

func TestGetPayloadByID(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblPayload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		Template:         "ssrf",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	retrieved, err := service.GetPayloadByID(payload.ID)
	if err != nil {
		t.Errorf("Failed to get payload: %v", err)
	}
	if retrieved.Token != payload.Token {
		t.Errorf("Expected token %s, got %s", payload.Token, retrieved.Token)
	}
}

func TestUpdatePayload(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblPayload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	req := &PayloadCreateRequest{
		Template:         "ssrf",
		Variables:        map[string]string{},
		ExpectedProtocol: "dns",
	}

	payload, err := service.CreatePayload(req, "1", "example.com")
	if err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	updateReq := &PayloadUpdateRequest{
		Status: "hit",
	}

	err = service.UpdatePayload(payload.ID, updateReq)
	if err != nil {
		t.Errorf("Failed to update payload: %v", err)
	}

	retrieved, err := service.GetPayloadByID(payload.ID)
	if err != nil {
		t.Errorf("Failed to get updated payload: %v", err)
	}
	if retrieved.Template != "xxe" {
		t.Errorf("Expected updated template, got %s", retrieved.Template)
	}
}

func TestListPayloads(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	err = engine.Sync2(new(models.TblPayload))
	if err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	service := NewService(engine)

	for i := 0; i < 5; i++ {
		req := &PayloadCreateRequest{
			Template:         "ssrf",
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
		t.Errorf("Failed to list payloads: %v", err)
	}
	if len(payloads.Items) != 5 {
		t.Errorf("Expected 5 payloads in list, got %d", len(payloads.Items))
	}
}
