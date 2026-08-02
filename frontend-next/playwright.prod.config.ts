import { defineConfig, devices } from '@playwright/test'

// Production smoke/E2E suite. Unlike the local e2e config (which boots a dev
// server and mocks APIs), this targets the deployed site and uses real auth.
//
//   ADMIN_PASSWORD=<prod admin pw> \
//   pnpm exec playwright test --config=playwright.prod.config.ts
//
// Optional env: E2E_PROD_URL (default https://www.godnslog.com),
//               E2E_PROD_SHORT_ID (admin subdomain id),
//               CHROMIUM_PATH (default /usr/bin/chromium-browser).
export default defineConfig({
  testDir: './e2e-prod',
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: 0,
  workers: 1,
  reporter: 'line',
  timeout: 60000,
  use: {
    baseURL: process.env.E2E_PROD_URL || 'https://www.godnslog.com',
    ignoreHTTPSErrors: true,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    actionTimeout: 15000,
    navigationTimeout: 20000,
  },
  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        launchOptions: {
          executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium-browser',
          args: ['--no-sandbox'],
        },
      },
    },
  ],
})
