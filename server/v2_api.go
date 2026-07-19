package server

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/internal/agentpolicy"
	"github.com/chennqqi/godnslog/internal/agentrun"
	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/canary"
	"github.com/chennqqi/godnslog/internal/evidencehub"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/listener"
	"github.com/chennqqi/godnslog/internal/marketplace"
	"github.com/chennqqi/godnslog/internal/mcp"
	"github.com/chennqqi/godnslog/internal/notification"
	"github.com/chennqqi/godnslog/internal/payload"
	"github.com/chennqqi/godnslog/internal/retention"
	"github.com/chennqqi/godnslog/internal/scannerhub"
	"github.com/chennqqi/godnslog/internal/scannerhub/search"
	"github.com/dgrijalva/jwt-go"

	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/rebinding"
	"github.com/chennqqi/godnslog/internal/workflow"
	"github.com/chennqqi/godnslog/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// getSetting retrieves a setting value by key from the Settings table.
func (self *WebServer) getSetting(key string) string {
	var setting v2models.Settings
	_, err := self.orm.Where("key = ?", key).Get(&setting)
	if err != nil || setting.ID == "" {
		return ""
	}
	return setting.Value
}

// registerV2API registers v2 API routes
func (self *WebServer) registerV2API(r *gin.Engine) {
	v2 := r.Group("/api/v2")
	{
		// Auth
		v2.POST("/auth/login", self.v2Login)
		v2.POST("/auth/logout", self.authHandler, self.v2Logout)
		v2.GET("/auth/info", self.authHandler, self.v2UserInfo)

		// Cases
		cases := v2.Group("/cases", self.authHandler)
		{
			cases.GET("", self.v2ListCases)
			cases.POST("", self.v2CreateCase)
			cases.GET("/stats", self.v2CaseStats)
			cases.GET("/:id", self.v2GetCase)
			cases.PUT("/:id", self.v2UpdateCase)
			cases.DELETE("/:id", self.v2DeleteCase)
			cases.GET("/:id/stats", self.v2GetCaseStats)
			cases.GET("/:id/payloads", self.v2GetCasePayloads)
			cases.GET("/:id/interactions", self.v2GetCaseInteractions)
		}

		// Payloads
		payloads := v2.Group("/payloads", self.authHandler)
		{
			payloads.GET("", self.v2ListPayloads)
			payloads.POST("", self.v2CreatePayload)
			// Static path segments must be registered before /:id so Gin does not treat e.g. "batch" as an id.
			payloads.POST("/batch", self.v2BatchCreatePayloads)
			payloads.GET("/:id", self.v2GetPayload)
			payloads.PUT("/:id", self.v2UpdatePayload)
			payloads.POST("/:id/revoke", self.v2RevokePayload)
			payloads.POST("/:id/preview", self.v2PreviewPayload)
			payloads.GET("/:id/interactions", self.v2ListPayloadInteractions)
		}

		// Interactions
		interactions := v2.Group("/interactions", self.authHandler)
		{
			interactions.GET("", self.v2ListInteractions)
			// Register /stats, /timeline, /stream before /:id so paths are not captured as ids.
			interactions.GET("/stats", self.v2InteractionStats)
			interactions.GET("/stats/daily", self.v2InteractionDailyStats)
			interactions.GET("/timeline", self.v2InteractionTimeline)
			interactions.GET("/stream", self.v2InteractionStream)
			interactions.POST("/delete", self.v2DeleteInteractions)
			interactions.POST("/export", self.v2ExportInteractions)
			interactions.GET("/:id", self.v2GetInteraction)
		}

		// Attack chains
		v2.GET("/attack-chains", self.authHandler, self.v2ListAttackChains)
		v2.GET("/attack-chains/:token", self.authHandler, self.v2GetAttackChainDetail)

		// APIKeys
		apikeys := v2.Group("/apikeys", self.authHandler)
		{
			apikeys.GET("", self.v2ListAPIKeys)
			apikeys.POST("", self.v2CreateAPIKey)
			apikeys.GET("/:id", self.v2GetAPIKey)
			apikeys.PUT("/:id", self.v2UpdateAPIKey)
			apikeys.DELETE("/:id", self.v2DeleteAPIKey)
		}

		agentPolicy := v2.Group("/agent-policy", self.authHandler)
		{
			agentPolicy.GET("/scopes", self.v2ListAgentPolicyScopes)
		}

		// Notifications
		notifications := v2.Group("/notifications", self.authHandler)
		{
			notifications.GET("/channels", self.v2ListNotificationChannels)
			notifications.POST("/channels", self.v2CreateNotificationChannel)
			notifications.GET("/channels/:id", self.v2GetNotificationChannel)
			notifications.PUT("/channels/:id", self.v2UpdateNotificationChannel)
			notifications.DELETE("/channels/:id", self.v2DeleteNotificationChannel)
			notifications.GET("/logs", self.v2ListNotificationLogs)
		}

		// Users (admin only)
		users := v2.Group("/users", self.authHandler)
		{
			users.GET("", self.v2ListUsers)
			users.POST("", self.v2CreateUser)
			users.PUT("/:id", self.v2UpdateUser)
			users.DELETE("/:id", self.v2DeleteUser)
		}

		// Marketplace
		marketplaceRoutes := v2.Group("/marketplace", self.authHandler)
		{
			marketplaceRoutes.GET("/plugins", self.v2ListPlugins)
			marketplaceRoutes.POST("/plugins", self.v2CreatePlugin)
			marketplaceRoutes.GET("/plugins/:id", self.v2GetPlugin)
			marketplaceRoutes.POST("/plugins/:id/install", self.v2InstallPlugin)
			marketplaceRoutes.GET("/templates", self.v2ListTemplates)
			marketplaceRoutes.POST("/templates", self.v2CreateTemplate)
			marketplaceRoutes.GET("/templates/:id", self.v2GetTemplate)
			marketplaceRoutes.GET("/installed", self.v2ListInstalledPlugins)
			marketplaceRoutes.DELETE("/installed/:id", self.v2UninstallPlugin)
		}

		// Rules/Workflow
		rules := v2.Group("/rules", self.authHandler)
		{
			rules.GET("", self.v2ListRules)
			rules.POST("", self.v2CreateRule)
			rules.GET("/:id", self.v2GetRule)
			rules.PUT("/:id", self.v2UpdateRule)
			rules.DELETE("/:id", self.v2DeleteRule)
		}

		// Evidence
		evidence := v2.Group("/evidence", self.authHandler)
		{
			evidence.POST("/generate", self.v2GenerateEvidence)
			evidence.POST("/summary", self.v2SummarizeEvidence)
			evidence.GET("/:id", self.v2GetEvidence)
		}

		// Audit
		audit := v2.Group("/audit", self.authHandler)
		{
			audit.GET("/logs", self.v2ListAuditLogs)
			audit.POST("/logs", self.v2CreateAuditLog)
		}

		// Canary
		canary := v2.Group("/canary", self.authHandler)
		{
			canary.GET("", self.v2ListCanaries)
			canary.POST("", self.v2CreateCanary)
			// Register /:id/hits before /:id for consistent matching across Gin versions.
			canary.GET("/:id/hits", self.v2ListCanaryHits)
			canary.GET("/:id", self.v2GetCanary)
			canary.PUT("/:id", self.v2UpdateCanary)
			canary.DELETE("/:id", self.v2DeleteCanary)
		}

		// Rebinding
		rebinding := v2.Group("/rebinding", self.authHandler)
		{
			rebinding.GET("/rules", self.v2ListRebindingRules)
			rebinding.POST("/rules", self.v2CreateRebindingRule)
			rebinding.GET("/rules/:id", self.v2GetRebindingRule)
			rebinding.PUT("/rules/:id", self.v2UpdateRebindingRule)
			rebinding.DELETE("/rules/:id", self.v2DeleteRebindingRule)
			rebinding.GET("/rules/:id/sessions", self.v2ListRebindingSessions)
			rebinding.GET("/scenarios", self.v2ListRebindingScenarios)
			rebinding.POST("/scenarios/:name/rules", self.v2CreateRebindingFromScenario)
		}

		// DNS Records
		dnsRecords := v2.Group("/dns/records", self.authHandler)
		{
			dnsRecords.GET("", self.v2ListDNSRecords)
			dnsRecords.POST("", self.v2CreateDNSRecord)
			dnsRecords.PUT("/:id", self.v2UpdateDNSRecord)
			dnsRecords.DELETE("/:id", self.v2DeleteDNSRecord)
		}

		// XIP encoding query
		v2.GET("/dns/xip/:ip", self.authHandler, self.v2QueryXip)

		// Listeners
		listeners := v2.Group("/listeners", self.authHandler)
		{
			listeners.GET("", self.v2ListListeners)
			listeners.POST("", self.v2CreateListener)
			listeners.GET("/:id/interactions", self.v2ListListenerInteractions)
			listeners.GET("/:id", self.v2GetListener)
			listeners.PUT("/:id", self.v2UpdateListener)
			listeners.DELETE("/:id", self.v2DeleteListener)
		}

		// Settings
		settings := v2.Group("/settings", self.authHandler)
		{
			settings.GET("", self.v2ListSettings)
			settings.POST("", self.v2CreateSetting)
			settings.GET("/:key", self.v2GetSetting)
			settings.PUT("/:key", self.v2UpdateSetting)
			settings.DELETE("/:key", self.v2DeleteSetting)
		}

		// Retention
		retentionGroup := v2.Group("/retention", self.authHandler)
		{
			retentionGroup.GET("/policies", self.v2ListRetentionPolicies)
			retentionGroup.POST("/policies", self.v2CreateRetentionPolicy)
			retentionGroup.GET("/policies/:id", self.v2GetRetentionPolicy)
			retentionGroup.PUT("/policies/:id", self.v2UpdateRetentionPolicy)
			retentionGroup.DELETE("/policies/:id", self.v2DeleteRetentionPolicy)
			retentionGroup.POST("/policies/:id/run", self.v2RunRetentionPolicy)
			retentionGroup.GET("/jobs", self.v2ListRetentionJobs)
			retentionGroup.GET("/archives", self.v2ListRetentionArchives)
		}

		// Scanner Hub
		scannerHub := v2.Group("/scanner-hub", self.authHandler)
		{
			scannerHub.GET("/adapters", self.v2ListScannerAdapters)
		}

		// Search engines (ZoomEye, Shodan, Fofa)
		searchGroup := v2.Group("/search", self.authHandler)
		{
			searchGroup.GET("/zoomeye", self.v2SearchZoomEye)
			searchGroup.GET("/shodan", self.v2SearchShodan)
			searchGroup.GET("/fofa", self.v2SearchFofa)
		}

		scannerRuns := v2.Group("/scanner-runs", self.authHandler)
		{
			scannerRuns.GET("", self.v2ListScannerRuns)
			scannerRuns.POST("", self.v2CreateScannerRun)
			scannerRuns.POST("/from-search", self.v2CreateScannerRunFromSearch)
			scannerRuns.GET("/:id", self.v2GetScannerRun)
			scannerRuns.PUT("/:id/status", self.v2UpdateScannerRunStatus)
			scannerRuns.POST("/:id/backfill", self.v2BackfillScannerResults)
		}

		// Agent Runs
		agentRuns := v2.Group("/agent-runs", self.authHandler)
		{
			agentRuns.GET("", self.v2ListAgentRuns)
			agentRuns.POST("", self.v2CreateAgentRun)
			agentRuns.GET("/:id", self.v2GetAgentRun)
			agentRuns.GET("/:id/review", self.v2GetAgentRunReview)
			agentRuns.PUT("/:id/status", self.v2UpdateAgentRunStatus)
			agentRuns.POST("/:id/operations", self.v2AppendAgentOperation)
			agentRuns.POST("/:id/followups", self.v2CreateAgentRunFollowup)
			agentRuns.POST("/:id/review-decision", self.v2RecordReviewDecision)
			agentRuns.POST("/:id/review-export", self.v2ExportReviewPackage)
			agentRuns.POST("/:id/review-delivery", self.v2DeliverReviewPackage)
			agentRuns.GET("/:id/review-deliveries", self.v2ListReviewDeliveries)
			agentRuns.GET("/review-package-trace", self.v2TraceReviewPackage)
			agentRuns.GET("/review-queue", self.v2ListReviewQueue)
			agentRuns.GET("/:id/followups", self.v2ListFollowupHistory)
			agentRuns.POST("/:id/complete", self.v2CompleteAgentRun)
		}

		// HA Cluster
		haCluster := v2.Group("/cluster", self.authHandler)
		{
			haCluster.GET("/nodes", self.v2ListClusterNodes)
			haCluster.POST("/nodes", self.v2CreateClusterNode)
			haCluster.GET("/nodes/:id", self.v2GetClusterNode)
			haCluster.PUT("/nodes/:id", self.v2UpdateClusterNode)
			haCluster.DELETE("/nodes/:id", self.v2DeleteClusterNode)
			haCluster.GET("/nodes/:id/health", self.v2NodeHealthCheck)
			haCluster.GET("/config", self.v2GetClusterConfig)
			haCluster.PUT("/config", self.v2UpdateClusterConfig)
			haCluster.GET("/status", self.v2ClusterStatus)
			haCluster.GET("/leader", self.v2GetLeader)
		}
		// Poll API (Burp Collaborator style cursor-based polling)
		v2.GET("/poll", self.authHandler, self.v2Poll)
	}

	// Health endpoints (no auth required)
	v2.GET("/health", self.v2HealthCheck)
	v2.GET("/ready", self.v2ReadinessCheck)

	// MCP Streamable HTTP transport (JSON-RPC 2.0)
	// Uses APIKey auth via Bearer token; no session-based authHandler
	v2.POST("/mcp", self.v2MCPHandler)
}

// v2Login handles v2 login
// @Summary User login
// @Description Authenticate with username/password and receive JWT token
// @Tags v2, auth
// @Accept json
// @Produce json
// @Param body body LoginRequest true "Login credentials"
// @Success 200 {object} gin.H "Login successful with token"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 401 {object} gin.H "Invalid credentials"
// @Router /api/v2/login [post]
func (self *WebServer) v2Login(c *gin.Context) {
	T := getTranslateFunc(c)

	var req LoginRequest
	err := c.BindJSON(&req)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2Login] BindJSON error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": T("bad input"),
		})
		return
	}

	logrus.Infof("[v2_api.go::v2Login] login request: username=%s", req.Username)

	session := self.orm.NewSession()
	defer session.Close()
	var user = new(models.TblUser)
	// Only use username for query (email is optional in frontend)
	exist, err := session.Where(`name=?`, req.Username).Get(user)

	if err != nil {
		logrus.Errorf("[v2_api.go::v2Login] orm.Get error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": T("bad service"),
		})
		return
	} else if !exist {
		logrus.Infof("[v2_api.go::v2Login] user not found: username=%s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": T("bad request"),
		})
		return
	}

	logrus.Infof("[v2_api.go::v2Login] user found: id=%d, name=%s", user.Id, user.Name)

	err = comparePassword(req.Password, user.Pass)
	if err != nil {
		logrus.Infof("[v2_api.go::v2Login] password not match for user: %s", req.Username)
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    401,
			"message": T("bad request"),
		})
		return
	}

	now := time.Now()
	seed := getSecuritySeed()
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, MyClaims{
		seed,
		jwt.StandardClaims{
			Id:        fmt.Sprintf("%v", user.Id),
			Audience:  user.Name,
			Subject:   user.Email,
			ExpiresAt: now.Add(3600 * 24 * time.Second).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    self.Domain,
		},
	})

	tokenString, err := token.SignedString([]byte(self.verifyKey))
	if err != nil {
		logrus.Errorf("[v2_api.go::v2Login] token.SignedString error: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{
			"code":    502,
			"message": T("bad service"),
		})
		return
	}
	store := self.store

	seedKey := fmt.Sprintf("%v.seed", user.Id)
	userKey := fmt.Sprintf("%v.user", user.Id)
	store.Set(seedKey, seed, self.AuthExpire)
	store.Set(userKey, user, cache.NoExpiration)

	logrus.Infof("[v2_api.go::v2Login] login success: username=%s", req.Username)

	// Return data in format expected by frontend: { code: 0, message: "OK", data: { token, user } }
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": T("OK"),
		"data": gin.H{
			"token": tokenString,
			"user": gin.H{
				"id":       user.Id,
				"username": user.Name,
				"email":    user.Email,
				"role":     user.Role,
				"lang":     user.Lang,
			},
		},
	})
}

// v2Logout handles v2 logout
// @Summary User logout
// @Description Invalidate the current session token
// @Tags v2, auth
// @Produce json
// @Success 200 {object} gin.H "Logout successful"
// @Router /api/v2/logout [post]
func (self *WebServer) v2Logout(c *gin.Context) {
	T := getTranslateFunc(c)

	store := self.store
	id := c.GetInt64("id")
	store.Delete(fmt.Sprintf("%v.seed", id))
	store.Delete(fmt.Sprintf("%v.user", id))
	c.JSON(200, gin.H{
		"code":    0,
		"message": T("OK"),
	})
}

// v2UserInfo handles v2 user info
// @Summary Get current user info
// @Description Get information about the currently authenticated user
// @Tags v2, auth
// @Produce json
// @Success 200 {object} gin.H "User information"
// @Failure 401 {object} gin.H "Unauthorized"
// @Router /api/v2/userinfo [get]
func (self *WebServer) v2UserInfo(c *gin.Context) {
	T := getTranslateFunc(c)

	// Check if authenticated via API key
	if apiKeyFull, exists := c.Get("api_key_full"); exists {
		key, ok := apiKeyFull.(*v2models.APIKey)
		if !ok {
			c.JSON(500, gin.H{
				"code":    500,
				"message": "api key data type error",
			})
			return
		}

		c.JSON(200, gin.H{
			"code":    0,
			"message": T("OK"),
			"data": gin.H{
				"user_id":        key.CreatedBy,
				"api_key_id":     key.ID,
				"api_key_prefix": key.KeyPrefix,
				"scopes":         key.Scopes,
				"is_agent":       key.IsAgent,
				"risk_tolerance": key.RiskTolerance,
				"workspace_id":   key.WorkspaceID,
			},
		})
		return
	}

	// JWT authentication
	store := self.store
	id := c.GetInt64("id")
	userValue, found := store.Get(fmt.Sprintf("%v.user", id))
	if !found {
		c.JSON(404, gin.H{
			"code":    404,
			"message": T("user not found"),
		})
		return
	}

	user, ok := userValue.(models.TblUser)
	if !ok {
		c.JSON(500, gin.H{
			"code":    500,
			"message": "user data type error",
		})
		return
	}

	c.JSON(200, gin.H{
		"code":    0,
		"message": T("OK"),
		"data": gin.H{
			"id":       user.Id,
			"username": user.Name,
			"email":    user.Email,
			"role":     user.Role,
			"lang":     user.Lang,
		},
	})
}

// v2ListCases lists cases
// @Summary List cases
// @Description Get a paginated list of cases with optional status filter
// @Tags v2, cases
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Page size (default 20)"
// @Param status query string false "Filter by status"
// @Success 200 {object} gin.H "List of cases"
// @Failure 401 {object} gin.H "Unauthorized"
// @Router /api/v2/cases [get]
func (self *WebServer) v2ListCases(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")
	search := c.Query("search")

	session := self.orm.NewSession()
	defer session.Close()

	var cases []models.TblCase
	query := session.Table(new(models.TblCase))

	if status != "" {
		query = query.Where("status = ?", status)
	}
	if search != "" {
		query = query.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListCases] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	err = query.OrderBy("created_at DESC").Limit(pageSize, (page-1)*pageSize).Find(&cases)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListCases] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	items := make([]models.Case, len(cases))
	for i, item := range cases {
		var tags []string
		if item.Tags != "" {
			json.Unmarshal([]byte(item.Tags), &tags)
		}
		items[i] = models.Case{
			Id:          strconv.FormatInt(item.Id, 10),
			Title:       item.Title,
			Description: item.Description,
			Target:      item.Target,
			Status:      item.Status,
			Tags:        tags,
			CreatedBy:   strconv.FormatInt(item.CreatedBy, 10),
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   item.UpdatedAt.Format(time.RFC3339),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.CaseListResponse{
			Items:      items,
			Total:      int(total),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// v2CreateCase creates a case
// @Summary Create a case
// @Description Create a new case for tracking vulnerability verification
// @Tags v2, cases
// @Accept json
// @Produce json
// @Param body body models.CaseCreateRequest true "Case creation request"
// @Success 201 {object} gin.H "Created case"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Router /api/v2/cases [post]
func (self *WebServer) v2CreateCase(c *gin.Context) {
	var req models.CaseCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	if req.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "title is required",
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)

	caseItem := models.TblCase{
		Title:       req.Title,
		Description: req.Description,
		Target:      req.Target,
		Status:      "active",
		CreatedBy:   user.Id,
	}

	if req.Tags != nil {
		tagsJson, _ := json.Marshal(req.Tags)
		caseItem.Tags = string(tagsJson)
	}

	session := self.orm.NewSession()
	defer session.Close()

	_, err := session.Insert(&caseItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateCase] insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	var tags []string
	if caseItem.Tags != "" {
		json.Unmarshal([]byte(caseItem.Tags), &tags)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.Case{
			Id:          strconv.FormatInt(caseItem.Id, 10),
			Title:       caseItem.Title,
			Description: caseItem.Description,
			Target:      caseItem.Target,
			Status:      caseItem.Status,
			Tags:        tags,
			CreatedBy:   strconv.FormatInt(caseItem.CreatedBy, 10),
			CreatedAt:   caseItem.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   caseItem.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// v2GetCase gets a case
// @Summary Get a case
// @Description Get detailed information about a specific case by ID
// @Tags v2, cases
// @Produce json
// @Param id path int true "Case ID"
// @Success 200 {object} gin.H "Case details"
// @Failure 404 {object} gin.H "Case not found"
// @Router /api/v2/cases/{id} [get]
func (self *WebServer) v2GetCase(c *gin.Context) {
	id := c.Param("id")
	caseId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid case id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var caseItem models.TblCase
	has, err := session.ID(caseId).Get(&caseItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCase] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    6,
			"message": "case not found",
		})
		return
	}

	var tags []string
	if caseItem.Tags != "" {
		json.Unmarshal([]byte(caseItem.Tags), &tags)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.Case{
			Id:          strconv.FormatInt(caseItem.Id, 10),
			Title:       caseItem.Title,
			Description: caseItem.Description,
			Target:      caseItem.Target,
			Status:      caseItem.Status,
			Tags:        tags,
			CreatedBy:   strconv.FormatInt(caseItem.CreatedBy, 10),
			CreatedAt:   caseItem.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   caseItem.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// v2UpdateCase updates a case
// @Summary Update a case
// @Description Update case fields such as title, description, status, tags
// @Tags v2, cases
// @Accept json
// @Produce json
// @Param id path int true "Case ID"
// @Param body body models.CaseUpdateRequest true "Case update request"
// @Success 200 {object} gin.H "Updated case"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 404 {object} gin.H "Case not found"
// @Router /api/v2/cases/{id} [put]
func (self *WebServer) v2UpdateCase(c *gin.Context) {
	id := c.Param("id")
	caseId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid case id",
		})
		return
	}

	var req models.CaseUpdateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var caseItem models.TblCase
	has, err := session.ID(caseId).Get(&caseItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateCase] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    6,
			"message": "case not found",
		})
		return
	}

	if req.Title != "" {
		caseItem.Title = req.Title
	}
	if req.Description != "" {
		caseItem.Description = req.Description
	}
	if req.Target != "" {
		caseItem.Target = req.Target
	}
	if req.Status != "" {
		caseItem.Status = req.Status
	}
	if req.Tags != nil {
		tagsJson, _ := json.Marshal(req.Tags)
		caseItem.Tags = string(tagsJson)
	}

	_, err = session.ID(caseId).Update(&caseItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateCase] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	var tags []string
	if caseItem.Tags != "" {
		json.Unmarshal([]byte(caseItem.Tags), &tags)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.Case{
			Id:          strconv.FormatInt(caseItem.Id, 10),
			Title:       caseItem.Title,
			Description: caseItem.Description,
			Target:      caseItem.Target,
			Status:      caseItem.Status,
			Tags:        tags,
			CreatedBy:   strconv.FormatInt(caseItem.CreatedBy, 10),
			CreatedAt:   caseItem.CreatedAt.Format(time.RFC3339),
			UpdatedAt:   caseItem.UpdatedAt.Format(time.RFC3339),
		},
	})
}

// v2DeleteCase deletes a case
// @Summary Delete a case
// @Description Delete a case and its associated data
// @Tags v2, cases
// @Produce json
// @Param id path int true "Case ID"
// @Success 200 {object} gin.H "Deletion successful"
// @Failure 404 {object} gin.H "Case not found"
// @Router /api/v2/cases/{id} [delete]
func (self *WebServer) v2DeleteCase(c *gin.Context) {
	id := c.Param("id")
	caseId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid case id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	_, err = session.ID(caseId).Delete(new(models.TblCase))
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteCase] delete error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2GetCaseStats gets case statistics
// @Summary Get case statistics
// @Description Get aggregated statistics for a specific case
// @Tags v2, cases
// @Produce json
// @Param id path int true "Case ID"
// @Success 200 {object} gin.H "Case statistics"
// @Router /api/v2/cases/{id}/stats [get]
func (self *WebServer) v2GetCaseStats(c *gin.Context) {
	id := c.Param("id")
	caseId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid case id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var caseItem models.TblCase
	has, err := session.ID(caseId).Get(&caseItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCaseStats] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "case not found",
		})
		return
	}

	// Count payloads
	payloadCount, err := session.Where("case_id = ?", caseId).Count(new(models.TblPayload))
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCaseStats] count payloads error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Count interactions
	interactionCount, err := session.Table("interactions").Where("case_id = ?", id).Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCaseStats] count interactions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Count hit payloads
	hitCount, err := session.Where("case_id = ? AND status = ?", caseId, "hit").Count(new(models.TblPayload))
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCaseStats] count hit payloads error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"payload_count":     payloadCount,
			"interaction_count": interactionCount,
			"hit_payload_count": hitCount,
		},
	})
}

// v2CaseStats returns aggregate case statistics across all cases
// @Summary Get case stats
// @Description Get aggregated statistics for all cases (total, active, archived, batch counts)
// @Tags v2, cases
// @Accept json
// @Produce json
// @Success 200 {object} gin.H "case statistics"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/v2/cases/stats [get]
func (self *WebServer) v2CaseStats(c *gin.Context) {
	type CaseStats struct {
		Total    int64 `json:"total"`
		Active   int64 `json:"active"`
		Archived int64 `json:"archived"`
		Batch    int64 `json:"batch"`
	}
	var stats CaseStats
	stats.Total, _ = self.orm.Count(&v2models.Case{})
	stats.Active, _ = self.orm.Where("status = 'active'").Count(&v2models.Case{})
	stats.Archived, _ = self.orm.Where("status = 'archived'").Count(&v2models.Case{})
	stats.Batch, _ = self.orm.Where("type = 'batch' OR type = 'scan'").Count(&v2models.Case{})
	c.JSON(200, gin.H{"code": 0, "data": stats})
}

// v2GetCasePayloads gets payloads associated with a case
// @Summary List payloads for a case
// @Description Get all payloads associated with a specific case
// @Tags v2, cases, payloads
// @Produce json
// @Param id path int true "Case ID"
// @Success 200 {object} gin.H "List of payloads"
// @Router /api/v2/cases/{id}/payloads [get]
func (self *WebServer) v2GetCasePayloads(c *gin.Context) {
	id := c.Param("id")
	caseId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid case id",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	session := self.orm.NewSession()
	defer session.Close()

	var payloads []models.TblPayload
	query := session.Where("case_id = ?", caseId)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	total, err := query.Count(new(models.TblPayload))
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCasePayloads] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	offset := (page - 1) * pageSize
	err = query.OrderBy("created_at DESC").Limit(pageSize, offset).Find(&payloads)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCasePayloads] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       payloads,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// v2GetCaseInteractions gets interactions associated with a case
// @Summary List interactions for a case
// @Description Get paginated interactions associated with a specific case
// @Tags v2, cases, interactions
// @Produce json
// @Param id path int true "Case ID"
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Success 200 {object} gin.H "List of interactions"
// @Router /api/v2/cases/{id}/interactions [get]
func (self *WebServer) v2GetCaseInteractions(c *gin.Context) {
	id := c.Param("id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	protocol := c.Query("protocol")

	session := self.orm.NewSession()
	defer session.Close()

	var interactions []v2models.Interaction
	query := session.Table("interactions").Where("case_id = ?", id)

	if protocol != "" {
		query = query.Where("type = ?", protocol)
	}

	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCaseInteractions] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	offset := (page - 1) * pageSize
	err = query.OrderBy("timestamp DESC").Limit(pageSize, offset).Find(&interactions)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetCaseInteractions] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       interactions,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// v2ListPayloads lists payloads
func (self *WebServer) v2ListPayloads(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	caseId := c.Query("case_id")
	status := c.Query("status")

	session := self.orm.NewSession()
	defer session.Close()

	var payloads []models.TblPayload
	query := session.Table(new(models.TblPayload))

	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListPayloads] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	err = query.OrderBy("created_at DESC").Limit(pageSize, (page-1)*pageSize).Find(&payloads)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListPayloads] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	items := make([]models.Payload, len(payloads))
	for i, item := range payloads {
		var variables map[string]string
		if item.Variables != "" {
			json.Unmarshal([]byte(item.Variables), &variables)
		}
		items[i] = models.Payload{
			Id:               strconv.FormatInt(item.Id, 10),
			CaseId:           strconv.FormatInt(item.CaseId, 10),
			Token:            item.Token,
			Template:         item.Template,
			RenderedPayload:  item.RenderedPayload,
			Variables:        variables,
			Status:           item.Status,
			ExpectedProtocol: item.ExpectedProtocol,
			CreatedBy:        strconv.FormatInt(item.CreatedBy, 10),
			CreatedAt:        item.CreatedAt.Format(time.RFC3339),
			UpdatedAt:        item.UpdatedAt.Format(time.RFC3339),
		}
		if !item.ExpiresAt.IsZero() {
			items[i].ExpiresAt = item.ExpiresAt.Format(time.RFC3339)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.PayloadListResponse{
			Items:      items,
			Total:      int(total),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// v2CreatePayload creates a payload
func (self *WebServer) v2CreatePayload(c *gin.Context) {
	var req models.PayloadCreateRequest
	var err error
	if err = c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	if req.TemplateID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "template_id is required",
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	// Convert ExpiresAt string to *time.Time if provided
	var expiresAt *time.Time
	if req.ExpiresAt != "" {
		parsedTime, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			expiresAt = &parsedTime
		}
	}

	// Create unified request for payload service
	unifiedReq := v2models.PayloadCreateRequest{
		CaseID:           req.CaseID,
		TemplateID:       req.TemplateID,
		Variables:        req.Variables,
		ExpiresAt:        expiresAt,
		ExpectedProtocol: req.ExpectedProtocol,
	}

	payloadService := payload.NewService(self.orm)
	payloadItem, err := payloadService.CreatePayload(&unifiedReq, userID, self.Domain)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreatePayload] create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    payloadItem,
	})
}

// v2GetPayload gets a payload
func (self *WebServer) v2GetPayload(c *gin.Context) {
	id := c.Param("id")

	payloadService := payload.NewService(self.orm)
	payloadItem, err := payloadService.GetPayloadByID(id)
	if err != nil {
		if err == payload.ErrPayloadNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "payload not found",
			})
			return
		}
		logrus.Errorf("[v2_api.go::v2GetPayload] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    payloadItem,
	})
}

// v2RevokePayload revokes a payload
func (self *WebServer) v2RevokePayload(c *gin.Context) {
	id := c.Param("id")
	payloadId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid payload id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	_, err = session.ID(payloadId).Update(&models.TblPayload{Status: "archived"})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2RevokePayload] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2UpdatePayload updates a payload
func (self *WebServer) v2UpdatePayload(c *gin.Context) {
	id := c.Param("id")
	payloadId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid payload id",
		})
		return
	}

	var req models.PayloadUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	_, err = session.ID(payloadId).Cols("status", "expected_protocol").Update(&models.TblPayload{
		Status:           req.Status,
		ExpectedProtocol: req.ExpectedProtocol,
	})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdatePayload] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2PreviewPayload previews payload rendering
func (self *WebServer) v2PreviewPayload(c *gin.Context) {
	id := c.Param("id")

	payloadService := payload.NewService(self.orm)
	payloadItem, err := payloadService.GetPayloadByID(id)
	if err != nil {
		if err == payload.ErrPayloadNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "payload not found",
			})
			return
		}
		logrus.Errorf("[v2_api.go::v2PreviewPayload] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"rendered_payload": payloadItem.TemplateRendered,
		},
	})
}

// v2BatchCreatePayloads creates multiple payloads
func (self *WebServer) v2BatchCreatePayloads(c *gin.Context) {
	var req struct {
		CaseID    string            `json:"case_id" binding:"required"`
		Template  string            `json:"template" binding:"required"`
		Count     int               `json:"count" binding:"required,min=1,max=100"`
		Variables map[string]string `json:"variables"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	caseId, err := strconv.ParseInt(req.CaseID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid case id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var payloads []models.TblPayload
	for i := 0; i < req.Count; i++ {
		token := genRandomString(8)
		renderedPayload := fmt.Sprintf("http://%s.%s", token, self.Domain)

		payload := models.TblPayload{
			CaseId:          caseId,
			Token:           token,
			Template:        req.Template,
			RenderedPayload: renderedPayload,
			Status:          "draft",
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		}

		payloads = append(payloads, payload)
	}

	_, err = session.Insert(&payloads)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2BatchCreatePayloads] insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": payloads,
			"count": len(payloads),
		},
	})
}

// v2ListInteractions lists interactions
func (self *WebServer) v2ListInteractions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	interactionType := c.Query("type")

	session := self.orm.NewSession()
	defer session.Close()

	var interactions []v2models.Interaction
	query := session.Table(new(v2models.Interaction))

	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if payloadId != "" {
		query = query.Where("payload_id = ?", payloadId)
	}
	if interactionType != "" {
		query = query.Where("type = ?", interactionType)
	}

	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListInteractions] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	err = query.OrderBy("timestamp DESC").Limit(pageSize, (page-1)*pageSize).Find(&interactions)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListInteractions] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	items := make([]models.Interaction, len(interactions))
	for i, item := range interactions {
		headers := item.Headers
		token := ""
		if item.Token != nil {
			token = *item.Token
		}
		domain := ""
		if item.Domain != nil {
			domain = *item.Domain
		}
		dnsType := ""
		if item.DNSType != nil {
			dnsType = *item.DNSType
		}
		method := ""
		if item.Method != nil {
			method = *item.Method
		}
		path := ""
		if item.Path != nil {
			path = *item.Path
		}
		body := ""
		if item.Body != nil {
			body = *item.Body
		}
		userAgent := ""
		if item.UserAgent != nil {
			userAgent = *item.UserAgent
		}
		contentType := ""
		if item.ContentType != nil {
			contentType = *item.ContentType
		}

		caseId := ""
		if item.CaseID != nil {
			caseId = *item.CaseID
		}
		payloadId := ""
		if item.PayloadID != nil {
			payloadId = *item.PayloadID
		}

		items[i] = models.Interaction{
			Id:          item.ID,
			Type:        item.Type,
			CaseId:      caseId,
			PayloadId:   payloadId,
			Token:       token,
			Timestamp:   item.Timestamp.Format(time.RFC3339),
			SourceIp:    item.SourceIP,
			Domain:      domain,
			DnsType:     dnsType,
			Method:      method,
			Path:        path,
			Headers:     headers,
			Body:        body,
			UserAgent:   userAgent,
			ContentType: contentType,
			RawData:     item.RawData,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.InteractionListResponse{
			Items:      items,
			Total:      int(total),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// v2ListPayloadInteractions lists interactions associated with a specific payload
func (self *WebServer) v2ListPayloadInteractions(c *gin.Context) {
	payloadId := c.Param("id")
	if payloadId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "payload id is required"})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	session := self.orm.NewSession()
	defer session.Close()

	var interactions []v2models.Interaction
	query := session.Table(new(v2models.Interaction)).Where("payload_id = ?", payloadId)

	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListPayloadInteractions] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}

	err = query.OrderBy("timestamp DESC").Limit(pageSize, (page-1)*pageSize).Find(&interactions)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListPayloadInteractions] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}

	items := make([]models.Interaction, len(interactions))
	for i, item := range interactions {
		token := ""
		if item.Token != nil {
			token = *item.Token
		}
		domain := ""
		if item.Domain != nil {
			domain = *item.Domain
		}
		dnsType := ""
		if item.DNSType != nil {
			dnsType = *item.DNSType
		}
		method := ""
		if item.Method != nil {
			method = *item.Method
		}
		path := ""
		if item.Path != nil {
			path = *item.Path
		}
		body := ""
		if item.Body != nil {
			body = *item.Body
		}
		userAgent := ""
		if item.UserAgent != nil {
			userAgent = *item.UserAgent
		}
		contentType := ""
		if item.ContentType != nil {
			contentType = *item.ContentType
		}
		caseId := ""
		if item.CaseID != nil {
			caseId = *item.CaseID
		}

		items[i] = models.Interaction{
			Id:          item.ID,
			Type:        item.Type,
			CaseId:      caseId,
			PayloadId:   payloadId,
			Token:       token,
			Timestamp:   item.Timestamp.Format(time.RFC3339),
			SourceIp:    item.SourceIP,
			Domain:      domain,
			DnsType:     dnsType,
			Method:      method,
			Path:        path,
			Headers:     item.Headers,
			Body:        body,
			UserAgent:   userAgent,
			ContentType: contentType,
			RawData:     item.RawData,
			CreatedAt:   item.CreatedAt.Format(time.RFC3339),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.InteractionListResponse{
			Items:      items,
			Total:      int(total),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// v2GetInteraction gets an interaction
func (self *WebServer) v2GetInteraction(c *gin.Context) {
	id := c.Param("id")

	session := self.orm.NewSession()
	defer session.Close()

	var interactionItem v2models.Interaction
	has, err := session.Where("id = ?", id).Get(&interactionItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetInteraction] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "interaction not found",
		})
		return
	}

	headers := interactionItem.Headers
	token := ""
	if interactionItem.Token != nil {
		token = *interactionItem.Token
	}
	domain := ""
	if interactionItem.Domain != nil {
		domain = *interactionItem.Domain
	}
	dnsType := ""
	if interactionItem.DNSType != nil {
		dnsType = *interactionItem.DNSType
	}
	method := ""
	if interactionItem.Method != nil {
		method = *interactionItem.Method
	}
	path := ""
	if interactionItem.Path != nil {
		path = *interactionItem.Path
	}
	body := ""
	if interactionItem.Body != nil {
		body = *interactionItem.Body
	}
	userAgent := ""
	if interactionItem.UserAgent != nil {
		userAgent = *interactionItem.UserAgent
	}
	contentType := ""
	if interactionItem.ContentType != nil {
		contentType = *interactionItem.ContentType
	}

	caseId := ""
	if interactionItem.CaseID != nil {
		caseId = *interactionItem.CaseID
	}
	payloadId := ""
	if interactionItem.PayloadID != nil {
		payloadId = *interactionItem.PayloadID
	}

	result := models.Interaction{
		Id:          interactionItem.ID,
		Type:        interactionItem.Type,
		CaseId:      caseId,
		PayloadId:   payloadId,
		Token:       token,
		Timestamp:   interactionItem.Timestamp.Format(time.RFC3339),
		SourceIp:    interactionItem.SourceIP,
		Domain:      domain,
		DnsType:     dnsType,
		Method:      method,
		Path:        path,
		Headers:     headers,
		Body:        body,
		UserAgent:   userAgent,
		ContentType: contentType,
		RawData:     interactionItem.RawData,
		CreatedAt:   interactionItem.CreatedAt.Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// v2DeleteInteractions deletes interactions
func (self *WebServer) v2DeleteInteractions(c *gin.Context) {
	var req struct {
		Ids []string `json:"ids"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	for _, id := range req.Ids {
		session.ID(id).Delete(new(v2models.Interaction))
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ExportInteractions exports interactions
func (self *WebServer) v2ExportInteractions(c *gin.Context) {
	var req struct {
		Ids    []string `json:"ids"`
		Format string   `json:"format"` // json, markdown
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"export_url": "/api/v2/interactions/export/" + req.Format,
		},
	})
}

// v2ListAPIKeys lists API keys
func (self *WebServer) v2ListAPIKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	// Use auth service to list API keys
	response, err := self.authService.ListAPIKeys(userID, page, pageSize)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListAPIKeys] list error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Mask keys in response (never return full key in list)
	for i := range response.Items {
		response.Items[i].Key = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    response,
	})
}

// v2CreateAPIKey creates an API key
func (self *WebServer) v2CreateAPIKey(c *gin.Context) {
	var req v2models.APIKeyCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "name is required",
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	// Use auth service to create API key
	apiKey, err := self.authService.CreateAPIKey(&req, userID)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateAPIKey] create error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	// Write audit log
	userIDPtr := &userID
	resourceIDPtr := &apiKey.ID
	auditLog := &v2models.AuditLog{
		ID:           generateRandomString(36),
		UserID:       userIDPtr,
		Action:       "api_key.created",
		ResourceType: "api_key",
		ResourceID:   resourceIDPtr,
		Details: v2models.AuditDetails{
			"api_key_id":     apiKey.ID,
			"key_prefix":     apiKey.KeyPrefix,
			"is_agent":       apiKey.IsAgent,
			"scopes":         apiKey.Scopes,
			"risk_tolerance": apiKey.RiskTolerance,
		},
		Timestamp: time.Now(),
	}
	if err := self.authService.CreateAuditLog(auditLog); err != nil {
		logrus.Errorf("[v2_api.go::v2CreateAPIKey] audit log error: %v", err)
	}

	// Return API key with full key (only shown on creation)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    apiKey,
	})
}

// v2DeleteAPIKey deletes an API key
func (self *WebServer) v2DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")

	// Get API key for audit log before revoking
	apiKey, err := self.authService.GetAPIKeyByID(id)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteAPIKey] get error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "api key not found",
		})
		return
	}

	// Revoke API key
	if err := self.authService.RevokeAPIKey(id); err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteAPIKey] revoke error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Write audit log
	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)
	userIDPtr := &userID
	resourceIDPtr := &id
	auditLog := &v2models.AuditLog{
		ID:           generateRandomString(36),
		UserID:       userIDPtr,
		Action:       "api_key.revoked",
		ResourceType: "api_key",
		ResourceID:   resourceIDPtr,
		Details: v2models.AuditDetails{
			"api_key_id":     apiKey.ID,
			"key_prefix":     apiKey.KeyPrefix,
			"is_agent":       apiKey.IsAgent,
			"scopes":         apiKey.Scopes,
			"risk_tolerance": apiKey.RiskTolerance,
		},
		Timestamp: time.Now(),
	}
	if err := self.authService.CreateAuditLog(auditLog); err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteAPIKey] audit log error: %v", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2GetAPIKey gets an API key by ID
func (self *WebServer) v2GetAPIKey(c *gin.Context) {
	id := c.Param("id")

	// Use auth service to get API key
	apiKey, err := self.authService.GetAPIKeyByID(id)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetAPIKey] get error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "api key not found",
		})
		return
	}

	// Mask the key for security (never return full key in get)
	apiKey.Key = ""

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    apiKey,
	})
}

// v2UpdateAPIKey updates an API key
func (self *WebServer) v2UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")

	var req v2models.APIKeyCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	// Get existing API key
	apiKey, err := self.authService.GetAPIKeyByID(id)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateAPIKey] get error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "api key not found",
		})
		return
	}

	// Update fields if provided
	if req.Name != "" {
		apiKey.Name = req.Name
	}
	if req.Scopes != nil {
		// Validate scopes
		for _, scope := range req.Scopes {
			if !auth.ValidScopes[scope] {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "invalid scope",
				})
				return
			}
		}
		// Validate agent scopes if this is an agent key
		if apiKey.IsAgent {
			if !v2models.ValidateAgentScopes(req.Scopes) {
				c.JSON(http.StatusBadRequest, gin.H{
					"code":    400,
					"message": "invalid agent scope",
				})
				return
			}
		}
		apiKey.Scopes = v2models.Scopes(req.Scopes)
	}
	if req.ExpiresAt != nil {
		apiKey.ExpiresAt = req.ExpiresAt
	}

	// Update in database
	session := self.orm.NewSession()
	defer session.Close()
	if _, err := session.ID(id).Update(apiKey); err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateAPIKey] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListAgentPolicyScopes lists the shared Agent scope and risk catalog.
func (self *WebServer) v2ListAgentPolicyScopes(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    agentpolicy.ListScopes(),
	})
}

func generateAPIKey() string {
	return "gdl_" + generateRandomString(32)
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

// v2ListUsers lists users (admin only)
func (self *WebServer) v2ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	// TblUser uses Atime/Utime (DB columns atime/utime), not created_at — avoid invalid ORDER BY.
	total, err := self.orm.Count(&models.TblUser{})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListUsers] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	var users []models.TblUser
	err = self.orm.Desc("id").Limit(pageSize, (page-1)*pageSize).Find(&users)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListUsers] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	items := make([]map[string]interface{}, len(users))
	for i, user := range users {
		created := user.Atime
		if created.IsZero() {
			created = user.Utime
		}
		items[i] = map[string]interface{}{
			"id":         strconv.FormatInt(user.Id, 10),
			"username":   user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": created.Format(time.RFC3339),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       items,
			"total":       int(total),
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// v2InteractionStats gets interaction statistics
func (self *WebServer) v2InteractionStats(c *gin.Context) {
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	period := c.Query("period")

	session := self.orm.NewSession()
	defer session.Close()

	query := session.Table(new(v2models.Interaction))

	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if payloadId != "" {
		query = query.Where("payload_id = ?", payloadId)
	}
	if period == "today" {
		t := time.Now().UTC()
		start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
		query = query.Where("timestamp >= ?", start)
	}

	// Count total interactions
	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2InteractionStats] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Count by type
	type InteractionTypeStats struct {
		Type  string `xorm:"type"`
		Count int64  `xorm:"count"`
	}
	var typeStats []InteractionTypeStats
	err = query.GroupBy("type").Select("type, count(*) as count").Find(&typeStats)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2InteractionStats] group by type error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Convert to map
	typeCountMap := make(map[string]int64)
	for _, stat := range typeStats {
		typeCountMap[stat.Type] = stat.Count
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total":      total,
			"by_type":    typeCountMap,
			"dns_count":  typeCountMap["dns"],
			"http_count": typeCountMap["http"],
			"smtp_count": typeCountMap["smtp"],
			"ldap_count": typeCountMap["ldap"],
		},
	})
}

// v2InteractionDailyStats returns daily interaction counts for the last N days.
func (self *WebServer) v2InteractionDailyStats(c *gin.Context) {
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	days := 7
	if d, err := strconv.Atoi(c.Query("days")); err == nil && d > 0 && d <= 90 {
		days = d
	}

	now := time.Now().UTC()
	start := time.Date(now.Year(), now.Month(), now.Day()-days+1, 0, 0, 0, 0, time.UTC)

	session := self.orm.NewSession()
	defer session.Close()
	query := session.Table(new(v2models.Interaction)).Where("timestamp >= ?", start)
	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if payloadId != "" {
		query = query.Where("payload_id = ?", payloadId)
	}

	type dailyStat struct {
		Date  string `xorm:"date" json:"date"`
		Count int64  `xorm:"count" json:"count"`
	}
	var rows []dailyStat
	if err := query.Select("DATE(timestamp) as date, count(*) as count").
		GroupBy("DATE(timestamp)").OrderBy("date ASC").Find(&rows); err != nil {
		logrus.Errorf("[v2_api.go::v2InteractionDailyStats] query error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}

	// Fill missing days with zero counts for a continuous chart.
	countMap := make(map[string]int64, len(rows))
	for _, r := range rows {
		countMap[r.Date] = r.Count
	}
	result := make([]gin.H, 0, days)
	for i := 0; i < days; i++ {
		day := time.Date(now.Year(), now.Month(), now.Day()-days+1+i, 0, 0, 0, 0, time.UTC)
		dateStr := day.Format("2006-01-02")
		result = append(result, gin.H{"date": dateStr, "count": countMap[dateStr]})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// v2InteractionTimeline gets interaction timeline
func (self *WebServer) v2InteractionTimeline(c *gin.Context) {
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	interval := c.DefaultQuery("interval", "hour")

	session := self.orm.NewSession()
	defer session.Close()

	query := session.Table(new(v2models.Interaction))

	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if payloadId != "" {
		query = query.Where("payload_id = ?", payloadId)
	}

	var interactions []v2models.Interaction
	if err := query.OrderBy("timestamp ASC").Find(&interactions); err != nil {
		logrus.Errorf("[v2_api.go::v2InteractionTimeline] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Group by time interval
	groupedEvents := make(map[string][]v2models.Interaction)
	for _, interaction := range interactions {
		key := getIntervalKey(interaction.Timestamp, interval)
		groupedEvents[key] = append(groupedEvents[key], interaction)
	}

	// Convert to array
	var timelineGroups []gin.H
	for key, items := range groupedEvents {
		timelineGroups = append(timelineGroups, gin.H{
			"time":   key,
			"count":  len(items),
			"events": items,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"total":          len(interactions),
			"grouped_events": timelineGroups,
		},
	})
}

// getIntervalKey returns the time interval key for grouping
func getIntervalKey(t time.Time, interval string) string {
	switch interval {
	case "hour":
		return t.Format("2006-01-02 15:00")
	case "day":
		return t.Format("2006-01-02")
	case "week":
		year, week := t.ISOWeek()
		return fmt.Sprintf("%d-W%02d", year, week)
	case "month":
		return t.Format("2006-01")
	default:
		return t.Format("2006-01-02 15:04")
	}
}

// v2InteractionStream streams new interactions via Server-Sent Events (SSE).
// Query params:
//   - case_id: filter by case
//   - payload_id: filter by payload
//   - type: filter by interaction type
//
// The client sends a "since" query param (RFC3339 timestamp) to get interactions newer than that time.
// The server polls the database every 2 seconds and sends any new interactions as SSE "interaction" events.
// A "heartbeat" event is sent every 30 seconds to keep the connection alive.
func (self *WebServer) v2InteractionStream(c *gin.Context) {
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	interactionType := c.Query("type")

	// Parse "since" timestamp; default to now
	sinceStr := c.Query("since")
	var since time.Time
	if sinceStr != "" {
		if parsed, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			since = parsed
		} else {
			since = time.Now()
		}
	} else {
		since = time.Now()
	}

	// Set SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "streaming not supported"})
		return
	}

	// Send initial connected event
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"since\":\"%s\"}\n\n", since.Format(time.RFC3339))
	flusher.Flush()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	ctx := c.Request.Context()

	for {
		select {
		case <-ctx.Done():
			return
		case <-heartbeatTicker.C:
			fmt.Fprintf(c.Writer, "event: heartbeat\ndata: {\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
			flusher.Flush()
		case <-ticker.C:
			// Query for new interactions since last check
			session := self.orm.NewSession()
			if caseId != "" {
				session = session.Where("case_id = ?", caseId)
			}
			if payloadId != "" {
				session = session.Where("payload_id = ?", payloadId)
			}
			if interactionType != "" {
				session = session.Where("type = ?", interactionType)
			}

			var newInteractions []v2models.Interaction
			err := session.Where("timestamp > ?", since).OrderBy("timestamp ASC").Limit(100, 0).Find(&newInteractions)
			session.Close()

			if err != nil {
				logrus.Errorf("[v2_api.go::v2InteractionStream] query error: %v", err)
				continue
			}

			for _, interaction := range newInteractions {
				data, err := json.Marshal(interaction)
				if err != nil {
					continue
				}
				fmt.Fprintf(c.Writer, "event: interaction\ndata: %s\n\n", data)
				flusher.Flush()
				since = interaction.Timestamp
			}
		}
	}
}

// v2Poll implements Burp Collaborator-style cursor-based polling.
// Returns interactions created after the cursor timestamp.
// GET /api/v2/poll?cursor={ISO8601}&limit={n}
// @Summary Poll for new interactions (cursor-based)
// @Description Burp Collaborator-style cursor-based polling for new interactions since a given timestamp
// @Tags v2, interactions
// @Accept json
// @Produce json
// @Param cursor query string false "ISO8601 timestamp cursor, defaults to 24h ago"
// @Param limit query int false "Maximum results (1-100, default 20)"
// @Success 200 {object} gin.H "poll results with next_cursor and has_more"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/v2/poll [get]
func (self *WebServer) v2Poll(c *gin.Context) {
	cursorStr := c.Query("cursor")
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 20
	}

	var cursorTime time.Time
	if cursorStr != "" {
		cursorTime, err = time.Parse(time.RFC3339, cursorStr)
		if err != nil {
			cursorTime = time.Now().Add(-24 * time.Hour)
		}
	} else {
		cursorTime = time.Now().Add(-24 * time.Hour)
	}

	iaSvc := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	// Fetch one more than limit to detect has_more
	interactions, err := iaSvc.ListInteractions("", "", "", &cursorTime, nil, 1, limit+1)
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	hasMore := len(interactions.Items) > limit
	if hasMore {
		interactions.Items = interactions.Items[:limit]
	}

	nextCursor := time.Now().Format(time.RFC3339)
	if len(interactions.Items) > 0 {
		nextCursor = interactions.Items[len(interactions.Items)-1].Timestamp.Format(time.RFC3339)
	}

	c.JSON(200, gin.H{
		"code": 0,
		"data": gin.H{
			"interactions": interactions.Items,
			"next_cursor":  nextCursor,
			"has_more":     hasMore,
		},
	})
}

// v2ListPlugins lists marketplace plugins
func (self *WebServer) v2ListPlugins(c *gin.Context) {
	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	filters := marketplace.PluginFilters{
		Type:     c.Query("type"),
		Category: c.Query("category"),
	}
	if published := c.Query("is_published"); published != "" {
		isPub := published == "true" || published == "1"
		filters.IsPublished = &isPub
	}
	if official := c.Query("is_official"); official != "" {
		isOff := official == "true" || official == "1"
		filters.IsOfficial = &isOff
	}

	plugins, err := svc.ListPlugins(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list plugins"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": plugins,
			"total": len(plugins),
		},
	})
}

// v2CreatePlugin creates a new marketplace plugin
func (self *WebServer) v2CreatePlugin(c *gin.Context) {
	var plugin marketplace.Plugin
	if err := c.ShouldBindJSON(&plugin); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if plugin.ID == "" {
		plugin.ID = v2models.GenerateID()
	}
	plugin.IsPublished = true

	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)
	if err := svc.CreatePlugin(c, &plugin); err != nil {
		logrus.Errorf("[v2_api.go::v2CreatePlugin] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create plugin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plugin})
}

// v2GetPlugin gets a specific plugin
func (self *WebServer) v2GetPlugin(c *gin.Context) {
	id := c.Param("id")
	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	plugin, err := svc.GetPlugin(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Plugin not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": plugin})
}

// v2ListTemplates lists marketplace templates
func (self *WebServer) v2ListTemplates(c *gin.Context) {
	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	filters := marketplace.TemplateFilters{
		Type:     c.Query("type"),
		Category: c.Query("category"),
	}
	if published := c.Query("is_published"); published != "" {
		isPub := published == "true" || published == "1"
		filters.IsPublished = &isPub
	}
	if official := c.Query("is_official"); official != "" {
		isOff := official == "true" || official == "1"
		filters.IsOfficial = &isOff
	}

	templates, err := svc.ListTemplates(c, filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list templates"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": templates,
			"total": len(templates),
		},
	})
}

// v2CreateTemplate creates a new marketplace template
func (self *WebServer) v2CreateTemplate(c *gin.Context) {
	var template marketplace.Template
	if err := c.ShouldBindJSON(&template); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": fmt.Sprintf("invalid request: %v", err)})
		return
	}
	if template.ID == "" {
		template.ID = v2models.GenerateID()
	}
	template.IsPublished = true

	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)
	if err := svc.CreateTemplate(c, &template); err != nil {
		logrus.Errorf("[v2_api.go::v2CreateTemplate] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create template"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": template})
}

// v2GetTemplate gets a specific template
func (self *WebServer) v2GetTemplate(c *gin.Context) {
	id := c.Param("id")
	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	template, err := svc.GetTemplate(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Template not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": template})
}

// v2InstallPlugin installs a marketplace plugin
func (self *WebServer) v2InstallPlugin(c *gin.Context) {
	pluginID := c.Param("id")

	var req struct {
		Version string `json:"version"`
		Config  string `json:"config"`
	}
	// Body is optional; ignore decode errors
	_ = c.ShouldBindJSON(&req)

	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	// Verify plugin exists
	plugin, err := svc.GetPlugin(c, pluginID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Plugin not found"})
		return
	}

	version := req.Version
	if version == "" {
		version = plugin.Version
	}

	installation, err := svc.InstallPlugin(c, pluginID, version, req.Config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to install plugin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": installation})
}

// v2ListInstalledPlugins lists all installed plugins
func (self *WebServer) v2ListInstalledPlugins(c *gin.Context) {
	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	installations, err := svc.ListPluginInstallations(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list installed plugins"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"items": installations, "total": len(installations)}})
}

// v2UninstallPlugin uninstalls a plugin
func (self *WebServer) v2UninstallPlugin(c *gin.Context) {
	id := c.Param("id")

	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)

	if err := svc.UninstallPlugin(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to uninstall plugin"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// v2ListRules lists workflow rules
func (self *WebServer) v2ListRules(c *gin.Context) {
	page := 1
	pageSize := 20
	if p := c.Query("page"); p != "" {
		fmt.Sscanf(p, "%d", &page)
	}
	if s := c.Query("page_size"); s != "" {
		fmt.Sscanf(s, "%d", &pageSize)
	}

	workflowService := workflow.NewService(self.orm)
	resp, err := workflowService.ListWorkflows("", nil, page, pageSize)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListRules] ListWorkflows: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list workflows",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       resp.Items,
			"total":       resp.Total,
			"page":        resp.Page,
			"page_size":   resp.PageSize,
			"total_pages": resp.TotalPages,
		},
	})
}

// v2CreateRule creates a new workflow rule
func (self *WebServer) v2CreateRule(c *gin.Context) {
	var req v2models.Workflow
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	if req.CreatedBy == "" {
		if uid, ok := c.Get("id"); ok {
			req.CreatedBy = fmt.Sprintf("%v", uid)
		} else {
			req.CreatedBy = "0"
		}
	}

	workflowService := workflow.NewService(self.orm)
	if err := workflowService.CreateWorkflow(&req); err != nil {
		logrus.Errorf("[v2_api.go::v2CreateRule] CreateWorkflow: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create workflow",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2GetRule gets a specific rule
func (self *WebServer) v2GetRule(c *gin.Context) {
	id := c.Param("id")

	workflowService := workflow.NewService(self.orm)
	workflow, err := workflowService.GetWorkflowByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Workflow not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    workflow,
	})
}

// v2UpdateRule updates a rule
func (self *WebServer) v2UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var req v2models.Workflow
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	req.ID = id
	workflowService := workflow.NewService(self.orm)
	if err := workflowService.UpdateWorkflow(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update workflow",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2DeleteRule deletes a rule
func (self *WebServer) v2DeleteRule(c *gin.Context) {
	id := c.Param("id")

	workflowService := workflow.NewService(self.orm)
	if err := workflowService.DeleteWorkflow(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete workflow",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2GenerateEvidence generates evidence report
func (self *WebServer) v2GenerateEvidence(c *gin.Context) {
	var req struct {
		CaseID    string `json:"case_id"`
		PayloadID string `json:"payload_id"`
		Format    string `json:"format" binding:"required,oneof=json markdown"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body: either case_id or payload_id is required",
		})
		return
	}

	// Validate that at least one of case_id or payload_id is provided
	if len(req.CaseID) == 0 && len(req.PayloadID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Either case_id or payload_id is required",
		})
		return
	}

	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)

	resp, err := evidenceService.GenerateEvidence(req.CaseID, req.PayloadID, req.Format)
	if err != nil {
		if err == interaction.ErrEvidenceNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "No evidence found for the specified case or payload",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate evidence",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// v2GetEvidence gets evidence report by ID
func (self *WebServer) v2GetEvidence(c *gin.Context) {
	// Evidence reports are generated on-demand and not persisted
	// Use v2GenerateEvidence endpoint to generate evidence reports
	c.JSON(http.StatusNotFound, gin.H{
		"code":    404,
		"message": "Evidence reports are generated on-demand. Use /evidence/generate endpoint to create evidence reports.",
	})
}

// v2SummarizeEvidence returns an Agent-friendly evidence summary bundle.
func (self *WebServer) v2SummarizeEvidence(c *gin.Context) {
	var req evidencehub.SummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	baseURL := fmt.Sprintf("http://%s", c.Request.Host)
	service := evidencehub.NewService(self.orm)
	resp, err := service.BuildSummary(&req, baseURL)
	if err != nil {
		status := http.StatusInternalServerError
		code := 500
		message := "Failed to summarize evidence"
		if strings.Contains(err.Error(), "is required") {
			status = http.StatusBadRequest
			code = 1
			message = err.Error()
		} else if err == interaction.ErrEvidenceNotFound {
			status = http.StatusNotFound
			code = 404
			message = "No evidence found for the specified case, payload, or scanner run"
		} else if strings.Contains(err.Error(), "scanner run not found") {
			status = http.StatusNotFound
			code = 404
			message = err.Error()
		}
		c.JSON(status, gin.H{
			"code":    code,
			"message": message,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// v2ListCanaries lists canary tokens
func (self *WebServer) v2ListCanaries(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	canaryService := canary.NewService(self.orm)
	canaries, total, err := canaryService.ListCanaries(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list canaries",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": canaries,
			"total": total,
		},
	})
}

// v2CreateCanary creates a new canary token
func (self *WebServer) v2CreateCanary(c *gin.Context) {
	var req v2models.Canary
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	canaryService := canary.NewService(self.orm)
	if err := canaryService.CreateCanary(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create canary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2GetCanary gets a specific canary token
func (self *WebServer) v2GetCanary(c *gin.Context) {
	id := c.Param("id")

	canaryService := canary.NewService(self.orm)
	canary, err := canaryService.GetCanary(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Canary not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    canary,
	})
}

// v2UpdateCanary updates a canary token
func (self *WebServer) v2UpdateCanary(c *gin.Context) {
	id := c.Param("id")

	var req v2models.Canary
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	req.ID = id
	canaryService := canary.NewService(self.orm)
	if err := canaryService.UpdateCanary(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update canary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2DeleteCanary deletes a canary token
func (self *WebServer) v2DeleteCanary(c *gin.Context) {
	id := c.Param("id")

	canaryService := canary.NewService(self.orm)
	if err := canaryService.DeleteCanary(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete canary",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListCanaryHits lists hits for a canary token
func (self *WebServer) v2ListCanaryHits(c *gin.Context) {
	id := c.Param("id")

	canaryService := canary.NewService(self.orm)
	hits, err := canaryService.ListCanaryHits(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list canary hits",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"canary_id": id,
			"hits":      hits,
			"total":     len(hits),
		},
	})
}

// v2ListRebindingRules lists rebinding rules
func (self *WebServer) v2ListRebindingRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	rebindingService := rebinding.NewService(self.orm)
	rules, total, err := rebindingService.ListRebindingRules(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list rebinding rules",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": rules,
			"total": total,
		},
	})
}

// v2CreateRebindingRule creates a new rebinding rule
func (self *WebServer) v2CreateRebindingRule(c *gin.Context) {
	var req v2models.RebindingRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	rebindingService := rebinding.NewService(self.orm)
	if err := rebindingService.CreateRebindingRule(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create rebinding rule",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2GetRebindingRule gets a specific rebinding rule
func (self *WebServer) v2GetRebindingRule(c *gin.Context) {
	id := c.Param("id")

	rebindingService := rebinding.NewService(self.orm)
	rule, err := rebindingService.GetRebindingRule(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Rebinding rule not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    rule,
	})
}

// v2UpdateRebindingRule updates a rebinding rule
func (self *WebServer) v2UpdateRebindingRule(c *gin.Context) {
	id := c.Param("id")

	var req v2models.RebindingRule
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	req.ID = id
	rebindingService := rebinding.NewService(self.orm)
	if err := rebindingService.UpdateRebindingRule(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update rebinding rule",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2DeleteRebindingRule deletes a rebinding rule
func (self *WebServer) v2DeleteRebindingRule(c *gin.Context) {
	id := c.Param("id")

	rebindingService := rebinding.NewService(self.orm)
	if err := rebindingService.DeleteRebindingRule(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete rebinding rule",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListRebindingSessions lists sessions for a rebinding rule
func (self *WebServer) v2ListRebindingSessions(c *gin.Context) {
	id := c.Param("id")

	rebindingService := rebinding.NewService(self.orm)
	sessions, err := rebindingService.ListRebindingSessions(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list rebinding sessions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"rule_id":  id,
			"sessions": sessions,
			"total":    len(sessions),
		},
	})
}

// v2ListRebindingScenarios lists predefined rebinding scenarios
func (self *WebServer) v2ListRebindingScenarios(c *gin.Context) {
	scenarios := rebinding.GetPredefinedScenarios()
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    scenarios,
	})
}

// v2CreateRebindingFromScenario creates a rebinding rule from a predefined scenario
func (self *WebServer) v2CreateRebindingFromScenario(c *gin.Context) {
	scenarioName := c.Param("name")

	var req struct {
		Domain string `json:"domain" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "domain is required",
		})
		return
	}

	rebindingService := rebinding.NewService(self.orm)
	rule, err := rebindingService.CreateRuleFromScenario(scenarioName, req.Domain)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    rule,
	})
}

// v2ListListeners lists protocol listeners
func (self *WebServer) v2ListListeners(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

	listenerService := listener.NewService(self.orm)
	listeners, total, err := listenerService.ListListeners(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list listeners",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": listeners,
			"total": total,
		},
	})
}

// v2CreateListener creates a new protocol listener
func (self *WebServer) v2CreateListener(c *gin.Context) {
	var req v2models.Listener
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	listenerService := listener.NewService(self.orm)
	if err := listenerService.CreateListener(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create listener",
		})
		return
	}

	// Start the listener if it's enabled
	if req.IsEnabled && self.listenerMgr != nil {
		if err := self.listenerMgr.StartListener(&req); err != nil {
			logrus.Errorf("[v2_api.go::v2CreateListener] failed to start listener: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2GetListener gets a specific listener
func (self *WebServer) v2GetListener(c *gin.Context) {
	id := c.Param("id")

	listenerService := listener.NewService(self.orm)
	listener, err := listenerService.GetListener(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "Listener not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    listener,
	})
}

// v2UpdateListener updates a listener
func (self *WebServer) v2UpdateListener(c *gin.Context) {
	id := c.Param("id")

	var req v2models.Listener
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	req.ID = id
	listenerService := listener.NewService(self.orm)
	if err := listenerService.UpdateListener(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update listener",
		})
		return
	}

	// Restart or stop the listener via manager
	if self.listenerMgr != nil {
		if req.IsEnabled {
			if err := self.listenerMgr.RestartListener(&req); err != nil {
				logrus.Errorf("[v2_api.go::v2UpdateListener] failed to restart listener: %v", err)
			}
		} else {
			_ = self.listenerMgr.StopListener(id)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    req,
	})
}

// v2DeleteListener deletes a listener
func (self *WebServer) v2DeleteListener(c *gin.Context) {
	id := c.Param("id")

	listenerService := listener.NewService(self.orm)
	if err := listenerService.DeleteListener(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to delete listener",
		})
		return
	}

	// Stop the listener via manager
	if self.listenerMgr != nil {
		_ = self.listenerMgr.StopListener(id)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListListenerInteractions lists interactions for a listener
func (self *WebServer) v2ListListenerInteractions(c *gin.Context) {
	id := c.Param("id")

	listenerService := listener.NewService(self.orm)
	interactions, err := listenerService.ListListenerInteractions(id)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListListenerInteractions] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list listener interactions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"listener_id":  id,
			"interactions": interactions,
			"total":        len(interactions),
		},
	})
}

// v2ListNotificationChannels lists notification channels
func (self *WebServer) v2ListNotificationChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	notifyService := notification.NewService(self.orm)
	channels, total, err := notifyService.ListChannels(page, pageSize)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListNotificationChannels] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       channels,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// v2CreateNotificationChannel creates a notification channel
func (self *WebServer) v2CreateNotificationChannel(c *gin.Context) {
	var req models.NotificationChannelCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	notifyService := notification.NewService(self.orm)
	channel, err := notifyService.CreateChannel(req.Name, req.Type, req.Config, user.Id)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateNotificationChannel] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    channel,
	})
}

// v2GetNotificationChannel gets a notification channel
func (self *WebServer) v2GetNotificationChannel(c *gin.Context) {
	id := c.Param("id")
	channelId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid channel id",
		})
		return
	}

	notifyService := notification.NewService(self.orm)
	channel, err := notifyService.GetChannel(channelId)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetNotificationChannel] error: %v", err)
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "channel not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    channel,
	})
}

// v2UpdateNotificationChannel updates a notification channel
func (self *WebServer) v2UpdateNotificationChannel(c *gin.Context) {
	id := c.Param("id")
	channelId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid channel id",
		})
		return
	}

	var req models.NotificationChannelUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	notifyService := notification.NewService(self.orm)
	err = notifyService.UpdateChannel(channelId, req.Name, req.Config, req.Enabled)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateNotificationChannel] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2DeleteNotificationChannel deletes a notification channel
func (self *WebServer) v2DeleteNotificationChannel(c *gin.Context) {
	id := c.Param("id")
	channelId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid channel id",
		})
		return
	}

	notifyService := notification.NewService(self.orm)
	err = notifyService.DeleteChannel(channelId)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteNotificationChannel] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListNotificationLogs lists notification logs
func (self *WebServer) v2ListNotificationLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	channelIdStr := c.Query("channel_id")

	var channelId *int64
	if channelIdStr != "" {
		id, err := strconv.ParseInt(channelIdStr, 10, 64)
		if err == nil {
			channelId = &id
		}
	}

	notifyService := notification.NewService(self.orm)
	logs, total, err := notifyService.ListLogs(page, pageSize, channelId)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListNotificationLogs] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       logs,
			"total":       total,
			"page":        page,
			"page_size":   pageSize,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// v2ListSettings lists system settings
func (self *WebServer) v2ListSettings(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	var settings []v2models.Settings
	total, err := self.orm.Limit(pageSize, (page-1)*pageSize).FindAndCount(&settings)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListSettings] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": v2models.SettingsListResponse{
			Items:      settings,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// v2GetSetting gets a specific setting by key
func (self *WebServer) v2GetSetting(c *gin.Context) {
	key := c.Param("key")

	var setting v2models.Settings
	_, err := self.orm.Where("key = ?", key).Get(&setting)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetSetting] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	if setting.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "setting not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    setting,
	})
}

// v2UpdateSetting updates a setting
func (self *WebServer) v2UpdateSetting(c *gin.Context) {
	key := c.Param("key")

	var req v2models.SettingsUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var setting v2models.Settings
	_, err := session.Where("key = ?", key).Get(&setting)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateSetting] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	if setting.ID == "" {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "setting not found",
		})
		return
	}

	setting.Value = req.Value
	_, err = session.Cols("value").Update(&setting)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateSetting] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    setting,
	})
}

// v2CreateSetting creates a new setting
func (self *WebServer) v2CreateSetting(c *gin.Context) {
	var req v2models.SettingsCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	setting := v2models.Settings{
		ID:    v2models.GenerateID(),
		Key:   req.Key,
		Value: req.Value,
	}

	_, err := session.Insert(&setting)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateSetting] insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    setting,
	})
}

// v2DeleteSetting deletes a setting
func (self *WebServer) v2DeleteSetting(c *gin.Context) {
	key := c.Param("key")

	session := self.orm.NewSession()
	defer session.Close()

	_, err := session.Where("key = ?", key).Delete(&v2models.Settings{})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteSetting] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListAuditLogs lists audit logs with pagination and filtering
func (self *WebServer) v2ListAuditLogs(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "50"))
	userID := c.Query("user_id")
	action := c.Query("action")
	resourceType := c.Query("resource_type")
	resourceID := c.Query("resource_id")

	// Parse time range
	var startTime, endTime *time.Time
	if startTimeStr := c.Query("start_time"); startTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, startTimeStr); err == nil {
			startTime = &t
		}
	}
	if endTimeStr := c.Query("end_time"); endTimeStr != "" {
		if t, err := time.Parse(time.RFC3339, endTimeStr); err == nil {
			endTime = &t
		}
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 50
	}

	// RBAC: restrict audit log access by role
	role := c.GetInt("role")
	currentUserID := fmt.Sprintf("%d", c.GetInt64("id"))

	switch role {
	case roleSuper, roleAdmin:
		// Admin and super can view all logs, optional user_id filter applies
	case roleNormal:
		// Normal users can only see their own logs
		userID = currentUserID
	case roleGuest:
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "guest users cannot access audit logs",
		})
		return
	default:
		c.JSON(http.StatusForbidden, gin.H{
			"code":    403,
			"message": "insufficient permissions to access audit logs",
		})
		return
	}

	authService := auth.NewService(self.orm)
	resp, err := authService.ListAuditLogs(userID, action, resourceType, resourceID, startTime, endTime, page, pageSize)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListAuditLogs] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list audit logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// v2CreateAuditLog creates an audit log entry
func (self *WebServer) v2CreateAuditLog(c *gin.Context) {
	var req v2models.AuditLog
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	// Set timestamp if not provided
	if req.Timestamp.IsZero() {
		req.Timestamp = time.Now()
	}

	// Create audit log using auth service
	authService := auth.NewService(self.orm)
	if err := authService.CreateAuditLog(&req); err != nil {
		logrus.Errorf("[v2_api.go::v2CreateAuditLog] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create audit log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2ListScannerRuns lists scanner runs with filtering
func (self *WebServer) v2ListScannerRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	caseID := c.Query("case_id")
	payloadID := c.Query("payload_id")
	scanner := c.Query("scanner")
	status := c.Query("status")

	scannerHubService := scannerhub.NewService(self.orm)
	resp, err := scannerHubService.ListScannerRuns(caseID, payloadID, scanner, status, page, pageSize)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListScannerRuns] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list scanner runs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// v2ListScannerAdapters lists supported Scanner Hub adapters.
func (self *WebServer) v2ListScannerAdapters(c *gin.Context) {
	scannerHubService := scannerhub.NewService(self.orm)
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    scannerHubService.ListAdapters(),
	})
}

// v2CreateScannerRun creates a new scanner run
func (self *WebServer) v2CreateScannerRun(c *gin.Context) {
	var req v2models.ScannerRunCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	// Build base URL from request
	baseURL := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.Host)
	if baseURL == "://" {
		baseURL = "http://" + c.Request.Host
	}

	scannerHubService := scannerhub.NewService(self.orm)
	scannerRun, err := scannerHubService.CreateScannerRun(&req, userID, baseURL)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateScannerRun] error: %v", err)
		if err == scannerhub.ErrInvalidCase {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "case not found",
			})
			return
		}
		if err == scannerhub.ErrInvalidPayload {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "payload not found",
			})
			return
		}
		if err == scannerhub.ErrPayloadNotInCase {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "payload does not belong to case",
			})
			return
		}
		if err == scannerhub.ErrInvalidScanner || err == scannerhub.ErrInvalidDelivery {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    400,
				"message": "invalid scanner or delivery method",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create scanner run",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    scannerRun,
	})
}

// v2GetScannerRun gets a scanner run by ID with derived fields
func (self *WebServer) v2GetScannerRun(c *gin.Context) {
	id := c.Param("id")

	// Build base URL from request
	baseURL := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.Host)
	if baseURL == "://" {
		baseURL = "http://" + c.Request.Host
	}

	scannerHubService := scannerhub.NewService(self.orm)
	detail, err := scannerHubService.GetScannerRunDetail(id, baseURL)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetScannerRun] error: %v", err)
		if err == scannerhub.ErrScannerRunNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "scanner run not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get scanner run",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    detail,
	})
}

// v2UpdateScannerRunStatus updates the status of a scanner run
func (self *WebServer) v2UpdateScannerRunStatus(c *gin.Context) {
	id := c.Param("id")

	var req v2models.ScannerRunUpdateStatusRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	scannerHubService := scannerhub.NewService(self.orm)
	err := scannerHubService.UpdateScannerRunStatus(id, &req, userID)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateScannerRunStatus] error: %v", err)
		if err == scannerhub.ErrScannerRunNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "scanner run not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update scanner run status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2CreateScannerRunFromSearch creates scanner runs from search engine results.
func (self *WebServer) v2CreateScannerRunFromSearch(c *gin.Context) {
	var req v2models.ScannerRunCreateFromSearchRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	baseURL := fmt.Sprintf("%s://%s", c.Request.URL.Scheme, c.Request.Host)
	if baseURL == "://" {
		baseURL = "http://" + c.Request.Host
	}

	scannerHubService := scannerhub.NewService(self.orm)
	runs, err := scannerHubService.CreateScannerRunsFromSearch(&req, userID, baseURL)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateScannerRunFromSearch] error: %v", err)
		if err == scannerhub.ErrInvalidCase {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "case not found"})
			return
		}
		if err == scannerhub.ErrInvalidPayload || err == scannerhub.ErrPayloadNotInCase {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid payload"})
			return
		}
		if err == scannerhub.ErrInvalidScanner || err == scannerhub.ErrInvalidDelivery {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid scanner or delivery method"})
			return
		}
		if err == scannerhub.ErrNoResults {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "no results provided"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create scanner runs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": runs,
			"total": len(runs),
		},
	})
}

// v2BackfillScannerResults imports scan results and associates them with a scanner run
func (self *WebServer) v2BackfillScannerResults(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Format     string `json:"format"`
		RawResults string `json:"raw_results"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	if req.Format == "" {
		req.Format = "jsonl"
	}

	scannerHubService := scannerhub.NewService(self.orm)
	result, err := scannerHubService.BackfillResults(&scannerhub.BackfillResultsRequest{
		Format:       req.Format,
		RawResults:   req.RawResults,
		ScannerRunID: id,
	})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2BackfillScannerResults] error: %v", err)
		if err == scannerhub.ErrScannerRunNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "scanner run not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to backfill scanner results",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// v2ListAgentRuns lists agent runs with filtering
func (self *WebServer) v2ListAgentRuns(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	agentID := c.Query("agent_id")
	caseID := c.Query("case_id")
	payloadID := c.Query("payload_id")
	status := c.Query("status")

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	req := &v2models.AgentRunListRequest{
		AgentID:   agentID,
		CaseID:    caseID,
		PayloadID: payloadID,
		Status:    status,
		Page:      page,
		PageSize:  pageSize,
	}

	resp, err := agentRunService.ListAgentRuns(req)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListAgentRuns] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list agent runs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// v2CreateAgentRun creates a new agent run
func (self *WebServer) v2CreateAgentRun(c *gin.Context) {
	var req v2models.AgentRunCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	agentRun, err := agentRunService.CreateAgentRun(&req, userID)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateAgentRun] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to create agent run",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    agentRun,
	})
}

// v2GetAgentRun retrieves an agent run by ID
func (self *WebServer) v2GetAgentRun(c *gin.Context) {
	id := c.Param("id")

	baseURL := ""
	if c.Request.TLS != nil {
		baseURL = "https://" + c.Request.Host
	} else {
		baseURL = "http://" + c.Request.Host
	}

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	detail, err := agentRunService.GetAgentRunDetail(id, baseURL)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetAgentRun] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to get agent run",
		})
		return
	}

	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "agent run not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    detail,
	})
}

// v2GetAgentRunReview generates a review packet for an agent run
func (self *WebServer) v2GetAgentRunReview(c *gin.Context) {
	id := c.Param("id")
	format := c.DefaultQuery("format", "json")

	// Validate format
	if format != "json" && format != "markdown" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid format, must be 'json' or 'markdown'",
		})
		return
	}

	baseURL := ""
	if c.Request.TLS != nil {
		baseURL = "https://" + c.Request.Host
	} else {
		baseURL = "http://" + c.Request.Host
	}

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := agentrun.NewReviewService(self.orm, agentRunService, authService, evidenceService, interactionService)

	packet, err := reviewService.BuildReviewPacket(id, format, baseURL)
	if err != nil {
		if err == agentrun.ErrAgentRunNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Agent run not found",
			})
			return
		}
		logrus.Errorf("[v2_api.go::v2GetAgentRunReview] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to generate review packet",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    packet,
	})
}

// v2UpdateAgentRunStatus updates the status of an agent run
func (self *WebServer) v2UpdateAgentRunStatus(c *gin.Context) {
	id := c.Param("id")

	var req v2models.AgentRunUpdateStatusRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	err := agentRunService.UpdateAgentRunStatus(id, &req, userID)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateAgentRunStatus] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to update agent run status",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2AppendAgentOperation appends an operation to an agent run
func (self *WebServer) v2AppendAgentOperation(c *gin.Context) {
	id := c.Param("id")

	var req v2models.AgentOperationCreateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": fmt.Sprintf("invalid request: %v", err),
		})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	err := agentRunService.AppendAgentOperation(id, &req, userID)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2AppendAgentOperation] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to append agent operation",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// v2CreateAgentRunFollowup creates a follow-up action for an agent run
func (self *WebServer) v2CreateAgentRunFollowup(c *gin.Context) {
	id := c.Param("id")
	var req v2models.AgentRunFollowupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	resp, err := agentRunService.CreateFollowupAction(id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid followup") ||
			strings.Contains(err.Error(), "reason") ||
			strings.Contains(err.Error(), "request is required") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		logrus.Errorf("[v2_api.go::v2CreateAgentRunFollowup] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create followup"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2RecordReviewDecision records a review decision for an agent run
func (self *WebServer) v2RecordReviewDecision(c *gin.Context) {
	id := c.Param("id")
	var req v2models.AgentRunReviewDecisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	resp, err := agentRunService.RecordReviewDecision(id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid decision") ||
			strings.Contains(err.Error(), "reason too long") ||
			strings.Contains(err.Error(), "review_packet_id") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		logrus.Errorf("[v2_api.go::v2RecordReviewDecision] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to record review decision"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2ExportReviewPackage exports a review evidence package for an agent run
func (self *WebServer) v2ExportReviewPackage(c *gin.Context) {
	id := c.Param("id")
	var req v2models.AgentRunReviewExportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := agentrun.NewReviewService(self.orm, agentRunService, authService, evidenceService, interactionService)

	resp, err := reviewService.ExportReviewPackage(id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid format") ||
			strings.Contains(err.Error(), "review_packet_id") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		logrus.Errorf("[v2_api.go::v2ExportReviewPackage] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to export review package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2DeliverReviewPackage delivers a review evidence package to a webhook
func (self *WebServer) v2DeliverReviewPackage(c *gin.Context) {
	id := c.Param("id")
	var req v2models.AgentRunReviewDeliveryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := agentrun.NewReviewService(self.orm, agentRunService, authService, evidenceService, interactionService)

	resp, err := reviewService.DeliverReviewPackage(id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		if strings.Contains(err.Error(), "invalid format") ||
			strings.Contains(err.Error(), "invalid webhook URL") ||
			strings.Contains(err.Error(), "invalid headers") ||
			strings.Contains(err.Error(), "review_packet_id") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "timed out") {
			c.JSON(http.StatusGatewayTimeout, gin.H{"code": 504, "message": "Webhook request timed out"})
			return
		}
		if strings.Contains(err.Error(), "delivery failed") {
			c.JSON(http.StatusBadGateway, gin.H{"code": 502, "message": err.Error()})
			return
		}
		logrus.Errorf("[v2_api.go::v2DeliverReviewPackage] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to deliver review package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2ListReviewDeliveries lists the delivery history for an agent run
func (self *WebServer) v2ListReviewDeliveries(c *gin.Context) {
	id := c.Param("id")

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := agentrun.NewReviewService(self.orm, agentRunService, authService, evidenceService, interactionService)

	resp, err := reviewService.ListReviewDeliveries(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		logrus.Errorf("[v2_api.go::v2ListReviewDeliveries] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list delivery history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2TraceReviewPackage traces a review package by its hash
func (self *WebServer) v2TraceReviewPackage(c *gin.Context) {
	packageHash := c.Query("package_hash")

	// Validate package_hash parameter
	if packageHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "package_hash is required"})
		return
	}

	// Validate package_hash format
	if err := agentrun.ValidatePackageHash(packageHash); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid package_hash: must be 64-character hex string"})
		return
	}

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := agentrun.NewReviewService(self.orm, agentRunService, authService, evidenceService, interactionService)

	resp, err := reviewService.TraceReviewPackageByHash(packageHash)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2TraceReviewPackage] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to trace package"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2ListReviewQueue lists the review queue with filters
func (self *WebServer) v2ListReviewQueue(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	reviewState := c.Query("review_state")
	status := c.Query("status")
	evidenceStrength := c.Query("evidence_strength")
	agentID := c.Query("agent_id")
	caseID := c.Query("case_id")
	payloadID := c.Query("payload_id")

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	filters := agentrun.ReviewQueueFilters{
		ReviewState:      reviewState,
		Status:           status,
		EvidenceStrength: evidenceStrength,
		AgentID:          agentID,
		CaseID:           caseID,
		PayloadID:        payloadID,
		Page:             page,
		PageSize:         pageSize,
	}

	resp, err := agentRunService.ListReviewQueue(filters)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListReviewQueue] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list review queue",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    resp,
	})
}

// v2CompleteAgentRun orchestrates the full agent run completion loop
func (self *WebServer) v2CompleteAgentRun(c *gin.Context) {
	id := c.Param("id")

	var req agentrun.CompleteAgentRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "Invalid request body"})
		return
	}

	user := c.MustGet("user").(*models.TblUser)
	userID := strconv.FormatInt(user.Id, 10)

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)
	interactionService := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	evidenceService := interaction.NewEvidenceService(interactionService)
	reviewService := agentrun.NewReviewService(self.orm, agentRunService, authService, evidenceService, interactionService)

	resp, err := reviewService.CompleteAgentRun(id, &req, userID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Agent run not found"})
			return
		}
		if strings.Contains(err.Error(), "already in terminal status") {
			c.JSON(http.StatusConflict, gin.H{"code": 409, "message": err.Error()})
			return
		}
		if strings.Contains(err.Error(), "invalid format") || strings.Contains(err.Error(), "invalid decision") {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
			return
		}
		logrus.Errorf("[v2_api.go::v2CompleteAgentRun] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to complete agent run"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": resp})
}

// v2ListDNSRecords lists DNS resolve records with pagination
func (self *WebServer) v2ListDNSRecords(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	total, err := self.orm.Count(&models.TblResolve{})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListDNSRecords] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}

	var records []models.TblResolve
	err = self.orm.Desc("id").Limit(pageSize, (page-1)*pageSize).Find(&records)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListDNSRecords] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}

	items := make([]map[string]interface{}, len(records))
	for i, r := range records {
		items[i] = map[string]interface{}{
			"id":         strconv.FormatInt(r.Id, 10),
			"host":       r.Host,
			"type":       r.Type,
			"value":      r.Value,
			"ttl":        r.Ttl,
			"created_at": r.Ctime.Format(time.RFC3339),
			"updated_at": r.Utime.Format(time.RFC3339),
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":       items,
			"total":       int(total),
			"page":        page,
			"page_size":   pageSize,
			"total_pages": totalPages,
		},
	})
}

// v2CreateDNSRecord creates a new DNS resolve record
func (self *WebServer) v2CreateDNSRecord(c *gin.Context) {
	var req struct {
		Host  string `json:"host" binding:"required"`
		Type  string `json:"type" binding:"required"`
		Value string `json:"value" binding:"required"`
		Ttl   uint32 `json:"ttl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}
	if req.Ttl == 0 {
		req.Ttl = 300
	}

	record := models.TblResolve{
		Host:  req.Host,
		Type:  req.Type,
		Value: req.Value,
		Ttl:   req.Ttl,
	}

	_, err := self.orm.Insert(&record)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateDNSRecord] insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to create DNS record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":         strconv.FormatInt(record.Id, 10),
			"host":       record.Host,
			"type":       record.Type,
			"value":      record.Value,
			"ttl":        record.Ttl,
			"created_at": record.Ctime.Format(time.RFC3339),
			"updated_at": record.Utime.Format(time.RFC3339),
		},
	})
}

// v2UpdateDNSRecord updates an existing DNS resolve record
func (self *WebServer) v2UpdateDNSRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var record models.TblResolve
	has, err := self.orm.ID(id).Get(&record)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateDNSRecord] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "DNS record not found"})
		return
	}

	var req struct {
		Host  *string `json:"host"`
		Type  *string `json:"type"`
		Value *string `json:"value"`
		Ttl   *uint32 `json:"ttl"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	if req.Host != nil {
		record.Host = *req.Host
	}
	if req.Type != nil {
		record.Type = *req.Type
	}
	if req.Value != nil {
		record.Value = *req.Value
	}
	if req.Ttl != nil {
		record.Ttl = *req.Ttl
	}

	_, err = self.orm.ID(id).Update(&record)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateDNSRecord] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to update DNS record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":         strconv.FormatInt(record.Id, 10),
			"host":       record.Host,
			"type":       record.Type,
			"value":      record.Value,
			"ttl":        record.Ttl,
			"created_at": record.Ctime.Format(time.RFC3339),
			"updated_at": record.Utime.Format(time.RFC3339),
		},
	})
}

// v2DeleteDNSRecord deletes a DNS resolve record
func (self *WebServer) v2DeleteDNSRecord(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var record models.TblResolve
	has, err := self.orm.ID(id).Get(&record)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteDNSRecord] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "DNS record not found"})
		return
	}

	_, err = self.orm.ID(id).Delete(&models.TblResolve{})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteDNSRecord] delete error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to delete DNS record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// v2QueryXip returns xip encoding formats for a given IPv4 address
func (self *WebServer) v2QueryXip(c *gin.Context) {
	ipStr := c.Param("ip")
	ip := net.ParseIP(ipStr)
	if ip == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid IP address"})
		return
	}
	ip4 := ip.To4()
	if ip4 == nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "only IPv4 addresses are supported"})
		return
	}

	dotted := fmt.Sprintf("%d.%d.%d.%d", ip4[0], ip4[1], ip4[2], ip4[3])
	hex := fmt.Sprintf("%02x%02x%02x%02x", ip4[0], ip4[1], ip4[2], ip4[3])
	binary := fmt.Sprintf("0b%08b%08b%08b%08b", ip4[0], ip4[1], ip4[2], ip4[3])

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"dotted": dotted,
			"hex":    hex,
			"binary": binary,
			"examples": gin.H{
				"dotted_decimal": dotted + ".example.com",
				"hex":            hex + ".example.com",
				"binary":         binary + ".example.com",
			},
		},
	})
}

// v2CreateUser creates a new user (admin only)
func (self *WebServer) v2CreateUser(c *gin.Context) {
	role := c.GetInt("role")
	if role != roleSuper && role != roleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "admin access required"})
		return
	}

	var req struct {
		Username string `json:"username" binding:"required"`
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
		Role     int    `json:"role"`
		Lang     string `json:"lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}
	if len(req.Password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "password must be at least 8 characters"})
		return
	}
	if req.Role == roleSuper {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "cannot create super user"})
		return
	}

	// Check for duplicate username
	existing, _ := self.orm.Where("name = ?", req.Username).Exist(&models.TblUser{})
	if existing {
		c.JSON(http.StatusConflict, gin.H{"code": 409, "message": "username already exists"})
		return
	}

	lang := req.Lang
	if lang == "" {
		lang = self.DefaultLanguage
	}

	user := models.TblUser{
		Name:          req.Username,
		Email:         req.Email,
		Role:          req.Role,
		Token:         genRandomToken(),
		ShortId:       genShortId(),
		Lang:          lang,
		Pass:          makePassword(req.Password),
		CleanInterval: self.DefaultCleanInterval,
	}

	_, err := self.orm.Insert(&user)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateUser] insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to create user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":         strconv.FormatInt(user.Id, 10),
			"username":   user.Name,
			"email":      user.Email,
			"role":       user.Role,
			"created_at": user.Atime.Format(time.RFC3339),
		},
	})
}

// v2UpdateUser updates an existing user
func (self *WebServer) v2UpdateUser(c *gin.Context) {
	role := c.GetInt("role")
	if role != roleSuper && role != roleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "admin access required"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var user models.TblUser
	has, err := self.orm.ID(id).Get(&user)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateUser] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "user not found"})
		return
	}

	// Prevent demoting super user
	if user.Role == roleSuper {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "cannot modify super user"})
		return
	}

	var req struct {
		Email    *string `json:"email"`
		Password *string `json:"password"`
		Role     *int    `json:"role"`
		Lang     *string `json:"lang"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid request: " + err.Error()})
		return
	}

	if req.Email != nil {
		user.Email = *req.Email
	}
	if req.Role != nil {
		if *req.Role == roleSuper {
			c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "cannot promote to super user"})
			return
		}
		user.Role = *req.Role
	}
	if req.Lang != nil {
		user.Lang = *req.Lang
	}
	if req.Password != nil {
		if len(*req.Password) < 8 {
			c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "password must be at least 8 characters"})
			return
		}
		user.Pass = makePassword(*req.Password)
	}

	_, err = self.orm.ID(id).Update(&user)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2UpdateUser] update error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to update user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"id":       strconv.FormatInt(user.Id, 10),
			"username": user.Name,
			"email":    user.Email,
			"role":     user.Role,
		},
	})
}

// v2DeleteUser deletes a user
func (self *WebServer) v2DeleteUser(c *gin.Context) {
	role := c.GetInt("role")
	if role != roleSuper && role != roleAdmin {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "admin access required"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": "invalid id"})
		return
	}

	var user models.TblUser
	has, err := self.orm.ID(id).Get(&user)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteUser] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "server internal error"})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "user not found"})
		return
	}

	// Prevent deleting super user
	if user.Role == roleSuper {
		c.JSON(http.StatusForbidden, gin.H{"code": 403, "message": "cannot delete super user"})
		return
	}

	_, err = self.orm.ID(id).Delete(&models.TblUser{})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteUser] delete error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "failed to delete user"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// v2ListFollowupHistory lists the follow-up history for an agent run
func (self *WebServer) v2ListFollowupHistory(c *gin.Context) {
	id := c.Param("id")

	authService := auth.NewService(self.orm)
	agentRunService := agentrun.NewService(self.orm, authService)

	history, err := agentRunService.ListFollowupHistory(id)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListFollowupHistory] error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "Failed to list follow-up history",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    history,
	})
}

// v2MCPHandler handles MCP Streamable HTTP transport requests.
// It authenticates via Bearer API key, creates an MCP Server instance,
// and delegates to MCPHandler for JSON-RPC 2.0 protocol processing.
func (self *WebServer) v2MCPHandler(c *gin.Context) {
	// Extract API key from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32001, "message": "missing Authorization header"},
		})
		return
	}

	// Strip "Bearer " prefix to get the API key
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	if apiKey == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{
			"jsonrpc": "2.0",
			"error":   map[string]interface{}{"code": -32001, "message": "invalid Authorization format, expected Bearer token"},
		})
		return
	}

	// Build base URL for internal API calls
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, c.Request.Host)

	// Create MCP Server instance with the API key
	mcpServer := mcp.NewServer(baseURL, apiKey)

	// Get registered tools and tool map
	tools, toolMap := mcpServer.GetTools()

	handler := mcp.NewMCPHandlerWithRedis(mcpServer, toolMap, tools, self.redisClient)
	handler.ServeHTTP(c.Writer, c.Request)
}

// v2ListRetentionPolicies lists retention policies
func (self *WebServer) v2ListRetentionPolicies(c *gin.Context) {
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	policies, err := svc.ListPolicies(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list policies"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"items": policies, "total": len(policies)}})
}

// v2CreateRetentionPolicy creates a retention policy
func (self *WebServer) v2CreateRetentionPolicy(c *gin.Context) {
	var policy retention.RetentionPolicy
	if err := c.ShouldBindJSON(&policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	if policy.ID == "" {
		policy.ID = v2models.GenerateID()
	}
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	if err := svc.CreatePolicy(c, &policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": policy})
}

// v2GetRetentionPolicy gets a specific retention policy
func (self *WebServer) v2GetRetentionPolicy(c *gin.Context) {
	id := c.Param("id")
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	policy, err := svc.GetPolicy(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Policy not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": policy})
}

// v2UpdateRetentionPolicy updates a retention policy
func (self *WebServer) v2UpdateRetentionPolicy(c *gin.Context) {
	id := c.Param("id")
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	policy, err := svc.GetPolicy(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Policy not found"})
		return
	}
	if err := c.ShouldBindJSON(policy); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	policy.ID = id
	if err := svc.UpdatePolicy(c, policy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": policy})
}

// v2DeleteRetentionPolicy deletes a retention policy
func (self *WebServer) v2DeleteRetentionPolicy(c *gin.Context) {
	id := c.Param("id")
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	if err := svc.DeletePolicy(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// v2RunRetentionPolicy manually triggers a retention policy
func (self *WebServer) v2RunRetentionPolicy(c *gin.Context) {
	id := c.Param("id")
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	job, err := svc.RunPolicy(c, id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "message": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": job})
}

// v2ListRetentionJobs lists retention jobs
func (self *WebServer) v2ListRetentionJobs(c *gin.Context) {
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	jobs, err := svc.ListJobs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list jobs"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"items": jobs, "total": len(jobs)}})
}

// v2ListRetentionArchives lists retention archives
func (self *WebServer) v2ListRetentionArchives(c *gin.Context) {
	store := retention.NewXormStore(self.orm)
	svc := retention.NewService(store)
	archives, err := svc.ListArchives(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list archives"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": gin.H{"items": archives, "total": len(archives)}})
}

// v2ListAttackChains lists attack chains grouped by token
// @Summary List attack chains
// @Description List all attack chains grouped by token with pagination
// @Tags v2, attack-chains
// @Accept json
// @Produce json
// @Param page query int false "Page number (default 1)"
// @Param page_size query int false "Items per page (1-100, default 20)"
// @Success 200 {object} gin.H "list of attack chains"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/v2/attack-chains [get]
func (self *WebServer) v2ListAttackChains(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	iaSvc := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	chains, err := iaSvc.GetAttackChains(page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, chains)
}

// @Summary Get attack chain detail
// @Description Get detailed information about a specific attack chain by token
// @Tags v2, attack-chains
// @Accept json
// @Produce json
// @Param token path string true "Attack chain token"
// @Success 200 {object} gin.H "attack chain detail"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 404 {object} gin.H "Not found"
// @Failure 500 {object} gin.H "Internal server error"
// @Router /api/v2/attack-chains/{token} [get]
func (self *WebServer) v2GetAttackChainDetail(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "token is required"})
		return
	}

	iaSvc := interaction.NewService(self.orm, nil, self.fingerprinter, false)
	detail, err := iaSvc.GetAttackChainDetail(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "attack chain not found"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// @Summary Search ZoomEye
// @Description Query the ZoomEye search engine for internet-connected devices
// @Tags v2, search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number (default 1)"
// @Success 200 {object} gin.H "search results"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 502 {object} gin.H "Bad gateway"
// @Router /api/v2/search/zoomeye [get]
func (self *WebServer) v2SearchZoomEye(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "query is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	key := self.getSetting("zoomeye_api_key")
	if key == "" {
		c.JSON(400, gin.H{"error": "ZoomEye API key not configured"})
		return
	}

	s := search.NewZoomEye(key)
	result, err := s.Search(query, page)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "data": result})
}

// @Summary Search Shodan
// @Description Query the Shodan search engine for internet-connected devices
// @Tags v2, search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number (default 1)"
// @Success 200 {object} gin.H "search results"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 502 {object} gin.H "Bad gateway"
// @Router /api/v2/search/shodan [get]
func (self *WebServer) v2SearchShodan(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "query is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	key := self.getSetting("shodan_api_key")
	if key == "" {
		c.JSON(400, gin.H{"error": "Shodan API key not configured"})
		return
	}

	s := search.NewShodan(key)
	result, err := s.Search(query, page)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "data": result})
}

// @Summary Search Fofa
// @Description Query the Fofa search engine for internet-connected devices
// @Tags v2, search
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Param page query int false "Page number (default 1)"
// @Success 200 {object} gin.H "search results"
// @Failure 400 {object} gin.H "Bad request"
// @Failure 401 {object} gin.H "Unauthorized"
// @Failure 502 {object} gin.H "Bad gateway"
// @Router /api/v2/search/fofa [get]
func (self *WebServer) v2SearchFofa(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(400, gin.H{"error": "query is required"})
		return
	}
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	if page < 1 {
		page = 1
	}

	email := self.getSetting("fofa_email")
	key := self.getSetting("fofa_api_key")
	if email == "" || key == "" {
		c.JSON(400, gin.H{"error": "Fofa email/api key not configured"})
		return
	}

	s := search.NewFofa(email, key)
	result, err := s.Search(query, page)
	if err != nil {
		c.JSON(502, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"code": 0, "data": result})
}
