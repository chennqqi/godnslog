import { test, expect } from '@playwright/test';

test.describe('Workflow Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test.skip('should display workflow page', async ({ page }) => {
    await page.goto('/dashboard/workflow');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').or(page.locator('h1'))).toBeVisible();
  });

  test.skip('should display condition list', async ({ page }) => {
    await page.goto('/dashboard/workflow');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=协议').or(page.locator('text=来源IP'))).toBeVisible();
  });

  test.skip('should display action list', async ({ page }) => {
    await page.goto('/dashboard/workflow');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=发送通知').or(page.locator('text=打标签'))).toBeVisible();
  });
});
