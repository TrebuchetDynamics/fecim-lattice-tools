import { expect, test } from "@playwright/test";

const DESKTOP_PROJECTS = new Set([
  "desktop-chromium",
  "desktop-compact-chromium",
]);
const MOBILE_PROJECTS = new Set(["mobile-chromium", "mobile-small-chromium"]);

function isDesktopProject(projectName) {
  return DESKTOP_PROJECTS.has(projectName);
}

function isMobileProject(projectName) {
  return MOBILE_PROJECTS.has(projectName);
}

async function viewportMetrics(page) {
  return page.evaluate(() => ({
    innerWidth: window.innerWidth,
    innerHeight: window.innerHeight,
    scrollWidth: document.documentElement.scrollWidth,
    scrollHeight: document.documentElement.scrollHeight,
  }));
}

async function expectNoHorizontalOverflow(page) {
  const metrics = await viewportMetrics(page);
  expect(metrics.scrollWidth - metrics.innerWidth).toBeLessThanOrEqual(2);
  return metrics;
}

async function expectLocatorInsideViewport(page, locator, label) {
  const box = await locator.boundingBox();
  expect(box, `${label} should render`).not.toBeNull();
  const metrics = await viewportMetrics(page);
  expect(
    box.x,
    `${label} should not start off the left edge`,
  ).toBeGreaterThanOrEqual(-1);
  expect(
    box.x + box.width,
    `${label} should not overflow right edge`,
  ).toBeLessThanOrEqual(metrics.innerWidth + 2);
  return { box, metrics };
}

async function expectTouchSized(locator, label) {
  const box = await locator.boundingBox();
  expect(box, `${label} should render`).not.toBeNull();
  expect(
    box.height,
    `${label} should be at least a 44px touch target`,
  ).toBeGreaterThanOrEqual(44);
  expect(
    box.width,
    `${label} should be wide enough to tap/click reliably`,
  ).toBeGreaterThanOrEqual(120);
  return box;
}

test("home workbench links open every destination by keyboard", async ({ page }) => {
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  const study = page.getByRole("navigation", { name: "Start a study" });
  const labs = page.getByRole("navigation", { name: "Explore a lab" });
  const moduleIds = ["hysteresis", "circuits", "crossbar", "comparison", "mnist", "eda", "docs"];
  const destinations = [
    [study, "/projects"],
    ...moduleIds.map((id) => [labs, `/modules/${id}`]),
  ];

  for (const [region, href] of destinations) {
    await page.goto("/");
    const heading = page.getByRole("main").getByRole("heading", { level: 1 });
    await expect(study).toBeVisible();
    await expect(labs).toBeVisible();
    await expect(heading).toBeFocused();
    const link = region.locator(`a[href="${href}"]`);
    await expect(link).toBeVisible();
    for (let i = 0; i < 9; i++) {
      await page.keyboard.press("Tab");
      if (await link.evaluate((el) => el === document.activeElement)) break;
    }
    await expect(link).toBeFocused();
    await expectLocatorInsideViewport(page, link, href);
    const metrics = await viewportMetrics(page);
    expect(metrics.scrollWidth).toBeLessThanOrEqual(page.viewportSize().width);
    await page.keyboard.press("Enter");
    await expect(page).toHaveURL(new RegExp(`${href}$`));
    await expect(heading).toBeVisible();
    if (href === "/projects" && process.env.FECIM_WEB_PROFILE === "local") {
      // The static local preview has no authenticated grant. Projects owns
      // focus for this error and tells the user how to open a real session.
      const alert = page.getByRole("alert");
      await expect(alert).toContainText("Authentication required. Open the one-time launch link from the terminal.");
      await expect(alert).toBeFocused();
    } else {
      await expect(heading).toBeFocused();
    }
  }
  expect(errors).toEqual([]);
});

test("home presents the workbench, profile, and simulation caveat", async ({ page }) => {
  await page.goto("/");

  await expect(
    page.getByRole("heading", { name: "Research workbench", level: 1 }),
  ).toBeVisible();
  await expect(page.getByText(/Education-phase simulation/i)).toBeVisible();
  const study = page.getByRole("navigation", { name: "Start a study" });
  await expect(study).toContainText(
    process.env.FECIM_WEB_PROFILE === "local" ? "Local application:" : "Hosted browser:",
  );
  const labs = page.getByRole("navigation", { name: "Explore a lab" });
  await expect(labs.getByRole("link")).toHaveCount(7);
  await expect(labs).toContainText("MNIST inference and EDA generation require the authenticated local application");
  await expect(labs).toContainText("their hosted views provide read-only help and imports");
  await expect(page.getByRole("link", { name: /Run locally/i })).toBeVisible();
  await expect(page.getByText(/Fyne browser demo/i)).toHaveCount(0);
  await expectNoHorizontalOverflow(page);
});

test("desktop landing keeps primary content above the fold", async ({
  page,
}, testInfo) => {
  test.skip(
    !isDesktopProject(testInfo.project.name),
    "desktop layout contract",
  );

  await page.goto("/");

  const title = page.getByRole("heading", { name: "Research workbench", level: 1 });
  const modules = page.getByRole("navigation", { name: "Explore a lab" });
  const runLocal = page.getByRole("link", { name: /Run locally/i });

  await expect(title).toBeVisible();
  await expect(modules).toBeVisible();
  await expect(runLocal).toBeVisible();

  const titleBox = await title.boundingBox();
  const runLocalBox = await runLocal.boundingBox();
  const metrics = await expectNoHorizontalOverflow(page);
  expect(titleBox, "workbench title should render on desktop").not.toBeNull();
  expect(
    runLocalBox,
    "run-local button should render on desktop",
  ).not.toBeNull();
  expect(titleBox.y).toBeLessThan(metrics.innerHeight * 0.22);
  expect(runLocalBox.y + runLocalBox.height).toBeLessThanOrEqual(
    Math.min(metrics.innerHeight - 12, 760),
  );
});

test("mobile landing keeps CTA and honesty content usable", async ({
  page,
}, testInfo) => {
  test.skip(!isMobileProject(testInfo.project.name), "mobile layout contract");

  await page.goto("/");

  const runLocal = page.getByRole("link", { name: /Run locally/i });
  const honesty = page.getByText(/Education-phase simulation/i);
  await expect(
    page.getByRole("heading", { name: "Research workbench", level: 1 }),
  ).toBeVisible();
  await expect(runLocal).toBeVisible();
  await expect(honesty).toBeVisible();

  const { box: runLocalBox, metrics } = await expectLocatorInsideViewport(
    page,
    runLocal,
    "mobile run-local button",
  );
  await expectTouchSized(runLocal, "mobile run-local button");
  expect(runLocalBox.y + runLocalBox.height).toBeLessThanOrEqual(
    metrics.innerHeight - 16,
  );

  const honestyBox = await honesty.boundingBox();
  expect(honestyBox, "honesty note should render on mobile").not.toBeNull();
  expect(honestyBox.y).toBeLessThan(metrics.innerHeight);
  await expectNoHorizontalOverflow(page);
});

test("landing navigation anchors work on every viewport project", async ({
  page,
}) => {
  const anchors = [
    {
      href: "/#modules",
      id: "modules",
      heading: "Explore a lab",
    },
    {
      href: "/#science",
      id: "science",
      heading: /Built to teach the physics/i,
    },
    { href: "/#start", id: "start", heading: /Clone, test, and launch/i },
  ];

  for (const anchor of anchors) {
    await page.goto("/");
    await page.locator(`.nav__links a[href="${anchor.href}"]`).click();
    await expect(page).toHaveURL(new RegExp(`${anchor.href}$`));
    await expect(
      page.getByRole("heading", { name: anchor.heading }),
    ).toBeVisible();
    await expect(page.locator(`#${anchor.id}`)).toBeFocused();
    await page.waitForFunction((id) => {
      const section = document.getElementById(id);
      if (!section) return false;
      const box = section.getBoundingClientRect();
      return box.top >= -40 && box.top < window.innerHeight * 0.7;
    }, anchor.id);
    await expectNoHorizontalOverflow(page);
  }
});

test("landing sections stay readable across viewport projects", async ({
  page,
}, testInfo) => {
  await page.goto("/");

  const stats = page.locator(".stats article");
  await expect(stats).toHaveCount(3);
  await page.locator("#modules").scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("heading", { name: "Explore a lab" }),
  ).toBeVisible();

  const moduleCards = page.locator(".module-card");
  await expect(moduleCards).toHaveCount(7);
  const cardBounds = await moduleCards.evaluateAll((cards) =>
    cards.map((card) => {
      const box = card.getBoundingClientRect();
      return { x: box.x, width: box.width, text: card.textContent.trim() };
    }),
  );
  const metrics = await expectNoHorizontalOverflow(page);
  const minCardWidth = isMobileProject(testInfo.project.name) ? 280 : 240;
  for (const card of cardBounds) {
    expect(
      card.width,
      `${card.text} should remain readable`,
    ).toBeGreaterThanOrEqual(minCardWidth);
    expect(card.x, `${card.text} should not spill left`).toBeGreaterThanOrEqual(
      -1,
    );
    expect(
      card.x + card.width,
      `${card.text} should not spill right`,
    ).toBeLessThanOrEqual(metrics.innerWidth + 2);
  }

  await page.locator("#science").scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("heading", { name: /Built to teach the physics/i }),
  ).toBeVisible();
  await expect(page.locator("#science li")).toHaveCount(3);
  await expectNoHorizontalOverflow(page);

  await page.locator("#start").scrollIntoViewIfNeeded();
  await expect(
    page.getByRole("heading", { name: /Clone, test, and launch/i }),
  ).toBeVisible();
  await expect(page.locator("#start pre")).toBeVisible();
  await expectLocatorInsideViewport(
    page,
    page.locator("#start"),
    "start section",
  );
});
