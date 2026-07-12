package executor

import (
	"regexp"
	"strings"

	"github.com/chennqqi/godnslog/internal/models"
)

func extractField(interaction *models.Interaction, field string) string {
	switch field {
	case "domain":
		if interaction.Domain != nil {
			return *interaction.Domain
		}
	case "path":
		if interaction.Path != nil {
			return *interaction.Path
		}
	case "body":
		if interaction.Body != nil {
			return *interaction.Body
		}
	case "source_ip":
		return interaction.SourceIP
	case "raw_data":
		return interaction.RawData
	}
	return ""
}

func matchRule(value string, rule MatchRule) (bool, string) {
	switch rule.Type {
	case "regex":
		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			return false, ""
		}
		m := re.FindString(value)
		return m != "", m
	case "contains":
		return strings.Contains(value, rule.Pattern), rule.Pattern
	case "equals":
		return value == rule.Pattern, rule.Pattern
	}
	return false, ""
}
