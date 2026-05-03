package listener

import (
	"context"
	"time"

	"xorm.io/xorm"
)

// Store defines the interface for listener storage
type Store interface {
	// Listener operations
	CreateListener(ctx context.Context, listener *Listener) error
	GetListener(ctx context.Context, id string) (*Listener, error)
	GetListenerByToken(ctx context.Context, token string) (*Listener, error)
	GetAllListeners(ctx context.Context) ([]Listener, error)
	UpdateListener(ctx context.Context, listener *Listener) error
	DeleteListener(ctx context.Context, id string) error

	// Listener interaction operations
	SaveListenerInteraction(ctx context.Context, interaction *ListenerInteraction) error
	GetListenerInteractions(ctx context.Context, listenerID string) ([]ListenerInteraction, error)
	DeleteListenerInteraction(ctx context.Context, id string) error

	// SMTP message operations
	SaveSMTPMessage(ctx context.Context, message *SMTPMessage) error
	GetSMTPMessages(ctx context.Context, listenerID string) ([]SMTPMessage, error)
	GetSMTPMessage(ctx context.Context, id string) (*SMTPMessage, error)
	DeleteSMTPMessage(ctx context.Context, id string) error

	// LDAP query operations
	SaveLDAPQuery(ctx context.Context, query *LDAPQuery) error
	GetLDAPQueries(ctx context.Context, listenerID string) ([]LDAPQuery, error)
	GetLDAPQuery(ctx context.Context, id string) (*LDAPQuery, error)
	DeleteLDAPQuery(ctx context.Context, id string) error
}

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreateListener creates a new listener
func (s *XormStore) CreateListener(ctx context.Context, listener *Listener) error {
	listener.CreatedAt = time.Now()
	listener.UpdatedAt = time.Now()
	_, err := s.engine.Insert(listener)
	return err
}

// GetListener retrieves a listener by ID
func (s *XormStore) GetListener(ctx context.Context, id string) (*Listener, error) {
	var listener Listener
	_, err := s.engine.ID(id).Get(&listener)
	if err != nil {
		return nil, err
	}
	return &listener, nil
}

// GetListenerByToken retrieves a listener by token
func (s *XormStore) GetListenerByToken(ctx context.Context, token string) (*Listener, error) {
	var listener Listener
	_, err := s.engine.Where("token = ?", token).Get(&listener)
	if err != nil {
		return nil, err
	}
	return &listener, nil
}

// GetAllListeners retrieves all listeners
func (s *XormStore) GetAllListeners(ctx context.Context) ([]Listener, error) {
	var listeners []Listener
	err := s.engine.Find(&listeners)
	return listeners, err
}

// UpdateListener updates a listener
func (s *XormStore) UpdateListener(ctx context.Context, listener *Listener) error {
	listener.UpdatedAt = time.Now()
	_, err := s.engine.ID(listener.ID).Update(listener)
	return err
}

// DeleteListener deletes a listener
func (s *XormStore) DeleteListener(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&Listener{})
	return err
}

// SaveListenerInteraction saves a listener interaction
func (s *XormStore) SaveListenerInteraction(ctx context.Context, interaction *ListenerInteraction) error {
	_, err := s.engine.Insert(interaction)
	return err
}

// GetListenerInteractions retrieves interactions for a listener
func (s *XormStore) GetListenerInteractions(ctx context.Context, listenerID string) ([]ListenerInteraction, error) {
	var interactions []ListenerInteraction
	err := s.engine.Where("listener_id = ?", listenerID).Desc("timestamp").Find(&interactions)
	return interactions, err
}

// DeleteListenerInteraction deletes a listener interaction
func (s *XormStore) DeleteListenerInteraction(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&ListenerInteraction{})
	return err
}

// SaveSMTPMessage saves an SMTP message
func (s *XormStore) SaveSMTPMessage(ctx context.Context, message *SMTPMessage) error {
	_, err := s.engine.Insert(message)
	return err
}

// GetSMTPMessages retrieves SMTP messages for a listener
func (s *XormStore) GetSMTPMessages(ctx context.Context, listenerID string) ([]SMTPMessage, error) {
	var messages []SMTPMessage
	err := s.engine.Where("listener_id = ?", listenerID).Desc("timestamp").Find(&messages)
	return messages, err
}

// GetSMTPMessage retrieves an SMTP message by ID
func (s *XormStore) GetSMTPMessage(ctx context.Context, id string) (*SMTPMessage, error) {
	var message SMTPMessage
	_, err := s.engine.ID(id).Get(&message)
	if err != nil {
		return nil, err
	}
	return &message, nil
}

// DeleteSMTPMessage deletes an SMTP message
func (s *XormStore) DeleteSMTPMessage(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&SMTPMessage{})
	return err
}

// SaveLDAPQuery saves an LDAP query
func (s *XormStore) SaveLDAPQuery(ctx context.Context, query *LDAPQuery) error {
	_, err := s.engine.Insert(query)
	return err
}

// GetLDAPQueries retrieves LDAP queries for a listener
func (s *XormStore) GetLDAPQueries(ctx context.Context, listenerID string) ([]LDAPQuery, error) {
	var queries []LDAPQuery
	err := s.engine.Where("listener_id = ?", listenerID).Desc("timestamp").Find(&queries)
	return queries, err
}

// GetLDAPQuery retrieves an LDAP query by ID
func (s *XormStore) GetLDAPQuery(ctx context.Context, id string) (*LDAPQuery, error) {
	var query LDAPQuery
	_, err := s.engine.ID(id).Get(&query)
	if err != nil {
		return nil, err
	}
	return &query, nil
}

// DeleteLDAPQuery deletes an LDAP query
func (s *XormStore) DeleteLDAPQuery(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&LDAPQuery{})
	return err
}
