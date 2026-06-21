package workspace

import (
	"context"
	"testing"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupWorkspaceEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	t.Cleanup(func() { engine.Close() })

	if err := engine.Sync2(
		new(Workspace),
		new(WorkspaceMember),
		new(WorkspaceDomain),
	); err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}
	return engine
}

func TestWorkspaceTableName(t *testing.T) {
	if (Workspace{}).TableName() != "workspaces" {
		t.Fatal("expected table name 'workspaces'")
	}
	if (WorkspaceMember{}).TableName() != "workspace_members" {
		t.Fatal("expected table name 'workspace_members'")
	}
	if (WorkspaceDomain{}).TableName() != "workspace_domains" {
		t.Fatal("expected table name 'workspace_domains'")
	}
}

func TestXormStore_CreateWorkspace(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{
		ID:          "ws-1",
		Name:        "Test Workspace",
		Description: "A test workspace",
		OwnerID:     "user-1",
		IsEnabled:   true,
	}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}
	if ws.CreatedAt.IsZero() {
		t.Fatal("expected CreatedAt to be set")
	}
}

func TestXormStore_GetWorkspace(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-2", Name: "Get WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	got, err := store.GetWorkspace(ctx, "ws-2")
	if err != nil {
		t.Fatalf("GetWorkspace failed: %v", err)
	}
	if got.Name != "Get WS" {
		t.Fatalf("expected name 'Get WS', got '%s'", got.Name)
	}
}

func TestXormStore_GetWorkspaceByOwner(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		ws := &Workspace{
			ID:        string(rune('a' + i)),
			Name:      "Owner WS",
			OwnerID:   "owner-1",
			IsEnabled: true,
		}
		if err := store.CreateWorkspace(ctx, ws); err != nil {
			t.Fatalf("CreateWorkspace failed: %v", err)
		}
	}

	workspaces, err := store.GetWorkspaceByOwner(ctx, "owner-1")
	if err != nil {
		t.Fatalf("GetWorkspaceByOwner failed: %v", err)
	}
	if len(workspaces) != 2 {
		t.Fatalf("expected 2 workspaces, got %d", len(workspaces))
	}
}

func TestXormStore_GetAllWorkspaces(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		ws := &Workspace{
			ID:        string(rune('x' + i)),
			Name:      "All WS",
			OwnerID:   "owner",
			IsEnabled: true,
		}
		if err := store.CreateWorkspace(ctx, ws); err != nil {
			t.Fatalf("CreateWorkspace failed: %v", err)
		}
	}

	workspaces, err := store.GetAllWorkspaces(ctx)
	if err != nil {
		t.Fatalf("GetAllWorkspaces failed: %v", err)
	}
	if len(workspaces) != 3 {
		t.Fatalf("expected 3 workspaces, got %d", len(workspaces))
	}
}

func TestXormStore_UpdateWorkspace(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-upd", Name: "Original", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	ws.Name = "Updated"
	if err := store.UpdateWorkspace(ctx, ws); err != nil {
		t.Fatalf("UpdateWorkspace failed: %v", err)
	}

	got, _ := store.GetWorkspace(ctx, "ws-upd")
	if got.Name != "Updated" {
		t.Fatalf("expected name 'Updated', got '%s'", got.Name)
	}
}

func TestXormStore_DeleteWorkspace(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-del", Name: "Delete WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	if err := store.DeleteWorkspace(ctx, "ws-del"); err != nil {
		t.Fatalf("DeleteWorkspace failed: %v", err)
	}
}

func TestXormStore_AddAndGetWorkspaceMembers(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-mbr", Name: "Member WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	m1 := &WorkspaceMember{ID: "m1", WorkspaceID: "ws-mbr", UserID: "user-2", Role: "member"}
	m2 := &WorkspaceMember{ID: "m2", WorkspaceID: "ws-mbr", UserID: "user-3", Role: "viewer"}
	if err := store.AddWorkspaceMember(ctx, m1); err != nil {
		t.Fatalf("AddWorkspaceMember failed: %v", err)
	}
	if err := store.AddWorkspaceMember(ctx, m2); err != nil {
		t.Fatalf("AddWorkspaceMember failed: %v", err)
	}

	members, err := store.GetWorkspaceMembers(ctx, "ws-mbr")
	if err != nil {
		t.Fatalf("GetWorkspaceMembers failed: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}

	// Get specific member
	got, err := store.GetWorkspaceMember(ctx, "ws-mbr", "user-2")
	if err != nil {
		t.Fatalf("GetWorkspaceMember failed: %v", err)
	}
	if got.Role != "member" {
		t.Fatalf("expected role 'member', got '%s'", got.Role)
	}
}

func TestXormStore_RemoveWorkspaceMember(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-rm", Name: "Remove WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	m := &WorkspaceMember{ID: "rm1", WorkspaceID: "ws-rm", UserID: "user-2", Role: "member"}
	if err := store.AddWorkspaceMember(ctx, m); err != nil {
		t.Fatalf("AddWorkspaceMember failed: %v", err)
	}

	if err := store.RemoveWorkspaceMember(ctx, "ws-rm", "user-2"); err != nil {
		t.Fatalf("RemoveWorkspaceMember failed: %v", err)
	}

	members, _ := store.GetWorkspaceMembers(ctx, "ws-rm")
	if len(members) != 0 {
		t.Fatalf("expected 0 members after removal, got %d", len(members))
	}
}

func TestXormStore_AddAndGetWorkspaceDomains(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-dom", Name: "Domain WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	d1 := &WorkspaceDomain{ID: "d1", WorkspaceID: "ws-dom", Domain: "example.com", IsPrimary: true}
	d2 := &WorkspaceDomain{ID: "d2", WorkspaceID: "ws-dom", Domain: "test.example.com", IsPrimary: false}
	if err := store.AddWorkspaceDomain(ctx, d1); err != nil {
		t.Fatalf("AddWorkspaceDomain failed: %v", err)
	}
	if err := store.AddWorkspaceDomain(ctx, d2); err != nil {
		t.Fatalf("AddWorkspaceDomain failed: %v", err)
	}

	domains, err := store.GetWorkspaceDomains(ctx, "ws-dom")
	if err != nil {
		t.Fatalf("GetWorkspaceDomains failed: %v", err)
	}
	if len(domains) != 2 {
		t.Fatalf("expected 2 domains, got %d", len(domains))
	}

	// Get primary domain
	primary, err := store.GetPrimaryDomain(ctx, "ws-dom")
	if err != nil {
		t.Fatalf("GetPrimaryDomain failed: %v", err)
	}
	if primary.Domain != "example.com" {
		t.Fatalf("expected primary domain 'example.com', got '%s'", primary.Domain)
	}
}

func TestXormStore_RemoveWorkspaceDomain(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-rd", Name: "Remove Domain WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	d := &WorkspaceDomain{ID: "rd1", WorkspaceID: "ws-rd", Domain: "remove.example.com", IsPrimary: true}
	if err := store.AddWorkspaceDomain(ctx, d); err != nil {
		t.Fatalf("AddWorkspaceDomain failed: %v", err)
	}

	if err := store.RemoveWorkspaceDomain(ctx, "rd1"); err != nil {
		t.Fatalf("RemoveWorkspaceDomain failed: %v", err)
	}

	domains, _ := store.GetWorkspaceDomains(ctx, "ws-rd")
	if len(domains) != 0 {
		t.Fatalf("expected 0 domains after removal, got %d", len(domains))
	}
}

func TestXormStore_GetWorkspaceConfig(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	config, err := store.GetWorkspaceConfig(ctx, "any-ws")
	if err != nil {
		t.Fatalf("GetWorkspaceConfig failed: %v", err)
	}
	if config.MaxCases != 1000 {
		t.Fatalf("expected MaxCases 1000, got %d", config.MaxCases)
	}
	if !config.EnableCanary {
		t.Fatal("expected EnableCanary=true")
	}
}

func TestXormStore_GetWorkspaceStats(t *testing.T) {
	engine := setupWorkspaceEngine(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	ws := &Workspace{ID: "ws-stats", Name: "Stats WS", OwnerID: "user-1", IsEnabled: true}
	if err := store.CreateWorkspace(ctx, ws); err != nil {
		t.Fatalf("CreateWorkspace failed: %v", err)
	}

	// Add members and domains
	store.AddWorkspaceMember(ctx, &WorkspaceMember{ID: "s1", WorkspaceID: "ws-stats", UserID: "u1", Role: "member"})
	store.AddWorkspaceDomain(ctx, &WorkspaceDomain{ID: "sd1", WorkspaceID: "ws-stats", Domain: "stats.example.com", IsPrimary: true})

	stats, err := store.GetWorkspaceStats(ctx, "ws-stats")
	if err != nil {
		t.Fatalf("GetWorkspaceStats failed: %v", err)
	}
	if stats.MemberCount != 1 {
		t.Fatalf("expected 1 member, got %d", stats.MemberCount)
	}
	if stats.DomainCount != 1 {
		t.Fatalf("expected 1 domain, got %d", stats.DomainCount)
	}
}
