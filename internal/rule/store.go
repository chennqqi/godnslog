package rule

import (
	"context"
	"fmt"
	"time"

	"xorm.io/xorm"
)

// XormStore implements the Store interface using XORM
type XormStore struct {
	db *xorm.Engine
}

// NewXormStore creates a new XORM-based rule store
func NewXormStore(db *xorm.Engine) *XormStore {
	return &XormStore{db: db}
}

// GetEnabledRules retrieves all enabled rules from the database
func (s *XormStore) GetEnabledRules(ctx context.Context) ([]*Rule, error) {
	var rules []*Rule
	err := s.db.Where("enabled = ?", true).OrderBy("priority DESC").Find(&rules)
	if err != nil {
		return nil, fmt.Errorf("failed to get enabled rules: %w", err)
	}
	return rules, nil
}

// GetRule retrieves a rule by ID
func (s *XormStore) GetRule(ctx context.Context, id string) (*Rule, error) {
	var rule Rule
	has, err := s.db.ID(id).Get(&rule)
	if err != nil {
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	if !has {
		return nil, fmt.Errorf("rule not found")
	}
	return &rule, nil
}

// CreateRule creates a new rule
func (s *XormStore) CreateRule(ctx context.Context, rule *Rule) error {
	rule.ID = generateID()
	rule.CreatedAt = time.Now()
	rule.UpdatedAt = time.Now()
	
	_, err := s.db.Insert(rule)
	if err != nil {
		return fmt.Errorf("failed to create rule: %w", err)
	}
	return nil
}

// UpdateRule updates an existing rule
func (s *XormStore) UpdateRule(ctx context.Context, rule *Rule) error {
	rule.UpdatedAt = time.Now()
	
	_, err := s.db.ID(rule.ID).Update(rule)
	if err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}
	return nil
}

// DeleteRule deletes a rule by ID
func (s *XormStore) DeleteRule(ctx context.Context, id string) error {
	_, err := s.db.ID(id).Delete(&Rule{})
	if err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}
	return nil
}

// ListRules lists all rules with pagination
func (s *XormStore) ListRules(ctx context.Context, page, pageSize int) ([]*Rule, int64, error) {
	var rules []*Rule
	total, err := s.db.Limit(pageSize, (page-1)*pageSize).FindAndCount(&rules)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list rules: %w", err)
	}
	return rules, total, nil
}

// SaveExecution saves a rule execution record
func (s *XormStore) SaveExecution(ctx context.Context, exec *RuleExecution) error {
	_, err := s.db.Insert(exec)
	if err != nil {
		return fmt.Errorf("failed to save execution: %w", err)
	}
	return nil
}

// GetExecutions retrieves execution records for a rule
func (s *XormStore) GetExecutions(ctx context.Context, ruleID string, page, pageSize int) ([]*RuleExecution, int64, error) {
	var executions []*RuleExecution
	total, err := s.db.Where("rule_id = ?", ruleID).
		Limit(pageSize, (page-1)*pageSize).
		OrderBy("executed_at DESC").
		FindAndCount(&executions)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get executions: %w", err)
	}
	return executions, total, nil
}
