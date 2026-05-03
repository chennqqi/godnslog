package payload

import (
	"xorm.io/xorm"
)

// MigratePayload runs database migration for payload tables
func MigratePayload(engine *xorm.Engine) error {
	return engine.Sync(new(Payload))
}
