# WASM parity gate — implementation receipt and proposed budgets

Scope: prerequisites only, before any React feature implementation. Related: [gate/run instructions](../../../validation/wasm/README.md), [one-app decision](../adr/0005-unified-react-hosted-and-local.md), [package feasibility](../architecture/browser-wasm-feasibility.md) and [HZO first-slice plan](../superpowers/plans/2026-10-25-hosted-hzo-first-slice.md). No implementation of that slice or commit is authorized by this receipt.

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

## Verification and remaining gate

Final explicit three-runtime parity, numerical-coverage diagnostics, first-field diagnostics and Linux detached-descendant cleanup passed. `CGO_ENABLED=1 go test -race ./validation/wasm -count=1` passed (the expensive parity test is separately enabled as documented). Fresh `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` passed after harness hardening; full log: `artifacts/wasm-parity/full-suite.log`. The full suite's default integration-test skip is not the source of the separate parity claim. Independent final code review reported no remaining consequential correctness/safety findings, with the Linux process cleanup and example-only evidence limits retained; review artifact: `artifacts/wasm-parity/final-code-review.md`. A generated tracked DEF fixture left empty during earlier validation was restored only to its verified original tracked bytes; no unrelated authored changes were reverted. No Orca pass is claimed.

Next scope is **plan only**: one hosted HZO loop in one Worker, one accessible table and Go summary, supplementary plot and persistent disclaimer. Do not start that slice, delete Astro/Fyne, implement four views or create commits before the owner's review.
