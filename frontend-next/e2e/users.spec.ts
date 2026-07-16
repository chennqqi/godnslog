import { test, expect } from '@playwright/test';

test.describe('Users Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/users')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: '1', username: 'admin', email: 'admin@godnslog.com', role: 0, created_at: new Date().toISOString() },
                { id: '2', username: 'tester', email: 'tester@godnslog.com', role: 2, created_at: new Date().toISOString() },
              ],
              total: 2,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/users');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display users page', async ({ page }) => {
    await expect(page).toHaveURL('/users');
  });

  test('should display user management title', async ({ page }) => {
    await expect(page.getByText('User Management')).toBeVisible({ timeout: 5000 });
  });

  test('should display user list heading', async ({ page }) => {
    await expect(page.getByText('User List')).toBeVisible({ timeout: 5000 });
  });

  test('should display users in table', async ({ page }) => {
    await expect(page.getByText('admin').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('tester').first()).toBeVisible();
  });

  test('should display new user button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create User' })).toBeVisible({ timeout: 5000 });
  });
});
