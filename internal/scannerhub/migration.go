package scannerhub

import (
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// MigrateScannerHub runs database migration for scanner hub tables
func MigrateScannerHub(engine *xorm.Engine) error {
	return engine.Sync2(new(models.ScannerRun))
}
