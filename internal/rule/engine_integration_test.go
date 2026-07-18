package rule

import (
	"context"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/stretchr/testify/assert"
)

// mockStore implements Store for testing
type mockStore struct {
	rules      []*Rule
	executions []*RuleExecution
	err        error
}

func (m *mockStore) GetEnabledRules(ctx context.Context) ([]*Rule, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.rules, nil
}

func (m *mockStore) GetRule(ctx context.Context, id string) (*Rule, error) {
	return nil, nil
}

func (m *mockStore) CreateRule(ctx context.Context, rule *Rule) error {
	return nil
}

func (m *mockStore) UpdateRule(ctx context.Context, rule *Rule) error {
	return nil
}

func (m *mockStore) DeleteRule(ctx context.Context, id string) error {
	return nil
}

func (m *mockStore) ListRules(ctx context.Context, page, pageSize int) ([]*Rule, int64, error) {
	return nil, 0, nil
}

func (m *mockStore) SaveExecution(ctx context.Context, exec *RuleExecution) error {
	if m.err != nil {
		return m.err
	}
	m.executions = append(m.executions, exec)
	return nil
}

func (m *mockStore) GetExecutions(ctx context.Context, ruleID string, page, pageSize int) ([]*RuleExecution, int64, error) {
	return nil, 0, nil
}

// Helper to create a test interaction
func testInteraction() *interaction.Interaction {
	token := "test-token-123"
	path := "/admin/login"
	body := "sensitive data"
	ua := "curl/8.0"
	return &interaction.Interaction{
		ID:        "inter-1",
		Type:      "http",
		Token:     &token,
		SourceIP:  "192.168.1.100",
		Path:      &path,
		Body:      &body,
		UserAgent: &ua,
		Headers:   map[string]string{"X-Forwarded-For": "10.0.0.1"},
	}
}

func TestMatchRule_ProtocolFilter(t *testing.T) {
	inter := testInteraction()
	inter.Type = "dns"

	rule := &Rule{
		ID: "rule-1",
		Conditions: Conditions{
			Protocol: []string{"http"},
		},
	}

	engine := NewEngine(nil)
	matched, err := engine.matchRule(rule, inter)
	assert.NoError(t, err)
	assert.False(t, matched, "dns interaction should not match http filter")
}

func TestMatchRule_TokenFilter(t *testing.T) {
	inter := testInteraction()

	t.Run("matching token", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				Tokens: []string{"test-token-123"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("non-matching token", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				Tokens: []string{"other-token"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.False(t, matched)
	})
}

func TestMatchRule_SourceIPFilter(t *testing.T) {
	inter := testInteraction()

	t.Run("exact match", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				SourceIP: []string{"192.168.1.100"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("CIDR match", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				SourceIP: []string{"192.168.0.0/16"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("no match", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				SourceIP: []string{"10.0.0.0/8"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.False(t, matched)
	})
}

func TestMatchRule_PathFilter(t *testing.T) {
	inter := testInteraction()

	t.Run("match regex", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				Path: []string{"/admin/.*"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("no match regex", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				Path: []string{"/api/.*"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.False(t, matched)
	})

	t.Run("nil path skips filter", func(t *testing.T) {
		interCopy := *inter
		interCopy.Path = nil
		rule := &Rule{
			Conditions: Conditions{
				Path: []string{".*"},
			},
		}
		engine := NewEngine(nil)
		// When Path is nil, the path filter is skipped (not failed)
		matched, err := engine.matchRule(rule, &interCopy)
		assert.NoError(t, err)
		assert.True(t, matched)
	})
}

func TestMatchRule_HeaderFilter(t *testing.T) {
	inter := testInteraction()

	t.Run("match header", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				Headers: map[string][]string{
					"X-Forwarded-For": {"10.0.0.1"},
				},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("missing header", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				Headers: map[string][]string{
					"X-Nonexistent": {"value"},
				},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.False(t, matched)
	})
}

func TestMatchRule_BodyFilter(t *testing.T) {
	inter := testInteraction()

	rule := &Rule{
		Conditions: Conditions{
			Body: []string{"sensitive"},
		},
	}
	engine := NewEngine(nil)
	matched, err := engine.matchRule(rule, inter)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestMatchRule_KeywordsFilter(t *testing.T) {
	inter := testInteraction()

	rule := &Rule{
		Conditions: Conditions{
			Keywords: []string{"curl/8.0"},
		},
	}
	engine := NewEngine(nil)
	matched, err := engine.matchRule(rule, inter)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestMatchRule_CaseIDFilter(t *testing.T) {
	inter := testInteraction()
	caseID := "case-123"
	inter.CaseID = &caseID

	t.Run("matching case", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				CaseIDs: []string{"case-123"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.True(t, matched)
	})

	t.Run("non-matching case", func(t *testing.T) {
		rule := &Rule{
			Conditions: Conditions{
				CaseIDs: []string{"other-case"},
			},
		}
		engine := NewEngine(nil)
		matched, err := engine.matchRule(rule, inter)
		assert.NoError(t, err)
		assert.False(t, matched)
	})
}

func TestMatchRule_CombinedConditions(t *testing.T) {
	inter := testInteraction()
	caseID := "case-123"
	inter.CaseID = &caseID

	rule := &Rule{
		Conditions: Conditions{
			Protocol: []string{"http"},
			Tokens:   []string{"test-token-123"},
			SourceIP: []string{"192.168.0.0/16"},
			Path:     []string{"/admin/.*"},
			Keywords: []string{"sensitive"},
			CaseIDs:  []string{"case-123"},
		},
	}
	engine := NewEngine(nil)
	matched, err := engine.matchRule(rule, inter)
	assert.NoError(t, err)
	assert.True(t, matched)
}

func TestMatchRule_InvalidPathRegex(t *testing.T) {
	inter := testInteraction()

	rule := &Rule{
		Conditions: Conditions{
			Path: []string{"[invalid"},
		},
	}
	engine := NewEngine(nil)
	_, err := engine.matchRule(rule, inter)
	assert.Error(t, err)
}

func TestParseTime(t *testing.T) {
	assert.Equal(t, 0, parseTime(""))
	assert.Equal(t, 0, parseTime("invalid"))
	assert.Equal(t, 8*60, parseTime("08:00"))
	assert.Equal(t, 22*60+30, parseTime("22:30"))
}

func TestMatchTimeRange(t *testing.T) {
	engine := NewEngine(nil)

	t.Run("within range (same day)", func(t *testing.T) {
		tr := &TimeRange{Start: "00:00", End: "23:59"}
		assert.True(t, engine.matchTimeRange(tr))
	})

	t.Run("outside range", func(t *testing.T) {
		tr := &TimeRange{Start: "23:00", End: "23:30"}
		// Current time is likely not between 23:00 and 23:30 during tests
		// This test verifies logic, not actual time
		now := time.Now()
		hour, min, _ := now.Clock()
		currentMin := hour*60 + min
		startMin := 23*60 + 0
		endMin := 23*60 + 30

		expected := currentMin >= startMin && currentMin <= endMin
		assert.Equal(t, expected, engine.matchTimeRange(tr))
	})

	t.Run("overnight range", func(t *testing.T) {
		tr := &TimeRange{Start: "22:00", End: "06:00"}
		result := engine.matchTimeRange(tr)
		// Should be true during 22:00-23:59 or 00:00-06:00
		now := time.Now()
		hour, min, _ := now.Clock()
		currentMin := hour*60 + min
		expected := currentMin >= 22*60 || currentMin <= 6*60
		assert.Equal(t, expected, result)
	})
}

func TestCollectInteractionData(t *testing.T) {
	inter := testInteraction()
	engine := NewEngine(nil)
	data := engine.collectInteractionData(inter)

	assert.Contains(t, data, "http")
	assert.Contains(t, data, "192.168.1.100")
	assert.Contains(t, data, "test-token-123")
	assert.Contains(t, data, "/admin/login")
	assert.Contains(t, data, "curl/8.0")
	assert.Contains(t, data, "sensitive data")
	assert.Contains(t, data, "X-Forwarded-For:10.0.0.1")
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2)
}

func TestEvaluate_NoRules(t *testing.T) {
	store := &mockStore{rules: []*Rule{}}
	engine := NewEngine(store)
	inter := testInteraction()

	results, err := engine.Evaluate(context.Background(), inter)
	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestEvaluate_MatchingRule(t *testing.T) {
	store := &mockStore{
		rules: []*Rule{
			{
				ID: "rule-match",
				Conditions: Conditions{
					Protocol: []string{"http"},
					Keywords: []string{"sensitive"},
				},
			},
		},
	}
	engine := NewEngine(store)
	inter := testInteraction()

	results, err := engine.Evaluate(context.Background(), inter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.True(t, results[0].Matched)
	assert.Equal(t, "rule-match", results[0].RuleID)
}

func TestEvaluate_NonMatchingRule(t *testing.T) {
	store := &mockStore{
		rules: []*Rule{
			{
				ID: "rule-no-match",
				Conditions: Conditions{
					Protocol: []string{"smtp"},
				},
			},
		},
	}
	engine := NewEngine(store)
	inter := testInteraction()

	results, err := engine.Evaluate(context.Background(), inter)
	assert.NoError(t, err)
	assert.Len(t, results, 1)
	assert.False(t, results[0].Matched)
}

func TestEvaluate_MultipleRules(t *testing.T) {
	store := &mockStore{
		rules: []*Rule{
			{
				ID: "rule-1",
				Conditions: Conditions{
					Protocol: []string{"http"},
				},
			},
			{
				ID: "rule-2",
				Conditions: Conditions{
					Protocol: []string{"dns"},
				},
			},
			{
				ID: "rule-3",
				Conditions: Conditions{
					Protocol: []string{"http"},
					Keywords: []string{"sensitive"},
				},
			},
		},
	}
	engine := NewEngine(store)
	inter := testInteraction()

	results, err := engine.Evaluate(context.Background(), inter)
	assert.NoError(t, err)
	assert.Len(t, results, 3)
	assert.True(t, results[0].Matched)
	assert.False(t, results[1].Matched)
	assert.True(t, results[2].Matched)
}

func TestEvaluate_StoreError(t *testing.T) {
	store := &mockStore{rules: nil, err: assert.AnError}
	engine := NewEngine(store)
	inter := testInteraction()

	_, err := engine.Evaluate(context.Background(), inter)
	assert.Error(t, err)
}

func TestContains(t *testing.T) {
	assert.True(t, contains([]string{"a", "b", "c"}, "a"))
	assert.False(t, contains([]string{"a", "b", "c"}, "d"))
	assert.False(t, contains([]string{}, "a"))
}
