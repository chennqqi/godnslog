package models

import "time"

// Canary represents a canary token for long-term monitoring
type Canary struct {
	ID          string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	Type        string    `json:"type" xorm:"varchar(16) notnull index"` // dns, http, document, config, ci, storage, email
	Token       string    `json:"token" xorm:"varchar(255) notnull"`
	Description string    `json:"description" xorm:"text"`
	Context     string    `json:"context" xorm:"text"` // Encoded context (project, asset, location, owner)
	ExpiresAt   time.Time `json:"expires_at" xorm:"datetime"`
	IsEnabled   bool      `json:"is_enabled" xorm:"bool notnull default true"`
	CreatedAt   time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"datetime updated"`
}

// CanaryHit represents a hit on a canary token
type CanaryHit struct {
	ID           string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	CanaryID     string    `json:"canary_id" xorm:"varchar(36) notnull index"`
	SourceIP     string    `json:"source_ip" xorm:"varchar(64) notnull"`
	UserAgent    string    `json:"user_agent" xorm:"text"`
	Headers      string    `json:"headers" xorm:"mediumtext"` // JSON string
	Body         string    `json:"body" xorm:"mediumtext"`
	Timestamp    time.Time `json:"timestamp" xorm:"datetime notnull created"`
	IsCompressed bool     `json:"is_compressed" xorm:"bool notnull default false"`
}

// CanaryConfig holds canary configuration
type CanaryConfig struct {
	MaxRetentionDays      int      `json:"max_retention_days"` // How long to keep canary tokens
	DefaultExpiry         string   `json:"default_expiry"`     // Default expiration (e.g., "90d")
	SilentWindow          int      `json:"silent_window"`      // Silent window in seconds
	CompressionThreshold  int      `json:"compression_threshold"` // Hits before compression
	NotificationLevels    []string `json:"notification_levels"` // low, medium, high, critical
}

// CanaryType represents the type of canary
type CanaryType string

const (
	CanaryTypeDNS      CanaryType = "dns"
	CanaryTypeHTTP     CanaryType = "http"
	CanaryTypeDocument CanaryType = "document"
	CanaryTypeConfig   CanaryType = "config"
	CanaryTypeCI       CanaryType = "ci"
	CanaryTypeStorage  CanaryType = "storage"
	CanaryTypeEmail    CanaryType = "email"
)

// CanaryContext represents decoded canary context
type CanaryContext struct {
	Project  string `json:"project"`
	Asset    string `json:"asset"`
	Location string `json:"location"`
	Owner    string `json:"owner"`
	Purpose  string `json:"purpose"`
}

// TableName returns the table name for Canary
func (Canary) TableName() string {
	return "canaries"
}

// TableName returns the table name for CanaryHit
func (CanaryHit) TableName() string {
	return "canary_hits"
}

// CanaryListResponse represents the response for listing canaries
type CanaryListResponse struct {
	Items      []Canary `json:"items"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// CanaryHitListResponse represents the response for listing canary hits
type CanaryHitListResponse struct {
	Items      []CanaryHit `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
