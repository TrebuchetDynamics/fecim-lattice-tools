+++
schema = 1
id = "T-012-f1aebb7e"
title = "Generate reports through confined root handles"
short = "Confine report output"
status = "done"
goal = "BWC-STORAGE"
section = "Now"
rank = 2000
depends_on = []
touches = ["workbench/report/report.go", "workbench/report/report_test.go"]
resources = []
kind = "implementation"
ownership = "worker"
scope = "Add report.GenerateRoot(root *os.Root, bundle project.Bundle, analysis experiment.Analysis) error over existing WriteJSON/CSV/Markdown serializers; keep Generate as path adapter. All namespace/temp/write/sync/rename/cleanup I/O uses pinned root handles, never Bundle.Root/Root.Name or global temporary files. Preserve existing per-file atomicity and previous report on precommit failure; no all-report-set transaction promise."
acceptance = ["Given outside reports links, misleading Bundle.Root, deterministic retarget/precommit failure and an existing report marker, when GenerateRoot runs, then outside data/previous output stay intact, only its own scratch is removed and caller root remains open. Prove in workbench/report/report_test.go::TestGenerateRootConfinement (new, first RED).", "Given valid identical bundle/analysis, when rooted or legacy report output is generated, then existing deterministic scientific JSON/CSV/Markdown bytes and writer-error behavior remain compatible. Prove in TestGenerateRootConfinement and existing deterministic/generation/writer-error tests."]
checks = ["CGO_ENABLED=0 go test ./workbench/report -run 'TestGenerateRootConfinement|TestReportsAreDeterministicAndExcludeRuntimeMetadata|TestGenerateWritesThreeReports|TestWriteMarkdownStopsAtFirstWriterError' -count=1", "go test -race ./workbench/report -count=1"]
exclusions = ["No copied serializers/scientific analysis, HTTP/UI, run-store/editor changes, all-file transaction/power-loss guarantee, dependencies/manifests/host changes."]
closure = "coordinator_verified"
closure_evidence = "tasks/evidence/T-012-f1aebb7e.toml"
+++
- Sources: docs/internals/superpowers/plans/2026-10-25-browser-rooted-persistence.md#accepted-rooted-persistence; docs/internals/superpowers/plans/2026-10-25-browser-rooted-persistence.md#s2-reports.
- Authority: Planning only; no implementation, commit, dispatch, hosted compute or live authorization granted. A later explicit execution instruction is required.
- Invariants: Preserve model equations/values, scientific schema/run identities/canonical records, evidence, cancellation/progress and existing CLI success/error contracts. Fyne is fully deprecated; no Fyne feature/parity work. Existing Full check and relevant race gates remain required at execution closeout.
- Checks status: New focused selectors are proposed/unrun. Existing CGO_ENABLED=0 go test ./workbench/... -count=1 passed for seven packages during planning, not full-suite/security/Orca acceptance.
- Produces: report.GenerateRoot(root *os.Root,bundle project.Bundle,analysis experiment.Analysis) error. Existing Generate signature delegates through owned handles; supplied handles remain open.
- Planner calls: Same-root exclusive scratch, sync/close then root-relative rename; no Root.Name path fallback. Existing per-output atomicity, not an all-three-output transaction. Worker chooses cohesive private hook placement.
- Next action: Add TestGenerateRootConfinement first, run focused RED, then minimally share existing serializers through rooted persistence.
