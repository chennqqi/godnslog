package workspace

import (
	"context"
	"time"

	"xorm.io/xorm"
)

// Store defines the interface for workspace storage
type Store interface {
	// Workspace operations
	CreateWorkspace(ctx context.Context, workspace *Workspace) error
	GetWorkspace(ctx context.Context, id string) (*Workspace, error)
	GetWorkspaceByOwner(ctx context.Context, ownerID string) ([]Workspace, error)
	GetAllWorkspaces(ctx context.Context) ([]Workspace, error)
	UpdateWorkspace(ctx context.Context, workspace *Workspace) error
	DeleteWorkspace(ctx context.Context, id string) error

	// Workspace member operations
	AddWorkspaceMember(ctx context.Context, member *WorkspaceMember) error
	GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error)
	GetWorkspaceMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error)
	UpdateWorkspaceMember(ctx context.Context, member *WorkspaceMember) error
	RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error

	// Workspace domain operations
	AddWorkspaceDomain(ctx context.Context, domain *WorkspaceDomain) error
	GetWorkspaceDomains(ctx context.Context, workspaceID string) ([]WorkspaceDomain, error)
	GetPrimaryDomain(ctx context.Context, workspaceID string) (*WorkspaceDomain, error)
	UpdateWorkspaceDomain(ctx context.Context, domain *WorkspaceDomain) error
	RemoveWorkspaceDomain(ctx context.Context, id string) error

	// Workspace config operations
	GetWorkspaceConfig(ctx context.Context, workspaceID string) (*WorkspaceConfig, error)
	UpdateWorkspaceConfig(ctx context.Context, workspaceID string, config *WorkspaceConfig) error

	// Workspace stats
	GetWorkspaceStats(ctx context.Context, workspaceID string) (*WorkspaceStats, error)
}

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreateWorkspace creates a new workspace
func (s *XormStore) CreateWorkspace(ctx context.Context, workspace *Workspace) error {
	workspace.CreatedAt = time.Now()
	workspace.UpdatedAt = time.Now()
	_, err := s.engine.Insert(workspace)
	return err
}

// GetWorkspace retrieves a workspace by ID
func (s *XormStore) GetWorkspace(ctx context.Context, id string) (*Workspace, error) {
	var workspace Workspace
	_, err := s.engine.ID(id).Get(&workspace)
	if err != nil {
		return nil, err
	}
	return &workspace, nil
}

// GetWorkspaceByOwner retrieves workspaces by owner
func (s *XormStore) GetWorkspaceByOwner(ctx context.Context, ownerID string) ([]Workspace, error) {
	var workspaces []Workspace
	err := s.engine.Where("owner_id = ?", ownerID).Find(&workspaces)
	return workspaces, err
}

// GetAllWorkspaces retrieves all workspaces
func (s *XormStore) GetAllWorkspaces(ctx context.Context) ([]Workspace, error) {
	var workspaces []Workspace
	err := s.engine.Find(&workspaces)
	return workspaces, err
}

// UpdateWorkspace updates a workspace
func (s *XormStore) UpdateWorkspace(ctx context.Context, workspace *Workspace) error {
	workspace.UpdatedAt = time.Now()
	_, err := s.engine.ID(workspace.ID).Update(workspace)
	return err
}

// DeleteWorkspace deletes a workspace
func (s *XormStore) DeleteWorkspace(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&Workspace{})
	return err
}

// AddWorkspaceMember adds a member to a workspace
func (s *XormStore) AddWorkspaceMember(ctx context.Context, member *WorkspaceMember) error {
	member.JoinedAt = time.Now()
	_, err := s.engine.Insert(member)
	return err
}

// GetWorkspaceMembers retrieves all members of a workspace
func (s *XormStore) GetWorkspaceMembers(ctx context.Context, workspaceID string) ([]WorkspaceMember, error) {
	var members []WorkspaceMember
	err := s.engine.Where("workspace_id = ?", workspaceID).Find(&members)
	return members, err
}

// GetWorkspaceMember retrieves a specific workspace member
func (s *XormStore) GetWorkspaceMember(ctx context.Context, workspaceID, userID string) (*WorkspaceMember, error) {
	var member WorkspaceMember
	_, err := s.engine.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Get(&member)
	if err != nil {
		return nil, err
	}
	return &member, nil
}

// UpdateWorkspaceMember updates a workspace member
func (s *XormStore) UpdateWorkspaceMember(ctx context.Context, member *WorkspaceMember) error {
	_, err := s.engine.ID(member.ID).Update(member)
	return err
}

// RemoveWorkspaceMember removes a member from a workspace
func (s *XormStore) RemoveWorkspaceMember(ctx context.Context, workspaceID, userID string) error {
	_, err := s.engine.Where("workspace_id = ? AND user_id = ?", workspaceID, userID).Delete(&WorkspaceMember{})
	return err
}

// AddWorkspaceDomain adds a domain to a workspace
func (s *XormStore) AddWorkspaceDomain(ctx context.Context, domain *WorkspaceDomain) error {
	domain.CreatedAt = time.Now()
	_, err := s.engine.Insert(domain)
	return err
}

// GetWorkspaceDomains retrieves all domains of a workspace
func (s *XormStore) GetWorkspaceDomains(ctx context.Context, workspaceID string) ([]WorkspaceDomain, error) {
	var domains []WorkspaceDomain
	err := s.engine.Where("workspace_id = ?", workspaceID).Find(&domains)
	return domains, err
}

// GetPrimaryDomain retrieves the primary domain of a workspace
func (s *XormStore) GetPrimaryDomain(ctx context.Context, workspaceID string) (*WorkspaceDomain, error) {
	var domain WorkspaceDomain
	_, err := s.engine.Where("workspace_id = ? AND is_primary = ?", workspaceID, true).Get(&domain)
	if err != nil {
		return nil, err
	}
	return &domain, nil
}

// UpdateWorkspaceDomain updates a workspace domain
func (s *XormStore) UpdateWorkspaceDomain(ctx context.Context, domain *WorkspaceDomain) error {
	_, err := s.engine.ID(domain.ID).Update(domain)
	return err
}

// RemoveWorkspaceDomain removes a domain from a workspace
func (s *XormStore) RemoveWorkspaceDomain(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&WorkspaceDomain{})
	return err
}

// GetWorkspaceConfig retrieves workspace configuration
func (s *XormStore) GetWorkspaceConfig(ctx context.Context, workspaceID string) (*WorkspaceConfig, error) {
	// For MVP, return default config
	// In production, store config in a separate table
	return &WorkspaceConfig{
		MaxCases:        1000,
		MaxPayloads:     10000,
		MaxInteractions: 100000,
		RetentionDays:   90,
		EnableCanary:    true,
		EnableRebinding: true,
		EnableListeners: true,
	}, nil
}

// UpdateWorkspaceConfig updates workspace configuration
func (s *XormStore) UpdateWorkspaceConfig(ctx context.Context, workspaceID string, config *WorkspaceConfig) error {
	// For MVP, config is not persisted
	// In production, store config in a separate table
	return nil
}

// GetWorkspaceStats retrieves workspace statistics
func (s *XormStore) GetWorkspaceStats(ctx context.Context, workspaceID string) (*WorkspaceStats, error) {
	stats := &WorkspaceStats{
		WorkspaceID: workspaceID,
	}
	
	// Count members
	members, err := s.GetWorkspaceMembers(ctx, workspaceID)
	if err == nil {
		stats.MemberCount = len(members)
	}
	
	// Count domains
	domains, err := s.GetWorkspaceDomains(ctx, workspaceID)
	if err == nil {
		stats.DomainCount = len(domains)
	}
	
	// For MVP, case/payload/interaction counts would require joins
	// In production, implement proper counting
	
	return stats, nil
}
