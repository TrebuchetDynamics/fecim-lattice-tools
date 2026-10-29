import { expect, test } from "@playwright/test";
import { spawn, execFileSync } from "node:child_process";
import { mkdtempSync, rmSync, readFileSync, writeFileSync } from "node:fs";
import { join } from "node:path";

const local = process.env.FECIM_WEB_PROFILE === "local";
let child: ReturnType<typeof spawn> | undefined;
let root = "", launch = "", projectNumber = 0;
let currentProject = "";
let cookies: Awaited<ReturnType<import("@playwright/test").BrowserContext["cookies"]>> = [];
const port = Number(process.env.AUTOWORK_PORT_BASE || 20200) + 12;
test.beforeAll(async () => {
  if (!local) return;
  test.setTimeout(180_000);
  root = mkdtempSync(join(process.env.TMPDIR!, "research-local-"));
  const binary = join(root, "fecim");
  execFileSync("node", ["src/local/build-native.mjs", binary], { timeout: 120_000 });
  child = spawn(binary, ["serve", "--root", root, "--port", String(port)]);
  let output = "";
  child.stdout!.on("data", data => { output += data; });
  child.stderr!.on("data", data => { output += data; });
  await expect.poll(() => output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)?.[0], { timeout: 30_000 }).toBeTruthy();
  launch = output.match(/http:\/\/127\.0\.0\.1:\d+\/#bootstrap=[a-f0-9]+/)![0];
});
test.afterAll(async () => {
  if (child && child.exitCode === null) {
    const exited = new Promise(resolve => child!.once("exit", resolve));
    child.kill("SIGTERM"); await exited;
  }
  if (root) rmSync(root, { recursive: true, force: true });
});
async function projects(page: import("@playwright/test").Page) {
  if (local) {
    if (cookies.length) { await page.context().addCookies(cookies); await page.goto(`http://127.0.0.1:${port}/hysteresis`); }
    else await page.goto(launch);
    await expect(page.getByLabel("Material", { exact: true })).toBeEnabled();
    cookies = await page.context().cookies();
  }
  else await page.goto("/");
  await page.getByRole("link", { name: "Projects", exact: true }).first().click();
  await expect(page.getByRole("heading", { name: "Projects", exact: true })).toBeFocused();
  if (local) {
    currentProject = `research-${++projectNumber}`;
    await page.getByLabel("Project name", { exact: true }).fill(currentProject);
    await page.getByRole("button", { name: "Create project", exact: true }).click();
  } else await page.getByRole("button", { name: "Open small example", exact: true }).click();
  await page.getByRole("link", { name: "Configuration", exact: true }).click();
}
test("four-view real Go study, fields, validation, stable selection and comparison", async ({ page }) => {
  await projects(page);
  const rows = page.getByLabel("Array rows", { exact: true });
  await expect(rows).toBeVisible();
  if (local) {
    await rows.fill("not-a-number");
    await page.getByRole("button", { name: "Save design.yaml", exact: true }).click();
    await expect(page.getByRole("alert")).toBeFocused();
    await expect(rows).toHaveValue("not-a-number");
    await expect(page.getByRole("alert")).toContainText("1 error");
    await page.getByRole("alert").getByRole("link", {name:"Array rows",exact:true}).click();
    await expect(rows).toBeFocused();
    await rows.fill("16");
    await page.getByRole("button", { name: "Save design.yaml", exact: true }).click();
    await expect(page.getByTestId("research-status")).toContainText("Saved");
  }
  await page.getByRole("button", { name: "Validate configuration", exact: true }).click();
  await expect(page.getByTestId("validation")).toContainText("valid");
  await page.getByRole("link", { name: "Experiment", exact: true }).click();
  await page.getByRole("button", { name: "Run study", exact: true }).click();
  await expect(page.getByTestId("experiment-state")).toContainText(/Execution succeeded|finished/, { timeout: 60_000 });
  await page.getByRole("link", { name: "Results", exact: true }).click();
  await expect(page.locator("#runs tbody tr")).toHaveCount(8);
  await expect(page.getByRole("figure", {name:"Result plot"})).toBeVisible();
  await expect(page.locator("#results-plot svg")).toHaveAttribute("viewBox", "0 0 1000 600");
  await expect(page.locator("#results-plot circle")).toHaveCount(8);
  await expect(page.locator("#results-plot [tabindex], #results-plot button, #results-plot a")).toHaveCount(0);
  await page.locator("#runs input[type=checkbox]").nth(0).check();
  await page.locator("#runs input[type=checkbox]").nth(1).check();
  const ids = await page.locator("#selected-runs li").allTextContents();
  await expect(page.locator('#results-plot circle[data-selected="true"]')).toHaveCount(2);
  await page.locator("#runs tbody th button").first().click();
  await expect(page.getByRole("heading", {name:"Run details and provenance",exact:true})).toBeVisible();
  await page.getByRole("button", {name:"Close details",exact:true}).click();
  await page.getByLabel("Sort runs", { exact: true }).selectOption("energy_pj");
  await page.getByRole("button", { name: "Apply query", exact: true }).click();
  await expect(page.locator("#runs caption")).toBeFocused();
  await expect(page.locator("#selected-runs li")).toHaveText(ids);
  await page.getByRole("button", { name: "Compare metrics", exact: true }).click();
  await expect(page.getByRole("heading", { name: "Compare metrics", exact: true })).toBeFocused();
  await expect(page.locator("#comparison tbody tr")).toHaveCount(4);
  await expect(page.locator("#comparison")).toContainText("picojoules");
  if (local) {
    await page.getByRole("button", { name: "Verify stored runs", exact: true }).click();
    await expect(page.getByTestId("result-action")).toContainText("Verified");
    await page.getByRole("button", { name: "Generate reports", exact: true }).click();
    await expect(page.getByRole("link", { name: "Download results.json", exact: true })).toBeVisible();
    await page.locator("#selected-runs button").last().click();
    await page.getByRole("button",{name:"Export selected run",exact:true}).click();
    await expect(page.getByTestId("result-action")).toContainText("Export generated");
    const download=page.getByRole("link",{name:/Download .*\.spice$/}).first();
    await expect(download).toBeVisible();
    const downloaded=page.waitForEvent("download");await download.click();expect((await downloaded).suggestedFilename()).toMatch(/\.spice$/);
    await page.getByRole("link",{name:"Experiment",exact:true}).click();
    await page.getByRole("button",{name:"Reconnect authoritative state",exact:true}).click();
    await page.getByRole("button",{name:"Run study",exact:true}).click();
    await expect(page.getByTestId("experiment-state")).toContainText(/Execution succeeded|finished/,{timeout:60_000});
    await expect(page.getByTestId("experiment-state")).toContainText("8 reused");
    await page.getByRole("link",{name:"Projects",exact:true}).first().click();
    await page.getByRole("button",{name:"Close project",exact:true}).click();
    await page.getByLabel("Relative project path",{exact:true}).fill(currentProject);
    await page.getByRole("button",{name:"Open project",exact:true}).click();
    await expect(page.getByTestId("research-status")).toContainText("Project opened");
  }
});
test("cancel is authoritative, navigation keeps the owner and previous results", async ({ page }) => {
  await projects(page);
  await page.getByRole("link", { name: "Experiment", exact: true }).click();
  await page.getByRole("button", { name: "Run study", exact: true }).click();
  const cancel=page.getByRole("button",{name:"Cancel experiment",exact:true});
  await expect(cancel).toBeEnabled();
  const cancellation=local?page.waitForResponse(r=>r.url().endsWith("/cancel")&&r.request().method()==="POST"):undefined;
  await cancel.click();if(cancellation)expect((await cancellation).ok()).toBeTruthy();
  await page.getByRole("link", { name: "Documentation", exact: true }).first().click();
  await expect(page.getByRole("heading", { name: "Documentation", exact: true })).toBeFocused();
  await page.goBack();
  await expect(page.getByRole("heading", { name: "Experiment", exact: true })).toBeFocused();
  await expect(page.getByTestId("experiment-state")).toContainText(/succeeded|cancelled|Cancellation settled|Cancellation requested/, { timeout: 60_000 });
});

test("selection survives pages and refuses a fifth without moving focus or losing four", async ({page}) => {
  await projects(page);
  await page.getByRole("link",{name:"Experiment",exact:true}).click();
  await page.getByRole("button",{name:"Run study",exact:true}).click();
  await expect(page.getByTestId("experiment-state")).toContainText(/Execution succeeded|finished/,{timeout:60_000});
  await page.getByRole("link",{name:"Results",exact:true}).click();
  await expect(page.locator("#runs tbody tr")).toHaveCount(8);
  const boxes=page.locator("#runs input[type=checkbox]");
  for(let i=0;i<4;i++)await boxes.nth(i).check();
  await boxes.nth(4).click();
  await expect(boxes.nth(4)).not.toBeChecked();
  await expect(boxes.nth(4)).toBeFocused();
  await expect(page.locator("#selected-runs li")).toHaveCount(4);
  await expect(page.getByTestId("selection-status")).toContainText("fifth");
  await page.getByLabel("Page size",{exact:true}).selectOption("2");
  await page.getByRole("button",{name:"Apply query",exact:true}).click();
  await expect(page.locator("#runs tbody tr")).toHaveCount(2);
  await expect(page.locator("#runs caption")).toBeFocused();
  await page.getByRole("button",{name:"Next runs",exact:true}).click();
  await expect(page.locator("#runs caption")).toContainText("page 2");
  await expect(page.locator("#runs caption")).toBeFocused();
  await expect(page.locator("#selected-runs li")).toHaveCount(4);
  await page.getByRole("button",{name:"Compare metrics",exact:true}).click();
  await expect(page.locator("#comparison thead th")).toHaveCount(5);
});

test("native stale field save retains edits and explicitly reviews current disk text", async ({page}) => {
  test.skip(!local,"native filesystem conflict");
  await projects(page);
  const file=join(root,currentProject,"design.yaml");
  const before=readFileSync(file,"utf8");
  await page.getByLabel("Array rows",{exact:true}).fill("16");
  writeFileSync(file,before.replace("rows: 32","rows: 24"));
  await page.getByRole("button",{name:"Save design.yaml",exact:true}).click();
  await expect(page.getByRole("heading",{name:"Configuration conflict",exact:true})).toBeVisible();
  await expect(page.getByRole("alert")).toBeFocused();
  await expect(page.getByRole("alert")).toContainText("on disk");
  await expect(page.getByLabel("Array rows",{exact:true})).toHaveValue("16");
  expect(readFileSync(file,"utf8")).toContain("rows: 24");
  await page.getByRole("button",{name:"Reapply reviewed edits",exact:true}).click();
  await expect(page.getByTestId("research-status")).toContainText("Saved");
  expect(readFileSync(file,"utf8")).toContain("rows: 16");
});


test("late Go rejection cannot steal focus from a newer research view", async ({page}) => {
  test.skip(!local,"native response delay");
  await projects(page);
  await page.getByRole("link",{name:"Experiment",exact:true}).click();
  await page.getByRole("button",{name:"Run study",exact:true}).click();
  await expect(page.getByTestId("experiment-state")).toContainText(/Execution succeeded|finished/,{timeout:60_000});
  await page.getByRole("link",{name:"Results",exact:true}).click();
  await expect(page.locator("#runs tbody tr")).toHaveCount(8);
  await page.locator("#runs input[type=checkbox]").first().check();
  let release!:()=>void,received!:()=>void;
  const gate=new Promise<void>(r=>release=r),arrived=new Promise<void>(r=>received=r);
  await page.route("**/analysis?*",async route=>{if(!new URL(route.request().url()).searchParams.has("compare"))return route.continue();const response=await route.fetch();expect(response.status()).toBe(400);received();await gate;await route.fulfill({response});});
  await page.getByRole("button",{name:"Compare metrics",exact:true}).click();await arrived;
  await page.getByRole("link",{name:"Configuration",exact:true}).click();
  await expect(page.getByRole("heading",{name:"Configuration",exact:true})).toBeFocused();
  const delivered=page.waitForResponse(r=>r.url().includes("compare="));release();await delivered;
  await expect(page.getByRole("alert")).toHaveCount(0);
  await expect(page.getByRole("heading",{name:"Configuration",exact:true})).toBeFocused();
});

test("pending native saves preserve newer drafts and cannot restore a closed project", async ({page}) => {
  test.skip(!local,"native save responses");await projects(page);
  const rows=page.getByLabel("Array rows",{exact:true});
  for(const changeProject of [false,true]) {
    let release!:()=>void,received!:()=>void;
    const gate=new Promise<void>(r=>release=r),arrived=new Promise<void>(r=>received=r);
    await page.route("**/configuration",async route=>{if(route.request().method()!=="PUT")return route.continue();const response=await route.fetch();expect(response.ok()).toBeTruthy();received();await gate;await route.fulfill({response});});
    await rows.fill(changeProject?"24":"16");await page.getByRole("button",{name:"Save design.yaml",exact:true}).click();await arrived;
    if(changeProject){await page.getByRole("link",{name:"Projects",exact:true}).first().click();await page.getByRole("button",{name:"Close project",exact:true}).click();}
    else await rows.fill("20");
    const delivered=page.waitForResponse(r=>r.request().method()==="PUT");release();await delivered;
    if(changeProject){await page.getByRole("link",{name:"Configuration",exact:true}).click();await expect(rows).toHaveCount(0);}
    else await expect(rows).toHaveValue("20");
    await page.unroute("**/configuration");
  }
});
