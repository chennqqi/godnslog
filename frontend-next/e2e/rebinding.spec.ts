import { test, expect } from '@playwright/test';

test.describe('Rebinding Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/rebinding/scenarios')) {
        return route.fulfill({
          json: {
            code: 0,
            data: [
              { name: 'browser-rebinding', description: 'Browser rebinding attack', stages: [{ target_ip: '1.2.3.4', ttl: 60 }, { target_ip: '127.0.0.1', ttl: 60 }] },
              { name: 'cloud-metadata', description: 'Cloud metadata extraction', stages: [{ target_ip: '169.254.169.254', ttl: 0 }, { target_ip: '127.0.0.1', ttl: 60 }] },
              { name: 'internal-management', description: 'Internal management access', stages: [{ target_ip: '10.0.0.1', ttl: 60 }, { target_ip: '127.0.0.1', ttl: 60 }] },
            ],
          },
        });
      }
      if (url.includes('/rebinding/rules')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                {
                  id: 'rule-1',
                  domain: 'rebind1.test.example.com',
                  is_enabled: true,
                  stages: [{ name: 'stage1', target_ip: '1.2.3.4', ttl: 60 }],
                  created_at: new Date().toISOString(),
                },
              ],
              total: 1,
              page: 1,
              page_size: 100,
              total_pages: 1,
            },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/rebinding');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display rebinding page', async ({ page }) => {
    await expect(page).toHaveURL('/dashboard/rebinding');
  });

  test('should display rebinding lab title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Rebinding Lab', exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('should display predefined scenarios section', async ({ page }) => {
    await expect(page.getByText('Predefined Scenarios').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display scenario cards', async ({ page }) => {
    await expect(page.getByText('browser-rebinding')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('cloud-metadata')).toBeVisible();
  });

  test('should display rebinding rules section', async ({ page }) => {
    await expect(page.getByText('Rebinding Rules').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display existing rule in list', async ({ page }) => {
    await expect(page.getByText('rebind1.test.example.com')).toBeVisible({ timeout: 5000 });
  });

  test('should display scenario description', async ({ page }) => {
    await expect(page.getByText('Browser rebinding attack')).toBeVisible({ timeout: 5000 });
  });

  test('should redirect to login when no auth token exists', async ({ browser }) => {
    const context = await browser.newContext();
    const page = await context.newPage();

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/rebinding');
    await page.waitForURL('**/login', { timeout: 10000 });
    await expect(page).toHaveURL(/\/login/);

    await context.close();
  });
});
