# WASM parity gate — historical example and production qualification

**Current scope:** [Shared React route qualification](#5-shared-react-route-qualification) adds actual hosted/local asset-profile, workflow, isolation and route-budget evidence. [Production hysteresis qualification](#4-production-hysteresis-worker-qualification) retains the preceding CORE/PROGRAM/ANALYSIS Worker checkpoint. Sections 1–3 and the historical verification retain the earlier **example-only** receipt; their unimplemented/next-scope statements describe those checkpoints, not current candidate status. [Authenticated embedded local qualification](#6-authenticated-embedded-local-profile) records the additive serve/file boundary separately. No deployment or Orca qualification is inferred.

Historical scope: prerequisites only, before any React feature implementation. Related: [gate/run instructions](../../../validation/wasm/README.md), [one-app decision](../adr/0005-unified-react-hosted-and-local.md), [package feasibility](../architecture/browser-wasm-feasibility.md) and [HZO first-slice plan](../superpowers/plans/2026-10-25-hosted-hzo-first-slice.md). No implementation of that slice or commit is authorized by this receipt.

## 1. Exact three-runtime parity

The implemented gate is `validation/wasm/parity_test.go::TestExampleResultsJSONNativeWASM`. Native uses the existing `study.Run` and reads its actual persisted report, with eight successful points and four finite numerical metrics per point required. The private Go WASM entry point uses the resolved same input, shared Expand/ID/Evaluate/Analyze/WriteJSON and identical repository revision. Node executes that module directly; headless Firefox executes it in a real dedicated Worker. TypeScript/JS performs no scientific arithmetic or JSON number normalization.

Observed successful runs used native CGO disabled and enabled. Representative final retained receipt: `artifacts/wasm-parity/parity-702123619/`, source HEAD `c5e4504f22645ef03835641b448cdbd99ad38d41`, plus the uncommitted validation code. No model/preset/equation was changed. Subsequent source/serializer/evaluator changes invalidate this receipt and require rerunning the gate.

| Output | SHA-256 |
|---|---|
| native results.json | `23aed010a1af37767d580c7a2eec8c2840efb772948ff04a53c5b2b57e7b8324` |
| Node WASM results.json | `23aed010a1af37767d580c7a2eec8c2840efb772948ff04a53c5b2b57e7b8324` |
| Firefox Worker WASM results.json | `23aed010a1af37767d580c7a2eec8c2840efb772948ff04a53c5b2b57e7b8324` |

**Result: byte-identical; no differing field.** On mismatch the test retains original outputs when the artifact directory is configured, and diagnoses the first field with exact numeric lexemes. Pure whitespace/serialization differences also fail. Six additional Firefox calculations check repeatability.

Tools: Go 1.26.10, Node 24.19.0, Playwright 1.60.0, explicitly selected installed Firefox 146.0.1; Linux x86-64, AMD Ryzen 9 5900X, 24 logical CPUs. This records the tested combination, not a compatibility promise for untested browser versions.

### RED / GREEN evidence

- Initial RED: the explicit gate generated the native report, then failed because `validation/wasm/testdata/runner` did not exist. GREEN: the actual Go WASM module ran in Node/Firefox and compared equal; no physics production code changed.
- Diagnostic RED: root-object versus root-array mismatch was incorrectly labelled serialization. GREEN: `TestFirstDifference` reports `<root>` and preserves deterministic field paths/numeric lexemes.
- Independent review found that identical failed studies could satisfy raw equality. RED: numerical-coverage guard was absent; GREEN: `TestExampleRequiresNumericalResults` rejects failed/empty/missing-metric example reports. The integration test checks the real persisted numerical report before comparing it.
- Review also found timeout could kill Node but leave detached Firefox descendants. Behavioral RED: `TestParityCommandKillsDetachedDescendants` observed the detached child survive. GREEN: Linux process-tree ownership kills only the command's owned descendants/groups; the focused regression and its race run passed. No user/browser process outside that tree is targeted.

The integration test is opt-in because it needs browser tooling. Ordinary full-suite execution is not a parity receipt: CI/release must run the explicit command in the README. This change does not claim a CI lane was already wired or a browser production API was delivered. The private fixture runner is validation-only; shared production input validation and production Worker admission/cancellation still need implementation.

## 2. Native GPU/CPU audit

Observed call path: standard study orchestration binds `fecim.Evaluate`; that evaluator uses scalar shared peripheral/system calculations and Go design validation. `experiment.Run` is a generic evaluator mechanism, but the production workbench route does not select a GPU evaluator.

Commands inspected current source and enumerated actual dependencies with both settings:

```bash
CGO_ENABLED=0 go list -deps ./workbench/fecim
CGO_ENABLED=1 go list -deps ./workbench/fecim
```

Neither dependency graph contains `github.com/vulkan-go/vulkan`, `shared/compute`, `gpuperiph` or `go-gl`. Source inspection of `workbench/fecim/evaluator.go`, `workbench/study/study.go`, `shared/system` and `shared/peripherals` found no GPU execution branch. Retained dependency lists are alongside the representative receipt as `evaluator-deps-cgo-0.txt` and `evaluator-deps-cgo-1.txt`.

**Conclusion:** the currently supported workbench evaluator is CPU-only even in a CGO-enabled build. GPU-vs-CPU bit equality is **not applicable to that path**, not an unrun GPU test disguised as a pass. Other native modules contain Vulkan/GPU implementations, such as shared crossbar acceleration and gpuperiph; this audit does not certify those kernels equal to CPU.

Recommendation: retain CPU-only workbench evaluation for now. If a future evaluator selects GPU, first compare actual kernels; if bit-identical equivalence is not proved, backend/precision/kernel version must become behavior-affecting identity input with an evaluator/schema compatibility decision. Do not silently reuse CPU run IDs for different GPU arithmetic. No backend/model/identity change is needed or made in this pass.

## 3. Measured size and browser time

The measured module is the optimized **validation example module**, not a shipped HZO UI bundle. The actual HZO solver adapter may change its size and must be remeasured. `gzip` uses level 9; Brotli uses quality 11, computed over emitted bytes separately for module/runtime/worker shims.

| Payload | Raw bytes | gzip bytes | Brotli bytes |
|---|---:|---:|---:|
| Go WASM | 4,417,480 | 1,230,933 | 909,398 |
| Go runtime shim | 16,992 | 4,348 | 3,752 |
| Worker shim | 757 | 369 | 321 |
| Total | 4,435,229 | 1,235,650 | 913,471 |

Headless Firefox measurements: **118 ms** from Worker creation to ready; first eight-point calculation **38 ms**; subsequent **22, 23, 25, 24, 24, 24 ms**, median **24 ms**. Node startup was 30.1 ms and its calculation 19.7 ms. Earlier repeated runs varied (Worker-ready 194–200 ms, first example 40–47 ms), so a single best result is not a universal latency claim.

Conditions: uncompressed assets over loopback HTTP, cache disabled, one new Worker, seven calculations, no CPU/network throttling. The run interval measures the Go callback, including serialization, not a rendered React table. Compression was measured, not actually negotiated by that fixture server. These measurements do **not** establish public-hosted cold-load time, low-end-device performance, mobile support, peak memory, paint time or accessibility.

### Proposed budgets — for owner review, not asserted production SLOs

| Budget | Proposed limit | Current observation / next enforcement |
|---|---|---|
| WASM + Go runtime + Worker transfer | <= 1.5 MiB Brotli; <= 2 MiB gzip fallback | 0.87 MiB / 1.18 MiB; remeasure actual HZO adapter and enforce in CI after review |
| Raw module/runtime/Worker payload | <= 6 MiB | 4.23 MiB; not a memory-usage claim |
| Worker ready on documented unthrottled Linux reference | <= 1,000 ms | 118–200 ms; independent from network SLO |
| First example compute + serialization | <= 250 ms | 38–47 ms; repeat on selected CI reference |
| Warm example median (six measured repeats) | <= 100 ms | 24–31 ms; no p95 claim from six samples |
| Complete first-slice JS/CSS/WASM compressed transfer | <= 2 MiB Brotli total | Unmeasured: no React slice exists yet |
| User starts loop → usable result, cold assets, 10 Mbit/s + 100 ms RTT reference profile | <= 5 seconds | Unmeasured; add a browser-load test with documented shaping, never infer from loopback timings |

Do not load WASM as a prerequisite for reading home/disclaimer/help. The first slice can lazy-load its same-origin Worker on the explicit loop action while keeping the React shell usable. A 30-second load is outside the proposed budget, but no current test is claimed to enforce an unimplemented hosted load profile. Budget limits are proposed for review; the current gate collects measurements and enforces numerical parity/timeouts, not these as-yet-unapproved release thresholds.

## Historical verification and remaining gate

Final explicit three-runtime parity, numerical-coverage diagnostics, first-field diagnostics and Linux detached-descendant cleanup passed. `CGO_ENABLED=1 go test -race ./validation/wasm -count=1` passed (the expensive parity test is separately enabled as documented). Fresh `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` passed after harness hardening; full log: `artifacts/wasm-parity/full-suite.log`. The full suite's default integration-test skip is not the source of the separate parity claim. Independent final code review reported no remaining consequential correctness/safety findings, with the Linux process cleanup and example-only evidence limits retained; review artifact: `artifacts/wasm-parity/final-code-review.md`. A generated tracked DEF fixture left empty during earlier validation was restored only to its verified original tracked bytes; no unrelated authored changes were reverted. No Orca pass is claimed.

Next scope is **plan only**: one hosted HZO loop in one Worker, one accessible table and Go summary, supplementary plot and persistent disclaimer. Do not start that slice, delete Astro/Fyne, implement four views or create commits before the owner's review.

## 4. Production hysteresis Worker qualification

This receipt exercises the actual [Go entry](../../../cmd/fecim-hysteresis-wasm/main_js.go), [sole Go registry/envelope](../../../workbench/hysteresis/bridge.go), [production Worker/client](../../../web/src/worker/hysteresis-worker.mjs) and all three producer APIs. Build/dispatch/recovery instructions and exact payload contracts are in the [validation README](../../../validation/wasm/README.md). This is Worker-driver qualification, not a shipped React route or public deployment.

### Scientific parity and bounded repairs

`TestHysteresisProductionParity` compared **1,734,584 original bytes / 225 replies** from native, Node and real headless Firefox. All three SHA-256 values were:

`a72e0cb86d034406be368229c099e2c13eadb739cb36ba7ad2be2527d1e62a6a`

The trace includes signed/manual minor-loop history and negative zero, both loop models, Sine/Triangle, completed target programming, completed Preisach calibration plus settings invalidation, seeded maximum-domain write/read advance/cancel, LK calibration advance/cancel, unavailable time-resolved status, PUND with independently identified Preisach profile under selected LK context, unverified imported reference evidence and complete **64-reversal/256-resolution FORC**. A native guard rejects identical incomplete/failed science before equality can qualify it. Raw request strings also remain intact across JS; parsing/stringifying signed zero previously changed Go display output and was caught RED by the production trace.

The retained original REDs and [bounded corrections](../../tasks/HYS-WORKER-contract-repair.md) identified native assembly `math.Exp` evaluation differences, not altered model equations. The candidate uses Go's portable range-reduced exponential evaluation for existing NLS, Everett/reversible Tanh and the independent GUI FORC Gaussian weight. Focused numerical tests failed before each repair. Parameters, Gaussian width, history/wipe-out/guards, diagnostic identities, goldens and scientific tolerances remain unchanged. Numerical suites and existing curve regressions passed without golden regeneration.

Production build identity:

`7ad5dfa5788c2ea92b12a8153be5f182cb4545c258b7d40b44778f3b3487e97b`

Identity also pins `fecim-hysteresis/1`, `hysteresis-producers/1` and **go1.26.10**. It hashes the sorted Go source/dependency inventory, canonical compiled configuration, production Worker and matching runtime bytes as documented in the README. It identifies tested source bytes, not a claim that an uncommitted/modified checkout equals a repository revision. Both entries use `-buildvcs=false`; incidental commit/module stamps are excluded from these source-pinned assets.

| Actual production asset | Raw bytes | gzip-9 | Brotli-11 | SHA-256 |
|---|---:|---:|---:|---|
| `hysteresis.wasm` | 5,654,116 | 1,562,344 | 1,157,759 | `6aa37c031cf1ea29d960f4dbd33aea3e68188d46c924b811cfe3c69ab2a4db70` |
| matching `wasm_exec.js` | 16,992 | 4,348 | 3,752 | `0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14` |
| Worker/client | 7,831 | 2,346 | 1,993 | `9bf274cc57c3e047e956a20b61684298d767d127a58d3b909829c0e2b8b80373` |
| **Total** | **5,678,939** | **1,569,038** | **1,163,504** | — |

These satisfy the accepted production 6 MiB raw / 2 MiB gzip / 1.5 MiB Brotli guardrails. Asset hashes are verified before Worker/runtime execution and module instantiation; build/runtime identity must agree before operations are accepted. The immutable production catalogue compiles the existing canonical YAML in both runtimes, rather than native filesystem overrides versus a different WASM fallback. Legacy loaders remain unchanged.

### Lifecycle and admitted work

`TestHysteresisWorkerLifecycle` uses real Go Worker results for the happy paths. It checks cooperative cancellation, unsupported-operation rejection without mutation, stale requests, pending-request hard termination, replacement session creation, retired-generation rejection, explicit stale-response corruption, missing assets and independent Worker/module/runtime hash/build drift. Faults only corrupt failure paths; no mocked numerical result supplies a happy path. Bridge tests also exercise oversize/unsupported/malformed/duplicate/non-finite envelopes, unsafe JSON integers, stale revisions and truthful terminal work-limit results.

Go admits one active session/job, at most **256 core/program outer steps per advance**, **one analysis unit per advance**, 100,000 retained samples, 2,000 supplementary plot points, 32 page rows, 128 programming domains and FORC dimensions up to 64 reversals/256 resolution. Request/reference envelope is at most 1 MiB; producer lifetime solver/calibration caps remain enforced. These are bounded scheduling/allocation limits, not claims that a maximum lifetime job converges rapidly. LK calibration and continuous write/read are qualified here for advance/cancel, not complete convergence.

The dedicated budget run gated every operation at 1,000 ms per dispatch and cancellation at 500 ms. Recorded maxima (ms):

| Operation | Maximum | Operation | Maximum |
|---|---:|---|---:|
| session.new | 110 | core.snapshot | 8 |
| core.apply | 27 | core.advance | 11 |
| core.page | 6 | program.new | 10 |
| program.apply | 152 | program.advance | 266 |
| program.snapshot | 5 | analysis.new | 113 |
| analysis.advance | 3 | analysis.result | 64 |
| analysis.page | 5 | job.cancel | 9 |

Earlier candidate parity/lifecycle runs on this shared host recorded scheduling outliers up to 17,582 ms; these were not the dedicated budget run and are not hidden as a real-time guarantee. Receipts retain every input/timing. Host load can change wall time without changing scientific bytes. The example-only 250 ms target is not applied to these production operations.

### Explicit cold-loop measurement

`TestHysteresisWorkerBudgets` used three new Firefox contexts/Workers, no-store responses and actual Brotli transport. Only a private static server added **100 ms request RTT** and scheduled all response chunks on one aggregate **10,000,000 bit/s** timeline; concurrent module/runtime requests shared bandwidth. No host networking was modified.

The marker starts before importing the production client and ends after the Go-default Sine session executes one complete signed cycle (64 physical steps), pauses and returns its first 32-row Go table page. Complete-result times were **1,549 / 1,557 / 1,565 ms**, each below 5 seconds. The transfer filter includes two Worker/client fetches (1,993 bytes each), module (1,157,759), runtime (3,752): **1,165,497 Brotli bytes** inside the marker. Only the empty HTML navigation before the marker is excluded. No React shell, CSS, paint or accessibility result is counted.

Host/tools: Linux x86-64, AMD Ryzen 9 5900X, 24 logical CPUs; Go 1.26.10, Node 24.19.0, locked Playwright 1.60.0 and its Firefox 150.0.2 (`firefox-1522`). No CPU throttling or low-end-device claim. Public/CDN conditions are not represented.

### Verification and evidence inventory

The explicit enabled `CGO_ENABLED=1 go test ./validation/wasm -count=1 -v` passed all three production gates, the independent eight-point example gate, incomplete-science guard, first-field diagnostics and detached-descendant cleanup. `go test ./workbench/hysteresis -count=1`, `CGO_ENABLED=1 go test -race ./validation/wasm ./workbench/hysteresis -count=1`, `go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric -count=1` and `go test ./validation -run '^TestPhysicsRegressionCurves' -count=1` also passed. The default race command skips only the explicitly opt-in browser gates; their enabled non-race gate ran separately. The merged-tree full check remains the coordinator's gate, not a worker claim.

Independent review reproduced two additional boundary defects: a same-turn send/recreate race and typed reference allocation before the 32-branch cap. Focused REDs reproduced an unsettled retired send and 128,183,296 allocated native bytes for 200,000 null references inside a 1 MiB request. GREEN binds sends to their generation before awaiting readiness and stops reference-array token preflight at the 33rd entry before typed decoding; state/revision remain unchanged. The real Firefox same-turn revision-zero regression and the allocation-bound regression now pass, as do all final checks listed above. No broader peak-memory or scheduling guarantee is inferred. Final evidence was regenerated after both fixes; the review's earlier findings are not left open.

Retained receipts under `/var/tmp/aw-1000-bc334d01-1/final-evidence/`:

- `hysteresis-parity-1152220463/`: source-bound production input, executable/WASM and three original equal transcripts.
- `hysteresis-lifecycle-1440421539/`: same scientific bytes plus actual lifecycle/fault outcomes.
- `hysteresis-budgets-530014594/`: same scientific bytes, per-operation admission and three cold-transfer records.

Detailed logs are in that scratch directory's parent (`final-parity.log`, `final-race.log`, `final-numerical.log`, `final-bridge.log`, `final-curves.log`). Scratch retention is host-local; rerun the documented gate when qualifying different bytes/tools. These are not deployed artifacts or durable public measurements.

**Remaining at this Worker checkpoint** (see section 5 for subsequent automated React evidence): HYS-UI owns the complete React-route 2 MiB Brotli transfer and cold usable-route budget, rendering/focus/speech behavior and real Orca/keyboard qualification. No hosted upload/persistence, local capability escalation, Fyne removal, launch-default change or hardware validation is asserted here.

## 5. Shared React route qualification

The HYS-UI candidate replaces Astro with one React/TypeScript/Vite route tree: public home and `/hysteresis`, built as hosted `web/dist` and local `web/dist-local` asset profiles. The production Worker shim is unchanged. [web/README.md](../../../web/README.md) owns setup, operation and capability limits. Source owners: [shared routes](../../../web/src/main.tsx), [lifecycle scheduler](../../../web/src/hysteresis/engine.ts), [Go asset builder](../../../web/src/build/build.mjs) and [Go-generated catalogue](../../../web/src/build/catalog.go). This is automated candidate evidence, not deployment or manual screen-reader acceptance.

### Scope and scientific ownership

Go supplies materials, defaults, controls, availability, diagnostic identity, computation and scientific display/speech. The interface covers CORE, target/write-read programming, in-memory calibration, unavailable time-resolved status, independent PUND/FORC and imported unverified reference context through production envelopes. Native 32-row pages retain revisions, headers, Go speech and canonical SI access; the bounded plot is supplementary. There is no TypeScript solver, unit conversion or derived metric.

Additive Go presentation fields (`TimeValue`, `CurrentValue`, `DensityValue`, `ScientificValues`, `AppliedParameters`) expose previously raw scalar values using the existing formatter. Material physical parameters also receive Go units/speech; Ea and StressGPa are presented as SI joules and pascals without altering their source preset fields. Canonical numerical fields, equations, defaults, diagnostic protocols, goldens and tolerances are unchanged. The Go catalogue selects applicable diagnostic columns instead of displaying unrelated struct zero values.

### Actual route and parity measurements

`web/test/e2e/hosted-isolation.spec.ts` measures the complete compiled route, not just the Worker driver. Three fresh Chromium contexts time navigation through explicit Advance batch, paused Go-default Sine results and the first complete signed 32-row page. Actual HTTP assets include HTML, JS/CSS, lazy Hysteresis code, manifest, Worker, WASM and runtime; same-origin Blob execution transfers no HTTP bytes. The private server negotiates Brotli, uses no-store responses, adds 100 ms request RTT and schedules chunks over one aggregate 10,000,000 bit/s link. Compression is precomputed server-side; no host-wide networking is changed.

| Asset profile | Complete route Brotli bytes | Cold completed-result times (ms) |
|---|---:|---|
| Hosted | 1,243,900 | 2,252 / 2,245 / 2,264 |
| Local assets, no local adapter | 1,243,984 | 2,248 / 2,254 / 2,293 |

Both pass the 2 MiB Brotli route and 5 s per cold result guardrails. Module/runtime/Worker totals: 5,736,889 raw, 1,171,023 Brotli-11 and 1,581,766 gzip-9 bytes, below the independent 6 MiB / 1.5 MiB / 2 MiB caps. These are controlled private-host measurements, not CDN, low-end-device, peak-memory or real-time guarantees.

Test host: Linux x86-64, AMD Ryzen 9 5900X (24 logical CPUs), Node 24.19.0, Go 1.26.10, Playwright 1.60.0, Chromium 152.0.7977.75 and compatible Firefox 150.0.2. Chromium performs the controlled budget measurement; both engines exercise workflow/DOM/accessibility contracts.

Tested source/runtime build identity: `e102ef610c8d0ed4b006c1720d2f2a275c885a0fb5f3872e0e48f71b0aba512d`; frontend content: `81fa00d18a24863e3f68e6e17cdedc6ef059b54f14c06d6fbe6c520d6258ebe7`. Protocol/model remain `fecim-hysteresis/1` and `hysteresis-producers/1`. Both profiles emit the same Go identity/catalogue; the local callback profile changes compiled JS, not scientific code or views.

The enabled native/Node/Firefox gate compared 2,346,420 original production transcript bytes in each runtime; each SHA-256 is `9492eab33efc259d7c275a6872159ff0e61d9f096203a65e413963d1b1feaf48`. Independent example parity, Worker lifecycle and admission budgets also pass. No field stripping, parsed-object comparison, rounding or tolerance weakening was used.

### RED/GREEN and verification

The retry observed RED at the independent landing contract (missing React home), real browser workflow (missing Material input), and producer JSON scalar contract (missing Go TimeValue/units/speech). GREEN uses the shared React app and actual Go envelopes. Reapplying the prior CI step rename reproduced `TestCIRunsWebTestsAndProductionBuild` failure; GREEN preserves its exact original required install/test/build step and adds typecheck/explicit dual-profile builds, both browser profiles and enabled scientific parity alongside it. No independent CI guard was weakened.

Fresh read-only review reproduced an earlier rejected command stranding a newer Run, pending reference reads undoing Clear, missing material units/speech, valid JSON provenance rejected by invented UI metadata, and edited diagnostic inputs leaving old results unmarked. Each has a focused RED→GREEN regression (five browser cases and the extended Go JSON scalar test). Command errors now retire only their own pump; reference reads are generation-bound, catch read errors and gate Analyze; JSON provenance stays with Go; material presentation stays Go-owned; retained diagnostics label edited inputs stale. Metadata exposes Go speech as readable text rather than relying on an aria-label on an unnamed span.

Fresh checks on the source identities above:

- `cd web && npm ci && npm test`: pass, locked installation and retained React/content/deployment/no-Fyne guards.
- `cd web && npm run typecheck && npm run build:hosted && npm run build:local`: pass.
- `cd web && npm run test:e2e -- test/e2e/hysteresis.spec.ts test/e2e/hosted-isolation.spec.ts`: 51 passed, 4 measurement-only skips (one controlled Chromium measurement per profile).
- `cd web && npm run test:e2e -- test/e2e/fecim-web.spec.mjs`: 19 passed, 6 viewport-specific skips; existing home/content/layout/navigation checks retained.
- `cd web && FECIM_WEB_PROFILE=local npm run test:e2e`: 70 passed, 10 documented measurement/viewport skips.
- `FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1`: pass with compatible Firefox and retained artifacts.
- `go test ./workbench/hysteresis ./workbench/presentation -count=1` and the same packages with `-race`: pass.
- `go test ./scripts -run 'TestCI|TestExternalValidationWorkflowRejectsNonGatingMutations' -count=1`: pass, including the previously held CI contract.

Servers used assigned private ports: hosted 20100/20101, local 20110/20111, parity 20120/20121. Logs/artifacts under `/var/tmp/aw-1000-bc334d01-1/`: `ui-required-hosted-final.log`, `ui-home-final.log`, `ui-local-final.log`, `ui-parity-final.log`, `ui-parity-final/`, `ui-race-final.log`. Playwright emits source-bound `ROUTE_BUDGET` log records and `route-budget.json` attachments. Scratch evidence is host-local, not a durable public endpoint; rerun for different source/tools.

### Remaining separate gates

Real Orca + Firefox/Linux speech/keyboard qualification remains **NOT RUN** for hosted and authenticated local profiles. Automated DOM/focus/keyboard, narrow-width/200% text, quiet-progress and accessibility checks do not attest pronunciation or usable Orca table reading. The [manual acceptance owner](../superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md) must supply that evidence.

At this HYS-UI checkpoint, authenticated local adapter, embedding and serve/file workflows remained HYS-LOCAL work; section 6 records subsequent implementation. A static local asset profile alone grants no filesystem capability. Hosted buttons give instructions without upload, persistence, download or localhost discovery. No public deployment, push, shared service change, Fyne retirement, launch-default change or hardware validation occurred. The coordinator owns merged-tree full-suite verification and shared task/history reconciliation.

## 6. Authenticated embedded local profile

HYS-LOCAL adds `fecim-lattice-tools serve --root DIRECTORY` without changing
legacy CLI/Fyne/default launches. The design's `fecim serve` remains executable
shorthand. [Command/security reference](../../../cmd/fecim-lattice-tools/README.md)
owns launch/renewal/logout, exact HTTP routes, file formats, conflicts and trust
limits; [installation](../../guides/installation.md#optional-local-react-hysteresis-application)
owns prerequisites. Both profiles retain the same Hysteresis components and Go
Worker; local adds only a same-origin authenticated file adapter.

### Source/build and security boundary

`npm run build:local` stages generated assets under `internal/browserui/assets`;
`npm run build:native` rejects source/UI drift and pins native UI/core identities.
The server hashes every admitted asset before file operations. Plain Go builds
compile with the checked-in sentinel and fail `serve` clearly with
`browser-assets-unavailable`; no generated binary is committed. `make build-browser`
builds both profiles and the embedded binary. CI installs native graphics build
dependencies before real embedded tests.

The loopback listener uses exact Host/Origin/Fetch Metadata, JSON/custom-header
checks, bounded bodies/admission/timeouts, single-use 60-second fragment bootstrap,
HttpOnly Strict session cookies and logout/shutdown revocation. A renewed link
also removes/exchanges its fragment when already signed in. No request logging,
CORS, external computation endpoint or public service configuration is added.

Opaque launch-granted roots use the existing traversal-resistant `os.Root`
primitive through read, staging and exclusive publication. Existing destinations
conflict rather than overwrite. Reads enforce an 8 MiB stat-before-read cap and
bounded recheck on the opened file. Typed enumerated exports use Go serialization
and numeric SVG generation, not arbitrary HTML/SVG. Browser snapshots and imported
references remain unverified simulation data; client experimental-validation flags
are not promoted. Existing preset engine names and version-4 calibration JSON
are retained; legacy version-1 calibration is also admitted read-only. Non-default stress, seed and history settings round-trip in optional
preset keys. Loaded calibration is explicitly read-only/imported-unverified, not
installed as fresh programming calibration; users explicitly recalibrate.

### Executed candidate checks

- `env CGO_ENABLED=1 xvfb-run -a go test ./cmd/fecim-lattice-tools ./workbench/server ./internal/browserui -count=1`: pass, including default Fyne repo-surface guards.
- `env CGO_ENABLED=1 go test -race ./workbench/server -count=1`: pass.
- `cd web && npm run build:hosted && npm run build:local && npm run test:e2e -- test/e2e/hysteresis-local.spec.ts`: **10 passed**, across desktop/compact/mobile/small Chromium and desktop Firefox.
- `go test ./workbench/hysteresis -count=1`: pass; serializer extraction retains existing producer behavior.
- `cd web && npm run typecheck && npm test`: pass.

`TestHysteresisLocalSession` started RED (missing server API), then passed asset
absence/drift, foreign Host/Origin/CSRF, expiry/replay/logout and oversize checks.
A chunked trailing-body regression started RED (400 instead of 413), then passed.
`TestHysteresisLocalFiles` checks every admitted artifact kind, existing formats,
8 MiB pre-read rejection, temporary-root traversal/link/absolute/reserved names,
no outside writes, conflicts, invalid diagnostics and unverified provenance.
Review reproduced preset stress loss; its non-default L-K round-trip started RED
and now preserves stress/seed/history. Authenticated renewal's real-browser RED
retained the fragment; the regression now reaches the token-free Hysteresis route.

The integration test builds/embeds **real** local assets, starts the actual CLI
with a temporary root on assigned private port 20210, uses actual Go Worker
preset/calibration/PUND/loop data, explicitly saves/loads/exports, rejects
conflict/traversal, verifies replay/logout/renewal, no HTTP token leakage and no
browser storage. Its hosted case makes no local API request. Final run logs are
worker scratch `required-go.log` and `required-browser.log` under the assigned
`TMPDIR`; these are not public deployment evidence. Tested source fingerprint:
`f2fc8fa49afeea96b5bb22e7b89ae4820a30a24879bcc1e24d02c5f52a4590c3`;
UI: `fde1de15c96828f7c7313e64e816a4dc10dc5957ab4f4e66f04bf58289c3c715`.
Generated profiles must be rebuilt for different bytes/tools.

**Remaining:** real Orca + Firefox/Linux speech/keyboard qualification for both
profiles is NOT RUN. Automated browser workflows do not supply that manual gate,
physical-device evidence, power-loss durability, stored-study verification or a
sandbox against same-user software. No Fyne retirement/default switch, public
deployment or scientific golden/tolerance change is claimed. The coordinator
owns merged-tree full-suite verification and task/history closeout.

## 7. Combined post-LOCAL automated checkpoint

This checkpoint re-executes the existing gates on candidate
`dbd49104bebb2930cadf50b63fede04cc99e0d9d` (2026-10-28). It follows the
[final-candidate plan](../../tasks/HYS-NEXT-PLAN.md#chosen-frontier-final-candidate-automated-evidence)
and [ADR 0005 acceptance](../adr/0005-unified-react-hosted-and-local.md#required-acceptance).
Sections 4–6 remain historical evidence, not substitutes for these runs.
The HYS-UI and HYS-LOCAL reports were read in place from the main repository's
execution common directory. Their commits are `525eed4cbfc589eb8f2932f3e7bb9e6599b357c9`
and `1cb09424cae1af4b914fdeac443c26f9c3eef0db`.

**Automated result: PASS for the selected gates below.** No failed or incomplete
automated gate occurred in this run. No product repair was inferred or made.
This checkpoint changes documentation only. TDD: N/A, because no behavior,
fixture, assertion, equation, threshold, golden or build input changes.
The documentation commit has the same scientific/frontend inputs as the candidate.
This is not a task receipt, manual qualification, goal-met claim or deployment.

### Candidate and build coherence

Setup follows [web/README.md](../../../web/README.md#build-and-develop) and the
[enabled parity instructions](../../../validation/wasm/README.md#run).
The locked install used Node **24.19.0**, npm **11.17.0**, Playwright **1.60.0**,
Go **1.26.10** exactly as pinned in `go.mod`, and that Go toolchain's runtime.
Firefox **150.0.2** came from the existing `playwright-1.60.0/firefox-1522` cache.
The browser configuration selected `/usr/bin/chromium`, version **152.0.7977.75**.
Host: Linux x86-64, AMD Ryzen 9 5900X, 24 logical CPUs. No CPU throttling,
shared-service change or host-network shaping was performed.

| Identity | Value |
|---|---|
| Protocol / model | `fecim-hysteresis/1` / `hysteresis-producers/1` |
| Core/runtime-bound build | `f2fc8fa49afeea96b5bb22e7b89ae4820a30a24879bcc1e24d02c5f52a4590c3` |
| Frontend content | `fde1de15c96828f7c7313e64e816a4dc10dc5957ab4f4e66f04bf58289c3c715` |
| Engine manifest SHA-256 | `423f729ba144cd3a186fc91a2de2402eefd080bead22f28501c61111432ad857` |
| Embedded local manifest SHA-256 | `6e3486ada7c3a260943954e1115e8a9f522d9180455dbd67c93ef4223e459211` |

`node src/build/build.mjs --fingerprint` reproduced these core/UI identities.
Readback compared the complete generated `web/node_modules/.fecim/manifest.json`
with hosted, local and embedded `engine/manifest.json`: all four are byte-identical.
Every engine asset hash agrees with each production parity/lifecycle/budget bundle.
The embedded local manifest identifies the same core/UI and hashes all **10**
staged files, each byte-identical to `web/dist-local`. Both profiles were rebuilt,
not copied from historical reports. Native embedding's existing drift check ran
before the real authenticated browser tests.

### Original-byte parity, lifecycle and admission

`FECIM_WASM_PARITY=1` explicitly enabled all four independent gates:
`TestHysteresisProductionParity`, `TestHysteresisWorkerLifecycle`,
`TestHysteresisWorkerBudgets` and `TestExampleResultsJSONNativeWASM`.
All passed, with **zero skips** in the enabled package's nine tests.
The incomplete-science, example numerical-results, first-difference and
subprocess-cleanup guards also passed. Default-suite skips did not supply this proof.

Each production gate retained **225 replies / 2,346,420 original bytes** per
runtime. Native, Node and Firefox files are byte-identical in all three bundles,
SHA-256 `7e2c6ea050a1426d775893826e0748d0f138982019a738233d56dcc61ac39919`.
The unchanged trace covers the finite successful science and modes described in
section 4, including complete target programming/calibration and maximum FORC.
Native request arrays are **69,038 bytes**, SHA-256
`67aaacdf41efb00398663bdacf9b41b7ecf744849eb63edf37ec0657ad0f6bae`.
Readback also checked that JS transport fixtures retain those original request strings.
No parsed-object comparison, field stripping or normalization replaced byte equality.

The independent eight-run example produced **21,972 original bytes** in each
runtime, SHA-256 `aa7a3ee41f95b44c9200d604db67a42578b4078892f8ce438586802a96061da8`.
Its existing guard requires eight successful runs with four finite metrics each.
Its report revision is the candidate above, not the later documentation-only commit.

The lifecycle bundle records real cooperative cancel, unsupported-operation and
stale-request rejection, pending termination, recreation, same-turn retired-send
and retired-generation rejection, injected stale replies, independent asset hash
drift, build drift and missing-asset rejection. No fault fixture supplies science.
Dedicated budget maxima (ms) were:

| Operation | Maximum | Operation | Maximum |
|---|---:|---|---:|
| session.new | 115 | core.snapshot | 9 |
| core.apply | 29 | core.advance | 13 |
| core.page | 8 | program.new | 13 |
| program.apply | 156 | program.advance | 262 |
| program.snapshot | 7 | analysis.new | 111 |
| analysis.advance | 3 | analysis.result | 94 |
| analysis.page | 8 | job.cancel | 11 |

All registry operations meet the existing **1,000 ms** dispatch bound and
**500 ms** cancellation bound at the unchanged admitted caps in section 4.
This does not prove full LK calibration convergence or a real-time guarantee.

### Payload and complete-route measurements

| Production asset | Raw bytes | Brotli-11 | gzip-9 | SHA-256 |
|---|---:|---:|---:|---|
| `hysteresis.wasm` | 5,712,064 | 1,165,895 | 1,575,075 | `d83a655b941ad4888e936a2f880473bc2d0b137812e85516e2834ad4e325f603` |
| `wasm_exec.js` | 16,992 | 3,752 | 4,348 | `0c949f4996f9a89698e4b5c586de32249c3b69b7baadb64d220073cc04acba14` |
| Worker/client | 7,831 | 1,993 | 2,346 | `9bf274cc57c3e047e956a20b61684298d767d127a58d3b909829c0e2b8b80373` |
| **Total** | **5,736,887** | **1,171,640** | **1,581,769** | — |

The totals satisfy **6 MiB raw / 1.5 MiB Brotli / 2 MiB gzip**.
Worker-driver cold signed Sine completion took **1,573 / 1,566 / 1,564 ms**.
Each transferred **1,173,633 Brotli bytes**, including two Worker/client fetches,
module and runtime, excluding only the empty HTML navigation before its marker.

The existing [complete-route case](../../../web/test/e2e/hosted-isolation.spec.ts)
then measured both compiled React profiles separately:

| Profile | Complete route Brotli bytes, each run | Cold completed-result times (ms) |
|---|---:|---|
| Hosted | 1,244,745 | 2,248.331 / 2,260.701 / 2,292.294 |
| Local static assets with local adapter | 1,245,725 | 2,269.724 / 2,258.116 / 2,247.156 |

All six route runs meet **2 MiB Brotli** and **5,000 ms** bounds.
These are three fresh contexts per profile with no-store responses, private-server
100 ms request RTT and one aggregate 10,000,000 bit/s chunk schedule.
Route timing starts at navigation and ends at the paused Go Sine result with its
complete signed 32-row first page. The filter includes HTML, CSS, main/lazy JS,
manifest, Worker, WASM and runtime. Local additionally records adapter JS and its
intentional `/api/v1/roots` request. The static server returns 404 for that API
(no asset-byte header), not authenticated file capability. Real file workflows
were tested separately against the embedded binary. These measurements are not
CDN, low-end-device, paint, physical-device or authenticated-route latency claims.

### Executed commands and selected coverage

Commands ran on unchanged candidate sources. Long groups used `autowork-run`
and explicit `timeout` bounds. Browser servers used assigned ports **20100/20101**,
embedded launch **20110**, and enabled parity **20120/20121**.

| Command (from repository root unless prefixed) | Actual result |
|---|---|
| `cd web && npm ci --cache "$AUTOWORK_CACHE/npm-node24" && npm test && npm run typecheck && npm run build:hosted && npm run build:local` | PASS, locked install, 1 web contract test, typecheck, both profiles and local staging |
| `env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -v` | PASS, 9 tests, no skips, retained native/Node/Firefox production and example outputs |
| `cd web && npm run test:e2e -- test/e2e/hysteresis.spec.ts test/e2e/hosted-isolation.spec.ts test/e2e/hysteresis-local.spec.ts` | PASS, 65 selected: 61 passed, 4 intentional measurement skips |
| `cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hysteresis.spec.ts` | PASS, 45 selected/passed, no skips |
| `cd web && FECIM_WEB_PROFILE=local npm run test:e2e -- test/e2e/hosted-isolation.spec.ts --grep 'complete real route transfer and cold Go sine result meet the production guardrails$' --project=desktop-chromium` | PASS, exactly 1 selected/passed, no skips, also confirmed by `--list` |
| `env CGO_ENABLED=1 xvfb-run -a go test ./cmd/fecim-lattice-tools ./workbench/server ./internal/browserui -count=1` | PASS, CLI/default-Fyne guards and authenticated session/files tests, browserui compiles (no test files) |
| `env CGO_ENABLED=1 go test -race ./validation/wasm ./workbench/hysteresis ./workbench/server -count=1` | PASS, producer/server and harness checks, 4 opt-in browser gates skipped only in this separate default race run |
| `go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric -count=1` | PASS, no scientific changes or golden regeneration |
| `go test ./validation -run '^TestPhysicsRegressionCurves' -count=1` | PASS |

Both browser workflow groups selected all five configured projects:
`desktop-chromium`, `desktop-compact-chromium`, `mobile-chromium`,
`mobile-small-chromium` and `desktop-firefox`. Hosted's four skips are only the
complete-route measurement on non-desktop-Chromium projects, as the existing case
requires one controlled measurement per profile. The hosted group includes **five
real embedded local sessions**, one per project, plus five hosted instruction-only
file cases. Real sessions used actual Go Worker data and temporary granted roots
for preset/calibration save/load, loop/diagnostic exports, conflicts, traversal,
replay, signed-in renewal, logout and no browser storage/token-request leakage.
Each session stopped its child server and removed its temporary root.
The hosted-only isolation workflow was not misapplied to the local adapter.

### Retained evidence and remaining limits

Private host-local evidence parent:
`/var/tmp/aw-1000-bc334d01-1/t015-evidence/`.
`FECIM_WASM_ARTIFACT_DIR` was set to its `wasm/` subdirectory, with the locked
Playwright browser cache selected through `PLAYWRIGHT_BROWSERS_PATH`.
No runtime bundle, trace, launch token or temporary grant is committed.

- `wasm/hysteresis-parity-3983452131/`,
  `wasm/hysteresis-lifecycle-2514327016/` and
  `wasm/hysteresis-budgets-354685648/`: exact requests/replies, native/WASM
  executables and `measurements-*.json`, including every admission input/timing.
- `wasm/parity-2021736445/`: independent resolved example/native project,
  original three-runtime reports, private module and measurements.
- `build.log`, `parity.log`, `hosted.log`, `local-shared.log`,
  `local-budget.log`, `focused-go.log` and matching `.exit` files: complete
  command results. `hosted-route-budget.json` and `local-budget-route-budget.json`
  retain the exact `ROUTE_BUDGET` records, including unrounded timings/resources.
- `verify-assets.py`, `identity-readback.json` and `identity-readback.log`:
  source fingerprint, all manifests/asset hashes, original-byte comparisons and
  transcript/request lengths/hashes. Scratch evidence is not a public endpoint
  or a durable qualification for later changed sources/tools.

**Still NOT RUN:** real Orca + Firefox/Linux speech and keyboard qualification
for **both hosted and authenticated local profiles**. [HYS-ORCA's manual gate](../superpowers/specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md)
remains with its real tester/owner. Automated keyboard/focus/DOM/accessibility
checks cannot supply that evidence. No human sign-off, Fyne retirement,
launch-default switch, deployment, push or hardware validation is asserted.
All results remain simulation-only. The coordinator owns the merged full check,
task receipts, shared-history closeout and final planning. Rebuild and rerun affected
gates if any bound Go, frontend, Worker, dependency or runtime input changes.
