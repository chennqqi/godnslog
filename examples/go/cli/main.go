package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/chennqqi/godnslog/client"
)

func main() {
	var (
		domain string
		secret string
		query  string
		class  string
		blur   bool
		ssl    bool
	)
	flag.StringVar(&domain, "domain", "", "input your test domain")
	flag.StringVar(&secret, "secret", "", "input your api secret")
	flag.StringVar(&query, "query", "", "input your query param")
	flag.StringVar(&class, "type", "dns", "set query type")
	flag.BoolVar(&ssl, "ssl", false, "query server by https")
	flag.BoolVar(&blur, "blur", false, "query param not exactly")

	flag.Parse()

	if domain == "" || secret == "" || query == "" {
		log.Println("Both domain, secret, query are required")
		return
	}

	c, err := client.NewClient(domain, secret, ssl)
	if err != nil {
		log.Println("NewClient: %v", err)
		return
	}
	switch class {
	case "dns":
		r, err := c.QueryDns(query, blur)
		if err != nil {
			log.Fatal("query DNS", err)
		}
		fmt.Printf("Result: %#v\n", r)
	case "http":
		r, err := c.QueryHttp(query, blur)
		if err != nil {
			log.Fatal("query HTTP", err)
		}
		fmt.Printf("Result: %#v\n", r)
	default:
		log.Println("unknow type:", class)
	}
}
