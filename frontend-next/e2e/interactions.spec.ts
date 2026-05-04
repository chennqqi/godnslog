import { test, expect } from '@playwright/test';

test.describe('Interactions Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display interactions page', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display empty state for interactions', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=暂无命中记录').first()).toBeVisible();
  });

  test('should display filter controls', async ({ page }) => {
    await page.goto('/dashboard/interactions');
    await page.waitForTimeout(2000);
    await expect(page.locator('select').first()).toBeVisible();
  });
});
