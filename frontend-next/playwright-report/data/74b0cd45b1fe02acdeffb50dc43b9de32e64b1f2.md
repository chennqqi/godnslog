# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: cases.spec.ts >> Cases Board >> should show create case button
- Location: e2e/cases.spec.ts:20:7

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('button').filter({ hasText: 'New Case' }).first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('button').filter({ hasText: 'New Case' }).first()

```

```yaml
- heading "GODNSLOG 2.0" [level=2]
- button "EN"
- button "中"
- paragraph: Sign in to your account
- text: Username
- textbox "Username"
- text: Password
- textbox "Password"
- button "Sign In"
- alert
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
> 22  |     await expect(createButton).toBeVisible();
      |                                ^ Error: expect(locator).toBeVisible() failed
  23  |   });
  24  | 
  25  |   test('should open create case modal', async ({ page }) => {
  26  |     const createButton = page.locator('button').filter({ hasText: 'New Case' }).first();
  27  |     await createButton.click();
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
```