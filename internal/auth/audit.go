package auth

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
type AuditLog struct {
	ID           string        `xorm:"pk varchar(36) notnull" json:"id"`
	UserID       string        `xorm:"varchar(36) notnull index" json:"user_id"`
	Action       string        `xorm:"varchar(64) notnull index" json:"action"`       // create, update, delete, login, logout, etc.
	ResourceType string        `xorm:"varchar(64) notnull index" json:"resource_type"` // case, payload, interaction, apikey, user
	ResourceID   *string       `xorm:"varchar(36) index" json:"resource_id"`
	IPAddress    string        `xorm:"varchar(64)" json:"ip_address"`
	UserAgent    string        `xorm:"text" json:"user_agent"`
	Details      AuditDetails  `xorm:"json" json:"details"`
	Timestamp    time.Time     `xorm:"datetime notnull index" json:"timestamp"`
	CreatedAt    time.Time     `xorm:"datetime created" json:"created_at"`
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
