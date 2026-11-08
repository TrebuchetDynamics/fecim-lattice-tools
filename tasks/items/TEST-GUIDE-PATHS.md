+++
schema = 1
id = "TEST-GUIDE-PATHS"
title = "Repair nonexistent go test package paths in the offline test guide"
short = "Fix stale test paths in the test guide"
status = "done"
goal = "DEV-TOOLING"
section = "Next"
rank = 2000
depends_on = []
touches = ["docs/internals/testing/TEST_GUIDE.md", "scripts/testing_doc_paths_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Correct only the `go test` package selectors in docs/internals/testing/TEST_GUIDE.md that no longer resolve, and add one focused offline guard. Module-prefixed paths: ./crossbar/pkg/crossbar/... -> ./shared/crossbar/... (Phase 0 moved the engine), ./eda/pkg/export/... -> ./module6-eda/pkg/export/..., ./mnist/pkg/core/... -> ./module3-mnist/pkg/core/..., ./hysteresis/pkg/ferroelectric/... and ./hysteresis/pkg/controller/... -> ./module1-hysteresis/.... GUI commands in section H: the E2EGUI/VisualXvfb/UICrawler/LayoutAudit commands must target ./cmd/fecim-lattice-tools-fyne/... where those tests live, not ./cmd/fecim-lattice-tools/... which selects no tests and exits 0. Keep the headless ./cmd/fecim-lattice-tools/... -run E2E and -run Integration commands (those tests exist there), the ./shared/physics, ./shared and ./validation selectors, and the FECIM_RUN_XVFB/FECIM_UI_CRAWL/FECIM_LAYOUT_AUDIT markers. Add one offline test that extracts the selectors and fails on a nonexistent package or a GUI selector that matches no test. Benefit: a reader or agent following the guide runs real tests instead of a missing package or a silent no-tests pass."
acceptance = ["Every `go test` package selector in docs/internals/testing/TEST_GUIDE.md resolves to an existing Go package: ./crossbar/pkg/crossbar/... becomes ./shared/crossbar/... (Phase 0 moved the crossbar engine), ./eda/pkg/export/... becomes ./module6-eda/pkg/export/..., ./mnist/pkg/core/... becomes ./module3-mnist/pkg/core/..., and the two ./hysteresis/pkg/... selectors become ./module1-hysteresis/pkg/...; the ./shared/physics, ./shared and ./validation selectors are unchanged.", "The GUI E2E/visual/crawler/layout commands in section H target ./cmd/fecim-lattice-tools-fyne/... (where TestE2EGUI, VisualXvfb, UICrawler and LayoutAudit live) instead of ./cmd/fecim-lattice-tools/..., which currently prints `[no tests to run]` and exits 0; the headless E2E and Integration commands that do exist under cmd/fecim-lattice-tools, and the FECIM_RUN_XVFB/FECIM_UI_CRAWL/FECIM_LAYOUT_AUDIT env markers, are preserved.", "A new offline scripts/testing_doc_paths_test.go extracts the `go test ... ./<selector>` tokens from docs/internals/testing/TEST_GUIDE.md, requires each to resolve with `go list`, and requires the E2EGUI/VisualXvfb/UICrawler/LayoutAudit selectors to name a package whose test list contains a matching test; it fails against the current file (RED) and passes after the repair (GREEN). It must not read docs/internals/testing/TESTING.md.", "No behavior, source, test, generated file, script, CI, Makefile or other documentation changes; TESTING.md and the module AGENTS guides are untouched. `go test ./scripts -count=1` and the final `env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m` pass."]
checks = ["go test ./scripts -run 'TestTestingDocPaths' -count=1", "go test ./scripts -count=1", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m", "git diff --check"]
exclusions = ["Only docs/internals/testing/TEST_GUIDE.md and the new scripts/testing_doc_paths_test.go. Do not edit TESTING.md, module AGENTS guides, docs/modules/*, docs/internals/automation/*, scripts, Makefile, CI or any Go/source file, and do not change physics, thresholds, goldens or test behavior.", "Offline building only: no network, deployment, push, Live authorization, runtime/service change, launch-default switch or planning chain. Preserve the four owner-only qualification tasks (HYS-ORCA, T-013-a31914c8, WEB-ORCA, WEB-TOOLS) and the HYS/WEB final planning stops. Recheck actual file/worktree ownership before implementation; docs/internals/testing/TESTING.md is held by the split-eda-export and staged-proposals-validation worktrees, and docs/internals/testing/physics-race-lane.md and widget-progress-flake.md by staged-proposals-validation."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/TEST-GUIDE-PATHS.toml"
+++
- Sources: docs/internals/testing/TEST_GUIDE.md#how-to-run-each-category; docs/internals/architecture/shared-refactor-plan.md#phase-0

## Discovery evidence and user benefit

Round 21, current source at cb6ce710. docs/internals/testing/TEST_GUIDE.md is the canonical how-to-run-the-suite guide, but ten of its package selectors do not exist, so the commands fail (`directory not found`) or silently report success with no tests.

Missing module-prefixed paths (checked with `go list`):
- line 74/146/330: ./crossbar/pkg/crossbar/... -> the Phase 0 refactor moved the engine to shared/crossbar/; module2-crossbar/pkg/crossbar/ now holds only AGENTS.md.
- line 75: ./eda/pkg/export/... -> module6-eda/pkg/export/...
- line 91: ./mnist/pkg/core/... -> module3-mnist/pkg/core/...
- lines 105/106/324: ./hysteresis/pkg/ferroelectric/... and ./hysteresis/pkg/controller/... -> module1-hysteresis/...

Silent GUI false pass (line 160/166/167/168): `go test ./cmd/fecim-lattice-tools/... -run 'E2EGUI'|'VisualXvfb'|'UICrawler'|'LayoutAudit'` prints `ok ... [no tests to run]` and exits 0, because TestE2EGUI*, TestVisualXvfb*, TestUICrawler* and TestLayoutAudit* live in cmd/fecim-lattice-tools-fyne. `go test ./cmd/fecim-lattice-tools-fyne -list 'E2EGUI|VisualXvfb|UICrawler|LayoutAudit'` lists 12 matching tests. The headless `-run E2E` on cmd/fecim-lattice-tools does select TestE2E* there, so only the GUI commands change.

Benefit: an agent or contributor following the guide executes the intended test package instead of a missing path or a green no-op.

Read-only reproduction:
- `go list ./crossbar/pkg/crossbar/... ./eda/pkg/export/... ./mnist/pkg/core/... ./hysteresis/pkg/ferroelectric/... ./hysteresis/pkg/controller/...` -> matched no packages / errors
- `go test ./cmd/fecim-lattice-tools/... -run 'LayoutAudit' -count=1` -> `ok ... [no tests to run]`, exit 0
- `go test ./cmd/fecim-lattice-tools-fyne -list 'E2EGUI|VisualXvfb|UICrawler|LayoutAudit'` -> 12 tests

## Admission and boundaries

Read AGENTS.md, README.md, CLAUDE.md, docs/internals/TODO.md Policies, BLOCKERS.md, task/goal records and runner/worktree state. The runner service is active (running, paused_until=0); the /home/xel/.config/pi-autowork/fecim-lattice-tools.paused marker is stale pause state contradicted by the running service (rounds 18-20). The four open Now tasks are owner-only and remain intact. No existing task or goal owns these paths; TEST_GUIDE.md and the new test file are unclaimed in every extant worktree. The referenced corrections (the shared-refactor plan Phase 0 row and the AGENTS.md modules already using moduleN- prefixes) are normative. Offline only: no network, tool install, dataset, Live authorization or unavailable input.

TDD for queueing: N/A (new task metadata only). The worker must add the failing scripts/testing_doc_paths_test.go first, record RED against the current guide, then repair the selectors and record GREEN plus the final checks.