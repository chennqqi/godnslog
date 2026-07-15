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
	service := &Service{engine: engine}
	if engine != nil {
		service.authService = auth.NewService(engine)
	}
	return service
}

// ListAdapters returns the stable public Scanner Hub adapter catalog.
func (s *Service) ListAdapters() *models.ScannerAdapterListResponse {
	return &models.ScannerAdapterListResponse{Items: scannerAdapterCatalog()}
}

// CreateScannerRun creates a new scanner run
func (s *Service) CreateScannerRun(req *models.ScannerRunCreateRequest, userID, baseURL string) (*models.ScannerRun, error) {
	if err := validateScannerDelivery(req.Scanner, req.DeliveryMethod); err != nil {
		return nil, err
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

	command, jsonl, packageManifest, packageHash, err := s.generateScannerArtifacts(req, &payload, baseURL)
	if err != nil {
		return nil, err
	}

	scannerRun := &models.ScannerRun{
		ID:              models.GenerateID(),
		CaseID:          req.CaseID,
		PayloadID:       req.PayloadID,
		Scanner:         req.Scanner,
		Target:          req.Target,
		Template:        req.Template,
		DeliveryMethod:  req.DeliveryMethod,
		Command:         command,
		Jsonl:           jsonl,
		PackageManifest: packageManifest,
		PackageHash:     packageHash,
		Status:          models.ScannerRunStatusCreated,
		CreatedBy:       userID,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
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

	// Evidence count: count enriched interactions (those with exploit type or decoded data)
	evidenceCount64, err := s.engine.Where("payload_id = ? AND (exploit_type IS NOT NULL AND exploit_type != '' OR decoded_data IS NOT NULL AND decoded_data != '')", scannerRun.PayloadID).Count(&models.Interaction{})
	evidenceCount := 0
	if err == nil {
		evidenceCount = int(evidenceCount64)
	}

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

func scannerAdapterCatalog() []models.ScannerAdapter {
	return []models.ScannerAdapter{
		{
			ID:               models.ScannerNuclei,
			Name:             "Nuclei",
			Category:         "Script",
			Maturity:         "L3 Official Script",
			SupportedMethods: []string{models.DeliveryMethodNucleiJsonl, models.DeliveryMethodNucleiVar},
			DefaultMethod:    models.DeliveryMethodNucleiJsonl,
			Description:      "Automated vulnerability scanning with template variables and JSONL distribution packages.",
		},
		{
			ID:               models.ScannerBurp,
			Name:             "Burp Suite",
			Category:         "Native",
			Maturity:         "L4 Native Plugin Path",
			SupportedMethods: []string{models.DeliveryMethodBurpExtension},
			DefaultMethod:    models.DeliveryMethodBurpExtension,
			Description:      "Manual and semi-automated verification package for Burp Suite extension workflows.",
		},
		{
			ID:               models.ScannerYakit,
			Name:             "Yakit/Yak",
			Category:         "Script",
			Maturity:         "L3 Official Script",
			SupportedMethods: []string{models.DeliveryMethodYakitScript},
			DefaultMethod:    models.DeliveryMethodYakitScript,
			Description:      "Yak script package for hybrid manual and automated security validation.",
		},
		{
			ID:               models.ScannerZap,
			Name:             "ZAP",
			Category:         "Script",
			Maturity:         "L3 Official Script",
			SupportedMethods: []string{models.DeliveryMethodZapScript},
			DefaultMethod:    models.DeliveryMethodZapScript,
			Description:      "OWASP ZAP script package for OAST probe injection and polling.",
		},
		{
			ID:               models.ScannerXray,
			Name:             "xray",
			Category:         "Webhook",
			Maturity:         "L2 Webhook Bridge",
			SupportedMethods: []string{models.DeliveryMethodXrayWebhook},
			DefaultMethod:    models.DeliveryMethodXrayWebhook,
			Description:      "Webhook bridge package for mapping xray findings to GODNSLOG evidence.",
		},
		{
			ID:               models.ScannerRad,
			Name:             "rad",
			Category:         "Webhook",
			Maturity:         "L2 Webhook Bridge",
			SupportedMethods: []string{models.DeliveryMethodRadWebhook},
			DefaultMethod:    models.DeliveryMethodRadWebhook,
			Description:      "Webhook bridge package for rad crawler and scanner automation.",
		},
		{
			ID:               models.ScannerPostman,
			Name:             "Postman",
			Category:         "Environment",
			Maturity:         "L2 Environment Bridge",
			SupportedMethods: []string{models.DeliveryMethodPostmanEnv},
			DefaultMethod:    models.DeliveryMethodPostmanEnv,
			Description:      "Environment variable package for API testing with OAST verification.",
		},
		{
			ID:               models.ScannerApifox,
			Name:             "Apifox",
			Category:         "Environment",
			Maturity:         "L2 Environment Bridge",
			SupportedMethods: []string{models.DeliveryMethodApifoxEnv},
			DefaultMethod:    models.DeliveryMethodApifoxEnv,
			Description:      "Environment variable package for Apifox API testing workflows.",
		},
	}
}

func validateScannerDelivery(scanner, delivery string) error {
	knownScanner := false
	for _, adapter := range scannerAdapterCatalog() {
		if adapter.ID != scanner {
			continue
		}
		knownScanner = true
		for _, method := range adapter.SupportedMethods {
			if method == delivery {
				return nil
			}
		}
	}
	if !knownScanner {
		return ErrInvalidScanner
	}
	return ErrInvalidDelivery
}

// generateScannerArtifacts generates an operator-facing integration package and a structured JSONL record.
func (s *Service) generateScannerArtifacts(req *models.ScannerRunCreateRequest, payload *models.Payload, baseURL string) (string, string, models.ScannerPackageManifest, string, error) {
	interactionsURL := fmt.Sprintf("%s/api/v2/interactions?payload_id=%s", baseURL, req.PayloadID)
	evidenceURL := fmt.Sprintf("%s/dashboard/evidence?payload_id=%s", baseURL, req.PayloadID)
	command := generateScannerCommand(req, payload, interactionsURL, evidenceURL)

	jsonlRecord := map[string]interface{}{
		"scanner":          req.Scanner,
		"delivery_method":  req.DeliveryMethod,
		"case_id":          req.CaseID,
		"payload_id":       req.PayloadID,
		"token":            payload.Token,
		"target":           req.Target,
		"template":         req.Template,
		"rendered_payload": payload.TemplateRendered,
		"interactions_url": interactionsURL,
		"evidence_url":     evidenceURL,
		"created_at":       time.Now().Format(time.RFC3339),
	}

	jsonlBytes, err := jsonEncode(jsonlRecord)
	if err != nil {
		return "", "", models.ScannerPackageManifest{}, "", err
	}

	jsonl := string(jsonlBytes)
	manifest := buildScannerPackageManifest(req, interactionsURL, evidenceURL)
	packageHash, err := computeScannerPackageHash(req, command, jsonl, manifest)
	if err != nil {
		return "", "", models.ScannerPackageManifest{}, "", err
	}
	manifest.PackageHash = packageHash

	return command, jsonl, manifest, packageHash, nil
}

func buildScannerPackageManifest(req *models.ScannerRunCreateRequest, interactionsURL, evidenceURL string) models.ScannerPackageManifest {
	files := []models.ScannerPackageFile{
		{
			Name:        "README.md",
			Kind:        "instructions",
			Description: "Operator and Agent instructions for distributing this GODNSLOG scanner package.",
		},
		{
			Name:        "godnslog-package.jsonl",
			Kind:        "jsonl",
			Description: "Single-line GODNSLOG scanner package record with payload, target, and callback URLs.",
		},
	}

	switch req.DeliveryMethod {
	case models.DeliveryMethodNucleiJsonl:
		files = append(files, models.ScannerPackageFile{Name: "godnslog-nuclei.jsonl", Kind: "jsonl", Description: "Nuclei JSONL import/distribution record."})
	case models.DeliveryMethodNucleiVar:
		files = append(files, models.ScannerPackageFile{Name: fmt.Sprintf("godnslog-%s.yaml", req.Template), Kind: "nuclei-template", Description: "Nuclei template using the godnslog_payload variable."})
	case models.DeliveryMethodBurpExtension:
		files = append(files, models.ScannerPackageFile{Name: "burp-extension-config.json", Kind: "burp-extension-config", Description: "Burp Suite extension configuration inputs and polling URLs."})
	case models.DeliveryMethodYakitScript:
		files = append(files, models.ScannerPackageFile{Name: "godnslog-oast.yak", Kind: "yak-script", Description: "Yak script skeleton for OAST payload injection and interaction polling."})
	case models.DeliveryMethodZapScript:
		files = append(files, models.ScannerPackageFile{Name: "godnslog-oast.js", Kind: "zap-script", Description: "ZAP script skeleton for OAST payload injection."})
	case models.DeliveryMethodXrayWebhook, models.DeliveryMethodRadWebhook:
		files = append(files, models.ScannerPackageFile{Name: "webhook-bridge.json", Kind: "webhook-bridge", Description: "Webhook bridge mapping scanner findings to GODNSLOG evidence references."})
	case models.DeliveryMethodPostmanEnv:
		files = append(files, models.ScannerPackageFile{Name: "godnslog.postman_environment.json", Kind: "postman-environment", Description: "Postman environment variables for OAST validation."})
	case models.DeliveryMethodApifoxEnv:
		files = append(files, models.ScannerPackageFile{Name: "godnslog.apifox_environment.json", Kind: "apifox-environment", Description: "Apifox environment variables for OAST validation."})
	}

	return models.ScannerPackageManifest{
		SchemaVersion:   "scanner-package.v1",
		Scanner:         req.Scanner,
		DeliveryMethod:  req.DeliveryMethod,
		HashAlgorithm:   "sha256",
		Files:           files,
		InteractionsURL: interactionsURL,
		EvidenceURL:     evidenceURL,
		NextActions: []string{
			"Distribute the generated payload through the selected scanner adapter.",
			"Poll interactions_url until the expected callback is observed.",
			"Open evidence_url or call the Evidence API to produce the final proof chain.",
		},
	}
}

func computeScannerPackageHash(req *models.ScannerRunCreateRequest, command, jsonl string, manifest models.ScannerPackageManifest) (string, error) {
	manifest.PackageHash = ""
	return models.ComputeDeterministicHash(map[string]interface{}{
		"scanner":          req.Scanner,
		"delivery_method":  req.DeliveryMethod,
		"target":           req.Target,
		"template":         req.Template,
		"command":          command,
		"jsonl":            jsonlWithoutCreatedAt(jsonl),
		"package_manifest": manifest,
	})
}

func jsonlWithoutCreatedAt(jsonl string) map[string]interface{} {
	var record map[string]interface{}
	if err := json.Unmarshal([]byte(jsonl), &record); err != nil {
		return map[string]interface{}{"raw": jsonl}
	}
	delete(record, "created_at")
	return record
}

func generateScannerCommand(req *models.ScannerRunCreateRequest, payload *models.Payload, interactionsURL, evidenceURL string) string {
	target := shellQuote(req.Target)
	payloadVar := shellQuote(fmt.Sprintf("godnslog_payload=%s", payload.TemplateRendered))

	switch req.Scanner {
	case models.ScannerNuclei:
		return fmt.Sprintf("nuclei -u %s -t godnslog-%s.yaml -var %s",
			target, req.Template, payloadVar)
	case models.ScannerBurp:
		return strings.Join([]string{
			fmt.Sprintf("# Burp Suite Extension — OAST probe for %s", req.Target),
			fmt.Sprintf("# 1. Build: cd $GODNSLOG_HOME/examples/burp-suite && mvn package"),
			fmt.Sprintf("# 2. Load godnslog-burp-extension-*.jar into Burp Extender"),
			fmt.Sprintf("# 3. Configure extension with:"),
			fmt.Sprintf("#    GODNSLOG_URL=%s", interactionsURL[:strings.LastIndex(interactionsURL, "/interactions")]),
			fmt.Sprintf("#    GODNSLOG_API_KEY=<your-api-key>"),
			fmt.Sprintf("# 4. Right-click HTTP request → Extensions → GODNSLOG → Create OAST Probe"),
			"",
			fmt.Sprintf("# Alternatively, create payload via API:"),
			fmt.Sprintf("curl -s -X POST \"${GODNSLOG_URL}/api/v2/payloads\" \\"),
			fmt.Sprintf("  -H \"Content-Type: application/json\" \\"),
			fmt.Sprintf("  -H \"Authorization: Bearer ${GODNSLOG_API_KEY}\" \\"),
			fmt.Sprintf("  -d '{\"case_id\":\"%s\",\"template\":\"%s\",\"scenario\":\"Burp OAST probe for %s\"}'", req.CaseID, req.Template, req.Target),
			fmt.Sprintf("# Poll: curl -s \"${GODNSLOG_URL}/api/v2/interactions?payload_id=<id>\" | jq ."),
		}, "\n")
	case models.ScannerYakit:
		return strings.Join([]string{
			fmt.Sprintf("export GODNSLOG_PAYLOAD=%s", payload.TemplateRendered),
			fmt.Sprintf("export GODNSLOG_TARGET=%s", req.Target),
			fmt.Sprintf("export GODNSLOG_INTERACTIONS=%s", interactionsURL),
			fmt.Sprintf("export GODNSLOG_EVIDENCE=%s", evidenceURL),
			"yak run godnslog-oast.yak",
		}, "\n")
	case models.ScannerZap:
		return strings.Join([]string{
			fmt.Sprintf("# ZAP Script — OAST probe for %s", req.Target),
			fmt.Sprintf("export GODNSLOG_PAYLOAD=\"%s\"", payload.TemplateRendered),
			fmt.Sprintf("export GODNSLOG_TARGET=\"%s\"", req.Target),
			fmt.Sprintf("export GODNSLOG_INTERACTIONS=\"%s\"", interactionsURL),
			"",
			fmt.Sprintf("zap.sh -cmd -script godnslog-oast.js \\"),
			fmt.Sprintf("  -scriptVars 'godnslog_payload=$GODNSLOG_PAYLOAD,godnslog_target=$GODNSLOG_TARGET,godnslog_url=$GODNSLOG_INTERACTIONS' \\"),
			fmt.Sprintf("  -port 8080 -host 127.0.0.1"),
		}, "\n")
	case models.ScannerXray, models.ScannerRad:
		webhookURL := interactionsURL[:strings.LastIndex(interactionsURL, "/interactions")] + "/webhook/xray"
		return strings.Join([]string{
			fmt.Sprintf("# xray/rad Webhook Bridge — OAST probe for %s", req.Target),
			fmt.Sprintf("export GODNSLOG_PAYLOAD=\"%s\"", payload.TemplateRendered),
			fmt.Sprintf("export GODNSLOG_TARGET=\"%s\"", req.Target),
			fmt.Sprintf("export GODNSLOG_WEBHOOK=\"%s\"", webhookURL),
			fmt.Sprintf("export GODNSLOG_INTERACTIONS=\"%s\"", interactionsURL),
			fmt.Sprintf("export GODNSLOG_EVIDENCE=\"%s\"", evidenceURL),
			"",
			fmt.Sprintf("./xray webhook --target %s \\", target),
			fmt.Sprintf("  --webhook-url ${GODNSLOG_WEBHOOK} \\"),
			fmt.Sprintf("  --json-output xray-output.json"),
		}, "\n")
	case models.ScannerPostman, models.ScannerApifox:
		return strings.Join([]string{
			fmt.Sprintf("%s Environment Package", req.Scanner),
			fmt.Sprintf("GODNSLOG_TARGET=%s", req.Target),
			fmt.Sprintf("GODNSLOG_PAYLOAD=%s", payload.TemplateRendered),
			fmt.Sprintf("GODNSLOG_INTERACTIONS_URL=%s", interactionsURL),
			fmt.Sprintf("GODNSLOG_EVIDENCE_URL=%s", evidenceURL),
		}, "\n")
	default:
		return ""
	}
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
