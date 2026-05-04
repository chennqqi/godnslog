package ha

import (
	"testing"
	"time"
)

// TestClusterNodeModel tests cluster node model
func TestClusterNodeModel(t *testing.T) {
	now := time.Now()
	node := &ClusterNode{
		ID:         "node-1",
		Name:       "Primary Node",
		Host:       "192.168.1.1",
		Port:       8080,
		Role:       "primary",
		Status:     "online",
		LastPing:   now,
		LoadCPU:    50.0,
		LoadMemory: 60.0,
		LoadDisk:   40.0,
		IsEnabled:  true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	if node.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if node.Host == "" {
		t.Fatal("Host should not be empty")
	}

	if node.Role == "" {
		t.Fatal("Role should not be empty")
	}
}

// TestClusterConfigModel tests cluster config model
func TestClusterConfigModel(t *testing.T) {
	now := time.Now()
	config := &ClusterConfig{
		ID:                "config-1",
		EnableFailover:    true,
		FailoverTimeout:   30,
		EnableLoadBalance: true,
		BalanceAlgorithm:  "round_robin",
		EnableReplication: false,
		ReplicationMode:   "async",
		HealthCheckInterval: 10,
		HealthCheckTimeout: 5,
		EnableQuorum:      false,
		QuorumSize:        2,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	if config.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if config.FailoverTimeout < 0 {
		t.Fatal("FailoverTimeout should be non-negative")
	}
}

// TestHealthCheckModel tests health check model
func TestHealthCheckModel(t *testing.T) {
	now := time.Now()
	check := &HealthCheck{
		ID:           "health-1",
		NodeID:       "node-1",
		CheckType:    "dns",
		Status:       "healthy",
		ResponseTime: 10,
		Timestamp:    now,
	}

	if check.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if check.NodeID == "" {
		t.Fatal("NodeID should not be empty")
	}

	if check.Status == "" {
		t.Fatal("Status should not be empty")
	}
}

// TestTableName tests table names
func TestTableName(t *testing.T) {
	node := ClusterNode{}
	if node.TableName() != "cluster_nodes" {
		t.Fatalf("Expected 'cluster_nodes', got '%s'", node.TableName())
	}

	config := ClusterConfig{}
	if config.TableName() != "cluster_configs" {
		t.Fatalf("Expected 'cluster_configs', got '%s'", config.TableName())
	}

	check := HealthCheck{}
	if check.TableName() != "health_checks" {
		t.Fatalf("Expected 'health_checks', got '%s'", check.TableName())
	}
}
