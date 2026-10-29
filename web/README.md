# FeCIM shared React application

One React + TypeScript + Vite application owns `/`, `/hysteresis`, `/projects`, `/configuration`, `/experiment`, `/results`, `/docs` and fixed `/modules/<id>` slots. All seven module views and the research workflow are source-built candidates, not a qualified release. Hosted provides Hysteresis, bounded studies, Crossbar, Circuits, Comparison and read-only Documentation. MNIST inference and EDA generation require the authenticated local application; their hosted views provide read-only help/imports. Real installed Icarus/Yosys qualification remains NOT RUN. Real Orca + Firefox/Linux qualification is **NOT RUN**. The home retains the module overview, literature links and education-phase caveats. Hysteresis runs the production Go model in a dedicated Worker; TypeScript owns controls, transport scheduling, focus and rendering, not scientific defaults, equations, conversions or speech.

## Build and develop

Build/test pins: Node **24.19.0**, npm **11.17.0**, and the **exact Go version in `../go.mod`** (currently 1.26.10). Dependencies are locked. Go's toolchain manager must already have that version or be able to provision it at build time.

```bash
cd web
npm ci
npm run generate:api
npm test
npm run typecheck
npm run build:hosted     # dist/
npm run build:local      # dist-local/ + generated local embedding stage
npm run build:native     # matching embedded fecim-lattice-tools binary
npm run dev -- --port 20200
# Or serve compiled hosted assets:
npm run preview -- --host 127.0.0.1 --port 20200 --strictPort
# Serve the local asset profile, without claiming a local Go API:
npm run preview -- --host 127.0.0.1 --port 20200 --strictPort --outDir dist-local
```

Use your assigned `AUTOWORK_PORT_BASE` instead of 20200 under Autowork. Users of compiled assets need **no runtime Node installation**. `build:native` explicitly selects CGO=1 and requires the native desktop build libraries; plain CGO=0 Go builds support headless command dispatch, not default desktop startup. Existing CLI/Fyne commands and launch defaults are unchanged.

Both builds invoke `src/build/build.mjs`. It pins Go from `go.mod`, hashes the same source/dependency/config/Worker/runtime inventory as the production parity gate, builds `cmd/fecim-hysteresis-wasm`, and uses **that toolchain's** `lib/wasm/wasm_exec.js`. `src/build/catalog.go` calls existing Go defaults and enum APIs to generate the supplementary presentation catalogue. Temporary generated assets live under `node_modules/.fecim`, never in source control.

The emitted `engine/manifest.json` binds frontend content, protocol/model/build/Go identities and SHA-256 asset hashes. The compiled frontend checks its fingerprint before starting the unchanged production Worker client, which checks module/runtime/Worker bytes and Go identity. After editing source, rebuild the complete profile; do not copy an old WASM module or runtime into a newer asset set. A restart discards the ephemeral session; a Retry creates a new verified Worker/session.

## Shared research workflow

**Projects → Configuration → Experiment → Results** reuses Go configuration, study, immutable runs, analysis and original report serializers. Local mode supports rooted creation/open/edit/validate, asynchronous run/cancel/reconnect/reuse, inspection, 2–4-run comparison, explicit stored-run verification, reports and selected-run exports. Exact quoted ETags protect saves and starts. Stale/conflicting work cannot bypass the revision or mutation gate. GETs do not regenerate reports or write scientific state.

Hosted studies stay in memory. Select YAML/input bytes or a `documents` JSON map of base64 bytes, or open the small eight-point example. Go enforces 8 MiB total bundle bytes, 1 MiB per YAML file, at most 256 points and 32×32 arrays before expansion. The small fixture uses 16/32-row arrays; the original example with 64-row cases is native-only. No hosted edit, report/export/download, persistent reuse, scientific upload, browser storage or filesystem API is added. See the [research guide](../docs/guides/research-design-workbench.md#browser-research-workflow-candidate) for limits, conflict review, result semantics and recovery.

Research ownership survives SPA navigation, fragments and Back/Forward. Go supplies stable pages/revisions, table captions, numeric display/speech/evidence and full-filtered-scope plot coordinates. Selection mirrors immutable IDs; the supplementary plot has no unique actions or point Tab stops. A retired response cannot overwrite newer field drafts or steal focus in another view. A Worker transport fault is unknown ephemeral execution, not a successful or cancelled scientific outcome. Explicit discard permits a fresh import/example; previous results remain readable until then.

## Shared module transport and extension conventions

- [moduleapi/types.go](../workbench/moduleapi/types.go) owns `Command`, `Snapshot`, fields, pages, jobs, capabilities, `Port`, optional `JobPort` and `Provider`. The Go `Port` methods are `Descriptor`, `Snapshot`, `Apply`, `Close`; `Provider.WebPort()` can adapt an existing memory-only module constructor. Legacy status badges or display metrics do not imply a scientific web port.
- [server/modules.go](../workbench/server/modules.go) is the sole seven-module native constructor/factory list. Hysteresis uses the existing bridge adapter; the other six module ports adapt those constructors through `Port` or `Provider`, without a second central route/test registry. No constructor probe may start work, open cwd/home files or acquire tools.
- Native callbacks receive optional [resource](../workbench/moduleapi/native/resources.go) and [project](../workbench/moduleapi/native/projects.go) capabilities. `Projects(ctx).WithProject` leases an already opened opaque project reference, shares study mutation exclusion and holds its rooted lifetime through the callback. Do not retain/reopen `root.Name()`, leak handles into DTOs or outlive the callback. Ports must honor cancellation and drain callbacks/readers before `Close` returns. Explicit large-file bindings remain bounded regular-file reads, not arbitrary browser filesystem access.
- Put a React default export at `src/modules/<id>/Screen.tsx`. Vite uses a fixed build-time glob; URL strings and server data never become arbitrary dynamic imports. An unavailable descriptor or absent screen remains an explanatory route. Mounted owners survive navigation. Docs may coexist with active science; competing scientific owners report busy.
- Hosted producers put `main_js.go` and a native-runnable `identity.go` under `src/modules/<id>/wasm/`. [build-models.mjs](scripts/build-models.mjs) discovers and builds each bundle independently. The sidecar emits either its existing four-key `{protocol,model,build,go}` identity (research compatibility), or `{identity,descriptor}`. Descriptors must match a known module ID, shared schema version and approved hosted/local profiles. Available bundles merge into the Go baseline catalogue; no future central capability edit is required. This registration is not numerical qualification.
- The [internal Huma router](../workbench/moduleapi/httpapi/api.go) stays behind the existing listener/session/Host/Origin/Fetch-Metadata checks. Only configuration PUT is added; no automatic PATCH or public docs/schema route. Mutations retain Origin/JSON/custom-header checks and the <=8 MiB body limit. Pagination queries are route-specific, bounded and contain no tokens. Jobs and SSE use separate admission/lifetimes; stalled consumers cannot retain the global session lock or runner.
- Go DTO/router inputs generate [openapi.json](../workbench/moduleapi/openapi.json); pinned `openapi-typescript` generates [generated.ts](src/api/generated.ts), consumed by `openapi-fetch` in the local adapter. Run `npm run generate:api` after Go contract changes. `make check-api` regenerates and checks drift. CI does the same; `scripts/ci_contract_test.go::TestCIRunsWebTestsAndProductionBuild` guards the required locked install/generation/test/production-build chain, generated drift rejection and both profile builds. Run `go test ./scripts -count=1` from the repository root after changing that CI pipeline. Private runtime docs/schema routes stay disabled. Assets bind UI/API/core/toolchain/Worker digests. Rebuild both profiles and the matching native binary after source changes.

## Hysteresis workflow

- Starts paused with the Go-selected material, model and Sine drive. Apply edited settings explicitly. Dirty settings and revision-bound stale tables/diagnostics are labelled, not presented as newly computed results.
- Run continuously, Pause, advance one bounded batch, Cancel or Reset. DT is explicit physical seconds per Go step; changing frequency does not silently retune DT. Select Manual and apply before direct-field or programming operations. Playback is not a scientific timestep.
- Read loop results through native, revision-bound pages, with Go display/speech and canonical SI rows. Material metadata also has Go-owned units and speech; source preset Ea (eV) and StressGPa are presented in SI joules and pascals without changing the model. Every retained row is reachable, including the last page and last-complete-cycle view. The bounded plot is supplementary and has no unique actions. Go's history ring reports dropped samples rather than silently claiming unbounded retention.
- Initialize target-write, write/read or the explicitly **unavailable legacy time-resolved protocol**. Inspect controller phase, target/queued target, read level, limits and reasons. Write/read accepts queued targets. Cancel and Reset use Go ownership rules. A work-limit or nonconvergence result is not success.
- Calibrate in memory, inspect Go freshness/fallback status and all ascending/descending values. Changing scientific settings invalidates calibration; explicitly recalibrate. Fresh simulation calibration is not evidence of measured device convergence.
- Run PUND or FORC, inspect actual independent Preisach diagnostic identity, parameters, selected-loop context, charges, sweep/density pages, availability, illustrative bands and citations. Reference files are explicit read-only in-memory inputs; Go validates CSV/JSON units, branch order, byte limits and provenance. Imports remain **unverified**, never fitted or experimentally validated. Analyze waits for a selected reference to finish reading; Clear/reselection discards earlier pending reads, and read errors must be cleared or corrected. Embedded JSON source/material provenance is parsed by Go rather than replaced by the filename or selected-loop material. Edited diagnostic inputs label retained results stale until explicitly analyzed again.

## Capability and privacy boundary

Hosted computation has no scientific upload, analytics, persistence, filesystem writes, automatic downloads, clipboard export, remote solver or localhost discovery. Ordinary same-origin static asset requests still occur and the host sees HTTP metadata. Citation links are deliberate navigation. Reload/discard loses ephemeral science.

Hosted local-only buttons remain navigable instructions without constructing a local transport. The local build stages generated assets under `internal/browserui/assets`; `npm run build:native` verifies source/UI drift and pins matching native identities. Start `fecim-lattice-tools serve --root DIRECTORY` and open its short-lived launch link for authenticated file operations. A local static preview alone grants no capability. See the [command/security reference](../cmd/fecim-lattice-tools/README.md) for explicit file formats, bootstrap renewal, conflicts, imported calibration limits and threat boundaries. Existing Fyne launch defaults and entry points remain unchanged.

## Browser and scientific checks

```bash
# Provision compatible engines if absent (build/test tools, not end-user dependencies):
npx playwright install chromium firefox
npm run test:e2e -- test/e2e/hysteresis.spec.ts test/e2e/workbench.spec.ts test/e2e/module-router.spec.ts test/e2e/hosted-isolation.spec.ts
FECIM_WEB_PROFILE=local npm run test:e2e
npm run test:e2e         # also retains existing home/content/layout/anchor guards
```

`FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/workbench.spec.ts test/e2e/module-router.spec.ts` exercises the real rooted research service (port base +12), not just static local-profile rendering. `npm run test:e2e -- test/e2e/hysteresis-local.spec.ts` also builds and starts the actual embedded Go binary with a temporary granted root on `AUTOWORK_PORT_BASE + 10`, exercises explicit files/replay/renewal/logout, then cleans up. It is not a static-preview substitute for the local API.

Playwright owns two private static servers at `AUTOWORK_PORT_BASE` and `+1` (standalone default 20200/20201); it does not reuse unrelated servers. `PLAYWRIGHT_CHROMIUM_EXECUTABLE` / `PLAYWRIGHT_FIREFOX_EXECUTABLE` select compatible installed engines. Worker caches under Autowork are reused when present. CI builds both profiles and runs `FECIM_WEB_SUITE=shared npm run test:e2e` against each, discovering new feature-owned specs automatically. The existing embedded Hysteresis file fixture remains a separate lane; hosted-isolation cases run only in the hosted shared lane. Private test setup keeps an explicit `TMPDIR`, otherwise uses `RUNNER_TEMP` or the platform temporary root. CI also runs the separate exact native/Node/Firefox Go parity gate.

The route budget case measures **all actual HTTP assets** from cold navigation through the first paused signed Go-default Sine result and complete first table page, in three fresh contexts. Its server negotiates Brotli and adds 100 ms request RTT on one aggregate 10 Mb/s response schedule; no host networking changes. Gates: full route ≤2 MiB Brotli, each cold result ≤5 s; module/runtime/Worker ≤6 MiB raw, ≤1.5 MiB Brotli and ≤2 MiB gzip. Results include identity, asset totals, resource filter and per-run timing in `route-budget.json` attachments and `ROUTE_BUDGET` log records. These private-host results are not CDN, low-end-device or real-time guarantees.

See [WASM gate evidence](../docs/internals/testing/wasm-parity-gate.md#5-shared-react-route-qualification) and [production transport/parity instructions](../validation/wasm/README.md). DOM, keyboard/focus, narrow-width/200% text and automated accessibility checks **do not qualify Orca speech**. Real Orca + Firefox/Linux acceptance for hosted and authenticated local profiles remains the separate [manual gate](../docs/internals/superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md).

## Cloudflare Pages output

`wrangler.toml` still selects `./dist`, project `fecim`, and `public/_redirects` preserves direct Hysteresis, research, docs and module navigation. `npm run deploy` builds the hosted profile before the existing Wrangler command; deployment is an explicit external operation, not part of validation and was not performed here. A hosting CSP must permit the verified Blob Worker/runtime and Go WASM compilation described in the [Worker documentation](../validation/wasm/README.md#production-protocol-and-ownership).
