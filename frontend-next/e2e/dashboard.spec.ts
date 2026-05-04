import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display dashboard', async ({ page }) => {
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('h1')).toContainText('GODNSLOG 2.0');
  });

  test('should display empty state for cases', async ({ page }) => {
    await page.waitForTimeout(5000);
    await expect(page.locator('text=暂无 Cases').first()).toBeVisible();
  });

  test('should display empty state for interactions', async ({ page }) => {
    await page.waitForTimeout(5000);
    await expect(page.locator('text=暂无命中记录').first()).toBeVisible();
  });

  test('should display stats with zero values', async ({ page }) => {
    await page.waitForTimeout(5000);
    await expect(page.locator('text=0').first()).toBeVisible();
  });
});
