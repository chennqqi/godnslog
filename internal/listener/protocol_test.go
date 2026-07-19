package listener

import (
	"bufio"
	"context"
	"io"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
)

func TestExtractEmail(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"MAIL FROM:<user@example.com>", "user@example.com"},
		{"RCPT TO:<admin@test.com>", "admin@test.com"},
		{"<simple@domain.com>", "simple@domain.com"},
		{"no-angle-brackets", "no-angle-brackets"},
		{"", ""},
		{"<only-open", "<only-open"},
		{">only-close", ">only-close"},
	}
	for _, tc := range tests {
		got := extractEmail(tc.input)
		if got != tc.want {
			t.Errorf("extractEmail(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestParseIPFromAddr(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"192.168.1.1:12345", "192.168.1.1"},
		{"10.0.0.1:80", "10.0.0.1"},
		{"[::1]:443", "::1"},
		{"no-port", "no-port"},
		{"", ""},
	}
	for _, tc := range tests {
		got := parseIPFromAddr(tc.input)
		if got != tc.want {
			t.Errorf("parseIPFromAddr(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExtractDN(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"dc=example,dc=com", "dc=example"},
		{"cn=admin,dc=test", "dc=test"},
		{"no-dc-here", ""},
		{"", ""},
	}
	for _, tc := range tests {
		got := extractDN(tc.input)
		if got != tc.want {
			t.Errorf("extractDN(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestExtractFilter(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"(objectClass=user)", "(objectClass=user)"},
		{"prefix (cn=test) suffix", "(cn=test)"},
		{"no-parentheses", ""},
		{"(unmatched", ""},
		{"", ""},
	}
	for _, tc := range tests {
		got := extractFilter(tc.input)
		if got != tc.want {
			t.Errorf("extractFilter(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestBuildLDAPResponse(t *testing.T) {
	l := &LDAPListener{}
	resp := l.buildLDAPResponse()
	if len(resp) == 0 {
		t.Fatal("expected non-empty LDAP response")
	}
	// Should start with sequence tag
	if resp[0] != 0x30 {
		t.Errorf("expected sequence tag 0x30, got 0x%02x", resp[0])
	}
}

func TestReadLength(t *testing.T) {
	t.Run("short form", func(t *testing.T) {
		data := []byte{0x05}
		offset := 0
		length := readLength(data, &offset)
		if length != 5 {
			t.Errorf("expected 5, got %d", length)
		}
		if offset != 1 {
			t.Errorf("expected offset 1, got %d", offset)
		}
	})

	t.Run("offset past end", func(t *testing.T) {
		data := []byte{0x01}
		offset := 5
		length := readLength(data, &offset)
		if length != 0 {
			t.Errorf("expected 0, got %d", length)
		}
	})
}

func TestReadInteger(t *testing.T) {
	t.Run("simple integer", func(t *testing.T) {
		data := []byte{0x02, 0x01, 0x2A} // tag, length, value(42)
		offset := 1                      // skip tag
		value := readInteger(data, &offset)
		if value != 42 {
			t.Errorf("expected 42, got %d", value)
		}
	})

	t.Run("offset past end", func(t *testing.T) {
		data := []byte{0x00}
		offset := 5
		value := readInteger(data, &offset)
		if value != 0 {
			t.Errorf("expected 0, got %d", value)
		}
	})
}

func TestReadString(t *testing.T) {
	t.Run("valid string", func(t *testing.T) {
		data := []byte{0x04, 0x05, 'h', 'e', 'l', 'l', 'o'} // tag, length, "hello"
		offset := 1                                         // skip tag
		str := readString(data, &offset)
		if str != "hello" {
			t.Errorf("expected 'hello', got '%s'", str)
		}
	})

	t.Run("offset past end", func(t *testing.T) {
		data := []byte{0x00}
		offset := 5
		str := readString(data, &offset)
		if str != "" {
			t.Errorf("expected empty, got '%s'", str)
		}
	})
}

func TestEncodeLDAPMessage(t *testing.T) {
	result := encodeLDAPMessage(1, 0x61, []byte{0x0a, 0x01, 0x00})
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
	if result[0] != 0x30 {
		t.Errorf("expected sequence tag 0x30, got 0x%02x", result[0])
	}
}

func TestEncodeInteger(t *testing.T) {
	t.Run("small value", func(t *testing.T) {
		result := encodeInteger(42)
		if len(result) != 1 || result[0] != 42 {
			t.Errorf("expected [42], got %v", result)
		}
	})

	t.Run("large value", func(t *testing.T) {
		result := encodeInteger(256)
		if len(result) != 2 {
			t.Errorf("expected 2 bytes, got %d", len(result))
		}
	})
}

func TestEncodeString(t *testing.T) {
	result := encodeString("hello")
	if string(result) != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestEncodeLength(t *testing.T) {
	t.Run("short form", func(t *testing.T) {
		result := encodeLength(5)
		if len(result) != 1 || result[0] != 5 {
			t.Errorf("expected [5], got %v", result)
		}
	})

	t.Run("long form", func(t *testing.T) {
		result := encodeLength(256)
		if len(result) != 3 {
			t.Errorf("expected 3 bytes, got %d (%v)", len(result), result)
		}
		if result[0] != 0x82 {
			t.Errorf("expected long-form indicator 0x82, got 0x%02x", result[0])
		}
	})
}

func TestDefaultSMTPConfig(t *testing.T) {
	cfg := DefaultSMTPConfig()
	if cfg.MaxConnections != 1000 {
		t.Errorf("expected MaxConnections 1000, got %d", cfg.MaxConnections)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("expected Timeout 30s, got %v", cfg.Timeout)
	}
	if cfg.RateLimitMax != 1000 {
		t.Errorf("expected RateLimitMax 1000, got %d", cfg.RateLimitMax)
	}
}

func TestDefaultLDAPConfig(t *testing.T) {
	cfg := DefaultLDAPConfig()
	if cfg.MaxConnections != 1000 {
		t.Errorf("expected MaxConnections 1000, got %d", cfg.MaxConnections)
	}
	if cfg.MaxConcurrentConnections != 500 {
		t.Errorf("expected MaxConcurrentConnections 500, got %d", cfg.MaxConcurrentConnections)
	}
}

func TestDefaultSMBConfig(t *testing.T) {
	cfg := DefaultSMBConfig()
	if cfg.MaxConnections != 500 {
		t.Errorf("expected MaxConnections 500, got %d", cfg.MaxConnections)
	}
}

func TestDefaultFTPConfig(t *testing.T) {
	cfg := DefaultFTPConfig()
	if cfg.MaxConnections != 500 {
		t.Errorf("expected MaxConnections 500, got %d", cfg.MaxConnections)
	}
}

func TestDefaultRMIConfig(t *testing.T) {
	cfg := DefaultRMIConfig()
	if cfg.MaxConnections != 500 {
		t.Errorf("expected MaxConnections 500, got %d", cfg.MaxConnections)
	}
}

func TestGenerateID_Functions(t *testing.T) {
	// Test that all generate functions produce non-empty, unique IDs
	id1 := generateListenerID()
	id2 := generateListenerID()
	if id1 == "" || id2 == "" {
		t.Error("expected non-empty IDs")
	}
	if id1 == id2 {
		t.Error("expected unique listener IDs")
	}

	ftp1 := generateFTPID()
	ftp2 := generateFTPID()
	if ftp1 == ftp2 {
		t.Error("expected unique FTP IDs")
	}

	query1 := generateQueryID()
	query2 := generateQueryID()
	if query1 == query2 {
		t.Error("expected unique query IDs")
	}

	msg1 := generateMessageID()
	msg2 := generateMessageID()
	if msg1 == msg2 {
		t.Error("expected unique message IDs")
	}

	inter1 := generateInteractionIDLDAP()
	inter2 := generateInteractionIDLDAP()
	if inter1 == inter2 {
		t.Error("expected unique interaction IDs")
	}

	interS1 := generateInteractionIDSMTP()
	interS2 := generateInteractionIDSMTP()
	if interS1 == interS2 {
		t.Error("expected unique SMTP interaction IDs")
	}

	interR1 := generateInteractionID()
	interR2 := generateInteractionID()
	if interR1 == interR2 {
		t.Error("expected unique rejected interaction IDs")
	}

	smbID := generateID()
	if smbID == "" {
		t.Error("expected non-empty SMB ID")
	}
}

func TestParseSMBRequest(t *testing.T) {
	s := &SMBListener{}

	t.Run("SMB magic bytes", func(t *testing.T) {
		req := &SMBRequest{}
		data := []byte{0xFF, 'S', 'M', 'B', 0x00}
		s.parseSMBRequest(req, data)
		if req.Command != "NEGOTIATE" {
			t.Errorf("expected NEGOTIATE, got %s", req.Command)
		}
	})

	t.Run("short data", func(t *testing.T) {
		req := &SMBRequest{Command: "CONNECT"}
		data := []byte{0x01, 0x02}
		s.parseSMBRequest(req, data)
		if req.Command != "CONNECT" {
			t.Errorf("expected unchanged CONNECT, got %s", req.Command)
		}
	})

	t.Run("SMB_COM_CREATE_DIRECTORY", func(t *testing.T) {
		req := &SMBRequest{}
		data := make([]byte, 32)
		data[4] = 0x00
		s.parseSMBRequest(req, data)
		if req.Command != "SMB_COM_CREATE_DIRECTORY" {
			t.Errorf("expected SMB_COM_CREATE_DIRECTORY, got %s", req.Command)
		}
	})

	t.Run("SMB_COM_DELETE", func(t *testing.T) {
		req := &SMBRequest{}
		data := make([]byte, 32)
		data[4] = 0x01
		s.parseSMBRequest(req, data)
		if req.Command != "SMB_COM_DELETE" {
			t.Errorf("expected SMB_COM_DELETE, got %s", req.Command)
		}
	})

	t.Run("SMB_COM_TREE_CONNECT", func(t *testing.T) {
		req := &SMBRequest{}
		data := make([]byte, 32)
		data[4] = 0x2F
		s.parseSMBRequest(req, data)
		if req.Command != "SMB_COM_TREE_CONNECT" {
			t.Errorf("expected SMB_COM_TREE_CONNECT, got %s", req.Command)
		}
	})

	t.Run("SMB_COM_NT_CREATE_ANDX", func(t *testing.T) {
		req := &SMBRequest{}
		data := make([]byte, 32)
		data[4] = 0x34
		s.parseSMBRequest(req, data)
		if req.Command != "SMB_COM_NT_CREATE_ANDX" {
			t.Errorf("expected SMB_COM_NT_CREATE_ANDX, got %s", req.Command)
		}
	})

	t.Run("unknown command", func(t *testing.T) {
		req := &SMBRequest{}
		data := make([]byte, 32)
		data[4] = 0xFF
		s.parseSMBRequest(req, data)
		if req.Command != "UNKNOWN(0xFF)" {
			t.Errorf("expected UNKNOWN(0xFF), got %s", req.Command)
		}
	})

	t.Run("extract share path", func(t *testing.T) {
		req := &SMBRequest{}
		data := make([]byte, 48)
		data[4] = 0x2F
		copy(data[32:], "\\SHARE\\path")
		s.parseSMBRequest(req, data)
		if req.Command != "SMB_COM_TREE_CONNECT" {
			t.Errorf("expected SMB_COM_TREE_CONNECT, got %s", req.Command)
		}
		if req.FilePath == "" {
			t.Error("expected non-empty FilePath")
		}
	})
}

func TestGenerateListenerID(t *testing.T) {
	id := generateListenerID()
	if len(id) < 9 {
		t.Errorf("listener ID too short: %s", id)
	}
	// Should start with "listener-"
	if len(id) < 10 || id[:9] != "listener-" {
		t.Errorf("expected listener- prefix, got %s", id)
	}
}

func TestParseSMBRequest_ShortDataReturns(t *testing.T) {
	// Edge case: data with more than 4 bytes but less than 32 bytes (no command extraction)
	s := &SMBListener{}
	req := &SMBRequest{Command: "CONNECT"}
	data := []byte{0xFF, 0x00, 0x00, 0x00, 0x00}
	s.parseSMBRequest(req, data)
	// 4 bytes is not SMB magic, and len < 32, so command stays CONNECT
	if req.Command != "CONNECT" {
		t.Errorf("expected CONNECT, got %s", req.Command)
	}
}

func TestLDAPParser_MinimalMessage(t *testing.T) {
	// Message too short to parse
	l := &LDAPListener{}
	msg, err := l.parseLDAPMessage([]byte{0x00})
	if err == nil {
		t.Error("expected error for too-short message")
	}
	if msg != nil {
		t.Errorf("expected nil message, got %+v", msg)
	}
}

func TestLDAPParser_SimpleMessage(t *testing.T) {
	// Minimal valid-length message
	l := &LDAPListener{}
	msg, err := l.parseLDAPMessage([]byte{
		0x30, 0x0C, // Sequence of length 12
		0x02, 0x01, 0x01, // Integer: message ID = 1
		0x63, 0x07, // BindRequest tag
		0x04, 0x05, 'a', 'd', 'm', 'i', 'n', // string "admin"
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg == nil {
		t.Fatal("expected non-nil message")
	}
	if msg.MessageID != 2 {
		t.Errorf("expected MessageID 2, got %d", msg.MessageID)
	}
}

func TestLDAPParser_WithDNFilter(t *testing.T) {
	// Message containing DN and filter-like content
	l := &LDAPListener{}
	data := []byte("dc=example,dc=com (objectClass=*)")
	msg, err := l.parseLDAPMessage(data)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.BaseDN != "dc=example" {
		t.Errorf("expected BaseDN 'dc=example', got '%s'", msg.BaseDN)
	}
	if msg.Filter != "(objectClass=*)" {
		t.Errorf("expected Filter '(objectClass=*)', got '%s'", msg.Filter)
	}
}

func TestSecurityContext_RateLimiterGetter(t *testing.T) {
	rl := NewRateLimiter(1, 10)
	cl := NewConnLimiter(5)
	sc := NewSecurityContext(rl, cl, nil, nil, "test", ProtocolSMTP, nil)

	retrievedRL := sc.RateLimiter()
	if retrievedRL != rl {
		t.Error("RateLimiter() should return the same instance")
	}

	retrievedCL := sc.ConnLimiter()
	if retrievedCL != cl {
		t.Error("ConnLimiter() should return the same instance")
	}
}

func TestSecurityContext_RateLimiterGetterNil(t *testing.T) {
	var sc *SecurityContext
	if rl := sc.RateLimiter(); rl != nil {
		t.Error("nil SecurityContext should return nil RateLimiter")
	}
	if cl := sc.ConnLimiter(); cl != nil {
		t.Error("nil SecurityContext should return nil ConnLimiter")
	}
}

func TestLoadTLSConfig_InvalidFiles(t *testing.T) {
	_, err := loadTLSConfig("/nonexistent/cert.pem", "/nonexistent/key.pem")
	if err == nil {
		t.Error("expected error for nonexistent TLS files")
	}
}

func TestParseRMIHandshake(t *testing.T) {
	// Already tested in model_test.go, but include additional edge cases
	t.Run("stream protocol no port", func(t *testing.T) {
		data := []byte{0x4a, 0x52, 0x4d, 0x49, 0x00, 0x02, 0x4b}
		urn, ok := parseRMIHandshake(data)
		if !ok {
			t.Error("expected recognized")
		}
		if urn != "JRMP StreamProtocol" {
			t.Errorf("expected 'JRMP StreamProtocol', got '%s'", urn)
		}
	})

	t.Run("stream protocol with port", func(t *testing.T) {
		data := []byte{0x4a, 0x52, 0x4d, 0x49, 0x00, 0x02, 0x4b, 0x08, 0x00}
		urn, ok := parseRMIHandshake(data)
		if !ok {
			t.Error("expected recognized")
		}
		if urn != "JRMP StreamProtocol (callback port: 2048)" {
			t.Errorf("unexpected URN: %s", urn)
		}
	})

	t.Run("unknown protocol type", func(t *testing.T) {
		data := []byte{0x4a, 0x52, 0x4d, 0x49, 0x00, 0x02, 0x01}
		urn, ok := parseRMIHandshake(data)
		if !ok {
			t.Error("expected recognized")
		}
		if urn != "JRMP type: 0x01" {
			t.Errorf("unexpected URN: %s", urn)
		}
	})
}

func TestNewFTPListener(t *testing.T) {
	// Test with nil config
	l := NewFTPListener(nil, nil, &Listener{ID: "ftp-1", Host: "0.0.0.0", Port: 21})
	if l == nil {
		t.Fatal("expected non-nil listener")
	}
	if l.config == nil {
		t.Fatal("expected non-nil config")
	}
	if l.config.MaxConnections != 10 {
		t.Errorf("expected default MaxConnections 10, got %d", l.config.MaxConnections)
	}
}

func TestNewSMBListener(t *testing.T) {
	// Test with nil config
	l := NewSMBListener(nil, nil, &Listener{ID: "smb-1"})
	if l == nil {
		t.Fatal("expected non-nil listener")
	}
	if l.config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestNewRMIListener(t *testing.T) {
	l := NewRMIListener(&Listener{ID: "rmi-1"}, nil, nil, nil)
	if l == nil {
		t.Fatal("expected non-nil listener")
	}
	if l.config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestNewSMTPListener(t *testing.T) {
	l := NewSMTPListener(&Listener{ID: "smtp-1"}, nil, nil, nil)
	if l == nil {
		t.Fatal("expected non-nil listener")
	}
	if l.config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestNewLDAPListener(t *testing.T) {
	l := NewLDAPListener(&Listener{ID: "ldap-1"}, nil, nil, nil)
	if l == nil {
		t.Fatal("expected non-nil listener")
	}
	if l.config == nil {
		t.Fatal("expected non-nil config")
	}
}

func TestNetPipeSMTPHandleConnection(t *testing.T) {
	// Test SMTP handleConnection using net.Pipe
	l := &SMTPListener{
		listener: &Listener{ID: "smtp-pipe", Token: "test-token"},
		config:   &ListenerConfig{Timeout: 30 * time.Second, BufferSize: 4096},
		store:    newListenerHandlerMockStore(),
		logger:   logrus.New(),
	}

	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	// Read server responses in background to prevent pipe deadlock
	go func() {
		io.Copy(io.Discard, client)
	}()

	// Start handleConnection in background BEFORE writing data
	// net.Pipe will block writes until the other end reads
	done := make(chan struct{})
	go func() {
		l.handleConnection(context.Background(), server)
		close(done)
	}()

	// Give handleConnection time to send greeting
	time.Sleep(50 * time.Millisecond)

	// Write SMTP commands - these will be read by handleConnection
	client.Write([]byte("HELO test\r\n"))
	time.Sleep(5 * time.Millisecond)
	client.Write([]byte("MAIL FROM:<sender@test.com>\r\n"))
	time.Sleep(5 * time.Millisecond)
	client.Write([]byte("RCPT TO:<rcpt@test.com>\r\n"))
	time.Sleep(5 * time.Millisecond)
	client.Write([]byte("QUIT\r\n"))

	select {
	case <-done:
		// OK
	case <-time.After(2 * time.Second):
		t.Fatal("handleConnection timed out")
	}
}

func TestSMTPReadData(t *testing.T) {
	server, client := net.Pipe()
	defer server.Close()
	defer client.Close()

	l := &SMTPListener{
		config: &ListenerConfig{Timeout: 5},
	}

	go func() {
		client.Write([]byte("Subject: Test\r\n"))
		client.Write([]byte("\r\n"))
		client.Write([]byte("Hello world\r\n"))
		client.Write([]byte(".\r\n"))
	}()

	reader := bufio.NewReader(server)
	var headers []string
	var body strings.Builder

	done := make(chan struct{})
	go func() {
		l.readData(reader, &headers, &body)
		close(done)
	}()

	select {
	case <-done:
		if len(headers) == 0 {
			t.Error("expected at least one header")
		}
		if body.String() == "" {
			t.Error("expected non-empty body")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("readData timed out")
	}
}
