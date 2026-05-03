# Payload Templates

This directory contains predefined OAST payload templates for common vulnerability scenarios.

## Template Categories

### SSRF (Server-Side Request Forgery)
- `ssrf-http`: Basic HTTP SSRF detection
- `ssrf-cloud-metadata`: Cloud metadata endpoint detection (AWS, GCP, Azure)

### Injection
- `xxe-external-entity`: XML External Entity injection
- `rfi-remote-file`: Remote File Inclusion
- `ssti-template`: Server-Side Template Injection
- `smtp-injection`: SMTP header injection

### RCE (Remote Code Execution)
- `rce-command`: Command injection detection

### SQL Injection
- `blind-sqli-dns`: Blind SQL injection via DNS exfiltration

### Misconfiguration
- `cors-jsonp`: CORS misconfiguration and JSONP

### Client-Side
- `pdf-html-rendering`: PDF/HTML rendering with external resources

### API & DevOps
- `webhook`: Webhook endpoint detection
- `ci-cd-variable`: CI/CD pipeline variable injection

## Usage

### CLI

```bash
# Use a specific template
godnslog payload create --template "{{.Token}}.oast.example.com"

# Load from templates file
godnslog payload create --template @templates/payloads.json --template-id ssrf-http
```

### API

```bash
curl -X POST http://localhost:8080/api/v2/payloads \
  -H "Authorization: Bearer $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "template": "{{.Token}}.oast.example.com",
    "case_id": "case-123",
    "expires_in": "24h"
  }'
```

### Frontend

Use Payload Studio to select and customize templates.

## Custom Templates

Add custom templates to `payloads.json`:

```json
{
  "id": "custom-template",
  "name": "Custom Template",
  "description": "Description of your template",
  "template": "{{.Token}}.custom-domain.com/path",
  "variables": {
    "custom_var": "value"
  },
  "category": "custom",
  "risk": "medium"
}
```

## Variables

Templates support the following variables:

- `{{.Token}}`: The unique payload token
- `{{.Case}}`: Case ID
- `{{.Domain}}`: Base domain
- `{{.CallbackURL}}`: Full callback URL
- `{{.Base32Context}}`: Base32 encoded context

Example:
```
{{.Token}}.{{.Domain}}/{{.Case}}
```

## Risk Levels

- **critical**: Immediate attention required (e.g., cloud metadata access)
- **high**: High severity vulnerabilities (e.g., SSRF, RCE)
- **medium**: Medium severity (e.g., CORS misconfiguration)
- **low**: Low severity (e.g., webhooks)

## Best Practices

1. **Use Descriptive Names**: Make template names clear and specific
2. **Add Descriptions**: Explain when to use each template
3. **Set Appropriate Risk**: Assign correct risk levels for prioritization
4. **Categorize Properly**: Use existing categories or create new ones
5. **Test Templates**: Verify templates work before using in production
6. **Document Variables**: List any custom variables required

## Template Migration

To migrate from 1.0 format:

1. Extract template definitions from old configuration
2. Convert to new JSON format
3. Add metadata (category, risk)
4. Test with CLI or API
5. Update documentation

## Contributing

To contribute new templates:

1. Add template to `payloads.json`
2. Add description to this README
3. Test with real scenarios
4. Submit pull request
