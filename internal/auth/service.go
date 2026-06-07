package auth

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/hex"
	"errors"
	"time"

	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

var (
	ErrAPIKeyNotFound = errors.New("api key not found")
	ErrAPIKeyRevoked  = errors.New("api key is revoked")
	ErrAPIKeyExpired  = errors.New("api key is expired")
	ErrInvalidScope   = errors.New("invalid scope")
)

// Service provides authentication and authorization services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new auth service
func NewService(engine *xorm.Engine) *Service {
	// Ensure audit logs table exists with correct schema
	SyncSchema(engine)
	return &Service{engine: engine}
}

// ValidScopes defines all valid API key scopes
var ValidScopes = map[string]bool{
	// Human/General scopes
	"case:read":         true,
	"case:write":        true,
	"payload:read":      true,
	"payload:write":     true,
	"interaction:read":  true,
	"interaction:write": true,
	"evidence:read":     true,
	"evidence:write":    true,
	"admin:all":         true,
	// Agent scopes (Sprint K naming convention)
	"agent:create_probe":       true,
	"agent:wait_interaction":   true,
	"agent:read_interactions":  true,
	"agent:summarize_evidence": true,
	"agent:export_report":      true,
	"agent:read_runs":          true,
	"agent:revoke_token":       true,
	"agent:delete_payload":     true,
	"agent:modify_config":      true,
}

// generateAPIKey generates a new API key
func generateAPIKey() (string, string, error) {
	// Generate 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// Encode to hex
	key := hex.EncodeToString(bytes)

	// Return key and prefix (first 8 characters)
	prefix := key[:8]
	return key, prefix, nil
}

// CreateAPIKey creates a new API key
func (s *Service) CreateAPIKey(req *models.APIKeyCreateRequest, userID string) (*models.APIKey, error) {
	// Validate scopes
	for _, scope := range req.Scopes {
		if !ValidScopes[scope] {
			return nil, ErrInvalidScope
		}
	}

	// Validate agent scopes if this is an agent key
	if req.IsAgent {
		if !models.ValidateAgentScopes(req.Scopes) {
			return nil, errors.New("invalid agent scope")
		}

		// Force expiration time for agent keys (default 24 hours if not set)
		if req.ExpiresAt == nil {
			defaultExpires := time.Now().Add(24 * time.Hour)
			req.ExpiresAt = &defaultExpires
		}

		// Force risk tolerance to medium or lower if not set
		if req.RiskTolerance == "" {
			req.RiskTolerance = "medium"
		}
	}

	key, prefix, err := generateAPIKey()
	if err != nil {
		return nil, err
	}

	apiKey := &models.APIKey{
		ID:            generateID(),
		Key:           key,
		KeyPrefix:     prefix,
		Name:          req.Name,
		Scopes:        models.Scopes(req.Scopes),
		WorkspaceID:   req.WorkspaceID,
		RiskTolerance: req.RiskTolerance,
		IsAgent:       req.IsAgent,
		ExpiresAt:     req.ExpiresAt,
		CreatedBy:     userID,
	}

	if _, err := s.engine.Insert(apiKey); err != nil {
		return nil, err
	}

	return apiKey, nil
}

// GetAPIKeyByPrefix retrieves an API key by its prefix
func (s *Service) GetAPIKeyByPrefix(prefix string) (*models.APIKey, error) {
	var apiKey models.APIKey
	has, err := s.engine.Where("key_prefix = ? AND is_revoked = ?", prefix, false).Get(&apiKey)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrAPIKeyNotFound
	}
	return &apiKey, nil
}

// GetAPIKeyByID retrieves an API key by its ID
func (s *Service) GetAPIKeyByID(id string) (*models.APIKey, error) {
	var apiKey models.APIKey
	has, err := s.engine.ID(id).Get(&apiKey)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrAPIKeyNotFound
	}
	return &apiKey, nil
}

// ListAPIKeys retrieves API keys for a user
func (s *Service) ListAPIKeys(userID string, page, pageSize int) (*models.APIKeyListResponse, error) {
	var apiKeys []models.APIKey
	offset := (page - 1) * pageSize

	total, err := s.engine.Where("created_by = ?", userID).Count(&models.APIKey{})
	if err != nil {
		return nil, err
	}

	if err := s.engine.Where("created_by = ?", userID).
		Desc("created_at").
		Limit(pageSize, offset).
		Find(&apiKeys); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.APIKeyListResponse{
		Items:      apiKeys,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// RevokeAPIKey revokes an API key
func (s *Service) RevokeAPIKey(id string) error {
	now := time.Now()
	_, err := s.engine.ID(id).Cols("is_revoked", "revoked_at").Update(&models.APIKey{
		IsRevoked: true,
		RevokedAt: &now,
	})
	return err
}

// UpdateLastUsed updates the last used timestamp for an API key by full key
func (s *Service) UpdateLastUsed(fullKey string) error {
	// Extract prefix from full key
	if len(fullKey) < 8 {
		return errors.New("invalid api key format")
	}
	prefix := fullKey[:8]

	now := time.Now()
	_, err := s.engine.Where("key_prefix = ? AND key = ?", prefix, fullKey).
		Cols("last_used_at").
		Update(&models.APIKey{LastUsedAt: &now})
	return err
}

// ValidateAPIKey validates an API key by full key and returns it if valid
func (s *Service) ValidateAPIKey(fullKey string) (*models.APIKey, error) {
	// Extract prefix from full key
	if len(fullKey) < 8 {
		return nil, errors.New("invalid api key format")
	}
	prefix := fullKey[:8]

	// First find by prefix
	apiKey, err := s.GetAPIKeyByPrefix(prefix)
	if err != nil {
		return nil, err
	}

	// Then verify the full key matches
	if apiKey.Key != fullKey {
		return nil, ErrAPIKeyNotFound
	}

	if !apiKey.IsValid() {
		if apiKey.IsRevoked {
			return nil, ErrAPIKeyRevoked
		}
		return nil, ErrAPIKeyExpired
	}

	return apiKey, nil
}

// CreateAuditLog creates an audit log entry
func (s *Service) CreateAuditLog(log *models.AuditLog) error {
	if _, err := s.engine.Insert(log); err != nil {
		return err
	}
	return nil
}

// ListAuditLogs retrieves audit logs with filtering
func (s *Service) ListAuditLogs(userID, action, resourceType, resourceID string, startTime, endTime *time.Time, page, pageSize int) (*models.AuditLogListResponse, error) {
	var logs []models.AuditLog
	session := s.engine.NewSession()
	defer session.Close()

	if userID != "" {
		session = session.Where("user_id = ?", userID)
	}
	if action != "" {
		session = session.Where("action = ?", action)
	}
	if resourceType != "" {
		session = session.Where("resource_type = ?", resourceType)
	}
	if resourceID != "" {
		session = session.Where("resource_id = ?", resourceID)
	}
	if startTime != nil {
		session = session.Where("timestamp >= ?", startTime)
	}
	if endTime != nil {
		session = session.Where("timestamp <= ?", endTime)
	}

	total, err := session.Count(&models.AuditLog{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	// Re-apply WHERE conditions for Find to ensure they are not lost
	findSession := s.engine.NewSession()
	defer findSession.Close()
	if userID != "" {
		findSession = findSession.Where("user_id = ?", userID)
	}
	if action != "" {
		findSession = findSession.Where("action = ?", action)
	}
	if resourceType != "" {
		findSession = findSession.Where("resource_type = ?", resourceType)
	}
	if resourceID != "" {
		findSession = findSession.Where("resource_id = ?", resourceID)
	}
	if startTime != nil {
		findSession = findSession.Where("timestamp >= ?", startTime)
	}
	if endTime != nil {
		findSession = findSession.Where("timestamp <= ?", endTime)
	}
	if err := findSession.Desc("timestamp").Limit(pageSize, offset).Find(&logs); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.AuditLogListResponse{
		Items:      logs,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// generateID generates a unique ID using base32 encoding
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}
