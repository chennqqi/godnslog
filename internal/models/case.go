package models

import (
	"time"
)

// Case represents a test task, vulnerability verification, or project
// Unified from internal/case/case.go and models/v2.go TblCase
type Case struct {
	ID          string    `json:"id" xorm:"pk varchar(36) notnull"`
	Title       string    `json:"title" xorm:"varchar(255) notnull"`
	Description string    `json:"description" xorm:"text"`
	Target      string    `json:"target" xorm:"varchar(255)"`
	Status      string    `json:"status" xorm:"varchar(32) notnull default('active') index"` // active, archived, completed
	Tags        string    `json:"tags" xorm:"varchar(500)"` // JSON array
	CreatedBy   string    `json:"created_by" xorm:"varchar(36) notnull index"` // User ID
	CreatedAt   time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"datetime updated"`
}

// TableName returns the table name for Case model
func (Case) TableName() string {
	return "cases"
}

// Status constants
const (
	CaseStatusActive    = "active"
	CaseStatusArchived  = "archived"
	CaseStatusCompleted = "completed"
)

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
