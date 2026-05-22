package interaction

import (
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/payload"
)

// Evidence represents aggregated, scored, and explained summary of Interactions for a Case
// This matches the unified terminology definition in docs/unified-terminology.md
type Evidence struct {
	ID               string               `json:"id"`
	CaseID           string               `json:"case_id"`
	PayloadID        string               `json:"payload_id,omitempty"` // Optional: can generate evidence for case or specific payload
	Payload          *payload.Payload     `json:"payload,omitempty"`
	EvidenceStrength EvidenceStrength     `json:"evidence_strength"` // low, medium, high, critical
	Confidence       int                  `json:"confidence"`        // 0-100
	InteractionCount int                  `json:"interaction_count"`
	UniqueSources    int                  `json:"unique_sources"`
	Interactions     []models.Interaction `json:"interactions"`
	Timeline         []TimelineItem       `json:"timeline"`
	Explainability   string               `json:"explainability"` // Human-readable explanation
	GeneratedAt      time.Time            `json:"generated_at"`
}

// EvidenceStrength represents qualitative assessment of evidence
type EvidenceStrength string

const (
	EvidenceStrengthLow      EvidenceStrength = "low"
	EvidenceStrengthMedium   EvidenceStrength = "medium"
	EvidenceStrengthHigh     EvidenceStrength = "high"
	EvidenceStrengthCritical EvidenceStrength = "critical"
)

// TimelineItem represents a timeline item in the evidence chain
type TimelineItem struct {
	Type        string                 `json:"type"` // payload_created, interaction, note
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EvidenceRequest represents the request for evidence generation
// Either case_id or payload_id must be provided (at least one is required)
type EvidenceRequest struct {
	CaseID    string `json:"case_id,omitempty"`
	PayloadID string `json:"payload_id,omitempty"`
	Format    string `json:"format" binding:"required,oneof=json markdown"` // json, markdown
}

// EvidenceResponse represents the response for evidence generation
// The main semantic is the structured Evidence, with format/content for export
type EvidenceResponse struct {
	Evidence *Evidence              `json:"evidence"` // Structured evidence result (main semantic)
	Format   string                 `json:"format"`   // json | markdown
	Content  string                 `json:"content"`  // Exported content (JSON string or Markdown text)
	Metadata map[string]interface{} `json:"metadata"` // Auxiliary information
}
