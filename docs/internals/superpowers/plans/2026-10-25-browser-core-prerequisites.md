# Accessible browser workbench — core prerequisites planning round

Status: Planning only; no implementation, commits, dispatch, hosting or publication authorized. This is an autonomous discovery frontier, not a replacement for the recovery program or a full browser implementation plan.

**Latest owner scope — one React application:** [ADR 0005](../../adr/0005-unified-react-hosted-and-local.md) and the [current UI policy](../specs/2026-10-25-accessible-browser-workbench-design.md#react-primary-deprecation-policy) supersede this historical plan's delivery assumptions. Keep Go and CLI. The same React source builds hosted/static with Go/WASM Worker computation and local/embedded with full same-origin server capabilities; hosted never calls localhost. React home replaces Astro. Fyne stays deprecated, is retired module-by-module when its React replacement is qualified at behavior/accessibility parity, then deleted under a separately authorized implementation. This latest removal condition supersedes the intervening no-parity-condition wording, not completed IDs/receipts. No new Fyne features, code removal, implementation or commits are authorized here.

**Goal:** Qualify four bounded core prerequisites of the approved local-first browser direction without changing model equations, scientific identities or existing CLI/Fyne success behavior.

**Architecture:** Retain the headless Go packages and introduce only capability-safe project reads, observable runner counts, study progress propagation and pure numeric presentation. The concurrent RDW-LIFECYCLE frontier's T-006-d5e490fd is the explicit partial-result/availability producer; T-005 is its Fyne consumer. This batch does not duplicate that repair. Do not introduce HTTP, React or a second source of scientific truth in this batch.

**Tech stack:** Existing Go standard library and `gopkg.in/yaml.v3`; traversal-resistant `os.Root` APIs on a patched supported Go toolchain. React/TypeScript/Vite/React Aria remain approved for the later browser client, not dependencies of these tasks.

**Spec:** [Browser design and approved decisions](../specs/2026-10-25-accessible-browser-workbench-design.md); [manual Orca gate](../specs/2026-10-25-accessible-browser-workbench-orca-acceptance.md). Those documents' written-consolidation review status is preserved; this plan does not invent a final-spec approval or manual accessibility results.

## Accepted core requirements

Status: Accepted

Source of acceptance: the owner's explicit approvals recorded as BW-01, BW-04, BW-05, BW-06, BW-08 and BW-13 in the browser design, including the final seven additions. The bounded engineering goal below is a **planner-derived prerequisite**, not a newly approved product direction. This heading records the existing accepted requirements so task-store initialization can reference them; it does not grant implementation permission or approve every detail of the written consolidation.

- Project configuration and input access must stay within explicitly granted roots; check-then-open symlink validation is insufficient.
- Progress and final summaries must come from Go without changing scientific identities or depending on worker completion order for persisted results.
- Cancellation/systemic/report errors must preserve honest access to committed results and distinguish a request from terminal completion.
- Go supplies display and speakable values, including units, signed exponents, precision and unavailable states. The browser is not the scientific calculator.

## Project model and evidence

Revision inspected: `816d17c7a007c620b0e7d834701e9c35e756e6bd`, plus the dirty working tree. Do not infer that HEAD alone covers the current candidate. The `workbench/` production/test paths below were clean at planning time; source fingerprints are recorded below. Many unrelated publication, calibration, GUI and repository-layout edits are staged/unstaged and remain untouched.

| Finding | Class / confidence | Evidence and implication |
|---|---|---|
| Simulation-only, reproducible pre-silicon studies are the product direction; researchers need inspectable assumptions/evidence, not hardware-performance promises. | DOCUMENTED / high | README; core workbench architecture; browser BW-01..BW-04. |
| CLI and Fyne use the same study operations and plain files; the marketing site is a separate Astro project. | OBSERVED / high | `cmd/fecim-lattice-tools/workbench_subcommand.go`; `workbench/study/study.go`; `web/README.md`. |
| Project loader checks input symlinks, then performs ordinary path-based reads; top-level YAML/citation access is also path-based. | OBSERVED / high | `workbench/project/load.go:23-36,95-123,141-174`. This does not establish the browser's capability confinement. |
| Runner already has bounded workers, cancellation, immutable cache reuse and ordered summaries, but no public progress seam. | OBSERVED / high | `workbench/experiment/runner.go`; existing determinism/cancellation tests. |
| Study discards the runner's partial outcome on error and discards a successful summary when report generation fails. | OBSERVED / high | `workbench/study/study.go:184-205`. Browser terminal counts cannot rely on the current wrapper. |
| Numeric metrics retain value, unit, model, assumptions and evidence; no browser speech/display representation exists in inspected core packages. | OBSERVED / high | `workbench/experiment/types.go:26-33`; `workbench/fecim/evaluator.go:63-66`. |
| Existing core regression baseline passes. | OBSERVED / high | `CGO_ENABLED=0 go test ./workbench/... -count=1`: all six packages passed in this planning occurrence. This is not a full-suite, browser or accessibility result. |
| Recovery Phase 1/2 and the core slice have completion records; architecture, docs/metrics and owner/legal work remain. | DOCUMENTED / medium | Existing recovery-ledger IDs TR-ARCH-02/03, TR-DOC-01, TR-METRIC-01 and TR-LICENSE-01. No fresh complete recovery/release qualification claimed. |
| Release/JOSS preparation has separate human, provenance and publication gates. | DOCUMENTED / high | `docs/internals/release/research-software-readiness.md`; no submission/release work belongs here. |
| A browser release is not yet safe merely after these four tasks. | INFERRED / high | Run-store, report, export, creation and study persistence still need end-to-end capability adaptation; no listener may be added on the strength of project reads alone. |

### Goal classes and prioritization

- **Accepted:** reproducible scientific workbench; local-first accessible React browser client; retain Go/CLI; no hosted compute. Fyne is fully deprecated under the subsequent owner clarification above, not a required retained UI target. Do not reopen the stack choice.
- **Derived engineering (planner call):** BWC-FOUNDATIONS — qualify the four core seams below. Confinement and truthful lifecycle precede browser convenience; pure presentation is an independent accessibility prerequisite.
- **Existing accepted work, not reprioritized:** technical recovery and release/provenance tasks retain their IDs and scope. They have concurrent dirty work; do not duplicate or mark complete based on this pass.
- **Proposed strategic / excluded:** hosted compute, collaboration accounts, database storage, physics/language rewrite and a WASM demo. No runnable task is created for them.

Expected value is qualitative: the four seams have direct approved requirements, strong source evidence, offline tests and no dependency on installing a browser stack. This is preferable to speculative frontend scaffolding or duplicating completed P0 recovery tasks. It is not a measured ROI or performance claim.

### Input binding

SHA-256 fingerprints for the unchanged inspected inputs:

| Path | SHA-256 |
|---|---|
| `workbench/project/load.go` | `caa68c25c5df087a941caecdded146462b6f700ae069d17e16ec35cccdfadf77` |
| `workbench/project/types.go` | `43622e3bdb90ed439ad5bdf636e869f65a9e12d34df280246fab6e89599f9bf2` |
| `workbench/experiment/runner.go` | `2a090f26492ac47d3e1d9c6c1be0182c92f56d0746b24c9c3ba243a68eef66ce` |
| `workbench/experiment/types.go` | `f56c649b17cf398f5f4f31572fd209d86a1e31f32bbf8be40547f16dec8d2a01` |
| `workbench/study/study.go` | `29f192005d17dbfdb4c6a33bee56da3cb75b0b9b66afb97269a35b851325e6b1` |
| `workbench/experiment/runner_test.go` | `67671169edeae99814a73ed7eca4a351cec60f2d25535bb13835c1436c7f1ff0` |
| `workbench/study/study_test.go` | `837f99a28c85ec67a77c4406c631f82c2a1e7bb298a753dd50566867655ef114` |

Recheck relevant dirty inputs before implementation; do not invalidate unrelated evidence simply because the session changed.

## Scope, invariants and planner calls

**Goals:** four tested Go capabilities with source-compatible existing entry points and explicit partial-result semantics. **Non-goals:** a launchable web UI/API, YAML editing, frontend tooling, end-to-end persistence sandbox, publication, a new scientific format or expert-module migration.

**Anti-goals:** claiming all filesystem paths are confined after only loader work; treating missing metrics as zero; claiming cancellation before workers settle; blocking the solver on UI progress; changing run IDs/artifacts for presentation; attributing unrun tests or human speech review as passed.

Planner calls (engineering choices, not new owner approvals):

1. Use `*os.Root` handles directly instead of creating a speculative generic virtual filesystem. Caller owns supplied handles; legacy path wrappers open/close their own handles. `Bundle.Root` remains informational/legacy-compatible and grants no future API authority.
2. Add `project.LoadRoot(root, citations *os.Root) (Bundle, error)`, where nil citations preserves today's optional citation-existence checks. `Load(path, LoadOptions)` remains and delegates after explicitly opening its project/citation directories. No unrestricted path fallback inside `LoadRoot`.
3. Use an optional send-only buffered progress channel, not callbacks into browser code. A single runner collector emits snapshots with nonblocking sends; an unread channel cannot block, and the caller must not close it while Run is active. The runner never closes caller-owned channels. Nil preserves current behavior. A lost update, including final progress, is repaired from the authoritative returned outcome/state, not interpreted as completion.
4. New `experiment.ProgressSnapshot` has exactly five integer fields: `Total`, `Completed`, `Reused`, `Failed`, `Cancelled`, with lower-case JSON names. Completed includes reused/failed committed points; the latter counts are subsets. Cancelled counts unresolved abandoned points only after actual context cancellation has settled. Systemic error does not invent scientific failed/cancelled records. Mid-run snapshots need not be byte-identical across schedules.
5. Study forwards `RunOptions.Progress` unchanged and leaves result/error/report behavior to the explicit [RDW-LIFECYCLE](2026-07-19-research-design-workbench-core.md#accepted-lifecycle-reliability-frontier) producer T-006-d5e490fd. That record owns partial-result preservation, ResultsState and AnalysisAvailable. T-004 must not duplicate those changes or claim to produce them; its field-only progress addition merges with that independent study repair.
6. Presentation is a new pure `workbench/presentation` package. Use canonical numeric values for sorting and exact scientific records; presentation never mutates them. V1 formatting uses Go's shortest round-trippable decimal representation (`strconv.FormatFloat` with `g`, `-1`, `64`), avoiding invented rounding/precision. Expand signs/exponents and unit vocabulary deterministically; no browser unit parsing. A declared missing/unsupported/nonfinite value is not a zero.

## Execution frontier and dependencies

Canonical allocation: F1 = `T-001-6be26d93`; F2 = `T-002-39b1f21c`; F3 = `T-004-c1000a0d`; F4 = `T-003-81f43547`, all under `BWC-FOUNDATIONS`. The CLI-generated records in `tasks/items/` are authoritative. Symbolic labels F1–F4 below describe outcomes, not a second status ledger.

| Label | Outcome | Placement | True prerequisites | Expected write set |
|---|---|---|---|---|
| F1 | Confine project/config/input/citation reads through root handles | Now, rank 1 | None | `workbench/project/load.go`, `workbench/project/load_test.go` |
| F2 | Count-only nonblocking runner progress with deterministic artifacts | Now, rank 2 | None | `workbench/experiment/runner.go`, `workbench/experiment/progress.go`, `workbench/experiment/progress_test.go` |
| F3 | Forward count-only progress through study orchestration | Next, rank 1 | F2 | `workbench/study/study.go`, `workbench/study/study_test.go` |
| F4 | Pure Go display/speakable metric presentation | Now, rank 3 | None | `workbench/presentation/value.go`, `workbench/presentation/value_test.go` |

F1, F2 and F4 have disjoint write sets within this batch. F3 consumes F2's new exact progress type/option, so its dependency is semantic. T-006-d5e490fd independently owns study partial-result/availability repair; its shared study.go/test edits require normal merge/review coordination, not a fabricated dependency. T-005-2f8387e5 depends on and consumes T-006, not F3. Keep this final explicit ownership stable; do not restore the earlier overlapping T-004 proposal. No chain longer than two, no live resources, no exclusive port/device and no protected manifest/CI writes. All future tasks retain existing repository-required full/race gates at actual implementation closeout; this planning pass does not run them or authorize a commit.

### F1 — Root-scoped project reader

**Produces:** `project.LoadRoot(root, citations *os.Root) (project.Bundle, error)`; root is already-open project scope, citations is an optional separately authorized read-only handle. Supplied handles remain usable after return. `Load` remains compatible. The method may use `root.Name()` as descriptive Bundle.Root metadata but all reads use handle-relative APIs.

**First RED seam:** add `TestLoadRootConfinement` in `workbench/project/load_test.go`. Table cases cover outside symlinks for each YAML file and an input, symlink retarget before open, nested in-root symlinks, invalid/oversized documents and optional citation scope. Use deterministic barriers/seams when needed; no sleep-race success claims. Assert no outside marker bytes were read and errors leave fixture files unchanged. The existing strict/hash/expandable-bundle cases remain the compatibility guards.

**GREEN scope:** extract reusable decode-from-reader logic while retaining strict known-field/single-document and 1 MiB checks; read input/hash data through project handle, citations through citation handle, and bound reads using opened-file information. Treat supplied handles as caller-owned. `Load` explicitly opens legacy directories and delegates; no HTTP or generic filesystem framework.

**Checks:** `CGO_ENABLED=0 go test ./workbench/project -run 'TestLoadRootConfinement|TestLoad' -count=1`; then package regression and `go test -race ./workbench/project`. The new focused test is proposed/unrun.

### F2 — Count-only runner progress

**Produces:** `experiment.ProgressSnapshot{Total, Completed, Reused, Failed, Cancelled int}` with JSON keys `total`, `completed`, `reused`, `failed`, `cancelled`; `experiment.RunOptions.Progress chan<- experiment.ProgressSnapshot` (nil optional). No new fields are inserted into canonical manifests/results.

**First RED seam:** add table-driven `TestRunProgressCounts` in a focused `progress_test.go`, reusing `runnerBundle`, `successResult` and `deterministicRunOptions` from the existing package tests. Cover all-reused, mixed committed failure/success, cancellation with committed records, nil and unread/full channels, and multiple worker schedules. Compare actual artifact bytes and ordered outcomes with/without observers under the same deterministic clock. Validate invariants rather than exact intermediate SSE histories.

**GREEN scope:** centralize emission in runner collection/reuse accounting; copy small snapshots and use nonblocking channel sends. Drain/reconcile worker outcomes before terminal accounting; an already committed point must not be counted cancelled. Derive totals from expanded candidates, and never include other cached design points. Do not add SSE/server or mutate evaluator inputs. Callback/channel clients cannot affect run identity.

**Checks:** `CGO_ENABLED=0 go test ./workbench/experiment -run 'TestRunProgressCounts|TestRunWorkerDeterminism|TestRunCancellationPreservesAndReusesCommittedRuns' -count=1`; `go test -race ./workbench/experiment -run 'TestRunProgressCounts|TestRunBoundsConcurrentEvaluatorCalls|TestRunCancellationPreservesAndReusesCommittedRuns' -count=1`. New focused test proposed/unrun.

### F3 — Study progress propagation

**Consumes:** F2's exact `experiment.ProgressSnapshot` and `experiment.RunOptions.Progress`. **Produces:** `study.RunOptions.Progress chan<- experiment.ProgressSnapshot`, forwarded unchanged without closure; unchanged public study.Run signature. The channel is caller-owned. It does not produce partial outcomes for T-005; T-006 explicitly owns that result/availability contract.

**First RED seam:** add `TestRunForwardsProgress` in `study_test.go`: a deterministic valid study exposes the core counts; nil/full/unread observers cannot block or change returned error identity; study never closes the channel. Existing `TestRunAndGenerateReportUseStructuredStudyOutcome` guards successful report behavior. Use private deterministic in-package seams if needed, not a live endpoint or sleep-based race.

**GREEN scope:** add the option and pass it through without browser math or duplicate lifecycle counters. Preserve T-006's concurrent partial-result/availability/error work. Shared files need ordinary merge/review, not a duplicate repair or an artificial dependency.

**Checks:** `CGO_ENABLED=0 go test ./workbench/study -run 'TestRunForwardsProgress|TestRunAndGenerateReportUseStructuredStudyOutcome' -count=1`; `go test -race ./workbench/study -count=1`; `CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools -run 'TestExperimentSubcommandHonorsCancelledContext|TestExperimentAndReportCommandsEndToEnd' -count=1`. New focused test proposed/unrun; CLI regression names were confirmed in current sources.

### F4 — Scientific value display and speech

**Produces:** `presentation.Value{Display, Speakable, Availability, Reason string}` and `presentation.FormatMetric(metric experiment.Metric, unavailableReason string) presentation.Value`. If unavailableReason is nonempty, availability is `unavailable`; otherwise finite supported values use `available`. Unsupported/nonfinite values return explicit unavailable reason. Evidence/model/assumptions remain on the existing metric; consumers must retain them alongside this presentation, not replace the record.

**First RED seam:** add table-driven `TestFormatMetricSpeech` in `workbench/presentation/value_test.go`. Fixtures cover `um2`, `pJ`, `ns`, `dB` from the evaluator, plus the approved `µC/cm²`, `A` and `Ω` examples, zero, negative values, signed exponents, explicit missing, unknown units and nonfinite inputs. Compare input metric before/after to prove no mutation; do not introduce reference device measurements.

**GREEN scope:** shortest round-trip representation with no invented rounding; scientific notation is expanded to spoken mantissa, “times 10 to the minus/plus” exponent and plural unit vocabulary. Display preserves the stated scientific unit, with a defined mapping for human-facing `um2` to square-micrometer notation. Evidence remains text supplied by existing core labels. Empty unknown units are not guessed. Automated string tests prove the contract, not Orca pronunciation.

**Checks:** `CGO_ENABLED=0 go test ./workbench/presentation -run TestFormatMetricSpeech -count=1`; `go test -race ./workbench/presentation -count=1`; existing evaluator tests remain regression guards. These commands target a new package/test and are proposed/unrun until implementation.

## Coverage map and deferred roadmap

| Requirement | Current classification | This frontier / remaining proof |
|---|---|---|
| Reproducible sweep, cache integrity, analysis, reports and selected-run export | Implemented; scoped core baseline verified in this pass | Preserve existing tests; do not duplicate implementation |
| Capability-confined project reads | Not implemented | F1; complete storage confinement still requires separate follow-up |
| Count-only progress with unchanged artifacts | Not implemented | F2 |
| Correct partial terminal analysis | Partially implemented: runner supports partial data, wrapper loses it | RDW-LIFECYCLE T-006 is the producer and T-005 consumes it; F3 only forwards progress |
| Go display/speech values | Not implemented | F4; actual pronunciation remains a manual gate |
| Local auth/Host/Origin/CSRF/SSE, safe project creation, run/report/export confinement | Not implemented for browser use | Later evidence-bound batch; no listener in this frontier |
| ETag + comment-preserving field edits, server paging/filter/sort, comparisons, chart summaries, frontend/API generation and asset policy | Accepted direction, not implemented | Existing design owns requirements; decompose after core seams are qualified |
| Focus, nonvirtualized tables, one-stop plots, final announcements, complete Orca task scripts | Required but unverified | Manual scripts already saved; no human acceptance invented |
| Full release and JOSS readiness | Unverified / externally gated | Existing release owner; not a browser-core task |

This bounded goal finishes when F1–F4 meet offline criteria; it is **not** the full browser milestone. No self-replenishing planning chain or scheduled successor is created. A subsequent planning request should use landed core reports/signatures to plan the remaining storage/HTTP/UI batch, not reopen approved product choices. Nothing in the deferred map is presented as selection-ready work.

## Pre-mortem and verification

1. A loader-only fix is mistaken for full sandboxing: F1 tests its exact read boundary; all records exclude serving HTTP until every persistence path is adapted.
2. Progress/cancellation changes corrupt identity or lose committed points: F2 compares scientific artifacts under worker counts and settled accounting; F3 proves transparent channel forwarding; independent T-006 owns partial-result integrity and T-005 consumes it.
3. Speech hides sign/unit/availability while string snapshots pass: F4 pins data semantics, but Orca tasks 4 and 9 remain mandatory human gates before accessibility claims.

Future workers begin with the named focused failing test, record actual RED, implement minimally, run GREEN/race and preserve full-suite obligations. No golden-data regeneration, tolerance weakening, model/evidence relabeling, or publication changes. Repair a dependency that blocks a task's own checks only within its existing behavioral contract, not by silencing a failure.

Planning verification: existing six-package headless baseline passed; new tests do not exist and were not run. Task validation, independent builder findings and final local-link checks are recorded in the handoff below after persistence. TDD: N/A for this documentation/task-only planning batch. No code, tests, fixtures, manifests, dependencies or host configuration are changed.

## Queue preparation and handoff

Preparation initially returned `NO_STORE`, state `none`, no accepted headings, no runner or in-flight attempts. No root TODO+goals conversion was available; retained HEAD/working-tree legacy-consumer diagnostics still identify TODO references. Existing legacy history, staged moves and recovery IDs are preserved without editing them or accepting consumers. New `tasks/` records cover only this justified frontier; they do not pretend to migrate/close the historical recovery backlog.

The accepted heading above bootstraps only these already approved core requirements. Store operations use literal JSON through `autowork tasks`, never manual task-file writes. Preserve legacy diagnostics as unresolved inventory, not permission to force migration or repair consumer code during planning. No commit is made because the owner's no-commit restriction continues; committed-store list/why cannot prove these uncommitted records selection-ready. Saving worker-kind contracts is not dispatch or execution authority.

### Final saved handoff

- F1/T-001, F2/T-002 and F4/T-003 are complete open Now contracts with independent write sets; F3/T-004 is Next and waits on T-002. They are structurally ready in the worktree only; no committed-store eligibility, worker ownership admission or delivery is claimed.
- Next proposed action, only after execution authorization: T-001, start `workbench/project/load_test.go::TestLoadRootConfinement`, run focused RED, then implement the root-scoped reader. F2 and F4 are independent lanes if separately authorized. Preserve current goal priority/focus ordering; do not override other goals to run this suggestion.
- Concurrent planning first reused T-004 for partial outcomes, then created an explicit T-006-d5e490fd producer. Final ownership is now fixed: T-002 → T-004 is progress/forwarding, while T-006 → T-005 is partial-result/availability → Fyne presentation. T-004 keeps its allocated ID and progress-only scope; T-006 is not duplicated. The other planner updated T-005's dependency to T-006. No RDW/HZO record was edited by this pass.
- Read-only builder review caught the interim cross-frontier mismatch. A final independent review of live records confirmed the two corrected chains above and reported **nothing important left**; transcript `/tmp/fecim-bwc-builder-final-live.md`. This is a local planning artifact, not implementation/accessibility/release evidence. Earlier review results are superseded wherever the producer contract changed.
- `autowork tasks validate --worktree .` passed with no drafts after reconciliation; the shared store additionally contains concurrent HZO/RDW records, not created by this batch. `prepare --no-commit` reports READY with uncommitted records and legacy consumer/documentation diagnostics. READY does not mean those diagnostics were repaired or historical backlog conversion occurred.
- Remaining legacy consumer inventory includes `.pi/development-loop.json:32` in retained HEAD, `scripts/agent-test-loop.sh:110`, `scripts/cognee-ingest-full.py:264` and `scripts/repository_layout_test.go:27,118,142`. No force, consumer acceptance, code repair or old-ledger mutation was performed. Preserve active owner layout changes; do not add a migration task duplicating that work.
- The new accepted plan was staged so prepare's tracked-document inventory could discover its existing accepted requirements. Only this pass's new plan was added to the index; other staged content was neither committed nor unstaged. Task records remain CLI-written working-tree changes. No implementation or commits.
- Existing six-package headless baseline and source fingerprints above remain the product evidence for this planning occurrence. New focused/race/full/UI/Orca checks are not executed. Final local-link, task-source, dependency and whitespace checks cover the saved planning artifacts, not runtime behavior.
- No new owner product decision is needed for this bounded frontier. Written browser-consolidation review, full browser implementation planning and actual human Orca qualification remain separate later gates. No successor planning daemon, runner or schedule is created.
