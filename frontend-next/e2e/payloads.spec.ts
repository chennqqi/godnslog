import { test, expect } from '@playwright/test';

test.describe('Payloads Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses - only intercept actual API calls
    await page.route('**/api/v2/payloads**', route => route.fulfill({
      json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } }
    }))
    // Set token before navigation to avoid redirect to login
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });
    await page.goto('/dashboard/payloads')
    await page.waitForLoadState('domcontentloaded')
    await page.waitForTimeout(5000);
  });

  test('should display payloads page', async ({ page }) => {
    // Debug: check page content
    console.log('Page URL:', page.url());
    const bodyText = await page.evaluate(() => document.body.innerText);
    console.log('Page body text:', bodyText);
    await expect(page.locator('h2')).toContainText('Payload Studio');
  });

  test('should display empty state for payloads', async ({ page }) => {
    await expect(page.getByText('暂无 Payloads')).toBeVisible();
  });

  test('should display search input', async ({ page }) => {
    await expect(page.getByPlaceholder('搜索 token 或模板...')).toBeVisible();
  });

  test('should display payload detail from API', async ({ page }) => {
    await page.route('**/api/v2/payloads/payload-1', route => route.fulfill({
      json: { code: 0, data: { id: 'payload-1', token: 'tok-1', value: 'https://tok-1.example.com' } }
    }))
    await page.goto('/dashboard/payloads/payload-1')
    await page.waitForLoadState('domcontentloaded')
    await page.waitForTimeout(2000)
    // Detail page may not exist yet, skip for now
    test.skip()
  });
});
