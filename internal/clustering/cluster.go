package clustering

import (
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
)

// Cluster groups similar interactions together
type Cluster struct {
	ID           string                    `json:"id"`
	Type         string                    `json:"type"` // dns, http, etc.
	SourceIP     string                    `json:"source_ip"`
	Token        string                    `json:"token"`
	Pattern      string                    `json:"pattern"` // Common pattern (e.g., domain, path)
	Count        int                       `json:"count"`
	FirstSeen    string                    `json:"first_seen"`
	LastSeen     string                    `json:"last_seen"`
	Interactions []interaction.Interaction `json:"interactions"`
	IsNoise      bool                      `json:"is_noise"`
	NoiseReason  string                    `json:"noise_reason,omitempty"`
}

// ClusteringConfig holds configuration for interaction clustering
type ClusteringConfig struct {
	MaxClusterSize int            `json:"max_cluster_size"` // Max interactions per cluster
	TimeWindow     string         `json:"time_window"`      // Time window for clustering (e.g., "5m", "1h")
	NoiseThreshold int            `json:"noise_threshold"`  // Count threshold for noise detection
	NoisePatterns  []NoisePattern `json:"noise_patterns"`   // Known noise patterns
}

// NoisePattern defines a pattern that should be treated as noise
type NoisePattern struct {
	Type        string `json:"type"`    // dns, http
	Pattern     string `json:"pattern"` // Regex pattern
	Description string `json:"description"`
}

// DefaultClusteringConfig returns default clustering configuration
func DefaultClusteringConfig() *ClusteringConfig {
	return &ClusteringConfig{
		MaxClusterSize: 100,
		TimeWindow:     "5m",
		NoiseThreshold: 10,
		NoisePatterns: []NoisePattern{
			{
				Type:        "dns",
				Pattern:     `.*\.google\.com$`,
				Description: "Google DNS queries",
			},
			{
				Type:        "dns",
				Pattern:     `.*\.cloudflare\.com$`,
				Description: "Cloudflare DNS queries",
			},
			{
				Type:        "http",
				Pattern:     `favicon\.ico`,
				Description: "Favicon requests",
			},
			{
				Type:        "http",
				Pattern:     `robots\.txt`,
				Description: "Robots.txt requests",
			},
		},
	}
}

// Clusterer performs interaction clustering
type Clusterer struct {
	config *ClusteringConfig
}

// NewClusterer creates a new clusterer
func NewClusterer(config *ClusteringConfig) *Clusterer {
	if config == nil {
		config = DefaultClusteringConfig()
	}
	return &Clusterer{config: config}
}

// ClusterInteractions groups interactions into clusters
func (c *Clusterer) ClusterInteractions(interactions []interaction.Interaction) []*Cluster {
	clusters := make(map[string]*Cluster)

	for _, inter := range interactions {
		// Generate cluster key
		key := c.generateClusterKey(inter)

		if cluster, exists := clusters[key]; exists {
			// Add to existing cluster
			cluster.Count++
			cluster.Interactions = append(cluster.Interactions, inter)
			lastSeen, err := time.Parse(time.RFC3339, cluster.LastSeen)
			if err == nil && inter.Timestamp.After(lastSeen) {
				cluster.LastSeen = inter.Timestamp.String()
			}
		} else {
			// Create new cluster
			cluster := &Cluster{
				ID:           generateClusterID(),
				Type:         inter.Type,
				SourceIP:     inter.SourceIP,
				Token:        "",
				Pattern:      c.extractPattern(inter),
				Count:        1,
				FirstSeen:    inter.Timestamp.String(),
				LastSeen:     inter.Timestamp.String(),
				Interactions: []interaction.Interaction{inter},
				IsNoise:      false,
			}
			if inter.Token != nil {
				cluster.Token = *inter.Token
			}
			clusters[key] = cluster
		}
	}

	// Convert to slice
	result := make([]*Cluster, 0, len(clusters))
	for _, cluster := range clusters {
		// Check for noise
		c.checkNoise(cluster)
		result = append(result, cluster)
	}

	return result
}

// generateClusterKey generates a unique key for clustering
func (c *Clusterer) generateClusterKey(inter interaction.Interaction) string {
	// Key based on type, source IP, and token
	key := inter.Type + ":" + inter.SourceIP
	if inter.Token != nil {
		key += ":" + *inter.Token
	}
	return key
}

// extractPattern extracts a common pattern from the interaction
func (c *Clusterer) extractPattern(inter interaction.Interaction) string {
	// For DNS, use domain
	if inter.Domain != nil {
		return *inter.Domain
	}
	// For HTTP, use path
	if inter.Path != nil {
		return *inter.Path
	}
	return ""
}

// checkNoise checks if a cluster should be treated as noise
func (c *Clusterer) checkNoise(cluster *Cluster) {
	// Check count threshold
	if cluster.Count >= c.config.NoiseThreshold {
		cluster.IsNoise = true
		cluster.NoiseReason = "High frequency"
		return
	}

	// Check against noise patterns
	for _, pattern := range c.config.NoisePatterns {
		if pattern.Type == cluster.Type {
			// Simple pattern matching (in production, use regex)
			if cluster.Pattern == pattern.Pattern {
				cluster.IsNoise = true
				cluster.NoiseReason = pattern.Description
				return
			}
		}
	}
}

// generateClusterID generates a unique cluster ID
func generateClusterID() string {
	return "cluster-" + randomString(8)
}

// randomString generates a random string (simplified)
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}
