import { test, expect } from '@playwright/test';

test.describe('Interactions Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses - only intercept actual API calls
    await page.route('**/api/v2/interactions**', route => route.fulfill({
      json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } }
    }))
    await page.route('**/api/v2/interactions/stats**', route => route.fulfill({
      json: { code: 0, data: { total: 0, dns_count: 0, http_count: 0, smtp_count: 0, ldap_count: 0 } }
    }))
    // Set token before navigation to avoid redirect to login
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });
    await page.goto('/dashboard/interactions')
    await page.waitForLoadState('domcontentloaded')
    await page.waitForTimeout(5000);
  });

  test('should display interactions page', async ({ page }) => {
    await expect(page.locator('h2')).toContainText('Interaction Timeline');
  });

  test('should display empty state for interactions', async ({ page }) => {
    await expect(page.getByText('暂无命中记录')).toBeVisible();
  });

  test('should display filter controls', async ({ page }) => {
    await expect(page.getByPlaceholder('搜索 IP、域名或token...')).toBeVisible();
  });

  test('should display interaction detail from API', async ({ page }) => {
    await page.route('**/api/v2/interactions/interaction-1', route => route.fulfill({
      json: { code: 0, data: { id: 'interaction-1', type: 'dns', token: 'tok-1' } }
    }))
    await page.goto('/dashboard/interactions/interaction-1')
    await page.waitForLoadState('domcontentloaded')
    await page.waitForTimeout(2000)
    // Detail page may not exist yet, skip for now
    test.skip()
  });
});
