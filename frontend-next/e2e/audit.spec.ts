import { expect, test, type Page, type Route } from '@playwright/test'

const tableHash = 'abc123def4567890123456789012345678901234567890123456789012345678'
const emptyHash = '0000000000000000000000000000000000000000000000000000000000000000'

type TraceOptions = {
  packageHash?: string | null
  empty?: boolean
  includeSensitive?: boolean
}

const tracePayload = ({ packageHash = tableHash, empty = false, includeSensitive = false }: TraceOptions = {}) => {
  if (empty) {
    return {
      package_hash: packageHash,
      summary: {
        agent_run_count: 0,
        export_count: 0,
        delivery_count: 0,
        audit_count: 0,
        delivered: 0,
        failed: 0,
        timeout: 0,
      },
      agent_runs: [],
      exports: [],
      deliveries: [],
      audits: [],
    }
  }

  return {
    package_hash: packageHash,
    summary: {
      agent_run_count: 1,
      export_count: 1,
      delivery_count: 1,
      audit_count: 1,
      delivered: 1,
      failed: 0,
      timeout: 0,
    },
    agent_runs: [{
      agent_run_id: 'agent-run-1',
      agent_id: 'agent-123',
      operator_id: 'test-user',
      title: 'Test Agent Run',
      status: 'completed',
      case_id: 'case-1',
      payload_id: 'payload-1',
      started_at: '2026-06-10T00:00:00Z',
      ended_at: '2026-06-10T00:05:00Z',
    }],
    exports: [{
      agent_run_id: 'agent-run-1',
      operation_id: 'op-export-1',
      audit_ref_id: 'audit-export-1',
      format: 'json',
      created_at: '2026-06-10T00:01:00Z',
    }],
    deliveries: [{
      agent_run_id: 'agent-run-1',
      delivery_id: 'delivery-1',
      delivery_operation_id: 'op-delivery-1',
      audit_ref_id: 'audit-delivery-1',
      format: 'json',
      result: 'delivered',
      status_code: 200,
      destination_host: 'hooks.example.com',
      created_at: '2026-06-10T00:02:00Z',
      ...(includeSensitive ? {
        webhook_url: 'https://hooks.example.com/webhook/full-path?token=secret-query-token',
        headers: {
          Authorization: 'Bearer secret-token-12345',
          'X-API-Key': 'api-key-sensitive-67890',
          Cookie: 'session=secret-session-id',
        },
        response_body: '{"secret":"hidden-data"}',
        secret: 'top-secret-package-value',
      } : {}),
    }],
    audits: [{
      audit_ref_id: 'audit-1',
      agent_run_id: 'agent-run-1',
      action: 'agent_run.review_exported',
      resource_type: 'agent_run',
      resource_id: 'agent-run-1',
      result: 'success',
      timestamp: '2026-06-10T00:01:00Z',
    }],
  }
}

const fulfillTrace = async (route: Route, options: TraceOptions = {}) => {
  const url = new URL(route.request().url())
  await route.fulfill({
    status: 200,
    contentType: 'application/json',
    body: JSON.stringify({
      code: 0,
      message: 'success',
      data: tracePayload({
        packageHash: options.packageHash ?? url.searchParams.get('package_hash'),
        empty: options.empty,
        includeSensitive: options.includeSensitive,
      }),
    }),
  })
}

const gotoAudit = async (page: Page) => {
  await page.goto('/dashboard/audit')
  await expect(page.getByRole('heading', { name: 'Audit Log', exact: true })).toBeVisible()
  await expect(page.getByRole('table')).toBeVisible()
}

const waitForTraceRequest = (page: Page, expectedHash: string) =>
  page.waitForRequest(request => {
    if (!request.url().includes('/api/v2/agent-runs/review-package-trace')) return false
    return new URL(request.url()).searchParams.get('package_hash') === expectedHash
  })

const expectSummaryCard = async (page: Page, label: string, value: string) => {
  await expect(page.locator('div').filter({ hasText: new RegExp(`^${value}${label}$`) }).first()).toBeVisible()
}

const expectRenderedTrace = async (page: Page) => {
  await expectSummaryCard(page, 'Agent Runs', '1')
  await expectSummaryCard(page, 'Exports', '1')
  await expectSummaryCard(page, 'Deliveries', '1')
  await expectSummaryCard(page, 'Audits', '1')
  await expectSummaryCard(page, 'Delivered', '1')
  await expectSummaryCard(page, 'Failed', '0')
  await expectSummaryCard(page, 'Timeout', '0')

  await expect(page.getByText('Test Agent Run')).toBeVisible()
  await expect(page.getByText('Case: case-1')).toBeVisible()
  await expect(page.getByText('Payload: payload-1')).toBeVisible()
  await expect(page.getByText('Operation ID: op-export-1')).toBeVisible()
  await expect(page.getByText('Operation ID: op-delivery-1')).toBeVisible()
  await expect(page.getByText('Audit Ref: audit-1')).toBeVisible()
  await expect(page.getByText('Host: hooks.example.com')).toBeVisible()
  await expect(page.getByText('Status: 200')).toBeVisible()
}

test.describe('Audit Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'test-token')
    })

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

    await page.route('**/api/v2/audit/logs**', route => {
      route.fulfill({
        json: {
          code: 0,
          data: {
            items: [{
              id: 'audit-1',
              user_id: 'user-1',
              ip_address: '127.0.0.1',
              action: 'agent_run.review_exported',
              resource_type: 'agent_run',
              resource_id: 'agent-run-1',
              result: 'success',
              timestamp: '2026-06-10T00:00:00Z',
              details: {
                package_hash: tableHash,
              },
            }],
            total: 1,
            page: 1,
            page_size: 20,
            total_pages: 1,
          },
        },
      })
    })
  })

  test('displays audit table and package trace controls', async ({ page }) => {
    await gotoAudit(page)

    await expect(page.getByText('Package Hash Trace')).toBeVisible()
    await expect(page.getByPlaceholder('Paste 64-character package hash...')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Trace' })).toBeVisible()
    await expect(page.getByRole('columnheader', { name: 'Timestamp' })).toBeVisible()
    await expect(page.getByText('agent_run.review_exported').first()).toBeVisible()
    await expect(page.getByText('abc123def456...')).toBeVisible()
  })

  test('rejects invalid package hash before calling trace API', async ({ page }) => {
    let traceCalls = 0
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => {
      traceCalls += 1
      route.fulfill({ status: 500, body: 'trace API should not be called' })
    })

    await gotoAudit(page)
    await page.getByPlaceholder('Paste 64-character package hash...').fill('invalid-hash')
    await page.getByRole('button', { name: 'Trace' }).click()

    await expect(page.getByText('Invalid package hash: must be 64-character hex string')).toBeVisible()
    expect(traceCalls).toBe(0)
  })

  test('traces valid package hash and renders aggregated refs', async ({ page }) => {
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => fulfillTrace(route))

    await gotoAudit(page)
    const traceRequest = waitForTraceRequest(page, tableHash)
    await page.getByPlaceholder('Paste 64-character package hash...').fill(tableHash)
    await page.getByRole('button', { name: 'Trace' }).click()
    await traceRequest

    await expectRenderedTrace(page)
  })

  test('renders an empty trace result for a known hash with no records', async ({ page }) => {
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => fulfillTrace(route, { empty: true }))

    await gotoAudit(page)
    const traceRequest = waitForTraceRequest(page, emptyHash)
    await page.getByPlaceholder('Paste 64-character package hash...').fill(emptyHash)
    await page.getByRole('button', { name: 'Trace' }).click()
    await traceRequest

    await expectSummaryCard(page, 'Agent Runs', '0')
    await expectSummaryCard(page, 'Exports', '0')
    await expectSummaryCard(page, 'Deliveries', '0')
    await expectSummaryCard(page, 'Audits', '0')
    await expect(page.getByText('No package trace records found for this hash.')).toBeVisible()
  })

  test('clicks package hash in audit table and traces the clicked hash', async ({ page }) => {
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => fulfillTrace(route))

    await gotoAudit(page)
    const traceRequest = waitForTraceRequest(page, tableHash)
    await page.getByText('abc123def456...').click()
    await traceRequest

    await expect(page.getByPlaceholder('Paste 64-character package hash...')).toHaveValue(tableHash)
    await expectRenderedTrace(page)
  })

  test('does not render sensitive delivery fields from trace response', async ({ page }) => {
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => fulfillTrace(route, { includeSensitive: true }))

    await gotoAudit(page)
    const traceRequest = waitForTraceRequest(page, tableHash)
    await page.getByPlaceholder('Paste 64-character package hash...').fill(tableHash)
    await page.getByRole('button', { name: 'Trace' }).click()
    await traceRequest

    await expect(page.getByText('Host: hooks.example.com')).toBeVisible()
    await expect(page.getByText('Operation ID: op-delivery-1')).toBeVisible()

    const content = await page.content()
    for (const sensitive of [
      'secret-token-12345',
      'api-key-sensitive-67890',
      'secret-session-id',
      'hidden-data',
      'secret-query-token',
      'https://hooks.example.com/webhook/full-path',
      'Bearer',
      'Authorization',
      'X-API-Key',
      'Cookie',
      'top-secret-package-value',
    ]) {
      expect(content).not.toContain(sensitive)
    }
  })
})
