import type { Browser, BrowserContext, Page } from '@playwright/test'
import { solveCaptcha, submitLogin } from './captcha'

export const PROD_USER = 'admin'
/** Production admin password. MUST be provided via env; never hardcode. */
export const PROD_PASSWORD = process.env.ADMIN_PASSWORD || ''
/** Admin's DNS subdomain id (the 12-char random label under godnslog.com). */
export const PROD_SHORT_ID = process.env.E2E_PROD_SHORT_ID || 'mc5urjgapz1w'
export const BASE_URL = process.env.E2E_PROD_URL || 'https://www.godnslog.com'

export interface LoginResult {
  token: string
  status: number
  context: BrowserContext
  page: Page
}

/** Perform a real login (slide captcha + credentials) and return the JWT. */
export async function loginAndGetToken(browser: Browser): Promise<LoginResult> {
  const context = await browser.newContext({ ignoreHTTPSErrors: true })
  const page = await context.newPage()

  let status = 0
  page.on('response', async (res) => {
    if (res.url().includes('/auth/login')) {
      status = res.status()
    }
  })

  await solveCaptcha(page, BASE_URL + '/login')
  await submitLogin(page, PROD_USER, PROD_PASSWORD)
  await page.waitForTimeout(2500)

  const token = await page.evaluate(() => localStorage.getItem('token') || '')
  if (status !== 200 || !token) {
    await context.close()
    throw new Error(`production login failed (HTTP ${status}) — set ADMIN_PASSWORD if not already`)
  }
  return { token, status, context, page }
}

/** Trigger an HTTP hit to the production log endpoint (dual-writes a v2
 *  Interaction). Runs inside the page so the browser context's
 *  ignoreHTTPSErrors applies and the request is same-origin. */
export async function triggerHttpHit(page: Page, token: string): Promise<number> {
  const path = `/log/${PROD_SHORT_ID}/${token}`
  return page.evaluate(async (u) => {
    const r = await fetch(u)
    return r.status
  }, path)
}
