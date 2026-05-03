package auth

import (
	"xorm.io/xorm"
)

// MigrateAuth runs database migration for auth related tables
func MigrateAuth(engine *xorm.Engine) error {
	return engine.Sync(
		new(APIKey),
		new(AuditLog),
	)
}
