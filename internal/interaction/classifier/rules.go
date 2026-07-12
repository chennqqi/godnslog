package classifier

import (
	"strings"

	"github.com/chennqqi/godnslog/internal/models"
)

// Rule defines a single classification rule with a match function.
type Rule struct {
	ID          string
	Name        string
	ExploitType string
	Confidence  string
	Priority    int
	MatchFunc   func(*models.Interaction) (bool, string) // returns (matched, evidence)
}

// Classification represents the result of matching a rule against an interaction.
type Classification struct {
	ExploitType string `json:"exploit_type"`
	Confidence  string `json:"confidence"`
	RuleID      string `json:"rule_id"`
	Evidence    string `json:"evidence"`
}

// defaultRules returns the full set of classification rules ordered by priority.
func defaultRules() []Rule {
	return []Rule{
		// HIGH confidence rules (priority 10-19)
		{
			ID: "log4shell-jndi-dns", Name: "Log4Shell JNDI DNS Lookup",
			ExploitType: ExploitLog4shell, Confidence: ConfidenceHigh, Priority: 10,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Domain == nil {
					return false, ""
				}
				d := *i.Domain
				if strings.Contains(d, "${jndi:") || strings.Contains(d, "jndi:ldap") || strings.Contains(d, "jndi:rmi") {
					return true, d
				}
				return false, ""
			},
		},
		{
			ID: "log4shell-jndi-http", Name: "Log4Shell JNDI in HTTP",
			ExploitType: ExploitLog4shell, Confidence: ConfidenceHigh, Priority: 11,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Body != nil && strings.Contains(*i.Body, "${jndi:") {
					return true, "body:" + truncate(*i.Body, 100)
				}
				for k, v := range i.Headers {
					if strings.Contains(v, "${jndi:") {
						return true, "header:" + k
					}
				}
				return false, ""
			},
		},
		{
			ID: "fastjson-body", Name: "Fastjson Deserialization",
			ExploitType: ExploitFastjson, Confidence: ConfidenceHigh, Priority: 12,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Body == nil {
					return false, ""
				}
				if strings.Contains(*i.Body, `{"@type":"`) {
					return true, truncate(*i.Body, 100)
				}
				return false, ""
			},
		},
		{
			ID: "xxe-path", Name: "XXE via Path",
			ExploitType: ExploitXXE, Confidence: ConfidenceHigh, Priority: 13,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Path == nil {
					return false, ""
				}
				p := *i.Path
				if p == "/xxe" || strings.HasPrefix(p, "/xxe/") || strings.HasPrefix(p, "/xml") {
					return true, p
				}
				return false, ""
			},
		},
		{
			ID: "rce-path", Name: "RCE via Path",
			ExploitType: ExploitRCE, Confidence: ConfidenceHigh, Priority: 14,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Path == nil {
					return false, ""
				}
				p := *i.Path
				if p == "/cmd" || strings.HasPrefix(p, "/cmd/") || strings.HasPrefix(p, "/exec") {
					return true, p
				}
				return false, ""
			},
		},
		{
			ID: "deserialization-java", Name: "Java Deserialization",
			ExploitType: ExploitDeserialization, Confidence: ConfidenceHigh, Priority: 15,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Body == nil {
					return false, ""
				}
				b := *i.Body
				if strings.Contains(b, "aced0005") || strings.Contains(b, "rO0AB") {
					return true, truncate(b, 100)
				}
				return false, ""
			},
		},

		// MEDIUM confidence rules (priority 20-29)
		{
			ID: "ssrf-path", Name: "SSRF via Path",
			ExploitType: ExploitSSRF, Confidence: ConfidenceMedium, Priority: 20,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Path == nil {
					return false, ""
				}
				p := *i.Path
				if strings.Contains(p, "/redirect") || strings.Contains(p, "/proxy") ||
					strings.Contains(p, "/fetch") || strings.Contains(p, "/curl") {
					return true, p
				}
				return false, ""
			},
		},
		{
			ID: "sqli-path", Name: "SQLi via Path",
			ExploitType: ExploitSQLi, Confidence: ConfidenceMedium, Priority: 21,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Path == nil {
					return false, ""
				}
				p := *i.Path
				if strings.Contains(p, "sql") || strings.Contains(p, "?id=") ||
					strings.Contains(p, "?page=") || strings.Contains(p, "?query=") {
					return true, p
				}
				return false, ""
			},
		},
		{
			ID: "dns-rce-domain", Name: "RCE via DNS Domain",
			ExploitType: ExploitRCE, Confidence: ConfidenceMedium, Priority: 22,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Domain == nil {
					return false, ""
				}
				d := *i.Domain
				if strings.Contains(d, "rce") || strings.Contains(d, "cmd") ||
					strings.Contains(d, "exec") || strings.Contains(d, "whoami") {
					return true, d
				}
				return false, ""
			},
		},
		{
			ID: "dns-sqli-domain", Name: "SQLi via DNS Domain",
			ExploitType: ExploitSQLi, Confidence: ConfidenceMedium, Priority: 23,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Domain == nil {
					return false, ""
				}
				d := *i.Domain
				if strings.Contains(d, "sql") || strings.Contains(d, "sqli") ||
					strings.Contains(d, "dbname") || strings.Contains(d, "table") {
					return true, d
				}
				return false, ""
			},
		},
		{
			ID: "dns-xxe-domain", Name: "XXE via DNS Domain",
			ExploitType: ExploitXXE, Confidence: ConfidenceMedium, Priority: 24,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Domain == nil {
					return false, ""
				}
				d := *i.Domain
				if strings.Contains(d, "xxe") || strings.Contains(d, "dtd") || strings.Contains(d, "entity") {
					return true, d
				}
				return false, ""
			},
		},

		// LOW confidence rules (priority 30+)
		{
			ID: "dns-ldap-domain", Name: "LDAP via DNS Domain",
			ExploitType: ExploitLDAP, Confidence: ConfidenceLow, Priority: 30,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Domain == nil {
					return false, ""
				}
				d := *i.Domain
				if strings.Contains(d, "ldap") {
					return true, d
				}
				return false, ""
			},
		},
		{
			ID: "path-xss", Name: "XSS via Path",
			ExploitType: ExploitXSS, Confidence: ConfidenceLow, Priority: 31,
			MatchFunc: func(i *models.Interaction) (bool, string) {
				if i.Path == nil {
					return false, ""
				}
				p := *i.Path
				if strings.Contains(p, "/xss") || strings.Contains(p, "/callback") {
					return true, p
				}
				return false, ""
			},
		},
	}
}

// truncate truncates a string to the given maximum length, appending "..." if truncated.
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
