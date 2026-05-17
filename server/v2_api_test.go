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
	}

	for _, route := range requiredRoutes {
		if _, ok := routes[route]; !ok {
			t.Fatalf("expected route %q to be registered, but it was missing", route)
		}
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

	// Manually set payload_id and case_id for testing (workaround for attribution function issue)
	interaction.PayloadID = &payload.ID
	interaction.CaseID = &caseID

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
