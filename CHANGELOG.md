# Changelog

All notable changes to GODNSLOG will be documented in this file.

## [2.0.0] - 2026-05-17

### Added
- Complete rewrite with enterprise-grade architecture
- OAST Evidence Platform with Case/Payload/Interaction tracking
- Agent-Native MCP Server for AI/LLM integration
- Scanner Hub with Nuclei, Burp Suite, ZAP, Yakit/Yak, xray/rad integration
- Workflow Automation with rule-based notification triggers
- Webhook notifications (Enterprise WeChat, Feishu, DingTalk)
- Canary Tokens for long-term monitoring (DNS/HTTP/SMTP)
- DNS Rebinding Lab with multi-stage rebinding
- Multi-protocol listeners (SMTP/LDAP/SMB/FTP)
- Multi-workspace isolation
- Data retention policies
- Audit logging
- CLI tool for OAST probe creation
- 30+ payload templates (SSRF/XXE/RCE/Blind SQLi/deserialization/LDAP/SMB/FTP/DNS Rebinding/Log4j JNDI/cloud metadata SSRF)
- Interaction clustering and noise reduction
- Evidence timeline with scoring
- Agent-specific API keys with scopes
- AI evidence summarization

### Changed
- Frontend migrated to Next.js 16 with TypeScript
- Backend refactored with internal/ directory structure
- Unified data models in internal/models/
- API v2 endpoints for all core features
- Docker build updated to use frontend-next
- Go version updated to 1.22
- Node version updated to 24.13.0

### Fixed
- Login authentication flow
- API response format consistency
- E2E test framework with Playwright
- Frontend routing and state management

### Removed
- Old Vue frontend (replaced by Next.js)
- Legacy API v1 endpoints (maintained for compatibility)

## [1.0.0] - Earlier

- Initial DNS/HTTP log server
- Basic user management
- DNS rebinding support
- Canary tokens
- Web UI
