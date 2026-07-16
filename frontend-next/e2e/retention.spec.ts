import { test, expect } from '@playwright/test';

test.describe('Retention Page', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'en-US');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      const method = route.request().method();
      if (url.includes('/retention/policies') && method === 'GET') {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'ret-1', name: 'Default 90-day', description: 'Keep data for 90 days', retention_days: 90, max_records: 0, is_enabled: true, apply_to_interactions: true, apply_to_cases: false, apply_to_payloads: false, apply_to_evidence: false, apply_to_logs: false, run_daily: true, run_hourly: false, run_weekly: false, run_monthly: false, last_run_at: '2026-06-21T10:00:00Z' },
                { id: 'ret-2', name: 'Weekly Cleanup', description: 'Clean old cases weekly', retention_days: 365, max_records: 10000, is_enabled: false, apply_to_interactions: false, apply_to_cases: true, apply_to_payloads: true, apply_to_evidence: false, apply_to_logs: false, run_daily: false, run_hourly: false, run_weekly: true, run_monthly: false, last_run_at: '' },
              ],
              total: 2,
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
                { id: 'job-1', policy_id: 'ret-1', status: 'completed', started_at: '2026-06-21T10:00:00Z', finished_at: '2026-06-21T10:05:00Z', records_processed: 150, records_deleted: 30, duration: 5000 },
                { id: 'job-2', policy_id: 'ret-1', status: 'running', started_at: '2026-06-22T10:00:00Z', finished_at: '', records_processed: 0, records_deleted: 0, duration: 0 },
              ],
              total: 2,
              page: 1,
              page_size: 20,
              total_pages: 1,
            },
          },
        });
      }
      if (url.includes('/retention/policies') && (method === 'POST' || method === 'PUT' || method === 'DELETE')) {
        return route.fulfill({ json: { code: 0, data: { id: 'ret-new' } } });
      }
      if (url.includes('/retention/policies') && url.includes('/run')) {
        return route.fulfill({ json: { code: 0, data: { id: 'job-new' } } });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/retention');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display retention page', async ({ page }) => {
    await expect(page).toHaveURL('/retention');
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

  test('should display second policy name', async ({ page }) => {
    await expect(page.getByText('Weekly Cleanup')).toBeVisible({ timeout: 5000 });
  });

  test('should display policy description', async ({ page }) => {
    await expect(page.getByText('Keep data for 90 days')).toBeVisible({ timeout: 5000 });
  });

  test('should display enabled badge for active policy', async ({ page }) => {
    await expect(page.getByText('Enabled').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display disabled badge for inactive policy', async ({ page }) => {
    await expect(page.getByText('Disabled').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display retention days info', async ({ page }) => {
    await expect(page.getByText('Retain: 90d')).toBeVisible({ timeout: 5000 });
  });

  test('should display max records info when set', async ({ page }) => {
    await expect(page.getByText('Max: 10000')).toBeVisible({ timeout: 5000 });
  });

  test('should display apply-to badges', async ({ page }) => {
    await expect(page.getByText('Interactions', { exact: true }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should display schedule badges', async ({ page }) => {
    await expect(page.getByText('Daily', { exact: true }).first()).toBeVisible({ timeout: 5000 });
  });

  test('should display recent jobs section', async ({ page }) => {
    await expect(page.getByText('Recent Jobs')).toBeVisible({ timeout: 5000 });
  });

  test('should display job status badges', async ({ page }) => {
    await expect(page.getByText('completed').first()).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('running').first()).toBeVisible({ timeout: 5000 });
  });

  test('should display job processed/deleted info', async ({ page }) => {
    await expect(page.getByText(/Processed: 150.*Deleted: 30/)).toBeVisible({ timeout: 5000 });
  });

  test('should display action buttons for policies', async ({ page }) => {
    const runButtons = page.getByRole('button', { name: 'Run' });
    await expect(runButtons.first()).toBeVisible({ timeout: 5000 });
    const editButtons = page.getByRole('button', { name: 'Edit' });
    await expect(editButtons.first()).toBeVisible({ timeout: 5000 });
    const deleteButtons = page.getByRole('button', { name: 'Delete' });
    await expect(deleteButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should display disable button for enabled policy', async ({ page }) => {
    const disableButtons = page.getByRole('button', { name: 'Disable' });
    await expect(disableButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should display enable button for disabled policy', async ({ page }) => {
    const enableButtons = page.getByRole('button', { name: 'Enable' });
    await expect(enableButtons.first()).toBeVisible({ timeout: 5000 });
  });

  test('should open create dialog with form fields', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Name', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Description', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Retention Days', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Apply To', { exact: true })).toBeVisible({ timeout: 5000 });
    await expect(page.getByText('Schedule', { exact: true })).toBeVisible({ timeout: 5000 });
  });

  test('should have name input in create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const nameInput = page.locator('#name');
    await expect(nameInput).toBeVisible({ timeout: 5000 });
  });

  test('should have retention days input with default value', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const retentionInput = page.locator('#retention_days');
    await expect(retentionInput).toHaveValue('90', { timeout: 5000 });
  });

  test('should have save button in create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const saveButton = page.locator('[role="dialog"]').getByRole('button', { name: 'Save' });
    await expect(saveButton).toBeVisible({ timeout: 5000 });
  });

  test('should have cancel button in create dialog', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const cancelButton = page.locator('[role="dialog"]').getByRole('button', { name: 'Cancel' });
    await expect(cancelButton).toBeVisible({ timeout: 5000 });
  });

  test('should close dialog when clicking cancel', async ({ page }) => {
    await page.getByRole('button', { name: 'Create Policy' }).click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const cancelButton = page.locator('[role="dialog"]').getByRole('button', { name: 'Cancel' });
    await cancelButton.click();
    await expect(page.getByRole('dialog')).not.toBeVisible({ timeout: 5000 });
  });

  test('should open edit dialog with pre-filled values', async ({ page }) => {
    const editButtons = page.getByRole('button', { name: 'Edit' });
    await editButtons.first().click();
    await expect(page.getByRole('dialog')).toBeVisible({ timeout: 5000 });
    const nameInput = page.locator('#name');
    await expect(nameInput).toHaveValue('Default 90-day', { timeout: 5000 });
  });

  test('should display policy count in section header', async ({ page }) => {
    await expect(page.getByText(/2 policies configured/)).toBeVisible({ timeout: 5000 });
  });

  test('should display last run info for policy', async ({ page }) => {
    await expect(page.getByText(/Last run:/)).toBeVisible({ timeout: 5000 });
  });
});

test.describe('Retention Page - Empty State', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'en-US');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/retention/policies')) {
        return route.fulfill({
          json: {
            code: 0,
            data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 },
          },
        });
      }
      if (url.includes('/retention/jobs')) {
        return route.fulfill({
          json: {
            code: 0,
            data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 },
          },
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/retention');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display empty state message', async ({ page }) => {
    await expect(page.getByText('No retention policies yet')).toBeVisible({ timeout: 5000 });
  });

  test('should display create policy button in empty state', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Policy' })).toBeVisible({ timeout: 5000 });
  });

  test('should not display recent jobs section when empty', async ({ page }) => {
    await expect(page.getByText('Recent Jobs')).not.toBeVisible({ timeout: 5000 });
  });
});

test.describe('Retention Page - API Error', () => {
  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('language', 'en-US');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/retention')) {
        return route.fulfill({ status: 500, json: { code: 500, message: 'Internal Server Error' } });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/retention');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should still display page title on API error', async ({ page }) => {
    await expect(page.getByText('Data Retention', { exact: true })).toBeVisible({ timeout: 10000 });
  });

  test('should still display create button on API error', async ({ page }) => {
    await expect(page.getByRole('button', { name: 'Create Policy' })).toBeVisible({ timeout: 5000 });
  });
});
