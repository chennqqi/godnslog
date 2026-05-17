package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"database/sql/driver"
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

// Payload represents a generated OAST probe template with rendered values, ready for delivery to targets.
// Unified from internal/payload/payload.go and models/v2.go TblPayload
// Aligned with docs/unified-terminology.md
type Payload struct {
	ID               string     `json:"id" xorm:"pk varchar(36) notnull"`
	CaseID           string     `json:"case_id" xorm:"varchar(36) notnull index"`
	Token            string     `json:"token" xorm:"varchar(64) notnull unique index"`
	TemplateID       string     `json:"template_id" xorm:"varchar(64) notnull"`                    // Template identifier (e.g., ssrf-basic, xxe-basic)
	TemplateRendered string     `json:"template_rendered" xorm:"text"`                             // Rendered payload with variables substituted
	Variables        Variables  `json:"variables" xorm:"json"`                                     // Custom variable values used in rendering
	Status           string     `json:"status" xorm:"varchar(32) notnull default('active') index"` // active, expired, revoked
	ExpectedProtocol string     `json:"expected_protocol" xorm:"varchar(16)"`                      // dns, http, smtp, ldap
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
		ExpiresAt string `json:"expires_at,omitempty"`
	}{
		Alias:     (*Alias)(p),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt,
	})
}

// TableName returns the table name for Payload model
func (Payload) TableName() string {
	return "payloads"
}

// Status constants
const (
	PayloadStatusActive  = "active"
	PayloadStatusExpired = "expired"
	PayloadStatusRevoked = "revoked"
)

// PayloadTemplates defines available payload templates
var PayloadTemplates = map[string]string{
	"ssrf-basic":    "http://{token}.{domain}/",
	"xxe-basic":     "http://{token}.{domain}/xxe",
	"rce-basic":     "http://{token}.{domain}/cmd",
	"blind-sqli":    "http://{token}.{domain}/sql?id=1",
	"dns-rebinding": "http://{token}.{domain}/rebind",
}

// RenderTemplate renders a payload template with variables
// Aligned with docs/unified-terminology.md
// Uses simple {variable} substitution instead of Go template syntax
func RenderTemplate(tmpl string, variables map[string]string, token, domain string) (string, error) {
	// Add default variables
	if variables == nil {
		variables = make(map[string]string)
	}
	variables["token"] = token
	variables["domain"] = domain
	variables["callback_url"] = fmt.Sprintf("http://%s/log/%s/", domain, token)

	// Simple substitution for {variable} syntax
	result := tmpl
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// RenderTemplateWithCase renders a payload template with case variable
// Aligned with docs/unified-terminology.md
// Uses simple {variable} substitution instead of Go template syntax
func RenderTemplateWithCase(tmpl string, variables map[string]string, token, domain, caseID string) (string, error) {
	// Add default variables
	if variables == nil {
		variables = make(map[string]string)
	}
	variables["token"] = token
	variables["domain"] = domain
	variables["case"] = caseID
	variables["callback_url"] = fmt.Sprintf("http://%s/log/%s/", domain, token)

	// Simple substitution for {variable} syntax
	result := tmpl
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// PayloadCreateRequest represents the request to create a payload
// Aligned with docs/unified-terminology.md
type PayloadCreateRequest struct {
	CaseID           string            `json:"case_id" binding:"required"`
	TemplateID       string            `json:"template_id" binding:"required"` // Template identifier (e.g., ssrf-basic, xxe-basic)
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
