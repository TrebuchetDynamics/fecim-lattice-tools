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

## Run Identity and Reuse

A run ID is a SHA-256 hash of its resolved design, seed, project model version, evaluator version, schema version, and input hashes. Bump `model_version` when a study's model assumptions change and bump the evaluator version when evaluator behavior changes. Identical behavior-affecting inputs produce the same ID regardless of worker count.

Completed successful and model-domain-failed runs are immutable. Repeating an experiment reuses them. Cancellation preserves committed runs; running the command again resumes the missing points.

Runtime timestamps and repository revision remain in run provenance but do not change deterministic report content.

## Result States

- **Successful:** the evaluator produced valid metrics.
- **Failed:** the point was valid project input but outside a supported model domain.
- **Feasible:** successful, complete, and within every configured constraint.
- **Infeasible:** successful and inspectable, but at least one constraint failed.
- **Unusable:** required objective or constraint metrics were missing, duplicated, or had incompatible units.
- **Pareto:** feasible and not dominated across the configured objective directions.

Failures are preserved instead of silently disappearing. Constraints do not turn a valid simulation into an execution failure.

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

The evaluator composes existing system and peripheral models using documented defaults and scaling approximations. Its results support exploratory comparison only, not measured hardware claims or foundry sign-off. The literature P–E suite includes provisional model-generated curves and a directly digitized BTO curve with uncertainty-relaxed thresholds; a green suite is not held-out validation of the HZO workbench evaluator. Quantitative design claims require device- and circuit-specific calibration against suitable experimental data and independent validation.

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
