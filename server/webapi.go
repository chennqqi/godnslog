package server

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
	"sort"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/models"

	"github.com/chennqqi/goutils/ginutils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//==============================================================================
// web api
//==============================================================================
func (self *WebServer) dataPreHandler(c *gin.Context) {
	host := c.GetHeader("X-Forwarded-Host")

	if host == "" {
		host = c.GetHeader("host")
		if host == "" {
			host = c.Request.Host
		}
	}
	if strings.Contains(host, ":") {
		host, _, _ = net.SplitHostPort(host)
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
	c.Set("token", user.Token)
}

func (self *WebServer) dataAuthHandler(c *gin.Context) {
	//authorization 1: t=$timestamp
	token := c.GetString("token")

	t64, err := ginutils.GetQueryInt64(c, "t")
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

// dig ${q}.${shortId}.godnslog.com
func (self *WebServer) queryDnsRecord(c *gin.Context) {
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	id := c.GetInt64("uid")
	variable, domainExist := c.GetQuery("q")
	if !domainExist {
		self.resp(c, 400, &CR{
			Message: "domain parameter required",
			Code:    CodeBadData,
		})
		return
	}

	session = session.Where(`uid=?`, id)
	blur, _ := ginutils.GetQueryInt(c, "blur")
	if blur == 0 {
		session = session.And(`var = ?`, variable)
	} else {
		session = session.And(`var like ?`, "%"+variable+"%")
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

	items := make([]models.DnsRecord, len(rcds))
	for i := 0; i < len(rcds); i++ {
		item := &items[i]
		rcd := &rcds[i]
		item.Domain = rcd.Domain
		item.Ip = rcd.Ip
		item.Ctime = rcd.Ctime
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Result:  items,
	})
}

// curl http://${shortId}.godnslog.com/log/${q}
func (self *WebServer) queryHttpRecord(c *gin.Context) {
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	id := c.GetInt64("uid")
	q, domainExist := c.GetQuery("q")
	if !domainExist {
		self.resp(c, 400, &CR{
			Message: "domain parameter required",
			Code:    CodeBadData,
		})
		return
	}
	session = session.Where(`uid=?`, id)

	blur, _ := ginutils.GetQueryInt(c, "blur")
	if blur == 0 {
		session = session.And(`var = ?`, q)
	} else {
		session = session.And(`var like ?`, "%"+q+"%")
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

		item.Path = rcd.Path
		item.Ctype = rcd.Ctype
		item.Ip = rcd.Ip
		item.Method = rcd.Method
		item.Ua = rcd.Ua
		item.Data = rcd.Data
		item.Ctime = rcd.Ctime
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Result:  items,
	})
}

func (self *WebServer) record(c *gin.Context) {
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	var data bytes.Buffer
	io.Copy(&data, c.Request.Body)
	c.Request.Body.Close()

	path := c.Request.URL.EscapedPath()

	var uid int64
	shortId := c.Param("shortId")

	store := self.store
	v, exist := store.Get(shortId + ".suser")
	if exist {
		user := v.(*models.TblUser)
		uid = user.Id
	}

	_, err := session.InsertOne(&models.TblHttp{
		Uid:    uid,
		Ip:     c.ClientIP(),
		Path:   path,
		Ua:     c.GetHeader("User-Agent"),
		Ctype:  c.GetHeader("Content-Type"),
		Var:    c.Param("any"),
		Method: c.Request.Method,
		Ctime:  time.Now(),
		Data:   data.String(),
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
