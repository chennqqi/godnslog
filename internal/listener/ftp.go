package listener

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"time"
)

// FTPListener handles FTP protocol interactions
type FTPListener struct {
	config   *ListenerConfig
	store    Store
	listener *Listener
	server   net.Listener
	stopChan chan struct{}
}

// NewFTPListener creates a new FTP listener
func NewFTPListener(config *ListenerConfig, store Store, listener *Listener) *FTPListener {
	if config == nil {
		config = &ListenerConfig{
			MaxConnections: 10,
			Timeout:        30 * time.Second,
			BufferSize:     4096,
		}
	}
	return &FTPListener{
		config:   config,
		store:    store,
		listener: listener,
		stopChan: make(chan struct{}),
	}
}

// Start starts the FTP listener
func (f *FTPListener) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", f.listener.Host, f.listener.Port)

	var err error
	f.server, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start FTP listener: %w", err)
	}

	log.Printf("FTP listener started on %s", addr)

	go f.acceptConnections(ctx)

	return nil
}

// Stop stops the FTP listener
func (f *FTPListener) Stop() error {
	close(f.stopChan)
	if f.server != nil {
		return f.server.Close()
	}
	return nil
}

// acceptConnections accepts incoming FTP connections
func (f *FTPListener) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-f.stopChan:
			return
		default:
			conn, err := f.server.Accept()
			if err != nil {
				select {
				case <-f.stopChan:
					return
				default:
					log.Printf("FTP accept error: %v", err)
					continue
				}
			}

			go f.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles an FTP connection
func (f *FTPListener) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	remoteAddr := conn.RemoteAddr().String()
	sourceIP, sourcePortStr, err := net.SplitHostPort(remoteAddr)
	sourcePort := 0
	if err != nil {
		sourceIP = remoteAddr
	} else {
		// Parse port from string
		var port int
		_, err := fmt.Sscanf(sourcePortStr, "%d", &port)
		if err == nil {
			sourcePort = port
		}
	}

	log.Printf("FTP connection from %s", sourceIP)

	// Set timeout
	if f.config.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(f.config.Timeout))
	}

	// Send FTP welcome message
	conn.Write([]byte("220 GODNSLOG FTP Server\r\n"))

	reader := bufio.NewReader(conn)
	username := ""

	// Handle FTP commands
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			log.Printf("FTP read error from %s: %v", sourceIP, err)
			break
		}

		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		// Parse FTP command
		parts := strings.SplitN(line, " ", 2)
		command := strings.ToUpper(parts[0])
		argument := ""
		if len(parts) > 1 {
			argument = parts[1]
		}

		// Create FTP command record
		ftpCmd := &FTPCommand{
			ID:         generateFTPID(),
			ListenerID: f.listener.ID,
			Command:    command,
			Argument:   argument,
			Username:   username,
			SourceIP:   sourceIP,
			SourcePort: sourcePort,
			Timestamp:  time.Now(),
		}

		// Handle common FTP commands
		switch command {
		case "USER":
			username = argument
			conn.Write([]byte("331 Username OK, need password\r\n"))
		case "PASS":
			conn.Write([]byte("230 Login successful\r\n"))
		case "QUIT":
			conn.Write([]byte("221 Goodbye\r\n"))
			break
		case "SYST":
			conn.Write([]byte("215 UNIX Type: L8\r\n"))
		case "TYPE":
			conn.Write([]byte("200 Type set to I\r\n"))
		case "PASV":
			conn.Write([]byte("227 Entering Passive Mode (127,0,0,1,196,173)\r\n"))
		case "LIST":
			conn.Write([]byte("150 Here comes the directory listing\r\n"))
			conn.Write([]byte("226 Directory send OK\r\n"))
		case "RETR", "STOR":
			ftpCmd.Data = argument
			conn.Write([]byte("550 File not found\r\n"))
		default:
			conn.Write([]byte("502 Command not implemented\r\n"))
		}

		// Store the interaction
		if err := f.store.CreateListenerInteraction(ctx, &ListenerInteraction{
			ID:         generateFTPID(),
			ListenerID: f.listener.ID,
			Protocol:   ProtocolFTP,
			SourceIP:   sourceIP,
			Data:       line,
			Metadata: map[string]string{
				"command":  command,
				"argument": argument,
				"username": username,
			},
			Timestamp: time.Now(),
		}); err != nil {
			log.Printf("Failed to store FTP interaction: %v", err)
		}

		// Store detailed FTP command
		if err := f.store.CreateFTPCommand(ctx, ftpCmd); err != nil {
			log.Printf("Failed to store FTP command: %v", err)
		}
	}
}

// generateFTPID generates a unique ID for FTP
func generateFTPID() string {
	return fmt.Sprintf("ftp-%d", time.Now().UnixNano())
}
