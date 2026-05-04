package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/models"
	"github.com/gin-gonic/gin"
)

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

	// Initialize router
	r := gin.New()
	server.registerV2API(r)

	// Test login with valid credentials
	reqBody := `{"username": "admin", "password": "password"}`
	req := httptest.NewRequest("POST", "/api/v2/auth/login", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// Check response
	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			Token string `json:"token"`
			User  struct {
				Id       int64  `json:"id"`
				Username string `json:"username"`
				Email    string `json:"email"`
			} `json:"user"`
		} `json:"data"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response.Code != models.CodeOK {
		t.Errorf("Expected code %d, got %d", models.CodeOK, response.Code)
	}

	if response.Data.Token == "" {
		t.Error("Expected token in response data")
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
