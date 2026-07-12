package classifier

import (
	"sort"

	"github.com/chennqqi/godnslog/internal/models"
)

// Classify returns the best classification for the given interaction.
// It first sorts rules by priority (ascending). The first HIGH confidence match
// wins and is returned immediately. If no HIGH match is found, the best
// (lowest priority number) MEDIUM or LOW match is returned. Returns nil if
// no rule matches.
func Classify(interaction *models.Interaction) *Classification {
	rules := defaultRules()
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	var bestMediumLow *Classification

	for _, rule := range rules {
		matched, evidence := rule.MatchFunc(interaction)
		if !matched {
			continue
		}

		c := &Classification{
			ExploitType: rule.ExploitType,
			Confidence:  rule.Confidence,
			RuleID:      rule.ID,
			Evidence:    evidence,
		}

		if rule.Confidence == ConfidenceHigh {
			return c
		}

		// Track the first medium/low match (lowest priority number
		// since we iterate in ascending priority order).
		if bestMediumLow == nil {
			bestMediumLow = c
		}
	}

	return bestMediumLow
}

// ClassifyAll returns all matching classifications for the given interaction,
// sorted by priority (ascending).
func ClassifyAll(interaction *models.Interaction) []*Classification {
	rules := defaultRules()
	sort.Slice(rules, func(i, j int) bool {
		return rules[i].Priority < rules[j].Priority
	})

	var results []*Classification

	for _, rule := range rules {
		matched, evidence := rule.MatchFunc(interaction)
		if !matched {
			continue
		}

		results = append(results, &Classification{
			ExploitType: rule.ExploitType,
			Confidence:  rule.Confidence,
			RuleID:      rule.ID,
			Evidence:    evidence,
		})
	}

	return results
}
