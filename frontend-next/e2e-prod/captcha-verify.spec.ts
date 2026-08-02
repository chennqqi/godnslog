import { test, expect } from './fixtures'
import { PROD_PASSWORD, PROD_USER, BASE_URL } from './helpers/auth'
import { findHoleX } from './helpers/captcha'

// Captcha tolerance verification against production. The server accepts a drag
// within +-captchaTolerance (5) image px of the hole; deliberately off-target
// drags must be rejected.
const CASES: { label: string; jitter: number; expect: number }[] = [
  { label: 'perfect', jitter: 0, expect: 200 },
  { label: 'human +4px', jitter: 4, expect: 200 },
  { label: 'human -4px', jitter: -4, expect: 200 },
  { label: 'wrong +40px', jitter: 40, expect: 400 },
]

test.describe('Production Captcha Precision', () => {
  for (const { label, jitter, expect: expectedStatus } of CASES) {
    test(`${label} (jitter ${jitter}px) -> HTTP ${expectedStatus}`, async ({ browser }) => {
      const context = await browser.newContext({ ignoreHTTPSErrors: true })
      const page = await context.newPage()

      let captchaResp: { image_base64: string; block_dx: number } | null = null
      let loginStatus = 0
      page.on('response', async (res) => {
        const url = res.url()
        if (url.includes('/auth/captcha')) {
          try {
            const j = (await res.json()) as { data?: { image_base64: string; block_dx: number } }
            if (j?.data) captchaResp = j.data
          } catch {}
        } else if (url.includes('/auth/login')) {
          loginStatus = res.status()
        }
      })

      await page.goto(BASE_URL + '/login', { waitUntil: 'networkidle' })
      await page.waitForFunction(() => {
        const img = document.querySelector('img[alt="captcha background"]')
        return img && img.complete && img.naturalWidth > 0
      })
      await page.waitForTimeout(400)
      if (!captchaResp) throw new Error('no captcha response captured')

      const imgW = await page.evaluate(
        () => (document.querySelector('img[alt="captcha background"]') as HTMLImageElement | null)?.clientWidth ?? 0,
      )
      const scale = imgW / 300
      const blockDX = captchaResp.block_dx ?? 0
      const { predX } = await findHoleX(page, captchaResp.image_base64)
      const targetDisplay = Math.round((predX - blockDX + jitter) * scale)

      await page.fill('#username', PROD_USER)
      await page.fill('#password', PROD_PASSWORD)

      const track = page.locator('.h-12')
      await track.scrollIntoViewIfNeeded()
      const tb = await track.boundingBox()
      if (!tb) throw new Error('no track bounding box')
      const sx = tb.x + 24
      const sy = tb.y + tb.height / 2
      await page.mouse.move(sx, sy)
      await page.mouse.down()
      await page.mouse.move(sx + targetDisplay, sy, { steps: 14 })
      await page.mouse.up()
      await page.waitForTimeout(300)

      await page.evaluate(() => {
        const btn = document.querySelector('form button[type="submit"]') as HTMLButtonElement | null
        if (btn && !btn.disabled) btn.click()
      })
      await page.waitForTimeout(1500)

      expect(loginStatus).toBe(expectedStatus)
      await context.close()
    })
  }
})
