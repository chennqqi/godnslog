import { test, expect, type Page } from '@playwright/test'

const mockAgentRun = {
  id: 'agent-run-1',
  agent_id: 'agent-123',
  operator_id: 'operator-1',
  case_id: 'case-1',
  payload_id: 'payload-1',
  target: 'http://example.com',
  title: 'Test Agent Run',
  status: 'running',
  started_at: '2026-05-24T00:00:00Z',
  ended_at: null,
  created_at: '2026-05-24T00:00:00Z',
  updated_at: '2026-05-24T00:00:00Z',
  interaction_count: 5,
  last_interaction_at: '2026-05-24T00:05:00Z',
  operations: [
    {
      id: 'op-1',
      agent_run_id: 'agent-run-1',
      agent_id: 'agent-123',
      action: 'create_oast_probe',
      risk_level: 'medium',
      request: '{"title":"Test","template":"ssrf-basic"}',
      result: '{"success":true,"case_id":"case-1","payload_id":"payload-1"}',
      error: '',
      started_at: '2026-05-24T00:00:00Z',
      ended_at: '2026-05-24T00:01:00Z',
      created_at: '2026-05-24T00:00:00Z',
    },
    {
      id: 'op-2',
      agent_run_id: 'agent-run-1',
      agent_id: 'agent-123',
      action: 'wait_for_interaction',
      risk_level: 'low',
      request: '{"token":"tok-abc","timeout":30}',
      result: '{"success":true,"interactions":[]}',
      error: '',
      started_at: '2026-05-24T00:01:00Z',
      ended_at: '2026-05-24T00:02:00Z',
      created_at: '2026-05-24T00:01:00Z',
    },
  ],
  case_url: '/dashboard/cases/case-1',
  payload_url: '/dashboard/payloads/payload-1',
  interactions_url: '/api/v2/interactions?payload_id=payload-1',
  evidence_url: '/dashboard/evidence?payload_id=payload-1',
}

test.describe('Agent Runs', () => {
  test.beforeEach(async ({ context }) => {
    // Set auth token in localStorage before page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'test-token')
    })
  })

  test.beforeEach(async ({ page }) => {
    // Mock auth endpoint to bypass login
    await page.route('**/api/v2/auth/info', route => {
      route.fulfill({
        json: {
          code: 0,
          data: {
            id: 'user-1',
            username: 'test-user',
          },
        },
      })
    })

    // Mock agent runs API
    await page.route('**/api/v2/agent-runs**', route => {
      const url = new URL(route.request().url())
      const agentId = url.searchParams.get('agent_id')
      const status = url.searchParams.get('status')

      if (route.request().method() === 'GET') {
        // Return filtered results based on query params
        let items = [mockAgentRun]
        if (agentId && agentId !== 'agent-123') {
          items = []
        }
        if (status && status !== 'running') {
          items = []
        }

        route.fulfill({
          json: {
            code: 0,
            data: {
              items: items,
              total: items.length,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        })
      }
    })

    // Mock agent run detail API
    await page.route('**/api/v2/agent-runs/agent-run-1**', route => {
      if (route.request().method() === 'GET') {
        route.fulfill({
          json: {
            code: 0,
            data: {
              data: mockAgentRun,
            },
          },
        })
      } else if (route.request().method() === 'PUT') {
        route.fulfill({
          json: {
            code: 0,
            data: {
              data: { ...mockAgentRun, status: 'completed' },
            },
          },
        })
      }
    })

    // Mock agent run operations API
    await page.route('**/api/v2/agent-runs/agent-run-1/operations**', route => {
      if (route.request().method() === 'POST') {
        route.fulfill({
          json: {
            code: 0,
            data: {
              id: 'op-3',
              action: 'test_action',
            },
          },
        })
      }
    })
  })

  test('should display agent runs list with API call and filter query', async ({ page }) => {
    // Set up request listener before navigation
    const apiCalls: { method: string; url: string }[] = []
    page.on('request', request => {
      if (request.url().includes('/api/v2/agent-runs')) {
        apiCalls.push({ method: request.method(), url: request.url() })
      }
    })

    await page.goto('/dashboard/agent-runs')
    await page.waitForLoadState('networkidle')

    // Check page title
    await expect(page.getByRole('heading', { name: 'Agent Runs' })).toBeVisible()

    // Check agent run data is displayed
    await expect(page.getByText('Test Agent Run')).toBeVisible()
    await expect(page.getByText('agent-123')).toBeVisible()
    await expect(page.getByText('5 interactions')).toBeVisible()
    await expect(page.getByText('2 operations')).toBeVisible()

    // Verify API call was made
    expect(apiCalls.length).toBeGreaterThan(0)
    expect(apiCalls[0].method).toBe('GET')

    // Test filter by status
    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: 'Completed' }).click()
    await page.getByRole('button', { name: 'Apply Filters' }).click()
    await page.waitForLoadState('networkidle')

    // Verify filter was applied (data should be empty since status=completed doesn't match running)
    await expect(page.getByText('No agent runs found')).toBeVisible()
  })

  test('should display agent run detail with operations timeline and backlinks', async ({ page }) => {
    await page.goto('/dashboard/agent-runs/agent-run-1')
    await page.waitForLoadState('networkidle')

    // Check basic info
    await expect(page.getByText('Test Agent Run')).toBeVisible()
    await expect(page.getByText('agent-123')).toBeVisible()
    await expect(page.getByText('operator-1')).toBeVisible()
    await expect(page.getByText('http://example.com')).toBeVisible()
    await expect(page.getByText('交互数')).toBeVisible() // interaction count label
    await expect(page.getByText('5', { exact: true })).toBeVisible() // interaction count value

    // Check operations timeline
    await expect(page.getByText('操作历史 (2)')).toBeVisible()
    await expect(page.getByText('create_oast_probe')).toBeVisible()
    await expect(page.getByText('wait_for_interaction')).toBeVisible()
    await expect(page.getByText('medium', { exact: true })).toBeVisible()
    await expect(page.getByText('low', { exact: true })).toBeVisible()

    // Check quick links (Interactions/Evidence backlinks)
    const interactionsLink = page.getByText('查看交互记录')
    await expect(interactionsLink).toBeVisible()
    const interactionsHref = await interactionsLink.getAttribute('href')
    expect(interactionsHref).toContain('/api/v2/interactions')
    expect(interactionsHref).toContain('payload_id=payload-1')

    const evidenceLink = page.getByText('查看证据')
    await expect(evidenceLink).toBeVisible()
    const evidenceHref = await evidenceLink.getAttribute('href')
    expect(evidenceHref).toContain('/dashboard/evidence')
    expect(evidenceHref).toContain('payload_id=payload-1')

    // Check case/payload links
    const caseLink = page.getByRole('link', { name: 'case-1' })
    await expect(caseLink).toBeVisible()
    const caseHref = await caseLink.getAttribute('href')
    expect(caseHref).toContain('/dashboard/cases/case-1')

    const payloadLink = page.getByRole('link', { name: 'payload-1' })
    await expect(payloadLink).toBeVisible()
    const payloadHref = await payloadLink.getAttribute('href')
    expect(payloadHref).toContain('/dashboard/payloads/payload-1')
  })

  test('should update agent run status with API call', async ({ page }) => {
    await page.goto('/dashboard/agent-runs/agent-run-1')
    await page.waitForLoadState('networkidle')

    let putRequestCount = 0
    let putRequestUrl = ''
    page.on('request', request => {
      if (request.url().includes('/agent-runs/agent-run-1/status') && request.method() === 'PUT') {
        putRequestCount++
        putRequestUrl = request.url()
      }
    })

    await page.getByRole('button', { name: 'Completed' }).click()

    await expect(putRequestCount).toBeGreaterThan(0)
    expect(putRequestUrl).toContain('/agent-runs/agent-run-1/status')
  })
})
