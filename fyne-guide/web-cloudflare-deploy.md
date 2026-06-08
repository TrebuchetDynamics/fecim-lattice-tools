# FeCIM Fyne web + Cloudflare Wrangler deploy notes

## Current web target

- Fyne WASM entrypoint: `cmd/fecim-web-fyne`
- Static site root: `web/`
- Fyne browser route after build: `/fecim/`
- Cloudflare Pages project: `fecim`
- Cloudflare Pages production branch: `development`
- Cloudflare Pages output: `web/dist`
- Wrangler config: `web/wrangler.toml`

The browser app is a reduced interactive FeCIM demo with crossbar controls, physics sketches, circuit knobs, EDA previews, and honesty boundaries. It avoids the full native simulator graph while WebAssembly compatibility is expanded.

## Build

```bash
cd web
npm test
npm run build
```

`npm run build` runs:

1. `npm run build:wasm`
   - builds `../cmd/fecim-web-fyne` with `GOOS=js GOARCH=wasm CGO_ENABLED=0` and `-ldflags "-s -w"`
   - compresses the WASM to `web/public/fecim/fecim.wasm.gz`
   - copies Go `wasm_exec.js` into `web/public/fecim/wasm_exec.js`
   - fails if `fecim.wasm.gz` exceeds Cloudflare Pages' 25 MiB single-file limit
2. `astro build`
   - emits `web/dist`

Validated output in this session:

- `web/dist/index.html`
- `web/dist/fecim/index.html`
- `web/dist/fecim/fecim.wasm.gz` (~11.56 MiB)
- `web/dist/fecim/wasm_exec.js`

## Deploy

```bash
cd web
wrangler login
npm run deploy
```

Non-interactive/CI alternative:

```bash
cd web
CLOUDFLARE_API_TOKEN=... npm run deploy
```

In this session, `../gormes/.env` contained `CLOUDFLARE_API_KEY`; it was mapped to `CLOUDFLARE_API_TOKEN` for Wrangler without printing the secret value.

## Session validation receipts

Passed:

```bash
go test ./cmd/fecim-web-fyne -count=1
CGO_ENABLED=0 go test ./shared/accessibility -count=1
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -o /tmp/fecim-web-fyne.wasm ./cmd/fecim-web-fyne
cd web && npm test
cd web && npm run build
python3 - <<'PY'
from pathlib import Path
limit = 25 * 1024 * 1024
for path in Path('web/dist').rglob('*'):
    if path.is_file() and path.stat().st_size > limit:
        raise SystemExit(f'oversize {path}')
print('dist_file_size_check=ok')
PY
```

Deploy receipts:

```bash
wrangler pages project create fecim --production-branch development
cd web && npm run deploy
```

Deployment succeeded:

- Production URL: `https://fecim.pages.dev`
- Preview URL: `https://862d9478.fecim.pages.dev`
- HTTP checks passed for `/` and `/fecim/` with status 200 on both production and preview URLs.

Earlier deploy blockers resolved:

- Missing Wrangler auth in non-interactive environment → used env from `../gormes/.env` without printing secrets.
- Missing Pages project `fecim` → created the Pages project.
- Cloudflare Pages 25 MiB single-file limit for raw `fecim.wasm` (~39 MiB) → deploys `fecim.wasm.gz` (~11.56 MiB) and decompresses in-browser with `DecompressionStream`.

## 2026-06-06 `/fecim/` load diagnosis

User reported `https://fecim.pages.dev/fecim/` did not load.

Repro/diagnosis:

```bash
chromium --headless=new --no-sandbox --disable-gpu --virtual-time-budget=30000 --dump-dom https://fecim.pages.dev/fecim/
```

Observed fallback reason:

```text
WebGL context creation failed or WebGL is disabled
```

With an isolated Chromium profile and software WebGL flags, the page proceeded past WASM startup and emitted non-fatal Fyne metadata warnings:

```text
Fyne error: failed to lookup build executable
Cause: Executable not implemented for js
At: .../fyne/v2@v2.7.4/app/meta_development.go:59
```

Fix deployed:

- `web/public/fecim/index.html` now shows a static diagnostic/fallback panel instead of a dead loading state.
- Fallback reports WebAssembly, `DecompressionStream`, and WebGL availability.
- Fallback links to `/?wasm=off` and a retry link.
- `web/test/landing-contract.test.mjs` asserts the fallback/diagnostics contract.

Redeploy receipt:

```bash
cd web && npm test && npm run build
cd web && npm run deploy
```

Updated preview URL after diagnostics fallback: `https://7d1ccb4d.fecim.pages.dev`.

## 2026-06-06 docs-only demo replacement

User feedback: `/fecim/` only loaded documentation and was not good enough.

Fix:

- Replaced `cmd/fecim-web-fyne` docs-only entrypoint with a reduced interactive FeCIM Fyne app.
- New browser app tabs:
  - Overview
  - Crossbar Playground
  - Hysteresis Physics
  - Peripheral Circuits
  - EDA Export
  - Honesty
- Added tests for interactive panels, quantization/metrics helpers, and Fyne tab construction in `cmd/fecim-web-fyne/main_test.go`.
- Removed the `module7-docs/pkg/gui` dependency from the web entrypoint, reducing WASM size slightly.
- Updated wording in `web/public/fecim/index.html`, `web/src/pages/index.astro`, and `web/README.md` away from "documentation-first".

Validation receipts:

```bash
go test ./cmd/fecim-web-fyne -count=1
cd web && npm test && npm run build
GOOS=js GOARCH=wasm CGO_ENABLED=0 go build -ldflags='-s -w' -o /tmp/fecim-web-interactive-final.wasm ./cmd/fecim-web-fyne
```

WASM size after replacement:

- Raw stripped WASM: ~36.66 MiB
- Deployed gzip WASM: ~11.22 MiB

Deploy receipt:

```bash
cd web && npm run deploy
```

Deployment succeeded:

- Production URL: `https://fecim.pages.dev/fecim/`
- Latest preview URL: `https://8ac12a31.fecim.pages.dev/fecim/`

## 2026-06-06 Playwright browser testing

User feedback: browser quality was awful and needed Playwright testing.

Added:

- `web/playwright.config.mjs`
- `web/test/e2e/fecim-web.spec.mjs`
- `@playwright/test` dev dependency in `web/package.json` / `web/package-lock.json`
- npm scripts:
  - `npm run test:e2e`
  - `npm run test:e2e:build`

Playwright uses system Chromium by default (`/usr/bin/chromium`) and can be overridden with `PLAYWRIGHT_CHROMIUM_EXECUTABLE=/path/to/chromium`. Browser downloads are not required for the local harness.

E2E coverage:

- Landing page has an obvious Launch Fyne browser demo link to `/fecim/`.
- `/fecim/` HTML is interactive FeCIM copy, not docs-only/documentation-browser copy.
- WebGL-disabled browser path shows useful diagnostics and fallback links instead of a dead loader.

Validation receipt:

```bash
cd web && npm test && npm run test:e2e:build
```

Result: contract tests passed and 3/3 Playwright tests passed.

## 2026-06-06 wider desktop/mobile Playwright testing

User feedback: Playwright coverage was still too narrow; add mobile and desktop testing.

Added/changed:

- `web/playwright.config.mjs` now has explicit projects:
  - `desktop-chromium` at 1440×900
  - `mobile-chromium` using an iPhone-13-sized Chromium-compatible viewport/touch/DPR profile
- Playwright runs with `workers: 1` so system Chromium is stable in this environment.
- `web/test/e2e/fecim-web.spec.mjs` now covers:
  - shared landing launch route
  - desktop above-the-fold primary content
  - mobile first-screen launch usability
  - no horizontal overflow on desktop/mobile
  - docs-only copy rejection on `/fecim/`
  - WebGL-disabled fallback diagnostics
- `web/test/landing-contract.test.mjs` now asserts the desktop/mobile Playwright config and e2e spec names.
- Fixed landing hero typography and mobile nav spacing in `web/src/pages/index.astro` because Playwright caught:
  - desktop launch CTA below the intended fold gate
  - mobile launch CTA below the first viewport

RED evidence:

```bash
cd web && npm test
# failed: missing desktop-chromium/mobile-chromium projects

cd web && npm run test:e2e
# failed: desktop launch CTA bottom 1853px, later 813px, expected <760px
# failed: mobile project/profile initially unstable under system Chromium

cd web && npm run test:e2e -- --project mobile-chromium --grep "mobile landing"
# failed: mobile launch CTA bottom 888.95px, expected <=844px
```

GREEN evidence:

```bash
cd web && npm test && npm run test:e2e
# contract tests passed
# 8 passed, 2 project-specific tests skipped
```

Deploy receipt:

```bash
cd web && npm run deploy
```

Deployment succeeded:

- Latest preview URL: `https://319e7ea6.fecim.pages.dev`
- Production URL: `https://fecim.pages.dev`

Post-deploy layout probe:

```text
desktop: launchBottom=752 / innerHeight=900, overflow=0
mobile:  launchBottom=690 / innerHeight=877, overflow=0
```

## 2026-06-06 expanded desktop/mobile Playwright matrix

User feedback: Playwright coverage was still awful; make it wider for both mobile and desktop.

Added/changed:

- `web/playwright.config.mjs` now runs four Chromium viewport projects:
  - `desktop-chromium` at 1440×900
  - `desktop-compact-chromium` at 1024×768
  - `mobile-chromium` using an iPhone-13-sized Chromium-compatible touch/DPR profile
  - `mobile-small-chromium` at 360×740 with touch/DPR emulation
- `web/test/e2e/fecim-web.spec.mjs` now has broader behavior coverage:
  - launch route contract on every viewport project
  - desktop above-the-fold CTA on both desktop projects
  - mobile first-viewport launch/honesty usability on both mobile projects
  - nav anchor behavior for Modules, Science, and Start on every viewport project
  - module-card/section readability and horizontal-overflow checks across viewports
  - `/fecim/` interactive copy contract
  - WebGL-disabled fallback diagnostics
  - fallback controls reachable and at least 44px touch-sized
- `web/test/landing-contract.test.mjs` now asserts the expanded viewport matrix and new e2e coverage names.
- `web/src/pages/index.astro` layout fixes from the wider suite:
  - clipped decorative `.hero__grid` overflow at the hero boundary
  - tightened mobile honesty-pill margin so the honesty note starts within a 360×740 first viewport

RED evidence:

```bash
cd web && npm test
# failed: missing desktop-compact-chromium project

cd web && npm run test:e2e
# failed: desktop-compact-chromium had 66px horizontal overflow from .hero__grid
# failed: mobile-small-chromium honesty note y=748.06, expected <740
```

GREEN evidence:

```bash
cd web && npm test && npm run test:e2e
# contract tests passed
# 28 passed, 4 project-specific tests skipped

cd web && npm run test:e2e:build
# build passed
# 28 passed, 4 project-specific tests skipped
```

Deploy receipt:

```bash
cd web && npm run deploy
```

Deployment succeeded:

- Latest preview URL: `https://29ec2953.fecim.pages.dev`
- Production URL: `https://fecim.pages.dev`

Post-deploy four-profile probe:

```text
desktop:         launchBottom=752 / innerHeight=900, overflow=0
desktop-compact: launchBottom=666 / innerHeight=768, overflow=0
mobile:          launchBottom=690 / innerHeight=844, overflow=0
mobile-small:    launchBottom=674 / innerHeight=740, overflow=0
```
