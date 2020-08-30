package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/chennqqi/godnslog/client"
	"github.com/chennqqi/godnslog/models"
)

func main() {
	var (
		host   string
		domain string
		secret string
		ssl    bool
	)
	flag.StringVar(&domain, "domain", "", "input your test domain")
	flag.StringVar(&secret, "secret", "", "input your api secret")
	flag.BoolVar(&ssl, "ssl", false, "query server by https")

	flag.Parse()

	if domain == "" || secret == "" {
		log.Println("Both domain and secret are required")
		return
	}

	c, err := client.NewClient(host, secret, ssl)
	if err != nil {
		log.Println("NewClient: %v", err)
		return
	}

	//callback example
	{
		http.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
			defer w.WriteHeader(200)
			defer r.Body.Close()

			txt, err := ioutil.ReadAll(r.Body)
			if err != nil {
				log.Println("handleCallback ReadAll:", err)
				return
			}

			t := r.URL.Query().Get("type")
			switch t {
			case "dns":
				var rcd models.DnsRecord
				err := json.NewDecoder(r.Body).Decode(&rcd)
				if err != nil {
					log.Println("handleCallback Decode:", err)
					return
				}
				fmt.Printf("callback DNS: %#v\n", rcd)

			case "http":
				var rcd models.HttpRecord
				err := json.NewDecoder(r.Body).Decode(&rcd)
				if err != nil {
					log.Println("handleCallback Decode:", err)
					return
				}
				fmt.Printf("callback HTTP: %#v\n", rcd)
			}
		})
		go http.ListenAndServe(":8081", nil)
	}

	var key int64
	simulatorScanId := func() int64 {
		//simulate scanId
		return atomic.AddInt64(&key, 1)
	}

	//dns example
	//dig xxx.n128b25b768knzdz.godnslog.com
	{
		var store = make(map[string]int64)
		var prefix string
		for i := 0; i < 5; i++ {
			scanId := simulatorScanId()
			target := c.BuildDnsDomain(scanId)

			//simulate dns request
			net.LookupIP(target)
			store[target] = scanId
		}
		// simulate wait..
		time.Sleep(10 * time.Second)

		// exactly query
		for k, v := range store {
			rcds, err := c.QueryDns(k, false)
			if err != nil {
				log.Println("Query DNS:", err)
				continue
			}
			if len(rcds) > 0 {
				//match vulnerable
			}
		}
		// batch query
		{
			rcds, err := c.QueryDns(k, true)
			if err != nil {
				log.Println("Query DNS:", err)
			}
			if len(rcds) > 0 {
				//match vulnerable
			}
		}
	}

	//http example
	{
		var prefix string
		var store = make(map[string]int64)
		for i := 0; i < 5; i++ {
			scanId := simulatorScanId()
			target := fmt.Sprintf("%v.%v", name, domain)

			//simulate dns request
			net.LookupIP(target)
			store[target] = scanId
		}
		// simulate wait..
		time.Sleep(10 * time.Second)

		for k, v := range store {
			rcds, err := c.QueryDns(k, false)
			if err != nil {
				log.Println("Query DNS:", err)
				continue
			}
			if len(rcds) > 0 {
			}
		}
		// batch query
		{
			rcds, err := c.QueryDns(k)
			if err != nil {
				log.Println("Query DNS:", err)
				return
			}
			if len(rcds) > 0 {
				//match vulnerable
			}
		}
	}

}
