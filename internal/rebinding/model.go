package rebinding

import (
	"github.com/chennqqi/godnslog/internal/models"
)

// Re-export types from models package for backward compatibility
type RebindingRule = models.RebindingRule
type Stage = models.Stage
type Stages = models.Stages
type RebindingConfig = models.RebindingConfig
type RebindingSession = models.RebindingSession
type RebindingScenario = models.RebindingScenario
type RebindingRuleListResponse = models.RebindingRuleListResponse
type RebindingSessionListResponse = models.RebindingSessionListResponse

// Re-export constants
const (
	ScenarioBrowserRebinding   = models.ScenarioBrowserRebinding
	ScenarioCloudMetadata      = models.ScenarioCloudMetadata
	ScenarioInternalManagement = models.ScenarioInternalManagement
	ScenarioIoTDevice          = models.ScenarioIoTDevice
	ScenarioRouterExploit      = models.ScenarioRouterExploit
)
