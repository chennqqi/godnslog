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
	// Note: Some v1 endpoints are commented out as methods are not yet implemented
	v1 := r.engine.Group("/api/v1")
	{
		// 1.0 auth endpoints
		v1.POST("/login", r.server.userLogin)
		v1.POST("/logout", r.server.authHandler, r.server.userLogout)
		v1.GET("/info", r.server.authHandler, r.server.userInfo)

		// 1.0 record endpoints - TODO: Implement these methods
		// v1.GET("/record", r.server.authHandler, r.server.getRecord)
		// v1.POST("/record/delete", r.server.authHandler, r.server.deleteRecord)

		// 1.0 data endpoints - TODO: Implement these methods
		// v1.GET("/data", r.server.authHandler, r.server.getData)
		// v1.POST("/data/delete", r.server.authHandler, r.server.deleteData)

		// 1.0 user management - TODO: Implement these methods
		// v1.GET("/user", r.server.authHandler, r.server.getUserList)
		// v1.POST("/user", r.server.authHandler, r.server.createUser)
		// v1.PUT("/user", r.server.authHandler, r.server.updateUser)
		// v1.DELETE("/user", r.server.authHandler, r.server.deleteUser)

		// 1.0 settings - TODO: Implement these methods
		// v1.GET("/setting", r.server.authHandler, r.server.getSetting)
		// v1.POST("/setting", r.server.authHandler, r.server.setSetting)
	}

	// API v2 has a single source of truth in WebServer.registerV2API.
	r.server.registerV2API(r.engine)
}
