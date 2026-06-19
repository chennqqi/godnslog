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

// TestReplacePattern tests that replacePattern correctly replaces regex matches
func TestReplacePattern(t *testing.T) {
	tests := []struct {
		input       string
		pattern     string
		replacement string
		expected    string
	}{
		{"/users/123/posts/456", `/[0-9]+`, "/{id}", "/users/{id}/posts/{id}"},
		{"/api/a1b2c3d4-e5f6-a7b8-c9d0-e1f2a3b4c5d6/data", `/[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}`, "/{uuid}", "/api/{uuid}/data"},
		{"/hash/a1b2c3d4e5f6a7b8c9d0e1f2a3b4c5d6", `/[a-f0-9]{32,}`, "/{hash}", "/hash/{hash}"},
		{"/no-match-here", `/[0-9]+`, "/{id}", "/no-match-here"},
	}

	for _, tt := range tests {
		result := replacePattern(tt.input, tt.pattern, tt.replacement)
		if result != tt.expected {
			t.Errorf("replacePattern(%q, %q, %q) = %q, expected %q", tt.input, tt.pattern, tt.replacement, result, tt.expected)
		}
	}
}

// TestExtractPattern tests that extractPattern correctly normalizes paths
func TestExtractPattern(t *testing.T) {
	engine, err := MockEngine()
	if err != nil {
		t.Fatalf("Failed to create engine: %v", err)
	}
	defer engine.Close()

	service := NewService(engine)

	tests := []struct {
		input    string
		expected string
	}{
		{"/users/123", "/users/{id}"},
		{"/api/items/42/details", "/api/items/{id}/details"},
		{"/static/path", "/static/path"},
	}

	for _, tt := range tests {
		result := service.extractPattern(tt.input)
		if result != tt.expected {
			t.Errorf("extractPattern(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
