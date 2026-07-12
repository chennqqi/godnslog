package fingerprint

import (
	"testing"
)

func TestScannerDetectionByIP(t *testing.T) {
	f := NewFingerprinter("")

	tests := []struct {
		ip       string
		ua       string
		wantType string
		wantName string
	}{
		{"141.98.10.1", "", SourceScanner, "Shodan"},
		{"162.142.125.1", "", SourceScanner, "Censys"},
		{"104.16.1.1", "", SourceCloud, "Cloudflare"},
		{"52.0.1.1", "", SourceCloud, "AWS"},
		{"34.0.1.1", "", SourceCloud, "Google Cloud"},
		{"13.64.1.1", "", SourceCloud, "Azure"},
		{"8.128.1.1", "", SourceCloud, "Alibaba"},
		{"8.8.8.8", "curl/7.0", SourceUnknown, ""},
		{"1.1.1.1", "", SourceUnknown, ""},
	}

	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			fp := f.Lookup(tt.ip, tt.ua)
			if fp.SourceType != tt.wantType {
				t.Errorf("Lookup(%q).SourceType = %q, want %q", tt.ip, fp.SourceType, tt.wantType)
			}
			if fp.SourceName != tt.wantName {
				t.Errorf("Lookup(%q).SourceName = %q, want %q", tt.ip, fp.SourceName, tt.wantName)
			}
		})
	}
}

func TestScannerDetectionByUA(t *testing.T) {
	f := NewFingerprinter("")

	fp := f.Lookup("1.2.3.4", "Nuclei/3.0")
	if fp.SourceType != SourceScanner || fp.SourceName != "Nuclei" {
		t.Errorf("expected Nuclei scanner, got type=%q name=%q", fp.SourceType, fp.SourceName)
	}

	fp = f.Lookup("1.2.3.4", "Mozilla/5.0 sqlmap/1.5")
	if fp.SourceType != SourceScanner || fp.SourceName != "SQLMap" {
		t.Errorf("expected SQLMap scanner, got type=%q name=%q", fp.SourceType, fp.SourceName)
	}

	fp = f.Lookup("1.2.3.4", "PostmanRuntime/7.0")
	if fp.SourceType != SourceScanner || fp.SourceName != "Postman" {
		t.Errorf("expected Postman scanner, got type=%q name=%q", fp.SourceType, fp.SourceName)
	}

	// Normal browser should not be detected as scanner
	fp = f.Lookup("1.2.3.4", "Mozilla/5.0 Chrome/120")
	if fp.SourceType == SourceScanner {
		t.Errorf("normal browser should not be detected as scanner, got name=%q", fp.SourceName)
	}
}

func TestDetectInvalidIP(t *testing.T) {
	f := NewFingerprinter("")
	fp := f.Lookup("not-an-ip", "")
	if fp.SourceType != SourceUnknown {
		t.Errorf("invalid IP should return unknown, got %q", fp.SourceType)
	}
}
