package agentpolicy

import (
	"testing"
)

func TestListScopesContainsExpectedAgentScopes(t *testing.T) {
	catalog := ListScopes()
	byScope := catalog.ByScope()

	expected := []string{
		"agent:create_probe",
		"agent:wait_interaction",
		"agent:read_interactions",
		"agent:summarize_evidence",
		"agent:export_report",
		"agent:read_runs",
		"agent:revoke_token",
		"agent:delete_payload",
		"agent:modify_config",
	}

	for _, scope := range expected {
		if _, ok := byScope[scope]; !ok {
			t.Fatalf("expected scope %q in catalog", scope)
		}
	}

	if got := byScope["agent:create_probe"].RiskLevel; got != "medium" {
		t.Fatalf("expected create_probe risk medium, got %q", got)
	}
	if got := byScope["agent:wait_interaction"].RiskLevel; got != "low" {
		t.Fatalf("expected wait_interaction risk low, got %q", got)
	}
	if !byScope["agent:revoke_token"].HighRisk {
		t.Fatal("expected revoke_token to be high risk")
	}
	if byScope["agent:delete_payload"].DefaultAllowed {
		t.Fatal("expected delete_payload to be default denied")
	}
}

func TestDefaultAndHighRiskScopes(t *testing.T) {
	defaultScopes := DefaultScopes()
	highRiskScopes := HighRiskScopes()

	if contains(defaultScopes, "agent:revoke_token") {
		t.Fatal("default scopes must not include revoke_token")
	}
	if !contains(defaultScopes, "agent:create_probe") {
		t.Fatal("default scopes should include create_probe")
	}
	if !contains(defaultScopes, "agent:wait_interaction") {
		t.Fatal("default scopes should include wait_interaction")
	}
	if !contains(highRiskScopes, "agent:delete_payload") {
		t.Fatal("high risk scopes should include delete_payload")
	}
	if !contains(highRiskScopes, "agent:modify_config") {
		t.Fatal("high risk scopes should include modify_config")
	}
}

func TestValidateScopes(t *testing.T) {
	valid := []string{"agent:create_probe", "agent:read_runs", "agent:revoke_token"}
	if !ValidateScopes(valid) {
		t.Fatalf("expected valid scopes to pass validation")
	}

	invalid := []string{"agent:create_probe", "case:read"}
	if ValidateScopes(invalid) {
		t.Fatalf("expected non-agent scope to fail validation")
	}
}

func contains(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}
