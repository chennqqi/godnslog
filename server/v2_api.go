package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/internal/canary"
	"github.com/chennqqi/godnslog/internal/interaction"
	"github.com/chennqqi/godnslog/internal/listener"
	"github.com/chennqqi/godnslog/internal/notification"
	"github.com/chennqqi/godnslog/internal/payload"
	"github.com/dgrijalva/jwt-go"

	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/rebinding"
	"github.com/chennqqi/godnslog/internal/workflow"
	"github.com/chennqqi/godnslog/models"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

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
		}

		// Interactions
		interactions := v2.Group("/interactions", self.authHandler)
		{
			interactions.GET("", self.v2ListInteractions)
			// Register /stats and /timeline before /:id so paths like /interactions/stats are not captured as ids.
			interactions.GET("/stats", self.v2InteractionStats)
			interactions.GET("/timeline", self.v2InteractionTimeline)
			interactions.POST("/delete", self.v2DeleteInteractions)
			interactions.POST("/export", self.v2ExportInteractions)
			interactions.GET("/:id", self.v2GetInteraction)
		}

		// APIKeys
		apikeys := v2.Group("/apikeys", self.authHandler)
		{
			apikeys.GET("", self.v2ListAPIKeys)
			apikeys.POST("", self.v2CreateAPIKey)
			apikeys.GET("/:id", self.v2GetAPIKey)
			apikeys.PUT("/:id", self.v2UpdateAPIKey)
			apikeys.DELETE("/:id", self.v2DeleteAPIKey)
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
		}

		// Marketplace
		marketplace := v2.Group("/marketplace", self.authHandler)
		{
			marketplace.GET("/plugins", self.v2ListPlugins)
			marketplace.GET("/plugins/:id", self.v2GetPlugin)
			marketplace.GET("/templates", self.v2ListTemplates)
			marketplace.GET("/templates/:id", self.v2GetTemplate)
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
			evidence.GET("/:id", self.v2GetEvidence)
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
		}

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
	}
}

// v2Login handles v2 login
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

	store.Set(fmt.Sprintf("%v.seed", user.Id), seed, self.AuthExpire)
	store.Set(fmt.Sprintf("%v.user", user.Id), user, cache.NoExpiration)

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
func (self *WebServer) v2UserInfo(c *gin.Context) {
	T := getTranslateFunc(c)

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

// v2GetCasePayloads gets payloads associated with a case
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
	} else if req.ExpiresIn != "" {
		// Parse ExpiresIn duration string (e.g., "24h", "1h30m")
		duration, err := time.ParseDuration(req.ExpiresIn)
		if err == nil {
			expiresAt = &[]time.Time{time.Now().Add(duration)}[0]
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

	var interactions []models.TblInteraction
	query := session.Table(new(models.TblInteraction))

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
		var headers map[string]string
		if item.Headers != "" {
			json.Unmarshal([]byte(item.Headers), &headers)
		}
		items[i] = models.Interaction{
			Id:          strconv.FormatInt(item.Id, 10),
			Type:        item.Type,
			CaseId:      strconv.FormatInt(item.CaseId, 10),
			PayloadId:   strconv.FormatInt(item.PayloadId, 10),
			Token:       item.Token,
			Timestamp:   item.Timestamp.Format(time.RFC3339),
			SourceIp:    item.SourceIp,
			Domain:      item.Domain,
			DnsType:     item.DnsType,
			Method:      item.Method,
			Path:        item.Path,
			Headers:     headers,
			Body:        item.Body,
			UserAgent:   item.UserAgent,
			ContentType: item.ContentType,
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
	interactionId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid interaction id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var interactionItem models.TblInteraction
	has, err := session.ID(interactionId).Get(&interactionItem)
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
			"code":    6,
			"message": "interaction not found",
		})
		return
	}

	var headers map[string]string
	if interactionItem.Headers != "" {
		json.Unmarshal([]byte(interactionItem.Headers), &headers)
	}

	result := models.Interaction{
		Id:          strconv.FormatInt(interactionItem.Id, 10),
		Type:        interactionItem.Type,
		CaseId:      strconv.FormatInt(interactionItem.CaseId, 10),
		PayloadId:   strconv.FormatInt(interactionItem.PayloadId, 10),
		Token:       interactionItem.Token,
		Timestamp:   interactionItem.Timestamp.Format(time.RFC3339),
		SourceIp:    interactionItem.SourceIp,
		Domain:      interactionItem.Domain,
		DnsType:     interactionItem.DnsType,
		Method:      interactionItem.Method,
		Path:        interactionItem.Path,
		Headers:     headers,
		Body:        interactionItem.Body,
		UserAgent:   interactionItem.UserAgent,
		ContentType: interactionItem.ContentType,
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

	for _, idStr := range req.Ids {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			continue
		}
		session.ID(id).Delete(new(models.TblInteraction))
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

	session := self.orm.NewSession()
	defer session.Close()

	var apiKeys []models.TblAPIKey
	query := session.Table(new(models.TblAPIKey)).Where("is_revoked = ?", false)

	total, err := query.Count()
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListAPIKeys] count error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	err = query.OrderBy("created_at DESC").Limit(pageSize, (page-1)*pageSize).Find(&apiKeys)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2ListAPIKeys] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	items := make([]models.APIKey, len(apiKeys))
	for i, item := range apiKeys {
		var scopes []string
		if item.Scopes != "" {
			json.Unmarshal([]byte(item.Scopes), &scopes)
		}
		items[i] = models.APIKey{
			Id:        strconv.FormatInt(item.Id, 10),
			KeyPrefix: item.KeyPrefix,
			Name:      item.Name,
			Scopes:    scopes,
			CreatedBy: strconv.FormatInt(item.CreatedBy, 10),
			CreatedAt: item.CreatedAt.Format(time.RFC3339),
			IsRevoked: item.IsRevoked,
		}
		if !item.ExpiresAt.IsZero() {
			items[i].ExpiresAt = item.ExpiresAt.Format(time.RFC3339)
		}
		if !item.LastUsedAt.IsZero() {
			items[i].LastUsedAt = item.LastUsedAt.Format(time.RFC3339)
		}
		if !item.RevokedAt.IsZero() {
			items[i].RevokedAt = item.RevokedAt.Format(time.RFC3339)
		}
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": models.APIKeyListResponse{
			Items:      items,
			Total:      int(total),
			Page:       page,
			PageSize:   pageSize,
			TotalPages: totalPages,
		},
	})
}

// v2CreateAPIKey creates an API key
func (self *WebServer) v2CreateAPIKey(c *gin.Context) {
	var req models.APIKeyCreateRequest
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

	// Generate API key
	key := generateAPIKey()
	keyPrefix := key[:8]

	apiKeyItem := models.TblAPIKey{
		Key:       key,
		KeyPrefix: keyPrefix,
		Name:      req.Name,
		CreatedBy: user.Id,
		IsRevoked: false,
	}

	if req.Scopes != nil {
		scopesJson, _ := json.Marshal(req.Scopes)
		apiKeyItem.Scopes = string(scopesJson)
	}

	if req.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			apiKeyItem.ExpiresAt = expiresAt
		}
	}

	session := self.orm.NewSession()
	defer session.Close()

	_, err := session.Insert(&apiKeyItem)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2CreateAPIKey] insert error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	var scopes []string
	if apiKeyItem.Scopes != "" {
		json.Unmarshal([]byte(apiKeyItem.Scopes), &scopes)
	}

	result := models.APIKey{
		Id:        strconv.FormatInt(apiKeyItem.Id, 10),
		Key:       apiKeyItem.Key,
		KeyPrefix: apiKeyItem.KeyPrefix,
		Name:      apiKeyItem.Name,
		Scopes:    scopes,
		CreatedBy: strconv.FormatInt(apiKeyItem.CreatedBy, 10),
		CreatedAt: apiKeyItem.CreatedAt.Format(time.RFC3339),
		IsRevoked: apiKeyItem.IsRevoked,
	}
	if !apiKeyItem.ExpiresAt.IsZero() {
		result.ExpiresAt = apiKeyItem.ExpiresAt.Format(time.RFC3339)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    result,
	})
}

// v2DeleteAPIKey deletes an API key
func (self *WebServer) v2DeleteAPIKey(c *gin.Context) {
	id := c.Param("id")
	apiKeyId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid api key id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	_, err = session.ID(apiKeyId).Update(&models.TblAPIKey{IsRevoked: true, RevokedAt: time.Now()})
	if err != nil {
		logrus.Errorf("[v2_api.go::v2DeleteAPIKey] update error: %v", err)
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

// v2GetAPIKey gets an API key by ID
func (self *WebServer) v2GetAPIKey(c *gin.Context) {
	id := c.Param("id")
	apiKeyId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid api key id",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	var apiKey models.TblAPIKey
	has, err := session.ID(apiKeyId).Get(&apiKey)
	if err != nil {
		logrus.Errorf("[v2_api.go::v2GetAPIKey] get error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}
	if !has {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "api key not found",
		})
		return
	}

	// Mask the key for security
	apiKey.Key = apiKey.KeyPrefix + "********"

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    apiKey,
	})
}

// v2UpdateAPIKey updates an API key
func (self *WebServer) v2UpdateAPIKey(c *gin.Context) {
	id := c.Param("id")
	apiKeyId, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid api key id",
		})
		return
	}

	var req models.APIKeyUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    400,
			"message": "invalid request body",
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	// Build update fields
	updateData := &models.TblAPIKey{}
	if req.Name != "" {
		updateData.Name = req.Name
	}
	if req.Scopes != nil {
		scopesJson, _ := json.Marshal(req.Scopes)
		updateData.Scopes = string(scopesJson)
	}
	if req.ExpiresAt != "" {
		expiresAt, err := time.Parse(time.RFC3339, req.ExpiresAt)
		if err == nil {
			updateData.ExpiresAt = expiresAt
		}
	}

	_, err = session.ID(apiKeyId).Cols("name", "scopes", "expires_at").Update(updateData)
	if err != nil {
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

	query := session.Table(new(models.TblInteraction))

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

// v2InteractionTimeline gets interaction timeline
func (self *WebServer) v2InteractionTimeline(c *gin.Context) {
	caseId := c.Query("case_id")
	payloadId := c.Query("payload_id")
	interval := c.DefaultQuery("interval", "hour")

	session := self.orm.NewSession()
	defer session.Close()

	query := session.Table(new(models.TblInteraction))

	if caseId != "" {
		query = query.Where("case_id = ?", caseId)
	}
	if payloadId != "" {
		query = query.Where("payload_id = ?", payloadId)
	}

	var interactions []models.TblInteraction
	if err := query.OrderBy("timestamp ASC").Find(&interactions); err != nil {
		logrus.Errorf("[v2_api.go::v2InteractionTimeline] find error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": "server internal error",
		})
		return
	}

	// Group by time interval
	groupedEvents := make(map[string][]models.TblInteraction)
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

// v2ListPlugins lists marketplace plugins
func (self *WebServer) v2ListPlugins(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": []map[string]interface{}{},
			"total": 0,
		},
	})
}

// v2GetPlugin gets a specific plugin
func (self *WebServer) v2GetPlugin(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
}

// v2ListTemplates lists marketplace templates
func (self *WebServer) v2ListTemplates(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": []map[string]interface{}{},
			"total": 0,
		},
	})
}

// v2GetTemplate gets a specific template
func (self *WebServer) v2GetTemplate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data":    nil,
	})
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
		Format    string `json:"format" binding:"required,oneof=json markdown csv"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "Invalid request body",
		})
		return
	}

	interactionService := interaction.NewService(self.orm)
	evidenceService := interaction.NewEvidenceService(interactionService)

	resp, err := evidenceService.GenerateEvidence(req.CaseID, req.PayloadID, req.Format)
	if err != nil {
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
