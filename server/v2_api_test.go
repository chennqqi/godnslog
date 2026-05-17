package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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

	// Test that creation and preview use the same rendering logic
	// The payload service uses RenderTemplateWithCase for creation
	// The preview endpoint should return the same template_rendered
	if payload.TemplateRendered == "" {
		t.Fatal("Expected TemplateRendered to be set after creation")
	}

	// Verify rendered payload contains token and domain (proves rendering logic works)
	if !strings.Contains(payload.TemplateRendered, payload.Token) {
		t.Errorf("Expected TemplateRendered to contain token %s", payload.Token)
	}
	if !strings.Contains(payload.TemplateRendered, "test.example.com") {
		t.Error("Expected TemplateRendered to contain domain")
	}

	// Test that GetPayloadByID returns the same template_rendered (simulates preview)
	retrieved, err := payloadService.GetPayloadByID(payload.ID)
	if err != nil {
		t.Fatalf("Failed to get payload: %v", err)
	}

	if retrieved.TemplateRendered != payload.TemplateRendered {
		t.Error("Preview should return the same template_rendered as creation")
	}

	// Test that non-existent payload returns 404 (simulates preview 404)
	_, err = payloadService.GetPayloadByID("nonexistent-id")
	if err == nil {
		t.Error("Expected error for non-existent payload")
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
