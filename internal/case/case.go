package casemgmt

import (
	"time"
)

// Case represents a test task, vulnerability verification, or project
type Case struct {
	ID          string    `xorm:"pk varchar(36) notnull" json:"id"`
	Title       string    `xorm:"varchar(255) notnull" json:"title"`
	Description string    `xorm:"text" json:"description"`
	Target      string    `xorm:"varchar(255)" json:"target"`
	Status      string    `xorm:"varchar(32) notnull default('active') index" json:"status"` // active, archived, completed
	Tags        string    `xorm:"varchar(500)" json:"tags"`                                  // JSON array
	CreatedBy   string    `xorm:"varchar(36) notnull index" json:"created_by"`
	CreatedAt   time.Time `xorm:"datetime created" json:"created_at"`
	UpdatedAt   time.Time `xorm:"datetime updated" json:"updated_at"`
}

// TableName returns the table name for Case model
func (Case) TableName() string {
	return "cases"
}

// CaseCreateRequest represents the request to create a case
type CaseCreateRequest struct {
	Title       string   `json:"title" binding:"required"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Tags        []string `json:"tags"`
}

// CaseUpdateRequest represents the request to update a case
type CaseUpdateRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Status      string   `json:"status" binding:"omitempty,oneof=active archived completed"`
	Tags        []string `json:"tags"`
}

// CaseListResponse represents the response for listing cases
type CaseListResponse struct {
	Items      []Case `json:"items"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

// CaseStats represents statistics for a case
type CaseStats struct {
	PayloadCount     int `json:"payload_count"`
	InteractionCount int `json:"interaction_count"`
	HitPayloadCount  int `json:"hit_payload_count"`
}
