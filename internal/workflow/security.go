package workflow

import (
	"fmt"
	"net"
	"net/url"
	"sync"
	"time"
)

// OutboundSecurity enforces security constraints on outbound HTTP requests.
// It provides URL allowlist matching, SSRF protection, and per-action rate limiting.
type OutboundSecurity struct {
	allowlist  []string
	mu         sync.Mutex
	rateCounts map[string][]time.Time
	rateWindow time.Duration
	rateLimit  int
}

// NewOutboundSecurity creates a new OutboundSecurity with the given domain allowlist.
// If allowlist is nil or empty, all outbound requests are denied.
// A wildcard "*" entry allows any host (subject to SSRF checks).
func NewOutboundSecurity(allowlist []string) *OutboundSecurity {
	return &OutboundSecurity{
		allowlist:  allowlist,
		rateCounts: make(map[string][]time.Time),
		rateWindow: time.Minute,
		rateLimit:  10,
	}
}

// IsURLAllowed checks if the URL's host matches an entry in the allowlist.
// Returns false if allowlist is empty (default deny).
func (s *OutboundSecurity) IsURLAllowed(rawURL string) bool {
	if len(s.allowlist) == 0 {
		return false
	}
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := parsed.Hostname()
	for _, allowed := range s.allowlist {
		if host == allowed || allowed == "*" {
			return true
		}
	}
	return false
}

// IsSSRF checks if the URL targets a private, localhost, or cloud metadata endpoint.
// Returns true if the URL is considered a SSRF risk.
func (s *OutboundSecurity) IsSSRF(rawURL string) bool {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return true
	}
	host := parsed.Hostname()
	if host == "localhost" {
		return true
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	privateRanges := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"169.254.0.0/16",
	}
	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(ip) {
			return true
		}
	}
	return false
}

// CheckRate returns true if the action is within the rate limit, false otherwise.
// Uses a sliding window counter per action ID.
func (s *OutboundSecurity) CheckRate(actionID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-s.rateWindow)
	counts := s.rateCounts[actionID]
	filtered := counts[:0]
	for _, t := range counts {
		if t.After(cutoff) {
			filtered = append(filtered, t)
		}
	}
	if len(filtered) >= s.rateLimit {
		s.rateCounts[actionID] = filtered
		return false
	}
	s.rateCounts[actionID] = append(filtered, now)
	return true
}

// ValidateURL performs all security checks: allowlist + SSRF + rate limit.
// Returns an error if any check fails.
func (s *OutboundSecurity) ValidateURL(actionID, rawURL string) error {
	if !s.IsURLAllowed(rawURL) {
		return fmt.Errorf("url not in allowlist: %s", rawURL)
	}
	if s.IsSSRF(rawURL) {
		return fmt.Errorf("url targets blocked private/localhost range: %s", rawURL)
	}
	if !s.CheckRate(actionID) {
		return fmt.Errorf("rate limit exceeded for action: %s", actionID)
	}
	return nil
}
