import { test, expect } from '@playwright/test';

test.describe('Marketplace Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display marketplace page', async ({ page }) => {
    await page.goto('/dashboard/marketplace');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toContainText('插件和模板市场');
  });

  test('should display tab buttons', async ({ page }) => {
    await page.goto('/dashboard/marketplace');
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("插件市场")').first()).toBeVisible();
    await expect(page.locator('button:has-text("模板市场")').first()).toBeVisible();
    await expect(page.locator('button:has-text("已安装")').first()).toBeVisible();
  });

  test('should switch tabs', async ({ page }) => {
    await page.goto('/dashboard/marketplace');
    await page.waitForTimeout(2000);
    
    await page.click('button:has-text("模板市场")');
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("模板市场")').first()).toHaveClass(/bg-indigo-600/);
  });
});
