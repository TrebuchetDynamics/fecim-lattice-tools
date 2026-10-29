import { test, expect } from "@playwright/test";
import AxeBuilder from "@axe-core/playwright";
import { spawn, execFileSync } from "node:child_process";
import {
  mkdtempSync,
  rmSync,
  writeFileSync,
  readFileSync,
  existsSync,
} from "node:fs";
import { join } from "node:path";
const local = process.env.FECIM_WEB_PROFILE === "local";
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 14;
let child: ReturnType<typeof spawn> | undefined,
  root = "",
  launch = "";
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180000);
  root = mkdtempSync(join(process.env.TMPDIR!, "docs-local-"));
  writeFileSync(
    join(root, "lesson.md"),
    "# Granted lesson\n\nOnly the local corpus contains uniquegrantword.\n",
  );
  const binary = join(root, "fecim");
  execFileSync("node", ["src/local/build-native.mjs", binary], {
    timeout: 120000,
  });
  child = spawn(binary, ["serve", "--root", root, "--port", String(port)]);
  let output = "";
  child.stdout!.on("data", (d) => (output += d));
  child.stderr!.on("data", (d) => (output += d));
  await expect
    .poll(
      () =>
        output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)?.[0],
      { timeout: 30000 },
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

test("documentation reads before WASM, then searches the real Go index and navigates accessible content", async ({
  page,
}) => {
  const requests: { url: string; method: string }[] = [];
  page.on("request", (r) =>
    requests.push({ url: r.url(), method: r.method() }),
  );
  if (local) {
    await page.goto(launch);
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    await page
      .getByRole("link", { name: "Documentation", exact: true })
      .first()
      .click();
  } else await page.goto("/docs");
  await expect(
    page.getByRole("heading", { name: "Documentation", exact: true }),
  ).toBeVisible();
  await expect(
    page
      .getByLabel("Document reader")
      .getByRole("heading", {
        name: "Ferroelectric CIM: Complete ELI5 Guide - Overview",
      })
      .first(),
  ).toBeVisible();
  expect(requests.some((r) => r.url.endsWith("docs.wasm"))).toBe(false);
  const corpus = await (
    await page.request.get(new URL("/docs-manifest.json", page.url()).href)
  ).json();
  await expect(page.getByTestId("corpus-count")).toContainText(
    String(corpus.document_count),
  );
  await page.getByRole("button", { name: "Glossary", exact: true }).click();
  await expect(page.getByText("Recent (1)", { exact: true })).toBeVisible();
  await page.getByLabel("Search documentation").fill("zzzznotincorpus");
  await page.getByLabel("Search documentation").press("Enter");
  await expect(
    page.getByRole("status").filter({ hasText: "0 matches" }),
  ).toBeVisible();
  expect(requests.some((r) => r.url.endsWith("docs.wasm"))).toBe(true);
  await expect(page.getByText("Recent (1)", { exact: true })).toBeVisible();
  await page.getByLabel("Search documentation").fill("Materlik");
  await page.getByLabel("Category").selectOption("Citation");
  await page.getByRole("button", { name: "Search", exact: true }).click();
  await page
    .getByRole("button", { name: "Read Citation guide", exact: true })
    .click();
  await expect(
    page
      .getByLabel("Document reader")
      .getByRole("heading", { name: "Citation guide", exact: true }),
  ).toBeFocused();
  await expect(
    page.getByRole("link", { name: "DOI: 10.1063/1.4916229" }),
  ).toHaveAttribute("href", "https://doi.org/10.1063/1.4916229");
  await page.getByRole("button", { name: "Add favorite" }).click();
  await expect(
    page.getByRole("button", { name: "Remove favorite" }),
  ).toBeVisible();
  await page.getByRole("button", { name: "Glossary", exact: true }).click();
  await page
    .getByRole("navigation", { name: "On this page" })
    .getByRole("link", { name: "MVM", exact: true })
    .focus();
  await page.keyboard.press("Enter");
  await expect(
    page
      .getByLabel("Document reader")
      .getByRole("heading", { name: "MVM", exact: true }),
  ).toBeFocused();
  await page.goBack();
  await expect(
    page
      .getByLabel("Document reader")
      .getByRole("heading", { name: "FeCIM Technical Glossary", exact: true }),
  ).toBeFocused();
  await page
    .getByRole("button", { name: "Validation and evidence", exact: true })
    .click();
  await expect(page.getByTestId("evidence-status")).toHaveText(
    "not run/unverified",
  );
  await expect(
    page.getByLabel("Document reader").getByRole("table"),
  ).toBeVisible();
  await expect(page.getByLabel("Import evidence report")).toBeEnabled();
  await page.getByLabel("Import evidence report").setInputFiles({
    name: "report.json",
    mimeType: "application/json",
    buffer: Buffer.from(
      JSON.stringify({
        binding: {
          revision: "old",
          source_sha256: "a".repeat(64),
          fixture_sha256: "b".repeat(64),
          tool: "go1.26.10",
        },
        summary: "<script>not executed</script>",
      }),
    ),
  });
  await expect(page.getByTestId("evidence-status")).toHaveText(
    "imported/unverified",
  );
  await expect(
    page.getByText("<script>not executed</script>", { exact: true }),
  ).toBeVisible();
  expect(
    await page.evaluate(
      () =>
        document.querySelectorAll("article script, article iframe, article img")
          .length,
    ),
  ).toBe(0);
  await page.evaluate(() => {
    document.documentElement.style.fontSize = "200%";
  });
  await expect(
    page.getByLabel("Document reader").locator("pre code"),
  ).toBeVisible();
  expect(
    (await new AxeBuilder({ page }).include(".docs-screen").analyze())
      .violations,
  ).toEqual([]);
  expect(
    requests.filter(
      (r) =>
        !r.url.startsWith(new URL(page.url()).origin) &&
        !r.url.startsWith("blob:"),
    ),
  ).toEqual([]);
  if (local) {
    await page
      .getByRole("link", { name: "Projects", exact: true })
      .first()
      .click();
    await page
      .getByLabel("Project name", { exact: true })
      .fill("docs-history-project");
    await page
      .getByRole("button", { name: "Create project", exact: true })
      .click();
    await expect(
      page.getByRole("link", { name: "Configuration", exact: true }),
    ).toBeVisible();
    await page
      .getByRole("link", { name: "Documentation", exact: true })
      .first()
      .click();
    const panel = page.locator(".docs-local");
    await panel.locator("summary").click();
    await panel
      .getByRole("button", { name: "Connect local document grants" })
      .click();
    await panel
      .getByLabel("Markdown path within granted root")
      .fill("lesson.md");
    await panel.getByRole("button", { name: "Read granted Markdown" }).click();
    await expect(
      page
        .getByLabel("Document reader")
        .getByRole("heading", { name: "Granted lesson", exact: true }),
    ).toBeFocused();
    expect(
      existsSync(join(root, "docs-history-project", "docs-history.json")),
    ).toBe(false);
    await panel
      .getByLabel("Project path within granted root")
      .fill("docs-history-project");
    await panel.getByRole("checkbox").check();
    await expect(
      panel.getByRole("button", { name: "Save history" }),
    ).toBeEnabled();
    await page.getByRole("button", { name: "Add favorite" }).click();
    await panel.getByRole("button", { name: "Save history" }).click();
    await expect(
      page.getByRole("status").filter({ hasText: "History saved" }),
    ).toBeVisible();
    expect(
      readFileSync(
        join(root, "docs-history-project", "docs-history.json"),
        "utf8",
      ),
    ).toContain("local-lesson");
    expect(
      readFileSync(
        join(root, "docs-history-project", "docs-history.json"),
        "utf8",
      ),
    ).toContain("citations");
    const savedHistory = readFileSync(
      join(root, "docs-history-project", "docs-history.json"),
      "utf8",
    );
    await page.getByRole("button", { name: "Remove favorite" }).click();
    await expect(
      page.getByRole("button", { name: "Add favorite" }),
    ).toBeVisible();
    await panel.getByRole("button", { name: "Load history" }).click();
    await expect(
      page.getByRole("button", { name: "Remove favorite" }),
    ).toBeVisible();
    writeFileSync(
      join(root, "docs-history-project", "docs-history.json"),
      "{}",
    );
    await panel.getByRole("button", { name: "Load history" }).click();
    await expect(page.getByRole("alert")).toContainText(
      "Local documentation request rejected",
    );
    await expect(
      page
        .getByLabel("Document reader")
        .getByRole("heading", { name: "Granted lesson", exact: true }),
    ).toBeVisible();
    writeFileSync(
      join(root, "docs-history-project", "docs-history.json"),
      savedHistory,
    );
    await panel.getByRole("button", { name: "Load history" }).click();
    await expect(
      page.getByRole("status").filter({ hasText: "History loaded" }),
    ).toBeVisible();
    await panel.getByRole("checkbox").uncheck();
    await expect(
      panel.getByRole("button", { name: "Save history" }),
    ).toBeDisabled();
    // Temporary, explicitly granted documents exercise real Go paging without
    // adding filler pages to the approved publication corpus.
    for (let i = 0; i < 21; i++) {
      await panel
        .getByLabel("Stable local document ID")
        .fill("local-page-" + i);
      await panel
        .getByRole("button", { name: "Read granted Markdown" })
        .click();
      await expect(page.getByTestId("corpus-count")).toContainText(
        String(corpus.document_count + 2 + i),
      );
    }
    await page.getByLabel("Category").selectOption("");
    await page.getByLabel("Search documentation").fill("uniquegrantword");
    await page.getByRole("button", { name: "Search", exact: true }).click();
    const results = page.getByRole("region", { name: "Search results" });
    await expect(results.locator("ol > li")).toHaveCount(20);
    await page.getByRole("button", { name: "Next results" }).click();
    await expect(results.locator("ol > li")).toHaveCount(2);
    await page.getByRole("button", { name: "Previous results" }).click();
    await expect(results.locator("ol > li")).toHaveCount(20);
  } else {
    expect(
      requests.filter(
        (r) =>
          r.method !== "GET" || new URL(r.url).pathname.startsWith("/api/"),
      ),
    ).toEqual([]);
    expect(
      await page.evaluate(() => [localStorage.length, sessionStorage.length]),
    ).toEqual([0, 0]);
  }
  await page.evaluate(() => {
    history.pushState(null, "", "#doc-missing:section");
    window.dispatchEvent(new PopStateEvent("popstate"));
  });
  await expect(page.getByRole("alert")).toContainText("Unknown document ID");
  await expect(page.getByLabel("Document reader")).toHaveCount(0);
});
