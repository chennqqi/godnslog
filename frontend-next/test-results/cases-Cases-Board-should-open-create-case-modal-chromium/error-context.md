# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: cases.spec.ts >> Cases Board >> should open create case modal
- Location: e2e/cases.spec.ts:25:7

# Error details

```
TimeoutError: locator.click: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('button').filter({ hasText: 'New Case' }).first()

```

# Page snapshot

```yaml
- generic [active] [ref=e1]:
  - generic [ref=e3]:
    - generic [ref=e4]:
      - heading "GODNSLOG 2.0" [level=2] [ref=e5]
      - generic [ref=e6]:
        - button "EN" [ref=e7] [cursor=pointer]
        - button "中" [ref=e8] [cursor=pointer]
    - paragraph [ref=e9]: Sign in to your account
    - generic [ref=e10]:
      - generic [ref=e11]:
        - generic [ref=e12]:
          - generic [ref=e13]: Username
          - textbox "Username" [ref=e14]
        - generic [ref=e15]:
          - generic [ref=e16]: Password
          - textbox "Password" [ref=e17]
      - button "Sign In" [ref=e19] [cursor=pointer]
  - button "Open Next.js Dev Tools" [ref=e25] [cursor=pointer]:
    - img [ref=e26]
  - alert [ref=e29]
```

# Test source

```ts
  1   | import { test, expect } from '@playwright/test';
  2   | 
  3   | test.describe('Cases Board', () => {
  4   |   test.beforeEach(async ({ page }) => {
  5   |     // Set token before navigation to avoid redirect to login
  6   |     await page.goto('/');
  7   |     await page.evaluate(() => {
  8   |       localStorage.setItem('token', 'mock-token');
  9   |       localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
  10  |     });
  11  |     await page.goto('/dashboard/cases');
  12  |     await page.waitForLoadState('domcontentloaded');
  13  |     await page.waitForTimeout(1000);
  14  |   });
  15  | 
  16  |   test('should display cases page', async ({ page }) => {
  17  |     await expect(page.locator('h2').first()).toContainText('Case Board');
  18  |   });
  19  | 
  20  |   test('should show create case button', async ({ page }) => {
  21  |     const createButton = page.locator('button').filter({ hasText: 'New Case' }).first();
  22  |     await expect(createButton).toBeVisible();
  23  |   });
  24  | 
  25  |   test('should open create case modal', async ({ page }) => {
  26  |     const createButton = page.locator('button').filter({ hasText: 'New Case' }).first();
> 27  |     await createButton.click();
      |                        ^ TimeoutError: locator.click: Timeout 10000ms exceeded.
  28  |     await expect(page.getByRole('heading', { name: 'New Case' })).toBeVisible();
  29  |   });
  30  | 
  31  |   test('should display search input', async ({ page }) => {
  32  |     const searchInput = page.locator('input[placeholder*="Search"]');
  33  |     await expect(searchInput).toBeVisible();
  34  |   });
  35  | 
  36  |   test('should display status filter', async ({ page }) => {
  37  |     // Radix Select uses a button trigger, not a native select element
  38  |     const statusFilterTrigger = page.locator('button').filter({ hasText: 'All statuses' }).first();
  39  |     await expect(statusFilterTrigger).toBeVisible();
  40  |   });
  41  | 
  42  |   test('should not display batch operations', async ({ page }) => {
  43  |     // Batch operations should not exist in the simplified Cases Board
  44  |     const batchDeleteButton = page.locator('button').filter({ hasText: 'Delete selected' });
  45  |     await expect(batchDeleteButton).not.toBeVisible();
  46  |   });
  47  | 
  48  |   test('should not display edit/delete buttons in list', async ({ page }) => {
  49  |     // Edit and delete buttons should not exist in the simplified Cases Board
  50  |     const editButton = page.locator('button').filter({ hasText: 'Edit' });
  51  |     const deleteButton = page.locator('button').filter({ hasText: 'Delete' });
  52  |     await expect(editButton).not.toBeVisible();
  53  |     await expect(deleteButton).not.toBeVisible();
  54  |   });
  55  | 
  56  |   test('should navigate to case detail on click', async ({ page }) => {
  57  |     // Mock a case in the list
  58  |     await page.route('**/api/v2/cases**', route => route.fulfill({
  59  |       json: {
  60  |         code: 0,
  61  |         data: {
  62  |           items: [
  63  |             { id: 'case-1', title: 'Test Case', description: 'Test description', status: 'active', created_at: new Date().toISOString() }
  64  |           ],
  65  |           total: 1,
  66  |           page: 1,
  67  |           page_size: 20,
  68  |           total_pages: 1
  69  |         }
  70  |       }
  71  |     }));
  72  | 
  73  |     await page.reload();
  74  |     await page.waitForLoadState('domcontentloaded');
  75  |     await page.waitForTimeout(1000);
  76  | 
  77  |     // Click on the case row
  78  |     const caseRow = page.locator('li').filter({ hasText: 'Test Case' }).first();
  79  |     await caseRow.click();
  80  | 
  81  |     // Should navigate to case detail
  82  |     await page.waitForURL('**/dashboard/cases/case-1');
  83  |     expect(page.url()).toContain('/dashboard/cases/case-1');
  84  |   });
  85  | });
  86  | 
  87  | test.describe('Case Detail', () => {
  88  |   test.beforeEach(async ({ page }) => {
  89  |     await page.goto('/');
  90  |     await page.evaluate(() => {
  91  |       localStorage.setItem('token', 'mock-token');
  92  |       localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
  93  |     });
  94  | 
  95  |     // Mock case detail API
  96  |     await page.route('**/api/v2/cases/case-1', route => route.fulfill({
  97  |       json: {
  98  |         code: 0,
  99  |         data: {
  100 |           id: 'case-1',
  101 |           title: 'Test Case',
  102 |           description: 'Test description',
  103 |           target: 'example.com',
  104 |           status: 'active',
  105 |           created_at: new Date().toISOString()
  106 |         }
  107 |       }
  108 |     }));
  109 | 
  110 |     // Mock case stats API
  111 |     await page.route('**/api/v2/cases/case-1/stats', route => route.fulfill({
  112 |       json: {
  113 |         code: 0,
  114 |         data: {
  115 |           payload_count: 5,
  116 |           interaction_count: 12,
  117 |           hit_payload_count: 3
  118 |         }
  119 |       }
  120 |     }));
  121 | 
  122 |     // Mock payloads API
  123 |     await page.route('**/api/v2/payloads?case_id=case-1', route => route.fulfill({
  124 |       json: {
  125 |         code: 0,
  126 |         data: {
  127 |           items: [
```