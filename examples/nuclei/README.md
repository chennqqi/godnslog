# GODNSLOG Nuclei Templates

This directory contains Nuclei templates for using GODNSLOG with the Nuclei vulnerability scanner.

## Usage

### Using with Interactsh-compatible mode

GODNSLOG 2.0 provides an Interactsh-compatible API for seamless integration with Nuclei.

```bash
nuclei -u https://target.com -t examples/nuclei/ -interactsh-server http://your-godnslog-server
```

### Manual CLI Workflow

For more control, use the GODNSLOG CLI:

```bash
# Create a case
godnslog case create --title "SSRF Scan" --target "https://target.com"

# Generate a payload
godnslog payload create --template "{{.Token}}.yourdomain.com" --case-id <case-id>

# Use the payload in Nuclei
nuclei -u https://target.com -t examples/nuclei/ -var "payload=<payload-token>"

# Poll for interactions
godnslog interaction poll <payload-id> --timeout 10m

# Export report
godnslog report export <case-id> --format markdown --output report.md
```

## Templates

### dns-oast.yaml
Detects DNS-based out-of-band interactions. Useful for:
- DNS record exfiltration
- DNS rebinding detection
- Subdomain takeover verification

### http-oast.yaml
Detects HTTP-based out-of-band interactions. Useful for:
- SSRF detection
- Blind XSS detection
- Callback verification

### ssrf-oast.yaml
Specifically targets SSRF vulnerabilities. Useful for:
- Cloud metadata endpoint detection
- Internal network scanning
- File upload URL parsing

## Customization

You can customize the templates by modifying:
- The endpoint paths
- The HTTP methods
- The request headers
- The matching conditions

For more information, see the [Nuclei documentation](https://docs.projectdiscovery.io/templates/overview).
