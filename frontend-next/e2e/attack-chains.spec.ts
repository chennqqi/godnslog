import { test, expect } from '@playwright/test';

test.describe('Attack Chains Page', () => {
  const mockChains = {
    code: 0,
    data: {
      items: [
        {
          token: 'jndi-test.dnslog.fun',
          interaction_count: 5,
          protocols: ['dns', 'http', 'ldap'],
          exploit_types: ['log4shell'],
          first_seen: '2026-07-11T10:00:01Z',
          last_seen: '2026-07-11T10:00:05Z',
          confidence: 'high',
        },
        {
          token: 'xxe-test.dnslog.fun',
          interaction_count: 2,
          protocols: ['dns', 'http'],
          exploit_types: ['xxe'],
          first_seen: '2026-07-11T09:00:00Z',
          last_seen: '2026-07-11T09:00:01Z',
          confidence: 'medium',
        },
      ],
      total: 2,
      page: 1,
      page_size: 20,
      total_pages: 1,
    },
  };

  const mockDetail = {
    code: 0,
    data: {
      token: 'jndi-test.dnslog.fun',
      interaction_count: 3,
      protocols: ['dns', 'http'],
      exploit_types: ['log4shell'],
      first_seen: '2026-07-11T10:00:01Z',
      last_seen: '2026-07-11T10:00:03Z',
      confidence: 'high',
      interactions: [
        {
          id: 'int-1',
          type: 'dns',
          timestamp: '2026-07-11T10:00:01Z',
          source_ip: '1.2.3.4',
          domain: 'jndi-test.dnslog.fun',
          exploit_type: 'log4shell',
          confidence: 'high',
        },
        {
          id: 'int-2',
          type: 'http',
          timestamp: '2026-07-11T10:00:02Z',
          source_ip: '5.6.7.8',
          method: 'GET',
          path: '/shell',
          exploit_type: 'log4shell',
          decoded_data: 'admin:true',
          confidence: 'high',
        },
        {
          id: 'int-3',
          type: 'ldap',
          timestamp: '2026-07-11T10:00:03Z',
          source_ip: '9.10.11.12',
          exploit_type: 'log4shell',
          confidence: 'high',
        },
      ],
    },
  };

  test.beforeEach(async ({ context, page }) => {
    await context.addInitScript(() => {
      localStorage.setItem('token', 'mock-token');
      localStorage.setItem('user', JSON.stringify({
        id: 1, username: 'admin', email: 'admin@test.com', role: 0, lang: 'en-US',
      }));
    });
  });

  test('should display attack chain list', async ({ page }) => {
    await page.route('**/api/v2/attack-chains?page=1&page_size=20', route => {
      return route.fulfill({ json: mockChains });
    });
    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/interactions/attack-chains');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Check chain token is visible
    await expect(page.getByText('jndi-test.dnslog.fun')).toBeVisible();
    await expect(page.getByText('xxe-test.dnslog.fun')).toBeVisible();
    // Check protocol badges
    await expect(page.getByText('DNS').first()).toBeVisible();
    await expect(page.getByText('HTTP').first()).toBeVisible();
    // Check exploit type tags
    await expect(page.getByText('log4shell').first()).toBeVisible();
    await expect(page.getByText('xxe')).toBeVisible();
  });

  test('should navigate to attack chain detail', async ({ page }) => {
    // Initial list response
    await page.route('**/api/v2/attack-chains?page=1&page_size=20', route => {
      return route.fulfill({ json: mockChains });
    });
    // Detail response when clicking first chain
    await page.route('**/api/v2/attack-chains/*', route => {
      return route.fulfill({ json: mockDetail });
    });
    // Fallback for other API calls
    await page.route('**/api/**', route => {
      return route.fulfill({ json: { code: 0, data: {} } });
    });

    await page.goto('/interactions/attack-chains');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    // Click the first chain
    await page.getByText('jndi-test.dnslog.fun').click();
    await page.waitForTimeout(2000);

    // Should show detail view
    await expect(page.getByText('Back to chains')).toBeVisible();
    // Should show interaction count
    await expect(page.getByText('3 interactions')).toBeVisible();
    // Should show decoded data
    await expect(page.getByText('admin:true')).toBeVisible();
  });

  test('should show empty state when no chains exist', async ({ page }) => {
    await page.route('**/api/**', route => {
      return route.fulfill({
        json: { code: 0, data: { items: [], total: 0, page: 1, page_size: 20, total_pages: 0 } },
      });
    });

    await page.goto('/interactions/attack-chains');
    await page.waitForLoadState('networkidle');
    await page.waitForTimeout(2000);

    await expect(page.getByText('No Attack Chains')).toBeVisible();
  });
});
