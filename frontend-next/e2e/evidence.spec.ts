import { test, expect } from '@playwright/test';

test.describe('Evidence Page', () => {
  test.beforeEach(async ({ context, page }) => {
    // Set token in localStorage before any page loads
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
    });

    // Mock API responses
    await page.route('**/api/**', route => {
      const url = route.request().url();
      if (url.includes('/cases')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              items: [
                { id: 'case-1', title: 'Test Case', status: 'active', created_at: new Date().toISOString() }
              ],
              total: 1,
              page: 1,
              page_size: 20,
              total_pages: 1
            }
          }
        });
      }
      if (url.includes('/evidence/generate')) {
        return route.fulfill({
          json: {
            code: 0,
            data: {
              evidence: {
                evidence_strength: 'high',
                confidence: 85,
                interaction_count: 10,
                unique_sources: 5,
                explainability: 'This evidence shows strong correlation between DNS queries and HTTP requests.',
                timeline: [
                  { id: 'int-1', type: 'dns', source_ip: '1.2.3.4', timestamp: '2024-01-01T00:00:00Z', domain: 'example.com' },
                  { id: 'int-2', type: 'http', source_ip: '5.6.7.8', timestamp: '2024-01-01T00:01:00Z', method: 'GET', path: '/test' }
                ]
              },
              content: '# Evidence Report\n\n## Summary\nEvidence strength: high\nConfidence: 85%'
            }
          }
        });
      }
      return route.fulfill({ json: { code: 0, data: {} } });
    });
  });

  test('should display evidence page', async ({ page }) => {
    await page.goto('/dashboard/evidence');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(1000);
    await expect(page.locator('h2').first()).toContainText('证据报告');
  });

  test('should display case scoped evidence', async ({ page }) => {
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Case scoped: case-1')).toBeVisible();
    await expect(page.locator('button:has-text("Clear scope")')).toBeVisible();
  });

  test('should display payload scoped evidence', async ({ page }) => {
    await page.goto('/dashboard/evidence?payload_id=payload-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=Payload scoped: payload-1')).toBeVisible();
    await expect(page.locator('button:has-text("Clear scope")')).toBeVisible();
  });

  test('should auto-generate evidence for case_id', async ({ page }) => {
    const evidenceRequestPromise = page.waitForRequest(request => {
      if (!request.url().includes('/api/v2/evidence/generate')) return false;
      const body = request.postDataJSON() as { case_id?: string; payload_id?: string; format?: string };
      return body.case_id === 'case-1' && !body.payload_id && body.format === 'markdown';
    });

    await page.goto('/dashboard/evidence?case_id=case-1');
    await evidenceRequestPromise;
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=证据摘要')).toBeVisible();
    await expect(page.locator('text=证据强度')).toBeVisible();
  });

  test('should auto-generate evidence for payload_id', async ({ page }) => {
    const evidenceRequestPromise = page.waitForRequest(request => {
      if (!request.url().includes('/api/v2/evidence/generate')) return false;
      const body = request.postDataJSON() as { case_id?: string; payload_id?: string; format?: string };
      return body.payload_id === 'payload-1' && !body.case_id && body.format === 'markdown';
    });

    await page.goto('/dashboard/evidence?payload_id=payload-1');
    await evidenceRequestPromise;
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=证据摘要')).toBeVisible();
    await expect(page.locator('text=置信度')).toBeVisible();
  });

  test('should use requested evidence format from URL scope', async ({ page }) => {
    const evidenceRequestPromise = page.waitForRequest(request => {
      if (!request.url().includes('/api/v2/evidence/generate')) return false;
      const body = request.postDataJSON() as { case_id?: string; format?: string };
      return body.case_id === 'case-1' && body.format === 'json';
    });

    await page.goto('/dashboard/evidence?case_id=case-1&format=json');
    await evidenceRequestPromise;
    await page.waitForLoadState('networkidle');
    await expect(page.locator('text=Case scoped: case-1')).toBeVisible();
  });

  test('should display evidence summary after generation', async ({ page }) => {
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=证据摘要').first()).toBeVisible();
    await expect(page.locator('text=交互数量')).toBeVisible();
    await expect(page.locator('text=唯一来源')).toBeVisible();
  });

  test('should display explainability section', async ({ page }) => {
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=可解释性').first()).toBeVisible();
  });

  test('should display timeline section', async ({ page }) => {
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=时间线').first()).toBeVisible();
  });

  test('should display report preview', async ({ page }) => {
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=报告预览').first()).toBeVisible();
  });

  test('should display no evidence state when no data', async ({ page }) => {
    await page.route('**/api/v2/evidence/generate', route => {
      return route.fulfill({
        json: { code: 404, message: '未找到该证据数据' }
      });
    });
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('text=暂无证据数据')).toBeVisible();
  });

  test('should handle API error gracefully', async ({ page }) => {
    await page.route('**/api/v2/evidence/generate', route => {
      return route.fulfill({
        json: { code: 500, message: '生成证据失败' }
      });
    });
    await page.goto('/dashboard/evidence?case_id=case-1');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);
    await expect(page.locator('.bg-red-50').first()).toBeVisible();
  });
});
