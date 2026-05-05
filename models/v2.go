package models

import (
	"time"
)

// WARNING: This file contains v2 database models AND API response/request types.
// The non-Tbl-prefixed API types (Case, Payload, Interaction, APIKey, etc.)
// are REDUNDANT with types defined in internal/models/.
// TODO: Migrate v2_api.go to use internal/models types and remove API types from this file.
// The Tbl-prefixed database models (TblCase, TblPayload, TblInteraction, TblAPIKey)
// are kept here for backward compatibility with existing database tables.

// v2 database models

// TblCase represents a case/vulnerability project
type TblCase struct {
	Id          int64     `xorm:"pk autoincr"`
	Title       string    `xorm:"varchar(255) notnull"`
	Description string    `xorm:"text"`
	Target      string    `xorm:"varchar(255)"`
	Status      string    `xorm:"varchar(32) default('active') notnull"` // active, archived, completed
	Tags        string    `xorm:"text"`                                  // JSON array
	CreatedBy   int64     `xorm:"notnull"`
	CreatedAt   time.Time `xorm:"datetime created"`
	UpdatedAt   time.Time `xorm:"datetime updated"`
}

// TblPayload represents a payload token
type TblPayload struct {
	Id               int64     `xorm:"pk autoincr"`
	CaseId           int64     `xorm:"index"`
	Token            string    `xorm:"varchar(128) notnull unique"`
	Template         string    `xorm:"varchar(64) notnull"`
	RenderedPayload  string    `xorm:"text"`
	Variables        string    `xorm:"json"`                                 // JSON object
	Status           string    `xorm:"varchar(32) default('draft') notnull"` // draft, deployed, hit, archived, expired
	ExpectedProtocol string    `xorm:"varchar(32)"`                          // dns, http, smtp, ldap
	ExpiresAt        time.Time `xorm:"datetime"`
	CreatedBy        int64     `xorm:"notnull"`
	CreatedAt        time.Time `xorm:"datetime created"`
	UpdatedAt        time.Time `xorm:"datetime updated"`
}

// TblInteraction represents an interaction (DNS/HTTP callback)
type TblInteraction struct {
	Id          int64     `xorm:"pk autoincr"`
	Type        string    `xorm:"varchar(32) notnull"` // dns, http, smtp, ldap, smb, ftp
	CaseId      int64     `xorm:"index"`
	PayloadId   int64     `xorm:"index"`
	Token       string    `xorm:"varchar(128) index"`
	Timestamp   time.Time `xorm:"datetime notnull"`
	SourceIp    string    `xorm:"varchar(64) notnull"`
	Domain      string    `xorm:"varchar(255)"`
	DnsType     string    `xorm:"varchar(16)"` // A, AAAA, CNAME, etc.
	Method      string    `xorm:"varchar(16)"` // GET, POST, etc.
	Path        string    `xorm:"text"`
	Headers     string    `xorm:"json"` // JSON object
	Body        string    `xorm:"mediumtext"`
	UserAgent   string    `xorm:"text"`
	ContentType string    `xorm:"varchar(64)"`
	RawData     string    `xorm:"mediumtext"`
	CreatedAt   time.Time `xorm:"datetime created"`
}

// TblAPIKey represents an API key
type TblAPIKey struct {
	Id         int64     `xorm:"pk autoincr"`
	Key        string    `xorm:"varchar(128) notnull unique"`
	KeyPrefix  string    `xorm:"varchar(32) notnull"`
	Name       string    `xorm:"varchar(255) notnull"`
	Scopes     string    `xorm:"json"` // JSON array
	ExpiresAt  time.Time `xorm:"datetime"`
	LastUsedAt time.Time `xorm:"datetime"`
	CreatedBy  int64     `xorm:"notnull"`
	CreatedAt  time.Time `xorm:"datetime created"`
	RevokedAt  time.Time `xorm:"datetime"`
	IsRevoked  bool      `xorm:"default false notnull"`
}

// v2 API request/response models (DEPRECATED - use internal/models instead)

type Case struct {
	Id          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Status      string   `json:"status"`
	Tags        []string `json:"tags"`
	CreatedBy   string   `json:"created_by"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type CaseCreateRequest struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Target      string   `json:"target"`
	Tags        []string `json:"tags"`
}

type CaseUpdateRequest struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Target      string   `json:"target,omitempty"`
	Status      string   `json:"status,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type CaseListResponse struct {
	Items      []Case `json:"items"`
	Total      int    `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	TotalPages int    `json:"total_pages"`
}

type Payload struct {
	Id               string            `json:"id"`
	CaseId           string            `json:"case_id"`
	Token            string            `json:"token"`
	Template         string            `json:"template"`
	RenderedPayload  string            `json:"rendered_payload"`
	Variables        map[string]string `json:"variables"`
	Status           string            `json:"status"`
	ExpectedProtocol string            `json:"expected_protocol,omitempty"`
	ExpiresAt        string            `json:"expires_at,omitempty"`
	CreatedBy        string            `json:"created_by"`
	CreatedAt        string            `json:"created_at"`
	UpdatedAt        string            `json:"updated_at"`
}

type PayloadCreateRequest struct {
	CaseId           string            `json:"case_id,omitempty"`
	Template         string            `json:"template"`
	Variables        map[string]string `json:"variables,omitempty"`
	ExpiresAt        string            `json:"expires_at,omitempty"`
	ExpectedProtocol string            `json:"expected_protocol,omitempty"`
}

type PayloadListResponse struct {
	Items      []Payload `json:"items"`
	Total      int       `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}

type Interaction struct {
	Id          string            `json:"id"`
	Type        string            `json:"type"`
	CaseId      string            `json:"case_id,omitempty"`
	PayloadId   string            `json:"payload_id,omitempty"`
	Token       string            `json:"token,omitempty"`
	Timestamp   string            `json:"timestamp"`
	SourceIp    string            `json:"source_ip"`
	Domain      string            `json:"domain,omitempty"`
	DnsType     string            `json:"dns_type,omitempty"`
	Method      string            `json:"method,omitempty"`
	Path        string            `json:"path,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        string            `json:"body,omitempty"`
	UserAgent   string            `json:"user_agent,omitempty"`
	ContentType string            `json:"content_type,omitempty"`
	RawData     string            `json:"raw_data"`
	CreatedAt   string            `json:"created_at"`
}

type InteractionListResponse struct {
	Items      []Interaction `json:"items"`
	Total      int           `json:"total"`
	Page       int           `json:"page"`
	PageSize   int           `json:"page_size"`
	TotalPages int           `json:"total_pages"`
}

type APIKey struct {
	Id         string   `json:"id"`
	Key        string   `json:"key"`
	KeyPrefix  string   `json:"key_prefix"`
	Name       string   `json:"name"`
	Scopes     []string `json:"scopes"`
	ExpiresAt  string   `json:"expires_at,omitempty"`
	LastUsedAt string   `json:"last_used_at,omitempty"`
	CreatedBy  string   `json:"created_by"`
	CreatedAt  string   `json:"created_at"`
	RevokedAt  string   `json:"revoked_at,omitempty"`
	IsRevoked  bool     `json:"is_revoked"`
}

type APIKeyCreateRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresAt string   `json:"expires_at,omitempty"`
}

type APIKeyListResponse struct {
	Items      []APIKey `json:"items"`
	Total      int      `json:"total"`
	Page       int      `json:"page"`
	PageSize   int      `json:"page_size"`
	TotalPages int      `json:"total_pages"`
}

// Notification models
type NotificationChannel struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`   // webhook, wechat, feishu, dingtalk
	Config    string `json:"config"` // JSON string for channel-specific config
	Enabled   bool   `json:"enabled"`
	CreatedBy string `json:"created_by"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type NotificationChannelCreateRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type" binding:"required,oneof=webhook wechat feishu dingtalk"`
	Config string `json:"config"`
}

type NotificationChannelUpdateRequest struct {
	Name    string `json:"name,omitempty"`
	Config  string `json:"config,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

type NotificationChannelListResponse struct {
	Items      []NotificationChannel `json:"items"`
	Total      int                   `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

type NotificationLog struct {
	Id        string `json:"id"`
	ChannelId string `json:"channel_id"`
	Channel   string `json:"channel"`
	Type      string `json:"type"`
	Status    string `json:"status"` // success, failed
	Message   string `json:"message"`
	Payload   string `json:"payload"`
	CreatedAt string `json:"created_at"`
}

type NotificationLogListResponse struct {
	Items      []NotificationLog `json:"items"`
	Total      int               `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}
