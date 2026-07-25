package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/cache"
	"github.com/chennqqi/godnslog/internal/demo"
	"github.com/chennqqi/godnslog/internal/ha"
	"github.com/chennqqi/godnslog/internal/marketplace"
	v2models "github.com/chennqqi/godnslog/internal/models"
	"github.com/chennqqi/godnslog/internal/workflow"
	"github.com/chennqqi/godnslog/models"

	"github.com/chennqqi/goutils/ginutils"
	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type MyClaims struct {
	Seed string `json:"seed"`
	jwt.StandardClaims
}

// ==============================================================================
// ui standard api
// ==============================================================================
func (self *WebServer) respData(c *gin.Context, status, code int,
	message string, data interface{}) {
	c.JSON(status, &CR{
		Message:   message,
		Code:      code,
		Timestamp: time.Now().Unix(),
	})
}

func (self *WebServer) resp(c *gin.Context, status int, cr *CR) {
	cr.Timestamp = time.Now().Unix()
	c.JSON(status, cr)
}

func (self *WebServer) initDatabase() error {
	orm := self.orm
	orm.SetTZDatabase(time.Local)
	orm.SetTZLocation(time.Local)

	err := orm.Sync2(
		// 1.0 legacy models
		&models.TblDns{},
		&models.TblHttp{},
		&models.TblUser{},
		&models.TblResolve{},
		&models.TblCase{},
		&models.TblPayload{},
		&models.TblInteraction{},
		&models.TblAPIKey{},
		// 2.0 core models
		// Note: User/Resolve are wrappers around legacy tbl_user/tbl_resolve;
		// do not sync them as separate tables.
		&v2models.Interaction{},
		&v2models.Case{},
		&v2models.Payload{},
		&v2models.APIKey{},
		&v2models.Workflow{},
		&v2models.Response{},
		&v2models.Canary{},
		&v2models.RebindingRule{},
		&v2models.Listener{},
		&v2models.ListenerInteraction{},
		&v2models.SMTPMessage{},
		&v2models.LDAPQuery{},
		&v2models.SMBRequest{},
		&v2models.FTPCommand{},
		&v2models.ScannerRun{},
		&v2models.AuditLog{},
		&v2models.AgentRun{},
		&v2models.AgentOperation{},
		&v2models.Settings{},
		// HA models
		&ha.ClusterNode{},
		&ha.ClusterConfig{},
		&ha.HealthCheck{},
		&ha.LeaderElection{},
		// Workflow persistent action log
		&workflow.PersistentActionLog{},
		// Marketplace models
		&marketplace.Plugin{},
		&marketplace.PluginVersion{},
		&marketplace.PluginReview{},
		&marketplace.Template{},
		&marketplace.TemplateReview{},
		&marketplace.PluginInstallation{},
	)
	if err != nil {
		logrus.Errorf("[webui.go::initDatabase] orm.Sync: %v", err)
		return err
	}

	if err := self.initMarketplaceSeed(); err != nil {
		logrus.Errorf("[webui.go::initDatabase] initMarketplaceSeed: %v", err)
	}

	// check superUser
	{
		count, err := orm.Count(&models.TblUser{})
		if err != nil {
			logrus.Errorf("[webui.go::initDatabase] orm.Count(user): %v", err)
			return err
		}
		// if there is no supser user when system first init
		if count == 0 {
			randomPass := genRandomString(12)
			// In test mode, use fixed password
			if self.TestMode {
				randomPass = "test123"
			}
			_, err = orm.InsertOne(&models.TblUser{
				Name:          "admin",
				Email:         "admin@godnslog.com",
				ShortId:       genShortId(),
				Pass:          makePassword(randomPass),
				Token:         genRandomToken(),
				Role:          roleSuper,
				Lang:          self.DefaultLanguage,
				CleanInterval: self.DefaultCleanInterval,
			})
			if err != nil {
				logrus.Errorf("[webui.go::initDatabase] orm.InsertOne(user): %v", err)
				return err
			}
			fmt.Printf("Init super admin user with password: %v\n", randomPass)
		}
	}

	// check and add guest user
	if self.WithGuest {
		var guestUser models.TblUser
		exist, err := orm.Where(`name=?`, "guest").Get(&guestUser)
		if err != nil {
			logrus.Errorf("[webui.go::initDatabase] orm.Get(user.name='guest'): %v", err)
			return err
		} else if !exist {
			guestUser.Name = "guest"
			guestUser.Email = "guest@godnslog.com"
			guestUser.Pass = makePassword("guest123")
			guestUser.Token = genRandomToken()
			guestUser.Role = roleGuest
			guestUser.Lang = self.DefaultLanguage
			guestUser.CleanInterval = self.DefaultCleanInterval
			_, err = orm.InsertOne(&guestUser)
			if err != nil {
				logrus.Errorf("[webui.go::initDatabase] orm.InsertOne(user.name=guest): %v", err)
				return err
			}
		}
	}

	// Default DNS A records so the web interface is reachable via
	// domain without manual DNS setup. Both www and @ are needed.
	for _, host := range []string{"www", "@"} {
		var rcd models.TblResolve
		exist, err := orm.Where(`host=?`, host).And(`type=?`, `A`).Get(&rcd)
		if err != nil {
			logrus.Errorf("[webui.go::initDatabase] orm.Get(resolve host=%s): %v", host, err)
			return err
		}
		if !exist {
			rcd.Host = host
			rcd.Value = self.IP
			rcd.Type = "A"
			rcd.Ttl = 600
			orm.InsertOne(&rcd)
		} else if rcd.Value != self.IP {
			rcd.Value = self.IP
			orm.Update(&rcd)
		}
	}

	store := self.store
	// sync user
	orm.Iterate(new(models.TblUser), func(idx int, bean interface{}) error {
		user := bean.(*models.TblUser)
		userKey := fmt.Sprintf("%v.user", user.Id)
		store.Set(userKey, user, cache.NoExpiration)
		domainKey := fmt.Sprintf("%v.suser", user.ShortId)
		store.Set(domainKey, user, cache.NoExpiration)
		return nil
	})

	// sync standard dns service data
	var hosts []string
	err = orm.Table(&models.TblResolve{}).GroupBy("host").Cols("host").Find(&hosts)
	if err != nil {
		logrus.Panicf("[webui.go::initDatabase] orm.GroupBy: %v", err)
	}

	for i := 0; i < len(hosts); i++ {
		var all []models.TblResolve
		err := orm.Where(`host=?`, hosts[i]).Find(&all)
		if err != nil {
			logrus.Panicf("[webui.go::initDatabase] orm.Find(%v): %v", hosts[i], err)
		}
		self.updateResolveCache(hosts[i], "", all)
	}

	return nil
}

// initMarketplaceSeed inserts sample marketplace entries if the store is empty.
func (self *WebServer) initMarketplaceSeed() error {
	store := marketplace.NewXormStore(self.orm)
	svc := marketplace.NewService(store)
	ctx := context.Background()

	plugins, err := svc.ListPlugins(ctx, marketplace.PluginFilters{})
	if err != nil {
		return err
	}
	if len(plugins) == 0 {
		sample := &marketplace.Plugin{
			ID:          v2models.GenerateID(),
			Name:        "Sample DNS Logger",
			Description: "A sample processor plugin that logs DNS interactions to a webhook.",
			Version:     "1.0.0",
			Author:      "GODNSLOG Team",
			Type:        "processor",
			Category:    "dns",
			Code:        "// sample plugin code\nfunction process(interaction) {\n  console.log(interaction);\n}",
			Language:    "javascript",
			IsPublished: true,
			IsOfficial:  true,
		}
		if err := svc.CreatePlugin(ctx, sample); err != nil {
			return err
		}
	}

	templates, err := svc.ListTemplates(ctx, marketplace.TemplateFilters{})
	if err != nil {
		return err
	}
	if len(templates) == 0 {
		sample := &marketplace.Template{
			ID:          v2models.GenerateID(),
			Name:        "SSRF Basic Payload",
			Description: "Basic SSRF OAST payload template for scanner integration.",
			Type:        "payload",
			Category:    "ssrf",
			Content:     "id: ssrf-basic\ninfo:\n  name: SSRF Basic OAST\n  author: GODNSLOG Team\n  severity: high",
			Format:      "yaml",
			IsPublished: true,
			IsOfficial:  true,
		}
		if err := svc.CreateTemplate(ctx, sample); err != nil {
			return err
		}
	}
	return nil
}

func (self *WebServer) authHandler(c *gin.Context) {
	// Try JWT authentication first
	if user, err := self.authenticateJWT(c); err == nil && user != nil {
		c.Set("user", user)

		// Set legacy context values for backward compatibility
		var uid int64 = int64(user.Id)
		c.Set("id", uid)
		c.Set("username", user.Name)
		c.Set("email", user.Email)
		c.Set("role", user.Role)

		// Demo user restriction: block demo users from admin/system config operations
		if self.demoMgr != nil && demo.IsDemoUser(user.Name) {
			c.Set("is_demo_user", true)
			// Check if the request path is a restricted admin/config operation
			path := c.Request.URL.Path
			if isDemoRestrictedPath(path) {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"code":    403,
					"message": "demo users cannot perform admin or system configuration operations",
				})
				return
			}
		}

		c.Next()
		return
	}

	// Try API key authentication
	if key, err := self.authenticateAPIKey(c); err == nil && key != nil {
		c.Set("api_key", key)
		c.Next()
		return
	}

	// No valid credentials found
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code":    401,
		"message": "unauthorized",
	})
}

// isDemoRestrictedPath returns true if the given API path is restricted for demo users.
func isDemoRestrictedPath(path string) bool {
	// Admin user management
	if strings.Contains(path, "/api/admin/") {
		return true
	}
	// System settings
	if strings.Contains(path, "/api/setting/") {
		return true
	}
	// API key management
	if strings.Contains(path, "/api/v2/apikeys") {
		return true
	}
	// Security settings
	if strings.Contains(path, "/api/v2/settings") {
		return true
	}
	return false
}

func (self *WebServer) verifyAdminPermission(c *gin.Context) {
	T := getTranslateFunc(c)

	role := c.GetInt("role")
	switch role {
	case roleAdmin, roleSuper:
		return
	default:
		self.resp(c, 403, &CR{
			Message: T("bad permission"),
			Code:    CodeNoPermission,
		})
		c.Abort()
		return
	}
}

//==============================================================================
//									user auth
//==============================================================================

// @Summary userLogin
// @Description get Dns Record by user query
// @Accept  json
// @Produce  json
// @Param   some_id path int true "Some ID"
// @Success 200 {string} CR	"OK"
// @Failure 502 {object} CR "BadService"
// @Failure 403 {object} CR "Forbidden"
// @Failure 401 {object} CR "Unauthorized"
// @Router /user/login [post]
func (self *WebServer) getCaptcha(c *gin.Context) {
	if self.captchaSvc == nil {
		self.resp(c, 200, &CR{
			Code: CodeOK,
			Data: map[string]interface{}{
				"captcha_id":   "",
				"image_base64": "",
				            "thumb_base64": "",
                "block_y":      0,
			},
		})
		return
	}

	captchaID, imageBase64, thumbBase64, blockY, err := self.captchaSvc.Generate()
	if err != nil {
		logrus.Errorf("[webui.go::getCaptcha] Generate: %v", err)
		self.resp(c, 502, &CR{
			Code:    models.CodeServerInternal,
			Message: "Failed to generate captcha",
		})
		return
	}

	self.resp(c, 200, &CR{
		Code: models.CodeOK,
		Data: map[string]interface{}{
			"captcha_id":   captchaID,
			"image_base64": imageBase64,
			            "thumb_base64": thumbBase64,
            "block_y":      blockY,
		},
	})
}

func (self *WebServer) userLogin(c *gin.Context) {
	T := getTranslateFunc(c)

	var req LoginRequest
	err := c.BindJSON(&req)
	if err != nil {
		logrus.Infof("[webui.go::userLogin] bad input param")
		self.resp(c, 400, &CR{
			Code:    CodeBadData,
			Message: T("bad input"),
		})
		return
	}

	// Captcha verification (if enabled)
	if self.captchaSvc != nil {
		if req.CaptchaID == "" {
			self.resp(c, 400, &CR{
				Code:    CodeBadData,
				Message: T("bad request"),
			})
			return
		}
		if !self.captchaSvc.Verify(req.CaptchaID, req.CaptchaValue, req.CaptchaY) {
			self.resp(c, 400, &CR{
				Code:    CodeBadData,
				Message: T("bad request"),
			})
			return
		}
	}

	session := self.orm.NewSession()
	defer session.Close()
	var user = new(models.TblUser)
	exist, err := session.Where(`name=?`, req.Username).
		Or(`email=?`, req.Email).Get(user)

	if err != nil {
		logrus.Errorf("[webui.go::userLogin] orm.Get: %v", err)
		self.respData(c, 502, CodeServerInternal, T("bad service"), nil)
		return
	} else if !exist {
		logrus.Infof("[webui.go::userLogin] not found: %v", req)
		self.respData(c, 401, CodeBadData, T("bad request"), nil)
		return
	}
	err = comparePassword(req.Password, user.Pass)
	if err != nil {
		logrus.Infof("[webui.go::userLogin] password not match")
		self.respData(c, 401, CodeBadData, T("bad request"), nil)
		return
	}

	now := time.Now()
	seed := getSecuritySeed()
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, MyClaims{
		seed,
		jwt.StandardClaims{
			Id:        fmt.Sprintf("%v", user.Id),
			Audience:  user.Name,
			Subject:   user.Email,
			ExpiresAt: now.Add(3600 * 24 * time.Second).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    self.Domain,
		},
	})

	tokenString, err := token.SignedString([]byte(self.verifyKey))
	if err != nil {
		logrus.Errorf("[webui.go::userLogin] token.SignedString: %v", err)
		self.respData(c, 502, CodeServerInternal, T("bad service"), nil)
		return
	}
	store := self.store

	store.Set(fmt.Sprintf("%v.seed", user.Id), seed, self.AuthExpire)
	store.Set(fmt.Sprintf("%v.user", user.Id), user, cache.NoExpiration)

	// Return data in format expected by frontend: { code: 0, data: { token, user } }
	self.resp(c, 200, &CR{
		Code:    CodeOK,
		Message: T("OK"),
		Data: map[string]interface{}{
			"token": tokenString,
			"user": map[string]interface{}{
				"id":       user.Id,
				"username": user.Name,
				"email":    user.Email,
				"role":     user.Role,
				"lang":     user.Lang,
			},
		},
	})
}

// @Summary userLogout
// @Description get Dns Record by user query
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {string} CR	"OK"
// @Failure 502 {object} CR "BadService"
// @Failure 403 {object} CR "Forbidden"
// @Failure 401 {object} CR "Unauthorized"
// @Router /user/logout [post]
func (self *WebServer) userLogout(c *gin.Context) {
	T := getTranslateFunc(c)

	store := self.store
	id := c.GetInt64("id")
	store.Delete(fmt.Sprintf("%v.seed", id))
	store.Delete(fmt.Sprintf("%v.user", id))
	self.resp(c, 200, &CR{
		Message: T("OK"),
	})
}

func parseRole(roleValue int) models.Role {
	var role models.Role
	role.Id = "normal"
	role.Name = "普通用户"
	role.Permissions = []models.Permission{
		models.Permission{
			RoleId:         roleNormal,
			PermissionId:   "document",
			PermissionName: "文档",
		},
	}
	switch roleValue {
	case roleAdmin, roleSuper:
		role.Id = "admin"
		role.Name = "管理员"
		role.Permissions = append(role.Permissions, []models.Permission{
			models.Permission{
				RoleId:         roleNormal,
				PermissionId:   "setting",
				PermissionName: "设置",
			},
			models.Permission{
				RoleId:         roleAdmin,
				PermissionId:   "manage",
				PermissionName: "管理用户",
			},
			models.Permission{
				RoleId:         roleNormal,
				PermissionId:   "record",
				PermissionName: "记录",
			},
		}...)

	case roleNormal:
		role.Permissions = append(role.Permissions, []models.Permission{
			models.Permission{
				RoleId:         roleNormal,
				PermissionId:   "setting",
				PermissionName: "设置",
			},
			models.Permission{
				RoleId:         roleNormal,
				PermissionId:   "record",
				PermissionName: "记录",
			},
		}...)
	case roleGuest:
		role.Id = "guest"
		role.Name = "访客"
	}
	return role
}

// @Summary userInfo
// @Description get Dns Record by user query
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} CR "We need ID!!"
// @Failure 404 {object} CR "Can not find ID"
// @Failure 401 {object} CR "Can not find ID"
// @Router /user/info [get]
func (self *WebServer) userInfo(c *gin.Context) {
	T := getTranslateFunc(c)
	id := c.GetInt64("id")
	session := self.orm.NewSession()
	defer session.Close()

	store := self.store
	userKey := fmt.Sprintf("%v.user", id)
	v, exist := store.Get(userKey)
	var user *models.TblUser
	if !exist {
		user = new(models.TblUser)
		exist, err := session.ID(id).Get(user)
		if err != nil {
			logrus.Errorf("[webui.go::userInfo] orm.Get: %v", err)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		}
		if !exist {
			logrus.Errorf("[webui.go::userInfo] No such user")
			self.resp(c, 400, &CR{
				Message: T("No such user"),
				Code:    CodeBadData,
			})
			return
		}
		store.Set(userKey, user, cache.NoExpiration)
		domainKey := fmt.Sprintf("%v.suser", user.ShortId)
		store.Set(domainKey, user, cache.NoExpiration)
	} else {
		user = v.(*models.TblUser)
	}

	//TODO: UserInfo from cache, role & permissions
	self.resp(c, 200, &CR{
		Message: T("OK"),
		Code:    CodeOK,
		Data: UserInfo{
			Id:    user.Id,
			Name:  user.Name,
			Email: user.Email,
			Role:  parseRole(user.Role),
		},
	})
}

// @Summary userInfo
// @Description get Dns Record by user query
// @Accept  json
// @Produce  json
// @Param   some_id     path    int     true        "Some ID"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} CR "We need ID!!"
// @Failure 404 {object} CR "Can not find ID"
// @Failure 401 {object} CR "Can not find ID"
// @Router /admin/user/list [get]
func (self *WebServer) userList(c *gin.Context) {
	T := getTranslateFunc(c)

	pageNo, pageNoErr := ginutils.GetQueryInt(c, "pageNo")
	if pageNoErr != nil || pageNo <= 0 {
		pageNo = 1
	}
	pageSize, pageSizeErr := ginutils.GetQueryInt(c, "pageSize")
	if pageSizeErr != nil {
		pageSize = 10
	}

	session := self.orm.NewSession()
	defer session.Close()

	session = session.Where(`id>1`)
	var items []models.TblUser
	count, err := session.Limit(pageSize, (pageNo-1)*pageSize).FindAndCount(&items)
	if err != nil {
		self.resp(c, 502, &CR{
			Code:    CodeServerInternal,
			Message: T("Failed"),
		})
		return
	}

	var resp UserListResp
	resp.TotalCount = int(count)
	resp.PageSize = pageSize
	resp.PageNo = pageNo
	resp.TotalPage = (resp.TotalCount + (pageSize - 1)) / pageSize
	resp.Data = make([]models.UserInfo, len(items))
	for i := 0; i < len(items); i++ {
		rcd := &resp.Data[i]
		item := &items[i]
		rcd.Id = item.Id
		rcd.Name = item.Name
		rcd.Email = item.Email
		rcd.Utime = item.Utime
		rcd.Role = parseRole(item.Role)
	}

	self.resp(c, 200, &CR{
		Message: T("OK"),
		Data:    &resp,
	})
}

// @Summary userNav
// @Description get navigation links for current user
// @Accept  json
// @Produce  json
// @Success 200 {string} string	"ok"
// @Router /auth/nav [get]
func (self *WebServer) userNav(c *gin.Context) {
}

//==============================================================================
//							user manage
//==============================================================================

// @Summary delUser
// @Description delete user by id
// @Accept  json
// @Produce  json
// @Param   ids     body    []int     true        "User IDs to delete"
// @Success 200 {object} CR	"OK"
// @Failure 400 {object} CR "Bad request"
// @Router /admin/user [delete]
func (self *WebServer) delUser(c *gin.Context) {
	T := getTranslateFunc(c)

	var req DeleteRecordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Infof("[webapi.go::delUser] parameter required")
		self.resp(c, 400, &CR{
			Message: T("param required"),
			Code:    CodeBadData,
		})
		return
	}
	var ids = make([]interface{}, len(req.Ids))
	for i := 0; i < len(req.Ids); i++ {
		ids[i] = req.Ids[i]
	}

	session := self.orm.NewSession()
	defer session.Close()

	//do not delete super user
	_, err = session.In("id", ids...).Delete(&models.TblUser{})
	if err != nil {
		logrus.Errorf("[webapi.go::delUser] orm.Delete: %v", err)
		self.resp(c, 502, &CR{
			Message: T("failed"),
			Code:    CodeServerInternal,
		})
		return
	}
	session.In("uid", ids).Delete(&models.TblDns{})
	session.In("uid", ids).Delete(&models.TblHttp{})

	cache := self.store
	for i := 0; i < len(req.Ids); i++ {
		seedKey := fmt.Sprintf("%v.seed", req.Ids[i])
		userKey := fmt.Sprintf("%v.user", req.Ids[i])
		v, exist := cache.Get(userKey)
		if exist {
			domainKey := fmt.Sprintf("%v.suser", v.(*models.TblUser).ShortId)
			cache.Delete(domainKey)
		}

		//logout these users
		cache.Delete(seedKey)
		cache.Delete(userKey)
	}

	self.resp(c, 200, &CR{
		Message: T("OK"),
	})
}

// @Summary addUser
// @Description add a new User
// @Accept  json
// @Produce  json
// @Param   body     body    UserRequest     true        "User data"
// @Success 200 {string} CR	"OK"
// @Failure 502 {object} CR "BadService"
// @Failure 403 {object} CR "Forbidden"
// @Failure 401 {object} CR "Unauthorized"
// @Router /admin/user [put]
func (self *WebServer) addUser(c *gin.Context) {
	T := getTranslateFunc(c)

	var req UserRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Infof("[webui.go::addUser] parameter format invalid")
		self.resp(c, 400, &CR{
			Message: T("Bad param"),
			Code:    CodeBadData,
		})
		return
	}

	if isWeakPass(req.Password) {
		self.resp(c, 400, &CR{
			Message: T("password too weak"),
			Code:    CodeBadData,
		})
		return
	}

	// TODO: other checks

	//random api Token
	session := self.orm.NewSession()
	defer session.Close()

	var item = models.TblUser{
		Name:          req.Name,
		Email:         req.Email,
		Role:          req.Role,
		Token:         genRandomToken(),
		ShortId:       genShortId(),
		Lang:          self.DefaultLanguage,
		Pass:          makePassword(req.Password),
		CleanInterval: self.DefaultCleanInterval,
	}
	_, err = session.InsertOne(&item)
	if self.IsDuplicate(err) {
		self.resp(c, 400, &CR{
			Message: T("Failed"),
			Code:    CodeBadData,
		})
		return
	} else if err != nil {
		self.resp(c, 502, &CR{
			Message: T("Failed"),
			Code:    CodeServerInternal,
		})
		return
	}
	self.resp(c, 200, &CR{
		Message: T("OK"),
	})
}

func (self *WebServer) setUser(c *gin.Context) {
	T := getTranslateFunc(c)

	var req UserRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Infof("[webapi.go::setUser] parameter required")
		self.resp(c, 400, &CR{
			Message: T("param invaid: ") + err.Error(),
			Code:    CodeBadData,
		})
		return
	}
	if req.Id < 1 {
		self.resp(c, 400, &CR{
			Message: T("Can't change"),
			Code:    CodeBadData,
		})
		return
	}

	store := self.store
	id := c.GetInt64("id")
	role := c.GetInt("role")
	session := self.orm.NewSession()
	defer session.Close()

	var user *models.TblUser

	switch role {
	case roleSuper, roleAdmin:
		//change other user
		session = session.ID(req.Id)
		if req.Password != "" {
			newPass := makePassword(req.Password)
			session = session.SetExpr(`pass`, customQuote(newPass))
		}
		if req.Email != "" {
			session = session.SetExpr(`email`, customQuote(req.Email))
		}
		if req.Name != "" {
			session = session.SetExpr(`name`, customQuote(req.Name))
		}

		_, err = session.Update(&models.TblUser{})
		if err != nil {
			sql, _ := session.LastSQL()
			logrus.Errorf("[webapi.go::setUser] orm.Update error: %v, sql:%v", err, sql)
			self.resp(c, 400, &CR{
				Message: "failed",
				Code:    CodeServerInternal,
			})
			return
		}

		//logout req.Id
		cache := self.store
		cache.Delete(fmt.Sprintf("%v.seed", req.Id))
		cache.Delete(fmt.Sprintf("%v.user", req.Id))
		self.resp(c, 200, &CR{
			Message: T("OK"),
		})

	case roleNormal:
		//allow change language only
		userKey := fmt.Sprintf("%v.user", id)
		v, exist := store.Get(userKey)
		if !exist {
			user = new(models.TblUser)
			exist, err := session.ID(id).Get(user)
			if err != nil {
				sql, _ := session.LastSQL()
				logrus.Errorf("[webapi.go::setUser] orm.Get error: %v, sql:%v", err, sql)
				self.resp(c, 502, &CR{
					Message: T("failed"),
					Code:    CodeServerInternal,
				})
				return
			} else if !exist {
				//this should not happend
				self.resp(c, 400, &CR{
					Message: T("Failed"),
					Code:    CodeBadData,
				})
				return
			}

		} else {
			user = v.(*models.TblUser)
		}
		dupUser := new(models.TblUser)
		*dupUser = *user

		_, err := session.ID(id).Cols("lang").Update(dupUser)
		if err != nil {
			sql, _ := session.LastSQL()
			logrus.Errorf("[webapi.go::setUser] orm.Update error: %v, sql:%v", err, sql)
			self.resp(c, 400, &CR{
				Message: T("failed"),
				Code:    CodeServerInternal,
			})
			return
		}
		store.Set(userKey, dupUser, cache.NoExpiration)
		domainKey := fmt.Sprintf("%v.suser", dupUser.ShortId)
		store.Set(domainKey, dupUser, cache.NoExpiration)
		self.resp(c, 200, &CR{
			Message: T("OK"),
		})
	}
}

func (self *WebServer) getAppSetting(c *gin.Context) {
	T := getTranslateFunc(c)
	id := c.GetInt64("id")
	store := self.store
	userKey := fmt.Sprintf("%v.user", id)
	v, exist := store.Get(fmt.Sprintf(userKey, id))
	var user *models.TblUser
	if !exist {
		session := self.orm.NewSession()
		defer session.Close()

		user = new(models.TblUser)
		exist, err := session.ID(id).Get(user)
		if err != nil {
			sql, _ := session.LastSQL()
			logrus.Errorf("[webui.go::getSecuritySetting] orm.Get error: %v, sql: %v", err, sql)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		} else if !exist {
			logrus.Errorf("[webui.go::getSecuritySetting] not found user(id=%v), this should not happend", id)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		}
		store.Set(userKey, user, cache.NoExpiration)
		domainKey := fmt.Sprintf("%v.suser", user.ShortId)
		store.Set(domainKey, user, cache.NoExpiration)
	} else {
		user = v.(*models.TblUser)
	}

	self.resp(c, 200, &CR{
		Message: "OK",
		Data: AppSetting{
			Rebind:    user.Rebind,
			Callback:  user.Callback,
			CleanHour: user.CleanInterval / 3600,
		},
	})
}

func (self *WebServer) setAppSetting(c *gin.Context) {
	T := getTranslateFunc(c)
	var req AppSetting
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Infof("[webui.go::setAppSetting] parameter format invalid")
		self.resp(c, 400, &CR{
			Message: T("Bad param"),
			Code:    CodeBadData,
		})
		return
	}

	id := c.GetInt64("id")
	store := self.store
	userKey := fmt.Sprintf("%v.user", id)
	v, exist := store.Get(userKey)
	session := self.orm.NewSession()
	defer session.Close()

	var user *models.TblUser
	if !exist {
		user = new(models.TblUser)
		exist, err := session.ID(id).Get(user)
		if err != nil {
			logrus.Errorf("[webuig.go::setAppSetting] orm.Get error: %v", err)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		} else if !exist {
			logrus.Errorf("[webuig.go::setAppSetting] not found user(id=%v), this should not happend", id)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		}
		store.Set(userKey, user, cache.NoExpiration)
		domainkey := fmt.Sprintf("%v.suser", user.ShortId)
		store.Set(domainkey, user, cache.NoExpiration)
	} else {
		user = v.(*models.TblUser)
	}

	dupUser := new(models.TblUser)
	*dupUser = *user
	dupUser.Rebind = req.Rebind
	dupUser.Callback = req.Callback
	dupUser.CleanInterval = req.CleanHour * 3600
	_, err = session.ID(id).Cols("rebind", "callback", "clean_interval").Update(dupUser)
	if err != nil {
		logrus.Errorf("[webui.go::setAppSetting] orm.Update error: %v", err)
		self.resp(c, 502, &CR{
			Message: T("Failed"),
			Code:    CodeServerInternal,
		})
		return
	}

	//update cache
	{
		domainKey := fmt.Sprintf("%v.suser", user.ShortId)
		userKey := fmt.Sprintf("%v.user", user.Id)
		store.Set(userKey, dupUser, cache.NoExpiration)
		store.Set(domainKey, dupUser, cache.NoExpiration)
	}

	self.resp(c, 200, &CR{
		Message: T("OK"),
	})
}

// change self password
func (self *WebServer) getSecuritySetting(c *gin.Context) {
	T := getTranslateFunc(c)
	id := c.GetInt64("id")
	store := self.store
	userKey := fmt.Sprintf("%v.user", id)
	v, exist := store.Get(userKey)
	var user *models.TblUser
	if !exist {
		session := self.orm.NewSession()
		defer session.Close()
		user = new(models.TblUser)
		exist, err := session.ID(id).Get(user)
		if err != nil {
			sql, _ := session.LastSQL()
			logrus.Errorf("[webuig.go::getSecuritySetting] orm.Get error: %v, sql: %v", err, sql)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		} else if !exist {
			logrus.Errorf("[webuig.go::getSecuritySetting] not found user(id=%v), this should not happend", id)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		}
		store.Set(userKey, user, cache.NoExpiration)
		domainkey := fmt.Sprintf("%v.suser", user.ShortId)
		store.Set(domainkey, user, cache.NoExpiration)
	} else {
		user = v.(*models.TblUser)
	}

	self.resp(c, 200, &CR{
		Message: T("OK"),
		Data: AppSecurity{
			HttpAddr: fmt.Sprintf("http://%v/log/%v/", self.IP, user.ShortId),
			DnsAddr:  user.ShortId + "." + self.Domain,
			Token:    user.Token,
		},
	})
}

// change self password
func (self *WebServer) setSecuritySetting(c *gin.Context) {
	T := getTranslateFunc(c)
	var req AppSecuritySet
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Infof("[webuig.go::setSecuritySetting] bad data")
		self.resp(c, 400, &CR{
			Message: T("bad param"),
			Code:    CodeBadData,
		})
		return
	}
	if isWeakPass(req.Password) {
		logrus.Warnf("[webuig.go::setSecuritySetting] weak password data")
		self.resp(c, 400, &CR{
			Message: T("password too weak"),
			Code:    CodeBadData,
		})
		return
	}

	id := c.GetInt64("id")
	session := self.orm.NewSession()
	defer session.Close()

	newPass := makePassword(req.Password)
	//logrus.Debugf("password:%v, hashpass=%v", req.Password, string(newPass))
	_, err = session.ID(id).SetExpr(`pass`, customQuote(newPass)).Update(&models.TblUser{})
	if err != nil {
		sql, _ := session.LastSQL()
		logrus.Errorf("[webuig.go::setSecuritySetting] orm.Update(%v), last SQL: %v", err, sql)
		self.resp(c, 502, &CR{
			Message: T("update Failed"),
			Code:    CodeServerInternal,
		})
		return
	}

	//logout & resp success
	self.userLogout(c)
}

//==============================================================================
// data api
//==============================================================================

// @Summary getDnsRecord
// @Description get Dns Record by user query
// @Accept  json
// @Produce  json
// @Param   ip     query    string     false        "IP address filter"
// @Param   domain  query    string     false        "Domain filter"
// @Param   date    query    string     false        "Date filter (RFC3339)"
// @Param   pageNo  query    int        false        "Page number (default 1)"
// @Param   pageSize query   int        false        "Page size (default 10)"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} CR "We need ID!!"
// @Failure 404 {object} CR "Can not find ID"
// @Failure 401 {object} CR "Can not find ID"
// @Router /record/dns [get]
func (self *WebServer) getDnsRecord(c *gin.Context) {
	T := getTranslateFunc(c)
	ip, ipExist := c.GetQuery("ip")
	domain, domainExist := c.GetQuery("domain")
	date, dateExist := c.GetQuery("date")

	pageNo, pageNoErr := ginutils.GetQueryInt(c, "pageNo")
	if pageNoErr != nil || pageNo <= 0 {
		pageNo = 1
	}
	pageSize, pageSizeErr := ginutils.GetQueryInt(c, "pageSize")
	if pageSizeErr != nil {
		pageSize = 10
	}

	session := self.orm.NewSession()
	defer session.Close()

	role := c.GetInt("role")
	id := c.GetInt64("id")
	switch role {
	case roleAdmin, roleSuper:
		session = session.In("uid", 0, id)
	default:
		session = session.Where(`uid=?`, id)
	}

	if domainExist {
		session = session.And(`domain like ?`, "%"+domain+"%")
	}
	if ipExist {
		session = session.And(`ip like ?`, "%"+ip+"%")
	}
	if dateExist {
		t, _ := time.Parse(time.RFC3339, strings.Trim(date, `"`))
		driver := self.orm.DriverName()
		if driver == "sqlite3" || driver == "sqlite" { //sqlite not support timezone
			t = t.Local()
		}
		session = session.And(`ctime > ?`, t)
		// fmt.Println("QUERYDATE=[", date, "] = ", t)
	}

	var items []models.TblDns
	count, err := session.Desc("id").Limit(pageSize, (pageNo-1)*pageSize).FindAndCount(&items)
	if err != nil {
		logrus.Errorf("[webui.go::getDnsRecord] orm.FindAndCount: %v", err)
		self.resp(c, 502, &CR{
			Message: T("Failed"),
			Code:    CodeServerInternal,
		})
		return
	}

	var resp DnsRecordResp
	resp.TotalCount = int(count)
	resp.PageSize = pageSize
	resp.PageNo = pageNo
	resp.TotalPage = (resp.TotalCount + (pageSize - 1)) / pageSize
	resp.Data = make([]models.DnsRecord, len(items))
	for i := 0; i < len(items); i++ {
		rcd := &resp.Data[i]
		item := &items[i]
		rcd.Id = item.Id
		rcd.Domain = item.Domain
		rcd.Ip = item.Ip
		rcd.Ctime = item.Ctime
	}

	self.resp(c, 200, &CR{
		Message: T("OK"),
		Data:    &resp,
	})
}

// @Summary delDnsRecord
// @Description del Dns Record by query ids
// @Accept  json
// @Produce  json
// @Param   ids     body    []int     true        "Record IDs to delete"
// @Success 200 {string} string	"ok"
// @Failure 400 {object} CR "We need ID!!"
// @Router /record/dns [delete]
func (self *WebServer) delDnsRecord(c *gin.Context) {
	T := getTranslateFunc(c)
	var req DeleteRecordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf("[webui.go::delDnsRecord] orm.Delete: %v", err)
		self.resp(c, 400, &CR{
			Message: T("invalid Param"),
			Code:    CodeServerInternal,
			Error:   err,
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	role := c.GetInt("role")
	id := c.GetInt64("id")

	switch role {
	case roleAdmin, roleSuper:
		if len(req.Ids) == 0 {
			_, err := session.In(`uid`, id, 0).Delete(&models.TblDns{})
			if err != nil {
				logrus.Errorf("[webui.go::delDnsRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		} else {
			params := make([]interface{}, len(req.Ids))
			for i := 0; i < len(req.Ids); i++ {
				params[i] = req.Ids[i]
			}
			_, err := session.In(`uid`, id, 0).In("id", params...).Delete(&models.TblDns{})
			if err != nil {
				logrus.Errorf("[webui.go::delDnsRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		}
	default:
		if len(req.Ids) == 0 {
			_, err := session.Where(`uid=?`, id).Delete(&models.TblDns{})
			if err != nil {
				logrus.Errorf("[webui.go::delDnsRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		} else {
			params := make([]interface{}, len(req.Ids))
			for i := 0; i < len(req.Ids); i++ {
				params[i] = req.Ids[i]
			}
			_, err := session.Where(`uid=?`, id).In("id", params...).Delete(&models.TblDns{})
			if err != nil {
				logrus.Errorf("[webui.go::delDnsRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		}
	}
}

func (self *WebServer) getHttpRecord(c *gin.Context) {
	T := getTranslateFunc(c)
	ip, ipExist := c.GetQuery("ip")
	domain, domainExist := c.GetQuery("domain")
	date, dateExist := c.GetQuery("date")

	pageNo, pageNoErr := ginutils.GetQueryInt(c, "pageNo")
	if pageNoErr != nil || pageNo <= 0 {
		pageNo = 1
	}
	pageSize, pageSizeErr := ginutils.GetQueryInt(c, "pageSize")
	if pageSizeErr != nil || pageSize <= 0 {
		pageSize = 10
	}

	ctype, ctypeExist := c.GetQuery("ctype")
	data, dataExist := c.GetQuery("data")
	method, methodExist := c.GetQuery("method")

	session := self.orm.NewSession()
	defer session.Close()

	role := c.GetInt("role")
	id := c.GetInt64("id")
	switch role {
	case roleAdmin, roleSuper:
		session = session.Where(`id>0`)
	default:
		session = session.Where(`uid=?`, id)
	}

	if domainExist {
		session = session.And(`domain like ?`, "%"+domain+"%")
	}
	if ipExist {
		session = session.And(`ip like ?`, "%"+ip+"%")
	}
	if dateExist {
		t, _ := time.Parse(time.RFC3339, strings.Trim(date, `"`))
		driver := self.orm.DriverName()
		if driver == "sqlite3" || driver == "sqlite" { //sqlite不支持时区
			t = t.Local()
		}
		session = session.And(`ctime > ?`, t)
	}
	if ctypeExist {
		session = session.And(`ctype like ?`, "%"+ctype+"%")
	}
	if dataExist {
		session = session.And(`data like ?`, "%"+data+"%")
	}
	if methodExist {
		session = session.And(`method = ?`, method)
	}

	var items []models.TblHttp
	count, err := session.Desc("id").Limit(pageSize, (pageNo-1)*pageSize).FindAndCount(&items)
	if err != nil {
		logrus.Errorf("[webui.go::getHttpRecord] orm.FindAndCount: %v", err)
		self.resp(c, 502, &CR{
			Code:    CodeServerInternal,
			Message: T("Failed"),
		})
		return
	}

	var resp HttpRecordResp
	resp.TotalCount = int(count)
	resp.PageSize = pageSize
	resp.PageNo = pageNo
	resp.TotalPage = (resp.TotalCount + (pageSize - 1)) / pageSize
	resp.Data = make([]models.HttpRecord, len(items))

	for i := 0; i < len(items); i++ {
		rcd := &resp.Data[i]
		item := &items[i]
		rcd.Id = item.Id
		rcd.Path = item.Path
		rcd.Ip = item.Ip
		rcd.Ctime = item.Ctime
		rcd.Ctype = item.Ctype
		rcd.Data = item.Data
		rcd.Method = item.Method
		rcd.Ua = item.Ua
	}
	self.resp(c, 200, &CR{
		Message: T("OK"),
		Data:    &resp,
	})
}

func (self *WebServer) delHttpRecord(c *gin.Context) {
	T := getTranslateFunc(c)
	var req DeleteRecordRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		logrus.Errorf("[webui.go::delHttpRecord] orm.Delete: %v", err)
		self.resp(c, 400, &CR{
			Message: T(`invalid Param`),
			Code:    CodeServerInternal,
			Error:   err,
		})
		return
	}

	session := self.orm.NewSession()
	defer session.Close()

	role := c.GetInt("role")
	id := c.GetInt64("id")

	switch role {
	case roleAdmin, roleSuper:
		if len(req.Ids) == 0 {
			_, err := session.In(`uid`, id, 0).Delete(&models.TblHttp{})
			if err != nil {
				//TODO:
				logrus.Errorf("[webui.go::delHttpRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		} else {
			params := make([]interface{}, len(req.Ids))
			for i := 0; i < len(req.Ids); i++ {
				params[i] = req.Ids[i]
			}
			_, err := session.In(`uid`, id, 0).In("id", params...).Delete(&models.TblHttp{})
			if err != nil {
				logrus.Errorf("[webui.go::delHttpRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		}
	default:
		if len(req.Ids) == 0 {
			_, err := session.Where(`uid=?`, id).Delete(&models.TblHttp{})
			if err != nil {
				//TODO:
				logrus.Errorf("[webui.go::delHttpRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: fmt.Sprintf(T("Delete Record: %v"), err),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		} else {
			params := make([]interface{}, len(req.Ids))
			for i := 0; i < len(req.Ids); i++ {
				params[i] = req.Ids[i]
			}
			_, err := session.Where(`uid=?`, id).In("id", params...).Delete(&models.TblHttp{})
			if err != nil {
				logrus.Errorf("[webui.go::delHttpRecord] orm.Delete: %v", err)
				self.resp(c, 502, &CR{
					Message: T("Failed"),
					Code:    CodeServerInternal,
				})
				return
			}
			self.resp(c, 200, &CR{
				Message: T("OK"),
			})
			return
		}
	}
}

// POST
// TYPE=[CNAME/A/MX/TXT], HOST={}, RECORD={}, TTL={}
func (self *WebServer) getResolveRecord(c *gin.Context) {
	T := getTranslateFunc(c)
	pageNo, pageNoErr := ginutils.GetQueryInt(c, "pageNo")
	if pageNoErr != nil || pageNo <= 0 {
		pageNo = 1
	}
	pageSize, pageSizeErr := ginutils.GetQueryInt(c, "pageSize")
	if pageSizeErr != nil || pageSize <= 0 {
		pageSize = 10
	}
	field, fieldExist := c.GetQuery("sortField")
	order, _ := c.GetQuery("sortOrder")

	host, hostExist := c.GetQuery("keyword")
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()
	session = session.Where(`id>0`)
	if hostExist {
		session = session.And(`host like ?`, "%"+host+"%")
	}
	if fieldExist {
		if field == "timestamp" {
			field = "utime"
		}

		if order != "ascend" {
			session = session.Desc(field)
		} else {
			session = session.Asc(field)
		}
	}

	var resolves []models.TblResolve
	count, err := session.Limit(pageSize, (pageNo-1)*pageSize).FindAndCount(&resolves)
	if err != nil {
		logrus.Errorf("[webui.go::delResolveRecord] orm.Delete: %v", err)
		self.resp(c, 502, &CR{
			Message: T("Failed"),
			Code:    CodeServerInternal,
		})
		return
	}
	var ret ResolveResult
	ret.Pagination.PageNo = pageNo
	ret.Pagination.PageSize = pageSize
	ret.Pagination.TotalCount = int(count)
	ret.Pagination.TotalPage = (int(count) + pageSize - 1) / pageSize
	for i := 0; i < len(resolves); i++ {
		r := &resolves[i]
		ret.Data = append(ret.Data, &ResolveItem{
			Id:         r.Id,
			Type:       r.Type,
			Host:       r.Host,
			Value:      r.Value,
			Ttl:        r.Ttl,
			Utimestamp: r.Utime.Unix(),
		})
	}
	c.JSON(200, &ret)
}

func (self *WebServer) updateResolveCache(host, rType string, all []models.TblResolve) {
	if rType != "" {
		var rr []*Resolve
		for i := 0; i < len(all); i++ {
			rcd := &all[i]
			r := &Resolve{
				Type:  rcd.Type,
				Name:  rcd.Host,
				Value: rcd.Value,
				Ttl:   rcd.Ttl,
			}
			rr = append(rr, r)
		}
		key := fmt.Sprintf("%v#%v", host, rType)
		self.store.Delete(key)

		if len(rr) > 0 {
			self.store.Set(key, rr, cache.NoExpiration)
		}
		return
	}

	var rrMap = make(map[string][]*Resolve)
	for i := 0; i < len(all); i++ {
		rcd := &all[i]
		r := &Resolve{
			Type:  rcd.Type,
			Name:  rcd.Host,
			Value: rcd.Value,
			Ttl:   rcd.Ttl,
		}
		if rr, exist := rrMap[rcd.Type]; exist {
			rrMap[rcd.Type] = append(rr, r)
		} else {
			rrMap[rcd.Type] = []*Resolve{r}
		}
	}
	// clear old cache
	for _, t := range []string{"A", "TXT", "CNAME", "SRV", "NS", "MX"} {
		key := fmt.Sprintf("%v#%v", host, t)
		self.store.Delete(key)
	}

	for k, v := range rrMap {
		key := fmt.Sprintf("%v#%v", host, k)
		self.store.Set(key, v, cache.NoExpiration)
	}
}

func (self *WebServer) setResolveRecord(c *gin.Context) {
	T := getTranslateFunc(c)

	self.uiMutex.Lock()
	defer self.uiMutex.Unlock()

	var rcd models.Resolve
	err := c.ShouldBindJSON(&rcd)
	if err != nil {
		logrus.Errorf("[webui.go::setResolveRecord] orm.Delete: %v", err)
		self.resp(c, 400, &CR{
			Message: T("Failed"),
			Code:    CodeServerInternal,
		})
		return
	}
	session := self.orm.NewSession()
	defer session.Close()

	// verify input
	if rcd.Host == "" || len(rcd.Host) > 128 ||
		rcd.Value == "" || len(rcd.Value) > 128 {
		self.resp(c, 400, &CR{
			Message: T("Bad Resolve Data"),
			Code:    CodeServerInternal,
		})
		return
	}

	// branch 1: create new
	if rcd.Id == 0 {
		// check confilct
		var rcds []models.TblResolve
		err = session.Where(`host = ?`, rcd.Host).Find(&rcds)
		if err != nil {
			logrus.Errorf("[webui.go::setResolveRecord] orm.Find: %v", err)
			self.resp(c, 502, &CR{
				Message: T("Failed"),
				Code:    CodeServerInternal,
			})
			return
		}

		// check conflict by type
		conflict, groups := models.Resolves(rcds).GetTypeConflict(&rcd)
		if conflict != nil {
			logrus.Errorf("[webui.go::setResolveRecord] record(%v) conflict with (%v)",
				rcd.Type, conflict.Type)
			self.resp(c, 400, &CR{
				Message: fmt.Sprintf(T("Type:(%v) conflict with %v"),
					rcd.Type, conflict.Type),
				Code: CodeServerInternal,
			})
			return
		} //1.24  3.30

		// check conflict by value, 不允许插入相同值
		if len(groups) >= 1 {
			sort.Sort(models.Resolves(groups))
			conflictV := groups.GetValueConflict(&rcd)
			if conflictV != nil {
				self.resp(c, 400, &CR{
					Message: fmt.Sprintf(T("Value(%v) already exists"), rcd.Value),
					Code:    CodeServerInternal,
				})
				return
			}
		}

		// insert into orm
		_, err = session.InsertOne(&models.TblResolve{
			Type:  rcd.Type,
			Host:  rcd.Host,
			Value: rcd.Value,
			Ttl:   rcd.Ttl,
		})
		if err != nil {
			logrus.Errorf("[webui.go::setResolveRecord] orm.InsertOne: %v", err)
			self.resp(c, 502, &CR{
				Message: fmt.Sprintf(T("Insert record: %v"), err),
				Code:    CodeServerInternal,
			})
			return
		}
		// update cache by type
		groups = append(groups, models.TblResolve{
			Type:  rcd.Type,
			Host:  rcd.Host,
			Value: rcd.Value,
			Ttl:   rcd.Ttl,
		})
		self.updateResolveCache(rcd.Host, rcd.Type, groups)

		c.JSON(200, CR{
			Message: T("OK"),
		})
		return
	}

	// branch 2: modify exist
	// check confilct
	var oldRcd models.TblResolve
	exist, err := session.ID(rcd.Id).Get(&oldRcd)
	if err != nil {
		logrus.Errorf("[webui.go::setResolveRecord] orm.Get: %v", err)
		self.resp(c, 502, &CR{
			Message: fmt.Sprintf(T("Query record: %v"), err),
			Code:    CodeServerInternal,
		})
		return
	} else if !exist {
		self.resp(c, 400, &CR{
			Message: fmt.Sprintf(T("Record not exist"), err),
			Code:    CodeServerInternal,
		})
		return
	} else if oldRcd.Host != rcd.Host || oldRcd.Type != rcd.Type {
		self.resp(c, 400, &CR{
			Message: fmt.Sprintf(T("Can't Change Host/Type"), err),
			Code:    CodeServerInternal,
		})
		return
	}

	var rcds []models.TblResolve
	err = session.Where(`host = ?`, rcd.Host).
		And(`type = ?`, rcd.Type).
		And(`id != ?`, oldRcd.Id).
		Find(&rcds)
	if err != nil {
		logrus.Errorf("[webui.go::setResolveRecord] orm.Find: %v", err)
		self.resp(c, 502, &CR{
			Message: fmt.Sprintf(T("Query Record: %v"), err),
			Code:    CodeServerInternal,
		})
		return
	} else if len(rcds) > 0 { // not exist groups
		// 1. conflict check by type
		// ... skip

		// 2. conflict check by value, distinct, 不允许将记录修改为已有值
		sort.Sort(models.Resolves(rcds))

		conflictV := models.Resolves(rcds).GetValueConflict(&rcd)
		if conflictV != nil {
			self.resp(c, 400, &CR{
				Message: fmt.Sprintf(T("Value(%v) already exists"), rcd.Value),
				Code:    CodeServerInternal,
			})
			return
		}
	}

	oldRcd.Ttl = rcd.Ttl
	oldRcd.Value = rcd.Value
	_, err = session.ID(oldRcd.Id).Cols("ttl", "value").Update(&oldRcd)
	if err != nil {
		logrus.Errorf("[webui.go::setResolveRecord] orm.InsertOne: %v", err)
		self.resp(c, 502, &CR{
			Message: fmt.Sprintf(T("Update record: %v"), err),
			Code:    CodeServerInternal,
		})
		return
	}

	// update cache by type
	rcds = append(rcds, models.TblResolve{
		Type:  rcd.Type,
		Host:  rcd.Host,
		Value: rcd.Value,
		Ttl:   rcd.Ttl,
	})

	// update cache
	self.updateResolveCache(rcd.Host, rcd.Type, rcds)

	c.JSON(200, CR{
		Message: T("OK"),
	})
}

// DELETE
func (self *WebServer) delResolveRecord(c *gin.Context) {
	T := getTranslateFunc(c)
	id, err := ginutils.GetQueryInt64(c, "id")
	if err != nil {
		logrus.Errorf("[webui.go::delResolveRecord] not input id")
		self.resp(c, 502, &CR{
			Message: T("parameter id is required"),
			Code:    CodeServerInternal,
		})
		return
	}
	orm := self.orm
	session := orm.NewSession()
	defer session.Close()

	var rcd models.TblResolve
	exist, err := session.ID(id).Get(&rcd)
	if err != nil {
		logrus.Errorf("[webui.go::delResolveRecord] orm.Delete: %v", err)
		self.resp(c, 502, &CR{
			Message: fmt.Sprintf(T("delete: %v"), err),
			Code:    CodeServerInternal,
		})
		return
	} else if !exist {
		logrus.Errorf("[webui.go::delResolveRecord] orm.Delete: %v", err)
		self.resp(c, 400, &CR{
			Message: fmt.Sprintf(T("Delete resolve(%v), but not exist"), rcd.Id),
			Code:    CodeServerInternal,
		})
		return
	}

	_, err = session.ID(id).Delete(&models.TblResolve{})
	if err != nil {
		logrus.Errorf("[webui.go::delResolveRecord] orm.Delete: %v", err)
		self.resp(c, 502, &CR{
			Message: fmt.Sprintf(T("delete: %v"), err),
			Code:    CodeServerInternal,
		})
		return
	}

	// delete -> update cache
	var all []models.TblResolve
	session.Where(`host=?`, rcd.Host).Find(&all)
	self.updateResolveCache(rcd.Host, rcd.Type, all)

	c.JSON(200, CR{
		Message: T("OK"),
	})
}

func getTranslateFunc(c *gin.Context) func(id string) string {
	lang := c.GetHeader("Language")
	return func(id string) string {
		return translateByLang(lang, id)
	}
}
