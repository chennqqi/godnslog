import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[type="email"]', 'admin@example.com');
    await page.fill('input[type="password"]', 'password');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard');
  });

  test('should display dashboard', async ({ page }) => {
    await expect(page.locator('h1')).toContainText('Dashboard');
  });

  test('should navigate to cases page', async ({ page }) => {
    await page.click('a[href="/dashboard/cases"]');
    await expect(page).toHaveURL('/dashboard/cases');
  });
});
