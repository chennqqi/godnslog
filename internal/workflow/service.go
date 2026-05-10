package workflow

import (
	"errors"
	"time"

	"xorm.io/xorm"

	"github.com/chennqqi/godnslog/internal/models"
)

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
)

// Service provides workflow management services
type Service struct {
	engine *xorm.Engine
}

// NewService creates a new workflow service
func NewService(engine *xorm.Engine) *Service {
	return &Service{engine: engine}
}

// CreateWorkflow creates a new workflow
func (s *Service) CreateWorkflow(workflow *models.Workflow) error {
	if workflow.ID == "" {
		workflow.ID = models.GenerateID()
	}
	if workflow.Actions == nil {
		workflow.Actions = models.Actions{}
	}
	if workflow.CreatedAt.IsZero() {
		workflow.CreatedAt = time.Now()
	}
	if workflow.UpdatedAt.IsZero() {
		workflow.UpdatedAt = time.Now()
	}

	_, err := s.engine.Insert(workflow)
	return err
}

// GetWorkflowByID retrieves a workflow by its ID
func (s *Service) GetWorkflowByID(id string) (*models.Workflow, error) {
	var workflow models.Workflow
	has, err := s.engine.ID(id).Get(&workflow)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, ErrWorkflowNotFound
	}
	return &workflow, nil
}

// ListWorkflows retrieves workflows with filtering
func (s *Service) ListWorkflows(caseID string, enabled *bool, page, pageSize int) (*models.WorkflowListResponse, error) {
	var workflows []models.Workflow
	session := s.engine.NewSession()
	defer session.Close()

	if caseID != "" {
		session = session.Where("case_id = ?", caseID)
	}
	if enabled != nil {
		session = session.Where("enabled = ?", *enabled)
	}

	total, err := session.Count(&models.Workflow{})
	if err != nil {
		return nil, err
	}

	offset := (page - 1) * pageSize
	if err := session.Desc("created_at").Limit(pageSize, offset).Find(&workflows); err != nil {
		return nil, err
	}

	totalPages := int(total) / pageSize
	if int(total)%pageSize > 0 {
		totalPages++
	}

	return &models.WorkflowListResponse{
		Items:      workflows,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}, nil
}

// UpdateWorkflow updates an existing workflow
func (s *Service) UpdateWorkflow(workflow *models.Workflow) error {
	workflow.UpdatedAt = time.Now()

	_, err := s.engine.ID(workflow.ID).Update(workflow)
	return err
}

// DeleteWorkflow deletes a workflow by its ID
func (s *Service) DeleteWorkflow(id string) error {
	_, err := s.engine.ID(id).Delete(&models.Workflow{})
	return err
}

// ExecuteWorkflow executes a workflow's actions
func (s *Service) ExecuteWorkflow(workflowID string, interaction *models.Interaction) error {
	workflow, err := s.GetWorkflowByID(workflowID)
	if err != nil {
		return err
	}

	if !workflow.Enabled {
		return nil // Skip disabled workflows
	}

	// Execute each action in the workflow
	for _, action := range workflow.Actions {
		if !action.Enabled {
			continue
		}

		if err := s.executeAction(action, interaction); err != nil {
			// Log error but continue with other actions
			continue
		}
	}

	return nil
}

// executeAction executes a single action
func (s *Service) executeAction(action models.Action, interaction *models.Interaction) error {
	switch action.Type {
	case models.ActionTypeHTTP:
		return s.executeHTTPAction(action, interaction)
	case models.ActionTypeDNS:
		return s.executeDNSAction(action, interaction)
	case models.ActionTypeWebhook:
		return s.executeWebhookAction(action, interaction)
	case models.ActionTypeNotify:
		return s.executeNotifyAction(action, interaction)
	default:
		return errors.New("unsupported action type")
	}
}

// executeHTTPAction executes an HTTP action
func (s *Service) executeHTTPAction(action models.Action, interaction *models.Interaction) error {
	// TODO: Implement HTTP action execution
	return nil
}

// executeDNSAction executes a DNS action
func (s *Service) executeDNSAction(action models.Action, interaction *models.Interaction) error {
	// TODO: Implement DNS action execution
	return nil
}

// executeWebhookAction executes a webhook action
func (s *Service) executeWebhookAction(action models.Action, interaction *models.Interaction) error {
	// TODO: Implement webhook action execution
	return nil
}

// executeNotifyAction executes a notification action
func (s *Service) executeNotifyAction(action models.Action, interaction *models.Interaction) error {
	// TODO: Implement notification action execution
	return nil
}
