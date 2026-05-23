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
npx playwright test --reporter=line e2e/cases.spec.ts
```

Do not use interactive Playwright HTML report commands during verification. In particular, avoid any flow that triggers:

```text
Serving HTML report at http://localhost:9323. Press Ctrl+C to quit.
```

Forbidden during routine verification:

- `npx playwright show-report`
- `npm run test:e2e:ui`
- any command path that leaves a local HTML report server running

## MCP

```bash
GOCACHE=/tmp/gocache go test -count=1 ./internal/mcp
```
