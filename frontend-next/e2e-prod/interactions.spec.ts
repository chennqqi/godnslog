import { test, expect } from './fixtures'
import { BASE_URL, triggerHttpHit } from './helpers/auth'

test.describe('Production Core', () => {
  test('C1 real HTTP hit is visible in interactions (API + UI)', async ({ authedPage, authToken }) => {
    const hitToken = 'pwtest' + Date.now()

    // Navigate onto the origin first so the page-level fetch below is same-origin.
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(800)

    // Trigger a real hit against the production log endpoint. It is dual-written
    // into the v2 interactions table via webapi.go record handler.
    const status = await triggerHttpHit(authedPage, hitToken)
    expect(status).toBe(200)
    // Allow the async dual-write to land
    await authedPage.waitForTimeout(1500)

    // Verify via the API
    const apiBody = await authedPage.evaluate(
      async (t) => {
        const r = await fetch('/api/v2/interactions?page=1&page_size=20', {
          headers: { Authorization: 'Bearer ' + t },
        })
        return { status: r.status, body: (await r.json()) as { data?: { items?: unknown[] } } }
      },
      authToken,
    )
    expect(apiBody.status).toBe(200)
    const items = apiBody.body?.data?.items ?? []
    const hits = items.filter((i) => JSON.stringify(i).includes(hitToken))
    expect(hits.length).toBeGreaterThanOrEqual(1)

    // Verify in the UI
    await authedPage.goto(BASE_URL + '/interactions', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(2500)
    const bodyText = await authedPage.evaluate(() => document.body.innerText)
    expect(bodyText).toContain(hitToken)
  })

  test('C2 create a case via the UI', async ({ authedPage }) => {
    await authedPage.goto(BASE_URL + '/cases', { waitUntil: 'networkidle' }).catch(() => {})
    await authedPage.waitForTimeout(1200)

    await authedPage
      .locator('button:has-text("New")')
      .first()
      .click()
      .catch(async () => {
        await authedPage.evaluate(() => {
          const btn = [...document.querySelectorAll('button')].find((b) => /new/i.test(b.textContent || ''))
          btn && (btn as HTMLButtonElement).click()
        })
      })
    await authedPage.waitForTimeout(600)

    const title = 'pw-case-' + Date.now()
    await authedPage.fill('#title', title)
    await authedPage.fill('#description', 'created by production e2e')
    await authedPage.fill('#target', 'internal-api.corp.com')

    await authedPage
      .getByRole('button', { name: /create/i })
      .click()
      .catch(async () => {
        await authedPage.evaluate(() => {
          const btn = document.querySelector('form button[type="submit"]') as HTMLButtonElement | null
          btn && btn.click()
        })
      })
    await authedPage.waitForTimeout(2000)

    const body = await authedPage.evaluate(() => document.body.innerText)
    expect(body).toContain(title)
  })
})
