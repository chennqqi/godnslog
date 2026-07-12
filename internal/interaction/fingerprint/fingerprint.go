package fingerprint

import (
	"net"
	"strings"
)

// Source type constants
const (
	SourceScanner = "scanner"
	SourceCloud   = "cloud"
	SourceKnown   = "known"
	SourceUnknown = "unknown"
)

// Fingerprint represents the source fingerprint of an IP
type Fingerprint struct {
	IP         string `json:"ip"`
	SourceType string `json:"source_type"`
	SourceName string `json:"source_name,omitempty"`
	ASN        uint   `json:"asn,omitempty"`
	Org        string `json:"org,omitempty"`
	Country    string `json:"country,omitempty"`
}

// Fingerprinter identifies the source of an IP address
type Fingerprinter struct {
	geo *GeoIP // optional, nil if no GeoIP DB loaded
}

// NewFingerprinter creates a new fingerprinter.
// mmdbPath is optional path to GeoLite2-ASN.mmdb; empty string disables GeoIP.
func NewFingerprinter(mmdbPath string) *Fingerprinter {
	var geo *GeoIP
	if mmdbPath != "" {
		var err error
		geo, err = NewGeoIP(mmdbPath)
		if err != nil {
			geo = nil // graceful degradation
		}
	}
	return &Fingerprinter{geo: geo}
}

// Lookup identifies the source of an IP address.
// Priority: scanner > cloud > geoip > unknown
func (f *Fingerprinter) Lookup(ipStr, userAgent string) *Fingerprint {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return &Fingerprint{IP: ipStr, SourceType: SourceUnknown}
	}

	fp := &Fingerprint{IP: ipStr}

	// 1. Check scanner rules (highest priority)
	if name := matchScanner(ip, userAgent); name != "" {
		fp.SourceType = SourceScanner
		fp.SourceName = name
		return fp
	}

	// 2. Check cloud providers
	if name := matchCloud(ip); name != "" {
		fp.SourceType = SourceCloud
		fp.SourceName = name
		return fp
	}

	// 3. Optional GeoIP enrichment
	if f.geo != nil {
		if asn, org, country := f.geo.Lookup(ip); asn > 0 {
			fp.SourceType = SourceUnknown
			fp.ASN = asn
			fp.Org = org
			fp.Country = country
			return fp
		}
	}

	// 4. Unknown
	fp.SourceType = SourceUnknown
	return fp
}

// matchScanner checks if the IP or User-Agent matches known scanners
func matchScanner(ip net.IP, userAgent string) string {
	// Check by IP against scanner CIDRs
	for _, entry := range scannerCIDRs {
		if entry.cidr.Contains(ip) {
			return entry.name
		}
	}

	// Check by User-Agent keywords
	ua := strings.ToLower(userAgent)
	for _, entry := range scannerUAKeywords {
		if strings.Contains(ua, entry.keyword) {
			return entry.name
		}
	}

	return ""
}

// matchCloud checks if the IP belongs to a known cloud provider
func matchCloud(ip net.IP) string {
	for _, entry := range cloudCIDRs {
		if entry.cidr.Contains(ip) {
			return entry.name
		}
	}
	return ""
}
