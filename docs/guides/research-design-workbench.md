# FeCIM Research and Design Workbench

FeCIM Lattice Tools is a **simulation-only research and design workbench**. It turns a device/array/circuit hypothesis into reproducible parameter sweeps, feasibility results, Pareto trade-offs, and inspectable reports. Current energy, area, and peripheral models are simplified literature-informed estimates, not a calibrated device/layout prediction.

The outputs are simulation estimates. They are not measured silicon, foundry sign-off, or proof that a fabricated device will meet the reported metrics.

## Quick Start

Copy the checked-in example so generated runs stay outside the repository:

```bash
cp -R examples/research-design-workbench /tmp/fecim-study

go run ./cmd/fecim-lattice-tools project validate /tmp/fecim-study \
  -citation-dir citations/papers

go run ./cmd/fecim-lattice-tools experiment run /tmp/fecim-study \
  -workers 2 -citation-dir citations/papers

go run ./cmd/fecim-lattice-tools report generate /tmp/fecim-study \
  -citation-dir citations/papers
```

The experiment command also generates reports after a successful sweep. The separate report command regenerates them from committed runs without rerunning models.

### Desktop starter workflow

In the Fyne workbench, choose an empty folder and select **Configure starter sweep…**. Enter comma-separated values for conductance levels (2–64), array rows (1–4096), and ADC bits (1–16), then create and validate the project before running it. Each list is limited to 16 values and the Cartesian sweep to 10,000 points. Project creation refuses non-empty folders; it does not overwrite existing files. Existing studies can still be opened and validated, but their YAML is not edited by the GUI.

### Desktop simulation results and cancellation

**Cancel simulation** changes the status to **Cancel requested**, not a terminal result. Simulation and other work actions remain disabled until the study operation returns; Cancel is disabled after the request. Committed runs are preserved for resume.

After the operation returns, the status and details show its actual outcome and Go-owned `total`, `reused`, `failed`, `feasible`, and `pareto` counts. `total` counts trusted current-sweep records, including reused and model-domain-failed records; unevaluated points are not failed runs. **Partial** results describe only the collected subset, including subset-only feasibility and Pareto analysis. Missing or integrity-invalid analysis displays `unavailable`, not zero; valid empty partial analysis displays explicit zeros. If no trusted result set is available, record counts also display `unavailable`.

A systemic or report-write error remains **failed**, even if Cancel was also requested. Available results and the error remain readable together. A report-write failure can retain **complete** results and is not labelled partial. Conversely, a late Cancel request does not turn a successful operation return into cancellation.

`TestWorkbenchTerminalOutcome` covers result availability and terminal formatting. `TestWorkbenchCancellationLifecycle` uses a blocked study operation and deterministic barriers to check requested-versus-terminal cancellation, duplicate-work protection, and final counts/details. Run their offline checks with:

```bash
CGO_ENABLED=0 go test ./cmd/fecim-lattice-tools -run '^TestWorkbenchTerminalOutcome$' -count=1
xvfb-run -a go test ./cmd/fecim-lattice-tools -run 'TestWorkbenchCancellationLifecycle|TestResearchWorkbenchContentRenders|TestParseSweepValues' -count=1
xvfb-run -a go test -race ./cmd/fecim-lattice-tools -run 'TestWorkbenchCancellationLifecycle|TestWorkbenchTerminalOutcome' -count=1
```

These automated checks do not establish native screen-reader or Orca acceptance.

## Project Bundle

A project is a plain directory:

```text
study/
├── project.yaml
├── design.yaml
├── sweep.yaml
├── inputs/       optional local calibration inputs
├── runs/         immutable resolved results
└── reports/      deterministic JSON, CSV, and Markdown
```

### `project.yaml`

Defines:

- Research hypothesis
- Model version
- Objective metrics and whether each is minimized or maximized
- Feasibility constraints
- Citation keys from `citations/papers/`
- Optional input paths, SHA-256 hashes, citations, and evidence levels

### `design.yaml`

Defines one baseline HZO device, crossbar array, and peripheral circuit. Physical field names include units, such as `g_min_s`, `read_voltage_v`, and `tia_gain_ohm`. The current evaluator supports HZO and the documented CMOS technology nodes.

### `sweep.yaml`

Defines a finite Cartesian grid. Each parameter uses explicit values or a finite linear range. The first slice supports these paths:

```text
device.conductance_levels
device.g_min_s
device.g_max_s
array.rows
array.cols
array.read_voltage_v
circuit.adc_bits
circuit.dac_bits
circuit.tia_gain_ohm
```

`max_points` prevents accidental unbounded expansion. Every point receives a deterministic recorded seed.

### Go configuration preparation

[`project.ReadConfiguration(root)`](../../workbench/project/configuration.go) captures the original bytes of `project.yaml`, `design.yaml`, and `sweep.yaml` through a caller-owned `*os.Root`. It rejects links and nonregular configuration files and enforces the 1 MiB per-file limit. `ConfigurationSnapshot.Revision()` returns a quoted strong `cfg-v1` ETag over the exact bytes of all three files; comments and line endings affect it. Reading a snapshot does not establish semantic validity.

`project.PrepareConfiguration(root, citations, snapshot, document, changes)` prepares an in-memory candidate without writing files. Each `FieldChange{Path, Value}` targets an existing scalar with a slash path and literal input text, not a YAML fragment. The initial allowlist covers project name/hypothesis and constraint value/operator; design device/array/circuit scalars; and sweep seed/max-points, existing values and range start/stop/count. Identity, schema/model version, objectives, citations, input hashes, constraint metric/unit and sweep parameter paths remain read-only. Collection insertion, deletion and reordering are unsupported.

A patch permits at most 64 changes, with each scalar input limited to 4096 UTF-8 bytes. Go parses base-10 integers with overflow checks, finite float64 values and literal strings. It rejects unsupported or duplicate targets and ambiguous duplicate keys; nonempty patches also reject alias/merge-bearing documents. The full candidate passes the shared strict loader, input-hash/citation checks and sweep validation, with editor-only range/value-count preflight before expansion allocation. Both root handles remain caller-owned; nil citations skips existence checks as with `LoadRoot`.

`PreparedConfiguration.Changed()` is false for empty or semantically unchanged patches. These retain the original bytes and revision without encoding, including CRLF, quotes/comments and a missing final newline. Changed candidates preserve comments, ordering and unexposed supported values through YAML nodes; only the edited document may normalize layout. The other two documents remain byte-identical.

Preparation alone performs no persistence or later-disk revision check. The Go save primitive below adds guarded persistence, not a CLI/GUI editor or browser API.

The table-driven `workbench/project/configuration_test.go::TestPrepareConfiguration` covers preparation, preservation and rejection cases. Focused offline regression checks are:

```bash
CGO_ENABLED=0 go test ./workbench/project -run 'TestPrepareConfiguration|TestLoad' -count=1
go test -race ./workbench/project -run 'TestPrepareConfiguration|TestLoadRootConfinement' -count=1
```

### Go revision-checked configuration saves

[`project.NewConfigurationEditor(root, citations)`](../../workbench/project/configuration_save.go) borrows the same caller-owned roots. Keep them open while the editor is in use, and share **one editor per project handle** among participating saves. `editor.Save(ctx, expectedRevision, document, changes)` applies the preparation rules above and returns `ConfigurationSaveResult{Revision, Changed}`. It changes only the named configuration document; it does not change scientific defaults, run identities or artifacts.

Pass the exact quoted strong revision from `ReadConfiguration`. Even no-ops require it. `ErrRevisionRequired` distinguishes a missing revision; `ErrInvalidRevision` rejects weak, wildcard, unquoted, list and malformed revisions. A stale save returns `*ConfigurationConflict` with `CurrentRevision` and an ordered item for each attempted field: `Path`, literal `Yours`, raw scalar `OnDisk`, `OnDiskAvailable` and `Reason`. Missing, nonscalar or ambiguous on-disk values are unavailable with an explanation, not zero. Review/reapply against a freshly read revision; no force-save bypass is provided. These scalar strings are not yet accessible labels, units or speech forms.

A valid no-op leaves all three files' bytes and modification times unchanged and creates no temporary file. A changed save validates before writing, exclusively creates an unpredictable regular temporary file inside the same root, writes it, preserves the target's permission bits, syncs and closes it, then rechecks all three documents and cancellation immediately before replacing the named file through `root.Rename`. It refuses observed permission changes rather than broadening access. Ownership, ACLs and special mode bits are not preserved. Precommit failures remove only the call's owned temporary file, leaving the original document intact. Successful rename is the commit point: a later cancellation does not turn the committed save into a cancellation error. Callers retain ownership of both handles on success and failure.

The editor mutex serializes participating saves **within that editor only**. Use external coordination or stop other editors while saving: a nonparticipating writer can still change files between the last revision check and rename. This is **not cross-process compare-and-swap**, and reading the three files is not a transactional filesystem snapshot. Replacement behavior is qualified for local Linux filesystems only; no power-loss durability, directory-sync, WASM or other-filesystem guarantee is made. Configuration I/O is root-confined, but this does not establish full confinement for project creation, runs or exports. Report generation has its own confined primitive described below. There is no delivered browser editor, HTTP status mapping or Orca qualification in this primitive.

`workbench/project/configuration_save_test.go::TestSaveConfiguration` covers exact no-op preservation (including CRLF and missing final newline), revision/conflict handling, deterministic concurrent/external edits, root ownership and confinement, persistence faults and commit-point cancellation. Offline checks:

```bash
CGO_ENABLED=0 go test ./workbench/project -run '^TestSaveConfiguration$' -count=1
go test -race ./workbench/project -run 'TestSaveConfiguration|TestPrepareConfiguration|TestLoadRootConfinement' -count=1
CGO_ENABLED=0 go test ./workbench/project -count=1
```

## Run Identity and Reuse

A run ID is a SHA-256 hash of its resolved design, seed, project model version, evaluator version, schema version, and input hashes. Bump `model_version` when a study's model assumptions change and bump the evaluator version when evaluator behavior changes. Identical behavior-affecting inputs produce the same ID regardless of worker count.

Completed successful and model-domain-failed runs are immutable. Repeating an experiment reuses them. Cancellation preserves committed runs; running the command again resumes the missing points.

Runtime timestamps and repository revision remain in run provenance but do not change deterministic report content.

### Go rooted run persistence

[`experiment.RunRoot(ctx, root, bundle, opts)`](../../workbench/experiment/runner.go) borrows a caller-owned `*os.Root` for run cache reads and commits. Keep the handle open until the call returns and all started workers have settled. `bundle.Root` and `root.Name()` are metadata only; neither is reopened. Nil or closed handles return errors. The legacy `experiment.Run` resolves `bundle.Root` to its own root (or nearest existing ancestor) and uses the same execution/store path. A missing legacy workspace is created only when committing a valid result; precommit failures and cancellation leave it absent.

`experiment.LoadRunRecordRoot(root, id)` verifies an exact successful or failed record and returns `false` without error when absent. `experiment.LoadRunRoot(root, id)` requires a verified successful record for export. `experiment.LoadRunsRoot(root)` verifies all committed records and returns them in point-index/run-ID order, ignoring scratch entries. Their string-path counterparts remain available. All rooted operations retain caller ownership on success and failure.

The store pins the `runs/` namespace and each run directory through child handles. Commits exclusively create unpredictable root-relative scratch directories and atomically publish complete run directories. Publication failure verifies any winning record through the same pinned namespace; cleanup removes only owned scratch files, not unrelated temporary entries. Root-confined reads reject links outside their pinned directories. The 16 MiB artifact bounds, canonical bytes, schema, run identities, integrity checks, reuse, model failures and cancellation/progress behavior are unchanged.

This path-confinement behavior is qualified for local Linux filesystems. `os.Root` is not a sandbox against privileged mounts, hard links or same-user malware; no power-loss durability or other-platform guarantee is made. These primitives do not add HTTP/React endpoints or establish confinement for report generation or exports.

`workbench/experiment/root_store_test.go::TestRootedRunStore` covers outside links, metadata retargeting, namespace pinning, publication failure/recovery, scratch ownership, canonical compatibility, failed-point reuse, cancellation and caller-handle ownership. Offline checks:

```bash
CGO_ENABLED=0 go test ./workbench/experiment -run 'TestRootedRunStore|TestRunWorkerDeterminism|TestRunCancellation|TestRunProgressCounts|TestRunRejects' -count=1
go test -race ./workbench/experiment -count=1
```

### Go runner and study progress

Direct callers of `experiment.Run` or `experiment.RunRoot` can set `RunOptions.Progress` to a caller-owned `chan<- experiment.ProgressSnapshot`. Callers of [`study.Run`](../../workbench/study/study.go) can set `study.RunOptions.Progress` to the same channel type; study forwards it unchanged, preserving existing outcome, error, and report behavior. Nil disables notifications. Sends never block simulation; unread or full channels may lose any snapshot, including the final one. The caller must not close the channel while `Run` is active. The runner never closes it. Reconcile notifications with the returned summary/error and committed runs, not channel closure or the last received snapshot.

Snapshots contain only `total`, `completed`, `reused`, `failed`, and `cancelled` JSON counts. `total` is the expanded candidate count. `completed` includes reused and failed committed points; `reused` and `failed` are overlapping subsets of `completed`. After caller-context cancellation, `cancelled` counts unresolved candidates only once all workers settle. Committed points remain completed, and systemic errors alone do not count unresolved work as cancelled. Notifications do not change run IDs, artifact bytes, or ordered outcomes. This is a Go runner/study API, not a CLI or Fyne progress display.

`workbench/experiment/progress_test.go::TestRunProgressCounts` covers count bounds, reuse, settled cancellation, observer backpressure, and artifact/outcome equivalence under different worker counts. Run its headless regression checks with:

```bash
CGO_ENABLED=0 go test ./workbench/experiment -run 'TestRunProgressCounts|TestRunWorkerDeterminism|TestRunCancellationPreservesAndReusesCommittedRuns' -count=1
```

[`TestRunForwardsProgress`](../../workbench/study/study_test.go) covers study-level count forwarding, reuse, model failures, cancellation, cache-error identity, observer backpressure, and caller channel ownership. The existing structured-outcome test guards successful report generation:

```bash
CGO_ENABLED=0 go test ./workbench/study -run 'TestRunForwardsProgress|TestRunAndGenerateReportUseStructuredStudyOutcome' -count=1
```

## Result States

- **Successful:** the evaluator produced valid metrics.
- **Failed:** the point was valid project input but outside a supported model domain.
- **Feasible:** successful, complete, and within every configured constraint.
- **Infeasible:** successful and inspectable, but at least one constraint failed.
- **Unusable:** required objective or constraint metrics were missing, duplicated, or had incompatible units.
- **Pareto:** feasible and not dominated across the configured objective directions.

Failures are preserved instead of silently disappearing. Constraints do not turn a valid simulation into an execution failure.

### Go study outcomes

`study.Run` returns a `RunOutcome` alongside its error. Inspect both: cancellation or a systemic execution error can retain verified current-sweep records, and a report-write error retains the computed results.

- `ResultsUnavailable`: no trusted result set, including option, configuration, or initial record-load failures. Treat zero or unknown `ResultsState` values as unavailable too.
- `ResultsPartial`: a trusted collected subset, possibly empty after cancellation.
- `ResultsComplete`: all current-sweep points collected, even if a late cancellation or report-write error is returned.

Use `AnalysisAvailable` before reading analysis counts. Valid empty partial results have explicit zero counts. Reconciliation integrity errors retain independently trusted records when possible but leave analysis empty and unavailable; missing counts are not zeros. Partial feasibility and Pareto results describe only the collected subset, not the full experiment.

The study snapshots verified expected IDs before execution and reloads only those IDs after workers settle. Unrelated old runs do not enter the outcome or prevent collection. `experiment.LoadRunRecord` verifies an exact committed successful or failed record and distinguishes absence from corruption; export-only `LoadRun` still rejects failed records. Cancellation and systemic execution errors do not regenerate aggregate reports. These availability fields are in-memory Go API metadata, not persisted scientific schema changes.

## Reports

[`report.GenerateRoot(root, bundle, analysis)`](../../workbench/report/report.go) writes through a caller-owned `*os.Root`; keep the handle open during the call. `bundle.Root` is descriptive metadata, not output authority. `report.Generate(path, bundle, analysis)` remains the path adapter and opens/closes its own handle, creating missing project directories as before. Neither entry point changes the JSON, CSV or Markdown serializers or scientific records.

Generation creates and pins a `reports` directory handle. Escaping directory links fail; an output-file link is replaced rather than followed. Each output uses an unpredictable, exclusively created scratch file in that directory, then syncs/closes it before root-relative rename. Directory retargeting does not redirect writes or cleanup. A precommit failure preserves that output's previous report and removes only scratch still identified as belonging to the call; unrelated scratch and substitutions observed at the identity check are left alone. Identity checks and subsequent remove/rename operations are not atomic against other writers: coordinate concurrent namespace changes externally. Earlier reports in the same call may already have been replaced: this is **per-file atomic replacement**, not an all-three-report transaction, cross-process compare-and-swap or power-loss durability guarantee. Replacement is qualified for local Linux filesystems, not WASM or other-filesystem guarantees.

`workbench/report/report_test.go::TestGenerateRootConfinement` covers escaping links, misleading bundle metadata, moved project handles, deterministic directory retargeting/precommit failures, scratch ownership and serializer-byte compatibility. Offline checks:

```bash
CGO_ENABLED=0 go test ./workbench/report -run 'TestGenerateRootConfinement|TestReportsAreDeterministicAndExcludeRuntimeMetadata|TestGenerateWritesThreeReports|TestWriteMarkdownStopsAtFirstWriterError' -count=1
go test -race ./workbench/report -count=1
```

Markdown reports include the project model version, run evaluator/revision metadata, input paths and hashes, citation keys, and each metric's model and assumptions.

Each experiment writes:

- `reports/results.json` — structured design, result, provenance, feasibility, and Pareto data
- `reports/results.csv` — stable tabular metrics for analysis tools
- `reports/report.md` — human-readable hypothesis, counts, objectives, trade-offs, warnings, and evidence

Each metric includes:

- Value and unit
- Producing model
- Assumptions
- Evidence level
- Source run ID through its containing record

Evidence levels distinguish `literature-backed`, `experiment-calibrated`, `simulation-default`, and `derived` values. A simulation default is never presented as a measurement.

## Trust Boundary

The evaluator composes existing system and peripheral models using documented defaults and scaling approximations. Its results support exploratory comparison only, not measured hardware claims or foundry sign-off. The P–E suite includes provisional references and the explicitly synthetic `park2015_synthetic_shape.csv`; the disputed BTO extraction is quarantined and excluded, not smoothed. A green regression suite is not held-out validation of the HZO workbench evaluator. See the [dated correction audit](../research/validation/park-reference-correction.md). Quantitative design claims require device- and circuit-specific calibration against suitable experimental data and independent validation.

Input files must remain inside the project directory, pass their declared SHA-256 hash, and use recognized evidence labels. YAML rejects unknown fields and non-finite physical values; each YAML file is limited to 1 MiB.

[Project loading](../../workbench/project/load.go) reads YAML and inputs through an authorized `os.Root` directory handle, rejecting symlinks that escape it while allowing relative links within it. Optional citation-existence checks use a separately authorized citation-directory handle. Library callers use `project.LoadRoot(root, citations)` and retain ownership of both handles; nil `citations` skips existence checks, not citation-key validation. `project.Load(path, options)` remains available and opens/closes its own handles. `Bundle.Root` is descriptive metadata, not filesystem authority.

This boundary covers project loading, configuration preparation/saves, rooted run persistence, and report generation through the rooted primitives described above. It does not establish confinement for project creation, exports, or a complete study workflow that also generates reports.

## Selected-run EDA export

After `experiment run` has committed successful runs, export one by its full run ID:

```bash
fecim-lattice-tools export /path/to/study -run <64-character-run-id>
```

The command verifies the stored run identity and artifact hashes, rejects failed/tampered runs, and writes SPICE, Verilog, DEF, LEF, Liberty, and a SHA-256 export manifest to `exports/<run-id>/`. It never reruns the simulation and never overwrites an existing export directory. The existing EDA cell/layout generators use their default SKY130 assumptions; the manifest records that the source `tech_node` is not mapped to a foundry PDK. These outputs are design artifacts for exploration, not foundry sign-off.

## Current Scope

Implemented:

- Strict project validation, including sweep expansion and limits
- Deterministic finite sweeps and bounded parallel execution, cancellation, resume, and reuse
- Device/array/circuit evaluation, feasibility and Pareto analysis
- JSON, CSV, and Markdown reports
- CLI and Fyne Configure starter sweep → Validate → Simulate → Analyze → run-integrity/provenance validation workflow over shared study operations
- Integrity-checked selected-run SPICE, Verilog, DEF, LEF, and Liberty export
- Standalone Preisach calibration with a hashed JSON fit artifact ([guide](calibration-workflow.md))

Deferred:

- Automatically applying Preisach calibration artifacts to the workbench evaluator; calibration outputs must currently be applied explicitly
- Optimizer frameworks, remote services, databases, distributed execution, and Rust kernels
