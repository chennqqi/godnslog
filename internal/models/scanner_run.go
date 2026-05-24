package models

import (
	"encoding/json"
	"time"
)

// ScannerRun represents a scanner distribution context for external scanners.
// Aligned with docs/unified-terminology.md and Sprint I plan
type ScannerRun struct {
	ID             string    `json:"id" xorm:"pk varchar(36) notnull"`
	CaseID         string    `json:"case_id" xorm:"varchar(36) notnull index"`
	PayloadID      string    `json:"payload_id" xorm:"varchar(36) notnull index"`
	Scanner        string    `json:"scanner" xorm:"varchar(32) notnull index"` // nuclei (Sprint I only supports nuclei)
	Target         string    `json:"target" xorm:"varchar(512) notnull"`
	Template       string    `json:"template" xorm:"varchar(64) notnull"`                        // ssrf-basic, xxe-basic, etc.
	DeliveryMethod string    `json:"delivery_method" xorm:"varchar(32) notnull"`                 // nuclei-jsonl, nuclei-var
	Command        string    `json:"command" xorm:"text"`                                        // Generated scanner command
	Jsonl          string    `json:"jsonl" xorm:"text"`                                          // Generated JSONL record (single line)
	Status         string    `json:"status" xorm:"varchar(32) notnull default('created') index"` // created, distributed, observed, evidenced
	CreatedBy      string    `json:"created_by" xorm:"varchar(36) notnull"`
	CreatedAt      time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt      time.Time `json:"updated_at" xorm:"datetime updated"`
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
	ScannerNuclei = "nuclei"
)

// DeliveryMethod constants
const (
	DeliveryMethodNucleiJsonl = "nuclei-jsonl"
	DeliveryMethodNucleiVar   = "nuclei-var"
)

// ScannerRunCreateRequest represents the request to create a scanner run
type ScannerRunCreateRequest struct {
	CaseID         string `json:"case_id" binding:"required"`
	PayloadID      string `json:"payload_id" binding:"required"`
	Scanner        string `json:"scanner" binding:"required,oneof=nuclei"`
	Target         string `json:"target" binding:"required"`
	Template       string `json:"template" binding:"required"`
	DeliveryMethod string `json:"delivery_method" binding:"required,oneof=nuclei-jsonl nuclei-var"`
}

// ScannerRunUpdateStatusRequest represents the request to update scanner run status
type ScannerRunUpdateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=created distributed observed evidenced"`
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
