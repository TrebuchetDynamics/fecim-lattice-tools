import { expect, test } from "@playwright/test";
import { spawn, execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, rmSync } from "node:fs";
import { join } from "node:path";

const local = process.env.FECIM_WEB_PROFILE === "local";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 14;
let root = "",
  launch = "",
  child: ReturnType<typeof spawn> | undefined;
let cookies: Awaited<
  ReturnType<import("@playwright/test").BrowserContext["cookies"]>
> = [];
async function openLocal(page: import("@playwright/test").Page) {
  if (cookies.length) {
    await page.context().addCookies(cookies);
    await page.goto(`http://127.0.0.1:${port}/hysteresis`);
  } else await page.goto(launch);
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  cookies = await page.context().cookies();
}
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "crossbar-local-"));
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
    const exited = new Promise((r) => child!.once("exit", r));
    child.kill("SIGTERM");
    await exited;
  }
  if (root) rmSync(root, { recursive: true, force: true });
});

test("real crossbar controls, keyboard cell/vector tables, profiles, atomic errors and explicit files", async ({
  page,
}) => {
  const requests: { url: string; method: string }[] = [];
  page.on("request", (r) =>
    requests.push({ url: r.url(), method: r.method() }),
  );
  if (local) {
    await openLocal(page);
    await page
      .getByRole("link", { name: "Projects", exact: true })
      .first()
      .click();
    await page.getByLabel("Project name", { exact: true }).fill("crossbar");
    await page
      .getByRole("button", { name: "Create project", exact: true })
      .click();
    await expect(page.getByTestId("research-status")).toContainText(
      "Project opened",
    );
    await page.goto(`http://127.0.0.1:${port}/modules/crossbar`);
  } else await page.goto("/modules/crossbar");
  await expect(
    page.getByRole("heading", { name: "Crossbar", exact: true }),
  ).toBeFocused();
  await expect(page.getByLabel("Model profile", { exact: true })).toBeEnabled();
  await page.getByLabel("Rows", { exact: true }).fill("2");
  await page.getByLabel("Columns", { exact: true }).fill("3");
  await page
    .getByLabel("Input vector (JSON volts)", { exact: true })
    .fill("[0.5,1,0.25]");
  await page
    .getByRole("button", { name: "Apply configuration", exact: true })
    .click();
  await expect(page.locator("#crossbar-matrix tbody tr")).toHaveCount(6);
  await page.getByLabel("Cell row", { exact: true }).fill("1");
  await page.getByLabel("Cell column", { exact: true }).fill("2");
  await page.getByRole("button", { name: "Select cell", exact: true }).focus();
  await page.keyboard.press("Enter");
  await expect(page.locator("#crossbar-selection tbody")).toContainText(
    "1 count",
  );
  await page.getByLabel("Cell state", { exact: true }).fill("0.8");
  await page
    .getByRole("button", { name: "Program selected cell", exact: true })
    .click();
  await expect(page.locator("#crossbar-selection tbody")).toContainText(
    "0.7931034482758621",
  );
  for (const profile of ["basic", "enhanced", "parasitic", "fecap"]) {
    await page
      .getByLabel("Model profile", { exact: true })
      .selectOption(profile);
    if (profile === "enhanced")
      await page
        .getByLabel("Sneak paths", { exact: true })
        .selectOption("true");
    await page
      .getByRole("button", { name: "Apply configuration", exact: true })
      .click();
    await page.getByRole("button", { name: "Run MVM", exact: true }).click();
    await expect(page.getByTestId("crossbar-status")).toContainText("finished");
    await page
      .getByLabel("Analysis mode", { exact: true })
      .selectOption("vectors");
    await expect(page.locator("#crossbar-output tbody tr")).toHaveCount(
      profile === "parasitic" ? 3 : 2,
    );
    if (profile === "fecap")
      await expect(page.locator("#crossbar-output")).toContainText("coulombs");
    if (profile === "enhanced") {
      await page
        .getByLabel("Analysis mode", { exact: true })
        .selectOption("ir");
      await expect(page.locator("#crossbar-ir tbody tr")).toHaveCount(6);
      await page
        .getByLabel("Analysis mode", { exact: true })
        .selectOption("sneak");
      await expect(
        page.locator("#crossbar-sneak-trace tbody tr"),
      ).not.toHaveCount(0);
    }
    if (profile === "parasitic") {
      await page
        .getByLabel("Analysis mode", { exact: true })
        .selectOption("solver");
      await expect(page.locator("#crossbar-solver-cells tbody tr")).toHaveCount(
        6,
      );
    }
  }
  for (const mode of [
    "ir",
    "sneak",
    "accuracy",
    "fecap",
    "stack",
    "literature",
    "solver",
    "matrix",
  ]) {
    await page.getByLabel("Analysis mode", { exact: true }).selectOption(mode);
    await expect(page.getByLabel("Analysis mode", { exact: true })).toHaveValue(
      mode,
    );
  }
  await page.getByLabel("Analysis mode", { exact: true }).selectOption("stack");
  await page.getByLabel("Layer count", { exact: true }).fill("512");
  await page.getByLabel("Selected layer", { exact: true }).fill("511");
  await page
    .getByRole("button", { name: "Apply stack selection", exact: true })
    .click();
  await expect(page.locator("#crossbar-stack-cells caption")).toContainText(
    "511 of 512",
  );
  await expect(page.locator("#crossbar-stack-cells tbody tr")).toHaveCount(6);
  await page
    .getByLabel("Analysis mode", { exact: true })
    .selectOption("matrix");
  await page
    .getByLabel("Matrix palette", { exact: true })
    .selectOption("amber");
  const before = await page.locator("#crossbar-matrix tbody").textContent();
  await page
    .getByLabel("Input vector (JSON volts)", { exact: true })
    .fill("[1]");
  await page
    .getByRole("button", { name: "Apply configuration", exact: true })
    .click();
  await expect(page.getByRole("alert")).toBeFocused();
  expect(await page.locator("#crossbar-matrix tbody").textContent()).toBe(
    before,
  );
  await page
    .getByRole("button", { name: "Dismiss error", exact: true })
    .click();
  await page
    .getByLabel("Input vector (JSON volts)", { exact: true })
    .fill("[0.5,1,0.25]");
  if (local) {
    await page
      .getByLabel("Project name for crossbar files", { exact: true })
      .fill("crossbar");
    await page
      .getByRole("button", { name: "Open file project", exact: true })
      .click();
    await page
      .getByLabel("Relative JSON path", { exact: true })
      .fill("matrix.json");
    await page
      .getByRole("button", { name: "Save crossbar JSON", exact: true })
      .click();
    await expect(page.getByTestId("crossbar-files")).toContainText("Saved");
    expect(
      JSON.parse(readFileSync(join(root, "crossbar/matrix.json"), "utf8"))
        .config.profile,
    ).toBe("fecap");
    await page
      .getByRole("button", { name: "Reset seeded array", exact: true })
      .click();
    await page
      .getByRole("button", { name: "Load crossbar JSON", exact: true })
      .click();
    await expect(page.getByTestId("crossbar-files")).toContainText(
      "unverified",
    );
    await page
      .getByLabel("Relative JSON path", { exact: true })
      .fill("../escape.json");
    await page
      .getByRole("button", { name: "Save crossbar JSON", exact: true })
      .click();
    await expect(page.getByRole("alert")).toBeVisible();
  } else {
    await page
      .getByRole("button", { name: "Save crossbar JSON", exact: true })
      .click();
    await expect(page.getByTestId("crossbar-files")).toContainText("serve");
    expect(
      requests.filter(
        (r) =>
          new URL(r.url).pathname.startsWith("/api/") || r.method !== "GET",
      ),
    ).toEqual([]);
  }
});

test("lost local start response retains scientific ownership until authoritative recovery", async ({
  page,
}) => {
  test.skip(!local, "native response-loss ownership boundary");
  await openLocal(page);
  await page.goto(`http://127.0.0.1:${port}/modules/crossbar`);
  await expect(
    page.getByRole("button", { name: "Run MVM", exact: true }),
  ).toBeEnabled();
  await page.route("**/api/v1/module-sessions/*/jobs/*", (route) =>
    route.abort("failed"),
  );
  await page.getByRole("button", { name: "Run MVM", exact: true }).click();
  await expect(page.getByRole("alert")).toBeVisible();
  await page
    .getByRole("link", { name: "Hysteresis lab", exact: true })
    .first()
    .click();
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(/busy|active|another/i);
  await page.goBack();
  await page.unroute("**/api/v1/module-sessions/*/jobs/*");
  await page
    .getByRole("button", { name: "Reconnect crossbar state", exact: true })
    .click();
  await expect(page.getByTestId("crossbar-status")).toContainText("finished");
});

test("cold default crossbar result meets measured route budget", async ({
  browser,
}, info) => {
  test.skip(
    local || info.project.name !== "desktop-chromium",
    "hosted shaped cold-route measurement, native covered by real service workflow",
  );
  const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 1,
    measurements = [];
  for (let i = 0; i < 3; i++) {
    const context = await browser.newContext({ serviceWorkers: "block" }),
      page = await context.newPage();
    let bytes = 0;
    const assets: string[] = [];
    page.on("response", (r) => {
      bytes += Number(r.headers()["x-fecim-brotli-bytes"] || 0);
      assets.push(new URL(r.url()).pathname);
    });
    const started = performance.now();
    await page.goto(`http://127.0.0.1:${port}/modules/crossbar`);
    await page.getByRole("button", { name: "Run MVM", exact: true }).click();
    await expect(page.getByTestId("crossbar-status")).toContainText("finished");
    await page
      .getByLabel("Analysis mode", { exact: true })
      .selectOption("vectors");
    await expect(page.locator("#crossbar-output tbody tr")).toHaveCount(8);
    const ms = performance.now() - started;
    expect(bytes).toBeLessThanOrEqual(2 * 1024 ** 2);
    expect(ms).toBeLessThanOrEqual(5000);
    measurements.push({ ms, brotliRouteBytes: bytes, assets });
    await context.close();
  }
  console.log("CROSSBAR_ROUTE_BUDGET " + JSON.stringify(measurements));
  await info.attach("crossbar-route-budget.json", {
    body: JSON.stringify(measurements),
    contentType: "application/json",
  });
});
