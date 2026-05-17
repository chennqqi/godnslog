package server

import (
	"fmt"
	"net/http"

	"github.com/chennqqi/godnslog/models"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// authenticateJWT validates a JWT token from the request context and returns the user
func (s *WebServer) authenticateJWT(c *gin.Context) (*models.TblUser, error) {
	tokenString := c.GetHeader("Access-Token")
	if tokenString == "" {
		// Try Authorization header with Bearer token
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" && len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			tokenString = authHeader[7:]
		}
	}
	if tokenString == "" {
		return nil, nil
	}

	var claim MyClaims
	token, err := jwt.ParseWithClaims(tokenString, &claim, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.verifyKey), nil
	})
	if err != nil || !token.Valid {
		return nil, err
	}

	store := s.store
	key := fmt.Sprintf("%v.seed", claim.Id)
	realSeed, exist := store.Get(key)
	if !exist || realSeed.(string) != claim.Seed {
		return nil, fmt.Errorf("invalid seed")
	}

	u, exist := store.Get(fmt.Sprintf("%v.user", claim.Id))
	if !exist {
		return nil, fmt.Errorf("user not found")
	}

	return u.(*models.TblUser), nil
}

// authenticateAPIKey validates an API key from the request context and returns the key details
func (s *WebServer) authenticateAPIKey(c *gin.Context) (*models.TblAPIKey, error) {
	apiKey := c.GetHeader("X-API-Key")
	if apiKey == "" {
		return nil, nil
	}

	// Validate full API key using auth service
	key, err := s.authService.ValidateAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	// Update last used timestamp
	_ = s.authService.UpdateLastUsed(apiKey)

	// Convert to TblAPIKey for backward compatibility
	tblKey := &models.TblAPIKey{
		Key:       key.Key,
		KeyPrefix: key.KeyPrefix,
		Name:      key.Name,
		IsAgent:   key.IsAgent, // Propagate IsAgent to legacy model
	}

	// Convert ExpiresAt pointer to value
	if key.ExpiresAt != nil {
		tblKey.ExpiresAt = *key.ExpiresAt
	}

	// Convert Scopes array to JSON string
	// For now, leave as empty string since TblAPIKey expects JSON string
	// The actual scope checking will be done in the auth middleware

	return tblKey, nil
}

// CORSMiddleware handles CORS headers
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-API-Key")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// LoggingMiddleware logs all requests
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process request
		c.Next()

		// Log request
		status := c.Writer.Status()
		method := c.Request.Method
		path := c.Request.URL.Path
		clientIP := c.ClientIP()

		logrus.Infof("%s %s - Status: %d - IP: %s", method, path, status, clientIP)
	}
}

// RecoveryMiddleware recovers from panics
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logrus.Errorf("Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    5,
					"message": "Internal server error",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// AdminOnlyMiddleware restricts access to admin users only
func (s *WebServer) AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    4,
				"message": "Forbidden",
			})
			c.Abort()
			return
		}

		// Check if user is admin (role 0 or 1)
		if roleInt, ok := role.(int); ok {
			if roleInt > 1 {
				c.JSON(http.StatusForbidden, gin.H{
					"code":    4,
					"message": "Admin access required",
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
