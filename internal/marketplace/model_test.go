package marketplace

import (
	"testing"
	"time"
)

// TestPluginModel tests plugin model
func TestPluginModel(t *testing.T) {
	now := time.Now()
	plugin := &Plugin{
		ID:          "plugin-1",
		Name:        "Test Plugin",
		Description: "A test plugin",
		Version:     "1.0.0",
		Author:      "Test Author",
		Type:        "listener",
		Category:    "network",
		Code:        `function test() { return "hello"; }`,
		Language:    "javascript",
		ConfigSchema: `{"type": "object"}`,
		Downloads:   100,
		Rating:      4.5,
		Reviews:     10,
		IsPublished: true,
		IsOfficial:  false,
		LatestVersion: "1.0.0",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if plugin.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if plugin.Name == "" {
		t.Fatal("Name should not be empty")
	}

	if plugin.Type == "" {
		t.Fatal("Type should not be empty")
	}
}

// TestTemplateModel tests template model
func TestTemplateModel(t *testing.T) {
	now := time.Now()
	template := &Template{
		ID:          "template-1",
		Name:        "Test Template",
		Description: "A test template",
		Type:        "payload",
		Content:     `{"test": "data"}`,
		Format:      "json",
		Category:    "ssrf",
		Tags:        "ssrf,oob",
		Downloads:   50,
		Rating:      4.0,
		Reviews:     5,
		IsPublished: true,
		IsOfficial:  false,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if template.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if template.Name == "" {
		t.Fatal("Name should not be empty")
	}

	if template.Type == "" {
		t.Fatal("Type should not be empty")
	}
}

// TestPluginVersionModel tests plugin version model
func TestPluginVersionModel(t *testing.T) {
	now := time.Now()
	version := &PluginVersion{
		ID:        "version-1",
		PluginID:  "plugin-1",
		Version:   "1.0.0",
		Code:      `function test() { return "hello"; }`,
		Changelog: "Initial release",
		CreatedAt: now,
	}

	if version.ID == "" {
		t.Fatal("ID should not be empty")
	}

	if version.PluginID == "" {
		t.Fatal("PluginID should not be empty")
	}

	if version.Version == "" {
		t.Fatal("Version should not be empty")
	}
}

// TestTableName tests table names
func TestTableName(t *testing.T) {
	plugin := Plugin{}
	if plugin.TableName() != "marketplace_plugins" {
		t.Fatalf("Expected 'marketplace_plugins', got '%s'", plugin.TableName())
	}

	template := Template{}
	if template.TableName() != "marketplace_templates" {
		t.Fatalf("Expected 'marketplace_templates', got '%s'", template.TableName())
	}

	pluginVersion := PluginVersion{}
	if pluginVersion.TableName() != "marketplace_plugin_versions" {
		t.Fatalf("Expected 'marketplace_plugin_versions', got '%s'", pluginVersion.TableName())
	}

	pluginReview := PluginReview{}
	if pluginReview.TableName() != "marketplace_plugin_reviews" {
		t.Fatalf("Expected 'marketplace_plugin_reviews', got '%s'", pluginReview.TableName())
	}

	templateReview := TemplateReview{}
	if templateReview.TableName() != "marketplace_template_reviews" {
		t.Fatalf("Expected 'marketplace_template_reviews', got '%s'", templateReview.TableName())
	}

	installation := PluginInstallation{}
	if installation.TableName() != "marketplace_plugin_installations" {
		t.Fatalf("Expected 'marketplace_plugin_installations', got '%s'", installation.TableName())
	}
}
