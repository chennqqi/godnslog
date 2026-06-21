package canary

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupCanaryEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	t.Cleanup(func() { engine.Close() })

	if err := engine.Sync2(
		new(models.Canary),
		new(models.CanaryHit),
	); err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}
	return engine
}

func TestDetector_Detect_DNS(t *testing.T) {
	engine := setupCanaryEngine(t)
	store := NewXormStore(engine)
	detector := NewDetector(nil, store)
	ctx := context.Background()

	// Create an active canary
	canary := &models.Canary{
		ID:        "c1",
		Type:      string(CanaryTypeDNS),
		Token:     "test-dns-token",
		IsEnabled: true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := store.CreateCanary(ctx, canary); err != nil {
		t.Fatalf("CreateCanary failed: %v", err)
	}

	// Simulate a DNS interaction matching the token
	domain := "test-dns-token.example.com"
	inter := models.Interaction{
		Type:     "dns",
		Domain:   &domain,
		SourceIP: "192.168.1.100",
	}

	hit, err := detector.Detect(ctx, inter)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if hit == nil {
		t.Fatal("expected canary hit, got nil")
	}
	if hit.CanaryID != "c1" {
		t.Fatalf("expected CanaryID 'c1', got '%s'", hit.CanaryID)
	}
	if hit.SourceIP != "192.168.1.100" {
		t.Fatalf("expected SourceIP '192.168.1.100', got '%s'", hit.SourceIP)
	}
}

func TestDetector_Detect_NoMatch(t *testing.T) {
	engine := setupCanaryEngine(t)
	store := NewXormStore(engine)
	detector := NewDetector(nil, store)
	ctx := context.Background()

	canary := &models.Canary{
		ID:        "c2",
		Type:      string(CanaryTypeDNS),
		Token:     "unique-token",
		IsEnabled: true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := store.CreateCanary(ctx, canary); err != nil {
		t.Fatalf("CreateCanary failed: %v", err)
	}

	domain := "non-matching.example.com"
	inter := models.Interaction{
		Type:   "dns",
		Domain: &domain,
	}

	hit, err := detector.Detect(ctx, inter)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if hit != nil {
		t.Fatal("expected no hit for non-matching interaction")
	}
}

func TestDetector_Detect_HTTP(t *testing.T) {
	engine := setupCanaryEngine(t)
	store := NewXormStore(engine)
	detector := NewDetector(nil, store)
	ctx := context.Background()

	canary := &models.Canary{
		ID:        "c3",
		Type:      string(CanaryTypeHTTP),
		Token:     "http-token-123",
		IsEnabled: true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := store.CreateCanary(ctx, canary); err != nil {
		t.Fatalf("CreateCanary failed: %v", err)
	}

	path := "/http-token-123/test"
	inter := models.Interaction{
		Type:     "http",
		Path:     &path,
		SourceIP: "10.0.0.1",
	}

	hit, err := detector.Detect(ctx, inter)
	if err != nil {
		t.Fatalf("Detect failed: %v", err)
	}
	if hit == nil {
		t.Fatal("expected canary hit for HTTP interaction")
	}
}

func TestDetector_AssessRisk(t *testing.T) {
	engine := setupCanaryEngine(t)
	store := NewXormStore(engine)
	detector := NewDetector(nil, store)

	// Critical: localhost
	hit := &CanaryHit{SourceIP: "127.0.0.1"}
	canary := &Canary{}
	if level := detector.AssessRisk(hit, canary); level != "critical" {
		t.Fatalf("expected 'critical' for localhost, got '%s'", level)
	}

	// High: curl
	hit = &CanaryHit{SourceIP: "8.8.8.8", UserAgent: "curl"}
	if level := detector.AssessRisk(hit, canary); level != "high" {
		t.Fatalf("expected 'high' for curl, got '%s'", level)
	}

	// High: wget
	hit = &CanaryHit{SourceIP: "8.8.8.8", UserAgent: "wget"}
	if level := detector.AssessRisk(hit, canary); level != "high" {
		t.Fatalf("expected 'high' for wget, got '%s'", level)
	}

	// Medium: regular UA
	hit = &CanaryHit{SourceIP: "8.8.8.8", UserAgent: "Mozilla/5.0"}
	if level := detector.AssessRisk(hit, canary); level != "medium" {
		t.Fatalf("expected 'medium' for browser UA, got '%s'", level)
	}

	// Low: no UA
	hit = &CanaryHit{SourceIP: "8.8.8.8"}
	if level := detector.AssessRisk(hit, canary); level != "low" {
		t.Fatalf("expected 'low' for no UA, got '%s'", level)
	}
}

func TestEncodeDecodeContext(t *testing.T) {
	ctx := &CanaryContext{
		Project:  "test-project",
		Asset:    "config.yaml",
		Location: "server-1",
		Owner:    "user-1",
		Purpose:  "detect unauthorized access",
	}

	encoded, err := EncodeContext(ctx)
	if err != nil {
		t.Fatalf("EncodeContext failed: %v", err)
	}

	decoded, err := DecodeContext(encoded)
	if err != nil {
		t.Fatalf("DecodeContext failed: %v", err)
	}

	if decoded.Project != ctx.Project {
		t.Fatalf("expected Project '%s', got '%s'", ctx.Project, decoded.Project)
	}
	if decoded.Owner != ctx.Owner {
		t.Fatalf("expected Owner '%s', got '%s'", ctx.Owner, decoded.Owner)
	}
}

func TestDecodeContext_Invalid(t *testing.T) {
	_, err := DecodeContext("invalid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDefaultCanaryConfig(t *testing.T) {
	config := DefaultCanaryConfig()
	if config.MaxRetentionDays != 90 {
		t.Fatalf("expected MaxRetentionDays 90, got %d", config.MaxRetentionDays)
	}
	if config.SilentWindow != 300 {
		t.Fatalf("expected SilentWindow 300, got %d", config.SilentWindow)
	}
	if config.CompressionThreshold != 10 {
		t.Fatalf("expected CompressionThreshold 10, got %d", config.CompressionThreshold)
	}
}

func TestService_CreateCanary(t *testing.T) {
	engine := setupCanaryEngine(t)
	svc := NewService(engine)

	c := &models.Canary{
		ID:        "svc-1",
		Type:      string(CanaryTypeDNS),
		Token:     "svc-token",
		IsEnabled: true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := svc.CreateCanary(c); err != nil {
		t.Fatalf("CreateCanary failed: %v", err)
	}
}

func TestService_GetCanary(t *testing.T) {
	engine := setupCanaryEngine(t)
	svc := NewService(engine)

	c := &models.Canary{
		ID:        "svc-2",
		Type:      string(CanaryTypeDNS),
		Token:     "svc-token-2",
		IsEnabled: true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := svc.CreateCanary(c); err != nil {
		t.Fatalf("CreateCanary failed: %v", err)
	}

	got, err := svc.GetCanary("svc-2")
	if err != nil {
		t.Fatalf("GetCanary failed: %v", err)
	}
	if got.Token != "svc-token-2" {
		t.Fatalf("expected token 'svc-token-2', got '%s'", got.Token)
	}
}

func TestService_ListCanaries(t *testing.T) {
	engine := setupCanaryEngine(t)
	svc := NewService(engine)

	for i := 0; i < 5; i++ {
		c := &models.Canary{
			ID:        string(rune('a' + i)),
			Type:      string(CanaryTypeDNS),
			Token:     "token",
			IsEnabled: true,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		if err := svc.CreateCanary(c); err != nil {
			t.Fatalf("CreateCanary failed: %v", err)
		}
	}

	canaries, total, err := svc.ListCanaries(1, 3)
	if err != nil {
		t.Fatalf("ListCanaries failed: %v", err)
	}
	if total != 5 {
		t.Fatalf("expected total 5, got %d", total)
	}
	if len(canaries) != 3 {
		t.Fatalf("expected 3 canaries on page 1, got %d", len(canaries))
	}
}

func TestService_DeleteCanary(t *testing.T) {
	engine := setupCanaryEngine(t)
	svc := NewService(engine)

	c := &models.Canary{
		ID:        "svc-del",
		Type:      string(CanaryTypeDNS),
		Token:     "del-token",
		IsEnabled: true,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	if err := svc.CreateCanary(c); err != nil {
		t.Fatalf("CreateCanary failed: %v", err)
	}

	if err := svc.DeleteCanary("svc-del"); err != nil {
		t.Fatalf("DeleteCanary failed: %v", err)
	}

	// XORM Get returns nil error even when record not found; check ID is empty
	got, err := svc.GetCanary("svc-del")
	if err != nil {
		t.Fatalf("GetCanary returned error: %v", err)
	}
	if got.ID != "" {
		t.Fatalf("expected empty ID after deletion, got '%s'", got.ID)
	}
}

func TestGenerateHitID(t *testing.T) {
	id1 := generateHitID()
	id2 := generateHitID()
	if id1 == id2 {
		t.Fatal("expected unique hit IDs")
	}
}

func TestEncodeContext_Invalid(t *testing.T) {
	// EncodeContext should always succeed with valid context
	ctx := &CanaryContext{
		Project: "test",
	}
	encoded, err := EncodeContext(ctx)
	if err != nil {
		t.Fatalf("EncodeContext failed: %v", err)
	}

	// Verify it's valid base64
	if _, err := base64.StdEncoding.DecodeString(encoded); err != nil {
		t.Fatalf("encoded value is not valid base64: %v", err)
	}
}
