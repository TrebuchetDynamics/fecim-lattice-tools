# Main-check process-variation yield repair

## Contract and evidence

The DOCS-MAINTAINER request authorizes one first-in-Now repair task for the uncovered `TestM2_PV_03_YieldDecreasesWithNoiseLevel` full-check failure. `MAIN-M2-YIELD` reuses the unmet `MAIN-CHECK` goal. The three previous MAIN repair tasks and HYS-WORKER are coordinator-closed; no open task covers this assertion. Preserve their closures, all existing dependencies and goal statuses. No producer prerequisite is evidenced for this independent validation repair.

Read the bundle **in place**, never copy its files into a worktree or commit:

`/home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/3242627522d34812fd956325c7b5b3f43a834564`

- `main_check.json`: state `red`, result `TEST_FAILURE`, checked revision `3242627522d34812fd956325c7b5b3f43a834564`, last green `9bba4b85c86d64682af05ad2aba9ff2df0a8e7db`; sole listed failure `TestM2_PV_03_YieldDecreasesWithNoiseLevel`.
- `main-check.log:204-208`: `process_variation_yield_test.go:71` rejects yields `1.000 (200/200)`, `1.000 (200/200)`, `0.025 (5/200)` at sigma `0.01`, `0.10`, `0.30`: the first two yields are equal rather than strictly decreasing. `shared/crossbar` fails at lines 236-237 after 13.656s; the full run reaches final `FAIL` at line 310. This is not a check-tool crash.
- `main-check.exit=1`; `main-check-reruns.log` is empty; `flaky_history.json={}`. There are no recorded isolated reruns or three-revision flake evidence. Do not invent those results, classify a product defect from this single failure, or reclassify main as flaky on speculation.
- Commits since green, as recorded in the bundle:
  - `32426275 docs(tasks): close HYS-WORKER (pi-autowork)`
  - `6bacc1f1 Merge commit 'f85cc9671cd1c813f16ce85e7bea2210a26c494a' into HEAD`
  - `f85cc967 feat: qualify production hysteresis Worker across Go Node and Firefox`

## Source-backed boundaries

- `AGENTS.md:326` supplies the unchanged Full check: `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`.
- `shared/crossbar/process_variation_yield_test.go:11-81` programs constant 0.85 weights into an 8x8 array, drives unit inputs, uses 12-bit ADC/DAC and a variation-disabled ideal baseline, then checks 200 trials at each of the three sigmas with variation enabled, IR drop/sneak paths disabled and temperature 300 K. Yield means every output has strictly less than 10% relative error. Preserve strict yield ordering and the existing low/high-sigma sanity bounds (0.5). Lines 27 and 48 call global `rand.Seed`; investigate whether these actually seed the draws on the repository's Go 1.26.10 toolchain (`go.mod:3`), rather than assuming replayability.
- `shared/crossbar/array.go:239-271` initializes per-cell `NoiseFactor` using global `rand.NormFloat64`; `enhanced.go:263-298` applies variation before ADC quantization. These are diagnostic sources, not authorized production edit targets. Examine finite-sample saturation, fixture coupling, actual RNG behavior and package/test-order interference. A same-package, test-local controlled fixture can use existing cell state without introducing a production RNG API, if evidence supports that approach. Preserve replayable stochastic coverage rather than choosing a lucky seed.
- The listed commits change no crossbar code or yield test. The HYS-WORKER report at `.git/autowork-execution/slot-data/HYS-WORKER/attempt-1/report.json` identifies production bridge/numerical parity work, not crossbar causation. Its task is already closed; this task neither edits that record nor repeats its qualification.
- The changed `validation/wasm/hysteresis_parity_test.go`, `parity_test.go`, `hysteresis_harness.mjs` and `harness.mjs` are conditional validation targets only for demonstrated resource/process interference causing this failure. `hysteresis_parity_test.go:59-63` makes production browser gates opt-in; the ordinary transcript regression still runs. Temporal association is not evidence of interference. Other changed production, physics, Worker, documentation and task files are not repair targets.

## Worker contract and verification

Investigate isolation, resource load, fixture timing, randomness/replay and statistical saturation before choosing a repair. Record a deterministic failing automated regression for the demonstrated validation/fixture defect before the minimum validation-only change; a controlled reproducer is acceptable without reproducing uncontrolled host scheduling. Preserve the 8x8 workload, weights/input, three sigmas, 200 trials, 12-bit conversion, ideal/variation comparison, per-output 10% criterion, strict ordering and both existing sanity bounds. Evidence-backed controlled fixtures plus separate replayable stochastic coverage may resolve confounding; do not drop coverage, select a lucky seed, loosen assertions, skip or retry until passing.

Do not change scientific computation, physical parameters, defaults, quantization, goldens, production randomness/APIs, browser product features, dependencies or launch/build policy. No global suite serialization. Existing defects in the named validation/fixture paths may be repaired only as needed to restore this check. If the diagnosis requires a product or scientific change, report blocked with evidence rather than widening scope.

Run:

1. `env CGO_ENABLED=1 go test ./shared/crossbar -run '^TestM2_PV_03_YieldDecreasesWithNoiseLevel$' -count=20 -v`
2. `env CGO_ENABLED=1 go test ./shared/crossbar -count=1`
3. `env CGO_ENABLED=1 go test ./validation/wasm -count=1`
4. The unchanged Full check above.
5. Only if WASM validation paths change: provision existing locked prerequisites per `validation/wasm/README.md` and run `env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -v`. Preserve exact native/Node/Firefox scientific bytes, production lifecycle/budget checks, the original example gate and owned-process cleanup. Missing prerequisites or opt-in skips are not parity proof. A yield-test-only repair does not require the enabled browser gate.

Record actual RED/GREEN and final results with revisions. Report unrelated failures without altering other task contracts. Update Evidence on this task ID for later bundles or isolated reruns; never add a second task for the same failing test. This plan is not proof of repair or authority to close/qualify tasks or goals.

## Maintainer verification

On unchanged product HEAD `3242627522d34812fd956325c7b5b3f43a834564`, `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` exited 0 during this docs-only pass. `shared/crossbar` ran uncached in 12.481s. Ephemeral local log `/tmp/pi-docs-maintainer-m2-yield-full-check.log` is not committed. This later full-suite pass is not an isolated rerun, proof of repair or three-revision flake evidence; retain the original red bundle. Independent read-only review found only a misleading rank note, now removed. Authoritative section `Now`/rank `1000` selects this sole open Now task first.

## Maintainer checklist

- [x] Read all five bundle files in place; compare failing source, changed files, existing tasks and HYS-WORKER report.
- [x] Confirm the failure is uncovered; retain closures, dependencies and goal statuses.
- [x] Validate ledger, source references and scoped diff; run the repository Full check before committing.

TDD: N/A for this maintainer change: task documentation only, no executable behavior changes. Worker validation changes still require RED/GREEN.
