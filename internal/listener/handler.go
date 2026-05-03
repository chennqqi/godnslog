package listener

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles listener-related HTTP requests
type Handler struct {
	store  Store
	config *ListenerConfig
}

// NewHandler creates a new listener handler
func NewHandler(store Store, config *ListenerConfig) *Handler {
	return &Handler{
		store:  store,
		config: config,
	}
}

// CreateListenerRequest represents the request to create a listener
type CreateListenerRequest struct {
	Protocol  Protocol `json:"protocol" binding:"required"`
	Host      string   `json:"host" binding:"required"`
	Port      int      `json:"port" binding:"required"`
	Token     string   `json:"token" binding:"required"`
	IsEnabled bool     `json:"is_enabled"`
}

// UpdateListenerRequest represents the request to update a listener
type UpdateListenerRequest struct {
	IsEnabled *bool `json:"is_enabled"`
}

// CreateListener creates a new listener
// @Summary Create listener
// @Description Create a new protocol listener
// @Tags listener
// @Accept json
// @Produce json
// @Param request body CreateListenerRequest true "Listener details"
// @Success 200 {object} Response{data=Listener}
// @Router /listeners [post]
func (h *Handler) CreateListener(c *gin.Context) {
	var req CreateListenerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Validate protocol
	if !isValidProtocol(req.Protocol) {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": "Invalid protocol"})
		return
	}

	listener := &Listener{
		ID:        generateListenerID(),
		Protocol:  req.Protocol,
		Host:      req.Host,
		Port:      req.Port,
		Token:     req.Token,
		IsEnabled: req.IsEnabled,
	}

	if err := h.store.CreateListener(c, listener); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": listener})
}

// ListListeners lists all listeners
// @Summary List listeners
// @Description List all protocol listeners
// @Tags listener
// @Produce json
// @Param protocol query string false "Filter by protocol"
// @Success 200 {object} Response{data=[]Listener}
// @Router /listeners [get]
func (h *Handler) ListListeners(c *gin.Context) {
	protocolStr := c.Query("protocol")

	listeners, err := h.store.GetAllListeners(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	// Filter by protocol if specified
	if protocolStr != "" {
		protocol := Protocol(protocolStr)
		filtered := make([]Listener, 0)
		for _, l := range listeners {
			if l.Protocol == protocol {
				filtered = append(filtered, l)
			}
		}
		listeners = filtered
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": listeners})
}

// GetListener retrieves a listener by ID
// @Summary Get listener
// @Description Get a listener by ID
// @Tags listener
// @Produce json
// @Param id path string true "Listener ID"
// @Success 200 {object} Response{data=Listener}
// @Router /listeners/{id} [get]
func (h *Handler) GetListener(c *gin.Context) {
	id := c.Param("id")

	listener, err := h.store.GetListener(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Listener not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": listener})
}

// UpdateListener updates a listener
// @Summary Update listener
// @Description Update a listener
// @Tags listener
// @Accept json
// @Produce json
// @Param id path string true "Listener ID"
// @Param request body UpdateListenerRequest true "Update details"
// @Success 200 {object} Response{data=Listener}
// @Router /listeners/{id} [put]
func (h *Handler) UpdateListener(c *gin.Context) {
	id := c.Param("id")

	var req UpdateListenerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	listener, err := h.store.GetListener(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Listener not found"})
		return
	}

	if req.IsEnabled != nil {
		listener.IsEnabled = *req.IsEnabled
	}

	if err := h.store.UpdateListener(c, listener); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": listener})
}

// DeleteListener deletes a listener
// @Summary Delete listener
// @Description Delete a listener
// @Tags listener
// @Param id path string true "Listener ID"
// @Success 200 {object} Response
// @Router /listeners/{id} [delete]
func (h *Handler) DeleteListener(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteListener(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListInteractions lists interactions for a listener
// @Summary List listener interactions
// @Description List all interactions for a listener
// @Tags listener
// @Produce json
// @Param id path string true "Listener ID"
// @Success 200 {object} Response{data=[]ListenerInteraction}
// @Router /listeners/{id}/interactions [get]
func (h *Handler) ListInteractions(c *gin.Context) {
	id := c.Param("id")

	interactions, err := h.store.GetListenerInteractions(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": interactions})
}

// ListSMTPMessages lists SMTP messages for a listener
// @Summary List SMTP messages
// @Description List all SMTP messages for a listener
// @Tags listener
// @Produce json
// @Param id path string true "Listener ID"
// @Success 200 {object} Response{data=[]SMTPMessage}
// @Router /listeners/{id}/smtp [get]
func (h *Handler) ListSMTPMessages(c *gin.Context) {
	id := c.Param("id")

	messages, err := h.store.GetSMTPMessages(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": messages})
}

// ListLDAPQueries lists LDAP queries for a listener
// @Summary List LDAP queries
// @Description List all LDAP queries for a listener
// @Tags listener
// @Produce json
// @Param id path string true "Listener ID"
// @Success 200 {object} Response{data=[]LDAPQuery}
// @Router /listeners/{id}/ldap [get]
func (h *Handler) ListLDAPQueries(c *gin.Context) {
	id := c.Param("id")

	queries, err := h.store.GetLDAPQueries(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": queries})
}

// GetConfig returns the current listener configuration
// @Summary Get listener config
// @Description Get the current listener configuration
// @Tags listener
// @Produce json
// @Success 200 {object} Response{data=ListenerConfig}
// @Router /listeners/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": h.config})
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// isValidProtocol checks if protocol is valid
func isValidProtocol(protocol Protocol) bool {
	switch protocol {
	case ProtocolSMTP, ProtocolLDAP, ProtocolSMB, ProtocolFTP:
		return true
	default:
		return false
	}
}

// generateListenerID generates a unique listener ID
func generateListenerID() string {
	return "listener-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
