package listener

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// RMIListener implements a minimal RMI server for JNDI injection detection.
type RMIListener struct {
	listener *Listener
	config   *ListenerConfig
	server   net.Listener
	store    Store
	logger   *logrus.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewRMIListener creates a new RMI listener.
func NewRMIListener(listener *Listener, config *ListenerConfig, store Store, logger *logrus.Logger) *RMIListener {
	if config == nil {
		config = DefaultRMIConfig()
	}
	return &RMIListener{
		listener: listener,
		config:   config,
		store:    store,
		logger:   logger,
	}
}

// DefaultRMIConfig returns default RMI configuration.
func DefaultRMIConfig() *ListenerConfig {
	return &ListenerConfig{
		MaxConnections:           500,
		Timeout:                  30 * time.Second,
		BufferSize:               4096,
		RateLimitMax:             500,
		MaxConcurrentConnections: 200,
	}
}

// Start starts the RMI listener.
func (r *RMIListener) Start(ctx context.Context) error {
	r.ctx, r.cancel = context.WithCancel(ctx)
	addr := fmt.Sprintf("%s:%d", r.listener.Host, r.listener.Port)
	server, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("rmi listen on %s: %w", addr, err)
	}
	r.server = server
	r.logger.Infof("[rmi] listening on %s for token %s", addr, r.listener.Token)
	r.wg.Add(1)
	go r.acceptLoop()
	return nil
}

// Stop stops the RMI listener.
func (r *RMIListener) Stop() error {
	r.cancel()
	if r.server != nil {
		r.server.Close()
	}
	r.wg.Wait()
	return nil
}

// acceptLoop accepts incoming RMI connections.
func (r *RMIListener) acceptLoop() {
	defer r.wg.Done()
	sc := GetSecurityContext(r.listener.ID)
	for {
		conn, err := r.server.Accept()
		if err != nil {
			select {
			case <-r.ctx.Done():
				return
			default:
				r.logger.Warnf("[rmi] accept error: %v", err)
				continue
			}
		}

		if !sc.CheckConnection(conn.RemoteAddr()) {
			r.logger.Warnf("[rmi] connection rejected by rate/connection limit from %s", conn.RemoteAddr())
			conn.Close()
			continue
		}

		r.wg.Add(1)
		go func() {
			defer sc.ReleaseConnection()
			r.handleConnection(conn)
		}()
	}
}

const (
	rmiMagicJRMI        = 0x4a524d49 // "JRMI"
	rmiStreamProtocol   = 0x4b
	rmiSingleOpProtocol = 0x4c
)

// parseRMIHandshake extracts protocol info from JRMP handshake data.
// JRMP: 4 bytes magic + 2 bytes version + 1 byte protocol type
// StreamProtocol: followed by 2 bytes callback port.
func parseRMIHandshake(data []byte) (string, bool) {
	if len(data) < 7 {
		return "", false
	}
	magic := binary.BigEndian.Uint32(data[0:4])
	if magic != rmiMagicJRMI {
		return "", false
	}
	protoType := data[6]
	switch protoType {
	case rmiStreamProtocol:
		if len(data) >= 9 {
			port := binary.BigEndian.Uint16(data[7:9])
			return fmt.Sprintf("JRMP StreamProtocol (callback port: %d)", port), true
		}
		return "JRMP StreamProtocol", true
	case rmiSingleOpProtocol:
		return "JRMP SingleOpProtocol", true
	default:
		return fmt.Sprintf("JRMP type: 0x%02x", protoType), true
	}
}

// handleConnection handles a single RMI connection.
func (r *RMIListener) handleConnection(conn net.Conn) {
	defer conn.Close()
	r.wg.Done()

	conn.SetDeadline(time.Now().Add(r.config.Timeout))

	remoteAddr := conn.RemoteAddr().(*net.TCPAddr)
	buf := make([]byte, r.config.BufferSize)
	n, err := conn.Read(buf)
	if err != nil && err != io.EOF {
		r.logger.Debugf("[rmi] read error from %s: %v", remoteAddr.IP, err)
		return
	}
	data := buf[:n]
	urn, recognized := parseRMIHandshake(data)

	interaction := &RMIInteraction{
		ID:         fmt.Sprintf("rmi-%d-%d", time.Now().UnixNano(), remoteAddr.Port),
		ListenerID: r.listener.ID,
		SourceIP:   remoteAddr.IP.String(),
		SourcePort: remoteAddr.Port,
		URN:        urn,
		RawData:    fmt.Sprintf("%x", data),
		Timestamp:  time.Now(),
	}
	if err := r.store.SaveRMIInteraction(interaction); err != nil {
		r.logger.Warnf("[rmi] failed to save interaction: %v", err)
	}
	r.logger.Infof("[rmi] captured connection from %s:%d - %s (recognized: %v)",
		remoteAddr.IP, remoteAddr.Port, urn, recognized)
}
