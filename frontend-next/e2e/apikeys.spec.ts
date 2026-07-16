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

  test.beforeEach(async ({ page }) => {
    await page.route('**/api/v2/agent-policy/scopes', async route => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          code: 0,
          message: 'success',
          data: {
            items: [
              {
                scope: 'agent:create_probe',
                name: 'Create OAST Probe',
                risk_level: 'medium',
                default_allowed: true,
                high_risk: false,
                tool_names: ['create_oast_probe'],
                description: 'Create Case and Payload resources.',
              },
              {
                scope: 'agent:wait_interaction',
                name: 'Wait For Interaction',
                risk_level: 'low',
                default_allowed: true,
                high_risk: false,
                tool_names: ['wait_for_interaction'],
                description: 'Poll for interactions.',
              },
              {
                scope: 'agent:read_runs',
                name: 'Read Agent Runs',
                risk_level: 'low',
                default_allowed: true,
                high_risk: false,
                tool_names: ['list_agent_runs'],
                description: 'Read agent run status.',
              },
              {
                scope: 'agent:revoke_token',
                name: 'Revoke Token',
                risk_level: 'high',
                default_allowed: false,
                high_risk: true,
                tool_names: ['revoke_token'],
                description: 'Revoke API tokens.',
              },
            ],
            default_scopes: ['agent:create_probe', 'agent:wait_interaction', 'agent:read_runs'],
            high_risk_scopes: ['agent:revoke_token'],
          },
        }),
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

    await page.goto('/apikeys');
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

    await page.goto('/apikeys');
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

  test('should load agent policy and group scopes by risk', async ({ page }) => {
    let policyCalled = false
    let createRequestBody: Record<string, unknown> | null = null

    await page.route('**/api/v2/agent-policy/scopes', async route => {
      policyCalled = true
      await route.fulfill({
        json: {
          code: 0,
          data: {
            items: [
              {
                scope: 'agent:create_probe',
                name: 'Create OAST Probe',
                risk_level: 'medium',
                default_allowed: true,
                high_risk: false,
                tool_names: ['create_oast_probe'],
                description: 'Create Case and Payload resources.',
              },
              {
                scope: 'agent:wait_interaction',
                name: 'Wait For Interaction',
                risk_level: 'low',
                default_allowed: true,
                high_risk: false,
                tool_names: ['wait_for_interaction'],
                description: 'Poll for interactions.',
              },
              {
                scope: 'agent:revoke_token',
                name: 'Revoke Token',
                risk_level: 'high',
                default_allowed: false,
                high_risk: true,
                tool_names: ['revoke_token'],
                description: 'Revoke API tokens.',
              },
            ],
            default_scopes: ['agent:create_probe', 'agent:wait_interaction'],
            high_risk_scopes: ['agent:revoke_token'],
          },
        },
      })
    })

    await page.route('**/api/v2/apikeys**', async route => {
      if (route.request().method() === 'POST') {
        createRequestBody = await route.request().postDataJSON()
        return route.fulfill({
          json: {
            code: 0,
            data: {
              id: 'agent-key-risk',
              key: 'gdl_' + 'z'.repeat(32),
              key_prefix: 'gdl_risk',
              name: 'Agent Risk Key',
              scopes: createRequestBody?.scopes || [],
              is_agent: true,
              risk_tolerance: createRequestBody?.risk_tolerance || 'medium',
              is_revoked: false,
              created_at: '2024-01-01T00:00:00Z',
              created_by: 'user1',
            },
          },
        })
      }
      return route.fulfill({
        json: {
          code: 0,
          data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 },
        },
      })
    })

    await page.goto('/apikeys')
    await page.getByRole('button', { name: '创建 API Key' }).click()
    await page.getByLabel('名称').fill('Agent Risk Key')
    await page.getByLabel('Agent Key (AI Agent 专用)').check()

    expect(policyCalled).toBe(true)
    await expect(page.getByText('默认 Agent 作用域')).toBeVisible()
    await expect(page.getByText('高风险作用域')).toBeVisible()
    await expect(page.getByText('agent:revoke_token')).toBeVisible()
    await expect(page.getByLabel('agent:create_probe')).toBeChecked()
    await expect(page.getByLabel('agent:revoke_token')).not.toBeChecked()

    await page.getByLabel('agent:revoke_token').check()
    await page.getByLabel('风险容忍度').selectOption('high')
    const createRequest = page.waitForRequest(request => {
      return request.url().includes('/api/v2/apikeys') && request.method() === 'POST'
    })
    await page.locator('.fixed').getByRole('button', { name: '创建' }).click()
    await createRequest

    expect(createRequestBody?.is_agent).toBe(true)
    expect(createRequestBody?.risk_tolerance).toBe('high')
    expect(createRequestBody?.scopes).toContain('agent:create_probe')
    expect(createRequestBody?.scopes).toContain('agent:revoke_token')
  })

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

    // Handle custom confirm dialog (not native)
    await page.goto('/apikeys');
    await page.click('button:has-text("删除")');
    // Wait for the custom confirm dialog and click Delete
    await page.waitForTimeout(500);
    await page.click('button:has-text("Delete")');

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

    await page.goto('/apikeys');
    const keyText = await page.locator('text=Key:').textContent();
    expect(keyText).toContain('********');
    expect(keyText).not.toContain('gdl_abc123' + 'x'.repeat(24));
  });
});
