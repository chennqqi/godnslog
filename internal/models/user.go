package models

import (
	"time"

	"github.com/chennqqi/godnslog/models"
)

// User represents a user in the system (migrated from TblUser)
// This is a wrapper around models.TblUser for compatibility
type User struct {
	ID              int64     `json:"id" xorm:"pk autoincr"`
	Name            string    `json:"name" xorm:"varchar(64) notnull unique"`
	Email           string    `json:"email" xorm:"varchar(64) notnull unique"`
	Role            int       `json:"role" xorm:"tinyint notnull default 0"` // 0=super, 1=admin, 2=normal, 3=guest
	ShortID         string    `json:"short_id" xorm:"varchar(32) notnull unique"` // For subdomain
	Token           string    `json:"-" xorm:"varchar(128) notnull unique"` // API Token
	Pass            string    `json:"-" xorm:"varchar(128) notnull"` // Password hash
	
	// Settings
	Lang            string   `json:"lang" xorm:"varchar(16) default('en-US') notnull"`
	Callback        string   `json:"callback" xorm:"text"`
	CallbackMessage string   `json:"callback_message" xorm:"text"`
	Rebind          []string `json:"rebind" xorm:"json"`
	CleanInterval   int64    `json:"clean_interval" xorm:"default 3600"`
	
	CreatedAt       time.Time `json:"created_at" xorm:"datetime created"`
	UpdatedAt       time.Time `json:"updated_at" xorm:"datetime updated"`
}

// TableName returns the table name for User model
func (User) TableName() string {
	return "tbl_user"
}

// ToTblUser converts User to models.TblUser
func (u *User) ToTblUser() *models.TblUser {
	return &models.TblUser{
		Id:              u.ID,
		Name:            u.Name,
		Email:           u.Email,
		Role:            u.Role,
		ShortId:         u.ShortID,
		Token:           u.Token,
		Pass:            u.Pass,
		Lang:            u.Lang,
		Callback:        u.Callback,
		CallbackMessage: u.CallbackMessage,
		Rebind:          u.Rebind,
		CleanInterval:   u.CleanInterval,
		Atime:           u.CreatedAt,
		Utime:           u.UpdatedAt,
	}
}

// FromTblUser converts models.TblUser to User
func FromTblUser(tbl *models.TblUser) *User {
	return &User{
		ID:              tbl.Id,
		Name:            tbl.Name,
		Email:           tbl.Email,
		Role:            tbl.Role,
		ShortID:         tbl.ShortId,
		Token:           tbl.Token,
		Pass:            tbl.Pass,
		Lang:            tbl.Lang,
		Callback:        tbl.Callback,
		CallbackMessage: tbl.CallbackMessage,
		Rebind:          tbl.Rebind,
		CleanInterval:   tbl.CleanInterval,
		CreatedAt:       tbl.Atime,
		UpdatedAt:       tbl.Utime,
	}
}

// Role constants
const (
	RoleSuper  = 0
	RoleAdmin  = 1
	RoleNormal = 2
	RoleGuest  = 3
)

// RoleNames maps role IDs to names
var RoleNames = map[int]string{
	RoleSuper:  "super",
	RoleAdmin:  "admin",
	RoleNormal: "normal",
	RoleGuest:  "guest",
}

// GetRoleName returns the role name for a role ID
func (u *User) GetRoleName() string {
	if name, ok := RoleNames[u.Role]; ok {
		return name
	}
	return "unknown"
}

// IsSuper checks if user is super admin
func (u *User) IsSuper() bool {
	return u.Role == RoleSuper
}

// IsAdmin checks if user is admin or super
func (u *User) IsAdmin() bool {
	return u.Role == RoleSuper || u.Role == RoleAdmin
}

// CanReadAll checks if user can read all records
func (u *User) CanReadAll() bool {
	return u.IsAdmin()
}
