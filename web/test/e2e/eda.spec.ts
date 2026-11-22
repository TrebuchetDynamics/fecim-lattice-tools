import { expect, test } from "@playwright/test";
import { spawn, execFileSync } from "node:child_process";
import { mkdtempSync, mkdirSync, writeFileSync, rmSync } from "node:fs";
import { join } from "node:path";
const local = process.env.FECIM_WEB_PROFILE === "local";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 14;
let root = "",
  launch = "",
  child: ReturnType<typeof spawn> | undefined;
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "eda-native-"));
  const binary = join(root, "fecim");
  execFileSync("node", ["src/local/build-native.mjs", binary], {
    timeout: 120_000,
  });
  const bin = join(root, "private-tools");
  mkdirSync(bin);
  writeFileSync(
    join(bin, "iverilog"),
    "#!/bin/sh\nif [ \"$1\" = '-V' ]; then echo 'PRIVATE browser cancellation fixture, not real Icarus'; exit; fi\nsleep 30 &\nwait\n",
    { mode: 0o700 },
  );
  child = spawn(binary, ["serve", "--root", root, "--port", String(port)], {
    env: { ...process.env, PATH: bin + ":" + process.env.PATH },
  });
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

test("EDA real profile: typed builder/artifacts locally, escaped unverified Learn imports hosted", async ({
  page,
}, testInfo) => {
  const pageErrors: string[] = [];
  page.on("pageerror", (error) => pageErrors.push(error.message));
  const scientificRequests: string[] = [];
  page.on("request", (r) => {
    if (r.url().includes("/api/")) scientificRequests.push(r.url());
  });
  let snapshotURL = "";
  if (local) {
    await page.goto(launch);
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    await page
      .getByRole("link", { name: "Projects", exact: true })
      .first()
      .click();
    await page.getByLabel("Project name", { exact: true }).fill("eda-fixture");
    await page
      .getByRole("button", { name: "Create project", exact: true })
      .click();
    await expect(page.getByTestId("research-status")).toContainText(
      "Project opened",
    );
    await page
      .getByRole("link", { name: /FeCIM Lattice Tools/ })
      .first()
      .click();
    const sessionResponse = page.waitForResponse(
      (r) =>
        r.url().endsWith("/api/v1/modules/eda/sessions") &&
        r.request().method() === "POST",
    );
    await page
      .getByRole("link", { name: "Explore EDA Export", exact: true })
      .click();
    const response = await sessionResponse;
    snapshotURL = new URL(
      `/api/v1/module-sessions/${(await response.json()).session}`,
      response.url(),
    ).href;
  } else await page.goto("/modules/eda");
  await expect(
    page.getByRole("heading", { name: "EDA", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("heading", { name: "Learn and qualification", exact: true }),
  ).toBeVisible();
  await expect(page.getByRole("heading", { level: 1 })).toHaveCount(1);
  const regions = page.getByRole("main").getByRole("region");
  if (!local) {
    const imports = page.getByRole("region", {
      name: "Hosted read-only inspection",
    });
    await expect(regions.first()).toHaveAccessibleName(
      "Hosted read-only inspection",
    );
    await expect(page.getByRole("button", { name: /generate/i })).toHaveCount(
      0,
    );
    const importInput = imports.getByLabel(
      "Import artifact for read-only inspection",
      { exact: true },
    );
    await importInput.setInputFiles({
      name: "oversized.txt",
      mimeType: "text/plain",
      buffer: Buffer.alloc(8 * 1024 * 1024 + 1),
    });
    await expect(page.getByRole("alert")).toHaveText(
      "Read-only import exceeds 8 MiB.",
    );
    await expect(page.getByRole("alert")).toBeFocused();
    await expect(imports.getByTestId("eda-import")).toHaveCount(0);
    await importInput.setInputFiles({
      name: "unverified.spice",
      mimeType: "text/plain",
      buffer: Buffer.from(
        "<script>window.edaInjected=true</script>\n* imported",
      ),
    });
    await expect(page.getByTestId("eda-import")).toContainText(
      "<script>window.edaInjected=true</script>",
    );
    await expect(
      imports.getByRole("heading", { name: "unverified.spice", exact: true }),
    ).toBeVisible();
    await expect(
      imports.getByText(
        /Imported\/unverified — no generation or tool validation.*not foundry-ready/i,
      ),
    ).toBeVisible();
    await expect(page.getByRole("alert")).toHaveCount(0);
    await imports.screenshot({
      path: testInfo.outputPath("hosted-import.png"),
    });
    expect(await page.evaluate(() => "edaInjected" in window)).toBe(false);
    expect(scientificRequests).toEqual([]);
    expect(pageErrors).toEqual([]);
    return;
  }
  const source = page.getByRole("region", {
    name: "Source design",
    exact: true,
  });
  const generation = page.getByRole("region", {
    name: "Generation",
    exact: true,
  });
  const inspection = page.getByRole("region", {
    name: "Artifact inspection",
    exact: true,
  });
  await expect(regions.nth(0)).toHaveAccessibleName("Source design");
  await expect(regions.nth(1)).toHaveAccessibleName("Generation");
  await expect(regions.nth(2)).toHaveAccessibleName("Artifact inspection");
  await expect(
    source.getByText("Opened project: HZO Device-Array-Circuit Study"),
  ).toBeVisible();
  await expect(
    source.getByLabel("Typed builder config (JSON)", { exact: true }),
  ).toContainText('"array_rows": 8');
  await generation
    .getByRole("button", { name: "Generate all formats", exact: true })
    .focus();
  await page.keyboard.press("Enter");
  await expect(page.getByTestId("eda-status")).toContainText(
    "Generated educational artifacts",
    { timeout: 30_000 },
  );
  await expect(page.getByTestId("eda-origin")).toContainText("freeform");
  await expect(page.getByTestId("eda-origin")).toContainText(
    "ESTIMATED_SCALAR",
  );
  const inspect = (name: string) =>
    inspection.getByRole("button", { name: `Inspect ${name}`, exact: true });
  const selection = inspection.getByRole("region", {
    name: "Selected artifact",
    exact: true,
  });
  const displayedPreview = selection.getByTestId("eda-preview");
  await inspect("manifest.json").click();
  await expect(displayedPreview).toContainText('"files_sha256"');
  const manifest = JSON.parse((await displayedPreview.textContent())!);
  const bundleIdentity = await inspection
    .getByTestId("eda-origin")
    .textContent();
  await inspect("design.spice").press("Enter");
  await expect(displayedPreview).toContainText("* Array: 64 cells");
  await expect(selection.getByText(/^design\.spice bytes 0\.\./)).toBeVisible();
  await expect(
    selection.getByText(manifest.files_sha256["design.spice"], { exact: true }),
  ).toBeVisible();

  // A failed switch must not pair the previous text with a new artifact identity.
  await page.route(`${snapshotURL}/actions`, (route) => route.abort(), {
    times: 1,
  });
  await inspect("design.v").press("Enter");
  await expect(page.getByRole("alert")).toBeFocused();
  await expect(inspect("design.spice")).toHaveAttribute("aria-pressed", "true");
  await expect(selection.getByText(/^design\.spice bytes 0\.\./)).toBeVisible();
  await expect(displayedPreview).toContainText("* Array: 64 cells");

  await inspect("design.v").press("Enter");
  await expect(page.getByRole("alert")).toHaveCount(0);
  await expect(inspect("design.v")).toHaveAttribute("aria-pressed", "true");
  await expect(inspect("design.spice")).toHaveAttribute(
    "aria-pressed",
    "false",
  );
  await expect(displayedPreview).toContainText("parameter ROWS = 8;");
  await expect(displayedPreview).not.toContainText("* Array: 64 cells");
  await expect(selection.getByText(/^design\.v bytes 0\.\./)).toBeVisible();
  await expect(
    selection.getByText(manifest.files_sha256["design.v"], { exact: true }),
  ).toBeVisible();
  await expect(selection.getByTestId("eda-origin")).toHaveText(bundleIdentity!);
  await expect(selection.getByTestId("eda-origin")).toContainText(
    "Structural checks only",
  );
  await selection.screenshot({
    path: testInfo.outputPath("local-selection.png"),
  });
  const reconnect = page.getByRole("button", {
    name: "Reconnect authoritative state",
    exact: true,
  });
  const defaults = page.getByRole("button", {
    name: "Load Go defaults",
    exact: true,
  });
  const builder = page.getByLabel("Typed builder config (JSON)", {
    exact: true,
  });
  // Fetch the authoritative response before holding delivery, not job polling.
  async function gateResponse(
    url: string,
    method: string,
    trigger: () => Promise<void>,
    pending: () => Promise<void>,
  ) {
    let release!: () => void;
    const gate = new Promise<void>((resolve) => {
      release = resolve;
    });
    const bound = setTimeout(release, 20_000);
    let captured = false;
    const handler = async (route: import("@playwright/test").Route) => {
      if (route.request().method() !== method) return route.continue();
      const response = await route.fetch({ timeout: 10_000 });
      captured = true;
      await gate;
      await route.fulfill({ response });
    };
    await page.route(url, handler);
    try {
      await trigger();
      await expect.poll(() => captured).toBe(true);
      await pending();
    } finally {
      clearTimeout(bound);
      release();
      await page.unrouteAll({ behavior: "wait" });
    }
  }
  await gateResponse(
    snapshotURL,
    "GET",
    () => reconnect.click(),
    async () => {
      await expect(defaults).toBeDisabled();
      await expect(reconnect).toBeDisabled();
    },
  );
  await expect(reconnect).toBeEnabled();
  await page.getByLabel("Rows", { exact: true }).fill("4");
  await gateResponse(
    `${snapshotURL}/actions`,
    "POST",
    () => defaults.click(),
    async () => {
      await expect(reconnect).toBeDisabled();
    },
  );
  await expect(builder).toContainText('"array_rows": 4');
  await expect(reconnect).toBeEnabled();
  await page
    .getByRole("button", { name: "Validate structure", exact: true })
    .click();
  await expect(reconnect).toBeEnabled();
  await expect(builder).toContainText('"array_rows": 4');
  await expect(page.getByRole("alert")).toHaveCount(0);
  const retainedDraft = await builder.inputValue();
  const retainedPreview = await page.getByTestId("eda-preview").textContent();
  await page.route(snapshotURL, (route) => route.abort(), { times: 1 });
  await reconnect.click();
  await expect(page.getByRole("alert")).toBeFocused();
  await expect(reconnect).toBeEnabled();
  await expect(defaults).toBeEnabled();
  await expect(builder).toHaveValue(retainedDraft);
  await expect(page.getByTestId("eda-preview")).toHaveText(retainedPreview!);
  await reconnect.click();
  await expect(page.getByRole("alert")).toHaveCount(0);
  await expect(page.getByTestId("eda-tool")).toContainText("structural-only");
  const downloaded = page.waitForEvent("download");
  await page
    .getByRole("link", { name: "Download package.zip", exact: true })
    .click();
  expect((await downloaded).suggestedFilename()).toBe("package.zip");
  await page
    .getByRole("button", { name: "Discover installed tools", exact: true })
    .click();
  await expect(page.getByTestId("eda-tool")).toContainText(
    "discovered-not-validated",
  );
  // A real private process fixture remains cancellable after a polling outage.
  const jobPattern = "**/api/v1/module-sessions/*/jobs/*";
  const outage = async (route: import("@playwright/test").Route) => {
    await route.abort();
  };
  await page.route(jobPattern, outage, { times: 1 });
  await page
    .getByRole("button", { name: "Run installed tool", exact: true })
    .click();
  await expect(page.getByRole("alert")).toBeVisible();
  await page.unroute(jobPattern, outage);
  await page
    .getByRole("button", { name: "Reconnect authoritative state", exact: true })
    .click();
  await expect(page.getByTestId("eda-status")).toContainText("running");
  await expect(
    page.getByRole("button", { name: "Cancel EDA job", exact: true }),
  ).toBeEnabled();
  await page
    .getByRole("button", { name: "Cancel EDA job", exact: true })
    .click();
  await expect
    .poll(async () => {
      await page
        .getByRole("button", {
          name: "Reconnect authoritative state",
          exact: true,
        })
        .click();
      return page.getByTestId("eda-status").textContent();
    })
    .toContain("cancelled");
  await expect(page.getByTestId("eda-tool")).toContainText(
    "PRIVATE browser cancellation fixture",
  );
  const config = page.getByLabel("Typed builder config (JSON)", {
    exact: true,
  });
  const bad = JSON.parse(await config.inputValue());
  bad.config.array_rows = 257;
  await config.fill(JSON.stringify(bad));
  await page
    .getByRole("button", { name: "Generate all formats", exact: true })
    .click();
  await expect(page.getByRole("alert")).toBeFocused();
  // Selected science uses the real stored-run seam, not the dirty freeform draft.
  await page.getByRole("link", { name: "Results", exact: true }).click();
  await page.getByRole("link", { name: "Experiment", exact: true }).click();
  await page.getByRole("button", { name: "Run study", exact: true }).click();
  await expect(page.getByTestId("experiment-state")).toContainText(
    /Execution succeeded|finished/,
    { timeout: 60_000 },
  );
  await page.getByRole("link", { name: "Results", exact: true }).click();
  await page.locator("#runs input[type=checkbox]").first().check();
  await page
    .getByRole("link", { name: /FeCIM Lattice Tools/ })
    .first()
    .click();
  await page
    .getByRole("link", { name: "Explore EDA Export", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Export verified selected run", exact: true })
    .click();
  await expect(page.getByTestId("eda-origin")).toContainText("selected-run", {
    timeout: 30_000,
  });
  await expect(page.getByTestId("eda-preview")).toHaveText(
    "Choose an artifact to inspect",
  );
  await page
    .getByRole("button", { name: "Inspect manifest.json", exact: true })
    .click();
  await expect(page.getByTestId("eda-preview")).toContainText(
    '"resolved_design"',
  );
  await expect(page.getByTestId("eda-preview")).toContainText(
    "not mapped to a foundry PDK",
  );
  expect(pageErrors).toEqual([]);
});
