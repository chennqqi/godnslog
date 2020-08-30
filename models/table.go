package models

import (
	"time"
)

//==============================================================================
//database models
//==============================================================================

// tbl_user
type TblUser struct {
	Id      int64  `xorm:"pk autoincr"`
	Name    string `xorm:"varchar(64) notnull unique"`
	Email   string `xorm:"varchar(64) notnull unique"`
	Role    int    `xorm:"tinyint notnull default 0"`
	ShortId string `xorm:"varchar(32) notnull unique"`
	Token   string `xorm:"varchar(128) notnull unique"`
	Pass    string `xorm:"varchar(128) notnull"`

	//settings
	Lang            string   `xorm:"varchar(16) default('en-US') notnull"`
	Callback        string   `xorm:"text"`
	CallbackMessage string   `xorm:"text"`
	Rebind          []string `xorm:"json"`
	CleanInterval   int64    `xorm:"default 3600"`

	Atime time.Time `xorm:"datetime created"`
	Utime time.Time `xorm:"datetime updated"`
}

type TblDns struct {
	Id     int64     `xorm:"pk autoincr"`
	Uid    int64     `xorm:"notnull"` //TblUser.Id fk
	Domain string    `xorm:"varchar(255) notnull"`
	Var    string    `xorm:"varchar(255) index"`
	Ip     string    `xorm:"varchar(16) notnull"`
	Ctime  time.Time `xorm:"datetime"`
	Atime  time.Time `xorm:"datetime created"`
}

type TblHttp struct {
	Id     int64     `xorm:"pk autoincr"`
	Uid    int64     `xorm:"notnull"` //TblUser.Id fk
	Ip     string    `xorm:"varchar(16) notnull"`
	Var    string    `xorm:"varchar(255) index"`
	Path   string    `xorm:"text notnull"`
	Method string    `xorm:"varchar(16)"`
	Data   string    `xorm:"mediumtext"`
	Ctype  string    `xorm:"varchar(64)"`
	Ua     string    `xorm:"text"`
	Ctime  time.Time `xorm:"datetime"`
	Atime  time.Time `xorm:"datetime created"`
}
