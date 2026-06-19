# Postman/Apifox Integration

Use GODNSLOG as the OAST backend for API testing in Postman or Apifox.

## Setup

### 1. Import Collection

Import `godnslog-oast.postman_collection.json` into Postman or Apifox.

### 2. Configure Environment Variables

Create an environment with:

| Variable | Value | Description |
|----------|-------|-------------|
| `GODNSLOG_API_URL` | `http://your-godnslog:8080` | GODNSLOG server URL |
| `GODNSLOG_API_KEY` | `your-api-key` | Agent API Key with `agent:create_probe` and `agent:read_interactions` scopes |
| `CASE_ID` | (auto-set) | Filled by collection scripts |
| `PAYLOAD_TOKEN` | (auto-set) | Filled by collection scripts |

### 3. Configure Authentication

The collection uses API Key authentication via `X-API-Key` header. Ensure your API Key has the following scopes:

- `agent:create_probe` — for creating cases and payloads
- `agent:read_interactions` — for polling interactions
- `agent:summarize_evidence` — for evidence summary

## Usage Flow

1. **Create Case** — Creates a new OAST test case
2. **Create Payload** — Generates a tracked payload token
3. **Inject Payload** — Manual step: inject the token into your target request
4. **Poll Interactions** — Check for DNS/HTTP callbacks
5. **Generate Evidence Report** — Export structured evidence
6. **Get Evidence Summary** — Get summary bundle with next actions

## Pre-request Scripts

The collection includes Postman pre-request and test scripts that:

- Auto-extract `CASE_ID` and `PAYLOAD_TOKEN` from API responses
- Log interaction details to the Postman console
- Assert on interaction detection
- Display evidence strength and confidence

## Using with Apifox

Apifox is fully compatible with Postman collection format. Simply import the same JSON file.

### Apifox Environment

Create an environment in Apifox with the same variables listed above. The scripts and assertions work identically.

## Integration with CI

You can use Postman CLI or Apifox CLI to run this collection in CI pipelines:

```bash
# Postman CLI
postman collection run examples/postman/godnslog-oast.postman_collection.json \
  --env-var GODNSLOG_API_URL=$GODNSLOG_API_URL \
  --env-var GODNSLOG_API_KEY=$GODNSLOG_API_KEY

# Apifox CLI
apifox run examples/postman/godnslog-oast.postman_collection.json \
  --env-var GODNSLOG_API_URL=$GODNSLOG_API_URL \
  --env-var GODNSLOG_API_KEY=$GODNSLOG_API_KEY
```
