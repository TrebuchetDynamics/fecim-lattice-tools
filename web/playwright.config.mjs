import { existsSync } from 'node:fs';
import { defineConfig, devices } from '@playwright/test';

const chromiumPath = process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE || '/usr/bin/chromium';
const launchOptions = {
  args: ['--no-sandbox', '--disable-dev-shm-usage'],
};

if (existsSync(chromiumPath)) {
  launchOptions.executablePath = chromiumPath;
}

const iPhone13ChromiumProfile = {
  // iPhone 13 dimensions without the WebKit-only defaultBrowserType from the
  // built-in descriptor, so system Chromium can run the mobile project.
  viewport: { width: 390, height: 844 },
  screen: { width: 390, height: 844 },
  deviceScaleFactor: 3,
  isMobile: true,
  hasTouch: true,
};

const smallPhoneChromiumProfile = {
  // Narrow Android-class small-phone viewport to catch cramped mobile layouts.
  viewport: { width: 360, height: 740 },
  screen: { width: 360, height: 740 },
  deviceScaleFactor: 2,
  isMobile: true,
  hasTouch: true,
};

export default defineConfig({
  testDir: './test/e2e',
  timeout: 90_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [['list']],
  webServer: {
    command: 'npm run preview -- --host 127.0.0.1 --port 4321',
    url: 'http://127.0.0.1:4321',
    reuseExistingServer: !process.env.CI,
    timeout: 45_000,
  },
  use: {
    baseURL: 'http://127.0.0.1:4321',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
    launchOptions,
  },
  projects: [
    {
      name: 'desktop-chromium',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1440, height: 900 },
      },
    },
    {
      name: 'desktop-compact-chromium',
      use: {
        ...devices['Desktop Chrome'],
        viewport: { width: 1024, height: 768 },
      },
    },
    {
      name: 'mobile-chromium',
      use: iPhone13ChromiumProfile,
    },
    {
      name: 'mobile-small-chromium',
      use: smallPhoneChromiumProfile,
    },
  ],
});
