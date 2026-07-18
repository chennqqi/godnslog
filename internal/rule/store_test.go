package rule

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func setupRuleDB(t *testing.T) *xorm.Engine {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "rule_test.db")
	engine, err := xorm.NewEngine("sqlite", dbPath)
	require.NoError(t, err)
	require.NoError(t, engine.Sync2(new(Rule), new(RuleExecution)))
	return engine
}

func TestXormStore_CreateAndGetRule(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	rule := &Rule{
		Name:        "test-rule",
		Description: "A test rule",
		Enabled:     true,
		Priority:    10,
		Conditions: Conditions{
			Protocol: []string{"http"},
			Keywords: []string{"admin"},
		},
		Actions: Actions{
			DiscardNoise: true,
		},
	}

	err := store.CreateRule(ctx, rule)
	require.NoError(t, err)
	assert.NotEmpty(t, rule.ID)
	assert.False(t, rule.CreatedAt.IsZero())

	// Fetch by ID
	fetched, err := store.GetRule(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, "test-rule", fetched.Name)
	assert.True(t, fetched.Enabled)
	assert.Equal(t, 10, fetched.Priority)
}

func TestXormStore_GetEnabledRules(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	rules := []*Rule{
		{Name: "enabled-1", Enabled: true, Priority: 5},
		{Name: "disabled", Enabled: false, Priority: 1},
		{Name: "enabled-2", Enabled: true, Priority: 10},
	}
	for _, r := range rules {
		require.NoError(t, store.CreateRule(ctx, r))
	}

	enabled, err := store.GetEnabledRules(ctx)
	require.NoError(t, err)
	assert.Len(t, enabled, 2)
	// Should be ordered by priority DESC
	assert.Equal(t, "enabled-2", enabled[0].Name)
	assert.Equal(t, "enabled-1", enabled[1].Name)
}

func TestXormStore_UpdateRule(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	rule := &Rule{Name: "original", Enabled: true, Priority: 1}
	require.NoError(t, store.CreateRule(ctx, rule))

	rule.Name = "updated"
	rule.Priority = 100
	require.NoError(t, store.UpdateRule(ctx, rule))

	fetched, err := store.GetRule(ctx, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, "updated", fetched.Name)
	assert.Equal(t, 100, fetched.Priority)
}

func TestXormStore_DeleteRule(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	rule := &Rule{Name: "to-delete", Enabled: true}
	require.NoError(t, store.CreateRule(ctx, rule))

	require.NoError(t, store.DeleteRule(ctx, rule.ID))

	_, err := store.GetRule(ctx, rule.ID)
	assert.Error(t, err)
}

func TestXormStore_ListRules(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		require.NoError(t, store.CreateRule(ctx, &Rule{
			Name:     fmt.Sprintf("rule-%d", i),
			Enabled:  true,
			Priority: i,
		}))
	}

	// Page 1, size 3
	rules, total, err := store.ListRules(ctx, 1, 3)
	require.NoError(t, err)
	assert.Len(t, rules, 3)
	assert.Equal(t, int64(5), total)

	// Page 2, size 3
	rules2, total2, err := store.ListRules(ctx, 2, 3)
	require.NoError(t, err)
	assert.Len(t, rules2, 2)
	assert.Equal(t, int64(5), total2)
}

func TestXormStore_SaveAndGetExecutions(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	exec := &RuleExecution{
		ID:          "exec-1",
		RuleID:      "rule-1",
		Interaction: "inter-1",
		Matched:     true,
		ExecutedAt:  time.Now(),
	}
	require.NoError(t, store.SaveExecution(ctx, exec))

	execs, total, err := store.GetExecutions(ctx, "rule-1", 1, 10)
	require.NoError(t, err)
	assert.Len(t, execs, 1)
	assert.Equal(t, int64(1), total)
	assert.True(t, execs[0].Matched)
	assert.Equal(t, "inter-1", execs[0].Interaction)
	assert.Equal(t, "exec-1", execs[0].ID)
}

func TestXormStore_GetRule_NotFound(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	_, err := store.GetRule(ctx, "nonexistent")
	assert.Error(t, err)
}

func TestXormStore_GetExecutions_Empty(t *testing.T) {
	engine := setupRuleDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	execs, total, err := store.GetExecutions(ctx, "nonexistent", 1, 10)
	require.NoError(t, err)
	assert.Empty(t, execs)
	assert.Equal(t, int64(0), total)
}
