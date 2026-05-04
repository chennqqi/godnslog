import { test, expect } from '@playwright/test';

test.describe('Canary Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display canary page', async ({ page }) => {
    await page.goto('/dashboard/canary');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toContainText('Canary长期监测');
  });

  test('should display canary list', async ({ page }) => {
    await page.goto('/dashboard/canary');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Canary Token列表').first()).toBeVisible();
  });

  test('should display create button', async ({ page }) => {
    await page.goto('/dashboard/canary');
    await page.waitForTimeout(2000);
    await expect(page.locator('button').filter({ hasText: /创建/i }).first()).toBeVisible();
  });
});
