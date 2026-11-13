import { expect, test } from "@playwright/test";
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, rmSync } from "node:fs";
import { join } from "node:path";

const local = process.env.FECIM_WEB_PROFILE === "local";
let child: ReturnType<typeof spawn> | undefined, root = "", launch = "";
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180000);
  root = mkdtempSync(join(process.env.TMPDIR!, "fragment-focus-"));
  const binary = join(root, "fecim");
  execFileSync("node", ["src/local/build-native.mjs", binary], { timeout: 120000 });
  child = spawn(binary, ["serve", "--root", root, "--port", String(Number(process.env.AUTOWORK_PORT_BASE || 20200) + 15)]);
  let output = "";
  child.stdout!.on("data", d => output += d);
  child.stderr!.on("data", d => output += d);
  await expect.poll(() => output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)?.[0], { timeout: 30000 }).toBeTruthy();
  launch = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)![0];
});
test.beforeEach(async ({ page }) => {
  if (!local) return;
  // Renew per browser context: launch links are single-use.
  let output = "";
  const collect = (d: Buffer) => { output += d; };
  child!.stdout!.on("data", collect);
  child!.stdin!.write("renew\n");
  await expect.poll(() => output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)?.[0]).toBeTruthy();
  launch = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)![0];
  child!.stdout!.off("data", collect);
  await page.goto(launch);
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  // The native bootstrap strips all cold fragments. After real authentication,
  // isolate the shared React focus owner so it receives the URL under test.
  await page.route("**/bootstrap.js", route => route.fulfill({ contentType: "text/javascript", body: "" }));
});
test.afterAll(async () => {
  if (child && child.exitCode === null) {
    const exited = new Promise(resolve => child!.once("exit", resolve));
    child.kill("SIGTERM");
    await exited;
  }
  if (root) rmSync(root, { recursive: true, force: true });
});

for (const hash of ["%", "%E0%A4%A"]) test(`cold malformed #${hash} uses the visible heading`, async ({ page }) => {
  const errors: Error[] = [];
  page.on("pageerror", error => errors.push(error));
  await page.goto(`${local ? new URL(launch).origin : ""}/#${hash}`);
  await expect(page.getByRole("heading", { level: 1 })).toBeFocused();
  await expect(page).toHaveURL(new RegExp(`/#${hash}$`));
  expect(errors, "no page errors, including URIError").toEqual([]);
});

test("SPA and history fragments preserve fallback, valid targets and later control focus", async ({ page }) => {
  const errors: Error[] = [];
  page.on("pageerror", error => errors.push(error));
  const origin = local ? new URL(launch).origin : "";
  const heading = page.getByRole("heading", { level: 1 });
  for (const hash of ["modules", "mod%75les", "missing-target"]) {
    await page.goto(`${origin}/#${hash}`);
    expect(errors, "no page errors, including URIError").toEqual([]);
    await expect(heading).toBeVisible();
    await expect(hash === "modules" || hash === "mod%75les" ? page.locator("#modules") : heading).toBeFocused();
    await expect(page).toHaveURL(`${origin || new URL(page.url()).origin}/#${hash}`);
  }

  // Alter only the destination of a real shared link; click must use the SPA router.
  await page.getByRole("link", { name: "Hysteresis lab", exact: true }).first().evaluate(a => a.setAttribute("href", "/hysteresis#%E0%A4%A"));
  await page.evaluate(() => { (window as Window & { fragmentNavigation?: boolean }).fragmentNavigation = true; });
  await page.getByRole("link", { name: "Hysteresis lab", exact: true }).first().click();
  await expect(heading).toHaveText("Hysteresis lab");
  await expect(heading).toBeFocused();
  expect(await page.evaluate(() => (window as Window & { fragmentNavigation?: boolean }).fragmentNavigation)).toBe(true);
  await expect(page).toHaveURL(/\/hysteresis#%E0%A4%A$/);
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  const steps = page.getByLabel("Steps per batch", { exact: true });
  await steps.fill("128");
  await expect(steps).toBeFocused();
  await page.getByRole("link", { name: "Modules", exact: true }).click();
  await expect(page.locator("#modules")).toBeFocused();
  await page.goBack();
  await expect(page).toHaveURL(/\/hysteresis#%E0%A4%A$/);
  await expect(heading).toBeFocused();
  await expect(steps).toHaveValue("128");
  await page.goForward();
  await expect(page.locator("#modules")).toBeFocused();
  await page.goBack();
  await expect(heading).toBeFocused();
  await page.goBack();
  await expect(page).toHaveURL(/\/#missing-target$/);
  await expect(heading).toBeFocused();
  await page.goForward();
  await expect(page).toHaveURL(/\/hysteresis#%E0%A4%A$/);
  await expect(heading).toBeFocused();
  expect(errors, "no page errors, including URIError").toEqual([]);
});
