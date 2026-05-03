package canary

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles canary-related HTTP requests
type Handler struct {
	detector *Detector
	store    Store
	config   *CanaryConfig
}

// NewHandler creates a new canary handler
func NewHandler(detector *Detector, store Store, config *CanaryConfig) *Handler {
	return &Handler{
		detector: detector,
		store:    store,
		config:   config,
	}
}

// CreateCanaryRequest represents the request to create a canary
type CreateCanaryRequest struct {
	Type        string `json:"type" binding:"required"`
	Token       string `json:"token" binding:"required"`
	Description string `json:"description"`
	Context     string `json:"context"`
	ExpiresIn   string `json:"expires_in"` // e.g., "90d"
}

// UpdateCanaryRequest represents the request to update a canary
type UpdateCanaryRequest struct {
	Description string `json:"description"`
	IsEnabled   *bool  `json:"is_enabled"`
}

// CreateCanary creates a new canary token
// @Summary Create canary token
// @Description Create a new canary token for long-term monitoring
// @Tags canary
// @Accept json
// @Produce json
// @Param request body CreateCanaryRequest true "Canary details"
// @Success 200 {object} Response{data=Canary}
// @Router /canaries [post]
func (h *Handler) CreateCanary(c *gin.Context) {
	var req CreateCanaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Parse expiration
	expiresIn := h.config.DefaultExpiry
	if req.ExpiresIn != "" {
		expiresIn = req.ExpiresIn
	}
	expiresAt, err := parseExpiration(expiresIn)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "Invalid expiration format"})
		return
	}

	// Create canary
	canary := &Canary{
		ID:          generateCanaryID(),
		Type:        req.Type,
		Token:       req.Token,
		Description: req.Description,
		Context:     req.Context,
		ExpiresAt:   expiresAt,
		IsEnabled:   true,
	}

	if err := h.store.CreateCanary(c, canary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": canary})
}

// ListCanaries lists all canaries
// @Summary List canaries
// @Description List all canary tokens
// @Tags canary
// @Produce json
// @Param enabled query bool false "Filter by enabled status"
// @Success 200 {object} Response{data=[]Canary}
// @Router /canaries [get]
func (h *Handler) ListCanaries(c *gin.Context) {
	enabledStr := c.Query("enabled")
	
	var canaries []Canary
	var err error
	
	if enabledStr == "" {
		canaries, err = h.store.GetAllCanaries(c)
	} else {
		enabled := enabledStr == "true"
		if enabled {
			canaries, err = h.store.GetActiveCanaries(c)
		} else {
			// Get all and filter
			all, err := h.store.GetAllCanaries(c)
			if err == nil {
				for _, c := range all {
					if !c.IsEnabled {
						canaries = append(canaries, c)
					}
				}
			}
		}
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": canaries})
}

// GetCanary retrieves a canary by ID
// @Summary Get canary
// @Description Get a canary token by ID
// @Tags canary
// @Produce json
// @Param id path string true "Canary ID"
// @Success 200 {object} Response{data=Canary}
// @Router /canaries/{id} [get]
func (h *Handler) GetCanary(c *gin.Context) {
	id := c.Param("id")
	
	canary, err := h.store.GetCanary(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Canary not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": canary})
}

// UpdateCanary updates a canary
// @Summary Update canary
// @Description Update a canary token
// @Tags canary
// @Accept json
// @Produce json
// @Param id path string true "Canary ID"
// @Param request body UpdateCanaryRequest true "Update details"
// @Success 200 {object} Response{data=Canary}
// @Router /canaries/{id} [put]
func (h *Handler) UpdateCanary(c *gin.Context) {
	id := c.Param("id")
	
	var req UpdateCanaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	canary, err := h.store.GetCanary(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Canary not found"})
		return
	}

	if req.Description != "" {
		canary.Description = req.Description
	}
	if req.IsEnabled != nil {
		canary.IsEnabled = *req.IsEnabled
	}

	if err := h.store.UpdateCanary(c, canary); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": canary})
}

// DeleteCanary deletes a canary
// @Summary Delete canary
// @Description Delete a canary token
// @Tags canary
// @Param id path string true "Canary ID"
// @Success 200 {object} Response
// @Router /canaries/{id} [delete]
func (h *Handler) DeleteCanary(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.store.DeleteCanary(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListCanaryHits lists hits for a canary
// @Summary List canary hits
// @Description List all hits for a canary token
// @Tags canary
// @Produce json
// @Param id path string true "Canary ID"
// @Success 200 {object} Response{data=[]CanaryHit}
// @Router /canaries/{id}/hits [get]
func (h *Handler) ListCanaryHits(c *gin.Context) {
	id := c.Param("id")
	
	hits, err := h.store.GetCanaryHits(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": hits})
}

// GetCanaryStats retrieves statistics for a canary
// @Summary Get canary stats
// @Description Get statistics for a canary token
// @Tags canary
// @Produce json
// @Param id path string true "Canary ID"
// @Success 200 {object} Response{data=CanaryStats}
// @Router /canaries/{id}/stats [get]
func (h *Handler) GetCanaryStats(c *gin.Context) {
	id := c.Param("id")
	
	hits, err := h.store.GetCanaryHits(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	canary, err := h.store.GetCanary(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Canary not found"})
		return
	}

	stats := CanaryStats{
		CanaryID:    id,
		TotalHits:   len(hits),
		RiskLevel:   h.assessOverallRisk(hits, canary),
		FirstHit:    getFirstHitTime(hits),
		LastHit:     getLastHitTime(hits),
		UniqueIPs:   countUniqueIPs(hits),
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": stats})
}

// CanaryStats represents canary statistics
type CanaryStats struct {
	CanaryID    string `json:"canary_id"`
	TotalHits   int    `json:"total_hits"`
	RiskLevel   string `json:"risk_level"`
	FirstHit    string `json:"first_hit"`
	LastHit     string `json:"last_hit"`
	UniqueIPs   int    `json:"unique_ips"`
}

// parseExpiration parses expiration string to time
func parseExpiration(exp string) (time.Time, error) {
	// Simple parser for MVP
	// Supports: 90d, 24h, 60m
	duration, err := time.ParseDuration(exp)
	if err != nil {
		return time.Time{}, err
	}
	return time.Now().Add(duration), nil
}

// generateCanaryID generates a unique canary ID
func generateCanaryID() string {
	return "canary-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// assessOverallRisk assesses overall risk for canary
func (h *Handler) assessOverallRisk(hits []CanaryHit, canary *Canary) string {
	if len(hits) == 0 {
		return "none"
	}

	// Count high-risk hits
	highRiskCount := 0
	for _, hit := range hits {
		risk := h.detector.AssessRisk(&hit, canary)
		if risk == "critical" || risk == "high" {
			highRiskCount++
		}
	}

	// Assess overall risk
	if highRiskCount > 0 {
		return "high"
	}
	if len(hits) > 10 {
		return "medium"
	}
	return "low"
}

// getFirstHitTime gets the first hit time
func getFirstHitTime(hits []CanaryHit) string {
	if len(hits) == 0 {
		return ""
	}
	return hits[len(hits)-1].Timestamp.Format(time.RFC3339)
}

// getLastHitTime gets the last hit time
func getLastHitTime(hits []CanaryHit) string {
	if len(hits) == 0 {
		return ""
	}
	return hits[0].Timestamp.Format(time.RFC3339)
}

// countUniqueIPs counts unique source IPs
func countUniqueIPs(hits []CanaryHit) int {
	ips := make(map[string]bool)
	for _, hit := range hits {
		ips[hit.SourceIP] = true
	}
	return len(ips)
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
