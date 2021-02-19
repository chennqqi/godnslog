package models

import (
	"sort"
	"time"
)

var (
	resolveTypeConflictMap = map[string]map[string]bool{
		"CNAME": map[string]bool{
			"CNAME": true,
			"A":     true,
			"MX":    true,
			"TXT":   false,
		},
		"A": map[string]bool{
			"CNAME": true,
			"A":     false,
			"MX":    true,
			"TXT":   false,
		},
		"MX": map[string]bool{
			"CNAME": true,
			"A":     true,
			"MX":    true,
			"TXT":   false,
		},
		"TXT": map[string]bool{
			"CNAME": false,
			"A":     false,
			"MX":    false,
			"TXT":   true,
		},
	}
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

type TblResolve struct {
	Id    int64  `xorm:"pk autoincr"`
	Host  string `xorm:"index varchar(255) notnull"` //host record, eg. www
	Type  string `xorm:"varchar(16) notnull"`        //record type, eg. CNAME/A/MX/TXT/SRV/NS.
	Value string `xorm:"varchar(255) notnull"`
	Ttl   uint32
	Ctime time.Time `xorm:"created"`
	Utime time.Time `xorm:"updated"`
}

type Resolves []TblResolve

func (rs Resolves) Len() int           { return len(rs) }
func (rs Resolves) Less(i, j int) bool { return rs[i].Value < rs[j].Value }
func (rs Resolves) Swap(i, j int)      { rs[i], rs[j] = rs[j], rs[i] }

// CNAME,TXT,MX 同一个host只允许配置1个
func (rs Resolves) GetIndex(r *Resolve) int {
	for i := 0; i < len(rs); i++ {
		if rs[i].Id == r.Id {
			return i
		}
	}
	return -1
}

// CNAME,TXT,MX 同一个host只允许配置1个
func (rs Resolves) GetTypeConflict(r *Resolve) (*Resolve, Resolves) {
	var groups Resolves
	cm, ok := resolveTypeConflictMap[r.Type]
	if !ok {
		return nil, groups
	}
	for i := 0; i < len(rs); i++ {
		if cm[rs[i].Type] {
			return &Resolve{
				Id:         rs[i].Id,
				Type:       rs[i].Type,
				Value:      rs[i].Value,
				Ttl:        rs[i].Ttl,
				Utimestamp: rs[i].Utime.Unix(),
			}, groups
		}
		if rs[i].Type == r.Type {
			groups = append(groups, TblResolve{
				Type:  rs[i].Type,
				Value: rs[i].Value,
				Ttl:   rs[i].Ttl,
			})
		}
	}
	return nil, groups
}

// Value 不允许重复
func (rs Resolves) GetValueConflict(r *Resolve) *Resolve {
	idx := sort.Search(len(rs), func(i int) bool {
		return rs[i].Value >= r.Value
	})
	if idx < len(rs) && rs[idx].Value == r.Value {
		return &Resolve{
			Id:         rs[idx].Id,
			Type:       rs[idx].Type,
			Value:      rs[idx].Value,
			Ttl:        rs[idx].Ttl,
			Utimestamp: rs[idx].Utime.Unix(),
		}
	}

	return nil
}
