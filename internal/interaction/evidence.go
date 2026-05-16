package interaction

import (
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/payload"
)

// Evidence represents an evidence chain composed of payload and interactions
type Evidence struct {
	ID           string               `json:"id"`
	CaseID       string               `json:"case_id"`
	PayloadID    string               `json:"payload_id"`
	Payload      *payload.Payload     `json:"payload"`
	Interactions []models.Interaction `json:"interactions"`
	Timeline     []TimelineItem       `json:"timeline"`
	Score        float64              `json:"score"`
	CreatedAt    time.Time            `json:"created_at"`
}

// TimelineItem represents a timeline item in the evidence chain
type TimelineItem struct {
	Type        string                 `json:"type"` // payload_created, payload_deployed, interaction, note
	Description string                 `json:"description"`
	Timestamp   time.Time              `json:"timestamp"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// EvidenceResponse represents the response for evidence generation
type EvidenceResponse struct {
	Format   string                 `json:"format"`
	Content  string                 `json:"content"`
	Metadata map[string]interface{} `json:"metadata"`
}
