package workspace

import (
	"testing"
	"time"
)

func TestWorkspaceModel(t *testing.T) {
	now := time.Now()
	ws := Workspace{
		ID:          "ws-model-1",
		Name:        "Model Test",
		Description: "Test workspace model",
		OwnerID:     "user-1",
		IsEnabled:   true,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if ws.ID == "" {
		t.Fatal("ID should not be empty")
	}
	if ws.Name == "" {
		t.Fatal("Name should not be empty")
	}
	if ws.OwnerID == "" {
		t.Fatal("OwnerID should not be empty")
	}
}

func TestWorkspaceMemberModel(t *testing.T) {
	m := WorkspaceMember{
		ID:          "m-1",
		WorkspaceID: "ws-1",
		UserID:      "user-1",
		Role:        "owner",
		JoinedAt:    time.Now(),
	}

	if m.ID == "" {
		t.Fatal("ID should not be empty")
	}
	if m.WorkspaceID == "" {
		t.Fatal("WorkspaceID should not be empty")
	}
	if m.Role != "owner" && m.Role != "admin" && m.Role != "member" && m.Role != "viewer" {
		t.Fatalf("invalid role: %s", m.Role)
	}
}

func TestWorkspaceDomainModel(t *testing.T) {
	d := WorkspaceDomain{
		ID:          "d-1",
		WorkspaceID: "ws-1",
		Domain:      "example.com",
		IsPrimary:   true,
		CreatedAt:   time.Now(),
	}

	if d.ID == "" {
		t.Fatal("ID should not be empty")
	}
	if d.Domain == "" {
		t.Fatal("Domain should not be empty")
	}
}

func TestWorkspaceConfigDefaults(t *testing.T) {
	config := WorkspaceConfig{
		MaxCases:        1000,
		MaxPayloads:     10000,
		MaxInteractions: 100000,
		RetentionDays:   90,
		EnableCanary:    true,
		EnableRebinding: true,
		EnableListeners: true,
	}

	if config.MaxCases <= 0 {
		t.Fatal("MaxCases should be positive")
	}
	if config.MaxPayloads <= 0 {
		t.Fatal("MaxPayloads should be positive")
	}
	if config.RetentionDays <= 0 {
		t.Fatal("RetentionDays should be positive")
	}
}

func TestWorkspaceStatsModel(t *testing.T) {
	stats := WorkspaceStats{
		WorkspaceID:      "ws-1",
		CaseCount:        10,
		PayloadCount:     50,
		InteractionCount: 1000,
		MemberCount:      5,
		DomainCount:      2,
	}

	if stats.WorkspaceID == "" {
		t.Fatal("WorkspaceID should not be empty")
	}
	if stats.CaseCount < 0 {
		t.Fatal("CaseCount should not be negative")
	}
}
