package listener

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// protocolMockStore extends handlerMockStore with FTP/RMI/LDAP/SMB support
type protocolMockStore struct {
	*handlerMockStore
	ftpCommands     []FTPCommand
	rmiInteractions []RMIInteraction
	ldapQueries     []LDAPQuery
	smbRequests     []SMBRequest
	smtpMessages    []SMTPMessage
}

func newProtocolMockStore() *protocolMockStore {
	return &protocolMockStore{
		handlerMockStore: newListenerHandlerMockStore(),
	}
}

func (m *protocolMockStore) CreateFTPCommand(_ context.Context, cmd *FTPCommand) error {
	m.ftpCommands = append(m.ftpCommands, *cmd)
	return nil
}

func (m *protocolMockStore) SaveRMIInteraction(interaction *RMIInteraction) error {
	m.rmiInteractions = append(m.rmiInteractions, *interaction)
	return nil
}

func (m *protocolMockStore) CreateLDAPQuery(_ context.Context, q *LDAPQuery) error {
	m.ldapQueries = append(m.ldapQueries, *q)
	return nil
}

func (m *protocolMockStore) SaveLDAPQuery(_ context.Context, q *LDAPQuery) error {
	m.ldapQueries = append(m.ldapQueries, *q)
	return nil
}

func (m *protocolMockStore) CreateSMTPMessage(_ context.Context, msg *SMTPMessage) error {
	m.smtpMessages = append(m.smtpMessages, *msg)
	return nil
}

func (m *protocolMockStore) CreateSMBRequest(_ context.Context, req *SMBRequest) error {
	m.smbRequests = append(m.smbRequests, *req)
	return nil
}

// TestFTPHandleConnection verifies FTP protocol handling via real TCP connection.
func TestFTPHandleConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	addr := ln.Addr().(*net.TCPAddr)

	l := &FTPListener{
		listener: &Listener{ID: "ftp-pipe", Token: "ftp-token", Host: "127.0.0.1", Port: addr.Port},
		config:   &ListenerConfig{Timeout: 5 * time.Second, BufferSize: 4096},
		store:    newProtocolMockStore(),
		stopChan: make(chan struct{}),
	}

	// Replace server with our listener
	l.server = ln
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go l.acceptConnections(ctx)

	// Connect to FTP server
	client, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	defer client.Close()

	// Read greeting
	buf := make([]byte, 1024)
	n, _ := client.Read(buf)
	assert.Contains(t, string(buf[:n]), "220")

	// Send FTP commands
	client.Write([]byte("USER testuser\r\n"))
	n, _ = client.Read(buf)
	assert.Contains(t, string(buf[:n]), "331")

	client.Write([]byte("PASS testpass\r\n"))
	n, _ = client.Read(buf)
	assert.Contains(t, string(buf[:n]), "230")

	client.Write([]byte("SYST\r\n"))
	n, _ = client.Read(buf)
	assert.Contains(t, string(buf[:n]), "215")

	client.Write([]byte("QUIT\r\n"))
	n, _ = client.Read(buf)
	assert.Contains(t, string(buf[:n]), "221")

	time.Sleep(50 * time.Millisecond)

	mockStore, ok := l.store.(*protocolMockStore)
	assert.True(t, ok)
	assert.NotEmpty(t, mockStore.ftpCommands)
	assert.NotEmpty(t, mockStore.interactions)

	// Verify first command was USER
	assert.Equal(t, "USER", mockStore.ftpCommands[0].Command)
	assert.Equal(t, "testuser", mockStore.ftpCommands[0].Argument)

	l.Stop()
}

// TestFTPStartStop verifies FTP listener start and stop lifecycle.
func TestFTPStartStop(t *testing.T) {
	l := &FTPListener{
		listener: &Listener{ID: "ftp-lifecycle", Token: "ftp-token", Host: "127.0.0.1", Port: 0},
		config:   DefaultFTPConfig(),
		store:    newProtocolMockStore(),
		stopChan: make(chan struct{}),
	}

	// Use port 0 to get a random available port
	// We need to find an actual available port
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().(*net.TCPAddr)
	ln.Close()
	l.listener.Port = addr.Port

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Start(ctx)
	assert.NoError(t, err)

	// Connect and disconnect
	conn, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	conn.Close()

	time.Sleep(50 * time.Millisecond)

	err = l.Stop()
	assert.NoError(t, err)
}

// TestRMIHandleConnection verifies RMI protocol handling via real TCP connection.
func TestRMIHandleConnection(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	addr := ln.Addr().(*net.TCPAddr)

	l := &RMIListener{
		listener: &Listener{ID: "rmi-pipe", Token: "rmi-token", Host: "127.0.0.1", Port: addr.Port},
		config:   DefaultRMIConfig(),
		store:    newProtocolMockStore(),
		logger:   logger,
	}
	l.server = ln
	l.ctx, l.cancel = context.WithCancel(context.Background())
	l.wg.Add(1)
	go l.acceptLoop()

	// Connect and send JRMP handshake
	client, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	defer client.Close()

	handshake := []byte{
		0x4a, 0x52, 0x4d, 0x49, // "JRMI" magic
		0x00, 0x02, // version
		0x4b,       // StreamProtocol
		0x08, 0x00, // callback port
	}
	client.Write(handshake)

	time.Sleep(100 * time.Millisecond)

	l.Stop()

	mockStore, ok := l.store.(*protocolMockStore)
	assert.True(t, ok)
	assert.NotEmpty(t, mockStore.rmiInteractions)
	assert.Contains(t, mockStore.rmiInteractions[0].URN, "StreamProtocol")
}

// TestSMBHandleConnection verifies SMB protocol handling via real TCP connection.
func TestSMBHandleConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	addr := ln.Addr().(*net.TCPAddr)

	l := &SMBListener{
		listener: &Listener{ID: "smb-pipe", Token: "smb-token", Host: "127.0.0.1", Port: addr.Port},
		config:   DefaultSMBConfig(),
		store:    newProtocolMockStore(),
		stopChan: make(chan struct{}),
	}

	// Replace server with our listener
	l.server = ln
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go l.acceptConnections(ctx)

	client, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	defer client.Close()

	time.Sleep(50 * time.Millisecond)

	// Send SMB negotiate request
	smbData := make([]byte, 32)
	smbData[0] = 0xFF
	smbData[1] = 'S'
	smbData[2] = 'M'
	smbData[3] = 'B'
	smbData[4] = 0x00 // NEGOTIATE
	client.Write(smbData)

	time.Sleep(100 * time.Millisecond)

	l.Stop()
}

// TestSMBStartStop verifies SMB listener start and stop lifecycle.
func TestSMBStartStop(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().(*net.TCPAddr)
	ln.Close()

	l := &SMBListener{
		listener: &Listener{ID: "smb-lifecycle", Token: "smb-token", Host: "127.0.0.1", Port: addr.Port},
		config:   DefaultSMBConfig(),
		store:    newProtocolMockStore(),
		stopChan: make(chan struct{}),
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Start(ctx)
	assert.NoError(t, err)

	conn, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	conn.Close()

	time.Sleep(50 * time.Millisecond)

	err = l.Stop()
	assert.NoError(t, err)
}

// TestRMIStartStop verifies RMI listener start and stop lifecycle.
func TestRMIStartStop(t *testing.T) {
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	addr := ln.Addr().(*net.TCPAddr)
	ln.Close()

	logger := logrus.New()
	logger.SetLevel(logrus.WarnLevel)
	l := &RMIListener{
		listener: &Listener{ID: "rmi-lifecycle", Token: "rmi-token", Host: "127.0.0.1", Port: addr.Port},
		config:   DefaultRMIConfig(),
		store:    newProtocolMockStore(),
		logger:   logger,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := l.Start(ctx)
	assert.NoError(t, err)

	conn, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	conn.Close()

	time.Sleep(50 * time.Millisecond)

	err = l.Stop()
	assert.NoError(t, err)
}

// TestLDAPHandleConnection verifies LDAP protocol handling via real TCP connection.
func TestLDAPHandleConnection(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	assert.NoError(t, err)
	addr := ln.Addr().(*net.TCPAddr)

	l := &LDAPListener{
		listener: &Listener{ID: "ldap-pipe", Token: "ldap-token", Host: "127.0.0.1", Port: addr.Port},
		config:   DefaultLDAPConfig(),
		store:    newProtocolMockStore(),
		logger:   logrus.New(),
	}

	l.server = ln
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go l.acceptConnections(ctx)

	client, err := net.Dial("tcp", addr.String())
	assert.NoError(t, err)
	defer client.Close()

	time.Sleep(50 * time.Millisecond)

	// Send a minimal LDAP bind request
	ldapBind := []byte{
		0x30, 0x0C, // Sequence of length 12
		0x02, 0x01, 0x01, // Integer: message ID = 1
		0x63, 0x07, // BindRequest tag
		0x04, 0x05, 'a', 'd', 'm', 'i', 'n', // string "admin"
	}
	client.Write(ldapBind)

	time.Sleep(100 * time.Millisecond)

	l.Stop()
}
