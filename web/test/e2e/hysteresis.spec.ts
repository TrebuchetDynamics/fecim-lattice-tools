import { expect, test } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";

async function openLab(page) {
  await page.goto("/hysteresis");
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
}
async function apply(page) {
  await page
    .getByRole("button", { name: "Apply settings", exact: true })
    .click();
  await expect(page.getByTestId("settings-state")).toHaveText(
    "Current settings",
  );
}
async function advance(page) {
  const button = page.getByRole("button", {
    name: "Advance batch",
    exact: true,
  });
  await button.focus();
  await page.keyboard.press("Enter");
  await expect(page.locator("#core-table tbody tr").first()).toBeVisible();
}

test("core Go history, revisions, paging, focus and accessible narrow layout", async ({
  page,
}) => {
  await openLab(page);
  await expect(page.getByTestId("core-state")).toHaveText("paused");
  await advance(page);
  await page
    .getByText("Model assumptions, material and build identity", {
      exact: true,
    })
    .click();
  const coercivity = page
    .getByRole("row")
    .filter({
      has: page.getByRole("rowheader", {
        name: "Material / Parameters / Ec",
        exact: true,
      }),
    });
  await expect(coercivity).toContainText("V/m");
  await expect(coercivity).toContainText("megavolts per centimeter");
  await expect(page.locator("#core-table caption")).toContainText("Rows 1–32");
  await page.getByRole("button", { name: "Next core page" }).click();
  await expect(page.locator("#core-table caption")).toBeFocused();
  await expect(page.locator("#core-table caption")).toContainText("Rows 33");
  const cell = page.locator("#core-table tbody td[data-speech]").first();
  expect(await cell.getAttribute("aria-label")).toBe(
    await cell.getAttribute("data-speech"),
  );
  await page.getByLabel("Drive", { exact: true }).selectOption("manual");
  await expect(page.getByTestId("settings-state")).toContainText("stale");
  await apply(page);
  await page.getByLabel("Manual field (V/m)", { exact: true }).fill("-1000");
  await page.getByRole("button", { name: "Set field", exact: true }).click();
  await advance(page);
  await expect(page.locator("#core-table tbody")).toContainText("-");
  await page.getByLabel("Drive", { exact: true }).selectOption("triangle");
  await page
    .getByLabel("Model", { exact: true })
    .selectOption("landau-khalatnikov");
  await page.getByLabel("Material", { exact: true }).selectOption({ index: 0 });
  await apply(page);
  await advance(page);
  await page
    .getByLabel("Playback scale (ratio)", { exact: true })
    .fill("0.001");
  await apply(page);
  const beforePlayback = Number(
    await page.getByTestId("retained-count").textContent(),
  );
  await page.evaluate(() => {
    (window as any).announcements = 0;
    new MutationObserver(() => (window as any).announcements++).observe(
      document.querySelector(".status-line")!,
      { subtree: true, childList: true, characterData: true },
    );
  });
  await page.getByRole("button", { name: "Run", exact: true }).focus();
  await page.keyboard.press("Enter");
  await page.evaluate(
    () => new Promise((resolve) => setTimeout(resolve, 1200)),
  );
  await expect(
    page.getByRole("button", { name: "Run", exact: true }),
  ).toBeFocused();
  expect(
    await page.evaluate(() => (window as any).announcements),
  ).toBeLessThanOrEqual(4);
  expect(
    Number(await page.getByTestId("retained-count").textContent()),
  ).toBeLessThanOrEqual(beforePlayback + 64);
  await page.getByRole("button", { name: "Pause", exact: true }).click();
  await expect(page.getByTestId("core-state")).toHaveText("paused");
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await expect(page.getByTestId("core-state")).toHaveText("cancelled");
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await expect(page.locator("#core-table tbody tr")).toHaveCount(0);
  await page.getByRole("button", { name: "Read results", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "Loop results", exact: true }),
  ).toBeFocused();
  await page.setViewportSize({ width: 360, height: 740 });
  await page.evaluate(() => {
    document.documentElement.style.fontSize = "200%";
  });
  expect(
    await page.evaluate(
      () => document.documentElement.scrollWidth <= innerWidth + 2,
    ),
  ).toBe(true);
  const scan = await new AxeBuilder({ page }).analyze();
  expect(scan.violations).toEqual([]);
});

test("program modes, calibration invalidation and independent Go diagnostics", async ({
  page,
}) => {
  await openLab(page);
  await page.getByLabel("Drive", { exact: true }).selectOption("manual");
  await apply(page);
  await page.getByText("Programming and calibration", { exact: true }).click();
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("idle");
  await page.getByLabel("Target level", { exact: true }).fill("1");
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("complete", {
    timeout: 30_000,
  });
  await page
    .getByRole("button", { name: "Calibrate in memory", exact: true })
    .click();
  await expect(page.getByTestId("calibration-status")).toHaveText("fresh", {
    timeout: 30_000,
  });
  await page.getByLabel("Levels (levels)", { exact: true }).fill("4");
  await apply(page);
  await expect(page.getByTestId("calibration-status")).toHaveText("stale");
  await page
    .getByRole("button", { name: "Calibrate in memory", exact: true })
    .click();
  await expect(page.getByTestId("calibration-status")).toHaveText("fresh");
  await page
    .getByLabel("Programming mode", { exact: true })
    .selectOption("write-read");
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await page.getByLabel("Target level", { exact: true }).fill("2");
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await expect(page.getByTestId("program-status")).toHaveText("cancelled");
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await expect(page.getByTestId("program-status")).toHaveText("idle");
  await page
    .getByLabel("Programming mode", { exact: true })
    .selectOption("time-resolved");
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText(
    "legacy_protocol_unavailable",
  );
  await expect(
    page
      .getByText(/no switching protocol or numerical trace is implemented/)
      .first(),
  ).toBeVisible();
  await page
    .getByLabel("Model", { exact: true })
    .selectOption("landau-khalatnikov");
  await apply(page);
  await page
    .getByText("Diagnostics and reference context", { exact: true })
    .click();
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("complete");
  await expect(
    page.getByRole("button", { name: "Analyze", exact: true }),
  ).toBeFocused();
  await expect(page.getByTestId("analysis-context")).toContainText(
    "legacy-pund-demo",
  );
  await expect(page.locator("#analysis-table tbody tr")).not.toHaveCount(0);
  await expect(page.locator("#analysis-table thead")).not.toContainText(
    "Density",
  );
  await expect(page.locator("#analysis-table thead")).toContainText(
    "CurrentA (A)",
  );
  await expect(
    page
      .locator("#analysis-table tbody tr")
      .first()
      .locator("td[data-speech]")
      .last(),
  ).toHaveAttribute("aria-label", /amperes/);
  await page.getByLabel("Diagnostic", { exact: true }).selectOption("forc");
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("complete");
  await expect(page.getByTestId("analysis-context")).toContainText("gui-forc");
  await page
    .getByLabel("Analysis table", { exact: true })
    .selectOption("forc-density");
  await expect(page.locator("#analysis-table tbody tr")).toHaveCount(32);
  expect(
    await page.evaluate(
      () =>
        document.documentElement.scrollWidth <=
        document.documentElement.clientWidth + 2,
    ),
  ).toBe(true);
  await page.getByRole("button", { name: "Next analysis page" }).click();
  await expect(page.locator("#analysis-table caption")).toBeFocused();
  await expect(page.getByTestId("analysis-context")).toContainText(
    "not applied",
  );
  await page.getByLabel("EmaxVm", { exact: true }).fill("10");
  await expect(page.getByTestId("diagnostic-settings-state")).toContainText(
    "stale",
  );
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("complete");
  await expect(page.getByTestId("diagnostic-settings-state")).toHaveCount(0);
});

test("an earlier rejected batch cannot strand a newer valid Run", async ({
  page,
}) => {
  await page.addInitScript(() => {
    const NativeWorker = Worker;
    window.Worker = class extends NativeWorker {
      set onmessage(handler) {
        super.onmessage = (event) => {
          const reply =
            event.data.type === "reply"
              ? JSON.parse(event.data.json)
              : undefined;
          if (reply?.operation === "core.advance" && reply.error)
            setTimeout(() => handler?.call(this, event), 1200);
          else handler?.call(this, event);
        };
      }
      get onmessage() {
        return super.onmessage;
      }
    };
  });
  await openLab(page);
  await page.getByLabel("Steps per batch", { exact: true }).fill("0");
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await page.getByLabel("Steps per batch", { exact: true }).fill("64");
  await page.getByRole("button", { name: "Run", exact: true }).click();
  await expect
    .poll(async () =>
      Number(await page.getByTestId("retained-count").textContent()),
    )
    .toBeGreaterThan(0);
  await expect(
    page.getByRole("button", { name: "Run", exact: true }),
  ).toHaveAttribute("aria-disabled", "true");
  await expect(page.getByRole("alert")).toHaveCount(0);
  await page.getByRole("button", { name: "Pause", exact: true }).click();
  await expect(page.getByTestId("core-state")).toHaveText("paused");
});

test("reference read completion respects Clear and gates Analyze", async ({
  page,
}) => {
  await page.addInitScript(() => {
    const read = File.prototype.arrayBuffer;
    File.prototype.arrayBuffer = async function () {
      await new Promise((resolve) => setTimeout(resolve, 1200));
      if (this.name === "broken.csv") throw new Error("Reference read failed");
      return read.call(this);
    };
  });
  await openLab(page);
  await page
    .getByText("Diagnostics and reference context", { exact: true })
    .click();
  const input = page.getByLabel("Reference branch file (CSV or JSON)", {
    exact: false,
  });
  const csv = {
    name: "branch.csv",
    mimeType: "text/csv",
    buffer: Buffer.from("E_MV_cm,P_uC_cm2\n-1,-20\n1,20\n"),
  };
  await input.setInputFiles(csv);
  await expect(
    page.getByRole("button", { name: "Analyze", exact: true }),
  ).toHaveAttribute("aria-disabled", "true");
  await page
    .getByRole("button", { name: "Clear reference", exact: true })
    .click();
  await page.waitForTimeout(1500); // Wait past the deliberately delayed stale completion.
  await expect(page.getByText(/Reference ready in memory/)).toHaveCount(0);
  await input.setInputFiles({ ...csv, name: "broken.csv" });
  await expect(page.getByRole("alert")).toContainText("Reference read failed");
  await expect(
    page.getByRole("button", { name: "Analyze", exact: true }),
  ).toHaveAttribute("aria-disabled", "true");
  await input.setInputFiles(csv);
  await expect(page.getByText(/Reference ready in memory/)).toBeVisible();
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(
    page.getByRole("heading", { name: "branch.csv — imported, unverified" }),
  ).toBeVisible();
});

test("late Apply replies preserve newer input drafts", async ({ page }) => {
  await page.addInitScript(() => {
    const NativeWorker = Worker;
    window.Worker = class extends NativeWorker {
      set onmessage(handler) {
        super.onmessage = (event) => {
          const data = event.data;
          if (
            data.type === "reply" &&
            JSON.parse(data.json).operation === "core.apply"
          )
            setTimeout(() => handler?.call(this, event), 600);
          else handler?.call(this, event);
        };
      }
      get onmessage() {
        return super.onmessage;
      }
    };
  });
  await openLab(page);
  await page.getByLabel("Frequency (Hz)", { exact: true }).fill("1");
  await page
    .getByRole("button", { name: "Apply settings", exact: true })
    .click();
  await page.getByLabel("Frequency (Hz)", { exact: true }).fill("2");
  await expect(page.locator('.status-line[role="status"]')).toContainText(
    "Settings/state applied",
  );
  await expect(page.getByLabel("Frequency (Hz)", { exact: true })).toHaveValue(
    "2",
  );
  await expect(page.getByTestId("settings-state")).toContainText("stale");
});

test("unapplied loop settings cannot silently start jobs against old inputs", async ({
  page,
}) => {
  await openLab(page);
  await page.getByLabel("Frequency (Hz)", { exact: true }).fill("5");
  await page.getByText("Programming and calibration", { exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Initialize programmer", exact: true }),
  ).toBeDisabled();
  await page
    .getByText("Diagnostics and reference context", { exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Analyze", exact: true }),
  ).toHaveAttribute("aria-disabled", "true");
  await expect(page.getByTestId("analysis-status")).toHaveCount(0);
});

test("rejected queued target does not stop the active Go write/read pump", async ({
  page,
}) => {
  await openLab(page);
  await page.getByLabel("Drive", { exact: true }).selectOption("manual");
  await apply(page);
  await page.getByText("Programming and calibration", { exact: true }).click();
  await page
    .getByLabel("Programming mode", { exact: true })
    .selectOption("write-read");
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("idle");
  await page.getByLabel("Program steps per batch", { exact: true }).fill("1");
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("running");
  await page.getByLabel("Target level", { exact: true }).fill("0");
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText("target outside native");
  const revision = Number(await page.getByTestId("revision").textContent());
  await expect
    .poll(async () => Number(await page.getByTestId("revision").textContent()))
    .toBeGreaterThan(revision + 2);
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await expect(page.getByTestId("program-status")).toHaveText("cancelled");
});

test("Pause is explicitly unavailable for diagnostics without stranding cancellation", async ({
  page,
}) => {
  await openLab(page);
  await page
    .getByText("Diagnostics and reference context", { exact: true })
    .click();
  await page.getByLabel("Diagnostic", { exact: true }).selectOption("forc");
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await expect(
    page.getByRole("button", { name: "Pause", exact: true }),
  ).toHaveAttribute("aria-disabled", "true");
  await page.getByRole("button", { name: "Cancel", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("cancelled");
  await page.getByRole("button", { name: "Analyze", exact: true }).click();
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await expect(page.getByTestId("analysis-status")).toHaveText("cancelled");
  await expect(page.getByTestId("core-state")).toHaveText("paused");
});

test("invalid Go input is explicit and Worker fault can retry without losing focus", async ({
  page,
}) => {
  await openLab(page);
  await page.getByLabel("Steps per batch", { exact: true }).fill("0");
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText("Worker budget");
  await expect(page.getByTestId("core-state")).toHaveText("paused");
  await page.getByLabel("Steps per batch", { exact: true }).fill("64");
  await page.getByLabel("Frequency (Hz)", { exact: true }).fill("-1");
  await page
    .getByRole("button", { name: "Apply settings", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "frequency must be positive",
  );
  await page.getByLabel("Frequency (Hz)", { exact: true }).fill("1");
  await apply(page);
  await page.getByLabel("Drive", { exact: true }).selectOption("manual");
  await apply(page);
  await page.getByText("Programming and calibration", { exact: true }).click();
  await page.getByLabel("WorkLimit", { exact: true }).fill("1");
  await page
    .getByRole("button", { name: "Initialize programmer", exact: true })
    .click();
  await expect(page.getByTestId("program-status")).toHaveText("idle");
  await page
    .getByRole("button", { name: "Program target", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "solver step limit reached",
  );
  await expect(page.getByTestId("program-status")).toHaveText("work_limit");
  await page.getByRole("button", { name: "Reset", exact: true }).click();
  await page.route("**/hysteresis.wasm", (route) =>
    route.fulfill({ body: "corrupt module" }),
  );
  await page
    .getByRole("button", { name: "Discard and restart Worker", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText("asset hash mismatch");
  await page.unroute("**/hysteresis.wasm");
  await page.getByRole("button", { name: "Retry Worker", exact: true }).click();
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await advance(page);
  // Corrupt only transport metadata from a real Worker reply, never supply
  // replacement scientific results on a happy path.
  await page.addInitScript(() => {
    const NativeWorker = Worker;
    (window as any).corruptRevision = false;
    window.Worker = class extends NativeWorker {
      set onmessage(handler) {
        super.onmessage = (event) => {
          let data = event.data;
          if ((window as any).corruptRevision && data.type === "reply") {
            const reply = JSON.parse(data.json);
            if (reply.operation === "core.advance") {
              (window as any).corruptRevision = false;
              reply.revision--;
              data = { ...data, json: JSON.stringify(reply) };
            }
          }
          handler?.call(this, new MessageEvent("message", { data }));
        };
      }
      get onmessage() {
        return super.onmessage;
      }
    };
  });
  await page.reload();
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await page.evaluate(() => {
    (window as any).corruptRevision = true;
  });
  await page
    .getByRole("button", { name: "Advance batch", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText(
    "stale or mismatched Worker response",
  );
  await page.getByRole("button", { name: "Retry Worker", exact: true }).click();
  await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
  await advance(page);
});
