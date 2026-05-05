package models

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
// Unified from internal/payload/payload.go and models/v2.go TblPayload
type Payload struct {
	ID               string     `json:"id" xorm:"pk varchar(36) notnull"`
	CaseID           string     `json:"case_id" xorm:"varchar(36) notnull index"`
	Token            string     `json:"token" xorm:"varchar(64) notnull unique index"`
	Template         string     `json:"template" xorm:"varchar(64) notnull"` // SSRF, XXE, RCE, Blind SQLi, etc.
	RenderedPayload  string     `json:"rendered_payload" xorm:"text"`
	Variables        Variables  `json:"variables" xorm:"json"`
	Status           string     `json:"status" xorm:"varchar(32) notnull default('draft') index"` // draft, deployed, hit, archived, expired
	ExpectedProtocol string     `json:"expected_protocol" xorm:"varchar(16)"`                     // dns, http, smtp, ldap
	ExpiresAt        *time.Time `json:"expires_at" xorm:"datetime"`
	CreatedBy        string     `json:"created_by" xorm:"varchar(36) notnull"`
	CreatedAt        time.Time  `json:"created_at" xorm:"datetime created"`
	UpdatedAt        time.Time  `json:"updated_at" xorm:"datetime updated"`
}

// MarshalJSON implements json.Marshaler interface for Payload
func (p *Payload) MarshalJSON() ([]byte, error) {
	type Alias Payload
	expiresAt := ""
	if p.ExpiresAt != nil {
		expiresAt = p.ExpiresAt.Format(time.RFC3339)
	}
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
		ExpiresAt string `json:"expires_at,omitempty"`
	}{
		Alias:     (*Alias)(p),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		UpdatedAt: p.UpdatedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt,
	})
}

// TableName returns the table name for Payload model
func (Payload) TableName() string {
	return "payloads"
}

// Status constants
const (
	PayloadStatusDraft    = "draft"
	PayloadStatusDeployed = "deployed"
	PayloadStatusHit      = "hit"
	PayloadStatusArchived = "archived"
	PayloadStatusExpired  = "expired"
)

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
