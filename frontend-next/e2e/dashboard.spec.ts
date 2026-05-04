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

  test.skip('should navigate to cases page', async ({ page }) => {
    await page.waitForSelector('a[href="/dashboard/cases"]', { timeout: 5000 });
    const link = page.locator('a[href="/dashboard/cases"]');
    await link.click({ timeout: 5000 });
    await page.waitForURL('/dashboard/cases', { timeout: 5000 });
    await expect(page).toHaveURL('/dashboard/cases');
  });
});
