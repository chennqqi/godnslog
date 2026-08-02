import { test, expect } from './fixtures'
import { loginAndGetToken, PROD_PASSWORD, PROD_USER, BASE_URL } from './helpers/auth'
import { solveCaptcha, submitLogin } from './helpers/captcha'

test.describe('Production Auth', () => {
  test('A1 login succeeds via slide captcha', async ({ browser }) => {
    const { token, context } = await loginAndGetToken(browser)
    expect(token.length).toBeGreaterThan(10)
    await context.close()
  })

  test('A2 wrong password is rejected (401)', async ({ browser }) => {
    const context = await browser.newContext({ ignoreHTTPSErrors: true })
    const page = await context.newPage()
    await solveCaptcha(page, BASE_URL + '/login')

    let status = 0
    page.on('response', (res) => {
      if (res.url().includes('/auth/login')) status = res.status()
    })
    await submitLogin(page, PROD_USER, 'wrongpass-123')
    await page.waitForTimeout(2000)
    expect(status).toBe(401)
    await context.close()
  })

  test('A3 captcha refresh issues a new id', async ({ browser }) => {
    const context = await browser.newContext({ ignoreHTTPSErrors: true })
    const page = await context.newPage()

    const ids: (string | undefined)[] = []
    page.on('response', async (res) => {
      if (res.url().includes('/auth/captcha')) {
        try {
          const j = (await res.json()) as { data?: { captcha_id?: string } }
          ids.push(j?.data?.captcha_id)
        } catch {}
      }
    })
    await page.goto(BASE_URL + '/login', { waitUntil: 'networkidle' })
    // Wait for the initial captcha to load
    await page.waitForFunction(() => {
      const img = document.querySelector('img[alt="captcha background"]') as HTMLImageElement | null
      return img && img.complete && img.naturalWidth > 0
    })
    await page.waitForTimeout(400)

    // Click the refresh button
    await page.evaluate(() => {
      const btn = [...document.querySelectorAll('button')].find((b) => b.textContent?.includes('Refresh'))
      btn && (btn as HTMLButtonElement).click()
    })
    await page.waitForTimeout(1200)

    const id1 = ids[0]
    const id2 = ids[ids.length - 1]
    expect(id1).toBeTruthy()
    expect(id2).toBeTruthy()
    expect(id1).not.toBe(id2)
    await context.close()
  })

  test('A4 session persists across reload', async ({ authedPage }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' })
    await authedPage.reload({ waitUntil: 'networkidle' })
    expect(authedPage.url()).not.toContain('/login')
  })

  test('A5 logout redirects to /login and clears token', async ({ authedPage }) => {
    await authedPage.goto(BASE_URL + '/', { waitUntil: 'networkidle' })
    // Open the user menu (button labelled with the username) and click Sign out
    await authedPage.locator('button:has-text("admin")').last().click()
    await authedPage.waitForTimeout(600)
    await authedPage.getByText('Sign out', { exact: false }).last().click()
    await authedPage.waitForTimeout(2000)
    expect(authedPage.url()).toContain('/login')
    const token = await authedPage.evaluate(() => localStorage.getItem('token'))
    expect(token).toBeNull()
  })
})

// A sentinel so the fixture's skip (when ADMIN_PASSWORD is unset) is meaningful.
test('admin password is configured', () => {
  expect(PROD_PASSWORD.length).toBeGreaterThan(0)
})
