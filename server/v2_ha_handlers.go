package server

import (
	"net/http"
	"time"

	"github.com/chennqqi/godnslog/internal/ha"
	"github.com/gin-gonic/gin"
)

// v2HealthCheck returns basic liveness info (no auth required).
func (self *WebServer) v2HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ok",
		"data": gin.H{
			"status":    "alive",
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	})
}

// v2ReadinessCheck returns whether the service is ready to accept traffic (no auth required).
func (self *WebServer) v2ReadinessCheck(c *gin.Context) {
	// Check DB connectivity
	if err := self.orm.Ping(); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    503,
			"message": "not ready",
			"data": gin.H{
				"status": "not_ready",
				"error":  err.Error(),
			},
		})
		return
	}

	data := gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}

	if self.listenerMgr != nil {
		data["active_listeners"] = self.listenerMgr.ActiveCount()
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "ready",
		"data":    data,
	})
}

// v2ListClusterNodes lists all cluster nodes.
func (self *WebServer) v2ListClusterNodes(c *gin.Context) {
	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	nodes, err := svc.ListNodes(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to list cluster nodes"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"items": nodes,
			"total": len(nodes),
		},
	})
}

// v2CreateClusterNode creates a new cluster node.
func (self *WebServer) v2CreateClusterNode(c *gin.Context) {
	var node ha.ClusterNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "Invalid request body"})
		return
	}

	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	if err := svc.AddNode(c, &node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create cluster node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": node})
}

// v2GetClusterNode gets a specific cluster node.
func (self *WebServer) v2GetClusterNode(c *gin.Context) {
	id := c.Param("id")

	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	node, err := svc.GetNode(c, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "Cluster node not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": node})
}

// v2UpdateClusterNode updates a cluster node.
func (self *WebServer) v2UpdateClusterNode(c *gin.Context) {
	id := c.Param("id")

	var node ha.ClusterNode
	if err := c.ShouldBindJSON(&node); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "Invalid request body"})
		return
	}

	node.ID = id
	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	if err := svc.UpdateNode(c, &node); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to update cluster node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": node})
}

// v2DeleteClusterNode deletes a cluster node.
func (self *WebServer) v2DeleteClusterNode(c *gin.Context) {
	id := c.Param("id")

	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	if err := svc.DeleteNode(c, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to delete cluster node"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success"})
}

// v2NodeHealthCheck performs a health check on a specific node.
func (self *WebServer) v2NodeHealthCheck(c *gin.Context) {
	id := c.Param("id")

	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	check, err := svc.PerformHealthCheck(c, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to perform health check"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": check})
}

// v2GetClusterConfig gets the cluster configuration.
func (self *WebServer) v2GetClusterConfig(c *gin.Context) {
	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	config, err := svc.GetConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to get cluster config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
}

// v2UpdateClusterConfig updates the cluster configuration.
func (self *WebServer) v2UpdateClusterConfig(c *gin.Context) {
	var config ha.ClusterConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 1, "message": "Invalid request body"})
		return
	}

	if config.ID == "" {
		config.ID = "default"
	}

	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	// Check if config already exists in DB
	configs, err := store.ListConfigs(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to check cluster config"})
		return
	}

	if len(configs) > 0 {
		// Update existing
		config.CreatedAt = configs[0].CreatedAt
		if err := svc.UpdateConfig(c, &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to update cluster config"})
			return
		}
	} else {
		// Create new
		if err := store.CreateConfig(c, &config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to create cluster config"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
}

// v2ClusterStatus returns the overall cluster status.
func (self *WebServer) v2ClusterStatus(c *gin.Context) {
	store := ha.NewXormStore(self.orm)
	svc := ha.NewService(store)

	status, err := svc.GetClusterStatus(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": "Failed to get cluster status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": status})
}
