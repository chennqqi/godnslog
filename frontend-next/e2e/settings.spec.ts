import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/settings');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display settings page', async ({ page }) => {
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display tab buttons', async ({ page }) => {
    // Page uses Radix Tabs with Chinese labels
    await expect(page.getByRole('tab').filter({ hasText: '通用设置' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('tab').filter({ hasText: '域名设置' })).toBeVisible();
    await expect(page.getByRole('tab').filter({ hasText: '监听配置' })).toBeVisible();
  });
});
