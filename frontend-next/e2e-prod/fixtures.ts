import { test as base, expect, type Page } from '@playwright/test'
import { loginAndGetToken, PROD_PASSWORD } from './helpers/auth'

type Fixtures = {
  /** Real production JWT, obtained via one slide-captcha login. */
  authToken: string
  /** Page with the token/user injected into localStorage (fast, no re-login). */
  authedPage: Page
}

export const test = base.extend<Fixtures>({
  authToken: async ({ browser }, use) => {
    if (!PROD_PASSWORD) {
      test.skip('set ADMIN_PASSWORD env var to run production auth tests')
      return
    }
    const { token, context } = await loginAndGetToken(browser)
    await use(token)
    await context.close()
  },

  authedPage: async ({ browser, authToken }, use) => {
    const context = await browser.newContext({ ignoreHTTPSErrors: true })
    const page = await context.newPage()
    await page.addInitScript((token) => {
      localStorage.setItem('token', token)
      localStorage.setItem(
        'user',
        JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }),
      )
    }, authToken)
    await use(page)
    await context.close()
  },
})

export { expect }
