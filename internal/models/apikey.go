package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Scopes represents API key scopes
type Scopes []string

// Scan implements sql.Scanner interface for Scopes
func (s *Scopes) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, s)
}

// Value implements driver.Valuer interface for Scopes
func (s Scopes) Value() (driver.Value, error) {
	if s == nil {
		return nil, nil
	}
	return json.Marshal(s)
}

// APIKey represents an API key for programmatic access
// Unified from internal/auth/apikey.go and models/v2.go TblAPIKey
type APIKey struct {
	ID         string     `json:"id" xorm:"pk varchar(36) notnull"`
	Key        string     `json:"key" xorm:"varchar(128) notnull unique"` // Only shown on creation
	KeyPrefix  string     `json:"key_prefix" xorm:"varchar(16) notnull index"`
	Name       string     `json:"name" xorm:"varchar(128) notnull"`
	Scopes     Scopes     `json:"scopes" xorm:"json"`
	ExpiresAt  *time.Time `json:"expires_at" xorm:"datetime"`
	LastUsedAt *time.Time `json:"last_used_at" xorm:"datetime"`
	CreatedBy  string     `json:"created_by" xorm:"varchar(36) notnull"`
	CreatedAt  time.Time  `json:"created_at" xorm:"datetime created"`
	RevokedAt  *time.Time `json:"revoked_at" xorm:"datetime"`
	IsRevoked  bool       `json:"is_revoked" xorm:"bool notnull default(false)"`
}

// TableName returns the table name for APIKey model
func (APIKey) TableName() string {
	return "api_keys"
}

// APIKeyCreateRequest represents the request to create an API key
type APIKeyCreateRequest struct {
	Name      string   `json:"name" binding:"required"`
	Scopes    []string `json:"scopes" binding:"required"`
	ExpiresAt *time.Time `json:"expires_at"`
}

// APIKeyListResponse represents the response for listing API keys
type APIKeyListResponse struct {
	Items      []APIKey `json:"items"`
	Total      int64    `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// IsValid checks if the API key is valid (not revoked and not expired)
func (k *APIKey) IsValid() bool {
	if k.IsRevoked {
		return false
	}
	if k.ExpiresAt != nil && time.Now().After(*k.ExpiresAt) {
		return false
	}
	return true
}

// HasScope checks if the API key has the specified scope
func (k *APIKey) HasScope(scope string) bool {
	for _, s := range k.Scopes {
		if s == scope || s == "admin:all" {
			return true
		}
	}
	return false
}
