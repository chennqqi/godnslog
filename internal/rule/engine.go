package rule

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
)

// Engine evaluates rules against interactions
type Engine struct {
	ruleStore Store
}

// Store defines the interface for rule storage
type Store interface {
	GetEnabledRules(ctx context.Context) ([]*Rule, error)
	GetRule(ctx context.Context, id string) (*Rule, error)
	CreateRule(ctx context.Context, rule *Rule) error
	UpdateRule(ctx context.Context, rule *Rule) error
	DeleteRule(ctx context.Context, id string) error
	ListRules(ctx context.Context, page, pageSize int) ([]*Rule, int64, error)
	SaveExecution(ctx context.Context, exec *RuleExecution) error
	GetExecutions(ctx context.Context, ruleID string, page, pageSize int) ([]*RuleExecution, int64, error)
}

// NewEngine creates a new rule engine
func NewEngine(store Store) *Engine {
	return &Engine{
		ruleStore: store,
	}
}

// Evaluate evaluates all enabled rules against an interaction
func (e *Engine) Evaluate(ctx context.Context, inter *interaction.Interaction) ([]*RuleExecution, error) {
	rules, err := e.ruleStore.GetEnabledRules(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get rules: %w", err)
	}

	var results []*RuleExecution
	for _, rule := range rules {
		exec := &RuleExecution{
			ID:          generateID(),
			RuleID:      rule.ID,
			Interaction: inter.ID,
			ExecutedAt:  time.Now(),
		}

		matched, err := e.matchRule(rule, inter)
		if err != nil {
			exec.Error = err.Error()
			exec.Matched = false
		} else {
			exec.Matched = matched
		}

		if err := e.ruleStore.SaveExecution(ctx, exec); err != nil {
			return results, fmt.Errorf("failed to save execution: %w", err)
		}

		results = append(results, exec)
	}

	return results, nil
}

// matchRule checks if an interaction matches a rule's conditions
func (e *Engine) matchRule(rule *Rule, inter *interaction.Interaction) (bool, error) {
	conds := rule.Conditions

	// Protocol filter
	if len(conds.Protocol) > 0 {
		if !contains(conds.Protocol, inter.Type) {
			return false, nil
		}
	}

	// Token filter
	if len(conds.Tokens) > 0 && inter.Token != nil {
		if !contains(conds.Tokens, *inter.Token) {
			return false, nil
		}
	}

	// SourceIP filter
	if len(conds.SourceIP) > 0 {
		matched := false
		for _, ip := range conds.SourceIP {
			if strings.Contains(ip, "/") {
				// CIDR match
				if matchCIDR(ip, inter.SourceIP) {
					matched = true
					break
				}
			} else {
				// Exact match
				if ip == inter.SourceIP {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Path filter (HTTP only)
	if len(conds.Path) > 0 && inter.Path != nil {
		matched := false
		for _, pattern := range conds.Path {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return false, fmt.Errorf("invalid path regex %s: %w", pattern, err)
			}
			if re.MatchString(*inter.Path) {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Header filter
	if len(conds.Headers) > 0 && len(inter.Headers) > 0 {
		for key, values := range conds.Headers {
			headerValue, ok := inter.Headers[key]
			if !ok {
				return false, nil
			}
			for _, v := range values {
				if headerValue != v {
					return false, nil
				}
			}
		}
	}

	// Body filter
	if len(conds.Body) > 0 && inter.Body != nil {
		matched := false
		for _, pattern := range conds.Body {
			re, err := regexp.Compile(pattern)
			if err != nil {
				return false, fmt.Errorf("invalid body regex %s: %w", pattern, err)
			}
			if re.MatchString(*inter.Body) {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Keywords filter
	if len(conds.Keywords) > 0 {
		data := e.collectInteractionData(inter)
		matched := false
		for _, keyword := range conds.Keywords {
			if strings.Contains(data, keyword) {
				matched = true
				break
			}
		}
		if !matched {
			return false, nil
		}
	}

	// Case filter
	if len(conds.CaseIDs) > 0 {
		if inter.CaseID == nil || !contains(conds.CaseIDs, *inter.CaseID) {
			return false, nil
		}
	}

	// Time range filter
	if conds.TimeRange != nil {
		if !e.matchTimeRange(conds.TimeRange) {
			return false, nil
		}
	}

	return true, nil
}

// collectInteractionData collects all text data from an interaction for keyword matching
func (e *Engine) collectInteractionData(inter *interaction.Interaction) string {
	var data strings.Builder
	data.WriteString(inter.Type)
	data.WriteString(" ")
	data.WriteString(inter.SourceIP)
	if inter.Token != nil {
		data.WriteString(" ")
		data.WriteString(*inter.Token)
	}
	if inter.Domain != nil {
		data.WriteString(" ")
		data.WriteString(*inter.Domain)
	}
	if inter.Path != nil {
		data.WriteString(" ")
		data.WriteString(*inter.Path)
	}
	if inter.UserAgent != nil {
		data.WriteString(" ")
		data.WriteString(*inter.UserAgent)
	}
	if inter.Body != nil {
		data.WriteString(" ")
		data.WriteString(*inter.Body)
	}
	for k, v := range inter.Headers {
		data.WriteString(" ")
		data.WriteString(k)
		data.WriteString(":")
		data.WriteString(v)
		data.WriteString(" ")
	}
	return data.String()
}

// matchTimeRange checks if current time is within the specified time range
func (e *Engine) matchTimeRange(tr *TimeRange) bool {
	now := time.Now()
	hour, minute, _ := now.Clock()
	currentTime := hour*60 + minute

	startTime := parseTime(tr.Start)
	endTime := parseTime(tr.End)

	if startTime <= endTime {
		return currentTime >= startTime && currentTime <= endTime
	}
	// Handle overnight range (e.g., 22:00 - 06:00)
	return currentTime >= startTime || currentTime <= endTime
}

// parseTime parses HH:MM format to minutes since midnight
func parseTime(t string) int {
	if t == "" {
		return 0
	}
	parts := strings.Split(t, ":")
	if len(parts) != 2 {
		return 0
	}
	hour := 0
	minute := 0
	fmt.Sscanf(parts[0], "%d", &hour)
	fmt.Sscanf(parts[1], "%d", &minute)
	return hour*60 + minute
}

// matchCIDR checks if an IP matches a CIDR range
func matchCIDR(cidr, ip string) bool {
	// Simplified CIDR matching - in production use net.ParseCIDR
	// This is a placeholder for proper CIDR matching
	return strings.HasPrefix(ip, strings.TrimSuffix(cidr, "/32"))
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
