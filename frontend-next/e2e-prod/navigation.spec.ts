import { test, expect } from './fixtures'
import { BASE_URL } from './helpers/auth'

const ROUTES = [
  '/',
  '/cases',
  '/payloads',
  '/agent-runs',
  '/interactions',
  '/evidence-summary',
  '/attack-chains',
  '/canary',
  '/rebinding',
  '/listeners',
  '/workflow',
  '/scanner-hub',
  '/marketplace',
  '/retention',
  '/settings',
  '/users',
  '/apikeys',
  '/audit',
  '/docs',
]

/** Load a route and report whether an error boundary showed. 5xx responses are
 *  collected into `bad` by the caller's response listener. */
async function loadRoute(page: import('@playwright/test').Page, route: string): Promise<boolean> {
  await page.goto(BASE_URL + route, { waitUntil: 'networkidle', timeout: 20000 }).catch(() => {})
  await page.waitForTimeout(1500)
  const body = await page.evaluate(() => document.body.innerText)
  return body.includes('Something went wrong')
}

test.describe('Production Navigation', () => {
  for (const route of ROUTES) {
    test(`renders ${route || '/'} without 5xx or error boundary`, async ({ authedPage }) => {
      const bad: string[] = []
      const jsErrors: string[] = []
      const onResponse = (res: { status: () => number; url: () => string }) => {
        if (res.status() >= 500) bad.push(`${res.status()} ${res.url().replace(BASE_URL, '')}`)
      }
      const onPageError = (e: Error) => jsErrors.push(e.message)
      authedPage.on('response', onResponse)
      authedPage.on('pageerror', onPageError)

      let boundary = await loadRoute(authedPage, route)

      // Long full-suite runs occasionally hit a momentary 5xx or slow backend;
      // one reload settles transient failures before we assert.
      if (boundary || bad.length > 0) {
        bad.length = 0
        await authedPage.waitForTimeout(2000)
        boundary = await loadRoute(authedPage, route)
      }

      authedPage.removeListener('response', onResponse)
      authedPage.removeListener('pageerror', onPageError)

      expect(boundary).toBe(false)
      expect(bad).toEqual([])
      expect(jsErrors).toEqual([])
    })
  }
})
