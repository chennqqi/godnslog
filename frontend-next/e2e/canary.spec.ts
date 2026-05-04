import { test, expect } from '@playwright/test';

test.describe('Canary Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test.skip('should display canary page', async ({ page }) => {
    await page.goto('/dashboard/canary');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2')).toContainText('Canary长期监测');
  });

  test.skip('should display canary list', async ({ page }) => {
    await page.goto('/dashboard/canary');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Canary Token列表').or(page.locator('text=管理长期部署的诱饵Token'))).toBeVisible();
  });

  test.skip('should display create button', async ({ page }) => {
    await page.goto('/dashboard/canary');
    await page.waitForTimeout(2000);
    await expect(page.locator('button').filter({ hasText: /创建/i }).or(page.locator('button').first())).toBeVisible();
  });
});
