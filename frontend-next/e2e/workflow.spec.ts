import { test, expect } from '@playwright/test';

test.describe('Workflow Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display workflow page', async ({ page }) => {
    await page.goto('/dashboard/workflow');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toBeVisible();
  });

  test('should display condition list', async ({ page }) => {
    await page.goto('/dashboard/workflow');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=规则列表').first()).toBeVisible();
  });

  test('should display action list', async ({ page }) => {
    await page.goto('/dashboard/workflow');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=选择一个规则进行编辑').or(page.locator('text=新建')).first()).toBeVisible();
  });
});
