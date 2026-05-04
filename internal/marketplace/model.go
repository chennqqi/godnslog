package marketplace

import "time"

// Plugin represents a marketplace plugin
type Plugin struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Version     string    `json:"version"`
	Author      string    `json:"author"`
	
	// Plugin type and category
	Type        string    `json:"type"` // listener, processor, notifier, exporter
	Category    string    `json:"category"`
	
	// Plugin code
	Code        string    `json:"code"` // Plugin implementation code
	Language    string    `json:"language"` // go, javascript, python
	
	// Configuration schema
	ConfigSchema string   `json:"config_schema"` // JSON schema for configuration
	
	// Metadata
	Downloads    int       `json:"downloads"`
	Rating       float64   `json:"rating"` // 0-5
	Reviews      int       `json:"reviews"`
	
	// Status
	IsPublished  bool      `json:"is_published"`
	IsOfficial   bool      `json:"is_official"` // Official plugin from the team
	
	// Versioning
	LatestVersion string   `json:"latest_version"`
	
	CreatedAt    time.Time `json:"created_at" xorm:"created"`
	UpdatedAt    time.Time `json:"updated_at" xorm:"updated"`
}

// TableName returns the table name for Plugin
func (Plugin) TableName() string {
	return "marketplace_plugins"
}

// PluginVersion represents a version of a plugin
type PluginVersion struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	PluginID    string    `json:"plugin_id"`
	Version     string    `json:"version"`
	Code        string    `json:"code"`
	Changelog   string    `json:"changelog"`
	
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
}

// TableName returns the table name for PluginVersion
func (PluginVersion) TableName() string {
	return "marketplace_plugin_versions"
}

// PluginReview represents a plugin review
type PluginReview struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	PluginID    string    `json:"plugin_id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Rating      int       `json:"rating"` // 1-5
	Comment     string    `json:"comment"`
	
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
}

// TableName returns the table name for PluginReview
func (PluginReview) TableName() string {
	return "marketplace_plugin_reviews"
}

// Template represents a marketplace template
type Template struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	
	// Template type
	Type        string    `json:"type"` // payload, workflow, rule, notification
	
	// Template content
	Content     string    `json:"content"` // Template definition (JSON/YAML)
	Format      string    `json:"format"` // json, yaml
	
	// Metadata
	Category    string    `json:"category"`
	Tags        string    `json:"tags"` // Comma-separated tags
	
	// Usage stats
	Downloads   int       `json:"downloads"`
	Rating      float64   `json:"rating"` // 0-5
	Reviews     int       `json:"reviews"`
	
	// Status
	IsPublished bool      `json:"is_published"`
	IsOfficial  bool      `json:"is_official"`
	
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
}

// TableName returns the table name for Template
func (Template) TableName() string {
	return "marketplace_templates"
}

// TemplateReview represents a template review
type TemplateReview struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	TemplateID  string    `json:"template_id"`
	UserID      string    `json:"user_id"`
	UserName    string    `json:"user_name"`
	Rating      int       `json:"rating"` // 1-5
	Comment     string    `json:"comment"`
	
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
}

// TableName returns the table name for TemplateReview
func (TemplateReview) TableName() string {
	return "marketplace_template_reviews"
}

// PluginInstallation represents an installed plugin
type PluginInstallation struct {
	ID          string    `json:"id" xorm:"'id' pk"`
	PluginID    string    `json:"plugin_id"`
	PluginVersion string  `json:"plugin_version"`
	
	// Installation status
	Status      string    `json:"status"` // installed, disabled, error
	
	// Configuration
	Config      string    `json:"config"` // JSON configuration
	
	CreatedAt   time.Time `json:"created_at" xorm:"created"`
	UpdatedAt   time.Time `json:"updated_at" xorm:"updated"`
}

// TableName returns the table name for PluginInstallation
func (PluginInstallation) TableName() string {
	return "marketplace_plugin_installations"
}
