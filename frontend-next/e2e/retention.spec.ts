import { test, expect } from '@playwright/test';

test.describe('Retention Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/retention/policies')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'ret-1', name: 'Default 90-day', description: 'Keep data for 90 days', retention_days: 90, max_records: 0, is_enabled: true, apply_to_interactions: true, apply_to_cases: false, apply_to_payloads: false, apply_to_evidence: false, apply_to_logs: false },
              ],
              total: 1,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        });
      }
      if (url.includes('/retention/jobs')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'job-1', policy_id: 'ret-1', status: 'completed', started_at: new Date().toISOString(), finished_at: new Date().toISOString() },
              ],
              total: 1,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/dashboard/retention');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display retention page', async ({ page }) => {
    await expect(page).toHaveURL('/dashboard/retention');
  });

  test('should display data retention title', async ({ page }) => {
    await expect(page.getByText('Data Retention', { exact: true })).toBeVisible({ timeout: 10000 });
  });

  test('should display create policy button', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Policy' })).toBeVisible({ timeout: 5000 });
  });

  test('should display retention policies section', async ({ page }) => {
    await expect(page.getByText('Retention Policies')).toBeVisible({ timeout: 5000 });
  });

  test('should display policy name in list', async ({ page }) => {
    await expect(page.getByText('Default 90-day')).toBeVisible({ timeout: 5000 });
  });
});
