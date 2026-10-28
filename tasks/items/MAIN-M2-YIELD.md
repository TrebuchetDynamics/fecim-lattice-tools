+++
schema = 1
id = "MAIN-M2-YIELD"
title = "Restore the full-check process-variation yield regression"
short = "Repair process-variation yield validation"
status = "done"
goal = "MAIN-CHECK"
section = "Now"
rank = 1000
depends_on = []
touches = ["shared/crossbar/process_variation_yield_test.go", "validation/wasm/hysteresis_parity_test.go", "validation/wasm/parity_test.go", "validation/wasm/hysteresis_harness.mjs", "validation/wasm/harness.mjs"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Diagnose TestM2_PV_03_YieldDecreasesWithNoiseLevel from the named main bundle; investigate RNG/replay, finite-sample saturation, isolation, resource load and fixture timing. Make only the RED-backed minimum validation/fixture repair in the named paths. WASM validation edits require demonstrated causal interference; scientific production paths are read-only diagnostic sources. See docs/tasks/main-m2-yield.md."
acceptance = ["A deterministic failing automated regression exposes the diagnosed validation/fixture defect before the minimum repair; record causal evidence and RED/GREEN. Preserve the 8x8 workload, 0.85 weights/unit inputs, sigmas 0.01/0.10/0.30, 200 trials, 12-bit ADC/DAC, 300 K, ideal/variation comparison, IR-drop/sneak disabled, per-output less-than-10% relative-error yield, strict yield ordering and both existing 0.5 sanity bounds. Controlled fixtures plus separate replayable stochastic coverage are allowed with evidence, not lucky-seed selection, skips, loosened thresholds, removed coverage or retry-until-pass.", "The focused test passes repeatedly uncached, shared/crossbar passes uncached and the unchanged Full check completes and passes. If WASM validation changes, preserve exact native/Node/Firefox bytes, production lifecycle/budgets, original example parity and owned-process cleanup and run the enabled gate with locked prerequisites. Report revision/rerun evidence honestly: the original bundle contains no isolated reruns or three-revision flake evidence. Report unrelated failures without altering other task contracts."]
checks = ["env CGO_ENABLED=1 go test ./shared/crossbar -run '^TestM2_PV_03_YieldDecreasesWithNoiseLevel$' -count=20 -v", "env CGO_ENABLED=1 go test ./shared/crossbar -count=1", "env CGO_ENABLED=1 go test ./validation/wasm -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m", "If validation/wasm paths change: provision locked prerequisites per validation/wasm/README.md; env FECIM_WASM_PARITY=1 CGO_ENABLED=1 go test ./validation/wasm -count=1 -v"]
exclusions = ["No scientific/numerical/physics/default/material/quantization/golden or production RNG/API changes; no weakened assertions/coverage, global full-check serialization, new dependencies, browser features, launch/build-policy changes or task closure. WASM edits only for demonstrated interference causing this failure, never weakened parity/lifecycle/budgets/cleanup. Preserve all existing task contracts, dependencies and closures."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/MAIN-M2-YIELD.toml"
+++

- Sources: docs/tasks/main-m2-yield.md#contract-and-evidence
- Evidence: Bundle (read in place): /home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/3242627522d34812fd956325c7b5b3f43a834564. main_check.json: state red, TEST_FAILURE, checked revision 3242627522d34812fd956325c7b5b3f43a834564, last green 9bba4b85c86d64682af05ad2aba9ff2df0a8e7db; sole failure TestM2_PV_03_YieldDecreasesWithNoiseLevel. process_variation_yield_test.go:71 rejects yields 1.000 (200/200), 1.000 (200/200), 0.025 (5/200) at sigmas 0.01/0.10/0.30. shared/crossbar FAIL (13.656s); full run ends FAIL. main-check.exit=1; main-check-reruns.log empty; flaky_history.json={}: no isolated reruns or three-revision flake evidence. Commits since green: 32426275 docs(tasks): close HYS-WORKER (pi-autowork); 6bacc1f1 Merge commit f85cc9671cd1c813f16ce85e7bea2210a26c494a into HEAD; f85cc967 feat: qualify production hysteresis Worker across Go Node and Firefox. No crossbar files changed; no causal attribution. HYS-WORKER report: .git/autowork-execution/slot-data/HYS-WORKER/attempt-1/report.json. Source trace and conditional verification: docs/tasks/main-m2-yield.md.
- Maintainer verification: On unchanged product HEAD 3242627522d34812fd956325c7b5b3f43a834564, env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m exited 0 during this docs-only pass; shared/crossbar ran uncached in 12.481s. Ephemeral local log /tmp/pi-docs-maintainer-m2-yield-full-check.log is not committed. This later full-suite pass is not an isolated rerun, proof of repair or three-revision flake evidence; retain the original red bundle. Independent read-only review found only a misleading rank note, now removed; authoritative section Now/rank 1000 selects this sole open Now task first.
