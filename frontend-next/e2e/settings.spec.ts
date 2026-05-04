import { test, expect } from '@playwright/test';

test.describe('Settings Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test.skip('should display settings page', async ({ page }) => {
    await page.goto('/dashboard/settings');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2')).toContainText('系统设置');
  });

  test.skip('should display tab buttons', async ({ page }) => {
    await page.goto('/dashboard/settings');
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("通用设置")')).toBeVisible();
    await expect(page.locator('button:has-text("域名设置")')).toBeVisible();
    await expect(page.locator('button:has-text("监听配置")')).toBeVisible();
  });

  test.skip('should switch tabs', async ({ page }) => {
    await page.goto('/dashboard/settings');
    await page.waitForTimeout(2000);
    
    await page.click('button:has-text("域名设置")');
    await page.waitForTimeout(500);
    await expect(page.locator('button:has-text("域名设置")')).toHaveClass(/bg-indigo-600/);
  });
});
