package ha

import (
	"context"

	"xorm.io/xorm"
)

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreateNode creates a new cluster node
func (s *XormStore) CreateNode(ctx context.Context, node *ClusterNode) error {
	_, err := s.engine.Insert(node)
	return err
}

// GetNode retrieves a node by ID
func (s *XormStore) GetNode(ctx context.Context, id string) (*ClusterNode, error) {
	var node ClusterNode
	_, err := s.engine.ID(id).Get(&node)
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// ListNodes lists all cluster nodes
func (s *XormStore) ListNodes(ctx context.Context) ([]ClusterNode, error) {
	var nodes []ClusterNode
	err := s.engine.Find(&nodes)
	return nodes, err
}

// UpdateNode updates a cluster node
func (s *XormStore) UpdateNode(ctx context.Context, node *ClusterNode) error {
	_, err := s.engine.ID(node.ID).Update(node)
	return err
}

// DeleteNode deletes a cluster node
func (s *XormStore) DeleteNode(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&ClusterNode{})
	return err
}

// CreateConfig creates a new cluster config
func (s *XormStore) CreateConfig(ctx context.Context, config *ClusterConfig) error {
	_, err := s.engine.Insert(config)
	return err
}

// GetConfig retrieves a config by ID
func (s *XormStore) GetConfig(ctx context.Context, id string) (*ClusterConfig, error) {
	var config ClusterConfig
	_, err := s.engine.ID(id).Get(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

// ListConfigs lists all cluster configs
func (s *XormStore) ListConfigs(ctx context.Context) ([]ClusterConfig, error) {
	var configs []ClusterConfig
	err := s.engine.Find(&configs)
	return configs, err
}

// UpdateConfig updates a cluster config
func (s *XormStore) UpdateConfig(ctx context.Context, config *ClusterConfig) error {
	_, err := s.engine.ID(config.ID).Update(config)
	return err
}

// CreateHealthCheck creates a new health check
func (s *XormStore) CreateHealthCheck(ctx context.Context, check *HealthCheck) error {
	_, err := s.engine.Insert(check)
	return err
}

// ListHealthChecks lists health checks for a node
func (s *XormStore) ListHealthChecks(ctx context.Context, nodeID string) ([]HealthCheck, error) {
	var checks []HealthCheck
	err := s.engine.Where("node_id = ?", nodeID).Desc("timestamp").Find(&checks)
	return checks, err
}
