package listener

import (
	"errors"
	"time"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrListenerNotFound = errors.New("listener not found")
)

// Service provides listener management services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new listener service
func NewService(engine *xorm.Engine) *Service {
	return &Service{
		engine: engine,
	}
}

// CreateListener creates a new listener
func (s *Service) CreateListener(listener *models.Listener) error {
	if listener.ID == "" {
		listener.ID = models.GenerateID()
	}
	if listener.CreatedAt.IsZero() {
		listener.CreatedAt = time.Now()
	}
	if listener.UpdatedAt.IsZero() {
		listener.UpdatedAt = time.Now()
	}

	_, err := s.engine.Insert(listener)
	return err
}

// GetListener retrieves a listener by its ID
func (s *Service) GetListener(id string) (*models.Listener, error) {
	var listener models.Listener
	has, err := s.engine.ID(id).Get(&listener)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrListenerNotFound
	}
	return &listener, nil
}

// ListListeners retrieves listeners with pagination
func (s *Service) ListListeners(page, pageSize int) ([]models.Listener, int64, error) {
	var listeners []models.Listener
	total, err := s.engine.Count(&models.Listener{})
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := s.engine.Limit(pageSize, offset).Find(&listeners); err != nil {
		return nil, 0, err
	}

	return listeners, total, nil
}

// UpdateListener updates an existing listener
func (s *Service) UpdateListener(listener *models.Listener) error {
	listener.UpdatedAt = time.Now()
	_, err := s.engine.ID(listener.ID).Update(listener)
	return err
}

// DeleteListener deletes a listener by its ID
func (s *Service) DeleteListener(id string) error {
	_, err := s.engine.ID(id).Delete(&models.Listener{})
	return err
}

// ListListenerInteractions retrieves interactions for a specific listener
func (s *Service) ListListenerInteractions(listenerID string) ([]models.Interaction, error) {
	var interactions []models.Interaction
	err := s.engine.Where("listener_id = ?", listenerID).Find(&interactions)
	return interactions, err
}
