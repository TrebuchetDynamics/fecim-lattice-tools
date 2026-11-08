+++
schema = 1
id = "QA-PACKAGE-TERMINALS"
title = "Reject incomplete Go test streams instead of counting individual tests as packages"
short = "Require real package completion in QA summaries"
status = "done"
goal = "OFFLINE-QA"
section = "Next"
rank = 1000
depends_on = []
touches = ["scripts/gotest_pkgsum/main.go", "scripts/gotest_pkgsum/main_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Repair only the offline go-test JSON summarizer and focused tests. Distinguish test/subtest events from package terminal events, count only packages with real terminal summaries, and return a nonzero incomplete-stream diagnostic naming observed packages that never completed, including mixed streams with other completed packages. Preserve PKG_SUM fields for complete streams, failure dominance, noise tolerance, stdin/file interfaces and all caller exit/coverage protections. Start with deterministic failing automated fixtures before the minimum production fix. Benefit: truthful qa-a0 package coverage and retained test evidence after interrupted/partial capture, without changing scientific computation."
acceptance = ["Given individual test/subtest pass or skip without package completion, when runPkgSum reads stdin or a file, then it cannot return success or count these as completed passing/skipped packages. Add table-driven RED fixtures including a completed package plus another started package with only test-level pass/skip; GREEN returns nonzero, names the incomplete package and makes no all-complete claim.", "Given complete interleaved streams with package pass/fail/skip and test/subtest events, when summarized, then PKG_SUM counts package terminal events only; skipped tests inside a passing package do not make it skipped, any observed failure remains non-success, and package failure dominance stays intact. Preserve existing noise/open/scan/empty-input handling and complete-stream output format with focused tests.", "Given unchanged make qa-a0 -> scripts/qa_a0.sh semantics, when executable file/stdin fixtures run, then complete streams retain output/exit behavior and incomplete streams fail. Preserve all caller test-exit and LIST_TOTAL-vs-JSON_TOTAL protections without changing caller/workload. Record focused automated RED/GREEN and final checks. This record is not implementation or qualification."]
checks = ["go test ./scripts/gotest_pkgsum -count=1", "go test -race ./scripts/gotest_pkgsum -count=1", "go test ./scripts -count=1", "git diff --check", "env CGO_ENABLED=1 xvfb-run -a go test ./... -timeout=10m"]
exclusions = ["No physics/model/default/golden/tolerance changes, sample/coverage reductions, new dependencies/toolchains, weakened caller checks, CI/launch-policy changes, global serialization, instruction/task-history edits, UI features or qualification receipts.", "Offline building only: no network/deployment/publication/push, installed EDA tools, real Orca claims, Live authorization, service/runtime changes or new planning chain. Preserve all owner-owned tasks and HYS/WEB final planning stops."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/QA-PACKAGE-TERMINALS.toml"
+++

- Sources: README.md#validation-and-development; Makefile (qa-a0); scripts/qa_a0.sh; scripts/gotest_pkgsum/main.go; scripts/gotest_pkgsum/main_test.go

## Evidence and independence

At discovery HEAD 09ae0bc3982849e8309254fc32afcf9e9b3cbbe1, TestEvent has no Test field and runPkgSum accepts any pass/fail/skip carrying Package. Existing tests cover noise, package failure and missing-file handling, not test versus package completion.

Read-only reproduction: `printf '%s\n' '{"Action":"pass","Package":"example/interrupted","Test":"TestFirst"}' | go run ./scripts/gotest_pkgsum -` incorrectly prints `PKG_SUM pass=1 fail=0 skip=0 total=1` and exits 0 despite no package completion. Baseline `go test ./scripts/gotest_pkgsum -count=1` passes. This discovery evidence is not the worker's required automated RED/GREEN.

Actual user flow is Makefile qa-a0 -> scripts/qa_a0.sh, which consumes PKG_SUM and compares JSON_TOTAL with go list. The wrapper also preserves go test exit, so this is incorrect completion/count reporting, not a claim that it universally turns failed runs green.

Runner/parallel status reports zero active workers and no integration backlog; four owner-owned qualification obligations remain. All inspected slot records are closed. git worktree list/status shows no modified selected files in any extant worktree. Main has Hysteresis/Plot/engine/style/e2e edits; other worktrees preserve physics/parity, citation, EDA-export and staged-proposal edits, including agent-test-loop.sh. None overlap the two Touches. Recheck live ownership before implementation.

All 54 existing task scopes were inspected; none owns this package-summary defect. MAIN-CHECK is the historical named full-check regression frontier. A separate low-priority unfocused OFFLINE-QA goal avoids widening it or the stopped WEB/HYS planning chains. No missing prerequisites, credentials, fixture acquisition or Live operation.

Root AGENTS.md/README.md/CLAUDE.md, TODO policies, BLOCKERS.md and current tasks/goals were read. No separate repository standing decision-policy file was found; owner-only, no-publication, simulation-only and final-planning-stop restrictions remain.

TDD for discovery: N/A (new task/goal metadata only). Worker must add failing automated fixtures first, record RED, then implement and record GREEN/final checks.
