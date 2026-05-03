package workspace

import "time"

// Workspace represents a multi-tenant workspace
type Workspace struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     string    `json:"owner_id"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
}

// WorkspaceMember represents a workspace member
type WorkspaceMember struct {
	ID         string    `json:"id" xorm:"'id' pk"`
	WorkspaceID string   `json:"workspace_id"`
	UserID     string    `json:"user_id"`
	Role       string    `json:"role"` // owner, admin, member, viewer
	JoinedAt   time.Time `json:"joined_at" xorm:"created"`
}

// WorkspaceDomain represents a domain associated with a workspace
type WorkspaceDomain struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	WorkspaceID string    `json:"workspace_id"`
	Domain      string    `json:"domain"`
	IsPrimary   bool      `json:"is_primary"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
}

// WorkspaceConfig holds workspace configuration
type WorkspaceConfig struct {
	MaxCases       int    `json:"max_cases"`
	MaxPayloads    int    `json:"max_payloads"`
	MaxInteractions int   `json:"max_interactions"`
	RetentionDays  int    `json:"retention_days"`
	EnableCanary   bool   `json:"enable_canary"`
	EnableRebinding bool  `json:"enable_rebinding"`
	EnableListeners bool   `json:"enable_listeners"`
}

// WorkspaceStats represents workspace statistics
type WorkspaceStats struct {
	WorkspaceID      string `json:"workspace_id"`
	CaseCount        int    `json:"case_count"`
	PayloadCount     int    `json:"payload_count"`
	InteractionCount int    `json:"interaction_count"`
	MemberCount      int    `json:"member_count"`
	DomainCount      int    `json:"domain_count"`
}

// TableName returns the table name for Workspace
func (Workspace) TableName() string {
	return "workspaces"
}

// TableName returns the table name for WorkspaceMember
func (WorkspaceMember) TableName() string {
	return "workspace_members"
}

// TableName returns the table name for WorkspaceDomain
func (WorkspaceDomain) TableName() string {
	return "workspace_domains"
}
