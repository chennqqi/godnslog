import { test, expect } from '@playwright/test';

test.describe('Payload Cheat Sheet', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({
        id: 1, username: 'admin', email: 'admin@test.com', role: 0, lang: 'en-US',
      }));
    });
    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });
  });

  test('should display cheat sheet on dashboard', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    // Cheat sheet should be collapsed by default
    await expect(page.getByText('Payload Cheat Sheet')).toBeVisible();
  });

  test('should expand cheat sheet on click', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    // Click to expand
    await page.getByText('Payload Cheat Sheet').click();
    await page.waitForTimeout(1000);

    // Should show category headers
    await expect(page.getByText('SSRF')).toBeVisible();
    await expect(page.getByText('RCE')).toBeVisible();
  });

  test('should show SSRF payloads when category expanded', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    // Expand cheat sheet
    await page.getByText('Payload Cheat Sheet').click();
    await page.waitForTimeout(500);

    // SSRF category should be expanded by default, showing templates
    await expect(page.getByText('http://{token}.{domain}/')).toBeVisible();
  });

  test('should have copy buttons on payload items', async ({ page }) => {
    await page.goto('/dashboard');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(3000);

    // Expand cheat sheet
    await page.getByText('Payload Cheat Sheet').click();
    await page.waitForTimeout(500);

    // Should have Copy buttons
    const copyButtons = page.getByText('Copy');
    await expect(copyButtons.first()).toBeVisible();
  });
});
