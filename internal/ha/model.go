package ha

import "time"

// ClusterNode represents a node in the cluster
type ClusterNode struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Name        string    `json:"name"`
	Host        string    `json:"host"`
	Port        int       `json:"port"`
	Role        string    `json:"role"` // primary, secondary, standby
	
	// Health status
	Status      string    `json:"status"` // online, offline, degraded
	LastPing    time.Time `json:"last_ping"`
	
	// Load metrics
	LoadCPU     float64   `json:"load_cpu"`
	LoadMemory  float64   `json:"load_memory"`
	LoadDisk    float64   `json:"load_disk"`
	
	IsEnabled   bool      `json:"is_enabled"`
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
}

// TableName returns the table name for ClusterNode
func (ClusterNode) TableName() string {
	return "cluster_nodes"
}

// ClusterConfig represents cluster configuration
type ClusterConfig struct {
	ID                string    `json:"id" xorm:"'id' pk"`
	
	// Failover settings
	EnableFailover    bool      `json:"enable_failover"`
	FailoverTimeout   int       `json:"failover_timeout"` // seconds
	
	// Load balancing
	EnableLoadBalance bool      `json:"enable_load_balance"`
	BalanceAlgorithm  string    `json:"balance_algorithm"` // round_robin, least_connections, ip_hash
	
	// Replication settings
	EnableReplication bool      `json:"enable_replication"`
	ReplicationMode   string    `json:"replication_mode"` // sync, async
	
	// Health check settings
	HealthCheckInterval int      `json:"health_check_interval"` // seconds
	HealthCheckTimeout  int      `json:"health_check_timeout"` // seconds
	
	// Quorum settings
	EnableQuorum      bool      `json:"enable_quorum"`
	QuorumSize        int       `json:"quorum_size"` // minimum nodes required
	
	CreatedAt         time.Time `json:"created_at" xorm:"created"`
	UpdatedAt         time.Time `json:"updated_at" xorm:"updated"`
}

// TableName returns the table name for ClusterConfig
func (ClusterConfig) TableName() string {
	return "cluster_configs"
}

// HealthCheck represents a health check result
type HealthCheck struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	NodeID      string    `json:"node_id"`
	
	// Check details
	CheckType   string    `json:"check_type"` // dns, http, database, listener
	Status      string    `json:"status"` // healthy, unhealthy, warning
	
	// Response time
	ResponseTime int64    `json:"response_time"` // milliseconds
	
	// Error message
	Error       string    `json:"error"`
	
	Timestamp   time.Time `json:"timestamp" xorm:"created"`
}

// TableName returns the table name for HealthCheck
func (HealthCheck) TableName() string {
	return "health_checks"
}
