import { test, expect } from '@playwright/test';

test.describe('Canary Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/canary') && !url.includes('/hits')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                {
                  id: 'canary-1',
                  type: 'dns',
                  token: 'canary-abc123',
                  description: 'DNS canary for prod',
                  context: '',
                  is_enabled: true,
                  status: 'active',
                  created_at: new Date().toISOString(),
                  expires_at: '2026-12-31T23:59:59Z',
                },
                {
                  id: 'canary-2',
                  type: 'http',
                  token: 'canary-def456',
                  description: 'HTTP canary for staging',
                  context: '',
                  is_enabled: false,
                  status: 'revoked',
                  created_at: new Date().toISOString(),
                  expires_at: '2026-06-30T23:59:59Z',
                },
              ],
              total: 2,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/canary');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display canary page', async ({ page }) => {
    await expect(page).toHaveURL('/dashboard/canary');
  });

  test('should display canary tokens title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Canary Tokens', exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('should display new canary token button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'New Canary Token' })).toBeVisible({ timeout: 5000 });
  });

  test('should display summary cards', async ({ page }) => {
    await expect(page.getByText('Active').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Silent').first()).toBeVisible();
    await expect(page.getByText('Revoked').first()).toBeVisible();
  });

  test('should display token registry', async ({ page }) => {
    await expect(page.getByText('Token Registry')).toBeVisible({ timeout: 5000 });
  });

  test('should display canary tokens in list', async ({ page }) => {
    await expect(page.getByText('canary-abc123')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('canary-def456')).toBeVisible();
  });

  test('should display token type badges', async ({ page }) => {
    await expect(page.getByText('DNS').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('HTTP').first()).toBeVisible();
  });

  test('should redirect to login when no auth token exists', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/canary');
    await page.waitForURL('**/login', { timeout: 10000 });
    await expect(page).toHaveURL(/\/login/);

    await context.close();
  });
});
