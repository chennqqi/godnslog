package marketplace

import (
	"context"
	"fmt"
	"time"
)

// Service handles marketplace operations
type Service struct {
	store Store
}

// NewService creates a new marketplace service
func NewService(store Store) *Service {
	return &Service{store: store}
}

// Plugin operations

// CreatePlugin creates a new plugin
func (s *Service) CreatePlugin(ctx context.Context, plugin *Plugin) error {
	plugin.CreatedAt = time.Now()
	plugin.UpdatedAt = time.Now()
	plugin.Downloads = 0
	plugin.Rating = 0
	plugin.Reviews = 0
	return s.store.CreatePlugin(ctx, plugin)
}

// GetPlugin retrieves a plugin by ID
func (s *Service) GetPlugin(ctx context.Context, id string) (*Plugin, error) {
	return s.store.GetPlugin(ctx, id)
}

// ListPlugins lists all plugins
func (s *Service) ListPlugins(ctx context.Context, filters PluginFilters) ([]Plugin, error) {
	return s.store.ListPlugins(ctx, filters)
}

// UpdatePlugin updates a plugin
func (s *Service) UpdatePlugin(ctx context.Context, plugin *Plugin) error {
	plugin.UpdatedAt = time.Now()
	return s.store.UpdatePlugin(ctx, plugin)
}

// DeletePlugin deletes a plugin
func (s *Service) DeletePlugin(ctx context.Context, id string) error {
	return s.store.DeletePlugin(ctx, id)
}

// PublishPlugin publishes a plugin
func (s *Service) PublishPlugin(ctx context.Context, id string) error {
	plugin, err := s.store.GetPlugin(ctx, id)
	if err != nil {
		return err
	}
	plugin.IsPublished = true
	plugin.UpdatedAt = time.Now()
	return s.store.UpdatePlugin(ctx, plugin)
}

// UnpublishPlugin unpublishes a plugin
func (s *Service) UnpublishPlugin(ctx context.Context, id string) error {
	plugin, err := s.store.GetPlugin(ctx, id)
	if err != nil {
		return err
	}
	plugin.IsPublished = false
	plugin.UpdatedAt = time.Now()
	return s.store.UpdatePlugin(ctx, plugin)
}

// IncrementPluginDownloads increments the download count for a plugin
func (s *Service) IncrementPluginDownloads(ctx context.Context, id string) error {
	return s.store.IncrementPluginDownloads(ctx, id)
}

// AddPluginVersion adds a new version to a plugin
func (s *Service) AddPluginVersion(ctx context.Context, version *PluginVersion) error {
	version.CreatedAt = time.Now()
	if err := s.store.CreatePluginVersion(ctx, version); err != nil {
		return err
	}
	
	// Update plugin's latest version
	plugin, err := s.store.GetPlugin(ctx, version.PluginID)
	if err != nil {
		return err
	}
	plugin.LatestVersion = version.Version
	plugin.UpdatedAt = time.Now()
	return s.store.UpdatePlugin(ctx, plugin)
}

// GetPluginVersions gets all versions of a plugin
func (s *Service) GetPluginVersions(ctx context.Context, pluginID string) ([]PluginVersion, error) {
	return s.store.GetPluginVersions(ctx, pluginID)
}

// AddPluginReview adds a review for a plugin
func (s *Service) AddPluginReview(ctx context.Context, review *PluginReview) error {
	review.CreatedAt = time.Now()
	if err := s.store.CreatePluginReview(ctx, review); err != nil {
		return err
	}
	
	// Recalculate plugin rating
	s.recalculatePluginRating(ctx, review.PluginID)
	
	return nil
}

// GetPluginReviews gets reviews for a plugin
func (s *Service) GetPluginReviews(ctx context.Context, pluginID string) ([]PluginReview, error) {
	return s.store.GetPluginReviews(ctx, pluginID)
}

// recalculatePluginRating recalculates the rating for a plugin
func (s *Service) recalculatePluginRating(ctx context.Context, pluginID string) error {
	reviews, err := s.store.GetPluginReviews(ctx, pluginID)
	if err != nil {
		return err
	}
	
	if len(reviews) == 0 {
		return nil
	}
	
	sum := 0
	for _, review := range reviews {
		sum += review.Rating
	}
	
	plugin, err := s.store.GetPlugin(ctx, pluginID)
	if err != nil {
		return err
	}
	
	plugin.Rating = float64(sum) / float64(len(reviews))
	plugin.Reviews = len(reviews)
	plugin.UpdatedAt = time.Now()
	
	return s.store.UpdatePlugin(ctx, plugin)
}

// Template operations

// CreateTemplate creates a new template
func (s *Service) CreateTemplate(ctx context.Context, template *Template) error {
	template.CreatedAt = time.Now()
	template.UpdatedAt = time.Now()
	template.Downloads = 0
	template.Rating = 0
	template.Reviews = 0
	return s.store.CreateTemplate(ctx, template)
}

// GetTemplate retrieves a template by ID
func (s *Service) GetTemplate(ctx context.Context, id string) (*Template, error) {
	return s.store.GetTemplate(ctx, id)
}

// ListTemplates lists all templates
func (s *Service) ListTemplates(ctx context.Context, filters TemplateFilters) ([]Template, error) {
	return s.store.ListTemplates(ctx, filters)
}

// UpdateTemplate updates a template
func (s *Service) UpdateTemplate(ctx context.Context, template *Template) error {
	template.UpdatedAt = time.Now()
	return s.store.UpdateTemplate(ctx, template)
}

// DeleteTemplate deletes a template
func (s *Service) DeleteTemplate(ctx context.Context, id string) error {
	return s.store.DeleteTemplate(ctx, id)
}

// PublishTemplate publishes a template
func (s *Service) PublishTemplate(ctx context.Context, id string) error {
	template, err := s.store.GetTemplate(ctx, id)
	if err != nil {
		return err
	}
	template.IsPublished = true
	template.UpdatedAt = time.Now()
	return s.store.UpdateTemplate(ctx, template)
}

// UnpublishTemplate unpublishes a template
func (s *Service) UnpublishTemplate(ctx context.Context, id string) error {
	template, err := s.store.GetTemplate(ctx, id)
	if err != nil {
		return err
	}
	template.IsPublished = false
	template.UpdatedAt = time.Now()
	return s.store.UpdateTemplate(ctx, template)
}

// IncrementTemplateDownloads increments the download count for a template
func (s *Service) IncrementTemplateDownloads(ctx context.Context, id string) error {
	return s.store.IncrementTemplateDownloads(ctx, id)
}

// AddTemplateReview adds a review for a template
func (s *Service) AddTemplateReview(ctx context.Context, review *TemplateReview) error {
	review.CreatedAt = time.Now()
	if err := s.store.CreateTemplateReview(ctx, review); err != nil {
		return err
	}
	
	// Recalculate template rating
	s.recalculateTemplateRating(ctx, review.TemplateID)
	
	return nil
}

// GetTemplateReviews gets reviews for a template
func (s *Service) GetTemplateReviews(ctx context.Context, templateID string) ([]TemplateReview, error) {
	return s.store.GetTemplateReviews(ctx, templateID)
}

// recalculateTemplateRating recalculates the rating for a template
func (s *Service) recalculateTemplateRating(ctx context.Context, templateID string) error {
	reviews, err := s.store.GetTemplateReviews(ctx, templateID)
	if err != nil {
		return err
	}
	
	if len(reviews) == 0 {
		return nil
	}
	
	sum := 0
	for _, review := range reviews {
		sum += review.Rating
	}
	
	template, err := s.store.GetTemplate(ctx, templateID)
	if err != nil {
		return err
	}
	
	template.Rating = float64(sum) / float64(len(reviews))
	template.Reviews = len(reviews)
	template.UpdatedAt = time.Now()
	
	return s.store.UpdateTemplate(ctx, template)
}

// PluginInstallation operations

// InstallPlugin installs a plugin
func (s *Service) InstallPlugin(ctx context.Context, pluginID string, version string, config string) (*PluginInstallation, error) {
	installation := &PluginInstallation{
		ID:             generateInstallationID(),
		PluginID:       pluginID,
		PluginVersion:  version,
		Status:         "installed",
		Config:         config,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
	
	if err := s.store.CreatePluginInstallation(ctx, installation); err != nil {
		return nil, err
	}
	
	// Increment plugin download count
	s.store.IncrementPluginDownloads(ctx, pluginID)
	
	return installation, nil
}

// GetPluginInstallation gets an installed plugin
func (s *Service) GetPluginInstallation(ctx context.Context, id string) (*PluginInstallation, error) {
	return s.store.GetPluginInstallation(ctx, id)
}

// ListPluginInstallations lists all installed plugins
func (s *Service) ListPluginInstallations(ctx context.Context) ([]PluginInstallation, error) {
	return s.store.ListPluginInstallations(ctx)
}

// UninstallPlugin uninstalls a plugin
func (s *Service) UninstallPlugin(ctx context.Context, id string) error {
	return s.store.DeletePluginInstallation(ctx, id)
}

// Generate IDs

func generateInstallationID() string {
	return fmt.Sprintf("install-%d", time.Now().UnixNano())
}

// Filters

type PluginFilters struct {
	Type     string
	Category string
	IsOfficial *bool
	IsPublished *bool
}

type TemplateFilters struct {
	Type     string
	Category string
	IsOfficial *bool
	IsPublished *bool
}

// Store defines the storage interface for marketplace operations
type Store interface {
	// Plugin operations
	CreatePlugin(ctx context.Context, plugin *Plugin) error
	GetPlugin(ctx context.Context, id string) (*Plugin, error)
	ListPlugins(ctx context.Context, filters PluginFilters) ([]Plugin, error)
	UpdatePlugin(ctx context.Context, plugin *Plugin) error
	DeletePlugin(ctx context.Context, id string) error
	IncrementPluginDownloads(ctx context.Context, id string) error
	
	// Plugin version operations
	CreatePluginVersion(ctx context.Context, version *PluginVersion) error
	GetPluginVersions(ctx context.Context, pluginID string) ([]PluginVersion, error)
	
	// Plugin review operations
	CreatePluginReview(ctx context.Context, review *PluginReview) error
	GetPluginReviews(ctx context.Context, pluginID string) ([]PluginReview, error)
	
	// Template operations
	CreateTemplate(ctx context.Context, template *Template) error
	GetTemplate(ctx context.Context, id string) (*Template, error)
	ListTemplates(ctx context.Context, filters TemplateFilters) ([]Template, error)
	UpdateTemplate(ctx context.Context, template *Template) error
	DeleteTemplate(ctx context.Context, id string) error
	IncrementTemplateDownloads(ctx context.Context, id string) error
	
	// Template review operations
	CreateTemplateReview(ctx context.Context, review *TemplateReview) error
	GetTemplateReviews(ctx context.Context, templateID string) ([]TemplateReview, error)
	
	// Plugin installation operations
	CreatePluginInstallation(ctx context.Context, installation *PluginInstallation) error
	GetPluginInstallation(ctx context.Context, id string) (*PluginInstallation, error)
	ListPluginInstallations(ctx context.Context) ([]PluginInstallation, error)
	DeletePluginInstallation(ctx context.Context, id string) error
}
