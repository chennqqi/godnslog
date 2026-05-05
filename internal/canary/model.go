package canary

import (
	"github.com/chennqqi/godnslog/internal/models"
)

// Re-export types from models package for backward compatibility
type Canary = models.Canary
type CanaryHit = models.CanaryHit
type CanaryConfig = models.CanaryConfig
type CanaryType = models.CanaryType
type CanaryContext = models.CanaryContext
type CanaryListResponse = models.CanaryListResponse
type CanaryHitListResponse = models.CanaryHitListResponse

// Re-export constants
const (
	CanaryTypeDNS      = models.CanaryTypeDNS
	CanaryTypeHTTP     = models.CanaryTypeHTTP
	CanaryTypeDocument = models.CanaryTypeDocument
	CanaryTypeConfig   = models.CanaryTypeConfig
	CanaryTypeCI       = models.CanaryTypeCI
	CanaryTypeStorage  = models.CanaryTypeStorage
	CanaryTypeEmail    = models.CanaryTypeEmail
)
