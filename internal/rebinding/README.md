# DNS Rebinding Lab

DNS Rebinding system for controlled DNS resolution experiments and security testing.

## Features

- **Multi-Stage Rebinding**: Configure multiple resolution stages with different IPs
- **Session Tracking**: Track rebinding sessions per source IP
- **Predefined Scenarios**: Built-in scenarios for common use cases
- **Conditional Advancement**: Advance stages based on hit count or conditions
- **Security Controls**: C2 disabled by default with audit logging
- **TTL Control**: Configure TTL for each stage

## Scenarios

### Browser Rebinding
- Stage 0: Resolve to 127.0.0.1 (localhost)
- Stage 1: Rebind to 192.168.1.1 (internal IP)
- Use case: Browser-based attacks, SSRF detection

### Cloud Metadata
- Stage 0: Resolve to 169.254.169.254 (cloud metadata)
- Stage 1: Rebind to internal IP after detection
- Use case: Cloud metadata access detection

### Internal Management
- Stage 0: Resolve to management interface IP
- Stage 1: Rebind to internal network
- Use case: Internal management panel detection

### IoT Device
- Stage 0: Resolve to IoT gateway
- Stage 1: Rebind to localhost after detection
- Use case: IoT device exploitation

### Router Exploit
- Stage 0: Resolve to router admin interface
- Stage 1: Rebind to localhost for exploitation
- Use case: Router vulnerability testing

## Usage

### Create Custom Rule

```bash
curl -X POST http://localhost:8080/api/v2/rebinding/rules \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "rebind.example.com",
    "stages": [
      {
        "order": 0,
        "target_ip": "127.0.0.1",
        "ttl": 10,
        "max_hits": 1,
        "description": "Localhost"
      },
      {
        "order": 1,
        "target_ip": "192.168.1.1",
        "ttl": 60,
        "max_hits": 0,
        "description": "Internal IP"
      }
    ],
    "is_enabled": true
  }'
```

### Create Rule from Scenario

```bash
curl -X POST http://localhost:8080/api/v2/rebinding/rules/scenario \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "domain": "cloud-metadata.example.com",
    "scenario": "cloud-metadata"
  }'
```

### List Rules

```bash
curl http://localhost:8080/api/v2/rebinding/rules \
  -H "Authorization: Bearer $API_KEY"
```

### Get Rule Sessions

```bash
curl http://localhost:8080/api/v2/rebinding/rules/{id}/sessions \
  -H "Authorization: Bearer $API_KEY"
```

### Get Configuration

```bash
curl http://localhost:8080/api/v2/rebinding/config \
  -H "Authorization: Bearer $API_KEY"
```

## Configuration

### Default Configuration

```go
config := &rebinding.RebindingConfig{
    DefaultTTL:      60,  // 60 seconds
    MaxStages:       5,   // Maximum 5 stages
    EnableC2:        false,  // C2 disabled
    RequireAuth:     true,  // Require auth for C2
    AuditC2:         true,  // Audit C2 operations
}
```

### Stage Configuration

Each stage supports:
- `order`: Stage order (0, 1, 2, ...)
- `target_ip`: IP address to resolve to
- `ttl`: Time to live in seconds
- `max_hits`: Maximum hits before advancing (0 = unlimited)
- `condition`: Custom condition for advancement
- `description`: Stage description

## Integration

### With DNS Server

```go
resolver := rebinding.NewResolver(config, store)

// Handle DNS query
result, err := resolver.Resolve(ctx, domain, sourceIP)
if err != nil {
    return "", err
}

// Return DNS response
return result.IP, result.TTL
```

### With Interaction Pipeline

```go
// Log rebinding as interaction
if result.IsRebind {
    interaction := &interaction.Interaction{
        Type:      "dns",
        Domain:    &domain,
        SourceIP:  sourceIP,
        RawData:   result.IP,
        Metadata: map[string]interface{}{
            "rebinding": true,
            "stage": result.Stage,
            "rule_id": result.RuleID,
        },
    }
    store.SaveInteraction(ctx, interaction)
}
```

## Security Considerations

### C2 (Command and Control)

DNS C2 is disabled by default for security reasons:

1. **Disabled by Default**: C2 capabilities are not enabled
2. **Require Approval**: Enabling C2 requires additional approval
3. **Audit Logging**: All C2 operations are logged
4. **Authentication**: C2 requires authentication

### Best Practices

1. **Scope Limitation**: Limit rebinding to specific domains
2. **Session Tracking**: Monitor rebinding sessions
3. **TTL Management**: Use appropriate TTL values
4. **Audit Logging**: Review rebinding logs regularly
5. **Approval Process**: Use approval process for C2 enablement

## Use Cases

### Security Testing

- SSRF vulnerability testing
- DNS rebinding attack testing
- Cloud metadata access detection
- Internal network mapping

### Research

- Browser DNS caching behavior
- DNS rebinding mitigation techniques
- Network security research

### Authorized Penetration Testing

- Client-side attack simulation
- Internal network discovery
- Management panel access testing

## API Endpoints

### POST /rebinding/rules
Create a custom rebinding rule

### POST /rebinding/rules/scenario
Create a rule from a predefined scenario

### GET /rebinding/rules
List all rebinding rules

### GET /rebinding/rules/{id}
Get a rebinding rule by ID

### PUT /rebinding/rules/{id}
Update a rebinding rule

### DELETE /rebinding/rules/{id}
Delete a rebinding rule

### GET /rebinding/rules/{id}/sessions
List sessions for a rule

### GET /rebinding/config
Get rebinding configuration

### PUT /rebinding/config
Update rebinding configuration

## Troubleshooting

### No Rebinding Occurring

- Verify rule is enabled
- Check domain matches exactly
- Verify DNS server integration
- Check session tracking

### Stages Not Advancing

- Check max_hits configuration
- Verify hit count incrementing
- Review condition logic
- Check session state

### Invalid IP Addresses

- Verify IP format (IPv4/IPv6)
- Use ValidateIP function
- Check for typos in configuration

## Limitations

- C2 disabled by default for security
- Requires DNS server integration
- Session tracking per source IP
- Max 5 stages per rule (configurable)

## Future Enhancements

- WebSocket-based real-time monitoring
- Graph visualization of rebinding chains
- Advanced condition evaluation
- IPv6 support
- DNSSEC support
