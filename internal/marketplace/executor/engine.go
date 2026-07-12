package executor

import "github.com/chennqqi/godnslog/internal/models"

// MatchRule defines a matching condition
type MatchRule struct {
	Type    string `json:"type"`    // regex, contains, equals
	Pattern string `json:"pattern"` // matching pattern
	Field   string `json:"field"`   // domain, path, body, source_ip
}

// Action defines the action to take on match
type Action struct {
	Type  string `json:"type"`  // tag, block, notify
	Value string `json:"value"` // action parameter
}

// TemplateDef is a parsed template definition from marketplace Template.Content
type TemplateDef struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Protocols   []string   `json:"protocols"`
	Match       MatchRule  `json:"match"`
	Action      Action     `json:"action"`
	Severity    string     `json:"severity"`
}

// MatchResult represents a template match result
type MatchResult struct {
	TemplateName string `json:"template_name"`
	Severity     string `json:"severity"`
	Action       Action `json:"action"`
	Evidence     string `json:"evidence"`
}

// Engine matches interactions against loaded templates
type Engine struct {
	templates []TemplateDef
}

// NewEngine creates a new template engine
func NewEngine(templates []TemplateDef) *Engine {
	return &Engine{templates: templates}
}

// Match runs all templates against an interaction and returns matches
func (e *Engine) Match(interaction *models.Interaction) []MatchResult {
	var results []MatchResult
	for _, tmpl := range e.templates {
		// Protocol filter
		if len(tmpl.Protocols) > 0 {
			matched := false
			for _, p := range tmpl.Protocols {
				if p == interaction.Type {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		value := extractField(interaction, tmpl.Match.Field)
		if value == "" {
			continue
		}

		ok, evidence := matchRule(value, tmpl.Match)
		if ok {
			results = append(results, MatchResult{
				TemplateName: tmpl.Name,
				Severity:     tmpl.Severity,
				Action:       tmpl.Action,
				Evidence:     evidence,
			})
		}
	}
	return results
}
