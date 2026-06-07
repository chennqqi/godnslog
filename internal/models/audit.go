package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AuditDetails represents additional audit log details
type AuditDetails map[string]interface{}

// Scan implements sql.Scanner interface for AuditDetails
func (d *AuditDetails) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, d)
}

// Value implements driver.Valuer interface for AuditDetails
func (d AuditDetails) Value() (driver.Value, error) {
	if d == nil {
		return nil, nil
	}
	return json.Marshal(d)
}

// AuditLog represents an audit log entry
// Unified from internal/auth/audit.go and internal/audit/audit.go
type AuditLog struct {
	ID           string       `json:"id" xorm:"'id' pk varchar(36) notnull"`
	UserID       *string      `json:"user_id" xorm:"'user_id' varchar(36) index"`                     // User who performed the action
	APIKeyID     *string      `json:"api_key_id" xorm:"'api_key_id' varchar(36) index"`               // API key used (if applicable)
	APIKeyPrefix *string      `json:"api_key_prefix" xorm:"'api_key_prefix' varchar(16) index"`       // API key prefix
	IsAgent      bool         `json:"is_agent" xorm:"'is_agent' bool notnull index"`                  // Whether this is an agent action
	Action       string       `json:"action" xorm:"'action' varchar(64) notnull index"`               // create, update, delete, login, logout, create_case, create_payload, etc.
	ResourceType string       `json:"resource_type" xorm:"'resource_type' varchar(64) notnull index"` // case, payload, interaction, apikey, user
	ResourceID   *string      `json:"resource_id" xorm:"'resource_id' varchar(36) index"`             // Resource ID
	Parameters   string       `json:"parameters" xorm:"'parameters' text"`                            // JSON string of parameters
	Result       string       `json:"result" xorm:"'result' varchar(32)"`                             // success, failure
	ErrorMessage string       `json:"error_message" xorm:"'error_message' text"`                      // Error message if failed
	IPAddress    string       `json:"ip_address" xorm:"'ip_address' varchar(64)"`                     // IP address of the request
	UserAgent    string       `json:"user_agent" xorm:"'user_agent' text"`                            // User agent string
	Details      AuditDetails `json:"details" xorm:"'details' json"`                                  // Additional details as JSON
	Timestamp    time.Time    `json:"timestamp" xorm:"'timestamp' datetime notnull index"`
	CreatedAt    time.Time    `json:"created_at" xorm:"'created_at' datetime created"`
}

// TableName returns the table name for AuditLog model
func (AuditLog) TableName() string {
	return "audit_logs"
}

// AuditLogListResponse represents the response for listing audit logs
type AuditLogListResponse struct {
	Items      []AuditLog `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
