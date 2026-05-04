# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: dashboard.spec.ts >> Dashboard >> should display dashboard
- Location: e2e\dashboard.spec.ts:12:7

# Error details

```
Test timeout of 30000ms exceeded while running "beforeEach" hook.
```

```
Error: page.fill: Test timeout of 30000ms exceeded.
Call log:
  - waiting for locator('input[type="email"]')

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - generic [ref=e3]:
    - generic [ref=e4]:
      - heading "GODNSLOG 2.0" [level=2] [ref=e5]
      - paragraph [ref=e6]: 登录到您的账户
    - generic [ref=e7]:
      - generic [ref=e8]:
        - generic [ref=e9]:
          - generic [ref=e10]: 用户名
          - textbox "用户名" [ref=e11]
        - generic [ref=e12]:
          - generic [ref=e13]: 密码
          - textbox "密码" [ref=e14]
      - button "登录" [ref=e16] [cursor=pointer]
  - alert [ref=e17]
```

# Test source

```ts
  1  | import { test, expect } from '@playwright/test';
  2  | 
  3  | test.describe('Dashboard', () => {
  4  |   test.beforeEach(async ({ page }) => {
  5  |     await page.goto('/login');
> 6  |     await page.fill('input[type="email"]', 'admin@example.com');
     |                ^ Error: page.fill: Test timeout of 30000ms exceeded.
  7  |     await page.fill('input[type="password"]', 'password');
  8  |     await page.click('button[type="submit"]');
  9  |     await page.waitForURL('/dashboard');
  10 |   });
  11 | 
  12 |   test('should display dashboard', async ({ page }) => {
  13 |     await expect(page.locator('h1')).toContainText('Dashboard');
  14 |   });
  15 | 
  16 |   test('should navigate to cases page', async ({ page }) => {
  17 |     await page.click('a[href="/dashboard/cases"]');
  18 |     await expect(page).toHaveURL('/dashboard/cases');
  19 |   });
  20 | });
  21 | 
```