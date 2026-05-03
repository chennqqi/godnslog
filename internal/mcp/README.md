# GODNSLOG MCP Server

MCP (Model Context Protocol) server for AI Agent integration with GODNSLOG.

## Features

- **Tool Exposure**: Exposes GODNSLOG capabilities as MCP tools
- **Audit Logging**: All Agent operations are logged
- **Scoped API Keys**: Support for scoped API keys with minimal permissions
- **High-Risk Protection**: High-risk capabilities are disabled by default

## Available Tools

### create_case

Create a new testing case.

**Parameters**:
- `title` (string, required): Case title
- `description` (string, optional): Case description
- `target` (string, optional): Target URL or system
- `tags` (array, optional): Case tags

**Example**:
```json
{
  "title": "SSRF Vulnerability Test",
  "description": "Testing for SSRF in API endpoints",
  "target": "https://api.example.com",
  "tags": ["ssrf", "api"]
}
```

### create_payload

Create a new OAST payload.

**Parameters**:
- `template` (string, required): Payload template
- `case_id` (string, optional): Associated case ID
- `variables` (object, optional): Template variables
- `expires_in` (string, optional): Expiration time (e.g., "24h")

**Example**:
```json
{
  "template": "{{.Token}}.oast.example.com",
  "case_id": "case-123",
  "variables": {},
  "expires_in": "24h"
}
```

### list_interactions

List interactions for a case or all interactions.

**Parameters**:
- `case_id` (string, optional): Filter by case ID
- `limit` (number, optional): Maximum results (default: 50)

**Example**:
```json
{
  "case_id": "case-123",
  "limit": 100
}
```

### wait_for_interaction

Wait for an interaction to occur.

**Parameters**:
- `token` (string, required): Payload token to wait for
- `timeout` (number, optional): Timeout in seconds (default: 300)

**Example**:
```json
{
  "token": "abc123",
  "timeout": 600
}
```

### summarize_evidence

Generate an evidence summary for a case.

**Parameters**:
- `case_id` (string, required): Case ID

**Example**:
```json
{
  "case_id": "case-123"
}
```

### export_report

Export a report for a case.

**Parameters**:
- `case_id` (string, required): Case ID
- `format` (string, optional): Report format (markdown, json, csv)

**Example**:
```json
{
  "case_id": "case-123",
  "format": "markdown"
}
```

### revoke_token

Revoke an API key.

**Parameters**:
- `key_id` (string, required): API key ID to revoke

**Example**:
```json
{
  "key_id": "key-123"
}
```

## Installation

```bash
go build -o godnslog-mcp-server cmd/mcp-server/main.go
```

## Usage

### Environment Variables

- `GODNSLOG_API_URL`: GODNSLOG server URL (default: http://localhost:8080/api/v2)
- `GODNSLOG_API_KEY`: API key for authentication (required)

### Running the Server

```bash
export GODNSLOG_API_URL=http://localhost:8080/api/v2
export GODNSLOG_API_KEY=your-api-key
./godnslog-mcp-server
```

### MCP Client Configuration

Configure your MCP client to connect to the GODNSLOG MCP server:

```json
{
  "mcpServers": {
    "godnslog": {
      "command": "./godnslog-mcp-server",
      "env": {
        "GODNSLOG_API_URL": "http://localhost:8080/api/v2",
        "GODNSLOG_API_KEY": "your-api-key"
      }
    }
  }
}
```

## API Key Scopes

For Agent integration, create scoped API keys with minimal permissions:

### Read-Only Scope
- `interactions:read`
- `cases:read`

### Write Scope
- `interactions:read`
- `cases:read`
- `cases:write`
- `payloads:write`

### Admin Scope (High Risk - Disabled by Default)
- `*` (All permissions)

Create scoped API key via API:
```bash
curl -X POST http://localhost:8080/api/v2/apikeys \
  -H "Authorization: Bearer $ADMIN_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "agent-key",
    "scope": ["interactions:read", "cases:read"],
    "expires_in": "24h"
  }'
```

## Audit Logging

All Agent operations are logged with:
- Action performed
- Resource affected
- User agent (mcp-client)
- Timestamp
- Details

Audit logs can be viewed via:
- API: `GET /api/v2/audit/logs`
- Database: `audit_logs` table

## Security Considerations

1. **Scoped Permissions**: Always use scoped API keys for Agents
2. **Expiration**: Set appropriate expiration times for Agent keys
3. **Audit**: Regularly review audit logs for Agent activity
4. **High-Risk Protection**: Long-term Canary, response modification, DNS C2 are disabled by default
5. **Network Isolation**: Run MCP server in isolated network if possible

## Example Workflows

### Workflow 1: Automated SSRF Testing

1. Agent calls `create_case` for SSRF test
2. Agent calls `create_payload` with SSRF template
3. Agent injects payload into target application
4. Agent calls `wait_for_interaction` with timeout
5. Agent calls `summarize_evidence` to get summary
6. Agent calls `export_report` to generate report

### Workflow 2: Continuous Monitoring

1. Agent calls `create_case` for monitoring task
2. Agent calls `create_payload` with long expiration
3. Agent periodically calls `list_interactions`
4. Agent calls `revoke_token` when monitoring complete

## Troubleshooting

### Connection Issues

- Verify GODNSLOG API URL is accessible
- Check API key is valid and has required scopes
- Review MCP server logs for errors

### Tool Execution Failures

- Check tool parameters are correct
- Verify API key has required permissions
- Review audit logs for detailed error information

### High-Risk Capabilities Disabled

High-risk capabilities are disabled by default:
- Long-term Canary tokens
- Response modification
- DNS C2

To enable (not recommended):
1. Modify server configuration
2. Add explicit approval process
3. Implement additional authentication

## Development

### Adding New Tools

1. Add tool function in `server.go`
2. Register tool in `Run()` method
3. Add audit logging
4. Update documentation

### Testing

```bash
go test ./internal/mcp/...
```

## License

MIT
