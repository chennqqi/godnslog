package rebinding

import "time"

// RebindingRule represents a DNS rebinding rule
type RebindingRule struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Domain      string    `json:"domain"`
	Stages      []Stage   `json:"stages" xorm:"stages_json"`
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
}

// Stage represents a rebinding stage
type Stage struct {
	Order       int    `json:"order"`
	TargetIP    string `json:"target_ip"`
	TTL         int    `json:"ttl"`          // Time to live in seconds
	HitCount    int    `json:"hit_count"`    // Number of hits for this stage
	MaxHits     int    `json:"max_hits"`     // Maximum hits before moving to next stage
	Condition   string `json:"condition"`    // Condition to trigger this stage
	Description string `json:"description"`
}

// RebindingConfig holds rebinding configuration
type RebindingConfig struct {
	DefaultTTL      int `json:"default_ttl"`       // Default TTL in seconds
	MaxStages       int `json:"max_stages"`         // Maximum number of stages
	EnableC2        bool `json:"enable_c2"`         // Enable DNS C2 (disabled by default)
	RequireAuth     bool `json:"require_auth"`      // Require authentication for C2
	AuditC2         bool `json:"audit_c2"`          // Audit all C2 operations
}

// RebindingSession represents an active rebinding session
type RebindingSession struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	RuleID      string    `json:"rule_id"`
	SourceIP    string    `json:"source_ip"`
	CurrentStage int     `json:"current_stage"`
	HitCount    int       `json:"hit_count"`
	StartedAt   time.Time `json:"started_at" xorm:"created"`
	LastHit     time.Time `json:"last_hit"`
}

// RebindingScenario represents predefined rebinding scenarios
type RebindingScenario string

const (
	ScenarioBrowserRebinding   RebindingScenario = "browser-rebinding"
	ScenarioCloudMetadata      RebindingScenario = "cloud-metadata"
	ScenarioInternalManagement RebindingScenario = "internal-management"
	ScenarioIoTDevice          RebindingScenario = "iot-device"
	ScenarioRouterExploit      RebindingScenario = "router-exploit"
)

// TableName returns the table name for RebindingRule
func (RebindingRule) TableName() string {
	return "rebinding_rules"
}

// TableName returns the table name for RebindingSession
func (RebindingSession) TableName() string {
	return "rebinding_sessions"
}

// GetStageCount returns the number of stages
func (r *RebindingRule) GetStageCount() int {
	return len(r.Stages)
}

// GetCurrentStage returns the current stage for a session
func (r *RebindingRule) GetCurrentStage(stageIndex int) *Stage {
	if stageIndex < 0 || stageIndex >= len(r.Stages) {
		return nil
	}
	return &r.Stages[stageIndex]
}

// GetNextStage returns the next stage
func (r *RebindingRule) GetNextStage(currentStage int) *Stage {
	nextIndex := currentStage + 1
	if nextIndex >= len(r.Stages) {
		return nil
	}
	return &r.Stages[nextIndex]
}
