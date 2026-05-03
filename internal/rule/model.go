package rule

import (
	"time"
)

// Rule represents an automation rule for processing interactions
type Rule struct {
	ID          string    `xorm:"pk 'id'"`
	Name        string    `xorm:"name"`
	Description string    `xorm:"description"`
	Enabled     bool      `xorm:"enabled"`
	Priority    int       `xorm:"priority"`
	Conditions  Conditions `xorm:"conditions json"`
	Actions     Actions   `xorm:"actions json"`
	CreatedAt   time.Time `xorm:"created_at"`
	UpdatedAt   time.Time `xorm:"updated_at"`
}

// Conditions defines the matching criteria for a rule
type Conditions struct {
	// Protocol filter: dns, http, smtp, ldap
	Protocol []string `json:"protocol,omitempty"`
	
	// Token filter: match specific tokens
	Tokens []string `json:"tokens,omitempty"`
	
	// SourceIP filter: CIDR or exact match
	SourceIP []string `json:"source_ip,omitempty"`
	
	// Path filter: regex match for HTTP path
	Path []string `json:"path,omitempty"`
	
	// Header filter: key-value pairs
	Headers map[string][]string `json:"headers,omitempty"`
	
	// Body filter: regex match for body content
	Body []string `json:"body,omitempty"`
	
	// Keywords filter: match in any field
	Keywords []string `json:"keywords,omitempty"`
	
	// Case filter: match specific case IDs
	CaseIDs []string `json:"case_ids,omitempty"`
	
	// Risk level filter: low, medium, high, critical
	RiskLevels []string `json:"risk_levels,omitempty"`
	
	// Time range filter
	TimeRange *TimeRange `json:"time_range,omitempty"`
}

// TimeRange defines a time window for matching
type TimeRange struct {
	Start string `json:"start,omitempty"` // HH:MM format
	End   string `json:"end,omitempty"`   // HH:MM format
}

// Actions defines what to do when a rule matches
type Actions struct {
	// Notification actions
	Notifications []Notification `json:"notifications,omitempty"`
	
	// Tag actions
	Tags []TagAction `json:"tags,omitempty"`
	
	// Webhook forwarding
	Webhooks []Webhook `json:"webhooks,omitempty"`
	
	// Report generation
	Reports []Report `json:"reports,omitempty"`
	
	// Noise filtering
	DiscardNoise bool `json:"discard_noise,omitempty"`
}

// Notification defines a notification action
type Notification struct {
	Type     string                 `json:"type"` // feishu, wecom, dingtalk, slack, discord, telegram, email, webhook
	Channel  string                 `json:"channel"`
	Template string                 `json:"template"`
	Config   map[string]interface{} `json:"config"`
}

// TagAction defines a tag action
type TagAction struct {
	Add    []string `json:"add,omitempty"`
	Remove []string `json:"remove,omitempty"`
}

// Webhook defines a webhook forwarding action
type Webhook struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"` // POST, PUT
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

// Report defines a report generation action
type Report struct {
	Format string `json:"format"` // json, markdown, csv
	Title  string `json:"title"`
}

// RuleExecution represents a rule execution record
type RuleExecution struct {
	ID          string    `xorm:"pk 'id'"`
	RuleID      string    `xorm:"rule_id"`
	Interaction string    `xorm:"interaction_id"`
	Matched     bool      `xorm:"matched"`
	ExecutedAt  time.Time `xorm:"executed_at"`
	Error       string    `xorm:"error"`
}
