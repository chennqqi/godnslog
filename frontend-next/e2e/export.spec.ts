import { test, expect } from '@playwright/test';

test.describe('Export Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/export');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display export page', async ({ page }) => {
    await expect(page).toHaveURL('/export');
  });

  test('should display export title', async ({ page }) => {
    await expect(page.getByText('证据导出')).toBeVisible({ timeout: 5000 });
  });

  test('should display export config card', async ({ page }) => {
    await expect(page.getByText('导出配置')).toBeVisible({ timeout: 5000 });
  });

  test('should display format selector', async ({ page }) => {
    await expect(page.getByText('导出格式')).toBeVisible({ timeout: 5000 });
  });

  test('should display export button', async ({ page }) => {
    await expect(page.getByRole('button', { name: '开始导出' })).toBeVisible({ timeout: 5000 });
  });
});
