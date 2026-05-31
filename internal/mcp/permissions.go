package mcp

// RiskLevel represents the risk level of an MCP tool operation
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// ToolPermission defines the required scope and risk level for an MCP tool
type ToolPermission struct {
	ToolName      string
	RequiredScope string
	RiskLevel     RiskLevel
}

// ToolPermissions maps MCP tool names to their required scopes and risk levels
var ToolPermissions = map[string]ToolPermission{
	"create_oast_probe": {
		ToolName:      "create_oast_probe",
		RequiredScope: "agent:create_probe",
		RiskLevel:     RiskLevelMedium,
	},
	"create_case": {
		ToolName:      "create_case",
		RequiredScope: "case:create",
		RiskLevel:     RiskLevelMedium,
	},
	"wait_for_interaction": {
		ToolName:      "wait_for_interaction",
		RequiredScope: "agent:wait_interaction",
		RiskLevel:     RiskLevelLow,
	},
	"list_interactions": {
		ToolName:      "list_interactions",
		RequiredScope: "agent:read_interactions",
		RiskLevel:     RiskLevelLow,
	},
	"summarize_evidence": {
		ToolName:      "summarize_evidence",
		RequiredScope: "agent:summarize_evidence",
		RiskLevel:     RiskLevelLow,
	},
	"export_report": {
		ToolName:      "export_report",
		RequiredScope: "agent:export_report",
		RiskLevel:     RiskLevelLow,
	},
	"list_agent_runs": {
		ToolName:      "list_agent_runs",
		RequiredScope: "agent:read_runs",
		RiskLevel:     RiskLevelLow,
	},
	"get_agent_run": {
		ToolName:      "get_agent_run",
		RequiredScope: "agent:read_runs",
		RiskLevel:     RiskLevelLow,
	},
	"revoke_token": {
		ToolName:      "revoke_token",
		RequiredScope: "agent:revoke_token",
		RiskLevel:     RiskLevelHigh,
	},
}

// GetToolPermission returns the permission configuration for a given tool name
func GetToolPermission(toolName string) (ToolPermission, bool) {
	perm, exists := ToolPermissions[toolName]
	return perm, exists
}

// RiskLevelOrder defines the risk level hierarchy for comparison
var RiskLevelOrder = map[RiskLevel]int{
	RiskLevelLow:      0,
	RiskLevelMedium:   1,
	RiskLevelHigh:     2,
	RiskLevelCritical: 3,
}

// IsRiskLevelAllowed checks if a given risk level is allowed based on tolerance
func IsRiskLevelAllowed(toolRisk, tolerance RiskLevel) bool {
	toolLevel, toolExists := RiskLevelOrder[toolRisk]
	toleranceLevel, toleranceExists := RiskLevelOrder[tolerance]

	if !toolExists || !toleranceExists {
		return false
	}

	return toolLevel <= toleranceLevel
}
