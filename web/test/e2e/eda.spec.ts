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
}) => {
  const scientificRequests: string[] = [];
  page.on("request", (r) => {
    if (r.url().includes("/api/")) scientificRequests.push(r.url());
  });
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
    await page
      .getByRole("link", { name: "Explore EDA Export", exact: true })
      .click();
  } else await page.goto("/modules/eda");
  await expect(
    page.getByRole("heading", { name: "EDA", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("heading", { name: "Learn and qualification", exact: true }),
  ).toBeVisible();
  if (!local) {
    await expect(
      page.getByRole("button", { name: "Generate all formats", exact: true }),
    ).toHaveCount(0);
    await page
      .getByLabel("Import artifact for read-only inspection", { exact: true })
      .setInputFiles({
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
      page.getByText("Imported/unverified — no generation or tool validation"),
    ).toBeVisible();
    expect(await page.evaluate(() => "edaInjected" in window)).toBe(false);
    expect(scientificRequests).toEqual([]);
    return;
  }
  await expect(
    page.getByLabel("Typed builder config (JSON)", { exact: true }),
  ).toContainText('"array_rows": 8');
  await page
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
  await page
    .getByRole("button", { name: "Inspect design.spice", exact: true })
    .click();
  await expect(page.getByTestId("eda-preview")).toContainText(
    "* Array: 64 cells",
  );
  await page
    .getByRole("button", { name: "Validate structure", exact: true })
    .click();
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
});
