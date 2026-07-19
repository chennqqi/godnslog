package tlsmanager

import (
	"crypto/tls"
	"os"
	"path/filepath"
	"testing"
)

func TestParseMode(t *testing.T) {
	tests := []struct {
		input   string
		want    Mode
		wantErr bool
	}{
		{"disabled", ModeDisabled, false},
		{"off", ModeDisabled, false},
		{"", ModeDisabled, false},
		{"acme", ModeACME, false},
		{"letsencrypt", ModeACME, false},
		{"static", ModeStatic, false},
		{"manual", ModeStatic, false},
		{"self-signed", ModeSelfSigned, false},
		{"selfsigned", ModeSelfSigned, false},
		{"invalid", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for input %q, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for input %q: %v", tt.input, err)
			}
			if got != tt.want {
				t.Fatalf("ParseMode(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNewManagerDisabled(t *testing.T) {
	m, err := NewManager(Config{Mode: ModeDisabled})
	if err != nil {
		t.Fatalf("NewManager disabled: %v", err)
	}
	if m.IsTLSEnabled() {
		t.Fatal("TLS should not be enabled in disabled mode")
	}
	if m.TLSConfig() != nil {
		t.Fatal("TLSConfig should be nil in disabled mode")
	}
}

func TestNewManagerSelfSigned(t *testing.T) {
	tmpDir := t.TempDir()
	m, err := NewManager(Config{
		Mode:    ModeSelfSigned,
		Domain:  "test.example.com",
		CertDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("NewManager self-signed: %v", err)
	}
	if !m.IsTLSEnabled() {
		t.Fatal("TLS should be enabled in self-signed mode")
	}
	if m.TLSConfig() == nil {
		t.Fatal("TLSConfig should not be nil in self-signed mode")
	}
	if len(m.TLSConfig().Certificates) != 1 {
		t.Fatalf("expected 1 certificate, got %d", len(m.TLSConfig().Certificates))
	}
	if m.TLSConfig().MinVersion != tls.VersionTLS12 {
		t.Fatalf("expected MinVersion TLS 1.2, got %x", m.TLSConfig().MinVersion)
	}

	// Verify cert files were written
	certFile := filepath.Join(tmpDir, "self-signed.crt")
	keyFile := filepath.Join(tmpDir, "self-signed.key")
	if _, err := os.Stat(certFile); err != nil {
		t.Fatalf("cert file not created: %v", err)
	}
	if _, err := os.Stat(keyFile); err != nil {
		t.Fatalf("key file not created: %v", err)
	}
}

func TestSelfSignedCertReuse(t *testing.T) {
	tmpDir := t.TempDir()

	// First creation generates the cert
	_, err := NewManager(Config{
		Mode:    ModeSelfSigned,
		Domain:  "reuse.example.com",
		CertDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("first NewManager: %v", err)
	}

	// Second creation should reuse the cached cert
	m2, err := NewManager(Config{
		Mode:    ModeSelfSigned,
		Domain:  "reuse.example.com",
		CertDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("second NewManager: %v", err)
	}

	if !m2.IsTLSEnabled() {
		t.Fatal("TLS should be enabled on reuse")
	}
}

func TestNewManagerStatic(t *testing.T) {
	// Generate a self-signed cert to use as static
	tmpDir := t.TempDir()
	_, err := NewManager(Config{
		Mode:    ModeSelfSigned,
		Domain:  "static.example.com",
		CertDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("generate cert for static test: %v", err)
	}

	certFile := filepath.Join(tmpDir, "self-signed.crt")
	keyFile := filepath.Join(tmpDir, "self-signed.key")

	m, err := NewManager(Config{
		Mode:     ModeStatic,
		CertFile: certFile,
		KeyFile:  keyFile,
	})
	if err != nil {
		t.Fatalf("NewManager static: %v", err)
	}
	if !m.IsTLSEnabled() {
		t.Fatal("TLS should be enabled in static mode")
	}
}

func TestNewManagerStaticMissingFiles(t *testing.T) {
	_, err := NewManager(Config{
		Mode:     ModeStatic,
		CertFile: "/nonexistent/cert.pem",
		KeyFile:  "/nonexistent/key.pem",
	})
	if err == nil {
		t.Fatal("expected error for missing cert files")
	}
}

func TestNewManagerACMEFallback(t *testing.T) {
	// ACME without network access should fall back to self-signed
	tmpDir := t.TempDir()
	m, err := NewManager(Config{
		Mode:      ModeACME,
		Domain:    "acme.example.com",
		ACMEEmail: "test@example.com",
		CertDir:   tmpDir,
	})
	if err != nil {
		t.Fatalf("NewManager ACME fallback: %v", err)
	}
	if !m.IsTLSEnabled() {
		t.Fatal("TLS should be enabled after ACME fallback to self-signed")
	}
}

func TestNewManagerACMEMissingDomain(t *testing.T) {
	// ACME without domain falls back to self-signed
	tmpDir := t.TempDir()
	m, err := NewManager(Config{
		Mode:      ModeACME,
		ACMEEmail: "test@example.com",
		CertDir:   tmpDir,
	})
	if err != nil {
		t.Fatalf("expected fallback to self-signed, got error: %v", err)
	}
	if !m.IsTLSEnabled() {
		t.Fatal("TLS should be enabled after fallback")
	}
}

func TestNewManagerACMEMissingEmail(t *testing.T) {
	// ACME without email falls back to self-signed
	tmpDir := t.TempDir()
	m, err := NewManager(Config{
		Mode:    ModeACME,
		Domain:  "test.example.com",
		CertDir: tmpDir,
	})
	if err != nil {
		t.Fatalf("expected fallback to self-signed, got error: %v", err)
	}
	if !m.IsTLSEnabled() {
		t.Fatal("TLS should be enabled after fallback")
	}
}

func TestNewManagerInvalidMode(t *testing.T) {
	_, err := NewManager(Config{Mode: Mode("invalid")})
	if err == nil {
		t.Fatal("expected error for invalid mode")
	}
}
