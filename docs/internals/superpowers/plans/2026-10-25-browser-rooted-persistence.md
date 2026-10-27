# Browser rooted persistence — bounded implementation frontier

Status: Planning only. No implementation, commits, worker dispatch, services or publication authorized. Earlier no-commit restrictions remain binding. Preserve the uncommitted full-Fyne-deprecation edits in the other two browser documents.

## Accepted rooted persistence

Status: Accepted

This is a planner-derived engineering prerequisite of the owner's approved [local browser security contract](../specs/2026-10-25-accessible-browser-workbench-design.md#authentication-confinement-and-threat-model), not a new product direction. The owner requires launch-authorized project roots and confinement throughout configuration, inputs, runs, reports and exports. Fyne is fully deprecated; this frontier creates no Fyne feature, support promise or replacement-parity gate. Keep Go, CLI and the approved React direction. Existing build/security obligations are not waived by deprecation.

**Goal:** Add caller-owned root-handle variants for existing-project run persistence, reports, selected-run exports and study orchestration, with no scientific-format changes and no path escape through metadata.

**Users/journey:** researchers run a validated local study, reuse immutable results, cancel/reconcile outcomes, regenerate reports and export a selected design without granting the eventual browser arbitrary host filesystem access.

**Non-goals:** HTTP/auth/SSE, React, project discovery/creation, downloads, filesystem sandboxing against privileged or same-user hostile software, WASM, new physics, a database, changes to model tolerances, full Fyne removal or release/JOSS activity.

**Anti-goals:** reopening Root.Name/Bundle.Root to perform rooted I/O; passing an untrusted absolute path into a Module 6 writer; changing run IDs or canonical artifacts for security plumbing; treating partial/unavailable results as successful zeroes; claiming this slice alone makes a browser server ready.

## Discovery and evidence

Inspected baseline: HEAD `258e8ab6610de17ece9f5cd76bd8f94e5ebf4c36` plus the working tree. At discovery, only the two earlier deprecation documentation edits were dirty; no workbench source was dirty. `autowork capabilities` reports prepare_api 4; prepare reports READY, 15 tasks, 5 goals, no runnable order or in-flight tasks; a runner exists but this invocation does not start, stop or authorize it.

| Finding | Classification / confidence | Source / disposition |
|---|---|---|
| Core foundations, lifecycle, policy and configuration-editing tasks are closed. | OBSERVED / high | Existing T-001..T-010 and HZO-REF-1..5 records; preserve their IDs/closures. Do not recreate those features. Goal statuses are not automatically promoted from closed queues. |
| Root-bound project loading and byte-preserving/revision-guarded YAML editing exist. | OBSERVED / high | `workbench/project/load.go`, `configuration.go`, `configuration_save.go`; previous BWC-CONFIG plan. |
| Run cache/store reads, commits and listing still use ordinary path-based os operations. | OBSERVED / high | `workbench/experiment/store.go`, `runner.go`, `load.go`; this is the next missing capability boundary, not a reproduced exploit. |
| Reports use global temporary files/rename paths; EDA selected export passes paths to Module 6 writers. | OBSERVED / high | `workbench/report/report.go:163-179,220-236`; `workbench/edaexport/selected.go:53-106`. |
| Study reconciliation opens bundle.Root repeatedly; root-safe project loading alone cannot confine it. | OBSERVED / high | `workbench/study/study.go:196-313`; requires rooted preflight, evaluation, reconciliation and report/export calls. |
| Existing generator functions can produce SPICE/Verilog/DEF strings without host file writes. | OBSERVED / high | Module 6 GenerateSPICE, GenerateVerilogWithDefaults, GenerateDEFWithDefaults, GenerateLEF, GenerateLiberty. Reuse their output, not a new serializer. |
| Verilog/DEF output embeds current time. | OBSERVED / high | `module6-eda/pkg/export/verilog.go:82`, `def.go:127`; do not promise cross-time byte identity or silently change generators in a confinement task. |
| Needed traversal-resistant operations exist in the current toolchain. | OBSERVED / high | `go doc os.Root`: Open/OpenRoot, Mkdir/MkdirAll, Rename, Remove/RemoveAll, Stat/Lstat. No Root.MkdirTemp exists; create unpredictable exclusive names through the root instead. Keep current module Go requirement; install/change no toolchain here. |
| Headless baseline passes in this occurrence. | OBSERVED / high | `CGO_ENABLED=0 go test ./workbench/... -count=1`: seven packages passed. Not a full-suite, race, browser, security audit or human Orca result. |
| The full browser, safe root-granted project creation and UI qualification are still missing. | DOCUMENTED / high | Approved browser spec/manual scripts. This batch handles existing-project persistence only. |

### Goal classes and priority

Accepted direction: simulation-only reproducible research workbench, Go core/CLI, local React UI, full Fyne deprecation, no hosted compute. Derived engineering goal: BWC-STORAGE, root-confine the already implemented persistence operations before exposing them through HTTP. Proposed/out of scope: hosted services, a scientific rewrite and a static WASM demo. Existing release/scientific evidence owners and their gates stay separate.

Planner's priority call: security prerequisite before frontend scaffolding. Four tasks form a complete small frontier with independently testable outputs; they do not attempt to decompose the whole browser product. Priority 3, no focus takeover; do not rewrite existing priorities or mark unrelated goals met. No automated successor planner is needed for this bounded goal.

## Decisions and interfaces

### Capability and compatibility rules

- A supplied `*os.Root` is the authority and remains caller-owned. Rooted methods reject nil/closed handles with an error, not panic; never close caller handles. Child handles opened internally are closed on every path.
- `Bundle.Root`, `root.Name()` and legacy returned paths are descriptive only within rooted operations. A test must deliberately set Bundle.Root to an outside marker directory and prove rooted work never reads/writes there.
- Legacy string APIs remain source-compatible adapters: explicitly open their authorized root(s), invoke common rooted implementation, close owned handles and preserve return/error/output conventions. No duplicate scientific pipelines or root.Name-based fallback.
- Pin generated namespace directories with child OpenRoot handles before multi-step I/O. Use root-relative reads, exclusive writes and renames throughout. Validate fixed names/run IDs first. Reject escaping links and invalid namespace types; reads retain strict JSON/hash/identity/size validation. Do not weaken integrity checks to make a rooted load succeed.
- No host-global temp directory. For runs/reports, create cryptographically unpredictable names using root.Mkdir or root.OpenFile(O_CREATE|O_EXCL), retry collisions, then operate through pinned directory handles. Keep existing atomic per-run-directory / per-report-file commit points. Sync/close before publication; cleanup only scratch created by the current call. Preserve previous committed records/files on precommit error. Do not claim an all-three-reports transaction or power-loss durability.
- OS root traversal does not prevent bind mounts, hard-link attacks, special devices or privileged/same-user malware. Initial qualification is Linux (the approved Firefox/Orca target); do not claim GOOS=js confinement. Use deterministic link/retarget/failure barriers for tests, never a probabilistic sleep race.

### Decision map

| Choice | Selected / reason | Rejected / affected proof |
|---|---|---|
| Filesystem boundary | Native os.Root handles reused from existing project/editor APIs; operations remain in their owning packages. | A generic filesystem framework adds an unneeded abstraction. S1-S4 confinement tests. |
| Rooted experiment entry | `experiment.RunRoot(ctx context.Context, root *os.Root, bundle project.Bundle, opts RunOptions) (Summary,error)`. Same evaluator/options/progress and ordered result contract as Run. | Adding an optional path/handle switch into serialized Bundle conflates authority and scientific data. S1 tests metadata mismatch. |
| Rooted reads | `LoadRunRecordRoot(root *os.Root,id string) (RunRecord,bool,error)`, `LoadRunRoot(root *os.Root,id string) (RunRecord,error)`, `LoadRunsRoot(root *os.Root) ([]RunRecord,error)`. | Reopening string paths would undo confinement during reconciliation/export. S1/S3/S4. |
| Report boundary | `report.GenerateRoot(root *os.Root,bundle project.Bundle,analysis experiment.Analysis) error` shares existing WriteJSON/CSV/Markdown serializers. | Browser or alternate report math would fork scientific truth. S2 byte/report regression tests. |
| EDA boundary | `edaexport.ExportSelectedRunRoot(root *os.Root,id string) (string,error)` returns relative `exports/<id>`; string wrapper returns its existing joined path. | `/proc/self/fd` tricks and arbitrary Module 6 path writers are not the portability/security contract. S3 uses pure Generate functions plus rooted writes. |
| Study entry points | `RunRoot(ctx context.Context,root,citations *os.Root,opts RunOptions) (RunOutcome,error)`, `GenerateReportRoot(root,citations *os.Root) (experiment.Analysis,error)`, `ValidateProjectRoot(root,citations *os.Root) (project.Bundle,int,error)`, `ExportRunRoot(root *os.Root,id string) (string,error)`. | Calling string APIs within a rooted study would reopen authority. S4 inspects whole operation. |
| Citation option conflict | Rooted study rejects nonempty opts.CitationDir; caller supplies the authorized citation handle or nil. Legacy wrapper opens that directory and clears its internal copy of CitationDir. | Silently privileging a browser-supplied path over its handle is forbidden. S4 negative case. |
| Partial/error behavior | Preserve existing RunOutcome.ResultsState/AnalysisAvailable, current-sweep expected-ID reconciliation, reused flags, nonnil errors and nonblocking progress. | No regression to empty outcome on error or fabricated successful complete science. S4 regressions. |
| Export publication | Preserve exclusive creation/refusal of existing export destination; write via pinned export subroot and manifest last; remove only this call's incomplete new destination on failure. | An unconditional rename could replace an existing empty destination; do not invent no-replace atomic guarantees absent from the primitive. S3 existing-destination/failure tests. |

## S1 Run store

Outcome: root-handle experiment execution, exact-record loading, successful-run loading and ordered listing, retaining cache reuse and canonical bytes. Rooted core store must have no string-path escape. Adapt private old store fixtures/hooks if necessary without changing their logical corruption/rename-winner coverage.

Produces all experiment signatures in the decision map. Run opens bundle.Root as a legacy adapter after argument validation; RunRoot ignores Bundle.Root for I/O. Every cache/manifest/result/design read and scratch/commit/cleanup/listing uses the supplied root or a child handle. Rename-winner recovery revalidates through that same namespace, never global os functions. Keep 16 MiB artifact bounds, schema-v2 canonical encoding, failed-design-point reuse and all worker/progress behavior. An internally opened handle lives until all workers settle.

Files: `workbench/experiment/runner.go`, `store.go`, `load.go`, `runner_test.go` (private fixture adaptation only), new `root_store_test.go`. Worker-reversible detail: where private decode-from-open-file helpers live; public names/shapes and confinement semantics are fixed.

First RED: table-driven `TestRootedRunStore` in root_store_test.go, using existing runnerBundle/successResult/deterministicRunOptions. Cases: external runs/artifact links, deterministic retarget before I/O, invalid IDs/corrupt cache, scratch failure, supplied-handle lifetime, metadata pointing outside, cancellation/reuse and same-clock byte-equivalence with legacy Run. No new model data.

Checks (new focused test proposed/unrun): `CGO_ENABLED=0 go test ./workbench/experiment -run 'TestRootedRunStore|TestRunWorkerDeterminism|TestRunCancellation|TestRunProgressCounts|TestRunRejects' -count=1`; `go test -race ./workbench/experiment -count=1`.

## S2 Reports

Outcome: GenerateRoot shares existing serializers but performs all directory/temp/write/sync/rename/cleanup operations through pinned root handles. Generate remains a compatibility wrapper. Reports stay individually atomic, not a new set-wide transaction. An old report remains intact on that output's precommit failure; do not delete another call's scratch or previous report.

Files: `workbench/report/report.go`, `workbench/report/report_test.go`.

First RED: `TestGenerateRootConfinement` with outside reports links, deterministic precommit retarget/failure, existing marker reports, no leftover own scratch, supplied-handle lifetime and misleading bundle.Root. Existing deterministic report tests compare scientific bytes; invoke the same Write* functions, not copied serializers.

Checks (new test proposed/unrun): `CGO_ENABLED=0 go test ./workbench/report -run 'TestGenerateRootConfinement|TestReportsAreDeterministicAndExcludeRuntimeMetadata|TestGenerateWritesThreeReports|TestWriteMarkdownStopsAtFirstWriterError' -count=1`; `go test -race ./workbench/report -count=1`.

## S3 Selected EDA exports

Consumes S1's LoadRunRoot, including successful-only integrity validation. Produces ExportSelectedRunRoot with relative output path as specified. Use existing pure GenerateSPICE/GenerateVerilogWithDefaults/GenerateDEFWithDefaults/GenerateLEF/GenerateLiberty output and rooted exclusive file writes; do not call path-based Export* writers or change Module 6 generators. Retain the 16,384-cell bound, PDK caveat, run ID, resolved design and computed file hashes. No evaluator invocation. Existing generated timestamp comments remain; do not assert cross-time artifact determinism. Verify hashes over actual emitted bytes and unchanged scientific/configuration content instead.

Files: `workbench/edaexport/selected.go`, `workbench/edaexport/selected_test.go`.

First RED: `TestExportSelectedRunRootConfinement` with outside exports/run-artifact links, invalid/tampered/non-successful selected runs, destination collision, deterministic mid-write failure/cleanup, supplied-handle lifetime and artifact/manifest consistency. Existing selected-export regressions continue. Construct runs with existing test fixtures; no external EDA tool, Docker, PDK or live device is required for file-generation acceptance.

Checks (new test proposed/unrun): `CGO_ENABLED=0 go test ./workbench/edaexport -run 'TestExportSelectedRunRootConfinement|TestExportSelectedRun' -count=1`; `go test -race ./workbench/edaexport -count=1`.

## S4 Rooted study operations

Consumes S1 RunRoot/LoadRunRecordRoot/LoadRunsRoot, S2 GenerateRoot and S3 ExportSelectedRunRoot. Produces all study root signatures from the decision map. Shared orchestration must use project.LoadRoot plus rooted run preflight/execution/reconciliation, reports and selected export. Preserve existing private evaluator/report injection for deterministic tests, adapting internally as needed instead of duplicating result logic. All relevant path wrappers delegate to the common rooted pipeline; CreateStarterProject is excluded from this existing-project frontier and must not be advertised to HTTP as safe.

Files: `workbench/study/study.go`, `workbench/study/study_test.go`, optional cohesive `workbench/study/root.go` if needed; declare these exact possible paths in Touches. No command/Fyne wiring change is needed because existing path signatures stay compatible.

First RED: `TestStudyRootOperations`, exercising valid root-based validate/run/report/export, optional citation roots, path-option rejection, outside marker/path retarget, owned-handle lifetime and cancellation/report/reconciliation error availability. A root handle survives a rename: subsequent operations must use the original directory handle, not reopen its old name. Existing TestRunRetainsPartialOutcome/TestRunForwardsProgress/structured-success tests remain guards. Test output hashes/result states and absence of outside writes; do not merely mock that a rooted helper was called.

Checks (new test proposed/unrun): `CGO_ENABLED=0 go test ./workbench/study -run 'TestStudyRootOperations|TestRunRetainsPartialOutcome|TestRunForwardsProgress|TestRunAndGenerateReportUseStructuredStudyOutcome' -count=1`; `go test -race ./workbench/study -count=1`; `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools -run 'TestProjectCommandInitAndValidate|TestExperimentAndReportCommandsEndToEnd|TestExperimentSubcommandHonorsCancelledContext' -count=1`.

## Batch, coverage and gates

All four tasks serve BWC-STORAGE. S1 and S2 are independent; S3 depends on S1's exact rooted loader; S4 depends on S1/S2/S3's actual APIs. Longest dependency chain is three tasks. These edges describe consumed functionality, not mere file overlap. Expected write sets are disjoint between tasks; no protected manifests, CI, dependencies, fixed ports, exclusive resources or live authority. The queue owns status/IDs, not this document.

| Accepted requirement | Current state | Frontier / remaining proof |
|---|---|---|
| Go scientific core, progress, speech values, partial outcomes | Implemented; seven-package scoped baseline passed | Preserve existing tests; no duplicate feature tasks |
| Rooted loader and comment-preserving configuration saves | Implemented | Preserve T-001/T-009/T-010; no recreation |
| Root-confined run cache/commit/listing | Path-based, not yet rooted | S1 |
| Root-confined deterministic reports | Path-based I/O | S2 |
| Root-confined selected-run EDA file output | Path-based I/O, pure generators available | S3 |
| Root-confined existing-project study lifecycle | Reopens Bundle.Root | S4 |
| Safe launch-root project registry/creation, auth/API/SSE, frontend assets/forms/results | Approved direction, not delivered | Later bounded batch based on landed APIs; no listener now |
| Orca/Firefox acceptance and full browser behavior | Unverified | Existing numbered manual scripts; no offline waiver |
| Fyne removal | Fully deprecated policy; code remains | Separate retirement scope, not a browser prerequisite or supported target |

Pre-mortem: (1) one rooted helper silently reopens metadata → outside-marker and renamed-root tests in S1/S4; (2) scratch cleanup or publication clobbers committed data → failure/collision/previous-marker cases S1-S3; (3) refactoring storage changes identity or partial-state truth → existing deterministic/progress/integrity tests plus S4 error-state checks. Global invariant: no scientific values, schema versions, run IDs, golden data, tolerances or evidence labels change.

The worker begins with its named focused failing test, records RED, implements minimally and records GREEN/race evidence. New tests/fixtures are local task deliverables, not external blockers. Existing repository Full check remains required at actual integration/closeout; scoped green is not full-main qualification. No test code or production code is written in this planning pass.

## Decision closure and handoff

Planner-resolved: explicit root authority/lifetimes, signatures, option conflicts, namespace/temp/commit behavior, export collision and timestamp limits, partial-analysis compatibility and Linux threat boundary. Worker-reversible: private helper factoring and deterministic test-hook placement within declared files. Owner/external-blocked: none for this offline slice. No new browser/product approval is inferred from this plan.

Saved under goal `BWC-STORAGE` (priority 3, focus false), through literal JSON and `autowork tasks`:

| Task | Outcome | Placement / dependency |
|---|---|---|
| `T-011-ebf984dd` | Rooted experiment store/run/load/list | Now; independent |
| `T-012-f1aebb7e` | Rooted report output | Now; independent |
| `T-013-447af517` | Rooted selected-run export | Next; T-011 |
| `T-014-b8c042aa` | Rooted study operations | Next; T-011, T-012, T-013 |

The four task records are the sole work queue. Existing 15 completed IDs, receipts, goal status, ranks and focus remain untouched. No successor planning loop is created; a later request can plan remaining browser boundaries from actual landed results.

Builder check: the initial reviewer timed out and supplied no findings; no approval is inferred from that timeout. A bounded independent read-only builder review then reported **nothing important left**. It checked the four saved contracts and targeted current source, confirming exact producer/consumer signatures, root lifetimes, citation handling, availability/error semantics, publication limits, locally creatable fixtures/checks and independent S1/S2 lanes. Artifact: `/tmp/fecim-storage-builder-bounded.md` (local planning review, not a product receipt). No material finding remains undisposed: none was raised in the completed review; private helper factoring remains explicitly worker-reversible.

Planning receipts: the existing seven-package headless baseline passed before persistence. Final CLI task validation passed with 19 tasks, 6 goals and no drafts; all four record readbacks, source/dependency/write-set checks, the policy link/anchor and scoped whitespace checks passed. Prepare reports READY and lists T-011/T-012 as the new worktree frontier, with no in-flight contracts. Older GOAL_WITHOUT_WORK diagnostics concern five prior goals whose tasks are closed but aggregate goal qualification has not been established here; their statuses are preserved rather than fabricating met/waiver records or adding duplicate work to silence the checker. Legacy consumer/documentation diagnostics likewise remain outside this storage slice. Proposed new focused/race/full/security/Orca tests were not executed; tests that do not yet exist are not a measured zero or a passing baseline.

Next action, only under later execution authority: T-011's `TestRootedRunStore` RED test; T-012 is an independent lane. S3/S4 wait for their actual API prerequisites. The new plan is staged for tracked accepted-intent discovery; the new goal/task records are uncommitted CLI working-tree writes. No code/tests/fixtures/dependencies, services, workers, commits or publication actions were performed. TDD: N/A for this documentation/task-only batch.
