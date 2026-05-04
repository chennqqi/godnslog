# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: login.spec.ts >> Login Page >> should display login form
- Location: e2e\login.spec.ts:4:7

# Error details

```
Error: expect(locator).toContainText(expected) failed

Locator: locator('h1')
Expected substring: "GODNSLOG"
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toContainText" with timeout 5000ms
  - waiting for locator('h1')

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
  3  | test.describe('Login Page', () => {
  4  |   test('should display login form', async ({ page }) => {
  5  |     await page.goto('/login');
  6  |     
> 7  |     await expect(page.locator('h1')).toContainText('GODNSLOG');
     |                                      ^ Error: expect(locator).toContainText(expected) failed
  8  |   });
  9  | 
  10 |   test('should show error with invalid credentials', async ({ page }) => {
  11 |     await page.goto('/login');
  12 |     
  13 |     await page.fill('input[type="email"]', 'invalid@example.com');
  14 |     await page.fill('input[type="password"]', 'wrongpassword');
  15 |     await page.click('button[type="submit"]');
  16 |     
  17 |     await expect(page.locator('.error')).toBeVisible();
  18 |   });
  19 | });
  20 | 
```