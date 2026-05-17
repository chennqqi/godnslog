import { test, expect } from '@playwright/test';

test.describe('Workflow Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses - rules endpoint, not workflows
    await page.route('**/api/v2/rules', route => route.fulfill({
      json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } }
    }))
    // Set token before navigation to avoid redirect to login
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });
    await page.goto('/dashboard/workflow')
    await page.waitForLoadState('domcontentloaded')
    await page.waitForTimeout(5000);
  });

  test('should display workflow page', async ({ page }) => {
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display condition list', async ({ page }) => {
    await expect(page.locator('text=规则列表').first()).toBeVisible();
  });

  test('should display action list', async ({ page }) => {
    await expect(page.locator('text=选择一个规则进行编辑').or(page.locator('text=新建')).first()).toBeVisible();
  });
});
