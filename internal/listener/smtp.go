package listener

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// SMTPListener implements an SMTP server for OAST
type SMTPListener struct {
	listener *Listener
	config   *ListenerConfig
	server   net.Listener
	store    Store
	logger   *logrus.Logger
}

// NewSMTPListener creates a new SMTP listener
func NewSMTPListener(listener *Listener, config *ListenerConfig, store Store, logger *logrus.Logger) *SMTPListener {
	if config == nil {
		config = DefaultSMTPConfig()
	}
	return &SMTPListener{
		listener: listener,
		config:   config,
		store:    store,
		logger:   logger,
	}
}

// DefaultSMTPConfig returns default SMTP configuration
func DefaultSMTPConfig() *ListenerConfig {
	return &ListenerConfig{
		MaxConnections: 100,
		Timeout:        30 * time.Second,
		BufferSize:     4096,
		EnableTLS:      false,
	}
}

// Start starts the SMTP listener
func (l *SMTPListener) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", l.listener.Host, l.listener.Port)

	server, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start SMTP listener: %w", err)
	}

	l.server = server
	l.logger.Printf("SMTP listener started on %s", addr)

	// Accept connections
	go l.acceptConnections(ctx)

	return nil
}

// Stop stops the SMTP listener
func (l *SMTPListener) Stop() error {
	if l.server != nil {
		return l.server.Close()
	}
	return nil
}

// acceptConnections accepts incoming connections
func (l *SMTPListener) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := l.server.Accept()
			if err != nil {
				l.logger.Printf("SMTP accept error: %v", err)
				continue
			}

			go l.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles a single SMTP connection
func (l *SMTPListener) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Set timeout
	if l.config.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(l.config.Timeout))
	}

	remoteAddr := conn.RemoteAddr().String()
	sourceIP := parseIPFromAddr(remoteAddr)

	l.logger.Printf("SMTP connection from %s", remoteAddr)

	// Send greeting
	l.sendResponse(conn, "220 godnslog SMTP OAST Service")

	reader := bufio.NewReader(conn)

	// SMTP state machine
	var from string
	var to []string
	var headers []string
	var body strings.Builder

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)
		l.logger.Printf("SMTP command: %s", line)

		// Parse SMTP commands
		if strings.HasPrefix(strings.ToUpper(line), "HELO") || strings.HasPrefix(strings.ToUpper(line), "EHLO") {
			l.sendResponse(conn, "250 OK")
		} else if strings.HasPrefix(strings.ToUpper(line), "MAIL FROM:") {
			from = extractEmail(line)
			l.sendResponse(conn, "250 OK")
		} else if strings.HasPrefix(strings.ToUpper(line), "RCPT TO:") {
			email := extractEmail(line)
			to = append(to, email)
			l.sendResponse(conn, "250 OK")
		} else if strings.ToUpper(line) == "DATA" {
			l.sendResponse(conn, "354 End data with <CRLF>.<CRLF>")
			l.readData(reader, &headers, &body)
			l.sendResponse(conn, "250 OK")
		} else if strings.ToUpper(line) == "QUIT" {
			l.sendResponse(conn, "221 Bye")
			break
		} else if strings.ToUpper(line) == "RSET" {
			from = ""
			to = nil
			headers = nil
			body.Reset()
			l.sendResponse(conn, "250 OK")
		} else {
			l.sendResponse(conn, "500 Command not recognized")
		}
	}

	// Save message if we have data
	if from != "" && len(to) > 0 {
		// Convert to slice to JSON string
		toJSON, _ := json.Marshal(to)
		message := &SMTPMessage{
			ID:         generateMessageID(),
			ListenerID: l.listener.ID,
			From:       from,
			To:         string(toJSON),
			Body:       body.String(),
			Headers:    strings.Join(headers, "\n"),
			SourceIP:   sourceIP,
			Timestamp:  time.Now(),
		}

		if err := l.store.SaveSMTPMessage(ctx, message); err != nil {
			l.logger.Printf("Failed to save SMTP message: %v", err)
		} else {
			l.logger.Printf("SMTP message saved: %s -> %v", from, to)
		}

		// Also save as general interaction
		interaction := &ListenerInteraction{
			ID:         generateInteractionIDSMTP(),
			ListenerID: l.listener.ID,
			Protocol:   ProtocolSMTP,
			SourceIP:   sourceIP,
			Data:       fmt.Sprintf("SMTP: %s -> %v", from, to),
			Timestamp:  time.Now(),
		}

		if err := l.store.SaveListenerInteraction(ctx, interaction); err != nil {
			l.logger.Printf("Failed to save interaction: %v", err)
		}
	}
}

// readData reads the DATA section
func (l *SMTPListener) readData(reader *bufio.Reader, headers *[]string, body *strings.Builder) {
	inHeaders := true

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.TrimSpace(line)

		// End of data
		if line == "." {
			break
		}

		if inHeaders {
			if line == "" {
				inHeaders = false
			} else {
				*headers = append(*headers, line)
			}
		} else {
			body.WriteString(line)
			body.WriteString("\n")
		}
	}
}

// sendResponse sends an SMTP response
func (l *SMTPListener) sendResponse(conn net.Conn, message string) {
	_, err := conn.Write([]byte(message + "\r\n"))
	if err != nil {
		l.logger.Printf("Failed to send response: %v", err)
	}
}

// extractEmail extracts email from MAIL FROM: or RCPT TO: command
func extractEmail(line string) string {
	// Extract email from "<email@domain.com>" format
	start := strings.Index(line, "<")
	end := strings.Index(line, ">")
	if start >= 0 && end > start {
		return line[start+1 : end]
	}
	return line
}

// parseIPFromAddr parses IP from address string
func parseIPFromAddr(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// generateMessageID generates a unique message ID
func generateMessageID() string {
	return fmt.Sprintf("smtp-%d", time.Now().UnixNano())
}

// generateInteractionIDSMTP generates a unique interaction ID (SMTP version)
func generateInteractionIDSMTP() string {
	return fmt.Sprintf("inter-%d", time.Now().UnixNano())
}
