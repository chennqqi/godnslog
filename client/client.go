package client

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"sort"
	"time"

	"github.com/chennqqi/godnslog/models"
)

type Client struct {
	*http.Client

	host   string
	token  string
	domain string
}

func NewClient(host, token string) (*Client, error) {
	client := &Client{
		Client: &http.Client{},
	}
	return client, nil
}

func (self *Client) ToDnsDomain(v interface{}) string {
	return fmt.Sprintf("%v.%v", v, self.host)
}

func (self *Client) ToHttpURL(v interface{}) string {
	return fmt.Sprintf("%v/log/%v", self.host, v)
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

func (self *Client) QueryDns(domain string) ([]models.DnsRecord, error) {
	c := self.Client

	querys := make(url.Values)
	querys.Set("t", fmt.Sprintf("%v", time.Now().Unix()))
	querys.Set("domain", domain)

	hash := self.Hash(querys)
	querys.Set("hash", hash)

	u := fmt.Sprintf("%v/app/dns?%v", self.host, querys.Encode())
	req, err := http.NewRequest("GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rcds []models.DnsRecord
	var cr models.CR
	cr.Result = &rcds

	txt, _ := ioutil.ReadAll(resp.Body)
	return rcds, json.Unmarshal(txt, &cr)
}

func (self *Client) QueryHttp(suffix string) ([]models.HttpRecord, error) {
	c := self.Client

	querys := make(url.Values)
	querys.Set("t", fmt.Sprintf("%v", time.Now().Unix()))

	hash := self.Hash(querys)
	querys.Set("hash", hash)

	u := fmt.Sprintf("%v/app/http?%v", self.host, querys.Encode())
	req, err := http.NewRequest("GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rcds []models.HttpRecord
	var cr models.CR
	cr.Result = &rcds

	txt, _ := ioutil.ReadAll(resp.Body)
	return rcds, json.Unmarshal(txt, &cr)
}
