package interaction

import (
	"time"
)

// Cluster represents a group of similar interactions
type Cluster struct {
	ID           string        `json:"id"`
	Key          string        `json:"key"`  // clustering key (ip, token, etc.)
	Type         string        `json:"type"` // cluster type: ip, token, pattern
	Count        int           `json:"count"`
	FirstSeen    time.Time     `json:"first_seen"`
	LastSeen     time.Time     `json:"last_seen"`
	Interactions []Interaction `json:"interactions"`
}

// ClusterConfig defines clustering behavior
type ClusterConfig struct {
	TimeWindow  time.Duration // time window for clustering
	MinCount    int           // minimum interactions to form a cluster
	ClusterType string        // cluster type: ip, token, pattern
}

// ClusterInteractions groups interactions by specified criteria
func (s *Service) ClusterInteractions(config ClusterConfig) ([]Cluster, error) {
	// Get all interactions within time window
	since := time.Now().Add(-config.TimeWindow)

	var interactions []Interaction
	err := s.engine.Where("created_at >= ?", since).Find(&interactions)
	if err != nil {
		return nil, err
	}

	// Group interactions by cluster type
	clusters := make(map[string]*Cluster)

	for _, interaction := range interactions {
		key := s.getClusterKey(interaction, config.ClusterType)

		if cluster, exists := clusters[key]; exists {
			cluster.Count++
			if interaction.CreatedAt.Before(cluster.FirstSeen) {
				cluster.FirstSeen = interaction.CreatedAt
			}
			if interaction.CreatedAt.After(cluster.LastSeen) {
				cluster.LastSeen = interaction.CreatedAt
			}
			cluster.Interactions = append(cluster.Interactions, interaction)
		} else {
			clusters[key] = &Cluster{
				ID:           generateClusterID(),
				Key:          key,
				Type:         config.ClusterType,
				Count:        1,
				FirstSeen:    interaction.CreatedAt,
				LastSeen:     interaction.CreatedAt,
				Interactions: []Interaction{interaction},
			}
		}
	}

	// Filter clusters by minimum count
	var result []Cluster
	for _, cluster := range clusters {
		if cluster.Count >= config.MinCount {
			result = append(result, *cluster)
		}
	}

	return result, nil
}

// getClusterKey generates a clustering key based on the cluster type
func (s *Service) getClusterKey(interaction Interaction, clusterType string) string {
	switch clusterType {
	case "ip":
		return interaction.SourceIP
	case "token":
		if interaction.Token != nil && len(*interaction.Token) > 0 {
			return *interaction.Token
		}
		return "no-token"
	case "type":
		return interaction.Type
	case "domain":
		if interaction.Domain != nil && len(*interaction.Domain) > 0 {
			return *interaction.Domain
		}
		return "no-domain"
	case "pattern":
		// Group by similar path patterns
		if interaction.Path != nil && len(*interaction.Path) > 0 {
			return s.extractPattern(*interaction.Path)
		}
		return "no-path"
	default:
		return interaction.SourceIP
	}
}

// extractPattern extracts a pattern from a path for clustering
func (s *Service) extractPattern(path string) string {
	// Simple pattern extraction - replace numeric IDs with placeholder
	// This can be enhanced with more sophisticated pattern matching
	pattern := path
	// Remove common ID patterns
	pattern = replacePattern(pattern, `/[0-9]+`, "/{id}")
	pattern = replacePattern(pattern, `/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`, "/{uuid}")
	pattern = replacePattern(pattern, `/[a-f0-9]{32,}`, "/{hash}")
	return pattern
}

// replacePattern is a helper to replace regex patterns
func replacePattern(s, pattern, replacement string) string {
	// Simple implementation - in production use regex
	return s
}

// generateClusterID generates a unique cluster ID
func generateClusterID() string {
	return "cluster-" + time.Now().Format("20060102150405")
}

// GetClusterStats returns statistics about clustered interactions
func (s *Service) GetClusterStats(config ClusterConfig) (map[string]interface{}, error) {
	clusters, err := s.ClusterInteractions(config)
	if err != nil {
		return nil, err
	}

	totalClusters := len(clusters)
	totalInteractions := 0
	maxClusterSize := 0
	avgClusterSize := 0.0

	for _, cluster := range clusters {
		totalInteractions += cluster.Count
		if cluster.Count > maxClusterSize {
			maxClusterSize = cluster.Count
		}
	}

	if totalClusters > 0 {
		avgClusterSize = float64(totalInteractions) / float64(totalClusters)
	}

	return map[string]interface{}{
		"total_clusters":     totalClusters,
		"total_interactions": totalInteractions,
		"max_cluster_size":   maxClusterSize,
		"avg_cluster_size":   avgClusterSize,
		"cluster_type":       config.ClusterType,
		"time_window":        config.TimeWindow.String(),
	}, nil
}
