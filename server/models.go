package server

import (
	"github.com/chennqqi/godnslog/models"
)

//==============================================================================
// api models
//==============================================================================
const (
	CodeOK             = models.CodeOK
	CodeBadPermission  = models.CodeBadPermission
	CodeBadData        = models.CodeBadData
	CodeNoAuth         = models.CodeNoAuth
	CodeNoPermission   = models.CodeNoPermission
	CodeServerInternal = models.CodeServerInternal
	CodeNoData         = models.CodeNoData
	CodeExpire         = models.CodeExpire
)

const (
	roleSuper  = models.RoleSuper
	roleAdmin  = models.RoleAdmin
	roleNormal = models.RoleNormal
)

type LoginRequest models.LoginRequest
type LoginResponse models.LoginResponse
type Role models.Role
type PermissionActionSet models.PermissionActionSet
type Permission models.Permission
type UserInfo models.UserInfo
type UserRequest models.UserRequest
type DnsRecordResp models.DnsRecordResp
type HttpRecordResp models.HttpRecordResp
type UserListResp models.UserListResp
type AppSetting models.AppSetting
type DeleteRecordRequest models.DeleteRecordRequest
type AppSecurity models.AppSecurity
type AppSecuritySet models.AppSecuritySet
type DnsRecord models.DnsRecord
type HttpRecord models.HttpRecord

// commone response
type CR models.CR
type Pagination models.Pagination
