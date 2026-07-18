package listener

import (
	"context"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestManager_StartListener(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	// Create a listener and start it
	l := &Listener{
		ID:        "test-start",
		Protocol:  ProtocolSMTP,
		Host:      "127.0.0.1",
		Port:      18902,
		IsEnabled: true,
	}

	err := mgr.StartListener(l)
	if err != nil {
		t.Fatalf("StartListener failed: %v", err)
	}

	if !mgr.IsActive("test-start") {
		t.Fatal("expected listener to be active after StartListener")
	}

	_ = mgr.StopListener("test-start")
	mgr.Stop()
}

func TestManager_StartListener_UnsupportedProtocol(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	l := &Listener{
		ID:        "test-unsupported",
		Protocol:  "unsupported",
		Host:      "127.0.0.1",
		Port:      18999,
		IsEnabled: true,
	}

	err := mgr.StartListener(l)
	if err == nil {
		t.Fatal("expected error for unsupported protocol")
	}

	mgr.Stop()
}

func TestManager_StopListener_NotRunning(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	err := mgr.StopListener("nonexistent")
	if err == nil {
		t.Fatal("expected error when stopping non-running listener")
	}

	mgr.Stop()
}

func TestManager_RestartListener(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	l := &Listener{
		ID:        "test-restart",
		Protocol:  ProtocolSMTP,
		Host:      "127.0.0.1",
		Port:      18903,
		IsEnabled: true,
	}

	// Start
	err := mgr.StartListener(l)
	if err != nil {
		t.Fatalf("StartListener failed: %v", err)
	}

	if !mgr.IsActive("test-restart") {
		t.Fatal("expected listener to be active")
	}

	// Restart
	err = mgr.RestartListener(l)
	if err != nil {
		t.Fatalf("RestartListener failed: %v", err)
	}

	if !mgr.IsActive("test-restart") {
		t.Fatal("expected listener to be active after restart")
	}

	_ = mgr.StopListener("test-restart")
	mgr.Stop()
}

func TestManager_RestartListener_NotRunning(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	// Restart a listener that was never started should work
	l := &Listener{
		ID:        "test-restart-new",
		Protocol:  ProtocolSMTP,
		Host:      "127.0.0.1",
		Port:      18904,
		IsEnabled: true,
	}

	err := mgr.RestartListener(l)
	if err != nil {
		t.Fatalf("RestartListener on new listener should work: %v", err)
	}

	if !mgr.IsActive("test-restart-new") {
		t.Fatal("expected listener to be active after restart")
	}

	_ = mgr.StopListener("test-restart-new")
	mgr.Stop()
}

func TestManager_GetConfigForProtocol(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	mgr := NewManager(store, logger)

	t.Run("SMTP", func(t *testing.T) {
		cfg := mgr.getConfigForProtocol(ProtocolSMTP)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.MaxConnections != 1000 {
			t.Errorf("expected 1000, got %d", cfg.MaxConnections)
		}
	})

	t.Run("LDAP", func(t *testing.T) {
		cfg := mgr.getConfigForProtocol(ProtocolLDAP)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
	})

	t.Run("SMB", func(t *testing.T) {
		cfg := mgr.getConfigForProtocol(ProtocolSMB)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
	})

	t.Run("FTP", func(t *testing.T) {
		cfg := mgr.getConfigForProtocol(ProtocolFTP)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
	})

	t.Run("RMI", func(t *testing.T) {
		cfg := mgr.getConfigForProtocol(ProtocolRMI)
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
	})

	t.Run("unknown protocol", func(t *testing.T) {
		cfg := mgr.getConfigForProtocol("unknown")
		if cfg == nil {
			t.Fatal("expected non-nil config")
		}
		if cfg.MaxConnections != 100 {
			t.Errorf("expected default MaxConnections 100, got %d", cfg.MaxConnections)
		}
	})
}

func TestManager_MultipleListeners(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	mgr := NewManager(store, logger)
	ctx := context.Background()
	_ = mgr.Start(ctx)

	listeners := []*Listener{
		{ID: "multi-1", Protocol: ProtocolSMTP, Host: "127.0.0.1", Port: 18951, IsEnabled: true},
		{ID: "multi-2", Protocol: ProtocolLDAP, Host: "127.0.0.1", Port: 18952, IsEnabled: true},
		{ID: "multi-3", Protocol: ProtocolSMB, Host: "127.0.0.1", Port: 18953, IsEnabled: true},
	}

	for _, l := range listeners {
		err := mgr.StartListener(l)
		if err != nil {
			t.Fatalf("StartListener(%s) failed: %v", l.ID, err)
		}
	}

	if mgr.ActiveCount() != 3 {
		t.Fatalf("expected 3 active listeners, got %d", mgr.ActiveCount())
	}

	// Stop one
	_ = mgr.StopListener("multi-2")
	if mgr.ActiveCount() != 2 {
		t.Fatalf("expected 2 active listeners after stop, got %d", mgr.ActiveCount())
	}

	// Stop remaining
	_ = mgr.StopListener("multi-1")
	_ = mgr.StopListener("multi-3")

	mgr.Stop()
}

func TestManager_StartWithEnabled(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	// Pre-add an enabled listener to the store
	store.listeners = append(store.listeners, Listener{
		ID:        "pre-enabled",
		Protocol:  ProtocolSMTP,
		Host:      "127.0.0.1",
		Port:      18905,
		IsEnabled: true,
	})

	mgr := NewManager(store, logger)
	ctx := context.Background()

	err := mgr.Start(ctx)
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Give it time to start
	time.Sleep(50 * time.Millisecond)

	if mgr.ActiveCount() != 1 {
		t.Fatalf("expected 1 active listener, got %d", mgr.ActiveCount())
	}

	mgr.Stop()
}

func TestManager_ActiveCount(t *testing.T) {
	store := newMockStore()
	logger := logrus.New()
	mgr := NewManager(store, logger)

	c := mgr.ActiveCount()
	if c != 0 {
		t.Fatalf("expected 0, got %d", c)
	}
}
