package canary

import (
	"context"
	"time"

	"xorm.io/xorm"
)

// Store defines the interface for canary storage
type Store interface {
	// Canary operations
	CreateCanary(ctx context.Context, canary *Canary) error
	GetCanary(ctx context.Context, id string) (*Canary, error)
	GetCanaryByToken(ctx context.Context, token string) (*Canary, error)
	GetActiveCanaries(ctx context.Context) ([]Canary, error)
	GetAllCanaries(ctx context.Context) ([]Canary, error)
	UpdateCanary(ctx context.Context, canary *Canary) error
	DeleteCanary(ctx context.Context, id string) error
	DeleteExpiredCanaries(ctx context.Context, maxDays int) error

	// Canary hit operations
	SaveCanaryHit(ctx context.Context, hit *CanaryHit) error
	GetCanaryHit(ctx context.Context, id string) (*CanaryHit, error)
	GetCanaryHits(ctx context.Context, canaryID string) ([]CanaryHit, error)
	GetRecentCanaryHits(ctx context.Context, canaryID string, windowSeconds int) ([]CanaryHit, error)
	UpdateCanaryHit(ctx context.Context, hit *CanaryHit) error
	DeleteCanaryHit(ctx context.Context, id string) error
}

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreateCanary creates a new canary
func (s *XormStore) CreateCanary(ctx context.Context, canary *Canary) error {
	canary.CreatedAt = time.Now()
	canary.UpdatedAt = time.Now()
	_, err := s.engine.Insert(canary)
	return err
}

// GetCanary retrieves a canary by ID
func (s *XormStore) GetCanary(ctx context.Context, id string) (*Canary, error) {
	var canary Canary
	_, err := s.engine.ID(id).Get(&canary)
	if err != nil {
		return nil, err
	}
	return &canary, nil
}

// GetCanaryByToken retrieves a canary by token
func (s *XormStore) GetCanaryByToken(ctx context.Context, token string) (*Canary, error) {
	var canary Canary
	_, err := s.engine.Where("token = ?", token).Get(&canary)
	if err != nil {
		return nil, err
	}
	return &canary, nil
}

// GetActiveCanaries retrieves all active canaries
func (s *XormStore) GetActiveCanaries(ctx context.Context) ([]Canary, error) {
	var canaries []Canary
	err := s.engine.Where("is_enabled = ? AND expires_at > ?", true, time.Now()).Find(&canaries)
	return canaries, err
}

// GetAllCanaries retrieves all canaries
func (s *XormStore) GetAllCanaries(ctx context.Context) ([]Canary, error) {
	var canaries []Canary
	err := s.engine.Find(&canaries)
	return canaries, err
}

// UpdateCanary updates a canary
func (s *XormStore) UpdateCanary(ctx context.Context, canary *Canary) error {
	canary.UpdatedAt = time.Now()
	_, err := s.engine.ID(canary.ID).Update(canary)
	return err
}

// DeleteCanary deletes a canary
func (s *XormStore) DeleteCanary(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&Canary{})
	return err
}

// DeleteExpiredCanaries deletes expired canaries
func (s *XormStore) DeleteExpiredCanaries(ctx context.Context, maxDays int) error {
	cutoff := time.Now().AddDate(0, 0, -maxDays)
	_, err := s.engine.Where("expires_at < ?", cutoff).Delete(&Canary{})
	return err
}

// SaveCanaryHit saves a canary hit
func (s *XormStore) SaveCanaryHit(ctx context.Context, hit *CanaryHit) error {
	_, err := s.engine.Insert(hit)
	return err
}

// GetCanaryHit retrieves a canary hit by ID
func (s *XormStore) GetCanaryHit(ctx context.Context, id string) (*CanaryHit, error) {
	var hit CanaryHit
	_, err := s.engine.ID(id).Get(&hit)
	if err != nil {
		return nil, err
	}
	return &hit, nil
}

// GetCanaryHits retrieves all hits for a canary
func (s *XormStore) GetCanaryHits(ctx context.Context, canaryID string) ([]CanaryHit, error) {
	var hits []CanaryHit
	err := s.engine.Where("canary_id = ?", canaryID).Desc("timestamp").Find(&hits)
	return hits, err
}

// GetRecentCanaryHits retrieves recent hits within time window
func (s *XormStore) GetRecentCanaryHits(ctx context.Context, canaryID string, windowSeconds int) ([]CanaryHit, error) {
	cutoff := time.Now().Add(-time.Duration(windowSeconds) * time.Second)
	var hits []CanaryHit
	err := s.engine.Where("canary_id = ? AND timestamp > ?", canaryID, cutoff).Find(&hits)
	return hits, err
}

// UpdateCanaryHit updates a canary hit
func (s *XormStore) UpdateCanaryHit(ctx context.Context, hit *CanaryHit) error {
	_, err := s.engine.ID(hit.ID).Update(hit)
	return err
}

// DeleteCanaryHit deletes a canary hit
func (s *XormStore) DeleteCanaryHit(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&CanaryHit{})
	return err
}
