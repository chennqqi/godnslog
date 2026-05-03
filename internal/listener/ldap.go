package listener

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// LDAPListener implements an LDAP server for OAST
type LDAPListener struct {
	listener *Listener
	config   *ListenerConfig
	server   net.Listener
	store    Store
	logger   *logrus.Logger
}

// NewLDAPListener creates a new LDAP listener
func NewLDAPListener(listener *Listener, config *ListenerConfig, store Store, logger *logrus.Logger) *LDAPListener {
	if config == nil {
		config = DefaultLDAPConfig()
	}
	return &LDAPListener{
		listener: listener,
		config:   config,
		store:    store,
		logger:   logger,
	}
}

// DefaultLDAPConfig returns default LDAP configuration
func DefaultLDAPConfig() *ListenerConfig {
	return &ListenerConfig{
		MaxConnections: 100,
		Timeout:        30 * time.Second,
		BufferSize:     4096,
		EnableTLS:      false,
	}
}

// Start starts the LDAP listener
func (l *LDAPListener) Start(ctx context.Context) error {
	addr := fmt.Sprintf("%s:%d", l.listener.Host, l.listener.Port)

	server, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to start LDAP listener: %w", err)
	}

	l.server = server
	l.logger.Printf("LDAP listener started on %s", addr)

	// Accept connections
	go l.acceptConnections(ctx)

	return nil
}

// Stop stops the LDAP listener
func (l *LDAPListener) Stop() error {
	if l.server != nil {
		return l.server.Close()
	}
	return nil
}

// acceptConnections accepts incoming connections
func (l *LDAPListener) acceptConnections(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := l.server.Accept()
			if err != nil {
				l.logger.Printf("LDAP accept error: %v", err)
				continue
			}

			go l.handleConnection(ctx, conn)
		}
	}
}

// handleConnection handles a single LDAP connection
func (l *LDAPListener) handleConnection(ctx context.Context, conn net.Conn) {
	defer conn.Close()

	// Set timeout
	if l.config.Timeout > 0 {
		conn.SetDeadline(time.Now().Add(l.config.Timeout))
	}

	remoteAddr := conn.RemoteAddr().String()
	sourceIP := parseIPFromAddr(remoteAddr)

	l.logger.Printf("LDAP connection from %s", remoteAddr)

	// Read LDAP message
	buf := make([]byte, l.config.BufferSize)
	n, err := conn.Read(buf)
	if err != nil {
		l.logger.Printf("LDAP read error: %v", err)
		return
	}

	// Parse LDAP message
	message := buf[:n]
	ldapMsg, err := l.parseLDAPMessage(message)
	if err != nil {
		l.logger.Printf("LDAP parse error: %v", err)
		return
	}

	l.logger.Printf("LDAP message: %+v", ldapMsg)

	// Save LDAP query
	query := &LDAPQuery{
		ID:         generateQueryID(),
		ListenerID: l.listener.ID,
		BaseDN:     ldapMsg.BaseDN,
		Filter:     ldapMsg.Filter,
		BindDN:     ldapMsg.BindDN,
		SourceIP:   sourceIP,
		Timestamp:  time.Now(),
	}

	if err := l.store.SaveLDAPQuery(ctx, query); err != nil {
		l.logger.Printf("Failed to save LDAP query: %v", err)
	} else {
		l.logger.Printf("LDAP query saved: %s", query.Filter)
	}

	// Save as general interaction
	interaction := &ListenerInteraction{
		ID:         generateInteractionIDLDAP(),
		ListenerID: l.listener.ID,
		Protocol:   ProtocolLDAP,
		SourceIP:   sourceIP,
		Data:       fmt.Sprintf("LDAP: %s, Filter: %s", query.BaseDN, query.Filter),
		Timestamp:  time.Now(),
	}

	if err := l.store.SaveListenerInteraction(ctx, interaction); err != nil {
		l.logger.Printf("Failed to save interaction: %v", err)
	}

	// Send response (simplified)
	response := l.buildLDAPResponse()
	conn.Write(response)
}

// LDAPMessage represents a parsed LDAP message
type LDAPMessage struct {
	MessageID  int
	Operation  int
	BaseDN     string
	Filter     string
	BindDN     string
	Attributes []string
}

// parseLDAPMessage parses an LDAP message (simplified)
func (l *LDAPListener) parseLDAPMessage(data []byte) (*LDAPMessage, error) {
	msg := &LDAPMessage{}

	// LDAP message format (simplified ASN.1 BER)
	// This is a basic parser for MVP
	// In production, use a proper LDAP library

	if len(data) < 2 {
		return nil, fmt.Errorf("message too short")
	}

	// Skip tag and length
	offset := 2

	// Parse message ID
	if offset < len(data) {
		msg.MessageID = int(data[offset])
		offset++
	}

	// Parse operation (simplified)
	if offset < len(data) {
		msg.Operation = int(data[offset])
		offset++
	}

	// Extract strings from data (simplified)
	dataStr := string(data)

	// Try to extract DN and filter
	// This is very basic parsing
	if strings.Contains(dataStr, "dc=") || strings.Contains(dataStr, "cn=") {
		msg.BaseDN = extractDN(dataStr)
	}

	if strings.Contains(dataStr, "filter") || strings.Contains(dataStr, "(") {
		msg.Filter = extractFilter(dataStr)
	}

	return msg, nil
}

// extractDN extracts DN from LDAP data
func extractDN(data string) string {
	// Simple DN extraction
	// In production, use proper LDAP parsing
	if strings.Contains(data, "dc=") {
		start := strings.Index(data, "dc=")
		if start >= 0 {
			end := strings.Index(data[start:], ",")
			if end >= 0 {
				return data[start : start+end]
			}
			return data[start:]
		}
	}
	return ""
}

// extractFilter extracts filter from LDAP data
func extractFilter(data string) string {
	// Simple filter extraction
	// In production, use proper LDAP parsing
	if strings.Contains(data, "(") && strings.Contains(data, ")") {
		start := strings.Index(data, "(")
		end := strings.LastIndex(data, ")")
		if start >= 0 && end > start {
			return data[start : end+1]
		}
	}
	return ""
}

// buildLDAPResponse builds an LDAP response (simplified)
func (l *LDAPListener) buildLDAPResponse() []byte {
	// Simplified LDAP response
	// In production, use proper LDAP library

	// Basic ASN.1 BER response
	response := []byte{
		0x30, // Sequence tag
		0x0A, // Length
		0x02, // Integer tag
		0x01, // Length
		0x01, // Message ID
		0x61, // BindResponse tag
		0x05, // Length
		0x0A, // Result code tag
		0x01, // Length
		0x00, // Success (0x00)
	}

	return response
}

// parseIPFromAddrLDAP parses IP from address string (LDAP version)
func parseIPFromAddrLDAP(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return addr
	}
	return host
}

// generateQueryID generates a unique query ID
func generateQueryID() string {
	return fmt.Sprintf("ldap-%d", time.Now().UnixNano())
}

// generateInteractionIDLDAP generates a unique interaction ID (LDAP version)
func generateInteractionIDLDAP() string {
	return fmt.Sprintf("inter-%d", time.Now().UnixNano())
}

// Helper functions for ASN.1 parsing
func readLength(data []byte, offset *int) int {
	if *offset >= len(data) {
		return 0
	}

	length := int(data[*offset])
	*offset++

	if length&0x80 != 0 {
		// Long form
		numBytes := length & 0x7F
		length = 0
		for i := 0; i < numBytes; i++ {
			if *offset >= len(data) {
				break
			}
			length = length<<8 | int(data[*offset])
			*offset++
		}
	}

	return length
}

func readInteger(data []byte, offset *int) int {
	if *offset >= len(data) {
		return 0
	}

	length := readLength(data, offset)
	if length == 0 || *offset+length > len(data) {
		return 0
	}

	value := 0
	for i := 0; i < length; i++ {
		value = value<<8 | int(data[*offset])
		*offset++
	}

	return value
}

func readString(data []byte, offset *int) string {
	if *offset >= len(data) {
		return ""
	}

	length := readLength(data, offset)
	if length == 0 || *offset+length > len(data) {
		return ""
	}

	str := string(data[*offset : *offset+length])
	*offset += length

	return str
}

// encodeLDAPMessage encodes an LDAP message
func encodeLDAPMessage(messageID int, operation int, data []byte) []byte {
	buf := make([]byte, 0, 1024)

	// Message ID
	buf = append(buf, 0x02) // Integer tag
	buf = append(buf, byte(len([]byte{byte(messageID)})))
	buf = append(buf, byte(messageID))

	// Operation
	buf = append(buf, byte(operation))
	buf = append(buf, byte(len(data)))
	buf = append(buf, data...)

	// Wrap in sequence
	result := make([]byte, 0, len(buf)+2)
	result = append(result, 0x30) // Sequence tag
	result = append(result, byte(len(buf)))
	result = append(result, buf...)

	return result
}

// Helper for ASN.1 encoding
func encodeInteger(value int) []byte {
	if value < 0x80 {
		return []byte{byte(value)}
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(value))

	// Remove leading zeros
	for len(buf) > 1 && buf[0] == 0 {
		buf = buf[1:]
	}

	return buf
}

func encodeString(str string) []byte {
	return []byte(str)
}

func encodeLength(length int) []byte {
	if length < 0x80 {
		return []byte{byte(length)}
	}

	buf := make([]byte, 0, 4)
	for length > 0 {
		buf = append([]byte{byte(length & 0xFF)}, buf...)
		length >>= 8
	}

	result := append([]byte{byte(0x80 | len(buf))}, buf...)
	return result
}
