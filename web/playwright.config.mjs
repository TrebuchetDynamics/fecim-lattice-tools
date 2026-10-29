import "./test/setup.mjs";
import { existsSync } from "node:fs";
import { defineConfig, devices } from "@playwright/test";

const port = Number(process.env.AUTOWORK_PORT_BASE || 20200);
const firefoxPath =
  process.env.PLAYWRIGHT_FIREFOX_EXECUTABLE ||
  (process.env.AUTOWORK_CACHE
    ? `${process.env.AUTOWORK_CACHE}/playwright-1.60.0/firefox-1522/firefox/firefox`
    : "");
const chromiumPath =
  process.env.PLAYWRIGHT_CHROMIUM_EXECUTABLE || "/usr/bin/chromium";
const launchOptions = {
  args: ["--no-sandbox", "--disable-dev-shm-usage"],
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
  testDir: "./test/e2e",
  // Discover future feature specs without another per-module CI registry edit.
  // The explicit embedded-file fixture remains its own lane; hosted isolation
  // is meaningful only in the hosted shared lane.
  testIgnore: process.env.FECIM_WEB_SUITE === "shared"
    ? ["**/hysteresis-local.spec.ts", ...(process.env.FECIM_WEB_PROFILE === "local" ? ["**/hosted-isolation.spec.ts"] : [])]
    : [],
  outputDir: process.env.TMPDIR ? `${process.env.TMPDIR}/playwright-${process.env.FECIM_WEB_PROFILE || "hosted"}` : "test-results",
  timeout: 90_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  reporter: [["list"]],
  webServer: [false, true].map((shaped) => ({
    command: `node test/e2e/serve.mjs ${port + (shaped ? 1 : 0)} ${shaped ? "--shaped" : ""}`,
    url: `http://127.0.0.1:${port + (shaped ? 1 : 0)}`,
    reuseExistingServer: false,
    timeout: 90_000,
  })),
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure",
    launchOptions,
  },
  projects: [
    {
      name: "desktop-chromium",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1440, height: 900 },
      },
    },
    {
      name: "desktop-compact-chromium",
      use: {
        ...devices["Desktop Chrome"],
        viewport: { width: 1024, height: 768 },
      },
    },
    {
      name: "mobile-chromium",
      use: iPhone13ChromiumProfile,
    },
    {
      name: "mobile-small-chromium",
      use: smallPhoneChromiumProfile,
    },
    {
      name: "desktop-firefox",
      use: {
        ...devices["Desktop Firefox"],
        launchOptions: existsSync(firefoxPath)
          ? { executablePath: firefoxPath }
          : {},
      },
    },
  ],
});
