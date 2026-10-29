# MAIN-DEMO-STEPS: restore the full-check demo completion regression

## Contract and evidence

The unattended worker restores `TestDemoControllerRunsAllSteps` in the unchanged full offline check. This is an existing compatibility check, not authorization for new Fyne work or a launch-policy change.

Read the original bundle **in place**, never copy it into a worktree or commit:

`/home/xel/git/fecim-lattice-tools/.git/autowork-execution/main-evidence/b2c1a0374d09490a64016c16833126209c58a6e6`

- `main_check.json`: `red`, `TEST_FAILURE`, checked revision `b2c1a0374d09490a64016c16833126209c58a6e6`, last green `ecaa522c380993cc8fb78b262a5c4358ceb0642a`; sole listed failure `TestDemoControllerRunsAllSteps`.
- `main-check.log:141-144`: `demo_controller_test.go:88: Non-looping demo should stop after all steps`; `shared/widgets` fails after 3.714s. The full log reaches its final `FAIL`; this is not a crashed check.
- `main-check.exit`: `1`. `main-check-reruns.log` is empty; `flaky_history.json` is `{}`. There are no isolated reruns or three-revision flake evidence in this bundle. Do not reclassify main as flaky from the timing hypothesis or a later pass.
- Only commit since green: `b2c1a037 fix: retain private planning locally and refresh public source guidance`. Its changed paths are `.gitignore`, repository guidance/release/README documents, removed private planning files, and `scripts/public-release/test_private_ignore_policy.py` / `test_public_release_boundary.py`. Neither demo source nor its tests changed. Temporal association does not establish causation; these unrelated privacy/publication paths are not repair targets.

No existing open task covers this failure. The existing `MAIN-CHECK` goal remains unmet; earlier MAIN tasks and their closures are not reopened. There is no worker report for this new task and no source-backed producer prerequisite, so no dependency or resource reservation is added.

## Source boundaries and investigation

Docs map: the MAIN bundle supplies the failure; `AGENTS.md` supplies the Full check; `docs/internals/testing/TESTING.md#tdd-hard-rule` supplies RED/GREEN requirements. The failing test and its controller supply the bounded repair surface:

- `shared/widgets/demo_controller_test.go:65-102`: three actions each increment an atomic counter, each step has a 10ms duration, and a fixed 50ms sleep precedes the stopped-state and exactly-once assertions. A 20ms nominal margin is not evidence that the goroutine has finished under full-suite resource load.
- `shared/widgets/demo_controller.go:157-248`: `run` performs each action and callback before its duration wait, then calls `Stop` after the final wait for non-looping playback. `Stop` updates the running state before invoking `onStop` (`:102-123`). Existing callbacks and `IsRunning` provide observable completion without requiring a new production API.

Investigate isolation, package/test-order effects, resource load and fixture timing before deciding whether the defect is in the test or the existing controller lifecycle. Start with the test fixture: a bounded completion observation using existing state/callbacks can distinguish natural completion from the test merely waking up. A regression must still reject missing/duplicate actions or failure to auto-stop; cleanup calling `Stop` must not satisfy the natural-completion assertion. Use a controlled delayed-action/completion reproducer if uncontrolled host scheduling does not reproduce reliably. Do not require a coordinator receipt or an owner-supplied fixture.

Touch only `shared/widgets/demo_controller_test.go` by default. `shared/widgets/demo_controller.go` is a conditional target solely for a demonstrated existing lifecycle defect causing this failure, with a failing regression first; do not change production durations or lifecycle behavior merely to accommodate a sleeping test. Repairing an existing fixture/lifecycle defect within these boundaries is allowed. Preserve the public API, sequential actions, per-step waits, non-looping auto-stop and looping/manual-stop behavior. No scientific computation, material/default, unrelated widget, release/privacy path, dependency or launch-policy changes.

## Acceptance and verification

1. Record a controlled failing automated regression for the diagnosed timing/isolation/lifecycle defect and the minimum GREEN repair. Preserve all three exactly-once action assertions and prove natural non-looping completion. Use bounded observable synchronization and reliable cleanup; no arbitrary sleep enlargement, skips, dropped assertions, global suite serialization or retry-until-pass wrapper.
2. The focused test passes repeatedly uncached, the widgets package passes uncached, focused shuffled controller tests and race checks pass, and the unchanged Full check completes and passes. Report unrelated failures independently without changing their task contracts. Isolated passes are diagnostic evidence, not proof of a three-revision flake or a product defect.

Commands:

```sh
env CGO_ENABLED=1 xvfb-run -a go test ./shared/widgets -run '^TestDemoControllerRunsAllSteps$' -count=100 -timeout=10m
env CGO_ENABLED=1 xvfb-run -a go test ./shared/widgets -count=1 -timeout=10m
env CGO_ENABLED=1 xvfb-run -a go test ./shared/widgets -run '^TestDemoController' -shuffle=on -count=20 -timeout=10m
env CGO_ENABLED=1 xvfb-run -a go test -race ./shared/widgets -run '^TestDemoController' -count=20 -timeout=10m
env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m
git diff --check
```

Update this task's Evidence with later revision bundles and isolated reruns; do not add another task for the same failure. Refresh source references if the named test/controller changes. This contract queues a repair; it does not claim repair, main-green status or goal qualification.

## Maintainer verification

On unchanged product HEAD `0ff5bb24715eb201d184bb0e2349549fa4a60589`, `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` exited 0 during this docs-only pass; `shared/widgets` ran uncached in 2.284s. Ephemeral log: `/tmp/pi-docs-maintainer-demo-full-check.log` (not committed). This later full-suite pass is not an isolated rerun, a repair or three-revision flake proof; retain the original red evidence and investigate. Independent read-only review approved the contract with no findings. Task-ledger validation passes with 54 tasks and 10 goals.

## Maintainer checklist

- [x] Read all five original bundle files in place; inspect the failing test, controller, changed commit paths and existing open task records.
- [x] Confirm no existing open task covers this failure; retain closures, goal status and dependencies.
- [x] Validate the task ledger, source references and scoped diff; run the repository Full check before committing.

TDD: N/A for this maintainer change: task documentation only; executable behavior is unchanged. The worker's repair still requires RED/GREEN.
