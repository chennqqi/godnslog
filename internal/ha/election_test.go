package ha

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

func setupTestEngine(t *testing.T) *xorm.Engine {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "ha_test.db")
	engine, err := xorm.NewEngine("sqlite", dbPath)
	if err != nil {
		t.Fatalf("failed to create engine: %v", err)
	}
	if err := engine.Sync2(new(LeaderElection)); err != nil {
		t.Fatalf("failed to sync schema: %v", err)
	}
	return engine
}

func TestTryAcquireLock(t *testing.T) {
	engine := setupTestEngine(t)
	store := &XormStore{engine: engine}
	ctx := context.Background()

	// First node acquires lock
	ok, err := store.TryAcquireLock(ctx, "node-1", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected node-1 to acquire lock")
	}

	// Second node should fail
	ok, err = store.TryAcquireLock(ctx, "node-2", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatal("expected node-2 to be rejected")
	}

	// Release by node-1
	if err := store.ReleaseLock(ctx, "node-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Node-2 should now acquire
	ok, err = store.TryAcquireLock(ctx, "node-2", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected node-2 to acquire lock after release")
	}
}

func TestGetLeader(t *testing.T) {
	engine := setupTestEngine(t)
	store := &XormStore{engine: engine}
	ctx := context.Background()

	// No leader initially
	leader, err := store.GetLeader(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if leader != nil {
		t.Fatal("expected nil leader")
	}

	// Acquire lock
	store.TryAcquireLock(ctx, "node-1", 30*time.Second)

	leader, err = store.GetLeader(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if leader == nil {
		t.Fatal("expected non-nil leader")
	}
	if leader.LeaderID != "node-1" {
		t.Errorf("expected leader node-1, got %s", leader.LeaderID)
	}
}

func TestReleaseLockOnlyByOwner(t *testing.T) {
	engine := setupTestEngine(t)
	store := &XormStore{engine: engine}
	ctx := context.Background()

	store.TryAcquireLock(ctx, "node-1", 30*time.Second)

	// node-2 tries to release node-1's lock — should not succeed
	if err := store.ReleaseLock(ctx, "node-2"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// node-1 should still hold the lock
	ok, err := store.TryAcquireLock(ctx, "node-1", 30*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatal("expected node-1 to still hold lock after node-2 attempted release")
	}
}
