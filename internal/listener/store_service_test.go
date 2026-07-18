package listener

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/xorm"
)

func setupListenerTestDB(t *testing.T) *xorm.Engine {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "listener_test.db")
	engine, err := xorm.NewEngine("sqlite", dbPath)
	require.NoError(t, err)
	require.NoError(t, engine.Sync2(
		new(models.Listener),
		new(models.ListenerInteraction),
		new(models.SMTPMessage),
		new(models.LDAPQuery),
		new(models.SMBRequest),
		new(models.FTPCommand),
		new(models.RMIInteraction),
	))
	return engine
}

func TestXormStore_CreateAndGetListener(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	listener := &Listener{
		ID:        "listener-1",
		Protocol:  "smtp",
		Host:      "0.0.0.0",
		Port:      2525,
		Token:     "test-token",
		IsEnabled: true,
	}
	err := store.CreateListener(ctx, listener)
	require.NoError(t, err)

	fetched, err := store.GetListener(ctx, "listener-1")
	require.NoError(t, err)
	assert.Equal(t, "smtp", string(fetched.Protocol))
	assert.Equal(t, 2525, fetched.Port)
	assert.True(t, fetched.IsEnabled)
}

func TestXormStore_GetListenerByToken(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	require.NoError(t, store.CreateListener(ctx, &Listener{
		ID: "l1", Protocol: "dns", Token: "token-abc",
	}))
	require.NoError(t, store.CreateListener(ctx, &Listener{
		ID: "l2", Protocol: "http", Token: "token-xyz",
	}))

	fetched, err := store.GetListenerByToken(ctx, "token-xyz")
	require.NoError(t, err)
	assert.Equal(t, "http", string(fetched.Protocol))
}

func TestXormStore_GetAllListeners(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		require.NoError(t, store.CreateListener(ctx, &Listener{
			ID: string(rune('a'+i)), Protocol: "dns",
		}))
	}

	all, err := store.GetAllListeners(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 3)
}

func TestXormStore_UpdateAndDeleteListener(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	require.NoError(t, store.CreateListener(ctx, &Listener{
		ID: "l1", Protocol: "dns", Port: 53, IsEnabled: false,
	}))

	require.NoError(t, store.UpdateListener(ctx, &Listener{
		ID: "l1", Port: 5353,
	}))

	fetched, _ := store.GetListener(ctx, "l1")
	assert.Equal(t, 5353, fetched.Port)

	require.NoError(t, store.DeleteListener(ctx, "l1"))
	deleted, _ := store.GetListener(ctx, "l1")
	assert.Empty(t, deleted.ID)
}

func TestXormStore_ListenerInteraction(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	interaction := &ListenerInteraction{
		ID: "inter-1", ListenerID: "l1", Protocol: "smtp",
		SourceIP: "10.0.0.1", SourcePort: 12345, Data: "HELO test",
	}
	err := store.SaveListenerInteraction(ctx, interaction)
	require.NoError(t, err)

	interactions, err := store.GetListenerInteractions(ctx, "l1")
	require.NoError(t, err)
	assert.Len(t, interactions, 1)
	assert.Equal(t, "10.0.0.1", interactions[0].SourceIP)

	// Cleanup
	require.NoError(t, store.DeleteListenerInteraction(ctx, "inter-1"))
	interactions, _ = store.GetListenerInteractions(ctx, "l1")
	assert.Empty(t, interactions)
}

func TestXormStore_SMTPMessage(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	msg := &SMTPMessage{
		ID: "smtp-1", ListenerID: "l1",
		From: "sender@test.com", To: "rcpt@test.com",
		Subject: "Test", Body: "Hello",
	}
	err := store.SaveSMTPMessage(ctx, msg)
	require.NoError(t, err)

	fetched, err := store.GetSMTPMessage(ctx, "smtp-1")
	require.NoError(t, err)
	assert.Equal(t, "sender@test.com", fetched.From)
	assert.Equal(t, "Test", fetched.Subject)

	messages, err := store.GetSMTPMessages(ctx, "l1")
	require.NoError(t, err)
	assert.Len(t, messages, 1)

	require.NoError(t, store.DeleteSMTPMessage(ctx, "smtp-1"))
}

func TestXormStore_LDAPQuery(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	query := &LDAPQuery{
		ID: "ldap-1", ListenerID: "l1",
		BaseDN: "dc=example,dc=com", Filter: "(objectClass=*)",
	}
	err := store.SaveLDAPQuery(ctx, query)
	require.NoError(t, err)

	fetched, err := store.GetLDAPQuery(ctx, "ldap-1")
	require.NoError(t, err)
	assert.Equal(t, "dc=example,dc=com", fetched.BaseDN)

	queries, err := store.GetLDAPQueries(ctx, "l1")
	require.NoError(t, err)
	assert.Len(t, queries, 1)

	require.NoError(t, store.DeleteLDAPQuery(ctx, "ldap-1"))
}

func TestXormStore_SMBAndFTP(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)
	ctx := context.Background()

	// SMB
	smb := &SMBRequest{
		ID: "smb-1", ListenerID: "l1",
		Command: "TreeConnect", ShareName: "IPC$",
	}
	err := store.CreateSMBRequest(ctx, smb)
	require.NoError(t, err)

	fetched, err := store.GetSMBRequest(ctx, "smb-1")
	require.NoError(t, err)
	assert.Equal(t, "TreeConnect", fetched.Command)

	requests, err := store.GetSMBRequests(ctx, "l1")
	require.NoError(t, err)
	assert.Len(t, requests, 1)

	require.NoError(t, store.DeleteSMBRequest(ctx, "smb-1"))

	// FTP
	ftp := &FTPCommand{
		ID: "ftp-1", ListenerID: "l1",
		Command: "USER", Argument: "anonymous",
	}
	err = store.CreateFTPCommand(ctx, ftp)
	require.NoError(t, err)

	fetchedFTP, err := store.GetFTPCommand(ctx, "ftp-1")
	require.NoError(t, err)
	assert.Equal(t, "USER", fetchedFTP.Command)

	cmds, err := store.GetFTPCommands(ctx, "l1")
	require.NoError(t, err)
	assert.Len(t, cmds, 1)

	require.NoError(t, store.DeleteFTPCommand(ctx, "ftp-1"))
}

func TestXormStore_RMI(t *testing.T) {
	engine := setupListenerTestDB(t)
	store := NewXormStore(engine)

	rmi := &RMIInteraction{
		ID: "rmi-1", ListenerID: "l1",
		SourceIP: "10.0.0.1", URN: "rmi://test",
	}
	err := store.SaveRMIInteraction(rmi)
	require.NoError(t, err)

	// Verify via engine
	var saved models.RMIInteraction
	has, err := engine.ID("rmi-1").Get(&saved)
	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, "rmi://test", saved.URN)
}

func TestService_CreateAndGetListener(t *testing.T) {
	engine := setupListenerTestDB(t)
	svc := NewService(engine)

	listener := &models.Listener{
		Protocol: "smtp", Host: "0.0.0.0", Port: 2525, IsEnabled: true,
	}
	err := svc.CreateListener(listener)
	require.NoError(t, err)
	assert.NotEmpty(t, listener.ID)

	fetched, err := svc.GetListener(listener.ID)
	require.NoError(t, err)
	assert.Equal(t, "smtp", string(fetched.Protocol))
	assert.Equal(t, 2525, fetched.Port)
}

func TestService_GetListener_NotFound(t *testing.T) {
	engine := setupListenerTestDB(t)
	svc := NewService(engine)

	_, err := svc.GetListener("nonexistent")
	assert.ErrorIs(t, err, ErrListenerNotFound)
}

func TestService_ListListeners(t *testing.T) {
	engine := setupListenerTestDB(t)
	svc := NewService(engine)

	for i := 0; i < 3; i++ {
		require.NoError(t, svc.CreateListener(&models.Listener{
			Protocol: "dns", Host: "0.0.0.0", Port: 53,
		}))
	}

	listeners, total, err := svc.ListListeners(1, 10)
	require.NoError(t, err)
	assert.Len(t, listeners, 3)
	assert.Equal(t, int64(3), total)
}

func TestService_UpdateListener(t *testing.T) {
	engine := setupListenerTestDB(t)
	svc := NewService(engine)

	l := &models.Listener{Protocol: "dns", Port: 53}
	require.NoError(t, svc.CreateListener(l))

	l.Port = 5353
	err := svc.UpdateListener(l)
	require.NoError(t, err)

	fetched, _ := svc.GetListener(l.ID)
	assert.Equal(t, 5353, fetched.Port)
}

func TestService_DeleteListener(t *testing.T) {
	engine := setupListenerTestDB(t)
	svc := NewService(engine)

	l := &models.Listener{Protocol: "dns", Port: 53}
	require.NoError(t, svc.CreateListener(l))

	require.NoError(t, svc.DeleteListener(l.ID))

	_, err := svc.GetListener(l.ID)
	assert.ErrorIs(t, err, ErrListenerNotFound)
}

func TestService_ListListenerInteractions(t *testing.T) {
	engine := setupListenerTestDB(t)
	svc := NewService(engine)

	// Direct insert via engine
	inter := &models.ListenerInteraction{
		ID: "inter-1", ListenerID: "l1", Protocol: "smtp",
		SourceIP: "10.0.0.1", Data: "test", Timestamp: time.Now(),
	}
	_, err := engine.Insert(inter)
	require.NoError(t, err)

	interactions, err := svc.ListListenerInteractions("l1")
	require.NoError(t, err)
	assert.Len(t, interactions, 1)
	assert.Equal(t, "10.0.0.1", interactions[0].SourceIP)
}
