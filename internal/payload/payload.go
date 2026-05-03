package payload

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Variables represents custom variables for payload template rendering
type Variables map[string]string

// Scan implements sql.Scanner interface for Variables
func (v *Variables) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, v)
}

// Value implements driver.Valuer interface for Variables
func (v Variables) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// Payload represents a trackable payload with token and lifecycle
type Payload struct {
	ID                string     `xorm:"pk varchar(36) notnull" json:"id"`
	CaseID            string     `xorm:"varchar(36) notnull index" json:"case_id"`
	Token             string     `xorm:"varchar(64) notnull unique index" json:"token"`
	Template          string     `xorm:"varchar(64) notnull" json:"template"` // SSRF, XXE, RCE, Blind SQLi, etc.
	RenderedPayload   string     `xorm:"text" json:"rendered_payload"`
	Variables         Variables `xorm:"json" json:"variables"`
	Status            string     `xorm:"varchar(32) notnull default('draft') index" json:"status"` // draft, deployed, hit, archived, expired
	ExpectedProtocol  string     `xorm:"varchar(16)" json:"expected_protocol"`                     // dns, http, smtp, ldap
	ExpiresAt         *time.Time `xorm:"datetime" json:"expires_at"`
	CreatedBy         string     `xorm:"varchar(36) notnull" json:"created_by"`
	CreatedAt         time.Time  `xorm:"datetime created" json:"created_at"`
	UpdatedAt         time.Time  `xorm:"datetime updated" json:"updated_at"`
}

// TableName returns the table name for Payload model
func (Payload) TableName() string {
	return "payloads"
}

// PayloadCreateRequest represents the request to create a payload
type PayloadCreateRequest struct {
	CaseID           string            `json:"case_id" binding:"required"`
	Template         string            `json:"template" binding:"required"`
	Variables        map[string]string `json:"variables"`
	ExpiresAt        *time.Time        `json:"expires_at"`
	ExpectedProtocol string            `json:"expected_protocol" binding:"omitempty,oneof=dns http smtp ldap"`
}

// PayloadUpdateRequest represents the request to update a payload
type PayloadUpdateRequest struct {
	Status           string     `json:"status" binding:"omitempty,oneof=draft deployed hit archived expired"`
	ExpiresAt        *time.Time `json:"expires_at"`
	ExpectedProtocol string     `json:"expected_protocol" binding:"omitempty,oneof=dns http smtp ldap"`
}

// PayloadListResponse represents the response for listing payloads
type PayloadListResponse struct {
	Items      []Payload `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}
