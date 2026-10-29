import { expect, test } from "@playwright/test";
import { spawn, execFileSync } from "node:child_process";
import { mkdtempSync, copyFileSync, writeFileSync, rmSync } from "node:fs";
import { join } from "node:path";
const local = process.env.FECIM_WEB_PROFILE === "local";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 16;
let root = "",
  launch = "",
  expected: any;
let child: ReturnType<typeof spawn> | undefined;
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "mnist-native-"));
  for (const name of [
    "pretrained_weights.json",
    "t10k-images-idx3-ubyte.gz",
    "t10k-labels-idx1-ubyte.gz",
  ])
    copyFileSync(join("../module3-mnist/data", name), join(root, name));
  writeFileSync(join(root, "bad.json"), '{"layer1_weights":[[1]]}');
  const oracle = join(root, "oracle.go");
  writeFileSync(
    oracle,
    `package main
import("encoding/json";"os";"path/filepath";"fecim-lattice-tools/shared/neural";"fecim-lattice-tools/module3-mnist/pkg/mnist")
func main(){n:=neural.NewDualModeNetwork(784,128,10);if e:=n.LoadWeights(filepath.Join(os.Args[1],"pretrained_weights.json"));e!=nil{panic(e)};n.SetNoiseLevel(0);im,la,e:=mnist.LoadMNIST(os.Args[1],false);if e!=nil{panic(e)};json.NewEncoder(os.Stdout).Encode(map[string]any{"inference":n.Infer(im[17]),"metrics":neural.EvaluateDualModeDataset(n,im[17:20],la[17:20])})}`,
  );
  expected = JSON.parse(
    execFileSync("go", ["run", oracle, root], {
      encoding: "utf8",
      timeout: 120_000,
    })
      .trim()
      .split("\n")
      .at(-1)!,
  );
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

test("MNIST real local science and hosted read-only boundary", async ({
  page,
}) => {
  const requests: string[] = [];
  let mnistSessions = 0;
  page.on("request", (r) => {
    if (r.url().includes("/api/")) requests.push(r.url());
    if (
      r.method() === "POST" &&
      r.url().endsWith("/api/v1/modules/mnist/sessions")
    )
      mnistSessions++;
  });
  if (local) {
    await page.goto(launch);
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    await page
      .getByRole("link", { name: /FeCIM Lattice Tools/ })
      .first()
      .click();
    await page
      .getByRole("link", { name: "Explore MNIST Pipeline", exact: true })
      .click();
  } else await page.goto("/modules/mnist");
  await expect(
    page.getByRole("heading", { name: "MNIST", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("heading", { name: "Model, input, evidence", exact: true }),
  ).toBeVisible();
  if (!local) {
    await expect(
      page.getByRole("button", { name: "Run paired inference", exact: true }),
    ).toHaveCount(0);
    await page
      .getByLabel("Import result for read-only inspection", { exact: true })
      .setInputFiles({
        name: "result.json",
        mimeType: "application/json",
        buffer: Buffer.from("<script>window.mnistInjected=true</script>"),
      });
    await expect(page.getByTestId("mnist-import")).toContainText(
      "<script>window.mnistInjected=true</script>",
    );
    await expect(
      page.getByText("Imported/unverified — not a locally verified evaluation"),
    ).toBeVisible();
    expect(await page.evaluate(() => "mnistInjected" in window)).toBe(false);
    expect(requests).toEqual([]);
    return;
  }
  await expect(page.getByTestId("mnist-status")).toContainText(
    "unavailable: select a model",
  );
  await page
    .getByLabel("Model path", { exact: true })
    .fill("pretrained_weights.json");
  await page
    .getByRole("button", { name: "Load selected model", exact: true })
    .focus();
  await page.keyboard.press("Enter");
  await expect(page.getByTestId("mnist-identity")).toContainText(
    "Hidden neurons: 128",
  );
  await expect(
    page.getByRole("button", { name: "Load selected model", exact: true }),
  ).toBeEnabled();
  await page.getByLabel("Noise coefficient", { exact: true }).fill("0");
  await page
    .getByRole("button", { name: "Apply configuration", exact: true })
    .click();
  await expect(
    page.getByRole("button", { name: "Apply configuration", exact: true }),
  ).toBeEnabled();
  // Keyboard pixel editing is an accessible drawing alternative, not TS preprocessing.
  await page.getByLabel("Pixel row", { exact: true }).fill("2");
  await page.getByLabel("Pixel column", { exact: true }).fill("2");
  await page.getByRole("button", { name: "Set pixel", exact: true }).click();
  await page.getByRole("button", { name: "Use drawing", exact: true }).click();
  await expect(page.getByTestId("mnist-identity")).toContainText(
    "Current input source: drawing",
  );
  await expect(
    page.getByRole("button", { name: "Use drawing", exact: true }),
  ).toBeEnabled();
  await page
    .getByRole("button", { name: "Run paired inference", exact: true })
    .click();
  await expect(page.getByTestId("mnist-status")).toContainText("succeeded");
  await expect(page.getByTestId("mnist-evidence")).toContainText(
    "Accuracy unavailable",
  );
  await expect(page.locator("#mnist-probabilities tbody tr")).toHaveCount(10);
  await expect(page.locator("#mnist-layers tbody tr")).toHaveCount(128);
  await page
    .getByRole("button", { name: "Load labelled corpus", exact: true })
    .click();
  await expect(page.getByTestId("mnist-identity")).toContainText(
    "Actual selected corpus images: 10000",
  );
  await expect(
    page.getByRole("button", { name: "Load labelled corpus", exact: true }),
  ).toBeEnabled();
  // Terminal job publication precedes the refreshed snapshot. Keep the real
  // response held until a read has captured the previous retained result.
  const resultAfterRefresh = async (
    button: string,
    matches: (result: any) => void,
  ) => {
    const result = page.getByTestId("mnist-result");
    const previous = await result.textContent();
    let release!: () => void;
    const gate = new Promise<void>((resolve) => (release = resolve));
    let safetyReleased = false;
    const safety = setTimeout(() => {
      safetyReleased = true;
      release();
    }, 15_000);
    let held = false;
    await page.route(
      "**/api/v1/module-sessions/*",
      async (route) => {
        const response = await route.fetch({ timeout: 10_000 });
        held = true;
        await gate;
        await route.fulfill({ response });
      },
      { times: 1 },
    );
    try {
      await page.getByRole("button", { name: button, exact: true }).click();
      await expect.poll(() => held).toBe(true);
      await expect(page.getByTestId("mnist-status")).toContainText("succeeded");
      await expect(result).toHaveText(previous!);
      expect(safetyReleased).toBe(false);
      const read = async () => {
        const next = JSON.parse((await result.textContent())!);
        release();
        matches(next);
        return next;
      };
      let next: any;
      await expect(async () => {
        next = await read();
      }).toPass({ timeout: 10_000 });
      return next;
    } finally {
      release();
      clearTimeout(safety);
      await page.unrouteAll({ behavior: "wait" });
    }
  };
  await page.getByLabel("Dataset sample index", { exact: true }).fill("17");
  await page
    .getByRole("button", { name: "Use dataset sample", exact: true })
    .click();
  await expect(page.getByTestId("mnist-identity")).toContainText(
    "Current input source: dataset",
  );
  await expect(
    page.getByRole("button", { name: "Use dataset sample", exact: true }),
  ).toBeEnabled();
  let r = await resultAfterRefresh("Run paired inference", (next) => {
    expect(next).toMatchObject({
      source: "dataset",
      start: 17,
      count: 1,
      preprocessing: "none",
    });
    expect(next.inference).toEqual(expected.inference);
  });
  expect(r.preprocessing).toBe("none");
  expect(r.model_sha256).toMatch(/^[a-f0-9]{64}$/);
  expect(r.dataset_sha256).toMatch(/^[a-f0-9]{64}$/);
  await page.getByLabel("Subset start", { exact: true }).fill("17");
  await page.getByLabel("Selected images", { exact: true }).fill("3");
  r = await resultAfterRefresh("Evaluate labelled subset", (next) => {
    expect(next).toMatchObject({
      source: "dataset",
      start: 17,
      count: 3,
      preprocessing: "none",
    });
    expect(next.cases[0].total).toBe(3);
    expect(next.cases[0].cim_confusion).toEqual(expected.metrics.CIM.Confusion);
    expect(next.cases[0].fp_confusion).toEqual(expected.metrics.FP.Confusion);
  });
  await expect(page.locator("#mnist-confusion tbody tr")).toHaveCount(100);
  await page
    .getByRole("button", { name: "Next weight page", exact: true })
    .click();
  await expect(page.locator("#mnist-weights caption")).toContainText("page 2");
  await expect(page.locator("#mnist-weights caption")).toBeFocused();
  const download = page.waitForEvent("download");
  await page
    .getByRole("button", { name: "Export complete result JSON", exact: true })
    .click();
  expect((await download).suggestedFilename()).toBe("mnist-result.json");
  await page.getByLabel("Model path", { exact: true }).fill("bad.json");
  await page
    .getByRole("button", { name: "Load selected model", exact: true })
    .click();
  await expect(page.getByRole("alert")).toBeFocused();
  expect(
    JSON.parse((await page.getByTestId("mnist-result").textContent())!),
  ).toEqual(r);
  // Keep the job identity and cancellation usable when one poll is lost.
  await page.getByLabel("Subset start", { exact: true }).fill("0");
  await page.getByLabel("Selected images", { exact: true }).fill("10000");
  await page.route(
    "**/api/v1/module-sessions/*/jobs/*",
    (route) => route.abort(),
    { times: 1 },
  );
  await page
    .getByRole("button", { name: "Evaluate labelled subset", exact: true })
    .click();
  await expect(page.getByRole("alert")).toBeVisible();
  // Reload must recover the same native session, not orphan its active work.
  await page.reload();
  await expect(
    page.getByRole("button", { name: "Cancel MNIST job", exact: true }),
  ).toBeEnabled();
  expect(mnistSessions).toBe(1);
  expect(
    JSON.parse((await page.getByTestId("mnist-result").textContent())!),
  ).toEqual(r);
  await page
    .getByRole("button", { name: "Cancel MNIST job", exact: true })
    .click();
  await page
    .getByRole("button", { name: "Reconnect authoritative state", exact: true })
    .click();
  await expect(page.getByTestId("mnist-status")).toContainText("cancelled");
  expect(
    JSON.parse((await page.getByTestId("mnist-result").textContent())!),
  ).toEqual(r);
  await page.getByLabel("Subset start", { exact: true }).fill("0");
  await page.getByLabel("Selected images", { exact: true }).fill("10000");
  // Lose only the next start receipt after an earlier terminal job exists.
  // The server has accepted work; reconnect must prefer its new job identity.
  await page.route(
    "**/api/v1/module-sessions/*/jobs",
    async (route) => {
      await route.fetch();
      await route.abort();
    },
    { times: 1 },
  );
  await page
    .getByRole("button", { name: "Evaluate labelled subset", exact: true })
    .click();
  await expect(page.getByRole("alert")).toContainText("outcome unknown");
  await page
    .getByRole("button", { name: "Cancel MNIST job", exact: true })
    .click();
  await expect(page.getByRole("alert")).toHaveCount(0);
  await expect(page.getByTestId("mnist-status")).toContainText("cancelled");
  expect(
    JSON.parse((await page.getByTestId("mnist-result").textContent())!),
  ).toEqual(r);
});
