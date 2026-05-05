package server

import (
	"net/http"

	"github.com/chennqqi/godnslog/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// validateJWTToken validates a JWT token and returns the user
func (s *WebServer) validateJWTToken(token string) (*models.TblUser, error) {
	// TODO: Implement JWT token validation
	// For now, return nil to force API key authentication
	return nil, nil
}

// validateAPIKey validates an API key and returns the key details
func (s *WebServer) validateAPIKey(key string) (*models.TblAPIKey, error) {
	// TODO: Implement API key validation
	// For now, return nil
	return nil, nil
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
