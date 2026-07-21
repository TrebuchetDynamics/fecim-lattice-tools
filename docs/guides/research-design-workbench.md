# FeCIM Research and Design Workbench

FeCIM Lattice Tools is a **literature-calibrated pre-silicon research and design workbench**. It turns a device/array/circuit hypothesis into reproducible parameter sweeps, feasibility results, Pareto trade-offs, and inspectable reports.

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

A run ID is a SHA-256 hash of its resolved design, seed, evaluator version, schema version, and input hashes. Identical behavior-affecting inputs produce the same ID regardless of worker count.

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

The evaluator composes existing system and peripheral models. It provides literature-calibrated pre-silicon estimates for comparison, not hardware guarantees. Quantitative design claims require calibration against suitable experimental data and independent validation.

Input files must remain inside the project directory, pass their declared SHA-256 hash, and use recognized evidence labels. YAML rejects unknown fields and non-finite physical values.

## Current Scope

Implemented:

- Strict project validation
- Deterministic finite sweeps
- Device/array/circuit evaluation
- Bounded parallel execution, cancellation, resume, and reuse
- Feasibility and Pareto analysis
- JSON, CSV, and Markdown reports
- CLI workflow

Deferred:

- Fyne Define/Simulate/Analyze/Validate workflow and CLI parity checks
- Exporting one selected immutable run through the existing SPICE, Verilog, Liberty, DEF, and LEF tools
- Experimental-data calibration workflows
- Optimizer frameworks, remote services, databases, distributed execution, and Rust kernels
