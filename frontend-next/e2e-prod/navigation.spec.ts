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

      await authedPage.goto(BASE_URL + route, { waitUntil: 'networkidle', timeout: 20000 })
      await authedPage.waitForTimeout(1200)

      const body = await authedPage.evaluate(() => document.body.innerText)
      const boundary = body.includes('Something went wrong')

      authedPage.removeListener('response', onResponse)
      authedPage.removeListener('pageerror', onPageError)

      expect(boundary).toBe(false)
      expect(bad).toEqual([])
      expect(jsErrors).toEqual([])
    })
  }
})
