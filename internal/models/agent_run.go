package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// AgentRunStatus represents the status of an agent run
type AgentRunStatus string

const (
	AgentRunStatusCreated   AgentRunStatus = "created"
	AgentRunStatusRunning   AgentRunStatus = "running"
	AgentRunStatusWaiting   AgentRunStatus = "waiting"
	AgentRunStatusCompleted AgentRunStatus = "completed"
	AgentRunStatusFailed    AgentRunStatus = "failed"
	AgentRunStatusCancelled AgentRunStatus = "cancelled"
	AgentRunStatusTimedOut  AgentRunStatus = "timed_out"
)

// AgentRun represents a single task execution lifecycle for an Agent
type AgentRun struct {
	ID         string         `json:"id" xorm:"pk varchar(36) notnull"`
	AgentID    string         `json:"agent_id" xorm:"varchar(128) index"`
	OperatorID string         `json:"operator_id" xorm:"varchar(128) index"`
	CaseID     string         `json:"case_id" xorm:"varchar(36) index"`
	PayloadID  string         `json:"payload_id" xorm:"varchar(36) index"`
	Target     string         `json:"target" xorm:"text"`
	Title      string         `json:"title" xorm:"varchar(256)"`
	Status     AgentRunStatus `json:"status" xorm:"varchar(32) index"`
	StartedAt  *time.Time     `json:"started_at" xorm:"datetime index"`
	EndedAt    *time.Time     `json:"ended_at" xorm:"datetime index"`
	CreatedAt  time.Time      `json:"created_at" xorm:"created"`
	UpdatedAt  time.Time      `json:"updated_at" xorm:"updated"`
}

// TableName returns the table name for AgentRun
func (AgentRun) TableName() string {
	return "agent_runs"
}

// AgentRunDetail represents an agent run with computed fields
type AgentRunDetail struct {
	AgentRun
	InteractionCount  int              `json:"interaction_count"`
	LastInteractionAt *time.Time       `json:"last_interaction_at"`
	Operations        []AgentOperation `json:"operations"`
	CaseURL           string           `json:"case_url"`
	PayloadURL        string           `json:"payload_url"`
	InteractionsURL   string           `json:"interactions_url"`
	EvidenceURL       string           `json:"evidence_url"`
}

// AgentOperation represents an operation within an agent run
type AgentOperation struct {
	ID         string     `json:"id" xorm:"pk varchar(36) notnull"`
	AgentRunID string     `json:"agent_run_id" xorm:"varchar(36) index"`
	AgentID    string     `json:"agent_id" xorm:"varchar(128) index"`
	Action     string     `json:"action" xorm:"varchar(64) index"`
	RiskLevel  string     `json:"risk_level" xorm:"varchar(32)"`
	Request    string     `json:"request" xorm:"text"`
	Result     string     `json:"result" xorm:"text"`
	Error      string     `json:"error" xorm:"text"`
	StartedAt  time.Time  `json:"started_at" xorm:"datetime index"`
	EndedAt    *time.Time `json:"ended_at" xorm:"datetime"`
	CreatedAt  time.Time  `json:"created_at" xorm:"created"`
}

// TableName returns the table name for AgentOperation
func (AgentOperation) TableName() string {
	return "agent_operations"
}

// AgentOperations is a slice of AgentOperation for JSON handling
type AgentOperations []AgentOperation

// Scan implements sql.Scanner for AgentOperations
func (o *AgentOperations) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, o)
}

// Value implements driver.Valuer for AgentOperations
func (o AgentOperations) Value() (driver.Value, error) {
	if o == nil {
		return nil, nil
	}
	return json.Marshal(o)
}

// AgentRunCreateRequest represents a request to create an agent run
type AgentRunCreateRequest struct {
	AgentID    string `json:"agent_id" binding:"required"`
	OperatorID string `json:"operator_id" binding:"required"`
	CaseID     string `json:"case_id"`
	PayloadID  string `json:"payload_id"`
	Target     string `json:"target" binding:"required"`
	Title      string `json:"title" binding:"required"`
}

// AgentRunUpdateStatusRequest represents a request to update agent run status
type AgentRunUpdateStatusRequest struct {
	Status AgentRunStatus `json:"status" binding:"required"`
}

// AgentRunListRequest represents a request to list agent runs
type AgentRunListRequest struct {
	AgentID   string `form:"agent_id"`
	CaseID    string `form:"case_id"`
	PayloadID string `form:"payload_id"`
	Status    string `form:"status"`
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
}

// AgentRunListResponse represents the response for listing agent runs
type AgentRunListResponse struct {
	Items      []AgentRunDetail `json:"items"`
	Total      int64            `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// AgentOperationCreateRequest represents a request to create an agent operation
type AgentOperationCreateRequest struct {
	Action    string                 `json:"action" binding:"required"`
	RiskLevel string                 `json:"risk_level"`
	Request   map[string]interface{} `json:"request"`
	Result    map[string]interface{} `json:"result"`
	Error     string                 `json:"error"`
}

// AgentRunFollowupActionType constants
const (
	AgentRunFollowupRecheckEvidence      = "recheck_evidence"
	AgentRunFollowupWaitMoreInteractions = "wait_more_interactions"
	AgentRunFollowupCreateNote           = "create_followup_note"
)

// AgentRunFollowupRequest represents a request to create a follow-up action
type AgentRunFollowupRequest struct {
	ActionType     string `json:"action_type" binding:"required"`
	Reason         string `json:"reason" binding:"required"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
}

// AgentRunFollowupResponse represents the response for creating a follow-up action
type AgentRunFollowupResponse struct {
	AgentRunID     string         `json:"agent_run_id"`
	OperationID    string         `json:"operation_id"`
	ActionType     string         `json:"action_type"`
	Reason         string         `json:"reason"`
	ReviewPacketID string         `json:"review_packet_id,omitempty"`
	Operation      AgentOperation `json:"operation"`
	CreatedAt      time.Time      `json:"created_at"`
}

// IsAllowedAgentRunFollowupAction checks if the action type is allowed for follow-up
func IsAllowedAgentRunFollowupAction(action string) bool {
	switch action {
	case AgentRunFollowupRecheckEvidence,
		AgentRunFollowupWaitMoreInteractions,
		AgentRunFollowupCreateNote:
		return true
	default:
		return false
	}
}

// AgentRunReviewQueueItem represents an agent run in the review queue
type AgentRunReviewQueueItem struct {
	ID                 string     `json:"id"`
	AgentID            string     `json:"agent_id,omitempty"`
	OperatorID         string     `json:"operator_id,omitempty"`
	CaseID             string     `json:"case_id,omitempty"`
	PayloadID          string     `json:"payload_id,omitempty"`
	Target             string     `json:"target,omitempty"`
	Status             string     `json:"status"`
	ReviewState        string     `json:"review_state"`
	EvidenceStrength   string     `json:"evidence_strength,omitempty"`
	InteractionCount   int        `json:"interaction_count"`
	OperationCount     int        `json:"operation_count"`
	FollowupCount      int        `json:"followup_count"`
	LastFollowupAction string     `json:"last_followup_action,omitempty"`
	LastReviewedAt     *time.Time `json:"last_reviewed_at,omitempty"`
	LastFollowupAt     *time.Time `json:"last_followup_at,omitempty"`
	LastReviewDecision string     `json:"last_review_decision,omitempty"`
	LastDecisionReason string     `json:"last_decision_reason,omitempty"`
	LastDecisionAt     *time.Time `json:"last_decision_at,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
	DetailURL          string     `json:"detail_url"`
	EvidenceURL        string     `json:"evidence_url,omitempty"`
}

// AgentRunReviewQueueResponse represents the response for listing review queue
type AgentRunReviewQueueResponse struct {
	Items      []AgentRunReviewQueueItem  `json:"items"`
	Total      int64                      `json:"total"`
	Page       int                        `json:"page"`
	PageSize   int                        `json:"page_size"`
	TotalPages int                        `json:"total_pages"`
	Summary    AgentRunReviewQueueSummary `json:"summary"`
}

// AgentRunReviewQueueSummary represents summary statistics for the review queue
type AgentRunReviewQueueSummary struct {
	Total           int64 `json:"total"`
	NotReviewed     int64 `json:"not_reviewed"`
	Reviewed        int64 `json:"reviewed"`
	FollowupCreated int64 `json:"followup_created"`
	NeedsAttention  int64 `json:"needs_attention"`
}

// AgentRunFollowupHistoryItem represents a follow-up action in history
type AgentRunFollowupHistoryItem struct {
	OperationID    string    `json:"operation_id"`
	ActionType     string    `json:"action_type"`
	Reason         string    `json:"reason,omitempty"`
	ReviewPacketID string    `json:"review_packet_id,omitempty"`
	RiskLevel      string    `json:"risk_level"`
	AuditRefID     string    `json:"audit_ref_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// AgentRunReviewDecisionRequest represents a request to record a review decision
type AgentRunReviewDecisionRequest struct {
	Decision       string `json:"decision" binding:"required"`
	Reason         string `json:"reason,omitempty"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
	EvidenceID     string `json:"evidence_id,omitempty"`
}

// AgentRunReviewDecisionResponse represents the response for recording a review decision
type AgentRunReviewDecisionResponse struct {
	AgentRunID     string                 `json:"agent_run_id"`
	OperationID    string                 `json:"operation_id"`
	Decision       string                 `json:"decision"`
	ReviewPacketID string                 `json:"review_packet_id,omitempty"`
	AuditRefID     string                 `json:"audit_ref_id,omitempty"`
	Operation      *AgentOperation        `json:"operation,omitempty"`
	Audit          map[string]interface{} `json:"audit,omitempty"`
}

// AgentRunReviewExportRequest represents a request to export review evidence package
type AgentRunReviewExportRequest struct {
	Format         string `json:"format" binding:"required"`
	ReviewPacketID string `json:"review_packet_id,omitempty"`
	IncludeAudit   bool   `json:"include_audit,omitempty"`
}

// AgentRunReviewExportResponse represents the response for exporting review evidence package
type AgentRunReviewExportResponse struct {
	AgentRunID     string                 `json:"agent_run_id"`
	Format         string                 `json:"format"`
	OperationID    string                 `json:"operation_id"`
	AuditRefID     string                 `json:"audit_ref_id,omitempty"`
	ReviewPacketID string                 `json:"review_packet_id,omitempty"`
	Decision       string                 `json:"decision,omitempty"`
	Content        string                 `json:"content,omitempty"`
	Package        map[string]interface{} `json:"package,omitempty"`
	GeneratedAt    time.Time              `json:"generated_at"`
}

// AgentRunReviewDeliveryRequest represents a request to deliver review evidence package to webhook
type AgentRunReviewDeliveryRequest struct {
	Format         string            `json:"format" binding:"required"`
	ReviewPacketID string            `json:"review_packet_id,omitempty"`
	WebhookURL     string            `json:"webhook_url" binding:"required"`
	Headers        map[string]string `json:"headers,omitempty"`
	IncludeAudit   bool              `json:"include_audit,omitempty"`
}

// AgentRunReviewDeliveryResponse represents the response for delivering review evidence package
type AgentRunReviewDeliveryResponse struct {
	AgentRunID        string    `json:"agent_run_id"`
	Format            string    `json:"format"`
	DeliveryID        string    `json:"delivery_id"`
	DeliveryOperation string    `json:"delivery_operation_id"`
	ExportOperationID string    `json:"export_operation_id,omitempty"`
	AuditRefID        string    `json:"audit_ref_id,omitempty"`
	DestinationHost   string    `json:"destination_host"`
	StatusCode        int       `json:"status_code"`
	Result            string    `json:"result"`
	DeliveredAt       time.Time `json:"delivered_at"`
}

// AgentRunReviewDeliveryHistoryResponse represents the response for listing review delivery history
type AgentRunReviewDeliveryHistoryResponse struct {
	AgentRunID string                              `json:"agent_run_id"`
	Summary    AgentRunReviewDeliverySummary       `json:"summary"`
	Items      []AgentRunReviewDeliveryHistoryItem `json:"items"`
}

// AgentRunReviewDeliverySummary represents the summary of delivery attempts
type AgentRunReviewDeliverySummary struct {
	Total     int `json:"total"`
	Delivered int `json:"delivered"`
	Failed    int `json:"failed"`
	Timeout   int `json:"timeout"`
}

// AgentRunReviewDeliveryHistoryItem represents a single delivery attempt in history
type AgentRunReviewDeliveryHistoryItem struct {
	DeliveryID          string    `json:"delivery_id,omitempty"`
	DeliveryOperationID string    `json:"delivery_operation_id"`
	ExportOperationID   string    `json:"export_operation_id,omitempty"`
	AuditRefID          string    `json:"audit_ref_id,omitempty"`
	Format              string    `json:"format"`
	Result              string    `json:"result"`
	DestinationHost     string    `json:"destination_host"`
	StatusCode          int       `json:"status_code,omitempty"`
	HeaderNames         []string  `json:"header_names,omitempty"`
	ErrorSummary        string    `json:"error_summary,omitempty"`
	CreatedAt           time.Time `json:"created_at"`
	DeliveredAt         time.Time `json:"delivered_at,omitempty"`
}
