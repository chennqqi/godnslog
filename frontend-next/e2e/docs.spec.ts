import { test, expect } from '@playwright/test';

test.describe('Docs Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'zh-CN');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'zh-CN' }));
    });

    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/docs');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display docs page', async ({ page }) => {
    await expect(page).toHaveURL('/docs');
  });

  test('should display docs center title', async ({ page }) => {
    await expect(page.getByText('文档中心')).toBeVisible({ timeout: 5000 });
  });

  test('should display doc cards', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '快速开始' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'API 文档' })).toBeVisible();
    await expect(page.getByRole('heading', { name: '用户指南' })).toBeVisible();
  });

  test('should display help section', async ({ page }) => {
    await expect(page.getByRole('heading', { name: '获取帮助' })).toBeVisible({ timeout: 5000 });
  });

  test('should display view doc buttons', async ({ page }) => {
    await expect(page.getByRole('button', { name: '查看文档' }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should navigate to quick-start subpage', async ({ page }) => {
    await page.getByRole('button', { name: '查看文档' }).first().click();
    await page.waitForURL('**/docs/quick-start', { timeout: 5000 });
    await expect(page).toHaveURL(/\/\/docs\/quick-start/);
  });

  test('should display content on quick-start subpage', async ({ page }) => {
    await page.goto('/docs/quick-start');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: '快速开始' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Start the Server')).toBeVisible({ timeout: 5000 });
  });

  test('should display content on api subpage', async ({ page }) => {
    await page.goto('/docs/api');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: 'API 文档' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Authentication')).toBeVisible({ timeout: 5000 });
  });

  test('should display content on user-guide subpage', async ({ page }) => {
    await page.goto('/docs/user-guide');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: '用户指南' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('heading', { name: 'Cases', exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('should display content on config subpage', async ({ page }) => {
    await page.goto('/docs/config');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: '配置参考' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Command Line Options')).toBeVisible({ timeout: 5000 });
  });

  test('should display content on faq subpage', async ({ page }) => {
    await page.goto('/docs/faq');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: '常见问题' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('How do I change the admin password?')).toBeVisible({ timeout: 5000 });
  });

  test('should display content on security subpage', async ({ page }) => {
    await page.goto('/docs/security');
    await page.waitForLoadState('networkidle');
    await expect(page.getByRole('heading', { name: '安全指南' })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Authentication')).toBeVisible({ timeout: 5000 });
  });

  test('should navigate back to docs index from subpage', async ({ page }) => {
    await page.goto('/docs/quick-start');
    await page.waitForLoadState('networkidle');
    await page.getByRole('button', { name: /文档中心/ }).click();
    await page.waitForURL('**/docs', { timeout: 5000 });
    await expect(page).toHaveURL(/\/\/docs$/);
  });
});
