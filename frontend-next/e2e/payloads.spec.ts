import { test, expect } from '@playwright/test';

test.describe('Payloads Page', () => {
  test.beforeEach(async ({ page }) => {
    // Mock API responses - only intercept actual API calls
    await page.route('**/api/v2/payloads**', route => route.fulfill({
      json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } }
    }))
    // Set token before navigation to avoid redirect to login
    await page.goto('/')
    await page.evaluate(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });
    await page.goto('/dashboard/payloads')
    await page.waitForLoadState('domcontentloaded')
    await page.waitForTimeout(5000);
  });

  test('should display payloads page', async ({ page }) => {
    // Debug: check page content
    console.log('Page URL:', page.url());
    const bodyText = await page.evaluate(() => document.body.innerText);
    console.log('Page body text:', bodyText);
    await expect(page.locator('h2')).toContainText('Payload Studio');
  });

  test('should display empty state for payloads', async ({ page }) => {
    await expect(page.getByText('No payloads yet')).toBeVisible();
  });

  test('should display search input', async ({ page }) => {
    await expect(page.getByPlaceholder('Search by token or template...')).toBeVisible();
  });

  test('should display payload detail from API', async ({ page, context }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.match(/\/payloads\/payload-1$/)) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              id: 'payload-1',
              token: 'tok-1',
              template: 'ssrf_http',
              rendered_payload: 'https://tok-1.example.com/test',
              status: 'deployed',
              created_at: new Date().toISOString(),
            }
          }
        });
      }
      if (url.includes('/interactions')) {
        return route.fulfill({
          json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 5, total_pages: 0 } }
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/payloads/payload-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);

    // Verify payload detail page displays key elements
    await expect(page.locator('h2').first()).toContainText('ssrf_http');
    await expect(page.locator('text=tok-1').first()).toBeVisible();
    await expect(page.locator('text=https://tok-1.example.com/test').first()).toBeVisible();
  });
});
