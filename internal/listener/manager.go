package listener

import (
	"context"
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// ProtocolListener is the common interface for all protocol listeners.
type ProtocolListener interface {
	Start(ctx context.Context) error
	Stop() error
}

// Manager manages the lifecycle of all protocol listeners at runtime.
// It starts, stops, and monitors listeners based on their DB state.
type Manager struct {
	store  Store
	logger *logrus.Logger
	mu     sync.Mutex
	active map[string]*managedListener
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// managedListener wraps a running listener with its cancel context and metadata.
type managedListener struct {
	listener ProtocolListener
	cancel   context.CancelFunc
	config   *ListenerConfig
	rateLim  *RateLimiter
	connLim  *ConnLimiter
}

// NewManager creates a new listener manager.
func NewManager(store Store, logger *logrus.Logger) *Manager {
	if logger == nil {
		logger = logrus.StandardLogger()
	}
	return &Manager{
		store:  store,
		logger: logger,
		active: make(map[string]*managedListener),
	}
}

// Start initializes the manager and starts all enabled listeners from the store.
func (m *Manager) Start(ctx context.Context) error {
	m.ctx, m.cancel = context.WithCancel(ctx)

	listeners, err := m.store.GetAllListeners(ctx)
	if err != nil {
		return fmt.Errorf("failed to load listeners: %w", err)
	}

	for _, l := range listeners {
		if !l.IsEnabled {
			continue
		}
		if err := m.startListener(m.ctx, &l); err != nil {
			m.logger.Errorf("[listener/manager] failed to start listener %s (%s): %v", l.ID, l.Protocol, err)
		}
	}

	// Start cleanup goroutine for rate limiters
	m.wg.Add(1)
	go m.cleanupLoop()

	m.logger.Infof("[listener/manager] started %d listeners", len(m.active))
	return nil
}

// Stop gracefully stops all active listeners.
func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}

	m.mu.Lock()
	for id, ml := range m.active {
		if err := ml.listener.Stop(); err != nil {
			m.logger.Errorf("[listener/manager] error stopping listener %s: %v", id, err)
		}
	}
	m.active = make(map[string]*managedListener)
	m.mu.Unlock()

	m.wg.Wait()
	m.logger.Info("[listener/manager] all listeners stopped")
}

// StartListener starts a specific listener by its DB record.
func (m *Manager) StartListener(l *Listener) error {
	return m.startListener(m.ctx, l)
}

// StopListener stops a specific listener by ID.
func (m *Manager) StopListener(id string) error {
	m.mu.Lock()
	ml, ok := m.active[id]
	m.mu.Unlock()
	if !ok {
		return fmt.Errorf("listener %s is not running", id)
	}

	if err := ml.listener.Stop(); err != nil {
		return fmt.Errorf("failed to stop listener %s: %w", id, err)
	}

	m.mu.Lock()
	delete(m.active, id)
	m.mu.Unlock()

	m.logger.Infof("[listener/manager] stopped listener %s", id)
	return nil
}

// RestartListener restarts a specific listener.
func (m *Manager) RestartListener(l *Listener) error {
	if err := m.StopListener(l.ID); err != nil {
		// Not running is OK, continue to start
	}
	return m.startListener(m.ctx, l)
}

// IsActive returns true if the listener is currently running.
func (m *Manager) IsActive(id string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	_, ok := m.active[id]
	return ok
}

// ActiveCount returns the number of currently running listeners.
func (m *Manager) ActiveCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.active)
}

// startListener creates and starts a protocol listener from a DB record.
func (m *Manager) startListener(ctx context.Context, l *Listener) error {
	cfg := m.getConfigForProtocol(l.Protocol)

	// Create rate and connection limiters
	rateLim := NewRateLimiter(1*time.Minute, cfg.MaxConnections)
	connLim := NewConnLimiter(cfg.MaxConnections)

	// Create the store wrapper with rate limiting
	wrappedStore := &rateLimitedStore{
		Store:   m.store,
		rateLim: rateLim,
		connLim: connLim,
	}

	var protoListener ProtocolListener
	switch l.Protocol {
	case ProtocolSMTP:
		protoListener = NewSMTPListener(l, cfg, wrappedStore, m.logger)
	case ProtocolLDAP:
		protoListener = NewLDAPListener(l, cfg, wrappedStore, m.logger)
	case ProtocolSMB:
		protoListener = NewSMBListener(cfg, wrappedStore, l)
	case ProtocolFTP:
		protoListener = NewFTPListener(cfg, wrappedStore, l)
	default:
		return fmt.Errorf("unsupported protocol: %s", l.Protocol)
	}

	// Apply TLS if configured
	if cfg.EnableTLS && cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
		if tl, ok := protoListener.(TLSCapable); ok {
			tlsConfig, err := loadTLSConfig(cfg.TLSCertFile, cfg.TLSKeyFile)
			if err != nil {
				return fmt.Errorf("failed to load TLS config: %w", err)
			}
			tl.SetTLSConfig(tlsConfig)
		}
	}

	childCtx, cancel := context.WithCancel(ctx)
	if err := protoListener.Start(childCtx); err != nil {
		cancel()
		return err
	}

	m.mu.Lock()
	m.active[l.ID] = &managedListener{
		listener: protoListener,
		cancel:   cancel,
		config:   cfg,
		rateLim:  rateLim,
		connLim:  connLim,
	}
	m.mu.Unlock()

	m.logger.Infof("[listener/manager] started listener %s (%s) on %s:%d", l.ID, l.Protocol, l.Host, l.Port)
	return nil
}

// getConfigForProtocol returns sensible defaults for each protocol.
func (m *Manager) getConfigForProtocol(protocol Protocol) *ListenerConfig {
	switch protocol {
	case ProtocolSMTP:
		return DefaultSMTPConfig()
	case ProtocolLDAP:
		return DefaultLDAPConfig()
	case ProtocolSMB:
		return DefaultSMBConfig()
	case ProtocolFTP:
		return DefaultFTPConfig()
	default:
		return &ListenerConfig{
			MaxConnections: 100,
			Timeout:        30 * time.Second,
			BufferSize:     4096,
		}
	}
}

// cleanupLoop periodically cleans up expired rate limiter entries.
func (m *Manager) cleanupLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.mu.Lock()
			for _, ml := range m.active {
				ml.rateLim.Cleanup()
			}
			m.mu.Unlock()
		}
	}
}

// TLSCapable is implemented by listeners that support TLS.
type TLSCapable interface {
	SetTLSConfig(config *tls.Config)
}

// loadTLSConfig loads a TLS certificate pair.
func loadTLSConfig(certFile, keyFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// rateLimitedStore wraps a Store with rate limiting metadata.
// The actual rate limiting happens in the accept loop of each listener,
// but this wrapper makes the store aware of the limiters for future use.
type rateLimitedStore struct {
	Store
	rateLim *RateLimiter
	connLim *ConnLimiter
}
