import { test, expect } from '@playwright/test';

test.describe('Payloads Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test.skip('should display payloads page', async ({ page }) => {
    await page.goto('/dashboard/payloads');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').or(page.locator('h1'))).toBeVisible();
  });

  test.skip('should display empty state for payloads', async ({ page }) => {
    await page.goto('/dashboard/payloads');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=暂无 Payloads').or(page.locator('text=Payload Studio')).or(page.locator('text=加载中'))).toBeVisible();
  });

  test.skip('should display search input', async ({ page }) => {
    await page.goto('/dashboard/payloads');
    await page.waitForTimeout(2000);
    await expect(page.locator('input[type="text"]').or(page.locator('input[placeholder*="搜索"]'))).toBeVisible();
  });
});
