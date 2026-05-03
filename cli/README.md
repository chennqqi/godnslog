# GODNSLOG CLI

The GODNSLOG CLI provides command-line access to GODNSLOG 2.0 for automation and integration with other tools.

## Installation

```bash
go build -o godnslog ./cmd/cli
```

## Configuration

Set environment variables:

```bash
export GODNSLOG_API_URL="http://localhost:8080/api/v2"
export GODNSLOG_API_KEY="your-api-key"
```

Or use flags:

```bash
godnslog --api-url http://localhost:8080/api/v2 --api-key your-api-key <command>
```

## Commands

### Case Management

```bash
# Create a case
godnslog case create --title "Security Assessment" --description "Target: example.com" --target "example.com" --tags "ssrf,xss"

# List all cases
godnslog case list

# Get case details
godnslog case get <case-id>

# Delete a case
godnslog case delete <case-id>
```

### Payload Management

```bash
# Create a payload
godnslog payload create --template "{{.Token}}.yourdomain.com" --case-id <case-id> --var "domain=yourdomain.com" --expires 24h

# List all payloads
godnslog payload list
```

### Interaction Management

```bash
# List interactions
godnslog interaction list --case-id <case-id> --type dns --limit 50

# Poll for interactions from a payload
godnslog interaction poll <payload-id> --timeout 5m --interval 10s
```

### Report Export

```bash
# Export case report
godnslog report export <case-id> --format json --output report.json
godnslog report export <case-id> --format markdown --output report.md
godnslog report export <case-id> --format csv --output report.csv

# Include raw data in export
godnslog report export <case-id> --format json --include-raw
```

## Example Workflow

```bash
# 1. Create a case for your assessment
CASE_ID=$(godnslog case create --title "SSRF Assessment" --target "https://example.com" | grep "ID:" | awk '{print $2}')

# 2. Generate a DNS payload
PAYLOAD=$(godnslog payload create --template "{{.Token}}.dns.example.com" --case-id $CASE_ID | grep "Token:" | awk '{print $2}')

# 3. Use the payload in your testing (e.g., in Nuclei)
nuclei -u https://example.com -t ssrf-templates/ -var "payload=$PAYLOAD"

# 4. Poll for interactions
godnslog interaction poll $PAYLOAD --timeout 10m

# 5. Export the report
godnslog report export $CASE_ID --format markdown --output ssrf-report.md
```

## Integration with CI/CD

```yaml
# GitHub Actions example
- name: Run OAST Scan
  env:
    GODNSLOG_API_URL: ${{ secrets.GODNSLOG_API_URL }}
    GODNSLOG_API_KEY: ${{ secrets.GODNSLOG_API_KEY }}
  run: |
    CASE_ID=$(godnslog case create --title "CI Scan" --target ${{ inputs.target }})
    PAYLOAD=$(godnslog payload create --template "{{.Token}}.example.com" --case-id $CASE_ID)
    nuclei -u ${{ inputs.target }} -t oast-templates/ -var "payload=$PAYLOAD"
    godnslog interaction poll $PAYLOAD --timeout 5m
    godnslog report export $CASE_ID --format json --output report.json
```

## JSONL Output

For integration with other tools, you can output in JSONL format:

```bash
godnslog interaction list --format jsonl > interactions.jsonl
```

Each line represents a single interaction in JSON format.
