import { test, expect } from './fixtures'
import { BASE_URL } from './helpers/auth'

// Extended production feature coverage beyond the core preflight suite.
test.describe('Production Feature Coverage', () => {
  test('F1 core list APIs are reachable and authenticated', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)

    const endpoints = [
      '/api/v2/rules',
      '/api/v2/canary',
      '/api/v2/rebinding/rules',
      '/api/v2/retention/policies',
      '/api/v2/marketplace/plugins',
      '/api/v2/scanner-hub/adapters',
    ]
    for (const ep of endpoints) {
      const status = await authedPage.evaluate(
        async ({ endpoint, token }) => {
          const r = await fetch(endpoint, { headers: { Authorization: 'Bearer ' + token } })
          return r.status
        },
        { endpoint: ep, token: authToken },
      )
      expect(status, `${ep} should not 5xx`).toBe(200)
    }
  })

  test('F2 create a workflow rule via API', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/rules', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + t },
          body: JSON.stringify({
            name: 'pw-workflow-' + Date.now().toString().slice(-8),
            enabled: true,
            actions: [],
          }),
        })
        return { status: r.status, body: (await r.json()) as { code?: number; message?: string } }
      },
      authToken,
    )
    expect(resp.status, `workflow create: ${JSON.stringify(resp.body)}`).toBe(200)
  })

  test('F3 create a retention policy via API', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/retention/policies', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + t },
          body: JSON.stringify({
            name: 'pw-policy-' + Date.now().toString().slice(-8),
            apply_to_interactions: true,
            retention_days: 30,
          }),
        })
        return { status: r.status, body: (await r.json()) as { code?: number; message?: string } }
      },
      authToken,
    )
    expect(resp.status, `retention create: ${JSON.stringify(resp.body)}`).toBe(200)
  })
})
