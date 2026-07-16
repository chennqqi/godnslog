import { test, expect } from '@playwright/test';

test.describe('Evidence Summary Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/cases') && !url.includes('/stats') && !url.match(/\/cases\/[^/]+$/)) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'case-1', title: 'SSRF Test', status: 'active', created_at: new Date().toISOString() },
              ],
              total: 1,
              page: 1,
              page_size: 100,
              total_pages: 1,
            },
          },
        });
      }
      if (url.includes('/scanner-runs')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'run-1', scanner_type: 'nuclei', status: 'completed', created_at: new Date().toISOString() },
              ],
              total: 1,
              page: 1,
              page_size: 100,
              total_pages: 1,
            },
          },
        });
      }
      if (url.includes('/evidence/summary')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              scope: { type: 'case', id: 'case-1' },
              summary: {
                total_interactions: 5,
                unique_sources: 2,
                high_risk_count: 1,
                evidence_strength: 'high',
              },
              interactions: [],
              payloads: [],
            },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/evidence-summary');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display evidence summary page', async ({ page }) => {
    await expect(page).toHaveURL('/evidence-summary');
  });

  test('should display evidence summary title', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Evidence Summary' })).toBeVisible({ timeout: 5000 });
  });

  test('should display query scope section', async ({ page }) => {
    await expect(page.getByText('Query Scope')).toBeVisible({ timeout: 5000 });
  });

  test('should display scope type selector', async ({ page }) => {
    await expect(page.getByText('Scope Type')).toBeVisible({ timeout: 5000 });
  });

  test('should display case selector', async ({ page }) => {
    await expect(page.getByText('Select Case')).toBeVisible({ timeout: 5000 });
  });
});
