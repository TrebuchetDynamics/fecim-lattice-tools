import { expect, test } from "@playwright/test";
import { readFileSync, readdirSync, statSync } from "node:fs";
import { join } from "node:path";
import { brotliCompressSync, gzipSync, constants } from "node:zlib";

test("all hosted workflow families, imports, local actions and faults remain isolated and ephemeral", async ({
  page,
  context,
}) => {
  const requests: { url: string; method: string; body: string | null }[] = [];
  const downloads: string[] = [];
  context.on("request", (request) =>
    requests.push({
      url: request.url(),
      method: request.method(),
      body: request.postData(),
    }),
  );
  page.on("download", (download) =>
    downloads.push(download.suggestedFilename()),
  );
  await page.addInitScript(() => {
    (window as any).forbidden = [];
    const forbid = (name: string) => () => {
      (window as any).forbidden.push(name);
      throw new Error("Forbidden persistence: " + name);
    };
    Storage.prototype.setItem = forbid("storage");
    IDBFactory.prototype.open = forbid("indexedDB");
    if (navigator.storage)
      (navigator.storage as any).getDirectory = forbid("OPFS");
    for (const name of [
      "showSaveFilePicker",
      "showDirectoryPicker",
      "showOpenFilePicker",
    ])
      (window as any)[name] = forbid(name);
    if (navigator.clipboard)
      (navigator.clipboard as any).writeText = forbid("clipboard");
  });
  await page.goto("/");
  await expect(page.getByRole("heading", { level: 1 })).toContainText(
    "lattice-scale",
  );
  expect(requests.some((request) => request.url.includes("/engine/"))).toBe(
    false,
  );
  await page
    .getByRole("link", { name: "Open Hysteresis lab", exact: true })
    .click();
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await expect(page.locator("#core-table tbody tr")).toHaveCount(32);
  await page.getByLabel("Drive", { exact: true }).selectOption("manual");
  await page
    .getByRole("button", { name: "Apply settings", exact: true })
    .click();
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await expect(page.locator("#core-table tbody tr")).toHaveCount(0);
  await page.getByText("Programming and calibration", { exact: true }).click();
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("idle");
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("complete");
  await page
    .getByRole("button", { name: "Calibrate in memory", exact: true })
    .click();
  await expect(page.getByTestId("calibration-status")).toHaveText("fresh");
  await page
    .getByText("Diagnostics and reference context", { exact: true })
    .click();
  await page
    .getByLabel("Reference branch file (CSV or JSON)", { exact: false })
    .setInputFiles({
      name: "branch.csv",
      mimeType: "text/csv",
      buffer: Buffer.from("E_MV_cm,P_uC_cm2\n-1,-20\n1,20\n"),
    });
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("complete");
  await expect(
    page.getByRole("heading", { name: "branch.csv — imported, unverified" }),
  ).toBeVisible();
  await page
    .getByLabel("Reference branch file (CSV or JSON)", { exact: false })
    .setInputFiles({
      name: "branch.json",
      mimeType: "application/json",
      buffer: Buffer.from(
        JSON.stringify({
          source: "Published source",
          material: "Independent material",
          E_V_m: [-1, 1],
          P_C_m2: [-0.2, 0.2],
        }),
      ),
    });
  await expect(page.getByText(/Reference ready in memory/)).toBeVisible();
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(
    page.getByRole("heading", {
      name: "Published source — imported, unverified",
    }),
  ).toBeVisible();
  await expect(page.getByRole("alert")).toHaveCount(0);
  await page.getByLabel("Diagnostic", { exact: true }).selectOption("forc");
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(page.getByTestId("analysis-context")).toContainText("gui-forc");
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await page.getByLabel("Frequency (Hz)", { exact: true }).fill("-1");
  await page
    .getByRole("button", { name: "Apply settings", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "frequency must be positive",
  );
  await page.getByText("Local files and exports", { exact: true }).click();
  for (const name of [
    "Save preset",
    "Load preset",
    "Save calibration",
    "Load calibration",
    "Export loop",
    "Export diagnostics",
    "Verify stored run",
    "EDA tools",
  ])
    await page.getByRole("button", { name, exact: true }).click();
  await expect(page.getByText(/No localhost request was made/)).toBeVisible();
  await page.route("**/engine/manifest.json", async (route) => {
    const response = await route.fetch();
    const manifest = await response.json();
    manifest.UI = "wrong-ui";
    await route.fulfill({ json: manifest });
  });
  await page
    .getByRole("button", { name: "Discard and restart Worker", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "UI/Worker asset fingerprint mismatch",
  );
  await page.unroute("**/engine/manifest.json");
  await page.getByRole("button", { name: "Retry Worker", exact: true }).click();
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await page.getByRole("link", { name: "Projects", exact: true }).first().click();
  await page.getByRole("button", { name: "Open small example", exact: true }).click();
  await page.getByRole("link", { name: "Configuration", exact: true }).click();
  await expect(page.getByLabel("Array rows", { exact: true })).toHaveAttribute("readonly", "");
  await page.getByRole("link", { name: "Experiment", exact: true }).click();
  await page.getByRole("button", { name: "Run study", exact: true }).click();
  await expect(page.getByTestId("experiment-state")).toContainText(/Execution succeeded|finished/);
  await page.getByRole("link", { name: "Results", exact: true }).click();
  await expect(page.locator("#runs tbody tr")).toHaveCount(8);
  await page.getByRole("button", { name: "Verify stored runs", exact: true }).click();
  await expect(page.getByRole("alert")).toContainText("No localhost request was made");
  await page.getByRole("link", { name: "Projects", exact: true }).first().click();
  const documents = Object.fromEntries(["project.yaml", "design.yaml", "sweep.yaml"].map(name => [name, readFileSync(join("..", "examples/research-design-workbench", name)).toString("base64")]));
  await page.getByLabel("Import study byte bundle", { exact: true }).setInputFiles({name:"original.json",mimeType:"application/json",buffer:Buffer.from(JSON.stringify({documents}))});
  await expect(page.getByRole("alert")).toContainText("dimension");
  await page.getByRole("link", { name: "Documentation", exact: true }).first().click();
  await expect(page.getByRole("heading", { name: "Documentation", exact: true })).toBeVisible();
  await expect(page.getByRole("article", { name: "Document reader", exact: true })).toBeVisible();
  expect(await page.evaluate(() => (window as any).forbidden)).toEqual([]);
  expect(downloads).toEqual([]);
  const origin = new URL(page.url()).origin;
  for (const request of requests) {
    const url = new URL(request.url);
    expect(url.origin).toBe(origin); // only the explicitly opened test origin, never discovery
    if (url.protocol === "blob:") continue; // same-origin in-memory Worker/runtime, no network
    // Reviewed public corpus IDs only; not arbitrary documents or filesystem paths.
    expect(url.pathname).toMatch(/^\/(?:$|hysteresis$|assets\/|engine\/|docs-manifest\.json$|docs-content\/(?:curriculum|hysteresis|crossbar|mnist|circuits|comparison|eda|glossary|trust|honesty|park-correction|validation|citations)\.json$)/);
    expect(request.method).toBe("GET");
    expect(request.body).toBeNull();
  }
});

test("complete real route transfer and cold Go sine result meet the production guardrails", async ({
  browser,
}, testInfo) => {
  test.skip(
    testInfo.project.name !== "desktop-chromium",
    "one controlled measurement per build profile",
  );
  const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 1;
  const root =
    process.env.FECIM_WEB_PROFILE === "local" ? "dist-local" : "dist";
  const engine = [
    "hysteresis.wasm",
    "wasm_exec.js",
    "hysteresis-worker.mjs",
  ].map((name) => readFileSync(join(root, "engine", name)));
  const br = (bytes: Buffer) =>
    brotliCompressSync(bytes, {
      params: { [constants.BROTLI_PARAM_QUALITY]: 11 },
    }).length;
  const raw = engine.reduce((sum, bytes) => sum + bytes.length, 0),
    brotli = engine.reduce((sum, bytes) => sum + br(bytes), 0),
    gzip = engine.reduce(
      (sum, bytes) => sum + gzipSync(bytes, { level: 9 }).length,
      0,
    );
  expect(raw).toBeLessThanOrEqual(6 * 1024 * 1024);
  expect(brotli).toBeLessThanOrEqual(1.5 * 1024 * 1024);
  expect(gzip).toBeLessThanOrEqual(2 * 1024 * 1024);
  const measurements = [];
  for (let iteration = 0; iteration < 3; iteration++) {
    const context = await browser.newContext({ serviceWorkers: "block" });
    const page = await context.newPage();
    let bytes = 0;
    const resources: string[] = [];
    page.on("response", (response) => {
      if (!/^https?:/.test(response.url())) return;
      bytes += Number(response.headers()["x-fecim-brotli-bytes"] || 0);
      resources.push(new URL(response.url()).pathname);
    });
    const started = performance.now();
    await page.goto("http://127.0.0.1:" + port + "/hysteresis");
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    await page
      .getByRole("button", { name: "Advance batch", exact: true })
      .click();
    // Completed marker includes paused Go result and the first complete signed
    // 32-row page, not merely Worker readiness or a synthetic timing fixture.
    await expect(page.locator("#core-table tbody tr")).toHaveCount(32);
    await expect(page.getByTestId("core-state")).toHaveText("paused");
    await expect(page.locator("#core-table tbody")).toContainText("-");
    const ms = performance.now() - started;
    expect(bytes).toBeLessThanOrEqual(2 * 1024 * 1024);
    expect(ms).toBeLessThanOrEqual(5000);
    measurements.push({ iteration, ms, brotliRouteBytes: bytes, resources });
    await context.close();
  }
  const manifest = JSON.parse(
    readFileSync(join(root, "engine/manifest.json"), "utf8"),
  );
  const evidence = {
    profile: process.env.FECIM_WEB_PROFILE || "hosted",
    browser: browser.version(),
    node: process.version,
    platform: process.platform,
    arch: process.arch,
    identity: manifest.identity,
    UI: manifest.UI,
    engine: { raw, brotli, gzip },
    network:
      "worker-private static server, aggregate 10 Mb/s +100 ms per-request RTT; fresh context and cache:no-store",
    marker: "paused Go sine result + 32 signed first-page rows",
    measurements,
  };
  console.log("ROUTE_BUDGET " + JSON.stringify(evidence));
  await testInfo.attach("route-budget.json", {
    body: JSON.stringify(evidence, null, 2),
    contentType: "application/json",
  });
});
