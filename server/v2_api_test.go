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
	}

	for _, route := range requiredRoutes {
		if _, ok := routes[route]; !ok {
			t.Fatalf("expected route %q to be registered, but it was missing", route)
		}
	}
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

	t.Run("empty reason 400", func(t *testing.T) {
		body := strings.NewReader(`{"action_type":"recheck_evidence","reason":""}`)
		req := httptest.NewRequest(http.MethodPost, "/api/v2/agent-runs/"+agentRunID+"/followups", body)
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("too long reason 400", func(t *testing.T) {
		longReason := string(make([]byte, 501))
		body := strings.NewReader(fmt.Sprintf(`{"action_type":"recheck_evidence","reason":"%s"}`, longReason))
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
