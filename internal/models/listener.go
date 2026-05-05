package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Protocol represents the listener protocol type
type Protocol string

const (
	ProtocolSMTP Protocol = "smtp"
	ProtocolLDAP Protocol = "ldap"
	ProtocolSMB  Protocol = "smb"
	ProtocolFTP  Protocol = "ftp"
)

// Listener represents a protocol listener
type Listener struct {
	ID        string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	Protocol  Protocol  `json:"protocol" xorm:"varchar(16) notnull"`
	Host      string    `json:"host" xorm:"varchar(255) notnull"`
	Port      int       `json:"port" xorm:"int notnull"`
	Token     string    `json:"token" xorm:"varchar(64) notnull"`
	IsEnabled bool      `json:"is_enabled" xorm:"bool notnull default true"`
	CreatedAt time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt time.Time `json:"updated_at" xorm:"datetime updated"`
}

// ListenerInteraction represents an interaction from a listener
type ListenerInteraction struct {
	ID          string            `json:"id" xorm:"'id' pk varchar(36) notnull"`
	ListenerID  string            `json:"listener_id" xorm:"varchar(36) notnull index"`
	Protocol    Protocol          `json:"protocol" xorm:"varchar(16) notnull"`
	SourceIP    string            `json:"source_ip" xorm:"varchar(64) notnull"`
	SourcePort  int               `json:"source_port" xorm:"int notnull"`
	Data        string            `json:"data" xorm:"mediumtext"`
	Metadata    Metadata          `json:"metadata" xorm:"json"`
	Timestamp   time.Time         `json:"timestamp" xorm:"datetime notnull created"`
}

// Metadata represents metadata as a JSON map
type Metadata map[string]string

// Scan implements sql.Scanner interface for Metadata
func (m *Metadata) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, m)
}

// Value implements driver.Valuer interface for Metadata
func (m Metadata) Value() (driver.Value, error) {
	if m == nil {
		return nil, nil
	}
	return json.Marshal(m)
}

// SMTPMessage represents an SMTP message interaction
type SMTPMessage struct {
	ID         string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	ListenerID string    `json:"listener_id" xorm:"varchar(36) notnull index"`
	From       string    `json:"from" xorm:"varchar(255) notnull"`
	To         string    `json:"to" xorm:"text"` // JSON array
	Subject    string    `json:"subject" xorm:"text"`
	Body       string    `json:"body" xorm:"mediumtext"`
	Headers    string    `json:"headers" xorm:"mediumtext"` // JSON string
	SourceIP   string    `json:"source_ip" xorm:"varchar(64) notnull"`
	Timestamp  time.Time `json:"timestamp" xorm:"datetime notnull created"`
}

// LDAPQuery represents an LDAP query interaction
type LDAPQuery struct {
	ID         string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	ListenerID string    `json:"listener_id" xorm:"varchar(36) notnull index"`
	BaseDN     string    `json:"base_dn" xorm:"text"`
	Filter     string    `json:"filter" xorm:"text"`
	Attributes string   `json:"attributes" xorm:"mediumtext"` // JSON string
	BindDN     string    `json:"bind_dn" xorm:"varchar(255)"`
	SourceIP   string    `json:"source_ip" xorm:"varchar(64) notnull"`
	Timestamp  time.Time `json:"timestamp" xorm:"datetime notnull created"`
}

// SMBRequest represents an SMB request interaction
type SMBRequest struct {
	ID         string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	ListenerID string    `json:"listener_id" xorm:"varchar(36) notnull index"`
	Command    string    `json:"command" xorm:"varchar(64) notnull"` // SMB command
	ShareName  string    `json:"share_name" xorm:"varchar(255)"`
	FilePath   string    `json:"file_path" xorm:"text"`
	Username   string    `json:"username" xorm:"varchar(255)"`
	Data       string    `json:"data" xorm:"mediumtext"` // JSON string
	SourceIP   string    `json:"source_ip" xorm:"varchar(64) notnull"`
	SourcePort int      `json:"source_port" xorm:"int notnull"`
	Timestamp  time.Time `json:"timestamp" xorm:"datetime notnull created"`
}

// FTPCommand represents an FTP command interaction
type FTPCommand struct {
	ID         string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	ListenerID string    `json:"listener_id" xorm:"varchar(36) notnull index"`
	Command    string    `json:"command" xorm:"varchar(64) notnull"` // FTP command
	Argument   string    `json:"argument" xorm:"text"`
	Username   string    `json:"username" xorm:"varchar(255)"`
	Data       string    `json:"data" xorm:"mediumtext"` // For STOR/APPE commands
	SourceIP   string    `json:"source_ip" xorm:"varchar(64) notnull"`
	SourcePort int      `json:"source_port" xorm:"int notnull"`
	Timestamp  time.Time `json:"timestamp" xorm:"datetime notnull created"`
}

// ListenerConfig holds listener configuration
type ListenerConfig struct {
	MaxConnections int           `json:"max_connections"`
	Timeout        time.Duration `json:"timeout"`
	BufferSize     int           `json:"buffer_size"`
	EnableTLS      bool          `json:"enable_tls"`
	TLSCertFile    string        `json:"tls_cert_file"`
	TLSKeyFile     string        `json:"tls_key_file"`
}

// TableName returns the table name for Listener
func (Listener) TableName() string {
	return "listeners"
}

// TableName returns the table name for ListenerInteraction
func (ListenerInteraction) TableName() string {
	return "listener_interactions"
}

// TableName returns the table name for SMTPMessage
func (SMTPMessage) TableName() string {
	return "smtp_messages"
}

// TableName returns the table name for LDAPQuery
func (LDAPQuery) TableName() string {
	return "ldap_queries"
}

// TableName returns the table name for SMBRequest
func (SMBRequest) TableName() string {
	return "smb_requests"
}

// TableName returns the table name for FTPCommand
func (FTPCommand) TableName() string {
	return "ftp_commands"
}

// ListenerListResponse represents the response for listing listeners
type ListenerListResponse struct {
	Items      []Listener `json:"items"`
	Total      int64      `json:"total"`
	Page       int        `json:"page"`
	PageSize   int        `json:"page_size"`
	TotalPages int        `json:"total_pages"`
}
