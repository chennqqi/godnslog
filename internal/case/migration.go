package casemgmt

import (
	"xorm.io/xorm"
)

// MigrateCase runs database migration for case tables
func MigrateCase(engine *xorm.Engine) error {
	return engine.Sync(new(Case))
}
