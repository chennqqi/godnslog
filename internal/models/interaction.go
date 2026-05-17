package models

import (
	"database/sql/driver"
	"encoding/json"
	"strconv"
	"time"

	"github.com/chennqqi/godnslog/models"
	"xorm.io/xorm"
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

// Interaction represents a captured out-of-band event (DNS, HTTP, SMTP, LDAP, SMB, FTP) triggered by a Probe.
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

// MarshalJSON implements json.Marshaler interface for Interaction
func (i *Interaction) MarshalJSON() ([]byte, error) {
	type Alias Interaction
	return json.Marshal(&struct {
		*Alias
		Timestamp string `json:"timestamp"`
		CreatedAt string `json:"created_at"`
	}{
		Alias:     (*Alias)(i),
		Timestamp: i.Timestamp.Format(time.RFC3339),
		CreatedAt: i.CreatedAt.Format(time.RFC3339),
	})
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

// FromTblDnsWithAttribution converts models.TblDns to Interaction with payload/case attribution
func FromTblDnsWithAttribution(dns *models.TblDns, engine *xorm.Engine) *Interaction {
	domain := dns.Domain
	token := dns.Var
	dnsType := DNSTypeA

	interaction := &Interaction{
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

	// Auto-attribution: associate interaction with payload and case based on token
	if token != "" {
		type PayloadInfo struct {
			ID     int64 `xorm:"id"`
			CaseId int64 `xorm:"case_id"`
		}
		var payloadInfo PayloadInfo
		has, err := engine.Table("payloads").Where("token = ?", token).Get(&payloadInfo)
		if err == nil && has {
			payloadID := strconv.FormatInt(payloadInfo.ID, 10)
			caseID := strconv.FormatInt(payloadInfo.CaseId, 10)
			interaction.PayloadID = &payloadID
			interaction.CaseID = &caseID
		}
	}

	return interaction
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

// FromTblHttpWithAttribution converts models.TblHttp to Interaction with payload/case attribution
func FromTblHttpWithAttribution(http *models.TblHttp, engine *xorm.Engine) *Interaction {
	token := http.Var
	method := http.Method
	path := http.Path
	ua := http.Ua
	ctype := http.Ctype

	interaction := &Interaction{
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

	// Auto-attribution: associate interaction with payload and case based on token
	if token != "" {
		type PayloadInfo struct {
			ID     int64 `xorm:"id"`
			CaseId int64 `xorm:"case_id"`
		}
		var payloadInfo PayloadInfo
		has, err := engine.Table("payloads").Where("token = ?", token).Get(&payloadInfo)
		if err == nil && has {
			payloadID := strconv.FormatInt(payloadInfo.ID, 10)
			caseID := strconv.FormatInt(payloadInfo.CaseId, 10)
			interaction.PayloadID = &payloadID
			interaction.CaseID = &caseID
		}
	}

	return interaction
}
