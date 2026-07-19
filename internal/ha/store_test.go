package ha

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupHAStoreDB(t *testing.T) *xorm.Engine {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "ha_store_test.db")
	engine, err := xorm.NewEngine("sqlite", dbPath)
	require.NoError(t, err)
	require.NoError(t, engine.Sync2(new(ClusterNode), new(ClusterConfig), new(HealthCheck), new(LeaderElection)))
	return engine
}

func TestXormStore_CreateAndGetNode(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	node := &ClusterNode{
		ID:   "node-1",
		Name: "Test Node",
		Host: "10.0.0.1",
		Port: 8080,
		Role: "primary",
	}
	err := store.CreateNode(ctx, node)
	require.NoError(t, err)

	fetched, err := store.GetNode(ctx, "node-1")
	require.NoError(t, err)
	assert.Equal(t, "Test Node", fetched.Name)
	assert.Equal(t, "10.0.0.1", fetched.Host)
}

func TestXormStore_ListNode(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, store.CreateNode(ctx, &ClusterNode{
			ID:   fmt.Sprintf("node-%d", i),
			Name: fmt.Sprintf("Node %d", i),
			Host: fmt.Sprintf("10.0.0.%d", i),
		}))
	}

	nodes, err := store.ListNodes(ctx)
	require.NoError(t, err)
	assert.Len(t, nodes, 3)
}

func TestXormStore_UpdateNode(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	require.NoError(t, store.CreateNode(ctx, &ClusterNode{ID: "node-1", Name: "Original", Host: "10.0.0.1"}))
	require.NoError(t, store.UpdateNode(ctx, &ClusterNode{ID: "node-1", Name: "Updated", Host: "10.0.0.2"}))

	fetched, err := store.GetNode(ctx, "node-1")
	require.NoError(t, err)
	assert.Equal(t, "Updated", fetched.Name)
	assert.Equal(t, "10.0.0.2", fetched.Host)
}

func TestXormStore_DeleteNode(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	require.NoError(t, store.CreateNode(ctx, &ClusterNode{ID: "node-1", Name: "To Delete"}))
	require.NoError(t, store.DeleteNode(ctx, "node-1"))

	fetched, err := store.GetNode(ctx, "node-1")
	require.NoError(t, err)
	assert.Empty(t, fetched.ID) // deleted, so should return empty struct
}

func TestXormStore_CreateAndListConfig(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	cfg := &ClusterConfig{ID: "cfg-1", EnableFailover: true, FailoverTimeout: 30}
	err := store.CreateConfig(ctx, cfg)
	require.NoError(t, err)

	configs, err := store.ListConfigs(ctx)
	require.NoError(t, err)
	assert.Len(t, configs, 1)
	assert.True(t, configs[0].EnableFailover)
}

func TestXormStore_GetConfig_NotFound(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	cfg, err := store.GetConfig(ctx, "nonexistent")
	require.NoError(t, err)
	assert.Empty(t, cfg.ID)
	assert.Empty(t, cfg.BalanceAlgorithm)
}

func TestXormStore_CreateHealthCheck(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	check := &HealthCheck{ID: "h-1", NodeID: "node-1", Status: "healthy", ResponseTime: 5}
	err := store.CreateHealthCheck(ctx, check)
	require.NoError(t, err)

	// Verify by querying the engine directly
	var saved HealthCheck
	has, err := engine.ID("h-1").Get(&saved)
	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, "healthy", saved.Status)
	assert.Equal(t, int64(5), saved.ResponseTime)
}

func TestXormStore_LeaderElection_FullCycle(t *testing.T) {
	engine := setupHAStoreDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	// Initially no leader
	leader, err := store.GetLeader(ctx)
	require.NoError(t, err)
	assert.Nil(t, leader)

	// Acquire lock
	ok, err := store.TryAcquireLock(ctx, "node-1", 30*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)

	// Get leader
	leader, err = store.GetLeader(ctx)
	require.NoError(t, err)
	assert.NotNil(t, leader)
	assert.Equal(t, "node-1", leader.LeaderID)

	// Another node can't acquire
	ok, err = store.TryAcquireLock(ctx, "node-2", 30*time.Second)
	require.NoError(t, err)
	assert.False(t, ok)

	// Release lock
	err = store.ReleaseLock(ctx, "node-1")
	require.NoError(t, err)

	// node-2 can now acquire
	ok, err = store.TryAcquireLock(ctx, "node-2", 30*time.Second)
	require.NoError(t, err)
	assert.True(t, ok)
}
