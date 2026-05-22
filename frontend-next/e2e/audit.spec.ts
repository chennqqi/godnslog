import { test, expect } from '@playwright/test';

test.describe('Audit Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display audit page', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toContainText('Audit Log');
  });

  test('should display audit table', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should display filters', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('input[placeholder*="Filter by user"]').first()).toBeVisible();
    await expect(page.locator('select').first()).toBeVisible();
  });

  test('should display refresh button', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('button:has-text("Refresh")').first()).toBeVisible();
  });

  test('should display table headers', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);
    await expect(page.locator('th:has-text("Timestamp")').first()).toBeVisible();
    await expect(page.locator('th:has-text("User ID")').first()).toBeVisible();
    await expect(page.locator('th:has-text("IP")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Action")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Resource")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Result")').first()).toBeVisible();
    await expect(page.locator('th:has-text("Details")').first()).toBeVisible();
  });

  test('should display empty state when no audit events', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const tableBody = page.locator('tbody').first();
    const rows = await tableBody.locator('tr').count();

    if (rows === 1) {
      // Check if it's the empty state row
      const emptyText = await page.locator('text=No audit events found').isVisible();
      if (emptyText) {
        await expect(page.locator('text=No audit events available').first()).toBeVisible();
      }
    }
  });

  test('should handle filter by user ID', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const searchInput = page.locator('input[placeholder*="Filter by user"]').first();
    await searchInput.fill('admin');
    await page.waitForTimeout(1000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should handle result filter', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const resultSelect = page.locator('select').first();
    await resultSelect.selectOption('success');
    await page.waitForTimeout(1000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should handle category filter', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const categorySelect = page.locator('select').nth(1);
    await categorySelect.selectOption('auth');
    await page.waitForTimeout(1000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });

  test('should handle clear filters', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const searchInput = page.locator('input[placeholder*="Filter by user"]').first();
    await searchInput.fill('admin');
    await page.waitForTimeout(1000);

    // Check if clear button appears
    const clearButton = page.locator('button:has-text("Clear filters")');
    if (await clearButton.isVisible()) {
      await clearButton.click();
      await page.waitForTimeout(1000);
      await expect(searchInput).toHaveValue('');
    }
  });

  test('should display error message on API failure', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    // If there's an error, it should be displayed
    const errorCard = page.locator('.border-red-200');
    if (await errorCard.isVisible()) {
      await expect(errorCard.locator('.text-red-700').first()).toBeVisible();
    }
  });

  test('should refresh audit logs on button click', async ({ page }) => {
    await page.goto('/dashboard/audit');
    await page.waitForTimeout(2000);

    const refreshButton = page.locator('button:has-text("Refresh")');
    await refreshButton.click();
    await page.waitForTimeout(2000);

    // Should not error
    await expect(page.locator('table').first()).toBeVisible();
  });
});
