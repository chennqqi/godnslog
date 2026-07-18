package ha

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// mockStore implements Store for testing
type mockStore struct {
	nodes        []ClusterNode
	configs      []ClusterConfig
	healthChecks []HealthCheck
	leader       *LeaderElection
	createNodeFn func(ctx context.Context, node *ClusterNode) error
	getNodeFn    func(ctx context.Context, id string) (*ClusterNode, error)
	err          error
}

func (m *mockStore) CreateNode(ctx context.Context, node *ClusterNode) error {
	if m.createNodeFn != nil {
		return m.createNodeFn(ctx, node)
	}
	if m.err != nil {
		return m.err
	}
	m.nodes = append(m.nodes, *node)
	return nil
}

func (m *mockStore) GetNode(ctx context.Context, id string) (*ClusterNode, error) {
	if m.getNodeFn != nil {
		return m.getNodeFn(ctx, id)
	}
	if m.err != nil {
		return nil, m.err
	}
	for _, n := range m.nodes {
		if n.ID == id {
			return &n, nil
		}
	}
	return &ClusterNode{}, nil
}

func (m *mockStore) ListNodes(ctx context.Context) ([]ClusterNode, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.nodes, nil
}

func (m *mockStore) UpdateNode(ctx context.Context, node *ClusterNode) error {
	if m.err != nil {
		return m.err
	}
	for i, n := range m.nodes {
		if n.ID == node.ID {
			m.nodes[i] = *node
			return nil
		}
	}
	return nil
}

func (m *mockStore) DeleteNode(ctx context.Context, id string) error {
	if m.err != nil {
		return m.err
	}
	for i, n := range m.nodes {
		if n.ID == id {
			m.nodes = append(m.nodes[:i], m.nodes[i+1:]...)
			return nil
		}
	}
	return nil
}

func (m *mockStore) CreateConfig(ctx context.Context, config *ClusterConfig) error {
	if m.err != nil {
		return m.err
	}
	m.configs = append(m.configs, *config)
	return nil
}

func (m *mockStore) GetConfig(ctx context.Context, id string) (*ClusterConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.configs {
		if c.ID == id {
			return &c, nil
		}
	}
	return nil, nil
}

func (m *mockStore) ListConfigs(ctx context.Context) ([]ClusterConfig, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.configs, nil
}

func (m *mockStore) UpdateConfig(ctx context.Context, config *ClusterConfig) error {
	if m.err != nil {
		return m.err
	}
	for i, c := range m.configs {
		if c.ID == config.ID {
			m.configs[i] = *config
			return nil
		}
	}
	return nil
}

func (m *mockStore) CreateHealthCheck(ctx context.Context, check *HealthCheck) error {
	if m.err != nil {
		return m.err
	}
	m.healthChecks = append(m.healthChecks, *check)
	return nil
}

func (m *mockStore) ListHealthChecks(ctx context.Context, nodeID string) ([]HealthCheck, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []HealthCheck
	for _, h := range m.healthChecks {
		if h.NodeID == nodeID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (m *mockStore) TryAcquireLock(ctx context.Context, leaderID string, leaseDuration time.Duration) (bool, error) {
	return true, nil
}

func (m *mockStore) ReleaseLock(ctx context.Context, leaderID string) error {
	return nil
}

func (m *mockStore) GetLeader(ctx context.Context) (*LeaderElection, error) {
	return m.leader, nil
}

func TestService_AddNode(t *testing.T) {
	store := &mockStore{}
	svc := NewService(store)
	ctx := context.Background()

	node := &ClusterNode{
		ID:   "node-1",
		Name: "Node 1",
		Host: "192.168.1.10",
		Port: 8080,
		Role: "primary",
	}

	err := svc.AddNode(ctx, node)
	assert.NoError(t, err)
	assert.Len(t, store.nodes, 1)
	assert.Equal(t, "Node 1", store.nodes[0].Name)
}

func TestService_GetNode(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Name: "Node 1", Host: "10.0.0.1"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	node, err := svc.GetNode(ctx, "node-1")
	assert.NoError(t, err)
	assert.Equal(t, "Node 1", node.Name)
}

func TestService_ListNode(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Name: "Node 1"},
			{ID: "node-2", Name: "Node 2"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	nodes, err := svc.ListNodes(ctx)
	assert.NoError(t, err)
	assert.Len(t, nodes, 2)
}

func TestService_UpdateNode(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Name: "Node 1", Host: "10.0.0.1"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	err := svc.UpdateNode(ctx, &ClusterNode{ID: "node-1", Name: "Node 1 Updated", Host: "10.0.0.2"})
	assert.NoError(t, err)
	assert.Equal(t, "10.0.0.2", store.nodes[0].Host)
}

func TestService_DeleteNode(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Name: "Node 1"},
			{ID: "node-2", Name: "Node 2"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	err := svc.DeleteNode(ctx, "node-1")
	assert.NoError(t, err)
	assert.Len(t, store.nodes, 1)
	assert.Equal(t, "node-2", store.nodes[0].ID)
}

func TestService_GetConfig_DefaultWhenEmpty(t *testing.T) {
	store := &mockStore{}
	svc := NewService(store)
	ctx := context.Background()

	config, err := svc.GetConfig(ctx)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, "default", config.ID)
	assert.True(t, config.EnableFailover)
	assert.Equal(t, "round_robin", config.BalanceAlgorithm)
}

func TestService_GetConfig_Existing(t *testing.T) {
	store := &mockStore{
		configs: []ClusterConfig{
			{ID: "cfg-1", EnableFailover: false, HealthCheckInterval: 30},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	config, err := svc.GetConfig(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "cfg-1", config.ID)
	assert.False(t, config.EnableFailover)
	assert.Equal(t, 30, config.HealthCheckInterval)
}

func TestService_UpdateConfig(t *testing.T) {
	store := &mockStore{
		configs: []ClusterConfig{
			{ID: "cfg-1", EnableFailover: true},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	err := svc.UpdateConfig(ctx, &ClusterConfig{ID: "cfg-1", EnableFailover: false})
	assert.NoError(t, err)
	assert.False(t, store.configs[0].EnableFailover)
}

func TestService_PerformHealthCheck(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Name: "Node 1", Status: "online"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	check, err := svc.PerformHealthCheck(ctx, "node-1")
	assert.NoError(t, err)
	assert.NotNil(t, check)
	assert.Equal(t, "node-1", check.NodeID)
	assert.Equal(t, "healthy", check.Status)
	assert.Equal(t, "general", check.CheckType)
	assert.Len(t, store.healthChecks, 1)
}

func TestService_ListHealthChecks(t *testing.T) {
	store := &mockStore{
		healthChecks: []HealthCheck{
			{ID: "h-1", NodeID: "node-1", Status: "healthy"},
			{ID: "h-2", NodeID: "node-1", Status: "healthy"},
			{ID: "h-3", NodeID: "node-2", Status: "unhealthy"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	checks, err := svc.ListHealthChecks(ctx, "node-1")
	assert.NoError(t, err)
	assert.Len(t, checks, 2)

	checks2, err := svc.ListHealthChecks(ctx, "node-2")
	assert.NoError(t, err)
	assert.Len(t, checks2, 1)
}

func TestService_GetClusterStatus_AllOnline(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Status: "online"},
			{ID: "node-2", Status: "online"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	status, err := svc.GetClusterStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, status["total_nodes"])
	assert.Equal(t, 2, status["online_nodes"])
	assert.Equal(t, 0, status["offline_nodes"])
	assert.Equal(t, "healthy", status["cluster_status"])
}

func TestService_GetClusterStatus_Mixed(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Status: "online"},
			{ID: "node-2", Status: "offline"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	status, err := svc.GetClusterStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 2, status["total_nodes"])
	assert.Equal(t, 1, status["online_nodes"])
	assert.Equal(t, 1, status["offline_nodes"])
	assert.Equal(t, "degraded", status["cluster_status"])
}

func TestService_GetClusterStatus_AllOffline(t *testing.T) {
	store := &mockStore{
		nodes: []ClusterNode{
			{ID: "node-1", Status: "offline"},
			{ID: "node-2", Status: "offline"},
		},
	}
	svc := NewService(store)
	ctx := context.Background()

	status, err := svc.GetClusterStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, "down", status["cluster_status"])
}

func TestService_GetClusterStatus_Empty(t *testing.T) {
	store := &mockStore{}
	svc := NewService(store)
	ctx := context.Background()

	status, err := svc.GetClusterStatus(ctx)
	assert.NoError(t, err)
	assert.Equal(t, 0, status["total_nodes"])
	assert.Equal(t, "healthy", status["cluster_status"])
}
