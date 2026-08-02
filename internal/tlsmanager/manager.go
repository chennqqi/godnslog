package tlsmanager

import (
	"context"
	"crypto"
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
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/acme"
)

type Mode string

const (
	ModeDisabled   Mode = "disabled"
	ModeACME       Mode = "acme"
	ModeStatic     Mode = "static"
	ModeSelfSigned Mode = "self-signed"
)

// DNS01RecordWriter is called by the TLS manager to set or remove
// _acme-challenge TXT records during DNS-01 certificate issuance.
type DNS01RecordWriter interface {
	SetDNS01Record(fqdn, value string, ttl uint32) error
	DeleteDNS01Record(fqdn string) error
}

type Config struct {
	Mode      Mode
	Domain    string
	ACMEEmail string
	CertDir   string
	CertFile  string
	KeyFile   string
	HTTPAddr  string
	HTTPSAddr string

	// DNS01, if set, enables DNS-01 challenge (wildcard support).
	// The writer must persist TXT records into the DNS server's database.
	DNS01 DNS01RecordWriter
}

type Manager struct {
	cfg     Config
	tlsConf *tls.Config
	certDir string

	wildcardMu  sync.Mutex
	wildcardCert *tls.Certificate
}

func NewManager(cfg Config) (*Manager, error) {
	m := &Manager{cfg: cfg}

	switch cfg.Mode {
	case ModeDisabled:
		logrus.Warn("[tlsmanager] TLS disabled")
		return m, nil
	case ModeACME:
		tlsConf, err := m.initACME()
		if err != nil {
			logrus.Warnf("[tlsmanager] ACME init failed: %v", err)
			tlsConf, err = m.initSelfSigned()
			if err != nil {
				return nil, fmt.Errorf("ACME+self-signed both failed: %w", err)
			}
		}
		m.tlsConf = tlsConf
		return m, nil
	case ModeStatic:
		tlsConf, err := m.initStatic()
		if err != nil {
			return nil, err
		}
		m.tlsConf = tlsConf
		return m, nil
	case ModeSelfSigned:
		tlsConf, err := m.initSelfSigned()
		if err != nil {
			return nil, err
		}
		m.tlsConf = tlsConf
		return m, nil
	default:
		return nil, fmt.Errorf("unknown TLS mode: %s", cfg.Mode)
	}
}

func (m *Manager) TLSConfig() *tls.Config { return m.tlsConf }
func (m *Manager) IsTLSEnabled() bool     { return m.cfg.Mode != ModeDisabled && m.tlsConf != nil }
func (m *Manager) Shutdown(_ context.Context) error { return nil }

func (m *Manager) initStatic() (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(m.cfg.CertFile, m.cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("load X509 key pair: %w", err)
	}
	logrus.Warnf("[tlsmanager] static TLS certificate loaded")
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
}

func (m *Manager) initSelfSigned() (*tls.Config, error) {
	certDir := m.cfg.CertDir
	if certDir == "" {
		certDir = filepath.Join(os.TempDir(), "godnslog-certs")
	}
	os.MkdirAll(certDir, 0700)

	certFile := filepath.Join(certDir, "self-signed.crt")
	keyFile := filepath.Join(certDir, "self-signed.key")
	if cached, err := loadCertIfValid(certFile, keyFile, m.cfg.Domain); err == nil {
		logrus.Warn("[tlsmanager] reusing self-signed certificate")
		return cached, nil
	}

	cert, err := generateSelfSigned(m.cfg.Domain)
	if err != nil {
		return nil, err
	}
	writeCertToFile(cert, certFile, keyFile)
	logrus.Warnf("[tlsmanager] self-signed certificate generated")
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
}

// initACME prepares a TLS config that serves the DNS-01 wildcard certificate
// for any hostname under the configured domain. If no wildcard cert is
// available yet (PreObtainCert not run or failed), it falls back to a
// self-signed certificate so the HTTPS listener can still start.
func (m *Manager) initACME() (*tls.Config, error) {
	if m.cfg.Domain == "" {
		return nil, fmt.Errorf("ACME mode requires a domain")
	}
	if m.cfg.ACMEEmail == "" {
		return nil, fmt.Errorf("ACME mode requires an email")
	}

	m.certDir = m.cfg.CertDir
	if m.certDir == "" {
		m.certDir = filepath.Join(os.TempDir(), "godnslog-acme")
	}
	os.MkdirAll(m.certDir, 0700)

	// Reuse an existing valid wildcard cert on disk so we don't re-issue on every restart.
	if _, ok := m.loadWildcardCert(); ok {
		logrus.Warn("[tlsmanager] reusing existing wildcard certificate")
	}

	return &tls.Config{
		GetCertificate: m.getCertificate,
		MinVersion:     tls.VersionTLS12,
		NextProtos:     []string{"h2", "http/1.1"},
	}, nil
}

// getCertificate returns the DNS-01 wildcard certificate for any requested
// hostname (the wildcard covers all subdomains of the configured domain).
// Falls back to a self-signed certificate when no wildcard cert is loaded.
func (m *Manager) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	m.wildcardMu.Lock()
	defer m.wildcardMu.Unlock()
	if m.wildcardCert != nil {
		return m.wildcardCert, nil
	}
	fb, _ := generateSelfSigned(m.cfg.Domain)
	return &fb, nil
}

// PreObtainCert obtains a wildcard certificate via DNS-01 challenge, using
// GODNSLOG's own DNS server to serve _acme-challenge TXT records.
// If a valid wildcard cert already exists on disk, it is reused instead.
// Must be called BEFORE the HTTPS server starts.
func (m *Manager) PreObtainCert() error {
	if m.cfg.DNS01 == nil || m.cfg.Domain == "" {
		return nil
	}
	if m.certDir == "" {
		m.certDir = m.cfg.CertDir
		if m.certDir == "" {
			m.certDir = filepath.Join(os.TempDir(), "godnslog-acme")
		}
		os.MkdirAll(m.certDir, 0700)
	}

	// Reuse existing cert if still valid for at least a month.
	if _, ok := m.loadWildcardCert(); ok {
		logrus.Warn("[tlsmanager] reusing existing wildcard certificate (no re-issue needed)")
		return nil
	}

	wildcard := "*." + m.cfg.Domain
	logrus.Warnf("[tlsmanager] DNS-01 obtaining wildcard cert for %s and %s...", wildcard, m.cfg.Domain)

	// Load or create ACME account key.
	accountKey, err := loadOrCreateAccountKey(m.certDir)
	if err != nil {
		return fmt.Errorf("account key: %w", err)
	}

	client := &acme.Client{
		Key:          accountKey,
		DirectoryURL: acme.LetsEncryptURL,
	}

	// Register or look up account.
	acct := &acme.Account{Contact: []string{"mailto:" + m.cfg.ACMEEmail}}
	if _, err := client.Register(context.Background(), acct, acme.AcceptTOS); err != nil {
		if err != acme.ErrAccountAlreadyExists {
			return fmt.Errorf("register: %w", err)
		}
	}

	// DNS-01: create order for *.domain + domain (SAN).
	order, err := client.AuthorizeOrder(context.Background(),
		acme.DomainIDs(wildcard, m.cfg.Domain))
	if err != nil {
		return fmt.Errorf("authorize order: %w", err)
	}

	// Fulfill DNS-01 challenges.
	for _, authzURL := range order.AuthzURLs {
		authz, err := client.GetAuthorization(context.Background(), authzURL)
		if err != nil {
			return fmt.Errorf("get authz: %w", err)
		}

		// Use the identifier (domain name) from this authorization.
		domain := authz.Identifier.Value

		var dnsChal *acme.Challenge
		for i := range authz.Challenges {
			if authz.Challenges[i].Type == "dns-01" {
				dnsChal = authz.Challenges[i]
				break
			}
		}
		if dnsChal == nil {
			return fmt.Errorf("no DNS-01 challenge for %s", domain)
		}

		// Compute the TXT record value.
		resp, err := client.DNS01ChallengeRecord(dnsChal.Token)
		if err != nil {
			return fmt.Errorf("dns-01 record: %w", err)
		}

		// Write _acme-challenge TXT record into GODNSLOG's DNS database.
		recordName := "_acme-challenge." + domain + "."
		if err := m.cfg.DNS01.SetDNS01Record(recordName, resp, 60); err != nil {
			return fmt.Errorf("set DNS record: %w", err)
		}

		// Accept the challenge.
		if _, err := client.Accept(context.Background(), dnsChal); err != nil {
			m.cfg.DNS01.DeleteDNS01Record(recordName)
			return fmt.Errorf("accept: %w", err)
		}

		// Wait for LE to validate via DNS query.
		if _, err := client.WaitAuthorization(context.Background(), authz.URI); err != nil {
			m.cfg.DNS01.DeleteDNS01Record(recordName)
			return fmt.Errorf("wait authz: %w", err)
		}

		// Clean up the TXT record.
		m.cfg.DNS01.DeleteDNS01Record(recordName)
		logrus.Warnf("[tlsmanager] DNS-01 validated for %s", domain)
	}

	// Generate a cert key and finalize the order.
	certKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate cert key: %w", err)
	}

	csr, err := x509.CreateCertificateRequest(rand.Reader,
		&x509.CertificateRequest{DNSNames: []string{wildcard, m.cfg.Domain}},
		certKey,
	)
	if err != nil {
		return fmt.Errorf("create CSR: %w", err)
	}

	derChain, _, err := client.CreateOrderCert(context.Background(), order.FinalizeURL, csr, true)
	if err != nil {
		return fmt.Errorf("finalize: %w", err)
	}

	// Save certificate to disk for persistence across restarts.
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: derChain[0]})
	keyDER, _ := x509.MarshalECPrivateKey(certKey)
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER})

	certFile := filepath.Join(m.certDir, m.cfg.Domain+".crt")
	keyFile := filepath.Join(m.certDir, m.cfg.Domain+".key")
	if err := os.WriteFile(certFile, certPEM, 0600); err != nil {
		logrus.Warnf("[tlsmanager] failed to persist cert %s: %v", certFile, err)
	}
	if err := os.WriteFile(keyFile, keyPEM, 0600); err != nil {
		logrus.Warnf("[tlsmanager] failed to persist key %s: %v", keyFile, err)
	}

	// Load the new cert into memory for serving.
	if _, ok := m.loadWildcardCert(); ok {
		logrus.Warnf("[tlsmanager] wildcard certificate obtained for %s (and %s)", wildcard, m.cfg.Domain)
		return nil
	}

	return fmt.Errorf("obtained cert but failed to load %s/%s", certFile, keyFile)
}

// loadWildcardCert loads <domain>.crt/.key from the cert dir if the cert is
// still valid for more than a month. Stores it in wildcardCert and returns it.
func (m *Manager) loadWildcardCert() (*tls.Certificate, bool) {
	if m.certDir == "" {
		return nil, false
	}
	certFile := filepath.Join(m.certDir, m.cfg.Domain+".crt")
	keyFile := filepath.Join(m.certDir, m.cfg.Domain+".key")
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, false
	}
	leaf, err := x509.ParseCertificate(cert.Certificate[0])
	if err != nil {
		return nil, false
	}
	// Renew if the cert expires within a month.
	if time.Until(leaf.NotAfter) < 30*24*time.Hour {
		logrus.Warnf("[tlsmanager] wildcard cert expires %s, renewing", leaf.NotAfter.Format(time.RFC3339))
		return nil, false
	}

	m.wildcardMu.Lock()
	m.wildcardCert = &cert
	m.wildcardMu.Unlock()
	return &cert, true
}

func loadOrCreateAccountKey(certDir string) (crypto.Signer, error) {
	accountFile := filepath.Join(certDir, "acme_account+key")
	data, err := os.ReadFile(accountFile)
	if err == nil {
		block, _ := pem.Decode(data)
		if block != nil {
			return x509.ParseECPrivateKey(block.Bytes)
		}
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	der, _ := x509.MarshalECPrivateKey(key)
	os.WriteFile(accountFile, pem.EncodeToMemory(&pem.Block{
		Type: "EC PRIVATE KEY", Bytes: der,
	}), 0600)
	return key, nil
}

func generateSelfSigned(domain string) (tls.Certificate, error) {
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	serial, _ := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	var ipAddrs []net.IP
	if ip := net.ParseIP(domain); ip != nil {
		ipAddrs = []net.IP{ip}
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{Organization: []string{"GoDNSLog"}, CommonName: domain},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		DNSNames:     []string{domain},
		IPAddresses:  ipAddrs,
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &priv.PublicKey, priv)
	kp, _ := x509.MarshalECPrivateKey(priv)
	return tls.X509KeyPair(
		pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}),
		pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kp}),
	)
}

func writeCertToFile(cert tls.Certificate, certFile, keyFile string) error {
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Certificate[0]})
	kp, _ := x509.MarshalECPrivateKey(cert.PrivateKey.(*ecdsa.PrivateKey))
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: kp})
	os.WriteFile(certFile, certPEM, 0600)
	os.WriteFile(keyFile, keyPEM, 0600)
	return nil
}

func loadCertIfValid(certFile, keyFile, domain string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	leaf, _ := x509.ParseCertificate(cert.Certificate[0])
	if time.Now().After(leaf.NotAfter) {
		return nil, fmt.Errorf("certificate expired")
	}
	return &tls.Config{Certificates: []tls.Certificate{cert}, MinVersion: tls.VersionTLS12}, nil
}

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
