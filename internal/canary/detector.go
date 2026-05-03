package canary

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/chennqqi/godnslog/internal/interaction"
)

// Detector detects canary token hits from interactions
type Detector struct {
	config *CanaryConfig
	store  Store
}

// NewDetector creates a new canary detector
func NewDetector(config *CanaryConfig, store Store) *Detector {
	if config == nil {
		config = DefaultCanaryConfig()
	}
	return &Detector{
		config: config,
		store:  store,
	}
}

// DefaultCanaryConfig returns default canary configuration
func DefaultCanaryConfig() *CanaryConfig {
	return &CanaryConfig{
		MaxRetentionDays:     90,
		DefaultExpiry:        "90d",
		SilentWindow:         300, // 5 minutes
		CompressionThreshold: 10,
		NotificationLevels:   []string{"medium", "high", "critical"},
	}
}

// Detect checks if an interaction matches a canary token
func (d *Detector) Detect(ctx context.Context, inter interaction.Interaction) (*CanaryHit, error) {
	// Get active canaries
	canaries, err := d.store.GetActiveCanaries(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get canaries: %w", err)
	}

	// Check each canary
	for _, canary := range canaries {
		if d.matchesCanary(inter, canary) {
			// Create hit record
			hit := &CanaryHit{
				ID:        generateHitID(),
				CanaryID:  canary.ID,
				SourceIP:  inter.SourceIP,
				UserAgent: "",
				Timestamp: time.Now(),
			}

			// Extract additional info
			if inter.UserAgent != nil {
				hit.UserAgent = *inter.UserAgent
			}

			if inter.Headers != nil {
				headersJSON, _ := json.Marshal(inter.Headers)
				hit.Headers = string(headersJSON)
			}

			if inter.Body != nil {
				hit.Body = *inter.Body
			}

			// Check silent window
			if d.isInSilentWindow(ctx, canary.ID) {
				hit.IsCompressed = true
			}

			// Save hit
			if err := d.store.SaveCanaryHit(ctx, hit); err != nil {
				return nil, fmt.Errorf("failed to save canary hit: %w", err)
			}

			return hit, nil
		}
	}

	return nil, nil
}

// matchesCanary checks if interaction matches canary token
func (d *Detector) matchesCanary(inter interaction.Interaction, canary Canary) bool {
	switch canary.Type {
	case string(CanaryTypeDNS):
		if inter.Type == "dns" && inter.Domain != nil {
			return *inter.Domain == canary.Token || d.containsToken(*inter.Domain, canary.Token)
		}
	case string(CanaryTypeHTTP):
		if inter.Type == "http" {
			// Check in path, headers, or body
			if inter.Path != nil && d.containsToken(*inter.Path, canary.Token) {
				return true
			}
			if inter.Headers != nil {
				for _, v := range inter.Headers {
					if d.containsToken(v, canary.Token) {
						return true
					}
				}
			}
			if inter.Body != nil && d.containsToken(*inter.Body, canary.Token) {
				return true
			}
		}
	}

	return false
}

// containsToken checks if token is contained in the string
func (d *Detector) containsToken(s, token string) bool {
	// Simple contains check (in production, use more sophisticated matching)
	return len(s) > 0 && len(token) > 0 && s == token
}

// isInSilentWindow checks if we're in silent window for this canary
func (d *Detector) isInSilentWindow(ctx context.Context, canaryID string) bool {
	// Get recent hits
	recentHits, err := d.store.GetRecentCanaryHits(ctx, canaryID, d.config.SilentWindow)
	if err != nil {
		return false
	}

	return len(recentHits) > 0
}

// DecodeContext decodes the canary context
func DecodeContext(encoded string) (*CanaryContext, error) {
	// Base64 decode
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("failed to decode context: %w", err)
	}

	// JSON unmarshal
	var context CanaryContext
	if err := json.Unmarshal(decoded, &context); err != nil {
		return nil, fmt.Errorf("failed to unmarshal context: %w", err)
	}

	return &context, nil
}

// EncodeContext encodes the canary context
func EncodeContext(context *CanaryContext) (string, error) {
	// JSON marshal
	jsonData, err := json.Marshal(context)
	if err != nil {
		return "", fmt.Errorf("failed to marshal context: %w", err)
	}

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(jsonData)
	return encoded, nil
}

// AssessRisk assesses the risk level of a canary hit
func (d *Detector) AssessRisk(hit *CanaryHit, canary *Canary) string {
	// Risk assessment logic
	// In production, use more sophisticated risk scoring

	// High risk indicators
	if hit.SourceIP == "127.0.0.1" || hit.SourceIP == "::1" {
		return "critical" // Local access
	}

	if hit.UserAgent != "" && (hit.UserAgent == "curl" || hit.UserAgent == "wget") {
		return "high" // Command line tools
	}

	// Medium risk
	if hit.UserAgent != "" {
		return "medium"
	}

	// Low risk
	return "low"
}

// CompressHits compresses old canary hits
func (d *Detector) CompressHits(ctx context.Context) error {
	canaries, err := d.store.GetAllCanaries(ctx)
	if err != nil {
		return fmt.Errorf("failed to get canaries: %w", err)
	}

	for _, canary := range canaries {
		hits, err := d.store.GetCanaryHits(ctx, canary.ID)
		if err != nil {
			continue
		}

		// Compress if threshold exceeded
		if len(hits) > d.config.CompressionThreshold {
			// Compress older hits
			for i := 0; i < d.config.CompressionThreshold; i++ {
				hits[i].IsCompressed = true
				if err := d.store.UpdateCanaryHit(ctx, &hits[i]); err != nil {
					continue
				}
			}
		}
	}

	return nil
}

// CleanupExpiredCanaries removes expired canaries
func (d *Detector) CleanupExpiredCanaries(ctx context.Context) error {
	return d.store.DeleteExpiredCanaries(ctx, d.config.MaxRetentionDays)
}

// generateHitID generates a unique hit ID
func generateHitID() string {
	return fmt.Sprintf("hit-%d", time.Now().UnixNano())
}
