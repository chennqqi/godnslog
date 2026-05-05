package interaction

import (
	"testing"
	"time"

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
		new(models.Interaction),
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

func TestService_CreateInteraction(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	token := "test-token"
	interaction := &models.Interaction{
		ID:        models.GenerateID(),
		Type:      "dns",
		SourceIP:  "192.168.1.1",
		Token:     &token,
		Timestamp: time.Now(),
	}

	err = service.CreateInteraction(interaction)
	assert.NoError(t, err)
	assert.NotEqual(t, "", interaction.ID)

	// Verify interaction was created
	var retrieved models.Interaction
	_, err = engine.ID(interaction.ID).Get(&retrieved)
	assert.NoError(t, err)
	assert.Equal(t, "dns", retrieved.Type)
}

func TestService_GetInteractionByID(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	token := "test-token"
	interaction := &models.Interaction{
		ID:        models.GenerateID(),
		Type:      "dns",
		SourceIP:  "192.168.1.1",
		Token:     &token,
		Timestamp: time.Now(),
	}

	err = service.CreateInteraction(interaction)
	assert.NoError(t, err)

	// Get interaction
	retrieved, err := service.GetInteractionByID(interaction.ID)
	assert.NoError(t, err)
	assert.NotNil(t, retrieved)
	assert.Equal(t, "dns", retrieved.Type)
}

func TestService_ListInteractions(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	// Create multiple interactions
	for i := 0; i < 3; i++ {
		token := "test-token"
		interaction := &models.Interaction{
			ID:        models.GenerateID(),
			Type:      "dns",
			SourceIP:  "192.168.1.1",
			Token:     &token,
			Timestamp: time.Now(),
		}
		err = service.CreateInteraction(interaction)
		assert.NoError(t, err)
	}

	// List interactions
	response, err := service.ListInteractions("", "", "", nil, nil, 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.GreaterOrEqual(t, len(response.Items), 3)
}

func TestService_DeleteInteractions(t *testing.T) {
	engine, err := MockEngine()
	assert.NoError(t, err)

	service := NewService(engine)

	token := "test-token"
	interaction := &models.Interaction{
		ID:        models.GenerateID(),
		Type:      "dns",
		SourceIP:  "192.168.1.1",
		Token:     &token,
		Timestamp: time.Now(),
	}

	err = service.CreateInteraction(interaction)
	assert.NoError(t, err)

	// Delete interaction using engine directly
	_, err = engine.ID(interaction.ID).Delete(&models.Interaction{})
	assert.NoError(t, err)

	// Verify deletion
	retrieved, err := service.GetInteractionByID(interaction.ID)
	assert.Error(t, err)
	assert.Nil(t, retrieved)
}
