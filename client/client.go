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
	"strings"
	"time"

	"github.com/chennqqi/godnslog/models"
)

type Client struct {
	*http.Client

	shortId string
	host    string
	secret  string
	domain  string
}

func NewClient(domain, secret string, ssl bool) (*Client, error) {
	host := "http://" + domain
	if ssl {
		host = "https://" + domain
	}

	idx := strings.Index(domain, ".")
	if idx <= 0 {
		return nil, fmt.Errorf("Unexpect domain format")
	}

	client := &Client{
		Client:  &http.Client{},
		host:    host,
		domain:  domain,
		shortId: domain[:idx],
		secret:  secret,
	}
	return client, nil
}

func (self *Client) BuildDnsDomain(v interface{}) string {
	return fmt.Sprintf("%v.%v", v, self.domain)
}

func (self *Client) BuildHttpURL(v interface{}) string {
	return fmt.Sprintf("%v/log/%v/%v", self.host, self.shortId, v)
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

	h.Write([]byte(self.secret))
	return hex.EncodeToString(h.Sum(nil))
}

func (self *Client) QueryDns(varirable string, blur bool) ([]models.DnsRecord, error) {
	c := self.Client

	querys := make(url.Values)
	querys.Set("t", fmt.Sprintf("%v", time.Now().Unix()))
	querys.Set("q", varirable)
	if blur {
		querys.Set("blur", "1")
	} else {
		querys.Set("blur", "0")
	}

	hash := self.Hash(querys)
	querys.Set("hash", hash)

	u := fmt.Sprintf("%v/data/dns?%v", self.host, querys.Encode())
	req, err := http.NewRequest("GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var cr models.CR
		txt, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(txt, &cr)
		return nil, fmt.Errorf("code(%v), Reason:%v", resp.StatusCode, cr.Message)
	}

	var rcds []models.DnsRecord
	var cr models.CR
	cr.Result = &rcds

	txt, _ := ioutil.ReadAll(resp.Body)
	return rcds, json.Unmarshal(txt, &cr)
}

func (self *Client) QueryHttp(varirable string, blur bool) ([]models.HttpRecord, error) {
	c := self.Client

	querys := make(url.Values)
	querys.Set("t", fmt.Sprintf("%v", time.Now().Unix()))
	querys.Set("q", varirable)
	if blur {
		querys.Set("blur", "1")
	} else {
		querys.Set("blur", "0")
	}

	hash := self.Hash(querys)
	querys.Set("hash", hash)

	u := fmt.Sprintf("%v/data/http?%v", self.host, querys.Encode())
	req, err := http.NewRequest("GET", u, nil)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		var cr models.CR
		txt, _ := ioutil.ReadAll(resp.Body)
		json.Unmarshal(txt, &cr)
		return nil, fmt.Errorf("code(%v), Reason:%v", resp.StatusCode, cr.Message)
	}

	var rcds []models.HttpRecord
	var cr models.CR
	cr.Result = &rcds

	txt, _ := ioutil.ReadAll(resp.Body)
	return rcds, json.Unmarshal(txt, &cr)
}
