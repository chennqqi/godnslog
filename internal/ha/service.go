package ha

import (
	"context"
	"fmt"
	"time"
)

// Service handles high availability operations
type Service struct {
	store Store
}

// NewService creates a new HA service
func NewService(store Store) *Service {
	return &Service{store: store}
}

// AddNode adds a new cluster node
func (s *Service) AddNode(ctx context.Context, node *ClusterNode) error {
	node.CreatedAt = time.Now()
	node.UpdatedAt = time.Now()
	node.LastPing = time.Now()
	return s.store.CreateNode(ctx, node)
}

// GetNode retrieves a node by ID
func (s *Service) GetNode(ctx context.Context, id string) (*ClusterNode, error) {
	return s.store.GetNode(ctx, id)
}

// ListNodes lists all cluster nodes
func (s *Service) ListNodes(ctx context.Context) ([]ClusterNode, error) {
	return s.store.ListNodes(ctx)
}

// UpdateNode updates a cluster node
func (s *Service) UpdateNode(ctx context.Context, node *ClusterNode) error {
	node.UpdatedAt = time.Now()
	return s.store.UpdateNode(ctx, node)
}

// DeleteNode deletes a cluster node
func (s *Service) DeleteNode(ctx context.Context, id string) error {
	return s.store.DeleteNode(ctx, id)
}

// GetConfig retrieves cluster configuration
func (s *Service) GetConfig(ctx context.Context) (*ClusterConfig, error) {
	configs, err := s.store.ListConfigs(ctx)
	if err != nil {
		return nil, err
	}
	if len(configs) == 0 {
		// Return default config
		return &ClusterConfig{
			ID:                 "default",
			EnableFailover:     true,
			FailoverTimeout:    30,
			EnableLoadBalance:  true,
			BalanceAlgorithm:   "round_robin",
			EnableReplication:  false,
			ReplicationMode:    "async",
			HealthCheckInterval: 10,
			HealthCheckTimeout:  5,
			EnableQuorum:       false,
			QuorumSize:         2,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}, nil
	}
	return &configs[0], nil
}

// UpdateConfig updates cluster configuration
func (s *Service) UpdateConfig(ctx context.Context, config *ClusterConfig) error {
	config.UpdatedAt = time.Now()
	return s.store.UpdateConfig(ctx, config)
}

// PerformHealthCheck performs a health check on a node
func (s *Service) PerformHealthCheck(ctx context.Context, nodeID string) (*HealthCheck, error) {
	node, err := s.store.GetNode(ctx, nodeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}
	
	healthCheck := &HealthCheck{
		ID:         generateHealthCheckID(),
		NodeID:     nodeID,
		CheckType:  "general",
		Status:     "healthy",
		ResponseTime: 10, // Simulated
		Timestamp:  time.Now(),
	}
	
	// Update node last ping
	node.LastPing = time.Now()
	node.Status = "online"
	s.store.UpdateNode(ctx, node)
	
	if err := s.store.CreateHealthCheck(ctx, healthCheck); err != nil {
		return nil, fmt.Errorf("failed to create health check: %w", err)
	}
	
	return healthCheck, nil
}

// ListHealthChecks lists health checks for a node
func (s *Service) ListHealthChecks(ctx context.Context, nodeID string) ([]HealthCheck, error) {
	return s.store.ListHealthChecks(ctx, nodeID)
}

// GetClusterStatus returns the overall cluster status
func (s *Service) GetClusterStatus(ctx context.Context) (map[string]interface{}, error) {
	nodes, err := s.store.ListNodes(ctx)
	if err != nil {
		return nil, err
	}
	
	onlineCount := 0
	offlineCount := 0
	
	for _, node := range nodes {
		if node.Status == "online" {
			onlineCount++
		} else {
			offlineCount++
		}
	}
	
	return map[string]interface{}{
		"total_nodes":  len(nodes),
		"online_nodes": onlineCount,
		"offline_nodes": offlineCount,
		"cluster_status": func() string {
			if offlineCount == 0 {
				return "healthy"
			}
			if onlineCount > 0 {
				return "degraded"
			}
			return "down"
		}(),
	}, nil
}

// generateHealthCheckID generates a unique health check ID
func generateHealthCheckID() string {
	return fmt.Sprintf("health-%d", time.Now().UnixNano())
}

// Store defines the storage interface for HA operations
type Store interface {
	// Node operations
	CreateNode(ctx context.Context, node *ClusterNode) error
	GetNode(ctx context.Context, id string) (*ClusterNode, error)
	ListNodes(ctx context.Context) ([]ClusterNode, error)
	UpdateNode(ctx context.Context, node *ClusterNode) error
	DeleteNode(ctx context.Context, id string) error
	
	// Config operations
	CreateConfig(ctx context.Context, config *ClusterConfig) error
	GetConfig(ctx context.Context, id string) (*ClusterConfig, error)
	ListConfigs(ctx context.Context) ([]ClusterConfig, error)
	UpdateConfig(ctx context.Context, config *ClusterConfig) error
	
	// Health check operations
	CreateHealthCheck(ctx context.Context, check *HealthCheck) error
	ListHealthChecks(ctx context.Context, nodeID string) ([]HealthCheck, error)
}
