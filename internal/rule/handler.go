package rule

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles rule-related HTTP requests
type Handler struct {
	store Store
}

// NewHandler creates a new rule handler
func NewHandler(store Store) *Handler {
	return &Handler{store: store}
}

// CreateRuleRequest represents the request to create a rule
type CreateRuleRequest struct {
	Name        string     `json:"name" binding:"required"`
	Description string     `json:"description"`
	Enabled     bool       `json:"enabled"`
	Priority    int        `json:"priority"`
	Conditions  Conditions `json:"conditions"`
	Actions     Actions    `json:"actions"`
}

// UpdateRuleRequest represents the request to update a rule
type UpdateRuleRequest struct {
	Name        *string     `json:"name"`
	Description *string     `json:"description"`
	Enabled     *bool       `json:"enabled"`
	Priority    *int        `json:"priority"`
	Conditions  *Conditions `json:"conditions"`
	Actions     *Actions    `json:"actions"`
}

// CreateRule creates a new rule
// @Summary Create a rule
// @Description Create a new automation rule
// @Tags rules
// @Accept json
// @Produce json
// @Param request body CreateRuleRequest true "Rule data"
// @Success 200 {object} Response{data=Rule}
// @Router /rules [post]
func (h *Handler) CreateRule(c *gin.Context) {
	var req CreateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	rule := &Rule{
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
		Priority:    req.Priority,
		Conditions:  req.Conditions,
		Actions:     req.Actions,
	}

	if err := h.store.CreateRule(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// GetRule retrieves a rule by ID
// @Summary Get a rule
// @Description Get a rule by ID
// @Tags rules
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} Response{data=Rule}
// @Router /rules/{id} [get]
func (h *Handler) GetRule(c *gin.Context) {
	id := c.Param("id")

	rule, err := h.store.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// ListRules lists all rules
// @Summary List rules
// @Description List all rules with pagination
// @Tags rules
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} Response{data=[]Rule}
// @Router /rules [get]
func (h *Handler) ListRules(c *gin.Context) {
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if _, err := fmt.Sscanf(ps, "%d", &pageSize); err != nil {
			pageSize = 20
		}
	}

	rules, total, err := h.store.ListRules(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     rules,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// UpdateRule updates a rule
// @Summary Update a rule
// @Description Update an existing rule
// @Tags rules
// @Accept json
// @Produce json
// @Param id path string true "Rule ID"
// @Param request body UpdateRuleRequest true "Rule data"
// @Success 200 {object} Response{data=Rule}
// @Router /rules/{id} [put]
func (h *Handler) UpdateRule(c *gin.Context) {
	id := c.Param("id")

	var req UpdateRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	rule, err := h.store.GetRule(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": err.Error()})
		return
	}

	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Priority != nil {
		rule.Priority = *req.Priority
	}
	if req.Conditions != nil {
		rule.Conditions = *req.Conditions
	}
	if req.Actions != nil {
		rule.Actions = *req.Actions
	}

	if err := h.store.UpdateRule(c.Request.Context(), rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": rule})
}

// DeleteRule deletes a rule
// @Summary Delete a rule
// @Description Delete a rule by ID
// @Tags rules
// @Produce json
// @Param id path string true "Rule ID"
// @Success 200 {object} Response
// @Router /rules/{id} [delete]
func (h *Handler) DeleteRule(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteRule(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// GetExecutions retrieves execution records for a rule
// @Summary Get rule executions
// @Description Get execution records for a specific rule
// @Tags rules
// @Produce json
// @Param id path string true "Rule ID"
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} Response{data=[]RuleExecution}
// @Router /rules/{id}/executions [get]
func (h *Handler) GetExecutions(c *gin.Context) {
	id := c.Param("id")
	page := 1
	pageSize := 20

	if p := c.Query("page"); p != "" {
		if _, err := fmt.Sscanf(p, "%d", &page); err != nil {
			page = 1
		}
	}
	if ps := c.Query("page_size"); ps != "" {
		if _, err := fmt.Sscanf(ps, "%d", &pageSize); err != nil {
			pageSize = 20
		}
	}

	executions, total, err := h.store.GetExecutions(c.Request.Context(), id, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items":     executions,
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
