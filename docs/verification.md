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

**Note for Sprint F E2E verification:** Due to environment-specific issues with Playwright's webServer auto-start, use the following two-step process:

```bash
cd frontend-next
npm run dev &
npx playwright test --reporter=line e2e/cases.spec.ts
kill %1  # or Ctrl+C to stop the dev server
```

Alternatively, run in separate terminals:
1. Terminal 1: `cd frontend-next && npm run dev`
2. Terminal 2: `cd frontend-next && npx playwright test --reporter=line e2e/cases.spec.ts`
3. After tests complete, stop the dev server with Ctrl+C in Terminal 1

Clean up test artifacts after verification:

```bash
cd frontend-next
rm -rf playwright-report test-results
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
