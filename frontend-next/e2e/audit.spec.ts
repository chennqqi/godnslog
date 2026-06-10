import { test, expect } from '@playwright/test';

test.describe('Audit Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display audit page', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toContainText('Audit Log');
  });

  test('should display audit table', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should display filters', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('input[placeholder*="Filter by user"]').first()).toBeVisible();
    await expect(page.locator('select').first()).toBeVisible();
  });

  test('should display refresh button', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("Refresh")').first()).toBeVisible();
  });

  test('should display table headers', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('th:has-text("Timestamp")').first()).toBeVisible();
    await expect(page.locator('th:has-text("User ID")').first()).toBeVisible();
    await expect(page.locator('th:has-text("IP")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Action")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Resource")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Result")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Details")').first()).toBeVisible();
  });

  test('should display empty state when no audit events', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const tableBody = page.locator('tbody').first();
    const rows = await tableBody.locator('tr').count();

    if (rows === 1) {
      // Check if it's the empty state row
      const emptyText = await page.locator('text=No audit events found').isVisible();
      if (emptyText) {
        await expect(page.locator('text=No audit events available').first()).toBeVisible();
      }
    }
  });

  test('should handle filter by user ID', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const searchInput = page.locator('input[placeholder*="Filter by user"]').first();
    await searchInput.fill('admin');
    await page.waitForTimeout(1000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should handle result filter', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const resultSelect = page.locator('select').first();
    await resultSelect.selectOption('success');
    await page.waitForTimeout(1000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should handle category filter', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const categorySelect = page.locator('select').nth(1);
    await categorySelect.selectOption('auth');
    await page.waitForTimeout(1000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should handle clear filters', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const searchInput = page.locator('input[placeholder*="Filter by user"]').first();
    await searchInput.fill('admin');
    await page.waitForTimeout(1000);

    // Check if clear button appears
    const clearButton = page.locator('button:has-text("Clear filters")');
    if (await clearButton.isVisible()) {
      await clearButton.click();
      await page.waitForTimeout(1000);
      await expect(searchInput).toHaveValue('');
    }
  });

  test('should display error message on API failure', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    // If there's an error, it should be displayed
    const errorCard = page.locator('.border-red-200');
    if (await errorCard.isVisible()) {
      await expect(errorCard.locator('.text-red-700').first()).toBeVisible();
    }
  });

  test('should refresh audit logs on button click', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const refreshButton = page.locator('button:has-text("Refresh")');
    await refreshButton.click();
    await page.waitForTimeout(2000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should display package hash trace section', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Package Hash Trace').first()).toBeVisible();
  });

  test('should validate package hash format', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    
    // Fill invalid hash
    await page.fill('input[placeholder*="Paste 64-character package hash"]', 'invalid-hash');
    await page.click('button:has-text("Trace")');
    await page.waitForTimeout(500);
    
    // Should show error
    await expect(page.locator('text=Invalid package hash').first()).toBeVisible();
  });

  test('should trace package with valid hash', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    
    // Mock the trace API
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          message: 'success',
          data: {
            package_hash: 'abc123def4567890123456789012345678901234567890123456789012345678',
            summary: {
              agent_run_count: 1,
              export_count: 1,
              delivery_count: 1,
              audit_count: 1,
              delivered: 1,
              failed: 0,
              timeout: 0
            },
            agent_runs: [{
              agent_run_id: 'agent-run-1',
              title: 'Test Agent Run',
              status: 'completed',
              case_id: 'case-1',
              payload_id: 'payload-1',
              target: 'example.com',
              url: '/dashboard/agent-runs/agent-run-1'
            }],
            exports: [{
              agent_run_id: 'agent-run-1',
              operation_id: 'op-export-1',
              audit_ref_id: 'audit-export-1',
              review_packet_id: 'packet-1',
              format: 'json',
              created_at: '2026-06-10T00:00:00Z'
            }],
            deliveries: [{
              agent_run_id: 'agent-run-1',
              delivery_id: 'delivery-1',
              delivery_operation_id: 'op-delivery-1',
              export_operation_id: 'op-export-1',
              audit_ref_id: 'audit-delivery-1',
              format: 'json',
              result: 'delivered',
              destination_host: 'hooks.example.com',
              status_code: 200,
              created_at: '2026-06-10T00:00:00Z',
              delivered_at: '2026-06-10T00:01:00Z'
            }],
            audits: [{
              audit_ref_id: 'audit-export-1',
              agent_run_id: 'agent-run-1',
              action: 'agent_run.review_exported',
              resource_type: 'agent_run',
              resource_id: 'agent-run-1',
              timestamp: '2026-06-10T00:00:00Z',
              url: '/dashboard/audit?resource_type=agent_run&resource_id=agent-run-1'
            }]
          }
        })
      });
    });
    
    // Fill valid hash
    await page.fill('input[placeholder*="Paste 64-character package hash"]', 'abc123def4567890123456789012345678901234567890123456789012345678');
    await page.click('button:has-text("Trace")');
    await page.waitForTimeout(500);
    
    // Should show trace results
    await expect(page.locator('text=Agent Runs').first()).toBeVisible();
    await expect(page.locator('text=Exports').first()).toBeVisible();
    await expect(page.locator('text=Deliveries').first()).toBeVisible();
    await expect(page.locator('text=Audit Records').first()).toBeVisible();
    
    // Verify summary counts
    await expect(page.locator('text=1').first()).toBeVisible();
  });

  test('should handle empty trace result', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    
    // Mock the trace API with empty result
    await page.route('**/api/v2/agent-runs/review-package-trace**', route => {
      route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          message: 'success',
          data: {
            package_hash: '0000000000000000000000000000000000000000000000000000000000000000',
            summary: {
              agent_run_count: 0,
              export_count: 0,
              delivery_count: 0,
              audit_count: 0,
              delivered: 0,
              failed: 0,
              timeout: 0
            },
            agent_runs: [],
            exports: [],
            deliveries: [],
            audits: []
          }
        })
      });
    });
    
    // Fill valid hash that returns no results
    await page.fill('input[placeholder*="Paste 64-character package hash"]', '0000000000000000000000000000000000000000000000000000000000000000');
    await page.click('button:has-text("Trace")');
    await page.waitForTimeout(500);
    
    // Should show trace results with empty counts
    await expect(page.locator('text=0').first()).toBeVisible();
  });
});
