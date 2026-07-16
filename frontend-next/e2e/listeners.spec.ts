import { test, expect } from '@playwright/test';

test.describe('Listeners Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'en-US');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      const method = route.request().method();
      if (url.includes('/listeners') && method === 'GET') {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'lst-1', protocol: 'smtp', host: '0.0.0.0', port: 25, token: 'tok-abc', is_enabled: true },
                { id: 'lst-2', protocol: 'ldap', host: '0.0.0.0', port: 389, token: 'tok-def', is_enabled: false },
              ],
              total: 2,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        });
      }
      if (url.includes('/listeners') && (method === 'POST' || method === 'PUT' || method === 'DELETE')) {
        return route.fulfill({ json: { code: 0, data: { id: 'lst-new' } } });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/listeners');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display listeners page', async ({ page }) => {
    await expect(page).toHaveURL('/listeners');
  });

  test('should display protocol listeners title', async ({ page }) => {
    await expect(page.getByText('Protocol Listeners', { exact: true })).toBeVisible({ timeout: 10000 });
  });

  test('should display create listener button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Listener' })).toBeVisible({ timeout: 5000 });
  });

  test('should display listener cards', async ({ page }) => {
    await expect(page.getByText('SMTP').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('LDAP').first()).toBeVisible();
  });

  test('should display listener status badges', async ({ page }) => {
    await expect(page.getByText('Running').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Stopped').first()).toBeVisible();
  });

  test('should display listener host and port info', async ({ page }) => {
    await expect(page.getByText('0.0.0.0:25')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('0.0.0.0:389')).toBeVisible({ timeout: 5000 });
  });

  test('should display listener tokens', async ({ page }) => {
    await expect(page.getByText('tok-abc')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('tok-def')).toBeVisible({ timeout: 5000 });
  });

  test('should open create dialog when clicking create button', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Listener' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Protocol', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Host', { exact: true })).toBeVisible();
    await expect(page.getByText('Port', { exact: true })).toBeVisible();
  });

  test('should have protocol select in create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Listener' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const protocolSelect = page.locator('[role="dialog"]').getByRole('combobox');
    await expect(protocolSelect).toBeVisible({ timeout: 5000 });
  });

  test('should have host input with default value in create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Listener' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const hostInput = page.locator('[role="dialog"] input[value="0.0.0.0"]');
    await expect(hostInput).toBeVisible({ timeout: 5000 });
  });

  test('should have save button in create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Listener' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    await expect(page.getByRole('button', { name: 'Save' })).toBeVisible({ timeout: 5000 });
  });

  test('should close dialog when clicking cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Listener' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    await page.getByRole('button', { name: 'Cancel' }).click();
    await expect(page.getByRole('dialog')).not.toBeVisible({ timeout: 5000 });
  });

  test('should display edit buttons for listeners', async ({ page }) => {
    const editButtons = page.getByRole('button', { name: /Edit/i });
    await expect(editButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should display delete buttons for listeners', async ({ page }) => {
    const deleteButtons = page.getByRole('button', { name: /Delete/i });
    await expect(deleteButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should show toggle button for enabled listener', async ({ page }) => {
    const stopButtons = page.getByRole('button', { name: /Stop/i });
    await expect(stopButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should show toggle button for disabled listener', async ({ page }) => {
    const startButtons = page.getByRole('button', { name: /Start/i });
    await expect(startButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should open edit dialog with pre-filled values', async ({ page }) => {
    const editButtons = page.getByRole('button', { name: /Edit/i });
    await editButtons.first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const hostInput = page.locator('[role="dialog"] input[value="0.0.0.0"]');
    await expect(hostInput).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Listeners Page - Empty State', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'en-US');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/listeners')) {
        return route.fulfill({
          json: {
            code: 0,
            data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/listeners');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display empty state with create button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Listener' })).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Listeners Page - API Error', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'en-US');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/listeners')) {
        return route.fulfill({ status: 500, json: { code: 500, message: 'Internal Server Error' } });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/listeners');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display error message on API failure', async ({ page }) => {
    await expect(page.getByText(/Failed to load|error/i).first()).toBeVisible({ timeout: 10000 });
  });
});
