package clustering

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler handles clustering-related HTTP requests
type Handler struct {
	clusterer  *Clusterer
	compressor *Compressor
}

// NewHandler creates a new clustering handler
func NewHandler(clusterer *Clusterer, compressor *Compressor) *Handler {
	return &Handler{
		clusterer:  clusterer,
		compressor: compressor,
	}
}

// ClusterRequest represents the request to cluster interactions
type ClusterRequest struct {
	Interactions []interaction.Interaction `json:"interactions"`
}

// ClusterResponse represents the clustering result
type ClusterResponse struct {
	Clusters   []*Cluster `json:"clusters"`
	TotalCount int        `json:"total_count"`
	NoiseCount int        `json:"noise_count"`
}

// ClusterInteractions clusters interactions
// @Summary Cluster interactions
// @Description Group similar interactions together and detect noise
// @Tags clustering
// @Accept json
// @Produce json
// @Param request body ClusterRequest true "Interactions to cluster"
// @Success 200 {object} Response{data=ClusterResponse}
// @Router /clustering/cluster [post]
func (h *Handler) ClusterInteractions(c *gin.Context) {
	var req ClusterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	clusters := h.clusterer.ClusterInteractions(req.Interactions)
	
	noiseCount := 0
	for _, cluster := range clusters {
		if cluster.IsNoise {
			noiseCount++
		}
	}

	response := ClusterResponse{
		Clusters:   clusters,
		TotalCount: len(clusters),
		NoiseCount: noiseCount,
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": response})
}

// CompressRequest represents the request to compress interactions
type CompressRequest struct {
	Interactions []interaction.Interaction `json:"interactions"`
}

// CompressInteractions compresses interaction data
// @Summary Compress interactions
// @Description Reduce interaction data size by compressing similar interactions
// @Tags clustering
// @Accept json
// @Produce json
// @Param request body CompressRequest true "Interactions to compress"
// @Success 200 {object} Response{data=[]interaction.Interaction}
// @Router /clustering/compress [post]
func (h *Handler) CompressInteractions(c *gin.Context) {
	var req CompressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": -1, "message": err.Error()})
		return
	}

	compressed := h.compressor.CompressInteractions(req.Interactions)

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": compressed})
}

// GetConfig returns the current clustering configuration
// @Summary Get clustering config
// @Description Get the current clustering and compression configuration
// @Tags clustering
// @Produce json
// @Success 200 {object} Response{data=ConfigResponse}
// @Router /clustering/config [get]
func (h *Handler) GetConfig(c *gin.Context) {
	config := ConfigResponse{
		Clustering:  h.clusterer.config,
		Compression: h.compressor.config,
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "success", "data": config})
}

// ConfigResponse represents the clustering configuration
type ConfigResponse struct {
	Clustering  *ClusteringConfig  `json:"clustering"`
	Compression *CompressionConfig `json:"compression"`
}

// Response represents a standard API response
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
