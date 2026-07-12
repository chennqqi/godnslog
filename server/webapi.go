package server

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sort"
	"strings"
	"time"

	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/models"
	"github.com/chennqqi/goutils/ginutils"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ==============================================================================
// web api
// ==============================================================================
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
		Data:    items,
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
		Data:    items,
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

	httpRecord := &models.TblHttp{
		Uid:    uid,
		Ip:     c.ClientIP(),
		Path:   path,
		Ua:     c.GetHeader("User-Agent"),
		Ctype:  c.GetHeader("Content-Type"),
		Var:    c.Param("any"),
		Method: c.Request.Method,
		Ctime:  time.Now(),
		Data:   data.String(),
	}
	_, err := session.InsertOne(httpRecord)
	if err != nil {
		logrus.Errorf("[webapi.go::Record] orm.InsertOne: %v", err)
		self.resp(c, 502, &CR{
			Message: "Failed",
			Code:    CodeServerInternal,
		})
		return
	}
	// Dual-write to unified interactions table with attribution
	interaction := v2models.FromTblHttpWithAttribution(httpRecord, self.orm)
	if _, err2 := session.InsertOne(interaction); err2 != nil {
		logrus.Errorf("[webapi.go::Record] dual-write interactions: %v", err2)
	}

	// Trigger matching workflows asynchronously
	self.triggerWorkflows(interaction)

	// Check for Workflow-driven HTTP response overrides (SCA-02)
	// This runs before Payload-level CustomResponse so workflow rules
	// take highest priority.
	if wfResp := self.checkWorkflowResponse(); wfResp != nil {
		if wfResp.Redirect != "" {
			c.Redirect(wfResp.Status, wfResp.Redirect)
			return
		}
		status := wfResp.Status
		if status == 0 {
			status = 200
		}
		for k, v := range wfResp.Headers {
			c.Header(k, v)
		}
		c.String(status, wfResp.Body)
		return
	}

	// Check for custom HTTP response configured on the payload (SCA-02)
	token := c.Param("any")
	if token != "" {
		var payload v2models.Payload
		has, err := self.orm.Where("token = ?", token).Get(&payload)
		if err == nil && has && payload.CustomResponse != "" {
			var customResp struct {
				Status   int               `json:"status"`
				Headers  map[string]string `json:"headers"`
				Body     string            `json:"body"`
				Redirect string            `json:"redirect"`
			}
			if err := json.Unmarshal([]byte(payload.CustomResponse), &customResp); err == nil {
				if customResp.Redirect != "" {
					c.Redirect(customResp.Status, customResp.Redirect)
					return
				}
				status := customResp.Status
				if status == 0 {
					status = 200
				}
				for k, v := range customResp.Headers {
					c.Header(k, v)
				}
				c.String(status, customResp.Body)
				return
			}
		}
	}

	// Echo back the full request in the response body
	var echoBody strings.Builder
	echoBody.WriteString("=== Request Echo ===\n")
	echoBody.WriteString(c.Request.Method + " " + c.Request.URL.String() + "\n")
	for k, v := range c.Request.Header {
		echoBody.WriteString(k + ": " + strings.Join(v, ", ") + "\n")
	}
	echoBody.WriteString("\n")
	if bodyStr := data.String(); bodyStr != "" {
		echoBody.WriteString(bodyStr + "\n")
	}
	echoBody.WriteString("===================\n")

	c.String(200, echoBody.String())
}

// workflowRespConfig holds HTTP response override values from a workflow action.
type workflowRespConfig struct {
	Status   int
	Headers  map[string]string
	Body     string
	Redirect string
}

// checkWorkflowResponse synchronously checks all enabled workflows for
// HTTP response overrides (ActionTypeResponse). Returns the first matching
// response config, or nil if no workflow defines a response action.
func (self *WebServer) checkWorkflowResponse() *workflowRespConfig {
	if self.workflowSvc == nil {
		return nil
	}
	workflows, err := self.workflowSvc.ListWorkflows("", boolPtr(true), 1, 100)
	if err != nil {
		logrus.Errorf("[webapi.go::checkWorkflowResponse] ListWorkflows: %v", err)
		return nil
	}
	for _, wf := range workflows.Items {
		for _, action := range wf.Actions {
			if action.Type == v2models.ActionTypeResponse && action.Enabled {
				cfg := &workflowRespConfig{}
				if v, ok := action.Config["status"]; ok {
					if s, ok := v.(float64); ok {
						cfg.Status = int(s)
					}
				}
				if v, ok := action.Config["redirect"]; ok {
					cfg.Redirect, _ = v.(string)
				}
				if v, ok := action.Config["body"]; ok {
					cfg.Body, _ = v.(string)
				}
				if v, ok := action.Config["headers"]; ok {
					if h, ok := v.(map[string]interface{}); ok {
						cfg.Headers = make(map[string]string, len(h))
						for k, val := range h {
							cfg.Headers[k] = fmt.Sprintf("%v", val)
						}
					}
				}
				return cfg
			}
		}
	}
	return nil
}
