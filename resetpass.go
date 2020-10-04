package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/server"
	"github.com/google/subcommands"
	"github.com/slonzok/getpass"

	"github.com/sirupsen/logrus"
)

type resetPwCmd struct {
	driver string
	dsn    string
	user   string
}

func (*resetPwCmd) Name() string     { return "resetpw" }
func (*resetPwCmd) Synopsis() string { return "Reset password." }
func (*resetPwCmd) Usage() string {
	return `resetpw [-option] <some text>:
  reset password.
`
}

func (p *resetPwCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&p.user, "u", "admin", "set user name")
	f.StringVar(&p.dsn, "dsn", "file:godnslog.db?cache=shared&mode=rwc", "set database source name, option")
	f.StringVar(&p.driver, "driver", "sqlite3", "set database driver, [sqlite3/mysql], option")
}

func (p *resetPwCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	newPass1 := getpass.Prompt("Please input new password")
	fmt.Println()
	newPass2 := getpass.Prompt("Please input new password again")
	fmt.Println()
	if newPass1 != newPass2 {
		fmt.Println("Passwords not same!")
		return subcommands.ExitSuccess
	}

	store := cache.NewCache(24*3600*time.Second, 10*time.Minute)

	web, err := server.NewWebServer(&server.WebServerConfig{
		Driver:                       p.driver,
		Dsn:                          p.dsn,
		Domain:                       "example.com",
		IP:                           "127.0.0.1",
		Listen:                       ":8080",
		Swagger:                      false,
		AuthExpire:                   AuthExpire,
		DefaultCleanInterval:         DefaultCleanInterval,
		DefaultQueryApiMaxItem:       DefaultQueryApiMaxItem,
		DefaultMaxCallbackErrorCount: DefaultMaxCallbackErrorCount,
		DefaultLanguage:              DefaultLanguage,
	}, store)
	if err != nil {
		logrus.Fatalf("[main.go::main] NewWebServer: %v", err)
	}
	err = web.ResetPassword(p.user, newPass2)
	if err != nil {
		fmt.Println("reset password: %v", err)
		return subcommands.ExitFailure
	}
	fmt.Println("Sucess!")

	//TODO:
	fmt.Println()
	return subcommands.ExitSuccess
}
