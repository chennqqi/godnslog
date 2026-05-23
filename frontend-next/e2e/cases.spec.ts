import { test, expect } from '@playwright/test';

test.describe('Cases Board', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock cases API before navigation
    await page.route('**/api/v2/cases**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          items: [
            { id: 'case-1', title: 'Test Case', description: 'Test description', status: 'active', created_at: new Date().toISOString() }
          ],
          total: 1,
          page: 1,
          page_size: 20,
          total_pages: 1
        }
      }
    }));

    await page.goto('/dashboard/cases');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display cases page', async ({ page }) => {
    await expect(page.locator('h2').first()).toContainText('Case Board');
  });

  test('should show create case button', async ({ page }) => {
    const createButton = page.locator('button').filter({ hasText: 'New Case' }).first();
    await expect(createButton).toBeVisible();
  });

  test('should open create case modal', async ({ page }) => {
    const createButton = page.locator('button').filter({ hasText: 'New Case' }).first();
    await createButton.click();
    await expect(page.getByRole('heading', { name: 'New Case' })).toBeVisible();
  });

  test('should display search input', async ({ page }) => {
    const searchInput = page.locator('input[placeholder*="Search"]');
    await expect(searchInput).toBeVisible();
  });

  test('should display status filter', async ({ page }) => {
    // Radix Select uses a button trigger, not a native select element
    const statusFilterTrigger = page.locator('button').filter({ hasText: 'All statuses' }).first();
    await expect(statusFilterTrigger).toBeVisible();
  });

  test('should not display batch operations', async ({ page }) => {
    // Batch operations should not exist in the simplified Cases Board
    const batchDeleteButton = page.locator('button').filter({ hasText: 'Delete selected' });
    await expect(batchDeleteButton).not.toBeVisible();
  });

  test('should not display edit/delete buttons in list', async ({ page }) => {
    // Edit and delete buttons should not exist in the simplified Cases Board
    const editButton = page.locator('button').filter({ hasText: 'Edit' });
    const deleteButton = page.locator('button').filter({ hasText: 'Delete' });
    await expect(editButton).not.toBeVisible();
    await expect(deleteButton).not.toBeVisible();
  });

  test('should navigate to case detail on click', async ({ page }) => {
    // Click on the case row
    const caseRow = page.locator('li').filter({ hasText: 'Test Case' }).first();
    await caseRow.click();

    // Should navigate to case detail
    await page.waitForURL('**/dashboard/cases/case-1');
    expect(page.url()).toContain('/dashboard/cases/case-1');
  });
});

test.describe('Case Detail', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock case detail API
    await page.route('**/api/v2/cases/case-1**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          id: 'case-1',
          title: 'Test Case',
          description: 'Test description',
          target: 'example.com',
          status: 'active',
          created_at: new Date().toISOString()
        }
      }
    }));

    // Mock case stats API
    await page.route('**/api/v2/cases/case-1/stats**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          payload_count: 5,
          interaction_count: 12,
          hit_payload_count: 3
        }
      }
    }));

    // Mock payloads API
    await page.route('**/api/v2/payloads**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          items: [
            { id: 'payload-1', token: 'gdl_abc123', template: 'ssrf_http', status: 'deployed', created_at: new Date().toISOString() }
          ],
          total: 1,
          page: 1,
          page_size: 20,
          total_pages: 1
        }
      }
    }));

    await page.goto('/dashboard/cases/case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display case detail', async ({ page }) => {
    await expect(page.locator('h2').first()).toContainText('Test Case');
  });

  test('should display case stats', async ({ page }) => {
    await expect(page.locator('text=Payloads').first()).toBeVisible();
    await expect(page.locator('text=Interactions').first()).toBeVisible();
    await expect(page.locator('text=Hit Payloads').first()).toBeVisible();
  });

  test('should display payloads list', async ({ page }) => {
    await expect(page.locator('text=Payloads').first()).toBeVisible();
    await expect(page.locator('text=gdl_abc123').first()).toBeVisible();
  });

  test('should display create payload button', async ({ page }) => {
    const createButton = page.locator('button').filter({ hasText: 'Create Payload' }).first();
    await expect(createButton).toBeVisible();
  });

  test('should display quick actions', async ({ page }) => {
    await expect(page.locator('text=Quick Actions').first()).toBeVisible();
    await expect(page.locator('button').filter({ hasText: 'View Evidence' }).first()).toBeVisible();
    await expect(page.locator('button').filter({ hasText: 'View Interactions' }).first()).toBeVisible();
  });

  test('should navigate to new payload with case_id', async ({ page }) => {
    const createButton = page.locator('button').filter({ hasText: 'Create Payload' }).first();
    await createButton.click();

    // Should navigate to new payload page with case_id
    await page.waitForURL('**/dashboard/payloads/new?case_id=case-1');
    expect(page.url()).toContain('case_id=case-1');
  });
});

test.describe('New Payload', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock cases API
    await page.route('**/api/v2/cases**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          items: [
            { id: 'case-1', title: 'Test Case', description: 'Test description', status: 'active', created_at: new Date().toISOString() }
          ],
          total: 1,
          page: 1,
          page_size: 20,
          total_pages: 1
        }
      }
    }));
  });

  test('should display new payload page', async ({ page }) => {
    await page.goto('/dashboard/payloads/new');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2').first()).toContainText('New Payload');
  });

  test('should display associated case when case_id is provided', async ({ page }) => {
    await page.goto('/dashboard/payloads/new?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=Creating for Case').first()).toBeVisible();
    await expect(page.locator('text=Test Case').first()).toBeVisible();
  });

  test('should display step indicator', async ({ page }) => {
    await page.goto('/dashboard/payloads/new');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=Choose a template').first()).toBeVisible();
  });

  test('should display template selection', async ({ page }) => {
    await page.goto('/dashboard/payloads/new');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('text=SSRF HTTP').first()).toBeVisible();
  });
});

test.describe('Payload Detail', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock payload API
    await page.route('**/api/v2/payloads/payload-1**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          id: 'payload-1',
          token: 'gdl_abc123',
          template: 'ssrf_http',
          rendered_payload: 'http://gdl_abc123.example.com/test',
          status: 'hit',
          case_id: 'case-1',
          created_at: new Date().toISOString(),
          expires_at: new Date(Date.now() + 86400000).toISOString()
        }
      }
    }));

    // Mock associated case API
    await page.route('**/api/v2/cases/case-1**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          id: 'case-1',
          title: 'Test Case',
          description: 'Test description',
          target: 'example.com',
          status: 'active',
          created_at: new Date().toISOString()
        }
      }
    }));

    // Mock interactions API
    await page.route('**/api/v2/interactions**', route => route.fulfill({
      json: {
        code: 0,
        data: {
          items: [
            { id: 'int-1', type: 'dns', source_ip: '1.2.3.4', timestamp: new Date().toISOString() }
          ],
          total: 1,
          page: 1,
          page_size: 5,
          total_pages: 1
        }
      }
    }));

    await page.goto('/dashboard/payloads/payload-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
  });

  test('should display payload detail', async ({ page }) => {
    await expect(page.locator('h2').first()).toContainText('ssrf_http');
  });

  test('should display token', async ({ page }) => {
    await expect(page.locator('text=gdl_abc123').first()).toBeVisible();
  });

  test('should display rendered payload', async ({ page }) => {
    await expect(page.locator('text=http://gdl_abc123.example.com/test').first()).toBeVisible();
  });

  test('should display associated case', async ({ page }) => {
    await expect(page.locator('text=关联 Case').first()).toBeVisible();
    await expect(page.locator('text=Test Case').first()).toBeVisible();
  });

  test('should display recent interactions', async ({ page }) => {
    await expect(page.locator('text=最近交互').first()).toBeVisible();
    await expect(page.locator('text=DNS').first()).toBeVisible();
  });

  test('should display quick actions', async ({ page }) => {
    await expect(page.locator('text=快速操作').first()).toBeVisible();
    await expect(page.locator('button').filter({ hasText: '查看交互' }).first()).toBeVisible();
    await expect(page.locator('button').filter({ hasText: '查看证据' }).first()).toBeVisible();
  });

  test('should not display revoke button', async ({ page }) => {
    // Revoke button should not exist in the simplified Payload Detail
    const revokeButton = page.locator('button').filter({ hasText: '撤销' });
    await expect(revokeButton).not.toBeVisible();
  });

  test('should navigate to interactions on quick action click', async ({ page }) => {
    const interactionsButton = page.locator('button').filter({ hasText: '查看交互' }).first();
    await interactionsButton.click();
    await page.waitForURL('**/dashboard/interactions?payload_id=payload-1');
    expect(page.url()).toContain('payload_id=payload-1');
  });
});
