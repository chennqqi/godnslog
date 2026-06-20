package listener

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

// mockStore implements Store for testing the manager without a real DB.
type mockStore struct {
	Store
	listeners []Listener
}

func newMockStore() *mockStore {
	return &mockStore{
		listeners: []Listener{},
	}
}

func (m *mockStore) GetAllListeners(ctx context.Context) ([]Listener, error) {
	result := make([]Listener, len(m.listeners))
	copy(result, m.listeners)
	return result, nil
}

func (m *mockStore) CreateListener(ctx context.Context, l *Listener) error {
	m.listeners = append(m.listeners, *l)
	return nil
}

func (m *mockStore) GetListener(ctx context.Context, id string) (*Listener, error) {
	for _, l := range m.listeners {
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, ErrListenerNotFound
}

func (m *mockStore) UpdateListener(ctx context.Context, l *Listener) error {
	for i := range m.listeners {
		if m.listeners[i].ID == l.ID {
			m.listeners[i] = *l
			return nil
		}
	}
	return ErrListenerNotFound
}

func (m *mockStore) DeleteListener(ctx context.Context, id string) error {
	for i, l := range m.listeners {
		if l.ID == id {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			return nil
		}
	}
	return ErrListenerNotFound
}

func TestManager_StartStop(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)

	// Add a disabled listener — should not start
	store.listeners = append(store.listeners, Listener{
		ID:        "test-1",
		Protocol:  ProtocolSMTP,
		Host:      "127.0.0.1",
		Port:      18999,
		IsEnabled: false,
	})

	ctx := context.Background()
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if mgr.ActiveCount() != 0 {
		t.Fatalf("expected 0 active listeners, got %d", mgr.ActiveCount())
	}

	mgr.Stop()
}

func TestManager_StartEnabledListener(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)

	// Add an enabled SMTP listener on a random port
	store.listeners = append(store.listeners, Listener{
		ID:        "test-smtp-1",
		Protocol:  ProtocolSMTP,
		Host:      "127.0.0.1",
		Port:      18901,
		IsEnabled: true,
	})

	ctx := context.Background()
	if err := mgr.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give listener a moment to bind
	time.Sleep(100 * time.Millisecond)

	if !mgr.IsActive("test-smtp-1") {
		t.Fatal("expected listener test-smtp-1 to be active")
	}

	if mgr.ActiveCount() != 1 {
		t.Fatalf("expected 1 active listener, got %d", mgr.ActiveCount())
	}

	mgr.Stop()

	if mgr.ActiveCount() != 0 {
		t.Fatalf("expected 0 active after stop, got %d", mgr.ActiveCount())
	}
}

func TestManager_StopNonexistent(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	err := mgr.StopListener("nonexistent")
	if err == nil {
		t.Fatal("expected error stopping nonexistent listener")
	}

	mgr.Stop()
}
