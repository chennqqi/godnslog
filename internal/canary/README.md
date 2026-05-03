# Canary Token System

Canary token system for long-term monitoring of security events.

## Features

- **Multiple Token Types**: DNS, HTTP, Document, Config, CI, Storage, Email
- **Context Encoding**: Base64-encoded context for project, asset, location, owner
- **Silent Window**: Compress repeated hits within time window
- **Risk Assessment**: Automatic risk level assessment
- **Expiration**: Automatic cleanup of expired tokens
- **Hit Tracking**: Detailed hit logging with source IP, user agent, headers

## Token Types

### DNS Canary
- Monitors DNS queries for specific domains
- Useful for detecting DNS-based attacks
- Example: `canary-abc123.example.com`

### HTTP Canary
- Monitors HTTP requests for specific tokens
- Can be embedded in URLs, headers, or body
- Example: `canary-abc123` in URL path

### Document Canary
- Embedded in documents (PDF, Word, Excel)
- Triggers when document is opened
- Example: `http://canary-abc123.example.com/track`

### Config Canary
- Embedded in configuration files
- Detects config file access
- Example: `# canary: canary-abc123`

### CI Canary
- Embedded in CI/CD variables
- Detects CI/CD pipeline access
- Example: `CANARY_TOKEN=canary-abc123`

### Storage Canary
- Embedded in object storage metadata
- Detects storage access
- Example: S3 metadata with canary URL

### Email Canary
- Embedded in email headers or tracking pixels
- Detects email access
- Example: `<img src="http://canary-abc123.example.com/pixel">`

## Usage

### Create a Canary

```bash
curl -X POST http://localhost:8080/api/v2/canaries \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "dns",
    "token": "canary-abc123.example.com",
    "description": "DNS canary for project X",
    "context": "eyJwcm9qZWN0IjoicHJvamVjdC1YIiwiYXNzZXQiOiJzZXJ2ZXItMSIsImxvY2F0aW9uIjoiL2V0Yy9ob3N0cyIsIm93bmVyIjoiYWRtaW4iLCJwdXJwb3NlIjoibW9uaXRvcmluZyJ9",
    "expires_in": "90d"
  }'
```

### List Canaries

```bash
curl http://localhost:8080/api/v2/canaries \
  -H "Authorization: Bearer $API_KEY"
```

### Get Canary Hits

```bash
curl http://localhost:8080/api/v2/canaries/{id}/hits \
  -H "Authorization: Bearer $API_KEY"
```

### Get Canary Statistics

```bash
curl http://localhost:8080/api/v2/canaries/{id}/stats \
  -H "Authorization: Bearer $API_KEY"
```

## Context Encoding

### Encode Context

```go
context := &canary.CanaryContext{
    Project:  "project-X",
    Asset:    "server-1",
    Location: "/etc/hosts",
    Owner:    "admin",
    Purpose:  "monitoring",
}

encoded, err := canary.EncodeContext(context)
```

### Decode Context

```go
context, err := canary.DecodeContext(encoded)
```

## Risk Assessment

### Risk Levels

- **none**: No hits
- **low**: Few hits, low-risk sources
- **medium**: Multiple hits or medium-risk sources
- **high**: High-risk sources (local access, command-line tools)
- **critical**: Critical indicators (localhost access, automation tools)

### Risk Factors

- Local IP addresses (127.0.0.1, ::1)
- Command-line tools (curl, wget)
- Automation tools
- Repeated access patterns

## Configuration

### Default Configuration

```go
config := &canary.CanaryConfig{
    MaxRetentionDays:     90,
    DefaultExpiry:        "90d",
    SilentWindow:         300, // 5 minutes
    CompressionThreshold: 10,
    NotificationLevels:   []string{"medium", "high", "critical"},
}
```

### Silent Window

Compresses repeated hits within the time window to reduce noise:
- Default: 5 minutes
- Configurable per deployment

### Compression Threshold

Compresses old hits when threshold exceeded:
- Default: 10 hits
- Keeps recent hits, marks older as compressed

## Integration

### With Interaction Pipeline

```go
detector := canary.NewDetector(config, store)

// Process each interaction
hit, err := detector.Detect(ctx, interaction)
if err != nil {
    log.Printf("Canary detection error: %v", err)
}

if hit != nil {
    // Trigger notification
    risk := detector.AssessRisk(hit, canary)
    if contains(config.NotificationLevels, risk) {
        sendNotification(hit, risk)
    }
}
```

### With Rule Engine

Add canary detection as a rule action:
```yaml
rules:
  - name: Canary Alert
    conditions:
      - type: canary_hit
    actions:
      - type: notify
        channel: slack
        message: "Canary hit detected: {{.canary_id}}"
```

## Best Practices

1. **Context Encoding**: Always encode context with project, asset, location, owner
2. **Expiration**: Set appropriate expiration based on monitoring duration
3. **Risk Levels**: Configure notification levels based on your risk tolerance
4. **Silent Window**: Adjust silent window based on expected traffic patterns
5. **Regular Cleanup**: Schedule regular cleanup of expired canaries
6. **Unique Tokens**: Use unique tokens for different assets/projects

## Use Cases

### Supply Chain Leak Detection
- Embed canary in dependencies
- Monitor for unauthorized access
- Detect compromised packages

### Configuration Leak Detection
- Embed canary in config files
- Monitor for config file access
- Detect configuration drift

### Insider Threat Detection
- Embed canary in sensitive documents
- Monitor for document access
- Detect unauthorized access

### Lateral Movement Detection
- Embed canary in internal systems
- Monitor for network access
- Detect lateral movement

### Account Compromise Detection
- Embed canary in email signatures
- Monitor for email access
- Detect account compromise

## Troubleshooting

### No Hits Detected

- Verify canary is enabled
- Check token is correctly deployed
- Verify interaction pipeline is running
- Check silent window configuration

### Too Many Hits

- Adjust silent window
- Increase compression threshold
- Filter by risk level
- Review deployment location

### High False Positives

- Adjust risk assessment logic
- Add whitelist for known safe IPs
- Review notification levels
- Consider different token placement

## API Endpoints

### POST /canaries
Create a new canary token

### GET /canaries
List all canary tokens

### GET /canaries/{id}
Get a canary token by ID

### PUT /canaries/{id}
Update a canary token

### DELETE /canaries/{id}
Delete a canary token

### GET /canaries/{id}/hits
List hits for a canary

### GET /canaries/{id}/stats
Get statistics for a canary

## Security Considerations

1. **Token Uniqueness**: Use cryptographically random tokens
2. **Context Encryption**: Consider encrypting sensitive context
3. **Access Control**: Restrict canary management to authorized users
4. **Audit Logging**: Log all canary operations
5. **Rate Limiting**: Implement rate limiting on canary endpoints
