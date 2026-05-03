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
		To:         []string{"recipient@example.com"},
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
	q := LDAPQuery{
		ID:         "test-query-1",
		ListenerID: "test-listener-1",
		BaseDN:     "dc=example,dc=com",
		Filter:     "(objectClass=user)",
		SourceIP:   "192.168.1.1",
		Timestamp:  now,
	}

	if q.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if q.BaseDN == "" {
		t.Fatal("BaseDN should not be empty")
	}
}
