package main

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
)

const (
	roleSuper = iota
	roleAdmin
	roleNormal
)

type Setting struct {
	IPv6Bind []string `json:"ipv6_bind"`
	IPv4Bind []string `json:"ipv4_bind"`
}

// tbl_user
// type User struct {
// 	Id            int64  `xorm:"pk autoincr"`
// 	Name          string `xorm:"varchar(64) notnull unique"`
// 	Role          int    `xorm:"tinyint notnull default 0"`
// 	Token         string `xorm:"varchar(128) notnull"`
// 	Callback      string `xorm:"text"`
// 	CleanInterval int64
// }

type DnsRecord struct {
	Id       int64     `json:"id,omitempty"`
	Uid      int64     `json:"-"`
	Callback string    `json:"-"`
	Domain   string    `json:"domain"`
	Ip       string    `json:"addr"`
	Ctime    time.Time `json:"ctime"`
}

type HttpRecord struct {
	Id       int64     `json:"id,omitempty"`
	Uid      int64     `json:"-"`
	Callback string    `json:"-"`
	Url      string    `json:"url"`
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
