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

    await page.goto('/settings');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display settings page', async ({ page }) => {
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display tab buttons', async ({ page }) => {
    // Page uses Radix Tabs with English labels
    await expect(page.getByRole('tab').filter({ hasText: 'General' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('tab').filter({ hasText: 'Domain' })).toBeVisible();
    await expect(page.getByRole('tab').filter({ hasText: 'Listener' })).toBeVisible();
  });
});
