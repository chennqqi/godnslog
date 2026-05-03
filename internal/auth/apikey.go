package auth

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
type APIKey struct {
	ID           string     `xorm:"pk varchar(36) notnull" json:"id"`
	Key          string     `xorm:"varchar(128) notnull unique" json:"key"` // Only shown on creation
	KeyPrefix    string     `xorm:"varchar(16) notnull index" json:"key_prefix"`
	Name         string     `xorm:"varchar(128) notnull" json:"name"`
	Scopes       Scopes     `xorm:"json" json:"scopes"`
	ExpiresAt    *time.Time `xorm:"datetime" json:"expires_at"`
	LastUsedAt   *time.Time `xorm:"datetime" json:"last_used_at"`
	CreatedBy    string     `xorm:"varchar(36) notnull" json:"created_by"`
	CreatedAt    time.Time  `xorm:"datetime created" json:"created_at"`
	RevokedAt    *time.Time `xorm:"datetime" json:"revoked_at"`
	IsRevoked    bool       `xorm:"bool notnull default(false)" json:"is_revoked"`
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
