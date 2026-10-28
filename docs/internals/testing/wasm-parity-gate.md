# WASM parity gate — historical example and production qualification

**Current scope:** [Production hysteresis qualification](#4-production-hysteresis-worker-qualification) adds real CORE/PROGRAM/ANALYSIS Worker evidence. Sections 1–3 and the historical verification below retain the earlier **example-only** receipt; their unimplemented/next-scope statements describe that earlier checkpoint, not current production status. No React-route, deployment or Orca qualification is inferred.

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

**Remaining separate gates:** HYS-UI owns the complete React-route 2 MiB Brotli transfer and cold usable-route budget, rendering/focus/speech behavior and real Orca/keyboard qualification. No hosted upload/persistence, local capability escalation, Fyne removal, launch-default change or hardware validation is asserted here.
