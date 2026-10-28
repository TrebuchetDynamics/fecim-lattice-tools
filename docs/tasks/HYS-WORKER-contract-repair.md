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

## Execution and verification

- [x] Read current task, Sources, accepted plan and original worker contract/report.
- [x] Check slot state: held, worker process exited, report has no commit awaiting integration.
- [x] Preserve Scope, every Acceptance item, Exclusions, Dependencies, resources, section and rank.
- [ ] Worker: reproduce the completed-programming discrepancy through the real production native/Node/Firefox gate; retain identical input bytes, build identity, seeds and step schedule.
- [ ] Worker: write/run a focused failing numerical regression before the minimum shared calculation repair. Do not round, strip fields, compare tolerances instead of bytes, drop completed programming, alter presets or regenerate goldens.
- [ ] Worker: run all original Checks, plus `go test ./shared/physics ./module1-hysteresis/pkg/ferroelectric -count=1` and `go test ./validation -run '^TestPhysicsRegressionCurves' -count=1` without golden updates. Run `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` before delivery. Passing default tests alone do not enable or qualify production WASM parity.
- [ ] Worker: finish actual lifecycle/budget measurements and production documentation; retain the separate HYS-UI route-budget and real Orca qualification gaps.

Maintainer TDD: N/A — task metadata/documentation only; no production code, tests or runtime state changed. Contract validation and diff checks apply to this correction; numerical RED/GREEN and production qualification remain worker obligations.
