package rebinding

import (
	"context"
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupRebindingEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	t.Cleanup(func() { engine.Close() })

	if err := engine.Sync2(
		new(models.RebindingRule),
		new(models.RebindingSession),
	); err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}
	return engine
}

func TestService_CreateRebindingRule(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	rule := &models.RebindingRule{
		Domain: "test.example.com",
		Stages: models.Stages{
			{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1},
			{Order: 1, TargetIP: "127.0.0.1", TTL: 0, MaxHits: 0},
		},
		IsEnabled: true,
	}

	if err := svc.CreateRebindingRule(rule); err != nil {
		t.Fatalf("CreateRebindingRule failed: %v", err)
	}
	if rule.ID == "" {
		t.Fatal("expected non-empty ID")
	}
	if rule.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestService_GetRebindingRule(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	rule := &models.RebindingRule{
		Domain: "get.example.com",
		Stages: models.Stages{
			{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1},
		},
		IsEnabled: true,
	}
	if err := svc.CreateRebindingRule(rule); err != nil {
		t.Fatalf("CreateRebindingRule failed: %v", err)
	}

	got, err := svc.GetRebindingRule(rule.ID)
	if err != nil {
		t.Fatalf("GetRebindingRule failed: %v", err)
	}
	if got.Domain != "get.example.com" {
		t.Fatalf("expected domain 'get.example.com', got '%s'", got.Domain)
	}
}

func TestService_GetRebindingRule_NotFound(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	_, err := svc.GetRebindingRule("nonexistent")
	if err != ErrRebindingRuleNotFound {
		t.Fatalf("expected ErrRebindingRuleNotFound, got %v", err)
	}
}

func TestService_ListRebindingRules(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	for i := 0; i < 3; i++ {
		rule := &models.RebindingRule{
			Domain:    "list" + string(rune('0'+i)) + ".example.com",
			Stages:    models.Stages{{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1}},
			IsEnabled: true,
		}
		if err := svc.CreateRebindingRule(rule); err != nil {
			t.Fatalf("CreateRebindingRule failed: %v", err)
		}
	}

	rules, total, err := svc.ListRebindingRules(1, 10)
	if err != nil {
		t.Fatalf("ListRebindingRules failed: %v", err)
	}
	if total != 3 {
		t.Fatalf("expected total 3, got %d", total)
	}
	if len(rules) != 3 {
		t.Fatalf("expected 3 rules, got %d", len(rules))
	}
}

func TestService_UpdateRebindingRule(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	rule := &models.RebindingRule{
		Domain:    "update.example.com",
		Stages:    models.Stages{{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1}},
		IsEnabled: true,
	}
	if err := svc.CreateRebindingRule(rule); err != nil {
		t.Fatalf("CreateRebindingRule failed: %v", err)
	}

	rule.Domain = "updated.example.com"
	if err := svc.UpdateRebindingRule(rule); err != nil {
		t.Fatalf("UpdateRebindingRule failed: %v", err)
	}

	got, _ := svc.GetRebindingRule(rule.ID)
	if got.Domain != "updated.example.com" {
		t.Fatalf("expected domain 'updated.example.com', got '%s'", got.Domain)
	}
}

func TestService_DeleteRebindingRule(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	rule := &models.RebindingRule{
		Domain:    "delete.example.com",
		Stages:    models.Stages{{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1}},
		IsEnabled: true,
	}
	if err := svc.CreateRebindingRule(rule); err != nil {
		t.Fatalf("CreateRebindingRule failed: %v", err)
	}

	if err := svc.DeleteRebindingRule(rule.ID); err != nil {
		t.Fatalf("DeleteRebindingRule failed: %v", err)
	}

	if _, err := svc.GetRebindingRule(rule.ID); err == nil {
		t.Fatal("expected error getting deleted rule")
	}
}

func TestService_CreateRuleFromScenario(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	rule, err := svc.CreateRuleFromScenario(string(models.ScenarioBrowserRebinding), "scenario.example.com")
	if err != nil {
		t.Fatalf("CreateRuleFromScenario failed: %v", err)
	}
	if rule.Domain != "scenario.example.com" {
		t.Fatalf("expected domain 'scenario.example.com', got '%s'", rule.Domain)
	}
	if len(rule.Stages) != 2 {
		t.Fatalf("expected 2 stages, got %d", len(rule.Stages))
	}
	if !rule.IsEnabled {
		t.Fatal("expected rule to be enabled")
	}
}

func TestService_CreateRuleFromScenario_NotFound(t *testing.T) {
	engine := setupRebindingEngine(t)
	svc := NewService(engine)

	_, err := svc.CreateRuleFromScenario("nonexistent-scenario", "test.example.com")
	if err == nil {
		t.Fatal("expected error for unknown scenario")
	}
}

func TestGetPredefinedScenarios(t *testing.T) {
	scenarios := GetPredefinedScenarios()
	if len(scenarios) != 5 {
		t.Fatalf("expected 5 predefined scenarios, got %d", len(scenarios))
	}

	names := make(map[string]bool)
	for _, sc := range scenarios {
		names[sc.Name] = true
		if len(sc.Stages) == 0 {
			t.Fatalf("scenario '%s' has no stages", sc.Name)
		}
	}

	expected := []string{
		string(models.ScenarioBrowserRebinding),
		string(models.ScenarioCloudMetadata),
		string(models.ScenarioInternalManagement),
		string(models.ScenarioIoTDevice),
		string(models.ScenarioRouterExploit),
	}
	for _, name := range expected {
		if !names[name] {
			t.Fatalf("missing scenario: %s", name)
		}
	}
}

func TestValidateIP(t *testing.T) {
	validIPs := []string{"127.0.0.1", "192.168.1.1", "169.254.169.254", "::1", "10.0.0.1"}
	for _, ip := range validIPs {
		if !ValidateIP(ip) {
			t.Fatalf("expected '%s' to be valid", ip)
		}
	}

	invalidIPs := []string{"not-an-ip", "999.999.999.999", "", "hello"}
	for _, ip := range invalidIPs {
		if ValidateIP(ip) {
			t.Fatalf("expected '%s' to be invalid", ip)
		}
	}
}

func TestDefaultRebindingConfig(t *testing.T) {
	config := DefaultRebindingConfig()
	if config.DefaultTTL != 60 {
		t.Fatalf("expected DefaultTTL 60, got %d", config.DefaultTTL)
	}
	if config.MaxStages != 5 {
		t.Fatalf("expected MaxStages 5, got %d", config.MaxStages)
	}
	if config.EnableC2 {
		t.Fatal("expected EnableC2=false by default")
	}
	if !config.RequireAuth {
		t.Fatal("expected RequireAuth=true by default")
	}
}

func TestResolver_CreateScenarioRule(t *testing.T) {
	engine := setupRebindingEngine(t)
	store := NewXormStore(engine)
	resolver := NewResolver(nil, store)

	ctx := context.Background()
	rule, err := resolver.CreateScenarioRule(ctx, "resolver.example.com", ScenarioBrowserRebinding)
	if err != nil {
		t.Fatalf("CreateScenarioRule failed: %v", err)
	}
	if rule.Domain != "resolver.example.com" {
		t.Fatalf("expected domain 'resolver.example.com', got '%s'", rule.Domain)
	}
	// Note: XormStore.CreateRebindingRule clears Stages to nil for separate storage
	// Verify the rule was persisted by retrieving it
	got, err := store.GetRebindingRule(ctx, rule.ID)
	if err != nil {
		t.Fatalf("GetRebindingRule failed: %v", err)
	}
	if got.Domain != "resolver.example.com" {
		t.Fatalf("expected persisted domain 'resolver.example.com', got '%s'", got.Domain)
	}
}

func TestResolver_CreateScenarioRule_Unknown(t *testing.T) {
	engine := setupRebindingEngine(t)
	store := NewXormStore(engine)
	resolver := NewResolver(nil, store)

	ctx := context.Background()
	_, err := resolver.CreateScenarioRule(ctx, "test.example.com", RebindingScenario("unknown"))
	if err == nil {
		t.Fatal("expected error for unknown scenario")
	}
}
