package auth

import (
	"errors"
	"net/http"
	"strings"

	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

var (
	ErrUnauthorized = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
)

// AuthIdentity represents the unified identity in Gin Context
type AuthIdentity struct {
	UserID      string
	Username    string
	Role        int
	IsAgent     bool
	APIKeyID    string
	WorkspaceID *string
}

// Context keys for Gin Context
const (
	ContextKeyIdentity = "auth_identity"
	ContextKeyUser     = "user"
	ContextKeyRole     = "role"
)

// JWTConfig holds JWT configuration
type JWTConfig struct {
	SecretKey string
}

// APIKeyConfig holds API key configuration
type APIKeyConfig struct {
	Service *Service
}

// AuthMiddleware provides unified authentication middleware
type AuthMiddleware struct {
	jwtConfig    *JWTConfig
	apiKeyConfig *APIKeyConfig
}

// NewAuthMiddleware creates a new unified auth middleware
func NewAuthMiddleware(jwtSecret string, authService *Service) *AuthMiddleware {
	return &AuthMiddleware{
		jwtConfig:    &JWTConfig{SecretKey: jwtSecret},
		apiKeyConfig: &APIKeyConfig{Service: authService},
	}
}

// Authenticate authenticates requests using JWT or API Key
// Priority: JWT > API Key
func (m *AuthMiddleware) Authenticate(c *gin.Context) (*AuthIdentity, error) {
	// Try JWT first
	identity, err := m.authenticateJWT(c)
	if err == nil && identity != nil {
		c.Set(ContextKeyIdentity, identity)
		return identity, nil
	}

	// Fall back to API Key
	identity, err = m.authenticateAPIKey(c)
	if err == nil && identity != nil {
		c.Set(ContextKeyIdentity, identity)
		return identity, nil
	}

	return nil, ErrUnauthorized
}

// authenticateJWT validates JWT token and returns identity
func (m *AuthMiddleware) authenticateJWT(c *gin.Context) (*AuthIdentity, error) {
	tokenString := c.GetHeader("Access-Token")
	if tokenString == "" {
		// Try Authorization header with Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = authHeader[7:]
		}
	}
	if tokenString == "" {
		return nil, nil // No JWT token, try API key
	}

	var claim jwt.MapClaims
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(m.jwtConfig.SecretKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	// Extract user info from claims
	userID, ok := claim["id"].(string)
	if !ok {
		return nil, errors.New("invalid token: missing user id")
	}

	username, _ := claim["username"].(string)
	roleFloat, _ := claim["role"].(float64)
	role := int(roleFloat)

	return &AuthIdentity{
		UserID:   userID,
		Username: username,
		Role:     role,
		IsAgent:  false,
	}, nil
}

// authenticateAPIKey validates API key and returns identity
func (m *AuthMiddleware) authenticateAPIKey(c *gin.Context) (*AuthIdentity, error) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return nil, nil // No API key either
	}

	// Check if service is available
	if m.apiKeyConfig.Service == nil {
		return nil, errors.New("auth service not configured")
	}

	// Validate full API key from database
	key, err := m.apiKeyConfig.Service.ValidateAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Update last used timestamp
	_ = m.apiKeyConfig.Service.UpdateLastUsed(apiKey)

	return &AuthIdentity{
		UserID:      key.CreatedBy,
		IsAgent:     key.IsAgent,
		APIKeyID:    key.ID,
		WorkspaceID: nil, // TODO: add workspace_id to APIKey model
	}, nil
}

// RequireAuth creates middleware that requires authentication
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := m.Authenticate(c)
		if err != nil || identity == nil {
			c.JSON(http.StatusUnauthorized, v2models.UnauthorizedResponse(""))
			c.Abort()
			return
		}

		// Set legacy context keys for backward compatibility
		// TODO: Handle int64 ID conversion from string UserID
		// c.Set(ContextKeyUser, &models.TblUser{Id: identity.UserID})
		if identity.Role > 0 {
			c.Set(ContextKeyRole, identity.Role)
		}

		c.Next()
	}
}

// RequireAgent creates middleware that requires Agent API key
func (m *AuthMiddleware) RequireAgent() gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := m.Authenticate(c)
		if err != nil || identity == nil {
			c.JSON(http.StatusUnauthorized, v2models.UnauthorizedResponse(""))
			c.Abort()
			return
		}

		if !identity.IsAgent {
			c.JSON(http.StatusForbidden, v2models.ForbiddenResponse("Agent API key required"))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireScope creates middleware that requires specific scope
func (m *AuthMiddleware) RequireScope(scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		identity, err := m.Authenticate(c)
		if err != nil || identity == nil {
			c.JSON(http.StatusUnauthorized, v2models.UnauthorizedResponse(""))
			c.Abort()
			return
		}

		// If using JWT (not agent), check if user has admin role
		if !identity.IsAgent {
			if identity.Role > 1 {
				c.JSON(http.StatusForbidden, v2models.ForbiddenResponse(""))
				c.Abort()
				return
			}
			c.Next()
			return
		}

		// For agent keys, check scope
		if identity.APIKeyID != "" {
			key, err := m.apiKeyConfig.Service.GetAPIKeyByID(identity.APIKeyID)
			if err != nil {
				c.JSON(http.StatusForbidden, v2models.ForbiddenResponse(""))
				c.Abort()
				return
			}
			if !key.HasScope(scope) {
				c.JSON(http.StatusForbidden, v2models.ForbiddenResponse("Insufficient scope"))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// GetIdentity retrieves the authenticated identity from Gin Context
func GetIdentity(c *gin.Context) *AuthIdentity {
	if identity, exists := c.Get(ContextKeyIdentity); exists {
		return identity.(*AuthIdentity)
	}
	return nil
}
