# Main-check M3 scaling repair

## Contract and evidence

The DOCS-MAINTAINER request authorizes one first-in-Now task for the uncovered `TestPerformance_M3_PERF_03_Scaling` full-check failure. Reuse the unmet `MAIN-CHECK` goal: its outcome is a passing full offline check. `MAIN-LK-5TARGETS` covers a different failure and has a live worker; do not edit its contract, closure, priority or dependencies. No other open task covers scaling.

Read this evidence bundle **in place**, never copy it into a worktree or commit:

`/home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/cd69673e3138fa2523ff02f9021b4cb476cd49a8`

- `main_check.json`: state `red`, result class `TEST_FAILURE`, checked revision `cd69673e3138fa2523ff02f9021b4cb476cd49a8`, last green `c5e4504f22645ef03835641b448cdbd99ad38d41`; sole listed failure `TestPerformance_M3_PERF_03_Scaling`.
- Commits since green: `cd69673e Merge commit 'b3a8519731eb768409ce36790bb3a58011acdd79' into pi-autowork/20261025-233730`; `b3a85197 docs(tasks): queue evidence-backed LK full-check repair`; `23405de8 Merge commit '7048b840b5fe39b49e75c37ef673e8e7bdc56f39' into pi-autowork/20261025-233730`; `7048b840 test(wasm): gate exact native Node and Firefox report parity`.
- `main-check.log:104-122`: scaling test fails at `performance_scaling_test.go:103`; batch 100 achieves 1994.0 img/s, batch 1000 achieves 1445.5 img/s, a 27.5% drop against the 25% allowance. The separate 100 img/s baseline is exceeded. `shared/neural` fails after 197.750s; the log reaches final results and ends `FAIL` at line 185. The test's unconditional logged `PASS` is not a passing Go test result.
- `main-check.exit` is `1`; `main-check-reruns.log` is empty; `flaky_history.json` is `{}`. No isolated reruns or three-revision flake evidence exist here. Preserve the recorded classification; neither a later pass nor temporal association proves a flake, a product defect or WASM causation.

## Source-backed investigation boundaries

- `AGENTS.md:326`: unchanged Full check is `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`.
- `shared/neural/performance_scaling_test.go:14-123` loads existing MNIST images/weights before measurement, configures an 8-bit noise-free network, performs 20 warmup inferences, then takes one sequential wall-clock sample for each of batches 1, 10, 100 and 1000. Lines 99-114 enforce the 25% comparison allowance and 100 img/s baseline. Investigate sample duration/order, warmup, fixture isolation and concurrent full-suite resource pressure before choosing a repair.
- `shared/neural/network_inference.go:30-106` uses pooled scratch and returns copied results. This is an investigation lead for allocations/GC, not evidence authorizing an inference algorithm rewrite. No neural production file changed in the listed commits.
- `7048b840` changed the six executable/test paths in `validation/wasm` listed in Touches, plus its README and browser architecture/planning documentation. `b3a85197` changed only the prior LK task, its goal and evidence index. Those docs are context, not product repair targets.
- `validation/wasm/parity_test.go:28-30` makes three-runtime execution opt-in. Its ordinary process-control tests still run. Investigate those changed paths only for demonstrated cross-package resource/process interference; preserve process ownership and exact report parity. Follow `validation/wasm/README.md` if an enabled gate is needed. Shared paths do not make the live LK task a producer prerequisite; add no dependency for overlap.

## Worker contract and verification

Investigate the failure as recorded. If measurement/fixture timing is at fault, repair that validation path with bounded, representative measurement rather than changing inference behavior. First add a deterministic failing regression for the demonstrated measurement/fixture/process defect (injected timing samples are appropriate if supported by the diagnosis), then make the minimum repair and record RED/GREEN. Do not demand a deterministic host scheduling failure as the only possible RED. Preserve all batch sizes, real inference and fixture checks, the existing 25% comparison allowance and 100 img/s baseline; no skips, threshold relaxation, shortened coverage or retry-until-pass wrapper. Diagnostic benchmark runs are not acceptance proof.

Run repeated uncached focused checks, the unchanged full offline check, and ordinary WASM regressions. If WASM parity execution or process control changes, also run the enabled three-runtime gate using the existing locked Playwright/Firefox prerequisites from `validation/wasm/README.md`; a missing prerequisite or opt-in skip is not parity proof. No enabled browser gate is required for a scaling-test-only repair. Report any independent LK failure without editing its live task or claiming it is fixed here.

Update Evidence on this task ID for later bundles or isolated reruns; never queue a second task for this failure. If investigation supports only a product/inference change or no justified repair within the authorized boundaries, report blocked with diagnostic evidence rather than widen scope. No repair or closure is claimed by this maintainer pass.

Maintainer verification on unchanged product HEAD `cd69673e3138fa2523ff02f9021b4cb476cd49a8`: `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` exited 0; `shared/neural` ran uncached (164.895s). Ephemeral local log: `/tmp/pi-docs-maintainer-m3-full-check.log`, not committed. This one later full-suite pass is not an isolated rerun, proof of repair or three-revision flake evidence. Independent read-only review found no critical/important issues; its cause-neutral scope clarification was applied.

## Maintainer checklist

- [x] Read all five bundle files in place; compare commits, failing test and existing task sources.
- [x] Confirm scaling is uncovered and preserve the live LK task unchanged.
- [x] Validate ledger, source references and scoped diff; run the required repository check before committing.

TDD: N/A for the maintainer change: task documentation only, with no executable behavior changed. Worker behavior changes still require RED/GREEN.
