# GODNSLOG 2.0 Sprint DD Payload Template Library Expansion Acceptance

**Date:** 2026-06-19  
**Sprint:** DD — Payload Template Library Expansion

## Scope

Expand the payload template library from 5 basic templates to 33 templates covering all categories listed in Roadmap 2.0.

### In Scope

- Expand `PayloadTemplates` map in `internal/models/payload.go` from 5 to 33 templates
- Add `TemplateMetadata` struct and `PayloadTemplateMetadata` map with name, description, category, and risk for every template
- Add `TestPayloadTemplateMetadataConsistency` test verifying bidirectional consistency
- Categories covered: SSRF (6), XXE/Injection (13), RCE (2), Client-side (3), API (2), DevOps (1), Network (4), SMTP (1), Auth (1)

### Out of Scope

- No frontend changes (template listing UI already exists)
- No API endpoint changes (template listing endpoint already exists)
- No new payload service logic

## Verification

```bash
GOCACHE=/tmp/gocache go test ./internal/models -v -run 'TestPayload'
# PASS — 2 tests
```

```bash
GOCACHE=/tmp/gocache go test ./...
# PASS — all packages ok
```

```bash
git diff --check
# PASS
```

## Decision

Accepted. Sprint DD brings the payload template library to full Roadmap 2.0 coverage with 33 templates across 9 categories, each with complete metadata for UI display and API consumption.
