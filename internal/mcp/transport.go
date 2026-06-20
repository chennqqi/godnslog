package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
)

// Session represents an MCP session with state tracking
type Session struct {
	ID        string
	CreatedAt time.Time
	LastSeen  time.Time
	mu        sync.Mutex
	data      map[string]interface{}
}

// Touch updates the last seen time of the session
func (s *Session) Touch() {
	s.mu.Lock()
	s.LastSeen = time.Now()
	s.mu.Unlock()
}

// Set stores a value in the session
func (s *Session) Set(key string, value interface{}) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

// Get retrieves a value from the session
func (s *Session) Get(key string) (interface{}, bool) {
	s.mu.Lock()
	v, ok := s.data[key]
	s.mu.Unlock()
	return v, ok
}

// SessionStore manages MCP sessions with timeout cleanup
type SessionStore struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	timeout  time.Duration
}

// NewSessionStore creates a new session store with the given idle timeout
func NewSessionStore(timeout time.Duration) *SessionStore {
	return &SessionStore{
		sessions: make(map[string]*Session),
		timeout:  timeout,
	}
}

// Create creates a new session and returns it
func (store *SessionStore) Create() *Session {
	session := &Session{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		LastSeen:  time.Now(),
		data:      make(map[string]interface{}),
	}
	store.mu.Lock()
	store.sessions[session.ID] = session
	store.mu.Unlock()
	return session
}

// Get retrieves a session by ID. Returns nil if not found or expired.
func (store *SessionStore) Get(id string) *Session {
	store.mu.RLock()
	session, ok := store.sessions[id]
	store.mu.RUnlock()
	if !ok {
		return nil
	}

	// Check if session has expired
	if time.Since(session.LastSeen) > store.timeout {
		store.Delete(id)
		return nil
	}

	session.Touch()
	return session
}

// Delete removes a session from the store
func (store *SessionStore) Delete(id string) {
	store.mu.Lock()
	delete(store.sessions, id)
	store.mu.Unlock()
}

// Cleanup removes all expired sessions
func (store *SessionStore) Cleanup() {
	store.mu.Lock()
	defer store.mu.Unlock()
	now := time.Now()
	for id, session := range store.sessions {
		if now.Sub(session.LastSeen) > store.timeout {
			delete(store.sessions, id)
		}
	}
}

// StartCleanupRoutine runs periodic cleanup of expired sessions
func (store *SessionStore) StartCleanupRoutine(ctx context.Context) {
	ticker := time.NewTicker(store.timeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			store.Cleanup()
		}
	}
}

// JSONRPCRequest represents a JSON-RPC 2.0 request
type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      interface{}     `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *RPCError   `json:"error,omitempty"`
}

// RPCError represents a JSON-RPC 2.0 error object
type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// Standard JSON-RPC error codes
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError  = -32603
)

// MCPHandler handles MCP protocol requests over Streamable HTTP transport
type MCPHandler struct {
	server   *Server
	sessions *SessionStore
	toolMap  map[string]Tool
	tools    []Tool
}

// NewMCPHandler creates a new MCP protocol handler
func NewMCPHandler(server *Server, toolMap map[string]Tool, tools []Tool) *MCPHandler {
	return &MCPHandler{
		server:   server,
		sessions: NewSessionStore(30 * time.Minute),
		toolMap:  toolMap,
		tools:    tools,
	}
}

// StartCleanup starts the session cleanup routine
func (h *MCPHandler) StartCleanup(ctx context.Context) {
	go h.sessions.StartCleanupRoutine(ctx)
}

// ServeHTTP implements the Streamable HTTP transport for MCP
func (h *MCPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON-RPC request
	var req JSONRPCRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeRPCError(w, nil, ErrCodeParseError, "Parse error")
		return
	}

	if req.JSONRPC != "2.0" {
		writeRPCError(w, req.ID, ErrCodeInvalidRequest, "Invalid Request: jsonrpc must be '2.0'")
		return
	}

	// Handle different MCP methods
	var result interface{}
	var rpcErr *RPCError

	switch req.Method {
	case "initialize":
		result, rpcErr = h.handleInitialize(r, &req)
	case "notifications/initialized":
		// This is a notification, no response needed but we return 202
		w.WriteHeader(http.StatusAccepted)
		return
	case "tools/list":
		result, rpcErr = h.handleToolsList(r)
	case "tools/call":
		result, rpcErr = h.handleToolsCall(r, &req)
	case "ping":
		result = map[string]interface{}{}
	default:
		rpcErr = &RPCError{
			Code:    ErrCodeMethodNotFound,
			Message: fmt.Sprintf("Method not found: %s", req.Method),
		}
	}

	if rpcErr != nil {
		writeRPCError(w, req.ID, rpcErr.Code, rpcErr.Message)
		return
	}

	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  result,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// handleInitialize handles the MCP initialize method
func (h *MCPHandler) handleInitialize(r *http.Request, req *JSONRPCRequest) (interface{}, *RPCError) {
	// Create new session
	session := h.sessions.Create()

	// Set session ID in response header
	// (the writer is not available here, so we return it in the result)

	var params struct {
		ProtocolVersion string                 `json:"protocolVersion"`
		Capabilities    map[string]interface{} `json:"capabilities"`
		ClientInfo      map[string]interface{} `json:"clientInfo"`
	}
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "Invalid params"}
		}
	}

	return map[string]interface{}{
		"protocolVersion": "2024-11-05",
		"capabilities": map[string]interface{}{
			"tools": map[string]interface{}{
				"listChanged": false,
			},
		},
		"serverInfo": map[string]interface{}{
			"name":    "godnslog-mcp-server",
			"version": "1.0.0",
		},
		"sessionId": session.ID,
	}, nil
}

// handleToolsList returns the list of available MCP tools
func (h *MCPHandler) handleToolsList(r *http.Request) (interface{}, *RPCError) {
	tools := make([]map[string]interface{}, 0, len(h.tools))
	for _, tool := range h.tools {
		tools = append(tools, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
			"inputSchema": map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		})
	}
	return map[string]interface{}{
		"tools": tools,
	}, nil
}

// handleToolsCall executes a tool by name with given arguments
func (h *MCPHandler) handleToolsCall(r *http.Request, req *JSONRPCRequest) (interface{}, *RPCError) {
	var params struct {
		Name      string                 `json:"name"`
		Arguments map[string]interface{} `json:"arguments"`
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return nil, &RPCError{Code: ErrCodeInvalidParams, Message: "Invalid params"}
	}

	tool, ok := h.toolMap[params.Name]
	if !ok {
		return nil, &RPCError{
			Code:    ErrCodeMethodNotFound,
			Message: fmt.Sprintf("Tool not found: %s", params.Name),
		}
	}

	result, err := tool.Execute(r.Context(), params.Arguments)
	if err != nil {
		return nil, &RPCError{
			Code:    ErrCodeInternalError,
			Message: err.Error(),
		}
	}

	// Wrap result as MCP content
	return map[string]interface{}{
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": toJSONString(result),
			},
		},
	}, nil
}

// writeRPCError writes a JSON-RPC error response
func writeRPCError(w http.ResponseWriter, id interface{}, code int, message string) {
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &RPCError{
			Code:    code,
			Message: message,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(resp)
}

// toJSONString converts any value to a JSON string
func toJSONString(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprintf("%v", v)
	}
	return string(b)
}
