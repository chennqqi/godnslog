package auth

import (
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// SyncSchema synchronizes the database schema with the current models
func SyncSchema(engine *xorm.Engine) error {
	return engine.Sync2(
		new(models.APIKey),
		new(models.AuditLog),
	)
}
