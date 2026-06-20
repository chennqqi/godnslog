package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionStore_CreateAndGet(t *testing.T) {
	store := NewSessionStore(5 * time.Minute)
	session := store.Create()
	if session.ID == "" {
		t.Fatal("expected non-empty session ID")
	}

	got := store.Get(session.ID)
	if got == nil {
		t.Fatal("expected to find session by ID")
	}
	if got.ID != session.ID {
		t.Fatalf("expected ID %s, got %s", session.ID, got.ID)
	}
}

func TestSessionStore_GetExpired(t *testing.T) {
	store := NewSessionStore(1 * time.Millisecond)
	session := store.Create()
	time.Sleep(5 * time.Millisecond)

	got := store.Get(session.ID)
	if got != nil {
		t.Fatal("expected nil for expired session")
	}
}

func TestSessionStore_Delete(t *testing.T) {
	store := NewSessionStore(5 * time.Minute)
	session := store.Create()
	store.Delete(session.ID)

	got := store.Get(session.ID)
	if got != nil {
		t.Fatal("expected nil after delete")
	}
}

func TestSessionStore_Cleanup(t *testing.T) {
	store := NewSessionStore(1 * time.Millisecond)
	store.Create()
	store.Create()
	time.Sleep(5 * time.Millisecond)

	store.Cleanup()

	store.mu.RLock()
	count := len(store.sessions)
	store.mu.RUnlock()
	if count != 0 {
		t.Fatalf("expected 0 sessions after cleanup, got %d", count)
	}
}

func TestSession_SetGet(t *testing.T) {
	s := &Session{
		ID:        "test",
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		data:      make(map[string]interface{}),
	}
	s.Set("key", "value")
	v, ok := s.Get("key")
	if !ok {
		t.Fatal("expected to find key")
	}
	if v != "value" {
		t.Fatalf("expected 'value', got %v", v)
	}
}

func TestMCPHandler_Initialize(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	toolMap := map[string]Tool{}
	tools := []Tool{}
	handler := NewMCPHandler(server, toolMap, tools)

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params:  json.RawMessage(`{"protocolVersion":"2024-11-05"}`),
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected result to be a map")
	}
	if result["protocolVersion"] != "2024-11-05" {
		t.Fatalf("expected protocolVersion '2024-11-05', got %v", result["protocolVersion"])
	}
	if result["sessionId"] == nil || result["sessionId"] == "" {
		t.Fatal("expected non-empty sessionId")
	}
}

func TestMCPHandler_ToolsList(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	tools := []Tool{
		{Name: "test_tool", Description: "A test tool"},
	}
	toolMap := map[string]Tool{
		"test_tool": tools[0],
	}
	handler := NewMCPHandler(server, toolMap, tools)

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	result, ok := resp.Result.(map[string]interface{})
	if !ok {
		t.Fatal("expected result to be a map")
	}
	toolsList, ok := result["tools"].([]interface{})
	if !ok {
		t.Fatal("expected tools to be a list")
	}
	if len(toolsList) != 1 {
		t.Fatalf("expected 1 tool, got %d", len(toolsList))
	}
}

func TestMCPHandler_MethodNotFound(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	handler := NewMCPHandler(server, map[string]Tool{}, []Tool{})

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      3,
		Method:  "nonexistent/method",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	var resp JSONRPCResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != ErrCodeMethodNotFound {
		t.Fatalf("expected error code %d, got %d", ErrCodeMethodNotFound, resp.Error.Code)
	}
}

func TestMCPHandler_NotificationsInitialized(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	handler := NewMCPHandler(server, map[string]Tool{}, []Tool{})

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      nil,
		Method:  "notifications/initialized",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusAccepted {
		t.Fatalf("expected 202, got %d", w.Code)
	}
}

func TestMCPHandler_Ping(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	handler := NewMCPHandler(server, map[string]Tool{}, []Tool{})

	reqBody := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      4,
		Method:  "ping",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader(body))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestMCPHandler_MethodNotAllowed(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	handler := NewMCPHandler(server, map[string]Tool{}, []Tool{})

	req := httptest.NewRequest(http.MethodGet, "/mcp", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestMCPHandler_InvalidJSON(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	handler := NewMCPHandler(server, map[string]Tool{}, []Tool{})

	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestMCPHandler_StartCleanup(t *testing.T) {
	server := &Server{apiURL: "http://localhost", apiKey: "test"}
	handler := NewMCPHandler(server, map[string]Tool{}, []Tool{})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	handler.StartCleanup(ctx)

	// Just verify it doesn't panic
	time.Sleep(10 * time.Millisecond)
}
