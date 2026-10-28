# FeCIM shared React application

One React + TypeScript + Vite application owns `/` and `/hysteresis`. The home retains the module overview, literature links and education-phase caveats. Hysteresis runs the production Go model in a dedicated Worker; TypeScript owns controls, transport scheduling, focus and rendering, not scientific defaults, equations, conversions or speech.

## Build and develop

Prerequisites: Node 24+, npm, and the **exact Go version in `../go.mod`** (currently 1.26.10). Dependencies are locked. Go's toolchain manager must already have that version or be able to provision it at build time.

```bash
cd web
npm ci
npm test
npm run typecheck
npm run build:hosted     # dist/
npm run build:local      # dist-local/; identical views, local adapter callback seam
npm run dev -- --port 20200
# Or serve compiled hosted assets:
npm run preview -- --host 127.0.0.1 --port 20200 --strictPort
# Serve the local asset profile, without claiming a local Go API:
npm run preview -- --host 127.0.0.1 --port 20200 --strictPort --outDir dist-local
```

Use your assigned `AUTOWORK_PORT_BASE` instead of 20200 under Autowork. Users of compiled assets need **no runtime Node installation**. Existing CLI/Fyne commands and launch defaults are unchanged.

Both builds invoke `src/build/build.mjs`. It pins Go from `go.mod`, hashes the same source/dependency/config/Worker/runtime inventory as the production parity gate, builds `cmd/fecim-hysteresis-wasm`, and uses **that toolchain's** `lib/wasm/wasm_exec.js`. `src/build/catalog.go` calls existing Go defaults and enum APIs to generate the supplementary presentation catalogue. Temporary generated assets live under `node_modules/.fecim`, never in source control.

The emitted `engine/manifest.json` binds frontend content, protocol/model/build/Go identities and SHA-256 asset hashes. The compiled frontend checks its fingerprint before starting the unchanged production Worker client, which checks module/runtime/Worker bytes and Go identity. After editing source, rebuild the complete profile; do not copy an old WASM module or runtime into a newer asset set. A restart discards the ephemeral session; a Retry creates a new verified Worker/session.

## Hysteresis workflow

- Starts paused with the Go-selected material, model and Sine drive. Apply edited settings explicitly. Dirty settings and revision-bound stale tables/diagnostics are labelled, not presented as newly computed results.
- Run continuously, Pause, advance one bounded batch, Cancel or Reset. DT is explicit physical seconds per Go step; changing frequency does not silently retune DT. Select Manual and apply before direct-field or programming operations. Playback is not a scientific timestep.
- Read loop results through native, revision-bound pages, with Go display/speech and canonical SI rows. Material metadata also has Go-owned units and speech; source preset Ea (eV) and StressGPa are presented in SI joules and pascals without changing the model. Every retained row is reachable, including the last page and last-complete-cycle view. The bounded plot is supplementary and has no unique actions. Go's history ring reports dropped samples rather than silently claiming unbounded retention.
- Initialize target-write, write/read or the explicitly **unavailable legacy time-resolved protocol**. Inspect controller phase, target/queued target, read level, limits and reasons. Write/read accepts queued targets. Cancel and Reset use Go ownership rules. A work-limit or nonconvergence result is not success.
- Calibrate in memory, inspect Go freshness/fallback status and all ascending/descending values. Changing scientific settings invalidates calibration; explicitly recalibrate. Fresh simulation calibration is not evidence of measured device convergence.
- Run PUND or FORC, inspect actual independent Preisach diagnostic identity, parameters, selected-loop context, charges, sweep/density pages, availability, illustrative bands and citations. Reference files are explicit read-only in-memory inputs; Go validates CSV/JSON units, branch order, byte limits and provenance. Imports remain **unverified**, never fitted or experimentally validated. Analyze waits for a selected reference to finish reading; Clear/reselection discards earlier pending reads, and read errors must be cleared or corrected. Embedded JSON source/material provenance is parsed by Go rather than replaced by the filename or selected-loop material. Edited diagnostic inputs label retained results stale until explicitly analyzed again.

## Capability and privacy boundary

Hosted computation has no scientific upload, analytics, persistence, filesystem writes, automatic downloads, clipboard export, remote solver or localhost discovery. Ordinary same-origin static asset requests still occur and the host sees HTTP metadata. Citation links are deliberate navigation. Reload/discard loses ephemeral science.

Local-only buttons remain navigable instructions. The local profile exposes a callback seam (`window.fecimLocalAdapter.perform(action)`) but **the authenticated local adapter, embedded assets and `fecim serve` command are not implemented here**; HYS-LOCAL owns them. A local build flag alone grants no capability. This task does not change the executable's legacy Fyne launch default or retire any Fyne module.

## Browser and scientific checks

```bash
# Provision compatible engines if absent (build/test tools, not end-user dependencies):
npx playwright install chromium firefox
npm run test:e2e -- test/e2e/hysteresis.spec.ts test/e2e/hosted-isolation.spec.ts
FECIM_WEB_PROFILE=local npm run test:e2e
npm run test:e2e         # also retains existing home/content/layout/anchor guards
```

Playwright owns two private static servers at `AUTOWORK_PORT_BASE` and `+1` (standalone default 20200/20201); it does not reuse unrelated servers. `PLAYWRIGHT_CHROMIUM_EXECUTABLE` / `PLAYWRIGHT_FIREFOX_EXECUTABLE` select compatible installed engines. Worker caches under Autowork are reused when present. CI builds both profiles and runs the same browser contracts against each, plus the separate exact native/Node/Firefox Go parity gate.

The route budget case measures **all actual HTTP assets** from cold navigation through the first paused signed Go-default Sine result and complete first table page, in three fresh contexts. Its server negotiates Brotli and adds 100 ms request RTT on one aggregate 10 Mb/s response schedule; no host networking changes. Gates: full route ≤2 MiB Brotli, each cold result ≤5 s; module/runtime/Worker ≤6 MiB raw, ≤1.5 MiB Brotli and ≤2 MiB gzip. Results include identity, asset totals, resource filter and per-run timing in `route-budget.json` attachments and `ROUTE_BUDGET` log records. These private-host results are not CDN, low-end-device or real-time guarantees.

See [WASM gate evidence](../docs/internals/testing/wasm-parity-gate.md#5-shared-react-route-qualification) and [production transport/parity instructions](../validation/wasm/README.md). DOM, keyboard/focus, narrow-width/200% text and automated accessibility checks **do not qualify Orca speech**. Real Orca + Firefox/Linux acceptance for hosted and authenticated local profiles remains the separate [manual gate](../docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md).

## Cloudflare Pages output

`wrangler.toml` still selects `./dist`, project `fecim`, and `public/_redirects` preserves direct `/hysteresis` navigation. `npm run deploy` builds the hosted profile before the existing Wrangler command; deployment is an explicit external operation, not part of validation and was not performed here. A hosting CSP must permit the verified Blob Worker/runtime and Go WASM compilation described in the [Worker documentation](../validation/wasm/README.md#production-protocol-and-ownership).
