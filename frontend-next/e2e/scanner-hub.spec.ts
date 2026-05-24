import { test, expect, type Page } from '@playwright/test'

const mockCase = {
  id: 'case-1',
  title: 'Nuclei SSRF Scan',
  description: '',
  target: 'https://target.example',
  status: 'active',
  tags: [],
  created_by: 'admin',
  created_at: '2026-05-24T00:00:00Z',
  updated_at: '2026-05-24T00:00:00Z',
}

const mockPayload = {
  id: 'payload-1',
  case_id: 'case-1',
  token: 'tok-abc123',
  template: 'ssrf-basic',
  rendered_payload: 'http://tok-abc123.example.com/callback',
  variables: {},
  status: 'deployed',
  created_by: 'admin',
  created_at: '2026-05-24T00:00:00Z',
  updated_at: '2026-05-24T00:00:00Z',
}

async function installScannerHubMocks(page: Page) {
  await page.route('**/api/v2/cases**', route => {
    route.fulfill({
      json: {
        code: 0,
        data: {
          items: [mockCase],
          total: 1,
          page: 1,
          page_size: 100,
          total_pages: 1,
        },
      },
    })
  })

  await page.route('**/api/v2/payloads**', route => {
    if (route.request().method() === 'POST') {
      return route.fulfill({
        json: {
          code: 0,
          data: {
            data: mockPayload,
          },
        },
      })
    }

    return route.fulfill({
      json: {
        code: 0,
        data: {
          items: [mockPayload],
          total: 1,
          page: 1,
          page_size: 100,
          total_pages: 1,
        },
      },
    })
  })

  await page.route('**/api/v2/evidence/generate', route => {
    route.fulfill({
      json: {
        code: 0,
        data: {
          evidence: {
            id: 'evidence-1',
            case_id: 'case-1',
            payload_id: 'payload-1',
            evidence_strength: 'high',
            confidence: 90,
            interaction_count: 2,
            unique_sources: 1,
            timeline: [],
            explainability: 'Captured 2 interactions from 1 unique source.',
            generated_at: '2026-05-24T00:00:00Z',
          },
          content: '# Evidence Report',
        },
      },
    })
  })
}

async function openScannerHub(page: Page) {
  await installScannerHubMocks(page)
  await page.goto('/dashboard/scanner-hub')
  await page.waitForLoadState('networkidle')
}

async function generateScannerRun(page: Page) {
  await page.getByRole('combobox').nth(0).click()
  await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
  await page.getByPlaceholder('example.com').fill('https://target.example')
  await page.getByRole('combobox').nth(2).click()
  await page.getByRole('option', { name: 'tok-abc123' }).click()
  await page.getByRole('button', { name: '生成 Scanner Run' }).click()
}

test.describe('Scanner Hub', () => {
  test.beforeEach(async ({ context }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token')
      localStorage.setItem('user', JSON.stringify({
        id: 1,
        username: 'admin',
        email: 'admin@godnslog.com',
        role: 0,
        lang: 'en-US',
      }))
    })
  })

  test('should load scanner hub workspace', async ({ page }) => {
    await openScannerHub(page)
    await expect(page.getByRole('heading', { name: 'Scanner Hub' })).toBeVisible()
    await expect(page.getByText('Nuclei 集成工作台')).toBeVisible()
    await expect(page.getByRole('heading', { name: '选择 Case' })).toBeVisible()
    await expect(page.getByRole('heading', { name: '输入 Target' })).toBeVisible()
  })

  test('should create payload through the unified payload API', async ({ page }) => {
    await openScannerHub(page)
    const createRequest = page.waitForRequest(request => {
      if (!request.url().endsWith('/api/v2/payloads') || request.method() !== 'POST') return false
      const body = request.postDataJSON() as { case_id?: string; template?: string }
      return body.case_id === 'case-1' && body.template === 'ssrf-basic'
    })

    await page.getByRole('combobox').nth(0).click()
    await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
    await page.getByRole('button', { name: '创建新 Payload' }).click()
    await createRequest

    await page.getByRole('combobox').nth(2).click()
    await expect(page.getByRole('option', { name: 'tok-abc123' }).last()).toBeVisible()
  })

  test('should generate nuclei command and JSONL with stable scanner fields', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)

    await expect(page.getByText('Nuclei Command')).toBeVisible()
    await expect(page.locator('input[readonly]').nth(2)).toHaveValue(/nuclei -u 'https:\/\/target\.example'/)
    await expect(page.locator('input[readonly]').nth(2)).toHaveValue(/godnslog_payload=http:\/\/tok-abc123\.example\.com\/callback/)

    const jsonl = await page.locator('textarea').inputValue()
    expect(jsonl).not.toContain('\n')
    const record = JSON.parse(jsonl)
    expect(record).toMatchObject({
      scanner: 'nuclei',
      case_id: 'case-1',
      payload_id: 'payload-1',
      token: 'tok-abc123',
      target: 'https://target.example',
      template: 'ssrf-basic',
      rendered_payload: 'http://tok-abc123.example.com/callback',
      interactions_url: 'http://localhost:3000/api/v2/interactions?payload_id=payload-1',
      evidence_url: 'http://localhost:3000/dashboard/evidence?payload_id=payload-1',
    })
  })

  test('should expose copy controls for payload command and JSONL', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)
    await expect(page.getByRole('button', { name: '复制' })).toHaveCount(4)
    await expect(page.getByText('Rendered Payload')).toBeVisible()
    await expect(page.getByText('JSONL Preview')).toBeVisible()
  })

  test('should show current case and payload scope', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)
    await expect(page.getByText('当前 Scope')).toBeVisible()
    await expect(page.getByText('case-1', { exact: true })).toBeVisible()
    await expect(page.getByText('payload-1', { exact: true })).toBeVisible()
  })

  test('should navigate to payload scoped interactions', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)

    await page.getByRole('button', { name: '查看 Interactions' }).click()
    await page.waitForURL('**/dashboard/interactions?payload_id=payload-1')
    expect(page.url()).toContain('payload_id=payload-1')
  })

  test('should navigate to payload scoped evidence', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)
    await page.getByRole('button', { name: '查看 Evidence' }).click()
    await page.waitForURL('**/dashboard/evidence?payload_id=payload-1')
    expect(page.url()).toContain('payload_id=payload-1')
  })

  test('should keep evidence generation on the unified evidence contract', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)
    const evidenceRequest = page.waitForRequest(request => {
      if (!request.url().includes('/api/v2/evidence/generate')) return false
      const body = request.postDataJSON() as { payload_id?: string; format?: string }
      return body.payload_id === 'payload-1' && body.format === 'markdown'
    })
    await page.getByRole('button', { name: '查看 Evidence' }).click()
    await evidenceRequest
  })

  test('should show validation and API error states', async ({ page }) => {
    await openScannerHub(page)
    await page.getByRole('button', { name: '生成 Scanner Run' }).click()
    await expect(page.getByText('请选择Case、Payload并输入Target')).toBeVisible()

    await page.route('**/api/v2/payloads**', route => {
      if (route.request().method() === 'POST') {
        return route.fulfill({ status: 500, json: { code: 1, message: 'failed' } })
      }
      return route.fulfill({
        json: { code: 0, data: { items: [mockPayload], total: 1, page: 1, page_size: 100, total_pages: 1 } },
      })
    })

    await page.getByRole('combobox').nth(0).click()
    await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
    await page.getByRole('button', { name: '创建新 Payload' }).click()
    await expect(page.getByText('创建Payload失败')).toBeVisible()
  })
})
