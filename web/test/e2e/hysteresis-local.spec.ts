import { expect, test } from "@playwright/test";
import { spawn, execFileSync } from "node:child_process";
import { mkdtempSync, readFileSync, readdirSync, rmSync } from "node:fs";
import { join } from "node:path";

let child: ReturnType<typeof spawn>;
let root: string;
let launch: string;
let output = "";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 10;
async function launchLink(previous = "") {
  await expect
    .poll(
      () =>
        output
          .match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/g)
          ?.at(-1),
      { timeout: 30_000 },
    )
    .not.toBe(previous || undefined);
  return output
    .match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/g)!
    .at(-1)!;
}
test.beforeAll(async () => {
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "hys-local-"));
  const binary = join(root, "fecim-local");
  execFileSync("node", ["src/local/build-native.mjs", binary], {
    stdio: "pipe",
    timeout: 120_000,
  });
  child = spawn(binary, ["serve", "--root", root, "--port", String(port)], {
    stdio: ["pipe", "pipe", "pipe"],
  });
  child.stdout!.on("data", (data) => {
    output += data;
  });
  child.stderr!.on("data", (data) => {
    output += data;
  });
  output = "";
  launch = await launchLink();
});
test.afterAll(async () => {
  if (child && child.exitCode === null) {
    const exited = new Promise<void>((resolve) =>
      child.once("exit", () => resolve()),
    );
    child.kill("SIGTERM");
    await exited;
  }
  if (root) rmSync(root, { recursive: true, force: true });
});

test("real embedded local session, explicit file workflows, conflict, error and logout", async ({
  page,
  browser,
}) => {
  const requests: string[] = [];
  page.on("request", (request) => requests.push(request.url()));
  await page.goto(launch);
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await expect(page).toHaveURL(`http://127.0.0.1:${port}/hysteresis`);
  await page.getByText("Local files and exports", { exact: true }).click();
  await expect(page.getByLabel("Granted root")).toBeEnabled();
  expect(readdirSync(root).filter((name) => name !== "fecim-local")).toEqual(
    [],
  );
  await page
    .getByLabel("Relative file path", { exact: true })
    .fill("preset.json");
  await page.getByRole("button", { name: "Save preset", exact: true }).click();
  await expect(page.getByTestId("file-status")).toContainText(
    "Saved preset.json",
  );
  const preset = JSON.parse(readFileSync(join(root, "preset.json"), "utf8"));
  expect(preset.metadata.module).toBe("hysteresis");
  await page.getByRole("button", { name: "Save preset", exact: true }).click();
  await expect(page.getByTestId("file-status")).toContainText("Conflict");
  await page.getByRole("button", { name: "Load preset", exact: true }).click();
  await expect(page.getByTestId("file-status")).toContainText(
    "imported/unverified",
  );
  await page
    .getByLabel("Relative file path", { exact: true })
    .fill("../escape.json");
  await page.getByRole("button", { name: "Save preset", exact: true }).click();
  await expect(page.getByTestId("file-status")).toContainText(
    "invalid granted root or relative path",
  );
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await expect(page.locator("#core-table tbody tr").first()).toBeVisible();
  await page.getByLabel("Relative file path", { exact: true }).fill("loop.csv");
  await page.getByRole("button", { name: "Export loop", exact: true }).click();
  await expect(page.getByTestId("file-status")).toContainText("Saved loop.csv");
  expect(readFileSync(join(root, "loop.csv"), "utf8")).toContain(
    "browser simulation data",
  );
  // Same Go Worker produces the calibration. No automatic calibration file write.
  await page.getByText("Programming and calibration", { exact: true }).click();
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Calibrate in memory", exact: true })
    .click();
  await expect(page.getByTestId("calibration-status")).toHaveText("fresh", {
    timeout: 30_000,
  });
  await page
    .getByLabel("Relative file path", { exact: true })
    .fill("calibration.json");
  await page
    .getByRole("button", { name: "Save calibration", exact: true })
    .click();
  await expect(page.getByTestId("file-status")).toContainText(
    "Saved calibration.json",
  );
  await page
    .getByRole("button", { name: "Load calibration", exact: true })
    .click();
  await expect(page.getByTestId("file-status")).toContainText(
    "imported/unverified",
  );
  await page
    .getByText("Diagnostics and reference context", { exact: true })
    .click();
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("complete");
  await page
    .getByLabel("Relative file path", { exact: true })
    .fill("diagnostic.json");
  await page
    .getByRole("button", { name: "Export diagnostics", exact: true })
    .click();
  await expect(page.getByTestId("file-status")).toContainText(
    "Saved diagnostic.json",
  );
  const diagnostic = JSON.parse(
    readFileSync(join(root, "diagnostic.json"), "utf8"),
  );
  expect(diagnostic.ExperimentalValidation).toBe(false);
  expect(diagnostic.Evidence).toContain("browser simulation data");
  const second = await browser.newContext();
  const replay = await second.newPage();
  await replay.goto(launch);
  await expect(replay.getByRole("status")).toContainText(
    "expired, used or rejected",
  );
  await second.close();
  child.stdin!.write("renew\n");
  const signedInRenewal = await launchLink(launch);
  await page.goto(signedInRenewal);
  await expect(page).toHaveURL(`http://127.0.0.1:${port}/hysteresis`);
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await page.getByText("Local files and exports", { exact: true }).click();
  await page
    .getByRole("button", { name: "Log out local session", exact: true })
    .click();
  await expect(page.getByTestId("file-status")).toContainText("Logged out");
  await page.getByRole("button", { name: "Save preset", exact: true }).click();
  await expect(page.getByTestId("file-status")).toContainText(
    /authentication required/i,
  );
  child.stdin!.write("renew\n");
  const renewed = await launchLink(signedInRenewal);
  await page.goto(renewed);
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  expect(
    requests.every(
      (url) => !url.includes("bootstrap=") && !url.includes("secret="),
    ),
  ).toBe(true);
  expect(
    await page.evaluate(() => [localStorage.length, sessionStorage.length]),
  ).toEqual([0, 0]);
});

test("hosted file action remains instructions with no local transport", async ({
  page,
}) => {
  test.skip(
    process.env.FECIM_WEB_PROFILE === "local",
    "Hosted-only transport isolation case",
  );
  const calls: string[] = [];
  page.on("request", (r) => {
    if (r.url().includes("/api/")) calls.push(r.url());
  });
  await page.goto("/hysteresis");
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await page.getByText("Local files and exports", { exact: true }).click();
  await page.getByRole("button", { name: "Save preset", exact: true }).click();
  await expect(page.getByRole("status").last()).toContainText(
    "fecim-lattice-tools serve --root DIRECTORY",
  );
  expect(calls).toEqual([]);
});
