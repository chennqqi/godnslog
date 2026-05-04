# Plugin and Template Marketplace

## Overview

Marketplace system for managing plugins and templates, enabling community contributions and sharing of reusable components.

## Features

- **Plugin Management**: Create, publish, and manage plugins
- **Template Management**: Create, publish, and manage templates
- **Version Control**: Track multiple versions of plugins
- **Reviews and Ratings**: Community feedback system
- **Download Tracking**: Monitor plugin/template popularity
- **Installation Management**: Track installed plugins
- **Filtering and Search**: Filter by type, category, official status

## Data Models

### Plugin

Represents a marketplace plugin:

- **Type**: listener, processor, notifier, exporter
- **Category**: Plugin category
- **Code**: Plugin implementation code
- **Language**: go, javascript, python
- **ConfigSchema**: JSON schema for configuration
- **Downloads**: Download count
- **Rating**: Average rating (0-5)
- **IsPublished**: Publication status
- **IsOfficial**: Official plugin flag

### PluginVersion

Represents a version of a plugin:

- **PluginID**: Parent plugin ID
- **Version**: Version string
- **Code**: Version-specific code
- **Changelog**: Version changes

### PluginReview

Represents a plugin review:

- **PluginID**: Reviewed plugin ID
- **UserID**: Reviewer user ID
- **Rating**: Rating (1-5)
- **Comment**: Review text

### Template

Represents a marketplace template:

- **Type**: payload, workflow, rule, notification
- **Content**: Template definition (JSON/YAML)
- **Format**: json, yaml
- **Category**: Template category
- **Tags**: Comma-separated tags
- **Downloads**: Download count
- **Rating**: Average rating (0-5)
- **IsPublished**: Publication status
- **IsOfficial**: Official template flag

### TemplateReview

Represents a template review:

- **TemplateID**: Reviewed template ID
- **UserID**: Reviewer user ID
- **Rating**: Rating (1-5)
- **Comment**: Review text

### PluginInstallation

Represents an installed plugin:

- **PluginID**: Installed plugin ID
- **PluginVersion**: Installed version
- **Status**: installed, disabled, error
- **Config**: JSON configuration

## Usage

### Create Plugin

```go
plugin := &marketplace.Plugin{
    Name:        "Custom Listener",
    Description: "A custom protocol listener",
    Version:     "1.0.0",
    Author:      "John Doe",
    Type:        "listener",
    Category:    "network",
    Code:        `function handle(data) { ... }`,
    Language:    "javascript",
    ConfigSchema: `{"type": "object"}`,
    IsPublished: false,
}

err := service.CreatePlugin(ctx, plugin)
```

### Publish Plugin

```go
err := service.PublishPlugin(ctx, "plugin-1")
```

### List Plugins

```go
filters := marketplace.PluginFilters{
    Type: "listener",
    Category: "network",
    IsPublished: &published,
}

plugins, err := service.ListPlugins(ctx, filters)
for _, plugin := range plugins {
    fmt.Printf("Plugin: %s, Rating: %.1f, Downloads: %d\n",
        plugin.Name, plugin.Rating, plugin.Downloads)
}
```

### Add Plugin Version

```go
version := &marketplace.PluginVersion{
    PluginID:  "plugin-1",
    Version:   "1.1.0",
    Code:      `function handle(data) { ... }`,
    Changelog: "Added new features",
}

err := service.AddPluginVersion(ctx, version)
```

### Install Plugin

```go
installation, err := service.InstallPlugin(ctx, "plugin-1", "1.1.0", "{}")
```

### Create Template

```go
template := &marketplace.Template{
    Name:        "SSRF Payload",
    Description: "SSRF detection payload",
    Type:        "payload",
    Content:     `{"type": "ssrf", "url": "{{token}}.example.com"}`,
    Format:      "json",
    Category:    "ssrf",
    Tags:        "ssrf,oob",
    IsPublished: false,
}

err := service.CreateTemplate(ctx, template)
```

### List Templates

```go
filters := marketplace.TemplateFilters{
    Type: "payload",
    Category: "ssrf",
}

templates, err := service.ListTemplates(ctx, filters)
for _, template := range templates {
    fmt.Printf("Template: %s, Rating: %.1f, Downloads: %d\n",
        template.Name, template.Rating, template.Downloads)
}
```

### Add Review

```go
review := &marketplace.PluginReview{
    PluginID: "plugin-1",
    UserID:   "user-1",
    UserName: "John Doe",
    Rating:   5,
    Comment:  "Excellent plugin!",
}

err := service.AddPluginReview(ctx, review)
```

## Best Practices

1. **Semantic Versioning**: Use semantic versioning for plugins
2. **Documentation**: Provide clear descriptions and changelogs
3. **Configuration**: Define clear configuration schemas
4. **Testing**: Test plugins before publishing
5. **Security**: Review plugin code for security issues

## Security Considerations

1. **Code Review**: Review all community plugins before use
2. **Sandboxing**: Run plugins in sandboxed environments
3. **Access Control**: Restrict plugin permissions
4. **Audit Logging**: Log all plugin installations
5. **Validation**: Validate plugin configurations

## Limitations

- No actual plugin execution engine
- No plugin sandboxing
- No dependency management
- No automatic updates
- No plugin marketplace UI

## Future Enhancements

- Plugin execution engine
- Plugin sandboxing
- Dependency management
- Automatic updates
- Web UI for marketplace
- Plugin marketplace API
- Template validation
- Plugin signing
- Marketplace moderation
