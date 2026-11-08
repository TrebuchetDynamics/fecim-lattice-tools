+++
schema = 1
id = "RACE-LANE-PACKAGES"
title = "Run the documented headless race lane on existing packages"
short = "Fix stale race-lane package list"
status = "open"
goal = "DEV-TOOLING"
section = "Next"
rank = 1000
depends_on = []
touches = ["scripts/ci/go-test-race.sh", "scripts/race_packages_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Repair only the RACE_PKGS array in scripts/ci/go-test-race.sh so every entry names a package that exists in the current tree, and add one focused offline guard. The Phase 0 shared refactor renamed shared/utils to shared/canvas (same package, new import path) and migrated module2-crossbar/pkg/crossbar to shared/crossbar, but RACE_PKGS still lists the old paths. `go test -race ./shared/utils` fails with `directory not found` and `go test -race ./module2-crossbar/pkg/crossbar` fails with `no Go files`, so `exec go test ... \"${RACE_PKGS[@]}\"` fails at package setup before any race test runs. Replace those two entries with ./shared/canvas and ./shared/crossbar; preserve the other 21 entries, the -tags=ci -race -count=1 -shuffle=off -trimpath flags, the headless gate and the timeout. Add one offline test that extracts RACE_PKGS and fails on any entry that does not resolve. Benefit: the documented headless race lane actually executes the intended race checks instead of failing on stale paths."
acceptance = ["Every RACE_PKGS entry in scripts/ci/go-test-race.sh resolves to an existing Go package: the stale ./shared/utils becomes ./shared/canvas (Phase 0 renamed shared/utils) and ./module2-crossbar/pkg/crossbar becomes ./shared/crossbar (Phase 0 migrated the crossbar engine); the other 21 entries, the -tags=ci -race -count=1 -shuffle=off -trimpath flags, the headless DISPLAY/WAYLAND gate and the timeout are unchanged.", "A new offline scripts/race_packages_test.go reads the RACE_PKGS array, fails when any entry does not resolve with `go list`, and fails on an empty list or duplicate entries; it fails against the current script (RED) and passes after the path repair (GREEN).", "No package, source, test, CI workflow, Makefile or documentation file is added or changed beyond the script and the new test; no tool is installed and no race/qualification claim is made. `go test ./scripts -count=1` and the final `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` pass."]
checks = ["go test ./scripts -run 'TestRaceLanePackages' -count=1", "go test ./scripts -count=1", "env -u DISPLAY -u WAYLAND_DISPLAY go test -race -run '^$' ./shared/canvas ./shared/crossbar -count=1", "bash -n scripts/ci/go-test-race.sh", "git diff --check", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m"]
exclusions = ["Only scripts/ci/go-test-race.sh and the new scripts/race_packages_test.go. No new/renamed package, no shared/canvas or shared/crossbar source/test edit, no CI workflow, Makefile, TEST_GUIDE or other script change, no tool installation, and no claim that any race lane passed in hosted CI or that a feature is qualified.", "Offline building only: no network, deployment, push, Live authorization, runtime/service change, launch-default switch or planning chain. Preserve the four owner-only qualification tasks (HYS-ORCA, T-013-a31914c8, WEB-ORCA, WEB-TOOLS) and the HYS/WEB final planning stops. Recheck actual file/worktree ownership before implementation; do not touch the split-eda-export or staged-proposals-validation worktrees."]
+++
- Sources: docs/internals/testing/TEST_GUIDE.md#how-to-run-each-category; scripts/AGENTS.md; docs/internals/automation/module4-automated-testing-plan.md#4.3-release-gate-commands-required

## Discovery evidence and user benefit

Round 21, current source at cb6ce710. scripts/ci/go-test-race.sh is the documented headless CI-like race lane. Its RACE_PKGS array still lists two pre-Phase-0 paths:

- `./shared/utils` -> `stat .../shared/utils: directory not found`. Commit 58d1a901 ("Phase 0b/11/12 interfaces + canvas rename + CI script") renamed shared/utils/ to shared/canvas/ (the package name `utils` is preserved; only the import path changed) and introduced this script.
- `./module2-crossbar/pkg/crossbar` -> `no Go files in .../module2-crossbar/pkg/crossbar`. The same refactor migrated the crossbar engine (~86 files) to shared/crossbar/.

`exec go test -race -tags=ci ... "${RACE_PKGS[@]}"` therefore fails at package setup (setup failed) before any race test executes, so the "required" lane never runs. `go list` resolves the other 21 entries, so only these two are stale. TEST_GUIDE.md ("Headless, material-aware parity lanes") and the module4 automation plan both direct users to this script; scripts/AGENTS.md documents it as the race gate.

Benefit: users and agents get the intended race coverage from the documented command instead of a setup failure.

Read-only reproduction:
- `go list ./shared/utils` -> directory not found; `go list ./module2-crossbar/pkg/crossbar` -> no Go files
- `go list ./shared/canvas ./shared/crossbar` -> both resolve
- `git show --find-renames=40% --stat 58d1a901 | rg 'shared/\{utils => canvas\}'` and the shared-refactor plan row 'Crossbar -> shared/crossbar/ ... Complete'

## Admission and boundaries

Read AGENTS.md, README.md, CLAUDE.md, docs/internals/TODO.md Policies, BLOCKERS.md, task/goal records and runner/worktree state. The runner service is active (running, paused_until=0) and dispatch continues; the /home/xel/.config/pi-autowork/fecim-lattice-tools.paused marker is stale pause state contradicted by the running service (rounds 18-20). The four open Now tasks are owner-only and remain intact. `rg -n 'RACE_PKGS|go-test-race' tasks/` returns nothing: no task or goal owns this script. scripts/ci/go-test-race.sh and the new test file are unclaimed in every extant worktree. Offline only: no network, tool install, dataset, Live authorization or unavailable input.

TDD for queueing: N/A (new task metadata only). The worker must add the failing scripts/race_packages_test.go first, record RED against the current script, then make the two-entry RACE_PKGS change and record GREEN plus the final checks.