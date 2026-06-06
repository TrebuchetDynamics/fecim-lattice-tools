import { expect, test } from '@playwright/test';

const DESKTOP_PROJECTS = new Set(['desktop-chromium', 'desktop-compact-chromium']);
const MOBILE_PROJECTS = new Set(['mobile-chromium', 'mobile-small-chromium']);

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
  expect(box.x, `${label} should not start off the left edge`).toBeGreaterThanOrEqual(-1);
  expect(box.x + box.width, `${label} should not overflow right edge`).toBeLessThanOrEqual(metrics.innerWidth + 2);
  return { box, metrics };
}

async function expectTouchSized(locator, label) {
  const box = await locator.boundingBox();
  expect(box, `${label} should render`).not.toBeNull();
  expect(box.height, `${label} should be at least a 44px touch target`).toBeGreaterThanOrEqual(44);
  expect(box.width, `${label} should be wide enough to tap/click reliably`).toBeGreaterThanOrEqual(120);
  return box;
}

test('landing page presents Astro-only product story', async ({ page }) => {
  await page.goto('/');

  await expect(page.getByRole('heading', { name: /lattice-scale playground/i })).toBeVisible();
  await expect(page.getByText(/Education-phase simulation/i)).toBeVisible();
  await expect(page.getByRole('link', { name: /Explore the modules/i })).toBeVisible();
  await expect(page.getByRole('link', { name: /Run locally/i })).toBeVisible();
  await expect(page.getByText(/Fyne browser demo/i)).toHaveCount(0);
  await expectNoHorizontalOverflow(page);
});

test('desktop landing keeps primary content above the fold', async ({ page }, testInfo) => {
  test.skip(!isDesktopProject(testInfo.project.name), 'desktop layout contract');

  await page.goto('/');

  const hero = page.locator('.hero__content');
  const modules = page.getByRole('link', { name: /Explore the modules/i });
  const runLocal = page.getByRole('link', { name: /Run locally/i });

  await expect(hero).toBeVisible();
  await expect(modules).toBeVisible();
  await expect(runLocal).toBeVisible();

  const heroBox = await hero.boundingBox();
  const runLocalBox = await runLocal.boundingBox();
  const metrics = await expectNoHorizontalOverflow(page);
  expect(heroBox, 'hero content should render on desktop').not.toBeNull();
  expect(runLocalBox, 'run-local button should render on desktop').not.toBeNull();
  expect(heroBox.y).toBeLessThan(metrics.innerHeight * 0.22);
  expect(runLocalBox.y + runLocalBox.height).toBeLessThanOrEqual(Math.min(metrics.innerHeight - 12, 760));
});

test('mobile landing keeps CTA and honesty content usable', async ({ page }, testInfo) => {
  test.skip(!isMobileProject(testInfo.project.name), 'mobile layout contract');

  await page.goto('/');

  const runLocal = page.getByRole('link', { name: /Run locally/i });
  const honesty = page.getByText(/Education-phase simulation/i);
  await expect(page.getByRole('heading', { name: /lattice-scale playground/i })).toBeVisible();
  await expect(runLocal).toBeVisible();
  await expect(honesty).toBeVisible();

  const { box: runLocalBox, metrics } = await expectLocatorInsideViewport(page, runLocal, 'mobile run-local button');
  await expectTouchSized(runLocal, 'mobile run-local button');
  expect(runLocalBox.y + runLocalBox.height).toBeLessThanOrEqual(metrics.innerHeight - 16);

  const honestyBox = await honesty.boundingBox();
  expect(honestyBox, 'honesty note should render on mobile').not.toBeNull();
  expect(honestyBox.y).toBeLessThan(metrics.innerHeight);
  await expectNoHorizontalOverflow(page);
});

test('landing navigation anchors work on every viewport project', async ({ page }) => {
  const anchors = [
    { href: '#modules', id: 'modules', heading: /Seven modules, one simulation story/i },
    { href: '#science', id: 'science', heading: /Built to teach the physics/i },
    { href: '#start', id: 'start', heading: /Clone, test, and launch/i },
  ];

  for (const anchor of anchors) {
    await page.goto('/');
    await page.locator(`.nav__links a[href="${anchor.href}"]`).click();
    await expect(page).toHaveURL(new RegExp(`${anchor.href}$`));
    await expect(page.getByRole('heading', { name: anchor.heading })).toBeVisible();
    await page.waitForFunction((id) => {
      const section = document.getElementById(id);
      if (!section) return false;
      const box = section.getBoundingClientRect();
      return box.top >= -40 && box.top < window.innerHeight * 0.7;
    }, anchor.id);
    await expectNoHorizontalOverflow(page);
  }
});

test('landing sections stay readable across viewport projects', async ({ page }, testInfo) => {
  await page.goto('/');

  const stats = page.locator('.stats article');
  await expect(stats).toHaveCount(3);
  await page.locator('#modules').scrollIntoViewIfNeeded();
  await expect(page.getByRole('heading', { name: /Seven modules, one simulation story/i })).toBeVisible();

  const moduleCards = page.locator('.module-card');
  await expect(moduleCards).toHaveCount(7);
  const cardBounds = await moduleCards.evaluateAll((cards) => cards.map((card) => {
    const box = card.getBoundingClientRect();
    return { x: box.x, width: box.width, text: card.textContent.trim() };
  }));
  const metrics = await expectNoHorizontalOverflow(page);
  const minCardWidth = isMobileProject(testInfo.project.name) ? 280 : 240;
  for (const card of cardBounds) {
    expect(card.width, `${card.text} should remain readable`).toBeGreaterThanOrEqual(minCardWidth);
    expect(card.x, `${card.text} should not spill left`).toBeGreaterThanOrEqual(-1);
    expect(card.x + card.width, `${card.text} should not spill right`).toBeLessThanOrEqual(metrics.innerWidth + 2);
  }

  await page.locator('#science').scrollIntoViewIfNeeded();
  await expect(page.getByRole('heading', { name: /Built to teach the physics/i })).toBeVisible();
  await expect(page.locator('#science li')).toHaveCount(3);
  await expectNoHorizontalOverflow(page);

  await page.locator('#start').scrollIntoViewIfNeeded();
  await expect(page.getByRole('heading', { name: /Clone, test, and launch/i })).toBeVisible();
  await expect(page.locator('#start pre')).toBeVisible();
  await expectLocatorInsideViewport(page, page.locator('#start'), 'start section');
});
