package marketplace

import (
	"context"

	"xorm.io/xorm"
)

// XormStore implements Store using XORM
type XormStore struct {
	engine *xorm.Engine
}

// NewXormStore creates a new XORM-based store
func NewXormStore(engine *xorm.Engine) *XormStore {
	return &XormStore{engine: engine}
}

// CreatePlugin creates a new plugin
func (s *XormStore) CreatePlugin(ctx context.Context, plugin *Plugin) error {
	_, err := s.engine.Insert(plugin)
	return err
}

// GetPlugin retrieves a plugin by ID
func (s *XormStore) GetPlugin(ctx context.Context, id string) (*Plugin, error) {
	var plugin Plugin
	_, err := s.engine.ID(id).Get(&plugin)
	if err != nil {
		return nil, err
	}
	return &plugin, nil
}

// ListPlugins lists all plugins
func (s *XormStore) ListPlugins(ctx context.Context, filters PluginFilters) ([]Plugin, error) {
	var plugins []Plugin
	query := s.engine.NewSession()

	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.IsOfficial != nil {
		query = query.Where("is_official = ?", *filters.IsOfficial)
	}
	if filters.IsPublished != nil {
		query = query.Where("is_published = ?", *filters.IsPublished)
	}

	err := query.Find(&plugins)
	return plugins, err
}

// UpdatePlugin updates a plugin
func (s *XormStore) UpdatePlugin(ctx context.Context, plugin *Plugin) error {
	_, err := s.engine.ID(plugin.ID).Update(plugin)
	return err
}

// DeletePlugin deletes a plugin
func (s *XormStore) DeletePlugin(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&Plugin{})
	return err
}

// IncrementPluginDownloads increments the download count for a plugin
func (s *XormStore) IncrementPluginDownloads(ctx context.Context, id string) error {
	_, err := s.engine.Table(&Plugin{}).ID(id).Incr("downloads", 1).Update(&Plugin{})
	return err
}

// CreatePluginVersion creates a new plugin version
func (s *XormStore) CreatePluginVersion(ctx context.Context, version *PluginVersion) error {
	_, err := s.engine.Insert(version)
	return err
}

// GetPluginVersions gets all versions of a plugin
func (s *XormStore) GetPluginVersions(ctx context.Context, pluginID string) ([]PluginVersion, error) {
	var versions []PluginVersion
	err := s.engine.Where("plugin_id = ?", pluginID).Desc("created_at").Find(&versions)
	return versions, err
}

// CreatePluginReview creates a new plugin review
func (s *XormStore) CreatePluginReview(ctx context.Context, review *PluginReview) error {
	_, err := s.engine.Insert(review)
	return err
}

// GetPluginReviews gets reviews for a plugin
func (s *XormStore) GetPluginReviews(ctx context.Context, pluginID string) ([]PluginReview, error) {
	var reviews []PluginReview
	err := s.engine.Where("plugin_id = ?", pluginID).Desc("created_at").Find(&reviews)
	return reviews, err
}

// CreateTemplate creates a new template
func (s *XormStore) CreateTemplate(ctx context.Context, template *Template) error {
	_, err := s.engine.Insert(template)
	return err
}

// GetTemplate retrieves a template by ID
func (s *XormStore) GetTemplate(ctx context.Context, id string) (*Template, error) {
	var template Template
	_, err := s.engine.ID(id).Get(&template)
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// ListTemplates lists all templates
func (s *XormStore) ListTemplates(ctx context.Context, filters TemplateFilters) ([]Template, error) {
	var templates []Template
	query := s.engine.NewSession()

	if filters.Type != "" {
		query = query.Where("type = ?", filters.Type)
	}
	if filters.Category != "" {
		query = query.Where("category = ?", filters.Category)
	}
	if filters.IsOfficial != nil {
		query = query.Where("is_official = ?", *filters.IsOfficial)
	}
	if filters.IsPublished != nil {
		query = query.Where("is_published = ?", *filters.IsPublished)
	}

	err := query.Find(&templates)
	return templates, err
}

// UpdateTemplate updates a template
func (s *XormStore) UpdateTemplate(ctx context.Context, template *Template) error {
	_, err := s.engine.ID(template.ID).Update(template)
	return err
}

// DeleteTemplate deletes a template
func (s *XormStore) DeleteTemplate(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&Template{})
	return err
}

// IncrementTemplateDownloads increments the download count for a template
func (s *XormStore) IncrementTemplateDownloads(ctx context.Context, id string) error {
	_, err := s.engine.Exec("UPDATE marketplace_templates SET downloads = downloads + 1 WHERE id = ?", id)
	return err
}

// CreateTemplateReview creates a new template review
func (s *XormStore) CreateTemplateReview(ctx context.Context, review *TemplateReview) error {
	_, err := s.engine.Insert(review)
	return err
}

// GetTemplateReviews gets reviews for a template
func (s *XormStore) GetTemplateReviews(ctx context.Context, templateID string) ([]TemplateReview, error) {
	var reviews []TemplateReview
	err := s.engine.Where("template_id = ?", templateID).Desc("created_at").Find(&reviews)
	return reviews, err
}

// CreatePluginInstallation creates a new plugin installation
func (s *XormStore) CreatePluginInstallation(ctx context.Context, installation *PluginInstallation) error {
	_, err := s.engine.Insert(installation)
	return err
}

// GetPluginInstallation gets an installed plugin
func (s *XormStore) GetPluginInstallation(ctx context.Context, id string) (*PluginInstallation, error) {
	var installation PluginInstallation
	_, err := s.engine.ID(id).Get(&installation)
	if err != nil {
		return nil, err
	}
	return &installation, nil
}

// ListPluginInstallations lists all installed plugins
func (s *XormStore) ListPluginInstallations(ctx context.Context) ([]PluginInstallation, error) {
	var installations []PluginInstallation
	err := s.engine.Find(&installations)
	return installations, err
}

// DeletePluginInstallation deletes a plugin installation
func (s *XormStore) DeletePluginInstallation(ctx context.Context, id string) error {
	_, err := s.engine.ID(id).Delete(&PluginInstallation{})
	return err
}
