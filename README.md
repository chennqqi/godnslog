# GODNSLOG 2.0

![](https://z3.ax1x.com/2021/08/10/fGd4IJ.png)

A DNS & HTTP log server for verifying SSRF/XXE/RFI/RCE vulnerabilities

English Doc | [中文文档](https://github.com/chennqqi/godnslog/blob/master/README_CN.md)

## Version 2.0

GODNSLOG 2.0 is a complete rewrite with enterprise-grade features:

- **OAST Evidence Platform**: Full evidence chain with Case/Payload/Interaction tracking
- **Agent-Native MCP Server**: AI/LLM integration with minimal permissions
- **Scanner Hub**: Nuclei, Burp Suite, ZAP, Yakit/Yak, xray/rad integration
- **Workflow Automation**: Rule-based notification triggers via Webhook/Enterprise WeChat/Feishu/DingTalk
- **Canary Tokens**: Long-term monitoring with multiple token types
- **Rebinding Lab**: Multi-stage DNS rebinding with session tracking
- **Multi-Protocol Listeners**: DNS, HTTP, SMTP, LDAP, SMB, FTP
- **Enterprise Features**: Multi-workspace, data retention, audit logging

## Quick Start

### Docker

```bash
docker build -t "user/godnslog" .
docker run -p 8080:8080 -p 53:53/udp "user/godnslog" serve -domain example.com -4 127.0.0.1
```

For Chinese users:

```bash
docker build -t "user/godnslog" -f DockerfileCN .
docker run -p 8080:8080 -p 53:53/udp "user/godnslog" serve -domain example.com -4 127.0.0.1
```

### Build from Source

**Frontend (Next.js):**

```bash
cd frontend-next
npm install
npm run build
```

**Backend (Go):**

```bash
go build
```

## Configuration

### Domain Setup

1. Register your domain (e.g., `example.com`)
2. Set your DNS server to point to your host (e.g., `ns.example.com` → `100.100.100.100`)
3. Some registrars require NS hosts to point to different IPs initially
4. Access http://your-server-ip

### Default Admin

- Username: `admin`
- Password: Shown in console logs on first run
- Change password using: `go run . resetpw`

## Features

### Core OAST
- DNS/HTTP interaction capture
- Evidence timeline with scoring
- Case-based workflow
- Payload template system (30+ templates)
- Interaction clustering and noise reduction

### Scanner Integration
- Nuclei templates with JSONL output
- Burp Suite extension
- ZAP script
- Yakit/Yak script
- xray/rad integration
- CI/CD gate examples

### Agent Integration
- MCP Server (Streamable HTTP)
- Agent-specific API keys with scopes
- Audit logging for agent operations
- AI evidence summarization

### Monitoring
- Canary tokens (DNS/HTTP/SMTP)
- DNS Rebinding Lab
- Multi-protocol listeners (SMTP/LDAP/SMB/FTP)
- Webhook notifications
- Enterprise IM (WeChat/Feishu/DingTalk)

### Enterprise
- Multi-workspace isolation
- User and role management
- Data retention policies
- Audit logging
- API key management

## Documentation

- [Introduction](https://www.godnslog.com/document/introduce)
- [Payload Templates](https://www.godnslog.com/document/payload)
- [API Reference](https://www.godnslog.com/document/api)
- [Rebinding](https://www.godnslog.com/document/rebinding)
- [DNS Resolution](https://www.godnslog.com/document/resolve)

## Development

### Run Tests

```bash
# Backend tests
go test ./...

# Frontend E2E tests
cd frontend-next
npm install
npm run test:e2e
```

### CLI Usage

```bash
# View help
go run ./cmd/cli --help

# Create OAST probe
go run ./cmd/cli create-oast-probe --title "SSRF Test" --template "ssrf-basic"
```

## License

Open source, continuing under the original license.