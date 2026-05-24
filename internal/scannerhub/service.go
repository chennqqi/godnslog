package scannerhub

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/chennqqi/godnslog/internal/auth"
	"github.com/chennqqi/godnslog/internal/models"
	"xorm.io/xorm"
)

var (
	ErrScannerRunNotFound = errors.New("scanner run not found")
	ErrInvalidCase        = errors.New("case not found")
	ErrInvalidPayload     = errors.New("payload not found")
	ErrPayloadNotInCase   = errors.New("payload does not belong to case")
	ErrInvalidScanner     = errors.New("invalid scanner")
	ErrInvalidDelivery    = errors.New("invalid delivery method")
)

// Service provides scanner run management services
type Service struct {
	engine      *xorm.Engine
	authService *auth.Service
}

// NewService creates a new scanner hub service
func NewService(engine *xorm.Engine) *Service {
	return &Service{
		engine:      engine,
		authService: auth.NewService(engine),
	}
}

// CreateScannerRun creates a new scanner run
func (s *Service) CreateScannerRun(req *models.ScannerRunCreateRequest, userID, baseURL string) (*models.ScannerRun, error) {
	// Validate scanner
	if req.Scanner != models.ScannerNuclei {
		return nil, ErrInvalidScanner
	}

	// Validate delivery method
	if req.DeliveryMethod != models.DeliveryMethodNucleiJsonl && req.DeliveryMethod != models.DeliveryMethodNucleiVar {
		return nil, ErrInvalidDelivery
	}

	// Validate case exists
	var caseModel models.Case
	has, err := s.engine.ID(req.CaseID).Get(&caseModel)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrInvalidCase
	}

	// Validate payload exists
	var payload models.Payload
	has, err = s.engine.ID(req.PayloadID).Get(&payload)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrInvalidPayload
	}

	// Validate payload belongs to case
	if payload.CaseID != req.CaseID {
		return nil, ErrPayloadNotInCase
	}

	// Generate command and JSONL using Sprint H logic
	command, jsonl, err := s.generateNucleiCommandAndJsonl(req, &payload, baseURL)
	if err != nil {
		return nil, err
	}

	scannerRun := &models.ScannerRun{
		ID:             models.GenerateID(),
		CaseID:         req.CaseID,
		PayloadID:      req.PayloadID,
		Scanner:        req.Scanner,
		Target:         req.Target,
		Template:       req.Template,
		DeliveryMethod: req.DeliveryMethod,
		Command:        command,
		Jsonl:          jsonl,
		Status:         models.ScannerRunStatusCreated,
		CreatedBy:      userID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if _, err := s.engine.Insert(scannerRun); err != nil {
		return nil, err
	}

	// Create audit log for scanner_run.created
	userIDPtr := &userID
	resourceIDPtr := &scannerRun.ID
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "scanner_run.created",
		ResourceType: "scanner_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"target":   scannerRun.Target,
			"template": scannerRun.Template,
			"scanner":  scannerRun.Scanner,
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		// Return error to ensure audit logging is not silently failing
		return nil, fmt.Errorf("failed to create audit log: %w", err)
	}

	return scannerRun, nil
}

// GetScannerRunByID retrieves a scanner run by its ID
func (s *Service) GetScannerRunByID(id string) (*models.ScannerRun, error) {
	var scannerRun models.ScannerRun
	has, err := s.engine.ID(id).Get(&scannerRun)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrScannerRunNotFound
	}
	return &scannerRun, nil
}

// GetScannerRunDetail retrieves a scanner run with derived fields
func (s *Service) GetScannerRunDetail(id, baseURL string) (*models.ScannerRunDetail, error) {
	scannerRun, err := s.GetScannerRunByID(id)
	if err != nil {
		return nil, err
	}

	// Get interaction count for this payload
	interactionCount, err := s.engine.Where("payload_id = ?", scannerRun.PayloadID).Count(&models.Interaction{})
	if err != nil {
		return nil, err
	}

	// Get last interaction timestamp
	var lastInteraction models.Interaction
	has, err := s.engine.Where("payload_id = ?", scannerRun.PayloadID).OrderBy("created_at DESC").Limit(1).Get(&lastInteraction)
	var lastInteractionAt *time.Time
	if err == nil && has {
		lastInteractionAt = &lastInteraction.CreatedAt
	}

	// Generate URLs
	interactionsURL := fmt.Sprintf("%s/api/v2/interactions?payload_id=%s", baseURL, scannerRun.PayloadID)
	evidenceURL := fmt.Sprintf("%s/dashboard/evidence?payload_id=%s", baseURL, scannerRun.PayloadID)

	// Evidence count calculation
	// NOTE: Sprint I limitation - evidence table not yet implemented
	// Future Sprint I+ will implement proper evidence table and count query
	// Current implementation: evidence_count = 0 (placeholder for future evidence table)
	evidenceCount := 0
	// TODO: Implement proper evidence table and count query for Sprint I+
	// This will require:
	// 1. Create models.Evidence table
	// 2. Add evidence generation logic in interaction service
	// 3. Query evidence count by payload_id here
	// For now, evidence_count is 0 to avoid false assumptions

	detail := &models.ScannerRunDetail{
		ScannerRun:        *scannerRun,
		InteractionCount:  int(interactionCount),
		LastInteractionAt: lastInteractionAt,
		EvidenceCount:     evidenceCount,
		InteractionsURL:   interactionsURL,
		EvidenceURL:       evidenceURL,
	}

	return detail, nil
}

// ListScannerRuns retrieves scanner runs with filtering
func (s *Service) ListScannerRuns(caseID, payloadID, scanner, status string, page, pageSize int) (*models.ScannerRunListResponse, error) {
	var scannerRuns []models.ScannerRun
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}
	if payloadID != "" {
		session = session.Where("payload_id = ?", payloadID)
	}
	if scanner != "" {
		session = session.Where("scanner = ?", scanner)
	}
	if status != "" {
		session = session.Where("status = ?", status)
	}

	total, err := session.Count(&models.ScannerRun{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := session.Desc("created_at").Limit(pageSize, offset).Find(&scannerRuns); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.ScannerRunListResponse{
		Items:      scannerRuns,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateScannerRunStatus updates the status of a scanner run
func (s *Service) UpdateScannerRunStatus(id string, req *models.ScannerRunUpdateStatusRequest, userID string) error {
	// Validate status transition
	existingRun, err := s.GetScannerRunByID(id)
	if err != nil {
		return err
	}

	if !isValidScannerRunStatusTransition(existingRun.Status, req.Status) {
		return errors.New("invalid status transition")
	}

	scannerRun := &models.ScannerRun{
		Status:    req.Status,
		UpdatedAt: time.Now(),
	}

	_, err = s.engine.ID(id).Cols("status", "updated_at").Update(scannerRun)
	if err != nil {
		return err
	}

	// Create audit log for scanner_run.status_updated
	userIDPtr := &userID
	resourceIDPtr := &id
	auditLog := &models.AuditLog{
		ID:           generateID(),
		UserID:       userIDPtr,
		Action:       "scanner_run.status_updated",
		ResourceType: "scanner_run",
		ResourceID:   resourceIDPtr,
		Details: models.AuditDetails{
			"from_status": existingRun.Status,
			"to_status":   req.Status,
		},
		Timestamp: time.Now(),
	}
	if err := s.authService.CreateAuditLog(auditLog); err != nil {
		// Return error to ensure audit logging is not silently failing
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// isValidScannerRunStatusTransition validates if a scanner run status transition is allowed
func isValidScannerRunStatusTransition(from, to string) bool {
	validTransitions := map[string][]string{
		models.ScannerRunStatusCreated:     {models.ScannerRunStatusDistributed},
		models.ScannerRunStatusDistributed: {models.ScannerRunStatusObserved},
		models.ScannerRunStatusObserved:    {models.ScannerRunStatusEvidenced},
		models.ScannerRunStatusEvidenced:   {},
	}

	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}

	for _, status := range allowed {
		if status == to {
			return true
		}
	}

	return false
}

// generateID generates a unique ID using base32 encoding
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}

// generateNucleiCommandAndJsonl generates Nuclei command and JSONL record
// Uses Sprint H logic from frontend-next/src/lib/scanner-hub.ts
func (s *Service) generateNucleiCommandAndJsonl(req *models.ScannerRunCreateRequest, payload *models.Payload, baseURL string) (string, string, error) {
	// Generate Nuclei command with shell quoting
	target := shellQuote(req.Target)
	payloadVar := shellQuote(fmt.Sprintf("godnslog_payload=%s", payload.TemplateRendered))

	command := fmt.Sprintf("nuclei -u %s -t godnslog-%s.yaml -var %s",
		target, req.Template, payloadVar)

	// Generate JSONL record
	jsonlRecord := map[string]interface{}{
		"scanner":          req.Scanner,
		"case_id":          req.CaseID,
		"payload_id":       req.PayloadID,
		"token":            payload.Token,
		"target":           req.Target,
		"template":         req.Template,
		"rendered_payload": payload.TemplateRendered,
		"interactions_url": fmt.Sprintf("%s/api/v2/interactions?payload_id=%s", baseURL, req.PayloadID),
		"evidence_url":     fmt.Sprintf("%s/dashboard/evidence?payload_id=%s", baseURL, req.PayloadID),
		"created_at":       time.Now().Format(time.RFC3339),
	}

	// Convert to single-line JSON
	jsonlBytes, err := jsonEncode(jsonlRecord)
	if err != nil {
		return "", "", err
	}

	return command, string(jsonlBytes), nil
}

// shellQuote quotes a string for shell use
func shellQuote(value string) string {
	return fmt.Sprintf("'%s'", strings.ReplaceAll(value, "'", "'\\''"))
}

// jsonEncode encodes a value to JSON without newlines
func jsonEncode(v interface{}) ([]byte, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	// Remove newlines to ensure single-line JSONL
	return []byte(strings.ReplaceAll(string(data), "\n", "")), nil
}
