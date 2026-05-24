# Scanner Hub Integration Contract

GODNSLOG exposes one integration contract for scanners and proxy tools.

## Scanner Hub MVP Scope

Sprint H focuses on Nuclei JSONL / template variable integration as the MVP. This is the minimum viable closed-loop for scanner integration.

## Create Probe

POST `/api/v2/payloads`

Required fields: `template`, `case_id`.
Optional fields: `variables`, `expires_at`, `expected_protocol`.

**Note**: Older plural or duration-based fields such as `expected_protocols`, `expires_in`, and `tool` are deprecated for this contract. Use the unified Payload request fields from `internal/models/payload.go` and `frontend-next/src/types/index.ts`.

## Wait For Result

GET `/api/v2/interactions?token=<token>&page_size=10`

GET `/api/v2/interactions?payload_id=<payload_id>&page_size=10`

## Evidence Generation

POST `/api/v2/evidence/generate`

Request body:
```json
{
  "case_id": "<case_id>",
  "payload_id": "<payload_id>",
  "format": "markdown"
}
```

## Nuclei MVP Integration

### Delivery Method 1: Template Variable

Use Nuclei template variable to inject GODNSLOG payload:

```bash
nuclei -u https://target.com -t godnslog-ssrf.yaml -var godnslog_payload=<rendered_payload>
```

Template example:
```yaml
id: godnslog-ssrf
info:
  name: GODNSLOG SSRF Test
  severity: medium
requests:
  - raw:
      |
      GET /?url={{godnslog_payload}} HTTP/1.1
      Host: {{Host}}
```

### Delivery Method 2: JSONL Export

Export scanner probes in JSONL format for batch processing:

```jsonl
{"scanner":"nuclei","case_id":"case-123","payload_id":"payload-456","token":"tok-abc123","target":"example.com","template":"ssrf-basic","rendered_payload":"http://tok-abc123.example.com","interactions_url":"http://godnslog/api/v2/interactions?payload_id=payload-456","evidence_url":"http://godnslog/api/v2/evidence/generate","created_at":"2026-05-24T00:00:00Z"}
```

### JSONL Minimum Fields

Each JSONL record must include:
- `scanner`: Fixed value `"nuclei"`
- `case_id`: Associated Case ID
- `payload_id`: Associated Payload ID
- `token`: Payload token for correlation
- `target`: Target system (domain, IP, URL)
- `template`: Template type (e.g., `ssrf-basic`, `xxe-basic`, `rce-callback`)
- `rendered_payload`: Fully rendered payload string
- `interactions_url`: URL to query interactions with payload_id
- `evidence_url`: URL to open payload-scoped Evidence
- `created_at`: ISO 8601 timestamp

### Wait / Evidence Workflow

After distributing probes via Nuclei:

1. **Query Interactions**:
   - API: `GET /api/v2/interactions?payload_id=<payload_id>`
   - Web: `/dashboard/interactions?payload_id=<payload_id>`

2. **Generate Evidence**:
   - API: `POST /api/v2/evidence/generate` with `{"payload_id": "<payload_id>", "format": "markdown"}`
   - Web: `/dashboard/evidence?payload_id=<payload_id>` (auto-generates)

## Supported Tool Paths (Primary Only for MVP)

- **Nuclei**: CLI wrapper and template variables (MVP - Primary)
- Burp Suite: extension calls the REST API (Primary, Phase 2)
- Yakit/Yak: Yak script calls REST API and polls token (Primary, Phase 2)
- ZAP: script or add-on calls REST API and polls token (Secondary, Phase 2)
- xray/rad: CLI or webhook bridge maps scanner events to Case and Payload (Secondary, Phase 2)
- Postman/Apifox: environment variables and pre-request scripts (Secondary, Phase 2)

**Note**: Sprint H only implements Nuclei MVP integration. Other tools are documented for future phases per `docs/official-support-boundary.md`.
