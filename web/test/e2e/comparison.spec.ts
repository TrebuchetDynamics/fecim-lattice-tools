import { expect, test } from "@playwright/test";
import { readFileSync, mkdtempSync, rmSync } from "node:fs";
import { join } from "node:path";
import { spawn, execFileSync } from "node:child_process";
import { brotliCompressSync, gzipSync } from "node:zlib";

const local = process.env.FECIM_WEB_PROFILE === "local";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 14;
let root = "",
  launch = "",
  child: ReturnType<typeof spawn> | undefined;
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "comparison-local-"));
  const binary = join(root, "fecim");
  execFileSync("node", ["src/local/build-native.mjs", binary], {
    timeout: 120_000,
  });
  child = spawn(binary, ["serve", "--root", root, "--port", String(port)]);
  let output = "";
  child.stdout!.on("data", (data) => {
    output += data;
  });
  child.stderr!.on("data", (data) => {
    output += data;
  });
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
    const exited = new Promise((resolve) => child!.once("exit", resolve));
    child.kill("SIGTERM");
    await exited;
  }
  if (root) rmSync(root, { recursive: true, force: true });
});

test("real comparison scenarios, source evidence, sensitivity and local versioned exports", async ({
  page,
}, info) => {
  page.setDefaultTimeout(15000);
  const project = "comparison-" + info.project.name;
  if (local) {
    await page.goto(launch);
    await page
      .getByRole("link", { name: "Modules", exact: true })
      .first()
      .click();
    await page
      .getByRole("link", { name: "Explore Technology Comparison", exact: true })
      .click();
    await expect(
      page.getByLabel("Inference rate", { exact: true }),
    ).toBeEnabled();
    await expect(
      page.getByRole("combobox", { name: /^Opened export project/ }),
    ).toHaveValue("");
    await page
      .getByRole("link", { name: "Projects", exact: true })
      .first()
      .click();
    await page.getByLabel("Project name", { exact: true }).fill(project);
    await page
      .getByRole("button", { name: "Create project", exact: true })
      .click();
    await expect(page.getByTestId("research-status")).toContainText(
      "Project opened",
    );
    await page
      .getByRole("link", { name: "Modules", exact: true })
      .first()
      .click();
    await page
      .getByRole("link", { name: "Explore Technology Comparison", exact: true })
      .click();
    await expect(
      page.getByRole("heading", { name: "Comparison", exact: true }),
    ).toBeFocused();
    await page
      .getByRole("button", { name: "Refresh opened projects", exact: true })
      .click();
    await expect(
      page.getByLabel("Opened export project", { exact: true }),
    ).not.toHaveValue("");
  } else await page.goto("/modules/comparison");
  if (!local)
    await expect(
      page.getByRole("heading", { name: "Comparison", exact: true }),
    ).toBeFocused();
  await expect(
    page.getByLabel("Inference rate", { exact: true }),
  ).toBeEnabled();
  await page.getByLabel("Workload", { exact: true }).selectOption("MNIST");
  await page
    .getByLabel("Scenario", { exact: true })
    .selectOption("conservative");
  await page.getByLabel("Inference rate", { exact: true }).fill("2");
  await page
    .getByLabel("Sensitivity output", { exact: true })
    .selectOption("energy_reduction_pct");
  await page.getByRole("button", { name: "Calculate", exact: true }).click();
  await expect(page.locator("#comparison-energy tbody tr")).toHaveCount(3);
  await expect(
    page.getByLabel("Sensitivity output", { exact: true }),
  ).toHaveValue("energy_reduction_pct");
  await page
    .getByRole("button", { name: "Rank sensitivity", exact: true })
    .click();
  await expect(page.locator("#comparison-sensitivity caption")).toContainText(
    "energy_reduction_pct",
  );
  await expect(
    page.locator("#comparison-sensitivity tbody tr").first(),
  ).toContainText("gpu_energy_pj_per_mac");
  await expect(page.locator("#comparison-energy")).toContainText("172800");
  await expect(page.locator("#comparison-summary")).toContainText(
    "No physical energy-duration input",
  );
  await expect(page.locator("#comparison-architectures caption")).toContainText(
    "Separate architecture-demonstration",
  );
  await page
    .getByRole("button", { name: "Compare with baseline", exact: true })
    .click();
  await expect(page.locator("#comparison-diff caption")).toBeFocused();
  await expect(
    page
      .locator("#comparison-diff tbody tr")
      .filter({ hasText: "energy_reduction_pct" })
      .locator("td"),
  ).toContainText("-0.6666");
  await page
    .getByLabel("Sensitivity output", { exact: true })
    .selectOption("tco_reduction_pct");
  await page
    .getByRole("button", { name: "Rank sensitivity", exact: true })
    .click();
  await expect(page.locator("#comparison-sensitivity caption")).toBeFocused();
  await expect(page.locator("#comparison-sensitivity tbody tr")).toHaveCount(
    12,
  );
  await expect(
    page.locator("#comparison-sensitivity tbody tr").first(),
  ).toContainText("gpu_tco_usd_per_month");
  await expect(
    page
      .locator("#comparison-sensitivity tbody tr")
      .first()
      .locator("td")
      .nth(1),
  ).toContainText("-3.0409");
  await page
    .getByLabel("Sensitivity output", { exact: true })
    .selectOption("co2_reduction_pct");
  await page
    .getByRole("button", { name: "Rank sensitivity", exact: true })
    .click();
  await expect(page.locator("#comparison-sensitivity caption")).toContainText(
    "Not available",
  );
  await page
    .getByRole("button", { name: "Toggle presentation", exact: true })
    .click();
  await expect(page.getByTestId("comparison-status")).toContainText(
    "Presentation",
  );
  await expect(page.locator("#comparison-sources")).toHaveCount(0);
  await page.getByLabel("Inference rate", { exact: true }).fill("NaN");
  await page.getByRole("button", { name: "Calculate", exact: true }).click();
  await expect(page.getByRole("alert")).toBeFocused();
  await expect(page.getByLabel("Inference rate", { exact: true })).toHaveValue(
    "NaN",
  );
  await expect(page.locator("#comparison-energy")).toContainText("172800");
  await page
    .getByRole("button", { name: "Dismiss error", exact: true })
    .click();
  await page
    .getByText("Source evidence and limitations", { exact: true })
    .click();
  await expect(page.locator("#comparison-evidence")).toContainText(
    "not a statistical interval",
  );
  await page.getByRole("button", { name: "Export JSON", exact: true }).click();
  if (local) {
    // The route's export picker discovers already-opened opaque project IDs; no path input.
    await expect(page.getByTestId("comparison-status")).toContainText(
      "Exported exports/comparison/",
    );
    const status = await page.getByTestId("comparison-status").textContent();
    const path = status!.match(
      /exports\/comparison\/comparison-v2-[a-f0-9]+\.json/,
    )![0];
    const artifact = JSON.parse(
      readFileSync(join(root, project, path), "utf8"),
    );
    expect(artifact.schema_version).toBe("fecim-comparison/2");
    expect(artifact.result.inferences_per_second).toBe(2);
    expect(artifact.result.energy_cpu_gpu_fecim[0].daily_inferences).toBe(
      172800,
    );
    await page.getByRole("button", { name: "Export CSV", exact: true }).click();
    await expect(page.getByTestId("comparison-status")).toContainText(".csv");
  } else
    await expect(page.getByTestId("comparison-status")).toContainText(
      "No localhost request was made",
    );
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await expect(page.getByLabel("Inference rate", { exact: true })).toHaveValue(
    "10000",
  );
});

test("comparison hosted actions stay private and enforce one active science owner", async ({
  page,
  context,
}) => {
  test.skip(local, "hosted privacy boundary");
  const requests: { url: string; method: string; body: string | null }[] = [],
    downloads: string[] = [];
  context.on("request", (r) =>
    requests.push({ url: r.url(), method: r.method(), body: r.postData() }),
  );
  page.on("download", (d) => downloads.push(d.suggestedFilename()));
  await page.addInitScript(() => {
    (window as any).forbidden = [];
    const forbid = (name: string) => () => {
      (window as any).forbidden.push(name);
      throw new Error(name);
    };
    Storage.prototype.setItem = forbid("storage");
    IDBFactory.prototype.open = forbid("indexedDB");
    if (navigator.storage)
      (navigator.storage as any).getDirectory = forbid("OPFS");
    if (navigator.clipboard)
      (navigator.clipboard as any).writeText = forbid("clipboard");
  });
  await page.goto("/hysteresis");
  await page.getByRole("button", { name: "Run", exact: true }).click();
  await page
    .getByRole("link", { name: "Modules", exact: true })
    .first()
    .click();
  await page
    .getByRole("link", { name: "Explore Technology Comparison", exact: true })
    .click();
  await page.getByRole("button", { name: "Calculate", exact: true }).click();
  await expect(page.getByRole("alert")).toContainText(
    "Another scientific computation",
  );
  await page
    .getByRole("link", { name: "Hysteresis lab", exact: true })
    .first()
    .click();
  await page.getByRole("button", { name: "Pause", exact: true }).click();
  await page
    .getByRole("link", { name: "Modules", exact: true })
    .first()
    .click();
  await page
    .getByRole("link", { name: "Explore Technology Comparison", exact: true })
    .click();
  await page.getByRole("button", { name: "Calculate", exact: true }).click();
  await expect(page.getByTestId("comparison-status")).toContainText(
    "Calculated scenario",
  );
  await page.getByRole("button", { name: "Export CSV", exact: true }).click();
  await expect(page.getByTestId("comparison-status")).toContainText(
    "No localhost request was made",
  );
  expect(await page.evaluate(() => (window as any).forbidden)).toEqual([]);
  expect(downloads).toEqual([]);
  const origin = new URL(page.url()).origin;
  for (const r of requests) {
    expect(new URL(r.url).origin).toBe(origin);
    expect(r.method).toBe("GET");
    expect(r.body).toBeNull();
    expect(new URL(r.url).pathname).not.toMatch(/^\/api\//);
  }
});

test("comparison production Worker and full cold route meet transfer and latency budgets", async ({
  browser,
}, info) => {
  test.skip(
    info.project.name !== "desktop-chromium" || local,
    "one hosted controlled measurement",
  );
  const engine = [
    "comparison.wasm",
    "wasm_exec.js",
    "hysteresis-worker.mjs",
  ].map((n) => readFileSync(join("dist/engine", n)));
  const sizes = {
    raw: engine.reduce((n, b) => n + b.length, 0),
    brotli: engine.reduce((n, b) => n + brotliCompressSync(b).length, 0),
    gzip: engine.reduce((n, b) => n + gzipSync(b).length, 0),
  };
  expect(sizes.raw).toBeLessThanOrEqual(6 * 1024 * 1024);
  expect(sizes.brotli).toBeLessThanOrEqual(1.5 * 1024 * 1024);
  expect(sizes.gzip).toBeLessThanOrEqual(2 * 1024 * 1024);
  const samples = [];
  for (let i = 0; i < 3; i++) {
    const ctx = await browser.newContext({ serviceWorkers: "block" }),
      page = await ctx.newPage();
    let bytes = 0;
    const resources: string[] = [];
    page.on("response", (r) => {
      if (/^https?:/.test(r.url())) {
        bytes += Number(r.headers()["x-fecim-brotli-bytes"] || 0);
        resources.push(new URL(r.url()).pathname);
      }
    });
    const start = performance.now();
    await page.goto(
      `http://127.0.0.1:${Number(process.env.AUTOWORK_PORT_BASE || 20200) + 1}/modules/comparison`,
    );
    await expect(page.locator("#comparison-energy tbody tr")).toHaveCount(3);
    const ms = performance.now() - start;
    expect(bytes).toBeLessThanOrEqual(2 * 1024 * 1024);
    expect(ms).toBeLessThanOrEqual(5000);
    expect(resources.some((r) => r.endsWith("comparison.wasm"))).toBe(true);
    expect(resources.some((r) => r.endsWith("hysteresis.wasm"))).toBe(false);
    samples.push({ ms, bytes, resources });
    await ctx.close();
  }
  console.log("COMPARISON_ROUTE_BUDGET " + JSON.stringify({ sizes, samples }));
  await info.attach("comparison-route-budget.json", {
    body: JSON.stringify({ sizes, samples }),
    contentType: "application/json",
  });
});
