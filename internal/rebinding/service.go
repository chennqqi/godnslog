package rebinding

import (
	"errors"
	"fmt"
	"time"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrRebindingRuleNotFound = errors.New("rebinding rule not found")
)

// Service provides rebinding management services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new rebinding service
func NewService(engine *xorm.Engine) *Service {
	return &Service{
		engine: engine,
	}
}

// CreateRebindingRule creates a new rebinding rule
func (s *Service) CreateRebindingRule(rule *models.RebindingRule) error {
	if rule.ID == "" {
		rule.ID = models.GenerateID()
	}
	if rule.CreatedAt.IsZero() {
		rule.CreatedAt = time.Now()
	}
	if rule.UpdatedAt.IsZero() {
		rule.UpdatedAt = time.Now()
	}

	_, err := s.engine.Insert(rule)
	return err
}

// GetRebindingRule retrieves a rebinding rule by its ID
func (s *Service) GetRebindingRule(id string) (*models.RebindingRule, error) {
	var rule models.RebindingRule
	has, err := s.engine.ID(id).Get(&rule)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrRebindingRuleNotFound
	}
	return &rule, nil
}

// ListRebindingRules retrieves rebinding rules with pagination
func (s *Service) ListRebindingRules(page, pageSize int) ([]models.RebindingRule, int64, error) {
	var rules []models.RebindingRule
	total, err := s.engine.Count(&models.RebindingRule{})
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := s.engine.Limit(pageSize, offset).Find(&rules); err != nil {
		return nil, 0, err
	}

	return rules, total, nil
}

// UpdateRebindingRule updates an existing rebinding rule
func (s *Service) UpdateRebindingRule(rule *models.RebindingRule) error {
	rule.UpdatedAt = time.Now()
	_, err := s.engine.ID(rule.ID).Update(rule)
	return err
}

// DeleteRebindingRule deletes a rebinding rule by its ID
func (s *Service) DeleteRebindingRule(id string) error {
	_, err := s.engine.ID(id).Delete(&models.RebindingRule{})
	return err
}

// ListRebindingSessions retrieves sessions for a specific rebinding rule
func (s *Service) ListRebindingSessions(ruleID string) ([]models.RebindingSession, error) {
	var sessions []models.RebindingSession
	err := s.engine.Where("rule_id = ?", ruleID).Find(&sessions)
	return sessions, err
}

// ProcessDNSQuery handles a DNS query for a rebinding rule domain.
// It finds or creates a session for the source IP, determines the current stage,
// advances the stage if hit count exceeds max_hits, and returns the target IP.
func (s *Service) ProcessDNSQuery(domain, sourceIP string) (string, *models.RebindingSession, error) {
	// Find the rebinding rule for this domain
	var rule models.RebindingRule
	has, err := s.engine.Where("domain = ? AND is_enabled = ?", domain, true).Get(&rule)
	if err != nil {
		return "", nil, fmt.Errorf("failed to query rebinding rule: %w", err)
	}
	if !has {
		return "", nil, ErrRebindingRuleNotFound
	}

	if len(rule.Stages) == 0 {
		return "", nil, fmt.Errorf("rule has no stages configured")
	}

	// Find or create session for this source IP
	session, err := s.findOrCreateSession(rule.ID, sourceIP)
	if err != nil {
		return "", nil, fmt.Errorf("failed to manage session: %w", err)
	}

	// Get current stage
	currentStage := rule.GetCurrentStage(session.CurrentStage)
	if currentStage == nil {
		// Reset to first stage if out of bounds
		session.CurrentStage = 0
		currentStage = &rule.Stages[0]
	}

	// Increment hit count
	session.HitCount++
	session.LastHit = time.Now()

	// Check if we should advance to next stage
	if currentStage.MaxHits > 0 && session.HitCount >= currentStage.MaxHits {
		nextStage := rule.GetNextStage(session.CurrentStage)
		if nextStage != nil {
			session.CurrentStage++
			currentStage = nextStage
			session.HitCount = 0
		}
	}

	// Update session in DB
	_, _ = s.engine.ID(session.ID).Update(session)

	return currentStage.TargetIP, session, nil
}

// findOrCreateSession finds an existing session or creates a new one for the source IP
func (s *Service) findOrCreateSession(ruleID, sourceIP string) (*models.RebindingSession, error) {
	var session models.RebindingSession
	has, err := s.engine.Where("rule_id = ? AND source_ip = ?", ruleID, sourceIP).Get(&session)
	if err != nil {
		return nil, err
	}
	if !has {
		session = models.RebindingSession{
			ID:           models.GenerateID(),
			RuleID:       ruleID,
			SourceIP:     sourceIP,
			CurrentStage: 0,
			HitCount:     0,
			StartedAt:    time.Now(),
			LastHit:      time.Now(),
		}
		if _, err := s.engine.Insert(&session); err != nil {
			return nil, err
		}
	}
	return &session, nil
}

// GetPredefinedScenarios returns the 5 predefined rebinding scenarios with their stage configurations
func GetPredefinedScenarios() []PredefinedScenario {
	return []PredefinedScenario{
		{
			Name:        string(models.ScenarioBrowserRebinding),
			Description: "Classic browser-based DNS rebinding: first resolve to attacker IP, then to 127.0.0.1",
			Stages: models.Stages{
				{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1, Description: "Initial resolve to attacker IP"},
				{Order: 1, TargetIP: "127.0.0.1", TTL: 0, MaxHits: 0, Description: "Rebind to localhost"},
			},
		},
		{
			Name:        string(models.ScenarioCloudMetadata),
			Description: "Rebind to cloud metadata endpoint (169.254.169.254) to extract cloud credentials",
			Stages: models.Stages{
				{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1, Description: "Initial resolve to attacker IP"},
				{Order: 1, TargetIP: "169.254.169.254", TTL: 0, MaxHits: 0, Description: "Rebind to AWS/GCP metadata endpoint"},
			},
		},
		{
			Name:        string(models.ScenarioInternalManagement),
			Description: "Rebind to internal management interfaces (e.g., 192.168.1.1 admin panel)",
			Stages: models.Stages{
				{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1, Description: "Initial resolve to attacker IP"},
				{Order: 1, TargetIP: "192.168.1.1", TTL: 0, MaxHits: 0, Description: "Rebind to internal router admin"},
			},
		},
		{
			Name:        string(models.ScenarioIoTDevice),
			Description: "Rebind to IoT device management interface for credential extraction",
			Stages: models.Stages{
				{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1, Description: "Initial resolve to attacker IP"},
				{Order: 1, TargetIP: "10.0.0.1", TTL: 0, MaxHits: 0, Description: "Rebind to IoT device interface"},
			},
		},
		{
			Name:        string(models.ScenarioRouterExploit),
			Description: "Multi-stage rebinding: attacker IP -> router -> internal network pivot",
			Stages: models.Stages{
				{Order: 0, TargetIP: "1.2.3.4", TTL: 3, MaxHits: 1, Description: "Initial resolve to attacker IP"},
				{Order: 1, TargetIP: "192.168.0.1", TTL: 1, MaxHits: 2, Description: "Rebind to router for CSRF"},
				{Order: 2, TargetIP: "10.0.0.2", TTL: 0, MaxHits: 0, Description: "Pivot to internal service"},
			},
		},
	}
}

// PredefinedScenario represents a predefined rebinding scenario with stages
type PredefinedScenario struct {
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Stages      models.Stages `json:"stages"`
}

// CreateRuleFromScenario creates a rebinding rule from a predefined scenario
func (s *Service) CreateRuleFromScenario(scenarioName, domain string) (*models.RebindingRule, error) {
	scenarios := GetPredefinedScenarios()
	for _, sc := range scenarios {
		if sc.Name == scenarioName {
			rule := &models.RebindingRule{
				ID:        models.GenerateID(),
				Domain:    domain,
				Stages:    sc.Stages,
				IsEnabled: true,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			if _, err := s.engine.Insert(rule); err != nil {
				return nil, err
			}
			return rule, nil
		}
	}
	return nil, fmt.Errorf("scenario %s not found", scenarioName)
}
