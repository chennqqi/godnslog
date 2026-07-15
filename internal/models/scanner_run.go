package models

import (
	"encoding/json"
	"time"
)

// ScannerRun represents a scanner distribution context for external scanners.
// Aligned with docs/unified-terminology.md and Sprint I plan
type ScannerRun struct {
	ID              string                 `json:"id" xorm:"pk varchar(36) notnull"`
	CaseID          string                 `json:"case_id" xorm:"varchar(36) notnull index"`
	PayloadID       string                 `json:"payload_id" xorm:"varchar(36) notnull index"`
	Scanner         string                 `json:"scanner" xorm:"varchar(32) notnull index"` // nuclei, burp, yakit, zap, xray, rad, postman, apifox
	Target          string                 `json:"target" xorm:"varchar(512) notnull"`
	Template        string                 `json:"template" xorm:"varchar(64) notnull"`                        // ssrf-basic, xxe-basic, etc.
	DeliveryMethod  string                 `json:"delivery_method" xorm:"varchar(32) notnull"`                 // scanner adapter delivery method
	Command         string                 `json:"command" xorm:"text"`                                        // Generated scanner command
	Jsonl           string                 `json:"jsonl" xorm:"text"`                                          // Generated JSONL record (single line)
	PackageManifest ScannerPackageManifest `json:"package_manifest" xorm:"json"`                               // Machine-readable integration package manifest
	PackageHash     string                 `json:"package_hash" xorm:"varchar(64) index"`                      // SHA-256 hash of generated package contents
	Status          string                 `json:"status" xorm:"varchar(32) notnull default('created') index"` // created, distributed, observed, evidenced
	CreatedBy       string                 `json:"created_by" xorm:"varchar(36) notnull"`
	CreatedAt       time.Time              `json:"created_at" xorm:"datetime created"`
	UpdatedAt       time.Time              `json:"updated_at" xorm:"datetime updated"`
}

// ScannerPackageManifest describes the generated Scanner Hub package in a machine-readable form.
type ScannerPackageManifest struct {
	SchemaVersion   string               `json:"schema_version"`
	Scanner         string               `json:"scanner"`
	DeliveryMethod  string               `json:"delivery_method"`
	PackageHash     string               `json:"package_hash"`
	HashAlgorithm   string               `json:"hash_algorithm"`
	Files           []ScannerPackageFile `json:"files"`
	InteractionsURL string               `json:"interactions_url"`
	EvidenceURL     string               `json:"evidence_url"`
	NextActions     []string             `json:"next_actions"`
}

// ScannerPackageFile describes one generated package file or virtual artifact.
type ScannerPackageFile struct {
	Name        string `json:"name"`
	Kind        string `json:"kind"`
	Description string `json:"description"`
}

// MarshalJSON implements json.Marshaler interface for ScannerPackageManifest.
func (m ScannerPackageManifest) MarshalJSON() ([]byte, error) {
	type Alias ScannerPackageManifest
	return json.Marshal(Alias(m))
}

// MarshalJSON implements json.Marshaler interface for ScannerRun
func (s *ScannerRun) MarshalJSON() ([]byte, error) {
	type Alias ScannerRun
	return json.Marshal(&struct {
		*Alias
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}{
		Alias:     (*Alias)(s),
		CreatedAt: s.CreatedAt.Format(time.RFC3339),
		UpdatedAt: s.UpdatedAt.Format(time.RFC3339),
	})
}

// TableName returns the table name for ScannerRun model
func (ScannerRun) TableName() string {
	return "scanner_runs"
}

// Status constants for ScannerRun
const (
	ScannerRunStatusCreated     = "created"
	ScannerRunStatusDistributed = "distributed"
	ScannerRunStatusObserved    = "observed"
	ScannerRunStatusEvidenced   = "evidenced"
)

// Scanner constants
const (
	ScannerNuclei  = "nuclei"
	ScannerBurp    = "burp"
	ScannerYakit   = "yakit"
	ScannerZap     = "zap"
	ScannerXray    = "xray"
	ScannerRad     = "rad"
	ScannerPostman = "postman"
	ScannerApifox  = "apifox"
)

// DeliveryMethod constants
const (
	DeliveryMethodNucleiJsonl   = "nuclei-jsonl"
	DeliveryMethodNucleiVar     = "nuclei-var"
	DeliveryMethodBurpExtension = "burp-extension"
	DeliveryMethodYakitScript   = "yakit-script"
	DeliveryMethodZapScript     = "zap-script"
	DeliveryMethodXrayWebhook   = "xray-webhook"
	DeliveryMethodRadWebhook    = "rad-webhook"
	DeliveryMethodPostmanEnv    = "postman-env"
	DeliveryMethodApifoxEnv     = "apifox-env"
)

// ScannerAdapter describes a supported Scanner Hub integration adapter.
type ScannerAdapter struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Category         string   `json:"category"`
	Maturity         string   `json:"maturity"`
	SupportedMethods []string `json:"supported_methods"`
	DefaultMethod    string   `json:"default_method"`
	Description      string   `json:"description"`
}

// ScannerAdapterListResponse represents the official Scanner Hub adapter catalog.
type ScannerAdapterListResponse struct {
	Items []ScannerAdapter `json:"items"`
}

// ScannerRunCreateRequest represents the request to create a scanner run
type ScannerRunCreateRequest struct {
	CaseID         string `json:"case_id" binding:"required"`
	PayloadID      string `json:"payload_id" binding:"required"`
	Scanner        string `json:"scanner" binding:"required,oneof=nuclei burp yakit zap xray rad postman apifox"`
	Target         string `json:"target" binding:"required"`
	Template       string `json:"template" binding:"required"`
	DeliveryMethod string `json:"delivery_method" binding:"required,oneof=nuclei-jsonl nuclei-var burp-extension yakit-script zap-script xray-webhook rad-webhook postman-env apifox-env"`
}

// ScannerRunUpdateStatusRequest represents the request to update scanner run status
type ScannerRunUpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=created distributed observed evidenced"`
}

// ScannerRunCreateFromSearchRequest represents the request to create scanner runs from search results
type ScannerRunCreateFromSearchRequest struct {
	CaseID         string `json:"case_id" binding:"required"`
	PayloadID      string `json:"payload_id" binding:"required"`
	Source         string `json:"source" binding:"required"`
	Scanner        string `json:"scanner" binding:"omitempty,oneof=nuclei burp yakit zap xray rad postman apifox"`
	Template       string `json:"template" binding:"omitempty"`
	DeliveryMethod string `json:"delivery_method" binding:"omitempty,oneof=nuclei-jsonl nuclei-var burp-extension yakit-script zap-script xray-webhook rad-webhook postman-env apifox-env"`
	Results        []ScanTargetItem `json:"results" binding:"required,min=1"`
}

// ScanTargetItem represents a single search result item used as a scan target
type ScanTargetItem struct {
	IP       string `json:"ip"`
	Port     int    `json:"port"`
	Protocol string `json:"protocol,omitempty"`
	Hostname string `json:"hostname,omitempty"`
}

// ScannerRunListResponse represents the response for listing scanner runs
type ScannerRunListResponse struct {
	Items      []ScannerRun `json:"items"`
	Total      int64        `json:"total"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	TotalPages int          `json:"total_pages"`
}

// ScannerRunDetail represents the detailed scanner run with derived fields
type ScannerRunDetail struct {
	ScannerRun
	InteractionCount  int        `json:"interaction_count"`
	LastInteractionAt *time.Time `json:"last_interaction_at,omitempty"`
	EvidenceCount     int        `json:"evidence_count"`
	LatestEvidenceID  *string    `json:"latest_evidence_id,omitempty"`
	InteractionsURL   string     `json:"interactions_url"`
	EvidenceURL       string     `json:"evidence_url"`
}

// MarshalJSON implements json.Marshaler interface for ScannerRunDetail
func (s *ScannerRunDetail) MarshalJSON() ([]byte, error) {
	type Alias ScannerRunDetail
	lastInteractionAt := ""
	if s.LastInteractionAt != nil {
		lastInteractionAt = s.LastInteractionAt.Format(time.RFC3339)
	}
	return json.Marshal(&struct {
		*Alias
		CreatedAt         string `json:"created_at"`
		UpdatedAt         string `json:"updated_at"`
		LastInteractionAt string `json:"last_interaction_at,omitempty"`
	}{
		Alias:             (*Alias)(s),
		CreatedAt:         s.ScannerRun.CreatedAt.Format(time.RFC3339),
		UpdatedAt:         s.ScannerRun.UpdatedAt.Format(time.RFC3339),
		LastInteractionAt: lastInteractionAt,
	})
}
