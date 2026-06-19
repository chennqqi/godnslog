package clustering

import (
	"testing"
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
)

// TestClusterInteractionsBasic tests basic clustering by type+IP+token
func TestClusterInteractionsBasic(t *testing.T) {
	clusterer := NewClusterer(nil)

	token1 := "token-abc"
	token2 := "token-xyz"
	interactions := []interaction.Interaction{
		{
			ID:        "i-1",
			Type:      "dns",
			SourceIP:  "1.2.3.4",
			Token:     &token1,
			Timestamp: time.Now(),
			Domain:    &token1,
		},
		{
			ID:        "i-2",
			Type:      "dns",
			SourceIP:  "1.2.3.4",
			Token:     &token1,
			Timestamp: time.Now(),
			Domain:    &token1,
		},
		{
			ID:        "i-3",
			Type:      "http",
			SourceIP:  "5.6.7.8",
			Token:     &token2,
			Timestamp: time.Now(),
			Path:      strPtr("/admin"),
		},
	}

	clusters := clusterer.ClusterInteractions(interactions)

	if len(clusters) != 2 {
		t.Fatalf("Expected 2 clusters, got %d", len(clusters))
	}

	// Find the DNS cluster (should have 2 interactions)
	var dnsCluster *Cluster
	for _, c := range clusters {
		if c.Type == "dns" {
			dnsCluster = c
			break
		}
	}
	if dnsCluster == nil {
		t.Fatal("Expected to find a DNS cluster")
	}
	if dnsCluster.Count != 2 {
		t.Errorf("Expected DNS cluster count 2, got %d", dnsCluster.Count)
	}
}

// TestClusterNoiseDetectionByCount tests noise detection via high frequency threshold
func TestClusterNoiseDetectionByCount(t *testing.T) {
	config := &ClusteringConfig{
		MaxClusterSize: 100,
		TimeWindow:     "5m",
		NoiseThreshold: 5,
		NoisePatterns:  []NoisePattern{},
	}
	clusterer := NewClusterer(config)

	token := "token-noisy"
	interactions := make([]interaction.Interaction, 10)
	for i := range interactions {
		interactions[i] = interaction.Interaction{
			ID:        "i-" + string(rune('a'+i)),
			Type:      "dns",
			SourceIP:  "1.1.1.1",
			Token:     &token,
			Timestamp: time.Now(),
			Domain:    &token,
		}
	}

	clusters := clusterer.ClusterInteractions(interactions)

	if len(clusters) != 1 {
		t.Fatalf("Expected 1 cluster, got %d", len(clusters))
	}

	if !clusters[0].IsNoise {
		t.Error("Expected cluster to be marked as noise due to high frequency")
	}
	if clusters[0].NoiseReason != "High frequency" {
		t.Errorf("Expected noise reason 'High frequency', got '%s'", clusters[0].NoiseReason)
	}
}

// TestClusterNoiseDetectionByPattern tests noise detection via regex pattern matching
func TestClusterNoiseDetectionByPattern(t *testing.T) {
	config := &ClusteringConfig{
		MaxClusterSize: 100,
		TimeWindow:     "5m",
		NoiseThreshold: 100,
		NoisePatterns: []NoisePattern{
			{
				Type:        "dns",
				Pattern:     `.*\.google\.com$`,
				Description: "Google DNS queries",
			},
		},
	}
	clusterer := NewClusterer(config)

	googleDomain := "something.google.com"
	token := "token-google"
	interactions := []interaction.Interaction{
		{
			ID:        "i-1",
			Type:      "dns",
			SourceIP:  "8.8.8.8",
			Token:     &token,
			Timestamp: time.Now(),
			Domain:    &googleDomain,
		},
	}

	clusters := clusterer.ClusterInteractions(interactions)

	if len(clusters) != 1 {
		t.Fatalf("Expected 1 cluster, got %d", len(clusters))
	}

	if !clusters[0].IsNoise {
		t.Error("Expected cluster to be marked as noise due to google.com pattern")
	}
	if clusters[0].NoiseReason != "Google DNS queries" {
		t.Errorf("Expected noise reason 'Google DNS queries', got '%s'", clusters[0].NoiseReason)
	}
}

// TestCompressorRemoveDuplicates tests deduplication of identical interactions
func TestCompressorRemoveDuplicates(t *testing.T) {
	config := &CompressionConfig{
		MaxRawDataLength: 1024,
		CompressHeaders:  false,
		RemoveDuplicates: true,
		KeepFirstN:       10,
	}
	compressor := NewCompressor(config)

	token := "token-dup"
	path := "/api/data"
	interactions := []interaction.Interaction{
		{ID: "i-1", Type: "http", SourceIP: "1.2.3.4", Token: &token, Path: &path},
		{ID: "i-2", Type: "http", SourceIP: "1.2.3.4", Token: &token, Path: &path},
		{ID: "i-3", Type: "http", SourceIP: "1.2.3.4", Token: &token, Path: &path},
	}

	result := compressor.CompressInteractions(interactions)

	if len(result) != 1 {
		t.Fatalf("Expected 1 interaction after dedup, got %d", len(result))
	}
}

// TestCompressorTruncateRawData tests raw data truncation
func TestCompressorTruncateRawData(t *testing.T) {
	config := &CompressionConfig{
		MaxRawDataLength: 10,
		CompressHeaders:  false,
		RemoveDuplicates: false,
		KeepFirstN:       10,
	}
	compressor := NewCompressor(config)

	longData := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	interactions := []interaction.Interaction{
		{ID: "i-1", Type: "http", SourceIP: "1.2.3.4", RawData: longData},
	}

	result := compressor.CompressInteractions(interactions)

	if len(result) != 1 {
		t.Fatalf("Expected 1 interaction, got %d", len(result))
	}
	if len(result[0].RawData) > 30 {
		t.Errorf("Expected raw data to be truncated, got length %d", len(result[0].RawData))
	}
	if !contains(result[0].RawData, "truncated") {
		t.Errorf("Expected truncated marker in raw data, got: %s", result[0].RawData)
	}
}

// TestCompressorCompressCluster tests cluster compression keeping only N interactions
func TestCompressorCompressCluster(t *testing.T) {
	config := &CompressionConfig{
		MaxRawDataLength: 1024,
		CompressHeaders:  false,
		RemoveDuplicates: false,
		KeepFirstN:       3,
	}
	compressor := NewCompressor(config)

	interactions := make([]interaction.Interaction, 10)
	for i := range interactions {
		interactions[i] = interaction.Interaction{
			ID:   "i-" + string(rune('a'+i)),
			Type: "dns",
		}
	}

	cluster := &Cluster{
		ID:           "cluster-1",
		Type:         "dns",
		Count:        10,
		Interactions: interactions,
	}

	compressed := compressor.CompressCluster(cluster)

	if len(compressed.Interactions) != 3 {
		t.Fatalf("Expected 3 interactions after compression, got %d", len(compressed.Interactions))
	}
	if compressed.Count != 3 {
		t.Errorf("Expected count 3 after compression, got %d", compressed.Count)
	}
}

// TestDefaultClusteringConfig tests that default config is sensible
func TestDefaultClusteringConfig(t *testing.T) {
	config := DefaultClusteringConfig()

	if config.MaxClusterSize != 100 {
		t.Errorf("Expected max cluster size 100, got %d", config.MaxClusterSize)
	}
	if config.NoiseThreshold != 10 {
		t.Errorf("Expected noise threshold 10, got %d", config.NoiseThreshold)
	}
	if len(config.NoisePatterns) == 0 {
		t.Error("Expected default noise patterns to be non-empty")
	}
}

// strPtr helper
func strPtr(s string) *string {
	return &s
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
