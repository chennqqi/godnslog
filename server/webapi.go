package server

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sort"
	"time"

	"github.com/chennqqi/godnslog/models"

	"github.com/chennqqi/goutils/ginutils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//==============================================================================
// web api
//==============================================================================
func (self *WebServer) apiAuthHandler(c *gin.Context) {
	host := c.GetHeader("X-Forwarded-Host")
	var err error
	if host == "" {
		host, _, err = net.SplitHostPort(c.Request.Host)
		if err != nil {
			host = c.Request.Host
		}
	}

	_, shortId, _ := parseDomain(host, self.Domain)
	c.Set("host", host)
	c.Set("shortId", shortId)
	proto := c.GetHeader("X-Forwarded-Proto")
	if proto == "" {
		proto = "http"
	}
	c.Set("proto", proto)

	store := self.store
	domainKey := shortId + ".suser"
	v, exist := store.Get(domainKey)
	if !exist {
		self.resp(c, 401, &CR{
			Message: "No User",
			Code:    CodeNoAuth,
		})
		c.Abort()
		return
	}
	user := v.(*models.TblUser)
	c.Set("uid", user.Id)
	token := user.Token

	//authorization 1
	t64, err := ginutils.GetHeaderInt64(c, "t")
	if err != nil {
		self.resp(c, 401, &CR{
			Message: "No param time",
			Code:    CodeBadData,
		})
		c.Abort()
		return
	}

	//authorization1: verify time
	if time.Now().Unix()-t64 > 60 || time.Now().Unix()-t64 < -60 {
		self.resp(c, 400, &CR{
			Message: "Expire",
			Code:    CodeBadData,
		})
		c.Abort()
		return
	}

	//authorization2: verify hash
	hash, hashExist := c.GetQuery("hash")
	if !hashExist {
		self.resp(c, 400, &CR{
			Message: "No hash",
			Code:    CodeBadData,
		})
		c.Abort()
		return
	}
	querys := c.Request.URL.Query()
	var keys []string
	for k, _ := range querys {
		if k != "hash" {
			keys = append(keys, k)
		}
	}

	h := md5.New()
	sort.Strings(keys)
	for _, key := range keys {
		value := querys.Get(key)
		h.Write([]byte(value))
	}
	h.Write([]byte(token))
	expectHash := hex.EncodeToString(h.Sum(nil))

	if hash != expectHash {
		self.resp(c, 401, &CR{
			Message: "Auth failed",
			Code:    CodeNoAuth,
		})
		c.Abort()
		return
	}
}

func (self *WebServer) queryDnsRecord(c *gin.Context) {
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	id := c.GetInt64("uid")
	domain, domainExist := c.GetQuery("domain")
	if !domainExist {
		self.resp(c, 400, &CR{
			Message: "domain parameter required",
			Code:    CodeBadData,
		})
		return
	}

	session = session.Where(`uid=?`, id)
	exactly, _ := ginutils.GetQueryBoolean(c, "exactly")

	if exactly {
		session = session.And(`domain like ?`, "%"+domain+"%")
	} else {
		session = session.And(`domain like ?`, "%"+domain+"%")
	}

	var rcds []models.TblDns
	err := session.Limit(self.DefaultQueryApiMaxItem).Find(&rcds)
	if err != nil {
		self.resp(c, 502, &CR{
			Message: "domain parameter required",
			Code:    CodeServerInternal,
		})
		return
	}

	items := make([]DnsRecord, len(rcds))
	for i := 0; i < len(rcds); i++ {
		item := &items[i]
		rcd := &rcds[i]
		item.Domain = rcd.Domain
		item.Ip = rcd.Ip
		item.Ctime = rcd.Ctime
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Result: map[string]interface{}{
			"type": "dns",
			"data": items,
		},
	})
}

func (self *WebServer) queryHttpRecord(c *gin.Context) {
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	id := c.GetInt64("uid")
	suffix, domainExist := c.GetQuery("suffix")
	if !domainExist {
		self.resp(c, 400, &CR{
			Message: "domain parameter required",
			Code:    CodeBadData,
		})
		return
	}

	session = session.Where(`uid=?`, id)

	exactly, _ := ginutils.GetQueryBoolean(c, "exactly")
	if exactly {
		session = session.And(`url like ?`, "%"+suffix)
	} else {
		session = session.And(`url like ?`, "%"+suffix+"%")
	}

	var rcds []models.TblHttp
	err := session.Limit(self.DefaultQueryApiMaxItem).Find(&rcds)

	if err != nil {
		self.resp(c, 502, &CR{
			Message: "domain parameter required",
			Code:    CodeServerInternal,
		})
		return
	}
	items := make([]HttpRecord, len(rcds))
	for i := 0; i < len(rcds); i++ {
		item := &items[i]
		rcd := &rcds[i]

		item.Url = rcd.Url
		item.Ctype = rcd.Ctype
		item.Ip = rcd.Ip
		item.Method = rcd.Method
		item.Ua = rcd.Ua
		item.Data = rcd.Data
		item.Ctime = rcd.Ctime
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Result: map[string]interface{}{
			"type": "http",
			"data": items,
		},
	})
}

func (self *WebServer) record(c *gin.Context) {
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	var data bytes.Buffer
	io.Copy(&data, c.Request.Body)
	c.Request.Body.Close()

	var uid int64
	host := c.GetString("host")
	proto := c.GetString("proto")
	url := fmt.Sprintf("%v://%v%v", proto, host, c.Request.URL.EscapedPath())

	_, shortId, _ := parseDomain(host, self.Domain)
	store := self.store
	v, exist := store.Get(shortId + ".suser")
	if exist {
		user := v.(*models.TblUser)
		uid = user.Id
	}

	_, err := session.InsertOne(&models.TblHttp{
		Uid:    uid,
		Ip:     c.ClientIP(),
		Url:    url,
		Ua:     c.GetHeader("User-Agent"),
		Ctype:  c.GetHeader("Content-Type"),
		Method: c.Request.Method,
		// Path:   c.Param("any"),
		Ctime: time.Now(),
		Data:  data.String(), //TODO: anti-xss
	})
	if err != nil {
		logrus.Errorf("[webapi.go::Record] orm.InsertOne: %v", err)
		self.resp(c, 502, &CR{
			Message: "Failed",
			Code:    CodeServerInternal,
		})
		return
	}
	self.resp(c, 200, &CR{
		Message: "OK",
	})
}
