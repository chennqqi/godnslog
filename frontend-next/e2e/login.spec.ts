import { test, expect } from '@playwright/test';

test.describe('Login Page', () => {
  test('should display login form', async ({ page }) => {
    await page.goto('/login');

    await expect(page.locator('h2')).toContainText('GODNSLOG 2.0');
    await expect(page.locator('input[name="username"]')).toBeVisible();
    await expect(page.locator('input[name="password"]')).toBeVisible();
    await expect(page.locator('button[type="submit"]')).toBeVisible();
  });

  test('should not expose credentials in URL on form submit', async ({ page }) => {
    await page.goto('/login');

    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'secret123');

    // Submit the form
    await page.click('button[type="submit"]');

    // Wait a moment for any navigation or error
    await page.waitForTimeout(2000);

    // Verify URL does not contain username or password
    const currentUrl = page.url();
    expect(currentUrl).not.toContain('username=');
    expect(currentUrl).not.toContain('password=');
    expect(currentUrl).not.toContain('admin');
    expect(currentUrl).not.toContain('secret123');
  });

  test('should show error with invalid credentials', async ({ page }) => {
    await page.goto('/login');

    await page.fill('input[name="username"]', 'invaliduser');
    await page.fill('input[name="password"]', 'wrongpassword');
    await page.click('button[type="submit"]');

    // Wait for error message or stay on login page
    await page.waitForTimeout(3000);

    // Should still be on login page (not redirected to dashboard)
    await expect(page).toHaveURL(/\/login/);
  });

  test('should redirect to home after successful login', async ({ page }) => {
    const password = process.env.ADMIN_PASSWORD || process.env.PLAYWRIGHT_PASSWORD || ''
    await page.goto('/login');

    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', password);
    await page.click('button[type="submit"]');

    // Should redirect to home page
    await page.waitForURL((url) => !url.pathname.startsWith('/login'), { timeout: 15000 });
    expect(page.url()).not.toContain('/login');
  });
});
