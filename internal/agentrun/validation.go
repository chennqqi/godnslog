package agentrun

import (
	"fmt"
	"net"
	"net/url"
	"strings"
)

// ValidateWebhookURL checks if the webhook URL is safe for delivery
// Only https:// is allowed by default
// Rejects localhost, link-local, private RFC1918 ranges, metadata IPs
func ValidateWebhookURL(webhookURL string) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL cannot be empty")
	}

	parsedURL, err := url.Parse(webhookURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}

	// Parse host
	host := parsedURL.Hostname()
	if host == "" {
		return fmt.Errorf("URL must have a hostname")
	}

	// Reject cloud metadata IPs first (before link-local check)
	if isMetadataIP(host) {
		return fmt.Errorf("metadata IPs are not allowed")
	}

	// Reject localhost variants
	if isLocalhost(host) {
		return fmt.Errorf("localhost URLs are not allowed")
	}

	// Reject link-local addresses
	if isLinkLocal(host) {
		return fmt.Errorf("link-local addresses are not allowed")
	}

	// Reject private RFC1918 ranges
	if isPrivateIP(host) {
		return fmt.Errorf("private IP addresses are not allowed")
	}

	// Only allow https:// by default
	if parsedURL.Scheme != "https" {
		return fmt.Errorf("only https:// URLs are allowed, got: %s", parsedURL.Scheme)
	}

	return nil
}

// ValidateWebhookHeaders checks if headers are safe for delivery
// Only allows Content-Type and X-* headers
// Rejects Authorization, Cookie, Set-Cookie, Proxy-*, and hop-by-hop headers
func ValidateWebhookHeaders(headers map[string]string) error {
	forbiddenHeaders := map[string]bool{
		"authorization":       true,
		"cookie":              true,
		"set-cookie":          true,
		"proxy-authorization": true,
		"proxy-authenticate":  true,
		"te":                  true, // hop-by-hop
		"trailer":             true, // hop-by-hop
		"transfer-encoding":   true, // hop-by-hop
		"upgrade":             true, // hop-by-hop
	}

	for headerName := range headers {
		lowerName := strings.ToLower(headerName)

		// Check forbidden headers
		if forbiddenHeaders[lowerName] {
			return fmt.Errorf("header '%s' is not allowed", headerName)
		}

		// Only allow Content-Type or X-* headers
		if lowerName != "content-type" && !strings.HasPrefix(lowerName, "x-") {
			return fmt.Errorf("header '%s' is not allowed (only Content-Type and X-* headers are permitted)", headerName)
		}
	}

	return nil
}

// isLocalhost checks if the host is a localhost variant
func isLocalhost(host string) bool {
	lowerHost := strings.ToLower(host)
	return lowerHost == "localhost" ||
		lowerHost == "127.0.0.1" ||
		lowerHost == "::1" ||
		strings.HasSuffix(lowerHost, ".localhost")
}

// isLinkLocal checks if the host is a link-local address
func isLinkLocal(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}
	return ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast()
}

// isPrivateIP checks if the host is a private RFC1918 address
func isPrivateIP(host string) bool {
	ip := net.ParseIP(host)
	if ip == nil {
		return false
	}

	// Check for private IPv4 ranges
	if ipv4 := ip.To4(); ipv4 != nil {
		// 10.0.0.0/8
		if ipv4[0] == 10 {
			return true
		}
		// 172.16.0.0/12
		if ipv4[0] == 172 && ipv4[1] >= 16 && ipv4[1] <= 31 {
			return true
		}
		// 192.168.0.0/16
		if ipv4[0] == 192 && ipv4[1] == 168 {
			return true
		}
	}

	// Check for private IPv6 ranges (fc00::/7)
	return ip.IsPrivate()
}

// isMetadataIP checks if the host is a cloud metadata IP or hostname
func isMetadataIP(host string) bool {
	// Check for AWS metadata IP first
	ip := net.ParseIP(host)
	if ip != nil {
		if ipv4 := ip.To4(); ipv4 != nil {
			if ipv4[0] == 169 && ipv4[1] == 254 && ipv4[2] == 169 && ipv4[3] == 254 {
				return true
			}
		}
	}

	// GCP metadata IP: metadata.google.internal (resolves to various IPs)
	// We'll check for common metadata hostnames
	lowerHost := strings.ToLower(host)
	if lowerHost == "metadata.google.internal" ||
		lowerHost == "metadata" ||
		strings.HasSuffix(lowerHost, ".metadata.google.internal") {
		return true
	}

	return false
}
