import { test, expect } from '@playwright/test';

test.describe('Rebinding Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display rebinding page', async ({ page }) => {
    await page.goto('/dashboard/rebinding');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display scenario cards', async ({ page }) => {
    await page.goto('/dashboard/rebinding');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=浏览器').first()).toBeVisible();
  });

  test('should display add stage button', async ({ page }) => {
    await page.goto('/dashboard/rebinding');
    await page.waitForTimeout(2000);
    await expect(page.locator('button').filter({ hasText: /添加/i }).first()).toBeVisible();
  });
});
