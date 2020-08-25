package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"

	"models"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

//==============================================================================
// web api
//==============================================================================
func (self *WebServer) apiAuthHandler(c *gin.Context) {
	shortUserId := c.GetHeader("X-Api-User")
	hash := c.GetHeader("X-Api-AuthHash")

	t, tExist := c.GetQuery("t")
	q, qExist := c.GetQuery("q")

	var t64 int64
	fmt.Sscanf(t, "%d", &t64)

	if time.Now().Unix()-t64 > 60 || time.Now().Unix()-t64 < -60 {
		self.resp(c, 400, &CR{
			Message: "Expire",
			Code:    CodeBadData,
		})
		c.Abort()
		return
	}
	if !tExist || !qExist {
		self.resp(c, 400, &CR{
			Message: "param required",
			Code:    CodeExpire,
		})
		c.Abort()
		return
	}

	//lazy load
	cache := self.cache
	uidKey := shortUserId + ".uid"
	uidv, exist := cache.Get(uidKey)
	var tokenString string
	if exist {
		c.Set("uid", uidv.(int64))
		token, _ := cache.Get(shortUserId + ".token")
		tokenString = token.(string)
	} else {
		session := self.orm.NewSession()
		defer session.Close()

		var user models.TblUser
		exist, err := session.Where(`name=?`, shortUserId).Get(&user)
		if err != nil {
			logrus.Errorf("[webapi.go::apiAuthHandler] orm.Get: %v", err)
			self.resp(c, 401, &CR{
				Message: "Failed",
				Code:    CodeServerInternal,
			})
			c.Abort()
			return
		} else if !exist {
			self.resp(c, 401, &CR{
				Message: "param required",
				Code:    CodeNoAuth,
			})
			c.Abort()
		}
		cache.Set(uidKey, user.Id, AuthExpire)
		cache.Set(shortUserId+".token", user.Token, AuthExpire)
		tokenString = user.Token
	}

	h := md5.New()
	h.Write([]byte(tokenString))
	h.Write([]byte(t))
	h.Write([]byte(q))
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

	var rcd models.TblDns
	exist, err := session.Where(`uid=?`, id).And(`domain=?`, domain).Get(&rcd)

	if err != nil {
		self.resp(c, 502, &CR{
			Message: "domain parameter required",
			Code:    CodeServerInternal,
		})
		return
	} else if !exist {
		self.resp(c, 400, &CR{
			Message: "Not Found",
			Code:    CodeNoData,
		})
		return
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Result: &DnsRecord{
			Domain: rcd.Domain,
			//TODO: api
		},
	})
}

func (self *WebServer) queryHttpRecord(c *gin.Context) {
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

	var rcd models.TblHttp
	exist, err := session.Where(`uid=?`, id).And(`domain=?`, domain).Get(&rcd)

	if err != nil {
		self.resp(c, 502, &CR{
			Message: "domain parameter required",
			Code:    CodeServerInternal,
		})
		return
	} else if !exist {
		self.resp(c, 400, &CR{
			Message: "Not Found",
			Code:    CodeNoData,
		})
		return
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Result: &HttpRecord{
			Domain: rcd.Domain,
			//TODO: api
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

	host, _, err := net.SplitHostPort(c.Request.Host)
	if err != nil {
		host = c.Request.Host
	}

	var uid int64
	_, shortId, _ := parseDomain(host, self.Domain)
	cache := self.cache
	v, exist := cache.Get(shortId + ".suser")
	if exist {
		user := v.(*models.TblUser)
		uid = user.Id
	}

	_, err = session.InsertOne(&models.TblHttp{
		Uid:    uid,
		Ip:     c.ClientIP(),
		Domain: host,
		Ua:     c.GetHeader("User-Agent"),
		Ctype:  c.GetHeader("Content-Type"),
		Method: c.Request.Method,
		Path:   c.Param("any"),
		Ctime:  time.Now(),
		Data:   data.String(), //TODO: anti-xss
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
