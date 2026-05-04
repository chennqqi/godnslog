package listener

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// SMBListener handles SMB protocol interactions
type SMBListener struct {
	config   *ListenerConfig
	store    Store
	listener *Listener
	server   net.Listener
	stopChan chan struct{}
}

// NewSMBListener creates a new SMB listener
func NewSMBListener(config *ListenerConfig, store Store, listener *Listener) *SMBListener {
	if config == nil {
		config = &ListenerConfig{
			MaxConnections: 10,
			Timeout:        30 * time.Second,
			BufferSize:     4096,
		}
	}
	return &SMBListener{
		config:   config,
		store:    store,
		listener: listener,
		stopChan: make(chan struct{}),
	}
}

// Start starts the SMB listener
func (s *SMBListener) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", s.listener.Host, s.listener.Port)

	var err error
	s.server, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start SMB listener: %w", err)
	}

	log.Printf("SMB listener started on %s", addr)

	go s.acceptConnections(ctx)

	return nil
}

// Stop stops the SMB listener
func (s *SMBListener) Stop() error {
	close(s.stopChan)
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// acceptConnections accepts incoming SMB connections
func (s *SMBListener) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-s.stopChan:
			return
		default:
			conn, err := s.server.Accept()
			if err != nil {
				select {
				case <-s.stopChan:
					return
				default:
					log.Printf("SMB accept error: %v", err)
					continue
				}
			}

			go s.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles an SMB connection
func (s *SMBListener) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	sourceIP, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		sourceIP = remoteAddr
	}

	log.Printf("SMB connection from %s", sourceIP)

	// Set timeout
	if s.config.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(s.config.Timeout))
	}

	// SMB protocol is complex, this is a simplified implementation
	// In production, use a proper SMB library like github.com/hirochachacha/go-smb2
	// For MVP, we capture basic connection attempts

	smbReq := &SMBRequest{
		ID:         generateID(),
		ListenerID: s.listener.ID,
		Command:    "CONNECT",
		SourceIP:   sourceIP,
		Timestamp:  time.Now(),
	}

	// Read initial SMB negotiation (simplified)
	buf := make([]byte, s.config.BufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("SMB read error from %s: %v", sourceIP, err)
		smbReq.Data = fmt.Sprintf("Error: %v", err)
	} else {
		smbReq.Data = fmt.Sprintf("Received %d bytes: %x", n, buf[:n])
		// Try to parse basic SMB commands
		s.parseSMBRequest(smbReq, buf[:n])
	}

	// Store the interaction
	if err := s.store.CreateListenerInteraction(ctx, &ListenerInteraction{
		ID:         generateID(),
		ListenerID: s.listener.ID,
		Protocol:   ProtocolSMB,
		SourceIP:   sourceIP,
		Data:       smbReq.Data,
		Metadata: map[string]string{
			"command":    smbReq.Command,
			"share_name": smbReq.ShareName,
			"file_path":  smbReq.FilePath,
			"username":   smbReq.Username,
		},
		Timestamp: time.Now(),
	}); err != nil {
		log.Printf("Failed to store SMB interaction: %v", err)
	}

	// Store detailed SMB request
	if err := s.store.CreateSMBRequest(ctx, smbReq); err != nil {
		log.Printf("Failed to store SMB request: %v", err)
	}
}

// parseSMBRequest parses basic SMB request data
func (s *SMBListener) parseSMBRequest(req *SMBRequest, data []byte) {
	if len(data) < 4 {
		return
	}

	// SMB magic bytes: "\xFFSMB"
	if string(data[0:4]) == "\xFFSMB" {
		req.Command = "NEGOTIATE"
	} else if len(data) >= 32 {
		// Try to extract SMB command from header
		command := data[4]
		switch command {
		case 0x00:
			req.Command = "SMB_COM_CREATE_DIRECTORY"
		case 0x01:
			req.Command = "SMB_COM_DELETE"
		case 0x02:
			req.Command = "SMB_COM_RENAME"
		case 0x03:
			req.Command = "SMB_COM_QUERY_INFORMATION"
		case 0x04:
			req.Command = "SMB_COM_SET_INFORMATION"
		case 0x05:
			req.Command = "SMB_COM_READ"
		case 0x06:
			req.Command = "SMB_COM_WRITE"
		case 0x0A:
			req.Command = "SMB_COM_OPEN_ANDX"
		case 0x2F:
			req.Command = "SMB_COM_TREE_CONNECT"
		case 0x34:
			req.Command = "SMB_COM_NT_CREATE_ANDX"
		default:
			req.Command = fmt.Sprintf("UNKNOWN(0x%02X)", command)
		}

		// Try to extract share name or file path from data
		dataStr := string(data)
		if idx := strings.Index(dataStr, "\\"); idx != -1 {
			parts := strings.Split(dataStr[idx:], "\\")
			if len(parts) > 0 {
				req.ShareName = parts[0]
			}
			if len(parts) > 1 {
				req.FilePath = strings.Join(parts[1:], "\\")
			}
		}
	}

	// Store raw data as JSON
	dataJSON, _ := json.Marshal(data)
	req.Data = string(dataJSON)
}

// generateID generates a unique ID
func generateID() string {
	return fmt.Sprintf("smb-%d", time.Now().UnixNano())
}
