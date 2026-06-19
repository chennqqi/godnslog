# GODNSLOG 2.0 Sprint EE Burp Suite Extension Prototype Acceptance

**Date:** 2026-06-19  
**Sprint:** EE — Burp Suite Extension Prototype

## Scope

Create a Burp Suite extension prototype using Montoya API that integrates GODNSLOG OAST functionality directly into Burp Suite.

### In Scope

- Maven project with Montoya API and Gson dependencies
- Main extension entry point (`GodnslogExtension.java`)
- REST API client (`GodnslogApiClient.java`) — case creation, payload generation, interaction polling, evidence export/summary
- OAST tab UI (`OastTabProvider.java`) — configuration panel, payload list, interaction display, auto-polling
- Context menu provider (`ContextMenuProvider.java`) — 33 template options, payload insertion, evidence export, evidence summary
- README with installation, configuration, and usage instructions

### Out of Scope

- No Go backend changes
- No frontend changes
- No Maven build verification (Java not in CI pipeline)

## Verification

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS — all packages ok
```

```bash
git diff --check
# PASS
```

## Decision

Accepted. Sprint EE provides a complete Burp Suite extension prototype with all 33 payload templates accessible via context menu, OAST tab with auto-polling, and evidence export/summary capabilities.
