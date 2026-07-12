package interaction

import (
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// MigrateInteraction runs database migration for interaction tables
func MigrateInteraction(engine *xorm.Engine) error {
	return engine.Sync(new(models.Interaction))
}
