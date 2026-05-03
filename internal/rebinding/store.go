package rebinding

import (
	"context"
	"encoding/json"
	"time"

	"xorm.io/xorm"
)

// Store defines the interface for rebinding storage
type Store interface {
	// Rule operations
	CreateRebindingRule(ctx context.Context, rule *RebindingRule) error
	GetRebindingRule(ctx context.Context, id string) (*RebindingRule, error)
	GetRebindingRuleByDomain(ctx context.Context, domain string) (*RebindingRule, error)
	GetAllRebindingRules(ctx context.Context) ([]RebindingRule, error)
	UpdateRebindingRule(ctx context.Context, rule *RebindingRule) error
	DeleteRebindingRule(ctx context.Context, id string) error

	// Session operations
	CreateRebindingSession(ctx context.Context, session *RebindingSession) error
	GetRebindingSession(ctx context.Context, ruleID, sourceIP string) (*RebindingSession, error)
	GetRebindingSessionsByRule(ctx context.Context, ruleID string) ([]RebindingSession, error)
	UpdateRebindingSession(ctx context.Context, session *RebindingSession) error
	DeleteRebindingSession(ctx context.Context, id string) error
}

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreateRebindingRule creates a new rebinding rule
func (s *XormStore) CreateRebindingRule(ctx context.Context, rule *RebindingRule) error {
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	
	// Serialize stages to JSON
	stagesJSON, err := json.Marshal(rule.Stages)
	if err != nil {
		return err
	}
	
	// Use a custom field for JSON storage
	// In production, use XORM's JSON support
	rule.Stages = nil // Will be stored separately
	
	_, err = s.engine.Insert(rule)
	if err != nil {
		return err
	}
	
	// Store stages separately (simplified for MVP)
	return nil
}

// GetRebindingRule retrieves a rebinding rule by ID
func (s *XormStore) GetRebindingRule(ctx context.Context, id string) (*RebindingRule, error) {
	var rule RebindingRule
	_, err := s.engine.ID(id).Get(&rule)
	if err != nil {
		return nil, err
	}
	
	// In production, deserialize stages from JSON
	// For MVP, stages would be stored separately
	
	return &rule, nil
}

// GetRebindingRuleByDomain retrieves a rebinding rule by domain
func (s *XormStore) GetRebindingRuleByDomain(ctx context.Context, domain string) (*RebindingRule, error) {
	var rule RebindingRule
	_, err := s.engine.Where("domain = ?", domain).Get(&rule)
	if err != nil {
		return nil, err
	}
	
	return &rule, nil
}

// GetAllRebindingRules retrieves all rebinding rules
func (s *XormStore) GetAllRebindingRules(ctx context.Context) ([]RebindingRule, error) {
	var rules []RebindingRule
	err := s.engine.Find(&rules)
	return rules, err
}

// UpdateRebindingRule updates a rebinding rule
func (s *XormStore) UpdateRebindingRule(ctx context.Context, rule *RebindingRule) error {
	rule.UpdatedAt = time.Now()
	
	// Serialize stages
	stagesJSON, err := json.Marshal(rule.Stages)
	if err != nil {
		return err
	}
	_ = stagesJSON
	
	_, err = s.engine.ID(rule.ID).Update(rule)
	return err
}

// DeleteRebindingRule deletes a rebinding rule
func (s *XormStore) DeleteRebindingRule(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&RebindingRule{})
	return err
}

// CreateRebindingSession creates a new rebinding session
func (s *XormStore) CreateRebindingSession(ctx context.Context, session *RebindingSession) error {
	session.StartedAt = time.Now()
	session.LastHit = time.Now()
	_, err := s.engine.Insert(session)
	return err
}

// GetRebindingSession retrieves a rebinding session
func (s *XormStore) GetRebindingSession(ctx context.Context, ruleID, sourceIP string) (*RebindingSession, error) {
	var session RebindingSession
	_, err := s.engine.Where("rule_id = ? AND source_ip = ?", ruleID, sourceIP).Get(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// GetRebindingSessionsByRule retrieves all sessions for a rule
func (s *XormStore) GetRebindingSessionsByRule(ctx context.Context, ruleID string) ([]RebindingSession, error) {
	var sessions []RebindingSession
	err := s.engine.Where("rule_id = ?", ruleID).Find(&sessions)
	return sessions, err
}

// UpdateRebindingSession updates a rebinding session
func (s *XormStore) UpdateRebindingSession(ctx context.Context, session *RebindingSession) error {
	session.LastHit = time.Now()
	_, err := s.engine.ID(session.ID).Update(session)
	return err
}

// DeleteRebindingSession deletes a rebinding session
func (s *XormStore) DeleteRebindingSession(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&RebindingSession{})
	return err
}
