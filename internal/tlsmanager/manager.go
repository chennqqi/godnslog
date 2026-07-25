package tlsmanager

import (
	"context"
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
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme/autocert"
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
	Mode      Mode
	Domain    string
	ACMEEmail string
	CertDir   string
	CertFile  string
	KeyFile   string
	HTTPAddr  string // address for ACME HTTP-01 challenge listener (e.g. ":80")
	HTTPSAddr string // address for HTTPS listener (e.g. ":443")
}

// Manager manages TLS certificates for the web server.
// It supports Let's Encrypt ACME, static certificate files, and self-signed fallback.
type Manager struct {
	cfg          Config
	tlsConf      *tls.Config
	acmeManager  *autocert.Manager
	challengeSrv *http.Server
}

// NewManager creates a TLS manager based on the given config.
func NewManager(cfg Config) (*Manager, error) {
	m := &Manager{cfg: cfg}

	switch cfg.Mode {
	case ModeDisabled:
		logrus.Warn("[tlsmanager] TLS disabled (plain HTTP mode)")
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

// ACMEChallengeHandler returns the HTTP handler for ACME HTTP-01 challenges.
func (m *Manager) ACMEChallengeHandler() http.Handler {
	if m.acmeManager != nil {
		return m.acmeManager.HTTPHandler(nil)
	}
	return nil
}

// Shutdown stops the ACME challenge server if running.
func (m *Manager) Shutdown(ctx context.Context) error {
	if m.challengeSrv != nil {
		return m.challengeSrv.Shutdown(ctx)
	}
	return nil
}

// initStatic loads a TLS certificate from user-provided files.
func (m *Manager) initStatic() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(m.cfg.CertFile, m.cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load X509 key pair: %w", err)
	}
	logrus.Warnf("[tlsmanager] static TLS certificate loaded from %s", m.cfg.CertFile)
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
		logrus.Warn("[tlsmanager] reusing existing self-signed certificate")
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

	logrus.Warnf("[tlsmanager] self-signed certificate generated for domain %q", m.cfg.Domain)
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// initACME sets up automatic certificate management via Let's Encrypt ACME.
// Uses autocert.Manager for the full certificate lifecycle (issuance, caching, renewal).
// No fallback certificate — autocert returns a temporary cert during the ACME
// handshake and swaps in the real cert once Let's Encrypt validates domain ownership.
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

	// autocert handles HTTP-01 for apex + www domains (SAN cert).
	// DNS-01 wildcard is handled by a background goroutine after startup
	// using our own DNS server to serve _acme-challenge TXT records.
	m.acmeManager = &autocert.Manager{
		Cache:      autocert.DirCache(certDir),
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(m.cfg.Domain, "www."+m.cfg.Domain),
		Email:      m.cfg.ACMEEmail,
	}

	// Start background HTTP server for ACME HTTP-01 challenges (port 80)
	httpAddr := m.cfg.HTTPAddr
	if httpAddr == "" {
		httpAddr = ":80"
	}
	m.challengeSrv = &http.Server{
		Addr:    httpAddr,
		Handler: m.acmeManager.HTTPHandler(nil),
	}
	go func() {
		logrus.Warnf("[tlsmanager] ACME HTTP-01 challenge server on %s", httpAddr)
		if err := m.challengeSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Warnf("[tlsmanager] ACME challenge server: %v", err)
		}
	}()

	// Pre-obtain certificates in background by making real TLS connections
	// to ourselves. This triggers autocert's ACME flow correctly.
	preObtain := func(serverName string, after time.Duration) {
		go func() {
			time.Sleep(after)
			logrus.Warnf("[tlsmanager] pre-obtaining cert for %s...", serverName)
			conf := &tls.Config{
				ServerName:         serverName,
				InsecureSkipVerify: true,
			}
			addr := m.cfg.HTTPSAddr
			if addr == "" {
				addr = "127.0.0.1:443"
			}
			conn, err := tls.Dial("tcp", addr, conf)
			if err != nil {
				logrus.Warnf("[tlsmanager] pre-obtain %s: %v", serverName, err)
				return
			}
			conn.Close()
			logrus.Warnf("[tlsmanager] pre-obtain %s complete", serverName)
		}()
	}
	preObtain(m.cfg.Domain, 5*time.Second)
	preObtain("www."+m.cfg.Domain, 7*time.Second)

	// Self-signed fallback while ACME certificates are being obtained.
	// Once autocert obtains the real certs, GetCertificate returns them.
	return &tls.Config{
		GetCertificate: func(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
			cert, err := m.acmeManager.GetCertificate(hello)
			if err == nil {
				return cert, nil
			}
			if fb, fbErr := generateSelfSigned(m.cfg.Domain); fbErr == nil {
				return &fb, nil
			}
			return nil, err
		},
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"h2", "http/1.1", "acme-tls/1"},
	}, nil
}

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
