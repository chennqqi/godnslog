import { test, expect } from '@playwright/test';

test.describe('Canary Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } } });
    });

    await page.goto('/dashboard/canary');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display canary page', async ({ page }) => {
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display create button', async ({ page }) => {
    await expect(page.locator('button').filter({ hasText: /create|new|add/i }).first()).toBeVisible({ timeout: 5000 });
  });
});
