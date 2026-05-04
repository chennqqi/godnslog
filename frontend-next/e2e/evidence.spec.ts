import { test, expect } from '@playwright/test';

test.describe('Evidence Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display evidence page', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toContainText('证据报告');
  });

  test('should display generate report section', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=生成报告').first()).toBeVisible();
  });

  test('should display case selector', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    await expect(page.locator('select').first()).toBeVisible();
  });
});
