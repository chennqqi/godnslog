package rebinding

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles rebinding-related HTTP requests
type Handler struct {
	resolver *Resolver
	store    Store
	config   *RebindingConfig
}

// NewHandler creates a new rebinding handler
func NewHandler(resolver *Resolver, store Store, config *RebindingConfig) *Handler {
	return &Handler{
		resolver: resolver,
		store:    store,
		config:   config,
	}
}

// CreateRuleRequest represents the request to create a rebinding rule
type CreateRuleRequest struct {
	Domain    string  `json:"domain" binding:"required"`
	Stages    []Stage `json:"stages" binding:"required"`
	IsEnabled bool    `json:"is_enabled"`
}

// UpdateRuleRequest represents the request to update a rebinding rule
type UpdateRuleRequest struct {
	Stages    []Stage `json:"stages"`
	IsEnabled *bool   `json:"is_enabled"`
}

// CreateScenarioRequest represents the request to create a rule from scenario
type CreateScenarioRequest struct {
	Domain   string            `json:"domain" binding:"required"`
	Scenario RebindingScenario `json:"scenario" binding:"required"`
}

// CreateRule creates a new rebinding rule
// @Summary Create rebinding rule
// @Description Create a new DNS rebinding rule
// @Tags rebinding
// @Accept json
// @Produce json
// @Param request body CreateRuleRequest true "Rule details"
// @Success 200 {object} Response{data=RebindingRule}
// @Router /rebinding/rules [post]
func (h *Handler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Validate stages
	if len(req.Stages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "At least one stage is required"})
		return
	}

	// Validate IPs
	for _, stage := range req.Stages {
		if !ValidateIP(stage.TargetIP) {
			c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "Invalid IP address in stage"})
			return
		}
	}

	rule := &RebindingRule{
		ID:        generateRuleID(),
		Domain:    req.Domain,
		Stages:    req.Stages,
		IsEnabled: req.IsEnabled,
	}

	if err := h.store.CreateRebindingRule(c, rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// CreateScenarioRule creates a rebinding rule from a predefined scenario
// @Summary Create scenario rule
// @Description Create a rebinding rule from a predefined scenario
// @Tags rebinding
// @Accept json
// @Produce json
// @Param request body CreateScenarioRequest true "Scenario details"
// @Success 200 {object} Response{data=RebindingRule}
// @Router /rebinding/rules/scenario [post]
func (h *Handler) CreateScenarioRule(c *gin.Context) {
	var req CreateScenarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	rule, err := h.resolver.CreateScenarioRule(c, req.Domain, req.Scenario)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// ListRules lists all rebinding rules
// @Summary List rebinding rules
// @Description List all DNS rebinding rules
// @Tags rebinding
// @Produce json
// @Success 200 {object} Response{data=[]RebindingRule}
// @Router /rebinding/rules [get]
func (h *Handler) ListRules(c *gin.Context) {
	rules, err := h.store.GetAllRebindingRules(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rules})
}

// GetRule retrieves a rebinding rule by ID
// @Summary Get rebinding rule
// @Description Get a rebinding rule by ID
// @Tags rebinding
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} Response{data=RebindingRule}
// @Router /rebinding/rules/{id} [get]
func (h *Handler) GetRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.store.GetRebindingRule(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Rule not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// UpdateRule updates a rebinding rule
// @Summary Update rebinding rule
// @Description Update a rebinding rule
// @Tags rebinding
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param request body UpdateRuleRequest true "Update details"
// @Success 200 {object} Response{data=RebindingRule}
// @Router /rebinding/rules/{id} [put]
func (h *Handler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	rule, err := h.store.GetRebindingRule(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Rule not found"})
		return
	}

	if req.Stages != nil {
		rule.Stages = req.Stages
	}
	if req.IsEnabled != nil {
		rule.IsEnabled = *req.IsEnabled
	}

	if err := h.store.UpdateRebindingRule(c, rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// DeleteRule deletes a rebinding rule
// @Summary Delete rebinding rule
// @Description Delete a rebinding rule
// @Tags rebinding
// @Param id path string true "Rule ID"
// @Success 200 {object} Response
// @Router /rebinding/rules/{id} [delete]
func (h *Handler) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteRebindingRule(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListSessions lists sessions for a rule
// @Summary List rebinding sessions
// @Description List all sessions for a rebinding rule
// @Tags rebinding
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} Response{data=[]RebindingSession}
// @Router /rebinding/rules/{id}/sessions [get]
func (h *Handler) ListSessions(c *gin.Context) {
	id := c.Param("id")

	sessions, err := h.store.GetRebindingSessionsByRule(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": sessions})
}

// GetConfig returns the current rebinding configuration
// @Summary Get rebinding config
// @Description Get the current rebinding configuration
// @Tags rebinding
// @Produce json
// @Success 200 {object} Response{data=RebindingConfig}
// @Router /rebinding/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": h.config})
}

// UpdateConfig updates the rebinding configuration
// @Summary Update rebinding config
// @Description Update the rebinding configuration
// @Tags rebinding
// @Accept json
// @Produce json
// @Param config body RebindingConfig true "Configuration"
// @Success 200 {object} Response{data=RebindingConfig}
// @Router /rebinding/config [put]
func (h *Handler) UpdateConfig(c *gin.Context) {
	var config RebindingConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Security: C2 requires additional approval
	if config.EnableC2 && !h.config.EnableC2 {
		// In production, require additional approval
		c.JSON(http.StatusForbidden, gin.H{"code": -1, "message": "C2 enablement requires additional approval"})
		return
	}

	h.config = &config
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": h.config})
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
