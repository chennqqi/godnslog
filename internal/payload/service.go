package payload

import (
	"crypto/rand"
	"encoding/base32"
	"errors"
	"fmt"
	"strings"
	"text/template"
	"time"

	"xorm.io/xorm"
)

var (
	ErrPayloadNotFound = errors.New("payload not found")
	ErrInvalidTemplate = errors.New("invalid template")
)

// Service provides payload management services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new payload service
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// PayloadTemplates defines available payload templates
var PayloadTemplates = map[string]string{
	"ssrf":            "http://{token}.example.com/",
	"xxe":             "http://{token}.example.com/xxe",
	"rfi":             "http://{token}.example.com/file.php",
	"rce":             "http://{token}.example.com/cmd",
	"blind_sqli":      "http://{token}.example.com/sql?id=1",
	"ssti":            "http://{token}.example.com/template",
	"deserialization": "http://{token}.example.com/obj",
	"cors":            "http://{token}.example.com/cors",
	"jsonp":           "http://{token}.example.com/jsonp",
	"smtp_injection":  "{token}@example.com",
	"webhook":         "http://{token}.example.com/webhook",
	"ci_cd":           "http://{token}.example.com/build",
	"metadata":        "http://{token}.example.com/latest/meta-data/",
}

// generateToken generates a unique token for payload tracking
func generateToken() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return strings.ToLower(base32.StdEncoding.EncodeToString(bytes))
}

// renderPayload renders a payload template with variables
func renderPayload(tmpl string, variables map[string]string, token, domain string) (string, error) {
	// Add default variables
	if variables == nil {
		variables = make(map[string]string)
	}
	variables["token"] = token
	variables["domain"] = domain
	variables["callback_url"] = fmt.Sprintf("http://%s/log/%s/", domain, token)

	// Parse template
	t, err := template.New("payload").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Execute template
	var buf strings.Builder
	if err := t.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// renderPayloadWithCase renders a payload template with case variable
func renderPayloadWithCase(tmpl string, variables map[string]string, token, domain, caseID string) (string, error) {
	// Add default variables
	if variables == nil {
		variables = make(map[string]string)
	}
	variables["token"] = token
	variables["domain"] = domain
	variables["case"] = caseID
	variables["callback_url"] = fmt.Sprintf("http://%s/log/%s/", domain, token)

	// Parse template
	t, err := template.New("payload").Parse(tmpl)
	if err != nil {
		return "", err
	}

	// Execute template
	var buf strings.Builder
	if err := t.Execute(&buf, variables); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// CreatePayload creates a new payload
func (s *Service) CreatePayload(req *PayloadCreateRequest, userID, domain string) (*Payload, error) {
	// Validate template
	tmpl, ok := PayloadTemplates[req.Template]
	if !ok {
		return nil, ErrInvalidTemplate
	}

	// Generate token
	token := generateToken()

	// Render payload
	rendered, err := renderPayload(tmpl, req.Variables, token, domain)
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		ID:               generateID(),
		CaseID:           req.CaseID,
		Token:            token,
		Template:         req.Template,
		RenderedPayload:  rendered,
		Variables:        Variables(req.Variables),
		Status:           "draft",
		ExpectedProtocol: req.ExpectedProtocol,
		ExpiresAt:        req.ExpiresAt,
		CreatedBy:        userID,
	}

	if _, err := s.engine.Insert(payload); err != nil {
		return nil, err
	}

	return payload, nil
}

// GetPayloadByID retrieves a payload by its ID
func (s *Service) GetPayloadByID(id string) (*Payload, error) {
	var payload Payload
	has, err := s.engine.ID(id).Get(&payload)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrPayloadNotFound
	}
	return &payload, nil
}

// GetPayloadByToken retrieves a payload by its token
func (s *Service) GetPayloadByToken(token string) (*Payload, error) {
	var payload Payload
	has, err := s.engine.Where("token = ?", token).Get(&payload)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrPayloadNotFound
	}
	return &payload, nil
}

// ListPayloads retrieves payloads with filtering
func (s *Service) ListPayloads(caseID, status string, page, pageSize int) (*PayloadListResponse, error) {
	var payloads []Payload
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}
	if status != "" {
		session = session.Where("status = ?", status)
	}

	total, err := session.Count(&Payload{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := session.Desc("created_at").Limit(pageSize, offset).Find(&payloads); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &PayloadListResponse{
		Items:      payloads,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdatePayload updates a payload
func (s *Service) UpdatePayload(id string, req *PayloadUpdateRequest) error {
	// Get existing payload to validate status transition
	existingPayload, err := s.GetPayloadByID(id)
	if err != nil {
		return err
	}

	// Validate status transition
	if req.Status != "" && req.Status != existingPayload.Status {
		if !isValidPayloadStatusTransition(existingPayload.Status, req.Status) {
			return errors.New("invalid status transition")
		}
	}

	payload := &Payload{
		Status:           req.Status,
		ExpectedProtocol: req.ExpectedProtocol,
		ExpiresAt:        req.ExpiresAt,
	}

	_, err = s.engine.ID(id).Cols("status", "expected_protocol", "expires_at").Update(payload)
	return err
}

// isValidPayloadStatusTransition validates if a payload status transition is allowed
func isValidPayloadStatusTransition(from, to string) bool {
	validTransitions := map[string][]string{
		"draft":    {"deployed", "archived", "expired"},
		"deployed": {"hit", "archived", "expired"},
		"hit":      {"archived", "expired"},
		"archived": {"expired"},
		"expired":  {},
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

// RevokePayload revokes a payload by marking it as expired
func (s *Service) RevokePayload(id string) error {
	now := time.Now()
	payload := &Payload{
		Status:    "expired",
		ExpiresAt: &now,
	}

	_, err := s.engine.ID(id).Cols("status", "expires_at").Update(payload)
	return err
}

// MarkPayloadHit marks a payload as hit when an interaction is received
func (s *Service) MarkPayloadHit(token string) error {
	payload := &Payload{Status: "hit"}
	_, err := s.engine.Where("token = ?", token).Cols("status").Update(payload)
	return err
}

// generateID generates a unique ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}
