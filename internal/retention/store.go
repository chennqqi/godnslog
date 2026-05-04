package retention

import (
	"context"

	"xorm.io/xorm"
)

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreatePolicy creates a new retention policy
func (s *XormStore) CreatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	_, err := s.engine.Insert(policy)
	return err
}

// GetPolicy retrieves a policy by ID
func (s *XormStore) GetPolicy(ctx context.Context, id string) (*RetentionPolicy, error) {
	var policy RetentionPolicy
	_, err := s.engine.ID(id).Get(&policy)
	if err != nil {
		return nil, err
	}
	return &policy, nil
}

// ListPolicies lists all retention policies
func (s *XormStore) ListPolicies(ctx context.Context) ([]RetentionPolicy, error) {
	var policies []RetentionPolicy
	err := s.engine.Find(&policies)
	return policies, err
}

// UpdatePolicy updates a retention policy
func (s *XormStore) UpdatePolicy(ctx context.Context, policy *RetentionPolicy) error {
	_, err := s.engine.ID(policy.ID).Update(policy)
	return err
}

// DeletePolicy deletes a retention policy
func (s *XormStore) DeletePolicy(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&RetentionPolicy{})
	return err
}

// CreateJob creates a new retention job
func (s *XormStore) CreateJob(ctx context.Context, job *RetentionJob) error {
	_, err := s.engine.Insert(job)
	return err
}

// GetJob retrieves a job by ID
func (s *XormStore) GetJob(ctx context.Context, id string) (*RetentionJob, error) {
	var job RetentionJob
	_, err := s.engine.ID(id).Get(&job)
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// ListJobs lists all retention jobs
func (s *XormStore) ListJobs(ctx context.Context) ([]RetentionJob, error) {
	var jobs []RetentionJob
	err := s.engine.Desc("created_at").Find(&jobs)
	return jobs, err
}

// UpdateJob updates a retention job
func (s *XormStore) UpdateJob(ctx context.Context, job *RetentionJob) error {
	_, err := s.engine.ID(job.ID).Update(job)
	return err
}

// CreateArchive creates a new archive
func (s *XormStore) CreateArchive(ctx context.Context, archive *Archive) error {
	_, err := s.engine.Insert(archive)
	return err
}

// GetArchive retrieves an archive by ID
func (s *XormStore) GetArchive(ctx context.Context, id string) (*Archive, error) {
	var archive Archive
	_, err := s.engine.ID(id).Get(&archive)
	if err != nil {
		return nil, err
	}
	return &archive, nil
}

// ListArchives lists all archives
func (s *XormStore) ListArchives(ctx context.Context) ([]Archive, error) {
	var archives []Archive
	err := s.engine.Desc("created_at").Find(&archives)
	return archives, err
}

// UpdateArchive updates an archive
func (s *XormStore) UpdateArchive(ctx context.Context, archive *Archive) error {
	_, err := s.engine.ID(archive.ID).Update(archive)
	return err
}
