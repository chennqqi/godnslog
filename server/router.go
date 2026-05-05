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

	// API v2 - 2.0 unified API (current version)
	v2 := r.engine.Group("/api/v2")
	{
		// Auth
		v2.POST("/auth/login", r.server.v2Login)
		v2.POST("/auth/logout", r.server.authHandler, r.server.v2Logout)
		v2.GET("/auth/info", r.server.authHandler, r.server.v2UserInfo)

		// Cases
		cases := v2.Group("/cases", r.server.authHandler)
		{
			cases.GET("", r.server.v2ListCases)
			cases.POST("", r.server.v2CreateCase)
			cases.GET("/:id", r.server.v2GetCase)
			cases.PUT("/:id", r.server.v2UpdateCase)
			cases.DELETE("/:id", r.server.v2DeleteCase)
		}

		// Payloads
		payloads := v2.Group("/payloads", r.server.authHandler)
		{
			payloads.GET("", r.server.v2ListPayloads)
			payloads.POST("", r.server.v2CreatePayload)
			payloads.GET("/:id", r.server.v2GetPayload)
			payloads.POST("/:id/revoke", r.server.v2RevokePayload)
		}

		// Interactions
		interactions := v2.Group("/interactions", r.server.authHandler)
		{
			interactions.GET("", r.server.v2ListInteractions)
			interactions.GET("/:id", r.server.v2GetInteraction)
			interactions.POST("/delete", r.server.v2DeleteInteractions)
			interactions.POST("/export", r.server.v2ExportInteractions)
		}

		// APIKeys
		apikeys := v2.Group("/apikeys", r.server.authHandler)
		{
			apikeys.GET("", r.server.v2ListAPIKeys)
			apikeys.POST("", r.server.v2CreateAPIKey)
			apikeys.DELETE("/:id", r.server.v2DeleteAPIKey)
		}

		// Users (admin only)
		users := v2.Group("/users", r.server.authHandler, r.server.AdminOnlyMiddleware())
		{
			users.GET("", r.server.v2ListUsers)
		}

		// Settings
		settings := v2.Group("/settings", r.server.authHandler)
		{
			settings.GET("", r.server.v2ListSettings)
			settings.POST("", r.server.v2CreateSetting)
			settings.GET("/:key", r.server.v2GetSetting)
			settings.PUT("/:key", r.server.v2UpdateSetting)
			settings.DELETE("/:key", r.server.v2DeleteSetting)
		}

		// Marketplace
		marketplace := v2.Group("/marketplace", r.server.authHandler)
		{
			marketplace.GET("/plugins", r.server.v2ListPlugins)
			marketplace.GET("/plugins/:id", r.server.v2GetPlugin)
			marketplace.GET("/templates", r.server.v2ListTemplates)
			marketplace.GET("/templates/:id", r.server.v2GetTemplate)
		}

		// Rules/Workflow
		rules := v2.Group("/rules", r.server.authHandler)
		{
			rules.GET("", r.server.v2ListRules)
			rules.POST("", r.server.v2CreateRule)
			rules.GET("/:id", r.server.v2GetRule)
			rules.PUT("/:id", r.server.v2UpdateRule)
			rules.DELETE("/:id", r.server.v2DeleteRule)
		}
	}
}
