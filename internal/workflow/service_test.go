package workflow

import (
	"testing"

	"github.com/chennqqi/godnslog/internal/models"
	"github.com/stretchr/testify/assert"
	_ "modernc.org/sqlite"
	"xorm.io/xorm"
)

// MockEngine creates a mock xorm engine for testing
func MockEngine() (*xorm.Engine, error) {
	engine, err := xorm.NewEngine("sqlite", ":memory:")
	if err != nil {
		return nil, err
	}

	// Sync tables
	err = engine.Sync2(
		new(models.Workflow),
	)
	if err != nil {
		return nil, err
	}

	return engine, nil
}

func TestNewService(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)
	assert.NotNil(t, engine)

	service := NewService(engine)
	assert.NotNil(t, service)
}

func TestService_CreateWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)
	assert.NotEqual(t, "", workflow.ID)

	// Verify workflow was created
	var retrieved models.Workflow
	_, err = engine.ID(workflow.ID).Get(&retrieved)
	assert.NoError(t, err)
	assert.Equal(t, "Test Workflow", retrieved.Name)
}

func TestService_GetWorkflowByID(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Get workflow
	retrieved, err := service.GetWorkflowByID(workflow.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "Test Workflow", retrieved.Name)
}

func TestService_ListWorkflows(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	// Create multiple workflows
	for i := 0; i < 3; i++ {
		workflow := &models.Workflow{
			ID:          models.GenerateID(),
			Name:        "Test Workflow",
			Description: "Test Description",
			Enabled:     true,
			Actions:     models.Actions{},
			CreatedBy:   "test-user",
		}
		err = service.CreateWorkflow(workflow)
		assert.NoError(t, err)
	}

	// List workflows
	response, err := service.ListWorkflows("", nil, 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.GreaterOrEqual(t, len(response.Items), 3)
}

func TestService_UpdateWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Update workflow
	workflow.Name = "Updated Workflow"
	workflow.Description = "Updated Description"
	err = service.UpdateWorkflow(workflow)
	assert.NoError(t, err)

	// Verify update
	retrieved, err := service.GetWorkflowByID(workflow.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Workflow", retrieved.Name)
	assert.Equal(t, "Updated Description", retrieved.Description)
}

func TestService_DeleteWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions:     models.Actions{},
		CreatedBy:   "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Delete workflow
	err = service.DeleteWorkflow(workflow.ID)
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := service.GetWorkflowByID(workflow.ID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
}

func TestService_ExecuteWorkflow(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	workflow := &models.Workflow{
		ID:          models.GenerateID(),
		Name:        "Test Workflow",
		Description: "Test Description",
		Enabled:     true,
		Actions: models.Actions{
			{
				Type: "notify",
				Config: map[string]interface{}{
					"message": "Test notification",
				},
			},
		},
		CreatedBy: "test-user",
	}

	err = service.CreateWorkflow(workflow)
	assert.NoError(t, err)

	// Execute workflow
	token := "test-token"
	interaction := &models.Interaction{
		ID:       models.GenerateID(),
		Type:     "dns",
		SourceIP: "192.168.1.1",
		Token:    &token,
	}
	err = service.ExecuteWorkflow(workflow.ID, interaction)
	assert.NoError(t, err)
}
