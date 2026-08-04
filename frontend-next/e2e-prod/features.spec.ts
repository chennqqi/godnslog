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
      let status = await authedPage.evaluate(
        async ({ endpoint, token }) => {
          const r = await fetch(endpoint, { headers: { Authorization: 'Bearer ' + token } })
          return r.status
        },
        { endpoint: ep, token: authToken },
      )
      // One retry tolerates a momentary 5xx during long full-suite runs.
      if (status !== 200) {
        await authedPage.waitForTimeout(1500)
        status = await authedPage.evaluate(
          async ({ endpoint, token }) => {
            const r = await fetch(endpoint, { headers: { Authorization: 'Bearer ' + token } })
            return r.status
          },
          { endpoint: ep, token: authToken },
        )
      }
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

  test('F4 create a canary token via API', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async ({ endpoint, body, token }) => {
        const r = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
          body: JSON.stringify(body),
        })
        return { status: r.status, body: (await r.json()) as { code?: number; message?: string } }
      },
      {
        endpoint: '/api/v2/canary',
        token: authToken,
        body: { type: 'http', token: 'pw-canary-' + Date.now().toString().slice(-8), description: 'e2e' },
      },
    )
    expect(resp.status, `canary create: ${JSON.stringify(resp.body)}`).toBe(200)
  })

  test('F5 create a rebinding rule via API', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async ({ endpoint, body, token }) => {
        const r = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
          body: JSON.stringify(body),
        })
        return { status: r.status, body: (await r.json()) as { code?: number; message?: string } }
      },
      {
        endpoint: '/api/v2/rebinding/rules',
        token: authToken,
        body: {
          domain: 'pw-rebind-' + Date.now().toString().slice(-8) + '.example.com',
          stages: [{ order: 1, target_ip: '127.0.0.1', ttl: 60, hit_count: 0, max_hits: 1 }],
          is_enabled: true,
        },
      },
    )
    expect(resp.status, `rebinding create: ${JSON.stringify(resp.body)}`).toBe(200)
  })

  test('F6 create a disabled listener via API (no side effect)', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async ({ endpoint, body, token }) => {
        const r = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
          body: JSON.stringify(body),
        })
        return { status: r.status, body: (await r.json()) as { code?: number; message?: string } }
      },
      {
        endpoint: '/api/v2/listeners',
        token: authToken,
        body: {
          protocol: 'http',
          host: '127.0.0.1',
          port: 40000 + Math.floor(Math.random() * 1000),
          token: 'pw-listener-' + Date.now().toString().slice(-8),
          is_enabled: false,
        },
      },
    )
    expect(resp.status, `listener create: ${JSON.stringify(resp.body)}`).toBe(200)
  })

  test('F7 additional list APIs are reachable', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const endpoints = [
      '/api/v2/scanner-runs',
      '/api/v2/agent-runs',
      '/api/v2/listeners',
      '/api/v2/audit/logs',
      '/api/v2/marketplace/templates',
    ]
    for (const ep of endpoints) {
      let status = await authedPage.evaluate(
        async ({ endpoint, token }) => {
          const r = await fetch(endpoint, { headers: { Authorization: 'Bearer ' + token } })
          return r.status
        },
        { endpoint: ep, token: authToken },
      )
      if (status !== 200) {
        await authedPage.waitForTimeout(1500)
        status = await authedPage.evaluate(
          async ({ endpoint, token }) => {
            const r = await fetch(endpoint, { headers: { Authorization: 'Bearer ' + token } })
            return r.status
          },
          { endpoint: ep, token: authToken },
        )
      }
      expect(status, `${ep} should not 5xx`).toBe(200)
    }
  })

  test('F8 create a marketplace template via API', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async ({ endpoint, body, token }) => {
        const r = await fetch(endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + token },
          body: JSON.stringify(body),
        })
        return { status: r.status, body: (await r.json()) as { code?: number; message?: string } }
      },
      {
        endpoint: '/api/v2/marketplace/templates',
        token: authToken,
        body: {
          name: 'pw-template-' + Date.now().toString().slice(-8),
          description: 'e2e',
          type: 'payload',
          content: 'id: pw-template',
          format: 'yaml',
        },
      },
    )
    expect(resp.status, `marketplace template create: ${JSON.stringify(resp.body)}`).toBe(200)
  })
})
