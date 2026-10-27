# Main-check LK five-target repair

## Contract and evidence

The DOCS-MAINTAINER request authorizes one first-in-Now repair task for the uncovered main-check failure, not a product redesign or a claim that the WASM commit caused it. `MAIN-LK-5TARGETS` tracks that repair under the unmet `MAIN-CHECK` goal. All 19 pre-existing task records are closed; none covers this test. Their closures, dependencies and goal statuses remain untouched.

Read the original bundle **in place**, without copying it into the repository:

`/home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/7048b840b5fe39b49e75c37ef673e8e7bdc56f39`

- `main_check.json`: `state=red`, `result_class=TEST_FAILURE`; checked revision `7048b840b5fe39b49e75c37ef673e8e7bdc56f39`; last green `c5e4504f22645ef03835641b448cdbd99ad38d41`.
- Only commit since green: `7048b840 test(wasm): gate exact native Node and Firefox report parity`. Its changes are WASM validation code and browser architecture/planning documentation, not the failing command test or LK equations. Temporal association is not causation.
- `main-check.log:5212-5213`: `TestHeadlessLKRun_ISPP5Targets_NoNaNInf_AndEFieldUnitsConsistent` failed in `cmd/fecim-lattice-tools-fyne`, at `mode_lk_headless_ispp_5targets_test.go:111`: expected at least five distinct targets, got two (`[1 33]`). Package result is FAIL at lines 13177-13178; the full check reached results.
- `main-check.exit`: `1`.
- `main-check-reruns.log`: empty. `flaky_history.json`: `{}`. No isolated rerun or three-revision flake evidence exists in this bundle. Do not invent it or reclassify main on speculation.

## Source-backed investigation boundaries

- `AGENTS.md:326` supplies the Full check: `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m`. Preserve it.
- Both command packages have the same named five-target test and timestamp-based `newestHysteresisCSVAfter` helper (`mode_lk_headless_ispp_5targets_test.go:17-146`). Preserve finite-value and V/m-to-MV/cm assertions as well as target coverage.
- `cmd/fecim-lattice-tools-fyne/gui_test_main_test.go:45-55` and `cmd/fecim-lattice-tools/testmain_test.go:9-25` already isolate package log directories. Investigate any remaining per-test/environment/fixture interaction; do not assume packages share a directory.
- Both `mode.go` implementations parse target count and explicit target overrides, record headless snapshots and close loggers. The Fyne path uses `module1-hysteresis/pkg/gui/data_logger.go`; the other command uses `module1-hysteresis/pkg/headless/data_logger.go`. The GUI logger has bounded asynchronous delivery (`Record`, `Close`, `run`). Trace actual execution versus selected CSV completeness before selecting a fix.
- The changed `validation/wasm` files are investigation candidates for resource/process interference only. The normal parity test is opt-in (`parity_test.go:28-30`); preserve its exact-report comparison and process-cleanup contracts. The commit's architecture/planning files provide context, not authorization to change browser product scope.

## Verification and update trigger

The worker must record a deterministic RED before implementation, the minimal GREEN, repeated uncached focused runs in both commands and the unchanged Full check. If parity execution or process-control paths change, also run `env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -v` with the existing prerequisites provisioned as described in `validation/wasm/README.md`; ordinary Go tests skip that explicit gate. No enabled browser gate is required for an LK-only repair that leaves those paths unchanged. Existing defects in the named execution/logging/fixture paths may be repaired only as needed to restore this check, with no physics/default/tolerance/golden changes or weakened assertions. Investigate isolation, resource load and fixture timing even if an isolated rerun passes; report exact revisions and results.

Update the task Evidence if a later main bundle or isolated rerun becomes available. Reuse this task ID for the same failure; do not create a second repair/hardening task. This document is task planning, not proof of repair.

Maintainer verification on unchanged product HEAD `23405de887bfb5e17454105d35a8f48459f571e7` (the merge of `7048b840`): the exact Full check exited 0; both command packages ran uncached (35.961s and 47.821s). Ephemeral local output is `/tmp/pi-docs-maintainer-lk-full-check.log`, not a committed receipt or replacement for the original bundle. One later full-suite pass does not prove repair, establish three-revision flakiness or justify closing the task. Preserve the original failure and investigate intermittent target coverage. Independent read-only review identified the conditional parity-check requirement above; it is included in the task Checks.

TDD: N/A for this maintainer change: only task contracts and this evidence index change; no executable behavior is edited.
