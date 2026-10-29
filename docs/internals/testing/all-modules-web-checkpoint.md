# Combined all-module web automated checkpoint

## Verdict and scope

**FAILED / INCOMPLETE — not a combined automated qualification.** This checkpoint records the actual candidate checks on 2026-10-29. Both shared-browser profiles have failures. The aggregate race command has no terminal `shared/neural` result. Passing parity, individual cases and historical producer reports do not close these gaps.

Tested source revision: `77e23369ab729e1657ab52624a830a21f4f51462`, branch `autowork/WEB-AUTOMATION`. The checkout remained byte-identical to that baseline throughout execution, including isolated reruns. Only this document was then added. No code, assertion, threshold, golden, fixture, generated tracked file, manifest, lock or CI file was changed. TDD: N/A, evidence-only documentation. The document commit is not a newly tested scientific revision; the original example report retains the tested revision above.

Authority: [approved closeout plan](../../tasks/WEB-NEXT-PLAN.md#task-2--web-automation-combined-source-automated-checkpoint), [roadmap outcome](../superpowers/plans/2026-10-28-all-modules-web-roadmap.md#outcome-and-definition-of-done), [ADR required acceptance](../adr/0005-unified-react-hosted-and-local.md#required-acceptance), [browser checks](../../../web/README.md#browser-and-scientific-checks) and [parity instructions](../../../validation/wasm/README.md). This is evidence for coordinator/final-planner reconciliation, not a task receipt, WEB-ALL goal claim, release, default switch or Fyne retirement authorization.

Real Orca + Firefox/Linux qualification for hosted and authenticated local workflows: **NOT RUN**. Real installed Icarus/Yosys qualification: **NOT RUN**. Automated accessibility and private process fixtures do not substitute for either.

## Execution and private evidence

Commands ran serially, including parity and every Playwright invocation. Assigned loopback ports were 20200/20201 for static/shaped servers, 20210 for the embedded fixture, 20212 for rooted research, 20214 for Crossbar/Comparison/Docs/EDA and 20216 for Circuits/MNIST fixtures. No unrelated service was reused. Evidence lives under **`E=$TMPDIR/evidence`** for `WEB-AUTOMATION/attempt-1`; the worker report resolves the private absolute location. Logs, binaries, manifests, transcripts and browser traces are not committed. Temporary storage is not a published or permanent artifact service; retain it before host cleanup if needed.

Each substantive check used `bash "$TMPDIR/check.sh" ID 'COMMAND'`. That private wrapper exports `PLAYWRIGHT_BROWSERS_PATH="$AUTOWORK_CACHE/playwright-1.60.0"`, `FECIM_WASM_ARTIFACT_DIR="$E/wasm"` and the versioned npm cache. It executes `autowork-run timeout 1790 bash -c 'COMMAND'`, records UTC start/end, exit and log SHA-256 in `checks.jsonl`. The 1790-second outer safety limit respects the runner's default 1800-second command ceiling. Go's aggregate timeout was explicitly 30 minutes, not its default ten minutes. Neither limit changes a product budget.

Build began at 06:26:20Z; parity ran 06:26:46–06:29:52Z; final shared lanes ran hosted 06:34:44–06:43:42Z and local 06:44:07–06:55:51Z. Race ran 06:59:53–07:29:43Z. Fingerprint/manifest/native identity checks followed at 07:40:09–07:40:11Z. Exact command records are private; the final commands and outcomes follow below.

Tools actually used: Linux `7.0.0-38-generic`, x86_64; AMD Ryzen 9 5900X, 24 logical CPUs; Go `go1.26.10 linux/amd64`; Node `v24.19.0`; npm `11.17.0`; Playwright `1.60.0`; system Chromium `152.0.7977.75`; compatible Playwright Firefox `150.0.2` (cache build `firefox-1522`). Browser checks were headless. Locked `npm ci` succeeded. The first hosted attempt could not create normal test pages because Playwright FFmpeg was missing (165 setup failures, 12 skips, 3 custom-context budget passes). `node web/node_modules/playwright/cli.js install ffmpeg` provisioned pinned FFmpeg v1011 in the existing versioned browser cache, then the complete hosted lane was rerun. This resolved prerequisite attempt is not the final browser result. Two initial anchored-grep rerun attempts selected no tests; corrected suffix-grep runs below selected all five intended cases. None of those empty selections counts as proof.

## Source and asset binding

Fresh `node src/build/build.mjs --fingerprint` and all emitted engine manifests agree:

| Identity | Value |
| --- | --- |
| Core/build | `c5647704ca9d297d0f3332799ae9ba3c3c2c0aa1b438377fb134108876fddd6f` |
| UI | `b48773d3a0084abb5854c41c4ce693d0807788d381b2c67b94f7f170472a4c57` |
| API/OpenAPI | `42652ddb722d9df7cb8444930e9f53607fb51722cd2d28e503f88208850ec711` |
| Hysteresis protocol/model | `fecim-hysteresis/1` / `hysteresis-producers/1` |
| Go identity | `go1.26.10` |
| Engine manifest SHA-256 | `ddce69013c8ddded295b77b6dd91103915b6c1559a43127ec9f18f0f71496562` |
| Embedded local-manifest SHA-256 | `3ab1b29a99daa26aa8a4a55a3f9f5f03d2cb566866ce1820d3ad350d820b8ac0` |

`cmp` passed between `web/node_modules/.fecim/manifest.json` and `web/dist/engine/manifest.json`, `web/dist-local/engine/manifest.json`, and `internal/browserui/assets/engine/manifest.json`. Every listed engine asset was independently hashed against each manifest. All 53 local-profile files also match their staged embedded copies byte-for-byte; embedding adds two metadata files. The private `asset-inventory.json` contains path, length and SHA-256 for all 45 hosted, 53 local and 55 embedded files. Its SHA-256 is `fa0edfa019aff117a1ca197c9f9525893b2e7352cad68cb1e4837e7c5f5e149c`.

| Engine asset (identical in all profiles) | Bytes | SHA-256 |
| --- | ---: | --- |
| `hysteresis.wasm` | 5716918 | `49a54c26756701c52530627518c60eabf05b4f10dd0375cbb6a86622a2f224bb` |
| `wasm_exec.js` | 16992 | `0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14` |
| `hysteresis-worker.mjs` | 8061 | `887ddd17b87bd533363c96fcad1c1c033bf79cc097b1daba8c5b04698c36705b` |
| `circuits.wasm` | 4899002 | `b869b2564e7f276c7f3fe7a19d0d070ba1ef7bd7626aa59309c33fe7bf6ffb9b` |
| `comparison.wasm` | 4602608 | `ca70464175ab718bdadf14db7779d224279c97e73a1b380b69dc9e49a48a230c` |
| `crossbar.wasm` | 4813635 | `f341f0392ab8c29f167df8457e4b92e85a1b339fda121cf62191e96542dd563e` |
| `docs.wasm` | 5920518 | `99baa2981b5d29735529367968394325e946fd2c5c0ac799898161d8ab65e59d` |
| `research.wasm` | 5412886 | `c2cb675815aaf99776c487fbb83ec2b73aa2a2a243cb568066a6ebbbd47eb51d` |

Module manifest identities share the core/build and Go identities above. Circuits, Comparison, Crossbar and Docs use protocol `fecim-module/1`, with models `circuits-device/1`, `scenario-estimates/2`, `crossbar-cpu/1` and `docs-reader/1`. Research uses `fecim-module/1` / `research-memory/1`. MNIST and EDA have presentation-only registrations, not scientific WASM assets: descriptor models are `mnist-native/1` and `eda-native/1`; their presentation sidecars carry the Hysteresis identity. Descriptor availability is not scientific qualification.

Hosted entry `assets/index-D2R6pwgo.js`: 239713 bytes, SHA-256 `78237939ed637ccae91065e6dca5db50a85f5931428df9b46e27ab1a36c4ede1`. Local/embedded entry `assets/index-BDgkEfsD.js`: 237736 bytes, SHA-256 `634833b589f0ef486f9d3bec31a45621ac4be1824bfec1a83adf83a05932d8b6`. The complete frontend chunk/CSS/docs hashes are in the inventory, not inferred from the UI fingerprint alone.

A fresh `cd web && node src/local/build-native.mjs "$E/fecim-local"` passed source/UI/API drift admission. Binary SHA-256: `82c804be290d32e52913588001ed16f7b785bda57cb5b541de0f1df5a2e9f69c`; `go version -m` records Go 1.26.10, CGO=1, linux/amd64. The embedded suite separately built its real binaries and cleaned its temporary roots/binaries after each project. The retained binary is an independent identity check, not a retained copy of each fixture binary.

## Final check matrix

Commands below run from the repository root unless prefixed with `cd web`. The wrapper and environment above apply. Earlier setup/empty-selection attempts are excluded from final counts.

| Command | Actual result | Private log |
| --- | --- | --- |
| `cd web && npm ci && npm run generate:api && npm run check:docs && npm test && npm run typecheck && npm run build:hosted && npm run build:local` | PASS; 6 Node tests, 0 failures/skips; both profiles rebuilt and local staged | `build.log` |
| `git diff --exit-code -- workbench/moduleapi/openapi.json web/src/api/generated.ts workbench/modules/docs/corpus.json web/public/docs-manifest.json web/public/docs-content` | PASS; no tracked generated drift | `drift.log` |
| `env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -timeout=30m -v` | PASS; 16 top-level tests, 0 skips, 185.249s | `parity.log` |
| `cd web && FECIM_WEB_SUITE=shared npm run test:e2e` | FAIL; 180 selected = 137 pass + 5 fail + 38 skip | `hosted.log` |
| `cd web && FECIM_WEB_PROFILE=local FECIM_WEB_SUITE=shared npm run test:e2e` | FAIL; 170 selected = 137 pass + 7 fail + 26 skip | `local.log` |
| `cd web && npm run test:e2e -- test/e2e/hysteresis-local.spec.ts` | PASS; 10 selected/passed, 0 skips, all five projects | `embedded.log` |
| `cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hosted-isolation.spec.ts --grep 'complete real route transfer and cold Go sine result meet the production guardrails$' --project=desktop-chromium` | PASS; 1 selected/passed, 0 skips; local static, not authenticated API latency | `local-static.log` |
| `env CGO_ENABLED=1 go test -race ./workbench/server ./workbench/moduleapi/... ./workbench/modules/... ./workbench/edaexport ./shared/neural -count=1 -timeout=30m` | INCOMPLETE; outer timeout exit 124 at 1790s; no `shared/neural` result | `race.log` |
| `go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric ./shared/viewmodel -count=1` | PASS; all three packages | `numerical.log` |
| `env CGO_ENABLED=1 xvfb-run -a go test ./validation -run 'TestPhysicsRegressionCurves\|TestSenseChainRegression_4x4' -count=1` | PASS; selected native regression package | `curves.log` |
| `go test ./scripts -count=1` | PASS; includes both-profile CI discovery/generation/drift guards | `scripts.log` |
| `env CGO_ENABLED=1 xvfb-run -a go test ./cmd/fecim-lattice-tools -run 'TestLivingGuidanceDoesNotCallFyneRetired\|TestDefaultRepoGraphExposesFynePackages\|TestGoFilesDoNotCarryLegacyFyneBuildTags\|TestServeCommand\|TestBrowserShutdownCancelsBeforeHTTPDrain' -count=1` | PASS; additional default-Fyne/CLI/serve guards | `legacy.log` |
| `cd web && node src/build/build.mjs --fingerprint` | PASS; identities above | `fingerprint.log` |
| `cmp web/node_modules/.fecim/manifest.json web/dist/engine/manifest.json && cmp web/node_modules/.fecim/manifest.json web/dist-local/engine/manifest.json && cmp web/node_modules/.fecim/manifest.json internal/browserui/assets/engine/manifest.json` | PASS | `manifest-cmp.log` |
| `sha256sum web/dist/engine/manifest.json web/dist-local/engine/manifest.json internal/browserui/assets/engine/manifest.json` | PASS; identical manifest hashes above | `manifest-sha.log` |

Escaped `\|` in Markdown table cells represents the literal regex `|` used in the actual commands. Native non-verbose logs record package outcomes, not invented per-test timing/counts. The race log has 10 successful package results and three `[no test files]` packages; server 54.548s, MNIST port 95.655s, no race diagnostic in captured output. There is no neural aggregate pass, active-test identification or final aggregate completion record. Focused passes cannot fill this gap. The coordinator owns the merged full check; this worker did not run it and claims no result for it.

### Browser inventory and skips

Five projects: desktop Chromium 1440×900, compact Chromium 1024×768, mobile Chromium 390×844, small-mobile Chromium 360×740 and desktop Firefox. Shared discovery covers all 11 feature/landing spec files, excluding the separate embedded fixture. Local shared also excludes the hosted-isolation file (10 identities), not a hidden passing substitute. `hosted-selection.json` / `local-selection.json` were checked against unique recorded project/file/line/title identities. No attempt or retry counts were summed into these totals.

| Project | Hosted pass/fail/skip (36 each) | Local pass/fail/skip (34 each) |
| --- | --- | --- |
| desktop-chromium | 30 / 1 / 5 | 28 / 1 / 5 |
| desktop-compact-chromium | 27 / 1 / 8 | 27 / 2 / 5 |
| mobile-chromium | 27 / 1 / 8 | 28 / 1 / 5 |
| mobile-small-chromium | 27 / 1 / 8 | 27 / 2 / 5 |
| desktop-firefox | 26 / 1 / 9 | 27 / 1 / 6 |

Every intentional skip follows existing test source:

- Hosted: Crossbar local response-loss case 5; three native Workbench filesystem/late-response/save cases 5 each; Comparison, Crossbar and Hysteresis single-project cold budgets 4 each; desktop landing layout 3 and mobile landing layout 3. Total 38.
- Local: hosted Comparison privacy case 5; hosted-only Comparison and Crossbar cold budgets 5 each; hosted static Worker/API mismatch case 5; desktop landing layout 3 and mobile landing layout 3. Total 26.
- Desktop landing layout runs only on the two Chromium desktop projects. Mobile layout runs only on the two mobile projects. Cold hosted budgets run only on desktop-chromium. These are applicability limits, not additional successful executions.

Private complete accounting hashes: `hosted-accounting.json` `7d91e75da72a87ba314f358f72489a28091aad7416bc0644fdd1ae97ffc0e9e0`; `local-accounting.json` `1edc282899af0a15599dbfb814e2fa7889f465d132eb779f2e6c75c229ef0569`.

### Failure and isolated-rerun evidence

All failures below occurred on the unchanged baseline revision before this document existed. They are unrelated to this documentation diff, but remain candidate acceptance gaps, not passing checks. No repair or assertion rebaselining was attempted.

- **Both profiles, all five projects:** `hysteresis.spec.ts:327`, “late Apply replies preserve newer input drafts”. At line 353, `getByRole('status').first()` expected `Settings/state applied` but received an empty string for 10 seconds. The locator resolved the shared research announcement paragraph. Source corroboration: [App.tsx](../../../web/src/App.tsx) has that preceding status region; [Hysteresis.tsx](../../../web/src/hysteresis/Hysteresis.tsx) has its own status region. This observation does not establish whether the later draft assertion would pass; it was not reached.
- Isolated commands: `cd web && FECIM_WEB_SUITE=shared npm run test:e2e -- test/e2e/hysteresis.spec.ts --grep 'late Apply replies preserve newer input drafts$'`, and the same with `FECIM_WEB_PROFILE=local`. Each selected only that case in all five projects, one worker, and reproduced 5/5 failures. Logs: `hosted-late-apply-rerun.log`, `local-late-apply-rerun.log`.
- **Local desktop-compact-chromium and mobile-small-chromium:** `mnist.spec.ts:64`, “MNIST real local science and hosted read-only boundary”. Compact initially failed inference deep equality at line 161. Small-mobile initially read missing `r.cases[0]` at line 172. The existing test waits for generic `succeeded` status before reading the next result; stale-result/timing behavior needs investigation, not a scientific mismatch conclusion from this checkpoint.
- Isolated command: `cd web && FECIM_WEB_PROFILE=local FECIM_WEB_SUITE=shared npm run test:e2e -- test/e2e/mnist.spec.ts --project=desktop-compact-chromium --project=mobile-small-chromium`. Compact still failed (missing `r.cases[0]`, line 172); small-mobile passed. Log: `local-mnist-rerun.log`. The suite remains failed; this isolated success is not a replacement aggregate.

Full screenshots, videos, traces and error contexts remain in `hosted-artifacts`, `local-artifacts`, `hosted-late-apply-rerun-artifacts`, `local-late-apply-rerun-artifacts` and `local-mnist-rerun-artifacts` below E. The original log paths preceded these private directory moves. Empty-selection/setup attempts are retained separately, never counted as tests that ran successfully.

## Exact-byte scientific evidence

`parity.log` records all 16 enabled test passes: Circuits and Crossbar dependency boundaries; production Circuits, Comparison and Crossbar parity; Hysteresis production parity, lifecycle, budgets and incomplete-transcript rejection; shared hosted dependency boundary; research parity; original example parity and numerical-result validation; first-difference helper; detached-descendant cleanup and its helper. No enabled gate skipped.

Original bytes were compared in-test, not parsed/reformatted for equality. The successful test results establish that those assertions executed. Crossbar, Circuits, Comparison and research use temporary-only transcripts; their **post-run transcript hashes/lengths are unavailable**. Complete verbose output and emitted measurements are retained instead. Their fixture build labels are `crossbar-parity`, `circuits-parity`, `comparison-production-parity` and `module-parity`; these are not the packaging fingerprint. The source revision is common, but fixture-compiled module bytes are not asserted equal to packaged module bytes. Final packaged identities were measured independently above.

- Crossbar: 8 successful scientific fixtures, four 2×3 profiles plus four 32×32 profiles; native/Node/Firefox exact bytes. Real cancellation observed 7 interruptions and 1 completion race, with previous science retained.
- Circuits: 48 scenarios (1/2/4/8 sizes × Ideal/TierA/TierB × FeCIM/HZO/BTO/PZT), 5 scientific result records per scenario. Native CLI/measured replay and Node/Firefox original bytes agree. Cancellation observed 37 interruptions and 11 completion races.
- Comparison: 19 operations; native production runner, canonical scientific bytes and original numeric-table replies agree with Node/Firefox; stale requests rejected. Empty/unavailable/unit-invalid science cannot satisfy the test.
- Research: admitted 16/32-row eight-point fixture; persisted native report, native production bridge/runner and Node/Firefox original report bytes agree. Eight successful finite runs are required; cancellation reported 8 cancelled. The separate original 64-row example remains unchanged and is not admitted as a hosted fixture.

Retained original Hysteresis/example outputs are independently hashed below. Each row applies separately to `native-results.json`, `node-results.json` and `firefox-results.json` in the named directory under `E/wasm`:

| Retained run directory | Bytes per original transcript | SHA-256 (all three equal) |
| --- | ---: | --- |
| `hysteresis-parity-1772332244` | 2346420 | `e5689d5f7df8cfa5594dd22dc7de860821baf7b2fcc9b1ea4281febf2d25f82f` |
| `hysteresis-lifecycle-2913090928` | 2346420 | `e5689d5f7df8cfa5594dd22dc7de860821baf7b2fcc9b1ea4281febf2d25f82f` |
| `hysteresis-budgets-2778301454` | 2346420 | `e5689d5f7df8cfa5594dd22dc7de860821baf7b2fcc9b1ea4281febf2d25f82f` |
| `parity-2740109373` (original example) | 21972 | `0bc0303b02382def2415a4dfa1ebd7ccf8267a71303f9ad079ce29d4577c8852` |

Each retained Hysteresis directory has `native-input.json` 69038 bytes / SHA-256 `5a12a7f934d067b8e77640ef9ab62ffbcdc6fd89a607ede56dd7b976576a7bd3`, and transport `input.json` 80560 bytes / `655bcc677c1d76071d450fd9c3685cb006fb866c360fa31f6883ec753e63baa9`. The latter adds the Go-defined cold schedule; the original request strings are preserved. Original example `input.json`: 993 bytes / `e8c7e15afaea1b8d82d130c84efe14f153eed0c81d6941854a2241a4d900cfb8`. Its persisted `native-project/reports/results.json` matches the retained native output. Input/output hash inventory `transcripts.json`: SHA-256 `c94df9adb7c8849ad3bf9541c7a3b8333833d8869976c6a9b6cf43c0aa3fb0e0`.

Hysteresis lifecycle passed real cancel, unsupported/stale-request rejection, pending termination, recreation, same-turn retired sends, retired/stale-response rejection, asset/build hash drift and missing-asset checks. Programming/calibration completion, signed rows, finite diagnostics and unavailable legacy protocol checks remain required. LK bounded advance/cancel is not a completed convergence claim.

## Measurements and admission limits

No thresholds or resource filters changed. MiB means 1024² bytes. Existing Worker payload limits are 6 MiB raw, 1.5 MiB Brotli and 2 MiB gzip. Route gates are ≤2 MiB Brotli and each cold result ≤5 seconds, not a claim about all module routes. Budgets use their existing per-asset compression and server-owned shaping, not host network changes.

| Existing measured seam | Actual observations | Limits/filter |
| --- | --- | --- |
| Hysteresis production Worker | raw/Brotli/gzip 5741971 / 1172835 / 1583004 bytes; cold Firefox 1578, 1562, 1575 ms | module/runtime/Worker; cold includes driver twice, excludes empty pre-marker HTML; 64-step signed cycle, pause, first 32 rows |
| Hysteresis complete hosted route | 1251723 Brotli bytes each; 2264.630, 2249.971, 2257.367 ms | all actual route assets through paused signed result/first page |
| Hysteresis complete local-static route | 1252137 Brotli bytes each; 2277.629, 2264.939, 2245.722 ms | includes local adapter and static `/api/v1/roots` request; not real API processing latency |
| Crossbar parity Worker | raw/Brotli/gzip 4838632 / 998802 / 1354730 bytes | fixture identity; module/runtime/Worker payload, not route cost |
| Crossbar complete hosted route | 1075978 Brotli bytes each; 2195.607, 2198.313, 2187.547 ms | existing harness resource filter retained, including its Blob resource entries |
| Circuits parity Worker | raw/Brotli/gzip 4924005 / 1017484 / 1382143 bytes | fixture identity; no full-route latency measurement emitted |
| Circuits per-scenario replay | native max 95 ms / observed Go heap 6422992 bytes; Node max 172.707 ms / heap 4099384 / committed WASM memory 12058624 bytes; Firefox max 1213 ms / heap 4089928 bytes | 48 profiles per runtime; ≤15000 ms and ≤128 MiB at observed responses; not peak heap or process RSS |
| Comparison complete hosted route | Worker raw/Brotli/gzip 4627661 / 948714 / 1291266 bytes; route 1022852 Brotli bytes each; 2108.285, 2108.299, 2117.604 ms | existing complete-route filter and three fresh contexts |

Cold route samples used desktop Chromium, three fresh contexts, Brotli responses, +100 ms request RTT and one aggregate 10,000,000-bit/s schedule. Hysteresis Worker cold samples instead used Firefox. Full resource lists and identities remain in `ROUTE_BUDGET`, `CROSSBAR_ROUTE_BUDGET`, `COMPARISON_ROUTE_BUDGET` log records and Hysteresis measurement files. They are not CDN, low-end-device, real-time or human accessibility guarantees.

Hysteresis budget maxima: dispatch 290 ms versus 1000 ms admission; cancellation 12 ms versus 500 ms. Measured caps remain one session job, 256 core/program steps per advance, one analysis work unit, 100000 retained samples, 2000 plot points, 32 table rows, 128 programming domains, FORC 64 reversals/256 resolution, 1 MiB request envelope. Full per-operation inputs/timings are retained in `measurements-budgets.json`, not extrapolated to other modules.

Current descriptors retain hosted/local Crossbar caps 32×32/128×128 and Circuits hosted 8×8/native size choices through 128, CPU only. Research admission remains ≤8 MiB bundle, ≤1 MiB YAML, ≤256 points and ≤32×32 arrays. Docs manifest reports 13 reviewed sources / 62684 decoded bytes, digest `591706a656c4139dfacdef38be5d424858aaf4b96bf0d734b42c88de987e050d`; lazy reading/search and local rooted history tests passed in both profiles. This run emits no new Docs cold-route/search timing or MNIST/EDA complete-route budget. No missing route measurement is inferred from payload size, heap or core timing. MNIST/EDA scientific execution remains authenticated-local only; hosted tests cover read-only help/import boundaries, not inference/generation.

## Evidence integrity and next action

Selected log SHA-256 values (all below E):

| Log | SHA-256 |
| --- | --- |
| `build.log` | `ff1f62ff65c0d6f24425e700c8e429dfa5eb068139e1c5a288f2f1feb4ab56f7` |
| `parity.log` | `3206ef888164fd3a1f8df61f4b6a6bdedee037d5bf26b5c8cc30f72f5185491b` |
| `hosted.log` | `84fbec239c16be73ad130dc749ad7fd37d9da6881c6cafc858fe2cc8415406cc` |
| `local.log` | `9b66a3099d859c4d0bb4a9ce2b39700d15bfd1784c94ae1a753af309648ab2e0` |
| `embedded.log` | `e3b3bfc5651bf81f28b48120395d8585b080c859bc62f7cf12a3100bf88aeefe` |
| `local-static.log` | `0c972d031322ac8e9ee32e6c7974133115954d755782161c6b33b6199ab82e4e` |
| `race.log` | `a3e22068769ff308e17147e14733cf3be5a018bce57628acc39c0fc0e99068ae` |
| `hosted-late-apply-rerun.log` | `07e05aaa61c8391923171628d6efd7a48ee85ad20c7ba06e45301a8eff2a739e` |
| `local-late-apply-rerun.log` | `31e6f309b66a56c5102eabcf31022bc60e4dad9034242c55233eb0a25efd6780` |
| `local-mnist-rerun.log` | `501907c3219e20d063f3ef0571bf906f7805584330bf9deba8466de1a2b6219d` |

Historical reports were read in place for WEB-WORKBENCH, DOCS, CROSSBAR, CIRCUITS, MNIST, EDA, Comparison attempts 1/2 and Availability attempts 1/2. Their candidate hashes and historical proof remain in the [producer matrix](../../tasks/WEB-NEXT-PLAN.md#delivered-module-and-proof-matrix). Final Availability attempt 2 is `43ddcf512cec74578e729dd8fd4ce9ee336f7e4d`. Its earlier anchor failures are not present in these current shared runs. Historical MNIST focused race results still do not establish the missing aggregate. None of those reports was rewritten or promoted to combined proof.

Next action for the same coordinator/final planner: reconcile the two browser failure families and the incomplete aggregate race result using the exact revision/logs above. Keep this candidate unqualified. Any repair that changes source/UI/API/assets requires affected rebuilds and fresh combined evidence; preserve this checkpoint as historical evidence, not a transferable pass. The merged-tree full check, separate human Orca gate and installed-tool gate remain with their existing owners. No task records, shared history or qualification policy were edited.

Scoped Docs Sync Maintain is limited to this new evidence owner. `python3 "$TMPDIR/check-checkpoint.py"` passed source/asset/transcript readback, unique browser inventory reconciliation and all eight local links/anchors. `git diff --check` passed. Bundled `docs.py stale .` and `docs.py check .` exited zero but reported no `docs/roles.json`; no manifest-based coverage is claimed. Fresh read-only review found no important evidence-document issue; its port-inventory correction was source-verified and applied. Assigned-port listener inspection after all suites found none. The coordinator's Docs Maintainer owns closeout synchronization of shared status/history after integration.
