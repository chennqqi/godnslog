package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Workflow represents an automated workflow with actions
type Workflow struct {
	ID          string    `json:"id" xorm:"pk varchar(36) notnull"`
	Name        string    `json:"name" xorm:"varchar(255) notnull"`
	Description string    `json:"description" xorm:"text"`
	CaseID      *string   `json:"case_id" xorm:"varchar(36) index"`
	Enabled     bool      `json:"enabled" xorm:"bool notnull default true"`
	Actions     Actions   `json:"actions" xorm:"json notnull"`
	CreatedBy   string    `json:"created_by" xorm:"varchar(36) notnull"`
	CreatedAt   time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"datetime updated"`
}

// TableName returns the table name for Workflow model
func (Workflow) TableName() string {
	return "workflows"
}

// Actions represents a list of workflow actions
type Actions []Action

// Scan implements sql.Scanner interface for Actions
func (a *Actions) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, a)
}

// Value implements driver.Valuer interface for Actions
func (a Actions) Value() (driver.Value, error) {
	return json.Marshal(a)
}

// Action represents a single action in a workflow
type Action struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type" binding:"required,oneof=http dns smtp webhook notify"` // http, dns, smtp, webhook, notify
	Config   map[string]interface{} `json:"config"`
	Enabled  bool                   `json:"enabled"`
	Priority int                    `json:"priority"`
}

// WorkflowListResponse represents the response for listing workflows
type WorkflowListResponse struct {
	Items      []Workflow `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// Action types
const (
	ActionTypeHTTP    = "http"
	ActionTypeDNS     = "dns"
	ActionTypeSMTP    = "smtp"
	ActionTypeWebhook = "webhook"
	ActionTypeNotify  = "notify"
)
