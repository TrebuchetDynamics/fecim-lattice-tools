# Fresh combined all-module web automated recheck

## Verdict and scope

**COMPLETE: required matrix commands passed, with the declared applicability skips below.** Race proof is **exhaustive sharded coverage**, not an invocation or PASS of the historical aggregate. The original aggregate remains **INCOMPLETE**; [the historical failed checkpoint](all-modules-web-checkpoint.md) is unchanged. A passing scoped automated matrix is not WEB-ALL goal qualification, a release, default switch or Fyne retirement authorization.

Tested source: `4a9ef955bad72e2b9b1fd06a1491ce1389bb2dbe`, branch `autowork/WEB-RECHECK`, 2026-10-29. Only this evidence document was added afterward. No production/test/fixture/generated tracked/CI/manifest/lock bytes, assertions, samples, tolerances, goldens or budgets changed. TDD: **N/A**, evidence-only. The document commit is not a newly tested scientific revision; VCS-independent asset inputs remain the tested source inventory.

Prerequisite reports were read in place: WEB-HYS-STATUS `23d6232e686654b3f8130e8278d32e9a94690498`, WEB-MNIST-WAIT `4dbf518cbc8537bd49f54e5a3a0405d698cce4dc`, and WEB-MANUAL-FIXTURES `d8cad4c193ed84e781b57e015bc510991b107aa5`. All three are ancestors of the tested revision. Tagged fixture Go files enter the build inventory; these are fresh assets after that producer, not transferred fixture-build hashes.

Authority: [approved recheck contract](../../tasks/WEB-FINAL-PLAN.md#web-recheck--complete-and-truthfully-classify-remaining-automated-proof), [original matrix](../../tasks/WEB-NEXT-PLAN.md#task-2--web-automation-combined-source-automated-checkpoint), [roadmap outcome](../superpowers/plans/2026-10-28-all-modules-web-roadmap.md#outcome-and-definition-of-done), [browser checks](../../../web/README.md#browser-and-scientific-checks), and [parity/retention instructions](../../../validation/wasm/README.md).

Real Orca + Firefox/Linux (both profiles): **NOT RUN**. Actual installed Icarus/Yosys qualification: **NOT RUN**. Automated accessibility, fake-tool coverage and the finite synthetic fixture contracts cannot substitute for those separate actual proofs. Fyne/core/CLI and launch defaults remain intact.

## Execution and private evidence

Private evidence is retained at **`E=$TMPDIR/evidence`** for WEB-RECHECK attempt 1; the worker report resolves the absolute path. Raw logs, discovery inventories, JSON results/events, exits, manifests, transcripts and the proof index are private, not committed. This is temporary task storage, not a permanent artifact service: retain it before host cleanup. Old unavailable temporary evidence was not required or reused.

The matrix ran 09:23:59–10:41:28 UTC; race shards ran 09:51:38–10:40:18 UTC. All substantive matrix commands ran serially. Each used `bash "$TMPDIR/check.sh" ID 'COMMAND'`, which invokes **`autowork-run timeout 1790 bash -c 'COMMAND'`** and records UTC start/end, exit, wall seconds and log SHA-256. Each race shard is its own runner command, with **`-timeout=29m`** inside that outer ceiling; the series has no single outer timeout. No `-short`, sample reduction, missing-case exclusion or benchmark substitution. Browser commands add only `--reporter=list,json` plus `PLAYWRIGHT_JSON_OUTPUT_FILE` for machine-readable evidence; no selection changes. Discovery uses the same profile/selection with `--list --reporter=json`.

Pinned existing tools: Go 1.26.10, Node 24.19.0, npm 11.17.0, locked Playwright 1.60.0; Chromium 152.0.7977.75 and compatible Firefox 150.0.2 (`firefox-1522`), headless on Linux 7.0.0-38-generic x86_64 / Ryzen 9 5900X / 24 logical CPUs. Existing versioned browser/FFmpeg cache was reused; no engine, optional science/tool/model/dataset/PDK/container acquisition. Locked `npm ci` passed using the shared versioned npm cache. `tools.json` retains exact tool paths/versions. Exports: `PLAYWRIGHT_BROWSERS_PATH="$AUTOWORK_CACHE/playwright-1.60.0"`, `FECIM_WASM_ARTIFACT_DIR="$E/wasm"`.

Worker-selected listeners used assigned ports: 20200/20201 static/shaped; 20210 embedded; 20212 rooted research; 20214 Crossbar/Comparison/Docs/EDA; 20216 Circuits/MNIST. Finite tagged fixtures reused 20200 only after browser servers stopped, with actual embedded assets and matching BuildID/UIID linker values as documented in [the fixture README](../../../workbench/server/testdata/orca/README.md#build-and-bind-the-candidate). Fixture values are conspicuously synthetic, not physics qualification.

## Source and asset identities

| Identity | Fresh value |
| --- | --- |
| Core/build | `a0336d6b65da243713117e0dd3d774a3df428a7b1108f6fdaab9952938b06b45` |
| UI | `b48773d3a0084abb5854c41c4ce693d0807788d381b2c67b94f7f170472a4c57` |
| API/OpenAPI | `42652ddb722d9df7cb8444930e9f53607fb51722cd2d28e503f88208850ec711` |
| Protocol / model / Go | `fecim-hysteresis/1` / `hysteresis-producers/1` / `go1.26.10` |
| Engine manifest SHA-256 | `ada321dc5c757fedc7124b0e54316a79e52c0eb4cd7d967e7e5637aaa0d28018` |
| Embedded local-manifest SHA-256 | `7c78f5ffb6a5f0e408d5ece4821e591aeeaed3860cb6ef5fa5f97163fc9bbd44` |
| Native binary SHA-256 | `4791f14ec1cd8d6217c56cf7c04ef768808cd505dfd19775b4756ae060779523` |
| Source inventory SHA-256 | `af5668eeb8766ed27d719b11d3cda63c40e51ce3854b6a650f62689a5e8dc429` |

Fresh fingerprint agrees with all engine manifests. `cmp` passed from `web/node_modules/.fecim/manifest.json` to hosted, local and embedded copies. Every declared engine asset and every local-manifest asset was independently hashed. All 53 local-profile files match embedded copies byte-for-byte; embedding adds two metadata files. `asset-inventory.json` retains all 45 hosted / 53 local / 55 embedded paths, lengths and SHA-256 values, including all frontend chunks/CSS/docs. Its SHA-256 is `c6c982f6d9509d88fa34b1dd812f814352bbb272f06cf159719c02965f1de29b`.

| Engine asset (identical across profiles) | Bytes | SHA-256 |
| --- | --- | --- |
| `hysteresis.wasm` | 5716918 | `f1b435dbcdc8c67673bbc0c661a1add2c9dfe3ce9b63334fa4e9feda2f8715d5` |
| `wasm_exec.js` | 16992 | `0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14` |
| `hysteresis-worker.mjs` | 8061 | `887ddd17b87bd533363c96fcad1c1c033bf79cc097b1daba8c5b04698c36705b` |
| `circuits.wasm` | 4899002 | `f80cabc99a5b467527b42bcbd3d12c688c9a35a06cf50311c1ec943ce2482afc` |
| `comparison.wasm` | 4602608 | `a154ee9b1de3f4e002567d257dde850ce67a623e6ec2a18401b52e3a621368ba` |
| `crossbar.wasm` | 4813635 | `0d24aa89f73a81edd32fdc58af8947a066328284f40c424abceeefaa65ed036c` |
| `docs.wasm` | 5920518 | `d19222ad1362a7d8d9e9a19b0b39ef45d679f8b41fada5cc099dd1c4b8482d93` |
| `research.wasm` | 5412886 | `3f4182ef60ba65bb1ae06ee4739575bd870edffeebf9ecb21e7e75a48bf47ba4` |

Module manifests share the core/Go identity above. Circuits, Comparison, Crossbar, Docs and research use `fecim-module/1` with models `circuits-device/1`, `scenario-estimates/2`, `crossbar-cpu/1`, `docs-reader/1`, `research-memory/1`. MNIST/EDA registrations are presentation-only (`mnist-native/1`, `eda-native/1`), not scientific WASM assets or hosted inference/generation qualification. Full descriptors and module identities are retained in `identity.json` and copied engine manifests.

Hosted entry `assets/index-D2R6pwgo.js`: 239713 bytes, SHA-256 `78237939ed637ccae91065e6dca5db50a85f5931428df9b46e27ab1a36c4ede1`. Local entry `assets/index-BDgkEfsD.js`: 237736 bytes, SHA-256 `634833b589f0ef486f9d3bec31a45621ac4be1824bfec1a83adf83a05932d8b6`. Local entry bytes also match the embedded stage.

Fresh `cd web && node src/local/build-native.mjs "$TMPDIR/fecim-web-recheck"` passed source/UI/API admission; the retained binary is independently hashed above. `native-version.txt` records Go 1.26.10, CGO=1, linux/amd64. Fixture-built browser binaries were separately built/cleaned by their tests; this retained binary is not asserted to be a saved copy of each fixture binary.

## Final matrix

Commands run from the repository root unless prefixed with `cd web`; wrapper/environment above apply. The full merged-tree check is runner-owned and was not duplicated here.

| Command | Actual result | Private log |
| --- | --- | --- |
| `cd web && npm ci && npm run generate:api && npm run check:docs && npm test && npm run typecheck && npm run build:hosted && npm run build:local` | PASS; 6 Node tests, 0 failures/skips; generated contracts/docs checked, both profiles rebuilt/staged | `build.log` |
| `git diff --exit-code -- workbench/moduleapi/openapi.json web/src/api/generated.ts workbench/modules/docs/corpus.json web/public/docs-manifest.json web/public/docs-content` | PASS; no tracked generated drift | `drift.log` |
| `env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -timeout=30m -v` | PASS; all 16 enabled top-level tests, 0 skips; 188.019s package result | `parity.log` |
| `cd web && PLAYWRIGHT_JSON_OUTPUT_FILE="$TMPDIR/evidence/hosted-results.json" FECIM_WEB_SUITE=shared npm run test:e2e -- --reporter=list,json` | PASS; 180 selected = 142 pass + 38 intentional skip | `hosted.log` |
| `cd web && PLAYWRIGHT_JSON_OUTPUT_FILE="$TMPDIR/evidence/local-results.json" FECIM_WEB_PROFILE=local FECIM_WEB_SUITE=shared npm run test:e2e -- --reporter=list,json` | PASS; 170 selected = 144 pass + 26 intentional skip | `local.log` |
| `cd web && PLAYWRIGHT_JSON_OUTPUT_FILE="$TMPDIR/evidence/embedded-results.json" npm run test:e2e -- test/e2e/hysteresis-local.spec.ts --reporter=list,json` | PASS; 10/10, all five projects, 0 skips; actual embedded local session/file workflows | `embedded.log` |
| `cd web && PLAYWRIGHT_JSON_OUTPUT_FILE="$TMPDIR/evidence/local-static-results.json" FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hosted-isolation.spec.ts --grep "complete real route transfer and cold Go sine result meet the production guardrails$" --project=desktop-chromium --reporter=list,json` | PASS; 1/1, desktop Chromium; static local route, not authenticated API latency | `local-static.log` |
| `env CGO_ENABLED=1 go test -race -tags=orca_fixture ./workbench/study ./workbench/server -run 'TestOrcaFixtureContract' -count=1` | PASS; seven finite synthetic server contract subcases; study adapter exercised there, study package has no separately named test | `fixture.log` |
| `go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric ./shared/viewmodel -count=1` | PASS; all three packages | `numerical.log` |
| `env CGO_ENABLED=1 xvfb-run -a go test ./validation -run 'TestPhysicsRegressionCurves\|TestSenseChainRegression_4x4' -count=1` | PASS; named physics/sense-chain regressions | `curves.log` |
| `go test ./scripts -count=1` | PASS; CI/build/discovery/drift guards | `scripts.log` |
| `env CGO_ENABLED=1 xvfb-run -a go test ./cmd/fecim-lattice-tools -run 'TestLivingGuidanceDoesNotCallFyneRetired\|TestDefaultRepoGraphExposesFynePackages\|TestGoFilesDoNotCarryLegacyFyneBuildTags\|TestServeCommand\|TestBrowserShutdownCancelsBeforeHTTPDrain' -count=1` | PASS; named CLI/serve/default-Fyne guards | `legacy.log` |
| `cd web && node src/build/build.mjs --fingerprint` | PASS; fresh source/UI/API identity above | `fingerprint.log` |
| `cmp web/node_modules/.fecim/manifest.json web/dist/engine/manifest.json && cmp web/node_modules/.fecim/manifest.json web/dist-local/engine/manifest.json && cmp web/node_modules/.fecim/manifest.json internal/browserui/assets/engine/manifest.json` | PASS; all three comparisons | `manifest-cmp.log` |
| `sha256sum web/dist/engine/manifest.json web/dist-local/engine/manifest.json internal/browserui/assets/engine/manifest.json` | PASS; identical manifest hashes | `manifest-sha.log` |
| `cd web && node src/local/build-native.mjs "$TMPDIR/fecim-web-recheck"` | PASS; fresh source-admitted CGO native binary | `native.log` |

Literal `|` regex alternation is escaped only for Markdown table rendering. Native non-verbose checks prove package outcomes and the selected assertions, not invented per-test timing/counts. Tagged fixture environment supplies `FECIM_ORCA_PORT`, absolute `FECIM_ORCA_ASSETS`, and matching quoted `GOFLAGS` linker identities; ordinary race checks do not enable those hooks.

## Browser inventory and skips

Discovery and final JSON inventories match exactly by profile/project/file/title, with no duplicate identities, retries, missing tests or unclassified failures. Combined: **361 unique profile/test/project cases = 297 pass + 64 intentional skip**. This is not a sum of retry attempts.

| Project | Hosted pass / skip (36 selected) | Local pass / skip (34 selected) | Embedded pass / skip (2 selected) |
| --- | --- | --- | --- |
| desktop-chromium | 31 / 5 | 29 / 5 | 2 / 0 |
| desktop-compact-chromium | 28 / 8 | 29 / 5 | 2 / 0 |
| desktop-firefox | 27 / 9 | 28 / 6 | 2 / 0 |
| mobile-chromium | 28 / 8 | 29 / 5 | 2 / 0 |
| mobile-small-chromium | 28 / 8 | 29 / 5 | 2 / 0 |

Separate local-static desktop-Chromium budget: 1 pass / 0 skip. Shared configuration excludes the separate embedded file in both profiles and hosted-isolation in local; those are discovery boundaries, not extra skipped cases. All counts/skip annotations/source locations and exact discovery/results hashes are in `browser-accounting.json`.

| Lane | Existing skip reason | Count |
| --- | --- | --- |
| hosted | one hosted controlled measurement | 4 |
| hosted | native response-loss ownership boundary | 5 |
| hosted | hosted shaped cold-route measurement, native covered by real service workflow | 4 |
| hosted | mobile layout contract | 3 |
| hosted | desktop layout contract | 3 |
| hosted | one controlled measurement per build profile | 4 |
| hosted | native filesystem conflict | 5 |
| hosted | native response delay | 5 |
| hosted | native save responses | 5 |
| local | hosted privacy boundary | 5 |
| local | one hosted controlled measurement | 5 |
| local | hosted shaped cold-route measurement, native covered by real service workflow | 5 |
| local | mobile layout contract | 3 |
| local | desktop layout contract | 3 |
| local | hosted static Worker contract | 5 |

The repaired late-Apply case passed all five projects in both shared lanes. MNIST passed all five in both lanes, including real authenticated-local science/refreshed-result assertions. Individual success does not replace the complete-lane counts above.

## Exhaustive bounded race coverage

Discovery commands: `go list ./workbench/server ./workbench/moduleapi/... ./workbench/modules/... ./workbench/edaexport` and `go test -list '^(Test|Example)' ./shared/neural`. Non-neural packages also retain their own top-level Test/Example discovery. Results: **13 packages + 226 neural Tests, 0 neural Examples; 239 separate commands; 262 unique top-level cases** (36 non-neural + 226 neural).

Exact command templates (each name/package comes from the complete retained discovery, not a chosen subset):

```bash
env CGO_ENABLED=1 go test -race PACKAGE -count=1 -timeout=29m -json
env CGO_ENABLED=1 go test -race ./shared/neural -run '^EXACT_NAME$' -count=1 -timeout=29m -json
```

| Non-neural package (`fecim-lattice-tools/` prefix) | Top-level cases | Package terminal / exit |
| --- | --- | --- |
| workbench/server | 4 | pass / 0 |
| workbench/moduleapi | 1 | pass / 0 |
| workbench/moduleapi/cmd/generate | 0 | skip / 0 |
| workbench/moduleapi/httpapi | 0 | skip / 0 |
| workbench/moduleapi/native | 0 | skip / 0 |
| workbench/moduleapi/research | 2 | pass / 0 |
| workbench/modules/circuits | 5 | pass / 0 |
| workbench/modules/comparison | 3 | pass / 0 |
| workbench/modules/crossbar | 6 | pass / 0 |
| workbench/modules/docs | 5 | pass / 0 |
| workbench/modules/eda | 2 | pass / 0 |
| workbench/modules/mnist | 4 | pass / 0 |
| workbench/edaexport | 4 | pass / 0 |

Union reconciliation: 239/239 recorded shard exits; 262/262 unique top-level run events; 262/262 matching top-level terminal events. No duplicate planned/run/terminal identity or unexpected top-level test. Shard command outcomes: `{"pass": 239}`; test terminals are **261 pass + 1 existing opt-in skip**. The longest shard was `TestProcessVariation_M3_NOISE_03_MonteCarlo`, 861 wall seconds including runner/build overhead, below its independent ceiling. Three packages have no test files and report package `skip`, not scientific assertion passes. The complete ordered neural list, exact commands, per-case exits/seconds/log hashes, all subtest events, skips and package terminal events are retained in `neural-discovery.txt`, `race-plan.json`, `race-*.log`, `matrix-checks.jsonl`, and `race-accounting.json`.

Existing race skips (package/test and original reason retained, never counted as assertion passes):

| Test | Reason/output |
| --- | --- |
| `fecim-lattice-tools/shared/neural::TestLatency_Determinism` | performance_latency_test.go:220: Skipping latency determinism test unless FECIM_PERF_DETERMINISM=1 (too environment-sensitive for default CI) |

Every shard completed below its independent ceiling, with a matching test/package outcome and no missing required-data result. This establishes the declared exhaustive **sharded** proof only; it never relabels the former single aggregate PASS.

## Exact-byte parity and supported retention

All 16 enabled parity tests passed with no skips, including dependency boundaries, the three module science producers, Hysteresis production/lifecycle/budget and incomplete-transcript rejection, shared hosted boundary, research, original example/numerical-result validation, first-difference and detached-descendant cleanup/helper. Complete test names are in `module-measurements.json`; `parity.log` retains full verbose assertions/measurements.

Crossbar: 8 scientific profiles, native/Node/Firefox exact bytes; cancellation 6 interruptions / 2 completion races. Circuits: 48 scenarios (1/2/4/8 × Ideal/TierA/TierB × FeCIM/HZO/BTO/PZT), 5 scientific records each; native measured CLI and Worker bytes agree; cancellation 35 interruptions / 13 completion races. Comparison: 19 production operations; canonical science/numeric table bytes agree and stale requests reject. Research: eight successful finite 16/32-row runs and authoritative cancellation of 8; original native serializer bytes agree. The separate native-only original 64-row example remains unchanged.

These four module helpers delete temporary transcripts on test exit: their **post-run transcript lengths/hashes are unavailable**, not invented from logs. In-test original-byte comparisons passed. Their fixture build labels (`crossbar-parity`, `circuits-parity`, `comparison-production-parity`, `module-parity`) are not the packaging fingerprint; fixture module bytes are not asserted equal to packaged module bytes. Final packaged identities are independently bound above.

Supported retained runs under `E/wasm`: each row applies separately to native, Node and Firefox `*-results.json`; all three independently hashed byte-for-byte equal.

| Retained directory | Bytes per original transcript | SHA-256 (each of all three) |
| --- | --- | --- |
| `hysteresis-budgets-2844948725` | 2346420 | `af807379e3f63ae0380d634e0908890163f494ecc0d9d3ddcac495ec65e5f459` |
| `hysteresis-lifecycle-3425181328` | 2346420 | `af807379e3f63ae0380d634e0908890163f494ecc0d9d3ddcac495ec65e5f459` |
| `hysteresis-parity-3991317178` | 2346420 | `af807379e3f63ae0380d634e0908890163f494ecc0d9d3ddcac495ec65e5f459` |
| `parity-2275442054` | 21972 | `5ab80297901823c5d9131c87b31433eef893134f23be73058b058fd87c4189fa` |

`transcripts.json` additionally binds exact native/transport input bytes and the original persisted native example report, which matches its retained native output. Hysteresis transport input includes the Go-defined cold schedule without rewriting original request strings. Lifecycle covers real cancellation, stale/unsupported requests, pending termination/recreation, same-turn/retired sends and response/hash/asset rejection. Complete target-write/Preisach calibration, finite diagnostics, signed rows and unavailable legacy protocol remain required; bounded LK advance/cancel is not claimed completed convergence.

## Measurements and unchanged admission limits

No budget, resource filter, golden or source fixture changed. MiB = 1024² bytes. Worker totals retain ≤6 MiB raw / ≤1.5 MiB Brotli / ≤2 MiB gzip; applicable complete route budgets retain ≤2 MiB Brotli and each cold sample ≤5 s. Cold-route checks retain three fresh contexts, cache:no-store, +100 ms request RTT and one aggregate 10,000,000-bit/s server-owned response schedule. Compression remains harness-specific: Hysteresis sums separately Brotli-11/gzip-9-compressed assets; Crossbar/Circuits fixture payload gates compress concatenated module/runtime/driver bytes with Node defaults (`validation/wasm/{crossbar,circuits}_harness.mjs:40–44`); Comparison sums separately compressed assets with Node defaults (`web/test/e2e/comparison.spec.ts:302–305`). These methods are not normalized into one measurement. No host network configuration changed.

| Measured seam | Actual observations | Unchanged scope/filter |
| --- | --- | --- |
| Hysteresis production Worker | raw/Brotli/gzip 5741971 / 1172945 / 1583002 bytes; Firefox cold 1571, 1584, 1588 ms | client/Worker/module/runtime; driver counted twice in cold transfer; excludes empty pre-marker HTML; signed 64-step cycle, pause, first 32 rows |
| Comparison hosted complete route | Worker raw/Brotli/gzip 4627661 / 948090 / 1291263 bytes; route 1022222 Brotli bytes each; 2111.168, 2117.932, 2112.480 ms | existing complete-route filter, three fresh Chromium contexts |
| Crossbar hosted complete route | 1074037 Brotli bytes each; 2189.603, 2188.435, 2164.811 ms | existing resource filter preserved, including Blob resource entries |
| Hysteresis hosted complete route | 1251827 Brotli bytes each; 2252.123, 2248.676, 2253.905 ms | all actual route assets through paused signed result and first table page |
| Hysteresis local-static complete route | 1252241 Brotli bytes each; 2281.948, 2265.876, 2250.464 ms | all actual route assets; local-static includes adapter/static /api/v1/roots, not real API processing |
| Crossbar parity Worker | raw/Brotli/gzip 4838632 / 998802 / 1354730 bytes | fixture identity, payload only; not a route measurement |
| Circuits parity Worker | raw/Brotli/gzip 4924005 / 1017484 / 1382143 bytes | fixture identity, payload only; no full-route timing emitted |
| Circuits per-scenario replay | native max 95 ms / Go heap 6667512 bytes; Node max 177.663 ms / heap 4098456 / WASM memory 12058624 bytes; Firefox max 1157 ms / heap 4099280 bytes | 48 scenarios per runtime; ≤15000 ms and ≤128 MiB at observed responses; not peak heap or process RSS |

Hysteresis budget maxima: dispatch 288 ms versus 1000 ms; cancellation 11 ms versus 500 ms. Full admission/per-operation inputs, lifecycle/cold records, resource lists and exact byte filters remain in `measurements-budgets.json`, `module-measurements.json` and `route-measurements.json`, not extrapolated to other modules.

Caps remain one session job, 256 core/program steps per advance, one analysis work unit, 100000 samples, 2000 plot points, 32 table rows, 128 programming domains, FORC 64 reversals/256 resolution, 1 MiB request envelope. Research remains ≤8 MiB bundle / ≤1 MiB YAML / ≤256 points / ≤32×32 arrays. Crossbar hosted/local remains 32×32/128×128; Circuits hosted 8×8/native choices through 128, CPU only. No new Docs cold-route/search timing, Circuits full-route latency or MNIST/EDA complete-route budget is emitted; payload/heap/core timing does not establish those missing measurements. MNIST/EDA science remains authenticated-local only. No CDN, low-end-device, real-time, hardware or human speech claim follows.

## Integrity, readback and remaining ownership

| Retained matrix evidence | SHA-256 |
| --- | --- |
| `matrix-checks.jsonl` | `1fbe7f5b10c7e6c9e11871cccce6f12c467c6f4a3b5f50258d40a5bbfa25c4fb` |
| `race-packages.txt` | `d8b2419d2d5a2b71d060080679ac0b0abd9d6cbee64b4270deb95a9670d3e091` |
| `neural-discovery.txt` | `f09c4783bcfd718f8b876ffca3622989b5fe2bcb3688a11217dbf4e977167d69` |
| `race-plan.json` | `4f6b394f4b2410056632fe6b54dcffdc52d16750ca666c0522902230ba850a4c` |
| `race-accounting.json` | `611dcc902613096af9445620cf3cdb8fa58e0e0c0abac71f3f19028c2d13527a` |
| `browser-accounting.json` | `fda6f03a6ffdc2f47189678026421dd9735a027535e9e6c781dc27505b9b0268` |
| `transcripts.json` | `36dda5167c6bacfad003ca894586cde04a1f956f6d3c220ecc4c814a135ae08b` |
| `route-measurements.json` | `fde26725ec3f479fe2934cf4fb4679e1532cdb56e4e33c855675c84d411d2de0` |
| `proof-index.json` | `51e0e3f9a15372e9a4cda3a82d70bcecbf5470fe8b07fa0f0ac662eefc9dcc48` |

`proof-index.json` freezes private matrix files (including all command logs, JSON events/results/discovery, manifests and supported transcripts) before document-only checks; it does not recursively hash itself or the append-only live `checks.jsonl`. `matrix-checks.jsonl` is the frozen exit/command/log-hash record. Recheck script copies are retained so union/hash accounting can be reproduced, not substituted for the original test outputs.

Scoped Docs Sync Maintain affects only this new evidence owner. Source/asset/transcript/hash/count/skip/command/local-link/anchor readback, repeat-read byte stability, bundled docs checks and `git diff --check` are recorded in the worker report. The docs role checker has no role manifest in this checkout; no repository-wide manifest coverage is claimed. Shared task records, policy and historical receipts are untouched; coordinator Docs Maintainer owns their closeout synchronization. The merged full gate remains the runner’s responsibility. Actual Orca/HYS/manual-fixture applicability and installed-tool qualification remain with their existing owners, never inferred from task closure. Any later scientific/UI/API/source change needs affected rebuilds and fresh proof; this checkpoint is bound to the exact revision/inventories above.
