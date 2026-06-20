package retention

import (
	"context"
	"fmt"
	"time"
)

// Service handles retention and archival operations
type Service struct {
	store Store
}

// NewService creates a new retention service
func NewService(store Store) *Service {
	return &Service{store: store}
}

// CreatePolicy creates a new retention policy
func (s *Service) CreatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	if policy.RetentionDays < 0 {
		return fmt.Errorf("retention_days must be non-negative")
	}
	if policy.MaxRecords < 0 {
		return fmt.Errorf("max_records must be non-negative")
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	return s.store.CreatePolicy(ctx, policy)
}

// GetPolicy retrieves a policy by ID
func (s *Service) GetPolicy(ctx context.Context, id string) (*RetentionPolicy, error) {
	return s.store.GetPolicy(ctx, id)
}

// ListPolicies lists all retention policies
func (s *Service) ListPolicies(ctx context.Context) ([]RetentionPolicy, error) {
	return s.store.ListPolicies(ctx)
}

// UpdatePolicy updates a retention policy
func (s *Service) UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	policy.UpdatedAt = time.Now()
	return s.store.UpdatePolicy(ctx, policy)
}

// DeletePolicy deletes a retention policy
func (s *Service) DeletePolicy(ctx context.Context, id string) error {
	return s.store.DeletePolicy(ctx, id)
}

// RunPolicy executes a retention policy
func (s *Service) RunPolicy(ctx context.Context, policyID string) (*RetentionJob, error) {
	policy, err := s.store.GetPolicy(ctx, policyID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	if !policy.IsEnabled {
		return nil, fmt.Errorf("policy is not enabled")
	}

	job := &RetentionJob{
		ID:        generateJobID(),
		PolicyID:  policyID,
		JobType:   "retention",
		Status:    "running",
		StartedAt: time.Now(),
		CreatedAt: time.Now(),
	}

	if err := s.store.CreateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	// Execute retention logic
	recordsProcessed, recordsDeleted := s.executeRetention(ctx, policy)

	// Update job
	job.RecordsProcessed = recordsProcessed
	job.RecordsDeleted = recordsDeleted
	job.Status = "completed"
	now := time.Now()
	job.CompletedAt = &now
	job.Duration = time.Since(job.StartedAt).Milliseconds()

	// Update policy last run time
	policy.LastRunAt = &now
	s.store.UpdatePolicy(ctx, policy)

	if err := s.store.UpdateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("failed to update job: %w", err)
	}

	return job, nil
}

// executeRetention executes the retention logic for a policy
func (s *Service) executeRetention(ctx context.Context, policy *RetentionPolicy) (processed, deleted int) {
	cutoffDate := time.Now().AddDate(0, 0, -policy.RetentionDays)

	// Apply retention to different data types
	if policy.ApplyToInteractions {
		p, d := s.retainInteractions(ctx, cutoffDate, policy.MaxRecords)
		processed += p
		deleted += d
	}

	if policy.ApplyToCases {
		p, d := s.retainCases(ctx, cutoffDate, policy.MaxRecords)
		processed += p
		deleted += d
	}

	if policy.ApplyToPayloads {
		p, d := s.retainPayloads(ctx, cutoffDate, policy.MaxRecords)
		processed += p
		deleted += d
	}

	return processed, deleted
}

// retainInteractions retains interactions based on policy
func (s *Service) retainInteractions(ctx context.Context, cutoffDate time.Time, maxRecords int) (processed, deleted int) {
	xormStore, ok := s.store.(*XormStore)
	if !ok {
		return 0, 0
	}

	// Count interactions older than cutoff
	count, err := xormStore.engine.Where("created_at < ?", cutoffDate).Count(new(InteractionRecord))
	if err != nil {
		return 0, 0
	}
	processed = int(count)

	// Delete interactions older than cutoff
	affected, err := xormStore.engine.Where("created_at < ?", cutoffDate).Delete(new(InteractionRecord))
	if err != nil {
		return processed, 0
	}
	deleted = int(affected)

	// If maxRecords is set, delete oldest beyond the limit
	if maxRecords > 0 {
		totalCount, err := xormStore.engine.Count(new(InteractionRecord))
		if err == nil && int(totalCount) > maxRecords {
			excess := int(totalCount) - maxRecords
			_, _ = xormStore.engine.Where("created_at < ?", cutoffDate).Limit(excess, 0).Delete(new(InteractionRecord))
			deleted += excess
		}
	}

	return processed, deleted
}

// retainCases retains cases based on policy
func (s *Service) retainCases(ctx context.Context, cutoffDate time.Time, maxRecords int) (processed, deleted int) {
	xormStore, ok := s.store.(*XormStore)
	if !ok {
		return 0, 0
	}

	count, err := xormStore.engine.Where("created_at < ?", cutoffDate).Count(new(CaseRecord))
	if err != nil {
		return 0, 0
	}
	processed = int(count)

	affected, err := xormStore.engine.Where("created_at < ?", cutoffDate).Delete(new(CaseRecord))
	if err != nil {
		return processed, 0
	}
	deleted = int(affected)

	return processed, deleted
}

// retainPayloads retains payloads based on policy
func (s *Service) retainPayloads(ctx context.Context, cutoffDate time.Time, maxRecords int) (processed, deleted int) {
	xormStore, ok := s.store.(*XormStore)
	if !ok {
		return 0, 0
	}

	count, err := xormStore.engine.Where("created_at < ?", cutoffDate).Count(new(PayloadRecord))
	if err != nil {
		return 0, 0
	}
	processed = int(count)

	affected, err := xormStore.engine.Where("created_at < ?", cutoffDate).Delete(new(PayloadRecord))
	if err != nil {
		return processed, 0
	}
	deleted = int(affected)

	return processed, deleted
}

// CreateArchive creates an archive of data
func (s *Service) CreateArchive(ctx context.Context, policyID string, dataType string) (*Archive, error) {
	if _, err := s.store.GetPolicy(ctx, policyID); err != nil {
		return nil, fmt.Errorf("failed to get policy: %w", err)
	}

	archive := &Archive{
		ID:          generateArchiveID(),
		PolicyID:    policyID,
		DataType:    dataType,
		Status:      "pending",
		Compression: "gzip",
		CreatedAt:   time.Now(),
	}

	if err := s.store.CreateArchive(ctx, archive); err != nil {
		return nil, fmt.Errorf("failed to create archive: %w", err)
	}

	// Execute archival logic
	// This would:
	// 1. Query data based on policy
	// 2. Export to file
	// 3. Compress file
	// 4. Calculate checksum
	// 5. Store archive

	archive.Status = "completed"
	archive.RecordCount = 0 // Would be actual count
	now := time.Now()
	archive.CompletedAt = &now

	if err := s.store.UpdateArchive(ctx, archive); err != nil {
		return nil, fmt.Errorf("failed to update archive: %w", err)
	}

	return archive, nil
}

// ListArchives lists all archives
func (s *Service) ListArchives(ctx context.Context) ([]Archive, error) {
	return s.store.ListArchives(ctx)
}

// GetArchive retrieves an archive by ID
func (s *Service) GetArchive(ctx context.Context, id string) (*Archive, error) {
	return s.store.GetArchive(ctx, id)
}

// ListJobs lists all retention jobs
func (s *Service) ListJobs(ctx context.Context) ([]RetentionJob, error) {
	return s.store.ListJobs(ctx)
}

// GetJob retrieves a job by ID
func (s *Service) GetJob(ctx context.Context, id string) (*RetentionJob, error) {
	return s.store.GetJob(ctx, id)
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return fmt.Sprintf("job-%d", time.Now().UnixNano())
}

// generateArchiveID generates a unique archive ID
func generateArchiveID() string {
	return fmt.Sprintf("archive-%d", time.Now().UnixNano())
}

// Scheduler runs retention policies on a periodic basis.
// It iterates all enabled policies and executes those whose interval has elapsed
// since the last run.
type Scheduler struct {
	service  *Service
	interval time.Duration
	stopCh   chan struct{}
}

// NewScheduler creates a new retention scheduler that checks for due policies
// at the given interval.
func NewScheduler(service *Service, interval time.Duration) *Scheduler {
	return &Scheduler{
		service:  service,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start launches the scheduler goroutine.
func (sc *Scheduler) Start() {
	go sc.run()
}

// Stop signals the scheduler to stop.
func (sc *Scheduler) Stop() {
	close(sc.stopCh)
}

// run is the main scheduler loop.
func (sc *Scheduler) run() {
	ticker := time.NewTicker(sc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-sc.stopCh:
			return
		case <-ticker.C:
			sc.runDuePolicies()
		}
	}
}

// runDuePolicies executes all enabled policies whose run interval has elapsed.
func (sc *Scheduler) runDuePolicies() {
	ctx := context.Background()
	policies, err := sc.service.ListPolicies(ctx)
	if err != nil {
		return
	}

	for _, policy := range policies {
		if !policy.IsEnabled {
			continue
		}
		if !sc.isPolicyDue(&policy) {
			continue
		}
		// Run the policy, ignore errors to continue processing other policies
		_, _ = sc.service.RunPolicy(ctx, policy.ID)
	}
}

// isPolicyDue checks if a policy is due to run based on its schedule and last run time.
func (sc *Scheduler) isPolicyDue(policy *RetentionPolicy) bool {
	intervalHours := policy.RunIntervalHours
	if intervalHours <= 0 {
		// Derive from schedule flags
		switch {
		case policy.RunHourly:
			intervalHours = 1
		case policy.RunDaily:
			intervalHours = 24
		case policy.RunWeekly:
			intervalHours = 168
		case policy.RunMonthly:
			intervalHours = 720
		default:
			return false
		}
	}
	if policy.LastRunAt == nil {
		return true
	}
	nextRun := policy.LastRunAt.Add(time.Duration(intervalHours) * time.Hour)
	return time.Now().After(nextRun)
}

// InteractionRecord is a lightweight record for retention operations on the interactions table.
type InteractionRecord struct {
	ID        string    `xorm:"'id' pk varchar(36)"`
	CreatedAt time.Time `xorm:"'created_at' datetime"`
}

// TableName returns the table name for InteractionRecord
func (InteractionRecord) TableName() string {
	return "interactions"
}

// CaseRecord is a lightweight record for retention operations on the cases table.
type CaseRecord struct {
	ID        string    `xorm:"'id' pk varchar(36)"`
	CreatedAt time.Time `xorm:"'created_at' datetime"`
}

// TableName returns the table name for CaseRecord
func (CaseRecord) TableName() string {
	return "cases"
}

// PayloadRecord is a lightweight record for retention operations on the payloads table.
type PayloadRecord struct {
	ID        string    `xorm:"'id' pk varchar(36)"`
	CreatedAt time.Time `xorm:"'created_at' datetime"`
}

// TableName returns the table name for PayloadRecord
func (PayloadRecord) TableName() string {
	return "payloads"
}

// Store defines the storage interface for retention operations
type Store interface {
	// Policy operations
	CreatePolicy(ctx context.Context, policy *RetentionPolicy) error
	GetPolicy(ctx context.Context, id string) (*RetentionPolicy, error)
	ListPolicies(ctx context.Context) ([]RetentionPolicy, error)
	UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error
	DeletePolicy(ctx context.Context, id string) error

	// Job operations
	CreateJob(ctx context.Context, job *RetentionJob) error
	GetJob(ctx context.Context, id string) (*RetentionJob, error)
	ListJobs(ctx context.Context) ([]RetentionJob, error)
	UpdateJob(ctx context.Context, job *RetentionJob) error

	// Archive operations
	CreateArchive(ctx context.Context, archive *Archive) error
	GetArchive(ctx context.Context, id string) (*Archive, error)
	ListArchives(ctx context.Context) ([]Archive, error)
	UpdateArchive(ctx context.Context, archive *Archive) error
}
