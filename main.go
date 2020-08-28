package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/server"
	"github.com/sirupsen/logrus"
)

const (
	AuthExpire                   = 24 * 3600 * time.Second
	DefaultCleanInterval         = 7200 //seconds
	DefaultLanguage              = "en-US"
	DefaultQueryApiMaxItem       = 20
	DefaultMaxCallbackErrorCount = 5
)

func main() {
	var (
		swagger bool
		domain,
		ipv4, ipv6,
		logFile, logLevel,
		dsn, driver,
		defaultLanguage string
		httpListen string
	)

	flag.StringVar(&domain, "domain", "example.com", "set domain, required")
	flag.StringVar(&ipv4, "4", "", "set public IPv4, required")
	//flag.StringVar(&ipv6, "6", "", "set ipv6 publicIP, option")	// not support IPv6 now

	//https://github.com/mattn/go-sqlite3/issues/39
	flag.StringVar(&dsn, "dsn", "file:godnslog.db?cache=shared&mode=rwc", "set database source name, option")
	flag.StringVar(&driver, "driver", "sqlite3", "set database driver, [sqlite3/mysql], option")

	flag.BoolVar(&swagger, "swagger", false, "with swagger, option")

	flag.StringVar(&logFile, "log", "", "set log file, option")
	flag.StringVar(&logLevel, "level", "WARN", "set loglevel, option")
	flag.StringVar(&defaultLanguage, "lang", DefaultLanguage, "set default language, [en-US/zh-CN], option")

	flag.StringVar(&httpListen, "http", ":8080", "set http listen, option")

	flag.Parse()

	// log & log level
	{
		if logFile != "" {
			f, err := os.Create(logFile)
			if err != nil {
				log.Panicf("Open", logFile, err)
			}
			defer f.Close()
			buf := bufio.NewWriter(f)
			logrus.SetOutput(buf)
			defer buf.Flush()
		}
		switch strings.ToUpper(logLevel) {
		case "DEBUG":
			logrus.SetLevel(logrus.DebugLevel)
		case "WARN":
			logrus.SetLevel(logrus.WarnLevel)
		case "INFO":
			logrus.SetLevel(logrus.InfoLevel)
		default:
			logrus.SetLevel(logrus.WarnLevel)
		}
	}

	// verify input
	{
		if ipv4 == "" || domain == "" {
			logrus.Fatal("[main.go::main] You should set ipv4 and domain at least.")
			return
		}
		if swagger {
			logrus.Warnf("[main.go::main] We only suggest set this option in debug enviroment.")
			return
		}
	}

	var wg sync.WaitGroup

	//	cache store
	store := cache.NewCache(24*3600*time.Second, 10*time.Minute)

	web, err := server.NewWebServer(&server.WebServerConfig{
		Driver:                       driver,
		Dsn:                          dsn,
		Domain:                       domain,
		Listen:                       httpListen,
		Swagger:                      swagger,
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
		Domain:   domain,
		RTimeout: 3 * time.Second,
		WTimeout: 3 * time.Second,
		V4:       net.ParseIP(ipv4),
		V6:       net.ParseIP(ipv6),

		// custom resolve
		Fixed: []server.Resolve{
			server.Resolve{"www", "A", ipv4, 600},
			server.Resolve{"api", "A", ipv4, 600},
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
}
