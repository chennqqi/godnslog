package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/server"
	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
)

type servePwCmd struct {
	swagger   bool
	withGuest bool
	testMode  bool

	domain,
	driver, dsn,
	ipv4, ipv6,
	defaultLanguage string
	httpListen string
	upstream   string
	redisAddr  string

	geoipMMDBPath   string
	geoipLicenseKey string

	captchaEnabled bool
	captchaExpire  time.Duration

	tlsMode     string
	tlsCertFile string
	tlsKeyFile  string
	acmeEmail   string
	certDir     string

	demoMode       bool
	demoResetHours int
}

func (*servePwCmd) Name() string     { return "serve" }
func (*servePwCmd) Synopsis() string { return "Serve dnslog." }
func (*servePwCmd) Usage() string {
	return `serve [-options] <some text>:
  Print args to stdout.
`
}

func (p *servePwCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.domain, "domain", "example.com", "set domain, required")
	f.StringVar(&p.ipv4, "4", "", "set public IPv4, required")
	f.StringVar(&p.ipv6, "6", "", "set ipv6 publicIP, option")

	//https://github.com/mattn/go-sqlite3/issues/39
	f.StringVar(&p.dsn, "dsn", "file:godnslog.db?cache=shared&mode=rwc", "set database source name, option")
	f.StringVar(&p.driver, "driver", "sqlite", "set database driver, [sqlite/mysql], option")

	f.StringVar(&p.upstream, "upstreamp", "8.8.8.8:53", "set upstream dns")
	f.BoolVar(&p.swagger, "swagger", false, "with swagger, option")
	f.BoolVar(&p.withGuest, "guest", false, "init with guest user")
	f.BoolVar(&p.testMode, "test", false, "enable test mode with fixed password")

	f.StringVar(&p.defaultLanguage, "lang", DefaultLanguage, "set default language, [en-US/zh-CN], option")
	f.StringVar(&p.httpListen, "http", ":8080", "set http listen, option")
	f.StringVar(&p.redisAddr, "redis", "", "set Redis address for HA session sharing, option")
	f.StringVar(&p.geoipMMDBPath, "geoip-mmdb", os.Getenv("MMDB_PATH"), "path to GeoLite2-ASN.mmdb for source ASN enrichment (optional); auto-downloads if -geoip-license-key is set")
	f.StringVar(&p.geoipLicenseKey, "geoip-license-key", os.Getenv("MMDB_LICENSE_KEY"), "MaxMind license key for auto-downloading GeoLite2-ASN.mmdb; sign up free at https://www.maxmind.com/en/geolite2/signup")
	f.BoolVar(&p.captchaEnabled, "captcha-enabled", true, "enable captcha verification on login, option")
	f.DurationVar(&p.captchaExpire, "captcha-expire", DefaultCaptchaExpire, "captcha challenge TTL, option")

	// TLS / deployment mode flags
	f.StringVar(&p.tlsMode, "tls-mode", os.Getenv("GODNSLOG_TLS_MODE"), "TLS mode: disabled, acme, static, self-signed (env: GODNSLOG_TLS_MODE)")
	f.StringVar(&p.tlsCertFile, "tls-cert", os.Getenv("GODNSLOG_TLS_CERT"), "path to TLS certificate file (for static mode)")
	f.StringVar(&p.tlsKeyFile, "tls-key", os.Getenv("GODNSLOG_TLS_KEY"), "path to TLS private key file (for static mode)")
	f.StringVar(&p.acmeEmail, "acme-email", os.Getenv("GODNSLOG_ACME_EMAIL"), "email for Let's Encrypt registration (env: GODNSLOG_ACME_EMAIL)")
	f.StringVar(&p.certDir, "cert-dir", os.Getenv("GODNSLOG_CERT_DIR"), "directory for storing certificates (env: GODNSLOG_CERT_DIR)")

	// Demo mode flags
	f.BoolVar(&p.demoMode, "demo", getEnvBool("GODNSLOG_DEMO"), "enable demo mode with sample data (env: GODNSLOG_DEMO)")
	f.IntVar(&p.demoResetHours, "demo-reset-hours", 6, "interval in hours for demo data reset")
}

func (p *servePwCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	// verify input
	{
		if p.ipv4 == "" || p.domain == "" {
			logrus.Fatal("[main.go::main] You should set ipv4 and domain at least.")
			return subcommands.ExitUsageError
		}
		if p.swagger {
			logrus.Warnf("[main.go::main] We only suggest set this option in debug enviroment.")
			return subcommands.ExitUsageError
		}
	}

	var wg sync.WaitGroup

	//	cache store
	store := cache.NewCache(24*3600*time.Second, 10*time.Minute)

	web, err := server.NewWebServer(&server.WebServerConfig{
		Driver:                       p.driver,
		Dsn:                          p.dsn,
		Domain:                       p.domain,
		IP:                           p.ipv4,
		Listen:                       p.httpListen,
		Swagger:                      p.swagger,
		WithGuest:                    p.withGuest,
		TestMode:                     p.testMode,
		AuthExpire:                   AuthExpire,
		DefaultCleanInterval:         DefaultCleanInterval,
		DefaultQueryApiMaxItem:       DefaultQueryApiMaxItem,
		DefaultMaxCallbackErrorCount: DefaultMaxCallbackErrorCount,
		DefaultLanguage:              DefaultLanguage,
		RedisAddr:                    p.redisAddr,
		GeoIPMMDBPath:                p.geoipMMDBPath,
		GeoIPLicenseKey:              p.geoipLicenseKey,
		CaptchaEnabled:               p.captchaEnabled,
		CaptchaExpire:                p.captchaExpire,
		TLSMode:                      p.tlsMode,
		TLSCertFile:                  p.tlsCertFile,
		TLSKeyFile:                   p.tlsKeyFile,
		ACMEEmail:                    p.acmeEmail,
		CertDir:                      p.certDir,
		DemoMode:                     p.demoMode,
		DemoResetHours:               p.demoResetHours,
	}, store)
	if err != nil {
		logrus.Fatalf("[main.go::main] NewWebServer: %v", err)
	}

	//run async store routine
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			web.RunStoreRoutine()
		}()
	}

	//run web server routine
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			web.Run()
		}()
	}

	dns, err := server.NewDnsServer(&server.DnsServerConfig{
		Domain:   p.domain,
		RTimeout: 3 * time.Second,
		WTimeout: 3 * time.Second,
		V4:       net.ParseIP(p.ipv4),
		V6:       net.ParseIP(p.ipv6),
		Upstream: p.upstream,
	}, store)
	if err != nil {
		logrus.Fatalf("[main.go::main] NewWebServer: %v", err)
	}

	//run dns server
	{
		wg.Add(1)
		go func() {
			defer wg.Done()
			dns.Run()
		}()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, os.Interrupt)
	<-sigCh

	dns.Shutdown()
	store.Close()
	web.Shutdown(context.Background())

	wg.Wait()

	fmt.Println()
	return subcommands.ExitSuccess
}

// getEnvBool reads a boolean from an environment variable, returning false on
// missing or unparseable values.
func getEnvBool(key string) bool {
	v := os.Getenv(key)
	if v == "" {
		return false
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return false
	}
	return b
}
