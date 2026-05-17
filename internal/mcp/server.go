package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Server implements the MCP server for GODNSLOG
type Server struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewServer creates a new MCP server
func NewServer(apiURL, apiKey string) *Server {
	return &Server{
		apiURL:     apiURL,
		apiKey:     apiKey,
		httpClient: &http.Client{},
	}
}

// Run starts the MCP server
func (s *Server) Run(ctx context.Context) error {
	// Register tools
	tools := []Tool{
		{Name: "create_oast_probe", Description: "Create an agent-native OAST probe with a case and payload", Execute: s.createOASTProbe},
		{Name: "create_case", Description: "Create a new case", Execute: s.createCase},
		{Name: "create_payload", Description: "Create a new payload", Execute: s.createPayload},
		{Name: "list_interactions", Description: "List interactions", Execute: s.listInteractions},
		{Name: "wait_for_interaction", Description: "Wait for interaction", Execute: s.waitForInteraction},
		{Name: "summarize_evidence", Description: "Summarize evidence", Execute: s.summarizeEvidence},
		{Name: "export_report", Description: "Export report", Execute: s.exportReport},
		{Name: "revoke_token", Description: "Revoke API token", Execute: s.revokeToken},
	}

	// Start stdio transport (simplified for MVP)
	log.Printf("MCP Server listening on stdio with %d tools", len(tools))

	// Convert to map for lookup
	toolMap := make(map[string]Tool)
	for _, tool := range tools {
		toolMap[tool.Name] = tool
	}

	// In production, use proper MCP transport (stdio or SSE)
	// For MVP, we'll implement a simple HTTP server
	return s.runHTTPServer(ctx, toolMap, tools)
}

// Tool represents an MCP tool
type Tool struct {
	Name        string
	Description string
	Execute     func(ctx context.Context, args map[string]interface{}) (interface{}, error)
}

// ToolResult represents the result of a tool execution
type ToolResult struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// createOASTProbe creates a case and payload in one agent-friendly operation.
func (s *Server) createOASTProbe(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	title, ok := args["title"].(string)
	if !ok || strings.TrimSpace(title) == "" {
		return nil, fmt.Errorf("title is required")
	}

	templateID, ok := args["template_id"].(string)
	if !ok || strings.TrimSpace(templateID) == "" {
		return nil, fmt.Errorf("template_id is required")
	}

	description, _ := args["description"].(string)
	target, _ := args["target"].(string)
	agentID, _ := args["agent_id"].(string)
	expiresIn, _ := args["expires_in"].(string)
	variables := normalizeStringMap(args["variables"])
	expectedProtocols := normalizeStringSlice(args["expected_protocols"])

	// Convert expires_in duration to expires_at timestamp (RFC3339 format)
	var expiresAt string
	if expiresIn != "" {
		duration, err := time.ParseDuration(expiresIn)
		if err == nil {
			expiresAt = time.Now().Add(duration).Format(time.RFC3339)
		}
	}

	// Convert expected_protocols array to single string (take first if exists)
	var expectedProtocol string
	if len(expectedProtocols) > 0 {
		expectedProtocol = expectedProtocols[0]
	}

	caseResult, err := s.apiCall("POST", "/api/v2/cases", map[string]interface{}{
		"title":       title,
		"description": description,
		"target":      target,
		"tags":        []string{"agent", "oast-probe"},
	})
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	caseID := extractString(caseResult, "id")
	if caseID == "" {
		caseID = extractString(caseResult, "case_id")
	}
	if caseID == "" {
		return ToolResult{Success: false, Error: "case creation response did not include an id"}, nil
	}

	payloadResult, err := s.apiCall("POST", "/api/v2/payloads", map[string]interface{}{
		"template_id":       templateID,
		"case_id":           caseID,
		"variables":         variables,
		"expires_at":        expiresAt,
		"expected_protocol": expectedProtocol,
	})
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	payloadID := extractString(payloadResult, "id")
	if payloadID == "" {
		payloadID = extractString(payloadResult, "payload_id")
	}
	token := extractString(payloadResult, "token")

	// Generate agent_run_id if agent_id is provided
	agentRunID := ""
	if agentID != "" {
		agentRunID = agentID + ":" + caseID + ":" + payloadID
	}

	responseData := map[string]interface{}{
		"probe_id":           caseID + ":" + payloadID,
		"case_id":            caseID,
		"payload_id":         payloadID,
		"token":              token,
		"case":               caseResult,
		"payload":            payloadResult,
		"expected_protocols": expectedProtocols,
		"agent_next_action":  "Deliver the payload to the target, then call wait_for_interaction with the returned token.",
	}

	if agentRunID != "" {
		responseData["agent_run_id"] = agentRunID
	}

	return ToolResult{Success: true, Data: responseData}, nil
}

// create_case creates a new case
func (s *Server) createCase(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	title, ok := args["title"].(string)
	if !ok {
		return nil, fmt.Errorf("title is required")
	}

	description, _ := args["description"].(string)
	target, _ := args["target"].(string)
	tags, _ := args["tags"].([]string)

	// Call API
	result, err := s.apiCall("POST", "/api/v2/cases", map[string]interface{}{
		"title":       title,
		"description": description,
		"target":      target,
		"tags":        tags,
	})

	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// create_payload creates a new payload
func (s *Server) createPayload(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	templateID, ok := args["template_id"].(string)
	if !ok {
		return nil, fmt.Errorf("template_id is required")
	}

	caseID, _ := args["case_id"].(string)
	variables, _ := args["variables"].(map[string]string)
	expiresIn, _ := args["expires_in"].(string)

	// Convert expires_in duration to expires_at timestamp (RFC3339 format)
	var expiresAt string
	if expiresIn != "" {
		duration, err := time.ParseDuration(expiresIn)
		if err == nil {
			expiresAt = time.Now().Add(duration).Format(time.RFC3339)
		}
	}

	// Call API
	result, err := s.apiCall("POST", "/api/v2/payloads", map[string]interface{}{
		"template_id": templateID,
		"case_id":     caseID,
		"variables":   variables,
		"expires_at":  expiresAt,
	})

	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// list_interactions lists interactions
func (s *Server) listInteractions(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	caseID, _ := args["case_id"].(string)
	limit := 50
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	// Call API
	query := fmt.Sprintf("?page_size=%d", limit)
	if caseID != "" {
		query += fmt.Sprintf("&case_id=%s", caseID)
	}

	result, err := s.apiCall("GET", "/api/v2/interactions"+query, nil)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// wait_for_interaction waits for an interaction
func (s *Server) waitForInteraction(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	token, ok := args["token"].(string)
	if !ok {
		return nil, fmt.Errorf("token is required")
	}

	timeout := 300 // 5 minutes default
	if t, ok := args["timeout"].(float64); ok {
		timeout = int(t)
	}

	// Poll for interactions (simplified)
	// In production, use WebSocket or SSE
	result, err := s.pollInteractions(ctx, token, timeout)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// summarize_evidence summarizes evidence for a case
func (s *Server) summarizeEvidence(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	caseID, ok := args["case_id"].(string)
	if !ok {
		return nil, fmt.Errorf("case_id is required")
	}

	// Call API
	result, err := s.apiCall("GET", "/api/v2/evidence/"+caseID, nil)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// export_report exports a report
func (s *Server) exportReport(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	caseID, ok := args["case_id"].(string)
	if !ok {
		return nil, fmt.Errorf("case_id is required")
	}

	format := "markdown"
	if f, ok := args["format"].(string); ok {
		format = f
	}

	// Call API
	result, err := s.apiCall("POST", "/api/v2/evidence/generate", map[string]interface{}{
		"case_id": caseID,
		"format":  format,
	})

	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// revoke_token revokes an API key
func (s *Server) revokeToken(ctx context.Context, args map[string]interface{}) (interface{}, error) {
	keyID, ok := args["key_id"].(string)
	if !ok {
		return nil, fmt.Errorf("key_id is required")
	}

	// Call API
	result, err := s.apiCall("DELETE", "/api/v2/apikeys/"+keyID, nil)
	if err != nil {
		return ToolResult{Success: false, Error: err.Error()}, nil
	}

	return ToolResult{Success: true, Data: result}, nil
}

// apiCall makes an API call to GODNSLOG
func (s *Server) apiCall(method, path string, body interface{}) (interface{}, error) {
	url := s.apiURL + path

	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = new(bytes.Buffer)
		if err := json.NewEncoder(reqBody).Encode(body); err != nil {
			return nil, fmt.Errorf("failed to encode request body: %w", err)
		}
	}

	var req *http.Request
	var err error

	if method == "GET" && body == nil {
		req, err = http.NewRequest(method, url, nil)
	} else if body != nil {
		req, err = http.NewRequest(method, url, reqBody)
	} else {
		req, err = http.NewRequest(method, url, nil)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("API returned error %d: %s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result, nil
}

func normalizeStringMap(value interface{}) map[string]string {
	result := make(map[string]string)

	switch typed := value.(type) {
	case map[string]string:
		for k, v := range typed {
			result[k] = v
		}
	case map[string]interface{}:
		for k, v := range typed {
			result[k] = fmt.Sprint(v)
		}
	}

	return result
}

func normalizeStringSlice(value interface{}) []string {
	result := []string{}

	switch typed := value.(type) {
	case []string:
		result = append(result, typed...)
	case []interface{}:
		for _, item := range typed {
			result = append(result, fmt.Sprint(item))
		}
	}

	return result
}

func extractString(value interface{}, key string) string {
	if found := extractStringFromMap(value, key); found != "" {
		return found
	}

	if root, ok := value.(map[string]interface{}); ok {
		if data, ok := root["data"]; ok {
			return extractStringFromMap(data, key)
		}
	}

	return ""
}

func extractStringFromMap(value interface{}, key string) string {
	fields, ok := value.(map[string]interface{})
	if !ok {
		return ""
	}

	if raw, ok := fields[key]; ok {
		if text, ok := raw.(string); ok {
			return text
		}
	}

	return ""
}

// pollInteractions polls for interactions
func (s *Server) pollInteractions(ctx context.Context, token string, timeout int) (interface{}, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if result, found := s.findInteractions(token); found {
		return result, nil
	}

	deadline := time.After(time.Duration(timeout) * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled")
		case <-deadline:
			return map[string]interface{}{
				"message": "Timeout waiting for interaction",
				"token":   token,
			}, nil
		case <-ticker.C:
			if result, found := s.findInteractions(token); found {
				return result, nil
			}
		}
	}
}

func (s *Server) findInteractions(token string) (interface{}, bool) {
	query := fmt.Sprintf("?token=%s&page_size=10", token)
	result, err := s.apiCall("GET", "/api/v2/interactions"+query, nil)
	if err != nil {
		return nil, false
	}

	resp, ok := result.(map[string]interface{})
	if !ok {
		return nil, false
	}

	data, ok := resp["data"].(map[string]interface{})
	if !ok {
		return nil, false
	}

	items, ok := data["items"].([]interface{})
	if !ok || len(items) == 0 {
		return nil, false
	}

	return map[string]interface{}{
		"message":      "Interaction detected",
		"token":        token,
		"interactions": items,
	}, true
}

// runHTTPServer runs a simple HTTP server for MCP (MVP)
func (s *Server) runHTTPServer(ctx context.Context, toolMap map[string]Tool, tools []Tool) error {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// MCP endpoint
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"name":    "godnslog-mcp-server",
			"version": "1.0.0",
			"tools":   s.listTools(tools),
		})
	})

	mux.HandleFunc("/tool/", func(w http.ResponseWriter, r *http.Request) {
		// Tool execution endpoint
		toolName := r.URL.Path[len("/tool/"):]
		tool, ok := toolMap[toolName]
		if !ok {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}

		var args map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		result, err := tool.Execute(ctx, args)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(result)
	})

	server := &http.Server{Addr: ":8081"}
	go func() {
		<-ctx.Done()
		server.Shutdown(context.Background())
	}()

	log.Println("MCP HTTP server listening on :8081")
	return server.ListenAndServe()
}

// listTools returns a list of available tools
func (s *Server) listTools(tools []Tool) []map[string]interface{} {
	list := make([]map[string]interface{}, 0, len(tools))
	for _, tool := range tools {
		list = append(list, map[string]interface{}{
			"name":        tool.Name,
			"description": tool.Description,
		})
	}
	return list
}
