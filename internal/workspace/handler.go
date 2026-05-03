package workspace

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Handler handles workspace-related HTTP requests
type Handler struct {
	store Store
}

// NewHandler creates a new workspace handler
func NewHandler(store Store) *Handler {
	return &Handler{
		store: store,
	}
}

// CreateWorkspaceRequest represents the request to create a workspace
type CreateWorkspaceRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id" binding:"required"`
	IsEnabled   bool   `json:"is_enabled"`
}

// UpdateWorkspaceRequest represents the request to update a workspace
type UpdateWorkspaceRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	IsEnabled   *bool   `json:"is_enabled"`
}

// AddMemberRequest represents the request to add a member
type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
	Role   string `json:"role" binding:"required"`
}

// AddDomainRequest represents the request to add a domain
type AddDomainRequest struct {
	Domain    string `json:"domain" binding:"required"`
	IsPrimary bool   `json:"is_primary"`
}

// CreateWorkspace creates a new workspace
// @Summary Create workspace
// @Description Create a new multi-tenant workspace
// @Tags workspace
// @Accept json
// @Produce json
// @Param request body CreateWorkspaceRequest true "Workspace details"
// @Success 200 {object} Response{data=Workspace}
// @Router /workspaces [post]
func (h *Handler) CreateWorkspace(c *gin.Context) {
	var req CreateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	workspace := &Workspace{
		ID:          generateWorkspaceID(),
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     req.OwnerID,
		IsEnabled:   req.IsEnabled,
	}

	if err := h.store.CreateWorkspace(c, workspace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workspace})
}

// ListWorkspaces lists all workspaces
// @Summary List workspaces
// @Description List all workspaces
// @Tags workspace
// @Produce json
// @Param owner_id query string false "Filter by owner ID"
// @Success 200 {object} Response{data=[]Workspace}
// @Router /workspaces [get]
func (h *Handler) ListWorkspaces(c *gin.Context) {
	ownerID := c.Query("owner_id")

	var workspaces []Workspace
	var err error

	if ownerID != "" {
		workspaces, err = h.store.GetWorkspaceByOwner(c, ownerID)
	} else {
		workspaces, err = h.store.GetAllWorkspaces(c)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workspaces})
}

// GetWorkspace retrieves a workspace by ID
// @Summary Get workspace
// @Description Get a workspace by ID
// @Tags workspace
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} Response{data=Workspace}
// @Router /workspaces/{id} [get]
func (h *Handler) GetWorkspace(c *gin.Context) {
	id := c.Param("id")

	workspace, err := h.store.GetWorkspace(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Workspace not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workspace})
}

// UpdateWorkspace updates a workspace
// @Summary Update workspace
// @Description Update a workspace
// @Tags workspace
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body UpdateWorkspaceRequest true "Update details"
// @Success 200 {object} Response{data=Workspace}
// @Router /workspaces/{id} [put]
func (h *Handler) UpdateWorkspace(c *gin.Context) {
	id := c.Param("id")

	var req UpdateWorkspaceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	workspace, err := h.store.GetWorkspace(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": -1, "message": "Workspace not found"})
		return
	}

	if req.Name != nil {
		workspace.Name = *req.Name
	}
	if req.Description != nil {
		workspace.Description = *req.Description
	}
	if req.IsEnabled != nil {
		workspace.IsEnabled = *req.IsEnabled
	}

	if err := h.store.UpdateWorkspace(c, workspace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": workspace})
}

// DeleteWorkspace deletes a workspace
// @Summary Delete workspace
// @Description Delete a workspace
// @Tags workspace
// @Param id path string true "Workspace ID"
// @Success 200 {object} Response
// @Router /workspaces/{id} [delete]
func (h *Handler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")

	if err := h.store.DeleteWorkspace(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListMembers lists members of a workspace
// @Summary List workspace members
// @Description List all members of a workspace
// @Tags workspace
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} Response{data=[]WorkspaceMember}
// @Router /workspaces/{id}/members [get]
func (h *Handler) ListMembers(c *gin.Context) {
	id := c.Param("id")

	members, err := h.store.GetWorkspaceMembers(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": members})
}

// AddMember adds a member to a workspace
// @Summary Add workspace member
// @Description Add a member to a workspace
// @Tags workspace
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body AddMemberRequest true "Member details"
// @Success 200 {object} Response{data=WorkspaceMember}
// @Router /workspaces/{id}/members [post]
func (h *Handler) AddMember(c *gin.Context) {
	id := c.Param("id")

	var req AddMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	member := &WorkspaceMember{
		ID:          generateMemberID(),
		WorkspaceID: id,
		UserID:      req.UserID,
		Role:        req.Role,
	}

	if err := h.store.AddWorkspaceMember(c, member); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": member})
}

// RemoveMember removes a member from a workspace
// @Summary Remove workspace member
// @Description Remove a member from a workspace
// @Tags workspace
// @Param id path string true "Workspace ID"
// @Param user_id path string true "User ID"
// @Success 200 {object} Response
// @Router /workspaces/{id}/members/{user_id} [delete]
func (h *Handler) RemoveMember(c *gin.Context) {
	id := c.Param("id")
	userID := c.Param("user_id")

	if err := h.store.RemoveWorkspaceMember(c, id, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// ListDomains lists domains of a workspace
// @Summary List workspace domains
// @Description List all domains of a workspace
// @Tags workspace
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} Response{data=[]WorkspaceDomain}
// @Router /workspaces/{id}/domains [get]
func (h *Handler) ListDomains(c *gin.Context) {
	id := c.Param("id")

	domains, err := h.store.GetWorkspaceDomains(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": domains})
}

// AddDomain adds a domain to a workspace
// @Summary Add workspace domain
// @Description Add a domain to a workspace
// @Tags workspace
// @Accept json
// @Produce json
// @Param id path string true "Workspace ID"
// @Param request body AddDomainRequest true "Domain details"
// @Success 200 {object} Response{data=WorkspaceDomain}
// @Router /workspaces/{id}/domains [post]
func (h *Handler) AddDomain(c *gin.Context) {
	id := c.Param("id")

	var req AddDomainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	domain := &WorkspaceDomain{
		ID:          generateDomainID(),
		WorkspaceID: id,
		Domain:      req.Domain,
		IsPrimary:   req.IsPrimary,
	}

	if err := h.store.AddWorkspaceDomain(c, domain); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": domain})
}

// GetStats retrieves workspace statistics
// @Summary Get workspace stats
// @Description Get statistics for a workspace
// @Tags workspace
// @Produce json
// @Param id path string true "Workspace ID"
// @Success 200 {object} Response{data=WorkspaceStats}
// @Router /workspaces/{id}/stats [get]
func (h *Handler) GetStats(c *gin.Context) {
	id := c.Param("id")

	stats, err := h.store.GetWorkspaceStats(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": -1, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": stats})
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// generateWorkspaceID generates a unique workspace ID
func generateWorkspaceID() string {
	return "ws-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// generateMemberID generates a unique member ID
func generateMemberID() string {
	return "member-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// generateDomainID generates a unique domain ID
func generateDomainID() string {
	return "domain-" + strconv.FormatInt(time.Now().UnixNano(), 36)
}
