package clustering

import (
	"github.com/chennqqi/godnslog/internal/interaction"
)

// Compressor reduces interaction data size by compressing similar interactions
type Compressor struct {
	config *CompressionConfig
}

// CompressionConfig holds configuration for interaction compression
type CompressionConfig struct {
	MaxRawDataLength int    `json:"max_raw_data_length"` // Max length for raw data
	CompressHeaders  bool   `json:"compress_headers"`     // Compress header data
	RemoveDuplicates bool  `json:"remove_duplicates"`     // Remove duplicate interactions
	KeepFirstN       int    `json:"keep_first_n"`          // Keep first N interactions per cluster
}

// DefaultCompressionConfig returns default compression configuration
func DefaultCompressionConfig() *CompressionConfig {
	return &CompressionConfig{
		MaxRawDataLength: 1024,
		CompressHeaders:  true,
		RemoveDuplicates: true,
		KeepFirstN:       10,
	}
}

// NewCompressor creates a new compressor
func NewCompressor(config *CompressionConfig) *Compressor {
	if config == nil {
		config = DefaultCompressionConfig()
	}
	return &Compressor{config: config}
}

// CompressInteractions compresses interaction data
func (c *Compressor) CompressInteractions(interactions []interaction.Interaction) []interaction.Interaction {
	if c.config.RemoveDuplicates {
		interactions = c.removeDuplicates(interactions)
	}
	
	for i := range interactions {
		interactions[i] = c.compressInteraction(interactions[i])
	}
	
	return interactions
}

// compressInteraction compresses a single interaction
func (c *Compressor) compressInteraction(inter interaction.Interaction) interaction.Interaction {
	// Truncate raw data
	if len(inter.RawData) > c.config.MaxRawDataLength {
		inter.RawData = inter.RawData[:c.config.MaxRawDataLength] + "... (truncated)"
	}
	
	// Compress headers
	if c.config.CompressHeaders && len(inter.Headers) > 0 {
		inter.Headers = c.compressHeaders(inter.Headers)
	}
	
	return inter
}

// compressHeaders compresses header data
func (c *Compressor) compressHeaders(headers interaction.Headers) interaction.Headers {
	// Keep only important headers
	importantHeaders := map[string]string{
		"User-Agent":      "",
		"Content-Type":    "",
		"Authorization":  "",
		"Cookie":         "",
		"Referer":        "",
		"X-Forwarded-For": "",
	}
	
	compressed := make(interaction.Headers)
	for k, v := range headers {
		if _, important := importantHeaders[k]; important {
			compressed[k] = v
		}
	}
	
	return compressed
}

// removeDuplicates removes duplicate interactions
func (c *Compressor) removeDuplicates(interactions []interaction.Interaction) []interaction.Interaction {
	seen := make(map[string]bool)
	result := make([]interaction.Interaction, 0)
	
	for _, inter := range interactions {
		key := generateInteractionKey(inter)
		if !seen[key] {
			seen[key] = true
			result = append(result, inter)
		}
	}
	
	return result
}

// generateInteractionKey generates a unique key for an interaction
func generateInteractionKey(inter interaction.Interaction) string {
	key := inter.Type + ":" + inter.SourceIP
	if inter.Token != nil {
		key += ":" + *inter.Token
	}
	if inter.Path != nil {
		key += ":" + *inter.Path
	}
	return key
}

// CompressCluster compresses a cluster by keeping only representative interactions
func (c *Compressor) CompressCluster(cluster *Cluster) *Cluster {
	if len(cluster.Interactions) <= c.config.KeepFirstN {
		return cluster
	}
	
	// Keep first N interactions
	cluster.Interactions = cluster.Interactions[:c.config.KeepFirstN]
	cluster.Count = c.config.KeepFirstN
	
	return cluster
}
