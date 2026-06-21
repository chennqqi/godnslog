import { test, expect } from '@playwright/test';

test.describe('Listeners Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/listeners')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'lst-1', protocol: 'smtp', host: '0.0.0.0', port: 25, token: 'tok-abc', is_enabled: true },
                { id: 'lst-2', protocol: 'ldap', host: '0.0.0.0', port: 389, token: 'tok-def', is_enabled: false },
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

    await page.goto('/dashboard/listeners');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display listeners page', async ({ page }) => {
    await expect(page).toHaveURL('/dashboard/listeners');
  });

  test('should display protocol listeners title', async ({ page }) => {
    await expect(page.getByText('Protocol Listeners', { exact: true })).toBeVisible({ timeout: 10000 });
  });

  test('should display create listener button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Listener' })).toBeVisible({ timeout: 5000 });
  });

  test('should display listener cards', async ({ page }) => {
    await expect(page.getByText('SMTP').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('LDAP').first()).toBeVisible();
  });

  test('should display listener status badges', async ({ page }) => {
    await expect(page.getByText('Running').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Stopped').first()).toBeVisible();
  });
});
