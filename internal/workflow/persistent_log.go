package workflow

import "time"

// PersistentActionLog records the state and result of a workflow action execution in the database.
// This enables recovery of pending jobs after process restart.
type PersistentActionLog struct {
	ID         string    `json:"id" xorm:"'id' pk varchar(36) notnull"`
	WorkflowID string    `json:"workflow_id" xorm:"varchar(36) notnull index"`
	ActionID   string    `json:"action_id" xorm:"varchar(36) notnull"`
	ActionType string    `json:"action_type" xorm:"varchar(64) notnull"`

	// InteractionID is the ID of the interaction that triggered this action (may be empty).
	InteractionID string `json:"interaction_id" xorm:"varchar(36) index"`

	// Status: pending, running, completed, failed
	Status string `json:"status" xorm:"varchar(16) notnull default 'pending'"`

	// Attempt is the current retry attempt (0-based).
	Attempt int `json:"attempt" xorm:"int notnull default 0"`

	// Error stores the last error message if any.
	Error string `json:"error" xorm:"text"`

	// DurationMs is the execution duration in milliseconds.
	DurationMs int64 `json:"duration_ms" xorm:"bigint default 0"`

	CreatedAt  time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt  time.Time `json:"updated_at" xorm:"datetime updated"`
	FinishedAt time.Time `json:"finished_at" xorm:"datetime"`
}

// TableName returns the table name for PersistentActionLog
func (PersistentActionLog) TableName() string {
	return "workflow_action_logs"
}
