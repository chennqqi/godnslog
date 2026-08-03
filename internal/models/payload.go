package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"database/sql/driver"
)

// Variables represents custom variables for payload template rendering
type Variables map[string]string

// Scan implements sql.Scanner interface for Variables
func (v *Variables) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, v)
}

// Value implements driver.Valuer interface for Variables
func (v Variables) Value() (driver.Value, error) {
	if v == nil {
		return nil, nil
	}
	return json.Marshal(v)
}

// Payload represents a generated OAST probe template with rendered values, ready for delivery to targets.
// Unified from internal/payload/payload.go and models/v2.go TblPayload
// Aligned with docs/unified-terminology.md
type Payload struct {
	ID               string     `json:"id" xorm:"pk varchar(36) notnull"`
	CaseID           string     `json:"case_id" xorm:"varchar(36) notnull index"`
	Token            string     `json:"token" xorm:"varchar(64) notnull unique index"`
	TemplateID       string     `json:"template_id" xorm:"varchar(64) notnull"`                    // Template identifier (e.g., ssrf-basic, xxe-basic)
	TemplateRendered string     `json:"template_rendered" xorm:"text"`                             // Rendered payload with variables substituted
	Variables        Variables  `json:"variables" xorm:"json"`                                     // Custom variable values used in rendering
	Status           string     `json:"status" xorm:"varchar(32) notnull default('active') index"` // active, expired, revoked
	ExpectedProtocol string     `json:"expected_protocol" xorm:"varchar(16)"`                      // dns, http, smtp, ldap
	CustomResponse   string     `json:"custom_response,omitempty" xorm:"text"`                     // JSON: {status, headers, body, redirect}
	ExpiresAt        *time.Time `json:"expires_at" xorm:"datetime"`
	CreatedBy        string     `json:"created_by" xorm:"varchar(36) notnull"`
	CreatedAt        time.Time  `json:"created_at" xorm:"datetime created"`
	UpdatedAt        time.Time  `json:"updated_at" xorm:"datetime updated"`
}

// MarshalJSON implements json.Marshaler interface for Payload
func (p *Payload) MarshalJSON() ([]byte, error) {
	type Alias Payload
	expiresAt := ""
	if p.ExpiresAt != nil {
		expiresAt = p.ExpiresAt.Format(time.RFC3339)
	}
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		ExpiresAt string `json:"expires_at,omitempty"`
	}{
		Alias:     (*Alias)(p),
		CreatedAt: p.CreatedAt.Format(time.RFC3339),
		ExpiresAt: expiresAt,
	})
}

// TableName returns the table name for Payload model
func (Payload) TableName() string {
	return "payloads"
}

// Status constants
const (
	PayloadStatusActive  = "active"
	PayloadStatusExpired = "expired"
	PayloadStatusRevoked = "revoked"
)

// TemplateMetadata describes a payload template's category and risk level.
type TemplateMetadata struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	Risk        string `json:"risk"`
}

// PayloadTemplates defines available payload templates using {variable} substitution syntax.
var PayloadTemplates = map[string]string{
	// SSRF
	"ssrf-basic":          "http://{token}.{domain}/",
	"ssrf-redirect":       "http://{token}.{domain}/redirect",
	"ssrf-cloud-metadata": "{token}.169.254.169.254.{domain}",
	"ssrf-aws-metadata":   "169.254.169.254.{token}.{domain}",
	"ssrf-gcp-metadata":   "metadata.google.internal.{token}.{domain}",
	"ssrf-azure-metadata": "169.254.169.254.{token}.{domain}",

	// XXE / Injection
	"xxe-basic":             "http://{token}.{domain}/xxe",
	"rfi-remote-file":       "http://{token}.{domain}/file.php",
	"file-inclusion":        "http://{token}.{domain}/include",
	"blind-sqli":            "http://{token}.{domain}/sql?id=1",
	"blind-sqli-dns":        "{token}.{domain}",
	"ssti-template":         "{token}.{domain}",
	"template-injection":    "{token}.{domain}",
	"prototype-pollution":   "{token}.{domain}",
	"host-header-injection": "{token}.{domain}",
	"log4j-jndi":            "${jndi:ldap://{token}.{domain}/exp}",
	"yaml-deserialization":  "http://{token}.{domain}/yaml",
	"deserialization":       "http://{token}.{domain}/object",
	"ldap-injection":        "{token}.{domain}",

	// RCE
	"rce-basic":    "http://{token}.{domain}/cmd",
	"rce-command":  "curl http://{token}.{domain}",
	"rce-callback": "curl http://{token}.{domain}/cb",

	// Client-side
	"cors-jsonp":         "http://{token}.{domain}/callback",
	"pdf-html-rendering": "http://{token}.{domain}/resource",
	"xss-reflected":      "http://{token}.{domain}/xss",

	// API
	"webhook":               "https://{token}.{domain}/webhook",
	"graphql-introspection": "{token}.{domain}",

	// DevOps
	"ci-cd-variable": "{token}.{domain}",

	// Network
	"dns-rebinding":     "http://{token}.{domain}/rebind",
	"smb-relay":         "{token}.{domain}",
	"ftp-exfil":         "ftp://{token}.{domain}/data",
	"request-smuggling": "{token}.{domain}",

	// SMTP
	"smtp-injection": "{token}@{domain}",

	// Auth
	"jwt-confusion": "{token}.{domain}",
}

// PayloadTemplateMetadata provides metadata for each template ID.
var PayloadTemplateMetadata = map[string]TemplateMetadata{
	"ssrf-basic": {
		Name:        "SSRF HTTP",
		Description: "Server-Side Request Forgery detection via HTTP",
		Category:    "ssrf",
		Risk:        "high",
	},
	"ssrf-redirect": {
		Name:        "SSRF Redirect",
		Description: "Open redirect via SSRF",
		Category:    "ssrf",
		Risk:        "high",
	},
	"ssrf-cloud-metadata": {
		Name:        "SSRF Cloud Metadata",
		Description: "Cloud metadata endpoint detection",
		Category:    "ssrf",
		Risk:        "critical",
	},
	"ssrf-aws-metadata": {
		Name:        "SSRF AWS Metadata",
		Description: "AWS EC2 metadata endpoint",
		Category:    "ssrf",
		Risk:        "critical",
	},
	"ssrf-gcp-metadata": {
		Name:        "SSRF GCP Metadata",
		Description: "Google Cloud Platform metadata",
		Category:    "ssrf",
		Risk:        "critical",
	},
	"ssrf-azure-metadata": {
		Name:        "SSRF Azure Metadata",
		Description: "Azure cloud metadata",
		Category:    "ssrf",
		Risk:        "critical",
	},
	"xxe-basic": {
		Name:        "XXE External Entity",
		Description: "XML External Entity injection",
		Category:    "injection",
		Risk:        "high",
	},
	"rfi-remote-file": {
		Name:        "RFI Remote File Inclusion",
		Description: "Remote File Inclusion detection",
		Category:    "injection",
		Risk:        "high",
	},
	"file-inclusion": {
		Name:        "File Inclusion",
		Description: "Local/Remote file inclusion",
		Category:    "injection",
		Risk:        "high",
	},
	"blind-sqli": {
		Name:        "Blind SQLi HTTP",
		Description: "Blind SQL injection via HTTP callback",
		Category:    "sqli",
		Risk:        "high",
	},
	"blind-sqli-dns": {
		Name:        "Blind SQLi DNS",
		Description: "Blind SQL injection via DNS exfiltration",
		Category:    "sqli",
		Risk:        "high",
	},
	"ssti-template": {
		Name:        "SSTI Template Injection",
		Description: "Server-Side Template Injection",
		Category:    "injection",
		Risk:        "high",
	},
	"template-injection": {
		Name:        "Template Injection",
		Description: "Jinja2/ERB/Freemarker template injection",
		Category:    "injection",
		Risk:        "high",
	},
	"prototype-pollution": {
		Name:        "Prototype Pollution",
		Description: "JavaScript prototype pollution",
		Category:    "injection",
		Risk:        "high",
	},
	"host-header-injection": {
		Name:        "Host Header Injection",
		Description: "Host header injection",
		Category:    "injection",
		Risk:        "medium",
	},
	"log4j-jndi": {
		Name:        "Log4j JNDI",
		Description: "Log4j JNDI injection",
		Category:    "injection",
		Risk:        "critical",
	},
	"yaml-deserialization": {
		Name:        "YAML Deserialization",
		Description: "YAML deserialization attack",
		Category:    "injection",
		Risk:        "critical",
	},
	"deserialization": {
		Name:        "Deserialization",
		Description: "Java/Python/PHP deserialization attack",
		Category:    "injection",
		Risk:        "critical",
	},
	"ldap-injection": {
		Name:        "LDAP Injection",
		Description: "LDAP query injection",
		Category:    "injection",
		Risk:        "high",
	},
	"rce-basic": {
		Name:        "RCE Basic",
		Description: "Remote Code Execution via HTTP callback",
		Category:    "rce",
		Risk:        "critical",
	},
	"rce-command": {
		Name:        "RCE Command Injection",
		Description: "Remote Code Execution via command injection",
		Category:    "rce",
		Risk:        "critical",
	},
	"rce-callback": {
		Name:        "RCE Callback",
		Description: "Remote Code Execution via out-of-band callback",
		Category:    "rce",
		Risk:        "critical",
	},
	"cors-jsonp": {
		Name:        "CORS/JSONP",
		Description: "CORS misconfiguration and JSONP detection",
		Category:    "misconfiguration",
		Risk:        "medium",
	},
	"pdf-html-rendering": {
		Name:        "PDF/HTML Rendering",
		Description: "PDF or HTML rendering with external resources",
		Category:    "client-side",
		Risk:        "medium",
	},
	"xss-reflected": {
		Name:        "XSS Reflected",
		Description: "Reflected XSS detection",
		Category:    "client-side",
		Risk:        "medium",
	},
	"webhook": {
		Name:        "Webhook",
		Description: "Webhook endpoint detection",
		Category:    "api",
		Risk:        "low",
	},
	"graphql-introspection": {
		Name:        "GraphQL Introspection",
		Description: "GraphQL API introspection",
		Category:    "api",
		Risk:        "medium",
	},
	"ci-cd-variable": {
		Name:        "CI/CD Variable",
		Description: "CI/CD pipeline variable injection",
		Category:    "devops",
		Risk:        "high",
	},
	"dns-rebinding": {
		Name:        "DNS Rebinding",
		Description: "DNS rebinding attack for SSRF",
		Category:    "network",
		Risk:        "high",
	},
	"smb-relay": {
		Name:        "SMB Relay",
		Description: "SMB NTLM relay attack",
		Category:    "network",
		Risk:        "critical",
	},
	"ftp-exfil": {
		Name:        "FTP Exfiltration",
		Description: "FTP data exfiltration",
		Category:    "network",
		Risk:        "high",
	},
	"request-smuggling": {
		Name:        "Request Smuggling",
		Description: "HTTP request smuggling",
		Category:    "network",
		Risk:        "critical",
	},
	"smtp-injection": {
		Name:        "SMTP Injection",
		Description: "SMTP header injection",
		Category:    "injection",
		Risk:        "medium",
	},
	"jwt-confusion": {
		Name:        "JWT Confusion",
		Description: "JWT algorithm confusion",
		Category:    "auth",
		Risk:        "high",
	},
}

// RenderTemplate renders a payload template with variables
// Aligned with docs/unified-terminology.md
// Uses simple {variable} substitution instead of Go template syntax
func RenderTemplate(tmpl string, variables map[string]string, token, domain string) (string, error) {
	// Add default variables
	if variables == nil {
		variables = make(map[string]string)
	}
	variables["token"] = token
	variables["domain"] = domain
	variables["callback_url"] = fmt.Sprintf("http://%s/log/%s/", domain, token)

	// Simple substitution for {variable} syntax
	result := tmpl
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// RenderTemplateWithCase renders a payload template with case variable
// Aligned with docs/unified-terminology.md
// Uses simple {variable} substitution instead of Go template syntax
func RenderTemplateWithCase(tmpl string, variables map[string]string, token, domain, caseID string) (string, error) {
	// Add default variables
	if variables == nil {
		variables = make(map[string]string)
	}
	variables["token"] = token
	variables["domain"] = domain
	variables["case"] = caseID
	variables["callback_url"] = fmt.Sprintf("http://%s/log/%s/", domain, token)

	// Simple substitution for {variable} syntax
	result := tmpl
	for key, value := range variables {
		placeholder := fmt.Sprintf("{%s}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	return result, nil
}

// PayloadCreateRequest represents the request to create a payload
// Aligned with docs/unified-terminology.md
type PayloadCreateRequest struct {
	CaseID           string            `json:"case_id" binding:"required"`
	TemplateID       string            `json:"template_id" binding:"required"` // Template identifier (e.g., ssrf-basic, xxe-basic)
	Variables        map[string]string `json:"variables"`
	ExpiresAt        *time.Time        `json:"expires_at"`
	ExpectedProtocol string            `json:"expected_protocol" binding:"omitempty,oneof=dns http smtp ldap"`
}

// PayloadUpdateRequest represents the request to update a payload
type PayloadUpdateRequest struct {
	Status           string     `json:"status" binding:"omitempty,oneof=draft deployed hit archived expired"`
	ExpiresAt        *time.Time `json:"expires_at"`
	ExpectedProtocol string     `json:"expected_protocol" binding:"omitempty,oneof=dns http smtp ldap"`
}

// PayloadListResponse represents the response for listing payloads
type PayloadListResponse struct {
	Items      []Payload `json:"items"`
	Total      int64     `json:"total"`
	Page       int       `json:"page"`
	PageSize   int       `json:"page_size"`
	TotalPages int       `json:"total_pages"`
}
