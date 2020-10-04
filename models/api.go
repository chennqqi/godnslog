package models

import (
	"time"
)

//==============================================================================
// api models
//==============================================================================
const (
	CodeOK             = 0
	CodeBadPermission  = 1
	CodeBadData        = 2
	CodeNoAuth         = 3
	CodeNoPermission   = 4
	CodeServerInternal = 5
	CodeNoData         = 6
	CodeExpire         = 7

	RoleSuper  = 0
	RoleAdmin  = 1
	RoleNormal = 2

	GODNS_RFI_KEY   = "GODNSLOG"
	GODNS_RFI_VALUE = "694ef536e5d0245f203a1bcf8cbf3294" // md5sum($GODNS_RFI_KEY)
)

type LoginRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Islogin bool   `json:"isLogin"`
	Token   string `json:"token"`
	//TODO:
	Username string `json:"username"`
	RoleId   string `json:"roleId"`
	Lang     string `json:"lang"`
}

type Role struct {
	Id          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
}

type PermissionActionSet struct {
	Action       string `json:"action"`
	Description  string `json:"description"`
	DefaultCheck bool   `json:"defaultCheck"`
}

type Permission struct {
	RoleId          int                   `json:"roleId"`
	PermissionId    string                `json:"permissionId"`
	PermissionName  string                `json:"permissionName"`
	ActionEntitySet []PermissionActionSet `json:"ActionEntitySet"`
}

type UserInfo struct {
	Id       int64     `json:"id"`
	Name     string    `json:"username"`
	Email    string    `json:"email"`
	Avatar   string    `json:"avatar"`
	Language string    `json:"lang"`
	Role     Role      `json:"role"`
	Utime    time.Time `json:"utime"`
}

type UserRequest struct {
	Id       int64  `json:"id"`
	Name     string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     int    `json:"role"`
	Language string `json:"lang"`
}

type DnsRecordResp struct {
	Pagination
	Data []DnsRecord `json:"data"`
}

type HttpRecordResp struct {
	Pagination
	Data []HttpRecord `json:"data"`
}

type UserListResp struct {
	Pagination
	Data []UserInfo `json:"data"`
}

type AppSetting struct {
	Callback  string   `json:"callback"`
	CleanHour int64    `json:"cleanHour"`
	Rebind    []string `json:"rebind"`
}

type DeleteRecordRequest struct {
	Ids []int64 `json:"ids"`
}

type AppSecurity struct {
	Token    string `json:"token"`
	DnsAddr  string `json:"dns_addr"`
	HttpAddr string `json:"http_addr"`
}

type AppSecuritySet struct {
	Password string `json:"password"`
}

type DnsRecord struct {
	Id       int64     `json:"id,omitempty"`
	Uid      int64     `json:"-"`
	Callback string    `json:"-"`
	Var      string    `json:"-"`
	Domain   string    `json:"domain"`
	Ip       string    `json:"addr"`
	Ctime    time.Time `json:"ctime"`
}

type HttpRecord struct {
	Id       int64     `json:"id,omitempty"`
	Uid      int64     `json:"-"`
	Callback string    `json:"-"`
	Path     string    `json:"path"`
	Ip       string    `json:"addr"`
	Method   string    `json:"method"`
	Data     string    `json:"data"`
	Ctype    string    `json:"ctype"`
	Ua       string    `json:"ua"`
	Ctime    time.Time `json:"ctime"`
}

// commone response
type CR struct {
	Message   string      `json:"message"`
	Code      int         `json:"code"`
	Error     error       `json:"error,omitempty"`
	Timestamp int64       `json:"timestemp"`
	Result    interface{} `json:"result,omitempty"`
}

type Pagination struct {
	PageNo     int `json:"pageNo"`
	PageSize   int `json:"pageSize"`
	TotalCount int `json:"totalCount"`
	TotalPage  int `json:"totalPage"`
}

type Resolv struct {
	Host  string `json:"host"` //host record, eg. www
	Type  string `json:"Type"` //record type
	Value string `json:"Value"`
	Ttl   uint32 `json:"ttl"`
}
