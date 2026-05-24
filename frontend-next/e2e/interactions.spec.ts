import { test, expect } from '@playwright/test';

test.describe('Interactions Page', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock API responses
    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/interactions')) {
        if (url.includes('/stats')) {
          return route.fulfill({
            json: {
              code: 0,
              data: { total: 10, dns_count: 5, http_count: 3, smtp_count: 1, ldap_count: 1 }
            }
          });
        }
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'int-1', type: 'dns', source_ip: '1.2.3.4', token: 'tok-1', domain: 'example.com', case_id: 'case-1', payload_id: 'payload-1', timestamp: new Date().toISOString() },
                { id: 'int-2', type: 'http', source_ip: '5.6.7.8', method: 'GET', path: '/test', case_id: 'case-1', timestamp: new Date().toISOString() }
              ],
              total: 2,
              page: 1,
              page_size: 20,
              total_pages: 1
            }
          }
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });
  });

  test('should display interactions page', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2').first()).toContainText('Interaction Timeline');
  });

  test('should display case scoped interactions', async ({ page }) => {
    const listRequestPromise = page.waitForRequest(request => {
      const url = new URL(request.url());
      return url.pathname.endsWith('/api/v2/interactions') &&
        url.searchParams.get('case_id') === 'case-1';
    });
    const statsRequestPromise = page.waitForRequest(request => {
      const url = new URL(request.url());
      return url.pathname.endsWith('/api/v2/interactions/stats') &&
        url.searchParams.get('case_id') === 'case-1';
    });

    await page.goto('/dashboard/interactions?case_id=case-1');
    await Promise.all([listRequestPromise, statsRequestPromise]);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=Case scoped: case-1')).toBeVisible();
    await expect(page.locator('button:has-text("Clear scope")')).toBeVisible();
  });

  test('should display payload scoped interactions', async ({ page }) => {
    const listRequestPromise = page.waitForRequest(request => {
      const url = new URL(request.url());
      return url.pathname.endsWith('/api/v2/interactions') &&
        url.searchParams.get('payload_id') === 'payload-1';
    });
    const statsRequestPromise = page.waitForRequest(request => {
      const url = new URL(request.url());
      return url.pathname.endsWith('/api/v2/interactions/stats') &&
        url.searchParams.get('payload_id') === 'payload-1';
    });

    await page.goto('/dashboard/interactions?payload_id=payload-1');
    await Promise.all([listRequestPromise, statsRequestPromise]);
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=Payload scoped: payload-1')).toBeVisible();
    await expect(page.locator('button:has-text("Clear scope")')).toBeVisible();
  });

  test('should render stats from the scoped stats request', async ({ page }) => {
    await page.route('**/api/v2/interactions/stats**', route => {
      const url = new URL(route.request().url());
      const total = url.searchParams.get('payload_id') === 'payload-1' ? 2 : 10;
      return route.fulfill({
        json: {
          code: 0,
          data: { total, dns_count: 1, http_count: 1, smtp_count: 0, ldap_count: 0 }
        }
      });
    });

    await page.goto('/dashboard/interactions?payload_id=payload-1');
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=Payload scoped: payload-1')).toBeVisible();
    await expect(page.locator('text=总数').locator('..').getByText('2')).toBeVisible();
  });

  test('should clear scope and return to all interactions', async ({ page }) => {
    await page.goto('/dashboard/interactions?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("Clear scope")');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=All Interactions')).toBeVisible();
  });

  test('should display empty state for scoped interactions', async ({ page }) => {
    await page.route('**/api/v2/interactions**', route => {
      return route.fulfill({
        json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } }
      });
    });
    await page.goto('/dashboard/interactions?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.getByText('当前 Case/Payload 暂无交互')).toBeVisible();
  });

  test('should open triage panel and display interaction details', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("详情")');
    await page.waitForTimeout(500);
    await expect(page.locator('text=Interaction Triage')).toBeVisible();
    await expect(page.locator('text=归因信息')).toBeVisible();
  });

  test('should display case and payload attribution in triage panel', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("详情")');
    await page.waitForTimeout(500);
    await expect(page.locator('text=Case ID: case-1')).toBeVisible();
    await expect(page.locator('text=Payload ID: payload-1')).toBeVisible();
  });

  test('should have copy token button in triage panel', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("详情")');
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("Copy")').first()).toBeVisible();
  });

  test('should have navigation buttons in triage panel', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("详情")');
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("View Case")')).toBeVisible();
    await expect(page.locator('button:has-text("View Payload")')).toBeVisible();
  });

  test('should have evidence generation buttons in triage panel', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await page.click('button:has-text("详情")');
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("Generate Evidence (Case)")')).toBeVisible();
    await expect(page.locator('button:has-text("Generate Evidence (Payload)")')).toBeVisible();
  });
});
