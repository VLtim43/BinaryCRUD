import { defineConfig, devices } from '@playwright/test';

export default defineConfig({
  testDir: './tests',
  // Output folders inside tests/
  outputDir: './tests/test-results',
  reporter: [['html', { outputFolder: './tests/playwright-report' }]],
  // Run tests in sequence to avoid state conflicts
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  // Global timeout for each test
  timeout: 30000,
  use: {
    // Wails dev server URL
    baseURL: 'http://localhost:34115',
    trace: 'on-first-retry',
    // Screenshot on failure
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
});
