package marketplace

import (
	"context"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	t.Cleanup(func() { engine.Close() })

	if err := engine.Sync2(
		new(Plugin), new(PluginVersion), new(PluginReview),
		new(Template), new(TemplateReview),
		new(PluginInstallation),
	); err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}
	return engine
}

func TestService_CreatePlugin(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	plugin := &Plugin{
		ID:   "p1",
		Name: "Test Plugin",
		Type: "listener",
	}
	if err := svc.CreatePlugin(ctx, plugin); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}
	if plugin.Downloads != 0 || plugin.Rating != 0 || plugin.Reviews != 0 {
		t.Fatalf("expected zero defaults, got downloads=%d rating=%f reviews=%d", plugin.Downloads, plugin.Rating, plugin.Reviews)
	}
	if plugin.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestService_GetPlugin(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	plugin := &Plugin{ID: "p2", Name: "Get Plugin", Type: "scanner"}
	if err := svc.CreatePlugin(ctx, plugin); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	got, err := svc.GetPlugin(ctx, "p2")
	if err != nil {
		t.Fatalf("GetPlugin failed: %v", err)
	}
	if got.Name != "Get Plugin" {
		t.Fatalf("expected name 'Get Plugin', got '%s'", got.Name)
	}
}

func TestService_ListPlugins(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		p := &Plugin{
			ID:   string(rune('a' + i)),
			Name: "Plugin",
			Type: "listener",
		}
		if err := svc.CreatePlugin(ctx, p); err != nil {
			t.Fatalf("CreatePlugin failed: %v", err)
		}
	}

	plugins, err := svc.ListPlugins(ctx, PluginFilters{})
	if err != nil {
		t.Fatalf("ListPlugins failed: %v", err)
	}
	if len(plugins) != 3 {
		t.Fatalf("expected 3 plugins, got %d", len(plugins))
	}
}

func TestService_PublishUnpublishPlugin(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p3", Name: "Pub Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	if err := svc.PublishPlugin(ctx, "p3"); err != nil {
		t.Fatalf("PublishPlugin failed: %v", err)
	}
	got, _ := svc.GetPlugin(ctx, "p3")
	if !got.IsPublished {
		t.Fatal("expected IsPublished=true after publish")
	}

	if err := svc.UnpublishPlugin(ctx, "p3"); err != nil {
		t.Fatalf("UnpublishPlugin failed: %v", err)
	}
	got, _ = svc.GetPlugin(ctx, "p3")
	if got.IsPublished {
		t.Fatal("expected IsPublished=false after unpublish")
	}
}

func TestService_AddPluginVersion(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p4", Name: "Versioned Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	v := &PluginVersion{ID: "v1", PluginID: "p4", Version: "1.0.0"}
	if err := svc.AddPluginVersion(ctx, v); err != nil {
		t.Fatalf("AddPluginVersion failed: %v", err)
	}

	got, _ := svc.GetPlugin(ctx, "p4")
	if got.LatestVersion != "1.0.0" {
		t.Fatalf("expected LatestVersion '1.0.0', got '%s'", got.LatestVersion)
	}
}

func TestService_AddPluginReview(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p5", Name: "Reviewed Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	r := &PluginReview{ID: "r1", PluginID: "p5", Rating: 5}
	if err := svc.AddPluginReview(ctx, r); err != nil {
		t.Fatalf("AddPluginReview failed: %v", err)
	}

	got, _ := svc.GetPlugin(ctx, "p5")
	if got.Rating != 5.0 {
		t.Fatalf("expected rating 5.0, got %f", got.Rating)
	}
	if got.Reviews != 1 {
		t.Fatalf("expected 1 review, got %d", got.Reviews)
	}
}

func TestService_InstallPlugin_NewAndDuplicate(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p6", Name: "Installable Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	// First install
	inst, err := svc.InstallPlugin(ctx, "p6", "1.0.0", "{}")
	if err != nil {
		t.Fatalf("InstallPlugin failed: %v", err)
	}
	if inst.ID == "" {
		t.Fatal("expected non-empty installation ID")
	}
	if inst.Status != "installed" {
		t.Fatalf("expected status 'installed', got '%s'", inst.Status)
	}

	// Second install should return existing
	inst2, err := svc.InstallPlugin(ctx, "p6", "1.0.0", "{}")
	if err != nil {
		t.Fatalf("Second InstallPlugin failed: %v", err)
	}
	if inst2.ID != inst.ID {
		t.Fatalf("expected same installation ID '%s', got '%s'", inst.ID, inst2.ID)
	}
}

func TestService_ListPluginInstallations(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p7", Name: "List Install Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	if _, err := svc.InstallPlugin(ctx, "p7", "1.0.0", "{}"); err != nil {
		t.Fatalf("InstallPlugin failed: %v", err)
	}

	installs, err := svc.ListPluginInstallations(ctx)
	if err != nil {
		t.Fatalf("ListPluginInstallations failed: %v", err)
	}
	if len(installs) != 1 {
		t.Fatalf("expected 1 installation, got %d", len(installs))
	}
}

func TestService_UninstallPlugin(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p8", Name: "Uninstall Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	inst, err := svc.InstallPlugin(ctx, "p8", "1.0.0", "{}")
	if err != nil {
		t.Fatalf("InstallPlugin failed: %v", err)
	}

	if err := svc.UninstallPlugin(ctx, inst.ID); err != nil {
		t.Fatalf("UninstallPlugin failed: %v", err)
	}

	installs, _ := svc.ListPluginInstallations(ctx)
	if len(installs) != 0 {
		t.Fatalf("expected 0 installations after uninstall, got %d", len(installs))
	}
}

func TestService_CreateTemplate(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	tmpl := &Template{ID: "t1", Name: "Test Template", Type: "payload"}
	if err := svc.CreateTemplate(ctx, tmpl); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}
	if tmpl.Downloads != 0 || tmpl.Rating != 0 {
		t.Fatal("expected zero defaults")
	}
}

func TestService_PublishTemplate(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	tmpl := &Template{ID: "t2", Name: "Pub Template", Type: "payload"}
	if err := svc.CreateTemplate(ctx, tmpl); err != nil {
		t.Fatalf("CreateTemplate failed: %v", err)
	}

	if err := svc.PublishTemplate(ctx, "t2"); err != nil {
		t.Fatalf("PublishTemplate failed: %v", err)
	}
	got, _ := svc.GetTemplate(ctx, "t2")
	if !got.IsPublished {
		t.Fatal("expected IsPublished=true")
	}
}

func TestService_DeletePlugin(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p9", Name: "Delete Plugin", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	if err := svc.DeletePlugin(ctx, "p9"); err != nil {
		t.Fatalf("DeletePlugin failed: %v", err)
	}

	// XORM Get returns nil error even when record not found; check ID is empty
	got, err := svc.GetPlugin(ctx, "p9")
	if err != nil {
		t.Fatalf("GetPlugin returned error: %v", err)
	}
	if got.ID != "" {
		t.Fatalf("expected empty ID after deletion, got '%s'", got.ID)
	}
}

func TestGenerateInstallationID(t *testing.T) {
	id1 := generateInstallationID()
	id2 := generateInstallationID()
	if id1 == id2 {
		t.Fatal("expected unique IDs")
	}
	if len(id1) == 0 {
		t.Fatal("expected non-empty ID")
	}
}

func TestService_UpdatePlugin(t *testing.T) {
	engine := setupTestEngine(t)
	svc := NewService(NewXormStore(engine))
	ctx := context.Background()

	p := &Plugin{ID: "p10", Name: "Original", Type: "listener"}
	if err := svc.CreatePlugin(ctx, p); err != nil {
		t.Fatalf("CreatePlugin failed: %v", err)
	}

	p.Name = "Updated"
	if err := svc.UpdatePlugin(ctx, p); err != nil {
		t.Fatalf("UpdatePlugin failed: %v", err)
	}

	got, _ := svc.GetPlugin(ctx, "p10")
	if got.Name != "Updated" {
		t.Fatalf("expected name 'Updated', got '%s'", got.Name)
	}
	if !got.UpdatedAt.After(got.CreatedAt) && !got.UpdatedAt.Equal(got.CreatedAt) {
		t.Fatal("expected UpdatedAt >= CreatedAt")
	}
	_ = time.Now() // ensure time import used
}
