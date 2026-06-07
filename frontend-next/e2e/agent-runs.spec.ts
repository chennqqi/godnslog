import { test, expect } from '@playwright/test'

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

    // Mock agent run detail API (use exact path to avoid matching review API)
    await page.route('**/api/v2/agent-runs/agent-run-1', route => {
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

    // Mock agent run review API
    await page.route('**/api/v2/agent-runs/agent-run-1/review**', route => {
      if (route.request().method() === 'GET') {
        const url = new URL(route.request().url())
        const format = url.searchParams.get('format') || 'json'
        
        if (format === 'json') {
          route.fulfill({
            json: {
              code: 0,
              data: {
                data: {
                  id: 'agent-run-1',
                  agent_run: {
                    id: 'agent-run-1',
                    agent_id: 'agent-1',
                    operator_id: 'user-1',
                    case_id: 'case-1',
                    payload_id: 'payload-1',
                    target: 'example.com',
                    title: 'Test Agent Run',
                    status: 'completed',
                    created_at: '2026-05-31T10:00:00Z',
                    updated_at: '2026-05-31T10:00:00Z',
                    interaction_count: 5,
                    last_interaction_at: '2026-05-31T10:00:00Z',
                    operations: [],
                    case_url: '/dashboard/cases/case-1',
                    payload_url: '/dashboard/payloads/payload-1',
                    interactions_url: '/dashboard/interactions?case_id=case-1',
                    evidence_url: '/dashboard/evidence?case_id=case-1',
                  },
                  interaction_summary: {
                    total: 5,
                    dns_count: 3,
                    http_count: 2,
                    unique_sources: 2,
                    last_interaction_at: '2026-05-31T10:00:00Z',
                  },
                  evidence: {
                    id: 'evidence-1',
                    case_id: 'case-1',
                    payload_id: 'payload-1',
                    evidence_strength: 'high',
                    confidence: 85,
                    interaction_count: 5,
                    unique_sources: 2,
                    explainability: 'Captured 5 interactions from 2 unique sources',
                    generated_at: '2026-05-31T10:00:00Z',
                  },
                  audit_refs: [],
                  generated_at: '2026-05-31T10:00:00Z',
                  format: 'json',
                  content: undefined,
                },
              },
            },
          })
        } else if (format === 'markdown') {
          route.fulfill({
            json: {
              code: 0,
              data: {
                data: {
                  id: 'agent-run-1',
                  agent_run: {
                    id: 'agent-run-1',
                    agent_id: 'agent-1',
                    operator_id: 'user-1',
                    case_id: 'case-1',
                    payload_id: 'payload-1',
                    target: 'example.com',
                    title: 'Test Agent Run',
                    status: 'completed',
                    created_at: '2026-05-31T10:00:00Z',
                    updated_at: '2026-05-31T10:00:00Z',
                    interaction_count: 5,
                    last_interaction_at: '2026-05-31T10:00:00Z',
                    operations: [],
                    case_url: '/dashboard/cases/case-1',
                    payload_url: '/dashboard/payloads/payload-1',
                    interactions_url: '/dashboard/interactions?case_id=case-1',
                    evidence_url: '/dashboard/evidence?case_id=case-1',
                  },
                  interaction_summary: {
                    total: 5,
                    dns_count: 3,
                    http_count: 2,
                    unique_sources: 2,
                    last_interaction_at: '2026-05-31T10:00:00Z',
                  },
                  evidence: {
                    id: 'evidence-1',
                    case_id: 'case-1',
                    payload_id: 'payload-1',
                    evidence_strength: 'high',
                    confidence: 85,
                    interaction_count: 5,
                    unique_sources: 2,
                    explainability: 'Captured 5 interactions from 2 unique sources',
                    generated_at: '2026-05-31T10:00:00Z',
                  },
                  audit_refs: [],
                  generated_at: '2026-05-31T10:00:00Z',
                  format: 'markdown',
                  content: '# Agent Run Review\n\n**Evidence Strength**: high\n**Confidence**: 85%\n**Interaction Count**: 5\n\n## Summary\n\nThis agent run captured 5 interactions from 2 unique sources.',
                },
              },
            },
          })
        }
      } else {
        route.continue()
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
    const interactionsLink = page.getByRole('link', { name: '查看交互记录' })
    await expect(interactionsLink).toBeVisible()
    const interactionsHref = await interactionsLink.getAttribute('href')
    expect(interactionsHref).toContain('/api/v2/interactions')
    expect(interactionsHref).toContain('payload_id=payload-1')

    const evidenceLink = page.getByRole('link', { name: '查看证据' })
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

  test('should generate and display review packet with API calls', async ({ page }) => {
    await page.goto('/dashboard/agent-runs/agent-run-1')
    await page.waitForLoadState('networkidle')

    // Check if Review Packet section exists
    await expect(page.getByText('Review Packet')).toBeVisible()

    // Set up request listener for review API
    const reviewRequests: { url: string; format: string }[] = []
    page.on('request', request => {
      if (request.url().includes('/agent-runs/agent-run-1/review') && request.method() === 'GET') {
        const url = new URL(request.url())
        const format = url.searchParams.get('format') || 'json'
        reviewRequests.push({ url: request.url(), format })
      }
    })

    // Click "生成 JSON Review" button and wait for review API request
    const jsonReviewPromise = page.waitForRequest(request => 
      request.url().includes('/agent-runs/agent-run-1/review') && 
      request.url().includes('format=json')
    )
    await page.getByRole('button', { name: '生成 JSON Review' }).click()
    await jsonReviewPromise
    await page.waitForLoadState('networkidle')

    // Verify review API was called with format=json
    expect(reviewRequests.length).toBeGreaterThan(0)
    expect(reviewRequests[0].format).toBe('json')
    expect(reviewRequests[0].url).toContain('format=json')

    // Wait for React state update and UI rendering
    // Use waitForFunction to check if review packet content is in DOM
    await page.waitForFunction(() => {
      const body = document.body
      return body.textContent?.includes('Evidence Strength') && 
             body.textContent?.includes('high') &&
             body.textContent?.includes('Confidence') &&
             body.textContent?.includes('85%')
    }, { timeout: 10000 })
    
    await expect(page.getByText('Evidence Strength')).toBeVisible()
    await expect(page.getByText('high')).toBeVisible()
    await expect(page.getByText('Confidence')).toBeVisible()
    await expect(page.getByText('85%')).toBeVisible()
    await expect(page.getByText('Interaction Count')).toBeVisible()
    // Use locator with context to avoid strict mode violation
    await expect(page.getByText('Interaction Count').locator('..').getByText('5')).toBeVisible()
    await expect(page.getByText('Unique Sources')).toBeVisible()
    await expect(page.getByText('Unique Sources').locator('..').getByText('2')).toBeVisible()

    // Click "生成 Markdown Review" button and wait for review API request
    reviewRequests.length = 0 // clear previous requests
    const markdownReviewPromise = page.waitForRequest(request =>
      request.url().includes('/agent-runs/agent-run-1/review') &&
      request.url().includes('format=markdown')
    )
    await page.getByRole('button', { name: '生成 Markdown Review' }).click()
    await markdownReviewPromise
    await page.waitForLoadState('networkidle')

    // Verify review API was called with format=markdown
    expect(reviewRequests.length).toBeGreaterThan(0)
    expect(reviewRequests[0].format).toBe('markdown')
    expect(reviewRequests[0].url).toContain('format=markdown')

    // Wait for markdown preview to be displayed and verify markdown content rendering
    await page.waitForFunction(() => {
      const body = document.body
      return body.textContent?.includes('Markdown Preview') && 
             body.textContent?.includes('# Agent Run Review')
    }, { timeout: 10000 })
    await expect(page.getByText('Markdown Preview')).toBeVisible()
    await expect(page.getByText('# Agent Run Review')).toBeVisible()
    await expect(page.getByText('**Evidence Strength**: high')).toBeVisible()
    await expect(page.getByText('**Confidence**: 85%')).toBeVisible()
  })

  test('should create follow-up action', async ({ page }) => {
    // Mock followup API
    await page.route('**/api/v2/agent-runs/agent-run-1/followups', async route => {
      if (route.request().method() === 'POST') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            code: 0,
            message: 'success',
            data: {
              agent_run_id: 'agent-run-1',
              operation_id: 'op-followup-1',
              action_type: 'recheck_evidence',
              reason: 'Evidence needs second review',
              review_packet_id: 'agent-run-1',
              operation: {
                id: 'op-followup-1',
                agent_run_id: 'agent-run-1',
                action: 'followup.recheck_evidence',
                risk_level: 'low',
                started_at: new Date().toISOString(),
                created_at: new Date().toISOString(),
              },
              created_at: new Date().toISOString(),
            },
          }),
        })
      } else {
        // GET request for followup history
        await route.fulfill({
          json: {
            code: 0,
            data: {
              data: [
                {
                  operation_id: 'op-followup-1',
                  action_type: 'recheck_evidence',
                  risk_level: 'low',
                  reason: 'Evidence needs second review',
                  review_packet_id: 'agent-run-1',
                  audit_ref_id: 'audit-123',
                  created_at: new Date().toISOString(),
                },
              ],
            },
          },
        })
      }
    })

    // Track if followup was created to modify subsequent agent run detail responses
    let followupCreated = false
    await page.route('**/api/v2/agent-runs/agent-run-1', route => {
      if (route.request().method() === 'GET') {
        if (followupCreated) {
          route.fulfill({
            json: {
              code: 0,
              data: {
                data: {
                  ...mockAgentRun,
                  operations: [
                    ...mockAgentRun.operations,
                    {
                      id: 'op-followup-1',
                      agent_run_id: 'agent-run-1',
                      agent_id: 'agent-123',
                      action: 'followup.recheck_evidence',
                      risk_level: 'low',
                      started_at: new Date().toISOString(),
                      created_at: new Date().toISOString(),
                    },
                  ],
                },
              },
            },
          })
        } else {
          route.fulfill({
            json: {
              code: 0,
              data: {
                data: mockAgentRun,
              },
            },
          })
        }
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

    // Navigate to agent run detail page
    await page.goto('/dashboard/agent-runs/agent-run-1')
    await page.waitForLoadState('networkidle')

    // Generate review first to enable followup button
    const reviewPromise = page.waitForRequest(request =>
      request.url().includes('/agent-runs/agent-run-1/review')
    )
    await page.getByRole('button', { name: '生成 JSON Review' }).click()
    await reviewPromise
    await page.waitForLoadState('networkidle')

    // Wait for review packet to be displayed
    await page.waitForFunction(() => {
      const body = document.body
      return body.textContent?.includes('Evidence Strength')
    }, { timeout: 10000 })

    // Click "创建 Follow-up Action" button
    await page.getByRole('button', { name: '创建 Follow-up Action' }).click()

    // Wait for dialog to open - check for dialog title specifically
    await expect(page.getByRole('heading', { name: '创建 Follow-up Action' })).toBeVisible()

    // Select action type
    await page.getByRole('combobox').click()
    await page.getByRole('option', { name: 'Recheck Evidence' }).click()

    // Enter reason
    await page.getByPlaceholder('请输入原因...').fill('Evidence needs second review')

    // Click create button and wait for API request
    const followupPromise = page.waitForRequest(request =>
      request.url().includes('/agent-runs/agent-run-1/followups')
    )
    await page.getByRole('button', { name: '创建' }).click()
    const followupRequest = await followupPromise

    // Assert POST request body contains correct fields
    const postData = JSON.parse(followupRequest.postData() || '{}')
    expect(postData.action_type).toBe('recheck_evidence')
    expect(postData.reason).toBe('Evidence needs second review')
    expect(postData.review_packet_id).toBe('agent-run-1')

    followupCreated = true
    await page.waitForLoadState('networkidle')

    // Verify dialog is closed
    await expect(page.getByRole('heading', { name: '创建 Follow-up Action' })).not.toBeVisible()

    // Wait for both agent run detail and followup history to refresh
    const agentRunRefreshPromise = page.waitForRequest(request =>
      request.url().includes('/agent-runs/agent-run-1') && request.method() === 'GET'
    )
    const followupHistoryRefreshPromise = page.waitForRequest(request =>
      request.url().includes('/agent-runs/agent-run-1/followups') && request.method() === 'GET'
    )

    await Promise.all([
      agentRunRefreshPromise,
      followupHistoryRefreshPromise,
    ])
    await page.waitForLoadState('networkidle')

    // Verify followup operation appears in timeline
    await expect(page.getByText('followup.recheck_evidence')).toBeVisible()

    // Verify followup history section is refreshed and shows the new followup
    await expect(page.getByText('Follow-up History (1)')).toBeVisible()
    await expect(page.getByText('recheck_evidence', { exact: true })).toBeVisible()
    await expect(page.getByText('Evidence needs second review')).toBeVisible()
  })

  test('should display review queue with summary and filters', async ({ page }) => {
    // Mock review queue API
    await page.route('**/api/v2/agent-runs/review-queue**', route => {
      const url = new URL(route.request().url())
      const reviewState = url.searchParams.get('review_state')
      const evidenceStrength = url.searchParams.get('evidence_strength')

      const mockReviewQueueItem = {
        id: 'agent-run-1',
        agent_id: 'agent-123',
        operator_id: 'operator-1',
        case_id: 'case-1',
        payload_id: 'payload-1',
        target: 'http://example.com',
        title: 'Test Agent Run',
        status: 'completed',
        review_state: 'not_reviewed',
        evidence_strength: 'high',
        interaction_count: 5,
        operation_count: 2,
        followup_count: 0,
        needs_attention: false,
        created_at: '2026-05-24T00:00:00Z',
        updated_at: '2026-05-24T00:00:00Z',
        case_url: '/dashboard/cases/case-1',
        payload_url: '/dashboard/payloads/payload-1',
        evidence_url: '/dashboard/evidence?payload_id=payload-1',
      }

      let items = [mockReviewQueueItem]
      let summary = {
        total: 1,
        not_reviewed: 1,
        reviewed: 0,
        followup_created: 0,
        needs_attention: 0,
      }

      if (reviewState === 'needs_attention') {
        // Return different summary for needs_attention filter
        summary = {
          total: 1,
          not_reviewed: 0,
          reviewed: 0,
          followup_created: 0,
          needs_attention: 1,
        }
        mockReviewQueueItem.review_state = 'needs_attention'
        mockReviewQueueItem.needs_attention = true
        items = [mockReviewQueueItem]
      } else if (reviewState && reviewState !== 'not_reviewed') {
        items = []
        summary = {
          total: 0,
          not_reviewed: 0,
          reviewed: 0,
          followup_created: 0,
          needs_attention: 0,
        }
      }
      if (evidenceStrength && evidenceStrength !== 'high') {
        items = []
        summary = {
          total: 0,
          not_reviewed: 0,
          reviewed: 0,
          followup_created: 0,
          needs_attention: 0,
        }
      }

      route.fulfill({
        json: {
          code: 0,
          data: {
            items: items,
            summary: summary,
            total: items.length,
            page: 1,
            page_size: 20,
            total_pages: 1,
          },
        },
      })
    })

    await page.goto('/dashboard/agent-runs')
    await page.waitForLoadState('networkidle')

    // Switch to Review Queue tab
    await page.getByRole('tab', { name: 'Review Queue' }).click()
    await page.waitForLoadState('networkidle')

    // Check summary is displayed with specific context to avoid ambiguity
    await expect(page.getByText('Review Queue Summary')).toBeVisible()
    const summarySection = page.locator('text=Review Queue Summary').locator('..').locator('..')
    await expect(summarySection.getByText('Total')).toBeVisible()
    await expect(summarySection.getByText('Not Reviewed')).toBeVisible()
    await expect(summarySection.getByText('Reviewed', { exact: true })).toBeVisible()

    // Check review queue item is displayed
    await expect(page.getByText('Test Agent Run')).toBeVisible()

    // Switch to needs_attention filter and verify API request
    const reviewQueueRequestPromise = page.waitForRequest(request =>
      request.url().includes('/api/v2/agent-runs/review-queue') &&
      request.url().includes('review_state=needs_attention')
    )
    
    // Click on the review state dropdown
    await page.getByText('All States').click()
    await page.getByRole('option', { name: 'Needs Attention' }).click()
    
    await reviewQueueRequestPromise
    await page.waitForLoadState('networkidle')

    // Verify summary changed to reflect needs_attention filter
    // The mock returns summary with needs_attention=1, not_reviewed=0 when review_state=needs_attention
    await expect(summarySection.getByText('Needs Attention')).toBeVisible()
    
    // Assert specific summary values: needs_attention=1, not_reviewed=0
    // Check the summary text content contains the expected values
    const summaryText = await summarySection.textContent()
    expect(summaryText).toContain('Needs Attention')
    // The mock returns needs_attention=1, not_reviewed=0 for needs_attention filter
    // The UI displays values before labels (e.g., "1Needs Attention")
    expect(summaryText).toMatch(/1.*Needs Attention/)
    expect(summaryText).toMatch(/0.*Not Reviewed/)
  })

  test('should display follow-up history in agent run detail', async ({ page }) => {
    // Mock audit logs API globally before any navigation
    await page.route('**/api/v2/audit/logs**', route => {
      route.fulfill({
        json: {
          code: 0,
          data: {
            items: [
              {
                id: 'audit-123',
                action: 'agent_run.followup_created',
                resource_type: 'agent_run',
                resource_id: 'agent-run-1',
                timestamp: '2026-05-24T00:00:00Z',
              },
            ],
            total: 1,
            page: 1,
            page_size: 20,
            total_pages: 1,
          },
        },
      })
    })

    // Listen for console errors to catch network errors
    const consoleErrors: string[] = []
    page.on('console', msg => {
      if (msg.type() === 'error') {
        consoleErrors.push(msg.text())
      }
    })

    // Mock followup history API
    await page.route('**/api/v2/agent-runs/agent-run-1/followups**', route => {
      route.fulfill({
        json: {
          code: 0,
          data: {
            data: [
              {
                operation_id: 'op-followup-1',
                action_type: 'recheck_evidence',
                risk_level: 'low',
                reason: 'Evidence needs second review',
                review_packet_id: 'agent-run-1',
                audit_ref_id: 'audit-123',
                created_at: '2026-05-24T00:00:00Z',
              },
            ],
          },
        },
      })
    })

    await page.goto('/dashboard/agent-runs/agent-run-1')
    await page.waitForLoadState('networkidle')

    // Check follow-up history section is displayed
    await expect(page.getByText('Follow-up History (1)')).toBeVisible()
    await expect(page.getByText('recheck_evidence')).toBeVisible()
    await expect(page.getByText('Reason:')).toBeVisible()
    await expect(page.getByText('Evidence needs second review')).toBeVisible()
    await expect(page.getByText('Review Packet ID:')).toBeVisible()
    await expect(page.getByText('agent-run-1', { exact: true })).toBeVisible()
    await expect(page.getByText('Audit Ref:')).toBeVisible()

    // Check audit link exists with correct query parameters
    const auditLink = page.getByRole('link', { name: 'audit-123' })
    await expect(auditLink).toBeVisible()
    const auditHref = await auditLink.getAttribute('href')
    expect(auditHref).toContain('/dashboard/audit')
    expect(auditHref).toContain('resource_type=agent_run')
    expect(auditHref).toContain('resource_id=agent-run-1')

    // Click audit link and wait for audit logs API request
    const auditRequestPromise = page.waitForRequest(request =>
      request.url().includes('/api/v2/audit/logs') &&
      request.url().includes('resource_type=agent_run') &&
      request.url().includes('resource_id=agent-run-1')
    )
    await auditLink.click()
    await auditRequestPromise
    await page.waitForLoadState('networkidle')

    // Verify we are on the audit page
    await expect(page).toHaveURL(/\/dashboard\/audit/)
    
    // Verify that no network error occurred
    expect(consoleErrors.some(error => error.includes('Network Error'))).toBe(false)
    
    // Verify audit page displays the expected audit log entry
    // The action should be visible in the Action column
    await expect(page.getByText('agent_run.followup_created')).toBeVisible()
  })
})
