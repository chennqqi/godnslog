package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const (
	GODNS_MD5 = "694ef536e5d0245f203a1bcf8cbf3294"
)

type Client struct {
	*http.Client

	host  string
	token string
}

func NewClient(host, token string) (*Client, error) {
	client := &Client{
		Client: &http.Client{},
	}
	return client, nil
}

func (self *Client) Hash(querys url.Values) string {
	var keys []string
	for k, _ := range querys {
		keys = append(keys, k)
	}
	h := md5.New()
	sort.Strings(keys)
	for _, k := range keys {
		h.Write([]byte(querys.Get(k)))
	}

	h.Write([]byte(self.token))
	return hex.EncodeToString(h.Sum(nil))
}

func (self *Client) QueryDns(domain string) error {
	c := self.Client

	querys := make(url.Values)
	querys.Set("t", fmt.Sprintf("%v", time.Now().Unix()))
	querys.Set("domain", domain)

	hash := self.Hash(querys)
	querys.Set("hash", hash)

	u := fmt.Sprintf("%v/app/dns?%v", querys.Encode())
	req, err := http.NewRequest("GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {

	}
	defer resp.Body.Close()
	return nil
}

func (self *Client) QueryHttp(suffix string) error {
	c := self.Client

	querys := make(url.Values)
	querys.Set("t", fmt.Sprintf("%v", time.Now().Unix()))

	hash := self.Hash(querys)
	querys.Set("hash", hash)

	u := fmt.Sprintf("%v/app/http?%v", querys.Encode())
	req, err := http.NewRequest("GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func main() {
	client, err := NewClient(
		`http://xxx.godnslog.com`,
		`21231231`)

	if err != nil {

	}
	client.QueryDns("xxxx")
	client.QueryHttp("xxx")
}
