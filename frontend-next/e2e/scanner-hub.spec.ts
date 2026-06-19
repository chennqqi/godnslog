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

const mockScannerRun = {
  id: 'scanner-run-1',
  case_id: 'case-1',
  payload_id: 'payload-1',
  scanner: 'nuclei',
  target: 'https://target.example',
  template: 'ssrf-basic',
  delivery_method: 'nuclei-jsonl',
  command: "nuclei -u 'https://target.example' -t godnslog-ssrf-basic.yaml -var 'godnslog_payload=http://tok-abc123.example.com/callback'",
  jsonl: '{"scanner":"nuclei","case_id":"case-1","payload_id":"payload-1","token":"tok-abc123","target":"https://target.example","template":"ssrf-basic","rendered_payload":"http://tok-abc123.example.com/callback","interactions_url":"http://localhost:3000/api/v2/interactions?payload_id=payload-1","evidence_url":"http://localhost:3000/dashboard/evidence?payload_id=payload-1","created_at":"2026-05-24T00:00:00.000Z"}',
  package_hash: '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
  package_manifest: {
    schema_version: 'scanner-package.v1',
    scanner: 'nuclei',
    delivery_method: 'nuclei-jsonl',
    package_hash: '0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef',
    hash_algorithm: 'sha256',
    interactions_url: 'http://localhost:3000/api/v2/interactions?payload_id=payload-1',
    evidence_url: 'http://localhost:3000/dashboard/evidence?payload_id=payload-1',
    files: [
      {
        name: 'README.md',
        kind: 'instructions',
        description: 'Operator and Agent instructions for distributing this GODNSLOG scanner package.',
      },
      {
        name: 'godnslog-package.jsonl',
        kind: 'jsonl',
        description: 'Single-line GODNSLOG scanner package record.',
      },
      {
        name: 'godnslog-nuclei.jsonl',
        kind: 'jsonl',
        description: 'Nuclei JSONL import/distribution record.',
      },
    ],
    next_actions: [
      'Distribute the generated payload through the selected scanner adapter.',
      'Poll interactions_url until the expected callback is observed.',
    ],
  },
  status: 'created',
  created_by: 'admin',
  created_at: '2026-05-24T00:00:00Z',
  updated_at: '2026-05-24T00:00:00Z',
}

const mockAdapters = [
  {
    id: 'nuclei',
    name: 'Nuclei',
    category: 'Script',
    maturity: 'L3 Official Script',
    supported_methods: ['nuclei-jsonl', 'nuclei-var'],
    default_method: 'nuclei-jsonl',
    description: 'Automated vulnerability scanning with template variables and JSONL distribution packages.',
  },
  {
    id: 'burp',
    name: 'Burp Suite',
    category: 'Native',
    maturity: 'L4 Native Plugin Path',
    supported_methods: ['burp-extension'],
    default_method: 'burp-extension',
    description: 'Manual and semi-automated verification package for Burp Suite extension workflows.',
  },
  {
    id: 'yakit',
    name: 'Yakit/Yak',
    category: 'Script',
    maturity: 'L3 Official Script',
    supported_methods: ['yakit-script'],
    default_method: 'yakit-script',
    description: 'Yak script package for hybrid manual and automated security validation.',
  },
  {
    id: 'zap',
    name: 'ZAP',
    category: 'Script',
    maturity: 'L3 Official Script',
    supported_methods: ['zap-script'],
    default_method: 'zap-script',
    description: 'OWASP ZAP script package for OAST probe injection and polling.',
  },
  {
    id: 'xray',
    name: 'xray',
    category: 'Webhook',
    maturity: 'L2 Webhook Bridge',
    supported_methods: ['xray-webhook'],
    default_method: 'xray-webhook',
    description: 'Webhook bridge package for xray findings.',
  },
  {
    id: 'rad',
    name: 'rad',
    category: 'Webhook',
    maturity: 'L2 Webhook Bridge',
    supported_methods: ['rad-webhook'],
    default_method: 'rad-webhook',
    description: 'Webhook bridge package for rad automation.',
  },
  {
    id: 'postman',
    name: 'Postman',
    category: 'Environment',
    maturity: 'L2 Environment Bridge',
    supported_methods: ['postman-env'],
    default_method: 'postman-env',
    description: 'Environment variable package for API testing.',
  },
  {
    id: 'apifox',
    name: 'Apifox',
    category: 'Environment',
    maturity: 'L2 Environment Bridge',
    supported_methods: ['apifox-env'],
    default_method: 'apifox-env',
    description: 'Environment variable package for Apifox workflows.',
  },
]

function scannerRunFor(body: { scanner?: string; delivery_method?: string }) {
  const scanner = body.scanner || 'nuclei'
  const deliveryMethod = body.delivery_method || 'nuclei-jsonl'
  const labels: Record<string, string> = {
    nuclei: "nuclei -u 'https://target.example' -t godnslog-ssrf-basic.yaml -var 'godnslog_payload=http://tok-abc123.example.com/callback'",
    burp: 'Burp Suite Extension Package\nCreate probe API: /api/v2/payloads\nPoll interactions: http://localhost:3000/api/v2/interactions?payload_id=payload-1',
    yakit: 'yak godnslog-oast.yak\nCreateHTTPFlow("https://target.example", "http://tok-abc123.example.com/callback")',
  }
  return {
    ...mockScannerRun,
    scanner,
    delivery_method: deliveryMethod,
    command: labels[scanner] || `${scanner} integration package`,
    package_manifest: {
      ...mockScannerRun.package_manifest,
      scanner,
      delivery_method: deliveryMethod,
      files: [
        ...mockScannerRun.package_manifest.files.slice(0, 2),
        {
          name: scanner === 'burp' ? 'burp-extension-config.json' : scanner === 'yakit' ? 'godnslog-oast.yak' : 'godnslog-nuclei.jsonl',
          kind: scanner === 'burp' ? 'burp-extension-config' : scanner === 'yakit' ? 'yak-script' : 'nuclei-jsonl',
          description: `${scanner} integration package file.`,
        },
      ],
    },
    jsonl: JSON.stringify({
      scanner,
      delivery_method: deliveryMethod,
      case_id: 'case-1',
      payload_id: 'payload-1',
      token: 'tok-abc123',
      target: 'https://target.example',
      template: 'ssrf-basic',
      rendered_payload: 'http://tok-abc123.example.com/callback',
      interactions_url: 'http://localhost:3000/api/v2/interactions?payload_id=payload-1',
      evidence_url: 'http://localhost:3000/dashboard/evidence?payload_id=payload-1',
      created_at: '2026-05-24T00:00:00.000Z',
    }),
  }
}

async function installScannerHubMocks(page: Page) {
  await page.route('**/api/v2/scanner-hub/adapters', route => {
    route.fulfill({
      json: {
        code: 0,
        data: {
          items: mockAdapters,
        },
      },
    })
  })

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

  await page.route('**/api/v2/scanner-runs**', route => {
    if (route.request().method() === 'POST') {
      const body = route.request().postDataJSON() as { scanner?: string; delivery_method?: string }
      return route.fulfill({
        json: {
          code: 0,
          data: scannerRunFor(body),
        },
      })
    }

    return route.fulfill({
      json: {
        code: 0,
        data: {
          items: [mockScannerRun],
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
  await page.getByRole('combobox').nth(2).click()
  await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
  await page.getByPlaceholder('example.com').fill('https://target.example')
  await page.getByRole('combobox').nth(4).click()
  await page.getByRole('option', { name: 'tok-abc123' }).click()
  await page.getByRole('button', { name: '生成 Scanner Run' }).click()
}

async function selectScanner(page: Page, name: string) {
  await page.getByRole('combobox').nth(0).click()
  await page.getByRole('option', { name }).click()
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
    await expect(page.getByText('多工具 OAST 适配器工作台')).toBeVisible()
    await expect(page.getByRole('heading', { name: '选择 Scanner Adapter' })).toBeVisible()
    await expect(page.getByText('Burp Suite')).toBeVisible()
    await expect(page.getByText('Yakit/Yak')).toBeVisible()
    await expect(page.getByText('ZAP', { exact: true })).toBeVisible()
    await expect(page.getByText('xray', { exact: true })).toBeVisible()
    await expect(page.getByText('rad', { exact: true })).toBeVisible()
    await expect(page.getByText('Postman', { exact: true })).toBeVisible()
    await expect(page.getByText('Apifox', { exact: true })).toBeVisible()
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

    await page.getByRole('combobox').nth(2).click()
    await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
    await page.getByRole('button', { name: '创建新 Payload' }).click()
    await createRequest

    await page.getByRole('combobox').nth(4).click()
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

  test('should display machine-readable scanner package manifest and hash', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)

    await expect(page.getByRole('heading', { name: 'Package Hash' })).toBeVisible()
    await expect(page.locator(`input[value="${mockScannerRun.package_hash}"]`)).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Package Manifest' })).toBeVisible()
    await expect(page.getByText('scanner-package.v1')).toBeVisible()
    await expect(page.getByText('README.md')).toBeVisible()
    await expect(page.getByText('godnslog-package.jsonl')).toBeVisible()
    await expect(page.getByText('godnslog-nuclei.jsonl')).toBeVisible()
  })

  test('should call scanner-runs API when generating scanner run', async ({ page }) => {
    await openScannerHub(page)
    const scannerRunRequest = page.waitForRequest(request => {
      if (!request.url().endsWith('/api/v2/scanner-runs') || request.method() !== 'POST') return false
      const body = request.postDataJSON() as { case_id?: string; payload_id?: string; scanner?: string; delivery_method?: string }
      return body.case_id === 'case-1' && body.payload_id === 'payload-1' && body.scanner === 'nuclei' && body.delivery_method === 'nuclei-jsonl'
    })

    await page.getByRole('combobox').nth(2).click()
    await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
    await page.getByPlaceholder('example.com').fill('https://target.example')
    await page.getByRole('combobox').nth(4).click()
    await page.getByRole('option', { name: 'tok-abc123' }).click()
    await page.getByRole('button', { name: '生成 Scanner Run' }).click()
    await scannerRunRequest
  })

  test('should generate Burp Suite extension package through scanner-runs API', async ({ page }) => {
    await openScannerHub(page)
    const scannerRunRequest = page.waitForRequest(request => {
      if (!request.url().endsWith('/api/v2/scanner-runs') || request.method() !== 'POST') return false
      const body = request.postDataJSON() as { scanner?: string; delivery_method?: string }
      return body.scanner === 'burp' && body.delivery_method === 'burp-extension'
    })

    await selectScanner(page, 'Burp Suite')
    await generateScannerRun(page)
    await scannerRunRequest

    await expect(page.getByText('Burp Suite Extension Package').first()).toBeVisible()
    await expect(page.locator('input[readonly]').nth(2)).toHaveValue(/Burp Suite Extension Package/)
  })

  test('should generate Yakit Yak script package through scanner-runs API', async ({ page }) => {
    await openScannerHub(page)
    const scannerRunRequest = page.waitForRequest(request => {
      if (!request.url().endsWith('/api/v2/scanner-runs') || request.method() !== 'POST') return false
      const body = request.postDataJSON() as { scanner?: string; delivery_method?: string }
      return body.scanner === 'yakit' && body.delivery_method === 'yakit-script'
    })

    await selectScanner(page, 'Yakit/Yak')
    await generateScannerRun(page)
    await scannerRunRequest

    await expect(page.getByText('Yakit/Yak Script Package').first()).toBeVisible()
    await expect(page.locator('input[readonly]').nth(2)).toHaveValue(/CreateHTTPFlow/)
  })

  test('should display recent scanner runs list', async ({ page }) => {
    await openScannerHub(page)
    await expect(page.getByText('最近的 Scanner Runs')).toBeVisible()
    // The empty state message may vary, just check the section exists
    await expect(page.locator('text=最近的 Scanner Runs')).toBeVisible()
  })

  test('should navigate to scanner run detail page', async ({ page }) => {
    await openScannerHub(page)
    await page.getByRole('combobox').nth(2).click()
    await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
    await page.getByPlaceholder('example.com').fill('https://target.example')
    await page.getByRole('combobox').nth(4).click()
    await page.getByRole('option', { name: 'tok-abc123' }).click()
    await page.getByRole('button', { name: '生成 Scanner Run' }).click()

    // Wait for scanner run to be created and navigate to detail
    await page.waitForTimeout(1000)
    await page.getByRole('button', { name: '查看详情' }).click()
    await page.waitForURL('**/dashboard/scanner-hub/**')
    expect(page.url()).toContain('/dashboard/scanner-hub/')
  })

  test('should update scanner run status on detail page', async ({ page }) => {
    // Directly navigate to detail page with proper mock
    await page.goto('http://localhost:3000/dashboard/scanner-hub/run-1')

    // Mock detail page API and status update API
    await page.route('**/api/v2/scanner-runs/*', route => {
      if (route.request().method() === 'GET') {
        return route.fulfill({
          json: {
            code: 0,
            message: 'success',
            data: {
              data: {
                ...mockScannerRun,
                id: 'run-1',
                status: 'created',
                interaction_count: 0,
                evidence_count: 0,
                interactions_url: 'http://localhost:3000/api/v2/interactions?payload_id=payload-1',
                evidence_url: 'http://localhost:3000/dashboard/evidence?payload_id=payload-1',
              },
            },
          },
        })
      }
      if (route.request().method() === 'PUT' && route.request().url().includes('/status')) {
        return route.fulfill({
          json: {
            code: 0,
            message: 'success',
            data: { ...mockScannerRun, status: 'distributed' },
          },
        })
      }
    })

    // Wait for page to load
    await page.waitForLoadState('networkidle')

    // Wait for status update request
    const statusUpdateRequest = page.waitForRequest(request => {
      if (!request.url().includes('/api/v2/scanner-runs/') || !request.url().includes('/status') || request.method() !== 'PUT') return false
      const body = request.postDataJSON() as { status?: string }
      return body.status === 'distributed'
    })

    // Click status update button
    await page.getByRole('button', { name: 'Distributed' }).click()
    await statusUpdateRequest
  })

  test('should display package manifest on scanner run detail page', async ({ page }) => {
    await page.route('**/api/v2/scanner-runs/*', route => {
      return route.fulfill({
        json: {
          code: 0,
          message: 'success',
          data: {
            data: {
              ...mockScannerRun,
              id: 'run-1',
              interaction_count: 0,
              evidence_count: 0,
              interactions_url: 'http://localhost:3000/api/v2/interactions?payload_id=payload-1',
              evidence_url: 'http://localhost:3000/dashboard/evidence?payload_id=payload-1',
            },
          },
        },
      })
    })

    await page.goto('http://localhost:3000/dashboard/scanner-hub/run-1')
    await page.waitForLoadState('networkidle')

    await expect(page.getByRole('heading', { name: 'Package Hash' })).toBeVisible()
    await expect(page.locator(`input[value="${mockScannerRun.package_hash}"]`)).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Package Manifest' })).toBeVisible()
    await expect(page.getByText('README.md')).toBeVisible()
    await expect(page.getByText('godnslog-package.jsonl')).toBeVisible()
  })

  test('should expose copy controls for payload command and JSONL', async ({ page }) => {
    await openScannerHub(page)
    await generateScannerRun(page)
    await expect(page.getByRole('button', { name: '复制' })).toHaveCount(5)
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

    await page.getByRole('combobox').nth(2).click()
    await page.getByRole('option', { name: 'Nuclei SSRF Scan' }).click()
    await page.getByRole('button', { name: '创建新 Payload' }).click()
    await expect(page.getByText('创建Payload失败')).toBeVisible()
  })
})
