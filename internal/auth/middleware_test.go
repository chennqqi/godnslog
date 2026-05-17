package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// TestAuthMiddlewareJWT verifies JWT authentication
func TestAuthMiddlewareJWT(t *testing.T) {
	t.Run("JWT with Access-Token header", func(t *testing.T) {
		// This test requires a valid JWT token and secret
		// For now, we test the structure and flow
		middleware := NewAuthMiddleware("test-secret", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Access-Token", "test-token")

		identity, err := middleware.authenticateJWT(c)

		// Since we don't have a valid token, we expect an error
		assert.Error(t, err)
		assert.Nil(t, identity)
	})

	t.Run("JWT with Authorization Bearer header", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("Authorization", "Bearer test-token")

		identity, err := middleware.authenticateJWT(c)

		// Since we don't have a valid token, we expect an error
		assert.Error(t, err)
		assert.Nil(t, identity)
	})

	t.Run("No JWT token", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		identity, err := middleware.authenticateJWT(c)

		// No error, just nil identity (will try API key next)
		assert.NoError(t, err)
		assert.Nil(t, identity)
	})
}

// TestAuthMiddlewareAPIKey verifies API Key authentication
func TestAuthMiddlewareAPIKey(t *testing.T) {
	t.Run("API Key with X-API-Key header (no service)", func(t *testing.T) {
		// Test with nil service - should return error
		middleware := NewAuthMiddleware("test-secret", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-API-Key", "test-api-key-12345678")

		identity, err := middleware.authenticateAPIKey(c)

		// Since service is nil, we expect an error
		assert.Error(t, err)
		assert.Nil(t, identity)
	})

	t.Run("No API Key header", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)

		identity, err := middleware.authenticateAPIKey(c)

		// No error, just nil identity
		assert.NoError(t, err)
		assert.Nil(t, identity)
	})

	t.Run("API Key too short", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret", nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Request.Header.Set("X-API-Key", "short")

		identity, err := middleware.authenticateAPIKey(c)

		assert.Error(t, err)
		assert.Nil(t, identity)
	})
}

// TestAuthMiddlewareRequireAuth verifies RequireAuth middleware
func TestAuthMiddlewareRequireAuth(t *testing.T) {
	t.Run("No credentials returns 401", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret", nil)

		router := gin.New()
		router.Use(middleware.RequireAuth())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestAuthMiddlewareRequireAgent verifies RequireAgent middleware
func TestAuthMiddlewareRequireAgent(t *testing.T) {
	t.Run("No credentials returns 401", func(t *testing.T) {
		middleware := NewAuthMiddleware("test-secret", nil)

		router := gin.New()
		router.Use(middleware.RequireAgent())
		router.GET("/test", func(c *gin.Context) {
			c.String(http.StatusOK, "success")
		})

		w := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/test", nil)
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

// TestGetIdentity verifies GetIdentity helper function
func TestGetIdentity(t *testing.T) {
	t.Run("Returns identity when present", func(t *testing.T) {
		expectedIdentity := &AuthIdentity{
			UserID:   "user-1",
			Username: "testuser",
			IsAgent:  false,
		}

		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set(ContextKeyIdentity, expectedIdentity)

		identity := GetIdentity(c)

		assert.NotNil(t, identity)
		assert.Equal(t, expectedIdentity.UserID, identity.UserID)
		assert.Equal(t, expectedIdentity.Username, identity.Username)
		assert.Equal(t, expectedIdentity.IsAgent, identity.IsAgent)
	})

	t.Run("Returns nil when not present", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		identity := GetIdentity(c)

		assert.Nil(t, identity)
	})
}

// TestAuthIdentityStructure verifies AuthIdentity fields
func TestAuthIdentityStructure(t *testing.T) {
	identity := AuthIdentity{
		UserID:      "user-1",
		Username:    "testuser",
		Role:        0,
		IsAgent:     true,
		APIKeyID:    "key-1",
		WorkspaceID: nil,
	}

	assert.Equal(t, "user-1", identity.UserID)
	assert.Equal(t, "testuser", identity.Username)
	assert.Equal(t, 0, identity.Role)
	assert.True(t, identity.IsAgent)
	assert.Equal(t, "key-1", identity.APIKeyID)
	assert.Nil(t, identity.WorkspaceID)
}

// TestContextKeys verifies context key constants
func TestContextKeys(t *testing.T) {
	assert.Equal(t, "auth_identity", ContextKeyIdentity)
	assert.Equal(t, "user", ContextKeyUser)
	assert.Equal(t, "role", ContextKeyRole)
}
