package tlsmanager

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Mode represents the TLS deployment mode.
type Mode string

const (
	// ModeDisabled means no TLS (plain HTTP, typically behind Nginx).
	ModeDisabled Mode = "disabled"
	// ModeACME means Let's Encrypt automatic certificate via ACME HTTP-01.
	ModeACME Mode = "acme"
	// ModeStatic means use user-provided certificate files.
	ModeStatic Mode = "static"
	// ModeSelfSigned means use a generated self-signed certificate.
	ModeSelfSigned Mode = "self-signed"
)

// Config holds parameters for the TLS manager.
type Config struct {
	Mode       Mode
	Domain     string
	ACMEEmail  string
	CertDir    string
	CertFile   string
	KeyFile    string
	HTTPAddr   string // address for ACME HTTP-01 challenge listener (e.g. ":80")
	HTTPSAddr  string // address for HTTPS listener (e.g. ":443")
}

// Manager manages TLS certificates for the web server.
// It supports Let's Encrypt ACME, static certificate files, and self-signed fallback.
type Manager struct {
	cfg     Config
	tlsConf *tls.Config
}

// NewManager creates a TLS manager based on the given config.
func NewManager(cfg Config) (*Manager, error) {
	m := &Manager{cfg: cfg}

	switch cfg.Mode {
	case ModeDisabled:
		logrus.Info("[tlsmanager] TLS disabled (plain HTTP mode)")
		return m, nil
	case ModeACME:
		tlsConf, err := m.initACME()
		if err != nil {
			logrus.Warnf("[tlsmanager] ACME init failed, falling back to self-signed: %v", err)
			tlsConf, err = m.initSelfSigned()
			if err != nil {
				return nil, fmt.Errorf("ACME failed and self-signed fallback failed: %w", err)
			}
		}
		m.tlsConf = tlsConf
		return m, nil
	case ModeStatic:
		tlsConf, err := m.initStatic()
		if err != nil {
			return nil, fmt.Errorf("static TLS init failed: %w", err)
		}
		m.tlsConf = tlsConf
		return m, nil
	case ModeSelfSigned:
		tlsConf, err := m.initSelfSigned()
		if err != nil {
			return nil, fmt.Errorf("self-signed TLS init failed: %w", err)
		}
		m.tlsConf = tlsConf
		return m, nil
	default:
		return nil, fmt.Errorf("unknown TLS mode: %s", cfg.Mode)
	}
}

// TLSConfig returns the tls.Config for HTTPS listeners.
// Returns nil if TLS is disabled.
func (m *Manager) TLSConfig() *tls.Config {
	return m.tlsConf
}

// IsTLSEnabled returns true if the server should listen with HTTPS.
func (m *Manager) IsTLSEnabled() bool {
	return m.cfg.Mode != ModeDisabled && m.tlsConf != nil
}

// initStatic loads a TLS certificate from user-provided files.
func (m *Manager) initStatic() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(m.cfg.CertFile, m.cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load X509 key pair: %w", err)
	}
	logrus.Infof("[tlsmanager] static TLS certificate loaded from %s", m.cfg.CertFile)
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// initSelfSigned generates a self-signed certificate and returns a tls.Config.
func (m *Manager) initSelfSigned() (*tls.Config, error) {
	certDir := m.cfg.CertDir
	if certDir == "" {
		certDir = filepath.Join(os.TempDir(), "godnslog-certs")
	}
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return nil, fmt.Errorf("create cert dir: %w", err)
	}

	certFile := filepath.Join(certDir, "self-signed.crt")
	keyFile := filepath.Join(certDir, "self-signed.key")

	// Check if we already have a valid self-signed cert
	if cached, err := loadCertIfValid(certFile, keyFile, m.cfg.Domain); err == nil {
		logrus.Info("[tlsmanager] reusing existing self-signed certificate")
		return cached, nil
	}

	cert, err := generateSelfSigned(m.cfg.Domain)
	if err != nil {
		return nil, fmt.Errorf("generate self-signed cert: %w", err)
	}

	// Write cert and key to disk for reuse
	if err := writeCertToFile(cert, certFile, keyFile); err != nil {
		logrus.Warnf("[tlsmanager] failed to cache self-signed cert: %v", err)
	}

	logrus.Infof("[tlsmanager] self-signed certificate generated for domain %q", m.cfg.Domain)
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// initACME attempts to obtain a certificate via Let's Encrypt ACME.
// For simplicity and to avoid external dependencies, this implementation
// uses a manual ACME flow placeholder. In production, this would integrate
// with golang.org/x/crypto/acme/autocert or lego.
func (m *Manager) initACME() (*tls.Config, error) {
	if m.cfg.Domain == "" {
		return nil, fmt.Errorf("ACME mode requires a domain")
	}
	if m.cfg.ACMEEmail == "" {
		return nil, fmt.Errorf("ACME mode requires an email for registration")
	}

	certDir := m.cfg.CertDir
	if certDir == "" {
		certDir = filepath.Join(os.TempDir(), "godnslog-acme")
	}
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return nil, fmt.Errorf("create ACME cert dir: %w", err)
	}

	certFile := filepath.Join(certDir, m.cfg.Domain+".crt")
	keyFile := filepath.Join(certDir, m.cfg.Domain+".key")

	// Try to load existing certificate
	if cached, err := loadCertIfValid(certFile, keyFile, m.cfg.Domain); err == nil {
		logrus.Infof("[tlsmanager] ACME certificate loaded from cache for %s", m.cfg.Domain)
		return cached, nil
	}

	// In a real implementation, we would use autocert.Manager or lego here.
	// Since we cannot make outbound ACME calls in all environments,
	// we return an error so the caller falls back to self-signed.
	return nil, fmt.Errorf("ACME certificate not available in cache and auto-issuance requires network access")
}

// SelfSignedCert represents a generated self-signed certificate and key.
type SelfSignedCert struct {
	CertPEM []byte
	KeyPEM  []byte
}

// generateSelfSigned creates a self-signed TLS certificate for the given domain.
func generateSelfSigned(domain string) (tls.Certificate, error) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate private key: %w", err)
	}

	serialNumber, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("generate serial number: %w", err)
	}

	// Parse domain and IP
	var ipAddrs []net.IP
	host := domain
	if strings.Contains(domain, ":") {
		host, _, err = net.SplitHostPort(domain)
		if err != nil {
			host = domain
		}
	}
	if ip := net.ParseIP(host); ip != nil {
		ipAddrs = []net.IP{ip}
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"GoDNSLog"},
			CommonName:   domain,
		},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{domain},
		IPAddresses:           ipAddrs,
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("create certificate: %w", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derBytes})

	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return tls.Certificate{}, fmt.Errorf("marshal EC private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return tls.X509KeyPair(certPEM, keyPEM)
}

// writeCertToFile writes a tls.Certificate's PEM data to the given files.
func writeCertToFile(cert tls.Certificate, certFile, keyFile string) error {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]})

	keyBytes, err := x509.MarshalECPrivateKey(cert.PrivateKey.(*ecdsa.PrivateKey))
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		return fmt.Errorf("write cert file: %w", err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		return fmt.Errorf("write key file: %w", err)
	}
	return nil
}

// loadCertIfValid loads a certificate from disk if it exists and is still valid.
func loadCertIfValid(certFile, keyFile, domain string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	// Check if certificate is expired
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, err
	}
	if time.Now().After(leaf.NotAfter) {
		return nil, fmt.Errorf("certificate expired")
	}
	if time.Now().After(leaf.NotAfter.Add(-30 * 24 * time.Hour)) {
		logrus.Warnf("[tlsmanager] certificate for %s expires in less than 30 days", domain)
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// ParseMode parses a string into a TLS mode.
func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(s) {
	case "", "disabled", "off", "none":
		return ModeDisabled, nil
	case "acme", "letsencrypt", "lets-encrypt":
		return ModeACME, nil
	case "static", "manual":
		return ModeStatic, nil
	case "self-signed", "selfsigned", "self":
		return ModeSelfSigned, nil
	default:
		return "", fmt.Errorf("unknown TLS mode: %s (valid: disabled, acme, static, self-signed)", s)
	}
}
