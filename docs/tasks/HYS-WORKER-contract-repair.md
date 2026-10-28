# HYS-WORKER bounded contract correction

## Evidence and disposition

The accepted [ADR 0005](../internals/adr/0005-unified-react-hosted-and-local.md#required-acceptance) requires fixing shared calculation/serialization differences under scientific regression review. The [accepted plan](../internals/superpowers/plans/2026-10-25-hosted-hzo-first-slice.md#go-seams-and-invariant-contract) forbids target-specific normalization, changed equations/presets and weakened scientific regressions. HYS-WORKER's original obligations explicitly anticipated a bounded shared-math amendment. This correction supplies that missing file boundary for the existing parity outcome; it adds no capability or physical model.

Worker evidence remains in place under `.git/autowork-execution/slot-data/HYS-WORKER/attempt-1/report.json` (resolve `.git` through `git rev-parse --git-common-dir` in a worktree). The expanded production check first differs at `/34/result/Core/Plot/35/P`: native `-0.223257000829515`, Node and Firefox `-0.22325700082951497`. Request: `program.advance`, Steps=256, DT=0.1, after signed manual history. Sample: Index=35, Time=3.0000000060000014, E=-73950000. Build identity: `dc59e9fbc577b608258f75084c2e17983e0b448811a3034fe0eb99c746a5dff9`. Disabling native FMA did not fix it. This identifies an observed discrepancy, not its root cause.

The report names original artifacts at `/var/tmp/aw-1000-bc334d01-1/evidence/hysteresis-parity-3516752142/` and the failing log at `/var/tmp/aw-1000-bc334d01-1/parity-expanded.log`. They are not copied into the repository. The report records passing bridge tests, failing production parity and unfinished remaining gates, not qualification.

## Bounded write correction

Existing source traces the relevant Preisach path through:

- `workbench/hysteresis/session.go:364-398`: session stepping dispatches to the existing Preisach model (or L-K); this trace does not establish a need to rewrite orchestration.
- `module1-hysteresis/pkg/ferroelectric/preisach.go:380-420`: Update/advancePolarization combines stack, reversible contribution and quasi-static relaxation; `:500-504` evaluates the reversible Tanh.
- `shared/physics/preisach.go`: Update/ComputePolarization evaluates the existing Everett integral and compensated accumulation.
- `shared/physics/tanh_everett.go`: Calculate evaluates the product-form Everett integral with Tanh.
- `shared/physics/nls.go`: CalculateTau/Relax evaluates the existing exponential relaxation.

Add only those four numerical implementation files and their existing same-stem tests to Touches. Changes there must be justified by a focused reproduction of the reported discrepancy and preserve existing equations, inputs, parameters, history/guard semantics and scientific tolerances. Listing a path is not evidence it needs modification. No unrelated solver/model repair or toolchain modification is admitted. If investigation points elsewhere, retain evidence and obtain another evidence-backed bounded contract correction before editing outside Touches.

The worker report also names `config/hysteresis.go`, `workbench/hysteresis/catalog.go` and `validation/wasm/parity_test.go` among its draft changes. Add these exact paths for production WASM integration and harness reuse only, not changes to scientific defaults or the original example gate.

## Follow-on FORC boundary correction

The resumed attempt-1 worker report at `.git/autowork-execution/slot-data/HYS-WORKER/attempt-1/report.json` records focused numerical RED/GREEN and passing numerical, curve-regression and bridge suites, but **production parity still fails**. The original reply-34 programming mismatch is gone in that candidate; the next mismatch is the maximum 64-reversal/256-resolution FORC result at `/217/result/FORC/Matrix/density/29/88`: native `-0.011478234720019359`, Node/Firefox `-0.011478234720012474`. The first differing sweep sample is `curves[30].polarization_cm2[7]`, reversal `-0.04761904761904767`, applied field `-0.2698412698412699`: native `-0.30545939929820926`, WASM `-0.3054593992982093`. Final candidate build identity: `99e61ef0dfc0e06d6d4b5106d20813e7612943034ce5bc494495f92386a0f081`.

The report retains inputs, executables, WASM, original three-runtime transcripts and measurements at `/var/tmp/aw-1000-bc334d01-1/resume-evidence/hysteresis-parity-2445483615/`, with `resume-parity.log` and `resume-red.log` in its parent scratch directory. These paths are worker-reported evidence, not independently rerun maintainer qualification; artifacts stay in place, not in commits. Slot state was held, recorded worker PID 1324066 was absent, and the report's commit was null; no live worker or commit awaiting integration was found.

Live source corroborates the exact missing boundary:

- `workbench/hysteresis/analysis.go:17-18,154-176` admits the reported limits and constructs `physics.GUIFORCEverett` for the independently identified `gui-forc` profile.
- `module1-hysteresis/pkg/gui/forc_panel.go:47-48` delegates the GUI Everett calculation to that same shared implementation.
- `shared/physics/worldclass_forc.go:13-26` retains the original clamps and Gaussian weight, including the direct `math.Exp` evaluation at line 25.
- `shared/physics/worldclass_forc_test.go` contains the existing sweep monotonicity, density and JSON-export guards.

Add only `shared/physics/worldclass_forc.go` and `shared/physics/worldclass_forc_test.go` to the existing Touches. This is an implementation-level reproducibility repair of the already-required GUI Gaussian weight, not permission to change the equation, Gaussian width, clamps, material Ps, Emax/units, profile identity, sweep/history/differentiation/resampling semantics or admission limits. Reproduce the retained FORC RED and add a focused failing numerical regression before the smallest shared-evaluation repair. Do not assume the direct exponential is the only cause; prove the repair through the complete native/Node/Firefox transcript. Preserve completed-programming and maximum-FORC coverage, the independent example gate, all scientific tolerances and goldens. Other numerical files remain outside the boundary unless separately admitted from evidence.

Scope, every Acceptance item, Checks, Exclusions, Dependencies, resources, section and rank remain unchanged. All original checks already include the relevant numerical/scientific/full-suite guards. Lifecycle, budgets and final production documentation remain unfinished worker obligations; the task and goal remain open/unmet, not qualified by this correction.

## Execution and verification

- [x] Read current task, Sources, accepted plan and original worker contract/report.
- [x] Check slot state: held, worker process exited, report has no commit awaiting integration.
- [x] Preserve Scope, every Acceptance item, Exclusions, Dependencies, resources, section and rank.
- [ ] Worker: reproduce the resumed maximum-FORC discrepancy through the real production native/Node/Firefox gate; retain completed-programming coverage, identical input bytes, build identity, seeds and step schedule.
- [ ] Worker: write/run a focused failing numerical regression before the minimum shared calculation repair. Do not round, strip fields, compare tolerances instead of bytes, drop completed programming, alter presets or regenerate goldens.
- [ ] Worker: run all original Checks, plus `go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric -count=1` and `go test ./validation -run '^TestPhysicsRegressionCurves' -count=1` without golden updates. Run `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` before delivery. Passing default tests alone do not enable or qualify production WASM parity.
- [ ] Worker: finish actual lifecycle/budget measurements and production documentation; retain the separate HYS-UI route-budget and real Orca qualification gaps.

Maintainer TDD: N/A — task metadata/documentation only; no production code, tests or runtime state changed. Contract validation and diff checks apply to this correction; numerical RED/GREEN and production qualification remain worker obligations.

Follow-on maintainer verification: `tasks validate --worktree .` passed (30 tasks, 8 goals, no drafts); task-header comparison confirmed only the two evidenced Touches were added and every other header field was preserved; relative documentation links exist; `git diff --check` passed. `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` passed (log `/tmp/fecim-hys-worker-maintainer-fullcheck.log`). This default full check does not enable the opt-in production WASM parity/lifecycle/budget gates and does not qualify the retained worker candidate.
