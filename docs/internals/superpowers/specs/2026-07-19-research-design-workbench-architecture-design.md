# Research and Design Workbench Architecture

Date: 2026-07-19
Status: Approved for implementation planning

## Context

FeCIM Lattice Tools currently presents itself as an educational simulator, research-validation workspace, EDA toolkit, and public showcase. That breadth obscures the primary product and leaves the seven application modules as separate destinations rather than parts of one reproducible design workflow.

The revised product purpose is a literature-calibrated, pre-silicon FeCIM research and design workbench. It should turn a design hypothesis into reproducible simulation evidence, design trade-offs, and exportable artifacts. Educational material remains important, but its role is to explain models, assumptions, and results rather than define the architecture.

The existing Go models, validation data, tests, Fyne application, and export code are valuable assets. The redesign must reuse validated behavior rather than replace it with an unproven rewrite.

## Product Decision

The primary workflow is:

```text
Hypothesis
  → baseline device/array/circuit design
  → deterministic parameter sweep
  → simulation results
  → feasibility and trade-off analysis
  → provenance and validation review
  → report and selected-design exports
```

The product targets literature-calibrated pre-silicon design-space exploration. It does not claim foundry readiness or measured-device fidelity. User-provided experimental datasets may improve calibration later, but they do not change the initial trust boundary.

The first vertical slice is a device/array/circuit configuration that produces a parameter sweep and trade-off report. Full inference and EDA workflows remain downstream extensions.

## Goals

1. Make reproducible research and design the primary user experience.
2. Provide one headless Go core used identically by CLI and Fyne clients.
3. Store projects, resolved runs, provenance, reports, and exports in inspectable files.
4. Reuse existing validated physics and circuit models through thin adapters.
5. Preserve failures, warnings, units, assumptions, and evidence levels in results.
6. Produce feasible-design and Pareto trade-off views from deterministic sweeps.
7. Keep the seven current modules available as expert model workspaces.
8. Migrate without interrupting the current application or rewriting validated models.

## Non-Goals

The first vertical slice will not add:

- A Rust rewrite or Rust component
- A local or remote HTTP/gRPC service
- A database or proprietary project container
- Distributed execution
- An optimizer framework
- Foundry PDK integration or foundry-ready claims
- New physics models solely to support the workflow
- Full inference or complete EDA automation
- Removal of the existing module workspaces

Go remains the implementation language. Rust is reconsidered only for a benchmark-proven kernel bottleneck that cannot be addressed adequately in Go. Scientific validity comes from models, calibration, provenance, and validation—not the systems language.

## System Architecture

```text
                         Project Bundle
                               |
                               v
                    Headless Research Core
          configure → sweep → run → analyze → validate
                               |
                  results + provenance + report
                               |
                               v
                     Selected-Run Exporters

 Existing validated models                       Existing exporters
 device / array / circuits  <── thin adapters ──> SPICE / Verilog / etc.

               CLI                              Fyne Workbench
      automation and reproduction        interactive research and design
```

### Boundary rules

- Numerical model packages contain numerical behavior. They do not own GUI state, project storage, reports, or workflow navigation.
- The research core owns project loading, resolved configuration, sweep expansion, execution, persistence, comparison, and reporting.
- Thin adapters translate a resolved design point into calls to existing model APIs. They do not duplicate equations or maintain a second set of defaults.
- Fyne and CLI invoke the same research-core operations and consume the same persisted results.
- Exporters accept selected completed runs. They do not rerun simulations or silently resolve different defaults.
- Educational documentation and expert workspaces may explain or inspect a model, but they do not become runtime dependencies of the headless core.

## Initial Package Shape

The implementation plan should introduce the fewest packages that preserve these boundaries:

```text
workbench/
  project/       project bundle loading, validation, and default resolution
  experiment/    sweep expansion, run execution, persistence, and analysis
  report/        JSON, CSV, and Markdown report generation
  fecim/         thin device/array/circuit evaluator over existing model APIs
```

The experiment package should accept a small evaluator function contract rather than a hierarchy of model interfaces. The initial FeCIM evaluator is the only required implementation. Additional interfaces or plugin systems are deferred until multiple real implementations require them.

The existing launcher remains the delivery surface. Command examples use `fecim` as shorthand for the installed executable; the first slice does not require renaming binaries.

## Project Bundle

A project is a plain directory:

```text
project/
├── project.yaml
├── design.yaml
├── sweep.yaml
├── inputs/
├── runs/
├── reports/
└── exports/
```

### `project.yaml`

Contains project-level intent and governance:

- Schema version
- Project name and identifier
- Research hypothesis
- Objective metrics and optimization direction
- Feasibility constraints
- Model selection and model-version requirements
- Citation and dataset references

### `design.yaml`

Contains one baseline device/array/circuit design. Parameter names carry their units explicitly or inherit units from a versioned schema. Internal model calls use canonical SI units. Ambiguous unitless physical quantities are rejected at load time.

### `sweep.yaml`

Defines the initial exhaustive grid sweep:

- Parameter path
- Explicit values or a finite linear range
- Optional design constraints
- Random seed policy

Sweep paths must resolve to fields in the versioned design schema. Duplicate paths, empty value sets, non-finite values, invalid ranges, and combinations above the configured safety limit are rejected before execution. Adaptive search and optimization algorithms are outside the first slice.

### `inputs/`

Holds optional local calibration datasets or references needed by the project. Every used input is recorded by relative path, SHA-256 digest, format, citation key when available, and evidence level.

### `runs/`

Contains immutable run directories. Each committed run stores:

```text
runs/<run-id>/
├── manifest.json
├── resolved-design.json
└── result.json
```

The manifest records:

- Run ID and schema version
- Fully resolved parameters
- Random seed
- Evaluator and model versions
- Repository revision when available
- Input dataset hashes and citation keys
- Start and completion metadata
- Success or failed-design-point status
- Warnings and validation status

The result records metrics with names, values, units, model source, assumptions, and evidence levels. A failed design point records its typed failure instead of pretending that metrics exist.

### `reports/` and `exports/`

Reports are generated aggregate views over immutable runs. The first slice emits JSON, CSV, and Markdown. Exports are generated only from an explicitly selected successful run and retain a reference to that run ID.

## Stable Run Identity and Immutability

A run ID is derived from a canonical serialization of all behavior-affecting inputs:

- Resolved design parameters
- Random seed
- Evaluator and model versions
- Input dataset hashes
- Relevant schema versions

Presentation metadata such as timestamps and project display names does not affect identity. Repository revision is recorded for provenance; it affects identity when the evaluator/model version does not otherwise uniquely identify executable behavior.

Canonical serialization uses sorted JSON object keys, documented numeric encoding, and no map-order dependence. Identical behavior-affecting inputs therefore produce the same ID from CLI and Fyne.

A completed successful or failed-design-point run is immutable and reusable. Deterministic model-domain failures are reusable because identical inputs should produce the same failure. Systemic or transient execution failures do not commit a run; they stop the experiment after preserving previously completed runs. Temporary run directories are removed or recovered on the next invocation.

## Execution Flow

```text
Load bundle
  → validate files, schemas, units, citations, and sweep bounds
  → resolve defaults into a complete baseline design
  → expand the sweep in deterministic parameter/value order
  → derive run IDs
  → reuse committed identical runs
  → evaluate missing design points with bounded concurrency
  → atomically commit each result
  → classify feasible results
  → calculate Pareto-optimal results
  → generate reports
  → optionally export one selected successful run
```

### Determinism

- Sweep expansion order is stable.
- Each stochastic model receives an explicit recorded seed.
- Parallel scheduling does not affect run IDs, persisted ordering, or report ordering.
- Reports sort records by deterministic design-point order and then run ID.
- CLI and Fyne use the same resolved project and experiment operations.

### Concurrency and cancellation

The runner uses bounded local concurrency. The default worker count should be conservative and configurable through the command invocation, not persisted as scientific input because scheduling must not change results.

Cancellation stops scheduling new work, allows in-progress atomic writes to finish or be discarded safely, and preserves committed runs. Resume derives the same IDs and skips committed runs.

## Failure Semantics

Failures are divided into three categories:

1. **Project errors:** malformed files, unknown fields, invalid units, unresolved sweep paths, missing required inputs, invalid constraints, or unsupported schema versions. These stop before execution.
2. **Design-point failures:** a valid design falls outside a model domain, violates a hard physical precondition, or cannot produce required metrics. These commit a failed run and allow the sweep to continue.
3. **Systemic failures:** corrupt project state, incompatible model/result schemas, persistence failure, unavailable required model code, or an unexpected evaluator failure. These halt the experiment without deleting completed runs.

Constraint violations discovered from valid output metrics do not count as execution failures. They remain successful but infeasible runs so users can inspect why the design was rejected.

Models must not silently clamp invalid inputs unless the existing validated model contract explicitly defines and reports that behavior. Any defined clamp appears as a warning in the run result.

## Analysis Model

A run is feasible when:

- Evaluation succeeded
- Every required objective and constraint metric exists
- All configured constraints pass
- No validation rule marks the result unusable

Pareto analysis uses only feasible runs and the objective direction declared in `project.yaml`. A run is Pareto-optimal when no other feasible run is at least as good in every objective and strictly better in one objective.

The first slice reports:

- All successful, failed, feasible, and infeasible points
- Constraint outcomes
- Objective metrics and units
- Pareto membership
- Model warnings
- Evidence-level summary
- Input and model provenance

Sensitivity plots may consume the same results, but sophisticated statistical sensitivity analysis is not required for acceptance of the first slice.

## Trust and Evidence Model

Every reported metric carries:

- Unit
- Producing model or calculation
- Relevant assumptions
- Evidence level
- Source run ID

The initial evidence levels are:

- `literature-backed`
- `experiment-calibrated`
- `simulation-default`
- `derived`

A derived metric also identifies its source metrics. Reports visibly distinguish these evidence levels and never promote a simulation default to a measured or literature-reported fact.

Project citations refer to the existing citation registry by key. The workbench consumes that registry rather than creating a second citation authority.

## User Experience

### Fyne workbench

The primary navigation becomes workflow-oriented:

1. **Define** — hypothesis, objectives, constraints, baseline design, and evidence
2. **Simulate** — sweep definition, run queue, progress, cancellation, and failures
3. **Analyze** — result table, feasibility, Pareto views, and sensitivity views
4. **Validate** — provenance, citations, assumptions, warnings, and evidence gaps
5. **Export** — report generation and selected-run design artifacts

The seven existing modules remain accessible under **Expert Workspaces**. They support detailed model inspection and tuning but are not the main workflow.

### CLI

The existing executable gains equivalent commands:

```text
fecim project init <directory>
fecim project validate <directory>
fecim experiment run <directory>
fecim report generate <directory>
fecim export <directory> --run <run-id>
```

Commands default to machine-readable errors on stderr and a concise human summary on stdout. Report and run files are the automation contract; terminal prose is not.

GUI and CLI operations over the same bundle must resolve identical run IDs and persisted scientific content.

## Migration Strategy

The redesign is a clean v2 architecture with adapters, not a rewrite.

1. Introduce project, experiment, report, and evaluator packages without moving existing model code.
2. Adapt one validated device→array→circuit path using current public model APIs.
3. Add one checked-in example project and the CLI validate/run/report path.
4. Add the Fyne Define/Simulate/Analyze workflow over the same core operations.
5. Add provenance review and selected-run export integration.
6. Keep current module routes as Expert Workspaces.
7. Delete, move, or consolidate old orchestration code only after behavioral parity is proven by tests.

No broad folder reorganization is part of the first slice. Model extraction is allowed only when an adapter cannot call a validated model without importing GUI or command code; such extraction must preserve behavior through existing and new regression tests.

## Testing Strategy

Production behavior follows the repository's red-green-refactor policy.

### Project tests

- Supported and unsupported schema versions
- Unknown and missing fields
- Unit validation
- Missing or hash-mismatched inputs
- Deterministic default resolution
- Sweep path and size-limit validation

### Experiment tests

- Deterministic Cartesian expansion
- Stable identity across map order and client path
- Atomic persistence and committed-run reuse
- Failed-design-point continuation
- Systemic-failure halt
- Cancellation and resume
- Deterministic output under different worker counts

### Model-adapter tests

- Golden metrics for representative device/array/circuit points
- Unit conversions at adapter boundaries
- Warning and domain-failure propagation
- No duplicated or silently changed model defaults

### Analysis and report tests

- Feasibility classification
- Multi-objective Pareto membership
- Missing metric handling
- Units and evidence metadata
- Deterministic JSON, CSV, and Markdown scientific content
- Selected-run provenance in exports

### Client parity tests

- CLI and Fyne resolve the same example project to the same run IDs
- Both clients display or return the same failures and warnings
- Fyne remains responsive while experiments run and uses the established UI-thread mutation rules

The existing model, GUI, validation, citation, and full Go test suites remain required gates.

## First-Slice Acceptance Criteria

A checked-in example project must demonstrate all of the following:

1. A baseline device/array/circuit design loads and validates.
2. A finite grid sweep varies parameters across all three layers.
3. CLI and Fyne derive identical run IDs from the same project.
4. Successful, failed, feasible, and infeasible design points remain distinguishable.
5. Cancellation preserves completed runs, and resume does not repeat them.
6. JSON, CSV, and Markdown reports contain deterministic scientific content.
7. Feasible and Pareto-optimal designs are identified correctly.
8. Every reported metric traces to inputs, model, units, assumptions, evidence level, and run ID.
9. One successful run can be selected for existing export capabilities without rerunning its simulation.
10. Existing regression suites pass without weakening scientific tolerances.

## Decisions Summary

- Primary purpose: research and design; education is explanatory support.
- End-to-end outcome: study, candidate architecture, and export artifacts.
- Fidelity: literature-calibrated pre-silicon exploration.
- Architecture: headless Go core with Fyne and CLI clients.
- Storage: plain project directories and inspectable standard formats.
- Navigation: workflow first, expert module workspaces second.
- Migration: clean v2 boundaries with adapters around validated existing code.
- First slice: configuration → deterministic sweep → trade-off report.
- Rust: deferred unless measurement proves a kernel requires it.
