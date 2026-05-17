package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Tags represents a list of tags stored as JSON in database.
// It serializes to a JSON array in API responses.
type Tags []string

// Scan implements sql.Scanner interface for Tags.
func (t *Tags) Scan(value interface{}) error {
	if value == nil {
		*t = nil
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, t)
}

// Value implements driver.Valuer interface for Tags.
func (t Tags) Value() (driver.Value, error) {
	if t == nil {
		return nil, nil
	}
	return json.Marshal(t)
}

// Case represents a security engagement or testing session that groups related OAST activities.
// This struct serves as both the database entity and the API response model.
type Case struct {
	ID          string    `json:"id" xorm:"pk varchar(36) notnull"`
	Title       string    `json:"title" xorm:"varchar(255) notnull"`
	Description string    `json:"description" xorm:"text"`
	Status      string    `json:"status" xorm:"varchar(32) notnull default('active') index"` // active, completed, archived
	Tags        Tags      `json:"tags" xorm:"json"`                                          // Stored as JSON array
	CreatedBy   string    `json:"created_by" xorm:"varchar(36) notnull index"`               // User ID
	CreatedAt   time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"datetime updated"`
}

// MarshalJSON implements json.Marshaler interface for Case
func (c *Case) MarshalJSON() ([]byte, error) {
	type Alias Case
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(c),
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
		UpdatedAt: c.UpdatedAt.Format(time.RFC3339),
	})
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
