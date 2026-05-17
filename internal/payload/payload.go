package payload

import (
	"github.com/chennqqi/godnslog/internal/models"
)

// Re-export unified Payload model from internal/models
// This package will be deprecated in favor of internal/models
type Payload = models.Payload
type Variables = models.Variables
type PayloadCreateRequest = models.PayloadCreateRequest
type PayloadUpdateRequest = models.PayloadUpdateRequest
type PayloadListResponse = models.PayloadListResponse
