package agentrun

import (
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

// MigrateAgentHub runs database migration for agent hub tables
func MigrateAgentHub(engine *xorm.Engine) error {
	return engine.Sync2(
		new(models.AgentRun),
		new(models.AgentOperation),
	)
}
