import { expect, test } from "@playwright/test";

test("deep module routes are honest and Back/Forward uses heading focus without resetting Hysteresis", async ({ page }) => {
  await page.goto("/modules/not-ported");
  await expect(page.getByRole("heading", { level: 1 })).toHaveText("Not-ported");
  await expect(page.getByText(/not available|unavailable/i).first()).toBeVisible();
  await page.getByRole("link", { name: "Hysteresis lab", exact: true }).first().click();
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await page.getByLabel("Steps per batch", { exact: true }).fill("128");
  await page.getByRole("button", { name: "Run", exact: true }).click();
  await expect.poll(async()=>Number(await page.getByTestId("retained-count").textContent())).toBeGreaterThanOrEqual(128);
  await page.getByRole("link", { name: "Modules", exact: true }).click();
  await expect(page.locator("#modules")).toBeFocused();
  await page.goBack();
  await expect(page.getByRole("button", { name: "Pause", exact: true })).toBeEnabled();
  await expect.poll(async()=>Number(await page.getByTestId("retained-count").textContent())).toBeGreaterThanOrEqual(128);
  await page.getByRole("button", { name: "Pause", exact: true }).click();
  await page.getByRole("button", { name: "Read results", exact: true }).click();
  await page.getByRole("link", { name: "Documentation", exact: true }).first().click();
  await expect(page.getByRole("heading", { name: "Documentation", exact: true })).toBeFocused();
  await page.goBack();
  await expect(page.getByRole("heading", { name: "Hysteresis lab", exact: true })).toBeFocused();
  await expect(page.locator("#core-table tbody tr")).toHaveCount(32);
  await page.goForward();
  await expect(page.getByRole("heading", { name: "Documentation", exact: true })).toBeFocused();
});

test("module assets refuse an API-contract mismatch before instantiating science", async ({page}) => {
  test.skip(process.env.FECIM_WEB_PROFILE === "local", "hosted static Worker contract");
  await page.route("**/engine/manifest.json",async route=>{const response=await route.fetch();const manifest=await response.json();manifest.API="wrong-api";await route.fulfill({json:manifest});});
  await page.goto("/hysteresis");
  await expect(page.getByRole("alert")).toContainText("API-contract fingerprint mismatch");
  await page.getByRole("link",{name:"Projects",exact:true}).first().click();
  await page.getByRole("button",{name:"Open small example",exact:true}).click();
  await expect(page.getByRole("alert")).toContainText("API-contract fingerprint mismatch");
});
