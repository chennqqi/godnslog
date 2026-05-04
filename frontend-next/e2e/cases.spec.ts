import { test, expect } from '@playwright/test';

test.describe('Cases Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display cases page', async ({ page }) => {
    await page.goto('/dashboard/cases');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2')).toContainText('Case Board');
  });

  test('should show create case button', async ({ page }) => {
    await page.goto('/dashboard/cases');
    await page.waitForTimeout(2000);
    const createButton = page.locator('button').filter({ hasText: '创建 Case' }).first();
    await expect(createButton).toBeVisible();
  });

  test('should open create case modal', async ({ page }) => {
    await page.goto('/dashboard/cases');
    await page.waitForTimeout(2000);
    const createButton = page.locator('button').filter({ hasText: '创建 Case' }).first();
    await createButton.click();
    await expect(page.locator('h3:has-text("创建 Case")')).toBeVisible();
  });
});
