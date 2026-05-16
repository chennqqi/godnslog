# Scanner Hub Integration Contract

GODNSLOG exposes one integration contract for scanners and proxy tools.

## Create Probe

POST `/api/v2/payloads`

Required fields: `template`, `case_id`.
Optional fields: `variables`, `expires_in`, `expected_protocols`, `tool`.

## Wait For Result

GET `/api/v2/interactions?token=<token>&page_size=10`

## Result Formats

CLI integrations should support JSONL and SARIF. Webhook integrations should POST one JSON object per confirmed evidence event.

## Supported Tool Paths

- Nuclei: CLI wrapper and template variables.
- Burp Suite: extension calls the REST API.
- Yakit/Yak: Yak script calls REST API and polls token.
- ZAP: script or add-on calls REST API and polls token.
- xray/rad: CLI or webhook bridge maps scanner events to Case and Payload.
- Postman/Apifox: environment variables and pre-request scripts.
