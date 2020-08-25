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

	"github.com/sirupsen/logrus"
)

// @title GoDnsLog API
// @version 0.1
// @description This is a sample server Petstore server.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url https://github.com/chennqqi/godnslog/issue
// @contact.email chennqqi@qq.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host petstore.swagger.io
// @BasePath /v1

var (
	defaultLanguage = "en-US"
)

func main() {
	var ui, swagger bool
	var domain, ipv4, ipv6, logFile, logLevel, dsn, driver string
	flag.BoolVar(&ui, "ui", false, "with ui, option")
	flag.BoolVar(&swagger, "swagger", false, "with swagger, option")
	flag.StringVar(&domain, "domain", "example.com", "set domain, required")
	flag.StringVar(&logFile, "log", "", "set file, option")
	flag.StringVar(&logLevel, "level", "WARN", "set loglevel, option")
	flag.StringVar(&defaultLanguage, "lang", "en-US", "set default language, en-US/zh-CN")
	flag.StringVar(&ipv4, "4", "", "set ipv4 publicIP, required")
	//flag.StringVar(&ipv6, "6", "", "set ipv6 publicIP, option")

	//https://github.com/mattn/go-sqlite3/issues/39
	flag.StringVar(&dsn, "dsn", "file:godnslog.db?cache=shared&mode=rwc", "set database source name, option")
	flag.StringVar(&driver, "driver", "sqlite3", "set database driver, sqlite3/mysql")

	flag.Parse()

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

	var wg sync.WaitGroup
	cache := NewCache(24*3600*time.Second, 10*time.Minute)

	web, err := NewWebServer(&WebServerConfig{
		Driver: driver,
		Dsn:    dsn,
		Domain: domain,
		Listen: ":8080",
	}, cache)
	if err != nil {
		logrus.Fatalf("[main.go::main] NewWebServer: %v", err)
	}

	//run web server routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		web.Run()
	}()

	//run async store routine
	wg.Add(1)
	go func() {
		defer wg.Done()
		web.RunStoreRoutine()
	}()

	dns, err := NewDnsServer(&DnsServerConfig{
		Domain:   domain,
		RTimeout: 3 * time.Second,
		WTimeout: 3 * time.Second,
		V4:       net.ParseIP(ipv4),
		V6:       net.ParseIP(ipv6),
	}, cache)
	if err != nil {
		logrus.Fatalf("[main.go::main] NewWebServer: %v", err)
	}

	//run dns server
	wg.Add(1)
	go func() {
		defer wg.Done()
		dns.Run()
	}()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, os.Kill, os.Interrupt)
	<-sigCh

	dns.Shutdown()
	cache.Close()
	web.Shutdown(context.Background())

	wg.Wait()
}
