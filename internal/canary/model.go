package canary

import "time"

// Canary represents a canary token for long-term monitoring
type Canary struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Type        string    `json:"type"`        // dns, http, document, config, ci, storage, email
	Token       string    `json:"token"`
	Description string    `json:"description"`
	Context     string    `json:"context"`     // Encoded context (project, asset, location, owner)
	ExpiresAt   time.Time `json:"expires_at"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
}

// CanaryHit represents a hit on a canary token
type CanaryHit struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	CanaryID    string    `json:"canary_id"`
	SourceIP    string    `json:"source_ip"`
	UserAgent   string    `json:"user_agent"`
	Headers     string    `json:"headers"`     // JSON string
	Body        string    `json:"body"`
	Timestamp   time.Time `json:"timestamp" xorm:"created"`
	IsCompressed bool     `json:"is_compressed"`
}

// CanaryConfig holds canary configuration
type CanaryConfig struct {
	MaxRetentionDays int    `json:"max_retention_days"` // How long to keep canary tokens
	DefaultExpiry    string `json:"default_expiry"`     // Default expiration (e.g., "90d")
	SilentWindow     int    `json:"silent_window"`      // Silent window in seconds
	CompressionThreshold int `json:"compression_threshold"` // Hits before compression
	NotificationLevels []string `json:"notification_levels"` // low, medium, high, critical
}

// CanaryType represents the type of canary
type CanaryType string

const (
	CanaryTypeDNS       CanaryType = "dns"
	CanaryTypeHTTP      CanaryType = "http"
	CanaryTypeDocument  CanaryType = "document"
	CanaryTypeConfig    CanaryType = "config"
	CanaryTypeCI        CanaryType = "ci"
	CanaryTypeStorage   CanaryType = "storage"
	CanaryTypeEmail     CanaryType = "email"
)

// CanaryContext represents decoded canary context
type CanaryContext struct {
	Project   string `json:"project"`
	Asset     string `json:"asset"`
	Location  string `json:"location"`
	Owner     string `json:"owner"`
	Purpose   string `json:"purpose"`
}

// TableName returns the table name for Canary
func (Canary) TableName() string {
	return "canaries"
}

// TableName returns the table name for CanaryHit
func (CanaryHit) TableName() string {
	return "canary_hits"
}
