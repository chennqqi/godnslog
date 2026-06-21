import { test, expect } from '@playwright/test';

test.describe('Docs Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/docs');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display docs page', async ({ page }) => {
    await expect(page).toHaveURL('/dashboard/docs');
  });

  test('should display docs center title', async ({ page }) => {
    await expect(page.getByText('文档中心')).toBeVisible({ timeout: 5000 });
  });

  test('should display doc cards', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '快速开始' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'API文档' })).toBeVisible();
    await expect(page.getByRole('heading', { name: '用户指南' })).toBeVisible();
  });

  test('should display help section', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '获取帮助' })).toBeVisible({ timeout: 5000 });
  });

  test('should display view doc buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: '查看文档' }).first()).toBeVisible({ timeout: 5000 });
  });
});
