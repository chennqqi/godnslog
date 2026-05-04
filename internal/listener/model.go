package listener

import "time"

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
	ID        string    `json:"id" xorm:"'id' pk"`
	Protocol  Protocol  `json:"protocol"`
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Token     string    `json:"token"`
	IsEnabled bool      `json:"is_enabled"`
	CreatedAt time.Time `json:"created_at" xorm:"created"`
	UpdatedAt time.Time `json:"updated_at" xorm:"updated"`
}

// ListenerInteraction represents an interaction from a listener
type ListenerInteraction struct {
	ID         string            `json:"id" xorm:"'id' pk"`
	ListenerID string            `json:"listener_id"`
	Protocol   Protocol          `json:"protocol"`
	SourceIP   string            `json:"source_ip"`
	SourcePort int               `json:"source_port"`
	Data       string            `json:"data"`
	Metadata   map[string]string `json:"metadata"`
	Timestamp  time.Time         `json:"timestamp" xorm:"created"`
}

// SMTPMessage represents an SMTP message interaction
type SMTPMessage struct {
	ID         string    `json:"id" xorm:"'id' pk"`
	ListenerID string    `json:"listener_id"`
	From       string    `json:"from"`
	To         []string  `json:"to"`
	Subject    string    `json:"subject"`
	Body       string    `json:"body"`
	Headers    string    `json:"headers"` // JSON string
	SourceIP   string    `json:"source_ip"`
	Timestamp  time.Time `json:"timestamp" xorm:"created"`
}

// LDAPQuery represents an LDAP query interaction
type LDAPQuery struct {
	ID         string    `json:"id" xorm:"'id' pk"`
	ListenerID string    `json:"listener_id"`
	BaseDN     string    `json:"base_dn"`
	Filter     string    `json:"filter"`
	Attributes string    `json:"attributes"` // JSON string
	BindDN     string    `json:"bind_dn"`
	SourceIP   string    `json:"source_ip"`
	Timestamp  time.Time `json:"timestamp" xorm:"created"`
}

// SMBRequest represents an SMB request interaction
type SMBRequest struct {
	ID         string    `json:"id" xorm:"'id' pk"`
	ListenerID string    `json:"listener_id"`
	Command    string    `json:"command"` // SMB command (e.g., TREE_CONNECT, OPEN, READ, WRITE)
	ShareName  string    `json:"share_name"`
	FilePath   string    `json:"file_path"`
	Username   string    `json:"username"`
	Data       string    `json:"data"` // JSON string for request data
	SourceIP   string    `json:"source_ip"`
	SourcePort int       `json:"source_port"`
	Timestamp  time.Time `json:"timestamp" xorm:"created"`
}

// FTPCommand represents an FTP command interaction
type FTPCommand struct {
	ID         string    `json:"id" xorm:"'id' pk"`
	ListenerID string    `json:"listener_id"`
	Command    string    `json:"command"` // FTP command (e.g., USER, PASS, LIST, RETR, STOR)
	Argument   string    `json:"argument"`
	Username   string    `json:"username"`
	Data       string    `json:"data"` // For STOR/APPE commands
	SourceIP   string    `json:"source_ip"`
	SourcePort int       `json:"source_port"`
	Timestamp  time.Time `json:"timestamp" xorm:"created"`
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
