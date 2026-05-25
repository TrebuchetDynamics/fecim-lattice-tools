# Validation

## Audience

This folder is for researchers, graduate students, reviewers, and contributors who need to see how FeCIM Lattice Tools checks its simulation claims. It is not a replacement for silicon measurements. It is the evidence layer for a public educational and research simulator.

## What Validation Means Here

Unit tests show that code paths run. Validation shows whether the modeled behavior respects physics, numerical conservation laws, external tools, or published reference data.

Every strong project claim should map to:

- a command someone can run
- a pass/fail threshold
- a generated artifact or fixture
- a limitation statement when the evidence is partial

Generated validation artifacts are written under `output/validation/` and are intentionally not tracked in git unless they are promoted to a curated fixture under `validation/testdata/`.

## One-Command Reproduction

```bash
bash scripts/reproduce_validation.sh
```

For the public Module 2 crossbar gates only:

```bash
go test -v ./validation/module2/... -run 'TestModule2(KCLConservation|AnalyticalLimits|NgspiceComparisonReport)_PublicValidation'
```

Those tests write:

```text
output/validation/module2/kcl_conservation.json
output/validation/module2/analytical_limits.json
output/validation/module2/ngspice_comparison.json
output/validation/module2/ngspice_comparison.svg
output/validation/module2/ngspice_comparison/*.sp
```

## Current Executable Validation

| Area | Evidence | Command |
|---|---|---|
| Module 1 hysteresis | Literature-backed P-E loop checks against digitized datasets, including Park 2015 HZO | `go test -v ./validation/literature/...` |
| Module 2 crossbar | Kirchhoff current conservation, analytical-limit fixtures, and an optional ngspice comparison report for deterministic small-array resistive fixtures | `go test -v ./validation/module2/...` |
| Module 2 external comparison | NumPy/SciPy and ngspice comparison harnesses where external tools are installed | `go test -v ./validation/external/...` |
| Module 4 circuits | KCL/KVL and sense-chain regression checks | `go test -v ./validation/... -run 'Module4|SenseChain'` |
| Module 6 EDA | Verilog sanity/lint and OpenLane smoke tests where tools are installed | `go test -v ./validation/external/... -run 'Verilog|OpenLane'` |
| Configuration | YAML/JSON validation for array, calibration, Preisach, weight, and OpenLane configs | `go test -v ./validation/configvalidator/...` |

External tool checks are optional by design. If `ngspice`, Yosys, OpenLane, or Python scientific packages are not installed, those tests skip the external execution while keeping structural checks active.

## Module 2 Crossbar Gates

`validation/module2/kcl_conservation_test.go` checks the Module 2 parasitic crossbar solver against Kirchhoff's Current Law.

The test:

- builds 100 deterministic random arrays
- varies array shape, conductance, wire parasitics, and applied voltages
- solves each parasitic matrix-vector multiply
- reconstructs cumulative row and column currents
- checks that every node conserves current
- requires maximum KCL residual below `1e-9 A`
- emits a JSON report with seed, threshold, maximum residual, and worst case

`validation/module2/analytical_limits_test.go` adds deterministic analytical-limit fixtures:

- a 1x1 single-cell case requiring `I = G×V` exactly within `1e-12 A`
- a 2x2 zero-parasitic no-sneak case requiring each bit-line current to equal the independent analytical column sum `Σ_i G_ij×V_j` within `1e-12 A`

`validation/module2/ngspice_comparison_report_test.go` promotes the optional ngspice comparison into generated report artifacts:

- deterministic 1x1, 2x2, and 4x4 resistive crossbar SPICE decks under `output/validation/module2/ngspice_comparison/`
- `output/validation/module2/ngspice_comparison.json` with tool availability, structural checks, branch-current comparison metrics, threshold, and limitations
- `output/validation/module2/ngspice_comparison.svg` visualizing the relative-error threshold when comparison data exists
- quantitative pass criterion: maximum parsed WL source-current relative error must be `<= 1%` when ngspice is installed and emits parseable branch currents

The ngspice comparison is optional. If ngspice is missing, the test writes a structural report with `status: "skipped_ngspice_missing"` and then skips instead of failing. Together these gates prove current conservation and selected analytical limits inside the solver, plus optional small-array SPICE consistency when the external tool is available. They do not prove agreement with fabricated devices.

## Package Structure

| Path | Purpose |
|---|---|
| `literature/` | Digitized reference data and literature-backed physics validation |
| `module2/` | Public Module 2 conservation and crossbar validation gates |
| `external/` | Optional external-tool comparisons, including ngspice and OpenLane checks |
| `integration/` | Cross-module validation between physics, crossbar, inference, and EDA paths |
| `configvalidator/` | Rule-based config validation engine and CLI |
| `testdata/` | Curated fixtures and golden references |
| `output/` | Generated local artifacts, ignored by git |

Core statistical helpers live in the root validation package:

- `literature.go` loads published experimental datasets
- `statistics.go` provides KS, chi-square, RMSE, MAE, and correlation metrics
- `interfaces.go` defines shared validation interfaces
- `readiness_report.go` generates release-readiness summaries

## Trust Boundaries

The repository is currently an educational simulation toolkit. A passing validation suite means the simulator is internally consistent and matches selected external references within declared tolerances. It does not mean the repository reports new silicon measurements.

Known public-facing limits:

- Some literature datasets are digitized from figures and carry digitization uncertainty.
- External comparisons depend on installed local tools and versions.
- MNIST results are pipeline demonstrations unless accompanied by a full training/inference artifact and confusion matrix.
- EDA export validity is staged: syntax and smoke tests are not the same as a clean full physical implementation.

See `validation/PLANNED.md` for the remaining validation roadmap.
