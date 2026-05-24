# GODNSLOG Nuclei Integration

This directory contains Nuclei templates and integration examples for using GODNSLOG with the Nuclei vulnerability scanner.

## Sprint H MVP Scope

Sprint H implements Nuclei MVP integration with two delivery methods:
1. **Template Variable**: Inject GODNSLOG payload via Nuclei template variable
2. **JSONL Export**: Export scanner probes in JSONL format for batch processing

## Delivery Method 1: Template Variable

### Step 1: Create Case and Payload

```bash
# Create a case
godnslog case create --title "SSRF Scan" --target "https://target.com"

# Generate a payload
godnslog payload create --template "ssrf-basic" --case-id <case-id>
```

### Step 2: Use Payload in Nuclei

```bash
nuclei -u https://target.com -t godnslog-ssrf.yaml -var godnslog_payload=<rendered_payload>
```

### Template Example

```yaml
id: godnslog-ssrf
info:
  name: GODNSLOG SSRF Test
  severity: medium
  author: GODNSLOG
requests:
  - raw:
      |
      GET /?url={{godnslog_payload}} HTTP/1.1
      Host: {{Host}}
```

## Delivery Method 2: JSONL Export

### Step 1: Generate JSONL

Use the Scanner Hub page or API to generate JSONL records:

```jsonl
{"scanner":"nuclei","case_id":"case-123","payload_id":"payload-456","token":"tok-abc123","target":"example.com","template":"ssrf-basic","rendered_payload":"http://tok-abc123.example.com","interactions_url":"http://godnslog/api/v2/interactions?payload_id=payload-456","evidence_url":"http://godnslog/api/v2/evidence/generate","created_at":"2026-05-24T00:00:00Z"}
```

### Step 2: Use JSONL with Nuclei

Parse JSONL and extract payloads for Nuclei scanning.

## Wait / Evidence Workflow

After distributing probes via Nuclei:

### Query Interactions

**API**:
```bash
curl -X GET "http://godnslog/api/v2/interactions?payload_id=<payload_id>&page_size=10" \
  -H "Authorization: Bearer <token>"
```

**Web**:
```
http://godnslog/dashboard/interactions?payload_id=<payload_id>
```

### Generate Evidence

**API**:
```bash
curl -X POST "http://godnslog/api/v2/evidence/generate" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{"payload_id": "<payload_id>", "format": "markdown"}'
```

**Web**:
```
http://godnslog/dashboard/evidence?payload_id=<payload_id>
```

## Templates

### ssrf-basic.yaml
Detects SSRF vulnerabilities using GODNSLOG payload. Useful for:
- Cloud metadata endpoint detection
- Internal network scanning
- File upload URL parsing

### xxe-basic.yaml
Detects XXE vulnerabilities using GODNSLOG payload. Useful for:
- XML external entity injection
- File disclosure via XXE
- SSRF via XXE

### rce-callback.yaml
Detects RCE callback using GODNSLOG payload. Useful for:
- Command injection verification
- Blind RCE detection
- Outbound connection verification

## Scanner Hub Page

For a more integrated experience, use the Scanner Hub page at `/dashboard/scanner-hub` to:
- Select or create Case
- Input target
- Select template type
- Generate Nuclei command and JSONL
- Copy payload, command, or JSONL
- Navigate to Interactions and Evidence with payload scope

## Customization

You can customize the templates by modifying:
- The endpoint paths
- The HTTP methods
- The request headers
- The matching conditions

For more information, see the [Nuclei documentation](https://docs.projectdiscovery.io/templates/overview).
