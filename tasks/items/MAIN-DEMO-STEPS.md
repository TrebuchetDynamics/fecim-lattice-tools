+++
schema = 1
id = "MAIN-DEMO-STEPS"
title = "Restore the full-check demo completion regression"
short = "Repair demo completion check"
status = "done"
goal = "MAIN-CHECK"
section = "Now"
rank = 1000
depends_on = []
touches = ["shared/widgets/demo_controller_test.go", "shared/widgets/demo_controller.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Investigate and restore TestDemoControllerRunsAllSteps failing in the full offline check. Diagnose isolation, test-order effects, resource load and fixture timing first; minimally repair the demonstrated test/fixture defect, or an existing controller lifecycle defect only if causally demonstrated with a failing regression. Preserve API and playback semantics. No new feature, launch-policy or scientific change; see docs/tasks/main-demo-steps.md."
acceptance = ["Record a controlled failing automated regression for the diagnosed timing/isolation/lifecycle defect before the minimum GREEN repair. Preserve all three exactly-once action assertions and prove natural non-looping auto-stop with bounded observable completion and reliable cleanup; cleanup Stop must not satisfy completion. No arbitrary sleep enlargement, skips, removed assertions, global suite serialization or retry-until-pass wrapper. Production lifecycle edits require causal evidence and a failing regression; preserve sequential actions, per-step waits, public API and looping/manual-stop behavior.", "TestDemoControllerRunsAllSteps passes repeatedly uncached; shared/widgets passes uncached; focused shuffled controller and race checks pass; the unchanged Full check completes and passes. Report independent failures without altering other contracts. Preserve the original red bundle and distinguish later isolated passes from three-revision flake proof: the bundle contains no reruns and no flake history."]
checks = ["env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m", "env CGO_ENABLED=1 xvfb-run -a go test ./shared/widgets -run '^TestDemoControllerRunsAllSteps$' -count=100 -timeout=10m", "env CGO_ENABLED=1 xvfb-run -a go test ./shared/widgets -count=1 -timeout=10m", "env CGO_ENABLED=1 xvfb-run -a go test ./shared/widgets -run '^TestDemoController' -shuffle=on -count=20 -timeout=10m", "env CGO_ENABLED=1 xvfb-run -a go test -race ./shared/widgets -run '^TestDemoController' -count=20 -timeout=10m", "git diff --check"]
exclusions = ["No new UI/API/features, launch-policy, scientific/default/material changes, unrelated widgets, release/privacy edits, new dependencies, dropped assertions, arbitrary sleep enlargement or global serialization. Controller production edits only for a demonstrated existing lifecycle defect causing this failure. Preserve all task closures, goal statuses and dependencies."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/MAIN-DEMO-STEPS.toml"
+++
- Sources: docs/tasks/main-demo-steps.md#contract-and-evidence; AGENTS.md (Full check); docs/internals/testing/TESTING.md#tdd-hard-rule
- Evidence: Bundle (read in place): /home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/b2c1a0374d09490a64016c16833126209c58a6e6. main_check.json: state red, TEST_FAILURE, checked revision b2c1a0374d09490a64016c16833126209c58a6e6, last green ecaa522c380993cc8fb78b262a5c4358ceb0642a; sole listed failure TestDemoControllerRunsAllSteps. main-check.log: demo_controller_test.go:88 Non-looping demo should stop after all steps; shared/widgets FAIL (3.714s); full run reaches final FAIL. main-check.exit=1; main-check-reruns.log empty; flaky_history.json={}: no isolated reruns or three-revision flake evidence. Only commit since green: b2c1a037 fix: retain private planning locally and refresh public source guidance. No demo files changed; no causal attribution. Source trace and conditional repair boundaries: docs/tasks/main-demo-steps.md.
- Maintainer verification: On unchanged product HEAD 0ff5bb24715eb201d184bb0e2349549fa4a60589, env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m exited 0; shared/widgets ran uncached in 2.284s. Ephemeral log: /tmp/pi-docs-maintainer-demo-full-check.log (not committed). One later full-suite pass is not an isolated rerun, a repair or three-revision flake proof; retain original evidence. Independent read-only review approved with no findings. TDD: N/A for task documentation only; executable behavior unchanged.
