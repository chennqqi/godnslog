# Sprint S Acceptance: Package Integrity

**Date**: 2026-06-10
**Sprint**: S - Package Integrity
**Status**: Accepted

## Summary

Sprint S successfully implemented a deterministic integrity manifest and SHA-256 content hash for Agent Run Review Evidence export and delivery flows for both JSON and Markdown formats. The `package_hash` is now consistently recorded and displayed across export responses, delivery webhook payloads, Agent Operations, Audit details, and Delivery History. E2E tests verify the full hash closure loop through Export Result, Delivery Receipt, Delivery History, and Audit details.

## Implementation

### Backend Changes

1. **Added manifest model** (`internal/models/agent_run.go`):
   - `AgentRunReviewPackageManifest` struct with schema_version, agent_run_id, format, package_hash, hash_algorithm, generated_at, and refs fields

2. **Extended response models** (`internal/models/agent_run.go`):
   - `AgentRunReviewExportResponse`: Added `Manifest *AgentRunReviewPackageManifest` and `PackageHash string` fields
   - `AgentRunReviewDeliveryResponse`: Added `PackageHash string` field
   - `AgentRunReviewDeliveryHistoryItem`: Added `PackageHash string` field

3. **Added deterministic hash helper** (`internal/models/utils.go`):
   - `ComputeDeterministicHash` function that computes SHA-256 hash of canonical JSON (sorted keys)
   - Unit tests in `internal/models/utils_test.go` verifying deterministic behavior

4. **Integrated hash computation** (`internal/agentrun/review.go`):
   - `ExportReviewPackage`: Computes package_hash for JSON format (canonical JSON) and Markdown format (exact content bytes), creates manifest, includes hash in operation result and audit details
   - `DeliverReviewPackage`: Includes package_hash in webhook payload (both top-level and refs.package_hash)
   - `createDeliverySuccess` and `createDeliveryFailure`: Retrieve package_hash from export operation, include in delivery operation result and audit details
   - `ListReviewDeliveries`: Parse package_hash from operation result for delivery history items

### Frontend Changes

1. **Updated TypeScript types** (`frontend-next/src/types/index.ts`):
   - Added `AgentRunReviewPackageManifest` interface
   - Extended `AgentRunReviewExportResponse` with `manifest` and `package_hash` fields
   - Extended `AgentRunReviewDeliveryResponse` with `package_hash` field
   - Extended `AgentRunReviewDeliveryHistoryItem` with `package_hash` field

2. **Updated Agent Run Detail UI** (`frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`):
   - Export Result: Displays package_hash (first 12 characters) with click-to-copy full hash
   - Delivery Receipt: Displays package_hash (first 12 characters) with click-to-copy full hash
   - Delivery History: Displays package_hash (first 12 characters) with click-to-copy full hash

3. **Updated Audit Page UI** (`frontend-next/src/app/dashboard/audit/audit-page-content.tsx`):
   - Audit log table now displays Package Hash from operation result details
   - Package Hash shown with click-to-copy functionality

### Testing

1. **Go Unit Tests** (`internal/agentrun/review_test.go`):
   - `TestExportPackageHash`: Verifies export response includes package_hash and manifest, operation result includes hash, audit details include hash
   - `TestDeliveryPackageHash`: Verifies delivery response includes package_hash matching export, webhook payload includes hash, delivery operation and audit include hash
   - `TestDeliveryHistoryPackageHash`: Verifies delivery history items include package_hash
   - `Markdown export has package_hash`: Verifies Markdown exports compute hash from content bytes and include manifest
   - All tests verify hash consistency (64-character SHA-256 hex strings)

2. **Go Helper Tests** (`internal/models/utils_test.go`):
   - `TestComputeDeterministicHash`: Verifies deterministic behavior with various data structures and key orders

3. **E2E Tests** (`frontend-next/e2e/agent-runs.spec.ts`):
   - `should export review evidence via UI and verify closure loop`: Mocks export API with package_hash and manifest for both JSON and Markdown formats, asserts Package Hash display in Export Result dialog, asserts Package Hash display in audit page
   - `should deliver review evidence to webhook successfully`: Mocks delivery API with package_hash, asserts Package Hash display in Delivery Receipt
   - `should list delivery history and show hash consistency`: Mocks delivery history with package_hash for delivered item, asserts Package Hash display in Delivery History
   - `should display failed and timeout delivery history`: Mocks delivery history with package_hash for failed/timeout items, asserts Package Hash display in Delivery History
   - Hash closure loop verified through mocked API responses: Export Result -> Delivery Receipt -> Delivery History -> Audit details
   - All 14 E2E tests pass

## Verification Results

### Backend Tests
```bash
GOCACHE=/tmp/gocache go test ./internal/agentrun ./server
```
Result: PASS

```bash
GOCACHE=/tmp/gocache go test ./...
```
Result: All tests passed

### Frontend Lint
```bash
cd frontend-next && npx eslint src/app/dashboard/agent-runs/page.tsx src/app/dashboard/agent-runs/[id]/page.tsx src/lib/api-client.ts src/types/index.ts e2e/agent-runs.spec.ts src/app/dashboard/audit/audit-page-content.tsx
```
Result: No errors

### Frontend Build
```bash
cd frontend-next && npm run build
```
Result: Build successful

### E2E Tests
```bash
cd frontend-next && npx playwright test --reporter=line e2e/agent-runs.spec.ts
```
Result: 14 passed (26.6s)

## Data Contracts

### Export Response (JSON format)
```json
{
  "agent_run_id": "string",
  "format": "json",
  "operation_id": "string",
  "audit_ref_id": "string",
  "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678",
  "manifest": {
    "schema_version": "review-package-manifest/v1",
    "agent_run_id": "string",
    "format": "json",
    "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678",
    "hash_algorithm": "sha256",
    "generated_at": "2026-06-08T00:00:00Z",
    "refs": {
      "operation_id": "string",
      "audit_ref_id": "string"
    }
  },
  "package": { ... },
  "generated_at": "2026-06-08T00:00:00Z"
}
```

### Export Response (Markdown format)
```json
{
  "agent_run_id": "string",
  "format": "markdown",
  "operation_id": "string",
  "audit_ref_id": "string",
  "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678",
  "manifest": {
    "schema_version": "review-package-manifest/v1",
    "agent_run_id": "string",
    "format": "markdown",
    "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678",
    "hash_algorithm": "sha256",
    "generated_at": "2026-06-08T00:00:00Z",
    "refs": {
      "operation_id": "string",
      "audit_ref_id": "string"
    }
  },
  "content": "# Markdown content",
  "generated_at": "2026-06-08T00:00:00Z"
}
```

### Delivery Response
```json
{
  "agent_run_id": "string",
  "format": "string",
  "delivery_id": "string",
  "delivery_operation_id": "string",
  "export_operation_id": "string",
  "audit_ref_id": "string",
  "destination_host": "string",
  "status_code": 200,
  "result": "delivered",
  "delivered_at": "2026-06-08T00:00:00Z",
  "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678"
}
```

### Delivery History Item
```json
{
  "delivery_id": "string",
  "delivery_operation_id": "string",
  "export_operation_id": "string",
  "audit_ref_id": "string",
  "format": "string",
  "result": "delivered",
  "destination_host": "string",
  "status_code": 200,
  "created_at": "2026-06-08T00:00:00Z",
  "delivered_at": "2026-06-08T00:00:00Z",
  "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678"
}
```

### Webhook Payload
```json
{
  "event": "agent_run.review_evidence_delivered",
  "agent_run_id": "string",
  "format": "string",
  "delivery_id": "string",
  "generated_at": "2026-06-08T00:00:00Z",
  "package": { ... },
  "refs": {
    "export_operation_id": "string",
    "delivery_operation_id": "string",
    "audit_ref_id": "string",
    "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678"
  },
  "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678"
}
```

### Audit Log Details
```json
{
  "id": "string",
  "action": "agent_run.review_exported",
  "resource_type": "agent_run",
  "resource_id": "string",
  "result": "success",
  "details": {
    "package_hash": "abc123def4567890123456789012345678901234567890123456789012345678"
  },
  "timestamp": "2026-06-08T00:00:00Z"
}
```

## Security Considerations

- The package_hash is a SHA-256 hash:
  - For JSON: canonical JSON representation of the export package (sorted keys)
  - For Markdown: exact Markdown content bytes
- Sensitive information (webhook URLs, headers, API keys) is not included in the hash computation or manifest
- The hash is deterministic: the same package content will always produce the same hash
- Hash consistency is verified across the entire lifecycle: Export Result -> Delivery Receipt -> Delivery History -> Audit details

## Non-Implemented Items (Out of Scope for Sprint S)

The following items were explicitly excluded from Sprint S scope:
- Digital signatures, private key management, PKI, JWS, certificate chains
- Report storage, signed attestations
- PDF, DOCX, ZIP export formats
- Batch export/delivery
- Case-level packages
- Saved webhook connectors
- Notification center
- Retry queues
- Background tasks
- Scanner Hub extensions
- Workflow engines
- MCP auto-delivery
- Real LLM calls
- High-risk actions like deletion, revocation, or modification of production configuration

## Files Modified

- `internal/models/agent_run.go`: Added manifest model and hash fields
- `internal/models/utils.go`: Added ComputeDeterministicHash function
- `internal/models/utils_test.go`: Added unit tests for hash computation
- `internal/agentrun/review.go`: Integrated hash computation into export/delivery flows for both JSON and Markdown
- `internal/agentrun/review_test.go`: Added tests for hash consistency, updated Markdown test to verify hash presence
- `frontend-next/src/types/index.ts`: Extended TypeScript types
- `frontend-next/src/app/dashboard/agent-runs/[id]/page.tsx`: Updated UI to display hash
- `frontend-next/src/app/dashboard/audit/audit-page-content.tsx`: Updated audit page to display Package Hash from operation details
- `frontend-next/e2e/agent-runs.spec.ts`: Added package_hash and manifest mocks to export, delivery, delivery history, and audit API responses; added Package Hash assertions for Export Result, Delivery Receipt, Delivery History, and audit page

## Conclusion

Sprint S successfully implemented package integrity for Agent Run Review Evidence export and delivery flows for both JSON and Markdown formats. The deterministic SHA-256 hash is consistently computed, stored, and displayed across all relevant components. E2E tests verify the full hash closure loop through Export Result, Delivery Receipt, Delivery History, and Audit details. All verification tests pass, and the implementation follows the security guidelines outlined in the sprint plan. Hash consistency is verified through comprehensive Go unit tests and E2E tests covering the full lifecycle (export -> delivery -> history -> audit).
