# Instructions

- Following Playwright test failed.
- Explain why, be concise, respect Playwright best practices.
- Provide a snippet of code with the fix, if possible.

# Test info

- Name: cases.spec.ts >> Payload Detail >> should navigate to interactions on quick action click
- Location: e2e/cases.spec.ts:329:7

# Error details

```
TimeoutError: locator.click: Timeout 10000ms exceeded.
Call log:
  - waiting for locator('button').filter({ hasText: '查看交互' }).first()

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
  265 |           title: 'Test Case',
  266 |           description: 'Test description',
  267 |           target: 'example.com',
  268 |           status: 'active',
  269 |           created_at: new Date().toISOString()
  270 |         }
  271 |       }
  272 |     }));
  273 | 
  274 |     // Mock interactions API
  275 |     await page.route('**/api/v2/interactions?payload_id=payload-1', route => route.fulfill({
  276 |       json: {
  277 |         code: 0,
  278 |         data: {
  279 |           items: [
  280 |             { id: 'int-1', type: 'dns', source_ip: '1.2.3.4', timestamp: new Date().toISOString() }
  281 |           ],
  282 |           total: 1,
  283 |           page: 1,
  284 |           page_size: 5,
  285 |           total_pages: 1
  286 |         }
  287 |       }
  288 |     }));
  289 | 
  290 |     await page.goto('/dashboard/payloads/payload-1');
  291 |     await page.waitForLoadState('domcontentloaded');
  292 |     await page.waitForTimeout(1000);
  293 |   });
  294 | 
  295 |   test('should display payload detail', async ({ page }) => {
  296 |     await expect(page.locator('h2').first()).toContainText('ssrf_http');
  297 |   });
  298 | 
  299 |   test('should display token', async ({ page }) => {
  300 |     await expect(page.locator('text=gdl_abc123').first()).toBeVisible();
  301 |   });
  302 | 
  303 |   test('should display rendered payload', async ({ page }) => {
  304 |     await expect(page.locator('text=http://gdl_abc123.example.com/test').first()).toBeVisible();
  305 |   });
  306 | 
  307 |   test('should display associated case', async ({ page }) => {
  308 |     await expect(page.locator('text=关联 Case').first()).toBeVisible();
  309 |     await expect(page.locator('text=Test Case').first()).toBeVisible();
  310 |   });
  311 | 
  312 |   test('should display recent interactions', async ({ page }) => {
  313 |     await expect(page.locator('text=最近交互').first()).toBeVisible();
  314 |     await expect(page.locator('text=DNS').first()).toBeVisible();
  315 |   });
  316 | 
  317 |   test('should display quick actions', async ({ page }) => {
  318 |     await expect(page.locator('text=快速操作').first()).toBeVisible();
  319 |     await expect(page.locator('button').filter({ hasText: '查看交互' }).first()).toBeVisible();
  320 |     await expect(page.locator('button').filter({ hasText: '查看证据' }).first()).toBeVisible();
  321 |   });
  322 | 
  323 |   test('should not display revoke button', async ({ page }) => {
  324 |     // Revoke button should not exist in the simplified Payload Detail
  325 |     const revokeButton = page.locator('button').filter({ hasText: '撤销' });
  326 |     await expect(revokeButton).not.toBeVisible();
  327 |   });
  328 | 
  329 |   test('should navigate to interactions on quick action click', async ({ page }) => {
  330 |     const interactionsButton = page.locator('button').filter({ hasText: '查看交互' }).first();
> 331 |     await interactionsButton.click();
      |                              ^ TimeoutError: locator.click: Timeout 10000ms exceeded.
  332 |     await page.waitForURL('**/dashboard/interactions?payload_id=payload-1');
  333 |     expect(page.url()).toContain('payload_id=payload-1');
  334 |   });
  335 | });
  336 | 
```