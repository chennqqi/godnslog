package retention

import (
	"context"
	"fmt"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

// mockStore implements Store for testing
type mockStore struct {
	policies map[string]*RetentionPolicy
	jobs     map[string]*RetentionJob
	archives map[string]*Archive
}

func newMockStore() *mockStore {
	return &mockStore{
		policies: make(map[string]*RetentionPolicy),
		jobs:     make(map[string]*RetentionJob),
		archives: make(map[string]*Archive),
	}
}

func (m *mockStore) CreatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	m.policies[policy.ID] = policy
	return nil
}

func (m *mockStore) GetPolicy(ctx context.Context, id string) (*RetentionPolicy, error) {
	if p, ok := m.policies[id]; ok {
		return p, nil
	}
	return nil, fmt.Errorf("policy not found")
}

func (m *mockStore) ListPolicies(ctx context.Context) ([]RetentionPolicy, error) {
	var result []RetentionPolicy
	for _, p := range m.policies {
		result = append(result, *p)
	}
	return result, nil
}

func (m *mockStore) UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	m.policies[policy.ID] = policy
	return nil
}

func (m *mockStore) DeletePolicy(ctx context.Context, id string) error {
	delete(m.policies, id)
	return nil
}

func (m *mockStore) CreateJob(ctx context.Context, job *RetentionJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockStore) GetJob(ctx context.Context, id string) (*RetentionJob, error) {
	if j, ok := m.jobs[id]; ok {
		return j, nil
	}
	return nil, fmt.Errorf("job not found")
}

func (m *mockStore) ListJobs(ctx context.Context) ([]RetentionJob, error) {
	var result []RetentionJob
	for _, j := range m.jobs {
		result = append(result, *j)
	}
	return result, nil
}

func (m *mockStore) UpdateJob(ctx context.Context, job *RetentionJob) error {
	m.jobs[job.ID] = job
	return nil
}

func (m *mockStore) CreateArchive(ctx context.Context, archive *Archive) error {
	m.archives[archive.ID] = archive
	return nil
}

func (m *mockStore) GetArchive(ctx context.Context, id string) (*Archive, error) {
	if a, ok := m.archives[id]; ok {
		return a, nil
	}
	return nil, fmt.Errorf("archive not found")
}

func (m *mockStore) ListArchives(ctx context.Context) ([]Archive, error) {
	var result []Archive
	for _, a := range m.archives {
		result = append(result, *a)
	}
	return result, nil
}

func (m *mockStore) UpdateArchive(ctx context.Context, archive *Archive) error {
	m.archives[archive.ID] = archive
	return nil
}

// TestCreatePolicy tests creating a retention policy
func TestCreatePolicy(t *testing.T) {
	store := newMockStore()
	service := NewService(store)

	policy := &RetentionPolicy{
		ID:                  "policy-test-1",
		Name:                "Test Policy",
		Description:         "Test retention policy",
		ApplyToInteractions: true,
		RetentionDays:       30,
		IsEnabled:           true,
	}

	err := service.CreatePolicy(context.Background(), policy)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if policy.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if policy.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

// TestCreatePolicyValidation tests validation of retention policy fields
func TestCreatePolicyValidation(t *testing.T) {
	store := newMockStore()
	service := NewService(store)

	// Negative retention days
	err := service.CreatePolicy(context.Background(), &RetentionPolicy{
		ID:            "p1",
		RetentionDays: -1,
	})
	if err == nil {
		t.Fatal("Expected error for negative retention_days")
	}

	// Negative max records
	err = service.CreatePolicy(context.Background(), &RetentionPolicy{
		ID:         "p2",
		MaxRecords: -1,
	})
	if err == nil {
		t.Fatal("Expected error for negative max_records")
	}
}

// TestRunPolicyDisabled tests that running a disabled policy fails
func TestRunPolicyDisabled(t *testing.T) {
	store := newMockStore()
	service := NewService(store)

	policy := &RetentionPolicy{
		ID:            "policy-disabled",
		Name:          "Disabled Policy",
		RetentionDays: 30,
		IsEnabled:     false,
	}
	_ = service.CreatePolicy(context.Background(), policy)

	_, err := service.RunPolicy(context.Background(), "policy-disabled")
	if err == nil {
		t.Fatal("Expected error for disabled policy")
	}
}

// TestRunPolicyWithXorm tests running a retention policy with a real xorm engine
func TestRunPolicyWithXorm(t *testing.T) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	// Sync tables
	if err := engine.Sync2(new(InteractionRecord), new(CaseRecord), new(PayloadRecord), new(RetentionPolicy), new(RetentionJob)); err != nil {
		t.Fatalf("Failed to sync tables: %v", err)
	}

	// Insert old interactions
	oldTime := time.Now().AddDate(0, 0, -100)
	for i := 0; i < 5; i++ {
		_, _ = engine.Insert(&InteractionRecord{
			ID:        "old-i-" + string(rune('a'+i)),
			CreatedAt: oldTime,
		})
	}
	// Insert recent interactions
	for i := 0; i < 3; i++ {
		_, _ = engine.Insert(&InteractionRecord{
			ID:        "new-i-" + string(rune('a'+i)),
			CreatedAt: time.Now(),
		})
	}

	store := NewXormStore(engine)
	service := NewService(store)

	policy := &RetentionPolicy{
		ID:                  "policy-xorm-test",
		Name:                "Xorm Test Policy",
		ApplyToInteractions: true,
		RetentionDays:       30,
		IsEnabled:           true,
	}
	_ = service.CreatePolicy(context.Background(), policy)

	job, err := service.RunPolicy(context.Background(), "policy-xorm-test")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if job.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", job.Status)
	}
	if job.RecordsProcessed != 5 {
		t.Errorf("Expected 5 records processed, got %d", job.RecordsProcessed)
	}
	if job.RecordsDeleted != 5 {
		t.Errorf("Expected 5 records deleted, got %d", job.RecordsDeleted)
	}

	// Verify remaining interactions
	count, _ := engine.Count(new(InteractionRecord))
	if int(count) != 3 {
		t.Errorf("Expected 3 remaining interactions, got %d", int(count))
	}
}

// TestCreateArchive tests creating an archive
func TestCreateArchive(t *testing.T) {
	store := newMockStore()
	service := NewService(store)

	policy := &RetentionPolicy{
		ID:            "policy-archive",
		Name:          "Archive Test",
		RetentionDays: 30,
		IsEnabled:     true,
	}
	_ = service.CreatePolicy(context.Background(), policy)

	archive, err := service.CreateArchive(context.Background(), "policy-archive", "interactions")
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if archive.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", archive.Status)
	}
	if archive.Compression != "gzip" {
		t.Errorf("Expected compression 'gzip', got '%s'", archive.Compression)
	}
}

// TestRecordTableNames tests that record types map to correct tables
func TestRecordTableNames(t *testing.T) {
	if (InteractionRecord{}).TableName() != "interactions" {
		t.Error("InteractionRecord should map to 'interactions'")
	}
	if (CaseRecord{}).TableName() != "cases" {
		t.Error("CaseRecord should map to 'cases'")
	}
	if (PayloadRecord{}).TableName() != "payloads" {
		t.Error("PayloadRecord should map to 'payloads'")
	}
}
