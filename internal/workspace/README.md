# Multi-Workspace Support

Multi-tenant workspace system for GODNSLOG, enabling isolation and management of multiple organizations or teams.

## Features

- **Workspace Isolation**: Separate cases, payloads, and interactions per workspace
- **Multi-Domain Support**: Associate multiple domains with each workspace
- **Member Management**: Role-based access control (owner, admin, member, viewer)
- **Resource Quotas**: Configurable limits per workspace
- **Statistics**: Per-workspace usage statistics

## Workspace Model

### Workspace
- ID: Unique identifier
- Name: Workspace name
- Description: Workspace description
- OwnerID: Owner user ID
- IsEnabled: Enable/disable status
- CreatedAt/UpdatedAt: Timestamps

### WorkspaceMember
- ID: Unique identifier
- WorkspaceID: Associated workspace
- UserID: User ID
- Role: owner, admin, member, viewer
- JoinedAt: Join timestamp

### WorkspaceDomain
- ID: Unique identifier
- WorkspaceID: Associated workspace
- Domain: Domain name
- IsPrimary: Primary domain flag
- CreatedAt: Creation timestamp

### WorkspaceConfig
- MaxCases: Maximum number of cases
- MaxPayloads: Maximum number of payloads
- MaxInteractions: Maximum number of interactions
- RetentionDays: Data retention period
- EnableCanary: Canary feature enablement
- EnableRebinding: Rebinding feature enablement
- EnableListeners: Listener feature enablement

## Usage

### Create Workspace

```bash
curl -X POST http://localhost:8080/api/v2/workspaces \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Team Alpha",
    "description": "Security team workspace",
    "owner_id": "user-123",
    "is_enabled": true
  }'
```

### List Workspaces

```bash
curl http://localhost:8080/api/v2/workspaces \
  -H "Authorization: Bearer $API_KEY"
```

### Get Workspace Stats

```bash
curl http://localhost:8080/api/v2/workspaces/{id}/stats \
  -H "Authorization: Bearer $API_KEY"
```

### Add Member

```bash
curl -X POST http://localhost:8080/api/v2/workspaces/{id}/members \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-456",
    "role": "admin"
  }'
```

### Add Domain

```bash
curl -X POST http://localhost:8080/api/v2/workspaces/{id}/domains \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "team-alpha.example.com",
    "is_primary": true
  }'
```

## Programmatic Usage

### Create Workspace

```go
workspace := &workspace.Workspace{
    Name:        "Team Alpha",
    Description: "Security team workspace",
    OwnerID:     "user-123",
    IsEnabled:   true,
}

store := workspace.NewXormStore(engine)
err := store.CreateWorkspace(ctx, workspace)
```

### Add Member

```go
member := &workspace.WorkspaceMember{
    WorkspaceID: "ws-abc123",
    UserID:      "user-456",
    Role:        "admin",
}

err := store.AddWorkspaceMember(ctx, member)
```

### Add Domain

```go
domain := &workspace.WorkspaceDomain{
    WorkspaceID: "ws-abc123",
    Domain:      "team-alpha.example.com",
    IsPrimary:   true,
}

err := store.AddWorkspaceDomain(ctx, domain)
```

## Role-Based Access Control

### Owner
- Full access to all workspace resources
- Can manage members
- Can delete workspace

### Admin
- Can manage cases, payloads
- Can add/remove members (except owner)
- Cannot delete workspace

### Member
- Can create and manage own cases
- Can generate payloads
- View-only access to shared resources

### Viewer
- Read-only access to workspace
- Cannot create or modify resources

## Resource Quotas

### Default Quotas
- MaxCases: 1000
- MaxPayloads: 10000
- MaxInteractions: 100000
- RetentionDays: 90

### Quota Enforcement
When creating resources, check against workspace quotas:
```go
config, _ := store.GetWorkspaceConfig(ctx, workspaceID)
if config.MaxCases > 0 {
    currentCount := getCaseCount(ctx, workspaceID)
    if currentCount >= config.MaxCases {
        return errors.New("case quota exceeded")
    }
}
```

## Domain Management

### Primary Domain
Each workspace can have one primary domain used for:
- Default payload generation
- Canary token creation
- Rebinding rules

### Multiple Domains
Workspaces can have multiple domains for:
- Different environments (dev, staging, prod)
- Different regions
- Different purposes

## Statistics

### Available Stats
- Case count
- Payload count
- Interaction count
- Member count
- Domain count

### Usage Monitoring
Monitor workspace usage to:
- Identify heavy users
- Plan capacity
- Detect abuse
- Optimize resources

## Security Considerations

1. **Isolation**: Ensure data is properly isolated per workspace
2. **Access Control**: Enforce role-based permissions
3. **Quota Enforcement**: Prevent resource exhaustion
4. **Audit Logging**: Log all workspace operations
5. **Domain Validation**: Validate domain ownership

## Integration

### With Case Management
Associate cases with workspaces:
```go
case := &models.Case{
    WorkspaceID: workspaceID,
    // ... other fields
}
```

### With Payload Generation
Generate payloads for workspace domains:
```go
domains, _ := store.GetWorkspaceDomains(ctx, workspaceID)
primaryDomain, _ := store.GetPrimaryDomain(ctx, workspaceID)
payload := generatePayload(primaryDomain.Domain, token)
```

### With Interaction Tracking
Filter interactions by workspace:
```go
interactions, _ := store.GetInteractionsByWorkspace(ctx, workspaceID)
```

## Best Practices

1. **Naming**: Use descriptive workspace names
2. **Members**: Add only necessary members
3. **Domains**: Use subdomains for different environments
4. **Quotas**: Adjust quotas based on team size
5. **Monitoring**: Regularly review workspace statistics

## Troubleshooting

### Cannot Add Member
- Check if user already exists
- Verify role is valid
- Check workspace member limits

### Domain Not Working
- Verify DNS configuration
- Check if domain is primary
- Validate domain ownership

### Quota Exceeded
- Review workspace usage
- Clean up old data
- Request quota increase

## API Endpoints

### POST /workspaces
Create a new workspace

### GET /workspaces
List all workspaces

### GET /workspaces/{id}
Get a workspace by ID

### PUT /workspaces/{id}
Update a workspace

### DELETE /workspaces/{id}
Delete a workspace

### GET /workspaces/{id}/members
List workspace members

### POST /workspaces/{id}/members
Add a workspace member

### DELETE /workspaces/{id}/members/{user_id}
Remove a workspace member

### GET /workspaces/{id}/domains
List workspace domains

### POST /workspaces/{id}/domains
Add a workspace domain

### DELETE /workspaces/{id}/domains/{id}
Remove a workspace domain

### GET /workspaces/{id}/stats
Get workspace statistics

## Future Enhancements

- Workspace templates
- Automatic quota scaling
- Workspace cloning
- Cross-workspace sharing
- Advanced role permissions
- Workspace-level audit logs
- Billing integration
