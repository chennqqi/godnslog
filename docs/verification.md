# Verification Gates

Run these commands before claiming a GODNSLOG 2.0 change is complete.

## Backend

```bash
GOCACHE=/tmp/gocache go test ./...
```

For focused backend changes, run the affected package first, then the full command before release.

## Frontend

```bash
cd frontend-next
npm run lint
npm run build
```

For UI behavior changes, also run the relevant Playwright spec:

```bash
cd frontend-next
npm run test:e2e -- e2e/cases.spec.ts
```

## MCP

```bash
GOCACHE=/tmp/gocache go test -count=1 ./internal/mcp
```
