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

This is a Go preparation primitive, not a saver, CLI/GUI editor or browser API. It does not check a revision against later disk changes or provide atomic persistence/concurrency protection. It adds no confinement claim for runs, reports or exports and no native screen-reader qualification.

The table-driven `workbench/project/configuration_test.go::TestPrepareConfiguration` covers preparation, preservation and rejection cases. Focused offline regression checks are:

```bash
CGO_ENABLED=0 go test ./workbench/project -run 'TestPrepareConfiguration|TestLoad' -count=1
go test -race ./workbench/project -run 'TestPrepareConfiguration|TestLoadRootConfinement' -count=1
```

## Run Identity and Reuse

A run ID is a SHA-256 hash of its resolved design, seed, project model version, evaluator version, schema version, and input hashes. Bump `model_version` when a study's model assumptions change and bump the evaluator version when evaluator behavior changes. Identical behavior-affecting inputs produce the same ID regardless of worker count.

Completed successful and model-domain-failed runs are immutable. Repeating an experiment reuses them. Cancellation preserves committed runs; running the command again resumes the missing points.

Runtime timestamps and repository revision remain in run provenance but do not change deterministic report content.

### Go runner and study progress

Direct callers of `experiment.Run` can set `RunOptions.Progress` to a caller-owned `chan<- experiment.ProgressSnapshot`. Callers of [`study.Run`](../../workbench/study/study.go) can set `study.RunOptions.Progress` to the same channel type; study forwards it unchanged, preserving existing outcome, error, and report behavior. Nil disables notifications. Sends never block simulation; unread or full channels may lose any snapshot, including the final one. The caller must not close the channel while `Run` is active. The runner never closes it. Reconcile notifications with the returned summary/error and committed runs, not channel closure or the last received snapshot.

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

This boundary covers project loading only. It does not establish confinement for project creation, run storage, reports, or exports.

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
