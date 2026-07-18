package server

import (
	"github.com/gin-gonic/gin"
)

// Router manages API routing with version support
type Router struct {
	engine *gin.Engine
	server *WebServer
}

// NewRouter creates a new router
func NewRouter(engine *gin.Engine, server *WebServer) *Router {
	return &Router{
		engine: engine,
		server: server,
	}
}

// RegisterRoutes registers all API routes
func (r *Router) RegisterRoutes() {
	// Apply global middleware
	r.engine.Use(CORSMiddleware())
	r.engine.Use(LoggingMiddleware())
	r.engine.Use(RecoveryMiddleware())

	// API v1 - 1.0 compatibility (deprecated, will be removed in future)
	v1 := r.engine.Group("/api/v1")
	{
		v1.POST("/login", r.server.userLogin)
		v1.POST("/logout", r.server.authHandler, r.server.userLogout)
		v1.GET("/info", r.server.authHandler, r.server.userInfo)
	}

	// API v2 has a single source of truth in WebServer.registerV2API.
	r.server.registerV2API(r.engine)
}
