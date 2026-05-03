package casemgmt

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"errors"

	"xorm.io/xorm"
)

var (
	ErrCaseNotFound = errors.New("case not found")
)

// Service provides case management services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new case service
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// CreateCase creates a new case
func (s *Service) CreateCase(req *CaseCreateRequest, userID string) (*Case, error) {
	// Serialize tags to JSON
	tagsJSON := ""
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return nil, err
		}
		tagsJSON = string(tagsBytes)
	}

	caseRecord := &Case{
		ID:          generateID(),
		Title:       req.Title,
		Description: req.Description,
		Target:      req.Target,
		Status:      "active",
		Tags:        tagsJSON,
		CreatedBy:   userID,
	}

	if _, err := s.engine.Insert(caseRecord); err != nil {
		return nil, err
	}

	return caseRecord, nil
}

// GetCaseByID retrieves a case by its ID
func (s *Service) GetCaseByID(id string) (*Case, error) {
	var caseRecord Case
	has, err := s.engine.ID(id).Get(&caseRecord)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrCaseNotFound
	}
	return &caseRecord, nil
}

// ListCases retrieves cases with filtering
func (s *Service) ListCases(status, search string, page, pageSize int) (*CaseListResponse, error) {
	var caseRecords []Case
	session := s.engine.NewSession()
	defer session.Close()

	if status != "" {
		session = session.Where("status = ?", status)
	}
	if search != "" {
		session = session.Where("title LIKE ? OR description LIKE ?", "%"+search+"%", "%"+search+"%")
	}

	total, err := session.Count(&Case{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := session.Desc("created_at").Limit(pageSize, offset).Find(&caseRecords); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &CaseListResponse{
		Items:      caseRecords,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateCase updates a case
func (s *Service) UpdateCase(id string, req *CaseUpdateRequest) error {
	// Serialize tags to JSON
	tagsJSON := ""
	if len(req.Tags) > 0 {
		tagsBytes, err := json.Marshal(req.Tags)
		if err != nil {
			return err
		}
		tagsJSON = string(tagsBytes)
	}

	caseRecord := &Case{
		Title:       req.Title,
		Description: req.Description,
		Target:      req.Target,
		Status:      req.Status,
		Tags:        tagsJSON,
	}

	_, err := s.engine.ID(id).Cols("title", "description", "target", "status", "tags").Update(caseRecord)
	return err
}

// DeleteCase deletes a case
func (s *Service) DeleteCase(id string) error {
	_, err := s.engine.ID(id).Delete(&Case{})
	return err
}

// generateID generates a unique ID
func generateID() string {
	bytes := make([]byte, 16)
	rand.Read(bytes)
	return base32.StdEncoding.EncodeToString(bytes)
}
