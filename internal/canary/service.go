package canary

import (
	"context"
	"errors"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrCanaryNotFound = errors.New("canary not found")
)

// Service provides canary management services
type Service struct {
	engine *xorm.Engine
	store  Store
}

// NewService creates a new canary service
func NewService(engine *xorm.Engine) *Service {
	return &Service{
		engine: engine,
		store:  NewXormStore(engine),
	}
}

// CreateCanary creates a new canary token
func (s *Service) CreateCanary(canary *models.Canary) error {
	ctx := context.Background()
	return s.store.CreateCanary(ctx, canary)
}

// GetCanary retrieves a canary by its ID
func (s *Service) GetCanary(id string) (*models.Canary, error) {
	ctx := context.Background()
	return s.store.GetCanary(ctx, id)
}

// ListCanaries retrieves all canaries with pagination
func (s *Service) ListCanaries(page, pageSize int) ([]models.Canary, int64, error) {
	ctx := context.Background()
	canaries, err := s.store.GetAllCanaries(ctx)
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(canaries))
	start := (page - 1) * pageSize
	end := start + pageSize

	if start >= len(canaries) {
		return []models.Canary{}, total, nil
	}
	if end > len(canaries) {
		end = len(canaries)
	}

	return canaries[start:end], total, nil
}

// UpdateCanary updates an existing canary
func (s *Service) UpdateCanary(canary *models.Canary) error {
	ctx := context.Background()
	return s.store.UpdateCanary(ctx, canary)
}

// DeleteCanary deletes a canary by its ID
func (s *Service) DeleteCanary(id string) error {
	ctx := context.Background()
	return s.store.DeleteCanary(ctx, id)
}

// ListCanaryHits retrieves hits for a specific canary
func (s *Service) ListCanaryHits(canaryID string) ([]models.CanaryHit, error) {
	ctx := context.Background()
	return s.store.GetCanaryHits(ctx, canaryID)
}
