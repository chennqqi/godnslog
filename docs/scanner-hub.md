# Scanner Hub Integration Contract

GODNSLOG exposes one integration contract for scanners and proxy tools.

## Scanner Hub Scope

Sprint H delivered the Nuclei JSONL / template variable MVP. Sprint U expands Scanner Hub into a multi-tool adapter package generator while keeping the same GODNSLOG evidence loop.

Scanner Hub currently generates integration packages and persists Scanner Runs. It does **not** execute scanners, schedule scans, run plugin binaries, or provide bidirectional live scanner event streaming.

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

## Adapter Catalog API

GET `/api/v2/scanner-hub/adapters`

Returns the official Scanner Hub adapter catalog:

```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [
      {
        "id": "burp",
        "name": "Burp Suite",
        "category": "Native",
        "maturity": "L4 Native Plugin Path",
        "supported_methods": ["burp-extension"],
        "default_method": "burp-extension",
        "description": "Manual and semi-automated verification package for Burp Suite extension workflows."
      }
    ]
  }
}
```

## Supported Tool Paths

| Tool | Supported Delivery Method | Current Support | Boundary |
|------|---------------------------|-----------------|----------|
| Nuclei | `nuclei-jsonl`, `nuclei-var` | Command + JSONL package generation | No scan execution |
| Burp Suite | `burp-extension` | Extension package instructions with API/polling links | No compiled extension binary |
| Yakit/Yak | `yakit-script` | Yak script package instructions | No embedded Yak runtime |
| ZAP | `zap-script` | ZAP script package instructions | No add-on binary |
| xray | `xray-webhook` | Webhook bridge package metadata | No scanner process control |
| rad | `rad-webhook` | Webhook bridge package metadata | No crawler/scanner execution |
| Postman | `postman-env` | Environment variable package | No collection upload |
| Apifox | `apifox-env` | Environment variable package | No workspace sync |

## Create Scanner Run

POST `/api/v2/scanner-runs`

Required fields:

```json
{
  "case_id": "case-123",
  "payload_id": "payload-456",
  "scanner": "burp",
  "target": "https://target.example",
  "template": "ssrf-basic",
  "delivery_method": "burp-extension"
}
```

Supported compatibility matrix:

```text
nuclei  -> nuclei-jsonl, nuclei-var
burp    -> burp-extension
yakit   -> yakit-script
zap     -> zap-script
xray    -> xray-webhook
rad     -> rad-webhook
postman -> postman-env
apifox  -> apifox-env
```

Invalid scanner names and invalid scanner/delivery combinations return 400.

The response includes:

- `scanner`
- `delivery_method`
- `command`
- `jsonl`
- `package_manifest`
- `package_hash`
- Case/Payload linkage
- URLs back to Interactions and Evidence

## Machine-Readable Package Manifest

Sprint W adds a machine-readable integration package manifest to every Scanner Run. This is intended for AI Agent, CI, and scanner automation consumers that need to verify package contents without parsing human-facing command text.

Every create/list/detail response includes:

```json
{
  "package_hash": "0123456789abcdef...",
  "package_manifest": {
    "schema_version": "scanner-package.v1",
    "scanner": "burp",
    "delivery_method": "burp-extension",
    "package_hash": "0123456789abcdef...",
    "hash_algorithm": "sha256",
    "files": [
      {
        "name": "README.md",
        "kind": "instructions",
        "description": "Operator and Agent instructions for distributing this GODNSLOG scanner package."
      },
      {
        "name": "godnslog-package.jsonl",
        "kind": "jsonl",
        "description": "Single-line GODNSLOG scanner package record with payload, target, and callback URLs."
      },
      {
        "name": "burp-extension-config.json",
        "kind": "burp-extension-config",
        "description": "Burp Suite extension configuration inputs and polling URLs."
      }
    ],
    "interactions_url": "http://godnslog/api/v2/interactions?payload_id=payload-456",
    "evidence_url": "http://godnslog/dashboard/evidence?payload_id=payload-456",
    "next_actions": [
      "Distribute the generated payload through the selected scanner adapter.",
      "Poll interactions_url until the expected callback is observed.",
      "Open evidence_url or call the Evidence API to produce the final proof chain."
    ]
  }
}
```

The package hash is a deterministic SHA-256 hash over the generated command, JSONL package data, manifest file list, scanner, delivery method, target, template, and correlation URLs. The timestamp in the JSONL record is excluded from the hash input so equivalent package contents remain stable across repeated generation.

The manifest does not imply scanner execution. It describes package files and next actions only; GODNSLOG still does not run Nuclei, Burp Suite, Yakit/Yak, ZAP, xray, rad, Postman, or Apifox.

## Evidence Summary Lookup

Sprint X adds a read-only evidence summary endpoint for scanner-driven review:

```http
POST /api/v2/evidence/summary
```

For Scanner Hub workflows, pass the Scanner Run ID:

```json
{
  "scanner_run_id": "scanner-run-789"
}
```

The response resolves the Scanner Run to Case/Payload scope, generates on-demand evidence from captured Interactions, includes related Scanner Run metadata, returns `package_hashes`, and provides a deterministic `summary_hash`. This is the preferred contract for AI Agent and CI review loops that need one structured evidence bundle.
