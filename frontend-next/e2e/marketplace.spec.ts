import { test, expect } from '@playwright/test';

test.describe('Marketplace Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } } });
    });

    await page.goto('/marketplace');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display marketplace page', async ({ page }) => {
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display tab buttons', async ({ page }) => {
    // Page uses English tab labels
    await expect(page.locator('button').filter({ hasText: 'Plugins' }).first()).toBeVisible({ timeout: 5000 });
    await expect(page.locator('button').filter({ hasText: 'Templates' }).first()).toBeVisible();
    await expect(page.locator('button').filter({ hasText: 'Installed' }).first()).toBeVisible();
  });
});
