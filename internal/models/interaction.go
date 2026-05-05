package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/chennqqi/godnslog/models"
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
// Unified from internal/interaction/interaction.go and models/v2.go TblInteraction
// Also serves as unified storage for 1.0 TblDns and TblHttp
type Interaction struct {
	ID        string    `json:"id" xorm:"pk varchar(36) notnull"`
	Type      string    `json:"type" xorm:"varchar(16) notnull index"` // dns, http, smtp, ldap, smb, ftp
	CaseID    *string   `json:"case_id" xorm:"varchar(36) index"`
	PayloadID *string   `json:"payload_id" xorm:"varchar(36) index"`
	Token     *string   `json:"token" xorm:"varchar(64) index"`
	Timestamp time.Time `json:"timestamp" xorm:"datetime notnull"`
	SourceIP  string    `json:"source_ip" xorm:"varchar(64) notnull"`

	// DNS specific fields
	Domain  *string `json:"domain" xorm:"varchar(255)"`
	DNSType *string `json:"dns_type" xorm:"varchar(16)"` // A, AAAA, CNAME, etc.

	// HTTP specific fields
	Method      *string `json:"method" xorm:"varchar(16)"` // GET, POST, etc.
	Path        *string `json:"path" xorm:"text"`
	Headers     Headers `json:"headers" xorm:"json"`
	Body        *string `json:"body" xorm:"mediumtext"`
	UserAgent   *string `json:"user_agent" xorm:"text"`
	ContentType *string `json:"content_type" xorm:"varchar(128)"`

	// Common fields
	RawData   string    `json:"raw_data" xorm:"mediumtext"`
	CreatedAt time.Time `json:"created_at" xorm:"datetime created"`
}

// TableName returns the table name for Interaction model
func (Interaction) TableName() string {
	return "interactions"
}

// Type constants
const (
	InteractionTypeDNS  = "dns"
	InteractionTypeHTTP = "http"
	InteractionTypeSMTP = "smtp"
	InteractionTypeLDAP = "ldap"
	InteractionTypeSMB  = "smb"
	InteractionTypeFTP  = "ftp"
)

// DNS Type constants
const (
	DNSTypeA     = "A"
	DNSTypeAAAA  = "AAAA"
	DNSTypeCNAME = "CNAME"
	DNSTypeTXT   = "TXT"
	DNSTypeMX    = "MX"
	DNSTypeNS    = "NS"
)

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

// FromTblDns converts models.TblDns to Interaction (for migration)
func FromTblDns(dns *models.TblDns) *Interaction {
	domain := dns.Domain
	token := dns.Var
	dnsType := DNSTypeA

	return &Interaction{
		ID:        GenerateID(),
		Type:      InteractionTypeDNS,
		Token:     &token,
		Timestamp: dns.Ctime,
		SourceIP:  dns.Ip,
		Domain:    &domain,
		DNSType:   &dnsType,
		RawData:   dns.Domain,
		CreatedAt: dns.Atime,
	}
}

// FromTblHttp converts models.TblHttp to Interaction (for migration)
func FromTblHttp(http *models.TblHttp) *Interaction {
	token := http.Var
	method := http.Method
	path := http.Path
	ua := http.Ua
	ctype := http.Ctype

	return &Interaction{
		ID:          GenerateID(),
		Type:        InteractionTypeHTTP,
		Token:       &token,
		Timestamp:   http.Ctime,
		SourceIP:    http.Ip,
		Method:      &method,
		Path:        &path,
		Body:        &http.Data,
		UserAgent:   &ua,
		ContentType: &ctype,
		RawData:     http.Path,
		CreatedAt:   http.Atime,
	}
}
