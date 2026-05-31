import { test, expect } from '@playwright/test';

test.describe('API Keys Page', () => {
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
  })

  test('should display API keys list', async ({ page }) => {
    // Override the default mock for this specific test
    await page.route('**/api/v2/apikeys*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          message: 'success',
          data: {
            items: [
              {
                id: 'key1',
                key_prefix: 'gdl_abc123',
                name: 'Test Key 1',
                scopes: ['case:read', 'payload:read'],
                is_agent: false,
                is_revoked: false,
                created_at: '2024-01-01T00:00:00Z',
                created_by: 'user1',
              },
            ],
            total: 1,
            page: 1,
            page_size: 20,
            total_pages: 1,
          },
        }),
      });
    });

    await page.goto('/dashboard/apikeys');
    await expect(page.locator('h2')).toContainText('API Keys 管理');
    await expect(page.locator('text=Test Key 1')).toBeVisible();
  });

  test('should create agent API key', async ({ page }) => {
    let createCalled = false;
    let createRequestBody: Record<string, unknown> = null;
    // Set up mock BEFORE navigation - match both GET and POST
    await page.route('**/api/v2/apikeys**', async (route) => {
      const method = route.request().method();
      if (method === 'POST') {
        createCalled = true;
        createRequestBody = await route.request().postDataJSON();
        expect(createRequestBody.is_agent).toBe(true);
        expect(createRequestBody.risk_tolerance).toBe('medium');
        // Verify agent scopes
        expect(createRequestBody.scopes).toContain('agent:create_probe');
        expect(createRequestBody.scopes).toContain('agent:wait_interaction');
        // Verify expires_at is calculated from expires_in
        expect(createRequestBody.expires_at).toBeDefined();
        expect(createRequestBody.expires_at).toMatch(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}/);

        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            code: 0,
            message: 'success',
            data: {
              id: 'new-agent-key',
              key: 'gdl_' + 'x'.repeat(32),
              key_prefix: 'gdl_xyz789',
              name: 'Agent Test Key',
              scopes: ['agent:create_probe', 'agent:wait_interaction'],
              is_agent: true,
              risk_tolerance: 'medium',
              is_revoked: false,
              created_at: '2024-01-01T00:00:00Z',
              created_by: 'user1',
            },
          }),
        });
      } else {
        // Handle GET requests (initial load)
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            code: 0,
            message: 'success',
            data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 },
          }),
        });
      }
    });

    await page.goto('/dashboard/apikeys');
    await page.click('button:has-text("创建 API Key")');

    await page.fill('input[type="text"]', 'Agent Test Key');
    await page.check('input[type="checkbox"]'); // Enable Agent mode

    // Verify that scopes are limited to agent: prefix when Agent mode is enabled
    const scopeCheckboxes = await page.locator('input[type="checkbox"][name*="scope"]').all();
    for (const checkbox of scopeCheckboxes) {
      const label = await checkbox.evaluate(el => (el as HTMLInputElement).labels?.[0]?.textContent || '');
      if (label && !label.startsWith('agent:')) {
        // Non-agent scopes should be disabled or hidden in Agent mode
        const isDisabled = await checkbox.isDisabled();
        expect(isDisabled).toBe(true);
      }
    }

    // Click the create button inside the modal (use more specific selector)
    await page.locator('.fixed').locator('button:has-text("创建")').click();

    expect(createCalled).toBe(true);

    // Verify that the full key modal is shown
    await expect(page.locator('.fixed:has-text("API Key 已创建")')).toBeVisible();
    await expect(page.locator(".fixed").locator("text=gdl_")).toBeVisible();

    // Close the modal
    await page.locator('.fixed').locator('button:has-text("我已复制")').click();
  });

  test('should revoke API key', async ({ page }) => {
    // Set up mock BEFORE navigation - handle both GET and DELETE
    let deleteCalled = false;
    await page.route('**/api/v2/apikeys**', async (route) => {
      const method = route.request().method();

      if (method === 'DELETE') {
        deleteCalled = true;
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({ code: 0, message: 'success' }),
        });
      } else if (method === 'GET') {
        await route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify({
            code: 0,
            message: 'success',
            data: {
              items: [
                {
                  id: 'key1',
                  key_prefix: 'gdl_abc123',
                  name: 'Test Key 1',
                  scopes: ['case:read'],
                  is_agent: false,
                  is_revoked: false,
                  created_at: '2024-01-01T00:00:00Z',
                  created_by: 'user1',
                },
              ],
              total: 1,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          }),
        });
      } else {
        // Fallback for other methods
        await route.continue();
      }
    });

    // Handle native confirm dialog
    page.on('dialog', dialog => dialog.accept());

    await page.goto('/dashboard/apikeys');
    await page.click('button:has-text("删除")');

    expect(deleteCalled).toBe(true);
  });

  test('should not leak full API key in list', async ({ page }) => {
    await page.route('**/api/v2/apikeys*', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          message: 'success',
          data: {
            items: [
              {
                id: 'key1',
                key_prefix: 'gdl_abc123',
                name: 'Test Key 1',
                scopes: ['case:read'],
                is_agent: false,
                is_revoked: false,
                created_at: '2024-01-01T00:00:00Z',
                created_by: 'user1',
              },
            ],
            total: 1,
            page: 1,
            page_size: 20,
            total_pages: 1,
          },
        }),
      });
    });

    await page.goto('/dashboard/apikeys');
    const keyText = await page.locator('text=Key:').textContent();
    expect(keyText).toContain('********');
    expect(keyText).not.toContain('gdl_abc123' + 'x'.repeat(24));
  });
});
