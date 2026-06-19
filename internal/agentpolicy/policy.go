package agentpolicy

// ScopePolicy describes one agent permission scope and its operational risk.
type ScopePolicy struct {
	Scope          string   `json:"scope"`
	Name           string   `json:"name"`
	RiskLevel      string   `json:"risk_level"`
	DefaultAllowed bool     `json:"default_allowed"`
	HighRisk       bool     `json:"high_risk"`
	ToolNames      []string `json:"tool_names"`
	Description    string   `json:"description"`
}

// ScopeCatalog is the public agent policy catalog returned to API/UI callers.
type ScopeCatalog struct {
	Items          []ScopePolicy `json:"items"`
	DefaultScopes  []string      `json:"default_scopes"`
	HighRiskScopes []string      `json:"high_risk_scopes"`
}

// ByScope returns catalog items keyed by scope.
func (c ScopeCatalog) ByScope() map[string]ScopePolicy {
	byScope := make(map[string]ScopePolicy, len(c.Items))
	for _, item := range c.Items {
		byScope[item.Scope] = item
	}
	return byScope
}

var scopePolicies = []ScopePolicy{
	{
		Scope:          "agent:create_probe",
		Name:           "Create OAST Probe",
		RiskLevel:      "medium",
		DefaultAllowed: true,
		ToolNames:      []string{"create_oast_probe", "create_payload"},
		Description:    "Create Case and Payload resources for OAST validation.",
	},
	{
		Scope:          "agent:wait_interaction",
		Name:           "Wait For Interaction",
		RiskLevel:      "low",
		DefaultAllowed: true,
		ToolNames:      []string{"wait_for_interaction"},
		Description:    "Poll for DNS/HTTP OAST interactions.",
	},
	{
		Scope:          "agent:read_interactions",
		Name:           "Read Interactions",
		RiskLevel:      "low",
		DefaultAllowed: true,
		ToolNames:      []string{"list_interactions"},
		Description:    "Read captured interactions for attribution and analysis.",
	},
	{
		Scope:          "agent:summarize_evidence",
		Name:           "Summarize Evidence",
		RiskLevel:      "low",
		DefaultAllowed: true,
		ToolNames:      []string{"summarize_evidence"},
		Description:    "Generate structured evidence summaries.",
	},
	{
		Scope:          "agent:export_report",
		Name:           "Export Report",
		RiskLevel:      "low",
		DefaultAllowed: true,
		ToolNames:      []string{"export_report"},
		Description:    "Export evidence or review reports for downstream use.",
	},
	{
		Scope:          "agent:read_runs",
		Name:           "Read Agent Runs",
		RiskLevel:      "low",
		DefaultAllowed: true,
		ToolNames:      []string{"list_agent_runs", "get_agent_run"},
		Description:    "Read Agent Run status and operation history.",
	},
	{
		Scope:       "agent:revoke_token",
		Name:        "Revoke Token",
		RiskLevel:   "high",
		HighRisk:    true,
		ToolNames:   []string{"revoke_token"},
		Description: "Revoke API tokens. Requires explicit grant and high risk tolerance.",
	},
	{
		Scope:       "agent:delete_payload",
		Name:        "Delete Payload",
		RiskLevel:   "high",
		HighRisk:    true,
		Description: "Delete payload resources. Reserved for explicit high-risk future workflows.",
	},
	{
		Scope:       "agent:modify_config",
		Name:        "Modify Configuration",
		RiskLevel:   "critical",
		HighRisk:    true,
		Description: "Modify sensitive system configuration. Reserved for explicit critical-risk future workflows.",
	},
}

// ListScopes returns a stable copy of the agent policy catalog.
func ListScopes() ScopeCatalog {
	items := append([]ScopePolicy(nil), scopePolicies...)
	return ScopeCatalog{
		Items:          items,
		DefaultScopes:  DefaultScopes(),
		HighRiskScopes: HighRiskScopes(),
	}
}

// DefaultScopes returns scopes granted to a new agent key by default.
func DefaultScopes() []string {
	var scopes []string
	for _, policy := range scopePolicies {
		if policy.DefaultAllowed {
			scopes = append(scopes, policy.Scope)
		}
	}
	return scopes
}

// HighRiskScopes returns scopes that require explicit selection.
func HighRiskScopes() []string {
	var scopes []string
	for _, policy := range scopePolicies {
		if policy.HighRisk {
			scopes = append(scopes, policy.Scope)
		}
	}
	return scopes
}

// ValidateScopes checks that every supplied scope is an agent scope in the catalog.
func ValidateScopes(scopes []string) bool {
	for _, scope := range scopes {
		if _, ok := PolicyForScope(scope); !ok {
			return false
		}
	}
	return true
}

// PolicyForScope returns the policy for one scope.
func PolicyForScope(scope string) (ScopePolicy, bool) {
	for _, policy := range scopePolicies {
		if policy.Scope == scope {
			return policy, true
		}
	}
	return ScopePolicy{}, false
}
