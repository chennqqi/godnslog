package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/server"
	"github.com/google/subcommands"
	"github.com/sirupsen/logrus"
)

type servePwCmd struct {
	swagger bool
	domain,
	driver, dsn,
	ipv4, ipv6,
	defaultLanguage string
	httpListen string
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
	//flag.StringVar(&ipv6, "6", "", "set ipv6 publicIP, option")	// not support IPv6 now

	//https://github.com/mattn/go-sqlite3/issues/39
	f.StringVar(&p.dsn, "dsn", "file:godnslog.db?cache=shared&mode=rwc", "set database source name, option")
	f.StringVar(&p.driver, "driver", "sqlite3", "set database driver, [sqlite3/mysql], option")

	f.BoolVar(&p.swagger, "swagger", false, "with swagger, option")
	f.StringVar(&p.defaultLanguage, "lang", DefaultLanguage, "set default language, [en-US/zh-CN], option")
	f.StringVar(&p.httpListen, "http", ":8080", "set http listen, option")
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
		AuthExpire:                   AuthExpire,
		DefaultCleanInterval:         DefaultCleanInterval,
		DefaultQueryApiMaxItem:       DefaultQueryApiMaxItem,
		DefaultMaxCallbackErrorCount: DefaultMaxCallbackErrorCount,
		DefaultLanguage:              DefaultLanguage,
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

		// custom resolve
		Fixed: []server.Resolve{
			server.Resolve{"www", "A", p.ipv4, 600},
			server.Resolve{"api", "A", p.ipv4, 600},
		},
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

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Kill, os.Interrupt)
	<-sigCh

	dns.Shutdown()
	store.Close()
	web.Shutdown(context.Background())

	wg.Wait()

	fmt.Println()
	return subcommands.ExitSuccess
}
