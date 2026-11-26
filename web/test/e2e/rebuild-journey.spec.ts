import { expect, test } from "@playwright/test";
import { useResearchServer } from "./research-server";

const local = process.env.FECIM_WEB_PROFILE === "local";
const profile = local ? "local" : "hosted";
const server = useResearchServer();

const labs = [
  { id: "hysteresis", link: "Hysteresis Lab", heading: "Hysteresis lab", region: "Loop result" },
  { id: "crossbar", link: "Crossbar Array", heading: "Crossbar", region: "MVM output" },
  { id: "circuits", link: "Peripheral Circuits", heading: "Circuits", region: "Cell operations" },
  { id: "comparison", link: "Technology Comparison", heading: "Comparison", region: "Metric detail" },
  { id: "mnist", link: "MNIST Pipeline", heading: "MNIST", region: local ? "Run inference" : "Learn and import" },
  { id: "eda", link: "EDA Export", heading: "EDA", region: local ? "Generation" : "Hosted read-only inspection" },
  { id: "docs", link: "Integrated Docs", heading: "Documentation", region: "Browse and search" },
] as const;

test(`${profile}: one selected study run survives all seven lab round trips`, async ({ page, context, baseURL }, info) => {
  const origin = new URL(local ? server.launch : baseURL!).origin;
  const browserErrors: string[] = [];
  const hostedViolations: string[] = [];
  page.on("pageerror", error => browserErrors.push(error.message));
  if (!local) {
    context.on("request", request => {
      const url = new URL(request.url());
      if (url.origin !== origin || request.method() !== "GET" || url.pathname.startsWith("/api/")) {
        hostedViolations.push(`${request.method()} ${url.href}`);
      }
    });
  }

  const main = page.getByRole("main");
  const primary = page.getByRole("navigation", { name: "Primary navigation" });
  const heading = main.getByRole("heading", { level: 1 });
  const link = (name: string) => main.getByRole("link", { name, exact: true });
  const button = (name: string) => main.getByRole("button", { name, exact: true });
  const region = (name: string) => main.getByRole("region", { name, exact: true });
  const openedStudy = main.getByText(/^Opened study:/);
  const selectedIds = region("Selected runs").locator("code");
  const checkRun = (id: string) => main.getByRole("checkbox", { name: `Select run ${id}`, exact: true });
  const expectRoute = async (path: string, title: string) => {
    await expect(page, `${profile} ${path}: URL`).toHaveURL(origin + path);
    await expect(heading, `${profile} ${path}: active heading`).toHaveText(title);
    await expect(heading, `${profile} ${path}: heading focus`).toBeFocused();
  };

  await test.step(`${profile} /results → /projects: recover from no open study`, async () => {
    if (local) {
      await page.goto(server.launch);
      await expect(page).toHaveURL(origin + "/hysteresis");
    }
    await page.goto(origin + "/results");
    await expectRoute("/results", "Results");
    await expect(main.getByText("No study is open.", { exact: true })).toBeVisible();
    await link("Go to Projects").press("Enter");
    await expectRoute("/projects", "Projects");
    if (local) {
      await main.getByLabel("Project name", { exact: true }).fill("rebuild-journey");
      await button("Create project").click();
    } else {
      await button("Open small example").click();
    }
  });

  const { id, study } = await test.step(`${profile} /configuration → /experiment → /results: run study and select an ID`, async () => {
    await link("Continue to Configuration").click();
    await expectRoute("/configuration", "Configuration");
    const study = await openedStudy.innerText();
    await link("Continue to Experiment").click();
    await expectRoute("/experiment", "Experiment");
    await button("Run study").click();
    await expect(main.getByTestId("experiment-state")).toContainText(/Execution succeeded|finished/, { timeout: 60_000 });
    await link("View study results").click();
    await expectRoute("/results", "Results");
    const id = await main.locator("#runs tbody th button").first().innerText();
    await checkRun(id).check();
    await expect(selectedIds).toHaveText([id]);
    return { id, study };
  });

  for (const lab of labs) {
    const path = `/modules/${lab.id}`;
    await test.step(`${profile} ${path} → /results: preserve ${id}`, async () => {
      await primary.getByRole("link", { name: "Modules", exact: true }).click();
      await expect(page).toHaveURL(origin + "/#modules");
      await main.getByRole("navigation", { name: "Explore a lab" })
        .getByRole("link", { name: `Explore ${lab.link}`, exact: true }).click();
      await expect(region(lab.region), `${profile} ${path}: ${lab.region} region`).toBeVisible();
      await expectRoute(path, lab.heading);

      if (lab.id === "mnist") {
        if (local) {
          const inference = region("Run inference");
          await expect(inference.getByRole("button", { name: "Reconnect authoritative state" })).toBeEnabled();
          await expect(inference.getByRole("button", { name: "Run paired inference" })).toBeDisabled();
          await expect(inference).toContainText("Load a model under");
        } else {
          const learn = region("Learn and import");
          await expect(learn).toContainText("read-only result inspection");
          await expect(learn.getByLabel("Import result for read-only inspection")).toBeVisible();
          await expect(region("Run inference")).toHaveCount(0);
          await expect(button("Run paired inference")).toHaveCount(0);
        }
      } else if (lab.id === "eda") {
        if (local) {
          const generation = region("Generation");
          await expect(generation.getByRole("button", { name: "Generate all formats" })).toBeEnabled();
          await expect(generation.getByRole("button", { name: "Export verified selected run" })).toBeEnabled();
          await expect(region("Source design").getByLabel("Stored run ID")).toHaveAttribute("placeholder", id);
        } else {
          const inspection = region("Hosted read-only inspection");
          await expect(inspection).toContainText("require the authenticated local application");
          await expect(inspection.getByLabel("Import artifact for read-only inspection")).toBeVisible();
          await expect(region("Generation")).toHaveCount(0);
          await expect(button("Generate all formats")).toHaveCount(0);
        }
      }

      // Both Back entries were created by actual app links, not deep-link reloads.
      await page.goBack();
      await expect(page).toHaveURL(origin + "/#modules");
      await page.goBack();
      await expectRoute("/results", "Results");
      await expect(openedStudy, `${profile} ${path}: opened study after return`).toHaveText(study);
      await expect(selectedIds, `${profile} ${path}: selected ID after return`).toHaveText([id]);
      await expect(checkRun(id), `${profile} ${path}: checked result after return`).toBeChecked();
      expect(browserErrors, `${profile} ${path}: uncaught browser errors`).toEqual([]);
      expect(hostedViolations, `${profile} ${path}: hosted sent a non-static or off-origin request`).toEqual([]);
    });
  }

  await page.screenshot({ path: info.outputPath(`${profile}-returned-results.png`), fullPage: true });
});
