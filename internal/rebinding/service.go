package rebinding

import (
	"errors"
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
