package interaction

import (
	"xorm.io/xorm"
)

// MigrateInteraction runs database migration for interaction tables
func MigrateInteraction(engine *xorm.Engine) error {
	return engine.Sync(new(Interaction))
}
