package interaction

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
)

// Headers represents HTTP headers
type Headers map[string]string

// Scan implements sql.Scanner interface for Headers
func (h *Headers) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, h)
}

// Value implements driver.Valuer interface for Headers
func (h Headers) Value() (driver.Value, error) {
	if h == nil {
		return nil, nil
	}
	return json.Marshal(h)
}

// Interaction represents an external connection event (DNS, HTTP, SMTP, LDAP, etc.)
type Interaction struct {
	ID        string    `xorm:"pk varchar(36) notnull" json:"id"`
	Type      string    `xorm:"varchar(16) notnull index" json:"type"` // dns, http, smtp, ldap, smb, ftp
	CaseID    *string   `xorm:"varchar(36) index" json:"case_id"`
	PayloadID *string   `xorm:"varchar(36) index" json:"payload_id"`
	Token     *string   `xorm:"varchar(64) index" json:"token"`
	Timestamp time.Time `xorm:"datetime notnull index" json:"timestamp"`
	SourceIP  string    `xorm:"varchar(64) notnull" json:"source_ip"`
	// DNS specific fields
	Domain  *string `xorm:"varchar(255)" json:"domain"`
	DNSType *string `xorm:"varchar(16)" json:"dns_type"`
	// HTTP specific fields
	Method      *string `xorm:"varchar(16)" json:"method"`
	Path        *string `xorm:"text" json:"path"`
	Headers     Headers `xorm:"json" json:"headers"`
	Body        *string `xorm:"mediumtext" json:"body"`
	UserAgent   *string `xorm:"text" json:"user_agent"`
	ContentType *string `xorm:"varchar(128)" json:"content_type"`
	// Common fields
	RawData   string    `xorm:"text" json:"raw_data"`
	CreatedAt time.Time `xorm:"datetime created" json:"created_at"`
}

// TableName returns the table name for Interaction model
func (Interaction) TableName() string {
	return "interactions"
}

// InteractionListResponse represents the response for listing interactions
type InteractionListResponse struct {
	Items      []Interaction `json:"items"`
	Total      int64         `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

// ExportRequest represents the request to export interactions
type ExportRequest struct {
	Format     string     `json:"format" binding:"required,oneof=json markdown csv"`
	CaseID     *string    `json:"case_id"`
	PayloadID  *string    `json:"payload_id"`
	StartTime  *time.Time `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	IncludeRaw bool       `json:"include_raw"`
}

// DeleteRequest represents the request to delete interactions
type DeleteRequest struct {
	IDs []string `json:"ids" binding:"required"`
}

// TimelineResponse represents the response for timeline query
type TimelineResponse struct {
	Total         int64                `json:"total"`
	Interactions  []models.Interaction `json:"interactions"`
	GroupedEvents []TimelineGroup      `json:"grouped_events"`
}

// TimelineGroup represents a group of interactions in a time interval
type TimelineGroup struct {
	Time         string               `json:"time"`
	Count        int                  `json:"count"`
	Interactions []models.Interaction `json:"interactions"`
}
