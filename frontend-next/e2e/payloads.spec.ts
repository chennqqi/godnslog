import { test, expect } from '@playwright/test';

test.describe('Payloads Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display payloads page', async ({ page }) => {
    await page.goto('/dashboard/payloads');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display empty state for payloads', async ({ page }) => {
    await page.goto('/dashboard/payloads');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=暂无 Payloads').first()).toBeVisible();
  });

  test('should display search input', async ({ page }) => {
    await page.goto('/dashboard/payloads');
    await page.waitForTimeout(2000);
    await expect(page.locator('input[type="text"]').first()).toBeVisible();
  });
});
