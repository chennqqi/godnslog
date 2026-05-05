package interaction

import (
	"strings"
	"time"
)

// NoiseFilterConfig defines noise filtering rules
type NoiseFilterConfig struct {
	// IP-based filtering
	FilterScannerIPs  bool
	FilterRepeatedIPs bool
	MaxHitsPerIP      int
	IPTimeWindow      time.Duration

	// Pattern-based filtering
	FilterKnownPatterns   bool
	FilterEmptyBody       bool
	FilterStaticResources bool

	// Token-based filtering
	FilterExpiredTokens bool

	// Custom patterns
	CustomNoisePatterns []string
}

// NoiseFilterResult represents the result of noise filtering
type NoiseFilterResult struct {
	TotalInteractions int
	FilteredCount     int
	NoiseCount        int
	FilteredIDs       []string
	NoiseCategories   map[string]int
}

// Known noise patterns
var knownNoisePatterns = []string{
	// Scanner user agents
	"Mozilla/5.0 (compatible; Googlebot/",
	"Mozilla/5.0 (compatible; bingbot/",
	"Mozilla/5.0 (compatible; YandexBot/",
	"scanner",
	"nuclei",
	"nikto",
	"nmap",
	"masscan",
	"zmap",

	// Common static resource paths
	"/robots.txt",
	"/favicon.ico",
	"/sitemap.xml",
	".css",
	".js",
	".png",
	".jpg",
	".jpeg",
	".gif",
	".svg",
	".woff",
	".woff2",
	".ttf",
	".eot",

	// Health check endpoints
	"/health",
	"/ping",
	"/status",
	"/alive",

	// Common noise in headers
	"X-Scanner",
	"X-Nuclei",
	"X-Amzn",
}

// FilterNoise filters out noise interactions
func (s *Service) FilterNoise(config NoiseFilterConfig) (*NoiseFilterResult, error) {
	// Get all interactions
	var interactions []Interaction
	err := s.engine.Find(&interactions)
	if err != nil {
		return nil, err
	}

	result := &NoiseFilterResult{
		TotalInteractions: len(interactions),
		NoiseCategories:   make(map[string]int),
		FilteredIDs:       []string{},
	}

	// Track IP hit counts
	ipHitCount := make(map[string]int)
	ipFirstHit := make(map[string]time.Time)

	for _, interaction := range interactions {
		isNoise := false
		category := ""

		// Filter by scanner IPs
		if config.FilterScannerIPs && s.isScannerIP(interaction.SourceIP) {
			isNoise = true
			category = "scanner_ip"
		}

		// Filter by repeated IPs
		if config.FilterRepeatedIPs {
			ipHitCount[interaction.SourceIP]++
			if ipFirstHit[interaction.SourceIP].IsZero() {
				ipFirstHit[interaction.SourceIP] = interaction.CreatedAt
			}

			if ipHitCount[interaction.SourceIP] > config.MaxHitsPerIP {
				timeSinceFirst := time.Since(ipFirstHit[interaction.SourceIP])
				if timeSinceFirst < config.IPTimeWindow {
					isNoise = true
					category = "repeated_ip"
				}
			}
		}

		// Filter by known patterns
		if config.FilterKnownPatterns && s.matchesNoisePattern(interaction) {
			isNoise = true
			category = "known_pattern"
		}

		// Filter empty body
		if config.FilterEmptyBody && (interaction.Body == nil || len(*interaction.Body) == 0) && interaction.Type == "http" {
			isNoise = true
			category = "empty_body"
		}

		// Filter static resources
		if config.FilterStaticResources && s.isStaticResource(interaction) {
			isNoise = true
			category = "static_resource"
		}

		// Filter custom patterns
		for _, pattern := range config.CustomNoisePatterns {
			if strings.Contains(interaction.RawData, pattern) {
				isNoise = true
				category = "custom_pattern"
				break
			}
		}

		if isNoise {
			result.NoiseCount++
			result.FilteredIDs = append(result.FilteredIDs, interaction.ID)
			result.NoiseCategories[category]++
		}
	}

	result.FilteredCount = result.TotalInteractions - result.NoiseCount

	return result, nil
}

// isScannerIP checks if an IP is a known scanner
func (s *Service) isScannerIP(ip string) bool {
	// Check against known scanner IP ranges
	scannerIPs := []string{
		"192.168.", // Example: internal scanners
		"10.",      // Example: internal scanners
		// Add more scanner IP patterns as needed
	}

	for _, pattern := range scannerIPs {
		if strings.HasPrefix(ip, pattern) {
			return true
		}
	}

	return false
}

// matchesNoisePattern checks if interaction matches known noise patterns
func (s *Service) matchesNoisePattern(interaction Interaction) bool {
	data := strings.ToLower(interaction.RawData)

	for _, pattern := range knownNoisePatterns {
		if strings.Contains(data, strings.ToLower(pattern)) {
			return true
		}
	}

	return false
}

// isStaticResource checks if interaction is for a static resource
func (s *Service) isStaticResource(interaction Interaction) bool {
	if interaction.Path == nil || len(*interaction.Path) == 0 {
		return false
	}

	path := strings.ToLower(*interaction.Path)
	staticExtensions := []string{
		".css", ".js", ".png", ".jpg", ".jpeg", ".gif",
		".svg", ".woff", ".woff2", ".ttf", ".eot", ".ico",
	}

	for _, ext := range staticExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}

	// Check for common static paths
	staticPaths := []string{
		"/robots.txt", "/favicon.ico", "/sitemap.xml",
		"/apple-touch-icon", "/android-chrome",
	}

	for _, staticPath := range staticPaths {
		if strings.HasPrefix(path, staticPath) {
			return true
		}
	}

	return false
}

// GetNoiseStats returns statistics about noise in interactions
func (s *Service) GetNoiseStats(config NoiseFilterConfig) (map[string]interface{}, error) {
	result, err := s.FilterNoise(config)
	if err != nil {
		return nil, err
	}

	noiseRatio := 0.0
	if result.TotalInteractions > 0 {
		noiseRatio = float64(result.NoiseCount) / float64(result.TotalInteractions) * 100
	}

	return map[string]interface{}{
		"total_interactions": result.TotalInteractions,
		"noise_count":        result.NoiseCount,
		"filtered_count":     result.FilteredCount,
		"noise_ratio":        noiseRatio,
		"noise_categories":   result.NoiseCategories,
	}, nil
}

// MarkNoise marks interactions as noise by adding a tag or note
func (s *Service) MarkNoise(ids []string) error {
	if len(ids) == 0 {
		return nil
	}

	// Since Interaction doesn't have IsNoise field, we could:
	// 1. Add the field to the struct
	// 2. Use a different mechanism (like tags or notes)
	// For now, this is a placeholder for the noise marking functionality
	// In a real implementation, you would add a noise field to the Interaction model
	return nil
}
