import { test, expect } from '@playwright/test';

test.describe('Dashboard', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock all API endpoints
    await page.route('**/api/**', route => {
      const url = route.request().url();

      if (url.includes('/interactions/stats')) {
        return route.fulfill({
          json: { code: 0, data: { today: 3, total: 10, high_risk: 1 } }
        });
      }
      if (url.includes('/cases') && !url.includes('/stats') && !url.match(/\/cases\/[^/]+$/)) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'case-1', title: 'SSRF Test', description: 'Test SSRF vulnerabilities', status: 'active', created_at: new Date().toISOString() }
              ],
              total: 1,
              page: 1,
              page_size: 20,
              total_pages: 1
            }
          }
        });
      }
      if (url.includes('/interactions')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'int-1', type: 'dns', source_ip: '1.2.3.4', token: 'gdl_abc', created_at: new Date().toISOString() }
              ],
              total: 1,
              page: 1,
              page_size: 20,
              total_pages: 1
            }
          }
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display dashboard page', async ({ page }) => {
    await expect(page).toHaveURL('/');
  });

  test('should display stats from API', async ({ page }) => {
    // Stats should be visible somewhere on the page
    await expect(page.getByText('3').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display navigation sidebar', async ({ page }) => {
    // Check for sidebar navigation items
    await expect(page.locator('nav').first()).toBeVisible();
  });

  test('should redirect to login when no auth token exists', async ({ browser }) => {
    // Use a fresh context without the mock auth token
    const context = await browser.newContext()
    const page = await context.newPage()

    // Mock API to avoid errors
    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } })
    })

    // Navigate to dashboard - should redirect to login
    await page.goto('/')
    await page.waitForURL('**/login', { timeout: 10000 })
    await expect(page).toHaveURL(/\/login/)

    await context.close()
  })
});
