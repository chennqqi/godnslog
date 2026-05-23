# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: cases.spec.ts >> Case Detail >> should display quick actions
- Location: e2e/cases.spec.ts:163:7

# Error details

```
Error: expect(locator).toBeVisible() failed

Locator: locator('text=Quick Actions').first()
Expected: visible
Timeout: 5000ms
Error: element(s) not found

Call log:
  - Expect "toBeVisible" with timeout 5000ms
  - waiting for locator('text=Quick Actions').first()

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
  128 |             { id: 'payload-1', token: 'gdl_abc123', template: 'ssrf_http', status: 'deployed', created_at: new Date().toISOString() }
  129 |           ],
  130 |           total: 1,
  131 |           page: 1,
  132 |           page_size: 20,
  133 |           total_pages: 1
  134 |         }
  135 |       }
  136 |     }));
  137 | 
  138 |     await page.goto('/dashboard/cases/case-1');
  139 |     await page.waitForLoadState('domcontentloaded');
  140 |     await page.waitForTimeout(1000);
  141 |   });
  142 | 
  143 |   test('should display case detail', async ({ page }) => {
  144 |     await expect(page.locator('h2').first()).toContainText('Test Case');
  145 |   });
  146 | 
  147 |   test('should display case stats', async ({ page }) => {
  148 |     await expect(page.locator('text=Payloads').first()).toBeVisible();
  149 |     await expect(page.locator('text=Interactions').first()).toBeVisible();
  150 |     await expect(page.locator('text=Hit Payloads').first()).toBeVisible();
  151 |   });
  152 | 
  153 |   test('should display payloads list', async ({ page }) => {
  154 |     await expect(page.locator('text=Payloads').first()).toBeVisible();
  155 |     await expect(page.locator('text=gdl_abc123').first()).toBeVisible();
  156 |   });
  157 | 
  158 |   test('should display create payload button', async ({ page }) => {
  159 |     const createButton = page.locator('button').filter({ hasText: 'Create Payload' }).first();
  160 |     await expect(createButton).toBeVisible();
  161 |   });
  162 | 
  163 |   test('should display quick actions', async ({ page }) => {
> 164 |     await expect(page.locator('text=Quick Actions').first()).toBeVisible();
      |                                                              ^ Error: expect(locator).toBeVisible() failed
  165 |     await expect(page.locator('button').filter({ hasText: 'View Evidence' }).first()).toBeVisible();
  166 |     await expect(page.locator('button').filter({ hasText: 'View Interactions' }).first()).toBeVisible();
  167 |   });
  168 | 
  169 |   test('should navigate to new payload with case_id', async ({ page }) => {
  170 |     const createButton = page.locator('button').filter({ hasText: 'Create Payload' }).first();
  171 |     await createButton.click();
  172 | 
  173 |     // Should navigate to new payload page with case_id
  174 |     await page.waitForURL('**/dashboard/payloads/new?case_id=case-1');
  175 |     expect(page.url()).toContain('case_id=case-1');
  176 |   });
  177 | });
  178 | 
  179 | test.describe('New Payload', () => {
  180 |   test.beforeEach(async ({ page }) => {
  181 |     await page.goto('/');
  182 |     await page.evaluate(() => {
  183 |       localStorage.setItem('token', 'mock-token');
  184 |       localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
  185 |     });
  186 | 
  187 |     // Mock cases API
  188 |     await page.route('**/api/v2/cases', route => route.fulfill({
  189 |       json: {
  190 |         code: 0,
  191 |         data: {
  192 |           items: [
  193 |             { id: 'case-1', title: 'Test Case', description: 'Test description', status: 'active', created_at: new Date().toISOString() }
  194 |           ],
  195 |           total: 1,
  196 |           page: 1,
  197 |           page_size: 20,
  198 |           total_pages: 1
  199 |         }
  200 |       }
  201 |     }));
  202 |   });
  203 | 
  204 |   test('should display new payload page', async ({ page }) => {
  205 |     await page.goto('/dashboard/payloads/new');
  206 |     await page.waitForLoadState('domcontentloaded');
  207 |     await page.waitForTimeout(1000);
  208 |     await expect(page.locator('h2').first()).toContainText('New Payload');
  209 |   });
  210 | 
  211 |   test('should display associated case when case_id is provided', async ({ page }) => {
  212 |     await page.goto('/dashboard/payloads/new?case_id=case-1');
  213 |     await page.waitForLoadState('domcontentloaded');
  214 |     await page.waitForTimeout(1000);
  215 |     await expect(page.locator('text=Creating for Case').first()).toBeVisible();
  216 |     await expect(page.locator('text=Test Case').first()).toBeVisible();
  217 |   });
  218 | 
  219 |   test('should display step indicator', async ({ page }) => {
  220 |     await page.goto('/dashboard/payloads/new');
  221 |     await page.waitForLoadState('domcontentloaded');
  222 |     await page.waitForTimeout(1000);
  223 |     await expect(page.locator('text=Choose a template').first()).toBeVisible();
  224 |   });
  225 | 
  226 |   test('should display template selection', async ({ page }) => {
  227 |     await page.goto('/dashboard/payloads/new');
  228 |     await page.waitForLoadState('domcontentloaded');
  229 |     await page.waitForTimeout(1000);
  230 |     await expect(page.locator('text=SSRF HTTP').first()).toBeVisible();
  231 |   });
  232 | });
  233 | 
  234 | test.describe('Payload Detail', () => {
  235 |   test.beforeEach(async ({ page }) => {
  236 |     await page.goto('/');
  237 |     await page.evaluate(() => {
  238 |       localStorage.setItem('token', 'mock-token');
  239 |       localStorage.setItem('user', JSON.stringify({ id: 1, username: 'admin', email: 'admin@godnslog.com', role: 0, lang: 'en-US' }));
  240 |     });
  241 | 
  242 |     // Mock payload API
  243 |     await page.route('**/api/v2/payloads/payload-1', route => route.fulfill({
  244 |       json: {
  245 |         code: 0,
  246 |         data: {
  247 |           id: 'payload-1',
  248 |           token: 'gdl_abc123',
  249 |           template: 'ssrf_http',
  250 |           rendered_payload: 'http://gdl_abc123.example.com/test',
  251 |           status: 'hit',
  252 |           case_id: 'case-1',
  253 |           created_at: new Date().toISOString(),
  254 |           expires_at: new Date(Date.now() + 86400000).toISOString()
  255 |         }
  256 |       }
  257 |     }));
  258 | 
  259 |     // Mock associated case API
  260 |     await page.route('**/api/v2/cases/case-1', route => route.fulfill({
  261 |       json: {
  262 |         code: 0,
  263 |         data: {
  264 |           id: 'case-1',
```