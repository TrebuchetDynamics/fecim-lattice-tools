# Main-check architecture-comparison repair

## Contract and evidence

The DOCS-MAINTAINER request authorizes one first-in-Now repair task for the uncovered `TestStressArchitectureComparison` full-check failure. `MAIN-STRESS-ARCH` reuses the unmet `MAIN-CHECK` goal; it does not reopen the coordinator-closed `MAIN-LK-5TARGETS` or edit the running `MAIN-M3-SCALING`. Neither task covers this assertion. No producer prerequisite is evidenced; shared conditional WASM investigation paths are not dependencies.

Read the bundle **in place**, never copy its files into a worktree or commit:

`/home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/87e6e395bb62c4d1b7ee98faffbb928f6ba78c9d`

- `main_check.json`: state `red`, result class `TEST_FAILURE`, checked revision `87e6e395bb62c4d1b7ee98faffbb928f6ba78c9d`, last green `c5e4504f22645ef03835641b448cdbd99ad38d41`; sole listed failure `TestStressArchitectureComparison`.
- `main-check.log:186-188`: `mvm_stress_test.go:459` reports 2T1R RMSE `0.002369587700487687` exceeding 1T1R RMSE `0.0022433374440505908` within the existing `1.05x` margin. Logged 0T1R RMSE is `0.7748`. `shared/crossbar` fails at lines 222-223 after 17.549s; the full check reaches final results and ends `FAIL` at line 295.
- `main-check.exit` is `1`; `main-check-reruns.log` is empty; `flaky_history.json` is `{}`. No isolated reruns or three-revision flake evidence are recorded. Do not reclassify this as a confirmed flaky test or infer a product defect from one failure.
- Commits since last green, as recorded in the bundle:
  - `87e6e395 docs(tasks): close MAIN-LK-5TARGETS (pi-autowork)`
  - `075100e3 Merge commit 'a3413652d69592cc0dac9a7b9f21845722e173fb' into HEAD`
  - `e5192bbd Merge commit '49bfa3940fd4d7d0441a9046bdb70d3817ba198d' into pi-autowork/20261025-233730`
  - `49bfa394 docs(tasks): queue evidence-backed M3 scaling check repair`
  - `a3413652 fix: preserve headless LK target boundaries in Fyne CSV logs`
  - `cd69673e Merge commit 'b3a8519731eb768409ce36790bb3a58011acdd79' into pi-autowork/20261025-233730`
  - `b3a85197 docs(tasks): queue evidence-backed LK full-check repair`
  - `23405de8 Merge commit '7048b840b5fe39b49e75c37ef673e8e7bdc56f39' into pi-autowork/20261025-233730`
  - `7048b840 test(wasm): gate exact native Node and Firefox report parity`

## Source-backed boundaries

- `AGENTS.md:326` supplies the unchanged Full check: `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`.
- `shared/crossbar/mvm_stress_test.go:401-475` builds one 16x16 array with noise level 0.02 and 8-bit ADC/DAC, programs random weights and input, then enables IR drop, sneak paths and variation for all three architectures. It asserts 2T1R RMSE <= 1.05 times 1T1R RMSE, 1T1R RMSE <= 0T1R RMSE, and finite actual outputs. Lines 793-795 call the global `rand.Seed` with wall-clock time; investigate actual RNG behavior under the repository Go toolchain rather than assuming that call provides replayable seeds.
- `shared/crossbar/enhanced.go:134-313` computes architecture-dependent IR drop, variation, DAC/ADC quantization and sneak contributions. `shared/crossbar/array.go` supplies variation factors. Read these to distinguish assertion/fixture confounding from a numerical defect; they are diagnostic sources, not permission to change scientific computation.
- No crossbar implementation or stress-test file changed in the listed commits. `a3413652` changes only `module1-hysteresis/pkg/gui/data_logger.go` and its regression test; the LK worker report identifies bounded-queue target-boundary loss, not crossbar numerical causation. Preserve that integrated repair and closure. The task-only commits and browser planning docs are not repair targets.
- `7048b840` changes the six executable/test WASM paths listed in Touches. These are conditional investigation/repair targets only for demonstrated validation resource/process interference causing this failure. `validation/wasm/parity_test.go` makes three-runtime execution opt-in; ordinary process-control regressions still run. Temporal association alone does not establish interference. Preserve exact report parity and owned-process cleanup.

## Worker contract and verification

Investigate fixture randomness/replay, variation and quantization confounding, package/test-order isolation, resource load and fixture timing before choosing a repair. Start with a deterministic failing automated regression for the demonstrated validation/fixture defect, then make the minimum validation-only change and record RED/GREEN. A controlled fixture reproducer is acceptable; do not demand reproduction of an uncontrolled host scheduling failure. Keep all three architectures, the 16x16 workload, noise/variation coverage, IR-drop/sneak coverage, finite-output assertions and existing RMSE comparisons including the 1.05x allowance. Do not merely select a lucky seed, remove variation, skip a failure, loosen a threshold or retry until passing. If separating controlled architecture comparison from stochastic coverage is justified, preserve both checks with replayable evidence rather than dropping coverage.

No physics equation, architecture isolation factor, numerical implementation, preset, quantization rule, material, golden or production RNG/API change is authorized. If evidence requires such a change or a product decision, report blocked with the diagnosis instead of widening scope. Existing defects in the named validation/fixture paths may be repaired when necessary to restore this check. No global suite serialization, new dependency, browser feature or launch/build-policy change.

Run the focused test repeatedly uncached, the shared/crossbar suite uncached, and the unchanged Full check. Report independent failures without editing their task contracts or claiming they are repaired. Run ordinary WASM regressions; if any WASM path changes, also provision the existing locked prerequisites per `validation/wasm/README.md` and run the enabled native/Node/Firefox parity gate. Missing prerequisites or an opt-in skip are not parity proof. A stress-test-only repair does not require that enabled browser gate.

Update Evidence on this task ID for later revision bundles or isolated reruns; do not add a duplicate for the same test. This document authorizes investigation and bounded repair, not task closure or a claim that main is green.

## Maintainer verification

On unchanged product HEAD `87e6e395bb62c4d1b7ee98faffbb928f6ba78c9d`, `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` exited 0 during this docs-only pass. `shared/crossbar` ran uncached in 10.657s. Ephemeral local log: `/tmp/pi-docs-maintainer-stress-full-check.log` (not committed). This one later full-suite pass is not proof of repair, an isolated rerun or three-revision flake evidence; retain the original red bundle and investigate. Independent read-only review found no critical, important or minor findings.

## Maintainer checklist

- [x] Read all five bundle files in place and compare failing test, changed files, existing tasks and LK worker report.
- [x] Confirm the failure is uncovered; preserve the running M3 task, closed LK task, goal statuses and existing dependencies.
- [x] Validate task ledger, source references and scoped diff; run the repository Full check before committing.

TDD: N/A for this maintainer change: task documentation only, no executable behavior changes. The worker's validation changes still require RED/GREEN.
