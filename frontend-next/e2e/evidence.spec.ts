import { test, expect } from '@playwright/test';

test.describe('Evidence Page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login');
    await page.fill('input[name="username"]', 'admin');
    await page.fill('input[name="password"]', 'test123');
    await page.click('button[type="submit"]');
    await page.waitForURL('/dashboard', { timeout: 10000 });
  });

  test('should display evidence page', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    await expect(page.locator('h2').first()).toContainText('证据报告');
  });

  test('should display generate evidence section', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=生成证据').first()).toBeVisible();
  });

  test('should display case selector', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    await expect(page.locator('select').first()).toBeVisible();
  });

  test('should display format selector with json and markdown options', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    const formatSelect = page.locator('select').nth(1);
    await expect(formatSelect).toBeVisible();
    const options = await formatSelect.locator('option').allTextContents();
    expect(options).toContain('Markdown');
    expect(options).toContain('JSON');
    // CSV should not be present in the new unified API
    expect(options).not.toContain('CSV');
  });

  test('should not display include_raw checkbox', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);
    // The old include_raw checkbox should not exist
    await expect(page.locator('text=包含原始数据')).not.toBeVisible();
  });

  test('should display evidence summary after generation', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);

    // Select a case if available
    const caseSelect = page.locator('select').first();
    const options = await caseSelect.locator('option').all();
    if (options.length > 1) {
      await caseSelect.selectOption({ index: 1 });

      // Click generate button
      await page.click('button:has-text("生成证据")');
      await page.waitForTimeout(3000);

      // Check for evidence summary section
      await expect(page.locator('text=证据摘要').first()).toBeVisible();
    }
  });

  test('should display explainability section', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);

    const caseSelect = page.locator('select').first();
    const options = await caseSelect.locator('option').all();
    if (options.length > 1) {
      await caseSelect.selectOption({ index: 1 });
      await page.click('button:has-text("生成证据")');
      await page.waitForTimeout(3000);

      await expect(page.locator('text=可解释性').first()).toBeVisible();
    }
  });

  test('should display timeline section', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);

    const caseSelect = page.locator('select').first();
    const options = await caseSelect.locator('option').all();
    if (options.length > 1) {
      await caseSelect.selectOption({ index: 1 });
      await page.click('button:has-text("生成证据")');
      await page.waitForTimeout(3000);

      await expect(page.locator('text=时间线').first()).toBeVisible();
    }
  });

  test('should handle no evidence scenario gracefully', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForTimeout(2000);

    const caseSelect = page.locator('select').first();
    const options = await caseSelect.locator('option').all();
    if (options.length > 1) {
      await caseSelect.selectOption({ index: 1 });
      await page.click('button:has-text("生成证据")');
      await page.waitForTimeout(3000);

      // If no evidence, should show error message
      const errorElement = page.locator('.bg-red-50, .text-red-700');
      if (await errorElement.isVisible()) {
        await expect(errorElement).toContainText('未找到');
      }
    }
  });
});
