package main

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/subcommands"
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
		logFile, logLevel string
	)

	subcommands.Register(subcommands.FlagsCommand(), "")
	subcommands.Register(subcommands.HelpCommand(), "")
	subcommands.Register(&servePwCmd{}, "")
	subcommands.Register(&resetPwCmd{}, "")

	//https://github.com/mattn/go-sqlite3/issues/39
	flag.StringVar(&logFile, "log", "", "set log file, option")
	flag.StringVar(&logLevel, "level", "WARN", "set loglevel, option")
	flag.Parse()

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
	// log & log level
	{
		if logFile != "" {
			f, err := os.Create(logFile)
			if err != nil {
				log.Panicf("Open", logFile, err)
			}
			defer f.Close()
			buf := bufio.NewWriter(f)
			//async flush
			go func() {
				for {
					time.Sleep(60 * time.Second)
					buf.Flush()
				}
			}()
			logrus.SetOutput(buf)
			defer buf.Flush()
		}

	}

	ctx := context.Background()
	os.Exit(int(subcommands.Execute(ctx)))
}
