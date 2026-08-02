import type { Page } from '@playwright/test'

const IMG_WIDTH = 300

export interface CaptchaData {
  captcha_id: string
  image_base64: string
  thumb_base64: string
  block_dx: number
  block_dy: number
  block_width: number
  block_height: number
}

/** Detect the puzzle-hole X on the background image by finding the darkest
 *  ~60px vertical run (the shadow), then subtracting the ~4px pad offset. */
export async function findHoleX(page: Page, imageBase64: string): Promise<{ predX: number; runWidth: number }> {
  return page.evaluate(async (imageBase64) => {
    const img = new Image()
    img.src = imageBase64
    await img.decode()
    const W = img.naturalWidth
    const H = img.naturalHeight
    const cv = document.createElement('canvas')
    cv.width = W
    cv.height = H
    const ctx = cv.getContext('2d')!
    ctx.drawImage(img, 0, 0)
    const data = ctx.getImageData(0, 0, W, H).data

    const lum = new Float32Array(W * H)
    for (let i = 0; i < W * H; i++) {
      lum[i] = 0.299 * data[i * 4] + 0.587 * data[i * 4 + 1] + 0.114 * data[i * 4 + 2]
    }
    const sorted = Array.from(lum).sort((a, b) => a - b)
    const thr = sorted[Math.floor(W * H * 0.05)]
    const colcount = new Array(W).fill(0)
    for (let y = 0; y < H; y++) {
      for (let x = 0; x < W; x++) {
        if (lum[y * W + x] < thr) colcount[x]++
      }
    }
    let best = { n: 0, xs: 0, xe: 0 }
    let runStart: number | null = null
    for (let x = 0; x <= W; x++) {
      const c = x < W ? colcount[x] : 0
      if (c > 20) {
        if (runStart === null) runStart = x
      } else if (runStart !== null) {
        const n = x - runStart
        if (n > best.n) best = { n, xs: runStart, xe: x - 1 }
        runStart = null
      }
    }
    // after aligning overlay/shadow pads, shadow left edge is ~4px in from block.X
    return { predX: best.xs - 4, runWidth: best.n }
  }, imageBase64)
}

/** Solve the slide captcha. Navigates to loginUrl (listener attached BEFORE the
 *  navigation so the /auth/captcha response is captured), then drags the tile
 *  onto the detected hole. Returns the captcha response data + drag distance. */
export async function solveCaptcha(page: Page, loginUrl: string): Promise<{ captchaResp: CaptchaData | null; targetDisplay: number }> {
  let captchaResp: CaptchaData | null = null
  page.on('response', async (res) => {
    if (res.url().includes('/auth/captcha')) {
      try {
        const j = (await res.json()) as { data?: CaptchaData }
        if (j?.data) captchaResp = j.data
      } catch {}
    }
  })

  await page.goto(loginUrl, { waitUntil: 'networkidle' })
  await page.waitForFunction(() => {
    const img = document.querySelector('img[alt="captcha background"]')
    return img && img.complete && img.naturalWidth > 0
  })
  await page.waitForTimeout(400)

  if (!captchaResp) throw new Error('no captcha response captured')
  const blockDX = captchaResp.block_dx ?? 0
  const dims = await page.evaluate(() => ({
    imgW: (document.querySelector('img[alt="captcha background"]') as HTMLImageElement | null)?.clientWidth ?? 0,
  }))
  const scale = dims.imgW / IMG_WIDTH
  const { predX } = await findHoleX(page, captchaResp.image_base64)
  const targetDisplay = Math.round((predX - blockDX) * scale)

  const track = page.locator('.h-12')
  await track.scrollIntoViewIfNeeded()
  const tb = await track.boundingBox()
  if (!tb) throw new Error('captcha track has no bounding box')
  const sx = tb.x + 24
  const sy = tb.y + tb.height / 2
  await page.mouse.move(sx, sy)
  await page.mouse.down()
  await page.mouse.move(sx + targetDisplay, sy, { steps: 14 })
  await page.mouse.up()
  await page.waitForTimeout(300)

  return { captchaResp, targetDisplay }
}

/** Submit the login form with the given credentials on a captcha-solved page. */
export async function submitLogin(page: Page, user: string, pass: string): Promise<void> {
  await page.fill('#username', user)
  await page.fill('#password', pass)
  await page.evaluate(() => {
    const btn = document.querySelector('form button[type="submit"]') as HTMLButtonElement | null
    if (btn && !btn.disabled) btn.click()
  })
}
