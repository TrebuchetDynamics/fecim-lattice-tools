import { expect, test } from "@playwright/test";
import { execFileSync, spawn } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { join } from "node:path";
import AxeBuilder from "@axe-core/playwright";

const local = process.env.FECIM_WEB_PROFILE === "local";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 16;
let root = "",
  launch = "",
  child: ReturnType<typeof spawn> | undefined;
let cookies: Awaited<
  ReturnType<import("@playwright/test").BrowserContext["cookies"]>
> = [];
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "circuits-local-"));
  const binary = join(root, "fecim");
  execFileSync("node", ["src/local/build-native.mjs", binary], {
    timeout: 120_000,
  });
  child = spawn(binary, ["serve", "--root", root, "--port", String(port)]);
  let output = "";
  child.stdout!.on("data", (d) => (output += d));
  child.stderr!.on("data", (d) => (output += d));
  await expect
    .poll(
      () =>
        output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)?.[0],
      { timeout: 30_000 },
    )
    .toBeTruthy();
  launch = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)![0];
});
test.afterAll(async () => {
  if (child && child.exitCode === null) {
    const done = new Promise((r) => child!.once("exit", r));
    child.kill("SIGTERM");
    await done;
  }
  if (root) rmSync(root, { recursive: true, force: true });
});

test("persistent circuit WRITE READ COMPUTE, masks, keyboard tables, undo and rooted capability", async ({
  page,
}) => {
  const requests: { url: string; method: string }[] = [];
  page.on("request", (r) =>
    requests.push({ url: r.url(), method: r.method() }),
  );
  if (local) {
    await page.goto(launch);
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    cookies = await page.context().cookies();
    await page
      .getByRole("link", { name: "Projects", exact: true })
      .first()
      .click();
    await page.getByLabel("Project name", { exact: true }).fill("circuits");
    await page
      .getByRole("button", { name: "Create project", exact: true })
      .click();
    await expect(page.getByTestId("research-status")).toContainText(
      "Project opened",
    );
    await page.goto(`http://127.0.0.1:${port}/modules/circuits`);
  } else await page.goto("/modules/circuits");
  await expect(
    page.getByRole("heading", { name: "Circuits", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("button", { name: "WRITE selected cell", exact: true }),
  ).toBeEnabled();
  await page.getByLabel("Rows", { exact: true }).selectOption("2");
  await page.getByLabel("Columns", { exact: true }).selectOption("2");
  await page.getByLabel("Architecture", { exact: true }).selectOption("2T1R");
  await page
    .getByLabel("Signed input codes (JSON)", { exact: true })
    .fill("[255,-128]");
  await page
    .getByRole("button", { name: "Apply configuration", exact: true })
    .click();
  await expect(page.locator("#circuits-cells tbody tr")).toHaveCount(4);
  await page.getByLabel("Cell row", { exact: true }).fill("0");
  await page.getByLabel("Cell column", { exact: true }).fill("0");
  await page.getByRole("button", { name: "Select cell", exact: true }).focus();
  await page.keyboard.press("Enter");
  await page.getByLabel("Write target level", { exact: true }).fill("20");
  await page
    .getByRole("button", { name: "WRITE selected cell", exact: true })
    .focus();
  await page.keyboard.press("Enter");
  await expect(page.getByTestId("circuits-status")).toContainText(
    /verified|pulse limit/,
  );
  const programmed = await page
    .locator("#circuits-selection tbody")
    .textContent();
  await expect(
    page.locator("#circuits-selection tbody td").nth(1),
  ).toContainText("20 count");
  await page
    .getByRole("button", { name: "READ selected cell", exact: true })
    .click();
  await expect(page.getByTestId("circuits-status")).toContainText(
    "read completed",
  );
  await expect(page.locator("#circuits-sense tbody tr")).toHaveCount(2);
  await page
    .getByLabel("Sense preset", { exact: true })
    .selectOption("High Sensitivity");
  await page
    .getByLabel("Manual DAC voltages (JSON or null)", { exact: true })
    .fill("[0.1,-0.2]");
  for (const coupling of ["Ideal", "TierA", "TierB"]) {
    await page.getByLabel("Coupling", { exact: true }).selectOption(coupling);
    await page
      .getByRole("button", { name: "Apply configuration", exact: true })
      .click();
    await page
      .getByRole("button", { name: "COMPUTE vector", exact: true })
      .click();
    await expect(page.getByTestId("circuits-status")).toContainText(
      "compute completed",
    );
    await expect(page.locator("#circuits-cell_results tbody tr")).toHaveCount(
      4,
    );
    await expect(page.locator("#circuits-columns tbody tr")).toHaveCount(2);
    await expect(page.locator("#circuits-sense")).toContainText("amperes");
  }
  await page
    .getByLabel("2T1R read mask (JSON or null)", { exact: true })
    .fill("[[true,false],[false,true]]");
  await page
    .getByLabel("Active rows (JSON)", { exact: true })
    .fill("[true,false]");
  await page
    .getByRole("button", { name: "Apply configuration", exact: true })
    .click();
  await expect(page.getByText(/STALE:/)).toBeVisible();
  await page
    .getByRole("button", { name: "COMPUTE vector", exact: true })
    .click();
  await expect(page.getByTestId("circuits-status")).toContainText(
    "compute completed",
  );
  await page
    .getByRole("button", { name: "Randomize array", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Undo device change", exact: true })
    .click();
  await expect(page.locator("#circuits-selection tbody")).toHaveText(
    programmed!,
  );
  await page.getByRole("button", { name: "Step timing", exact: true }).click();
  await expect(page.locator("#circuits-timing")).toContainText("seconds");
  const retained = await page.locator("#circuits-cells tbody").textContent();
  await page
    .getByLabel("Signed input codes (JSON)", { exact: true })
    .fill("[1]");
  await page
    .getByRole("button", { name: "Apply configuration", exact: true })
    .click();
  await expect(page.getByRole("alert")).toBeFocused();
  expect(await page.locator("#circuits-cells tbody").textContent()).toBe(
    retained,
  );
  await page
    .getByRole("button", { name: "Dismiss error", exact: true })
    .click();
  await page
    .getByLabel("Signed input codes (JSON)", { exact: true })
    .fill("[255,-128]");
  await page.getByRole("button", { name: "Reset array", exact: true }).click();
  await page
    .getByRole("button", { name: "Files and capabilities", exact: true })
    .click();
  if (local) {
    await page
      .getByLabel("Project name for circuit exports", { exact: true })
      .fill("circuits");
    await page
      .getByRole("button", { name: "Open export project", exact: true })
      .click();
    await expect(page.getByTestId("circuits-files")).toContainText(
      "Rooted project opened",
    );
    await page
      .getByLabel("Relative JSON export path", { exact: true })
      .fill("circuit.json");
    await page
      .getByRole("button", { name: "Export circuit JSON", exact: true })
      .click();
    await expect(page.getByTestId("circuits-files")).toContainText("Saved");
    expect(
      JSON.parse(readFileSync(join(root, "circuits/circuit.json"), "utf8"))
        .Config.architecture,
    ).toBe("2T1R");
    await page
      .getByLabel("Relative JSON export path", { exact: true })
      .fill("../escape.json");
    await page
      .getByRole("button", { name: "Export circuit JSON", exact: true })
      .click();
    await expect(page.getByRole("alert")).toBeVisible();
  } else {
    await page
      .getByRole("button", { name: "Export circuit JSON", exact: true })
      .click();
    await expect(page.getByTestId("circuits-files")).toContainText("serve");
    expect(
      requests.filter(
        (r) =>
          new URL(r.url).pathname.startsWith("/api/") || r.method !== "GET",
      ),
    ).toEqual([]);
  }
  const a11y = await new AxeBuilder({ page })
    .include(".circuits-lab")
    .analyze();
  expect(a11y.violations).toEqual([]);
  await page.setViewportSize({ width: 390, height: 844 });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 1,
    ),
  ).toBe(true);
});

test("lost circuit job response retains ownership across navigation until authoritative recovery", async ({
  page,
}) => {
  if (local) {
    if (cookies.length) {
      await page.context().addCookies(cookies);
      await page.goto(`http://127.0.0.1:${port}/hysteresis`);
    } else await page.goto(launch);
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    await page.goto(`http://127.0.0.1:${port}/modules/circuits`);
  } else await page.goto("/modules/circuits");
  await expect(
    page.getByRole("button", { name: "COMPUTE vector", exact: true }),
  ).toBeEnabled();
  if (local)
    await page.route("**/api/v1/module-sessions/*/jobs/*", (route) =>
      route.abort("failed"),
    );
  else
    await page.evaluate(() => {
      const prototype = (globalThis as any).HysteresisWorkerClient.prototype;
      const original = prototype.request;
      let drop = false,
        injected = false;
      prototype.request = function (op: string, ...args: unknown[]) {
        if (op === "start" && !injected) {
          drop = true;
          injected = true;
        }
        const work = original.call(this, op, ...args);
        if (op === "state" && drop) {
          drop = false;
          return work.then(() => {
            throw new Error("Injected lost circuit job response");
          });
        }
        return work;
      };
    });
  await page
    .getByRole("button", { name: "COMPUTE vector", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(/unknown/i);
  await page
    .getByRole("link", { name: "Hysteresis lab", exact: true })
    .first()
    .click();
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(/busy|active|another/i);
  await page.goBack();
  if (local) await page.unroute("**/api/v1/module-sessions/*/jobs/*");
  await page
    .getByRole("button", { name: "Reconnect circuit state", exact: true })
    .click();
  await expect(page.getByTestId("circuits-status")).toContainText(
    "compute completed",
  );
  await page
    .getByRole("button", { name: "WRITE selected cell", exact: true })
    .click();
  // A short native job can finish before the start response is read. Cancel
  // only while the Go-owned control is enabled; both terminal outcomes are real.
  await page
    .getByRole("button", { name: "Cancel circuit computation", exact: true })
    .evaluate((button: HTMLButtonElement) => {
      if (!button.disabled) button.click();
    });
  await expect(page.getByTestId("circuits-status")).toContainText(
    /Cancelled|verified|pulse limit/,
  );
});
