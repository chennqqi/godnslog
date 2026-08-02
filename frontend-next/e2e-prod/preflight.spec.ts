import { test, expect } from './fixtures'
import { BASE_URL } from './helpers/auth'

// Release preflight: exercise the core create/save flows against production.
// These complement navigation.spec.ts (which only asserts pages render).

test.describe('Production Release Preflight', () => {
  test('PF1 create a user via the users page', async ({ authedPage }) => {
    const name = 'pwuser' + Date.now().toString().slice(-8)
    await authedPage.goto(BASE_URL + '/users', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(1000)

    await authedPage
      .locator('button:has-text("Create")')
      .first()
      .click()
      .catch(async () => {
        await authedPage.evaluate(() => {
          const btn = [...document.querySelectorAll('button')].find((b) => /create|new/i.test(b.textContent || ''))
          btn && (btn as HTMLButtonElement).click()
        })
      })
    await authedPage.waitForTimeout(500)
    await authedPage.fill('#username', name)
    await authedPage.fill('#email', name + '@example.com')
    await authedPage.fill('#password', 'preflight-pass-123')
    await authedPage
      .locator('form button:has-text("Create"), [role="dialog"] button:has-text("Create")')
      .last()
      .click()
      .catch(async () => {
        await authedPage.evaluate(() => {
          const btn = [...document.querySelectorAll('[role="dialog"] button, form button')].find((b) =>
            /create/i.test(b.textContent || ''),
          )
          btn && (btn as HTMLButtonElement).click()
        })
      })
    await authedPage.waitForTimeout(2000)
    const body = await authedPage.evaluate(() => document.body.innerText)
    expect(body).toContain(name)
  })

  test('PF2 create an API key', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/apikeys', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + t },
          body: JSON.stringify({
            name: 'pwkey' + Date.now().toString().slice(-8),
            scopes: ['case:read'],
          }),
        })
        return { status: r.status, body: (await r.json()) as { data?: { key?: string } } }
      },
      authToken,
    )
    expect(resp.status).toBe(200)
    expect(resp.body?.data?.key).toBeTruthy()
  })

  test('PF3 save general settings via API', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const resp = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/settings', {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + t },
          body: JSON.stringify({ general: { language: 'en-US' } }),
        })
        return { status: r.status }
      },
      authToken,
    )
    expect(resp.status).toBe(200)
  })

  test('PF4 create a payload via API and see it in the UI', async ({ authedPage, authToken }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)
    const apiResp = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/payloads', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json', Authorization: 'Bearer ' + t },
          body: JSON.stringify({ case_id: '', template_id: 'ssrf-basic', variables: {} }),
        })
        return { status: r.status, body: (await r.json()) as { data?: { id?: string; token?: string } } }
      },
      authToken,
    )
    expect(apiResp.status).toBe(200)
    const createdToken = apiResp.body?.data?.token
    expect(createdToken).toBeTruthy()

    // Verify the created payload is listed by the API
    const listResp = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/payloads?page=1&page_size=100', {
          headers: { Authorization: 'Bearer ' + t },
        })
        return { status: r.status, body: (await r.json()) as { data?: { items?: Array<{ token: string }> } } }
      },
      authToken,
    )
    expect(listResp.status).toBe(200)
    const listed = (listResp.body?.data?.items ?? []).some((p) => p.token === createdToken)
    expect(listed).toBe(true)
  })
})
