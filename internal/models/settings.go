package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Settings represents system configuration
type Settings struct {
	ID        string    `json:"id" xorm:"pk varchar(36) notnull"`
	Key       string    `json:"key" xorm:"varchar(128) notnull unique index"`
	Value     string    `json:"value" xorm:"text"`
	UpdatedAt  time.Time `json:"updated_at" xorm:"datetime updated"`
	CreatedAt  time.Time `json:"created_at" xorm:"datetime created"`
}

// TableName returns the table name for Settings model
func (Settings) TableName() string {
	return "settings"
}

// SettingsValue represents a typed settings value
type SettingsValue struct {
	// DNS Settings
	DNSDomain       string `json:"dns_domain,omitempty"`
	DNSPort         int    `json:"dns_port,omitempty"`
	DNSTTL          int    `json:"dns_ttl,omitempty"`
	
	// HTTP Settings
	HTTPPort        int    `json:"http_port,omitempty"`
	HTTPSTLSCert    string `json:"https_tls_cert,omitempty"`
	HTTPSTLSKey     string `json:"https_tls_key,omitempty"`
	
	// Security Settings
	EnableAuth      bool   `json:"enable_auth,omitempty"`
	SessionTimeout  int    `json:"session_timeout,omitempty"`
	
	// Notification Settings
	EnableNotification bool   `json:"enable_notification,omitempty"`
	NotificationURL    string `json:"notification_url,omitempty"`
	
	// Logging Settings
	LogLevel        string `json:"log_level,omitempty"`
	LogRetentionDays int   `json:"log_retention_days,omitempty"`
}

// SettingsCreateRequest represents the request to create/update settings
type SettingsCreateRequest struct {
	Key   string `json:"key" binding:"required"`
	Value string `json:"value" binding:"required"`
}

// SettingsUpdateRequest represents the request to update settings
type SettingsUpdateRequest struct {
	Value string `json:"value" binding:"required"`
}

// SettingsListResponse represents the response for listing settings
type SettingsListResponse struct {
	Items      []Settings `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}

// Scan implements sql.Scanner interface for SettingsValue
func (sv *SettingsValue) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, sv)
}

// Value implements driver.Valuer interface for SettingsValue
func (sv SettingsValue) Value() (driver.Value, error) {
	return json.Marshal(sv)
}
