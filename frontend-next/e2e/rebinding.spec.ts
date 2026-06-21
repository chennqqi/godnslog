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
          json: { code: 0, data: [] }
        });
      }
      return route.fulfill({ json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } } });
    });

    await page.goto('/dashboard/rebinding');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display rebinding page', async ({ page }) => {
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display predefined scenarios', async ({ page }) => {
    // Page uses scenario cards to create rebinding rules
    await expect(page.locator('text=Predefined Scenarios').first()).toBeVisible({ timeout: 5000 });
  });
});
