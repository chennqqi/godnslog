package rebinding

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
)

// Resolver handles DNS rebinding resolution logic
type Resolver struct {
	config *RebindingConfig
	store  Store
}

// NewResolver creates a new rebinding resolver
func NewResolver(config *RebindingConfig, store Store) *Resolver {
	if config == nil {
		config = DefaultRebindingConfig()
	}
	return &Resolver{
		config: config,
		store:  store,
	}
}

// DefaultRebindingConfig returns default rebinding configuration
func DefaultRebindingConfig() *RebindingConfig {
	return &RebindingConfig{
		DefaultTTL:      60,
		MaxStages:       5,
		EnableC2:        false,
		RequireAuth:     true,
		AuditC2:         true,
	}
}

// Resolve resolves a DNS query according to rebinding rules
func (r *Resolver) Resolve(ctx context.Context, domain, sourceIP string) (*ResolutionResult, error) {
	// Find matching rule
	rule, err := r.store.GetRebindingRuleByDomain(ctx, domain)
	if err != nil {
		return nil, fmt.Errorf("failed to get rebinding rule: %w", err)
	}

	if rule == nil || !rule.IsEnabled {
		// No rebinding rule, return default resolution
		return &ResolutionResult{
			IP:       "",
			TTL:      r.config.DefaultTTL,
			Stage:    0,
			IsRebind: false,
		}, nil
	}

	// Get or create session
	session, err := r.getOrCreateSession(ctx, rule.ID, sourceIP)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	// Get current stage
	currentStage := rule.GetCurrentStage(session.CurrentStage)
	if currentStage == nil {
		// No more stages, return last stage
		currentStage = &rule.Stages[len(rule.Stages)-1]
	}

	// Check if we should advance to next stage
	if r.shouldAdvanceStage(currentStage, session) {
		nextStage := rule.GetNextStage(session.CurrentStage)
		if nextStage != nil {
			session.CurrentStage++
			session.HitCount = 0
			currentStage = nextStage
			if err := r.store.UpdateRebindingSession(ctx, session); err != nil {
				return nil, fmt.Errorf("failed to update session: %w", err)
			}
		}
	}

	// Update hit count
	session.HitCount++
	session.LastHit = time.Now()
	if err := r.store.UpdateRebindingSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to update session: %w", err)
	}

	// Update stage hit count
	currentStage.HitCount++
	if err := r.store.UpdateRebindingRule(ctx, rule); err != nil {
		// Log but don't fail
		fmt.Printf("Failed to update rule: %v", err)
	}

	// Return resolution
	return &ResolutionResult{
		IP:       currentStage.TargetIP,
		TTL:      currentStage.TTL,
		Stage:    session.CurrentStage,
		IsRebind: true,
		RuleID:   rule.ID,
	}, nil
}

// shouldAdvanceStage checks if we should advance to next stage
func (r *Resolver) shouldAdvanceStage(stage *Stage, session *RebindingSession) bool {
	if stage.MaxHits > 0 && session.HitCount >= stage.MaxHits {
		return true
	}

	// Check condition-based advancement
	if stage.Condition != "" {
		// In production, implement condition evaluation
		// For MVP, use hit count only
		return false
	}

	return false
}

// getOrCreateSession gets or creates a rebinding session
func (r *Resolver) getOrCreateSession(ctx context.Context, ruleID, sourceIP string) (*RebindingSession, error) {
	// Try to get existing session
	session, err := r.store.GetRebindingSession(ctx, ruleID, sourceIP)
	if err == nil && session != nil {
		return session, nil
	}

	// Create new session
	session = &RebindingSession{
		ID:           generateSessionID(),
		RuleID:       ruleID,
		SourceIP:     sourceIP,
		CurrentStage: 0,
		HitCount:     0,
		StartedAt:    time.Now(),
		LastHit:      time.Now(),
	}

	if err := r.store.CreateRebindingSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// CreateScenarioRule creates a rebinding rule from a predefined scenario
func (r *Resolver) CreateScenarioRule(ctx context.Context, domain string, scenario RebindingScenario) (*RebindingRule, error) {
	var stages []Stage

	switch scenario {
	case ScenarioBrowserRebinding:
		stages = []Stage{
			{Order: 0, TargetIP: "127.0.0.1", TTL: 10, MaxHits: 1, Description: "Localhost (first request)"},
			{Order: 1, TargetIP: "192.168.1.1", TTL: 60, MaxHits: 0, Description: "Internal IP (rebinding)"},
		}
	case ScenarioCloudMetadata:
		stages = []Stage{
			{Order: 0, TargetIP: "169.254.169.254", TTL: 10, MaxHits: 1, Description: "Cloud metadata endpoint"},
			{Order: 1, TargetIP: "192.168.1.1", TTL: 60, MaxHits: 0, Description: "Internal IP after detection"},
		}
	case ScenarioInternalManagement:
		stages = []Stage{
			{Order: 0, TargetIP: "192.168.1.1", TTL: 30, MaxHits: 3, Description: "Management interface"},
			{Order: 1, TargetIP: "10.0.0.1", TTL: 60, MaxHits: 0, Description: "Internal network"},
		}
	case ScenarioIoTDevice:
		stages = []Stage{
			{Order: 0, TargetIP: "192.168.0.1", TTL: 60, MaxHits: 5, Description: "IoT gateway"},
			{Order: 1, TargetIP: "127.0.0.1", TTL: 10, MaxHits: 0, Description: "Localhost after detection"},
		}
	case ScenarioRouterExploit:
		stages = []Stage{
			{Order: 0, TargetIP: "192.168.1.1", TTL: 30, MaxHits: 2, Description: "Router admin interface"},
			{Order: 1, TargetIP: "127.0.0.1", TTL: 10, MaxHits: 0, Description: "Localhost for exploitation"},
		}
	default:
		return nil, fmt.Errorf("unknown scenario: %s", scenario)
	}

	rule := &RebindingRule{
		ID:        generateRuleID(),
		Domain:    domain,
		Stages:    stages,
		IsEnabled: true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := r.store.CreateRebindingRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return rule, nil
}

// ResolutionResult represents the result of a DNS resolution
type ResolutionResult struct {
	IP       string `json:"ip"`
	TTL      int    `json:"ttl"`
	Stage    int    `json:"stage"`
	IsRebind bool   `json:"is_rebind"`
	RuleID   string `json:"rule_id,omitempty"`
}

// ValidateIP validates an IP address
func ValidateIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// generateRuleID generates a unique rule ID
func generateRuleID() string {
	return fmt.Sprintf("rule-%d", time.Now().UnixNano())
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("session-%d", time.Now().UnixNano())
}
