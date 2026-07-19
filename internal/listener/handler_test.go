package listener

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// handlerMockStore implements Store for handler tests
type handlerMockStore struct {
	Store
	listeners    []Listener
	interactions []ListenerInteraction
	smtpMessages []SMTPMessage
	ldapQueries  []LDAPQuery
	smbRequests  []SMBRequest
	ftpCommands  []FTPCommand
}

func newListenerHandlerMockStore() *handlerMockStore {
	return &handlerMockStore{}
}

func (m *handlerMockStore) CreateListener(_ context.Context, l *Listener) error {
	m.listeners = append(m.listeners, *l)
	return nil
}

func (m *handlerMockStore) GetListener(_ context.Context, id string) (*Listener, error) {
	for _, l := range m.listeners {
		if l.ID == id {
			return &l, nil
		}
	}
	return nil, errors.New("listener not found")
}

func (m *handlerMockStore) GetAllListeners(_ context.Context) ([]Listener, error) {
	result := make([]Listener, len(m.listeners))
	copy(result, m.listeners)
	return result, nil
}

func (m *handlerMockStore) UpdateListener(_ context.Context, l *Listener) error {
	for i := range m.listeners {
		if m.listeners[i].ID == l.ID {
			m.listeners[i] = *l
			return nil
		}
	}
	return errors.New("listener not found")
}

func (m *handlerMockStore) DeleteListener(_ context.Context, id string) error {
	for i, l := range m.listeners {
		if l.ID == id {
			m.listeners = append(m.listeners[:i], m.listeners[i+1:]...)
			return nil
		}
	}
	return errors.New("listener not found")
}

func (m *handlerMockStore) GetListenerInteractions(_ context.Context, listenerID string) ([]ListenerInteraction, error) {
	var result []ListenerInteraction
	for _, inter := range m.interactions {
		if inter.ListenerID == listenerID {
			result = append(result, inter)
		}
	}
	return result, nil
}

func (m *handlerMockStore) GetSMTPMessages(_ context.Context, listenerID string) ([]SMTPMessage, error) {
	var result []SMTPMessage
	for _, msg := range m.smtpMessages {
		if msg.ListenerID == listenerID {
			result = append(result, msg)
		}
	}
	return result, nil
}

func (m *handlerMockStore) GetLDAPQueries(_ context.Context, listenerID string) ([]LDAPQuery, error) {
	var result []LDAPQuery
	for _, q := range m.ldapQueries {
		if q.ListenerID == listenerID {
			result = append(result, q)
		}
	}
	return result, nil
}

func (m *handlerMockStore) GetSMBRequests(_ context.Context, listenerID string) ([]SMBRequest, error) {
	var result []SMBRequest
	for _, r := range m.smbRequests {
		if r.ListenerID == listenerID {
			result = append(result, r)
		}
	}
	return result, nil
}

func (m *handlerMockStore) GetFTPCommands(_ context.Context, listenerID string) ([]FTPCommand, error) {
	var result []FTPCommand
	for _, c := range m.ftpCommands {
		if c.ListenerID == listenerID {
			result = append(result, c)
		}
	}
	return result, nil
}

func (m *handlerMockStore) SaveSMTPMessage(_ context.Context, msg *SMTPMessage) error {
	m.smtpMessages = append(m.smtpMessages, *msg)
	return nil
}

func (m *handlerMockStore) SaveListenerInteraction(_ context.Context, interaction *ListenerInteraction) error {
	m.interactions = append(m.interactions, *interaction)
	return nil
}

func (m *handlerMockStore) CreateListenerInteraction(_ context.Context, interaction *ListenerInteraction) error {
	m.interactions = append(m.interactions, *interaction)
	return nil
}

func setupListenerHandler() (*Handler, *handlerMockStore) {
	gin.SetMode(gin.TestMode)
	store := newListenerHandlerMockStore()
	config := &ListenerConfig{
		MaxConnections: 100,
		Timeout:        30 * time.Second,
		BufferSize:     4096,
	}
	h := NewHandler(store, config)
	return h, store
}

func TestListenerHandler_NewHandler(t *testing.T) {
	h, store := setupListenerHandler()
	assert.NotNil(t, h)
	assert.Equal(t, store, h.store)
	assert.NotNil(t, h.config)
}

func TestListenerHandler_CreateListener(t *testing.T) {
	h, _ := setupListenerHandler()

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := CreateListenerRequest{
			Protocol:  ProtocolSMTP,
			Host:      "0.0.0.0",
			Port:      2525,
			Token:     "test-token",
			IsEnabled: true,
		}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPost, "/listeners", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateListener(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("bad request - invalid JSON", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPost, "/listeners", bytes.NewReader([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateListener(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid protocol", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := CreateListenerRequest{
			Protocol: "invalid-proto",
			Host:     "0.0.0.0",
			Port:     9999,
			Token:    "token",
		}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPost, "/listeners", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.CreateListener(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestListenerHandler_GetListener(t *testing.T) {
	h, store := setupListenerHandler()

	listener := &Listener{
		ID:       "listener-1",
		Protocol: ProtocolSMTP,
		Host:     "0.0.0.0",
		Port:     25,
		Token:    "token-abc",
	}
	require.NoError(t, store.CreateListener(context.Background(), listener))

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "listener-1"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/listeners/listener-1", nil)

		h.GetListener(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/listeners/nonexistent", nil)

		h.GetListener(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListenerHandler_ListListeners(t *testing.T) {
	h, store := setupListenerHandler()

	store.listeners = append(store.listeners,
		Listener{ID: "l1", Protocol: ProtocolSMTP, Host: "0.0.0.0", Port: 25},
		Listener{ID: "l2", Protocol: ProtocolLDAP, Host: "0.0.0.0", Port: 389},
		Listener{ID: "l3", Protocol: ProtocolSMB, Host: "0.0.0.0", Port: 445},
	)

	t.Run("all listeners", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/listeners", nil)

		h.ListListeners(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, float64(0), resp["code"])
		data := resp["data"].([]interface{})
		assert.Len(t, data, 3)
	})

	t.Run("filter by protocol", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/listeners?protocol=smtp", nil)

		h.ListListeners(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		assert.Len(t, data, 1)
	})

	t.Run("filter no match", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodGet, "/listeners?protocol=ftp", nil)

		h.ListListeners(c)
		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		data := resp["data"].([]interface{})
		assert.Len(t, data, 0)
	})
}

func TestListenerHandler_UpdateListener(t *testing.T) {
	h, store := setupListenerHandler()

	listener := &Listener{ID: "l-update", Protocol: ProtocolSMTP}
	require.NoError(t, store.CreateListener(context.Background(), listener))

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "l-update"}}

		enabled := true
		req := UpdateListenerRequest{IsEnabled: &enabled}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPut, "/listeners/l-update", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateListener(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("bad request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "l-update"}}
		c.Request = httptest.NewRequest(http.MethodPut, "/listeners/l-update", bytes.NewReader([]byte("{invalid")))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateListener(c)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}

		enabled := true
		req := UpdateListenerRequest{IsEnabled: &enabled}
		body, _ := json.Marshal(req)
		c.Request = httptest.NewRequest(http.MethodPut, "/listeners/nonexistent", bytes.NewReader(body))
		c.Request.Header.Set("Content-Type", "application/json")

		h.UpdateListener(c)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestListenerHandler_DeleteListener(t *testing.T) {
	h, store := setupListenerHandler()

	listener := &Listener{ID: "l-delete", Protocol: ProtocolSMTP}
	require.NoError(t, store.CreateListener(context.Background(), listener))

	t.Run("success", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "l-delete"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/listeners/l-delete", nil)

		h.DeleteListener(c)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = []gin.Param{{Key: "id", Value: "nonexistent"}}
		c.Request = httptest.NewRequest(http.MethodDelete, "/listeners/nonexistent", nil)

		h.DeleteListener(c)
		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestListenerHandler_ListInteractions(t *testing.T) {
	h, store := setupListenerHandler()

	store.interactions = append(store.interactions,
		ListenerInteraction{ID: "inter-1", ListenerID: "l1", Protocol: ProtocolSMTP, SourceIP: "10.0.0.1"},
		ListenerInteraction{ID: "inter-2", ListenerID: "l1", Protocol: ProtocolSMTP, SourceIP: "10.0.0.2"},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/listeners/l1/interactions", nil)

	h.ListInteractions(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].([]interface{})
	assert.Len(t, data, 2)
}

func TestListenerHandler_ListSMTPMessages(t *testing.T) {
	h, store := setupListenerHandler()

	store.smtpMessages = append(store.smtpMessages,
		SMTPMessage{ID: "smtp-1", ListenerID: "l1", From: "a@b.com"},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/listeners/l1/smtp", nil)

	h.ListSMTPMessages(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestListenerHandler_ListLDAPQueries(t *testing.T) {
	h, store := setupListenerHandler()

	store.ldapQueries = append(store.ldapQueries,
		LDAPQuery{ID: "ldap-1", ListenerID: "l1", BaseDN: "dc=test"},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/listeners/l1/ldap", nil)

	h.ListLDAPQueries(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestListenerHandler_ListSMBRequests(t *testing.T) {
	h, store := setupListenerHandler()

	store.smbRequests = append(store.smbRequests,
		SMBRequest{ID: "smb-1", ListenerID: "l1", Command: "NEGOTIATE"},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/listeners/l1/smb", nil)

	h.ListSMBRequests(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestListenerHandler_ListFTPCommands(t *testing.T) {
	h, store := setupListenerHandler()

	store.ftpCommands = append(store.ftpCommands,
		FTPCommand{ID: "ftp-1", ListenerID: "l1", Command: "USER"},
	)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "l1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/listeners/l1/ftp", nil)

	h.ListFTPCommands(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
}

func TestListenerHandler_GetConfig(t *testing.T) {
	h, _ := setupListenerHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/listeners/config", nil)

	h.GetConfig(c)
	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, float64(0), resp["code"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, float64(100), data["max_connections"])
}

func TestIsValidProtocol(t *testing.T) {
	assert.True(t, isValidProtocol(ProtocolSMTP))
	assert.True(t, isValidProtocol(ProtocolLDAP))
	assert.True(t, isValidProtocol(ProtocolSMB))
	assert.True(t, isValidProtocol(ProtocolFTP))
	assert.True(t, isValidProtocol(ProtocolRMI))
	assert.False(t, isValidProtocol("invalid"))
	assert.False(t, isValidProtocol(""))
	assert.False(t, isValidProtocol("http"))
}
