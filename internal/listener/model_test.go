package listener

import (
	"testing"
	"time"
)

// TestListenerTableName tests table name
func TestListenerTableName(t *testing.T) {
	l := Listener{}
	tableName := l.TableName()
	if tableName != "listeners" {
		t.Fatalf("Expected table name 'listeners', got '%s'", tableName)
	}
}

// TestListenerInteractionTableName tests table name
func TestListenerInteractionTableName(t *testing.T) {
	i := ListenerInteraction{}
	tableName := i.TableName()
	if tableName != "listener_interactions" {
		t.Fatalf("Expected table name 'listener_interactions', got '%s'", tableName)
	}
}

// TestSMTPMessageTableName tests table name
func TestSMTPMessageTableName(t *testing.T) {
	m := SMTPMessage{}
	tableName := m.TableName()
	if tableName != "smtp_messages" {
		t.Fatalf("Expected table name 'smtp_messages', got '%s'", tableName)
	}
}

// TestLDAPQueryTableName tests table name
func TestLDAPQueryTableName(t *testing.T) {
	q := LDAPQuery{}
	tableName := q.TableName()
	if tableName != "ldap_queries" {
		t.Fatalf("Expected table name 'ldap_queries', got '%s'", tableName)
	}
}

// TestProtocol constants
func TestProtocol(t *testing.T) {
	protocols := []Protocol{
		ProtocolSMTP,
		ProtocolLDAP,
		ProtocolSMB,
		ProtocolFTP,
	}

	for _, protocol := range protocols {
		if protocol == "" {
			t.Fatal("Protocol should not be empty")
		}
	}
}

// TestListenerModel tests listener model
func TestListenerModel(t *testing.T) {
	now := time.Now()
	l := Listener{
		ID:        "test-listener-1",
		Protocol:  ProtocolSMTP,
		Host:      "0.0.0.0",
		Port:      25,
		Token:     "test-token",
		IsEnabled: true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if l.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if l.Protocol != ProtocolSMTP {
		t.Fatalf("Expected protocol '%s', got '%s'", ProtocolSMTP, l.Protocol)
	}
}

// TestListenerInteractionModel tests listener interaction model
func TestListenerInteractionModel(t *testing.T) {
	now := time.Now()
	i := ListenerInteraction{
		ID:         "test-interaction-1",
		ListenerID: "test-listener-1",
		Protocol:   ProtocolSMTP,
		SourceIP:   "192.168.1.1",
		Data:       "test data",
		Timestamp:  now,
	}

	if i.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if i.SourceIP == "" {
		t.Fatal("SourceIP should not be empty")
	}
}

// TestSMTPMessageModel tests SMTP message model
func TestSMTPMessageModel(t *testing.T) {
	now := time.Now()
	m := SMTPMessage{
		ID:         "test-message-1",
		ListenerID: "test-listener-1",
		From:       "sender@example.com",
		To:         `["recipient@example.com"]`,
		Subject:    "Test subject",
		Body:       "Test body",
		SourceIP:   "192.168.1.1",
		Timestamp:  now,
	}

	if m.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if m.From == "" {
		t.Fatal("From should not be empty")
	}
}

// TestLDAPQueryModel tests LDAP query model
func TestLDAPQueryModel(t *testing.T) {
	now := time.Now()
	query := &LDAPQuery{
		ID:         "query-1",
		ListenerID: "listener-1",
		BaseDN:     "dc=example,dc=com",
		Filter:     "(objectClass=user)",
		Attributes: `["cn", "mail"]`,
		BindDN:     "cn=admin,dc=example,dc=com",
		SourceIP:   "192.168.1.1",
		Timestamp:  now,
	}

	if query.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if query.BaseDN == "" {
		t.Fatal("BaseDN should not be empty")
	}

	if query.Filter == "" {
		t.Fatal("Filter should not be empty")
	}
}

// TestSMBRequestModel tests SMB request model
func TestSMBRequestModel(t *testing.T) {
	now := time.Now()
	request := &SMBRequest{
		ID:         "smb-1",
		ListenerID: "listener-1",
		Command:    "TREE_CONNECT",
		ShareName:  "share",
		FilePath:   "\\path\\to\\file",
		Username:   "user",
		Data:       `{"test": "data"}`,
		SourceIP:   "192.168.1.1",
		SourcePort: 445,
		Timestamp:  now,
	}

	if request.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if request.Command == "" {
		t.Fatal("Command should not be empty")
	}

	if request.SourceIP == "" {
		t.Fatal("SourceIP should not be empty")
	}
}

// TestFTPCommandModel tests FTP command model
func TestFTPCommandModel(t *testing.T) {
	now := time.Now()
	command := &FTPCommand{
		ID:         "ftp-1",
		ListenerID: "listener-1",
		Command:    "RETR",
		Argument:   "/path/to/file.txt",
		Username:   "user",
		Data:       "file content",
		SourceIP:   "192.168.1.1",
		SourcePort: 21,
		Timestamp:  now,
	}

	if command.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if command.Command == "" {
		t.Fatal("Command should not be empty")
	}

	if command.SourceIP == "" {
		t.Fatal("SourceIP should not be empty")
	}
}

func TestRMIInteractionModel(t *testing.T) {
	now := time.Now()
	interaction := &RMIInteraction{
		ID:         "rmi-test-1",
		ListenerID: "listener-1",
		SourceIP:   "192.168.1.1",
		SourcePort: 1099,
		URN:        "JRMP StreamProtocol",
		RawData:    "4a524d4900024b",
		Timestamp:  now,
	}

	if interaction.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if interaction.SourceIP == "" {
		t.Fatal("SourceIP should not be empty")
	}

	if interaction.URN == "" {
		t.Fatal("URN should not be empty")
	}
}

// TestRMIInteractionTableName tests table name
func TestRMIInteractionTableName(t *testing.T) {
	i := RMIInteraction{}
	tableName := i.TableName()
	if tableName != "rmi_interactions" {
		t.Fatalf("Expected table name 'rmi_interactions', got '%s'", tableName)
	}
}

// TestRMIProtocolParsing tests RMI handshake parsing
func TestRMIProtocolParsing(t *testing.T) {
	// Valid JRMP StreamProtocol
	data := []byte{0x4a, 0x52, 0x4d, 0x49, 0x00, 0x02, 0x4b, 0x00, 0x50}
	urn, ok := parseRMIHandshake(data)
	if !ok {
		t.Error("expected RMI handshake to be recognized")
	}
	if urn == "" {
		t.Error("expected non-empty URN")
	}

	// Not RMI
	_, ok2 := parseRMIHandshake([]byte{0x00, 0x00, 0x00, 0x00})
	if ok2 {
		t.Error("expected non-RMI data to not be recognized")
	}

	// Short data
	_, ok3 := parseRMIHandshake([]byte{0x4a})
	if ok3 {
		t.Error("expected short data to not be recognized")
	}

	// SingleOpProtocol
	data2 := []byte{0x4a, 0x52, 0x4d, 0x49, 0x00, 0x02, 0x4c}
	urn2, ok4 := parseRMIHandshake(data2)
	if !ok4 {
		t.Error("expected SingleOpProtocol to be recognized")
	}
	if urn2 != "JRMP SingleOpProtocol" {
		t.Errorf("expected 'JRMP SingleOpProtocol', got '%s'", urn2)
	}
}
